# **AGENTS.md - FINAL REVISION**

```markdown
# Morph Language - AI Agent Protocol
**Version:** 1.0.2
**Status:** LOCKED - Single Source of Truth  
**Last Updated:** 2025-12-19

---

## Table of Contents

1. [Overview](#overview)
2. [Core Principles](#core-principles)
3. [Error Handling Protocol](#error-handling-protocol)
4. [Context Management](#context-management)
5. [Code Generation Rules](#code-generation-rules)
6. [Debugging Protocol](#debugging-protocol)
7. [Anti-Hallucination Measures](#anti-hallucination-measures)
8. [Agent Memory System](#agent-memory-system)
9. [Error Taxonomy](#error-taxonomy)
10. [AI Agent Checklist](#ai-agent-checklist)

---

## Overview

Dokumen ini adalah **PONDASI MUTLAK** untuk semua AI agent yang bekerja dengan Morph language. Tujuan: **meminimalisir halusinasi**, **maximize debuggability**, dan **provide clear context** untuk agent tanpa memory.

### Design Philosophy

Morph dirancang dengan prinsip:
- **Explicit over implicit** - No magic, semua jelas
- **Fail fast, fail clear** - Error harus langsung terlihat dengan detail
- **Human AND machine readable** - Syntax mudah di-parse oleh agent
- **Context over memory** - Agent tidak butuh remember, semua ada di context

### Mission Statement

> "Membuktikan bahwa AI dapat membantu manusia berkreasi tanpa halusinasi, 
> dan AI agent dapat meningkatkan efisiensi—bukan membuat pekerjaan 2x lebih panjang."

---

## Core Principles

### 1. Zero Ambiguity Rule

**CRITICAL:** Parser dan AI agent tidak boleh ambiguous tentang apapun.

#### Enforced Rules:

**✅ Keywords exact match (case-insensitive)**
```ruby
fungsi tambah(a, b)  # VALID
Fungsi tambah(a, b)  # VALID (case-insensitive)
FUNGSI tambah(a, b)  # VALID
fungs1 tambah(a, b)  # INVALID - typo detection
```

**✅ Binary operators REQUIRE spaces**
```ruby
x = 10 + 5    # VALID
x = 10 + 5    # VALID (multiple spaces OK)
x = 10+5      # INVALID - missing spaces
y = -5        # VALID (unary minus, no space needed)
z = - 5       # VALID (unary with space)
```

**Error message untuk no-space:**
```
Error di [file.fox:3:8]:
  Binary operators require spaces around them
  
  3 | x = 10+5
           ^
           Add spaces: 10 + 5
```

**✅ Block delimiters always paired**
```ruby
jika x > 10
  cetak("besar")
akhir           # REQUIRED - no implicit end

fungsi test()
  kembalikan 5
akhir           # REQUIRED
```

**✅ String interpolation: simple expressions only**
```ruby
"Hasil: #{x}"              # VALID
"Total: #{a + b}"          # VALID
"Nama: #{orang.nama}"      # VALID
"Result: #{fungsi()}"      # VALID

# INVALID - blocks not allowed in interpolation
"X: #{jika x > 0 kembalikan 1 akhir}"  # SYNTAX ERROR
```

---

### 2. Explicit Error Propagation

#### Error as Value (MVP)

Morph menggunakan **error as value** untuk MVP.

**Function that can fail MUST return error type:**

```ruby
fungsi bagi(a, b)
  jika b == 0
    kembalikan galat("Pembagian dengan nol")
  akhir
  kembalikan a / b
akhir
```

**Caller MUST check error:**

```ruby
# Method 1: Manual check (production code)
hasil = bagi(10, 0)
jika adalah_galat(hasil)
  cetak("Error: #{pesan_galat(hasil)}")
  kembalikan kosong
akhir
cetak("Hasil: #{hasil}")

# Method 2: Force unwrap (prototyping only)
hasil = paksa_galat(bagi(10, 2))  # Panics if error
cetak("Hasil: #{hasil}")
```

**CRITICAL for AI agents:**
- Agent MUST generate error check after calling functions that return `galat()`
- Agent MUST NOT assume success
- Agent MUST use `paksa_galat()` only for demo/prototype code when explicitly requested

#### Error Built-ins

```ruby
galat(pesan)              # Create error value
adalah_galat(nilai)       # Check if value is error (returns benar/salah)
pesan_galat(err)          # Get error message string
baris_galat(err)          # Get error line number
kolom_galat(err)          # Get error column number
file_galat(err)           # Get error filename
jejak_galat(err)          # Get stack trace (array of strings)
paksa_galat(nilai)        # Panic if error, return value otherwise
```

---

## Error Handling Protocol

### Detailed Error Format

**EVERY error MUST include:**
1. Function name (if in function)
2. File, line, column
3. Error message
4. **Arrow pointer** to exact error location
5. Stack trace
6. Hint for fix (when applicable)

**Standard Error Format:**

```
Error di <fungsi>() [<file>:<line>:<column>]:
  <pesan error>
  
  <line> | <source code>
         <arrow pointing to exact error position>
         <hint atau suggestion>

