package sema

import (
	"testing"

	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
)

func check(src string) error {
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return err
	}
	c := New()
	return c.Check(prog)
}

func TestValidProgram(t *testing.T) {
	err := check(`
func main() {
	var x = 10
	var y = 20
	print(x + y)
}`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTypeMismatchAssign(t *testing.T) {
	err := check(`
func main() {
	var x int = "hello"
}`)
	if err == nil {
		t.Fatal("expected type error")
	}
}

func TestUndefinedVariable(t *testing.T) {
	err := check(`
func main() {
	print(x)
}`)
	if err == nil {
		t.Fatal("expected error for undefined variable")
	}
}

func TestReturnTypeMismatch(t *testing.T) {
	err := check(`
func add(a int, b int) int {
	return "hello"
}
func main() {
	var x = add(1, 2)
	print(x)
}`)
	if err == nil {
		t.Fatal("expected return type mismatch error")
	}
}

func TestCorrectReturn(t *testing.T) {
	err := check(`
func add(a int, b int) int {
	return a + b
}
func main() {
	var x = add(1, 2)
	print(x)
}`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUndefinedFunction(t *testing.T) {
	err := check(`
func main() {
	foo()
}`)
	if err == nil {
		t.Fatal("expected error for undefined function")
	}
}

func TestArrayLiteral(t *testing.T) {
	err := check(`
func main() {
	var nums = [1, 2, 3]
	print(len(nums))
}`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStringConcat(t *testing.T) {
	err := check(`
func main() {
	var s = "hello" + " world"
	print(s)
}`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIfConditionNotBool(t *testing.T) {
	err := check(`
func main() {
	if 42 {
		print("ok")
	}
}`)
	if err == nil {
		t.Fatal("expected error for non-bool if condition")
	}
}

func TestVariableTypeFixed(t *testing.T) {
	err := check(`
func main() {
	var x = 10
	x = "hello"
}`)
	if err == nil {
		t.Fatal("expected error for type change after declaration")
	}
}