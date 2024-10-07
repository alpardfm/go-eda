package main

import (
	"log"
	"net/http"

	"github.com/alpardfm/go-eda/config"
	"github.com/alpardfm/go-eda/handlers"
	"github.com/alpardfm/go-eda/rabbitmq"
	"github.com/gorilla/mux"
)

func main() {
	// Inisialisasi database
	config.InitDB()
	defer config.DB.Close()

	// Inisialisasi consumer RabbitMQ
	go rabbitmq.ConsumeFromQueue()

	// Inisialisasi router
	r := mux.NewRouter()

	// Routes untuk CRUD produk
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetProducts(config.DB, w, r)
	}).Methods("GET")
	r.HandleFunc("/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetProduct(config.DB, w, r)
	}).Methods("GET")
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateProduct(config.DB, w, r)
	}).Methods("POST")
	r.HandleFunc("/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateProduct(config.DB, w, r)
	}).Methods("PUT")
	r.HandleFunc("/products/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteProduct(config.DB, w, r)
	}).Methods("DELETE")

	// Menjalankan server
	log.Fatal(http.ListenAndServe(":8080", r))
}
