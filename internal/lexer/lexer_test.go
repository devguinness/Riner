package lexer

import (
	"testing"
)

func TestHelloWorld(t *testing.T) {
	source := `func main() {
    print("hello, world")
}`
	l := New(source)
	tokens := l.Tokenize()

	expected := []TokenType{
		TOKEN_FUNC, TOKEN_IDENT, TOKEN_LPAREN, TOKEN_RPAREN,
		TOKEN_LBRACE, TOKEN_IDENT, TOKEN_LPAREN, TOKEN_STRING,
		TOKEN_RPAREN, TOKEN_RBRACE, TOKEN_EOF,
	}

	for i, tt := range expected {
		if i >= len(tokens) {
			t.Fatalf("expected token %s at index %d but got nothing", tt, i)
		}
		if tokens[i].Type != tt {
			t.Errorf("index %d: expected %s, got %s (%q)", i, tt, tokens[i].Type, tokens[i].Value)
		}
	}
}

func TestVariables(t *testing.T) {
	source := `var age = 17`
	l := New(source)
	tokens := l.Tokenize()

	expected := []TokenType{TOKEN_VAR, TOKEN_IDENT, TOKEN_ASSIGN, TOKEN_INT, TOKEN_EOF}

	for i, tt := range expected {
		if tokens[i].Type != tt {
			t.Errorf("index %d: expected %s, got %s", i, tt, tokens[i].Type)
		}
	}
}

func TestOperators(t *testing.T) {
	source := `== != <= >= && ||`
	l := New(source)
	tokens := l.Tokenize()

	expected := []TokenType{TOKEN_EQ, TOKEN_NEQ, TOKEN_LTE, TOKEN_GTE, TOKEN_AND, TOKEN_OR, TOKEN_EOF}

	for i, tt := range expected {
		if tokens[i].Type != tt {
			t.Errorf("index %d: expected %s, got %s", i, tt, tokens[i].Type)
		}
	}
}

func TestLineTracking(t *testing.T) {
	source := "var x = 1\nvar y = 2"
	l := New(source)
	tokens := l.Tokenize()

	if tokens[0].Line != 1 {
		t.Errorf("expected line 1, got %d", tokens[0].Line)
	}

	// find second var
	for _, tok := range tokens {
		if tok.Type == TOKEN_VAR && tok.Line == 2 {
			return
		}
	}
	t.Error("expected second var on line 2")
}

func TestString(t *testing.T) {
	source := `"hello, world"`
	l := New(source)
	tokens := l.Tokenize()

	if tokens[0].Type != TOKEN_STRING {
		t.Errorf("expected STRING, got %s", tokens[0].Type)
	}
	if tokens[0].Value != "hello, world" {
		t.Errorf("expected 'hello, world', got %q", tokens[0].Value)
	}
}

func TestComment(t *testing.T) {
	source := `// this is a comment
var x = 1`
	l := New(source)
	tokens := l.Tokenize()

	if tokens[0].Type != TOKEN_VAR {
		t.Errorf("expected VAR after comment, got %s", tokens[0].Type)
	}
}