package interpreter

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/devguinness/Riner/internal/parser"
)

// errReturn is used to unwind the call stack on return
type errReturn struct {
	value *Value
}

func (e *errReturn) Error() string { return "__return__" }

// Value represents any runtime value in Riner
type Value struct {
	kind    string // "int", "float", "string", "bool", "nil", "struct", "func", "array", "map"
	intVal  int64
	fltVal  float64
	strVal  string
	boolVal bool
	fields  map[string]*Value  // for structs
	elems   []*Value           // for arrays
	pairs   map[string]*Value  // for maps
	fn      *parser.FuncDecl   // for functions
}

var Nil = &Value{kind: "nil"}

func intVal(n int64) *Value    { return &Value{kind: "int", intVal: n} }
func fltVal(f float64) *Value  { return &Value{kind: "float", fltVal: f} }
func strVal(s string) *Value   { return &Value{kind: "string", strVal: s} }
func boolVal(b bool) *Value    { return &Value{kind: "bool", boolVal: b} }

func (v *Value) String() string {
	switch v.kind {
	case "int":
		return strconv.FormatInt(v.intVal, 10)
	case "float":
		s := strconv.FormatFloat(v.fltVal, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	case "string":
		return v.strVal
	case "bool":
		if v.boolVal {
			return "true"
		}
		return "false"
	case "nil":
		return "nil"
	case "struct":
		return "<struct>"
	case "func":
		return fmt.Sprintf("<func %s>", v.fn.Name)
	case "array":
		parts := make([]string, len(v.elems))
		for i, e := range v.elems {
			parts[i] = e.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case "map":
		return "<map>"
	}
	return "?"
}

// Environment holds variables in the current scope
type Environment struct {
	vars   map[string]*Value
	parent *Environment
}

func newEnv(parent *Environment) *Environment {
	return &Environment{vars: make(map[string]*Value), parent: parent}
}

func (e *Environment) get(name string) (*Value, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.get(name)
	}
	return nil, false
}

func (e *Environment) set(name string, val *Value) {
	e.vars[name] = val
}

func (e *Environment) assign(name string, val *Value) bool {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = val
		return true
	}
	if e.parent != nil {
		return e.parent.assign(name, val)
	}
	return false
}

// Interpreter runs a Riner AST
type Interpreter struct {
	globals *Environment
	structs map[string]*parser.StructDecl
}

func New() *Interpreter {
	env := newEnv(nil)
	return &Interpreter{
		globals: env,
		structs: make(map[string]*parser.StructDecl),
	}
}

func (i *Interpreter) Run(program *parser.Program) {
	// first pass: register all top-level funcs and structs
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *parser.FuncDecl:
			i.globals.set(s.Name, &Value{kind: "func", fn: s})
		case *parser.StructDecl:
			i.structs[s.Name] = s
		}
	}

	// call main
	mainVal, ok := i.globals.get("main")
	if !ok {
		i.runtimeError(0, 0, "no main function found")
	}
	_, err := i.callFunc(mainVal.fn, []*Value{}, i.globals)
	if err != nil {
		i.runtimeError(0, 0, err.Error())
	}
}

func (i *Interpreter) execBlock(block *parser.Block, env *Environment) (*Value, error) {
	for _, stmt := range block.Statements {
		result, err := i.execStmt(stmt, env)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}
	return nil, nil
}

