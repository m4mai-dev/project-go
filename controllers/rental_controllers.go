package controllers

import (
	"belajar_go/config"
	"belajar_go/models"
	"belajar_go/services"
	"encoding/json"
	"net/http"
	"strings"
)

func RentalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		GetRentals(w, r)
	case http.MethodPost:
		CreateRental(w, r)
	case http.MethodPut:
		UpdateRental(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan"})
	}
}

func GetRentals(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, id_user, id_unit, delivery_address, start_date, return_date, 
	          actual_return_date, rental_fee, late_fee, rental_status FROM rentals`

	rows, err := config.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal mengambil data rental: " + err.Error()})
		return
	}
	defer rows.Close()

	var rentals []models.Rental

	for rows.Next() {
		var rental models.Rental
		err := rows.Scan(
			&rental.ID,
			&rental.IDUser,
			&rental.IDUnit,
			&rental.DeliveryAddress,
			&rental.StartDate,
			&rental.ReturnDate,
			&rental.ActualReturnDate,
			&rental.RentalFee,
			&rental.LateFee,
			&rental.RentalStatus,
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": "Gagal membaca data rental: " + err.Error()})
			return
		}
		rentals = append(rentals, rental)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil mengambil data transaksi rental",
		"data":    rentals,
	})
}

func CreateRental(w http.ResponseWriter, r *http.Request) {
	var input models.CreateRentalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	newRental, err := services.CreateRental(input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Transaksi rental berhasil dibuat!",
		"data":    newRental,
	})
}

func UpdateRental(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/rental/")
	if id == "" || id == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "ID rental harus diisi"})
		return
	}

	var input struct {
		ActualReturnDate string  `json:"actual_return_date"`
		LateFee          float64 `json:"late_fee"`
		RentalStatus     string  `json:"rental_status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Format JSON tidak valid"})
		return
	}

	query := `UPDATE rentals SET actual_return_date = ?, late_fee = ?, rental_status = ? WHERE id = ?`
	result, err := config.DB.Exec(query, input.ActualReturnDate, input.LateFee, input.RentalStatus, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Gagal memperbarui transaksi rental: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "Data rental tidak ditemukan"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transaksi rental dengan ID " + id + " berhasil diperbarui!",
	})
}
