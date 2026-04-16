package main

import (
	"fmt"
	"os"

	"os/exec"
	"path/filepath"
	"strings"

	"github.com/devguinness/Riner/internal/codegen"
	"github.com/devguinness/Riner/internal/interpreter"
	"github.com/devguinness/Riner/internal/ir"
	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
	"github.com/devguinness/Riner/internal/sema"
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
	case "build":
		buildFile(file)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", command)
		fmt.Fprintln(os.Stderr, "usage: riner run <file.rn>")
		fmt.Fprintln(os.Stderr, "       riner build <file.rn>")
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

	c := sema.New()
	if err := c.Check(prog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	i := interpreter.New()
	i.Run(prog)
}

func buildFile(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not read file %q: %s\n", path, err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	tokens := l.Tokenize()

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

	c := sema.New()
	if err := c.Check(prog); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	b := ir.NewBuilder()
	irprog := b.Build(prog)

	g := codegen.New()
	csrc := g.Generate(irprog)

	// write C file next to source
	base := strings.TrimSuffix(filepath.Base(path), ".rn")
	cfile := base + ".c"
	if err := os.WriteFile(cfile, []byte(csrc), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing C file: %s\n", err)
		os.Exit(1)
	}

	// compile with gcc
	outfile := base
	runtimeDir := filepath.Join(filepath.Dir(os.Args[0]), "runtime")
cmd := exec.Command("gcc", "-I", runtimeDir, "-o", outfile, cfile)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
    	fmt.Fprintf(os.Stderr, "compilation failed: %s\n", err)
    	os.Exit(1)
	}

	// remove intermediate C file
	// os.Remove(cfile)

	fmt.Printf("built: %s\n", outfile)
}