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
	"belajar_go/services"
)

// Helper untuk mengekstrak ID dari URL path /rental/{id} atau /api/rental/{id}
func extractRentalID(path string) string {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == "rental" {
		return ""
	}
	return last
}

// RentalHandler mengelola routing transaksi peminjaman dan pengembalian barang
func RentalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := extractRentalID(r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		if id != "" {
			GetRentalByID(w, r, id)
		} else {
			GetRentals(w, r)
		}
	case http.MethodPost:
		CreateRental(w, r)
	case http.MethodPut:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID transaksi rental harus disertakan pada URL path"})
			return
		}
		UpdateRental(w, r, id)
	case http.MethodDelete:
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": "ID transaksi rental harus disertakan pada URL path"})
			return
		}
		DeleteRental(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

// GetRentals mengambil seluruh transaksi rental dengan dukungan filter ?status= dan ?user_id=
func GetRentals(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	userIDFilter := r.URL.Query().Get("user_id")

	baseQuery := `
		SELECT 
			r.id, 
			r.id_user, 
			COALESCE(u.username, '') AS username,
			COALESCE(u.email, '') AS customer_email,
			r.id_unit, 
			COALESCE(i.asset_code, '') AS asset_code,
			COALESCE(p.product_name, '') AS product_name,
			r.delivery_address, 
			DATE_FORMAT(r.start_date, '%Y-%m-%d') AS start_date, 
			DATE_FORMAT(r.return_date, '%Y-%m-%d') AS return_date, 
			DATE_FORMAT(r.actual_return_date, '%Y-%m-%d') AS actual_return_date, 
			r.rental_fee, 
			r.late_fee, 
			r.rental_status,
			COALESCE(pm.payment_status, 'unpaid') AS payment_status
		FROM rentals r
		LEFT JOIN users u ON u.id_user = r.id_user
		LEFT JOIN item_instances i ON i.id = r.id_unit
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		LEFT JOIN payment pm ON pm.id_rentals = r.id
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	if strings.TrimSpace(statusFilter) != "" {
		conditions = append(conditions, "r.rental_status = ?")
		args = append(args, strings.TrimSpace(statusFilter))
	}

	if strings.TrimSpace(userIDFilter) != "" {
		if uid, err := strconv.Atoi(userIDFilter); err == nil {
			conditions = append(conditions, "r.id_user = ?")
			args = append(args, uid)
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY r.id DESC"

	rows, err := config.DB.Query(baseQuery, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data rental: " + err.Error()})
		return
	}
	defer rows.Close()

	var rentals []models.Rental
	for rows.Next() {
		var rental models.Rental
		var actualRet sql.NullString

		err := rows.Scan(
			&rental.ID,
			&rental.IDUser,
			&rental.Username,
			&rental.CustomerEmail,
			&rental.IDUnit,
			&rental.AssetCode,
			&rental.ProductName,
			&rental.DeliveryAddress,
			&rental.StartDate,
			&rental.ReturnDate,
			&actualRet,
			&rental.RentalFee,
			&rental.LateFee,
			&rental.RentalStatus,
			&rental.PaymentStatus,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data rental: " + err.Error()})
			return
		}

		if actualRet.Valid {
			str := actualRet.String
			rental.ActualReturnDate = &str
		}
		rental.IDRentals = rental.ID
		rentals = append(rentals, rental)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data transaksi rental",
		"total":   len(rentals),
		"data":    rentals,
	})
}

// GetRentalByID mengambil detail 1 transaksi rental berdasarkan ID
func GetRentalByID(w http.ResponseWriter, r *http.Request, id string) {
	query := `
		SELECT 
			r.id, 
			r.id_user, 
			COALESCE(u.username, '') AS username,
			COALESCE(u.email, '') AS customer_email,
			r.id_unit, 
			COALESCE(i.asset_code, '') AS asset_code,
			COALESCE(p.product_name, '') AS product_name,
			r.delivery_address, 
			DATE_FORMAT(r.start_date, '%Y-%m-%d') AS start_date, 
			DATE_FORMAT(r.return_date, '%Y-%m-%d') AS return_date, 
			DATE_FORMAT(r.actual_return_date, '%Y-%m-%d') AS actual_return_date, 
			r.rental_fee, 
			r.late_fee, 
			r.rental_status,
			COALESCE(pm.payment_status, 'unpaid') AS payment_status
		FROM rentals r
		LEFT JOIN users u ON u.id_user = r.id_user
		LEFT JOIN item_instances i ON i.id = r.id_unit
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		LEFT JOIN payment pm ON pm.id_rentals = r.id
		WHERE r.id = ?
	`

	var rental models.Rental
	var actualRet sql.NullString

	err := config.DB.QueryRow(query, id).Scan(
		&rental.ID,
		&rental.IDUser,
		&rental.Username,
		&rental.CustomerEmail,
		&rental.IDUnit,
		&rental.AssetCode,
		&rental.ProductName,
		&rental.DeliveryAddress,
		&rental.StartDate,
		&rental.ReturnDate,
		&actualRet,
		&rental.RentalFee,
		&rental.LateFee,
		&rental.RentalStatus,
		&rental.PaymentStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Transaksi rental dengan ID " + id + " tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data rental: " + err.Error()})
		return
	}

	if actualRet.Valid {
		str := actualRet.String
		rental.ActualReturnDate = &str
	}
	rental.IDRentals = rental.ID

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data transaksi rental",
		"data":    rental,
	})
}

// CreateRental membuat transaksi sewa baru
func CreateRental(w http.ResponseWriter, r *http.Request) {
	var input models.CreateRentalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	newRental, err := services.CreateRental(input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Transaksi rental berhasil dibuat! Unit fisik barang otomatis terkunci (status: rented)",
		"data":    newRental,
	})
}

// UpdateRental memperbarui status rental, mencatat pengembalian barang, menghitung denda otomatis, dan mengembalikan unit ke status available
func UpdateRental(w http.ResponseWriter, r *http.Request, id string) {
	var input models.UpdateRentalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	// 1. Ambil data rental saat ini untuk pengecekan
	var currentUnitID uint
	var scheduledReturnStr string
	var pricingPDay float64
	var currentRentalStatus string

	queryGet := `
		SELECT 
			r.id_unit, 
			DATE_FORMAT(r.return_date, '%Y-%m-%d'),
			COALESCE(p.pricing_p_day, 0),
			r.rental_status
		FROM rentals r
		LEFT JOIN item_instances i ON i.id = r.id_unit
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		WHERE r.id = ?
	`
	err := config.DB.QueryRow(queryGet, id).Scan(&currentUnitID, &scheduledReturnStr, &pricingPDay, &currentRentalStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Data rental dengan ID " + id + " tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data rental: " + err.Error()})
		return
	}

	// 2. Kalkulasi Denda Otomatis jika barang dikembalikan
	if input.ActualReturnDate != "" {
		actualDate, errActual := time.Parse("2006-01-02", input.ActualReturnDate)
		scheduledDate, errSched := time.Parse("2006-01-02", scheduledReturnStr)

		if errActual == nil && errSched == nil {
			if actualDate.After(scheduledDate) {
				daysLate := int(actualDate.Sub(scheduledDate).Hours() / 24)
				if daysLate > 0 && input.LateFee <= 0 {
					// Hitung denda otomatis: tarif sewa harian x hari keterlambatan x penalti 1.5
					input.LateFee = float64(daysLate) * pricingPDay * 1.5
				}
			}
		}
	}

	if input.RentalStatus == "" {
		input.RentalStatus = currentRentalStatus
	}

	// 3. Eksekusi pembaruan dalam database transaction
	tx, err := config.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memulai transaksi database: " + err.Error()})
		return
	}

	var updateQuery string
	var updateArgs []interface{}

	if input.ActualReturnDate != "" {
		updateQuery = `UPDATE rentals SET actual_return_date = ?, late_fee = ?, rental_status = ? WHERE id = ?`
		updateArgs = []interface{}{input.ActualReturnDate, input.LateFee, input.RentalStatus, id}
	} else {
		updateQuery = `UPDATE rentals SET late_fee = ?, rental_status = ? WHERE id = ?`
		updateArgs = []interface{}{input.LateFee, input.RentalStatus, id}
	}

	result, err := tx.Exec(updateQuery, updateArgs...)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui transaksi rental: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Data rental tidak ditemukan"})
		return
	}

	// 4. PENYELESAIAN BUG #3: Jika status sewa menjadi 'completed', 'returned', atau 'cancelled',
	// Kembalikan status unit fisik di item_instances menjadi 'available' agar dapat disewa kembali!
	unitReset := false
	if input.RentalStatus == "completed" || input.RentalStatus == "returned" || input.RentalStatus == "cancelled" {
		resetQuery := `UPDATE item_instances SET status = 'available' WHERE id = ?`
		_, errReset := tx.Exec(resetQuery, currentUnitID)
		if errReset != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengembalikan status unit fisik menjadi available: " + errReset.Error()})
			return
		}
		unitReset = true
	}

	if errCommit := tx.Commit(); errCommit != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan perubahan transaksi: " + errCommit.Error()})
		return
	}

	message := fmt.Sprintf("Transaksi rental dengan ID %s berhasil diperbarui!", id)
	if unitReset {
		message += fmt.Sprintf(" Unit fisik barang (ID: %d) telah otomatis dikembalikan ke status 'available'.", currentUnitID)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data": map[string]interface{}{
			"id":                 id,
			"actual_return_date": input.ActualReturnDate,
			"late_fee":           input.LateFee,
			"rental_status":      input.RentalStatus,
			"unit_status_reset":  unitReset,
		},
	})
}

// DeleteRental menghapus catatan rental jika transaksi dibatalkan
func DeleteRental(w http.ResponseWriter, r *http.Request, id string) {
	// Ambil id_unit untuk merestore ketersediaan jika status belum completed
	var idUnit uint
	var status string
	err := config.DB.QueryRow("SELECT id_unit, rental_status FROM rentals WHERE id = ?", id).Scan(&idUnit, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"message": "Transaksi rental tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca transaksi: " + err.Error()})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memulai database transaction: " + err.Error()})
		return
	}

	// Kembalikan unit menjadi available jika rental belum selesai
	if status == "pending" || status == "active" || status == "confirmed" {
		_, _ = tx.Exec("UPDATE item_instances SET status = 'available' WHERE id = ?", idUnit)
	}

	// Hapus catatan pembayaran jika ada
	_, _ = tx.Exec("DELETE FROM payment WHERE id_rentals = ?", id)

	_, err = tx.Exec("DELETE FROM rentals WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menghapus transaksi rental: " + err.Error()})
		return
	}

	if errCommit := tx.Commit(); errCommit != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal menyimpan perubahan: " + errCommit.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transaksi rental dengan ID " + id + " berhasil dibatalkan dan dihapus!",
	})
}
