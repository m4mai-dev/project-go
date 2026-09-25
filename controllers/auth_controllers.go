package controllers

import (
	"encoding/json"
	"net/http"

	"belajar_go/config"
	"belajar_go/models"
	"belajar_go/utils"

	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input models.RegisterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Data tidak valid", http.StatusBadRequest)
		return
	}

	if input.Username == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "Username, email, dan password wajib diisi", http.StatusBadRequest)
		return
	}

	if input.Role == "" {
		input.Role = "customer"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := "INSERT INTO users (username, email, phone, address, password_hash, role) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = config.DB.Exec(query, input.Username, input.Email, input.Phone, input.Address, string(hashedPassword), input.Role)
	if err != nil {
		http.Error(w, "Gagal mendaftar: username/email mungkin sudah terdaftar", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Registrasi berhasil!",
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	var storedPassword string
	var userID int
	var username string 
	var role string

	query := "SELECT id_user, username, password_hash, role FROM users WHERE email = ?"

	err := config.DB.QueryRow(query, input.Email).Scan(&userID, &username, &storedPassword, &role)

	if err != nil {
		http.Error(w, "Email atau password salah", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(input.Password))
	if err != nil {
		http.Error(w, "Email atau password salah", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(userID, username, role)
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login berhasil! ",
		"token":   token,
		"user": map[string]interface{}{
			"id":       userID,
			"username": username,
			"role":     role,
		},
	})
}
