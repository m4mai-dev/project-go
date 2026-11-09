package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"belajar_go/config"
	"belajar_go/models"
)

// Helper untuk mengekstrak ID dari URL path (mendukung /category/{id} dan /api/category/{id})
func extractCategoryID(path string) string {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	// Jika bagian terakhir adalah nama endpoint ('category'), berarti tidak ada ID
	if last == "category" {
		return ""
	}
	return last
}

// CategoryHandler menangani seluruh metode HTTP untuk /category dan /api/category/
func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := extractCategoryID(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if id != "" {
			GetCategoryByID(w, r, id)
		} else {
			GetAllCategories(w, r)
		}
	case http.MethodPost:
		CreateCategory(w, r)
	case http.MethodPut:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID kategori wajib diisi pada URL path"})
			return
		}
		UpdateCategory(w, r, id)
	case http.MethodDelete:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID kategori wajib diisi pada URL path"})
			return
		}
		DeleteCategory(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

// GetAllCategories mengambil semua data kategori dari tabel `category`
func GetAllCategories(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id_category, category_name FROM category ORDER BY id_category ASC"
	rows, err := config.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data kategori: " + err.Error()})
		return
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data kategori: " + err.Error()})
			return
		}
		categories = append(categories, cat)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data kategori",
		"data":    categories,
	})
}

// GetCategoryByID mengambil 1 kategori berdasarkan id_category
func GetCategoryByID(w http.ResponseWriter, r *http.Request, id string) {
	var category models.Category
	query := "SELECT id_category, category_name FROM category WHERE id_category = ?"
	err := config.DB.QueryRow(query, id).Scan(&category.ID, &category.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Kategori tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data kategori: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data kategori",
		"data":    category,
	})
}

// CreateCategory menambahkan kategori baru ke tabel `category`
func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input models.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	categoryName := strings.TrimSpace(input.Value())
	if categoryName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nama kategori tidak boleh kosong (field 'category_name' atau 'name')"})
		return
	}

	query := "INSERT INTO category (category_name) VALUES (?)"
	result, err := config.DB.Exec(query, categoryName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan data kategori: " + err.Error()})
		return
	}

	lastInsertID, _ := result.LastInsertId()
	category := models.Category{
		ID:   uint(lastInsertID),
		Name: categoryName,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Kategori berhasil ditambahkan!",
		"data":    category,
	})
}

// UpdateCategory memperbarui nama kategori berdasarkan id_category
func UpdateCategory(w http.ResponseWriter, r *http.Request, id string) {
	var input models.CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	categoryName := strings.TrimSpace(input.Value())
	if categoryName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nama kategori tidak boleh kosong"})
		return
	}

	query := "UPDATE category SET category_name = ? WHERE id_category = ?"
	result, err := config.DB.Exec(query, categoryName, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui database: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Kategori dengan ID " + id + " tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Kategori dengan ID " + id + " berhasil diperbarui!",
		"data": models.Category{
			ID:   uint(rowsAffected),
			Name: categoryName,
		},
	})
}

// DeleteCategory menghapus kategori berdasarkan id_category
func DeleteCategory(w http.ResponseWriter, r *http.Request, id string) {
	query := "DELETE FROM category WHERE id_category = ?"
	result, err := config.DB.Exec(query, id)
	if err != nil {
		// Menangani Foreign Key Constraint jika ada produk yang masih terkait
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") || strings.Contains(strings.ToLower(err.Error()), "constraint") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Gagal menghapus kategori: masih terdapat produk yang terhubung dengan kategori ini. Hapus atau pindahkan produk terkait terlebih dahulu.",
			})
			return
		}
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
