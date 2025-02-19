package evaluator

import (
	"RoLang/ast"
	"RoLang/evaluator/context"
	"RoLang/evaluator/objects"
	"maps"
	"slices"

	"fmt"
	"io"
	"os"
)

// global context variable to maintain
// evaluator states
var ctxt *context.Context

func Init(in io.Reader, out, err io.Writer) {
	ctxt = context.New(in, out, err)
}

func recoveryHandler() {
	err := recover()

	if err != nil {
		switch e := err.(type) {
		case objects.ReturnObject:
			switch e := e.Value.(type) {
			case int64:
				os.Exit(int(e))
			case nil:
				os.Exit(0)
			default:
				io.WriteString(ctxt.Err, "can only return integer exit codes at top level\n")
			}
		case error:
			io.WriteString(ctxt.Err, fmt.Sprintf("runtime error:%v", e)+"\n")
		}
	}
}

func Evaluate(program *ast.Program) {
	defer recoveryHandler()
	evalStatements(program.Statements)
}

func exprErrorHandler(expr ast.Expression) {
	err := recover()

	if err != nil {
		switch err.(type) {
		case objects.ReturnObject:
			panic(err)
		case error:
			panic(fmt.Errorf("\n%s %s", expr.Location(), err))
		}
	}
}

func stmtErrorHandler(stmt ast.Statement) {
	err := recover()

	if err != nil {
		switch err.(type) {
		case objects.ReturnObject:
			panic(err)
		case error:
			panic(fmt.Errorf("\n%s %s", stmt.Location(), err))
		}
	}
}

func evalStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		evalStatement(stmt)
	}
}

func evalStatement(stmt ast.Statement) {
	// error handler for statement panics
	// used for adding source location to the error
	defer stmtErrorHandler(stmt)

	switch s := stmt.(type) {
	case *ast.LetStatement:
		evalLetStatement(s)
	case *ast.FunctionStatement:
		evalFunctionStatement(s)
	case *ast.ReturnStatement:
		evalReturnStatement(s)
	case *ast.IfStatement:
		evalIfStatement(s)
	case *ast.BlockStatement:
		ctxt.CreateEnv()
		evalStatements(s.Statements)
		// should pop out the current environment no matter what
		defer ctxt.RestoreEnv()
	case *ast.ExpressionStatement:
		evalExpression(s.Expression)
	case *ast.LoopStatement:
		evalLoopStatement(s)
	}
}

func evalFunctionStatement(s *ast.FunctionStatement) {
	name := s.Ident.Value
	init := evalExpression(s.Value)
	if !ctxt.Env.Set(name, init) {
		panic(fmt.Errorf("variable %s already exists in current scope", name))
	}
}

func evalLoopStatement(s *ast.LoopStatement) {
	for {
		cond := s.Condition == nil || isTruthy(evalExpression(s.Condition))
		if !cond {
			break
		}

		evalStatement(s.Body)
	}
}

func evalLetStatement(s *ast.LetStatement) {
	name := s.Ident.Value
	init := evalExpression(s.InitValue)
	if !ctxt.Env.Set(name, init) {
		panic(fmt.Errorf("variable %s already exists in current scope", name))
	}
}

func evalReturnStatement(s *ast.ReturnStatement) {
	var retValue any
	if s.ReturnValue != nil {
		retValue = evalExpression(s.ReturnValue)
	}
	panic(objects.ReturnObject{Value: retValue})
}

func evalIfStatement(s *ast.IfStatement) {
	condition := evalExpression(s.Condition)

	if isTruthy(condition) {
		evalStatement(s.Then)
	} else if s.Else != nil {
		evalStatement(s.Else)
	}
}

func evalExpression(expr ast.Expression) any {
	defer exprErrorHandler(expr)

	switch e := expr.(type) {
	case *ast.InfixExpression:
		return evalInfixExpression(e)
	case *ast.PrefixExpression:
		return evalPrefixExpression(e)
	case *ast.Identifier:
		return evalIdentifier(e)
	case *ast.AssignExpression:
		return evalAssignExpression(e)
	case *ast.ArrayLiteral:
		return evalArrayLiteral(e)
	case *ast.MapLiteral:
		return evalMapLiteral(e)
	case *ast.StringLiteral:
		return e.Value
	case *ast.BoolLiteral:
		return e.Value
	case *ast.IntegerLiteral:
		return e.Value
	case *ast.FloatLiteral:
		return e.Value
	case *ast.FunctionLiteral:
		return evalFunctionLiteral(e)
	case *ast.CallExpression:
		return evalCallExpression(e)
	case *ast.IndexExpression:
		return evalIndexExpression(e)
	default:
		panic(fmt.Errorf("unknown expression type %T", expr))
	}
}

func evalArrayLiteral(e *ast.ArrayLiteral) *objects.ArrayObject {
	arr := &objects.ArrayObject{}

	for _, elem := range e.Elements {
		expr := evalExpression(elem)
		arr.List = append(arr.List, expr)
	}

	return arr
}

