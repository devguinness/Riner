package ir

import (
	"fmt"
	"strconv"

	"github.com/devguinness/Riner/internal/parser"
)

// Builder lowers a Riner AST into IR instructions
type Builder struct {
	prog    *Program
	tempCount  int
	labelCount int
}

func NewBuilder() *Builder {
	return &Builder{prog: &Program{}}
}

func (b *Builder) Build(program *parser.Program) *Program {
	for _, stmt := range program.Statements {
		b.lowerStmt(stmt)
	}
	return b.prog
}

func (b *Builder) newTemp() string {
	t := fmt.Sprintf("t%d", b.tempCount)
	b.tempCount++
	return t
}

func (b *Builder) newLabel() string {
	l := fmt.Sprintf("L%d", b.labelCount)
	b.labelCount++
	return l
}

func (b *Builder) emit(i Instr) {
	b.prog.Emit(i)
}

// --- statements ---

func (b *Builder) lowerStmt(node parser.Node) {
	switch s := node.(type) {

	case *parser.FuncDecl:
		b.emit(Instr{Op: OpFunc, Extra: s.Name})
		for _, stmt := range s.Body.Statements {
			b.lowerStmt(stmt)
		}
		b.emit(Instr{Op: OpEnd, Extra: s.Name})

	case *parser.VarDecl:
		src := b.lowerExpr(s.Value)
		b.emit(Instr{Op: OpAssign, Dst: s.Name, Src1: src})

	case *parser.AssignStmt:
		src := b.lowerExpr(s.Value)
		b.emit(Instr{Op: OpAssign, Dst: s.Name, Src1: src})

	case *parser.ReturnStmt:
		src := b.lowerExpr(s.Value)
		b.emit(Instr{Op: OpReturn, Src1: src})

	case *parser.ExprStmt:
		b.lowerExpr(s.Expr)

	case *parser.IfStmt:
		b.lowerIf(s)

	case *parser.ForStmt:
		b.lowerFor(s)

	case *parser.IndexAssign:
		obj := b.lowerExpr(s.Object)
		idx := b.lowerExpr(s.Index)
		val := b.lowerExpr(s.Value)
		b.emit(Instr{Op: OpStore, Dst: obj, Src1: idx, Src2: val})

	case *parser.MultiVarDecl:
		src := b.lowerExpr(s.Value)
		for i, name := range s.Names {
			tmp := b.newTemp()
			b.emit(Instr{Op: OpLoad, Dst: tmp, Src1: src, Src2: strconv.Itoa(i)})
			b.emit(Instr{Op: OpAssign, Dst: name, Src1: tmp})
		}

	case *parser.StructDecl:
		// structs are resolved at type check time, no IR needed

	case *parser.ImportStmt:
		// imports handled at link time
	}
}

func (b *Builder) lowerIf(s *parser.IfStmt) {
	cond := b.lowerExpr(s.Condition)
	elseLabel := b.newLabel()
	endLabel := b.newLabel()

	b.emit(Instr{Op: OpJumpFalse, Src1: cond, Extra: elseLabel})

	for _, stmt := range s.Then.Statements {
		b.lowerStmt(stmt)
	}

	if s.Else != nil {
		b.emit(Instr{Op: OpJump, Extra: endLabel})
		b.emit(Instr{Op: OpLabel, Extra: elseLabel})
		for _, stmt := range s.Else.Statements {
			b.lowerStmt(stmt)
		}
		b.emit(Instr{Op: OpLabel, Extra: endLabel})
	} else {
		b.emit(Instr{Op: OpLabel, Extra: elseLabel})
	}
}

func (b *Builder) lowerFor(s *parser.ForStmt) {
	startLabel := b.newLabel()
	endLabel := b.newLabel()

	if s.Init != nil {
		b.lowerStmt(s.Init)
	}

	b.emit(Instr{Op: OpLabel, Extra: startLabel})

	cond := b.lowerExpr(s.Condition)
	b.emit(Instr{Op: OpJumpFalse, Src1: cond, Extra: endLabel})

	for _, stmt := range s.Body.Statements {
		b.lowerStmt(stmt)
	}

	if s.Post != nil {
		b.lowerStmt(s.Post)
	}

	b.emit(Instr{Op: OpJump, Extra: startLabel})
	b.emit(Instr{Op: OpLabel, Extra: endLabel})
}

