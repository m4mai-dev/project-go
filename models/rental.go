package models

type CreateRentalInput struct {
	IDUser          uint    `json:"id_user"`
	IDUnit          uint    `json:"id_unit"`
	DeliveryAddress string  `json:"delivery_address"`
	StartDate       string  `json:"start_date"` // Format: YYYY-MM-DD
	ReturnDate      string  `json:"return_date"` // Format: YYYY-MM-DD
	RentalFee       float64 `json:"rental_fee"`  // Opsional: jika 0, dihitung otomatis
}

type UpdateRentalInput struct {
	ActualReturnDate string  `json:"actual_return_date"` // Format: YYYY-MM-DD
	RentalStatus     string  `json:"rental_status"`      // pending, confirmed, active, completed, cancelled
	LateFee          float64 `json:"late_fee"`           // Opsional: jika 0, dihitung otomatis jika terlambat
}

type Rental struct {
	ID               uint    `json:"id"`
	IDRentals        uint    `json:"id_rentals,omitempty"` // Kompatibilitas ERD
	IDUser           uint    `json:"id_user"`
	Username         string  `json:"username,omitempty"`
	CustomerEmail    string  `json:"customer_email,omitempty"`
	IDUnit           uint    `json:"id_unit"`
	AssetCode        string  `json:"asset_code,omitempty"`
	ProductName      string  `json:"product_name,omitempty"`
	DeliveryAddress  string  `json:"delivery_address"`
	StartDate        string  `json:"start_date"`
	ReturnDate       string  `json:"return_date"`
	ActualReturnDate *string `json:"actual_return_date"`
	RentalFee        float64 `json:"rental_fee"`
	LateFee          float64 `json:"late_fee"`
	RentalStatus     string  `json:"rental_status"`
	PaymentStatus    string  `json:"payment_status,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
}
