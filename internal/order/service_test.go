package order

import (
	"context"
	"testing"
	"time"

	"github.com/alpardfm/go-eda/internal/event"
)

// --- Fakes ---

type fakeRepo struct {
	orders map[string]*Order
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{orders: make(map[string]*Order)}
}

func (r *fakeRepo) Create(_ context.Context, order *Order) error {
	r.orders[order.ID] = order
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id string) (*Order, error) {
	order, ok := r.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	return order, nil
}

func (r *fakeRepo) List(_ context.Context) ([]Order, error) {
	result := make([]Order, 0, len(r.orders))
	for _, o := range r.orders {
		result = append(result, *o)
	}
	return result, nil
}

func (r *fakeRepo) UpdateStatus(_ context.Context, id string, status Status, updatedAt time.Time) error {
	order, ok := r.orders[id]
	if !ok {
		return ErrNotFound
	}
	order.Status = status
	order.UpdatedAt = updatedAt
	return nil
}

type fakePublisher struct {
	events []event.Event
}

func (p *fakePublisher) Publish(_ context.Context, evt event.Event) error {
	p.events = append(p.events, evt)
	return nil
}

func (p *fakePublisher) Close() error { return nil }

// --- Tests ---

func TestCreateOrderSuccess(t *testing.T) {
	repo := newFakeRepo()
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	order, err := svc.Create(context.Background(), CreateInput{
		Customer: "John Doe",
		Amount:   99.99,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ID == "" {
		t.Fatal("expected order ID to be generated")
	}
	if order.Status != StatusPending {
		t.Fatalf("expected pending status, got %q", order.Status)
	}
	if order.Customer != "John Doe" {
		t.Fatalf("expected customer John Doe, got %q", order.Customer)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.events))
	}
	if pub.events[0].Type != EventOrderCreated {
		t.Fatalf("expected order.created event, got %q", pub.events[0].Type)
	}
}

func TestCreateOrderRejectsEmptyCustomer(t *testing.T) {
	svc := NewService(newFakeRepo(), &fakePublisher{})

	_, err := svc.Create(context.Background(), CreateInput{
		Customer: "   ",
		Amount:   10.00,
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestCreateOrderRejectsZeroAmount(t *testing.T) {
	svc := NewService(newFakeRepo(), &fakePublisher{})

	_, err := svc.Create(context.Background(), CreateInput{
		Customer: "John",
		Amount:   0,
	})
	if err != ErrValidation {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestPayOrderSuccess(t *testing.T) {
	repo := newFakeRepo()
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	order, _ := svc.Create(context.Background(), CreateInput{
		Customer: "Jane",
		Amount:   50.00,
	})

	err := svc.Pay(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := svc.GetByID(context.Background(), order.ID)
	if updated.Status != StatusPaid {
		t.Fatalf("expected paid status, got %q", updated.Status)
	}
	if len(pub.events) != 2 {
		t.Fatalf("expected 2 events (created + paid), got %d", len(pub.events))
	}
	if pub.events[1].Type != EventOrderPaid {
		t.Fatalf("expected order.paid event, got %q", pub.events[1].Type)
	}
}

func TestPayOrderRejectsNonPending(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fakePublisher{})

	order, _ := svc.Create(context.Background(), CreateInput{Customer: "Jane", Amount: 50.00})
	svc.Pay(context.Background(), order.ID) // Now paid

	err := svc.Pay(context.Background(), order.ID) // Try pay again
	if err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestCancelOrderSuccess(t *testing.T) {
	repo := newFakeRepo()
	pub := &fakePublisher{}
	svc := NewService(repo, pub)

	order, _ := svc.Create(context.Background(), CreateInput{Customer: "Bob", Amount: 25.00})

	err := svc.Cancel(context.Background(), order.ID, "changed mind")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := svc.GetByID(context.Background(), order.ID)
	if updated.Status != StatusCancelled {
		t.Fatalf("expected cancelled status, got %q", updated.Status)
	}
	if pub.events[1].Type != EventOrderCancelled {
		t.Fatalf("expected order.cancelled event, got %q", pub.events[1].Type)
	}
}

func TestCancelOrderRejectsNonPending(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, &fakePublisher{})

	order, _ := svc.Create(context.Background(), CreateInput{Customer: "Bob", Amount: 25.00})
	svc.Pay(context.Background(), order.ID) // Now paid

	err := svc.Cancel(context.Background(), order.ID, "too late")
	if err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	svc := NewService(newFakeRepo(), &fakePublisher{})

	_, err := svc.GetByID(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
