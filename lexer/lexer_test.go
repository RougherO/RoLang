package lexer

import (
	"testing"

	"RoLang/token"
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
		expectType token.TokenType
		expectWord string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMCOL, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.FLOAT, "10.23"},
		{token.SEMCOL, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FN, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMCOL, ";"},
		{token.RBRACE, "}"},
		{token.SEMCOL, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMCOL, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.STAR, "*"},
		{token.INT, "5"},
		{token.SEMCOL, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMCOL, ";"},
		{token.IF, "if"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMCOL, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMCOL, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMCOL, ";"},
		{token.INT, "10"},
		{token.NE, "!="},
		{token.FLOAT, "9.99"},
		{token.SEMCOL, ";"},
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
