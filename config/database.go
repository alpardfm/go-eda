package config

import (
	"log"

	"github.com/alpardfm/go-eda/models"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var DB *gorm.DB

// InitDB menginisialisasi koneksi ke database
func InitDB() {
	var err error
	// Perbaiki koneksi string. Gantilah 'pass' dengan 'password' dan 'ssmode' dengan 'sslmode'.
	connStr := "host=localhost port=5432 user=postgres password=f3rm0c4rd dbname=product_db sslmode=disable"

	// Pastikan DB diinisialisasi tanpa deklarasi baru.
	DB, err = gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Automigrate untuk memigrasi model ke database.
	DB.AutoMigrate(&models.Product{})
}
