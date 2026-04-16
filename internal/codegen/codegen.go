package codegen

import (
	"fmt"
	"strings"

	"github.com/devguinness/Riner/internal/ir"
)

// Generator takes an IR program and emits C source code
type Generator struct {
	buf     strings.Builder
	indent  int
	inFunc  bool
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(prog *ir.Program) string {
	g.buf.Reset()

	g.line(`#include "runtime.h"`)
	g.line("")

	// forward declare all functions
	funcs := g.collectFuncs(prog)
	for _, name := range funcs {
		if name == "main" {
			g.line(`int main(void);`)
		} else {
			g.line(fmt.Sprintf(`riner_any %s();`, name))
		}
	}
	g.line("")

	// generate code
	for _, instr := range prog.Instrs {
		g.emitInstr(instr)
	}

	// entry point
	g.line("")

	return g.buf.String()
}

func (g *Generator) collectFuncs(prog *ir.Program) []string {
	var funcs []string
	seen := map[string]bool{}
	for _, instr := range prog.Instrs {
		if instr.Op == ir.OpFunc {
			if !seen[instr.Extra] {
				funcs = append(funcs, instr.Extra)
				seen[instr.Extra] = true
			}
		}
	}
	return funcs
}

func (g *Generator) emitInstr(instr ir.Instr) {
	switch instr.Op {

	case ir.OpFunc:
		if instr.Extra == "main" {
			g.line(`int main(void) {`)
		} else {
			g.line(fmt.Sprintf(`riner_any %s() {`, instr.Extra))
		}
		g.indent++

	case ir.OpEnd:
		g.indent--
		if instr.Extra == "main" {
			g.iline(`return 0;`)
		} else {
			g.iline(`return RINER_NIL;`)
		}
		g.line(`}`)
		g.line(``)

	case ir.OpConst:
		g.iline(fmt.Sprintf(`riner_any %s = %s;`, g.cvar(instr.Dst), g.cconst(instr.Extra)))

	case ir.OpAssign:
		g.iline(fmt.Sprintf(`riner_any %s = %s;`, g.cvar(instr.Dst), g.cvar(instr.Src1)))

	case ir.OpAdd:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)((riner_int)%s + (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpSub:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)((riner_int)%s - (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpMul:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)((riner_int)%s * (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpDiv:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)((riner_int)%s / (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpMod:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)((riner_int)%s %% (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpEq:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)(%s == %s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpNeq:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)(%s != %s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpLt:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)((riner_int)%s < (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpGt:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)((riner_int)%s > (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpLte:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)((riner_int)%s <= (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpGte:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)((riner_int)%s >= (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpAnd:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)((riner_int)%s && (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpOr:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)((riner_int)%s || (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpNeg:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)(-(riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1)))

	case ir.OpNot:
		g.iline(fmt.Sprintf(`riner_any %s = (riner_any)(riner_int)(!(riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1)))

	case ir.OpLabel:
		g.line(fmt.Sprintf(`%s:;`, instr.Extra))

	case ir.OpJump:
		g.iline(fmt.Sprintf(`goto %s;`, instr.Extra))

	case ir.OpJumpFalse:
		g.iline(fmt.Sprintf(`if (!(riner_int)%s) goto %s;`, g.cvar(instr.Src1), instr.Extra))

	case ir.OpArg:
		// args are handled inline in OpCall via a simple stack
		// we store them as local vars with __arg_ prefix
		g.iline(fmt.Sprintf(`riner_any __arg_%d = %s;`, g.argCount(), g.cvar(instr.Src1)))

	case ir.OpCall:
		g.iline(fmt.Sprintf(`riner_any %s = %s;`, g.cvar(instr.Dst), g.cCallExpr(instr)))

	case ir.OpReturn:
		if instr.Src1 == "" {
			g.iline(`return RINER_NIL;`)
		} else {
			g.iline(fmt.Sprintf(`return (riner_any)%s;`, g.cvar(instr.Src1)))
		}

	case ir.OpLoad:
		g.iline(fmt.Sprintf(`riner_any %s = riner_array_get((riner_array*)%s, (riner_int)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))

	case ir.OpStore:
		g.iline(fmt.Sprintf(`riner_array_set((riner_array*)%s, (riner_int)%s, (riner_any)%s);`,
			g.cvar(instr.Dst), g.cvar(instr.Src1), g.cvar(instr.Src2)))
	}
}

var _argCounter int

func (g *Generator) argCount() int {
	n := _argCounter
	_argCounter++
	return n
}

func (g *Generator) cCallExpr(instr ir.Instr) string {
	name := instr.Extra
	switch name {
	case "print", "println":
		n := _argCounter - 1
		arg := fmt.Sprintf("__arg_%d", n)
		_argCounter = 0
		return fmt.Sprintf("(riner_print((riner_string)%s), RINER_NIL)", arg)
	case "len":
		n := _argCounter - 1
		arg := fmt.Sprintf("__arg_%d", n)
		_argCounter = 0
		return fmt.Sprintf("(riner_any)(riner_int)riner_array_len((riner_array*)%s)", arg)
	default:
		_argCounter = 0
		return fmt.Sprintf("%s()", name)
	}
}

func (g *Generator) cvar(name string) string {
	if name == "" {
		return "RINER_NIL"
	}
	// sanitize: replace dots and brackets
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "[", "_")
	name = strings.ReplaceAll(name, "]", "_")
	return "v_" + name
}

func (g *Generator) cconst(val string) string {
	if val == "true" {
		return "(riner_any)(riner_int)RINER_TRUE"
	}
	if val == "false" {
		return "(riner_any)(riner_int)RINER_FALSE"
	}
	if val == "nil" || val == "[]" || val == "{}" {
		return "RINER_NIL"
	}
	if strings.HasPrefix(val, "\"") {
		return fmt.Sprintf("(riner_any)%s", val)
	}
	// number
	return fmt.Sprintf("(riner_any)(riner_int)%s", val)
}

func (g *Generator) line(s string) {
	g.buf.WriteString(s + "\n")
}

func (g *Generator) iline(s string) {
	g.buf.WriteString(strings.Repeat("    ", g.indent) + s + "\n")
}