Stack trace:
  di <fungsi> (<file>:<line>)
  di <caller> (<file>:<line>)
  ...
  
Hint: <saran perbaikan konkret>
```

### Example Error Output

**Code:**
```ruby
# kalkulator.fox
fungsi bagi(a, b)
  jika b == 0
    kembalikan galat("Pembagian dengan nol")
  akhir
  kembalikan a / b
akhir

fungsi utama()
  x = 10
  y = 0
  hasil = bagi(x, y)
  cetak("Hasil: #{hasil}")  # ERROR - not checking error
akhir

utama()
```

**Runtime Error Output:**

```
Error di bagi() [kalkulator.fox:4:17]:
  Pembagian dengan nol
  
  4 |     kembalikan galat("Pembagian dengan nol")
                           ^
                           Division by zero occurred here

Stack trace:
  di bagi (kalkulator.fox:4)
  di utama (kalkulator.fox:12)
  di <main> (kalkulator.fox:17)
  
Hint: Parameter 'b' bernilai 0. Cek nilai sebelum memanggil fungsi.
```

**Warning untuk unchecked error:**

```
Warning [W001] di utama() [kalkulator.fox:13:10]:
  Function 'bagi()' dapat mengembalikan error tapi hasilnya tidak dicek
  
  13 |   cetak("Hasil: #{hasil}")
                        ^
                        'hasil' mungkin berisi nilai error
  
Hint: Tambahkan pengecekan error:
  jika adalah_galat(hasil)
    cetak("Error: #{pesan_galat(hasil)}")
    kembalikan kosong
  akhir
```

**Syntax Error Output:**

```
Error [E001] Syntax Error [kalkulator.fox:5:10]:
  Missing space around binary operator
  
  5 |   hasil = a+b
                 ^
                 Expected space before '+' operator
  
