# Morph Routines (Konkurensi)

Morph menyediakan model konkurensi (concurrency) yang kuat namun sederhana, terinspirasi oleh Go Goroutines dan Channels. Di Morph, unit eksekusi konkuren disebut **"Morph Routine"** atau direpresentasikan sebagai objek **Utas** (`Thread`).

> **Status:** Stabil (Phase 5)
> **Lokasi Kode:** `pkg/vm`, `pkg/object/builtin_concurrency.go`

---

## 1. Konsep Dasar

Konkurensi di Morph dibangun di atas prinsip: **"Jangan berkomunikasi dengan membagi memori; bagilah memori dengan berkomunikasi."**

### 1.1 Utas (Thread)
*   Unit eksekusi independen yang ringan.
*   Setiap Utas memiliki stack VM dan isolasi variabel lokalnya sendiri.
*   Utas berkomunikasi satu sama lain menggunakan **Saluran** (Channel).

### 1.2 Saluran (Channel)
*   Pipa komunikasi *type-safe* (di level object) untuk mengirim dan menerima data antar utas.
*   Bisa *buffered* (memiliki kapasitas antrian) atau *unbuffered* (sinkron).
*   Operasi kirim dan terima bersifat *blocking* secara default.

---

## 2. API Konkurensi

Morph menyediakan fungsi bawaan (built-in) untuk mengelola konkurensi.

### 2.1 `luncurkan(fungsi)`
Memulai eksekusi fungsi di background (utas baru).

*   **Parameter:** Fungsi yang akan dijalankan.
*   **Return:** Objek `Utas`.
*   **Perilaku:**
    *   Fungsi dieksekusi secara asinkron.
    *   VM utama tidak menunggu fungsi selesai (kecuali dipanggil `gabung`).
    *   Variabel global disalin (shallow copy) ke VM utas baru untuk isolasi dasar.

```ruby
fungsi kerja_berat()
  tidur(1000)
  kembalikan "Selesai"
akhir

# Jalankan di background
utas_ku = luncurkan(kerja_berat)
```

### 2.2 `gabung(utas)`
Menunggu utas selesai dan mengambil nilai kembaliannya.

*   **Parameter:** Objek `Utas` yang ingin ditunggu.
*   **Return:** Nilai yang dikembalikan oleh fungsi dalam utas tersebut.
*   **Perilaku:** Memblokir eksekusi utas pemanggil sampai utas target selesai.

```ruby
hasil = gabung(utas_ku)
cetak(hasil) # Output: "Selesai"
```

### 2.3 `saluran_baru(ukuran_buffer?)`
Membuat saluran komunikasi baru.

*   **Parameter (Opsional):** Integer ukuran buffer. Jika 0 atau tidak ada, saluran adalah *unbuffered*.
*   **Return:** Objek `Saluran`.

```ruby
ch = saluran_baru()      # Unbuffered
ch_buf = saluran_baru(5) # Buffered kapasitas 5
```

### 2.4 `kirim(saluran, nilai)`
Mengirim data ke dalam saluran.

*   **Parameter:**
    1.  Objek `Saluran`.
    2.  Nilai (Objek) yang akan dikirim.
*   **Return:** `Kosong` (Null).
*   **Perilaku:** Memblokir jika saluran penuh (buffered) atau tidak ada penerima (unbuffered).

```ruby
kirim(ch, "Pesan Rahasia")
```

### 2.5 `terima(saluran)`
Menerima data dari saluran.

*   **Parameter:** Objek `Saluran`.
*   **Return:** Nilai (Objek) yang diterima.
*   **Perilaku:** Memblokir jika saluran kosong sampai ada data yang masuk.

```ruby
pesan = terima(ch)
```

---

## 3. Primitif Sinkronisasi Tingkat Rendah

Selain Channel, Morph juga menyediakan primitif sinkronisasi klasik untuk kasus penggunaan khusus (misalnya shared state yang kompleks).

### 3.1 Mutex
Kunci eksklusif (Mutual Exclusion).

*   `mutex_baru()`: Membuat kunci baru.
*   `kunci(mutex)`: Mengunci (Lock). Memblokir jika sudah dikunci orang lain.
*   `buka(mutex)`: Membuka kunci (Unlock).

### 3.2 Atom
Wadah data tunggal yang thread-safe.

*   `atom_baru(nilai)`: Membuat atom dengan nilai awal.
*   `atom_tukar(atom, nilai_baru)`: Mengganti nilai atom dan mengembalikan nilai lama (Atomic Swap).
*   `atom_baca(atom)`: Membaca nilai saat ini dengan aman.

---

## 4. Contoh Penggunaan (Pattern)

### Worker Pool Sederhana

```ruby
fungsi pekerja(id, ch_tugas, ch_hasil)
  selama benar
    tugas = terima(ch_tugas)
    jika tugas == "STOP"
      berhenti
    akhir

    # Simulasi proses
    hasil = tugas * 2
    kirim(ch_hasil, hasil)
  akhir
akhir

# Setup
tugas_ch = saluran_baru(10)
hasil_ch = saluran_baru(10)

# Spawn worker
luncurkan(fungsi()
  pekerja(1, tugas_ch, hasil_ch)
akhir)

# Kirim tugas
kirim(tugas_ch, 10)
kirim(tugas_ch, 20)

# Ambil hasil
cetak(terima(hasil_ch)) # 20
cetak(terima(hasil_ch)) # 40
```

---

## 5. Implementasi Internal

Di balik layar, Morph memetakan konsep konkurensinya langsung ke fitur Go:
*   `Utas` Morph = `goroutine` Go.
*   `Saluran` Morph = `channel` Go (`chan object.Object`).
*   `Mutex` Morph = `sync.Mutex` Go.

Ini memastikan bahwa konkurensi di Morph sama efisien dan stabilnya dengan konkurensi di Go.
