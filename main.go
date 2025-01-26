package main

import (
	"RoLang/repl"
	"fmt"
	"os"
)

const MESSAGE = `RoLang v0.1 Tree-Walk Interpreter`

func main() {
	fmt.Println(MESSAGE)
	repl.Start(os.Stdin, os.Stdout)
}