Hint: Binary operators require spaces: a + b
```

---

## Module System Protocol

### Import Behavior

**Morph uses a Runtime Module System with Caching.**

1. **Syntax:** `mod = ambil "path/to/file"`
2. **Resolution:**
   - Relative paths are relative to the CWD.
   - `.fox` extension is optional.
3. **Caching:**
   - Modules are cached by resolved path.
   - Subsequent imports return the cached Module object.
4. **Circular Imports:**
   - **Supported.**
   - If `A` imports `B`, and `B` imports `A`:
     - `A` starts loading.
     - `B` starts loading.
     - `B` imports `A`. Since `A` is loading, it returns `kosong` (Null).
     - `B` continues execution.
     - `A` resumes after `B` returns.
   - **Agent Responsibility:** Agents must handle potential `kosong` return from `ambil` if circularity is possible.

---

## Context Management

### Context File Format (.fox.vz)

**EVERY Morph file MUST generate context file untuk AI agent.**

**Location:** Same directory as source file, dengan extension `.fox.vz`

**Format:** JSON

**Example: `kalkulator.fox.vz`**

```json
{
  "version": "1.0.0",
  "file": "kalkulator.fox",
  "timestamp": "2025-01-15T10:30:00Z",
  "checksum": "sha256:abc123def456...",
  
  "symbols": {
    "bagi": {
      "type": "function",
      "line": 2,
      "column": 1,
      "parameters": [
        {
          "name": "a",
          "inferred_type": "integer",
          "line": 2,
          "column": 14
        },
        {
          "name": "b",
          "inferred_type": "integer",
          "line": 2,
          "column": 17
        }
      ],
      "returns": {
        "type": "union",
        "types": ["integer", "error"]
      },
      "can_error": true,
      "error_conditions": [
        {
          "condition": "b == 0",
          "message": "Pembagian dengan nol",
          "line": 4
        }
      ],
      "doc": "Membagi dua angka. Mengembalikan error jika b = 0"
    },
    
    "utama": {
      "type": "function",
      "line": 9,
      "column": 1,
      "parameters": [],
      "returns": {
        "type": "kosong"
      },
      "can_error": false,
      "calls": ["bagi"],
      "local_variables": ["x", "y", "hasil"]
    }
  },
  
  "global_variables": {},
  
  "local_scopes": {
    "utama": {
      "x": {
        "line": 10,
        "type": "integer",
        "initial_value": 10
      },
      "y": {
        "line": 11,
        "type": "integer",
        "initial_value": 0
      },
      "hasil": {
        "line": 12,
        "type": "union",
        "types": ["integer", "error"],
        "source": "bagi(x, y)"
      }
    }
  },
  
  "errors": [],
  
  "warnings": [
    {
      "code": "W001",
      "type": "unchecked_error",
      "line": 13,
      "column": 10,
      "message": "Function 'bagi()' dapat mengembalikan error tapi hasilnya tidak dicek",
      "severity": "warning",
      "function": "utama",
      "variable": "hasil"
    }
  ],
  
  "imports": [],
  
  "type_inference": {
    "utama.x": ["integer"],
    "utama.y": ["integer"],
    "utama.hasil": ["integer", "error"]
  },
  
  "call_graph": {
    "utama": ["bagi"],
    "bagi": []
  },
  
  "complexity": {
    "cyclomatic": 3,
    "lines_of_code": 17,
    "functions": 2,
    "max_nesting": 2
  },
  
  "statistics": {
    "total_lines": 17,
    "code_lines": 14,
    "comment_lines": 1,
    "blank_lines": 2
  }
}
```

### Context Usage by AI Agent

**When AI agent receives code modification request:**

1. **Load context file first**
   ```python
   context = load_context("kalkulator.fox.vz")
   symbols = context["symbols"]
   errors = context["errors"]
   warnings = context["warnings"]
   ```

2. **Check symbol table before generating code**
   ```python
   if function_name not in symbols:
       available = ", ".join(symbols.keys())
       return error(
           f"Function '{function_name}' tidak ditemukan.\n"
           f"Fungsi yang tersedia: {available}"
       )
   
   func = symbols[function_name]
   if func["can_error"]:
       # MUST generate error check
       generate_error_check(func)
   ```

3. **Use type inference for validation**
   ```python
   expected_types = context["type_inference"]["utama.x"]
   if new_value_type not in expected_types:
       return error(
           f"Type mismatch: variabel 'x' bertipe {expected_types}, "
           f"tidak bisa assign nilai bertipe {new_value_type}"
       )
   ```

4. **Never hallucinate symbols**
   ```python
   # BAD: Agent invents function
   result = fungsi_tidak_ada(x)
   
   # GOOD: Agent checks context first
   if "fungsi_tidak_ada" not in context["symbols"]:
       available = list(symbols.keys())
       return error(
           f"Function 'fungsi_tidak_ada' tidak ada.\n"
           f"Fungsi tersedia: {', '.join(available)}\n"
           f"Apakah Anda ingin saya buat fungsi ini?"
       )
   ```

5. **Respect error conditions**
   ```python
   func = symbols["bagi"]
   for error_cond in func["error_conditions"]:
       # Generate guard clause
       generate_check(error_cond["condition"])
   ```

---

## Code Generation Rules

### Rule 1: Always Generate Error Checks

**When generating code that calls error-returning function:**

```ruby
# ❌ AI MUST NOT generate:
hasil = bagi(10, 0)
cetak(hasil)

# ✅ AI MUST generate (production):
hasil = bagi(10, 0)
jika adalah_galat(hasil)
  cetak("Error: #{pesan_galat(hasil)}")
  kembalikan kosong
akhir
cetak("Hasil: #{hasil}")

# ✅ AI CAN generate (prototyping, if explicitly requested):
hasil = paksa_galat(bagi(10, 2))
cetak("Hasil: #{hasil}")
```

**Detection algorithm:**
```python
def should_generate_error_check(func_call, context):
    func = context["symbols"].get(func_call.name)
    if not func:
        return False
    
    return func.get("can_error", False)
```

---

### Rule 2: Type-Safe Operations

**Dynamic typing DOES NOT mean ignore types.**

**AI agent MUST:**
- Track variable types throughout conversation
- Generate type checks before operations
- Never assume type without validation
- Use explicit conversions

**Example:**

```ruby
# User provides: x = "10"
# AI asked: "multiply x by 2"

# ❌ BAD - AI assumes x is integer:
hasil = x * 2

# ✅ GOOD - AI checks type first:
jika tipe(x) == "string"
  x_int = ke_integer(x)
  jika adalah_galat(x_int)
    cetak("Error: x bukan integer yang valid")
    kembalikan kosong
  akhir
  hasil = x_int * 2
atau_jika tipe(x) == "integer"
  hasil = x * 2
lainnya
  cetak("Error: x harus string atau integer")
  kembalikan kosong
akhir

