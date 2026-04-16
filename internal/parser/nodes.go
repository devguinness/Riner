package parser

// Node is the base interface for all AST nodes.
type Node interface {
	nodeType() string
}

// Statements

type Program struct {
	Statements []Node
}

func (p *Program) nodeType() string { return "Program" }

type VarDecl struct {
	Name  string
	Type  string // empty if inferred
	Value Node
	Line  int
	Col   int
}

func (v *VarDecl) nodeType() string { return "VarDecl" }

type FuncDecl struct {
	Name       string
	Params     []Param
	ReturnType string // empty if omitted
	Body       *Block
	Line       int
	Col        int
}

func (f *FuncDecl) nodeType() string { return "FuncDecl" }

type Param struct {
	Name string
	Type string
}

type StructDecl struct {
	Name   string
	Fields []Param
	Line   int
	Col    int
}

func (s *StructDecl) nodeType() string { return "StructDecl" }

type Block struct {
	Statements []Node
}

func (b *Block) nodeType() string { return "Block" }

type ReturnStmt struct {
	Value Node
	Line  int
	Col   int
}

func (r *ReturnStmt) nodeType() string { return "ReturnStmt" }

type IfStmt struct {
	Condition Node
	Then      *Block
	Else      *Block // nil if no else
	Line      int
	Col       int
}

func (i *IfStmt) nodeType() string { return "IfStmt" }

type ForStmt struct {
	// Classic for: Init, Condition, Post
	// Boolean for: only Condition is set
	Init      Node // VarDecl or AssignStmt, nil for boolean for
	Condition Node
	Post      Node // AssignStmt, nil for boolean for
	Body      *Block
	Line      int
	Col       int
}

func (f *ForStmt) nodeType() string { return "ForStmt" }

type AssignStmt struct {
	Name  string
	Value Node
	Line  int
	Col   int
}

func (a *AssignStmt) nodeType() string { return "AssignStmt" }

type ExprStmt struct {
	Expr Node
}

func (e *ExprStmt) nodeType() string { return "ExprStmt" }

// Expressions

type Identifier struct {
	Name string
	Line int
	Col  int
}

func (i *Identifier) nodeType() string { return "Identifier" }

type IntLiteral struct {
	Value string
	Line  int
	Col   int
}

func (i *IntLiteral) nodeType() string { return "IntLiteral" }

type FloatLiteral struct {
	Value string
	Line  int
	Col   int
}

func (f *FloatLiteral) nodeType() string { return "FloatLiteral" }

type StringLiteral struct {
	Value string
	Line  int
	Col   int
}

func (s *StringLiteral) nodeType() string { return "StringLiteral" }

type BoolLiteral struct {
	Value bool
	Line  int
	Col   int
}

func (b *BoolLiteral) nodeType() string { return "BoolLiteral" }

type NilLiteral struct {
	Line int
	Col  int
}

func (n *NilLiteral) nodeType() string { return "NilLiteral" }

type BinaryExpr struct {
	Left     Node
	Operator string
	Right    Node
	Line     int
	Col      int
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }

type UnaryExpr struct {
	Operator string
	Right    Node
	Line     int
	Col      int
}

func (u *UnaryExpr) nodeType() string { return "UnaryExpr" }

type CallExpr struct {
	Function string
	Args     []Node
	Line     int
	Col      int
}

func (c *CallExpr) nodeType() string { return "CallExpr" }

type FieldAccess struct {
	Object Node
	Field  string
	Line   int
	Col    int
}

func (f *FieldAccess) nodeType() string { return "FieldAccess" }

type StructLiteral struct {
	Name   string
	Fields []FieldValue
	Line   int
	Col    int
}

func (s *StructLiteral) nodeType() string { return "StructLiteral" }

type FieldValue struct {
	Name  string
	Value Node
}