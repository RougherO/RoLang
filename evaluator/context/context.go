package context

import (
	"RoLang/evaluator/env"

	"fmt"
	"io"
)

type BuiltIn func(...any) any

type Context struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
	Env *env.Environment

	envStack []*env.Environment
	builtins map[string]BuiltIn
}

// Create a new context for the evaluator
func New(in io.Reader, out, err io.Writer) *Context {
	c := Context{
		In:  in,
		Out: out,
		Err: err,
		Env: env.New(nil),
	}

	c.builtins = map[string]BuiltIn{
		"puts": func(args ...any) any {
			var out string
			for i, arg := range args {
				if arg == nil {
					arg = "null"
				}
				if i == 0 {
					out += fmt.Sprintf("%v", arg)
				} else {
					out += ", " + fmt.Sprintf("%v", arg)
				}
			}
			io.WriteString(c.Out, out+"\n")
			return nil
		},
	}

	return &c
}

// for block scopes it should just enclose the current environment
func (c *Context) CreateEnv() {
	c.Env = env.New(c.Env)
}

// for function calls it should set up a new environment
// with outer function as the one provided in parameter
func (c *Context) SetEnv(e *env.Environment) {
	c.envStack = append(c.envStack, c.Env)
	c.Env = env.New(e)
}

// resets the environment set with `SetEnv“ function
func (c *Context) ResetEnv() {
	c.Env, c.envStack = c.envStack[len(c.envStack)-1], c.envStack[:len(c.envStack)-1]
}

// restores the environment set with `CreateEnv` function
// by just 'moving up' the environment chain
func (c *Context) RestoreEnv() {
	c.Env = c.Env.Outer()
}

// used to retrieve builtin functions
func (c *Context) GetBuiltIn(name string) (any, bool) {
	fn, ok := c.builtins[name]
	return fn, ok
}
