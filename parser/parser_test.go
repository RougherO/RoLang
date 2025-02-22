package parser

import (
	"RoLang/ast"
	"RoLang/lexer"
	"RoLang/token"

	"errors"
	"regexp"
	"testing"
)

func TestLetStatement(t *testing.T) {
	input := `
let x = 5;
let y = 10.23;
let foobar = x;
let neg = -1;
let add23 = 2 + 3;
let calladd = fn (x, y) { x + y; };
`
	l := lexer.New("parser_test_let", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 6 {
		t.Fatalf("program.Statements does not contain 6 statements. got=%d", n)
	}

	tests := []struct {
		expectIdent string
		expectInit  func(*testing.T, ast.Expression) bool
	}{
		{"x", func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, 5) }},
		{"y", func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, 10.23) }},
		{"foobar", func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, "x") }},
		{"neg", func(t *testing.T, expr ast.Expression) bool { return testPrefixExpression(t, expr, "-", 1) }},
		{"add23", func(t *testing.T, expr ast.Expression) bool { return testInfixExpression(t, expr, 2, "+", 3) }},
		{"calladd", func(t *testing.T, expr ast.Expression) bool {
			return testFunction(t, expr, "", []string{"x", "y"}, func(t *testing.T, body *ast.BlockStatement) bool {
				if n := len(body.Statements); n != 1 {
					t.Errorf("body.Statements contain incorrect number of statements. got=%d", n)
					return false
				}

				stmt, ok := body.Statements[0].(*ast.ExpressionStatement)
				if !ok {
					t.Errorf("body.Statements[0] not *ast.ExpressionStatement. got=%T", body.Statements[0])
					return false
				}

				if !testInfixExpression(t, stmt.Expression, "x", "+", "y") {
					return false
				}

				return true
			})
		}},
	}

	for i, test := range tests {
		stmt := program.Statements[i]

		ident, ok := stmt.(*ast.LetStatement)
		if !ok {
			t.Fatalf("stmt not *ast.LetStatement. got=%T", stmt)
		}

		if !testIdentifier(t, ident.Ident, test.expectIdent) {
			return
		}

		if !test.expectInit(t, ident.InitValue) {
			return
		}
	}
}

func TestLoopStatement(t *testing.T) {
	input1 := "loop x < 1 { true; }"
	input2 := "loop { true; }"

	l := lexer.New("parser_test_loop", input1)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.LoopStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not *ast.LoopStatement. got=%T",
			program.Statements)
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", 1) {
		return
	}

	body, ok := stmt.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt.Body.Statements[0] not *ast.ExpressionStatement. got=%T",
			stmt.Body.Statements[0])
	}

	if !testPrimaryExpression(t, body.Expression, true) {
		return
	}

	l1 := lexer.New("parser_test_loop", input2)
	p1 := New(l1)

	program1, errs := p1.Parse()
	checkErrors(t, errs)

	if n := len(program1.Statements); n != 1 {
		t.Fatalf("program1.Statements does not contain 1 statement. got=%d", n)
	}

	stmt1, ok := program1.Statements[0].(*ast.LoopStatement)
	if !ok {
		t.Fatalf("program1.Statements[0] not *ast.LoopStatement. got=%T",
			program1.Statements)
	}

	if stmt1.Condition != nil {
		t.Fatalf("stmt1.Condition is not nil")
	}

	body1, ok := stmt1.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt1.Body.Statements[0] not *ast.ExpressionStatement. got=%T",
			stmt1.Body.Statements[0])
	}

	if !testPrimaryExpression(t, body1.Expression, true) {
		return
	}
}

func TestJumpStatement(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"break;", "break"},
		{"continue;", "continue"},
	}

	for _, test := range tests {
		l := lexer.New("parser_test_jump", test.input)
		p := New(l)

		stmt := p.ParseStatement()

		jump, ok := stmt.(*ast.JumpStatement)
		if !ok {
			t.Fatalf("jump is not *ast.JumpStatement, got=%T", stmt)
		}

		if jump.Token.Word != test.expect {
			t.Fatalf("jump is not %q. got=%q", test.expect, jump.Token.Word)
		}
	}
}

func TestFunctionStatement(t *testing.T) {
	input := "fn add(x, y) { x + y; }"

	l := lexer.New("parser_test_func_stmt", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", n)
	}

	expectParams := []string{"x", "y"}

	if !testFunction(t, program.Statements[0], "add", expectParams, func(t *testing.T, body *ast.BlockStatement) bool {
		if n := len(body.Statements); n != 1 {
			t.Errorf("body.Statements contain incorrect number of statements. got=%d", n)
			return false
		}

		stmt, ok := body.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("body.Statements[0] not *ast.ExpressionStatement. got=%T", body.Statements[0])
			return false
		}

		if !testInfixExpression(t, stmt.Expression, "x", "+", "y") {
			return false
		}

		return true
	}) {
		return
	}
}

