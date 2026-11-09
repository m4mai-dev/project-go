# 📦 Sistem Rental Barang Backend RESTful API (Golang)

Backend service berbasis RESTful API untuk mengelola ekosistem persewaan barang secara terpusat, mulai dari manajemen inventaris unit fisik (*asset tracking*), peminjaman (*rental*), pembayaran (*payments*), perhitungan denda keterlambatan (*late fee*), hingga manajemen peran (*admin & customer*).

Dibangun menggunakan bahasa pemrograman **Go (Golang)** dengan pustaka standar `net/http` dan basis data relasional **MySQL**.

---

## 📑 Daftar Isi
- [Arsitektur & Diagram Sistem](#-arsitektur--diagram-sistem)
  - [Entity Relationship Diagram (ERD)](#entity-relationship-diagram-erd)
  - [Use Case Diagram](#use-case-diagram)
- [Teknologi yang Digunakan](#-teknologi-yang-digunakan)
- [Struktur Database (DDL SQL)](#-struktur-database-ddl-sql)
- [Panduan Instalasi & Menjalankan Proyek](#-panduan-instalasi--menjalankan-proyek)
- [Dokumentasi Endpoint API](#-dokumentasi-endpoint-api)
- [Struktur Direktori Proyek](#-struktur-direktori-proyek)
- [Dokumen Pendukung](#-dokumen-pendukung)

---

## 🏛 Arsitektur & Diagram Sistem

### Entity Relationship Diagram (ERD)
Sistem memiliki 6 tabel relasional yang dinormalisasi untuk memisahkan katalog produk dengan nomor seri aset fisik individual:

```
[category] 1 ──── N [products] 1 ──── N [item_instances] 1 ──── N [rentals] 1 ──── N [payments]
                                                                      │
                                                                 N ─── 1
                                                                  [user]
```

### Use Case Diagram
- **Aktor Admin:**
  - Login / Logout
  - CRUD Kategori Barang
  - CRUD Katalog Produk
  - CRUD Unit Fisik & Pesanan Rental
  - Memvalidasi Pesanan & Bukti Pembayaran Masuk
  - Memvalidasi Pengembalian Barang & Input/Kalkulasi Denda
  - Rekapitulasi Laporan Penjualan/Rental
- **Aktor Pengguna (Customer):**
  - Registrasi Akun & Login
  - Eksplorasi Katalog Produk & Cek Spesifikasi
  - Menentukan Durasi Peminjaman & Alamat Pengiriman
  - Melakukan Booking Rental Barang
  - Melakukan Pembayaran & Upload Bukti Transfer
  - Mengajukan / Melakukan Pengembalian Barang

---

## 🛠 Teknologi yang Digunakan
- **Bahasa Pemrograman:** [Go (Golang)](https://go.dev/) (v1.20+)
- **Routing & Server:** Standard Library `net/http` (Custom ServeMux & Middleware)
- **Database Engine:** MySQL 8.x / MariaDB (InnoDB Storage Engine)
- **Database Driver:** `github.com/go-sql-driver/mysql`
- **Autentikasi:** JWT (`github.com/golang-jwt/jwt/v5`)
- **Enkripsi Kata Sandi:** Bcrypt (`golang.org/x/crypto/bcrypt`)

---

## 🗄 Struktur Database (DDL SQL)

Jalankan script SQL berikut pada database MySQL Anda (`rental_barang`):

```sql
CREATE DATABASE IF NOT EXISTS rental_barang;
USE rental_barang;

-- 1. Tabel Kategori
CREATE TABLE IF NOT EXISTS category (
    id_category INT AUTO_INCREMENT PRIMARY KEY,
    category_name VARCHAR(100) NOT NULL
) ENGINE=InnoDB;

-- 2. Tabel Produk (Master Data Katalog)
CREATE TABLE IF NOT EXISTS products (
    id_product INT AUTO_INCREMENT PRIMARY KEY,
    product_name VARCHAR(150) NOT NULL,
    photo_url VARCHAR(255) DEFAULT '',
    deskripsi TEXT,
    spec TEXT,
    pricing_p_day DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    id_category INT NOT NULL,
    FOREIGN KEY (id_category) REFERENCES category(id_category) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 3. Tabel Unit Fisik (Serial / Aset)
CREATE TABLE IF NOT EXISTS item_instances (
    id INT AUTO_INCREMENT PRIMARY KEY,
    id_products INT NOT NULL,
    asset_code VARCHAR(100) NOT NULL UNIQUE,
    status ENUM('available', 'rented', 'maintenance') DEFAULT 'available',
    FOREIGN KEY (id_products) REFERENCES products(id_product) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 4. Tabel Pengguna
CREATE TABLE IF NOT EXISTS users (
    id_user INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(150) NOT NULL UNIQUE,
    phone VARCHAR(30) DEFAULT '',
    address TEXT,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('admin', 'customer') DEFAULT 'customer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- 5. Tabel Transaksi Rental
CREATE TABLE IF NOT EXISTS rentals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    id_user INT NOT NULL,
    id_unit INT NOT NULL,
    delivery_address TEXT NOT NULL,
    start_date DATE NOT NULL,
    return_date DATE NOT NULL,
    actual_return_date DATE NULL,
    rental_fee DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    late_fee DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    rental_status ENUM('pending', 'confirmed', 'active', 'completed', 'cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE,
    FOREIGN KEY (id_unit) REFERENCES item_instances(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 6. Tabel Pembayaran
CREATE TABLE IF NOT EXISTS payment (
    id INT AUTO_INCREMENT PRIMARY KEY,
    id_rentals INT NOT NULL,
    payment_date DATETIME NOT NULL,
    amount DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    payment_method VARCHAR(50) NOT NULL,
    payment_proof_url VARCHAR(255) DEFAULT '',
    payment_status ENUM('pending', 'completed', 'rejected') DEFAULT 'pending',
    FOREIGN KEY (id_rentals) REFERENCES rentals(id) ON DELETE CASCADE
) ENGINE=InnoDB;
```

---

## 🚀 Panduan Instalasi & Menjalankan Proyek

### 1. Clone Repository & Masuk ke Folder
```bash
git clone https://github.com/username/belajar_go.git
cd belajar_go
```

### 2. Konfigurasi Basis Data
Buka file `config/database.go` dan sesuaikan kredensial MySQL lokal Anda:
```go
DB, err = sql.Open(
    "mysql",
    "username:password@tcp(127.0.0.1:3306)/rental_barang?parseTime=true",
)
```

### 3. Unduh Dependencies
```bash
go mod tidy
```

### 4. Jalankan Server
```bash
go run main.go
```
Server akan aktif di: `http://localhost:8080`

---

## 📡 Dokumentasi Endpoint API

Semua endpoint dengan tanda 🔒 mewajibkan header:
`Authorization: Bearer <token_jwt>`

### 1. Otentikasi & Akun
| Method | Endpoint | Auth | Deskripsi |
|---|---|:---:|---|
| `POST` | `/register` | Publik | Mendaftarkan akun baru |
| `POST` | `/login` | Publik | Login dan mendapatkan token JWT |
| `GET` | `/users` | 🔒 | Mengambil daftar semua user |
| `PUT` | `/users/{id}` | 🔒 | Memperbarui profil user |
| `DELETE` | `/users/{id}` | 🔒 | Menghapus akun user |

### 2. Kategori Produk
| Method | Endpoint | Auth | Deskripsi |
|---|---|:---:|---|
| `POST` | `/category` | 🔒 | Menambah kategori baru |
| `PUT` | `/api/category/{id}` | 🔒 | Memperbarui nama kategori |
| `DELETE` | `/api/category/{id}` | 🔒 | Menghapus kategori |

### 3. Katalog Produk & Unit Fisik
| Method | Endpoint | Auth | Deskripsi |
|---|---|:---:|---|
| `GET` | `/products` / `/products/{id}` | Publik | Mengambil katalog produk (mendukung `?search=nama` & `?category_id=id`) atau detail produk beserta unit fisiknya |
| `POST` | `/products` | 🔒 | Menambah produk baru ke katalog (validasi kategori wajib ada) |
| `PUT` | `/products/{id}` | 🔒 | Mengubah data produk |
| `DELETE` | `/products/{id}` | 🔒 | Menghapus produk beserta unit fisiknya (ACID Transaction & proteksi jika unit aktif disewa) |
| `GET` | `/item-instances` / `/item-instances/{id}` | 🔒 | Mengambil unit fisik (mendukung filter `?product_id=id`, `?status=available`, `?search=kode`) atau detail 1 unit fisik |
| `POST` | `/item-instances` | 🔒 | Mendaftarkan unit fisik baru (validasi unik `asset_code` & produk wajib ada) |
| `PUT` | `/item-instances/{id}` | 🔒 | Mengubah status aset (`available`, `rented`, `maintenance`) |
| `DELETE` | `/item-instances/{id}` | 🔒 | Menghapus unit fisik (diproteksi jika unit sedang berstatus `rented`) |

### 4. Rental & Pembayaran
| Method | Endpoint | Auth | Deskripsi |
|---|---|:---:|---|
| `GET` | `/rental` / `/rental/{id}` | 🔒 | Mengambil daftar rental (filter `?status=pending` untuk pesanan masuk, `?user_id=1` untuk riwayat user) atau detail 1 transaksi |
| `POST` | `/rental` | 🔒 | Membuat transaksi rental baru (anti double-booking & auto kalkulasi biaya sewa, mengunci unit fisik ke `rented`) |
| `PUT` | `/rental/{id}` | 🔒 | Memperbarui transaksi: input `actual_return_date` otomatis menghitung `late_fee` jika terlambat, dan OTOMATIS merestore status unit fisik kembali ke `available` |
| `DELETE` | `/rental/{id}` | 🔒 | Membatalkan/menghapus transaksi rental dan memulihkan status unit menjadi `available` |
| `GET` | `/payments` / `/payments/{id}` | 🔒 | Mengambil daftar pembayaran (filter `?rental_id=1`, `?status=pending`) atau detail 1 pembayaran |
| `POST` | `/payments` | 🔒 | Mencatat pembayaran baru pelanggan (status default `pending`) |
| `PUT` | `/payments/{id}` | 🔒 | Memperbarui/memvalidasi pembayaran (saat status `completed`, pesanan rental otomatis disetujui / `confirmed`) |
| `DELETE` | `/payments/{id}` | 🔒 | Menghapus catatan pembayaran |

### 5. Laporan & Rekapitulasi (Admin)
| Method | Endpoint | Auth | Deskripsi |
|---|---|:---:|---|
| `GET` | `/reports/sales` / `/api/reports/sales` | 🔒 | Rekapitulasi penjualan admin: total revenue, total biaya sewa, total denda, rincian status pesanan, 5 produk terlaris, dan 5 transaksi terkini |

---

## 📂 Struktur Direktori Proyek

```
belajar_go/
├── config/
│   └── database.go              # Koneksi database MySQL
├── controllers/
│   ├── auth_controllers.go      # Handler login & registrasi
│   ├── category_controller.go   # Handler CRUD kategori
│   ├── item_instances_controllers.go # Handler unit fisik aset
│   ├── payment_controllers.go   # Handler pembayaran
│   ├── product_controllers.go   # Handler master katalog produk
│   ├── rental_controllers.go    # Handler peminjaman & pengembalian
│   └── user_controllers.go      # Handler manajemen user
├── middlewares/
│   └── auth.go                  # Verifikasi otentikasi JWT Bearer
├── models/
│   ├── category.go              # Model struct kategori
│   ├── item_instances.go        # Model struct unit fisik
│   ├── payment.go               # Model struct pembayaran
│   ├── product.go               # Model struct produk
│   ├── rental.go                # Model struct rental
│   └── user.go                  # Model struct pengguna
├── routes/
│   ├── routes.go                # Registrasi rute aplikasi & CORS
│   └── api.go                   # Setup rute modular
├── services/
│   ├── item_instances_services.go
│   ├── product_service.go
│   └── rental_service.go
├── uploads/
│   └── products/                # Direktori penyimpanan file gambar produk
├── utils/
│   └── jwt.go                   # Generator & validator token JWT
├── audit.html                   # Laporan audit interaktif untuk browser web
├── audit.md                     # Laporan audit teknis & temuan bug (Markdown)
├── catatan_presentasi.html      # Cheatsheet & alur bicara presentasi untuk web
├── catatan_presentasi.md        # Panduan presentasi & kisi-kisi Q&A (Markdown)
├── prd.md                       # Product Requirement Document lengkap
├── go.mod                       # Modul Go
└── main.go                      # Entry point utama aplikasi
```

---

## 📚 Dokumen Pendukung
Proyek ini dilengkapi dengan berkas dokumentasi lengkap:
1. **[Product Requirement Document (PRD)](prd.md)** &mdash; Spesifikasi kebutuhan sistem, alur use case, dan fitur.
2. **[Laporan Audit Teknis](audit.md)** | **[Versi Web](audit.html)** &mdash; Status penyelesaian fitur, analisis bug kritis, dan rencana perbaikan.
3. **[Panduan Presentasi](catatan_presentasi.md)** | **[Versi Web](catatan_presentasi.html)** &mdash; Naskah alur bicara, slide pembuka, kisi-kisi pertanyaan dosen penguji, dan skenario live demo.

