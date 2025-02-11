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

func New(in io.Reader, out, err io.Writer) *Context {
	c := Context{
		In:  in,
		Out: out,
		Err: err,
		Env: env.New(nil),
	}

	c.builtins = map[string]BuiltIn{
		"puts": func(args ...any) any {
			for _, value := range args {
				fmt.Printf("%v\n", value)
			}
			return nil
		},
	}

	return &c
}

func (c *Context) CreateEnv(e *env.Environment) {
	c.envStack = append(c.envStack, c.Env)
	c.Env = env.New(e)
}

func (c *Context) RestoreEnv() {
	c.Env, c.envStack = c.envStack[len(c.envStack)-1], c.envStack[:len(c.envStack)-1]
}

func (c *Context) GetBuiltIn(name string) (any, bool) {
	fn, ok := c.builtins[name]
	return fn, ok
}
