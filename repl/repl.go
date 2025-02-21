package repl

import (
	"RoLang/evaluator"
	"RoLang/lexer"
	"RoLang/parser"

	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const prompt = "|> "
const message = `RoLang v0.3 Tree-Walk Interpreter`

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
		program := p.Parse()

		if len(p.Errors()) != 0 {
			io.WriteString(os.Stderr, strings.Join(p.Errors(), "\n"))
			io.WriteString(os.Stderr, "\n")
			io.WriteString(os.Stdout, "null\n")
			continue
		}

		e.Evaluate(program)
		if len(e.Errors) != 0 {
			fmt.Println(errors.Join(e.Errors...))
			e.Errors = e.Errors[:0]
		}
	}
}
