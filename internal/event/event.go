// Package event defines the core event types and broker interfaces.
// The interfaces are broker-agnostic — implementations can use RabbitMQ, Kafka, NATS, etc.
package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event with metadata.
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
	Metadata  Metadata  `json:"metadata"`
}

// Metadata contains event processing metadata.
type Metadata struct {
	Source     string `json:"source"`
	RetryCount int    `json:"retry_count"`
}

// New creates a new event with a generated ID and current timestamp.
func New(eventType string, payload any) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
		Metadata: Metadata{
			Source: "api",
		},
	}
}

// Publisher publishes events to a message broker.
// Implementations: RabbitMQ, Kafka, NATS, in-memory (for testing).
type Publisher interface {
	Publish(ctx context.Context, event Event) error
	Close() error
}

// Consumer consumes events from a message broker.
// Implementations: RabbitMQ, Kafka, NATS.
type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
	Close() error
}

// Handler processes a single event. Return error to trigger retry/DLQ.
type Handler func(ctx context.Context, event Event) error
