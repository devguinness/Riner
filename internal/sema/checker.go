package sema

import (
	"fmt"

	"github.com/devguinness/Riner/internal/parser"
)

// Type represents a Riner type
type Type struct {
	Name string
}

var (
	TypeInt    = &Type{"int"}
	TypeFloat  = &Type{"float"}
	TypeString = &Type{"string"}
	TypeBool   = &Type{"bool"}
	TypeNil    = &Type{"nil"}
	TypeVoid   = &Type{"void"}
	TypeArray  = &Type{"array"}
	TypeMap    = &Type{"map"}
	TypeUnknown = &Type{"unknown"}
)

func (t *Type) String() string { return t.Name }

func (t *Type) Equal(other *Type) bool {
	if t == TypeNil || other == TypeNil {
		return true
	}
	return t.Name == other.Name
}

// Scope holds variable types in the current scope
type Scope struct {
	vars   map[string]*Type
	funcs  map[string]*FuncType
	parent *Scope
}

type FuncType struct {
	Params     []*Type
	ReturnType *Type
}

func newScope(parent *Scope) *Scope {
	return &Scope{
		vars:  make(map[string]*Type),
		funcs: make(map[string]*FuncType),
		parent: parent,
	}
}

func (s *Scope) getVar(name string) (*Type, bool) {
	if t, ok := s.vars[name]; ok {
		return t, true
	}
	if s.parent != nil {
		return s.parent.getVar(name)
	}
	return nil, false
}

func (s *Scope) setVar(name string, t *Type) {
	s.vars[name] = t
}

func (s *Scope) getFunc(name string) (*FuncType, bool) {
	if f, ok := s.funcs[name]; ok {
		return f, true
	}
	if s.parent != nil {
		return s.parent.getFunc(name)
	}
	return nil, false
}

func (s *Scope) setFunc(name string, f *FuncType) {
	s.funcs[name] = f
}

// Checker is the type checker
type Checker struct {
	globals    *Scope
	structs    map[string]*parser.StructDecl
	currentReturn *Type
}

func New() *Checker {
	globals := newScope(nil)

	// built-in functions
	globals.setFunc("print",   &FuncType{Params: nil, ReturnType: TypeVoid})
	globals.setFunc("println", &FuncType{Params: nil, ReturnType: TypeVoid})
	globals.setFunc("len",     &FuncType{Params: nil, ReturnType: TypeInt})
	globals.setFunc("append",  &FuncType{Params: nil, ReturnType: TypeArray})

	return &Checker{
		globals: globals,
		structs: make(map[string]*parser.StructDecl),
	}
}

func (c *Checker) Check(program *parser.Program) error {
	// first pass: register all top-level funcs and structs
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *parser.FuncDecl:
			ft, err := c.buildFuncType(s)
			if err != nil {
				return err
			}
			c.globals.setFunc(s.Name, ft)
		case *parser.StructDecl:
			c.structs[s.Name] = s
		}
	}

	// second pass: check all statements
	for _, stmt := range program.Statements {
		if err := c.checkStmt(stmt, c.globals); err != nil {
			return err
		}
	}

	return nil
}

func (c *Checker) buildFuncType(s *parser.FuncDecl) (*FuncType, error) {
	params := make([]*Type, len(s.Params))
	for i, p := range s.Params {
		t, err := c.resolveTypeName(p.Type, 0, 0)
		if err != nil {
			return nil, err
		}
		params[i] = t
	}

	retType := TypeVoid
	if s.ReturnType != "" {
		t, err := c.resolveTypeName(s.ReturnType, 0, 0)
		if err != nil {
			return nil, err
		}
		retType = t
	}

	return &FuncType{Params: params, ReturnType: retType}, nil
}

func (c *Checker) resolveTypeName(name string, line, col int) (*Type, error) {
	switch name {
	case "int":
		return TypeInt, nil
	case "float":
		return TypeFloat, nil
	case "string":
		return TypeString, nil
	case "bool":
		return TypeBool, nil
	case "void":
		return TypeVoid, nil
	case "":
		return TypeVoid, nil
	}
	return nil, c.errorf(line, col, "unknown type %q", name)
}