func (i *Interpreter) execStmt(node parser.Node, env *Environment) (*Value, error) {
	switch s := node.(type) {

	case *parser.VarDecl:
		val, err := i.evalExpr(s.Value, env)
		if err != nil {
			return nil, err
		}
		env.set(s.Name, val)
		return nil, nil

	case *parser.AssignStmt:
		val, err := i.evalExpr(s.Value, env)
		if err != nil {
			return nil, err
		}
		if !env.assign(s.Name, val) {
			return nil, i.errorf(s.Line, s.Col, "undefined variable %q", s.Name)
		}
		return nil, nil

	case *parser.ReturnStmt:
		val, err := i.evalExpr(s.Value, env)
		if err != nil {
			return nil, err
		}
		return val, &errReturn{value: val}

	case *parser.IfStmt:
		cond, err := i.evalExpr(s.Condition, env)
		if err != nil {
			return nil, err
		}
		if i.isTruthy(cond) {
			return i.execBlock(s.Then, newEnv(env))
		} else if s.Else != nil {
			return i.execBlock(s.Else, newEnv(env))
		}
		return nil, nil

	case *parser.ForStmt:
		return i.execFor(s, env)

	case *parser.ExprStmt:
		_, err := i.evalExpr(s.Expr, env)
		return nil, err

	case *parser.FuncDecl:
		env.set(s.Name, &Value{kind: "func", fn: s})
		return nil, nil

	case *parser.StructDecl:
		i.structs[s.Name] = s
		return nil, nil

	case *parser.MultiVarDecl:
		val, err := i.evalExpr(s.Value, env)
		if err != nil {
			return nil, err
		}
		// expect array of values for multi-return
		if val.kind == "array" && len(val.elems) == len(s.Names) {
			for idx, name := range s.Names {
				env.set(name, val.elems[idx])
			}
		} else {
			for _, name := range s.Names {
				env.set(name, val)
			}
		}
		return nil, nil

	case *parser.IndexAssign:
		obj, err := i.evalExpr(s.Object, env)
		if err != nil {
			return nil, err
		}
		idx, err := i.evalExpr(s.Index, env)
		if err != nil {
			return nil, err
		}
		val, err := i.evalExpr(s.Value, env)
		if err != nil {
			return nil, err
		}
		if obj.kind == "array" {
			if idx.kind != "int" {
				return nil, i.errorf(s.Line, s.Col, "array index must be int")
			}
			if idx.intVal < 0 || int(idx.intVal) >= len(obj.elems) {
				return nil, i.errorf(s.Line, s.Col, "index out of bounds")
			}
			obj.elems[idx.intVal] = val
		} else if obj.kind == "map" {
			obj.pairs[idx.String()] = val
		} else {
			return nil, i.errorf(s.Line, s.Col, "cannot index %s", obj.kind)
		}
		return nil, nil

	case *parser.ImportStmt:
		// imports are resolved at a later stage
		return nil, nil
	}

	return nil, fmt.Errorf("unknown statement type: %T", node)
}

