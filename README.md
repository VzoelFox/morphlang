# Morph Language

Morph adalah bahasa pemrograman yang dirancang untuk **meminimalisir halusinasi AI**, **memaksimalkan debuggability**, dan memberikan **konteks yang jelas**. Morph dibangun dengan prinsip "Explicit over Implicit" dan "Error as Value".

## Fitur Utama

- **Syntax Bahasa Indonesia**: Menggunakan keyword bahasa Indonesia (fungsi, jika, kembalikan, dll).
- **Explicit Error Handling**: Error diperlakukan sebagai nilai dan harus ditangani secara eksplisit.
- **Context-Aware**: Setiap file source code memiliki file konteks (`.fox.vz`) yang membantu AI agent memahami struktur kode.
- **Type Safety**: Meskipun dinamis, Morph mendorong pengecekan tipe yang eksplisit.

## Instalasi

(Akan datang)

## Penggunaan

```bash
morph compile program.fox
```

## Contoh Kode

```ruby
# hello.fox
fungsi sapa(nama)
  cetak("Halo, #{nama}!")
akhir

sapa("Dunia")
```

## Lisensi

(Lihat LICENSE)
