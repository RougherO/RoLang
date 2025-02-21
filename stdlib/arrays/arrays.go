package arrays

import (
	"RoLang/evaluator/objects"
	"RoLang/stdlib/builtin"
	"RoLang/stdlib/common"
	"slices"

	"fmt"
)

type Arrays struct {
	DispatchTable map[string]common.Sanitizer
}

func New() *Arrays {
	a := &Arrays{}
	a.DispatchTable = map[string]common.Sanitizer{
		"len":    a.lenSanitizer,
		"push":   a.pushSanitizer,
		"pop":    a.popSanitizer,
		"insert": a.insertSanitizer,
		"erase":  a.eraseSanitizer,
		"concat": a.concatSanitizer,
		"copy":   a.copySanitizer,
	}

	return a
}

func (a *Arrays) Dispatcher(name string) (common.Sanitizer, error) {
	sanitizer, ok := a.DispatchTable[name]
	if !ok {
		return nil, fmt.Errorf("no method %q found in arrays module", name)
	}

	return sanitizer, nil
}

func (a *Arrays) lenSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len expects one argument, got=%d", len(args))
	}

	arr, ok := args[0].(*objects.ArrayObject)
	if !ok {
		return nil, fmt.Errorf("len expects an array type, got=%s", builtin.TypeStr(args[0]))
	}

	return arr.Len(), nil
}

func (a *Arrays) pushSanitizer(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("push expects two arguments, got=%d", len(args))
	}

	arr, ok := args[0].(*objects.ArrayObject)
	if !ok {
		return nil, fmt.Errorf("push expects first argument to be array, got=%s",
			builtin.TypeStr(args[0]))
	}

	return nil, arr.Insert(int(arr.Len()), args[0])
}

func (a *Arrays) popSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("pop expects one argument, got=%d", len(args))
	}
	arr, ok := args[0].(*objects.ArrayObject)
	if !ok {
		return nil, fmt.Errorf("pop expects argument to be array, got=%s",
			builtin.TypeStr(args[0]))
	}

	return arr.Erase(int(arr.Len()) - 1)
}

func (a *Arrays) insertSanitizer(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("insert expects three arguments, got=%d", len(args))
	}

	arr, ok := args[0].(*objects.ArrayObject)
	if !ok {
		return nil, fmt.Errorf("insert expects first argument to be array, got=%s",
			builtin.TypeStr(args[0]))
	}

	index, ok := args[1].(int64)
	if !ok {
		return nil, fmt.Errorf("insert expects second argument to be int, got=%s",
			builtin.TypeStr(args[1]))
	}

	return nil, arr.Insert(int(index), args[2])
}

func (a *Arrays) eraseSanitizer(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("erase expects two arguments, got=%d", len(args))
	}

	arr, ok := args[0].(*objects.ArrayObject)
	if !ok {
		return nil, fmt.Errorf("erase expects first argument to be array, got=%s",
			builtin.TypeStr(args[0]))
	}

	index, ok := args[1].(int64)
	if !ok {
		return nil, fmt.Errorf("erase expects second argument to be int, got=%s",
			builtin.TypeStr(args[1]))
	}

	return arr.Erase(int(index))
}

func (a *Arrays) concatSanitizer(args ...any) (any, error) {
	concat := &objects.ArrayObject{}

	for i, arg := range args {
		arr, ok := arg.(*objects.ArrayObject)
		if !ok {
			return nil, fmt.Errorf("concat expects all arguments to be array, arg=%d got=%s",
				i+1, builtin.TypeStr(arg))
		}

		concat.List = append(concat.List, arr.List...)
	}

	return concat, nil
}

func (a *Arrays) copySanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("copy expects one argument, got=%d", len(args))
	}

	arr, ok := args[0].(*objects.ArrayObject)
	if !ok {
		return nil, fmt.Errorf("copy expects argument to be array, got=%s", builtin.TypeStr(args[0]))
	}

	arrCopy := &objects.ArrayObject{
		List: slices.Clone(arr.List),
	}

	return arrCopy, nil
}