func evalMapLiteral(e *ast.MapLiteral) *objects.MapObject {
	mp := &objects.MapObject{
		Map: make(map[any]any),
	}

	for _, elem := range e.Elements {
		key := evalExpression(elem.Key)
		val := evalExpression(elem.Value)
		mp.Map[key] = val
	}

	return mp
}

func evalIndexExpression(e *ast.IndexExpression) any {
	left := evalExpression(e.Left)
	switch v := left.(type) {
	case *objects.ArrayObject:
		right := evalExpression(e.Index)
		index, ok := right.(int64)
		if !ok {
			panic(fmt.Errorf("expect integer index, got=%s", typeStr(index)))
		}

		if index >= int64(len(v.List)) || index < 0 {
			panic(fmt.Errorf("index out of range [%d]", index))
		}

		return v.List[index]
	case *objects.MapObject:
		right := evalExpression(e.Index)
		switch right.(type) {
		case int64:
			return v.Map[right]
		case float64:
			return v.Map[right]
		case string:
			return v.Map[right]
		case bool:
			return v.Map[right]
		default:
			panic(fmt.Errorf("only int, float, string and bool is allowed as key. got=%s",
				typeStr(v)))
		}
	}
	panic(fmt.Errorf("cannot index on type %s", typeStr(left)))
}

func evalCallExpression(e *ast.CallExpression) any {
	value := evalExpression(e.Callee)
	if value == nil {
		// trying to call on a null object
		panic(fmt.Errorf("cannot function call on null objects"))
	}

	args := evalCallArgs(e.Arguments)
	return callFunction(value, args)
}

func evalCallArgs(args []ast.Expression) []any {
	result := make([]any, 0)
	for _, e := range args {
		arg := evalExpression(e)
		result = append(result, arg)
	}

	return result
}

func callFunction(fn any, args []any) (retValue any) {
	switch obj := fn.(type) {
	case objects.FuncObject:
		returnRetriever := func() {
			ctxt.ResetEnv()

			err := recover()
			switch val := err.(type) {
			case objects.ReturnObject:
				retValue = val.Value // is a return value
			default:
				panic(retValue) // some runtime error
			}
		}
		defer returnRetriever() // set return value or propagate error

		// create new scope with the function's
		ctxt.SetEnv(obj.Env)

		function := obj.Fn
		if len(args) != len(function.Parameters) {
			panic(fmt.Errorf("incorrect no of arguments. got=%d, expect=%d",
				len(args), len(function.Parameters)))
		}

		for i, param := range function.Parameters {
			ctxt.Env.Set(param.Value, args[i])
		}

		evalStatements(function.Body.Statements)
		// reaching here means function does not return any value
		// in one of the control flow paths
		panic(objects.ReturnObject{Value: nil})
	case context.BuiltIn:
		return obj(args...)
	default:
		panic(fmt.Errorf("not a callable %s", typeStr(fn)))
	}
}

func evalFunctionLiteral(e *ast.FunctionLiteral) objects.FuncObject {
	return objects.FuncObject{
		Env: ctxt.Env,
		Fn:  e,
	}
}

func evalIdentifier(e *ast.Identifier) any {
	if value, ok := ctxt.Env.Get(e.Value); ok {
		return value
	}
	if value, ok := ctxt.GetBuiltIn(e.Value); ok {
		return value
	}

	panic(fmt.Errorf("variable not found: %s", e.Value))
}

func evalInfixExpression(e *ast.InfixExpression) any {
	left := evalExpression(e.Left)
	right := evalExpression(e.Right)
	switch e.Operator {
	case "+":
		return evalAddOperator(left, right)
	case "-":
		return evalSubOperator(left, right)
	case "*":
		return evalMulOperator(left, right)
	case "/":
		return evalDivOperator(left, right)
	case "<":
		return evalLtOperator(left, right)
	case ">":
		return evalGtOperator(left, right)
	case "<=":
		return !evalGtOperator(left, right)
	case ">=":
		return !evalLtOperator(left, right)
	case "==":
		return evalEqOperator(left, right)
	case "!=":
		return !evalEqOperator(left, right)
	default:
		panic(fmt.Errorf("unknown operator %s", e.Operator))
	}
}

func evalAssignExpression(expr *ast.AssignExpression) any {
	right := evalExpression(expr.Right)

	switch left := expr.Left.(type) {
	case *ast.Identifier:
		if !ctxt.Env.Assign(left.Value, right) {
			panic(fmt.Errorf("variable %q does not exist in current scope", left.Value))
		}
	case *ast.IndexExpression:
		l := evalExpression(left.Left)
		switch v := l.(type) {
		case *objects.ArrayObject:
			index, ok := evalExpression(left.Index).(int64)
			if !ok {
				panic(fmt.Errorf("expect integer index, got=%s", typeStr(index)))
			}

			if index >= int64(len(v.List)) || index < 0 {
				panic(fmt.Errorf("index out of range [%d]", index))
			}

			v.List[index] = right
		case *objects.MapObject:
			index := evalExpression(left.Index)
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
				panic(fmt.Errorf("only int, float, string and bool is allowed as key. got=%s",
					typeStr(v)))
			}
		}
	}

	return right
}

