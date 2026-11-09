# 🎤 Catatan & Panduan Presentasi (Presentation Cheatsheet)
## Proyek: Sistem REST API Rental Barang Berbasis Golang & MySQL

Dokumen ini disusun khusus sebagai bahan panduan, contekan taktis, dan alur bicara (*speaking points*) saat Anda mempresentasikan proyek ini di hadapan dosen penguji, klien, atau audiens teknis.

---

## 📌 Bagian 1: Pembukaan & Latar Belakang Masalah (1-2 Menit)

### 🎙 Poin Pembicaraan (Script):
> *"Selamat pagi/siang Bapak/Ibu penguji. Hari ini saya mempresentasikan proyek backend service: **Sistem Rental Barang Berbasis RESTful API menggunakan Golang**.*
>
> *Pada bisnis rental barang (misalnya alat kamera, proyektor, atau perangkat audio), tantangan utamanya bukan sekadar menjual barang, melainkan **melacak ketersediaan setiap unit fisik individual**, menghitung durasi sewa, mencatat uang sewa dan denda keterlambatan secara presisi, serta memverifikasi bukti pembayaran. Proyek ini dibangun untuk menjawab kebutuhan otomasi tersebut dengan performa tinggi dan arsitektur database yang rapi."*

### 💡 Poin Kunci yang Ditekankan:
- Model bisnis rental berbeda dengan e-commerce biasa: Satu jenis produk (misal: Sony A7 IV) bisa memiliki 5 unit fisik berbeda, dan masing-masing unit memiliki nomor aset serta status peminjaman tersendiri.

---

## 🏗 Bagian 2: Arsitektur & Pilihan Teknologi (Tech Stack)

### 🎙 Poin Pembicaraan:
> *"Sistem ini dibangun dengan stack teknologi pilihan:*
> 1. ***Golang (Go):*** Menggunakan pustaka standar `net/http` untuk routing dan HTTP handling. Go dipilih karena kompilasi native yang sangat cepat, efisiensi memori tinggi, serta kapabilitas konkurensi (goroutines) yang siap menangani ribuan transaksi secara bersamaan.
> 2. ***MySQL & Driver `go-sql-driver/mysql`:*** Menggunakan database relasional dengan engine InnoDB guna menjamin transaksi bersifat **ACID** (Atomicity, Consistency, Isolation, Durability), terutama saat terjadi booking dan pengembalian unit.
> 3. ***Keamanan (JWT & Bcrypt):*** Autentikasi stateless menggunakan JSON Web Token (JWT) dengan algoritma enkripsi kata sandi menggunakan Bcrypt (One-way hashing + salt) sehingga tidak ada plaintext password yang tersimpan di basis data."*

---

## 🗄 Bagian 3: Penjelasan Desain Basis Data (ERD)

### 🎙 Poin Pembicaraan:
> *"Struktur basis data dirancang secara ternormalisasi dengan 6 tabel utama:*
> 1. ***`category`:*** Mengelompokkan jenis barang (contoh: Kamera, Audio, Lensa).
> 2. ***`products`:*** Menyimpan informasi katalog (nama produk, spesifikasi, foto, dan tarif sewa harian `pricing_p_day`).
> 3. ***`item_instances`:*** **(Tabel Krusial)** Menyimpan unit fisik nyata dengan atribut `asset_code` (nomor seri) dan status (`available`, `rented`, `maintenance`).
> 4. ***`user`:*** Menyimpan data pengguna beserta perannya (`admin` atau `customer`).
> 5. ***`rentals`:*** Mencatat perjanjian sewa antara user dan unit fisik tertentu, tanggal mulai, batas kembali, biaya sewa, serta denda keterlambatan (`late_fee`).
> 6. ***`payments`:*** Mencatat rincian transaksi pembayaran, metode bayar, bukti transfer, dan status verifikasi."*

### 🎯 Jawaban Rahasia Penguji: *"Kenapa memisahkan `products` dan `item_instances`?"*
> *"Jika kita hanya menyimpan kolom `stock` di tabel `products`, kita tidak bisa melacak unit fisik mana yang sedang rusak, unit mana yang disewa oleh si A, dan unit mana yang nomor serinya hilang. Dengan adanya tabel `item_instances`, kita menerapkan konsep **asset tracking** seperti sistem ERP perusahaan."*

---

## 👥 Bagian 4: Penjelasan Use Case & Alur Kerja Bisnis

Sistem ini memfasilitasi 2 aktor:
1. **Alur Pengguna (Customer Flow):**
   - Registrasi & Login &rarr; Mendapatkan Bearer JWT Token.
   - Melihat katalog produk & spesifikasi teknis.
   - Memilih barang & tanggal peminjaman.
   - Melakukan order rental (`POST /rental`).
   - Melakukan pembayaran dan upload bukti transfer (`POST /payments`).
   - Menunggu barang diserahkan dan melakukan pengembalian saat masa sewa selesai.

