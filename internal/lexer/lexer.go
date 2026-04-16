package lexer

type Lexer struct {
	source string
	pos    int
	line   int
	column int
}

func New(source string) *Lexer {
	return &Lexer{source: source, pos: 0, line: 1, column: 1}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF || tok.Type == TOKEN_ILLEGAL {
			break
		}
	}
	return tokens
}

func (l *Lexer) nextToken() Token {
	tok, ok := l.skipWhitespaceAndComments()
	if !ok {
		return tok
	}

	if l.pos >= len(l.source) {
		return Token{Type: TOKEN_EOF, Value: "", Line: l.line, Column: l.column}
	}

	ch := l.current()

	if ch == '"' {
		return l.readString()
	}
	if isDigit(ch) {
		return l.readNumber()
	}
	if isLetter(ch) {
		return l.readIdentOrKeyword()
	}
	return l.readSymbol()
}

func (l *Lexer) readString() Token {
	line, col := l.line, l.column
	l.advance()

	var result []byte
	for l.pos < len(l.source) && l.current() != '"' {
		if l.current() == '\\' {
			l.advance()
			if l.pos >= len(l.source) {
				return Token{Type: TOKEN_ILLEGAL, Value: "unterminated string escape", Line: line, Column: col}
			}
			switch l.current() {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			default:
				result = append(result, '\\', l.current())
			}
		} else {
			result = append(result, l.current())
		}
		l.advance()
	}

	if l.pos >= len(l.source) {
		return Token{Type: TOKEN_ILLEGAL, Value: "unterminated string", Line: line, Column: col}
	}

	l.advance()
	return Token{Type: TOKEN_STRING, Value: string(result), Line: line, Column: col}
}

func (l *Lexer) readNumber() Token {
	line, col := l.line, l.column
	start := l.pos

	for l.pos < len(l.source) && isDigit(l.current()) {
		l.advance()
	}

	if l.pos < len(l.source) && l.current() == '.' && isDigit(l.peek()) {
		l.advance()
		for l.pos < len(l.source) && isDigit(l.current()) {
			l.advance()
		}
		return Token{Type: TOKEN_FLOAT, Value: l.source[start:l.pos], Line: line, Column: col}
	}

	return Token{Type: TOKEN_INT, Value: l.source[start:l.pos], Line: line, Column: col}
}

func (l *Lexer) readIdentOrKeyword() Token {
	line, col := l.line, l.column
	start := l.pos

	for l.pos < len(l.source) && (isLetter(l.current()) || isDigit(l.current())) {
		l.advance()
	}

	word := l.source[start:l.pos]
	tt := LookupIdent(word)

	return Token{Type: tt, Value: word, Line: line, Column: col}
}

func (l *Lexer) readSymbol() Token {
	line, col := l.line, l.column
	ch := l.current()

	switch ch {
	case '+':
		l.advance()
		return Token{Type: TOKEN_PLUS, Value: "+", Line: line, Column: col}
	case '-':
		l.advance()
		return Token{Type: TOKEN_MINUS, Value: "-", Line: line, Column: col}
	case '*':
		l.advance()
		return Token{Type: TOKEN_STAR, Value: "*", Line: line, Column: col}
	case '/':
		l.advance()
		return Token{Type: TOKEN_SLASH, Value: "/", Line: line, Column: col}
	case '%':
		l.advance()
		return Token{Type: TOKEN_PERCENT, Value: "%", Line: line, Column: col}
	case '(':
		l.advance()
		return Token{Type: TOKEN_LPAREN, Value: "(", Line: line, Column: col}
	case ')':
		l.advance()
		return Token{Type: TOKEN_RPAREN, Value: ")", Line: line, Column: col}
	case '{':
		l.advance()
		return Token{Type: TOKEN_LBRACE, Value: "{", Line: line, Column: col}
	case '}':
		l.advance()
		return Token{Type: TOKEN_RBRACE, Value: "}", Line: line, Column: col}
	case ',':
		l.advance()
		return Token{Type: TOKEN_COMMA, Value: ",", Line: line, Column: col}
	case ';':
		l.advance()
		return Token{Type: TOKEN_SEMICOLON, Value: ";", Line: line, Column: col}
	case '.':
		l.advance()
		return Token{Type: TOKEN_DOT, Value: ".", Line: line, Column: col}
	case ':':
		l.advance()
		return Token{Type: TOKEN_COLON, Value: ":", Line: line, Column: col}
	case '=':
		l.advance()
		if l.pos < len(l.source) && l.current() == '=' {
			l.advance()
			return Token{Type: TOKEN_EQ, Value: "==", Line: line, Column: col}
		}
		return Token{Type: TOKEN_ASSIGN, Value: "=", Line: line, Column: col}
	case '!':
		l.advance()
		if l.pos < len(l.source) && l.current() == '=' {
			l.advance()
			return Token{Type: TOKEN_NEQ, Value: "!=", Line: line, Column: col}
		}
		return Token{Type: TOKEN_BANG, Value: "!", Line: line, Column: col}
	case '<':
		l.advance()
		if l.pos < len(l.source) && l.current() == '=' {
			l.advance()
			return Token{Type: TOKEN_LTE, Value: "<=", Line: line, Column: col}
		}
		return Token{Type: TOKEN_LT, Value: "<", Line: line, Column: col}
	case '>':
		l.advance()
		if l.pos < len(l.source) && l.current() == '=' {
			l.advance()
			return Token{Type: TOKEN_GTE, Value: ">=", Line: line, Column: col}
		}
		return Token{Type: TOKEN_GT, Value: ">", Line: line, Column: col}
	case '&':
		if l.peek() == '&' {
			l.advance()
			l.advance()
			return Token{Type: TOKEN_AND, Value: "&&", Line: line, Column: col}
		}
		l.advance()
		return Token{Type: TOKEN_ILLEGAL, Value: "&", Line: line, Column: col}
	case '|':
		if l.peek() == '|' {
			l.advance()
			l.advance()
			return Token{Type: TOKEN_OR, Value: "||", Line: line, Column: col}
		}
		l.advance()
		return Token{Type: TOKEN_ILLEGAL, Value: "|", Line: line, Column: col}
	}

	l.advance()
	return Token{Type: TOKEN_ILLEGAL, Value: string(ch), Line: line, Column: col}
}

func (l *Lexer) skipWhitespaceAndComments() (Token, bool) {
	for l.pos < len(l.source) {
		ch := l.current()

		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advance()
			continue
		}

		if ch == '/' && l.peek() == '/' {
			for l.pos < len(l.source) && l.current() != '\n' {
				l.advance()
			}
			continue
		}

		if ch == '/' && l.peek() == '*' {
			line, col := l.line, l.column
			l.advance()
			l.advance()
			for {
				if l.pos >= len(l.source) {
					return Token{Type: TOKEN_ILLEGAL, Value: "unterminated block comment", Line: line, Column: col}, false
				}
				if l.current() == '*' && l.peek() == '/' {
					l.advance()
					l.advance()
					break
				}
				l.advance()
			}
			continue
		}

		break
	}
	return Token{}, true
}

func (l *Lexer) current() byte {
	return l.source[l.pos]
}

func (l *Lexer) peek() byte {
	if l.pos+1 >= len(l.source) {
		return 0
	}
	return l.source[l.pos+1]
}

func (l *Lexer) advance() {
	if l.pos < len(l.source) && l.source[l.pos] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}