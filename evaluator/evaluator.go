package evaluator

import (
	"RoLang/ast"
	"RoLang/evaluator/context"
	"RoLang/evaluator/env"

	"fmt"
	"io"
	"os"
)

type (
	returnObject struct {
		value any
	}
	fnObject struct {
		env *env.Environment
		fn  *ast.FunctionLiteral
	}
)

// global context variable to maintain
// evaluator states
var ctxt *context.Context

func Init(in io.Reader, out, err io.Writer) {
	ctxt = context.New(in, out, err)
}

func recoverHandler() {
	err := recover()

	if err != nil {
		switch e := err.(type) {
		case returnObject:
			switch e := e.value.(type) {
			case int64:
				os.Exit(int(e))
			case nil:
				os.Exit(0)
			default:
				io.WriteString(ctxt.Err, "can only return integer exit codes at top level\n")
			}
		default:
			io.WriteString(ctxt.Err, fmt.Sprintf("recovering: %v", e))
		}
	}
}

func Evaluate(program *ast.Program) {
	defer recoverHandler()
	evalStatements(program.Statements)
}

func exprErrorHandler(expr ast.Expression) {
	err := recover()

	if err != nil {
		switch err.(type) {
		case returnObject:
			panic(err)
		default:
			io.WriteString(ctxt.Err, fmt.Sprintf("%s %s\n", expr.Location(), err))
		}
	}
}

func stmtErrorHandler(stmt ast.Statement) {
	err := recover()

	if err != nil {
		switch err.(type) {
		case returnObject:
			panic(err)
		default:
			io.WriteString(ctxt.Err, fmt.Sprintf("%s %s\n", stmt.Location(), err))
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
		ctxt.CreateEnv(ctxt.Env)
		evalStatements(s.Statements)
		// should pop out the current environment no matter what
		defer ctxt.RestoreEnv()
	case *ast.ExpressionStatement:
		evalExpression(s.Expression)
	}
}

func evalFunctionStatement(s *ast.FunctionStatement) {
	name := s.Ident.Value
	init := evalExpression(s.Value)
	if !ctxt.Env.Set(name, init) {
		panic(fmt.Sprintf("variable %s already exists in current scope", name))
	}
}

func evalLetStatement(s *ast.LetStatement) {
	init := evalExpression(s.InitValue)
	if init == nil {
		// some error occured
		// do not initialise variable
		return
	}

	name := s.Ident.Value
	if !ctxt.Env.Set(name, init) {
		panic(fmt.Sprintf("variable %s already exists in current scope", name))
	}
}

func evalReturnStatement(s *ast.ReturnStatement) {
	var retValue any
	if s.ReturnValue != nil {
		retValue = evalExpression(s.ReturnValue)
	}
	panic(returnObject{value: retValue})
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
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
}

func evalCallExpression(e *ast.CallExpression) any {
	value := evalExpression(e.Callee)
	if value == nil {
		return nil
	}

	args := evalCallArgs(e.Arguments)
	if args == nil {
		return nil
	}
	return callFunction(value, args)
}

func evalCallArgs(args []ast.Expression) []any {
	result := make([]any, 0)
	for _, e := range args {
		arg := evalExpression(e)
		if arg == nil {
			return nil
		}
		result = append(result, arg)
	}

	return result
}

func callFunction(fn any, args []any) (retValue any) {
	switch obj := fn.(type) {
	case *fnObject:
		returnRetriever := func() {
			ctxt.RestoreEnv()

			err := recover()
			switch val := err.(type) {
			case returnObject:
				retValue = val.value // is a return value
			default:
				panic(retValue) // some runtime error
			}
		}
		defer returnRetriever() // set return value or propagate error

		// create new scope with the function's
		ctxt.CreateEnv(obj.env)

		function := obj.fn
		if len(args) != len(function.Parameters) {
			panic(fmt.Sprintf("incorrect no of arguments. got=%d, expect=%d",
				len(args), len(function.Parameters)))
		}

		for i, param := range function.Parameters {
			ctxt.Env.Set(param.Value, args[i])
		}

		evalStatements(function.Body.Statements)
		panic("should not reach here") // unreachable
	case context.BuiltIn:
		return obj(args...)
	default:
		panic(fmt.Sprintf("not a callable %s", typeStr(fn)))
	}
}

func evalFunctionLiteral(e *ast.FunctionLiteral) *fnObject {
	return &fnObject{
		env: ctxt.Env,
		fn:  e,
	}
}

func evalIdentifier(e *ast.Identifier) any {
	if value, ok := ctxt.Env.Get(e.Value); ok {
		return value
	}
	if value, ok := ctxt.GetBuiltIn(e.Value); ok {
		return value
	}

	panic(fmt.Sprintf("variable not found: %s", e.Value))
}

func evalInfixExpression(e *ast.InfixExpression) any {
	left := evalExpression(e.Left)
	if left == nil {
		return nil
	}

	right := evalExpression(e.Right)
	if right == nil {
		return nil
	}

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
		panic(fmt.Sprintf("unknown operator %s", e.Operator))
	}
}

func evalAddOperator(left, right any) any {
	switch l := left.(type) {
	case int64:
		switch r := right.(type) {
		case int64:
			return l + r
		case float64:
			return float64(l) + r
		default:
			panic(fmt.Sprintf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l + float64(r)
		case float64:
			return l + r
		default:
			panic(fmt.Sprintf("addition not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Sprintf("addition not supported for %s", typeStr(l)))
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
			panic(fmt.Sprintf("subtraction not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l - float64(r)
		case float64:
			return l - r
		default:
			panic(fmt.Sprintf("subtraction not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Sprintf("subtraction not supported for %s", typeStr(l)))
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
			panic(fmt.Sprintf("multiplication not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l * float64(r)
		case float64:
			return l * r
		default:
			panic(fmt.Sprintf("multiplication not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Sprintf("multiplication not supported for %s", typeStr(l)))
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
			panic(fmt.Sprintf("division not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l / float64(r)
		case float64:
			return l / r
		default:
			panic(fmt.Sprintf("division not supported for %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Sprintf("division not supported for %s", typeStr(l)))
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
			panic(fmt.Sprintf("cannot compare types %s and %s", typeStr(l), typeStr(r)))
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l < float64(r)
		case float64:
			return l < r
		default:
			panic(fmt.Sprintf("cannot compare types %s and %s", typeStr(l), typeStr(r)))
		}
	default:
		panic(fmt.Sprintf("comparison not supported for %s", typeStr(l)))
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
	default:
		return false
	}
}

func evalPrefixExpression(e *ast.PrefixExpression) any {
	right := evalExpression(e.Right)
	if right == nil {
		return nil
	}

	switch e.Operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalNegateOperator(right)
	default:
		panic(fmt.Sprintf("unknown operator %s", e.Operator))
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
		panic(fmt.Sprintf("cannot negate value of type %s", typeStr(e)))
	}
}

func typeStr(ty any) string {
	if ty == nil {
		return "null"
	}

	switch ty.(type) {
	case int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case bool:
		return "bool"
	case fnObject:
		return "function"
	case context.BuiltIn:
		return "builtin"
	default:
		return "<unknown>"
	}
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
