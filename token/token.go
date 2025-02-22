package token

type TokenType uint

const (
	ERR = iota
	EOF

	// Identifiers and literals
	IDENT  // x, y, name
	INT    // 1032
	FLOAT  // 5.2, 0.23
	STRING // "hello" "world"

	// Operators
	DOT    // "."
	ASSIGN // "="
	PLUS   // "+"
	MINUS  // "-"
	BANG   // "!"
	STAR   // "*"
	SLASH  // "/"
	LT     // "<"
	GT     // ">"

	EQ // "=="
	NE // "!="
	LE // "<="
	GE // ">="

	// Delimeters
	COLON  // ":"
	COMMA  // ","
	SEMCOL // ";"

	// Brackets
	LPAREN // "("
	RPAREN // ")"
	LBRACE // "{"
	RBRACE // "}"
	LBRACK // "["
	RBRACK // "]"

	// Keywords
	FN     // "fn"
	RETURN // "return"
	LET    // "let"
	TRUE   // "true"
	FALSE  // "false"
	IF     // "if"
	ELSE   // "else"
	LOOP   // "loop"
	NULL   // "null"
	BREAK  // "break"
	CONT   // "continue"

	TOTAL // total number of tokens
)

var TokenString = []string{
	EOF:    "eof",
	ERR:    "error",
	IDENT:  "identifier",
	INT:    "integer",
	FLOAT:  "float",
	DOT:    ".",
	ASSIGN: "=",
	PLUS:   "+",
	MINUS:  "-",
	BANG:   "!",
	STAR:   "*",
	SLASH:  "/",
	LT:     "<",
	GT:     ">",
	EQ:     "==",
	NE:     "!=",
	LE:     "<=",
	GE:     ">=",
	COMMA:  ",",
	SEMCOL: ";",
	LPAREN: "(",
	RPAREN: ")",
	LBRACE: "{",
	RBRACE: "}",
	FN:     "fn",
	RETURN: "return",
	LET:    "let",
	TRUE:   "true",
	FALSE:  "false",
	IF:     "if",
	ELSE:   "else",
	COLON:  ":",
	LOOP:   "loop",
	NULL:   "null",
	BREAK:  "break",
	CONT:   "continue",
}

type Token struct {
	Loc  SrcLoc
	Type TokenType
	Word string
}

var keywords = map[string]TokenType{
	"fn":       FN,
	"return":   RETURN,
	"let":      LET,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"loop":     LOOP,
	"null":     NULL,
	"break":    BREAK,
	"continue": CONT,
}

func LookUpKeyword(word string) TokenType {
	if tokType, ok := keywords[word]; ok {
		return tokType
	}

	return IDENT
}
