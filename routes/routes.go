package routes

import (
	"belajar_go/controllers"
	"belajar_go/middlewares"
	"net/http"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Serve folder uploads beserta seluruh sub-foldernya (termasuk products)
	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("API Belajar Go Berhasil Jalan!"))
	})

	mux.HandleFunc("/register", controllers.Register)
	mux.HandleFunc("/login", controllers.Login)
	mux.HandleFunc("/users", middlewares.AuthMiddleware(controllers.GetUsers))
	mux.HandleFunc("/users/", middlewares.AuthMiddleware(controllers.UserHandler))
	mux.HandleFunc("/api/users", middlewares.AuthMiddleware(controllers.GetUsers))
	mux.HandleFunc("/products/", middlewares.AuthMiddleware(controllers.ProductHandler))
	mux.HandleFunc("/api/category/", middlewares.AuthMiddleware(controllers.CategoryHandler))
	mux.HandleFunc("/category", middlewares.AuthMiddleware(controllers.CreateCategory))
	mux.HandleFunc("/rental", middlewares.AuthMiddleware(controllers.RentalHandler))
	mux.HandleFunc("/rental/", middlewares.AuthMiddleware(controllers.RentalHandler))
	mux.HandleFunc("/payments", middlewares.AuthMiddleware(controllers.PaymentHandler))
	mux.HandleFunc("/payments/", middlewares.AuthMiddleware(controllers.PaymentHandler))
	mux.HandleFunc("/item-instances", middlewares.AuthMiddleware(controllers.ItemInstanceHandler))
	mux.HandleFunc("/item-instances/", middlewares.AuthMiddleware(controllers.ItemInstanceHandler))

	return mux
}