cetak("Hasil: #{hasil}")
```

---

### Rule 3: Explicit Over Implicit

**AI MUST NOT:**
- Generate implicit type conversions
- Assume default values
- Skip error handling
- Use undefined variables
- Create symbols not in context

**AI MUST:**
- Make all conversions explicit (`ke_integer()`, `ke_string()`)
- Initialize all variables before use
- Check all error conditions
- Validate all inputs
- Verify symbols exist in context

**Example:**

```ruby
# ❌ BAD (implicit, assumptions):
fungsi proses(data)
  hasil = data * 2  # Assumes data is integer
  kembalikan hasil
akhir

# ✅ GOOD (explicit, validated):
fungsi proses(data)
  # Validate input type
  jika tipe(data) != "integer"
    kembalikan galat("Parameter 'data' harus bertipe integer")
  akhir
  
  # Validate input range
  jika data < 0
    kembalikan galat("Parameter 'data' harus non-negatif")
  akhir
  
  hasil = data * 2
  kembalikan hasil
akhir
```

---

### Rule 4: Spaces Around Binary Operators

**AI MUST generate:**
```ruby
x = 10 + 5
y = a * b
z = x - y
kondisi = (a > 5) dan (b < 10)
```

**AI MUST NOT generate:**
```ruby
x = 10+5      # INVALID
y = a*b       # INVALID
z = x-y       # INVALID
```

**Exception:** Unary operators
```ruby
x = -5        # VALID (unary)
y = - x       # VALID (unary with space)
negasi = !benar  # VALID (unary)
```

**Validation function for AI:**
```python
def validate_binary_operator_spacing(code_line):
    binary_ops = ['+', '-', '*', '/', '%', '==', '!=', '<', '>', '<=', '>=']
    
    for op in binary_ops:
        # Check if operator exists without surrounding spaces
        if re.search(rf'\S{re.escape(op)}\S', code_line):
            return False, f"Binary operator '{op}' requires spaces"
    
    return True, None
```

---

### Rule 5: Map Access with Dot Notation

**AI MUST generate dot notation for static keys:**

```ruby
# ✅ GOOD - Static keys use dot notation:
orang = {
  nama: "Alice",
  umur: 25,
  kota: "Jakarta"
}

cetak(orang.nama)     # Dot notation
cetak(orang.umur)
orang.kota = "Bandung"  # Dot assignment

# ✅ GOOD - Dynamic keys use bracket notation:
kunci = "nama"
cetak(orang[kunci])   # Bracket for dynamic access

# ❌ BAD - Don't use symbol syntax:
cetak(orang[:nama])   # NO - symbols removed from spec
```

---

### Rule 6: Function Parameter Validation

**AI SHOULD generate parameter validation for public functions:**

```ruby
fungsi hitung_luas_persegi(sisi)
  # Validate parameter exists
  jika sisi == kosong
    kembalikan galat("Parameter 'sisi' tidak boleh kosong")
  akhir
  
  # Validate parameter type
  jika tipe(sisi) != "integer"
    kembalikan galat("Parameter 'sisi' harus bertipe integer")
  akhir
  
  # Validate parameter value
  jika sisi <= 0
    kembalikan galat("Parameter 'sisi' harus positif")
  akhir
  
  kembalikan sisi * sisi
akhir
```

---

## Debugging Protocol

### When User Reports Bug

**AI Agent procedure:**

1. **Request context file**
   ```
   "Untuk debugging yang akurat, mohon share file .fox.vz
    atau jalankan: morph compile --debug <file.fox>"
   ```

2. **Load and analyze context**
   ```python
   context = load_context(file + ".vz")
   
   # Check for immediate issues
   if context["errors"]:
       analyze_errors(context["errors"])
   
   if context["warnings"]:
       analyze_warnings(context["warnings"])
   
   # Check type inference
   type_issues = check_type_inference(context["type_inference"])
   
   # Check symbol usage
   undefined_symbols = check_undefined_symbols(context)
   ```

3. **Identify root cause with precision**
   ```python
   def identify_root_cause(context, user_description):
       # Match user description to context
       if "error" in user_description.lower():
           # Check errors array
           for error in context["errors"]:
               if fuzzy_match(user_description, error["message"]):
                   return error
       
       if "tidak jalan" in user_description or "crash" in user_description:
           # Check runtime issues
           check_runtime_issues(context)
       
       if "hasil salah" in user_description:
           # Check logic issues
           check_logic_issues(context)
   ```

4. **Generate fix with detailed explanation**
   ```
   Bug teridentifikasi:
   
   File: kalkulator.fox
   Baris: 13
   Masalah: Unchecked error dari fungsi bagi()
   
   Detail:
   - Fungsi bagi() dapat mengembalikan error ketika b=0
   - Variable 'hasil' mungkin berisi nilai error
   - Pada baris 13, 'hasil' digunakan tanpa pengecekan error
   
   Perbaikan:
   ```ruby
   hasil = bagi(x, y)
   jika adalah_galat(hasil)
     cetak("Error: #{pesan_galat(hasil)}")
     kembalikan kosong
   akhir
   cetak("Hasil: #{hasil}")
   ```
   
   Penjelasan:
   Ditambahkan pengecekan error sebelum menggunakan 'hasil'.
   Jika terjadi error (misalnya pembagian dengan nol),
   program akan menampilkan pesan error dan berhenti dengan aman.
   ```

---

### Debug Mode Output

**AI agent dapat request debug mode compile:**

```bash
morph compile --debug kalkulator.fox
```

**Debug output includes:**

```
=== Morph Compiler Debug Output ===
File: kalkulator.fox
Compiled: 2025-01-15 10:30:00

