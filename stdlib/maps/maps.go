package maps

import (
	"RoLang/evaluator/objects"
	"RoLang/stdlib/builtin"
	"RoLang/stdlib/common"
	"maps"

	"fmt"
)

type Map struct {
	DispatchTable map[string]common.Sanitizer
}

func New() *Map {
	m := &Map{}
	m.DispatchTable = map[string]common.Sanitizer{
		"len":    m.lenSanitizer,
		"insert": m.insertSanitizer,
		"erase":  m.eraseSanitizer,
		"clone":  m.cloneSanitizer,
	}

	return m
}

func (m *Map) Dispatcher(name string) (common.Sanitizer, error) {
	sanitizer, ok := m.DispatchTable[name]
	if !ok {
		return nil, fmt.Errorf("no method %q found in strings module", name)
	}

	return sanitizer, nil
}

func (m *Map) lenSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len expects one argument, got=%d", len(args))
	}

	mp, ok := args[0].(*objects.MapObject)
	if !ok {
		return nil, fmt.Errorf("len expects argument to be map, got=%s", builtin.TypeStr(args[0]))
	}

	return mp.Len(), nil
}

func (m *Map) insertSanitizer(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("insert expects three arguments, got=%d", len(args))
	}

	mp, ok := args[0].(*objects.MapObject)
	if !ok {
		return nil, fmt.Errorf("insert expects first argument to be map, got=%s",
			builtin.TypeStr(args[0]))
	}

	return mp.Insert(args[1], args[2]), nil
}

func (m *Map) eraseSanitizer(args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("erase expects two arguments, got=%d", len(args))
	}

	mp, ok := args[0].(*objects.MapObject)
	if !ok {
		return nil, fmt.Errorf("erase expects first argument to be a map, got=%s",
			builtin.TypeStr(args[0]))
	}

	return mp.Erase(args[1]), nil
}

func (m *Map) cloneSanitizer(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("clone expects one argument, got=%d", len(args))
	}

	mp, ok := args[0].(*objects.MapObject)
	if !ok {
		return nil, fmt.Errorf("clone expects argument to be a map, got=%s",
			builtin.TypeStr(args[0]))
	}

	mpClone := &objects.MapObject{
		Map: maps.Clone(mp.Map),
	}

	return mpClone, nil
}
