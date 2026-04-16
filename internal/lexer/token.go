package lexer

import "fmt"

type TokenType string

const (
	// Literals
	TOKEN_INT    TokenType = "INT"
	TOKEN_FLOAT  TokenType = "FLOAT"
	TOKEN_STRING TokenType = "STRING"
	TOKEN_IDENT  TokenType = "IDENT"

	// Keywords
	TOKEN_VAR    TokenType = "VAR"
	TOKEN_FUNC   TokenType = "FUNC"
	TOKEN_STRUCT TokenType = "STRUCT"
	TOKEN_IMPORT TokenType = "IMPORT"
	TOKEN_IF     TokenType = "IF"
	TOKEN_ELSE   TokenType = "ELSE"
	TOKEN_FOR    TokenType = "FOR"
	TOKEN_RETURN TokenType = "RETURN"

	// Built-in values
	TOKEN_TRUE  TokenType = "TRUE"
	TOKEN_FALSE TokenType = "FALSE"
	TOKEN_NIL   TokenType = "NIL"

	// Operators
	TOKEN_PLUS    TokenType = "+"
	TOKEN_MINUS   TokenType = "-"
	TOKEN_STAR    TokenType = "*"
	TOKEN_SLASH   TokenType = "/"
	TOKEN_PERCENT TokenType = "%"
	TOKEN_EQ      TokenType = "=="
	TOKEN_NEQ     TokenType = "!="
	TOKEN_LT      TokenType = "<"
	TOKEN_GT      TokenType = ">"
	TOKEN_LTE     TokenType = "<="
	TOKEN_GTE     TokenType = ">="
	TOKEN_AND     TokenType = "&&"
	TOKEN_OR      TokenType = "||"
	TOKEN_BANG    TokenType = "!"
	TOKEN_ASSIGN  TokenType = "="

	// Delimiters
	TOKEN_LPAREN    TokenType = "("
	TOKEN_RPAREN    TokenType = ")"
	TOKEN_LBRACE    TokenType = "{"
	TOKEN_RBRACE    TokenType = "}"
	TOKEN_COMMA     TokenType = ","
	TOKEN_SEMICOLON TokenType = ";"
	TOKEN_DOT       TokenType = "."
	TOKEN_COLON     TokenType = ":"

	// Special
	TOKEN_EOF     TokenType = "EOF"
	TOKEN_ILLEGAL TokenType = "ILLEGAL"
)

var keywords = map[string]TokenType{
	"var":    TOKEN_VAR,
	"func":   TOKEN_FUNC,
	"struct": TOKEN_STRUCT,
	"import": TOKEN_IMPORT,
	"if":     TOKEN_IF,
	"else":   TOKEN_ELSE,
	"for":    TOKEN_FOR,
	"return": TOKEN_RETURN,
	"true":   TOKEN_TRUE,
	"false":  TOKEN_FALSE,
	"nil":    TOKEN_NIL,
}

// LookupIdent returns the keyword token type for a given identifier,
// or TOKEN_IDENT if it's not a keyword.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("[%s:%q line:%d col:%d]", t.Type, t.Value, t.Line, t.Column)
}

// IsLiteral reports whether the token is a literal value.
func (t TokenType) IsLiteral() bool {
	return t == TOKEN_INT || t == TOKEN_FLOAT || t == TOKEN_STRING
}

// IsOperator reports whether the token is an operator.
func (t TokenType) IsOperator() bool {
	switch t {
	case TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT,
		TOKEN_EQ, TOKEN_NEQ, TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE,
		TOKEN_AND, TOKEN_OR, TOKEN_BANG, TOKEN_ASSIGN:
		return true
	}
	return false
}

// IsKeyword reports whether the token is a language keyword.
func (t TokenType) IsKeyword() bool {
	switch t {
	case TOKEN_VAR, TOKEN_FUNC, TOKEN_STRUCT, TOKEN_IMPORT,
		TOKEN_IF, TOKEN_ELSE, TOKEN_FOR, TOKEN_RETURN:
		return true
	}
	return false
}