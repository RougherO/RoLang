package token

type TokenType uint

const (
	ERR = iota
	EOF

	// Identifiers and literals
	IDENT // x, y, name
	INT   // 1032
	FLOAT // 5.2, 0.23

	// Operators
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
	COMMA  // ","
	SEMCOL // ";"

	// Brackets
	LPAREN // "("
	RPAREN // ")"
	LBRACE // "{"
	RBRACE // "}"

	// Keywords
	FN     // "fn"
	RETURN // "return"
	LET    // "let"
	TRUE   // "true"
	FALSE  // "false"
	IF     // "if"
	ELSE   // "else"

	TOTAL // total number of tokens
)

var TokenString = []string{
	EOF:    "eof",
	ERR:    "error",
	IDENT:  "identifier",
	INT:    "integer",
	FLOAT:  "float",
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
}

type Token struct {
	Loc  SrcLoc
	Type TokenType
	Word string
}

var keywords = map[string]TokenType{
	"fn":     FN,
	"return": RETURN,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
}

func LookUpKeyword(word string) TokenType {
	if tokType, ok := keywords[word]; ok {
		return tokType
	}

	return IDENT
}
