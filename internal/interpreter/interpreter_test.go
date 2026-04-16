package interpreter

import (
	"testing"

	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
)

func run(src string) {
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		panic(err)
	}
	i := New()
	i.Run(prog)
}

func eval(src string) (*Value, error) {
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return nil, err
	}
	i := New()

	// register funcs and structs
	for _, stmt := range prog.Statements {
		switch s := stmt.(type) {
		case *parser.FuncDecl:
			i.globals.set(s.Name, &Value{kind: "func", fn: s})
		case *parser.StructDecl:
			i.structs[s.Name] = s
		}
	}

	// eval last statement as expression if it's a var decl
	last := prog.Statements[len(prog.Statements)-1]
	if v, ok := last.(*parser.VarDecl); ok {
		val, err := i.evalExpr(v.Value, i.globals)
		return val, err
	}
	return nil, nil
}

func TestHelloWorld(t *testing.T) {
	run(`func main() { print("hello, world") }`)
}

func TestArithmetic(t *testing.T) {
	v, err := eval(`var x = 10 + 20`)
	if err != nil {
		t.Fatal(err)
	}
	if v.intVal != 30 {
		t.Errorf("expected 30, got %d", v.intVal)
	}
}

func TestStringConcat(t *testing.T) {
	v, err := eval(`var s = "hello" + ", " + "world"`)
	if err != nil {
		t.Fatal(err)
	}
	if v.strVal != "hello, world" {
		t.Errorf("expected 'hello, world', got %q", v.strVal)
	}
}

func TestBoolLogic(t *testing.T) {
	v, err := eval(`var x = true && false`)
	if err != nil {
		t.Fatal(err)
	}
	if v.boolVal != false {
		t.Error("expected false")
	}
}

func TestIfElse(t *testing.T) {
	run(`
func main() {
	var age = 18
	if age >= 18 {
		print("adult")
	} else {
		print("minor")
	}
}`)
}

func TestForClassic(t *testing.T) {
	run(`
func main() {
	for i = 0; i < 3; i = i + 1 {
		print(i)
	}
}`)
}

func TestForBoolean(t *testing.T) {
	run(`
func main() {
	var n = 0
	for n < 3 {
		print(n)
		n = n + 1
	}
}`)
}

func TestFuncCall(t *testing.T) {
	run(`
func add(a int, b int) int {
	return a + b
}
func main() {
	var result = add(10, 20)
	print(result)
}`)
}

func TestStruct(t *testing.T) {
	run(`
struct Person {
	name string
	age int
}
func main() {
	var p = Person{ name: "alice", age: 30 }
	print(p.name)
}`)
}

func TestDivisionByZero(t *testing.T) {
	l := lexer.New(`var x = 10 / 0`)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, _ := p.Parse()
	i := New()

	last := prog.Statements[0].(*parser.VarDecl)
	_, err := i.evalExpr(last.Value, i.globals)
	if err == nil {
		t.Error("expected division by zero error")
	}
}