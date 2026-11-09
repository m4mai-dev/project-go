package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"belajar_go/config"
	"belajar_go/models"
	"belajar_go/utils"
)

// Helper untuk mengekstrak ID dari URL path produk (misal: /products/1 atau /api/products/1)
func extractProductID(path string) string {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == "products" {
		return ""
	}
	return last
}

// Helper untuk mendeteksi nama kolom primary key pada tabel products (id_product vs id)
func detectProductPK() string {
	var dummy int
	err := config.DB.QueryRow("SELECT id_product FROM products LIMIT 1").Scan(&dummy)
	if err == nil || err == sql.ErrNoRows {
		return "id_product"
	}
	return "id"
}

// ProductHandler mengatur akses:
// - GET: Terbuka untuk umum (Tamu/Customer) sesuai Use Case diagram "melihat produk"
// - POST, PUT, DELETE: Wajib otentikasi JWT Bearer Token (Admin / Pengelola)
func ProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := extractProductID(r.URL.Path)

	// Jika metode manipulasi data (POST, PUT, DELETE), validasi token JWT
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Akses ditolak: Operasi penambahan/perubahan/penghapusan produk memerlukan autentikasi Bearer token",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := utils.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Token tidak valid atau telah kadaluarsa",
			})
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		if id != "" {
			GetProductByID(w, r, id)
		} else {
			GetProducts(w, r)
		}
	case http.MethodPost:
		CreateProduct(w, r)
	case http.MethodPut:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID produk wajib disertakan pada URL path"})
			return
		}
		UpdateProduct(w, r, id)
	case http.MethodDelete:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID produk wajib disertakan pada URL path"})
			return
		}
		DeleteProduct(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

// GetProducts mengambil daftar produk dengan dukungan filter pencarian (?search=...) & kategori (?category_id=...)
func GetProducts(w http.ResponseWriter, r *http.Request) {
	pk := detectProductPK()
	searchQuery := r.URL.Query().Get("search")
	categoryFilter := r.URL.Query().Get("category_id")

	baseQuery := fmt.Sprintf(`
		SELECT 
			p.%s, 
			p.product_name, 
			p.photo_url, 
			p.deskripsi, 
			p.spec, 
			p.pricing_p_day, 
			p.id_category,
			COALESCE(c.category_name, '') AS category_name,
			COALESCE(u.avail_count, 0) AS available_units,
			COALESCE(u.total_count, 0) AS total_units
		FROM products p
		LEFT JOIN category c ON c.id_category = p.id_category
		LEFT JOIN (
			SELECT 
				id_products,
				COUNT(*) AS total_count,
				SUM(CASE WHEN status = 'available' THEN 1 ELSE 0 END) AS avail_count
			FROM item_instances
			GROUP BY id_products
		) u ON u.id_products = p.%s
		WHERE 1=1
	`, pk, pk)

	var conditions []string
	var args []interface{}

	if strings.TrimSpace(searchQuery) != "" {
		conditions = append(conditions, "p.product_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(searchQuery)+"%")
	}

	if strings.TrimSpace(categoryFilter) != "" {
		if catID, err := strconv.Atoi(categoryFilter); err == nil {
			conditions = append(conditions, "p.id_category = ?")
			args = append(args, catID)
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += fmt.Sprintf(" ORDER BY p.%s DESC", pk)

	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data produk: " + err.Error()})
		return
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		var photoURL, deskripsi, spec, catName sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.ProductName,
			&photoURL,
			&deskripsi,
			&spec,
			&p.PricingPDay,
			&p.IDCategory,
			&catName,
			&p.AvailableUnits,
			&p.TotalUnits,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data produk: " + err.Error()})
			return
		}

		p.PhotoURL = photoURL.String
		p.Deskripsi = deskripsi.String
		p.Spec = spec.String
		p.CategoryName = catName.String
		p.ProductID = p.ID // Dukungan kompatibilitas kunci id & id_product

		products = append(products, p)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data katalog produk",
		"total":   len(products),
		"data":    products,
	})
}

// GetProductByID mengambil detail 1 produk lengkap dengan daftar unit fisik ketersediaannya
func GetProductByID(w http.ResponseWriter, r *http.Request, id string) {
	pk := detectProductPK()

	query := fmt.Sprintf(`
		SELECT 
			p.%s, 
			p.product_name, 
			p.photo_url, 
			p.deskripsi, 
			p.spec, 
			p.pricing_p_day, 
			p.id_category,
			COALESCE(c.category_name, '') AS category_name,
			COALESCE(u.avail_count, 0) AS available_units,
			COALESCE(u.total_count, 0) AS total_units
		FROM products p
		LEFT JOIN category c ON c.id_category = p.id_category
		LEFT JOIN (
			SELECT 
				id_products,
				COUNT(*) AS total_count,
				SUM(CASE WHEN status = 'available' THEN 1 ELSE 0 END) AS avail_count
			FROM item_instances
			GROUP BY id_products
		) u ON u.id_products = p.%s
		WHERE p.%s = ?
	`, pk, pk, pk)

	var p models.Product
	var photoURL, deskripsi, spec, catName sql.NullString

	err := config.DB.QueryRow(query, id).Scan(
		&p.ID,
		&p.ProductName,
		&photoURL,
		&deskripsi,
		&spec,
		&p.PricingPDay,
		&p.IDCategory,
		&catName,
		&p.AvailableUnits,
		&p.TotalUnits,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Produk dengan ID " + id + " tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil detail produk: " + err.Error()})
		return
	}

	p.PhotoURL = photoURL.String
	p.Deskripsi = deskripsi.String
	p.Spec = spec.String
	p.CategoryName = catName.String
	p.ProductID = p.ID

	// Ambil daftar unit fisik individual dari item_instances
	unitsQuery := "SELECT id, id_products, asset_code, status FROM item_instances WHERE id_products = ?"
	unitRows, err := config.DB.Query(unitsQuery, id)
	units := make([]models.ItemInstance, 0)
	if err == nil {
		defer unitRows.Close()
		for unitRows.Next() {
			var u models.ItemInstance
			if errScan := unitRows.Scan(&u.ID, &u.IDProducts, &u.AssetCode, &u.Status); errScan == nil {
				units = append(units, u)
			}
		}
	}

	detail := models.ProductDetailResponse{
		Product: p,
		Units:   units,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil detail produk",
		"data":    detail,
	})
}

