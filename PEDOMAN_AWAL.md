

---

## **KEYWORD MAPPING - BAHASA INDONESIA**

```
fn         → fungsi
return     → kembalikan
if         → jika
elsif      → atauJika (camelCase, satu kata - tidak ambigu)
else       → lainnya
while      → selama
for        → untuk
in         → dalam
true/false → benar/salah
nil        → kosong
end        → akhir
```

**CRITICAL: `atauJika` bukan `atau jika` (satu keyword, tidak ambigu)**

---

## **REVISED SYNTAX - FULL BAHASA INDONESIA**

### **1. VARIABLES**

```ruby
x = 10
nama = "Alice"
umur = 25
```

(Tetap simple - no keyword)

---

### **2. FUNCTIONS**

```ruby
fungsi tambah(a, b)
  kembalikan a + b
akhir

fungsi faktorial(n)
  jika n <= 1
    kembalikan 1
  akhir
  kembalikan n * faktorial(n - 1)
akhir
```

---

### **3. CONTROL FLOW**

**IF/ATAUIJIKA/LAINNYA:**

```ruby
jika x > 10
  cetak("besar")
atauJika x > 5
  cetak("sedang")
lainnya
  cetak("kecil")
akhir
```

**WHILE:**

```ruby
selama x < 10
  x = x + 1
akhir
```

**FOR:**

```ruby
untuk item dalam daftar
  cetak(item)
akhir

untuk i dalam rentang(0, 10)
  cetak(i)
akhir
```

---

### **4. DATA STRUCTURES**

**ARRAYS:**

```ruby
angka = [1, 2, 3, 4, 5]
nama = ["Alice", "Bob", "Charlie"]

cetak(angka[0])    # 1
angka[0] = 10      # mutasi
```

**MAPS:**

```ruby
orang = {
  nama: "Alice",
  umur: 25,
  kota: "Jakarta"
}

cetak(orang[:nama])   # "Alice"
orang[:umur] = 26     # mutasi

# String keys
config = {
  "host": "localhost",
  "port": 8080
}
cetak(config["host"])
```

---

### **5. STRINGS & INTERPOLATION**

```ruby
nama = "Alice"
umur = 25

# Double quotes: interpolasi
salam = "Halo, #{nama}!"
info = "#{nama} berumur #{umur} tahun"

# Single quotes: literal (no interpolasi)
literal = 'Halo, #{nama}!'  # output: Halo, #{nama}!
```

**Multiline:**

```ruby
teks = "Baris 1
Baris 2
Baris 3"
```

---

### **6. COMMENTS**

```ruby
# Komentar satu baris

# Komentar
# multi
# baris
```

---

### **7. ERROR HANDLING**

```ruby
fungsi bagi(a, b)
  jika b == 0
    kembalikan galat("Pembagian dengan nol")
  akhir
  kembalikan a / b
akhir

hasil = bagi(10, 0)
jika adalah_galat(hasil)
  cetak("Error: #{pesan_galat(hasil)}")
lainnya
  cetak("Hasil: #{hasil}")
akhir
```

**Built-in error functions:**
- `galat(pesan)` → create error
- `adalah_galat(nilai)` → check if error
- `pesan_galat(err)` → get error message

---

### **8. OPERATORS**

**Arithmetic:**
```ruby
+ - * / %
```

**Comparison:**
```ruby
== != < > <= >=
```

**Logical (Phase 2):**
```ruby
dan    # and
atau   # or
tidak  # not
```

---

## **COMPLETE EXAMPLE - FULL BAHASA INDONESIA**

```ruby
# Program: Kalkulator faktorial dengan error handling

fungsi faktorial(n)
  # Validasi input
  jika tipe(n) != "integer"
    kembalikan galat("faktorial() membutuhkan bilangan bulat")
  akhir
  
  jika n < 0
    kembalikan galat("faktorial() membutuhkan bilangan non-negatif")
  akhir
  
  # Base case
  jika n <= 1
    kembalikan 1
  akhir
  
  # Recursive case
  kembalikan n * faktorial(n - 1)
akhir

fungsi utama()
  kasus_tes = [5, -1, "invalid", 10]
  
  untuk input dalam kasus_tes
    hasil = faktorial(input)
    
    jika adalah_galat(hasil)
      cetak("Error untuk input #{input}: #{pesan_galat(hasil)}")
    lainnya
      cetak("faktorial(#{input}) = #{hasil}")
    akhir
  akhir
akhir

utama()
```

**Output:**
```
faktorial(5) = 120
Error untuk input -1: faktorial() membutuhkan bilangan non-negatif
Error untuk input invalid: faktorial() membutuhkan bilangan bulat
faktorial(10) = 3628800
```

---

## **BUILT-IN FUNCTIONS - BAHASA INDONESIA**

