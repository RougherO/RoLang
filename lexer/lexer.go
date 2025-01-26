package lexer

import (
	"RoLang/tokens"
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
	l := &Lexer{file: file, input: input} // allocating on heap
	l.readChar()                          // read the first char
	return l
}

func (l *Lexer) NextToken() tokens.Token {
	var tok tokens.Token

	l.skipWhiteSpace()

	switch l.char {
	case ';':
		tok = l.makeToken(tokens.SEMCOL, ";")
	case '(':
		tok = l.makeToken(tokens.LPAREN, "(")
	case ')':
		tok = l.makeToken(tokens.RPAREN, ")")
	case '{':
		tok = l.makeToken(tokens.LBRACE, "{")
	case '}':
		tok = l.makeToken(tokens.RBRACE, "}")
	case ',':
		tok = l.makeToken(tokens.COMMA, ",")
	case '+':
		tok = l.makeToken(tokens.PLUS, "+")
	case '-':
		tok = l.makeToken(tokens.MINUS, "-")
	case '*':
		tok = l.makeToken(tokens.STAR, "*")
	case '/':
		tok = l.makeToken(tokens.SLASH, "/")
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(tokens.EQ, "==")
		} else {
			tok = l.makeToken(tokens.ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(tokens.NE, "!=")
		} else {
			tok = l.makeToken(tokens.BANG, "!")
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(tokens.LE, "<=")
		} else {
			tok = l.makeToken(tokens.LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(tokens.GE, ">=")
		} else {
			tok = l.makeToken(tokens.GT, ">")
		}
	case 0:
		tok.Word = "EOF"
		tok.Type = tokens.EOF
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

func (l *Lexer) makeToken(tokenType tokens.TokenType, word string) tokens.Token {
	token := tokens.Token{
		Loc: tokens.SrcLoc{
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

func (l *Lexer) makeErr(message string) tokens.Token {
	return tokens.Token{
		Loc: tokens.SrcLoc{
			File: l.file, // current file
			Line: l.line, // current line
			Col:  l.col,  // current column
		},
		Type: tokens.ERR,
		Word: message,
	}
}

func (l *Lexer) readIdent() tokens.Token {
	var tokType tokens.TokenType

	start := l.offset - 1

	for isAlpha(l.char) { // [A-Za-z_]+
		l.readChar()
	}

	word := l.input[start : l.offset-1]
	tokType = tokens.LookUpKeyword(word) // lookup for keywords: fn, let, return...

	return l.makeToken(tokType, word)
}

func (l *Lexer) readNum() tokens.Token {
	var tokType tokens.TokenType

	start := l.offset - 1

	tokType = tokens.INT
	for isDigit(l.char) { // [0-9]+
		l.readChar()
	}

	if l.char == '.' { // floating point literal
		l.readChar()
		tokType = tokens.FLOAT
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
			l.col = 0
			l.line++
		case '\r':
			l.col = 0
		default:
			return
		}
		l.readChar()
	}
}

func isAlpha(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}
