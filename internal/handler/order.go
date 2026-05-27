// Package handler contains HTTP handlers for the API.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alpardfm/go-eda/internal/order"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	service *order.Service
}

// NewOrderHandler creates an order handler.
func NewOrderHandler(service *order.Service) *OrderHandler {
	return &OrderHandler{service: service}
}

// Create handles POST /orders.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input order.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.service.Create(r.Context(), input)
	if err != nil {
		if err == order.ErrValidation {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    created,
	})
}

// List handles GET /orders.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list orders")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    orders,
	})
}

// GetByID handles GET /orders/{id}.
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "order id is required")
		return
	}

	o, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == order.ErrNotFound {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get order")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    o,
	})
}

// Pay handles POST /orders/{id}/pay.
func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "order id is required")
		return
	}

	err := h.service.Pay(r.Context(), id)
	if err != nil {
		switch err {
		case order.ErrNotFound:
			writeError(w, http.StatusNotFound, "order not found")
		case order.ErrInvalidTransition:
			writeError(w, http.StatusConflict, "order cannot be paid in current state")
		default:
			writeError(w, http.StatusInternalServerError, "failed to pay order")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "order paid",
	})
}

// Cancel handles POST /orders/{id}/cancel.
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "order id is required")
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	err := h.service.Cancel(r.Context(), id, body.Reason)
	if err != nil {
		switch err {
		case order.ErrNotFound:
			writeError(w, http.StatusNotFound, "order not found")
		case order.ErrInvalidTransition:
			writeError(w, http.StatusConflict, "order cannot be cancelled in current state")
		default:
			writeError(w, http.StatusInternalServerError, "failed to cancel order")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "order cancelled",
	})
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"message": message,
	})
}
