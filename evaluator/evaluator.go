package evaluator

import (
	"RoLang/ast"
	"RoLang/evaluator/env"
	"RoLang/evaluator/objects"
	"RoLang/stdlib"
	"RoLang/stdlib/builtin"
	"RoLang/stdlib/common"
	"RoLang/stdlib/strings"

	"fmt"
	"maps"
	"os"
	"slices"
)

type Evaluator struct {
	errors   []error
	env      *env.Environment
	envStack []*env.Environment
	stdlib   *stdlib.StdLib
}

func New() *Evaluator {
	return &Evaluator{
		env:    env.New(nil),
		stdlib: stdlib.New(),
	}
}

func (e *Evaluator) Evaluate(program *ast.Program) []error {
	defer e.recoveryHandler()
	e.errors = nil

	err := e.evalStatements(program.Statements)
	if err != nil {
		e.addError(err)
	}

	return e.errors
}

// for block scopes it should just enclose the current environment
func (e *Evaluator) createEnv() {
	e.env = env.New(e.env)
}

// restores the environment set with `CreateEnv` function
// by just 'moving up' the environment chain
func (e *Evaluator) restoreEnv() {
	e.env = e.env.Outer()
}

// for function calls it should set up a new environment
// with by enclosing the one provided in parameter
func (e *Evaluator) setEnv(environment *env.Environment) {
	e.envStack = append(e.envStack, e.env)
	e.env = env.New(environment)
}

// resets the environment set with `SetEnv“ function
func (e *Evaluator) resetEnv() {
	e.env = e.envStack[len(e.envStack)-1]
	e.envStack = e.envStack[:len(e.envStack)-1]
}

// recovers from any kind of error that may have
// trickled down to the top most execution stack
// this handler is a best effort handler for handling
// any possible error that couldn't be handled in
// intermediate steps
func (e *Evaluator) recoveryHandler() {
	err := recover()

	if err != nil {
		switch v := err.(type) {
		case objects.ReturnObject:
			switch v := v.Value.(type) {
			case int64:
				os.Exit(int(v))
			case nil:
				os.Exit(0)
			default:
				e.addError(fmt.Errorf("can only return integer exit codes at top level"))
			}
		case error:
			e.addError(fmt.Errorf("runtime error:%v", e))
		}
	}
}

func (e *Evaluator) errorDecorator(node ast.Node, err error) error {
	if err != nil {
		return fmt.Errorf("\n%s %s", node.Location(), err)
	}

	return nil
}