func TestReturnStatement(t *testing.T) {
	input := `
return 5;
return 10;
return 10.233;
return x;
return -2;
return 1 + 2;
return "hello";
`
	l := lexer.New("parser_test_return", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 7 {
		t.Fatalf("program.Statements does not contain 7 statements. got=%d", n)
	}

	tests := []struct {
		expectReturn func(*testing.T, ast.Expression) bool
	}{
		{func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, 5) }},
		{func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, 10) }},
		{func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, 10.233) }},
		{func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, "x") }},
		{func(t *testing.T, expr ast.Expression) bool { return testPrefixExpression(t, expr, "-", 2) }},
		{func(t *testing.T, expr ast.Expression) bool { return testInfixExpression(t, expr, 1, "+", 2) }},
		{func(t *testing.T, expr ast.Expression) bool { return testPrimaryExpression(t, expr, `str(hello)`) }},
	}

	for i, test := range tests {
		stmt, ok := program.Statements[i].(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("program.Statements[%d] not *ast.ReturnStatement. got=%T", i, program.Statements[i])
		}

		if !test.expectReturn(t, stmt.ReturnValue) {
			return
		}
	}
}

func TestIfStatement(t *testing.T) {
	input := `if x < y { x; }`

	l := lexer.New("parser_test_if", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

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
	input := `if x < y { x; } else { y; }`

	l := lexer.New("parser_test_if_else", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

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
	input := `if x < y { x; } else if x > y { y; }`

	l := lexer.New("parser_test_if_else_if", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

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
	input := `if x < y { x; } else if x > y { y; } else { x + y; }`

	l := lexer.New("parser_test_if_else_if", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

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
		{"!a;", "!", "a"},
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!5.223;", "!", 5.223},
		{"-10.23;", "-", 10.23},
	}

	for _, test := range prefixIntTests {
		l := lexer.New("parser_test_prefix", test.input)
		p := New(l)

		program, errs := p.Parse()
		checkErrors(t, errs)

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

func TestAssignExpression(t *testing.T) {
	input := `x = 1; a[0] = 2;`

	l := lexer.New("parser_test_assign", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 2 {
		t.Fatalf("program.Statements does not contain 2 statements. got=%d", n)
	}

	exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exprStmt is not *ast.ExpressionStatement. got=%T", exprStmt)
	}

	assignId, ok := exprStmt.Expression.(*ast.AssignExpression)
	if !ok {
		t.Fatalf("exprStmt.Expression is not *ast.AssignExpression. got=%T",
			exprStmt.Expression)
	}

	if !testPrimaryExpression(t, assignId.Left, "x") {
		return
	}

	if !testPrimaryExpression(t, assignId.Right, 1) {
		return
	}

	indexStmt, ok := program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("indexStmt is not *ast.ExpressionStatement. got=%T", indexStmt)
	}

	assign, ok := indexStmt.Expression.(*ast.AssignExpression)
	if !ok {
		t.Fatalf("indexStmt.Expression is not *ast.AssignExpression. got=%T",
			indexStmt.Expression)
	}

	lhs, ok := assign.Left.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("assign.Left is not *ast.IndexExpression. got=%T",
			assign.Left)
	}

	if !testPrimaryExpression(t, lhs.Left, "a") {
		return
	}

	if !testPrimaryExpression(t, lhs.Index, 0) {
		return
	}

	if !testPrimaryExpression(t, assign.Right, 2) {
		return
	}
}

func TestInfixExpression(t *testing.T) {
	infixTests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5.23 + 5.23;", 5.23, "+", 5.23},
		{"5.23 - 5.23;", 5.23, "-", 5.23},
		{"5.23 * 5.23;", 5.23, "*", 5.23},
		{"5.23 / 5.23;", 5.23, "/", 5.23},
		{"5.23 > 5.23;", 5.23, ">", 5.23},
		{"5.23 < 5.23;", 5.23, "<", 5.23},
		{"5.23 == 5.23;", 5.23, "==", 5.23},
		{"5.23 != 5.23;", 5.23, "!=", 5.23},
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

		program, errs := p.Parse()
		checkErrors(t, errs)

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

func TestCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4.53 + 5.22);"

	l := lexer.New("parser_test_call_expr", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if !testIdentifier(t, expr.Callee, "add") {
		return
	}

	if len(expr.Arguments) != 3 {
		t.Fatalf("expr.Arguments has wrong arity. got=%d", len(expr.Arguments))
	}

	if !testIntLiteral(t, expr.Arguments[0], 1) ||
		!testInfixExpression(t, expr.Arguments[1], 2, "*", 3) ||
		!testInfixExpression(t, expr.Arguments[2], 4.53, "+", 5.22) {
		return
	}
}

func TestIndexExpression(t *testing.T) {
	input := "arr[1 + 2];"

	l := lexer.New("parser_test_index", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	expr, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IndexExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, expr.Left, "arr") {
		return
	}

	if !testInfixExpression(t, expr.Index, 1, "+", 2) {
		return
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	expectStr := "foobar"

	l := lexer.New("parser_test_ident", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not *ast.ExpressionStatement. got=%T", stmt)
	}

	if !testIdentifier(t, stmt.Expression, expectStr) {
		return
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3];"

	l := lexer.New("parser_test_array", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] not *ast.ExpressionStatement. got=%T", stmt)
	}

	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("stmt.Expression not *ast.ArrayLiteral. got=%T", stmt)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("len(arr.Elements) not 3. got=%d", len(arr.Elements))
	}

	if !testIntLiteral(t, arr.Elements[0], 1) {
		return
	}
	if !testInfixExpression(t, arr.Elements[1], 2, "*", 2) {
		return
	}
	if !testInfixExpression(t, arr.Elements[2], 3, "+", 3) {
		return
	}
}

func TestMapLiteral(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	l := lexer.New("parser_test_map", input)
	p := New(l)

	expr := p.ParseExpression(NONE)
	checkErrors(t, p.errors)

	hash, ok := expr.(*ast.MapLiteral)
	if !ok {
		t.Fatalf("expr is not ast.HashLiteral. got=%T", expr)
	}

	if len(hash.Elements) != 3 {
		t.Errorf("hash.Elements has wrong length. got=%d", len(hash.Elements))
	}

	e0 := hash.Elements[0]
	if !testStringLiteral(t, e0.Key, "one") ||
		!testIntLiteral(t, e0.Value, 1) {
		return
	}

	e1 := hash.Elements[1]
	if !testStringLiteral(t, e1.Key, "two") ||
		!testIntLiteral(t, e1.Value, 2) {
		return
	}

	e2 := hash.Elements[2]
	if !testStringLiteral(t, e2.Key, "three") ||
		!testIntLiteral(t, e2.Value, 3) {
		return
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"
	expectNum := 5

	l := lexer.New("parser_test_int", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if !testPrimaryExpression(t, stmt.Expression, expectNum) {
		return
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := "10.23;"
	expectNum := 10.23

	l := lexer.New("parser_test_int", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	if !testFloatLiteral(t, stmt.Expression, expectNum) {
		return
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`
	expectStr := "hello world"

	l := lexer.New("parser_test_string", input)
	p := New(l)

	program, errs := p.Parse()
	checkErrors(t, errs)

	if n := len(program.Statements); n != 1 {
		t.Fatalf("program has not enough statements. got=%d", n)
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	if !testStringLiteral(t, stmt.Expression, expectStr) {
		return
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
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
		{
			"x = y = z",
			"(x = (y = z))",
		},
		{
			"x.y == a.b",
			"((x . y) == (a . b))",
		},
	}

	for _, test := range tests {
		l := lexer.New("parser_test_operator_precedence", test.input)
		p := New(l)

		expr := p.ParseExpression(NONE)
		checkErrors(t, p.errors)

		if found := expr.String(); found != test.expected {
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

func testFunction(t *testing.T, node ast.Node, expectedName string, expectedParams []string, testBody func(*testing.T, *ast.BlockStatement) bool) bool {
	switch v := node.(type) {
	case *ast.FunctionStatement:
		if !testIdentifier(t, v.Ident, expectedName) {
			return false
		}
		if !testFunctionParameterParsing(t, v.Value.Parameters, expectedParams) {
			return false
		}
		if !testBody(t, v.Value.Body) {
			return false
		}
		return true
	case *ast.FunctionLiteral:
		if !testFunctionParameterParsing(t, v.Parameters, expectedParams) {
			return false
		}
		if !testBody(t, v.Body) {
			return false
		}
		return true
	default:
		t.Errorf("type of v not handled. got=%T", v)
		return false
	}
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

	return true
}

var re, _ = regexp.Compile(`str\((\w*)\)`)

func testPrimaryExpression(t *testing.T, expr ast.Expression, expect interface{}) bool {
	switch v := expect.(type) {
	case int64:
		return testIntLiteral(t, expr, v)
	case int:
		return testIntLiteral(t, expr, int64(v))
	case float64:
		return testFloatLiteral(t, expr, v)
	case string:
		// to distinguish between string literals and identifers
		// we surround string literals with str(...)
		if re.MatchString(v) {
			str := re.FindStringSubmatch(v)[1]
			return testStringLiteral(t, expr, str)
		}
		return testIdentifier(t, expr, v)
	case bool:
		return testBooleanLiteral(t, expr, v)
	default:
		t.Errorf("type of v not handled. got=%T", v)
		return false
	}
}

func testFunctionParameterParsing(t *testing.T, parameters []*ast.Identifier, expectedParams []string) bool {
	if len(parameters) != len(expectedParams) {
		t.Errorf("parameter arity wrong. expect %d, got=%d\n",
			len(expectedParams), len(parameters))
		return false
	}

	for i, ident := range expectedParams {
		if !testIdentifier(t, parameters[i], ident) {
			return false
		}
	}

	return true
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

	return true
}

func testStringLiteral(t *testing.T, expr ast.Expression, value string) bool {
	l, ok := expr.(*ast.StringLiteral)
	if !ok {
		t.Errorf("expr not *ast.StringLiteral. got=%T", expr)
		return false
	}

	if l.Value != value {
		t.Errorf("l.Value not %q. got=%q", value, l.Value)
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

	return true
}

func checkErrors(t *testing.T, errs []error) {
	if len(errs) != 0 {
		t.Fatal(errors.Join(errs...))
	}
}
