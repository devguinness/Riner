package parser

import (
	"fmt"

	"github.com/devguinness/Riner/internal/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{}

	for !p.isEOF() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		program.Statements = append(program.Statements, stmt)
	}

	return program, nil
}

// --- statements ---

func (p *Parser) parseStatement() (Node, error) {
	tok := p.current()

	switch tok.Type {
	case lexer.TOKEN_VAR:
		return p.parseVarDecl()
	case lexer.TOKEN_FUNC:
		return p.parseFuncDecl()
	case lexer.TOKEN_STRUCT:
		return p.parseStructDecl()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_FOR:
		return p.parseForStmt()
	case lexer.TOKEN_IDENT:
		return p.parseIdentStmt()
	}

	return nil, p.errorf("unexpected token %s", tok.Type)
}

func (p *Parser) parseVarDecl() (*VarDecl, error) {
	line, col := p.current().Line, p.current().Column
	p.advance() // consume 'var'

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	// optional explicit type
	typeName := ""
	if p.current().Type == lexer.TOKEN_IDENT {
		typeName = p.current().Value
		p.advance()
	}

	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}

	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &VarDecl{Name: name.Value, Type: typeName, Value: value, Line: line, Col: col}, nil
}

func (p *Parser) parseFuncDecl() (*FuncDecl, error) {
	line, col := p.current().Line, p.current().Column
	p.advance() // consume 'func'

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}

	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}

	// optional return type
	returnType := ""
	if p.current().Type == lexer.TOKEN_IDENT {
		returnType = p.current().Value
		p.advance()
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &FuncDecl{Name: name.Value, Params: params, ReturnType: returnType, Body: body, Line: line, Col: col}, nil
}

func (p *Parser) parseParams() ([]Param, error) {
	var params []Param

	for p.current().Type != lexer.TOKEN_RPAREN && !p.isEOF() {
		name, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}

		typeName, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}

		params = append(params, Param{Name: name.Value, Type: typeName.Value})

		if p.current().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}

	return params, nil
}

func (p *Parser) parseStructDecl() (*StructDecl, error) {
	line, col := p.current().Line, p.current().Column
	p.advance() // consume 'struct'

	name, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_LBRACE); err != nil {
		return nil, err
	}

	var fields []Param
	for p.current().Type != lexer.TOKEN_RBRACE && !p.isEOF() {
		fieldName, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		fieldType, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		fields = append(fields, Param{Name: fieldName.Value, Type: fieldType.Value})
	}

	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}

	return &StructDecl{Name: name.Value, Fields: fields, Line: line, Col: col}, nil
}

func (p *Parser) parseReturnStmt() (*ReturnStmt, error) {
	line, col := p.current().Line, p.current().Column
	p.advance() // consume 'return'

	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{Value: value, Line: line, Col: col}, nil
}

func (p *Parser) parseIfStmt() (*IfStmt, error) {
	line, col := p.current().Line, p.current().Column
	p.advance() // consume 'if'

	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseBlock *Block
	if p.current().Type == lexer.TOKEN_ELSE {
		p.advance()
		elseBlock, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &IfStmt{Condition: condition, Then: then, Else: elseBlock, Line: line, Col: col}, nil
}

func (p *Parser) parseForStmt() (*ForStmt, error) {
	line, col := p.current().Line, p.current().Column
	p.advance() // consume 'for'

	// boolean for: next token is an expression that is not an assignment
	// classic for: ident = expr ; expr ; ident = expr { }
	// we detect by checking if we see "ident = ... ;" pattern
	if p.current().Type == lexer.TOKEN_IDENT && p.peek().Type == lexer.TOKEN_ASSIGN {
		return p.parseClassicFor(line, col)
	}

	// boolean for
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ForStmt{Condition: condition, Body: body, Line: line, Col: col}, nil
}

func (p *Parser) parseClassicFor(line, col int) (*ForStmt, error) {
	init, err := p.parseAssignStmt()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_SEMICOLON); err != nil {
		return nil, err
	}

	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.TOKEN_SEMICOLON); err != nil {
		return nil, err
	}

	post, err := p.parseAssignStmt()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &ForStmt{Init: init, Condition: condition, Post: post, Body: body, Line: line, Col: col}, nil
}

// parseIdentStmt handles: assignment or function call as a statement
func (p *Parser) parseIdentStmt() (Node, error) {
	if p.peek().Type == lexer.TOKEN_ASSIGN {
		return p.parseAssignStmt()
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ExprStmt{Expr: expr}, nil
}

func (p *Parser) parseAssignStmt() (*AssignStmt, error) {
	line, col := p.current().Line, p.current().Column
	name := p.current().Value
	p.advance() // consume ident

	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}

	value, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	return &AssignStmt{Name: name, Value: value, Line: line, Col: col}, nil
}

func (p *Parser) parseBlock() (*Block, error) {
	if _, err := p.expect(lexer.TOKEN_LBRACE); err != nil {
		return nil, err
	}

	block := &Block{}
	for p.current().Type != lexer.TOKEN_RBRACE && !p.isEOF() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		block.Statements = append(block.Statements, stmt)
	}

	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}

	return block, nil
}

// --- expressions ---

