package evaluator

import "RoLang/ast"

func Eval(node ast.Node) any {
	switch node := node.(type) {
	case *ast.Program:
		evalStatements(node.Statements)
		return nil
	case ast.Statement:
		evalStatement(node)
		return nil
	case ast.Expression:
		return evalExpression(node)
	default:
		return nil // TODO: replace with runtime errors
	}
}

func evalStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		evalStatement(stmt)
	}
}

func evalStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.IfStatement:
		evalIfStatement(s)
	case *ast.BlockStatement:
		evalStatements(s.Statements)
	case *ast.ExpressionStatement:
		evalExpression(s.Expression)
	}
}

func evalIfStatement(s *ast.IfStatement) {
	condition := Eval(s.Condition)

	if isTruthy(condition) {
		Eval(s.Then)
	} else if s.Else != nil {
		Eval(s.Else)
	}
}

func evalExpression(expr ast.Expression) any {
	switch e := expr.(type) {
	case *ast.InfixExpression:
		return evalInfixExpression(e)
	case *ast.PrefixExpression:
		return evalPrefixExpression(e)
	case *ast.BoolLiteral:
		return e.Value
	case *ast.IntegerLiteral:
		return e.Value
	case *ast.FloatLiteral:
		return e.Value
	default:
		return nil // TODO: replace with runtime errors
	}
}

func evalInfixExpression(e *ast.InfixExpression) any {
	left := Eval(e.Left)
	right := Eval(e.Right)
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
		return nil // TODO: replace with runtime errors
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
			return nil // TODO: replace with runtime errors
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l + float64(r)
		case float64:
			return l + r
		default:
			return nil // TODO: replace with runtime errors
		}
	default:
		return nil // TODO: replace with runtime errors
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
			return nil // TODO: replace with runtime errors
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l - float64(r)
		case float64:
			return l - r
		default:
			return nil // TODO: replace with runtime errors
		}
	default:
		return nil // TODO: replace with runtime errors
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
			return nil // TODO: replace with runtime errors
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l * float64(r)
		case float64:
			return l * r
		default:
			return nil // TODO: replace with runtime errors
		}
	default:
		return nil // TODO: replace with runtime errors
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
			return nil // TODO: replace with runtime errors
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l / float64(r)
		case float64:
			return l / r
		default:
			return nil // TODO: replace with runtime errors
		}
	default:
		return nil // TODO: replace with runtime errors
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
			return false // TODO: replace with runtime errors
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l < float64(r)
		case float64:
			return l < r
		default:
			return false // TODO: replace with runtime errors
		}
	default:
		return false // TODO: replace with runtime errors
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
			return false // TODO: replace with runtime errors
		}
	case float64:
		switch r := right.(type) {
		case int64:
			return l == float64(r)
		case float64:
			return l == r
		default:
			return false // TODO: replace with runtime errors
		}
	case bool:
		switch r := right.(type) {
		case bool:
			return l == r
		default:
			return false // TODO: replace with runtime errors
		}
	default:
		return false // TODO: replace with runtime errors
	}
}

func evalPrefixExpression(e *ast.PrefixExpression) any {
	right := Eval(e.Right)
	switch e.Operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalNegateOperator(right)
	default:
		return nil // TODO: replace with runtime errors
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
		return nil // TODO: replace with runtime errors
	}
}

func isTruthy(value any) bool {
	switch value {
	case false:
		fallthrough
	case nil:
		fallthrough
	case 0:
		fallthrough
	case 0.0:
		fallthrough
	case "":
		return false
	default:
		return true
	}
}
