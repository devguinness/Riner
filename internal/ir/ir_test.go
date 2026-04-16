package ir

import (
	"strings"
	"testing"

	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
)

func build(src string) *Program {
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		panic(err)
	}
	b := NewBuilder()
	return b.Build(prog)
}

func TestHelloWorld(t *testing.T) {
	ir := build(`func main() { print("hello, world") }`)
	dump := ir.Dump()
	if !strings.Contains(dump, "func main:") {
		t.Error("expected func main in IR")
	}
	if !strings.Contains(dump, "call print") {
		t.Error("expected call print in IR")
	}
}

func TestArithmetic(t *testing.T) {
	ir := build(`func main() { var x = 10 + 20 }`)
	dump := ir.Dump()
	if !strings.Contains(dump, "ADD") {
		t.Error("expected ADD in IR")
	}
}

func TestIfElse(t *testing.T) {
	ir := build(`
func main() {
	var age = 18
	if age >= 18 {
		print("adult")
	} else {
		print("minor")
	}
}`)
	dump := ir.Dump()
	if !strings.Contains(dump, "if !") {
		t.Error("expected conditional jump in IR")
	}
	if !strings.Contains(dump, "goto") {
		t.Error("expected goto in IR")
	}
}

func TestForLoop(t *testing.T) {
	ir := build(`
func main() {
	for i = 0; i < 10; i = i + 1 {
		print(i)
	}
}`)
	dump := ir.Dump()
	if !strings.Contains(dump, "L0:") {
		t.Error("expected loop label in IR")
	}
	if !strings.Contains(dump, "goto L0") {
		t.Error("expected back jump in IR")
	}
}

func TestFuncCall(t *testing.T) {
	ir := build(`
func add(a int, b int) int {
	return a + b
}
func main() {
	var result = add(1, 2)
	print(result)
}`)
	dump := ir.Dump()
	if !strings.Contains(dump, "func add:") {
		t.Error("expected func add in IR")
	}
	if !strings.Contains(dump, "return") {
		t.Error("expected return in IR")
	}
	if !strings.Contains(dump, "call add") {
		t.Error("expected call add in IR")
	}
}