2. **Alur Pengelola (Admin Flow):**
   - Mengelola master data kategori, produk, dan unit fisik barang.
   - Memantau pesanan masuk berstatus `pending`.
   - Memvalidasi pembayaran dan menyetujui pesanan (`confirmed`).
   - Menerima pengembalian barang: memeriksa kondisi unit dan memasukkan `actual_return_date`. Sistem menghitung denda otomatis jika melewati batas `return_date`.
   - Mengubah status unit kembali menjadi `available`.
   - Melihat laporan rekapitulasi penjualan.

---

## 💻 Bagian 5: Panduan Skenario Live Demo (Langkah Demi Langkah)

Siapkan Postman atau cURL dengan skenario pengujian runut berikut:

| Langkah | Aksi / Endpoint | Payload / Header | Respon yang Diharapkan |
|---|---|---|---|
| **Step 1** | Registrasi Akun<br>`POST /register` | `{"username":"budi", "email":"budi@mail.com", "password":"123", "role":"customer"}` | HTTP 201 Created ("Registrasi berhasil!") |
| **Step 2** | Login Akun<br>`POST /login` | `{"email":"budi@mail.com", "password":"123"}` | HTTP 200 OK & Token JWT |
| **Step 3** | Lihat Katalog Produk<br>`GET /products/` | Header: `Authorization: Bearer <token>` | HTTP 200 OK & List produk beserta harga harian |
| **Step 4** | Buat Transaksi Sewa<br>`POST /rental` | `{"id_user": 1, "id_unit": 2, "delivery_address":"Jl. Sudirman", "start_date":"2026-11-10", "return_date":"2026-11-12", "rental_fee": 300000}` | HTTP 201 Created & status unit otomatis berubah menjadi `rented` |
| **Step 5** | Bayar Sewa<br>`POST /payments` | `{"id_rentals": 1, "amount": 300000, "payment_method": "transfer_bca", "payment_proof_url": "bukti.jpg"}` | HTTP 201 Created & Bukti bayar tersimpan |
| **Step 6** | Pengembalian Barang & Denda<br>`PUT /rental/1` | `{"actual_return_date": "2026-11-13", "late_fee": 50000, "rental_status": "completed"}` | HTTP 200 OK & Unit fisik kembali berstatus `available` |

---

## ❓ Bagian 6: Cheatsheet Jawaban Pertanyaan Dosen Penguji (Q&A)

### Q1: *"Bagaimana sistem menangani jika dua pengguna menyewa satu barang yang sama secara bersamaan (Concurrency / Race Condition)?"*
> **Jawaban:**  
> *"Pada sistem backend, kami menerapkan pengecekan status unit secara langsung pada transaksi basis data. Ketika pesanan baru dibuat, kami memastikan status unit adalah `available`. Pada query update status unit ke `rented`, dilakukan secara atomic. Jika unit sudah berstatus `rented`, transaksi berikutnya akan ditolak dengan pesan error bahwa unit tidak lagi tersedia."*

### Q2: *"Kenapa menggunakan JWT, apa keuntungannya dibandingkan session biasa?"*
> **Jawaban:**  
> *"JWT bersifat **stateless**, artinya server tidak perlu menyimpan status login pengguna di memori server atau tabel session database. Semua informasi penting (User ID, username, dan role) terenkripsi di dalam signature token. Hal ini membuat REST API kami sangat mudah untuk di-*scale up* ke banyak server tanpa kendala sinkronisasi session."*

### Q3: *"Bagaimana cara sistem menghitung denda keterlambatan (Late Fee)?"*
> **Jawaban:**  
> *"Denda dihitung berdasarkan selisih antara `actual_return_date` (tanggal aktual barang kembali) dengan `return_date` (jadwal perjanjian). Jika selisih hari positif (terlambat), sistem mengalikan jumlah hari keterlambatan dengan tarif denda harian (misalnya tarif sewa harian + denda penalti)."*

### Q4: *"Kenapa saat menghapus produk, Anda menggunakan Database Transaction (`tx.Begin()`)?"*
> **Jawaban:**  
> *"Karena saat menghapus data produk, kita juga harus membersihkan atau memvalidasi data anak di tabel `item_instances`. Dengan database transaction, jika terjadi kegagalan di tengah jalan (misal gagal hapus unit fisik), sistem akan melakukan `tx.Rollback()` sehingga data master produk tidak menjadi yatim piatu (*orphan record*) dan data tetap konsisten."*

---

## 🏆 Bagian 7: Penutup & Kesimpulan (Closing Pitch)
> *"Kesimpulannya, sistem ini telah mengimplementasikan seluruh kebutuhan fundamental dari siklus bisnis rental barang, mulai dari manajemen aset, peminjaman, pembayaran, hingga pengembalian. Arsitekturnya yang modular dan berbasis standar RESTful API memudahkan sistem ini untuk diintegrasikan dengan frontend web (React/Vue) maupun aplikasi mobile (Flutter/Android). Sekian presentasi dari saya, terima kasih dan saya terbuka untuk pertanyaan maupun masukan dari Bapak/Ibu."*

