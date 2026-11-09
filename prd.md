# 📋 Product Requirement Document (PRD)
## Sistem Manajemen Rental Barang Berbasis Go (Golang) REST API

---

## 1. 📌 Informasi Dokumen & Ringkasan Proyek
- **Nama Proyek:** Sistem Rental Barang Backend Service (REST API)
- **Teknologi Utama:** Golang (Standard Library `net/http`), MySQL Driver (`go-sql-driver/mysql`), JWT (`golang-jwt/jwt/v5`), Bcrypt (`golang.org/x/crypto/bcrypt`)
- **Tujuan Sistem:** Menyediakan platform backend andal untuk mengelola seluruh siklus bisnis penyewaan barang, mulai dari katalog produk multi-unit fisik, transaksi sewa, verifikasi pembayaran, manajemen pengembalian & denda, hingga laporan rekapitulasi penjualan bagi admin.

---

## 2. 🎯 Analisis Diagram & Kebutuhan Sistem

### 2.1. Berdasarkan Entity Relationship Diagram (ERD)
Sistem memiliki 6 entitas utama dengan relasi relasional sebagai berikut:

```
[Category] 1 ──── N [Products] 1 ──── N [Item Instances] 1 ──── N [Rentals] 1 ──── N [Payments]
                                                                        │
                                                                   N ─── 1
                                                                    [User]
```

1. **`category` (Kategori Produk):**
   - Atribut: `id_category` (PK), `category_name`.
   - Relasi: Satu kategori dapat memiliki banyak produk (1-to-N ke `products`).
2. **`products` (Katalog Produk / Master Data):**
   - Atribut: `id_produk` (PK), `id_kategori` (FK), `product_name`, `photo_url`, `deskripsi`, `spec`, `pricing_p_day`.
   - Relasi: Memiliki banyak unit fisik fisik (1-to-N ke `item_instances`).
3. **`item_instances` (Unit Fisik Barang / Serial Unit):**
   - Atribut: `id_unit` (PK), `id_produk` (FK), `asset_code`, `status` (`available`, `rented`, `maintenance`).
   - Fungsi: Memisahkan konsep "Katalog Produk" dengan "Unit Fisik Nyata", sehingga stok tercatat presisi per nomor aset fisik.
4. **`user` (Pengguna & Hak Akses):**
   - Atribut: `id_user` (PK), `username`, `email`, `phone`, `address`, `password_hash`, `role` (`admin`, `customer`/`pengguna`).
   - Relasi: Satu pengguna dapat melakukan banyak transaksi rental (1-to-N ke `rentals`).
5. **`rentals` (Transaksi Sewa / Peminjaman):**
   - Atribut: `id_rentals` (PK), `id_user` (FK), `id_unit` (FK), `delivery_address`, `start_date`, `return_date`, `actual_return_date`, `rental_fee`, `late_fee`, `rental_status` (`pending`, `confirmed`, `active`, `completed`, `cancelled`).
   - Relasi: Dibuat oleh `user`, meminjam `item_instances`, dan memiliki catatan pembayaran di `payments`.
6. **`payments` (Pembayaran Transaksi):**
   - Atribut: `id_payment` (PK), `id_rentals` (FK), `payment_date`, `amount`, `payment_method`, `payment_proof_url`, `payment_status` (`pending`, `completed`, `rejected`).

---

### 2.2. Berdasarkan Use Case Diagram

Sistem memiliki dua aktor utama:
1. **Admin (Pengelola Sistem):**
   - Login & Logout
   - CRUD Kategori (`category`)
   - CRUD Produk (`products`)
   - CRUD Pesanan & Manajemen Unit Fisik (`rentals` & `item_instances`)
   - CRUD Pengembalian (`rentals` actual return)
   - Melihat pesanan masuk (`filter rental_status = pending`)
   - Memvalidasi pesanan (Konfirmasi sewa & kunci unit barang)
   - Melihat riwayat transaksi (Seluruh riwayat rental)
   - Memvalidasi pengembalian barang (Pemeriksaan fisik unit & kalkulasi denda keterlambatan / `late_fee`)
   - Rekap penjualan / rental (Laporan total pemasukan & statistik transaksi)

2. **Pengguna / Customer:**
   - Registrasi akun (Include: Login otomatis setelah registrasi atau direct login)
   - Login sistem (Include: Melihat katalog produk)
   - Melihat katalog produk & detail spesifikasi
   - Menentukan pilihan produk & cek ketersediaan unit
   - Memesan barang (Booking rental dengan menentukan tanggal & alamat pengiriman)
   - Menentukan durasi peminjaman (Perhitungan estimasi biaya sewa berdasarkan hari)
   - Menunggu konfirmasi admin
   - Melakukan pembayaran (Upload bukti transfer / payment gateway simulator)
   - Melakukan pengembalian barang (Mengajukan pengembalian unit fisik)
   - Logout

