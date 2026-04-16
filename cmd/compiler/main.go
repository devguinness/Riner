package main

import (
	"fmt"
	"os"

	"github.com/devguinness/Riner/internal/interpreter"
	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: riner run <file.rn>")
		os.Exit(1)
	}

	command := os.Args[1]
	file := os.Args[2]

	switch command {
	case "run":
		runFile(file)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "usage: riner run <file.rn>")
		os.Exit(1)
	}
}

func runFile(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	tokens := l.Tokenize()

	// check for illegal tokens
	for _, tok := range tokens {
		if tok.Type == lexer.TOKEN_ILLEGAL {
			fmt.Fprintf(os.Stderr, "lexer error at line %d, col %d: unexpected character %q\n", tok.Line, tok.Column, tok.Value)
			os.Exit(1)
		}
	}

	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	i := interpreter.New()
	i.Run(prog)
}