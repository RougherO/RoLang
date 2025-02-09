package evaluator

import (
	"RoLang/lexer"
	"RoLang/parser"

	"math"
	"testing"
)

func TestInfixOperator(t *testing.T) {
	tests := []struct {
		input  string
		expect any
	}{
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"5.5 + 4.5 + 5.0 - 10.0", 5.0},
		{"2.5 * 2.0 * 2.0", 10.0},
		{"-50.0 + 100.5 + -50.5", 0.0},
		{"5.0 * 2.5 + 10.0", 22.5},
		{"5.5 + 2.0 * 10.0", 25.5},
		{"20.0 + 2.0 * -10.5", -1.0},
		{"50.0 / 2.5 * 2.0 + 10.0", 50.0},
		{"2.5 * (5.0 + 10.0)", 37.5},
		{"3.1 * 3.1 * 3.1 + 10.0", 39.791},
		{"3.5 * (3.0 * 3.0) + 10.0", 41.5},
		{"(5.5 + 10.5 * 2.0 + 15.0 / 3.0) * 2.0 + -10.5", 52.5},
		{"5 > 3", true},
		{"10 < 5", false},
		{"2.5 * 2 == 5.0", true},
		{"3.5 * 3 > 10.5", false},
		{"(5 + 5) == (2 * 5)", true},
		{"10.0 / 2.0 != 5.0", false},
		{"(3.0 + 2.0) * 2.0 == 10.0", true},
		{"5.0 >= 5.0", true},
		{"7.5 <= 7.4", false},
		{"!(3.5 == 3.5)", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for i, test := range tests {
		eval := testEvalExpression(test.input)
		if !testPrimaryObject(t, eval, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestPrefixOperator(t *testing.T) {
	tests := []struct {
		input  string
		expect any
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!5.5", false},
		{"!!5.5", true},
		{"-10", -10},
		{"-5.5", -5.5},
	}

	for i, test := range tests {
		eval := testEvalExpression(test.input)
		if !testPrimaryObject(t, eval, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestIntegerExpression(t *testing.T) {
	tests := []struct {
		input  string
		expect int64
	}{
		{"5", 5},
		{"10", 10},
	}

	for i, test := range tests {
		eval := testEvalExpression(test.input)
		if !testIntegerObject(t, eval, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestFloatExpression(t *testing.T) {
	tests := []struct {
		input  string
		expect float64
	}{
		{"5.5", 5.5},
		{"10.23", 10.23},
	}

	for i, test := range tests {
		eval := testEvalExpression(test.input)
		if !testFloatObject(t, eval, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input  string
		expect bool
	}{
		{"true", true},
		{"false", false},
	}

	for i, test := range tests {
		eval := testEvalExpression(test.input)
		if !testBooleanObject(t, eval, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func testEvalStatement(input string) any {
	l := lexer.New("evaluator_test_statement", input)
	p := parser.New(l)

	stmt := p.ParseStatement()
	return Eval(stmt)
}

func testEvalExpression(input string) any {
	l := lexer.New("evaluator_test_expression", input)
	p := parser.New(l)

	expr := p.ParseExpression(parser.NONE)
	return Eval(expr)
}

func testPrimaryObject(t *testing.T, obj any, expect any) bool {
	switch e := expect.(type) {
	case int64:
		return testIntegerObject(t, obj, e)
	case int:
		return testIntegerObject(t, obj, int64(e))
	case float64:
		return testFloatObject(t, obj, e)
	case bool:
		return testBooleanObject(t, obj, e)
	default:
		t.Errorf("type of e not handled. got=%T", e)
		return false
	}
}

func testBooleanObject(t *testing.T, obj any, expect bool) bool {
	result, ok := obj.(bool)
	if !ok {
		t.Errorf("object is not Boolean. got=%T", obj)
		return false
	}

	if result != expect {
		t.Errorf("object has wrong value. got=%t, want=%t", result, expect)
		return false
	}

	return true
}

func testIntegerObject(t *testing.T, obj any, expect int64) bool {
	result, ok := obj.(int64)
	if !ok {
		t.Errorf("object is not Integer. got=%T", obj)
		return false
	}

	if result != expect {
		t.Errorf("object has wrong value. got=%d, want=%d", result, expect)
		return false
	}

	return true
}

func testFloatObject(t *testing.T, obj any, expect float64) bool {
	result, ok := obj.(float64)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", obj, obj)
		return false
	}

	const tol = 1e-9

	if math.Abs(result-expect) > tol {
		t.Errorf("object has wrong value. got=%f, want=%f",
			result, expect)
		return false
	}

	return true
}
