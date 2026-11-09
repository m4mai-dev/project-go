package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"belajar_go/config"
	"belajar_go/models"
)

// CreateRental memproses pesanan sewa baru dengan validasi anti-double booking, durasi sewa, dan transaksi ACID
func CreateRental(input models.CreateRentalInput) (models.Rental, error) {
	if input.IDUser == 0 {
		return models.Rental{}, errors.New("id_user wajib diisi")
	}
	if input.IDUnit == 0 {
		return models.Rental{}, errors.New("id_unit wajib diisi")
	}
	if input.DeliveryAddress == "" {
		return models.Rental{}, errors.New("delivery_address wajib diisi")
	}
	if input.StartDate == "" || input.ReturnDate == "" {
		return models.Rental{}, errors.New("start_date dan return_date wajib diisi (format: YYYY-MM-DD)")
	}

	// 1. Validasi parsing tanggal
	startTime, errStart := time.Parse("2006-01-02", input.StartDate)
	retTime, errRet := time.Parse("2006-01-02", input.ReturnDate)
	if errStart != nil || errRet != nil {
		return models.Rental{}, errors.New("format tanggal tidak valid, gunakan format YYYY-MM-DD (contoh: 2026-11-15)")
	}

	if retTime.Before(startTime) {
		return models.Rental{}, errors.New("tanggal pengembalian (return_date) tidak boleh lebih awal dari tanggal mulai (start_date)")
	}

	// Hitung durasi hari sewa (minimal 1 hari)
	days := int(retTime.Sub(startTime).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	// 2. Cek apakah user terdaftar di database
	var username string
	errUser := config.DB.QueryRow("SELECT username FROM users WHERE id_user = ?", input.IDUser).Scan(&username)
	if errUser != nil {
		return models.Rental{}, fmt.Errorf("user dengan ID %d tidak ditemukan di sistem", input.IDUser)
	}

	// 3. Mulai Database Transaction untuk menjamin tidak terjadi double-booking secara konkuren
	tx, err := config.DB.Begin()
	if err != nil {
		return models.Rental{}, fmt.Errorf("gagal memulai database transaction: %w", err)
	}

	// Kunci baris unit fisik untuk update ketersediaan (FOR UPDATE)
	var unitStatus, assetCode, productName string
	var pricingPDay float64
	unitQuery := `
		SELECT 
			i.status, 
			i.asset_code,
			p.pricing_p_day,
			p.product_name
		FROM item_instances i
		JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		WHERE i.id = ?
		FOR UPDATE
	`
	errUnit := tx.QueryRow(unitQuery, input.IDUnit).Scan(&unitStatus, &assetCode, &pricingPDay, &productName)
	if errUnit != nil {
		tx.Rollback()
		if errUnit == sql.ErrNoRows {
			return models.Rental{}, fmt.Errorf("unit fisik dengan ID %d tidak ditemukan", input.IDUnit)
		}
		return models.Rental{}, fmt.Errorf("gagal memeriksa ketersediaan unit: %w", errUnit)
	}

	// Validasi ketersediaan unit fisik (Anti Double-Booking Protection)
	if unitStatus != "available" {
		tx.Rollback()
		return models.Rental{}, fmt.Errorf("unit fisik barang dengan kode aset '%s' saat ini TIDAK TERSEDIA (status saat ini: '%s')", assetCode, unitStatus)
	}

	// Hitung biaya sewa otomatis jika tidak dimasukkan secara manual
	if input.RentalFee <= 0 {
		input.RentalFee = float64(days) * pricingPDay
	}

	// 4. Masukkan data ke tabel rentals dengan status default 'pending'
	insertQuery := `
		INSERT INTO rentals (id_user, id_unit, delivery_address, start_date, return_date, rental_fee, rental_status) 
		VALUES (?, ?, ?, ?, ?, ?, 'pending')
	`
	result, err := tx.Exec(insertQuery, input.IDUser, input.IDUnit, input.DeliveryAddress, input.StartDate, input.ReturnDate, input.RentalFee)
	if err != nil {
		tx.Rollback()
		return models.Rental{}, fmt.Errorf("gagal mencatat transaksi rental: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return models.Rental{}, fmt.Errorf("gagal mengambil ID transaksi rental: %w", err)
	}

	// 5. Ubah status unit fisik di item_instances menjadi 'rented'
	updateUnitQuery := `UPDATE item_instances SET status = 'rented' WHERE id = ?`
	_, err = tx.Exec(updateUnitQuery, input.IDUnit)
	if err != nil {
		tx.Rollback()
		return models.Rental{}, fmt.Errorf("gagal mengunci status unit fisik menjadi rented: %w", err)
	}

	// Simpan seluruh perubahan secara permanen (Commit)
	if errCommit := tx.Commit(); errCommit != nil {
		return models.Rental{}, fmt.Errorf("gagal menyimpan transaksi sewa: %w", errCommit)
	}

	newRental := models.Rental{
		ID:              uint(lastID),
		IDRentals:       uint(lastID),
		IDUser:          input.IDUser,
		Username:        username,
		IDUnit:          input.IDUnit,
		AssetCode:       assetCode,
		ProductName:     productName,
		DeliveryAddress: input.DeliveryAddress,
		StartDate:       input.StartDate,
		ReturnDate:      input.ReturnDate,
		RentalFee:       input.RentalFee,
		LateFee:         0,
		RentalStatus:    "pending",
	}

	return newRental, nil
}
