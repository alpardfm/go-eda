// Consumer worker — processes events from the message broker.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alpardfm/go-eda/internal/config"
	"github.com/alpardfm/go-eda/internal/event"
	"github.com/alpardfm/go-eda/internal/platform/broker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	// Broker
	rmq, err := broker.NewRabbitMQ(cfg.Broker, logger)
	if err != nil {
		logger.Error("failed to connect to broker", "error", err)
		os.Exit(1)
	}
	defer rmq.Close()

	// Event handler — processes each event
	handler := func(ctx context.Context, evt event.Event) error {
		switch evt.Type {
		case "order.created":
			logger.Info("📦 order created",
				"event_id", evt.ID,
				"payload", fmt.Sprintf("%v", evt.Payload),
			)
			// In production: send confirmation email, update analytics, etc.

		case "order.paid":
			logger.Info("💰 order paid",
				"event_id", evt.ID,
				"payload", fmt.Sprintf("%v", evt.Payload),
			)
			// In production: trigger fulfillment, update inventory, etc.

		case "order.cancelled":
			logger.Info("❌ order cancelled",
				"event_id", evt.ID,
				"payload", fmt.Sprintf("%v", evt.Payload),
			)
			// In production: refund payment, restore inventory, etc.

		default:
			logger.Warn("unknown event type", "type", evt.Type)
		}

		return nil
	}

	// Start consuming with graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("received shutdown signal")
		cancel()
	}()

	logger.Info("consumer starting", "queue", cfg.Broker.Queue)
	if err := rmq.Consume(ctx, handler); err != nil && err != context.Canceled {
		logger.Error("consumer error", "error", err)
		os.Exit(1)
	}

	logger.Info("consumer stopped")
}