```ruby
# I/O
cetak(nilai)          # Print ke stdout
masukan(prompt)       # Baca dari stdin (kembalikan string)

# Type checking
tipe(nilai)           # Kembalikan "integer", "string", "array", "map", dll
adalah_galat(nilai)   # Kembalikan benar/salah

# Error handling
galat(pesan)          # Buat nilai error
pesan_galat(err)      # Dapatkan pesan error

# Array operations
panjang(array)        # Panjang array
tambah(array, item)   # Tambah item (mutasi array)
ambil(array)          # Hapus item terakhir (mutasi array)

# String operations
panjang(string)       # Panjang string

# Conversion
ke_string(nilai)      # Konversi ke string
ke_integer(string)    # Konversi ke integer (kembalikan error jika invalid)

# Utility
rentang(awal, akhir)  # Generate array [awal, awal+1, ..., akhir-1]
```

---

## **TOKEN DEFINITIONS**

```go
package token

type TokenType string

const (
    // Literals
    INTEGER = "INTEGER"
    STRING  = "STRING"
    SYMBOL  = "SYMBOL"
    
    // Identifiers & Keywords
    IDENT     = "IDENT"
    FUNGSI    = "FUNGSI"     // fungsi
    KEMBALIKAN = "KEMBALIKAN" // kembalikan
    JIKA      = "JIKA"       // jika
    ATAUIJIKA = "ATAUIJIKA"  // atauJika
    LAINNYA   = "LAINNYA"    // lainnya
    SELAMA    = "SELAMA"     // selama
    UNTUK     = "UNTUK"      // untuk
    DALAM     = "DALAM"      // dalam
    BENAR     = "BENAR"      // benar
    SALAH     = "SALAH"      // salah
    KOSONG    = "KOSONG"     // kosong
    AKHIR     = "AKHIR"      // akhir
    
    // Operators
    ASSIGN   = "="
    PLUS     = "+"
    MINUS    = "-"
    ASTERISK = "*"
    SLASH    = "/"
    PERCENT  = "%"
    
    EQ     = "=="
    NOT_EQ = "!="
    LT     = "<"
    GT     = ">"
    LTE    = "<="
    GTE    = ">="
    
    BANG = "!"
    
    // Delimiters
    COMMA     = ","
    COLON     = ":"
    NEWLINE   = "NEWLINE"
    
    LPAREN   = "("
    RPAREN   = ")"
    LBRACE   = "{"
    RBRACE   = "}"
    LBRACKET = "["
    RBRACKET = "]"
    
    // Special
    EOF     = "EOF"
    ILLEGAL = "ILLEGAL"
    COMMENT = "COMMENT"
)

var keywords = map[string]TokenType{
    "fungsi":     FUNGSI,
    "kembalikan": KEMBALIKAN,
    "jika":       JIKA,
    "atauJika":   ATAUIJIKA,
    "lainnya":    LAINNYA,
    "selama":     SELAMA,
    "untuk":      UNTUK,
    "dalam":      DALAM,
    "benar":      BENAR,
    "salah":      SALAH,
    "kosong":     KOSONG,
    "akhir":      AKHIR,
}

func LookupIdent(ident string) TokenType {
    if tok, ok := keywords[ident]; ok {
        return tok
    }
    return IDENT
}
```

---

## **AST NODE NAMES - BAHASA INDONESIA INTERNAL**

```go
// Tetap pakai nama Go (English) untuk AST nodes
// Karena ini internal compiler, bukan user-facing

type Statement interface {
    statementNode()
}

type AssignmentStatement struct {
    Name  string
    Value Expression
    IsDeclaration bool
}

type ReturnStatement struct {
    Value Expression
}

type IfStatement struct {
    Condition     Expression
    Consequence   *BlockStatement
    AtauJikaBranches []AtauJikaBranch  // elsif branches
    Alternative   *BlockStatement
}

type AtauJikaBranch struct {
    Condition   Expression
    Consequence *BlockStatement
}

// dst...
```

**Rationale:** Internal code tetap English (Go convention), tapi keyword user-facing Bahasa Indonesia.

---

## **GRAMMAR (EBNF) - UPDATED**

```
program = statement* ;

statement = assignment_stmt
          | return_stmt
          | if_stmt
          | while_stmt
          | for_stmt
          | expression_stmt
          | fn_declaration ;

assignment_stmt = IDENTIFIER "=" expression ;
return_stmt = "kembalikan" expression ;
expression_stmt = expression ;

if_stmt = "jika" expression block
          ("atauJika" expression block)*
          ("lainnya" block)?
          "akhir" ;

while_stmt = "selama" expression block "akhir" ;

for_stmt = "untuk" IDENTIFIER "dalam" expression block "akhir" ;

block = statement* ;

fn_declaration = "fungsi" IDENTIFIER "(" parameter_list? ")" block "akhir" ;

parameter_list = IDENTIFIER ("," IDENTIFIER)* ;

expression = binary_expr ;

binary_expr = unary_expr (OPERATOR unary_expr)* ;

unary_expr = UNARY_OP unary_expr
           | primary ;

primary = INTEGER
        | STRING
        | SYMBOL
        | "benar" | "salah" | "kosong"
        | IDENTIFIER
        | array_literal
        | map_literal
        | call_expr
        | index_expr
        | "(" expression ")" ;

array_literal = "[" (expression ("," expression)*)? "]" ;

map_literal = "{" (map_pair ("," map_pair)*)? "}" ;

map_pair = (SYMBOL | STRING) ":" expression ;

call_expr = IDENTIFIER "(" argument_list? ")" ;

argument_list = expression ("," expression)* ;

index_expr = primary "[" expression "]" ;
```

