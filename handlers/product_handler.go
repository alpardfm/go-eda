package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alpardfm/go-eda/models"
	"github.com/alpardfm/go-eda/rabbitmq"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func GetProducts(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	db.Find(&products)
	json.NewEncoder(w).Encode(products)
}

func GetProduct(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var product models.Product
	db.First(&product, id)
	json.NewEncoder(w).Encode(product)
}

func CreateProduct(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var product models.Product
	_ = json.NewDecoder(r.Body).Decode(&product)
	db.Create(&product)
	json.NewEncoder(w).Encode(product)

	// Kirim event ke RabbitMQ
	message := "Product created: " + product.Name
	rabbitmq.PublishToQueue(message)
}

func UpdateProduct(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var product models.Product
	db.First(&product, id)
	_ = json.NewDecoder(r.Body).Decode(&product)
	db.Save(&product)
	json.NewEncoder(w).Encode(product)

	// Kirim event ke RabbitMQ
	message := "Product updated: " + product.Name
	rabbitmq.PublishToQueue(message)
}

func DeleteProduct(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var product models.Product
	db.Delete(&product, id)
	w.WriteHeader(http.StatusNoContent)

	// Kirim event ke RabbitMQ
	message := "Product deleted: " + strconv.Itoa(id)
	rabbitmq.PublishToQueue(message)
}
