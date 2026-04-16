package codegen

import (
	"strings"
	"testing"

	"github.com/devguinness/Riner/internal/ir"
	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
)

func generate(src string) string {
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		panic(err)
	}
	b := ir.NewBuilder()
	irprog := b.Build(prog)
	g := New()
	return g.Generate(irprog)
}

func TestIncludesRuntime(t *testing.T) {
	c := generate(`func main() { print("hello") }`)
	if !strings.Contains(c, `#include "runtime.h"`) {
		t.Error("expected runtime.h include")
	}
}

func TestMainFunction(t *testing.T) {
	c := generate(`func main() { print("hello") }`)
	if !strings.Contains(c, `int main(void)`) {
		t.Error("expected int main(void)")
	}
	if !strings.Contains(c, `return 0;`) {
		t.Error("expected return 0")
	}
}

func TestUserFunction(t *testing.T) {
	c := generate(`
func add(a int, b int) int {
	return a + b
}
func main() {
	print("ok")
}`)
	if !strings.Contains(c, `riner_any add()`) {
		t.Error("expected riner_any add()")
	}
}

func TestIfElse(t *testing.T) {
	c := generate(`
func main() {
	var x = 1
	if x >= 1 {
		print("yes")
	} else {
		print("no")
	}
}`)
	if !strings.Contains(c, "goto") {
		t.Error("expected goto in C output")
	}
	if !strings.Contains(c, "if (!") {
		t.Error("expected conditional jump in C output")
	}
}

func TestForLoop(t *testing.T) {
	c := generate(`
func main() {
	for i = 0; i < 3; i = i + 1 {
		print("x")
	}
}`)
	if !strings.Contains(c, "goto") {
		t.Error("expected goto for loop in C output")
	}
}

func TestArithmetic(t *testing.T) {
	c := generate(`func main() { var x = 10 + 20 }`)
	if !strings.Contains(c, "+ (riner_int)") {
		t.Error("expected arithmetic in C output")
	}
}