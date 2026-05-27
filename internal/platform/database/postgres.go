// Package database provides PostgreSQL connection and repository implementations.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/alpardfm/go-eda/internal/order"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a PostgreSQL connection pool.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

// OrderRepository implements order.Repository using PostgreSQL.
type OrderRepository struct {
	pool *pgxpool.Pool
}

// NewOrderRepository creates a PostgreSQL-backed order repository.
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO orders (id, customer, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, o.ID, o.Customer, o.Amount, o.Status, o.CreatedAt, o.UpdatedAt)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	var o order.Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer, amount, status, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(&o.ID, &o.Customer, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, order.ErrNotFound
	}
	return &o, nil
}

func (r *OrderRepository) List(ctx context.Context) ([]order.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer, amount, status, created_at, updated_at
		FROM orders ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []order.Order
	for rows.Next() {
		var o order.Order
		if err := rows.Scan(&o.ID, &o.Customer, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status order.Status, updatedAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3
	`, status, updatedAt, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return order.ErrNotFound
	}
	return nil
}