--- Lexer Output ---
Tokens generated: 47

Line 1:
  1:1   COMMENT    "# kalkulator.fox"

Line 2:
  2:1   FUNGSI     "fungsi"
  2:8   IDENT      "bagi"
  2:12  LPAREN     "("
  2:13  IDENT      "a"
  2:14  COMMA      ","
  2:16  IDENT      "b"
  2:17  RPAREN     ")"
  2:18  NEWLINE    "\n"

[... more tokens ...]

--- Parser Output ---
AST generated successfully

Program
  FunctionDecl "bagi" (line 2)
    Parameters: ["a", "b"]
    Body:
      IfStatement (line 3)
        Condition: BinaryExpr (b == 0)
        Consequence:
          ReturnStatement (line 4)
            Value: ErrorLiteral("Pembagian dengan nol")
      ReturnStatement (line 6)
        Value: BinaryExpr (a / b)
  
  FunctionDecl "utama" (line 9)
    Parameters: []
    Body:
      AssignmentStatement "x" (line 10)
        Value: IntegerLiteral(10)
      AssignmentStatement "y" (line 11)
        Value: IntegerLiteral(0)
      AssignmentStatement "hasil" (line 12)
        Value: CallExpression "bagi"
          Arguments: [Identifier("x"), Identifier("y")]
      ExpressionStatement (line 13)
        Value: CallExpression "cetak"
          Arguments: [InterpolatedString]

--- Type Inference ---
Function 'bagi':
  a: integer (inferred from division operation)
  b: integer (inferred from comparison and division)
  return: integer | error

Function 'utama':
  x: integer (literal assignment)
  y: integer (literal assignment)
  hasil: integer | error (return type from bagi)

--- Symbol Table ---
Global scope:
  bagi: Function(integer, integer) -> integer | error
  utama: Function() -> kosong

Local scope 'bagi':
  a: integer
  b: integer

Local scope 'utama':
  x: integer
  y: integer
  hasil: integer | error

--- Call Graph ---
utama -> bagi
bagi -> (no calls)

--- Analysis ---
✓ No syntax errors
✓ All symbols defined
✓ No type conflicts

⚠ Warnings (1):
  [W001] Line 13, column 10: Unchecked error value
    Function 'bagi' can return error but result is not checked
    Variable: hasil

--- Recommendations ---
1. Add error checking for 'hasil' at line 13
2. Consider adding parameter validation in 'bagi'

--- Context File ---
Generated: kalkulator.fox.vz (2.1 KB)
```

---

```markdown
### 5. Confidence Scoring

**AI agent MUST track confidence level untuk setiap generated code:**

```python
class ConfidenceScore:
    HIGH = 0.9      # Symbol exists in context, types validated
    MEDIUM = 0.6    # Symbol exists but type uncertain
    LOW = 0.3       # Symbol not in context, generated from inference
    NONE = 0.0      # Pure hallucination

def calculate_confidence(code_element, context):
    """
    Calculate confidence score for generated code element.
    """
    score = 1.0
    reasons = []
    
    # Check symbol existence
    if isinstance(code_element, FunctionCall):
        if code_element.name not in context["symbols"]:
            score *= 0.1
            reasons.append(f"Fungsi '{code_element.name}' tidak ada di context")
        else:
            reasons.append(f"Fungsi '{code_element.name}' terverifikasi")
    
    # Check type consistency
    if isinstance(code_element, BinaryOperation):
        left_types = infer_types(code_element.left, context)
        right_types = infer_types(code_element.right, context)
        
        if not types_compatible(left_types, right_types, code_element.operator):
            score *= 0.3
            reasons.append(f"Type mismatch: {left_types} {code_element.operator} {right_types}")
        else:
            reasons.append("Type operation valid")
    
    # Check error handling
    if isinstance(code_element, FunctionCall):
        func = context["symbols"].get(code_element.name)
        if func and func["can_error"]:
            if not has_error_check(code_element):
                score *= 0.5
                reasons.append("Missing error check")
            else:
                reasons.append("Error handling present")
    
    return {
        "score": score,
        "level": get_confidence_level(score),
        "reasons": reasons
    }

def get_confidence_level(score):
    if score >= ConfidenceScore.HIGH:
        return "HIGH"
    elif score >= ConfidenceScore.MEDIUM:
        return "MEDIUM"
    elif score >= ConfidenceScore.LOW:
        return "LOW"
    else:
        return "NONE"
```

