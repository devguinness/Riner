package ir

import "fmt"

// OpCode represents an IR instruction type
type OpCode string

const (
	// Variables
	OpAssign  OpCode = "ASSIGN"  // dst = src
	OpConst   OpCode = "CONST"   // dst = literal value

	// Arithmetic
	OpAdd OpCode = "ADD" // dst = left + right
	OpSub OpCode = "SUB" // dst = left - right
	OpMul OpCode = "MUL" // dst = left * right
	OpDiv OpCode = "DIV" // dst = left / right
	OpMod OpCode = "MOD" // dst = left % right
	OpNeg OpCode = "NEG" // dst = -src

	// Comparison
	OpEq  OpCode = "EQ"  // dst = left == right
	OpNeq OpCode = "NEQ" // dst = left != right
	OpLt  OpCode = "LT"  // dst = left < right
	OpGt  OpCode = "GT"  // dst = left > right
	OpLte OpCode = "LTE" // dst = left <= right
	OpGte OpCode = "GTE" // dst = left >= right

	// Logic
	OpAnd OpCode = "AND" // dst = left && right
	OpOr  OpCode = "OR"  // dst = left || right
	OpNot OpCode = "NOT" // dst = !src

	// Control flow
	OpLabel    OpCode = "LABEL"    // label:
	OpJump     OpCode = "JUMP"     // goto label
	OpJumpFalse OpCode = "JUMPF"   // if !cond goto label

	// Functions
	OpCall   OpCode = "CALL"   // dst = call func(args...)
	OpReturn OpCode = "RETURN" // return src
	OpArg    OpCode = "ARG"    // push arg before CALL
	OpFunc   OpCode = "FUNC"   // function definition start
	OpEnd    OpCode = "END"    // function definition end

	// Memory
	OpLoad  OpCode = "LOAD"  // dst = array/map[index]
	OpStore OpCode = "STORE" // array/map[index] = src

	// Built-ins
	OpPrint OpCode = "PRINT" // print(args...)
)

// Instr is a single IR instruction
type Instr struct {
	Op    OpCode
	Dst   string // destination temp or variable
	Src1  string // first operand
	Src2  string // second operand (optional)
	Extra string // extra info (label name, function name, literal value)
}

func (i Instr) String() string {
	switch i.Op {
	case OpConst:
		return fmt.Sprintf("%s = %s", i.Dst, i.Extra)
	case OpAssign:
		return fmt.Sprintf("%s = %s", i.Dst, i.Src1)
	case OpAdd, OpSub, OpMul, OpDiv, OpMod:
		return fmt.Sprintf("%s = %s %s %s", i.Dst, i.Src1, i.Op, i.Src2)
	case OpEq, OpNeq, OpLt, OpGt, OpLte, OpGte, OpAnd, OpOr:
		return fmt.Sprintf("%s = %s %s %s", i.Dst, i.Src1, i.Op, i.Src2)
	case OpNeg, OpNot:
		return fmt.Sprintf("%s = %s %s", i.Dst, i.Op, i.Src1)
	case OpLabel:
		return fmt.Sprintf("%s:", i.Extra)
	case OpJump:
		return fmt.Sprintf("goto %s", i.Extra)
	case OpJumpFalse:
		return fmt.Sprintf("if !%s goto %s", i.Src1, i.Extra)
	case OpCall:
		return fmt.Sprintf("%s = call %s", i.Dst, i.Extra)
	case OpReturn:
		return fmt.Sprintf("return %s", i.Src1)
	case OpArg:
		return fmt.Sprintf("arg %s", i.Src1)
	case OpFunc:
		return fmt.Sprintf("func %s:", i.Extra)
	case OpEnd:
		return fmt.Sprintf("end %s", i.Extra)
	case OpPrint:
		return fmt.Sprintf("print %s", i.Src1)
	case OpLoad:
		return fmt.Sprintf("%s = %s[%s]", i.Dst, i.Src1, i.Src2)
	case OpStore:
		return fmt.Sprintf("%s[%s] = %s", i.Dst, i.Src1, i.Src2)
	}
	return fmt.Sprintf("%s %s %s %s", i.Op, i.Dst, i.Src1, i.Src2)
}

// Program is the full IR output
type Program struct {
	Instrs []Instr
}

func (p *Program) Emit(i Instr) {
	p.Instrs = append(p.Instrs, i)
}

func (p *Program) Dump() string {
	out := ""
	for _, i := range p.Instrs {
		out += i.String() + "\n"
	}
	return out
}