# Spesifikasi Bahasa Morph (The Morph Standard)

**Versi:** 1.0.0-DRAFT
**Status:** STANDARD
**Tanggal:** 2025-12-14

Dokumen ini adalah spesifikasi teknis resmi untuk bahasa pemrograman Morph. Semua implementasi compiler dan runtime Morph (MorphVM) **WAJIB** mematuhi spesifikasi ini untuk menjamin kompatibilitas 100%.

Tujuan dokumen ini adalah memastikan bahwa kode Morph yang dikompilasi di satu mesin akan berjalan identik di mesin lain, terlepas dari platform atau bahasa implementasi compiler (Go, Rust, C++, dll).

---

## 1. Grammar (Sintaks)

Morph menggunakan grammar bebas konteks (Context-Free Grammar) yang didefinisikan dalam EBNF.

*(Referensi: Lihat `docs/grammar.md` untuk definisi EBNF lengkap)*

### Prinsip Utama: Zero Ambiguity
Aturan berikut bersifat **Normatif**:
1.  **Spasi Operator:** Binary operators (`+`, `-`, `*`, `/`, `==`, dll) **WAJIB** diapit spasi.
    *   Benar: `a + b`
    *   Salah: `a+b` (Harus ditolak oleh Parser dengan Error Syntax)
2.  **Blok Kode:** Wajib diakhiri kata kunci `akhir`.
    *   Tidak menggunakan kurung kurawal `{}` atau indentasi (Python-style) sebagai penentu blok logika utama.
3.  **String Interpolation:** Menggunakan sintaks `#{ekspresi}` di dalam double-quotes.
    *   Contoh: `"Hasil: #{x + y}"`

---

## 2. Semantik (Arti Kode)

### 2.1 Scoping (Cakupan Variabel)
Morph menerapkan **Lexical Scoping** (Static Scoping).
- **Global Scope:** Variabel yang didefinisikan di luar fungsi/blok manapun.
- **Local Scope:** Variabel yang didefinisikan di dalam blok (`fungsi`, `jika`, `selama`).
- **Visibility:** Scope dalam dapat melihat variabel scope luar. Scope luar *tidak bisa* melihat scope dalam.
- **Shadowing:** Variabel lokal dengan nama yang sama dengan variabel luar akan "menutupi" (shadow) variabel luar tersebut selama dalam scope lokal.

### 2.2 Siklus Hidup Variabel
- **Deklarasi:** Bersifat implisit melalui assignment pertama (`identifier = value`).
- **Inisialisasi:** Variabel dianggap ada sejak baris assignment dieksekusi.
- **Akses:** Mengakses variabel yang belum di-assign (atau salah eja) akan memicu **Runtime Error** (`Undefined Symbol`), bukan mengembalikan `null`.

### 2.3 Fungsi
- **First-class Citizens:** Fungsi adalah nilai (value). Bisa disimpan dalam variabel, array, atau dikirim sebagai argumen ke fungsi lain.
- **Higher-Order Functions:** Mendukung fungsi yang menerima atau mengembalikan fungsi.
- **Closure:** Fungsi "menangkap" (capture) lingkungan (environment) variabel tempat ia didefinisikan.
- **Parameter Passing:**
    - Tipe Primitif (Integer, Boolean): **Pass-by-value** (Salinan nilai).
    - Tipe Kompleks (Error, Function, Future: Map/List): **Pass-by-reference** (Pointer ke objek yang sama).

### 2.4 Control Flow
- `jika` dan `selama` adalah **Ekspresi**. Mereka mengevaluasi dan mengembalikan nilai dari statement terakhir di blok yang dieksekusi.
- Jika blok tidak dieksekusi (kondisi salah), mengembalikan `kosong` (null).

---

## 3. Type System (Sistem Tipe)

Morph adalah bahasa **Dynamically Typed** (tipe dicek saat runtime) namun **Strongly Typed** (tipe ketat, tidak ada koersi implisit yang berbahaya).