func evalAddOperator(left, right any) any {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l + r
		case float64:
			return float64(l) + r
		case string:
			return valueStr(l) + valueStr(r)
		default:
			panic(fmt.Errorf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l + float64(r)
		case float64:
			return l + r
		case string:
			return valueStr(l) + valueStr(r)
		default:
			panic(fmt.Errorf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case string:
		switch r := right.(type) {
		case string:
			return valueStr(l) + valueStr(r)
		case int64:
			return valueStr(l) + valueStr(r)
		case float64:
			return valueStr(l) + valueStr(r)
		default:
			panic(fmt.Errorf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case *objects.ArrayObject:
		switch r := right.(type) {
		case *objects.ArrayObject:
			return &objects.ArrayObject{
				List: slices.Concat(l.List, r.List),
			}
		default:
			panic(fmt.Errorf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case *objects.MapObject:
		switch r := right.(type) {
		case *objects.MapObject:
			mapObj := &objects.MapObject{
				Map: make(map[any]any),
			}
			maps.Copy(mapObj.Map, l.Map)
			maps.Copy(mapObj.Map, r.Map)
			return mapObj
		default:
			panic(fmt.Errorf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Errorf("addition not supported for %s", typeStr(l)))
	}
}

func evalSubOperator(left, right any) any {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l - r
		case float64:
			return float64(l) - r
		default:
			panic(fmt.Errorf("subtraction not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l - float64(r)
		case float64:
			return l - r
		default:
			panic(fmt.Errorf("subtraction not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Errorf("subtraction not supported for %s", typeStr(l)))
	}
}

func evalMulOperator(left, right any) any {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l * r
		case float64:
			return float64(l) * r
		default:
			panic(fmt.Errorf("multiplication not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l * float64(r)
		case float64:
			return l * r
		default:
			panic(fmt.Errorf("multiplication not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Errorf("multiplication not supported for %s", typeStr(l)))
	}
}

func evalDivOperator(left, right any) any {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l / r
		case float64:
			return float64(l) / r
		default:
			panic(fmt.Errorf("division not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l / float64(r)
		case float64:
			return l / r
		default:
			panic(fmt.Errorf("division not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Errorf("division not supported for %s", typeStr(l)))
	}
}

func evalLtOperator(left, right any) bool {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l < r
		case float64:
			return float64(l) < r
		default:
			panic(fmt.Errorf("cannot compare types %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l < float64(r)
		case float64:
			return l < r
		default:
			panic(fmt.Errorf("cannot compare types %s and %s", typeStr(l), typeStr(r)))
		}
	case string:
		switch r := right.(type) {
		case string:
			return l < r
		default:
			panic(fmt.Errorf("cannot compare types %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Errorf("comparison not supported for %s", typeStr(l)))
	}
}

func evalGtOperator(left, right any) bool {
	return !evalLtOperator(left, right) && !evalEqOperator(left, right)
}

func evalEqOperator(left, right any) bool {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l == r
		case float64:
			return float64(l) == r
		default:
			return false
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l == float64(r)
		case float64:
			return l == r
		default:
			return false
		}
	case bool:
		switch r := right.(type) {
		case bool:
			return l == r
		default:
			return false
		}
	case string:
		switch r := right.(type) {
		case string:
			return l == r
		default:
			return false
		}
	case *objects.ArrayObject:
		switch r := right.(type) {
		case *objects.ArrayObject:
			return slices.Equal(l.List, r.List)
		default:
			return false
		}
	case *objects.MapObject:
		switch r := right.(type) {
		case *objects.MapObject:
			return maps.Equal(l.Map, r.Map)
		default:
			return false
		}
	default:
		return false
	}
}

func evalPrefixExpression(e *ast.PrefixExpression) any {
	right := evalExpression(e.Right)
	switch e.Operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalNegateOperator(right)
	default:
		panic(fmt.Errorf("unknown operator %s", e.Operator))
	}
}

func evalBangOperator(e any) any {
	return !isTruthy(e)
}

func evalNegateOperator(e any) any {
	switch v := e.(type) {
	case int64:
		return -v
	case float64:
		return -v
	default:
		panic(fmt.Errorf("cannot negate value of type %s", typeStr(e)))
	}
}

func valueStr(val any) string {
	f, _ := ctxt.GetBuiltIn("str")
	str := f.(context.BuiltIn)

	return str(val).(string)
}

func typeStr(val any) string {
	f, _ := ctxt.GetBuiltIn("type")
	ty := f.(context.BuiltIn)

	return ty(val).(string)
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