---

## 3. 🔍 Status Fitur Saat Ini (Yang Sudah Dibuat)

| Modul | Endpoint | Method | Keterangan Status Saat Ini |
|---|---|---|---|
| **Auth** | `/register` | POST | Registrasi akun baru, enkripsi kata sandi dengan bcrypt. |
| **Auth** | `/login` | POST | Verifikasi email & password, pembuatan token JWT (24 jam). |
| **User** | `/users`, `/users/{id}` | GET, PUT, DELETE | Ambil semua user, update profil, hapus user. |
| **Category** | `/category`, `/category/{id}`, `/api/category`, `/api/category/{id}` | GET, POST, PUT, DELETE | **Selesai 100%**: CRUD Kategori lengkap, validasi nama, standardisasi nama tabel `category` & kolom `id_category`, proteksi error relasi foreign key. |
| **Products** | `/products`, `/products/{id}`, `/api/products`, `/api/products/{id}` | GET (Publik), POST, PUT, DELETE (Auth) | **Selesai 100%**: Katalog publik terbuka tanpa token, search query (`?search=...`), filter kategori (`?category_id=...`), detail produk + unit fisik, proteksi hapus jika unit aktif disewa, deteksi otomatis primary key (`id_product` / `id`). |
| **Item Instances** | `/item-instances`, `/{id}` | GET, POST, PUT, DELETE | CRUD unit fisik barang & status aset. |
| **Rentals** | `/rental`, `/rental/{id}` | GET, POST, PUT | Pembuatan sewa (update status unit ke `rented`), list rental, update status pengembalian. |
| **Payments** | `/payments`, `/payments/{id}` | GET, POST, PUT, DELETE | Input pembayaran, list pembayaran, update status bukti bayar. |

---

## 4. 🚀 Backlog & Fitur yang Akan Dibuat / Disempurnakan (Yang Belum Selesai)

1. **Role-Based Access Control (RBAC):**
   - Saat ini `AuthMiddleware` hanya memvalidasi token JWT tanpa mengecek peran (`role`).
   - Diperlukan middleware `AdminOnly` untuk mengamankan fungsi admin (CRUD produk, kategori, validasi pesanan, rekap penjualan).
2. **Katalog Produk Publik:**
   - Endpoint GET `/products` saat ini terkunci di balik `AuthMiddleware`. Pengguna tamu (guest) harus bisa melihat produk tanpa harus login terlebih dahulu sesuai Use Case.
3. **Kalkulasi Otomatis Durasi Sewa & Biaya:**
   - Kalkulasi otomatis total sewa: `durasi_hari * pricing_p_day`.
   - Validasi tanggal kembali (`return_date` >= `start_date`).
4. **Validasi Ketersediaan Unit (Concurrency Check):**
   - Memastikan unit yang dipilih (`id_unit`) statusnya benar-benar `available` sebelum mengizinkan sewa dibuat.
5. **Workflow Validasi Admin:**
   - Endpoint khusus Admin: `PUT /rental/{id}/validate-order` (Mengubah status dari `pending` ke `confirmed`).
   - Endpoint khusus Admin: `PUT /rental/{id}/validate-return` (Kalkulasi denda otomatis jika `actual_return_date > return_date`, lalu mengubah status unit kembali ke `available`).
6. **Laporan & Rekapitulasi Penjualan (Admin):**
   - Endpoint `GET /api/reports/sales` (Ringkasan total pendapatan, total unit tersewa, total denda yang terkumpul).
7. **Sinkronisasi Penamaan Skema DB:**
   - Penyelarasan nama kolom dan tabel agar konsisten 100% antara ERD (`id_produk`, `id_rentals`, `id_unit`, `id_category`) dan query SQL di controller.

---

## 5. 🏗 Arsitektur Teknis & Pola Desain
- **Pola Arsitektur:** MVC / Layered (Models, Services, Controllers, Routes, Middlewares, Config).
- **Format Pertukaran Data:** Standard REST JSON (`application/json`).
- **Autentikasi:** Stateless Bearer JWT Token dengan payload `user_id`, `username`, `role`.
- **Database Engine:** Relational MySQL dengan InnoDB storage engine untuk integritas Foreign Key dan transaksi ACID (`BEGIN`, `COMMIT`, `ROLLBACK`).

