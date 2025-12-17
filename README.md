# Morph Language

Morph adalah bahasa pemrograman yang dirancang untuk **meminimalisir halusinasi AI**, **memaksimalkan debuggability**, dan memberikan **konteks yang jelas**. Morph dibangun dengan prinsip "Explicit over Implicit" dan "Error as Value".

## Fitur Utama

- **Syntax Bahasa Indonesia**: Menggunakan keyword bahasa Indonesia (fungsi, jika, kembalikan, dll).
- **Explicit Error Handling**: Error diperlakukan sebagai nilai dan harus ditangani secara eksplisit.
- **Context-Aware**: Setiap file source code memiliki file konteks (`.fox.vz`) yang membantu AI agent memahami struktur kode.
- **Type Safety**: Meskipun dinamis, Morph mendorong pengecekan tipe yang eksplisit.

## Instalasi

Pastikan Anda telah menginstal [Go](https://go.dev/) (versi 1.18+).

1. Clone repository ini:
   ```bash
   git clone https://github.com/VzoelFox/morphlang.git
   cd morphlang
   ```

2. Build executable:
   ```bash
   go build -o morph cmd/morph/main.go
   ```

3. (Opsional) Tambahkan ke PATH atau pindahkan ke `/usr/local/bin`:
   ```bash
   mv morph /usr/local/bin/
   ```

## Penggunaan

Morph dilengkapi dengan CLI tool untuk mengkompilasi dan menganalisis kode.

### Kompilasi & Generate Context

Perintah `compile` akan memparsing kode Morph dan menghasilkan file konteks (`.vz`) yang digunakan oleh AI Agent.

```bash
./morph compile <file.fox>
```

Contoh:
```bash
./morph compile examples/hello.fox
```

Output:
- Jika sukses, akan muncul pesan `Successfully compiled ...`
- File baru `examples/hello.fox.vz` akan dibuat. File ini berisi metadata, simbol, statistik, dan struktur kode dalam format JSON.

### Debug Mode

Gunakan flag `--debug` untuk melihat output detail dari Lexer dan Parser:

```bash
./morph compile --debug examples/hello.fox
```

## Fitur Context & Session (Anti-Halusinasi)

Salah satu fitur unik Morph adalah **Context Generation**. Setiap kali kode dikompilasi, Morph menganalisis kode tersebut dan menyimpan informasinya ke dalam file `.fox.vz`.

File context ini berisi:
- **Simbol**: Daftar fungsi, parameter, dan variabel global.
- **Statistik**: Jumlah baris kode, komentar, dan kompleksitas.
- **Call Graph**: Relasi antar fungsi (siapa memanggil siapa).
- **Type Inference**: Prediksi tipe data variabel.

**Kegunaan:** AI Agent (seperti saya) membaca file ini untuk "memahami" kode Anda tanpa perlu menebak-nebak, sehingga mengurangi risiko halusinasi saat membantu coding.

## Pengujian (Testing)

Untuk memverifikasi sistem analisis dan context generation, Anda dapat menjalankan unit test yang tersedia.

### Menjalankan Session/Analysis Test

Test ini memvalidasi bahwa `analyzer` menghasilkan data konteks yang akurat.

```bash
go test -v ./pkg/analysis
```

Output yang diharapkan:
```
=== RUN   TestGenerateContext
--- PASS: TestGenerateContext (0.00s)
=== RUN   TestAnalyzeErrorFunction
--- PASS: TestAnalyzeErrorFunction (0.00s)
PASS
ok      github.com/VzoelFox/morphlang/pkg/analysis  ...
```

### Menjalankan Semua Test

Untuk menjalankan seluruh test suite (Lexer, Parser, Object, Analysis):

```bash
go test ./...
```

## COTC (Standard Library)

Morph menyertakan **COTC (Core of The Core)**, pustaka standar yang ditulis dalam Morph.
Gunakan perintah `ambil "cotc/nama_modul"` untuk menggunakannya.

Lihat [Dokumentasi COTC](docs/cotc.md) untuk detail modul yang tersedia.

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
