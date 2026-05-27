// Package order contains the order domain model and events.
package order

import "time"

// Status represents the order lifecycle state.
type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusCancelled Status = "cancelled"
)

// Order represents a customer order.
type Order struct {
	ID        string    `json:"id"`
	Customer  string    `json:"customer"`
	Amount    float64   `json:"amount"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event types for the order domain.
const (
	EventOrderCreated   = "order.created"
	EventOrderPaid      = "order.paid"
	EventOrderCancelled = "order.cancelled"
)

// CreatedPayload is the event payload when an order is created.
type CreatedPayload struct {
	OrderID  string  `json:"order_id"`
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
}

// PaidPayload is the event payload when an order is paid.
type PaidPayload struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

// CancelledPayload is the event payload when an order is cancelled.
type CancelledPayload struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}
