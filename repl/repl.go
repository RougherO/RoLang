package repl

import (
	"RoLang/lexer"
	"RoLang/parser"

	"bufio"
	"fmt"
	"io"
	"strings"
)

const prompt = "|> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Print(prompt)

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		lexer := lexer.New("repl", line)
		parser := parser.New(lexer)

		program := parser.Parse()
		if program == nil {
			fmt.Println(strings.Join(parser.Errors(), "\n"))
			continue
		}

		for _, stmt := range program.Statements {
			fmt.Printf("%s\n", stmt)
		}
	}
}
