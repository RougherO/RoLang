package repl

import (
	"RoLang/evaluator"
	"RoLang/lexer"
	"RoLang/parser"

	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const prompt = "|> "
const message = `RoLang v0.5 Tree-Walk Interpreter`

func checkError(errs []error) bool {
	if len(errs) != 0 {
		fmt.Fprintln(os.Stderr, errors.Join(errs...))
		return true
	}

	return false
}

func Start() {
	fmt.Println(message)
	scanner := bufio.NewScanner(os.Stdin)
	e := evaluator.New()
	for {
		fmt.Print(prompt)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := strings.TrimSpace(scanner.Text())
		size := len(line)
		if size == 0 {
			continue
		}

		l := lexer.New("repl", line)
		p := parser.New(l)

		program, errs := p.Parse()
		if checkError(errs) {
			continue
		}

		errs = e.Evaluate(program)
		if checkError(errs) {
			continue
		}
	}
}
