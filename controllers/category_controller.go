package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"belajar_go/config"
	"belajar_go/models"
)

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := strings.TrimPrefix(r.URL.Path, "/api/category/")

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := "UPDATE category SET category_name = ? WHERE id_category = ?"
	_, err := config.DB.Exec(query, input.Name, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal update database: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Kategori dengan ID " + id + " berhasil diperbarui di database!",
	})
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := strings.TrimPrefix(r.URL.Path, "/api/category/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID kategori wajib diisi"})
		return
	}
	query := "DELETE FROM category WHERE id_category = ?"
	result, err := config.DB.Exec(query, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus data: " + err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Kategori tidak ditemukan"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Kategori dengan ID " + id + " berhasil dihapus!",
	})
}

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

	case http.MethodPost:

	case http.MethodPut:

		UpdateCategory(w, r)
	case http.MethodDelete:

		DeleteCategory(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(category.Name) == "" {
		http.Error(w, `{"message":"Nama kategori tidak boleh kosong"}`, http.StatusBadRequest)
		return
	}

	query := "INSERT INTO categories (name) VALUES (?)"
	result, err := config.DB.Exec(query, category.Name)
	if err != nil {
		http.Error(w, `{"message":"Gagal menyimpan data ke database: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	lastInsertID, _ := result.LastInsertId()
	category.ID = uint(lastInsertID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func GetAllCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, name FROM categories")
	if err != nil {
		http.Error(w, `{"message":"Gagal mengambil data kategori: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := make([]models.Category, 0)

	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			http.Error(w, `{"message":"Gagal membaca data"}`, http.StatusInternalServerError)
			return
		}
		categories = append(categories, cat)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, `{"message":"Error saat iterasi data: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(categories)
}

func GetCategoryByID(w http.ResponseWriter, r *http.Request, id string) {
	var category models.Category

	query := "SELECT id, name FROM categories WHERE id = ?"
	err := config.DB.QueryRow(query, id).Scan(&category.ID, &category.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"message":"Kategori tidak ditemukan"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"message":"Gagal mengambil data kategori: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(category)
}
