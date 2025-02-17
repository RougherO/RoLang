package context

import (
	"RoLang/evaluator/env"
	"RoLang/evaluator/objects"

	"fmt"
	"io"
	"slices"
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
				if i == 0 {
					out += c.builtins["str"](arg).(string)
				} else {
					out += " " + c.builtins["str"](arg).(string)
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
				case *objects.ArrayObject:
					out += "["
					for i, e := range v.List {
						if i == 0 {
							out += c.builtins["str"](e).(string)
						} else {
							out += ", " + c.builtins["str"](e).(string)
						}
					}
					out += "]"
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
			case *objects.ArrayObject:
				return len(v.List)
			default:
				panic(fmt.Errorf("\nargument type not supported for `len`, got=%v",
					c.builtins["type"](v)))
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
			case *objects.ArrayObject:
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
		"push": func(args ...any) any {
			if len(args) != 2 {
				panic(fmt.Errorf("`push` expects 2 arguments, got=%d", len(args)))
			}
			arr, ok := args[0].(*objects.ArrayObject)
			if !ok {
				panic(fmt.Errorf("expect first argument to be array type, got=%s",
					c.builtins["type"](args[0])))
			}

			arr.Push(args[1])
			return nil
		},
		"pop": func(args ...any) any {
			if len(args) != 1 {
				panic(fmt.Errorf("`pop` expects 1 argument, got=%d", len(args)))
			}
			arr, ok := args[0].(*objects.ArrayObject)
			if !ok {
				panic(fmt.Errorf("`pop` expects array type, got=%s", c.builtins["type"](args[0])))
			}

			return arr.Pop(len(arr.List) - 1)
		},
		"first": func(args ...any) any {
			var out []any

			for i, arg := range args {
				arr, ok := arg.(*objects.ArrayObject)
				if !ok {
					panic(fmt.Errorf("`first` expects all arguments to be of array type, got argument %d=%s",
						i, c.builtins["type"](arg)))
				}
				if len(arr.List) == 0 {
					continue
				}
				out = append(out, arr.List[0])
			}

			return out
		},
		"last": func(args ...any) any {
			var out []any

			for i, arg := range args {
				arr, ok := arg.(*objects.ArrayObject)
				if !ok {
					panic(fmt.Errorf("`last` expects all arguments to be of array type, got argument %d=%s",
						i, c.builtins["type"](arg)))
				}
				if len(arr.List) == 0 {
					continue
				}
				out = append(out, arr.List[len(arr.List)-1])
			}

			return out
		},
		"erase": func(args ...any) any {
			if len(args) != 2 {
				panic(fmt.Errorf("`erase` expects 2 arguments, got=%d", len(args)))
			}
			switch v := args[0].(type) {
			case *objects.ArrayObject:
				index, ok := args[1].(int64)
				if !ok {
					panic(fmt.Errorf("expect integer index, got=%s", c.builtins["type"](index)))
				}
				e := v.Pop(int(index))
				return e // return the erased element
			}
			panic(fmt.Errorf("`erase` expects an array or map type, got=%s", c.builtins["type"](args[0])))
		},
		"clone": func(args ...any) any {
			if len(args) != 1 {
				panic(fmt.Errorf("`clone` expects 1 argument, got=%d", len(args)))
			}
			switch v := args[0].(type) {
			case *objects.ArrayObject:
				return &objects.ArrayObject{
					List: slices.Clone(v.List),
				}
			}

			panic(fmt.Errorf("`clone` expects an array or map type, got=%s", c.builtins["type"](args[0])))
		},
		"clear": func(args ...any) any {
			if len(args) != 1 {
				panic(fmt.Errorf("`clear` expects 1 argument, got=%d", len(args)))
			}

			switch v := args[0].(type) {
			case *objects.ArrayObject:
				v.List = v.List[:0]
				return nil
			}

			panic(fmt.Errorf("`clear` expects an array or map type, got=%s", c.builtins["type"](args[0])))
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
