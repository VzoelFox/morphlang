# Sistem Memori Morph (Experimental)

Dokumen ini menjelaskan arsitektur sistem manajemen memori kustom yang digunakan oleh Morph. Sistem ini dirancang untuk memberikan kontrol penuh atas alokasi objek, mendukung snapshot/rollback instan, dan mengelola memori fisik secara efisien melalui mekanisme paging (swapping).

> **Status:** Experimental (Phase X)
> **Lokasi Kode:** `pkg/memory`

---

## 1. Arsitektur Hirarkis

Sistem memori Morph menggunakan metafora "Lemari Arsip" untuk mengorganisir memori, yang terdiri dari tiga tingkat abstraksi utama: **Cabinet (Lemari)**, **Drawer (Laci)**, dan **Tray (Nampan)**.

### 1.1 Cabinet (Lemari) - The Heap
*   **Representasi:** Struct `Cabinet`
*   **Fungsi:** Merupakan representasi dari seluruh *Heap* memori yang tersedia untuk VM.
*   **Kapasitas:** Secara teoritis mampu menampung hingga 1024 *Virtual Drawers*.
*   **Manajemen Fisik:** Mengelola pemetaan antara *Virtual Drawer ID* ke *Physical RAM Slot*.
*   **Komponen:**
    *   `Drawers[]`: Slice dari semua virtual drawer.
    *   `RAMSlots[]`: Array ukuran tetap (16 slot) yang merepresentasikan memori fisik (RAM) yang sebenarnya dialokasikan.

### 1.2 Drawer (Laci) - Memory Region
*   **Representasi:** Struct `Drawer`
*   **Ukuran:** 128 KB per Drawer.
*   **Fungsi:** Unit terkecil yang dapat di-*swap* (pindahkan) antara RAM dan Disk.
*   **Struktur:** Setiap Drawer dibagi menjadi dua bagian semi-space untuk mendukung *Copying Garbage Collection* di masa depan.
*   **State:**
    *   `Resident`: Berada di RAM (memiliki `PhysicalSlot` valid).
    *   `Swapped`: Berada di file swap (RAM slot -1).

### 1.3 Tray (Nampan) - Allocation Block
*   **Representasi:** Struct `Tray`
*   **Ukuran:** 64 KB per Tray (Setengah dari Drawer).
*   **Fungsi:** Area kontigu di mana objek dialokasikan menggunakan strategi *Bump Pointer*.
*   **Mekanisme:**
    *   Setiap Drawer memiliki `PrimaryTray` dan `SecondaryTray`.
    *   Alokasi baru selalu terjadi di `PrimaryTray`.
    *   Ketika `PrimaryTray` penuh, strategi GC (belum diimplementasi) akan menyalin objek hidup ke `SecondaryTray` atau Drawer baru akan dibuat.

---

## 2. Alokasi Memori (Bump Pointer)

Morph menggunakan alokasi *Bump Pointer* yang sangat cepat karena hanya melibatkan penambahan offset pointer.

1.  **Request Alokasi:** VM meminta memori sebesar `N` bytes.
2.  **Align Size:** Ukuran dibulatkan ke kelipatan 8 byte terdekat (padding) untuk alignment.
3.  **Cek Kapasitas Tray:**
    *   Jika sisa ruang di *Tray Aktif* cukup: Pointer `Current` digeser maju sebesar `N`. Alamat lama dikembalikan.
    *   Jika Tray penuh:
        1.  **Drawer Baru** dibuat.
        2.  Drawer baru di-set sebagai *Active Drawer*.
        3.  Alokasi dicoba kembali di drawer baru.

---

## 3. Virtual Memory & Swapping (Draft Otomatis)

Untuk memungkinkan program menggunakan memori lebih besar dari RAM fisik yang dialokasikan, Morph mengimplementasikan sistem *Demand Paging* kustom.

### 3.1 Spesifikasi Fisik
*   **Limit RAM:** 16 Slot x 128 KB = **2 MB** (Physical Memory Cap).
*   **Swap File:** `.morph_cache.z`

### 3.2 Mekanisme Page Fault
Ketika sistem mencoba mengakses alamat di Drawer yang tidak ada di RAM:
1.  **Deteksi:** `MMU` (Memory Management Unit) mendeteksi bahwa `PhysicalSlot` drawer tersebut adalah `-1`.
2.  **Eviksi (Jika RAM Penuh):**
    *   Sistem memilih "korban" (Drawer lain di RAM) untuk dikeluarkan.
    *   Data korban disalin dari RAM ke file swap `.morph_cache.z`.
    *   Metadata korban diperbarui (`IsSwapped = true`).
3.  **Load:**
    *   Data drawer yang diminta dibaca dari swap file ke slot RAM yang kosong.
    *   Metadata diperbarui (`PhysicalSlot = [Slot Baru]`).
4.  **Resolve:** Akses memori dilanjutkan seolah-olah data selalu ada di sana.

---

## 4. Snapshot & Rollback (Time Travel)

Fitur unik Morph adalah kemampuan untuk mengambil potret (snapshot) status memori secara instan dan mengembalikannya (rollback).

### 4.1 Snapshot
*   Menyimpan metadata Cabinet dan seluruh isi data Drawer.
*   Dapat dilakukan secara *fine-grained* (per Drawer) atau global (seluruh Heap).
*   Disimpan di memori (map `Snapshots`) atau diserialisasi ke disk menggunakan `gob`.

### 4.2 Rollback (Restore)
*   Mengembalikan state memori ke titik waktu tertentu.
*   **Strategi:** *Lazy Restore* / Demand Paging.
    *   Saat restore global, semua data tidak langsung dimuat ke RAM.
    *   Semua data di-dump ke Swap File terlebih dahulu.
    *   Semua Drawer ditandai sebagai `Swapped`.
    *   Data hanya dimuat ke RAM saat benar-benar diakses oleh program.

---

## 5. Pointer System

Morph menggunakan `Ptr` (Virtual Pointer) kustom, bukan pointer Go standar.

*   **Format:** `uint64`
    *   **32-bit Atas:** `Drawer ID` (Indeks Laci)
    *   **32-bit Bawah:** `Offset` (Posisi dalam Laci)
*   **Keuntungan:**
    *   **Relocatable:** Data fisik bisa dipindah (swap in/out) tanpa mengubah nilai pointer yang dipegang VM.
    *   **Safe:** Validasi batas laci mudah dilakukan.

---

## 6. Garbage Collection

Saat ini, Morph **belum** mengimplementasikan Garbage Collector otomatis. Memori hanya dibebaskan saat VM dimatikan atau di-reset total. Desain *Semi-Space* pada Drawer disiapkan untuk algoritma *Copying GC* (Cheney's Algorithm) di masa depan.
