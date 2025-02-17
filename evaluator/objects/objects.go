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
	ArrayObject struct {
		List []any
	}
)

func (o *ArrayObject) Push(e any) {
	o.List = append(o.List, e)
}

func (o *ArrayObject) Pop(index int) any {
	if index < 0 || index >= len(o.List) {
		return nil
	}
	e := o.List[index]
	o.List = append(o.List[:index], o.List[index+1:]...)
	return e
}
