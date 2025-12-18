## **STRUKTUR FOLDER PROJECT**

```
morph/
├── cmd/
│   ├── morph/              # Main compiler executable
│   │   └── main.go
│   ├── morphi/             # Interactive REPL
│   │   └── main.go
│   └── morphfmt/           # Code formatter (nanti)
│       └── main.go
│
├── pkg/
│   ├── lexer/              # Tokenizer
│   │   ├── lexer.go
│   │   ├── lexer_test.go
│   │   ├── token.go
│   │   └── token_test.go
│   │
│   ├── parser/             # AST builder
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   ├── ast.go
│   │   └── ast_test.go
│   │
│   ├── analysis/           # Static Analysis & Context
│   │   ├── analyzer.go
│   │   └── context.go
│   │
│   ├── evaluator/          # Tree-walking interpreter (MVP)
│   │   ├── evaluator.go
│   │   ├── evaluator_test.go
│   │   ├── environment.go
│   │   └── builtins.go
│   │
│   ├── compiler/           # Bytecode compiler (Phase 2)
│   │   ├── compiler.go
│   │   ├── compiler_test.go
│   │   ├── symbol_table.go
│   │   └── opcodes.go
│   │
│   ├── vm/                 # Virtual machine (Phase 2)
│   │   ├── vm.go
│   │   ├── vm_test.go
│   │   ├── stack.go
│   │   └── frame.go
│   │
│   ├── object/             # Runtime object system
│   │   ├── object.go
│   │   ├── environment.go
│   │   └── builtins.go
│   │
│   └── stdlib/             # Standard library (Phase 3)
│       ├── io/
│       │   └── ...
│       ├── string/
│       │   └── ...
│       └── math/
│           └── ...
│
├── test/
│   ├── fixtures/           # Test morph programs
│   │   ├── valid/
│   │   │   ├── hello.fox
│   │   │   ├── fibonacci.fox
│   │   │   └── ...
│   │   └── invalid/
│   │       ├── syntax_error_1.fox
│   │       └── ...
│   │
│   └── integration/        # Integration tests
│       └── ...
│
├── examples/               # Example programs
│   ├── hello_world.fox
│   ├── calculator.fox
│   └── ...
│
├── docs/
│   ├── specification.md   # Language specification (The Standard)
│   ├── grammar.md         # Formal grammar (EBNF)
│   ├── tutorial.md        # Getting started
│   └── stdlib.md          # Standard library docs
│
├── scripts/
│   ├── test.sh            # Run all tests
│   ├── benchmark.sh       # Performance benchmarks
│   └── build.sh           # Build binaries
│
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── CHANGELOG.md
```

---

## **TODO LIST - REALISTIC PHASES**

### **PHASE 0: Project Setup (Week 1)**

**Deliverables:**
- [x] Initialize Go module (`go mod init github.com/VzoelFox/morphlang`)
- [x] Create folder structure (sesuai di atas)
- [x] Setup `.gitignore` (Go standard + IDE files)
- [x] Write `README.md` (project description, build instructions)
- [x] Write `docs/specification.md` (Complete Language Spec)
  - [x] Syntax examples
  - [x] Type system
  - [x] Control flow
  - [x] Function declaration
  - [x] Scoping rules
  - [x] Bytecode Standard
- [x] Setup CI/CD (GitHub Actions untuk run tests)

**Output:** Empty project structure dengan spec document

---

### **PHASE 1: Lexer (Week 2-3)**

**Patch 1.1: Token Definitions**
- [x] Define `TokenType` enum (`pkg/lexer/token.go`)
  - [x] Keywords (fungsi, kembalikan, jika, etc)
  - [x] Literals (angka, teks, benar/salah)
  - [x] Operators (+, -, *, /, ==, !=, etc)
  - [x] Delimiters ({, }, (, ), [, ], etc)
- [x] Write 20 test cases untuk token types

**Patch 1.2: Basic Lexer**
- [x] Implement `Lexer` struct (`pkg/lexer/lexer.go`)
- [x] Implement `NextToken()` method
- [x] Handle whitespace, comments
- [x] Write 30 test cases (identifiers, numbers, strings)

