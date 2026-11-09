package config

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	DB, err = sql.Open(
		"mysql",
		"root:keonhoal@tcp(127.0.0.1:3306)/rental_barang?parseTime=true",
	)

	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Database berhasil terhubung")
}
