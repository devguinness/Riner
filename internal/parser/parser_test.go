package parser

import (
	"testing"

	"github.com/devguinness/Riner/internal/lexer"
)

func parse(src string) (*Program, error) {
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := New(tokens)
	return p.Parse()
}

func TestHelloWorld(t *testing.T) {
	src := `func main() {
		print("hello, world")
	}`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*FuncDecl)
	if !ok {
		t.Fatal("expected FuncDecl")
	}
	if fn.Name != "main" {
		t.Errorf("expected main, got %s", fn.Name)
	}
}

func TestVarDecl(t *testing.T) {
	src := `var age = 17`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := prog.Statements[0].(*VarDecl)
	if !ok {
		t.Fatal("expected VarDecl")
	}
	if v.Name != "age" {
		t.Errorf("expected age, got %s", v.Name)
	}
	lit, ok := v.Value.(*IntLiteral)
	if !ok {
		t.Fatal("expected IntLiteral")
	}
	if lit.Value != "17" {
		t.Errorf("expected 17, got %s", lit.Value)
	}
}

func TestVarDeclExplicitType(t *testing.T) {
	src := `var score int = 0`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	v := prog.Statements[0].(*VarDecl)
	if v.Type != "int" {
		t.Errorf("expected int, got %s", v.Type)
	}
}

func TestIfElse(t *testing.T) {
	src := `if age >= 18 {
		print("adult")
	} else {
		print("minor")
	}`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	stmt, ok := prog.Statements[0].(*IfStmt)
	if !ok {
		t.Fatal("expected IfStmt")
	}
	if stmt.Else == nil {
		t.Error("expected else block")
	}
}

func TestForClassic(t *testing.T) {
	src := `for i = 0; i < 10; i = i + 1 {
		print(i)
	}`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := prog.Statements[0].(*ForStmt)
	if !ok {
		t.Fatal("expected ForStmt")
	}
	if f.Init == nil || f.Post == nil {
		t.Error("expected Init and Post in classic for")
	}
}

func TestForBoolean(t *testing.T) {
	src := `for active {
		print("running")
	}`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := prog.Statements[0].(*ForStmt)
	if !ok {
		t.Fatal("expected ForStmt")
	}
	if f.Init != nil || f.Post != nil {
		t.Error("expected no Init or Post in boolean for")
	}
}

func TestStructDecl(t *testing.T) {
	src := `struct Person {
		name string
		age int
	}`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := prog.Statements[0].(*StructDecl)
	if !ok {
		t.Fatal("expected StructDecl")
	}
	if s.Name != "Person" {
		t.Errorf("expected Person, got %s", s.Name)
	}
	if len(s.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(s.Fields))
	}
}

func TestStructLiteral(t *testing.T) {
	src := `var p = Person{ name: "alice", age: 30 }`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	v := prog.Statements[0].(*VarDecl)
	sl, ok := v.Value.(*StructLiteral)
	if !ok {
		t.Fatal("expected StructLiteral")
	}
	if sl.Name != "Person" {
		t.Errorf("expected Person, got %s", sl.Name)
	}
	if len(sl.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(sl.Fields))
	}
}

func TestBinaryExpr(t *testing.T) {
	src := `var x = 10 + 20`
	prog, err := parse(src)
	if err != nil {
		t.Fatal(err)
	}
	v := prog.Statements[0].(*VarDecl)
	b, ok := v.Value.(*BinaryExpr)
	if !ok {
		t.Fatal("expected BinaryExpr")
	}
	if b.Operator != "+" {
		t.Errorf("expected +, got %s", b.Operator)
	}
}