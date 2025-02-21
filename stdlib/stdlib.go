package stdlib

import (
	"RoLang/stdlib/arrays"
	"RoLang/stdlib/builtin"
	"RoLang/stdlib/common"
	"RoLang/stdlib/io"
	"RoLang/stdlib/maps"
	"RoLang/stdlib/strings"

	"fmt"
)

type Module interface {
	Dispatcher(string) (common.Sanitizer, error)
}

type StdLib struct {
	Modules map[string]Module
}

func New() *StdLib {
	return &StdLib{
		Modules: map[string]Module{
			"arrays":  arrays.New(),
			"builtin": builtin.New(),
			"io":      io.New(),
			"maps":    maps.New(),
			"strings": strings.New(),
		},
	}
}

func (s *StdLib) GetModuleDispatcher(name string, method string) (common.Sanitizer, error) {
	module, ok := s.Modules[name]
	if !ok {
		return nil, fmt.Errorf("no module named %q", name)
	}

	return module.Dispatcher(method)
}
