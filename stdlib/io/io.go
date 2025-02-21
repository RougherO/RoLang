package io

import (
	"RoLang/stdlib/common"
	"RoLang/stdlib/strings"

	"bufio"
	"fmt"
	"os"
)

type Io struct {
	DispatchTable map[string]common.Sanitizer

	scanner *bufio.Reader
	printer *bufio.Writer
}

func New() *Io {
	io := &Io{}
	io.DispatchTable = map[string]common.Sanitizer{
		"readln":  io.readlnSanitizer,
		"print":   io.printSanitizer,
		"println": io.printlnSanitizer,
	}
	io.scanner = bufio.NewReader(os.Stdin)
	io.printer = bufio.NewWriter(os.Stdout)
	return io
}

func (io *Io) Dispatcher(name string) (common.Sanitizer, error) {
	sanitizer, ok := io.DispatchTable[name]
	if !ok {
		return nil, fmt.Errorf("no method %q found in io module", name)
	}

	return sanitizer, nil
}

func (io *Io) readlnSanitizer(args ...any) (any, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("extra arguments in readln")
	}

	return io.scanner.ReadString('\n')
}

func (io *Io) printSanitizer(args ...any) (any, error) {
	var err error
	for _, value := range args {
		_, err = io.printer.WriteString(strings.From(value))
		if err != nil {
			return nil, err
		}
	}
	err = io.printer.Flush()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (io *Io) printlnSanitizer(args ...any) (any, error) {
	var err error
	for _, value := range args {
		_, err = io.printer.WriteString(strings.From(value))
		if err != nil {
			return nil, err
		}
	}
	err = io.printer.WriteByte('\n')
	if err != nil {
		return nil, err
	}

	err = io.printer.Flush()
	if err != nil {
		return nil, err
	}

	return nil, nil
}
