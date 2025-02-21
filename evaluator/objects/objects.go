package objects

import (
	"RoLang/ast"
	"RoLang/evaluator/env"
	"fmt"

	"slices"
)

type (
	ReturnObject struct {
		Value any
	}
	FuncObject struct {
		Env      *env.Environment
		Function *ast.FunctionLiteral
	}
	ArrayObject struct {
		List []any
	}
	MapObject struct {
		Map map[any]any
	}
)

func (o *ArrayObject) Insert(index int, e any) error {
	if index < 0 || index > len(o.List) {
		return fmt.Errorf("index out of bounds [%d]", index)
	}
	o.List = slices.Insert(o.List, index, e)

	return nil
}

func (o *ArrayObject) Erase(index int) (any, error) {
	if index < 0 || index >= len(o.List) {
		return nil, fmt.Errorf("index out of bounds [%d]", index)
	}
	e := o.List[index]
	o.List = append(o.List[:index], o.List[index+1:]...)
	return e, nil
}

func (o *ArrayObject) Len() int64 {
	return int64(len(o.List))
}

func (o *MapObject) Insert(key any, val any) bool {
	if _, ok := o.Map[key]; ok {
		return false
	}

	o.Map[key] = val
	return true
}

func (o *MapObject) Erase(key any) any {
	val, ok := o.Map[key]
	if ok {
		delete(o.Map, key)
		return val
	}

	return nil
}

func (o *MapObject) Len() int64 {
	return int64(len(o.Map))
}