// --- expressions ---

func (b *Builder) lowerExpr(node parser.Node) string {
	switch e := node.(type) {

	case *parser.IntLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: e.Value})
		return tmp

	case *parser.FloatLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: e.Value})
		return tmp

	case *parser.StringLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: fmt.Sprintf("%q", e.Value)})
		return tmp

	case *parser.BoolLiteral:
		tmp := b.newTemp()
		val := "false"
		if e.Value {
			val = "true"
		}
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: val})
		return tmp

	case *parser.NilLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: "nil"})
		return tmp

	case *parser.Identifier:
		return e.Name

	case *parser.BinaryExpr:
		return b.lowerBinary(e)

	case *parser.UnaryExpr:
		src := b.lowerExpr(e.Right)
		tmp := b.newTemp()
		if e.Operator == "-" {
			b.emit(Instr{Op: OpNeg, Dst: tmp, Src1: src})
		} else {
			b.emit(Instr{Op: OpNot, Dst: tmp, Src1: src})
		}
		return tmp

	case *parser.CallExpr:
		return b.lowerCall(e)

	case *parser.FieldAccess:
		obj := b.lowerExpr(e.Object)
		tmp := b.newTemp()
		b.emit(Instr{Op: OpLoad, Dst: tmp, Src1: obj, Src2: e.Field})
		return tmp

	case *parser.IndexExpr:
		obj := b.lowerExpr(e.Object)
		idx := b.lowerExpr(e.Index)
		tmp := b.newTemp()
		b.emit(Instr{Op: OpLoad, Dst: tmp, Src1: obj, Src2: idx})
		return tmp

	case *parser.ArrayLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: "[]"})
		for i, el := range e.Elements {
			val := b.lowerExpr(el)
			b.emit(Instr{Op: OpStore, Dst: tmp, Src1: strconv.Itoa(i), Src2: val})
		}
		return tmp

	case *parser.MapLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: "{}"})
		for _, pair := range e.Pairs {
			key := b.lowerExpr(pair.Key)
			val := b.lowerExpr(pair.Value)
			b.emit(Instr{Op: OpStore, Dst: tmp, Src1: key, Src2: val})
		}
		return tmp

	case *parser.StructLiteral:
		tmp := b.newTemp()
		b.emit(Instr{Op: OpConst, Dst: tmp, Extra: fmt.Sprintf("struct:%s", e.Name)})
		for _, f := range e.Fields {
			val := b.lowerExpr(f.Value)
			b.emit(Instr{Op: OpStore, Dst: tmp, Src1: f.Name, Src2: val})
		}
		return tmp
	}

	return "?"
}

func (b *Builder) lowerBinary(e *parser.BinaryExpr) string {
	left := b.lowerExpr(e.Left)
	right := b.lowerExpr(e.Right)
	tmp := b.newTemp()

	opMap := map[string]OpCode{
		"+":  OpAdd,
		"-":  OpSub,
		"*":  OpMul,
		"/":  OpDiv,
		"%":  OpMod,
		"==": OpEq,
		"!=": OpNeq,
		"<":  OpLt,
		">":  OpGt,
		"<=": OpLte,
		">=": OpGte,
		"&&": OpAnd,
		"||": OpOr,
	}

	op, ok := opMap[e.Operator]
	if !ok {
		op = OpAdd
	}

	b.emit(Instr{Op: op, Dst: tmp, Src1: left, Src2: right})
	return tmp
}

func (b *Builder) lowerCall(e *parser.CallExpr) string {
	for _, arg := range e.Args {
		src := b.lowerExpr(arg)
		b.emit(Instr{Op: OpArg, Src1: src})
	}

	tmp := b.newTemp()
	b.emit(Instr{Op: OpCall, Dst: tmp, Extra: e.Function})
	return tmp
}