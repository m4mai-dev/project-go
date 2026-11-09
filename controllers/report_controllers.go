package controllers

import (
	"encoding/json"
	"net/http"

	"belajar_go/config"
	"belajar_go/models"
)

// SalesReportHandler menyediakan rekapitulasi data penjualan dan performa rental untuk Admin
func SalesReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"message": "Method tidak diizinkan. Gunakan GET untuk melihat laporan."})
		return
	}

	var report models.SalesReportResponse

	// 1. Ambil ringkasan keuangan
	// Total Revenue dari pembayaran yang statusnya 'completed'
	queryRevenue := "SELECT COALESCE(SUM(amount), 0) FROM payment WHERE payment_status = 'completed'"
	_ = config.DB.QueryRow(queryRevenue).Scan(&report.Summary.TotalRevenue)

	// Total Biaya Sewa dari seluruh transaksi yang tidak dibatalkan
	queryRentalFees := "SELECT COALESCE(SUM(rental_fee), 0) FROM rentals WHERE rental_status != 'cancelled'"
	_ = config.DB.QueryRow(queryRentalFees).Scan(&report.Summary.TotalRentalFees)

	// Total Denda Keterlambatan yang tercatat
	queryLateFees := "SELECT COALESCE(SUM(late_fee), 0) FROM rentals"
	_ = config.DB.QueryRow(queryLateFees).Scan(&report.Summary.TotalLateFees)

	// Total Transaksi Keseluruhan
	queryTotalTx := "SELECT COUNT(*) FROM rentals"
	_ = config.DB.QueryRow(queryTotalTx).Scan(&report.Summary.TotalTransactions)

	// 2. Ambil rincian transaksi per status
	queryStatus := "SELECT rental_status, COUNT(*) FROM rentals GROUP BY rental_status"
	statusRows, err := config.DB.Query(queryStatus)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var st string
			var count int
			if errScan := statusRows.Scan(&st, &count); errScan == nil {
				switch st {
				case "pending":
					report.StatusBreakdown.Pending = count
					report.Summary.TotalPendingOrders = count
				case "confirmed":
					report.StatusBreakdown.Confirmed = count
				case "active":
					report.StatusBreakdown.Active = count
					report.Summary.TotalActiveRentals += count
				case "completed", "returned":
					report.StatusBreakdown.Completed += count
					report.Summary.TotalCompletedRentals += count
				case "cancelled":
					report.StatusBreakdown.Cancelled = count
				}
			}
		}
	}

	// 3. Ambil Top 5 Produk Paling Sering Disewa
	queryTopProducts := `
		SELECT 
			COALESCE(p.id_product, p.id) AS product_id,
			p.product_name,
			COALESCE(c.category_name, '') AS category_name,
			COUNT(r.id) AS rent_count,
			COALESCE(SUM(r.rental_fee), 0) AS total_revenue
		FROM rentals r
		JOIN item_instances i ON i.id = r.id_unit
		JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		LEFT JOIN category c ON c.id_category = p.id_category
		WHERE r.rental_status != 'cancelled'
		GROUP BY product_id, p.product_name, category_name
		ORDER BY rent_count DESC
		LIMIT 5
	`
	topRows, err := config.DB.Query(queryTopProducts)
	report.TopRentedProducts = make([]models.TopRentedProduct, 0)
	if err == nil {
		defer topRows.Close()
		for topRows.Next() {
			var tp models.TopRentedProduct
			if errScan := topRows.Scan(&tp.ProductID, &tp.ProductName, &tp.CategoryName, &tp.RentCount, &tp.TotalRevenue); errScan == nil {
				report.TopRentedProducts = append(report.TopRentedProducts, tp)
			}
		}
	}

	// 4. Ambil 5 Transaksi Terkini
	queryRecent := `
		SELECT 
			r.id,
			COALESCE(u.username, '') AS username,
			COALESCE(p.product_name, '') AS product_name,
			COALESCE(i.asset_code, '') AS asset_code,
			DATE_FORMAT(r.start_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(r.return_date, '%Y-%m-%d') AS return_date,
			r.rental_fee,
			r.late_fee,
			r.rental_status
		FROM rentals r
		LEFT JOIN users u ON u.id_user = r.id_user
		LEFT JOIN item_instances i ON i.id = r.id_unit
		LEFT JOIN products p ON (p.id_product = i.id_products OR p.id = i.id_products)
		ORDER BY r.id DESC
		LIMIT 5
	`
	recentRows, err := config.DB.Query(queryRecent)
	report.RecentTransactions = make([]models.RecentTransactionReport, 0)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			var rt models.RecentTransactionReport
			if errScan := recentRows.Scan(&rt.ID, &rt.Username, &rt.ProductName, &rt.AssetCode, &rt.StartDate, &rt.ReturnDate, &rt.RentalFee, &rt.LateFee, &rt.RentalStatus); errScan == nil {
				report.RecentTransactions = append(report.RecentTransactions, rt)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Berhasil menghasilkan rekapitulasi data penjualan & rental",
		"data":    report,
	})
}

