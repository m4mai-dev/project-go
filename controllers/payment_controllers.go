package controllers

import (
	"belajar_go/config"
	"belajar_go/models"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		GetPayments(w, r)
	case http.MethodPost:
		CreatePayment(w, r)
	case http.MethodPut:
		UpdatePayment(w, r)
	case http.MethodDelete:
		DeletePayment(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

func GetPayments(w http.ResponseWriter, r *http.Request) {
	// Ganti id_payment menjadi id
	query := "SELECT id, id_rentals, payment_date, amount, payment_method, payment_proof_url, payment_status FROM payment"

	rows, err := config.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data pembayaran: " + err.Error()})
		return
	}
	defer rows.Close()

	var payments []models.Payment

	for rows.Next() {
		var p models.Payment
		err := rows.Scan(
			&p.IDPayment,
			&p.IDRentals,
			&p.PaymentDate,
			&p.Amount,
			&p.PaymentMethod,
			&p.PaymentProofURL,
			&p.PaymentStatus,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data pembayaran: " + err.Error()})
			return
		}
		payments = append(payments, p)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data pembayaran",
		"data":    payments,
	})
}

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var input models.Payment
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	if input.PaymentDate == "" {
		input.PaymentDate = time.Now().Format("2006-01-02 15:04:05")
	} else {
		t, err := time.Parse(time.RFC3339, input.PaymentDate)
		if err == nil {
			input.PaymentDate = t.Format("2006-01-02 15:04:05")
		} else {
			// Coba parse format sederhana YYYY-MM-DD jika input dari user berupa "2026-08-12"
			t2, err2 := time.Parse("2006-01-02", input.PaymentDate)
			if err2 == nil {
				input.PaymentDate = t2.Format("2006-01-02 15:04:05")
			}
		}
	}

	if input.PaymentStatus == "" {
		input.PaymentStatus = "completed"
	}

	query := "INSERT INTO payment (id_rentals, payment_date, amount, payment_method, payment_proof_url, payment_status) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := config.DB.Exec(query, input.IDRentals, input.PaymentDate, input.Amount, input.PaymentMethod, input.PaymentProofURL, input.PaymentStatus)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mencatat pembayaran: " + err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	input.IDPayment = int(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Pembayaran berhasil dicatat!",
		"data":    input,
	})
}

func UpdatePayment(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/payments/")
	id = strings.TrimPrefix(id, "/payment/")

	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID payment harus diisi"})
		return
	}

	var input models.Payment
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := "UPDATE payment SET payment_status = ?, payment_proof_url = ? WHERE id = ?"
	result, err := config.DB.Exec(query, input.PaymentStatus, input.PaymentProofURL, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui pembayaran: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Data pembayaran tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Data pembayaran dengan ID " + id + " berhasil diperbarui!",
	})
}

func DeletePayment(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/payments/")
	id = strings.TrimPrefix(id, "/payment/")

	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID payment harus diisi"})
		return
	}

	query := "DELETE FROM payment WHERE id = ?"
	result, err := config.DB.Exec(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus pembayaran: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Data pembayaran tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Data pembayaran dengan ID " + id + " berhasil dihapus!",
	})
}
