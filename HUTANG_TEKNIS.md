# Laporan Analisis Hutang Teknis Morph

**Tanggal:** 20 Desember 2025
**Auditor:** Jules (AI Agent)
**Status:** Deep Analysis

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
**Status:** IMPROVED (Type Inference & Error Logic Added)
**Lokasi:** `pkg/analysis/analyzer.go`

Output `.fox.vz` telah ditingkatkan (20 Des 2025).
- **Type Inference:** Kini mendukung recursive expression analysis (variable lookup, binary ops).
- **Symbol Tracking:** Kondisi error (`if ... return galat`) kini diekstrak.
- **Sisa Gap:** Dokumentasi (`doc`) belum diekstrak (perlu dukungan AST).
- **Rekomendasi Lanjutan:** Tambahkan dukungan komentar di Parser untuk ekstraksi dokumentasi.

## 4. Ekosistem (Standard Library / COTC)
**Status:** MAJOR
**Lokasi:** `lib/cotc` vs `pkg/object/builtins.go`

- **Masalah:** `lib/cotc` hampir kosong. Sebagian besar logika "standar" (String ops, Math, IO) masih di-hardcode sebagai Go Builtins di `pkg/object/builtins.go`.
- **Dampak:** Portabilitas bahasa rendah dan ukuran binary VM membengkak.
- **Rekomendasi:** Freeze fitur VM. Mulai migrasi logika (misal: `gabung`, `pisah`) ke kode Morph native di `lib/cotc`, tinggalkan hanya primitive paling dasar di VM (Native Interface).

## 5. Kepatuhan Error Handling (VM Robustness)
**Status:** MEDIUM
**Lokasi:** `pkg/vm/vm.go`

- **Masalah:** VM mengembalikan Go `error` (Crash/Panic) untuk kesalahan runtime seperti "Argument Mismatch", melanggar prinsip "Error as Value" di `AGENTS.md` yang mengharuskan VM tidak boleh crash.
- **Dampak:** Program berhenti total alih-alih memberikan objek error yang bisa ditangani user.
- **Rekomendasi:** Ubah logic VM untuk mem-push Objek Error ke stack alih-alih me-return Go error untuk kesalahan non-sistem.

---

**Kesimpulan:**
Pondasi Morph solid secara konsep, namun implementasi saat ini masih sangat bergantung pada "scaffolding" Go Runtime. Untuk mencapai visi "Robust Foundation", prioritas utama adalah **memutuskan ketergantungan Go GC** dan **memperbaiki Analyzer** agar sesuai janji "Context-Aware".
