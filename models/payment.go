package models

type Payment struct {
	ID              int     `json:"id"`
	IDPayment       int     `json:"id_payment,omitempty"` // Kompatibilitas ERD
	IDRentals       int     `json:"id_rentals"`
	PaymentDate     string  `json:"payment_date"`
	Amount          float64 `json:"amount"`
	PaymentMethod   string  `json:"payment_method"`
	PaymentProofURL string  `json:"payment_proof_url"`
	PaymentStatus   string  `json:"payment_status"` // 'pending', 'completed', 'rejected'
	RentalStatus    string  `json:"rental_status,omitempty"`
	CustomerName    string  `json:"customer_name,omitempty"`
	ProductName     string  `json:"product_name,omitempty"`
}

type CreatePaymentInput struct {
	IDRentals       int     `json:"id_rentals"`
	PaymentDate     string  `json:"payment_date"`
	Amount          float64 `json:"amount"`
	PaymentMethod   string  `json:"payment_method"`
	PaymentProofURL string  `json:"payment_proof_url"`
	PaymentStatus   string  `json:"payment_status"`
}

type UpdatePaymentStatusInput struct {
	PaymentStatus   string  `json:"payment_status"` // 'pending', 'completed', 'rejected'
	PaymentProofURL string  `json:"payment_proof_url"`
	Amount          float64 `json:"amount,omitempty"`
}

func (Payment) TableName() string {
	return "payment"
}