func (e *Evaluator) evalStatements(stmts []ast.Statement) error {
	for _, stmt := range stmts {
		err := e.evalStatement(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Evaluator) evalStatement(statement ast.Statement) error {
	// error handler for statement panics
	// used for adding source location to the error
	var err error

	switch stmt := statement.(type) {
	case *ast.LetStatement:
		err = e.evalLetStatement(stmt)
	case *ast.FunctionStatement:
		err = e.evalFunctionStatement(stmt)
	case *ast.ReturnStatement:
		err = e.evalReturnStatement(stmt)
	case *ast.IfStatement:
		err = e.evalIfStatement(stmt)
	case *ast.BlockStatement:
		e.createEnv()
		defer e.restoreEnv()
		err = e.evalStatements(stmt.Statements)
		// should pop out the current environment no matter what
	case *ast.ExpressionStatement:
		_, err = e.evalExpression(stmt.Expression)
	case *ast.LoopStatement:
		err = e.evalLoopStatement(stmt)
	}

	return e.errorDecorator(statement, err)
}

func (e *Evaluator) evalFunctionStatement(function *ast.FunctionStatement) error {
	name := function.Ident.Value
	init, err := e.evalExpression(function.Value)
	if err != nil {
		return err
	}
	if !e.env.Set(name, init) {
		return fmt.Errorf("variable %s already exists in current scope", name)
	}

	return nil
}

func (e *Evaluator) evalLoopStatement(loop *ast.LoopStatement) error {
	for {
		// should continue looping if there is no condition
		// or check the condition repeatedly
		expr, err := e.evalExpression(loop.Condition)
		if err != nil {
			return err
		}

		cond := loop.Condition == nil || isTruthy(expr)
		if !cond {
			break
		}

		err = e.evalStatement(loop.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Evaluator) evalLetStatement(let *ast.LetStatement) error {
	name := let.Ident.Value
	init, err := e.evalExpression(let.InitValue)
	if err != nil {
		return err
	}
	if !e.env.Set(name, init) {
		return fmt.Errorf("variable %s already exists in current scope", name)
	}

	return nil
}

func (e *Evaluator) evalReturnStatement(ret *ast.ReturnStatement) error {
	var retValue any
	if ret.ReturnValue != nil {
		expr, err := e.evalExpression(ret.ReturnValue)
		if err != nil {
			return err
		}
		retValue = expr
	}

	panic(objects.ReturnObject{Value: retValue})
}

func (e *Evaluator) evalIfStatement(ifStmt *ast.IfStatement) error {
	condition, err := e.evalExpression(ifStmt.Condition)
	if err != nil {
		return err
	}

	if isTruthy(condition) {
		err := e.evalStatement(ifStmt.Then)
		if err != nil {
			return err
		}
	} else if ifStmt.Else != nil {
		err := e.evalStatement(ifStmt.Else)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Evaluator) evalExpression(expression ast.Expression) (value any, err error) {
	switch expr := expression.(type) {
	case *ast.InfixExpression:
		value, err = e.evalInfixExpression(expr)
	case *ast.PrefixExpression:
		value, err = e.evalPrefixExpression(expr)
	case *ast.Identifier:
		value, err = e.evalIdentifier(expr)
	case *ast.AssignExpression:
		value, err = e.evalAssignExpression(expr)
	case *ast.ArrayLiteral:
		value, err = e.evalArrayLiteral(expr)
	case *ast.MapLiteral:
		value, err = e.evalMapLiteral(expr)
	case *ast.StringLiteral:
		value, err = expr.Value, nil
	case *ast.BoolLiteral:
		value, err = expr.Value, nil
	case *ast.IntegerLiteral:
		value, err = expr.Value, nil
	case *ast.FloatLiteral:
		value, err = expr.Value, nil
	case *ast.FunctionLiteral:
		value, err = e.evalFunctionLiteral(expr)
	case *ast.CallExpression:
		value, err = e.evalCallExpression(expr)
	case *ast.IndexExpression:
		value, err = e.evalIndexExpression(expr)
	default:
		panic(fmt.Errorf("unknown expression type %T", expression))
	}

	err = e.errorDecorator(expression, err)
	return
}

func (e *Evaluator) evalArrayLiteral(expr *ast.ArrayLiteral) (*objects.ArrayObject, error) {
	arr := &objects.ArrayObject{}

	for _, elem := range expr.Elements {
		expr, err := e.evalExpression(elem)
		if err != nil {
			return nil, err
		}
		arr.List = append(arr.List, expr)
	}

	return arr, nil
}

func (e *Evaluator) evalMapLiteral(expr *ast.MapLiteral) (*objects.MapObject, error) {
	mp := &objects.MapObject{
		Map: make(map[any]any),
	}

	for _, elem := range expr.Elements {
		key, err := e.evalExpression(elem.Key)
		if err != nil {
			return nil, err
		}
		val, err := e.evalExpression(elem.Value)
		if err != nil {
			return nil, err
		}
		mp.Map[key] = val
	}

	return mp, nil
}

func (e *Evaluator) evalIndexExpression(expr *ast.IndexExpression) (any, error) {
	left, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	switch v := left.(type) {
	case *objects.ArrayObject:
		right, err := e.evalExpression(expr.Index)
		if err != nil {
			return nil, err
		}
		index, ok := right.(int64)
		if !ok {
			return nil, fmt.Errorf("expect integer index, got=%s", e.typeStr(index))
		}

		if index >= int64(len(v.List)) || index < 0 {
			return nil, fmt.Errorf("index out of range [%d]", index)
		}

		return v.List[index], nil
	case *objects.MapObject:
		right, err := e.evalExpression(expr.Index)
		if err != nil {
			return nil, err
		}
		switch right.(type) {
		case int64:
			return v.Map[right], nil
		case float64:
			return v.Map[right], nil
		case string:
			return v.Map[right], nil
		case bool:
			return v.Map[right], nil
		default:
			return nil, fmt.Errorf("only int, float, string and bool is allowed as key. got=%s",
				e.typeStr(v))
		}
	}

	return nil, fmt.Errorf("cannot index on type %s", e.typeStr(left))
}

func (e *Evaluator) evalCallExpression(expr *ast.CallExpression) (any, error) {
	value, err := e.evalExpression(expr.Callee)
	if err != nil {
		return nil, err
	}
	// if value == nil {
	// 	// trying to call on a null object
	// 	return fmt.Errorf("cannot function call on null objects")
	// }

	args, err := e.evalCallArgs(expr.Arguments)
	if err != nil {
		return nil, err
	}

	return e.callFunction(value, args)
}

func (e *Evaluator) evalCallArgs(args []ast.Expression) ([]any, error) {
	result := []any{}
	for _, expr := range args {
		arg, err := e.evalExpression(expr)
		if err != nil {
			return nil, err
		}
		result = append(result, arg)
	}

	return result, nil
}

func (e *Evaluator) callFunction(function any, args []any) (retValue any, errValue error) {
	switch obj := function.(type) {
	case *objects.FuncObject:
		returnRetriever := func() {
			e.resetEnv()

			err := recover()
			switch val := err.(type) {
			case objects.ReturnObject:
				retValue = val.Value // is a return value
			case error:
				errValue = val
			}
		}
		defer returnRetriever() // set return value or propagate error

		// create new scope with the function's
		e.setEnv(obj.Env)

		function := obj.Function
		if len(args) != len(function.Parameters) {
			return nil, fmt.Errorf("incorrect no of arguments. got=%d, expect=%d",
				len(args), len(function.Parameters))
		}

		for i, param := range function.Parameters {
			if !e.env.Set(param.Value, args[i]) {
				return nil, fmt.Errorf("redeclaration of variable %s", param.Value)
			}
		}

		err := e.evalStatements(function.Body.Statements)
		if err != nil {
			return nil, err
		}
		// reaching here means function does not return any value
		// in one of the control flow paths
		return nil, nil
	case common.Sanitizer:
		return obj(args...)
	default:
		return nil, fmt.Errorf("not a callable %s", e.typeStr(function))
	}
}

func (e *Evaluator) evalFunctionLiteral(expr *ast.FunctionLiteral) (*objects.FuncObject, error) {
	return &objects.FuncObject{
		Env:      e.env,
		Function: expr,
	}, nil
}

func (e *Evaluator) evalIdentifier(expr *ast.Identifier) (any, error) {
	if value, ok := e.env.Get(expr.Value); ok {
		return value, nil
	}
	if value, err := e.stdlib.GetModuleDispatcher("builtin", expr.Value); err == nil {
		return value, nil
	}

	return nil, fmt.Errorf("variable not found: %s", expr.Value)
}

func (e *Evaluator) evalInfixExpression(expr *ast.InfixExpression) (any, error) {
	if expr.Operator == "." {
		return e.evalModuleOperator(expr.Left, expr.Right)
	}

	left, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "+":
		return e.evalAddOperator(left, right)
	case "-":
		return e.evalSubOperator(left, right)
	case "*":
		return e.evalMulOperator(left, right)
	case "/":
		return e.evalDivOperator(left, right)
	case "<":
		return e.evalLtOperator(left, right)
	case ">":
		return e.evalGtOperator(left, right)
	case "<=":
		expr, err := e.evalGtOperator(left, right)
		if err != nil {
			return nil, err
		}
		return !expr.(bool), nil
	case ">=":
		expr, err := e.evalLtOperator(left, right)
		if err != nil {
			return nil, err
		}
		return !expr.(bool), nil
	case "==":
		return e.evalEqOperator(left, right)
	case "!=":
		expr, err := e.evalEqOperator(left, right)
		if err != nil {
			return nil, err
		}
		return !expr.(bool), nil
	default:
		return nil, fmt.Errorf("unknown operator %s", expr.Operator)
	}
}

func (e *Evaluator) evalAssignExpression(expr *ast.AssignExpression) (any, error) {
	right, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch left := expr.Left.(type) {
	case *ast.Identifier:
		if !e.env.Assign(left.Value, right) {
			return nil, fmt.Errorf("variable %q does not exist in current scope", left.Value)
		}
	case *ast.IndexExpression:
		l, err := e.evalExpression(left.Left)
		if err != nil {
			return nil, err
		}
		switch v := l.(type) {
		case *objects.ArrayObject:
			indexExpr, err := e.evalExpression(left.Index)
			if err != nil {
				return nil, err
			}

			index, ok := indexExpr.(int64)
			if !ok {
				return nil, fmt.Errorf("expect integer index, got=%s", e.typeStr(index))
			}

			if index >= int64(len(v.List)) || index < 0 {
				return nil, fmt.Errorf("index out of range [%d]", index)
			}

			v.List[index] = right
		case *objects.MapObject:
			index, err := e.evalExpression(left.Index)
			if err != nil {
				return nil, err
			}

			switch index.(type) {
			case int64:
				v.Map[index] = right
			case float64:
				v.Map[index] = right
			case string:
				v.Map[index] = right
			case bool:
				v.Map[index] = right
			default:
				return nil, fmt.Errorf("only int, float, string and bool is allowed as key. got=%s",
					e.typeStr(v))
			}
		}
	}

	return right, nil
}

func (e *Evaluator) evalModuleOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case *ast.Identifier:
		switch r := right.(type) {
		case *ast.Identifier:
			return e.stdlib.GetModuleDispatcher(l.Value, r.Value)
		default:
			return nil, fmt.Errorf("expect identifier after dot operator found %s", e.valueStr(r))
		}
	default:
		return nil, fmt.Errorf("cannot use dot operator with %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalAddOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l + r, nil
		case float64:
			return float64(l) + r, nil
		case string:
			return e.valueStr(l) + e.valueStr(r), nil
		default:
			return nil, fmt.Errorf("addition not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l + float64(r), nil
		case float64:
			return l + r, nil
		case string:
			return e.valueStr(l) + e.valueStr(r), nil
		default:
			return nil, fmt.Errorf("addition not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case string:
		switch r := right.(type) {
		case string:
			return e.valueStr(l) + e.valueStr(r), nil
		case int64:
			return e.valueStr(l) + e.valueStr(r), nil
		case float64:
			return e.valueStr(l) + e.valueStr(r), nil
		default:
			return nil, fmt.Errorf("addition not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case *objects.ArrayObject:
		switch r := right.(type) {
		case *objects.ArrayObject:
			return &objects.ArrayObject{
				List: slices.Concat(l.List, r.List),
			}, nil
		default:
			return nil, fmt.Errorf("addition not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case *objects.MapObject:
		switch r := right.(type) {
		case *objects.MapObject:
			mapObj := &objects.MapObject{
				Map: make(map[any]any),
			}
			maps.Copy(mapObj.Map, l.Map)
			maps.Copy(mapObj.Map, r.Map)
			return mapObj, nil
		default:
			return nil, fmt.Errorf("addition not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	default:
		return nil, fmt.Errorf("addition not supported for %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalSubOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l - r, nil
		case float64:
			return float64(l) - r, nil
		default:
			return nil, fmt.Errorf("subtraction not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l - float64(r), nil
		case float64:
			return l - r, nil
		default:
			return nil, fmt.Errorf("subtraction not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	default:
		return nil, fmt.Errorf("subtraction not supported for %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalMulOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l * r, nil
		case float64:
			return float64(l) * r, nil
		default:
			return nil, fmt.Errorf("multiplication not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l * float64(r), nil
		case float64:
			return l * r, nil
		default:
			return nil, fmt.Errorf("multiplication not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	default:
		return nil, fmt.Errorf("multiplication not supported for %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalDivOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l / r, nil
		case float64:
			return float64(l) / r, nil
		default:
			return nil, fmt.Errorf("division not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l / float64(r), nil
		case float64:
			return l / r, nil
		default:
			return nil, fmt.Errorf("division not supported for %s and %s", e.typeStr(l), e.typeStr(r))
		}
	default:
		return nil, fmt.Errorf("division not supported for %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalLtOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l < r, nil
		case float64:
			return float64(l) < r, nil
		default:
			return nil, fmt.Errorf("cannot compare types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l < float64(r), nil
		case float64:
			return l < r, nil
		default:
			return nil, fmt.Errorf("cannot compare types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case string:
		switch r := right.(type) {
		case string:
			return l < r, nil
		default:
			return nil, fmt.Errorf("cannot compare types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	default:
		return nil, fmt.Errorf("comparison not supported for %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalGtOperator(left, right any) (any, error) {
	e1, err := e.evalLtOperator(left, right)
	if err != nil {
		return nil, err
	}
	e2, err := e.evalEqOperator(left, right)
	if err != nil {
		return nil, err
	}
	return !e1.(bool) && !e2.(bool), nil
}

func (e *Evaluator) evalEqOperator(left, right any) (any, error) {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l == r, nil
		case float64:
			return float64(l) == r, nil
		default:
			return nil, fmt.Errorf("equality not supported for types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l == float64(r), nil
		case float64:
			return l == r, nil
		default:
			return nil, fmt.Errorf("equality not supported for types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case bool:
		switch r := right.(type) {
		case bool:
			return l == r, nil
		default:
			return nil, fmt.Errorf("equality not supported for types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case string:
		switch r := right.(type) {
		case string:
			return l == r, nil
		default:
			return nil, fmt.Errorf("equality not supported for types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case *objects.ArrayObject:
		switch r := right.(type) {
		case *objects.ArrayObject:
			return slices.Equal(l.List, r.List), nil
		default:
			return nil, fmt.Errorf("equality not supported for types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	case *objects.MapObject:
		switch r := right.(type) {
		case *objects.MapObject:
			return maps.Equal(l.Map, r.Map), nil
		default:
			return nil, fmt.Errorf("equality not supported for types %s and %s", e.typeStr(l), e.typeStr(r))
		}
	default:
		return nil, fmt.Errorf("equality not supported for %s", e.typeStr(l))
	}
}

func (e *Evaluator) evalPrefixExpression(expr *ast.PrefixExpression) (any, error) {
	right, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "!":
		return e.evalBangOperator(right)
	case "-":
		return e.evalNegateOperator(right)
	default:
		return nil, fmt.Errorf("unknown operator %s", expr.Operator)
	}
}

func (e *Evaluator) evalBangOperator(expr any) (any, error) {
	return !isTruthy(expr), nil
}

func (e *Evaluator) evalNegateOperator(expr any) (any, error) {
	switch v := expr.(type) {
	case int64:
		return -v, nil
	case float64:
		return -v, nil
	default:
		return nil, fmt.Errorf("cannot negate value of type %s", e.typeStr(e))
	}
}

func (e *Evaluator) valueStr(val any) string {
	return strings.From(val)
}

func (e *Evaluator) typeStr(val any) string {
	return builtin.TypeStr(val)
}

func isTruthy(value any) bool {
	switch value {
	case false:
		fallthrough
	case nil:
		fallthrough
	case int64(0):
		fallthrough
	case 0.0:
		fallthrough
	case "":
		return false
	default:
		return true
	}
}

func (e *Evaluator) addError(err error) {
	e.errors = append(e.errors, err)
}
