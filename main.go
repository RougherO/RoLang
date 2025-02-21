package main

import (
	"RoLang/driver"
	"RoLang/repl"

	"fmt"
	"io"
	"os"
)

const usage = `Usage: RoLang [FILE]
	If FILE is absent starts the RoLang interpreter.
	Otherwise interpretes the FILE.`

func main() {
	if len(os.Args) == 1 {
		repl.Start()
	} else if len(os.Args) == 2 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(1)
		}

		bytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			os.Exit(1)
		}

		driver.Execute(file.Name(), string(bytes))
	} else {
		fmt.Println(usage)
	}
}
