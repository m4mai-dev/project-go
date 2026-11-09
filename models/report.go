package models

type SalesSummary struct {
	TotalRevenue          float64 `json:"total_revenue"`           // Total penerimaan dari pembayaran valid
	TotalRentalFees       float64 `json:"total_rental_fees"`      // Total biaya sewa
	TotalLateFees         float64 `json:"total_late_fees"`        // Total denda keterlambatan
	TotalTransactions     int     `json:"total_transactions"`     // Total transaksi
	TotalPendingOrders    int     `json:"total_pending_orders"`   // Pesanan belum diproses
	TotalActiveRentals    int     `json:"total_active_rentals"`   // Sedang disewa
	TotalCompletedRentals int     `json:"total_completed_rentals"`// Selesai dikembalikan
}

type StatusBreakdown struct {
	Pending   int `json:"pending"`
	Confirmed int `json:"confirmed"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
}

type TopRentedProduct struct {
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	CategoryName string  `json:"category_name"`
	RentCount    int     `json:"rent_count"`
	TotalRevenue float64 `json:"total_revenue"`
}

type RecentTransactionReport struct {
	ID           uint    `json:"id"`
	Username     string  `json:"username"`
	ProductName  string  `json:"product_name"`
	AssetCode    string  `json:"asset_code"`
	StartDate    string  `json:"start_date"`
	ReturnDate   string  `json:"return_date"`
	RentalFee    float64 `json:"rental_fee"`
	LateFee      float64 `json:"late_fee"`
	RentalStatus string  `json:"rental_status"`
}

type SalesReportResponse struct {
	Summary            SalesSummary              `json:"summary"`
	StatusBreakdown    StatusBreakdown           `json:"status_breakdown"`
	TopRentedProducts  []TopRentedProduct        `json:"top_rented_products"`
	RecentTransactions []RecentTransactionReport `json:"recent_transactions"`
}
