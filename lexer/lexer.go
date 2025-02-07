package lexer

import (
	"RoLang/token"
	"fmt"
)

type Lexer struct {
	file   string
	input  string
	line   uint // current line number
	col    uint // current column number
	offset uint // next position to read
	char   byte // current ASCII character
}

func New(file, input string) *Lexer {
	// Allocating on heap
	l := &Lexer{
		file:  file,
		input: input,
		line:  1,
		col:   1,
	}

	// Read the first char to set the state
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhiteSpace()

	switch l.char {
	case ';':
		tok = l.makeToken(token.SEMCOL, ";")
	case '(':
		tok = l.makeToken(token.LPAREN, "(")
	case ')':
		tok = l.makeToken(token.RPAREN, ")")
	case '{':
		tok = l.makeToken(token.LBRACE, "{")
	case '}':
		tok = l.makeToken(token.RBRACE, "}")
	case ',':
		tok = l.makeToken(token.COMMA, ",")
	case '+':
		tok = l.makeToken(token.PLUS, "+")
	case '-':
		tok = l.makeToken(token.MINUS, "-")
	case '*':
		tok = l.makeToken(token.STAR, "*")
	case '/':
		tok = l.makeToken(token.SLASH, "/")
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(token.EQ, "==")
		} else {
			tok = l.makeToken(token.ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(token.NE, "!=")
		} else {
			tok = l.makeToken(token.BANG, "!")
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(token.LE, "<=")
		} else {
			tok = l.makeToken(token.LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(token.GE, ">=")
		} else {
			tok = l.makeToken(token.GT, ">")
		}
	case 0:
		tok.Word = "EOF"
		tok.Type = token.EOF
	default:
		if isAlpha(l.char) { // check [A-Za-z_]
			tok = l.readIdent()
			return tok
		} else if isDigit(l.char) { // check [0-9]
			tok = l.readNum()
			return tok
		} else {
			tok = l.makeErr(fmt.Sprintf("Unknown token %c", l.char))
		}
	}

	l.readChar() // ready the next character

	return tok
}

func (l *Lexer) makeToken(tokenType token.TokenType, word string) token.Token {
	token := token.Token{
		Loc: token.SrcLoc{
			File: l.file, // current file
			Line: l.line, // current line
			Col:  l.col,  // current column
		},
		Type: tokenType,
		Word: word,
	}

	l.col += uint(len(word))
	return token
}

func (l *Lexer) makeErr(message string) token.Token {
	return token.Token{
		Loc: token.SrcLoc{
			File: l.file, // current file
			Line: l.line, // current line
			Col:  l.col,  // current column
		},
		Type: token.ERR,
		Word: message,
	}
}

func (l *Lexer) readIdent() token.Token {
	var tokType token.TokenType

	start := l.offset - 1

	for isAlpha(l.char) || isDigit(l.char) || l.char == '_' { // [A-Za-z_0-9]+
		l.readChar()
	}

	word := l.input[start : l.offset-1]
	tokType = token.LookUpKeyword(word) // lookup for keywords: fn, let, return...

	return l.makeToken(tokType, word)
}

func (l *Lexer) readNum() token.Token {
	var tokType token.TokenType

	start := l.offset - 1

	tokType = token.INT
	for isDigit(l.char) { // [0-9]+
		l.readChar()
	}

	if l.char == '.' { // floating point literal
		l.readChar()
		tokType = token.FLOAT
		for isDigit(l.char) {
			l.readChar()
		}
	}

	word := l.input[start : l.offset-1]

	return l.makeToken(tokType, word)
}

func (l *Lexer) readChar() {
	if l.offset >= uint(len(l.input)) {
		l.char = 0
	} else {
		l.char = l.input[l.offset]
	}

	l.offset++
}

func (l *Lexer) peekChar() byte {
	if l.offset == uint(len(l.input)) {
		return 0
	}

	return l.input[l.offset]
}

func (l *Lexer) skipWhiteSpace() {
	for {
		switch l.char {
		case ' ':
			l.col++
		case '\t': // assume tab characters take 4 spaces
			l.col += 4
		case '\n':
			l.col = 1
			l.line++
		case '\r':
			l.col = 1
		default:
			return
		}
		l.readChar()
	}
}

func isAlpha(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
