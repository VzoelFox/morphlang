# Spesifikasi Bahasa Morph

**Versi:** 0.1.0 (Draft)

## 1. Syntax Dasar

Morph menggunakan sintaks yang terinspirasi dari Ruby tetapi dengan keyword Bahasa Indonesia.

### Komentar
```ruby
# Ini adalah komentar satu baris
```

### Variabel
Variabel dideklarasikan dengan assignment.
```ruby
x = 10
nama = "Budi"
aktif = benar
```

## 2. Tipe Data

- **Integer**: Bilangan bulat (`10`, `-5`)
- **String**: Teks (`"Halo"`)
- **Boolean**: `benar` atau `salah`
- **Error**: Tipe khusus untuk kesalahan (`galat("pesan")`)
- **Kosong**: Nilai null/nil (`kosong`)

## 3. Kontrol Alur

### Jika (If-Else)
```ruby
jika x > 10
  cetak("Besar")
atau_jika x == 10
  cetak("Pas")
lainnya
  cetak("Kecil")
akhir
```

### While Loop (Selama)
(Akan didefinisikan nanti, mungkin `selama`?)

## 4. Fungsi

Fungsi dideklarasikan dengan kata kunci `fungsi` dan diakhiri dengan `akhir`.

```ruby
fungsi tambah(a, b)
  kembalikan a + b
akhir
```

## 5. Error Handling

Morph menggunakan error as value.

```ruby
fungsi bagi(a, b)
  jika b == 0
    kembalikan galat("Pembagian nol")
  akhir
  kembalikan a / b
akhir
```

Pengecekan error:
```ruby
hasil = bagi(10, 0)
jika adalah_galat(hasil)
  cetak("Error terjadi")
akhir
```
