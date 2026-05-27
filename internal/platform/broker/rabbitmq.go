// Package broker provides message broker implementations.
package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/alpardfm/go-eda/internal/config"
	"github.com/alpardfm/go-eda/internal/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ implements event.Publisher and event.Consumer using RabbitMQ.
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.BrokerConfig
	logger  *slog.Logger
}

// NewRabbitMQ creates a new RabbitMQ connection with exchange and queue setup.
func NewRabbitMQ(cfg config.BrokerConfig, logger *slog.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
		logger:  logger,
	}

	if err := rmq.setup(); err != nil {
		rmq.Close()
		return nil, err
	}

	return rmq, nil
}

// setup declares exchange, queues, and bindings including DLQ.
func (r *RabbitMQ) setup() error {
	// Declare main exchange (topic type for routing by event type)
	if err := r.channel.ExchangeDeclare(
		r.cfg.Exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	// Declare dead letter exchange
	dlxName := r.cfg.Exchange + ".dlx"
	if err := r.channel.ExchangeDeclare(
		dlxName,
		"topic",
		true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare DLX: %w", err)
	}

	// Declare main queue with DLQ routing
	_, err := r.channel.QueueDeclare(
		r.cfg.Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    dlxName,
			"x-dead-letter-routing-key": r.cfg.DeadLetterQueue,
		},
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	// Bind main queue to exchange (all order events)
	if err := r.channel.QueueBind(
		r.cfg.Queue,
		"order.*", // routing key pattern
		r.cfg.Exchange,
		false, nil,
	); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	// Declare dead letter queue
	_, err = r.channel.QueueDeclare(
		r.cfg.DeadLetterQueue,
		true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("declare DLQ: %w", err)
	}

	// Bind DLQ to DLX
	if err := r.channel.QueueBind(
		r.cfg.DeadLetterQueue,
		r.cfg.DeadLetterQueue,
		dlxName,
		false, nil,
	); err != nil {
		return fmt.Errorf("bind DLQ: %w", err)
	}

	return nil
}

// Publish publishes an event to the exchange with the event type as routing key.
func (r *RabbitMQ) Publish(ctx context.Context, evt event.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return r.channel.PublishWithContext(ctx,
		r.cfg.Exchange,
		evt.Type, // routing key = event type (e.g., "order.created")
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    evt.ID,
			Timestamp:    evt.Timestamp,
			Body:         body,
		},
	)
}

// Consume starts consuming events from the queue and processes them with the handler.
// Failed events are retried with exponential backoff up to MaxRetries, then sent to DLQ.
func (r *RabbitMQ) Consume(ctx context.Context, handler event.Handler) error {
	if err := r.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	msgs, err := r.channel.Consume(
		r.cfg.Queue,
		"",    // consumer tag (auto-generated)
		false, // auto-ack (manual for retry control)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	r.logger.Info("consumer started", "queue", r.cfg.Queue)

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("consumer stopping", "reason", ctx.Err())
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			r.processMessage(ctx, msg, handler)
		}
	}
}

func (r *RabbitMQ) processMessage(ctx context.Context, msg amqp.Delivery, handler event.Handler) {
	var evt event.Event
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		r.logger.Error("unmarshal event failed", "error", err)
		msg.Nack(false, false) // Send to DLQ (don't requeue)
		return
	}

	r.logger.Info("processing event",
		"event_id", evt.ID,
		"type", evt.Type,
		"retry", evt.Metadata.RetryCount,
	)

	if err := handler(ctx, evt); err != nil {
		r.handleFailure(ctx, msg, evt, err)
		return
	}

	msg.Ack(false)
	r.logger.Info("event processed", "event_id", evt.ID, "type", evt.Type)
}

func (r *RabbitMQ) handleFailure(ctx context.Context, msg amqp.Delivery, evt event.Event, err error) {
	evt.Metadata.RetryCount++

	if evt.Metadata.RetryCount > r.cfg.MaxRetries {
		r.logger.Error("max retries exceeded, sending to DLQ",
			"event_id", evt.ID,
			"type", evt.Type,
			"retries", evt.Metadata.RetryCount,
			"error", err,
		)
		msg.Nack(false, false) // Send to DLQ
		return
	}

	// Exponential backoff: baseDelay * 2^(retryCount-1)
	delay := r.cfg.RetryBaseDelay * time.Duration(math.Pow(2, float64(evt.Metadata.RetryCount-1)))
	r.logger.Warn("event processing failed, retrying",
		"event_id", evt.ID,
		"type", evt.Type,
		"retry", evt.Metadata.RetryCount,
		"delay", delay,
		"error", err,
	)

	time.Sleep(delay)

	// Re-publish with incremented retry count
	body, _ := json.Marshal(evt)
	r.channel.PublishWithContext(ctx,
		r.cfg.Exchange,
		evt.Type,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    evt.ID,
			Timestamp:    evt.Timestamp,
			Body:         body,
		},
	)

	msg.Ack(false) // Ack original (retry is a new message)
}

// Close closes the channel and connection.
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
