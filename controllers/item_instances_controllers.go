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
)

// Helper untuk mengekstrak ID dari URL path /item-instances/{id} atau /api/item-instances/{id}
func extractItemInstanceID(path string) string {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == "item-instances" {
		return ""
	}
	return last
}

// ItemInstanceHandler mengelola routing RESTful untuk unit fisik barang
func ItemInstanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := extractItemInstanceID(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if id != "" {
			GetItemInstanceByID(w, r, id)
		} else {
			GetItemInstances(w, r)
		}
	case http.MethodPost:
		CreateItemInstance(w, r)
	case http.MethodPut:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID item instance harus disertakan pada URL path"})
			return
		}
		UpdateItemInstance(w, r, id)
	case http.MethodDelete:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID item instance harus disertakan pada URL path"})
			return
		}
		DeleteItemInstance(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

// GetItemInstances mengambil daftar unit fisik dengan filter ?product_id=, ?status=, dan ?search=
func GetItemInstances(w http.ResponseWriter, r *http.Request) {
	productIDFilter := r.URL.Query().Get("product_id")
	statusFilter := r.URL.Query().Get("status")
	searchQuery := r.URL.Query().Get("search")

	baseQuery := `
		SELECT 
			i.id, 
			i.id_products, 
			i.asset_code, 
			i.status,
			COALESCE(p.product_name, '') AS product_name
		FROM item_instances i
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	if strings.TrimSpace(productIDFilter) != "" {
		if pid, err := strconv.Atoi(productIDFilter); err == nil {
			conditions = append(conditions, "i.id_products = ?")
			args = append(args, pid)
		}
	}

	if strings.TrimSpace(statusFilter) != "" {
		conditions = append(conditions, "i.status = ?")
		args = append(args, strings.TrimSpace(statusFilter))
	}

	if strings.TrimSpace(searchQuery) != "" {
		conditions = append(conditions, "(i.asset_code LIKE ? OR p.product_name LIKE ?)")
		term := "%" + strings.TrimSpace(searchQuery) + "%"
		args = append(args, term, term)
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY i.id ASC"

	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data item: " + err.Error()})
		return
	}
	defer rows.Close()

	var items []models.ItemInstance
	for rows.Next() {
		var item models.ItemInstance
		var prodName sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.IDProducts,
			&item.AssetCode,
			&item.Status,
			&prodName,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data item: " + err.Error()})
			return
		}
		item.ProductName = prodName.String
		item.IDUnit = item.ID // Dukungan penamaan ERD (id_unit)
		items = append(items, item)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data unit item",
		"total":   len(items),
		"data":    items,
	})
}

// GetItemInstanceByID mengambil detail 1 unit fisik berdasarkan ID
func GetItemInstanceByID(w http.ResponseWriter, r *http.Request, id string) {
	query := `
		SELECT 
			i.id, 
			i.id_products, 
			i.asset_code, 
			i.status,
			COALESCE(p.product_name, '') AS product_name
		FROM item_instances i
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		WHERE i.id = ?
	`

	var item models.ItemInstance
	var prodName sql.NullString

	err := config.DB.QueryRow(query, id).Scan(
		&item.ID,
		&item.IDProducts,
		&item.AssetCode,
		&item.Status,
		&prodName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Unit item dengan ID " + id + " tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data unit item: " + err.Error()})
		return
	}

	item.ProductName = prodName.String
	item.IDUnit = item.ID

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data unit item",
		"data":    item,
	})
}

// CreateItemInstance menambahkan unit fisik baru dengan validasi nomor aset unik & eksistensi produk
func CreateItemInstance(w http.ResponseWriter, r *http.Request) {
	var input models.CreateItemInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	productID := input.GetProductID()
	if productID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID produk (id_products) wajib diisi"})
		return
	}

	assetCode := strings.TrimSpace(input.AssetCode)
	if assetCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Kode nomor aset (asset_code) wajib diisi"})
		return
	}

	if input.Status == "" {
		input.Status = "available"
	}

	// Validasi status diperbolehkan: available, rented, maintenance
	validStatus := map[string]bool{"available": true, "rented": true, "maintenance": true}
	if !validStatus[input.Status] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Status tidak valid. Pilihan status: 'available', 'rented', 'maintenance'"})
		return
	}

	// 1. Cek apakah produk benar-benar ada di database
	var prodName string
	errProd := config.DB.QueryRow("SELECT product_name FROM products WHERE id_product = ? OR id = ?", productID, productID).Scan(&prodName)
	if errProd != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Produk dengan ID %d tidak ditemukan di database", productID),
		})
		return
	}

	// 2. Cek apakah asset_code sudah pernah dipakai sebelumnya
	var existingID int
	errCheck := config.DB.QueryRow("SELECT id FROM item_instances WHERE asset_code = ?", assetCode).Scan(&existingID)
	if errCheck == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Kode aset '%s' sudah terdaftar pada unit lain", assetCode),
		})
		return
	}

	query := "INSERT INTO item_instances (id_products, asset_code, status) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, productID, assetCode, input.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menambahkan unit item: " + err.Error()})
		return
	}

	lastID, _ := result.LastInsertId()
	createdUnit := models.ItemInstance{
		ID:          uint(lastID),
		IDUnit:      uint(lastID),
		IDProducts:  productID,
		ProductName: prodName,
		AssetCode:   assetCode,
		Status:      input.Status,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Unit fisik barang berhasil ditambahkan!",
		"data":    createdUnit,
	})
}

// UpdateItemInstance memperbarui data unit fisik barang
func UpdateItemInstance(w http.ResponseWriter, r *http.Request, id string) {
	var input models.CreateItemInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	productID := input.GetProductID()
	assetCode := strings.TrimSpace(input.AssetCode)

	if productID == 0 || assetCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "id_products dan asset_code wajib diisi"})
		return
	}

	if input.Status == "" {
		input.Status = "available"
	}

	validStatus := map[string]bool{"available": true, "rented": true, "maintenance": true}
	if !validStatus[input.Status] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Status tidak valid. Pilihan: 'available', 'rented', 'maintenance'"})
		return
	}

	// Cek apakah asset_code dipakai oleh unit lain selain ID ini
	var conflictID int
	errCheck := config.DB.QueryRow("SELECT id FROM item_instances WHERE asset_code = ? AND id != ?", assetCode, id).Scan(&conflictID)
	if errCheck == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Kode aset '%s' sudah digunakan oleh unit lain", assetCode),
		})
		return
	}

	query := "UPDATE item_instances SET id_products = ?, asset_code = ?, status = ? WHERE id = ?"
	result, err := config.DB.Exec(query, productID, assetCode, input.Status, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui unit item: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unit item tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Unit item dengan ID " + id + " berhasil diperbarui!",
	})
}

// DeleteItemInstance menghapus unit item berdasarkan ID dengan proteksi sewa aktif
func DeleteItemInstance(w http.ResponseWriter, r *http.Request, id string) {
	// Proteksi: Cek apakah unit ini saat ini sedang berstatus 'rented'
	var currentStatus string
	errStatus := config.DB.QueryRow("SELECT status FROM item_instances WHERE id = ?", id).Scan(&currentStatus)
	if errStatus != nil {
		if errStatus == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Unit item tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memeriksa status unit: " + errStatus.Error()})
		return
	}

	if currentStatus == "rented" {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Tidak dapat menghapus unit fisik: unit sedang dalam status sewa aktif ('rented') oleh pelanggan",
		})
		return
	}

	query := "DELETE FROM item_instances WHERE id = ?"
	result, err := config.DB.Exec(query, id)
	if err != nil {
		// Jika ada riwayat transaksi di rentals yang menolak penghapusan
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") || strings.Contains(strings.ToLower(err.Error()), "constraint") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Tidak dapat menghapus unit fisik: unit memiliki riwayat transaksi peminjaman di database",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus unit item: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Unit item tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Unit item dengan ID " + id + " berhasil dihapus!",
	})
}
