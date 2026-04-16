# Riner - Project Notes

## The idea

Riner is a general-purpose, statically typed language inspired by Go's philosophy: less is more.
The compiler is written in Go. The language compiles to C, which then generates a native binary.
Open source, MIT license.

Core rules:
- If you can do it with less, do it with less
- No magic, no hidden behavior
- Human-friendly error messages (the compiler is your ally)
- Fast by default

---

## Language decisions so far

- Blocks delimited by curly braces `{ }`
- Variables declared with `var`, type inferred automatically
- Explicit type when needed: `var x int = 0`
- Functions with `func`
- Loops: `for i in range(0, 10)` and `while`
- Structs with `struct`
- File extension: `.rn`
- Error handling: not decided yet (options are explicit return, Result type, or propagation with `!`)

---

## Stage 1 - Lexer

The lexer reads raw source code and turns it into tokens.

A token is the smallest unit of the language: keywords (`var`, `func`, `if`), identifiers (`name`, `age`), literals (`10`, `"hello"`), symbols (`{`, `}`, `=`).

Goals:
- [ ] Define all token types
- [ ] Read source file character by character
- [ ] Emit a list of tokens
- [ ] Handle whitespace, comments, newlines
- [ ] Report position (line + column) for every token so errors show exact location

Test: given `var x = 10`, the lexer should emit `[VAR] [IDENT:x] [EQUALS] [INT:10]`

---

## Stage 2 - Parser

The parser takes the token list and builds an AST (abstract syntax tree). The AST is a tree structure that represents the program in a way the compiler can work with.

Goals:
- [ ] Define all AST node types (VarDecl, FuncDecl, IfStmt, BinaryExpr, etc.)
- [ ] Parse variable declarations
- [ ] Parse function declarations
- [ ] Parse expressions (math, comparisons, function calls)
- [ ] Parse control flow (if/else, for, while)
- [ ] Parse structs
- [ ] Good error messages when syntax is wrong

Test: given the token list from stage 1, produce a tree like `VarDecl(name=x, value=IntLiteral(10))`

---

## Stage 3 - Tree-walking interpreter

Before compiling to C, build a simple interpreter that walks the AST and executes it directly.
This lets you test the language design fast without dealing with code generation yet.

Goals:
- [ ] Evaluate expressions
- [ ] Execute variable declarations and assignments
- [ ] Execute if/else, for, while
- [ ] Call functions
- [ ] Basic built-ins: `print`

This is temporary. Once the language feels right, move to the real compiler pipeline.

---

## Stage 4 - Type checker

The type checker walks the AST and verifies that types are correct before the program runs.

Goals:
- [ ] Infer types from expressions (`var x = 10` -> x is int)
- [ ] Check that types match in assignments and function calls
- [ ] Catch undefined variables
- [ ] Catch wrong number of arguments in function calls
- [ ] Scope resolution (variables only accessible in their block)
- [ ] Human-friendly error messages with line and column

---

## Stage 5 - IR (intermediate representation)

The IR is a simplified version of the program, between the AST and the final code. Easier to optimize than the AST, easier to generate than machine code.

Goals:
- [ ] Define IR instruction set
- [ ] Lower AST into IR
- [ ] Basic optimizations: constant folding, dead code elimination

---

## Stage 6 - Code generation (C backend)

Translate IR into C code. Then use GCC or Clang to compile the C into a native binary.

Goals:
- [ ] Emit valid C from IR
- [ ] Handle all basic types
- [ ] Handle functions, structs, loops
- [ ] Link with the Riner runtime

Why C and not LLVM: simpler to implement, portable, GCC/Clang handle all the hard optimization work.
LLVM can come later if needed.

---

## Stage 7 - Runtime

Every Riner binary includes a small C runtime that handles:
- [ ] Memory allocation
- [ ] Garbage collector (if the language has one)
- [ ] Panic / crash handling
- [ ] Built-in functions that are too low-level to write in Riner itself

---

## Stage 8 - Standard library

The stdlib is written in Riner itself (once the language is stable enough).

Planned packages:
- `io` - print, read, file operations
- `math` - basic math functions
- `strings` - split, trim, contains, replace
- `collections` - arrays, maps, sets
- `net` - http, tcp (later stage)

---

## Stage 9 - Tooling

- [ ] `rinerfmt` - code formatter (like gofmt)
- [ ] VS Code extension - syntax highlighting via tmLanguage.json (easy to do early, makes the language feel real)
- [ ] LSP (language server) - autocomplete, go to definition, errors inline
- [ ] REPL - run Riner code interactively in the terminal

---

## Stage 10 - Bootstrap (optional, long term)

Rewrite the Riner compiler in Riner itself. Use the Go compiler to compile the new one, then the new one compiles itself.

This is what Go and Rust did. Not required for the language to be useful, but it's a milestone that means the language is mature.

---

## Compiler written in Go

The compiler itself is a Go project. Structure:

```
cmd/compiler    - CLI entrypoint
internal/lexer  - stage 1
internal/parser - stage 2
internal/ast    - AST node definitions
internal/sema   - type checker (stage 4)
internal/ir     - IR (stage 5)
internal/codegen - C backend (stage 6)
internal/diagnostics - error formatting
runtime/        - C runtime (stage 7)
stdlib/         - standard library (stage 8)
tests/          - unit, integration, conformance, fuzz
docs/spec.md    - language specification
docs/grammar.ebnf - formal grammar
```

---

## Error message philosophy

One of the main differentiators from Go. Every error should tell you:
- What went wrong
- Where exactly (file, line, column)
- Why it happened
- What to do about it

Example of what we want:

```
error: type mismatch on line 5, column 10
  var x int = "hello"
              ^^^^^^^
  expected int, got string
  hint: remove the quotes or change the type to string
```

---

## Notes on syntax still to decide

- Error handling strategy (explicit return, Result type, or `!` propagation)
- How imports work (`import "io"` or something else)
- Whether to have interfaces or not
- Concurrency (goroutine-style, async/await, or nothing for now)
- Arrays and maps syntax
- Null / nil handling