**Patch 1.3: Advanced Lexing**
- [x] Handle multi-character operators (==, !=, <=, >=)
- [x] Handle string escapes (\n, \t, \", etc)
- [x] Error reporting (line/column numbers)
- [x] Write 30 test cases (edge cases, errors)

**Deliverables:**
- [x] 80+ passing lexer tests
- [x] Can tokenize all valid Morph syntax

---

### **PHASE 2: Parser (Week 4-6)**

**Patch 2.1: AST Definitions**
- [x] Define AST node interfaces (`pkg/parser/ast.go`)
  - [x] `Expression` interface
  - [x] `Statement` interface
  - [x] `Program` node (root)
- [x] Implement concrete nodes:
  - [x] `IntegerLiteral`, `StringLiteral`, `BooleanLiteral`
  - [x] `Identifier`
  - [x] `BinaryExpression` (+, -, *, /, etc)
- [x] Write 15 test cases (AST node creation)

**Patch 2.2: Expression Parser**
- [x] Implement `Parser` struct (`pkg/parser/parser.go`)
- [x] Implement Pratt parser untuk expressions
  - [x] Precedence table
  - [x] Prefix parsers (literals, identifiers, -x, !x)
  - [x] Infix parsers (binary ops, function calls)
- [x] Write 40 test cases (expressions)

**Patch 2.3: Statement Parser**
- [x] Parse variable declarations (`var x = 10`)
- [x] Parse assignments (`x = 20`)
- [x] Parse return statements
- [x] Parse if/else
- [x] Parse while loops
- [x] Write 40 test cases (statements)

**Patch 2.4: Function Parser**
- [x] Parse function declarations
- [x] Parse function calls
- [x] Parse blocks `{ ... }`
- [x] Write 30 test cases (functions)

**Patch 2.5: Error Recovery & Strictness**
- [x] Implement error reporting (line, column, message)
- [x] Add panic mode recovery
- [x] Write 20 test cases (syntax errors)
- [x] **Strict Whitespace & Interpolation (Refactored)**

**Deliverables:**
- [x] 145+ passing parser tests
- [x] Can parse all valid Morph syntax
- [x] Clear error messages untuk invalid syntax

---

### **PHASE 3: Tree-Walking Interpreter (Week 7-9)**

**Patch 3.1: Object System**
- [x] Define `Object` interface (`pkg/object/object.go`)
- [x] Implement types:
  - [x] `Integer`, `String`, `Boolean`
  - [x] `Function`
  - [x] `Null`, `Error`
- [x] Write 15 test cases

**Patch 3.2: Environment**
- [x] Implement `Environment` (variable storage)
- [x] Handle scoping (nested environments)
- [x] Write 20 test cases

**Patch 3.3: Expression Evaluator**
- [x] Implement `Eval()` untuk literals
- [x] Implement binary operators (+, -, *, /, ==, !=, etc)
- [x] Implement unary operators (-, !)
- [x] Write 40 test cases

**Patch 3.4: Statement Evaluator**
- [x] Eval variable declarations (Implicit via assignment)
- [x] Eval assignments
- [x] Eval return statements
- [x] Eval if/else
- [x] Eval while loops
- [x] Write 40 test cases

**Patch 3.5: Function Calls**
- [x] Eval function declarations
- [x] Eval function calls
- [x] Handle parameters dan arguments
- [x] Handle recursion
- [x] Write 30 test cases

**Patch 3.6: Built-in Functions**
- [x] Implement `cetak()` (print)
- [x] Implement `panjang()` (length)
- [x] Implement `tipe()` (type)
- [x] Write 15 test cases

**Deliverables:**
- [x] 160+ passing evaluator tests
- [x] Can run simple Morph programs
- [x] REPL working (`cmd/morphi`)

---

### **PHASE 4: Integration & Polish (Week 10)**

**Patch 4.1: End-to-End Testing**
- [ ] Write 20 integration tests (`test/integration/`)
- [ ] Test full programs dari `test/fixtures/valid/`
- [ ] Verify error handling untuk `test/fixtures/invalid/`

**Patch 4.2: CLI Tool (Accelerated)**
- [x] Implement `cmd/morph/main.go` (Basic)
  - [x] Read file
  - [x] Lex → Parse → Analyze
  - [x] Generate `.fox.vz` context file
- [ ] Write usage documentation

**Patch 4.3: Examples**
- [ ] Write `hello_world.fox`
- [ ] Write `fibonacci.fox`
- [ ] Write `calculator.fox`
- [ ] Write `faktorial.fox`
- [ ] Write `sorting.fox`

**Patch 4.4: Documentation**
- [ ] Write `docs/tutorial.md` (getting started guide)
- [x] Write `docs/grammar.md` (formal EBNF grammar)
- [ ] Update `README.md` (install, usage, examples)

**Deliverables:**
- [ ] 200+ passing tests total
- [ ] Working compiler executable
- [ ] Working REPL
- [ ] 5+ example programs
- [ ] Complete documentation

---

### **PHASE 5: Bytecode VM (Week 11-16)**

**Patch 5.1: Opcode Definitions**
- [x] Define bytecode instructions (`pkg/compiler/opcodes.go`)
  - [x] LOAD_CONST, LOAD_VAR, STORE_VAR
  - [x] ADD, SUB, MUL, DIV
  - [x] JUMP, JUMP_IF_FALSE
  - [x] CALL, RETURN
- [x] Write opcode tests

**Patch 5.2: Symbol Table**
- [x] Implement symbol table (variable → index mapping)
- [x] Handle scopes
- [x] Write 20 test cases

**Patch 5.3: Compiler**
- [x] Implement AST → bytecode compiler
- [x] Compile expressions
- [x] Compile statements
- [x] Compile functions
- [x] Write 50 test cases

**Patch 5.4: Virtual Machine**
- [x] Implement stack-based VM
- [x] Implement instruction dispatch
- [x] Handle function calls (call frames)
- [x] Write 50 test cases

**Patch 5.5: Integration**
- [x] Wire compiler + VM into CLI
- [ ] Benchmark vs tree-walking interpreter
- [ ] Write 30 integration tests

**Deliverables:**
- [x] 150+ compiler/VM tests
- [x] 5-10x performance improvement over interpreter
- [x] Backward compatible (same CLI interface)

---

### **PHASE 6: COTC (Core of The Core) - Standard Library**

**Concept:** COTC is the foundational library written in Morph (`.fox`), serving as the "Standard Library" for the language.

**Patch 6.1: I/O Module**
- [x] `baca_file(path)` - read file
- [x] `tulis_file(path, content)` - write file
- [x] `input(prompt)` - read user input

**Patch 6.2: String Module**
- [x] `pisah(str, delim)` - split
- [x] `gabung(list, delim)` - join
- [x] `huruf_besar(str)` - uppercase
- [x] `huruf_kecil(str)` - lowercase

**Patch 6.3: Math Module**
- [x] `abs(x)`, `max(a, b)`, `min(a, b)`
- [x] `pow(x, y)`, `sqrt(x)`

**Deliverables:**
- [x] 50+ stdlib tests (Integrated in VM tests)
- [ ] Documented stdlib API

---

### **PHASE X: Deterministic Execution Model (The "OS" Vision)**

**Concept:** Transforming Morph from a simple VM into a deterministic execution engine with memory ownership, snapshot/rollback, and FIFO scheduling.

**Migration Strategy (Krusial):**
1. **Tahap 1 (Hybrid):** VM tetap menggunakan interface Object Go, tapi di dalamnya, data primitif (int, float) mulai dialokasikan ke Custom Allocator (pkg/memory).
2. **Tahap 2 (Complex Types):** String dan Array mulai dipindah ke Custom Allocator.
3. **Tahap 3 (Full Swap):** Melepas ketergantungan pada Go GC sepenuhnya.

**Patch X.1: Hybrid Integration (Tahap 1)**
- [x] Fix Memory Leak: Link `object.Integer` to `memory.Ptr`.
- [x] Implement `AllocFloat` in `pkg/memory`.
- [x] Implement `DrawerLease` struct (`pkg/memory/lease.go`).
- [x] Implement `AcquireDrawer(unitID)` - Exclusive ownership.
- [x] Integrate with `Cabinet` structure.

**Patch X.2: Complex Types & Snapshot (Tahap 2)**
- [x] Migrate String to Custom Allocator.
- [x] Migrate Array to Custom Allocator.
- [ ] Implement `OP_SNAPSHOT`, `OP_ROLLBACK`, `OP_COMMIT`.
- [ ] Update VM to support State Checkpointing.

**Patch X.3: Full Swap & Scheduler (Tahap 3)**
- [ ] Remove Go GC dependency.
- [ ] Implement FIFO Queue mechanism.
- [ ] Implement Atomic Shard Assignment (CAS).
- [ ] Implement Worker Units logic (Morph Routine).

---

## **PATCH VERSIONING**

```
v0.1.0 - Phase 0 complete (Setup)
v0.2.0 - Phase 1 complete (Lexer)
v0.3.0 - Phase 2 complete (Parser)
v0.4.0 - Phase 3 complete (Interpreter MVP) ← MILESTONE
v0.5.0 - Phase 4 complete (Polish & Docs)
v1.0.0 - Phase 5 complete (Bytecode VM) ← PUBLIC RELEASE
v1.1.0 - Phase 6 complete (Stdlib)
v1.2.0 - Phase X.1 complete (Drawer Lease)
```

---


# Founder Vzoel Fox's 
