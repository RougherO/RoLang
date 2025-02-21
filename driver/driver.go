package driver

import (
	"RoLang/evaluator"
	"RoLang/lexer"
	"RoLang/parser"

	"errors"
	"fmt"
	"os"
)

func Execute(file string, code string) {
	lexer := lexer.New(file, code)
	parser := parser.New(lexer)
	evaluator := evaluator.New()

	program, errs := parser.Parse()
	if len(errs) != 0 {
		fmt.Fprintln(os.Stderr, errors.Join(errs...))
	}

	errs = evaluator.Evaluate(program)
	if len(errs) != 0 {
		fmt.Fprintln(os.Stderr, errors.Join(errs...))
	}
}