func (c *Checker) checkStmt(node parser.Node, scope *Scope) error {
	switch s := node.(type) {

	case *parser.VarDecl:
		valType, err := c.checkExpr(s.Value, scope)
		if err != nil {
			return err
		}
		if s.Type != "" {
			declared, err := c.resolveTypeName(s.Type, s.Line, s.Col)
			if err != nil {
				return err
			}
			if !declared.Equal(valType) {
				return c.errorf(s.Line, s.Col, "type mismatch: variable %q declared as %s but got %s", s.Name, declared, valType)
			}
			scope.setVar(s.Name, declared)
		} else {
			scope.setVar(s.Name, valType)
		}
		return nil

	case *parser.MultiVarDecl:
		_, err := c.checkExpr(s.Value, scope)
		if err != nil {
			return err
		}
		for _, name := range s.Names {
			scope.setVar(name, TypeUnknown)
		}
		return nil

	case *parser.AssignStmt:
		existing, ok := scope.getVar(s.Name)
		if !ok {
			return c.errorf(s.Line, s.Col, "undefined variable %q", s.Name)
		}
		valType, err := c.checkExpr(s.Value, scope)
		if err != nil {
			return err
		}
		if !existing.Equal(valType) && existing != TypeUnknown {
			return c.errorf(s.Line, s.Col, "cannot assign %s to variable %q of type %s", valType, s.Name, existing)
		}
		return nil

	case *parser.ReturnStmt:
		retType, err := c.checkExpr(s.Value, scope)
		if err != nil {
			return err
		}
		if c.currentReturn != nil && !c.currentReturn.Equal(retType) {
			return c.errorf(s.Line, s.Col, "return type mismatch: expected %s, got %s", c.currentReturn, retType)
		}
		return nil

	case *parser.IfStmt:
		condType, err := c.checkExpr(s.Condition, scope)
		if err != nil {
			return err
		}
		if condType != TypeBool && condType != TypeUnknown {
			return c.errorf(s.Line, s.Col, "if condition must be bool, got %s", condType)
		}
		if err := c.checkBlock(s.Then, newScope(scope)); err != nil {
			return err
		}
		if s.Else != nil {
			if err := c.checkBlock(s.Else, newScope(scope)); err != nil {
				return err
			}
		}
		return nil

	case *parser.ForStmt:
		loopScope := newScope(scope)
		if s.Init != nil {
			if err := c.checkStmt(s.Init, loopScope); err != nil {
				return err
			}
		}
		if s.Condition != nil {
			condType, err := c.checkExpr(s.Condition, loopScope)
			if err != nil {
				return err
			}
			if condType != TypeBool && condType != TypeUnknown {
				return c.errorf(s.Line, s.Col, "for condition must be bool, got %s", condType)
			}
		}
		if s.Post != nil {
			if err := c.checkStmt(s.Post, loopScope); err != nil {
				return err
			}
		}
		return c.checkBlock(s.Body, newScope(loopScope))

	case *parser.ExprStmt:
		_, err := c.checkExpr(s.Expr, scope)
		return err

	case *parser.FuncDecl:
		return c.checkFunc(s, scope)

	case *parser.StructDecl:
		c.structs[s.Name] = s
		return nil

	case *parser.IndexAssign:
		_, err := c.checkExpr(s.Object, scope)
		if err != nil {
			return err
		}
		_, err = c.checkExpr(s.Index, scope)
		if err != nil {
			return err
		}
		_, err = c.checkExpr(s.Value, scope)
		return err

	case *parser.ImportStmt:
		return nil
	}

	return fmt.Errorf("unknown statement type: %T", node)
}

func (c *Checker) checkBlock(block *parser.Block, scope *Scope) error {
	for _, stmt := range block.Statements {
		if err := c.checkStmt(stmt, scope); err != nil {
			return err
		}
	}
	return nil
}

func (c *Checker) checkFunc(s *parser.FuncDecl, scope *Scope) error {
	ft, err := c.buildFuncType(s)
	if err != nil {
		return err
	}

	fnScope := newScope(scope)
	for i, p := range s.Params {
		fnScope.setVar(p.Name, ft.Params[i])
	}

	prev := c.currentReturn
	c.currentReturn = ft.ReturnType
	err = c.checkBlock(s.Body, fnScope)
	c.currentReturn = prev

	return err
}

