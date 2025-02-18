package evaluator

import (
	"RoLang/evaluator/env"
	"RoLang/evaluator/objects"
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

func TestLoopStatement(t *testing.T) {
	// TODO need tests
}

func TestBlockStatement(t *testing.T) {
	// TODO need tests
}

func TestFunctionStatement(t *testing.T) {
	// TODO need tests
}

func TestBuiltInFunctions(t *testing.T) {
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
		{
			`let s = "hello world";`,
			[]expectType{
				{"s", "hello world"},
			},
		},
	}

	for i, test := range tests {
		if !testLetStatements(t, test.input, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestAssignExpression(t *testing.T) {
	input := `let x = 1; x = 2;`

	l := lexer.New("evaluator_test_assign", input)
	p := parser.New(l)

	program := p.Parse()

	Init(nil, nil, nil)
	Evaluate(program)

	if !testIdentifier(t, "x", int64(2)) {
		return
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
		// {"(10.0 > 5.0) && (2.0 < 4.0)", true},
		// {"(10.0 < 5.0) || (2.0 > 1.0)", true},
		{"!(3.5 == 3.5)", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{`"hello" + 1`, "hello1"},
		{`1 + "hello" + 2.23`, "1hello2.23"},
	}

	for i, test := range tests {
		eval := testEvalExpression(test.input)
		if !testPrimaryObject(t, eval, test.expect) {
			t.Logf("test[%d]\n", i)
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	expr := testEvalExpression(input)
	obj, ok := expr.(*objects.ArrayObject)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", expr, expr)
	}
	result := obj.List

	if len(result) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result))
	}

	if !testIntegerObject(t, result[0], 1) {
		return
	}
	if !testIntegerObject(t, result[1], 4) {
		return
	}
	if !testIntegerObject(t, result[2], 6) {
		return
	}
}

func TestMapLiteral(t *testing.T) {
	input := `{"hello": 1, "world": 2}`

	expr := testEvalExpression(input)
	obj, ok := expr.(*objects.MapObject)
	if !ok {
		t.Fatalf("object is not Map. got=%T (%+v)", expr, expr)
	}
	result := obj.Map

	if len(result) != 2 {
		t.Fatalf("map has wrong num of elements. got=%d", len(result))
	}

	val1, ok := result["hello"]
	if !ok {
		t.Fatal(`map has no key "hello"`)
	}

	if !testPrimaryObject(t, val1, 1) {
		return
	}

	val2, ok := result["world"]
	if !ok {
		t.Fatal(`map has no key "world"`)
	}

	if !testPrimaryObject(t, val2, 2) {
		return
	}
}

func TestIndexExpression(t *testing.T) {
	tests := []struct {
		input  string
		expect any
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"[1, 2, 3][1 - 1]",
			1,
		},
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
	case string:
		return testStringObject(t, obj, e)
	case bool:
		return testBooleanObject(t, obj, e)
	default:
		t.Errorf("type of e not handled. got=%T", e)
		return false
	}
}

func testBooleanObject(t *testing.T, obj any, expect bool) bool {
	result, _ := obj.(bool)

	if result != expect {
		t.Errorf("object has wrong value. got=%t, want=%t", result, expect)
		return false
	}

	return true
}

func testStringObject(t *testing.T, obj any, expect string) bool {
	result, _ := obj.(string)

	if result != expect {
		t.Errorf("object has wrong value. got=%q, want=%q", result, expect)
		return false
	}

	return true
}

func testIntegerObject(t *testing.T, obj any, expect int64) bool {
	result, _ := obj.(int64)

	if result != expect {
		t.Errorf("object has wrong value. got=%d, want=%d", result, expect)
		return false
	}

	return true
}

func testFloatObject(t *testing.T, obj any, expect float64) bool {
	result, _ := obj.(float64)

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