### Tipe Data Primitif
| Tipe | Keyword/Nama | Deskripsi | Contoh |
|------|--------------|-----------|--------|
| **Integer** | `integer` | Bilangan bulat 64-bit signed (`int64`) | `42`, `-10` |
| **String** | `string` | Urutan karakter UTF-8 immutable | `"Halo"` |
| **Boolean** | `boolean` | Logika benar/salah | `benar`, `salah` |
| **Error** | `error` | Objek khusus kesalahan | `galat("Pesan")` |
| **Kosong** | `kosong` | Representasi ketiadaan nilai (Null/Nil) | `kosong` |
| **Fungsi** | `function` | Blok kode eksekutabel | `fungsi() ...` |

### Aturan Type Checking (Matriks Kompatibilitas)
Runtime WAJIB memeriksa tipe operand sebelum operasi dijalankan.

| Operasi | Operand A | Operand B | Hasil | Perilaku / Error |
|---------|-----------|-----------|-------|------------------|
| `+`     | Integer   | Integer   | Integer | Penjumlahan aritmatika |
| `+`     | String    | String    | String  | Konkatenasi string |
| `+`     | Integer   | String    | **ERROR** | **E003: Type Mismatch** (Gunakan `ke_string()`) |
| `+`     | String    | Integer   | **ERROR** | **E003: Type Mismatch** |
| `-`, `*`, `/` | Integer | Integer | Integer | Aritmatika |
| `-`, `*`, `/` | String | Any | **ERROR** | Operasi tidak valid untuk string |
| `==`, `!=` | Any | Any | Boolean | Perbandingan nilai (Deep equality untuk primitif) |
| `<`, `>` | Integer | Integer | Boolean | Perbandingan numerik |
| `!` | Boolean | - | Boolean | Negasi logika |
| `!` | Any (Non-Bool) | - | **ERROR** | **E003: Type Mismatch** (Harus boolean) |

---

## 4. Code Generation & Virtual Machine (MorphVM)

Untuk menjamin **"Identical Code Generation"** dan **"Identical Runtime Behavior"**, spesifikasi ini mendefinisikan arsitektur **Morph Virtual Machine (MVM)**.

### 4.1 Arsitektur VM
MorphVM adalah **Stack-based Virtual Machine**.
- **Operand Stack:** Tempat nilai sementara disimpan dan dimanipulasi.
- **Memory Areas:**
    - **Code/Text:** Instruksi bytecode (Read-only).
    - **Constant Pool:** Literal statis (angka, string) yang dimuat saat kompilasi.
    - **Global Store:** Variabel global (persisten selama program berjalan).
    - **Heap:** Alokasi dinamis untuk objek kompleks (String, Function, Error).
- **Execution Frame:** Setiap pemanggilan fungsi membuat frame baru berisi:
    - Return Address (Instruction Pointer).
    - Local Variables (Array akses cepat).
    - Reference ke Closure (jika ada).

### 4.2 Instruction Set Architecture (ISA) - Standard Opcodes

Setiap instruksi terdiri dari **1 byte Opcode**, diikuti oleh **0 atau lebih byte Operand**.
Semua operand integer bersifat **Big-Endian**.

#### Stack Manipulation
| Opcode | Hex | Mnemonic | Operand | Deskripsi |
|--------|-----|----------|---------|-----------|
| 0x01 | `POP` | - | Pop nilai teratas stack. |

#### Constants & Variables
| Opcode | Hex | Mnemonic | Operand | Deskripsi |
|--------|-----|----------|---------|-----------|
| 0x10 | `LOAD_CONST` | `u16 index` | Push konstanta dari `Pool[index]`. |
| 0x11 | `LOAD_GLOBAL`| `u16 index` | Push nilai variabel global dengan ID `index`. |
| 0x12 | `STORE_GLOBAL`| `u16 index` | Pop nilai, simpan ke global dengan ID `index`. |
| 0x13 | `LOAD_LOCAL` | `u8 index` | Push nilai variabel lokal pada frame index `index`. |
| 0x14 | `STORE_LOCAL`| `u8 index` | Pop nilai, simpan ke lokal index `index`. |

