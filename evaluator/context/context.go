package context

import (
	"RoLang/evaluator/env"
	"RoLang/evaluator/objects"

	"fmt"
	"io"
	"strconv"
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
		"str": func(args ...any) any {
			var out string
			for _, arg := range args {
				switch v := arg.(type) {
				case int64:
					out += strconv.FormatInt(v, 10)
				case float64:
					out += strconv.FormatFloat(v, 'f', -1, 64)
				case string:
					out += v
				case objects.FuncObject:
					out += "function"
				case nil:
					out += "null"
				default:
				}
			}
			return out
		},
		"len": func(args ...any) any {
			if len(args) != 1 {
				panic(fmt.Errorf("\n`len` expects only 1 argument, got=%d", len(args)))
			}
			switch v := args[0].(type) {
			case string:
				return len(v)
			case []any:
				return len(v)
			default:
				panic(fmt.Errorf("\nargument type not supported for `len`, got=%v", c.builtins["type"](v)))
			}
		},
		"type": func(args ...any) any {
			if len(args) != 1 {
				panic(fmt.Errorf("\n`type` expects only 1 argument, got=%d", len(args)))
			}

			switch args[0].(type) {
			case int64:
				return "int"
			case float64:
				return "float"
			case string:
				return "string"
			case bool:
				return "bool"
			case []any:
				return "array"
			case objects.FuncObject:
				return "function"
			case BuiltIn:
				return "builtin"
			case nil:
				return "null"
			default:
				return "<unknown>"
			}
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
