# Riner

*From French: to clean. A language that removes what doesn't need to exist.*

Riner is a general-purpose, statically typed programming language built around one idea: if you can do it with less, do it with less. No unnecessary abstractions, no hidden magic, no 200 lines when 50 will do.

Inspired by the philosophy of Go, but with a sharper focus on clarity, human-friendly error messages, and a minimal footprint.

## Philosophy

- **Less is more.** Every feature must justify its existence.
- **Honest errors.** The compiler is your ally, not your enemy. Errors tell you exactly what went wrong, where, and why.
- **No magic.** What you read is what runs. No implicit behavior hiding behind syntax.
- **Fast by default.** Compiles to native code. No VM, no interpreter overhead in production.

## Quick look

```riner
func greet(name string) string {
    return "hello, " + name
}

func main() {
    var message = greet("world")
    print(message)
}
```

Variables, types, functions. Nothing surprising. Anyone who has seen a C-family language reads Riner instantly.

## Syntax

### Variables
```riner
var name = "riner"        // type inferred automatically
var count int = 0         // explicit type when needed
var pi = 3.14
var active = true
```

### Functions
```riner
func add(a int, b int) int {
    return a + b
}
```

### Conditions
```riner
if age >= 18 {
    print("adult")
} else {
    print("minor")
}
```

### Loops
```riner
for i in range(0, 10) {
    print(i)
}

while active {
    print("running...")
}
```

### Structs
```riner
struct Person {
    name string
    age  int
}

func main() {
    var p = Person{ name: "alice", age: 30 }
    print(p.name)
}
```

## Status

Riner is in early development. The language spec and compiler are being built from scratch.

| Component        | Status      |
|-----------------|-------------|
| Lexer            | In progress |
| Parser           | Planned     |
| Type checker     | Planned     |
| Code generation  | Planned     |
| Standard library | Planned     |
| LSP / tooling    | Planned     |

## Roadmap

### Phase 1 - Foundation
- [ ] Lexer (tokenizer)
- [ ] Parser -> AST
- [ ] Tree-walking interpreter
- [ ] Basic types: `int`, `float`, `string`, `bool`
- [ ] Variables, functions, conditionals, loops

### Phase 2 - Type system
- [ ] Static type checker
- [ ] Structs
- [ ] Type inference
- [ ] Error handling strategy

### Phase 3 - Native compilation
- [ ] IR (intermediate representation)
- [ ] C backend (Riner -> C -> native binary)
- [ ] Optimizer (constant folding, dead code elimination)

### Phase 4 - Ecosystem
- [ ] Standard library (`io`, `math`, `strings`, `collections`)
- [ ] Package manager
- [ ] VS Code extension (syntax highlighting, LSP)
- [ ] Formatter (`rinerfmt`)

## Project structure

```
riner/
├── cmd/
│   ├── compiler/       # main binary: file -> executable
│   ├── repl/           # interactive mode
│   └── fmt/            # code formatter
├── internal/
│   ├── lexer/          # text -> tokens
│   ├── parser/         # tokens -> AST
│   ├── ast/            # AST node definitions
│   ├── sema/           # type checker, scope resolver
│   ├── ir/             # intermediate representation
│   ├── codegen/        # IR -> C
│   └── diagnostics/    # human-friendly error messages
├── runtime/            # C runtime (GC, memory, panic)
├── stdlib/             # standard library in Riner itself
├── tests/
│   ├── unit/
│   ├── integration/
│   └── conformance/    # official language test suite
├── docs/
│   ├── spec.md         # language specification
│   └── grammar.ebnf    # formal grammar
├── examples/           # sample programs written in Riner
├── go.mod
├── Makefile
└── LICENSE
```

## Building from source

Prerequisites: Go 1.22+ and GCC or Clang (for the C backend, phase 3+).

```bash
git clone https://github.com/yourusername/riner.git
cd riner
go mod tidy
make build
```

Run a file:

```bash
./riner run examples/hello.rn
```

## Contributing

Riner is open source under the MIT License. Contributions, issues, and ideas are welcome.

If you find a bug, open an issue. If you want to contribute code, open a pull request. Keep it focused and in line with the project philosophy.

## License

MIT. Do whatever you want with it, just keep the credit.
