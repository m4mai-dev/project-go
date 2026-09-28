package controllers

import (
	"belajar_go/config"
	"belajar_go/models"
	"encoding/json"
	"net/http"
	"strings"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		GetUsers(w, r)
	case http.MethodPut:
		UpdateUser(w, r)
	case http.MethodDelete:
		DeleteUser(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id_user, username, email, phone, address, role FROM users"

	rows, err := config.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data user: " + err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Phone,
			&user.Address,
			&user.Role,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data user: " + err.Error()})
			return
		}
		users = append(users, user)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data user",
		"data":    users,
	})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID user harus diisi"})
		return
	}

	var input models.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := "UPDATE users SET username = ?, email = ?, phone = ?, address = ?, role = ? WHERE id_user = ?"
	result, err := config.DB.Exec(query, input.Username, input.Email, input.Phone, input.Address, input.Role, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui data user: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "User tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Data user dengan ID " + id + " berhasil diperbarui!",
	})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID user harus diisi"})
		return
	}

	query := "DELETE FROM users WHERE id_user = ?"
	result, err := config.DB.Exec(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus user: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "User tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User dengan ID " + id + " berhasil dihapus!",
	})
}