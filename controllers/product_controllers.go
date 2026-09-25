package controllers

import (
	"belajar_go/config"
	"belajar_go/models"
	"encoding/json"
	"net/http"
	"strings"
)

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		GetProducts(w, r)
	case http.MethodPost:
		CreateProduct(w, r)
	case http.MethodPut:
		UpdateProduct(w, r)
	case http.MethodDelete:
		DeleteProduct(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

func GetProducts(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, product_name, photo_url, deskripsi, spec, pricing_p_day, id_category FROM products"

	rows, err := config.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data produk: " + err.Error()})
		return
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID,
			&p.ProductName,
			&p.PhotoURL,
			&p.Deskripsi,
			&p.Spec,
			&p.PricingPDay,
			&p.IDCategory,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data produk: " + err.Error()})
			return
		}
		products = append(products, p)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data produk",
		"data":    products,
	})
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := `INSERT INTO products (product_name, photo_url, deskripsi, spec, pricing_p_day, id_category) 
	          VALUES (?, ?, ?, ?, ?, ?)`

	result, err := config.DB.Exec(query, p.ProductName, p.PhotoURL, p.Deskripsi, p.Spec, p.PricingPDay, p.IDCategory)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menambahkan produk: " + err.Error()})
		return
	}

	lastID, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Produk berhasil ditambahkan!",
		"id":      lastID,
	})
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID produk harus diisi"})
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := `UPDATE products 
	          SET product_name = ?, photo_url = ?, deskripsi = ?, spec = ?, pricing_p_day = ?, id_category = ? 
	          WHERE id = ?`

	result, err := config.DB.Exec(query, p.ProductName, p.PhotoURL, p.Deskripsi, p.Spec, p.PricingPDay, p.IDCategory, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui produk: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Produk tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Produk dengan ID " + id + " berhasil diperbarui!",
	})
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID produk wajib diisi"})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memulai transaksi: " + err.Error()})
		return
	}

	_, err = tx.Exec("DELETE FROM item_instances WHERE id_products = ?", id)
	if err != nil {
		tx.Rollback() // Batal transaksi jika gagal
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus unit produk: " + err.Error()})
		return
	}

	result, err := tx.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		tx.Rollback() // Batal transaksi jika gagal
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus produk: " + err.Error()})
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
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan perubahan: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Produk beserta unit fisiknya dengan ID " + id + " berhasil dihapus!",
	})
}