package objects

import (
	"RoLang/ast"
	"RoLang/evaluator/env"
)

type (
	ReturnObject struct {
		Value any
	}
	FuncObject struct {
		Env *env.Environment
		Fn  *ast.FunctionLiteral
	}
)
