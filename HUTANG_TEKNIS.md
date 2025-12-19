# Laporan Analisis Hutang Teknis Morph

**Tanggal:** 20 Desember 2025
**Auditor:** Jules (AI Agent)
**Status:** Deep Analysis (Updated)

---

## 1. Arsitektur & Memory (Phase X - Hybrid State)
**Status:** FIXED (Pop Rehydration Removed)
**Lokasi:** `pkg/memory`, `pkg/vm`, `pkg/object`

Perbaikan signifikan telah diterapkan (20 Des 2025):
- **Pop Optimization:** Fungsi `vm.pop()` tidak lagi memanggil `Rehydrate()`. Wrapper object Go hanya dibuat saat dibutuhkan (lazy) melalui `vm.GetLastPopped()` untuk keperluan testing/debug.
- **Pure Bytecode Ops:** Operasi aritmatika di `vm_ops.go` sudah beroperasi langsung pada `memory.Ptr` tanpa wrapper.
- **Hasil:** Eliminasi alokasi Go wrapper di hot-path VM loop. Double GC pressure berkurang drastis.

## 2. Struktur Data (Missing Structs)
**Status:** FIXED (Structs Implemented)
**Lokasi:** `pkg/memory`, `pkg/vm`, `pkg/compiler`

Fitur `Struct` telah diimplementasikan (20 Des 2025):
- **Syntax:** `struktur Nama { field }` dan `Nama(val)`.
- **Memory:** `AllocSchema` dan `AllocStruct` dengan layout tetap.
- **VM:** `OpStruct` untuk definisi, `OpCall` (Schema) untuk instansiasi, dan `OpIndex` (Struct) untuk akses field.
- **Hasil:** Struktur data berkinerja tinggi dengan offset-based field access (via Schema lookup).

## 3. Tooling Debt (Analyzer / Context Generator)
**Status:** FIXED (Documentation Extraction Added)
**Lokasi:** `pkg/analysis/analyzer.go`, `pkg/lexer`, `pkg/parser`

Output `.fox.vz` telah mencapai standar `AGENTS.md` (20 Des 2025).
- **Type Inference:** Recursive expression analysis.
- **Error Tracking:** Kondisi error (`if ... return galat`) diekstrak.
- **Dokumentasi:** Lexer & Parser dimodifikasi untuk menangkap komentar `# doc` dan menempelkannya ke AST Node (Fungsi/Struct). Analyzer kini mengisi field `doc` di context.
- **Hasil:** AI Agent memiliki visibilitas penuh terhadap maksud kode (dokumentasi) dan perilaku (tipe/error).

## 4. Ekosistem (Standard Library / COTC)
**Status:** MAJOR (Fake Completion in Roadmap)
**Lokasi:** `lib/cotc` vs `pkg/object/builtins.go`

- **Masalah:** `lib/cotc` hampir kosong. Sebagian besar logika "standar" (String ops, Math, IO) masih di-hardcode sebagai Go Builtins di `pkg/object/builtins.go`.
- **Discrepancy:** `ROADMAP.md` menandai Phase 6 (COTC) sebagai "Selesai" (v1.1.0), padahal implementasinya masih hybrid/scaffolding.
- **Dampak:** Portabilitas bahasa rendah dan ukuran binary VM membengkak.
- **Rekomendasi:** Freeze fitur VM. Mulai migrasi logika (misal: `gabung`, `pisah`) ke kode Morph native di `lib/cotc`, tinggalkan hanya primitive paling dasar di VM (Native Interface).

## 5. Kepatuhan Error Handling (VM Robustness)
**Status:** CRITICAL
**Lokasi:** `pkg/vm/vm.go`, `pkg/vm/vm_ops.go`

- **Masalah:** VM mengembalikan Go `error` (Crash/Panic) untuk kesalahan runtime seperti "Argument Mismatch", "Type Mismatch", atau "Index Out of Bounds".
- **Pelanggaran:** Ini melanggar prinsip "Error as Value" di `AGENTS.md` yang mengharuskan VM **tidak boleh crash/panic** kecuali untuk kesalahan sistem (OOM). Kesalahan logika user harus mengembalikan objek `Error` ke stack.
- **Dampak:** Program berhenti total (crash) alih-alih memberikan kesempatan recovery (`jika adalah_galat(...)`).
- **Rekomendasi:** Refactor `vm.Run` loop dan opcode handlers agar mem-push Objek Error ke stack saat terjadi kesalahan runtime.

## 6. Dokumentasi & Spesifikasi (Documentation Debt)
**Status:** HIGH
**Lokasi:** `AGENTS.md`, `PEDOMAN_AWAL.md`

- **CLI Discrepancy:** `AGENTS.md` mencontohkan perintah `morph compile --debug <file>`, namun implementasi `main.go` menggunakan parser flag Go standar yang mengharuskan flag sebelum subcommand (`morph --debug compile <file>`). Ini menyesatkan users/agents.
- **Obsolete Spec:** File `PEDOMAN_AWAL.md` masih tersimpan di repo namun isinya bertentangan dengan `AGENTS.md` (Source of Truth).
  - Ekstensi file: `.morph` (Pedoman) vs `.fox` (Agents/Roadmap).
  - Akses Map: `orang[:nama]` (Pedoman) vs `orang.nama` (Agents).
- **Rekomendasi:** Refactor parser CLI di `main.go` agar mendukung urutan argumen sesuai contoh di `AGENTS.md` (`morph compile --debug ...`), dan tandai `PEDOMAN_AWAL.md` sebagai **DEPRECATED** atau hapus.

---

**Kesimpulan:**
Pondasi Morph solid secara konsep, namun implementasi saat ini "bergerak terlalu cepat" meninggalkan hutang teknis pada kepatuhan spesifikasi (Error Handling) dan dokumentasi. Prioritas perbaikan harus pada **VM Robustness** agar sesuai dengan janji "Zero Ambiguity / Error as Value".