func (p *Parser) parseExpr() (Node, error) {
	return p.parseComparison()
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}

	for {
		op := p.current().Type
		if op != lexer.TOKEN_EQ && op != lexer.TOKEN_NEQ &&
			op != lexer.TOKEN_LT && op != lexer.TOKEN_GT &&
			op != lexer.TOKEN_LTE && op != lexer.TOKEN_GTE &&
			op != lexer.TOKEN_AND && op != lexer.TOKEN_OR {
			break
		}
		line, col := p.current().Line, p.current().Column
		opVal := p.current().Value
		p.advance()

		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: opVal, Right: right, Line: line, Col: col}
	}

	return left, nil
}

func (p *Parser) parseAddSub() (Node, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}

	for p.current().Type == lexer.TOKEN_PLUS || p.current().Type == lexer.TOKEN_MINUS {
		line, col := p.current().Line, p.current().Column
		opVal := p.current().Value
		p.advance()

		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: opVal, Right: right, Line: line, Col: col}
	}

	return left, nil
}

func (p *Parser) parseMulDiv() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.current().Type == lexer.TOKEN_STAR || p.current().Type == lexer.TOKEN_SLASH || p.current().Type == lexer.TOKEN_PERCENT {
		line, col := p.current().Line, p.current().Column
		opVal := p.current().Value
		p.advance()

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Operator: opVal, Right: right, Line: line, Col: col}
	}

	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.current().Type == lexer.TOKEN_BANG || p.current().Type == lexer.TOKEN_MINUS {
		line, col := p.current().Line, p.current().Column
		op := p.current().Value
		p.advance()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Operator: op, Right: right, Line: line, Col: col}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Node, error) {
	tok := p.current()

	switch tok.Type {
	case lexer.TOKEN_INT:
		p.advance()
		return &IntLiteral{Value: tok.Value, Line: tok.Line, Col: tok.Column}, nil

	case lexer.TOKEN_FLOAT:
		p.advance()
		return &FloatLiteral{Value: tok.Value, Line: tok.Line, Col: tok.Column}, nil

	case lexer.TOKEN_STRING:
		p.advance()
		return &StringLiteral{Value: tok.Value, Line: tok.Line, Col: tok.Column}, nil

	case lexer.TOKEN_TRUE:
		p.advance()
		return &BoolLiteral{Value: true, Line: tok.Line, Col: tok.Column}, nil

	case lexer.TOKEN_FALSE:
		p.advance()
		return &BoolLiteral{Value: false, Line: tok.Line, Col: tok.Column}, nil

	case lexer.TOKEN_NIL:
		p.advance()
		return &NilLiteral{Line: tok.Line, Col: tok.Column}, nil

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TOKEN_IDENT:
		return p.parseIdentExpr()
	}

	return nil, p.errorf("unexpected token %s %q", tok.Type, tok.Value)
}

func (p *Parser) parseIdentExpr() (Node, error) {
	tok := p.current()
	p.advance()

	// struct literal: Name { field: value, ... }
	// only uppercase identifiers are types (Person, User, etc.)
	if p.current().Type == lexer.TOKEN_LBRACE && len(tok.Value) > 0 && tok.Value[0] >= 'A' && tok.Value[0] <= 'Z' {
		return p.parseStructLiteral(tok)
	}

	// function call: name(args)
	if p.current().Type == lexer.TOKEN_LPAREN {
		return p.parseCallExpr(tok)
	}

	// field access: name.field
	node := Node(&Identifier{Name: tok.Value, Line: tok.Line, Col: tok.Column})
	for p.current().Type == lexer.TOKEN_DOT {
		p.advance()
		field, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		node = &FieldAccess{Object: node, Field: field.Value, Line: field.Line, Col: field.Column}
	}

	return node, nil
}

func (p *Parser) parseCallExpr(tok lexer.Token) (*CallExpr, error) {
	p.advance() // consume '('

	var args []Node
	for p.current().Type != lexer.TOKEN_RPAREN && !p.isEOF() {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.current().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}

	return &CallExpr{Function: tok.Value, Args: args, Line: tok.Line, Col: tok.Column}, nil
}

func (p *Parser) parseStructLiteral(tok lexer.Token) (*StructLiteral, error) {
	p.advance() // consume '{'

	var fields []FieldValue
	for p.current().Type != lexer.TOKEN_RBRACE && !p.isEOF() {
		name, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TOKEN_COLON); err != nil {
			return nil, err
		}
		value, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		fields = append(fields, FieldValue{Name: name.Value, Value: value})

		if p.current().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.TOKEN_RBRACE); err != nil {
		return nil, err
	}

	return &StructLiteral{Name: tok.Value, Fields: fields, Line: tok.Line, Col: tok.Column}, nil
}

// --- helpers ---

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() lexer.Token {
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) isEOF() bool {
	return p.current().Type == lexer.TOKEN_EOF
}

func (p *Parser) expect(t lexer.TokenType) (lexer.Token, error) {
	tok := p.current()
	if tok.Type != t {
		return lexer.Token{}, p.errorf("expected %s, got %s %q", t, tok.Type, tok.Value)
	}
	p.advance()
	return tok, nil
}

func (p *Parser) errorf(format string, args ...any) error {
	tok := p.current()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("parse error at line %d, col %d: %s", tok.Line, tok.Column, msg)
}