---

## **EXAMPLE PROGRAMS**

### **1. Hello World**

```ruby
# hello.morph
cetak("Halo, dunia!")
```

---

### **2. Fibonacci**

```ruby
# fibonacci.morph

fungsi fibonacci(n)
  jika n <= 1
    kembalikan n
  akhir
  kembalikan fibonacci(n - 1) + fibonacci(n - 2)
akhir

fungsi utama()
  untuk i dalam rentang(0, 10)
    hasil = fibonacci(i)
    cetak("fibonacci(#{i}) = #{hasil}")
  akhir
akhir

utama()
```

---

### **3. Array Processing**

```ruby
# array_processing.morph

fungsi jumlah_array(arr)
  total = 0
  untuk angka dalam arr
    total = total + angka
  akhir
  kembalikan total
akhir

fungsi rata_rata(arr)
  jika panjang(arr) == 0
    kembalikan galat("Array kosong")
  akhir
  
  total = jumlah_array(arr)
  kembalikan total / panjang(arr)
akhir

fungsi utama()
  angka = [10, 20, 30, 40, 50]
  
  total = jumlah_array(angka)
  cetak("Total: #{total}")
  
  avg = rata_rata(angka)
  jika adalah_galat(avg)
    cetak("Error: #{pesan_galat(avg)}")
  lainnya
    cetak("Rata-rata: #{avg}")
  akhir
akhir

utama()
```

---

### **4. Map (Dictionary)**

```ruby
# map_example.morph

fungsi utama()
  # Buat database sederhana
  pengguna = {
    nama: "Alice",
    umur: 25,
    kota: "Jakarta",
    hobi: ["coding", "reading", "gaming"]
  }
  
  cetak("Nama: #{pengguna[:nama]}")
  cetak("Umur: #{pengguna[:umur]}")
  cetak("Kota: #{pengguna[:kota]}")
  
  cetak("Hobi:")
  untuk hobi dalam pengguna[:hobi]
    cetak("  - #{hobi}")
  akhir
  
  # Update data
  pengguna[:umur] = 26
  cetak("Umur baru: #{pengguna[:umur]}")
akhir

utama()
```

---

### **5. Input/Output**

```ruby
# kalkulator.morph

fungsi tambah(a, b)
  kembalikan a + b
akhir

fungsi kurang(a, b)
  kembalikan a - b
akhir

fungsi kali(a, b)
  kembalikan a * b
akhir

fungsi bagi(a, b)
  jika b == 0
    kembalikan galat("Tidak bisa membagi dengan nol")
  akhir
  kembalikan a / b
akhir

fungsi utama()
  cetak("=== Kalkulator Sederhana ===")
  
  a_str = masukan("Masukkan angka pertama: ")
  a = ke_integer(a_str)
  jika adalah_galat(a)
    cetak("Input tidak valid!")
    kembalikan kosong
  akhir
  
  operasi = masukan("Operasi (+, -, *, /): ")
  
  b_str = masukan("Masukkan angka kedua: ")
  b = ke_integer(b_str)
  jika adalah_galat(b)
    cetak("Input tidak valid!")
    kembalikan kosong
  akhir
  
  hasil = kosong
  
  jika operasi == "+"
    hasil = tambah(a, b)
  atauJika operasi == "-"
    hasil = kurang(a, b)
  atauJika operasi == "*"
    hasil = kali(a, b)
  atauJika operasi == "/"
    hasil = bagi(a, b)
  lainnya
    cetak("Operasi tidak dikenal!")
    kembalikan kosong
  akhir
  
  jika adalah_galat(hasil)
    cetak("Error: #{pesan_galat(hasil)}")
  lainnya
    cetak("Hasil: #{hasil}")
  akhir
akhir

utama()
```

---

## **FILE EXTENSION**

```
.morph
```

Example: `hello.morph`, `faktorial.morph`

---

## **CHECKLIST FINAL**

- ✅ **Bahasa Indonesia konsisten** (semua keywords)
- ✅ **Verbose tapi Ruby-inspired** (struktur jelas)
- ✅ **Dynamic typing** (no annotations)
- ✅ **Error handling fokus** (galat(), adalah_galat(), pesan_galat())
- ✅ **AST non-ambiguous** (atauJika satu keyword, akhir explicit)
- ✅ **Human-readable** (mudah dibaca Bahasa Indonesia)
- ✅ **Computer-friendly** (parser clear, no lookahead complex)

---

