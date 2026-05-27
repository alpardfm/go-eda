package order

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/alpardfm/go-eda/internal/event"
	"github.com/google/uuid"
)

var (
	ErrValidation        = errors.New("validation error")
	ErrNotFound          = errors.New("order not found")
	ErrInvalidTransition = errors.New("invalid status transition")
)

// Repository defines the persistence contract for orders.
type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	List(ctx context.Context) ([]Order, error)
	UpdateStatus(ctx context.Context, id string, status Status, updatedAt time.Time) error
}

// Service contains order use cases.
type Service struct {
	repo      Repository
	publisher event.Publisher
}

// NewService creates an order service.
func NewService(repo Repository, publisher event.Publisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

// CreateInput contains the payload for creating an order.
type CreateInput struct {
	Customer string  `json:"customer"`
	Amount   float64 `json:"amount"`
}

// Create creates a new order and publishes an order.created event.
func (s *Service) Create(ctx context.Context, input CreateInput) (*Order, error) {
	customer := strings.TrimSpace(input.Customer)
	if customer == "" || input.Amount <= 0 {
		return nil, ErrValidation
	}

	now := time.Now().UTC()
	order := &Order{
		ID:        uuid.New().String(),
		Customer:  customer,
		Amount:    input.Amount,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	evt := event.New(EventOrderCreated, CreatedPayload{
		OrderID:  order.ID,
		Customer: order.Customer,
		Amount:   order.Amount,
	})
	if err := s.publisher.Publish(ctx, evt); err != nil {
		// Log but don't fail — order is already persisted.
		// In production, use outbox pattern for guaranteed delivery.
		_ = err
	}

	return order, nil
}

// Pay marks an order as paid and publishes an order.paid event.
func (s *Service) Pay(ctx context.Context, id string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != StatusPending {
		return ErrInvalidTransition
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateStatus(ctx, id, StatusPaid, now); err != nil {
		return err
	}

	evt := event.New(EventOrderPaid, PaidPayload{
		OrderID: order.ID,
		Amount:  order.Amount,
	})
	_ = s.publisher.Publish(ctx, evt)

	return nil
}

// Cancel marks an order as cancelled and publishes an order.cancelled event.
func (s *Service) Cancel(ctx context.Context, id, reason string) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != StatusPending {
		return ErrInvalidTransition
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateStatus(ctx, id, StatusCancelled, now); err != nil {
		return err
	}

	evt := event.New(EventOrderCancelled, CancelledPayload{
		OrderID: order.ID,
		Reason:  reason,
	})
	_ = s.publisher.Publish(ctx, evt)

	return nil
}

// GetByID returns an order by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns all orders.
func (s *Service) List(ctx context.Context) ([]Order, error) {
	return s.repo.List(ctx)
}