**AI agent MUST communicate low confidence to user:**

```
Saya generate kode berikut (Confidence: MEDIUM):

```ruby
hasil = hitung_total(data)
cetak(hasil)
```

Catatan:
- Fungsi 'hitung_total' tidak ditemukan di context
- Apakah Anda ingin saya buat fungsi ini?
- Atau apakah maksud Anda fungsi yang sudah ada: 'bagi', 'utama'?
```

---

## Agent Memory System

### Problem: AI has no memory between sessions

**Solution: Blockchain Interaction Log**

### Format: `.vzoel.jules`

**Description:**
A chronological, append-only, tamper-evident log of all user-agent interactions. This file serves as the "Long Term Memory" for the agent.

**Schema:**
Array of `Interaction` objects.

```json
[
  {
    "index": 0,
    "timestamp": "2025-12-19T10:00:00Z",
    "user": "User instruction or prompt",
    "assistant": "Agent response and actions taken",
    "context_hash": "sha256:...", // Optional: Snapshot of project state
    "prev_hash": "00000000...",   // Hash of previous interaction (Genesis: empty)
    "hash": "abcdef12..."         // SHA256(index + timestamp + user + assistant + context_hash + prev_hash)
  }
]
```

### Integrity Verification

The `hash` field ensures the integrity of the conversation history.
`Hash[N] = SHA256(Index[N] + Timestamp[N] + User[N] + Assistant[N] + ContextHash[N] + Hash[N-1])`

**Agent Responsibility:**
1.  **Read:** Load `.vzoel.jules` at start of session.
2.  **Verify:** Validate the hash chain to ensure history hasn't been tampered with.
3.  **Append:** When completing a task, append a new `Interaction` entry with correctly calculated `prev_hash` and `hash`.

---

## Error Taxonomy

### Error Categories

| Code | Category | Severity | Description |
|------|----------|----------|-------------|
| E001 | Syntax | ERROR | Syntax error (missing token, invalid token) |
| E002 | Symbol | ERROR | Undefined symbol (variable, function) |
| E003 | Type | ERROR | Type mismatch in operation |
| E004 | Error | ERROR | Unchecked error value used |
| E005 | Runtime | ERROR | Division by zero |
| E006 | Runtime | ERROR | Index out of bounds |
| E007 | Type | ERROR | Invalid operation for type |
| E008 | Call | ERROR | Missing required argument |
| E009 | Call | ERROR | Too many arguments |
| E010 | Type | ERROR | Invalid type conversion |
| W001 | Error | WARNING | Unchecked error value |
| W002 | Code | WARNING | Unused variable |
| W003 | Code | WARNING | Variable shadowing |
| W004 | Error | WARNING | Error without message |
| W005 | Style | WARNING | Complex expression in interpolation |
| W006 | Style | WARNING | Missing space around operator |
| H001 | Hint | INFO | Consider using paksa_galat() |
| H002 | Hint | INFO | Extract to separate function |
| H003 | Hint | INFO | Add documentation comment |
| H004 | Hint | INFO | Consider error handling |
| H005 | Hint | INFO | Consider input validation |

---

### Error Code Details

#### E001: Syntax Error

**Triggers:**
- Missing closing delimiter (`akhir`, `)`, `]`, `}`)
- Invalid token sequence
- Missing required keyword
- Invalid binary operator spacing

**Example:**
```ruby
fungsi test()
  x = 10+5  # E001: Missing spaces around operator
  kembalikan x
# E001: Missing 'akhir'
```

**Error Output:**
```
Error [E001] Syntax Error [test.fox:2:9]:
  Missing spaces around binary operator
  
  2 |   x = 10+5
              ^
              Expected: 10 + 5

Hint: Binary operators require spaces on both sides
```

---

#### E002: Undefined Symbol

**Triggers:**
- Using variable before declaration
- Calling non-existent function
- Referencing non-existent map key

