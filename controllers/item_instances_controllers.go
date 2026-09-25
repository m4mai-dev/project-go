package controllers

import (
	"belajar_go/config"
	"belajar_go/models"
	"encoding/json"
	"net/http"
	"strings"
)

func ItemInstanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		GetItemInstances(w, r)
	case http.MethodPost:
		CreateItemInstance(w, r)
	case http.MethodPut:
		UpdateItemInstance(w, r)
	case http.MethodDelete:
		DeleteItemInstance(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

func GetItemInstances(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, id_products, asset_code, status FROM item_instances"

	rows, err := config.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data item: " + err.Error()})
		return
	}
	defer rows.Close()

	var items []models.ItemInstance

	for rows.Next() {
		var item models.ItemInstance
		err := rows.Scan(
			&item.ID,
			&item.IDProducts,
			&item.AssetCode,
			&item.Status,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data item: " + err.Error()})
			return
		}
		items = append(items, item)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data unit item",
		"data":    items,
	})
}

func CreateItemInstance(w http.ResponseWriter, r *http.Request) {
	var input models.CreateItemInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := "INSERT INTO item_instances (id_products, asset_code, status) VALUES (?, ?, ?)"
	result, err := config.DB.Exec(query, input.IDProducts, input.AssetCode, input.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menambahkan unit item: " + err.Error()})
		return
	}

	lastID, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Unit item berhasil ditambahkan!",
		"id":      lastID,
	})
}

func UpdateItemInstance(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/item-instances/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID item instance harus diisi"})
		return
	}

	var input models.CreateItemInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := "UPDATE item_instances SET id_products = ?, asset_code = ?, status = ? WHERE id = ?"
	result, err := config.DB.Exec(query, input.IDProducts, input.AssetCode, input.Status, id)
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

// DeleteItemInstance untuk menghapus unit item berdasarkan ID
func DeleteItemInstance(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil ID dari URL path (misal: /item-instances/1 -> "1")
	id := strings.TrimPrefix(r.URL.Path, "/item-instances/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID item instance harus diisi"})
		return
	}

	// 2. Eksekusi query DELETE
	query := "DELETE FROM item_instances WHERE id = ?"
	result, err := config.DB.Exec(query, id)
	if err != nil {
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