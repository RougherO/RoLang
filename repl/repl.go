package repl

import (
	"RoLang/evaluator"
	"RoLang/lexer"
	"RoLang/parser"

	"bufio"
	"fmt"
	"io"
	"strings"
)

const prompt = "|> "

func Start(in io.Reader, out io.Writer, err io.Writer) {
	scanner := bufio.NewScanner(in)
	evaluator.Init(in, out, err)
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

		// var node ast.Node
		// // lines ending with semicolon are statements
		// if c := line[size-1]; c == ';' {
		// 	node = p.ParseStatement()
		// } else {
		// 	node = p.ParseExpression(parser.NONE)
		// }
		program := p.Parse()

		if len(p.Errors()) != 0 {
			io.WriteString(err, strings.Join(p.Errors(), "\n"))
			io.WriteString(err, "\n")
			io.WriteString(out, "null\n")
			continue
		}

		evaluator.Evaluate(program)

		// if result != nil {
		// 	io.WriteString(out, fmt.Sprintf("%v", result))
		// 	io.WriteString(out, "\n")
		// } else if len(errors) != 0 {
		// 	io.WriteString(err, evaluator.Errors())
		// 	io.WriteString(err, "\n")
		// 	io.WriteString(out, "null\n")
		// 	evaluator.ClearErr()
		// }
	}
}
