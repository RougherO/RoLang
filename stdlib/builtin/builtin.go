package builtin

import (
	"RoLang/evaluator/objects"
	"RoLang/stdlib/common"

	"fmt"
)

type BuiltIn struct {
	DispatchTable map[string]common.Sanitizer
}

func New() *BuiltIn {
	return &BuiltIn{
		DispatchTable: map[string]common.Sanitizer{
			"type": typeStrSanitizer,
		},
	}
}

func (b *BuiltIn) Dispatcher(name string) (common.Sanitizer, error) {
	sanitizer, ok := b.DispatchTable[name]
	if !ok {
		return nil, fmt.Errorf("no builtin function %q found", name)
	}

	return sanitizer, nil
}

func typeStrSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("`type` expects only a single argument, got=%d",
			len(args))
	}

	return TypeStr(args[0]), nil
}

func TypeStr(value any) string {
	switch value.(type) {
	case int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case bool:
		return "bool"
	case *objects.MapObject:
		return "map"
	case *objects.ArrayObject:
		return "array"
	case *objects.FuncObject:
		return "function"
	case nil:
		return "null"
	default:
		return "<unknown>"
	}
}
