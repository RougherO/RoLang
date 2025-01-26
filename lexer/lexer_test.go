package lexer

import (
	"testing"

	"RoLang/tokens"
)

func TestNextToken(t *testing.T) {
	input := `
let five = 5;
let ten = 10.23;

let add = fn(x, y) {
	x + y;
};

let result = add(five, ten);

!-/*5;
5 < 10 > 5;

if 5 < 10 {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9.99;
`

	tests := []struct {
		expectType tokens.TokenType
		expectWord string
	}{
		{tokens.LET, "let"},
		{tokens.IDENT, "five"},
		{tokens.ASSIGN, "="},
		{tokens.INT, "5"},
		{tokens.SEMCOL, ";"},
		{tokens.LET, "let"},
		{tokens.IDENT, "ten"},
		{tokens.ASSIGN, "="},
		{tokens.FLOAT, "10.23"},
		{tokens.SEMCOL, ";"},
		{tokens.LET, "let"},
		{tokens.IDENT, "add"},
		{tokens.ASSIGN, "="},
		{tokens.FN, "fn"},
		{tokens.LPAREN, "("},
		{tokens.IDENT, "x"},
		{tokens.COMMA, ","},
		{tokens.IDENT, "y"},
		{tokens.RPAREN, ")"},
		{tokens.LBRACE, "{"},
		{tokens.IDENT, "x"},
		{tokens.PLUS, "+"},
		{tokens.IDENT, "y"},
		{tokens.SEMCOL, ";"},
		{tokens.RBRACE, "}"},
		{tokens.SEMCOL, ";"},
		{tokens.LET, "let"},
		{tokens.IDENT, "result"},
		{tokens.ASSIGN, "="},
		{tokens.IDENT, "add"},
		{tokens.LPAREN, "("},
		{tokens.IDENT, "five"},
		{tokens.COMMA, ","},
		{tokens.IDENT, "ten"},
		{tokens.RPAREN, ")"},
		{tokens.SEMCOL, ";"},
		{tokens.BANG, "!"},
		{tokens.MINUS, "-"},
		{tokens.SLASH, "/"},
		{tokens.STAR, "*"},
		{tokens.INT, "5"},
		{tokens.SEMCOL, ";"},
		{tokens.INT, "5"},
		{tokens.LT, "<"},
		{tokens.INT, "10"},
		{tokens.GT, ">"},
		{tokens.INT, "5"},
		{tokens.SEMCOL, ";"},
		{tokens.IF, "if"},
		{tokens.INT, "5"},
		{tokens.LT, "<"},
		{tokens.INT, "10"},
		{tokens.LBRACE, "{"},
		{tokens.RET, "return"},
		{tokens.TRUE, "true"},
		{tokens.SEMCOL, ";"},
		{tokens.RBRACE, "}"},
		{tokens.ELSE, "else"},
		{tokens.LBRACE, "{"},
		{tokens.RET, "return"},
		{tokens.FALSE, "false"},
		{tokens.SEMCOL, ";"},
		{tokens.RBRACE, "}"},
		{tokens.INT, "10"},
		{tokens.EQ, "=="},
		{tokens.INT, "10"},
		{tokens.SEMCOL, ";"},
		{tokens.INT, "10"},
		{tokens.NE, "!="},
		{tokens.FLOAT, "9.99"},
		{tokens.SEMCOL, ";"},
	}

	lexer := New("lexer_test", input)

	for i, test := range tests {
		tok := lexer.NextToken()

		if tok.Type != test.expectType {
			t.Fatalf("Test[%d] - wrong token type. expect=%d[%q], found=%d[%q]",
				i, test.expectType, test.expectWord, tok.Type, tok.Word)
		}

		if tok.Word != test.expectWord {
			t.Fatalf("Test[%d] - wrong token word. expect=%d[%q], found=%d[%q]",
				i, test.expectType, test.expectWord, tok.Type, tok.Word)
		}
	}
}