func (i *Interpreter) execFor(s *parser.ForStmt, env *Environment) (*Value, error) {
	loopEnv := newEnv(env)

	// classic for: init; condition; post
	// init is treated as a declaration so the variable exists in loopEnv
	if s.Init != nil {
		if assign, ok := s.Init.(*parser.AssignStmt); ok {
			val, err := i.evalExpr(assign.Value, loopEnv)
			if err != nil {
				return nil, err
			}
			loopEnv.set(assign.Name, val)
		} else {
			if _, err := i.execStmt(s.Init, loopEnv); err != nil {
				return nil, err
			}
		}
	}

	for {
		cond, err := i.evalExpr(s.Condition, loopEnv)
		if err != nil {
			return nil, err
		}
		if !i.isTruthy(cond) {
			break
		}

		ret, err := i.execBlock(s.Body, newEnv(loopEnv))
		if err != nil {
			return ret, err
		}
		_ = ret

		if s.Post != nil {
			if _, err := i.execStmt(s.Post, loopEnv); err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

func (i *Interpreter) evalExpr(node parser.Node, env *Environment) (*Value, error) {
	switch e := node.(type) {

	case *parser.IntLiteral:
		n, _ := strconv.ParseInt(e.Value, 10, 64)
		return intVal(n), nil

	case *parser.FloatLiteral:
		f, _ := strconv.ParseFloat(e.Value, 64)
		return fltVal(f), nil

	case *parser.StringLiteral:
		return strVal(e.Value), nil

	case *parser.BoolLiteral:
		return boolVal(e.Value), nil

	case *parser.NilLiteral:
		return Nil, nil

	case *parser.Identifier:
		val, ok := env.get(e.Name)
		if !ok {
			return nil, i.errorf(e.Line, e.Col, "undefined variable %q", e.Name)
		}
		return val, nil

	case *parser.BinaryExpr:
		return i.evalBinary(e, env)

	case *parser.UnaryExpr:
		return i.evalUnary(e, env)

	case *parser.CallExpr:
		return i.evalCall(e, env)

	case *parser.FieldAccess:
		return i.evalFieldAccess(e, env)

	case *parser.StructLiteral:
		return i.evalStructLiteral(e, env)

	case *parser.ArrayLiteral:
		return i.evalArrayLiteral(e, env)

	case *parser.MapLiteral:
		return i.evalMapLiteral(e, env)

	case *parser.IndexExpr:
		return i.evalIndexExpr(e, env)
	}

	return nil, fmt.Errorf("unknown expression type: %T", node)
}

func (i *Interpreter) evalArrayLiteral(e *parser.ArrayLiteral, env *Environment) (*Value, error) {
	elems := make([]*Value, len(e.Elements))
	for idx, el := range e.Elements {
		val, err := i.evalExpr(el, env)
		if err != nil {
			return nil, err
		}
		elems[idx] = val
	}
	return &Value{kind: "array", elems: elems}, nil
}

func (i *Interpreter) evalMapLiteral(e *parser.MapLiteral, env *Environment) (*Value, error) {
	pairs := make(map[string]*Value)
	for _, p := range e.Pairs {
		key, err := i.evalExpr(p.Key, env)
		if err != nil {
			return nil, err
		}
		val, err := i.evalExpr(p.Value, env)
		if err != nil {
			return nil, err
		}
		pairs[key.String()] = val
	}
	return &Value{kind: "map", pairs: pairs}, nil
}

func (i *Interpreter) evalIndexExpr(e *parser.IndexExpr, env *Environment) (*Value, error) {
	obj, err := i.evalExpr(e.Object, env)
	if err != nil {
		return nil, err
	}
	idx, err := i.evalExpr(e.Index, env)
	if err != nil {
		return nil, err
	}
	switch obj.kind {
	case "array":
		if idx.kind != "int" {
			return nil, i.errorf(e.Line, e.Col, "array index must be int")
		}
		if idx.intVal < 0 || int(idx.intVal) >= len(obj.elems) {
			return nil, i.errorf(e.Line, e.Col, "index out of bounds")
		}
		return obj.elems[idx.intVal], nil
	case "map":
		val, ok := obj.pairs[idx.String()]
		if !ok {
			return Nil, nil
		}
		return val, nil
	case "string":
		if idx.kind != "int" {
			return nil, i.errorf(e.Line, e.Col, "string index must be int")
		}
		if idx.intVal < 0 || int(idx.intVal) >= len(obj.strVal) {
			return nil, i.errorf(e.Line, e.Col, "index out of bounds")
		}
		return strVal(string(obj.strVal[idx.intVal])), nil
	}
	return nil, i.errorf(e.Line, e.Col, "cannot index %s", obj.kind)
}

func (i *Interpreter) evalBinary(e *parser.BinaryExpr, env *Environment) (*Value, error) {
	left, err := i.evalExpr(e.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := i.evalExpr(e.Right, env)
	if err != nil {
		return nil, err
	}

	switch e.Operator {
	case "+":
		if left.kind == "string" || right.kind == "string" {
			return strVal(left.String() + right.String()), nil
		}
		return i.arith(e, left, right, func(a, b int64) int64 { return a + b }, func(a, b float64) float64 { return a + b })
	case "-":
		return i.arith(e, left, right, func(a, b int64) int64 { return a - b }, func(a, b float64) float64 { return a - b })
	case "*":
		return i.arith(e, left, right, func(a, b int64) int64 { return a * b }, func(a, b float64) float64 { return a * b })
	case "/":
		if (left.kind == "int" && right.intVal == 0) || (left.kind == "float" && right.fltVal == 0) {
			return nil, i.errorf(e.Line, e.Col, "division by zero")
		}
		return i.arith(e, left, right, func(a, b int64) int64 { return a / b }, func(a, b float64) float64 { return a / b })
	case "%":
		if right.intVal == 0 {
			return nil, i.errorf(e.Line, e.Col, "division by zero")
		}
		return intVal(left.intVal % right.intVal), nil
	case "==":
		return boolVal(i.equal(left, right)), nil
	case "!=":
		return boolVal(!i.equal(left, right)), nil
	case "<":
		return boolVal(i.compare(left, right) < 0), nil
	case ">":
		return boolVal(i.compare(left, right) > 0), nil
	case "<=":
		return boolVal(i.compare(left, right) <= 0), nil
	case ">=":
		return boolVal(i.compare(left, right) >= 0), nil
	case "&&":
		return boolVal(i.isTruthy(left) && i.isTruthy(right)), nil
	case "||":
		return boolVal(i.isTruthy(left) || i.isTruthy(right)), nil
	}

	return nil, i.errorf(e.Line, e.Col, "unknown operator %q", e.Operator)
}

func (i *Interpreter) arith(e *parser.BinaryExpr, left, right *Value, iop func(int64, int64) int64, fop func(float64, float64) float64) (*Value, error) {
	if left.kind == "int" && right.kind == "int" {
		return intVal(iop(left.intVal, right.intVal)), nil
	}
	lf, err := i.toFloat(e.Line, e.Col, left)
	if err != nil {
		return nil, err
	}
	rf, err := i.toFloat(e.Line, e.Col, right)
	if err != nil {
		return nil, err
	}
	return fltVal(fop(lf, rf)), nil
}

func (i *Interpreter) toFloat(line, col int, v *Value) (float64, error) {
	switch v.kind {
	case "int":
		return float64(v.intVal), nil
	case "float":
		return v.fltVal, nil
	}
	return 0, i.errorf(line, col, "expected number, got %s", v.kind)
}

func (i *Interpreter) evalUnary(e *parser.UnaryExpr, env *Environment) (*Value, error) {
	right, err := i.evalExpr(e.Right, env)
	if err != nil {
		return nil, err
	}
	switch e.Operator {
	case "!":
		return boolVal(!i.isTruthy(right)), nil
	case "-":
		if right.kind == "int" {
			return intVal(-right.intVal), nil
		}
		if right.kind == "float" {
			return fltVal(-right.fltVal), nil
		}
		return nil, i.errorf(e.Line, e.Col, "cannot negate %s", right.kind)
	}
	return nil, i.errorf(e.Line, e.Col, "unknown unary operator %q", e.Operator)
}

func (i *Interpreter) evalCall(e *parser.CallExpr, env *Environment) (*Value, error) {
	// built-ins
	switch e.Function {
	case "append":
		if len(e.Args) < 2 {
			return nil, i.errorf(e.Line, e.Col, "append expects at least 2 arguments")
		}
		arr, err := i.evalExpr(e.Args[0], env)
		if err != nil {
			return nil, err
		}
		if arr.kind != "array" {
			return nil, i.errorf(e.Line, e.Col, "append expects array as first argument")
		}
		newElems := make([]*Value, len(arr.elems))
		copy(newElems, arr.elems)
		for _, argNode := range e.Args[1:] {
			val, err := i.evalExpr(argNode, env)
			if err != nil {
				return nil, err
			}
			newElems = append(newElems, val)
		}
		return &Value{kind: "array", elems: newElems}, nil

	case "print", "println":
		args, err := i.evalArgs(e.Args, env)
		if err != nil {
			return nil, err
		}
		parts := make([]string, len(args))
		for idx, a := range args {
			parts[idx] = a.String()
		}
		if e.Function == "println" {
			fmt.Println(strings.Join(parts, " "))
		} else {
			fmt.Println(strings.Join(parts, " "))
		}
		return Nil, nil

	case "len":
		if len(e.Args) != 1 {
			return nil, i.errorf(e.Line, e.Col, "len expects 1 argument")
		}
		arg, err := i.evalExpr(e.Args[0], env)
		if err != nil {
			return nil, err
		}
		switch arg.kind {
		case "string":
			return intVal(int64(len(arg.strVal))), nil
		case "array":
			return intVal(int64(len(arg.elems))), nil
		default:
			return nil, i.errorf(e.Line, e.Col, "len expects string or array, got %s", arg.kind)
		}
	}

	// user-defined function
	fnVal, ok := env.get(e.Function)
	if !ok {
		return nil, i.errorf(e.Line, e.Col, "undefined function %q", e.Function)
	}
	if fnVal.kind != "func" {
		return nil, i.errorf(e.Line, e.Col, "%q is not a function", e.Function)
	}

	args, err := i.evalArgs(e.Args, env)
	if err != nil {
		return nil, err
	}

	return i.callFunc(fnVal.fn, args, env)
}

func (i *Interpreter) callFunc(fn *parser.FuncDecl, args []*Value, _ *Environment) (*Value, error) {
	if len(args) != len(fn.Params) {
		return nil, fmt.Errorf("function %q expects %d args, got %d", fn.Name, len(fn.Params), len(args))
	}

	fnEnv := newEnv(i.globals)
	for idx, param := range fn.Params {
		fnEnv.set(param.Name, args[idx])
	}

	ret, err := i.execBlock(fn.Body, fnEnv)
	if r, ok := err.(*errReturn); ok {
		return r.value, nil
	}
	if err != nil {
		return nil, err
	}
	if ret != nil {
		return ret, nil
	}
	return Nil, nil
}

func (i *Interpreter) evalArgs(nodes []parser.Node, env *Environment) ([]*Value, error) {
	args := make([]*Value, len(nodes))
	for idx, n := range nodes {
		v, err := i.evalExpr(n, env)
		if err != nil {
			return nil, err
		}
		args[idx] = v
	}
	return args, nil
}

func (i *Interpreter) evalFieldAccess(e *parser.FieldAccess, env *Environment) (*Value, error) {
	obj, err := i.evalExpr(e.Object, env)
	if err != nil {
		return nil, err
	}
	if obj.kind != "struct" {
		return nil, i.errorf(e.Line, e.Col, "cannot access field on %s", obj.kind)
	}
	val, ok := obj.fields[e.Field]
	if !ok {
		return nil, i.errorf(e.Line, e.Col, "field %q not found", e.Field)
	}
	return val, nil
}

func (i *Interpreter) evalStructLiteral(e *parser.StructLiteral, env *Environment) (*Value, error) {
	decl, ok := i.structs[e.Name]
	if !ok {
		return nil, i.errorf(e.Line, e.Col, "undefined struct %q", e.Name)
	}

	fields := make(map[string]*Value)

	// initialize all fields to nil
	for _, f := range decl.Fields {
		fields[f.Name] = Nil
	}

	// set provided values
	for _, fv := range e.Fields {
		val, err := i.evalExpr(fv.Value, env)
		if err != nil {
			return nil, err
		}
		fields[fv.Name] = val
	}

	return &Value{kind: "struct", fields: fields}, nil
}

func (i *Interpreter) isTruthy(v *Value) bool {
	switch v.kind {
	case "bool":
		return v.boolVal
	case "nil":
		return false
	case "int":
		return v.intVal != 0
	case "float":
		return v.fltVal != 0 && !math.IsNaN(v.fltVal)
	case "string":
		return v.strVal != ""
	}
	return true
}

func (i *Interpreter) equal(a, b *Value) bool {
	if a.kind != b.kind {
		return false
	}
	switch a.kind {
	case "int":
		return a.intVal == b.intVal
	case "float":
		return a.fltVal == b.fltVal
	case "string":
		return a.strVal == b.strVal
	case "bool":
		return a.boolVal == b.boolVal
	case "nil":
		return true
	}
	return false
}

func (i *Interpreter) compare(a, b *Value) int {
	if a.kind == "int" && b.kind == "int" {
		if a.intVal < b.intVal {
			return -1
		}
		if a.intVal > b.intVal {
			return 1
		}
		return 0
	}
	af, _ := i.toFloat(0, 0, a)
	bf, _ := i.toFloat(0, 0, b)
	if af < bf {
		return -1
	}
	if af > bf {
		return 1
	}
	return 0
}

func (i *Interpreter) errorf(line, col int, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("runtime error at line %d, col %d: %s", line, col, msg)
}

func (i *Interpreter) runtimeError(line, col int, msg string) {
	if line > 0 {
		fmt.Fprintf(os.Stderr, "runtime error at line %d, col %d: %s\n", line, col, msg)
	} else {
		fmt.Fprintf(os.Stderr, "runtime error: %s\n", msg)
	}
	os.Exit(1)
}