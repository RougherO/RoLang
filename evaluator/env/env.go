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

// setting a new value using `let` statements
func (e *Environment) Set(name string, value any) bool {
	if _, ok := e.store[name]; ok {
		return false
	}

	e.store[name] = value
	return true
}

// setting an already existing value using `=`
func (e *Environment) Assign(name string, value any) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = value
		return true
	}

	if e.outer != nil {
		return e.Assign(name, value)
	}

	return false
}

func (e *Environment) Outer() *Environment {
	return e.outer
}
