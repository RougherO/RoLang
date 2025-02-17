package main

import (
	"RoLang/repl"

	"os"
)

func main() {
	repl.Start(os.Stdin, os.Stdout, os.Stderr)
}
