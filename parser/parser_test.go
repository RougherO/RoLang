package parser

import (
	"RoLang/ast"
	"RoLang/lexer"
	"RoLang/token"

	"fmt"
	"strconv"
	"testing"
)

func TestLetStatement(t *testing.T) {
	input := `
let x = 5;
let y = 10.23;
let foobar = 01923;
`
	l := lexer.New("parser_test_let", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", n)
	}

	tests := []struct {
		expectIdent string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, test := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, test.expectIdent) {
			return
		}
	}
}

func TestFunctionStatement(t *testing.T) {
	input := "fn add(x, y) { x + y; }"

	l := lexer.New("parser_test_func_stmt", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", n)
	}

	fn, ok := program.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not *ast.FunctionStatement. got=%T", program.Statements[0])
	}

	if !testIdentifier(t, fn.Ident, "add") {
		return
	}

	if len(fn.Value.Parameters) != 2 {
		t.Fatalf("fn.Value.Parameters has incorrect arity. got=%d", len(fn.Value.Parameters))
	}

	if !testIdentifier(t, fn.Value.Parameters[0], "x") ||
		!testIdentifier(t, fn.Value.Parameters[1], "y") {
		return
	}

	if n := len(fn.Value.Body.Statements); n != 1 {
		t.Fatalf("fn.Value.Body.Statements contain incorrect number of statements. got=%d", n)
	}

	stmt, ok := fn.Value.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("fn.Value.Body.Statements[0] not *ast.ExpressionStatement. got=%T",
			fn.Value.Body.Statements[0])
	}

	if !testInfixExpression(t, stmt.Expression, "x", "+", "y") {
		return
	}
}

func TestReturnStatement(t *testing.T) {
	input := `
return 5;
return 10;
return 10.233;
return x;
`
	l := lexer.New("parser_test_return", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d", n)
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got=%T", stmt)
		} else if word := returnStmt.TokenWord(); word != "return" {
			t.Errorf("returnStmt.TokenWord() not 'return'. got=%q", word)
		}
	}
}