#### Arithmetic & Logic
| Opcode | Hex | Mnemonic | Operand | Deskripsi |
|--------|-----|----------|---------|-----------|
| 0x20 | `ADD` | - | Pop `b`, Pop `a`, Push `a + b`. |
| 0x21 | `SUB` | - | Pop `b`, Pop `a`, Push `a - b`. |
| 0x22 | `MUL` | - | Pop `b`, Pop `a`, Push `a * b`. |
| 0x23 | `DIV` | - | Pop `b`, Pop `a`, Push `a / b`. |
| 0x24 | `EQ` | - | Pop `b`, Pop `a`, Push `a == b`. |
| 0x25 | `NEQ` | - | Pop `b`, Pop `a`, Push `a != b`. |
| 0x26 | `GT` | - | Pop `b`, Pop `a`, Push `a > b`. |
| 0x27 | `GTE` | - | Pop `b`, Pop `a`, Push `a >= b`. |
| 0x2F | `NEG` | - | Pop `a`, Push `-a` (Unary minus). |
| 0x2E | `NOT` | - | Pop `a`, Push `!a` (Unary not). |

#### Control Flow
| Opcode | Hex | Mnemonic | Operand | Deskripsi |
|--------|-----|----------|---------|-----------|
| 0x30 | `JUMP` | `u16 offset` | Ubah IP (Instruction Pointer) ke `offset`. |
| 0x31 | `JUMP_IF_FALSE`| `u16 offset` | Pop `kond`. Jika `salah`, jump ke `offset`. |

#### Functions
| Opcode | Hex | Mnemonic | Operand | Deskripsi |
|--------|-----|----------|---------|-----------|
| 0x40 | `CALL` | `u8 numArgs` | Panggil fungsi di stack dengan `N` argumen. |
| 0x41 | `RETURN` | - | Kembali dari fungsi (return `kosong`). |
| 0x42 | `RETURN_VAL` | - | Kembali dari fungsi dengan nilai di top stack. |

---

## 5. Runtime Environment

### 5.1 Error Handling (Error as Value)
Runtime tidak menggunakan Exception Throwing untuk logic flow.
- Jika operasi (misal pembagian nol) gagal, instruksi VM (misal `DIV`) **WAJIB** mempush objek `Error` ke stack, bukan crash.
- Kode pengguna harus memeriksa hasil operasi.
- **Panic Mode:** Jika error sistem kritis (Stack Overflow, Out of Memory), VM berhenti total.

### 5.2 Built-in Functions (Standard Library)
Setiap runtime Morph **WAJIB** menyediakan fungsi-fungsi berikut secara global:

1.  **`cetak(val)`**
    - Mencetak representasi string dari `val` ke Standard Output (stdout).
    - Mengembalikan: `kosong`.
2.  **`panjang(val)`**
    - Menerima `string` (jumlah karakter) atau `array` (jumlah elemen).
    - Mengembalikan: `integer`.
    - Error jika tipe salah.
3.  **`tipe(val)`**
    - Mengembalikan string nama tipe: `"integer"`, `"string"`, `"boolean"`, `"error"`, `"function"`, `"kosong"`.
4.  **`galat(pesan)`**
    - Membuat objek `Error` dengan pesan `pesan` (string).
    - Mengembalikan: `error`.
5.  **`adalah_galat(val)`**
    - Mengembalikan `benar` jika `val` bertipe `error`, `salah` jika tidak.
6.  **`pesan_galat(err)`**
    - Mengambil string pesan dari objek `error`.
    - Error jika `err` bukan tipe `error`.

### 5.3 Entry Point
Setiap program Morph dimulai dari statement tingkat atas (top-level) yang dieksekusi secara sekuensial dari baris pertama file utama. Tidak ada fungsi `main()` wajib, namun konvensi menyarankan penggunaan fungsi `utama()` yang dipanggil di akhir file.

---
**Dokumen ini bersifat mengikat untuk pengembangan Fase 3 (Interpreter) dan Fase 5 (Bytecode VM).**
