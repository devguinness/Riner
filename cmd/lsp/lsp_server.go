package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/devguinness/Riner/internal/lexer"
	"github.com/devguinness/Riner/internal/parser"
	"github.com/devguinness/Riner/internal/sema"
)

// ── JSON-RPC types ──

type Request struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RpcError   `json:"error,omitempty"`
}

type Notification struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ── LSP types ──

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=error, 2=warn, 3=info
	Message  string `json:"message"`
	Source   string `json:"source"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type HoverResult struct {
	Contents string `json:"contents"`
}

type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// ── Server ──

type Server struct {
	reader *bufio.Reader
	writer io.Writer
	docs   map[string]string // uri -> content
}

func NewServer() *Server {
	return &Server{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
		docs:   make(map[string]string),
	}
}

func (s *Server) Run() {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("read error: %v", err)
			continue
		}

		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}

		s.handle(req)
	}
}

func (s *Server) handle(req Request) {
	switch req.Method {
	case "initialize":
		s.respond(req.ID, map[string]interface{}{
			"capabilities": map[string]interface{}{
				"textDocumentSync": 1, // full sync
				"hoverProvider":    true,
				"definitionProvider": true,
			},
		})

	case "initialized":
		// no-op

	case "shutdown":
		s.respond(req.ID, nil)

	case "exit":
		os.Exit(0)

	case "textDocument/didOpen":
		var params DidOpenParams
		json.Unmarshal(req.Params, &params)
		s.docs[params.TextDocument.URI] = params.TextDocument.Text
		s.diagnose(params.TextDocument.URI, params.TextDocument.Text)

	case "textDocument/didChange":
		var params DidChangeParams
		json.Unmarshal(req.Params, &params)
		if len(params.ContentChanges) > 0 {
			text := params.ContentChanges[len(params.ContentChanges)-1].Text
			s.docs[params.TextDocument.URI] = text
			s.diagnose(params.TextDocument.URI, text)
		}

	case "textDocument/didClose":
		var params struct {
			TextDocument TextDocumentIdentifier `json:"textDocument"`
		}
		json.Unmarshal(req.Params, &params)
		delete(s.docs, params.TextDocument.URI)

	case "textDocument/hover":
		var params HoverParams
		json.Unmarshal(req.Params, &params)
		result := s.hover(params)
		s.respond(req.ID, result)

	case "textDocument/definition":
		var params DefinitionParams
		json.Unmarshal(req.Params, &params)
		result := s.definition(params)
		s.respond(req.ID, result)

	default:
		if req.ID != nil {
			s.respond(req.ID, nil)
		}
	}
}

func (s *Server) diagnose(uri, text string) {
	var diagnostics []Diagnostic

	// lex
	l := lexer.New(text)
	tokens := l.Tokenize()

	for _, tok := range tokens {
		if tok.Type == lexer.TOKEN_ILLEGAL {
			diagnostics = append(diagnostics, Diagnostic{
				Range: Range{
					Start: Position{Line: tok.Line - 1, Character: tok.Column - 1},
					End:   Position{Line: tok.Line - 1, Character: tok.Column},
				},
				Severity: 1,
				Message:  fmt.Sprintf("unexpected character %q", tok.Value),
				Source:   "riner",
			})
		}
	}

	// parse
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		line, col := extractPosition(err.Error())
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 1},
			},
			Severity: 1,
			Message:  cleanError(err.Error()),
			Source:   "riner",
		})
		s.publishDiagnostics(uri, diagnostics)
		return
	}

	// type check
	c := sema.New()
	if err := c.Check(prog); err != nil {
		line, col := extractPosition(err.Error())
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: line, Character: col},
				End:   Position{Line: line, Character: col + 1},
			},
			Severity: 1,
			Message:  cleanError(err.Error()),
			Source:   "riner",
		})
	}

	s.publishDiagnostics(uri, diagnostics)
}

func (s *Server) hover(params HoverParams) interface{} {
	text, ok := s.docs[params.TextDocument.URI]
	if !ok {
		return nil
	}

	word := wordAt(text, params.Position)
	if word == "" {
		return nil
	}

	// check if it's a keyword
	keywords := map[string]string{
		"var":    "keyword: declares a variable",
		"func":   "keyword: declares a function",
		"struct": "keyword: declares a struct",
		"if":     "keyword: conditional",
		"else":   "keyword: else branch",
		"for":    "keyword: loop",
		"return": "keyword: returns a value",
		"true":   "bool literal: true",
		"false":  "bool literal: false",
		"nil":    "nil: absence of value",
		"int":    "type: integer (64-bit)",
		"float":  "type: floating point (64-bit)",
		"string": "type: UTF-8 string",
		"bool":   "type: boolean",
		"print":  "builtin: print(args...) - prints to stdout",
		"println": "builtin: println(args...) - prints with newline",
		"len":    "builtin: len(s) - returns length of string or array",
		"append": "builtin: append(arr, val) - appends to array",
	}

	if desc, ok := keywords[word]; ok {
		return HoverResult{Contents: desc}
	}

	// try to infer type from source
	l := lexer.New(text)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return nil
	}
	c := sema.New()
	c.Check(prog)

	return nil
}

func (s *Server) definition(params DefinitionParams) interface{} {
	text, ok := s.docs[params.TextDocument.URI]
	if !ok {
		return nil
	}

	word := wordAt(text, params.Position)
	if word == "" {
		return nil
	}

	// find declaration in source
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, "var "+word) ||
			strings.Contains(line, "func "+word) ||
			strings.Contains(line, "struct "+word) {
			col := strings.Index(line, word)
			return Location{
				URI: params.TextDocument.URI,
				Range: Range{
					Start: Position{Line: i, Character: col},
					End:   Position{Line: i, Character: col + len(word)},
				},
			}
		}
	}

	return nil
}

func (s *Server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	s.notify("textDocument/publishDiagnostics", map[string]interface{}{
		"uri":         uri,
		"diagnostics": diagnostics,
	})
}

func (s *Server) respond(id interface{}, result interface{}) {
	resp := Response{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  result,
	}
	s.send(resp)
}

func (s *Server) notify(method string, params interface{}) {
	n := Notification{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}
	s.send(n)
}

func (s *Server) send(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)
	fmt.Fprint(s.writer, msg)
}

func (s *Server) readMessage() ([]byte, error) {
	// read headers
	contentLength := 0
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, _ = strconv.Atoi(val)
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("no content length")
	}

	buf := make([]byte, contentLength)
	_, err := io.ReadFull(s.reader, buf)
	return buf, err
}

// ── helpers ──

func wordAt(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line >= len(lines) {
		return ""
	}
	line := lines[pos.Line]
	if pos.Character >= len(line) {
		return ""
	}

	start := pos.Character
	for start > 0 && isIdentChar(line[start-1]) {
		start--
	}
	end := pos.Character
	for end < len(line) && isIdentChar(line[end]) {
		end++
	}

	return line[start:end]
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func extractPosition(msg string) (int, int) {
	// parse "... at line X, col Y: ..."
	var line, col int
	fmt.Sscanf(msg, "%*s error at line %d, col %d", &line, &col)
	if line > 0 {
		return line - 1, col - 1
	}
	return 0, 0
}

func cleanError(msg string) string {
	// remove "parse error at line X, col Y: " prefix
	if idx := strings.Index(msg, ": "); idx != -1 {
		return msg[idx+2:]
	}
	return msg
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	s := NewServer()
	s.Run()
}