func TestIfStatement(t *testing.T) {
	input := `if x < y { x }`

	l := lexer.New("parser_test_if", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, n)
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.IfStatement. got=%T",
			program.Statements[0])
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if n := len(stmt.Then.Statements); n != 1 {
		t.Errorf("then is not 1 statement. got=%d\n",
			len(stmt.Then.Statements))
	}

	then, ok := stmt.Then.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt.Then.Statements[0] is not ast.ExpressionStatement. got=%T",
			stmt.Then.Statements[0])
	}

	if !testIdentifier(t, then.Expression, "x") {
		return
	}

	if stmt.Else != nil {
		t.Errorf("stmt.Else.Statements was not nil. got=%+v", stmt.Else)
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if x < y { x } else { y }`

	l := lexer.New("parser_test_if_else", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n", 1, n)
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.IfStatement. got=%T", stmt)
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if n := len(stmt.Then.Statements); n != 1 {
		t.Fatalf("then is not 1 statement. got=%d", n)
	}

	then, ok := stmt.Then.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt.Then.Statements[0] is not *ast.ExpressionStatement. got=%T",
			stmt.Then.Statements[0])
	}

	if !testIdentifier(t, then.Expression, "x") {
		return
	}

	if stmt.Else == nil {
		t.Fatal("stmt.Else.Statements was nil.")
	}

	switch block := stmt.Else.(type) {
	case *ast.BlockStatement:
		if n := len(block.Statements); n != 1 {
			t.Fatalf("block is not 1 statement. got=%d", n)
		}

		expr, ok := block.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("block.Statements[0] is not *ast.ExpressionStatement. got=%T", expr)
		}

		if !testIdentifier(t, expr.Expression, "y") {
			return
		}
	default:
		t.Fatalf("stmt.Else is not *ast.BlockStatement. got=%T", block)
	}
}

func TestIfElseIfStatement(t *testing.T) {
	input := `if x < y { x } else if x > y { y }`

	l := lexer.New("parser_test_if_else_if", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n", 1, n)
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.IfStatement. got=%T", stmt)
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if n := len(stmt.Then.Statements); n != 1 {
		t.Fatalf("then is not 1 statement. got=%d", n)
	}

	then, ok := stmt.Then.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt.Then.Statements[0] is not *ast.ExpressionStatement. got=%T",
			stmt.Then.Statements[0])
	}

	if !testIdentifier(t, then.Expression, "x") {
		return
	}

	if stmt.Else == nil {
		t.Fatal("stmt.Else.Statements was nil.")
	}

	switch elseif := stmt.Else.(type) {
	case *ast.IfStatement:
		if n := len(elseif.Then.Statements); n != 1 {
			t.Fatalf("elseif is not 1 statement. got=%d", n)
		}

		if !testInfixExpression(t, elseif.Condition, "x", ">", "y") {
			return
		}

		if n := len(elseif.Then.Statements); n != 1 {
			t.Fatalf("elseif.then is not 1 statement. got=%d", n)
		}

		then, ok := elseif.Then.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("elseif.Then.Statements[0] is not *ast.ExpressionStatement. got=%T",
				elseif.Then.Statements[0])
		}

		if !testIdentifier(t, then.Expression, "y") {
			return
		}

	default:
		t.Fatalf("stmt.Else is not *ast.IfStatement. got=%T", elseif)
	}
}

func TestIfElseIfElseStatement(t *testing.T) {
	input := `if x < y { x } else if x > y { y } else { x + y }`

	l := lexer.New("parser_test_if_else_if", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n", 1, n)
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.IfStatement. got=%T", stmt)
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if n := len(stmt.Then.Statements); n != 1 {
		t.Fatalf("then is not 1 statement. got=%d", n)
	}

	then, ok := stmt.Then.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt.Then.Statements[0] is not *ast.ExpressionStatement. got=%T",
			stmt.Then.Statements[0])
	}

	if !testIdentifier(t, then.Expression, "x") {
		return
	}

	if stmt.Else == nil {
		t.Fatal("stmt.Else.Statements was nil.")
	}

	switch elseif := stmt.Else.(type) {
	case *ast.IfStatement:
		if n := len(elseif.Then.Statements); n != 1 {
			t.Fatalf("elseif is not 1 statement. got=%d", n)
		}

		if !testInfixExpression(t, elseif.Condition, "x", ">", "y") {
			return
		}

		if n := len(elseif.Then.Statements); n != 1 {
			t.Fatalf("elseif.then is not 1 statement. got=%d", n)
		}

		then, ok := elseif.Then.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("elseif.Then.Statements[0] is not *ast.ExpressionStatement. got=%T",
				elseif.Then.Statements[0])
		}

		if !testIdentifier(t, then.Expression, "y") {
			return
		}

		if elseif.Else == nil {
			t.Fatal("elseif.Else.Statements was nil.")
		}

		switch block := elseif.Else.(type) {
		case *ast.BlockStatement:
			if n := len(block.Statements); n != 1 {
				t.Fatalf("block is not 1 statement. got=%d", n)
			}

			expr, ok := block.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("block.Statements[0] is not *ast.ExpressionStatement. got=%T", expr)
			}

			if !testInfixExpression(t, expr.Expression, "x", "+", "y") {
				return
			}
		default:
			t.Fatalf("elseif.Else is not *ast.BlockStatement. got=%T", block)
		}

	default:
		t.Fatalf("stmt.Else is not *ast.IfStatement. got=%T", elseif)
	}
}

func TestPrefixExpression(t *testing.T) {
	prefixIntTests := []struct {
		input    string
		operator string
		right    interface{}
	}{
		{"!a", "!", "a"},
		{"!5;", "!", int64(5)},
		{"-15;", "-", int64(15)},
		{"!5.223;", "!", float64(5.223)},
		{"-10.23;", "-", float64(10.23)},
	}

	for _, test := range prefixIntTests {
		l := lexer.New("parser_test_prefix", test.input)
		p := New(l)

		program := p.Parse()
		checkErrors(t, p)

		if n := len(program.Statements); n != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d", n)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("stmt is not *ast.ExpressionStatement. got=%T", stmt)
		}

		if !testPrefixExpression(t, stmt.Expression, test.operator, test.right) {
			return
		}
	}
}

func TestInfixExpression(t *testing.T) {
	infixTests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", int64(5), "+", int64(5)},
		{"5 - 5;", int64(5), "-", int64(5)},
		{"5 * 5;", int64(5), "*", int64(5)},
		{"5 / 5;", int64(5), "/", int64(5)},
		{"5 > 5;", int64(5), ">", int64(5)},
		{"5 < 5;", int64(5), "<", int64(5)},
		{"5 == 5;", int64(5), "==", int64(5)},
		{"5 != 5;", int64(5), "!=", int64(5)},
		{"5.23 + 5.23;", float64(5.23), "+", float64(5.23)},
		{"5.23 - 5.23;", float64(5.23), "-", float64(5.23)},
		{"5.23 * 5.23;", float64(5.23), "*", float64(5.23)},
		{"5.23 / 5.23;", float64(5.23), "/", float64(5.23)},
		{"5.23 > 5.23;", float64(5.23), ">", float64(5.23)},
		{"5.23 < 5.23;", float64(5.23), "<", float64(5.23)},
		{"5.23 == 5.23;", float64(5.23), "==", float64(5.23)},
		{"5.23 != 5.23;", float64(5.23), "!=", float64(5.23)},
		{"a + a;", "a", "+", "a"},
		{"a - a;", "a", "-", "a"},
		{"a * a;", "a", "*", "a"},
		{"a / a;", "a", "/", "a"},
		{"a > a;", "a", ">", "a"},
		{"a < a;", "a", "<", "a"},
		{"a == a;", "a", "==", "a"},
		{"a != a;", "a", "!=", "a"},
		{"true + true;", true, "+", true},
		{"true - true;", true, "-", true},
		{"true * true;", true, "*", true},
		{"true / true;", true, "/", true},
		{"false > false;", false, ">", false},
		{"false < false;", false, "<", false},
		{"false == false;", false, "==", false},
		{"false != false;", false, "!=", false},
	}

	for _, test := range infixTests {
		l := lexer.New("parser_test_infix", test.input)
		p := New(l)

		program := p.Parse()
		checkErrors(t, p)

		if n := len(program.Statements); n != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, n)
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		testInfixExpression(t, stmt.Expression, test.left, test.operator, test.right)
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	expectStr := "foobar"

	l := lexer.New("parser_test_ident", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not ast.ExpressionStatement. got=%T", stmt)
	}

	if !testIdentifier(t, stmt.Expression, expectStr) {
		return
	}
}

func TestFunctionLiteralExpression(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New("parser_test_func_literal", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n", 1, n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if n := len(function.Parameters); n != 2 {
		t.Fatalf("incorrect number of function parameters. expect=2, got=%d\n", n)
	}

	testPrimaryExpression(t, function.Parameters[0], "x")
	testPrimaryExpression(t, function.Parameters[1], "y")

	if n := len(function.Body.Statements); n != 1 {
		t.Fatalf("function.Body.Statements has more than 1 statement. got=%d\n", n)
	}

	body, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not *ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(t, body.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}
	for _, test := range tests {

		l := lexer.New("parser_test_func_params", test.input)
		p := New(l)

		program := p.Parse()
		checkErrors(t, p)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		function, ok := stmt.Expression.(*ast.FunctionLiteral)
		if !ok {
			t.Fatalf("function is not *ast.FunctionLiteral. got=%T", function)
		}

		if len(function.Parameters) != len(test.expectedParams) {
			t.Errorf("parameter arity wrong. expect %d, got=%d\n",
				len(test.expectedParams), len(function.Parameters))
		}

		for i, ident := range test.expectedParams {
			testPrimaryExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"
	expectNum := int64(5)

	l := lexer.New("parser_test_int", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if !testIntLiteral(t, stmt.Expression, expectNum) {
		return
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "10.23;"
	expectNum := 10.23

	l := lexer.New("parser_test_int", input)
	p := New(l)

	program := p.Parse()
	checkErrors(t, p)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if !testFloatLiteral(t, stmt.Expression, expectNum) {

	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
	}

	for _, test := range tests {
		l := lexer.New("parser_test_operator_precedence", test.input)
		p := New(l)

		program := p.Parse()
		checkErrors(t, p)

		if found := program.String(); found != test.expected {
			t.Errorf("expected=%q, got=%q", test.expected, found)
		}
	}
}

func TestString(t *testing.T) {
	program := &ast.Program{
		Statements: []ast.Statement{
			// let myVar = anotherVar;
			// return myVar;
			&ast.LetStatement{
				Token: token.Token{
					Type: token.LET,
					Word: "let",
				},
				Ident: &ast.Identifier{
					Token: token.Token{
						Type: token.IDENT,
						Word: "myVar",
					},
					Value: "myVar",
				},
				InitValue: &ast.Identifier{
					Token: token.Token{
						Type: token.IDENT,
						Word: "anotherVar",
					},
					Value: "anotherVar",
				},
			},
			&ast.ReturnStatement{
				Token: token.Token{
					Type: token.RETURN,
					Word: "return",
				},
				ReturnValue: &ast.Identifier{
					Token: token.Token{
						Type: token.IDENT,
						Word: "myVar",
					},
					Value: "myVar",
				},
			},
		},
	}

	expectString := "let myVar = anotherVar;return myVar;"

	if str := program.String(); str != expectString {
		t.Errorf("program.String() wrong. got=%q, expect=%q",
			str, expectString)
	}
}

func testLetStatement(t *testing.T, s ast.Statement, Value string) bool {
	letStmt, ok := s.(*ast.LetStatement)

	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Ident.Value != Value {
		t.Errorf("letStmt.Ident.Value not %q. got=%q", Value, letStmt.Ident.Value)
		return false
	}

	if word := letStmt.Ident.TokenWord(); word != Value {
		t.Errorf("letStmt.Ident.TokenWord() not %q. got=%q", Value, word)
		return false
	}

	return true
}

func testInfixExpression(t *testing.T, expr ast.Expression,
	left interface{}, operator string, right interface{}) bool {
	infix, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Errorf("expr is not ast.InfixExpression. got=%T", expr)
		return false
	}

	if !testPrimaryExpression(t, infix.Left, left) {
		return false
	}

	if infix.Operator != operator {
		t.Errorf("infix.Operator is not %q. got=%q", operator, infix.Operator)
		return false
	}

	if !testPrimaryExpression(t, infix.Right, right) {
		return false
	}

	return true
}

func testPrefixExpression(t *testing.T, expr ast.Expression,
	operator string, right interface{}) bool {
	prefix, ok := expr.(*ast.PrefixExpression)
	if !ok {
		t.Errorf("expr is not ast.PrefixExpression. got=%T", expr)
		return false
	}

	if prefix.Operator != operator {
		t.Errorf("prefix.Operator is not %q. got=%q", operator, prefix.Operator)
		return false
	}

	if !testPrimaryExpression(t, prefix.Right, right) {
		return false
	}

	return false
}

func testPrimaryExpression(t *testing.T, expr ast.Expression, expect interface{}) bool {
	switch v := expect.(type) {
	case int64:
		return testIntLiteral(t, expr, v)
	case float64:
		return testFloatLiteral(t, expr, v)
	case string:
		return testIdentifier(t, expr, v)
	case bool:
		return testBooleanLiteral(t, expr, v)
	default:
		t.Errorf("type of v not handled. got=%T", v)
		return false
	}
}

func testIdentifier(t *testing.T, expr ast.Expression, value string) bool {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		t.Errorf("expr not *ast.Identifier. got=%T", expr)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %q. got=%q", value, ident.Value)
		return false
	}

	if word := ident.TokenWord(); word != value {
		t.Errorf("ident.TokenWord() not %q. got=%q", value, word)
		return false
	}

	return true
}

func testIntLiteral(t *testing.T, expr ast.Expression, value int64) bool {
	i, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("expr not *ast.IntegerLiteral. got=%T", expr)
		return false
	}

	if i.Value != value {
		t.Errorf("i.Value not %d. got=%d", value, i.Value)
		return false
	}

	if i.TokenWord() != fmt.Sprintf("%d", value) {
		t.Errorf("i.TokenWord() not '%d'. got=%q", value, i.TokenWord())
		return false
	}

	return true
}

func testFloatLiteral(t *testing.T, expr ast.Expression, value float64) bool {
	i, ok := expr.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("expr not *ast.IntegerLiteral. got=%T", expr)
		return false
	}

	if i.Value != value {
		t.Errorf("i.Value not %f. got=%f", value, i.Value)
		return false
	}

	if word := i.TokenWord(); word != strconv.FormatFloat(value, 'f', -1, 64) {
		t.Errorf("i.TokenWord() not '%f'. got=%q", value, word)
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, expr ast.Expression, value bool) bool {
	bl, ok := expr.(*ast.BoolLiteral)
	if !ok {
		t.Errorf("exp not *ast.BoolLiteral. got=%T", expr)
		return false
	}

	if bl.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bl.Value)
		return false
	}

	if bl.TokenWord() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not '%t'. got=%q",
			value, bl.TokenWord())
		return false
	}

	return true
}

func checkErrors(t *testing.T, p *Parser) {
	if len(p.Errors()) != 0 {
		logErrors(t, p)
		t.FailNow()
	}
}

func logErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	t.Errorf("parser has %d errors", len(errors))
	for _, message := range errors {
		t.Errorf("parser error: %s", message)
	}
}
