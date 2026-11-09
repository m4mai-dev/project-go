package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"belajar_go/config"
	"belajar_go/models"
)

// Helper untuk mengekstrak ID dari URL path /payments/{id} atau /api/payments/{id}
func extractPaymentID(path string) string {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == "payments" || last == "payment" {
		return ""
	}
	return last
}

// PaymentHandler mengelola seluruh metode HTTP untuk modul pembayaran
func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := extractPaymentID(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if id != "" {
			GetPaymentByID(w, r, id)
		} else {
			GetPayments(w, r)
		}
	case http.MethodPost:
		CreatePayment(w, r)
	case http.MethodPut:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID payment harus disertakan pada URL path"})
			return
		}
		UpdatePayment(w, r, id)
	case http.MethodDelete:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID payment harus disertakan pada URL path"})
			return
		}
		DeletePayment(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

// GetPayments mengambil data pembayaran dengan dukungan filter ?rental_id= dan ?status=
func GetPayments(w http.ResponseWriter, r *http.Request) {
	rentalIDFilter := r.URL.Query().Get("rental_id")
	statusFilter := r.URL.Query().Get("status")

	baseQuery := `
		SELECT 
			pm.id, 
			pm.id_rentals, 
			DATE_FORMAT(pm.payment_date, '%Y-%m-%d %H:%i:%s') AS payment_date, 
			pm.amount, 
			pm.payment_method, 
			pm.payment_proof_url, 
			pm.payment_status,
			COALESCE(r.rental_status, '') AS rental_status,
			COALESCE(u.username, '') AS customer_name,
			COALESCE(p.product_name, '') AS product_name
		FROM payment pm
		LEFT JOIN rentals r ON r.id = pm.id_rentals
		LEFT JOIN users u ON u.id_user = r.id_user
		LEFT JOIN item_instances i ON i.id = r.id_unit
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	if strings.TrimSpace(rentalIDFilter) != "" {
		if rid, err := strconv.Atoi(rentalIDFilter); err == nil {
			conditions = append(conditions, "pm.id_rentals = ?")
			args = append(args, rid)
		}
	}

	if strings.TrimSpace(statusFilter) != "" {
		conditions = append(conditions, "pm.payment_status = ?")
		args = append(args, strings.TrimSpace(statusFilter))
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY pm.id DESC"

	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data pembayaran: " + err.Error()})
		return
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var p models.Payment
		var proofURL, rStatus, cName, pName sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.IDRentals,
			&p.PaymentDate,
			&p.Amount,
			&p.PaymentMethod,
			&proofURL,
			&p.PaymentStatus,
			&rStatus,
			&cName,
			&pName,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data pembayaran: " + err.Error()})
			return
		}

		p.PaymentProofURL = proofURL.String
		p.RentalStatus = rStatus.String
		p.CustomerName = cName.String
		p.ProductName = pName.String
		p.IDPayment = p.ID
		payments = append(payments, p)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data pembayaran",
		"total":   len(payments),
		"data":    payments,
	})
}

// GetPaymentByID mengambil detail 1 catatan pembayaran berdasarkan ID
func GetPaymentByID(w http.ResponseWriter, r *http.Request, id string) {
	query := `
		SELECT 
			pm.id, 
			pm.id_rentals, 
			DATE_FORMAT(pm.payment_date, '%Y-%m-%d %H:%i:%s') AS payment_date, 
			pm.amount, 
			pm.payment_method, 
			pm.payment_proof_url, 
			pm.payment_status,
			COALESCE(r.rental_status, '') AS rental_status,
			COALESCE(u.username, '') AS customer_name,
			COALESCE(p.product_name, '') AS product_name
		FROM payment pm
		LEFT JOIN rentals r ON r.id = pm.id_rentals
		LEFT JOIN users u ON u.id_user = r.id_user
		LEFT JOIN item_instances i ON i.id = r.id_unit
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		WHERE pm.id = ?
	`

	var p models.Payment
	var proofURL, rStatus, cName, pName sql.NullString

	err := config.DB.QueryRow(query, id).Scan(
		&p.ID,
		&p.IDRentals,
		&p.PaymentDate,
		&p.Amount,
		&p.PaymentMethod,
		&proofURL,
		&p.PaymentStatus,
		&rStatus,
		&cName,
		&pName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Data pembayaran dengan ID " + id + " tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data pembayaran: " + err.Error()})
		return
	}

	p.PaymentProofURL = proofURL.String
	p.RentalStatus = rStatus.String
	p.CustomerName = cName.String
	p.ProductName = pName.String
	p.IDPayment = p.ID

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data pembayaran",
		"data":    p,
	})
}

// CreatePayment mencatat transaksi pembayaran baru dari pelanggan
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var input models.CreatePaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	if input.IDRentals == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "id_rentals wajib diisi"})
		return
	}

	if input.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Jumlah nominal pembayaran (amount) harus lebih dari 0"})
		return
	}

	if strings.TrimSpace(input.PaymentMethod) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Metode pembayaran (payment_method) wajib diisi (contoh: transfer_bca, qris, tunai)"})
		return
	}

	// 1. Verifikasi apakah transaksi rental ada di database
	var rentalFee, lateFee float64
	var currentRentalStatus string
	errRental := config.DB.QueryRow("SELECT rental_fee, late_fee, rental_status FROM rentals WHERE id = ?", input.IDRentals).Scan(&rentalFee, &lateFee, &currentRentalStatus)
	if errRental != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": fmt.Sprintf("Transaksi rental dengan ID %d tidak ditemukan di database", input.IDRentals),
		})
		return
	}

	// Format tanggal pembayaran
	if input.PaymentDate == "" {
		input.PaymentDate = time.Now().Format("2006-01-02 15:04:05")
	} else {
		if t, err := time.Parse(time.RFC3339, input.PaymentDate); err == nil {
			input.PaymentDate = t.Format("2006-01-02 15:04:05")
		} else if t2, err2 := time.Parse("2006-01-02", input.PaymentDate); err2 == nil {
			input.PaymentDate = t2.Format("2006-01-02 15:04:05")
		}
	}

	if input.PaymentStatus == "" {
		input.PaymentStatus = "pending" // Menunggu konfirmasi admin
	}

	query := "INSERT INTO payment (id_rentals, payment_date, amount, payment_method, payment_proof_url, payment_status) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := config.DB.Exec(query, input.IDRentals, input.PaymentDate, input.Amount, input.PaymentMethod, input.PaymentProofURL, input.PaymentStatus)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mencatat pembayaran: " + err.Error()})
		return
	}

	id, _ := result.LastInsertId()

	createdPayment := models.Payment{
		ID:              int(id),
		IDPayment:       int(id),
		IDRentals:       input.IDRentals,
		PaymentDate:     input.PaymentDate,
		Amount:          input.Amount,
		PaymentMethod:   input.PaymentMethod,
		PaymentProofURL: input.PaymentProofURL,
		PaymentStatus:   input.PaymentStatus,
		RentalStatus:    currentRentalStatus,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Pembayaran berhasil dicatat! Menunggu verifikasi admin.",
		"data":    createdPayment,
	})
}

// UpdatePayment memvalidasi pembayaran oleh admin dan OTOMATIS memperbarui status rental menjadi 'confirmed'
func UpdatePayment(w http.ResponseWriter, r *http.Request, id string) {
	var input models.UpdatePaymentStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	// Ambil data pembayaran saat ini beserta id_rentals
	var idRentals int
	var currentStatus string
	err := config.DB.QueryRow("SELECT id_rentals, payment_status FROM payment WHERE id = ?", id).Scan(&idRentals, &currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Data pembayaran tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data pembayaran: " + err.Error()})
		return
	}

	if input.PaymentStatus == "" {
		input.PaymentStatus = currentStatus
	}

	tx, err := config.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memulai database transaction: " + err.Error()})
		return
	}

	updateQuery := "UPDATE payment SET payment_status = ?, payment_proof_url = ? WHERE id = ?"
	_, err = tx.Exec(updateQuery, input.PaymentStatus, input.PaymentProofURL, id)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui status pembayaran: " + err.Error()})
		return
	}

	// FITUR INTEGRASI OTOMATIS:
	// Jika admin menyetujui/memvalidasi pembayaran ('completed'), ubah status rental yang tadinya 'pending' menjadi 'confirmed'!
	rentalUpdated := false
	if input.PaymentStatus == "completed" {
		_, errRental := tx.Exec("UPDATE rentals SET rental_status = 'confirmed' WHERE id = ? AND rental_status = 'pending'", idRentals)
		if errRental == nil {
			rentalUpdated = true
		}
	}

	if errCommit := tx.Commit(); errCommit != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan perubahan transaksi: " + errCommit.Error()})
		return
	}

	msg := fmt.Sprintf("Data pembayaran dengan ID %s berhasil diperbarui menjadi '%s'!", id, input.PaymentStatus)
	if rentalUpdated {
		msg += fmt.Sprintf(" Status pesanan sewa (Rental ID: %d) telah otomatis disetujui (status: confirmed).", idRentals)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": msg,
		"data": map[string]interface{}{
			"id":             id,
			"id_rentals":     idRentals,
			"payment_status": input.PaymentStatus,
			"rental_updated": rentalUpdated,
		},
	})
}

// DeletePayment menghapus data pembayaran
func DeletePayment(w http.ResponseWriter, r *http.Request, id string) {
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