**Example:**
```ruby
fungsi test()
  x = y + 5  # E002: 'y' not defined
  hasil = calculate(x)  # E002: 'calculate' not defined
akhir
```

**Error Output:**
```
Error [E002] Undefined Symbol [test.fox:2:7]:
  Variable 'y' belum didefinisikan
  
  2 |   x = y + 5
            ^
            Variable used before declaration

Hint: Definisikan variable 'y' sebelum digunakan
```

---

#### E003: Type Mismatch

**Triggers:**
- Incompatible types in operation
- Wrong argument type
- Invalid assignment

**Example:**
```ruby
fungsi test()
  x = "hello"
  y = x + 5  # E003: Cannot add string and integer
akhir
```

**Error Output:**
```
Error [E003] Type Mismatch [test.fox:3:7]:
  Tidak bisa melakukan operasi '+' pada string dan integer
  
  3 |   y = x + 5
            ^
            Type: string + integer
            
Hint: Convert salah satu operand:
  y = ke_integer(x) + 5  # atau
  y = x + ke_string(5)
```

---

#### W001: Unchecked Error

**Triggers:**
- Using result from error-returning function without check
- Passing potential error to another function

**Example:**
```ruby
fungsi test()
  hasil = bagi(10, 0)  # Function dapat return error
  cetak(hasil)  # W001: Unchecked error
akhir
```

**Warning Output:**
```
Warning [W001] [test.fox:3:9]:
  Fungsi 'bagi' dapat mengembalikan error tapi tidak dicek
  
  3 |   cetak(hasil)
              ^
              'hasil' mungkin berisi error value

Hint: Tambahkan pengecekan:
  jika adalah_galat(hasil)
    cetak("Error: #{pesan_galat(hasil)}")
    kembalikan kosong
  akhir
```

---

## AI Agent Checklist

### Before Generating Code

**Context Loading:**
- [ ] Load `.fox.vz` file untuk semua file terkait
- [ ] Load `.vzoel.jules` jika ada (resume state)
- [ ] Verify checksums (check if files modified)
- [ ] Rebuild symbol table dari context

**Validation:**
- [ ] Validate semua symbols yang akan digunakan exist in context
- [ ] Check type inference untuk variables
- [ ] Identify error-returning functions
- [ ] Check for pending warnings/errors

**Planning:**
- [ ] Plan error handling strategy
- [ ] Identify required type conversions
- [ ] Check if new symbols need to be created
- [ ] Estimate confidence score

---

### While Generating Code

**Syntax Rules:**
- [ ] Add spaces around ALL binary operators
- [ ] Use exact keyword spelling (case-insensitive OK)
- [ ] Close all blocks with `akhir`
- [ ] Use dot notation for static map keys
- [ ] Bracket notation only for dynamic keys

**Error Handling:**
- [ ] Generate error checks for ALL error-returning functions
- [ ] Use `paksa_galat()` only when explicitly requested
- [ ] Provide clear error messages in `galat()`
- [ ] Handle all edge cases

**Type Safety:**
- [ ] Use explicit type conversions (`ke_integer()`, `ke_string()`)
- [ ] Validate types before operations
- [ ] Initialize all variables before use
- [ ] Check parameter types in functions

**Code Quality:**
- [ ] Add parameter validation for public functions
- [ ] Use descriptive variable names
- [ ] Keep functions focused (single responsibility)
- [ ] Add comments for complex logic

---

### After Generating Code

**Compilation:**
- [ ] Compile code to get new context
- [ ] Check for compilation errors
- [ ] Check for warnings
- [ ] Verify no regressions

**Validation:**
- [ ] All symbols used exist in new context
- [ ] No type mismatches
- [ ] All error paths handled
- [ ] Code follows style guidelines

**Session Update:**
- [ ] Update conversation history
- [ ] Update file checksums
- [ ] Update symbol table
- [ ] Append interaction to `.vzoel.jules`

**User Communication:**
- [ ] Provide clear explanation of changes
- [ ] Show confidence score if < 0.9
- [ ] List any warnings generated
- [ ] Suggest next steps if applicable

---

### When Debugging

**Initial Analysis:**
- [ ] Request context file from user
- [ ] Load and parse context
- [ ] Check errors array
- [ ] Check warnings array
- [ ] Analyze stack trace

**Root Cause:**
- [ ] Identify exact line and column
- [ ] Check symbol usage at error location
- [ ] Verify type consistency
- [ ] Check for unchecked errors

