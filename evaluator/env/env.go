package env

type Environment struct {
	store map[string]any // symbol table for storing the symbols
	outer *Environment
}

func New(outer *Environment) *Environment {
	e := &Environment{
		store: make(map[string]any),
		outer: outer,
	}
	return e
}

func (e *Environment) Get(name string) (any, bool) {
	// search from current to outer for the variable
	value, ok := e.store[name]
	if !ok && e.outer != nil {
		value, ok = e.outer.Get(name)
	}

	return value, ok
}

func (e *Environment) Set(name string, value any) bool {
	if _, ok := e.store[name]; ok {
		return false
	}

	e.store[name] = value
	return true
}