func (c *Checker) checkExpr(node parser.Node, scope *Scope) (*Type, error) {
	switch e := node.(type) {

	case *parser.IntLiteral:
		return TypeInt, nil

	case *parser.FloatLiteral:
		return TypeFloat, nil

	case *parser.StringLiteral:
		return TypeString, nil

	case *parser.BoolLiteral:
		return TypeBool, nil

	case *parser.NilLiteral:
		return TypeNil, nil

	case *parser.Identifier:
		t, ok := scope.getVar(e.Name)
		if !ok {
			return nil, c.errorf(e.Line, e.Col, "undefined variable %q", e.Name)
		}
		return t, nil

	case *parser.BinaryExpr:
		return c.checkBinary(e, scope)

	case *parser.UnaryExpr:
		return c.checkUnary(e, scope)

	case *parser.CallExpr:
		return c.checkCall(e, scope)

	case *parser.FieldAccess:
		return TypeUnknown, nil

	case *parser.IndexExpr:
		_, err := c.checkExpr(e.Object, scope)
		if err != nil {
			return nil, err
		}
		_, err = c.checkExpr(e.Index, scope)
		if err != nil {
			return nil, err
		}
		return TypeUnknown, nil

	case *parser.ArrayLiteral:
		for _, el := range e.Elements {
			if _, err := c.checkExpr(el, scope); err != nil {
				return nil, err
			}
		}
		return TypeArray, nil

	case *parser.MapLiteral:
		for _, pair := range e.Pairs {
			if _, err := c.checkExpr(pair.Key, scope); err != nil {
				return nil, err
			}
			if _, err := c.checkExpr(pair.Value, scope); err != nil {
				return nil, err
			}
		}
		return TypeMap, nil

	case *parser.StructLiteral:
		return TypeUnknown, nil
	}

	return nil, fmt.Errorf("unknown expression type: %T", node)
}

func (c *Checker) checkBinary(e *parser.BinaryExpr, scope *Scope) (*Type, error) {
	left, err := c.checkExpr(e.Left, scope)
	if err != nil {
		return nil, err
	}
	right, err := c.checkExpr(e.Right, scope)
	if err != nil {
		return nil, err
	}

	switch e.Operator {
	case "+":
		if left == TypeString || right == TypeString {
			return TypeString, nil
		}
		if left == TypeUnknown || right == TypeUnknown {
			return TypeUnknown, nil
		}
		if !left.Equal(right) {
			return nil, c.errorf(e.Line, e.Col, "type mismatch in +: %s and %s", left, right)
		}
		return left, nil

	case "-", "*", "/", "%":
		if left == TypeUnknown || right == TypeUnknown {
			return TypeUnknown, nil
		}
		if left != TypeInt && left != TypeFloat {
			return nil, c.errorf(e.Line, e.Col, "operator %s requires number, got %s", e.Operator, left)
		}
		if !left.Equal(right) {
			return nil, c.errorf(e.Line, e.Col, "type mismatch in %s: %s and %s", e.Operator, left, right)
		}
		return left, nil

	case "==", "!=":
		return TypeBool, nil

	case "<", ">", "<=", ">=":
		if left == TypeUnknown || right == TypeUnknown {
			return TypeBool, nil
		}
		if left != TypeInt && left != TypeFloat {
			return nil, c.errorf(e.Line, e.Col, "operator %s requires number, got %s", e.Operator, left)
		}
		return TypeBool, nil

	case "&&", "||":
		return TypeBool, nil
	}

	return nil, c.errorf(e.Line, e.Col, "unknown operator %q", e.Operator)
}

func (c *Checker) checkUnary(e *parser.UnaryExpr, scope *Scope) (*Type, error) {
	right, err := c.checkExpr(e.Right, scope)
	if err != nil {
		return nil, err
	}
	switch e.Operator {
	case "!":
		return TypeBool, nil
	case "-":
		if right == TypeInt || right == TypeFloat || right == TypeUnknown {
			return right, nil
		}
		return nil, c.errorf(e.Line, e.Col, "cannot negate %s", right)
	}
	return nil, c.errorf(e.Line, e.Col, "unknown unary operator %q", e.Operator)
}

func (c *Checker) checkCall(e *parser.CallExpr, scope *Scope) (*Type, error) {
	ft, ok := scope.getFunc(e.Function)
	if !ok {
		return nil, c.errorf(e.Line, e.Col, "undefined function %q", e.Function)
	}

	// built-ins with variadic args skip param count check
	variadic := e.Function == "print" || e.Function == "println" ||
		e.Function == "len" || e.Function == "append"

	if !variadic && len(e.Args) != len(ft.Params) {
		return nil, c.errorf(e.Line, e.Col, "function %q expects %d args, got %d", e.Function, len(ft.Params), len(e.Args))
	}

	for _, arg := range e.Args {
		if _, err := c.checkExpr(arg, scope); err != nil {
			return nil, err
		}
	}

	return ft.ReturnType, nil
}

func (c *Checker) errorf(line, col int, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if line > 0 {
		return fmt.Errorf("type error at line %d, col %d: %s", line, col, msg)
	}
	return fmt.Errorf("type error: %s", msg)
}