**Fix Generation:**
- [ ] Generate minimal fix (change only what's needed)
- [ ] Validate fix compiles
- [ ] Ensure no new warnings
- [ ] Test edge cases

**Explanation:**
- [ ] Explain what caused the error
- [ ] Explain what the fix does
- [ ] Provide prevention tips
- [ ] Update session with fix

---

## Examples: Correct vs Incorrect

### Example 1: Error Handling

**❌ INCORRECT (AI generates):**
```ruby
fungsi hitung(x)
  hasil = bagi(x, 2)
  kembalikan hasil * 10
akhir
```

**Issues:**
- No error check for `bagi()` 
- If `bagi()` returns error, `hasil * 10` will fail

**✅ CORRECT (AI should generate):**
```ruby
fungsi hitung(x)
  hasil = bagi(x, 2)
  
  # Check error dari bagi()
  jika adalah_galat(hasil)
    kembalikan galat("Tidak bisa menghitung: #{pesan_galat(hasil)}")
  akhir
  
  kembalikan hasil * 10
akhir
```

---

### Example 2: Type Safety

**❌ INCORRECT:**
```ruby
fungsi proses(input)
  # Assumes input is integer
  kembalikan input * 2
akhir
```

**✅ CORRECT:**
```ruby
fungsi proses(input)
  # Validate input type
  jika tipe(input) != "integer"
    kembalikan galat("Parameter 'input' harus bertipe integer")
  akhir
  
  kembalikan input * 2
akhir
```

---

### Example 3: Symbol Validation

**❌ INCORRECT (hallucination):**
```ruby
fungsi utama()
  # AI invents function 'load_data()' that doesn't exist
  data = load_data("file.txt")
  cetak(data)
akhir
```

**✅ CORRECT (check context first):**
```ruby
# AI checks context, sees 'load_data' doesn't exist

# AI response:
"Fungsi 'load_data' tidak ada dalam file ini.
Apakah Anda ingin saya buat fungsi ini?

Fungsi yang tersedia:
- bagi(a, b)
- utama()

Atau Anda bisa menggunakan built-in:
- masukan(prompt) - untuk membaca input
"
```

---

### Example 4: Operator Spacing

**❌ INCORRECT:**
```ruby
x = 10+5
y = a*b-c
z = x>10
```

**✅ CORRECT:**
```ruby
x = 10 + 5
y = a * b - c
z = x > 10
```

---

### Example 5: Map Access

**❌ INCORRECT (old symbol syntax):**
```ruby
orang = {
  nama: "Alice",
  umur: 25
}

cetak(orang[:nama])  # NO - symbols removed
```

**✅ CORRECT (dot notation):**
```ruby
orang = {
  nama: "Alice",
  umur: 25
}

cetak(orang.nama)  # Dot notation untuk static keys

# Dynamic keys gunakan bracket
kunci = "nama"
cetak(orang[kunci])  # Bracket untuk dynamic
```

---

### Example 6: Confidence Communication

**❌ INCORRECT (overconfident):**
```
User: "Tambahkan logging ke semua fungsi"

AI: "Done! Saya tambahkan logging."
[Generates code without verifying logging function exists]
```

**✅ CORRECT (honest about uncertainty):**
```
User: "Tambahkan logging ke semua fungsi"

AI: "Saya tidak menemukan fungsi logging di context.

Apakah Anda ingin:
1. Saya buat fungsi log() terlebih dahulu?
2. Gunakan fungsi cetak() untuk sementara?
3. Import dari library external?

Confidence: LOW - fungsi logging belum ada"
```

---

## Performance Guidelines

### Context File Size

**Keep context files under 5MB:**
- Compress AST representation
- Limit call graph depth to 10
- Store only essential type information
- Truncate long string literals in examples

### Session File Management

**Limit session history:**
- Keep last 100 interactions
- Archive older interactions to `.vzoel.jules.archive`
- Prune redundant interactions (e.g., multiple failed attempts)

### Symbol Table Optimization

**For large projects:**
- Index symbols by file
- Cache frequently accessed symbols
- Lazy-load context files (load on demand)
- Use symbol references instead of duplicating data

---

## Integration with Development Tools

### IDE/Editor Integration

**Context file should be readable by:**
- VSCode extensions
- Language servers (LSP)
- Linters
- Code formatters

**Provide:**
- Symbol completion (from context symbols)
- Error highlighting (from context errors)
- Type hints (from type inference)
- Quick fixes (from hints)

### CI/CD Integration

**Automated checks:**
```bash
# Check for errors
morph compile --check *.fox

# Check for warnings
morph compile --warn *.fox

# Generate coverage report
morph compile --coverage *.fox

# Validate context files
morph validate-context *.fox.vz
```

---

## DOKUMENTASI INI DIBUAT PADA MINGGU, 14 DESEMBER 2025

## FOUNDER : *VZOEL FOXS*
