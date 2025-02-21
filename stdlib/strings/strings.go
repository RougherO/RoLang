package strings

import (
	"RoLang/evaluator/objects"
	"RoLang/stdlib/builtin"
	"RoLang/stdlib/common"
	"strconv"

	"fmt"
	"strings"
)

type String struct {
	DispatchTable map[string]common.Sanitizer
}

func New() *String {
	s := &String{}
	s.DispatchTable = map[string]common.Sanitizer{
		"from":       s.fromSanitizer,
		"len":        s.lenSanitizer,
		"trim":       s.trimSanitizer,
		"trimSpace":  s.trimSpaceSanitizer,
		"split":      s.splitSanitizer,
		"splitSpace": s.splitSpaceSanitizer,
	}
	return s
}

func (s *String) Dispatcher(name string) (common.Sanitizer, error) {
	sanitizer, ok := s.DispatchTable[name]
	if !ok {
		return nil, fmt.Errorf("no method %q found in strings module", name)
	}

	return sanitizer, nil
}

func (s *String) fromSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("from expects one argument, got=%d", len(args))
	}

	return From(args[0]), nil
}

func (s *String) lenSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len expects one argument, got=%d", len(args))
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("len expects its argument to be string, got=%s",
			builtin.TypeStr(args[0]))
	}
	return Len(str), nil
}

func (s *String) trimSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trim expects one argument, got=%d", len(args))
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("trim expects first argument to be string, got=%s",
			builtin.TypeStr(args[0]))
	}
	cut, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("trim expects second argument to be string, got=%s",
			builtin.TypeStr(args[1]))
	}

	return Trim(str, cut), nil
}

func (s *String) trimSpaceSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trimSpace expects one argument, got=%d", len(args))
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("trimSpace expects argument to be string, got=%s",
			builtin.TypeStr(args[0]))
	}

	return Trim(str, " "), nil
}

func (s *String) splitSanitizer(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("split expects two arguments, got=%d", len(args))
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("split expects first argument to be string, got=%s",
			builtin.TypeStr(args[0]))
	}
	sep, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("split expects second argument to be string, got=%s",
			builtin.TypeStr(args[0]))
	}

	return Split(str, sep), nil
}

func (s *String) splitSpaceSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("splitSpace expects one argument, got=%d", len(args))
	}
	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("splitSpace expects argument to be string, got=%s",
			builtin.TypeStr(args[0]))
	}

	return Split(str, " "), nil
}

func From(value any) string {
	var out string
	switch v := value.(type) {
	case int64:
		out += strconv.FormatInt(v, 10)
	case float64:
		out += strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		out += v
	case bool:
		out += strconv.FormatBool(v)
	case *objects.MapObject:
		out += "{"
		i := 0
		for k, v := range v.Map {
			key := From(k)
			val := From(v)
			elem := key + ": " + val
			if i == 0 {
				out += elem
			} else {
				out += ", " + elem
			}
			i++
		}
		out += "}"
	case *objects.ArrayObject:
		out += "["
		for i, e := range v.List {
			if i == 0 {
				out += From(e)
			} else {
				out += ", " + From(e)
			}
		}
		out += "]"
	case *objects.FuncObject:
		out += "function"
	case *common.Sanitizer:
		out += "function"
	case nil:
		out += "null"
	default:
		out += "<unknown>"
	}

	return out
}

func Len(str string) int64 {
	return int64(len(str))
}

func Trim(str string, cut string) string {
	return strings.TrimSpace(str)
}

func Split(str string, sep string) *objects.ArrayObject {
	result := &objects.ArrayObject{}
	splits := strings.Split(str, sep)

	for _, split := range splits {
		result.List = append(result.List, split)
	}

	return result
}
