package evaluator

import (
	"RoLang/evaluator/env"
	"RoLang/lexer"
	"RoLang/parser"
	"regexp"

	"bytes"
	"math"
	"strings"
	"testing"
)

type expectType struct {
	name  string
	value any
}

func TestBlockStatement(t *testing.T) {
	// TODO need tests
}

func TestFunctionStatement(t *testing.T) {
	// TODO need tests
}

func TestIfStatement(t *testing.T) {
	// TODO need tests
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input  string
		expect []expectType
	}{
		{
			"let a = 5;",
			[]expectType{
				{"a", int64(5)},
			}},
		{
			"let a = 5.2 * 2;",
			[]expectType{
				{"a", 10.4},
			},
		},
		{
			"let a = 5; let b = a;",
			[]expectType{
				{"a", int64(5)},
				{"b", int64(5)},
			},
		},
		{
			"let a = 5.5; let b = a; let c = a + b + 5;",
			[]expectType{
				{"a", 5.5},
				{"b", 5.5},
				{"c", 16.0},
			},
		},
	}

	for i, test := range tests {
		if !testLetStatements(t, test.input, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestCallExpressions(t *testing.T) {
	// TODO needs test
}

func TestClosureExpression(t *testing.T) {
	// TODO needs test
}

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

func TestErrorStatements(t *testing.T) {
	err := new(bytes.Buffer)
	Init(nil, nil, err)

	tests := []struct {
		input  string
		expect string
	}{
		// {"let x = 1", "expected next token to be \";\", got \"eof\" instead"},
		{"let x = y;", "variable not found: y"},
		{"let x = 1; let y = x();", "not a callable int"},
		{"fn f(){} let x = f(); x();", "cannot function call on null objects"},
	}

	for i, test := range tests {
		ctxt.Env = env.New(nil) // reset environement for each test to prevent name clash
		testEvalStatements(test.input)
		if !testErrors(t, err, test.expect) {
			t.Logf("test[%d]\n", i)
		}
		err.Reset()
	}
}

func TestOutStatements(t *testing.T) {
	out := new(bytes.Buffer)

	Init(nil, out, nil)

}

func testLetStatements(t *testing.T, input string, expects []expectType) bool {
	out := new(bytes.Buffer)
	err := new(bytes.Buffer)

	Init(nil, out, err)

	l := lexer.New("evaluator_test", input)
	p := parser.New(l)

	program := p.Parse()
	Evaluate(program)

	isValid := true
	for _, item := range expects {
		name := item.name
		expect := item.value

		isValid = isValid && testIdentifier(t, name, expect)
	}

	return isValid
}

func testIdentifier(t *testing.T, name string, expect any) bool {
	value, ok := ctxt.Env.Get(name)
	if !ok {
		t.Errorf("no identifier found %s", name)
		return false
	}

	if value != expect {
		t.Errorf("values are not equal %v(%T) %v(%T)", value, value, expect, expect)
		return false
	}

	return true
}

func testEvalStatements(input string) {
	l := lexer.New("evaluator_test", input)
	p := parser.New(l)

	program := p.Parse()
	Evaluate(program)
}

func testEvalExpression(input string) any {
	l := lexer.New("evaluator_test", input)
	p := parser.New(l)

	expr := p.ParseExpression(parser.NONE)
	return evalExpression(expr)
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
		t.Errorf("object is not Float. got=%T", obj)
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

var re, _ = regexp.Compile(`runtime error:(\nevaluator_test:\d+:\d+: )+`)

func testErrors(t *testing.T, err *bytes.Buffer, expect string) bool {
	errStr := err.String()
	errStr = strings.TrimPrefix(errStr, re.FindString(errStr)) // trim the location info
	errStr = strings.TrimSpace(errStr)                         // trime surrounding spaces
	if expect != errStr {
		t.Errorf("different error statement. got=%q expect=%q", errStr, expect)
		return false
	}
	return true
}