// CreateProduct menambahkan produk baru ke database (Memerlukan Admin Auth)
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	if strings.TrimSpace(input.ProductName) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nama produk wajib diisi"})
		return
	}

	if input.PricingPDay <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Tarif sewa harian (pricing_p_day) harus lebih dari 0"})
		return
	}

	if input.IDCategory == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID kategori (id_category) wajib diisi"})
		return
	}

	// Validasi apakah kategori benar-benar ada di database
	var catName string
	errCat := config.DB.QueryRow("SELECT category_name FROM category WHERE id_category = ?", input.IDCategory).Scan(&catName)
	if errCat != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Kategori dengan ID %d tidak ditemukan di database", input.IDCategory),
		})
		return
	}

	query := `INSERT INTO products (product_name, photo_url, deskripsi, spec, pricing_p_day, id_category) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	result, err := config.DB.Exec(query, input.ProductName, input.PhotoURL, input.Deskripsi, input.Spec, input.PricingPDay, input.IDCategory)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menambahkan produk: " + err.Error()})
		return
	}

	lastID, _ := result.LastInsertId()
	createdProduct := models.Product{
		ID:           uint(lastID),
		ProductID:    uint(lastID),
		ProductName:  input.ProductName,
		PhotoURL:     input.PhotoURL,
		Deskripsi:    input.Deskripsi,
		Spec:         input.Spec,
		PricingPDay:  input.PricingPDay,
		IDCategory:   input.IDCategory,
		CategoryName: catName,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Produk berhasil ditambahkan ke katalog!",
		"data":    createdProduct,
	})
}

// UpdateProduct memperbarui informasi produk yang sudah ada
func UpdateProduct(w http.ResponseWriter, r *http.Request, id string) {
	pk := detectProductPK()

	var input models.ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	if strings.TrimSpace(input.ProductName) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nama produk wajib diisi"})
		return
	}

	if input.IDCategory > 0 {
		var catName string
		if errCat := config.DB.QueryRow("SELECT category_name FROM category WHERE id_category = ?", input.IDCategory).Scan(&catName); errCat != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": fmt.Sprintf("Kategori dengan ID %d tidak ditemukan", input.IDCategory),
			})
			return
		}
	}

	query := fmt.Sprintf(`UPDATE products 
	          SET product_name = ?, photo_url = ?, deskripsi = ?, spec = ?, pricing_p_day = ?, id_category = ? 
	          WHERE %s = ?`, pk)

	result, err := config.DB.Exec(query, input.ProductName, input.PhotoURL, input.Deskripsi, input.Spec, input.PricingPDay, input.IDCategory, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui produk: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Produk dengan ID " + id + " tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Produk dengan ID " + id + " berhasil diperbarui!",
	})
}

// DeleteProduct menghapus produk beserta unit fisik anak menggunakan ACID Transaction
func DeleteProduct(w http.ResponseWriter, r *http.Request, id string) {
	pk := detectProductPK()

	// Cek apakah ada unit fisik produk yang saat ini sedang dalam status sewa (rented)
	var rentedCount int
	checkQuery := "SELECT COUNT(*) FROM item_instances WHERE id_products = ? AND status = 'rented'"
	err := config.DB.QueryRow(checkQuery, id).Scan(&rentedCount)
	if err == nil && rentedCount > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Tidak dapat menghapus produk: terdapat %d unit fisik yang sedang disewa pelanggan", rentedCount),
		})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memulai database transaction: " + err.Error()})
		return
	}

	// 1. Hapus unit fisik terkait di item_instances
	_, err = tx.Exec("DELETE FROM item_instances WHERE id_products = ?", id)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus unit fisik produk: " + err.Error()})
		return
	}

	// 2. Hapus data produk di tabel products
	deleteProductQuery := fmt.Sprintf("DELETE FROM products WHERE %s = ?", pk)
	result, err := tx.Exec(deleteProductQuery, id)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus data produk: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Produk tidak ditemukan"})
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan perubahan transaksi: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Produk beserta unit fisiknya dengan ID " + id + " berhasil dihapus dari sistem!",
	})
}
