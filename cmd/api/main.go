// API server — produces events on order state changes.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alpardfm/go-eda/internal/config"
	"github.com/alpardfm/go-eda/internal/handler"
	"github.com/alpardfm/go-eda/internal/order"
	"github.com/alpardfm/go-eda/internal/platform/broker"
	"github.com/alpardfm/go-eda/internal/platform/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	ctx := context.Background()

	// Database
	pool, err := database.Connect(ctx, cfg.Database.URL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Broker
	rmq, err := broker.NewRabbitMQ(cfg.Broker, logger)
	if err != nil {
		logger.Error("failed to connect to broker", "error", err)
		os.Exit(1)
	}
	defer rmq.Close()

	// Services
	orderRepo := database.NewOrderRepository(pool)
	orderService := order.NewService(orderRepo, rmq)

	// Handlers
	orderHandler := handler.NewOrderHandler(orderService)

	// Router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("POST /orders", orderHandler.Create)
	mux.HandleFunc("GET /orders", orderHandler.List)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetByID)
	mux.HandleFunc("POST /orders/{id}/pay", orderHandler.Pay)
	mux.HandleFunc("POST /orders/{id}/cancel", orderHandler.Cancel)

	// Server
	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("api server starting", "port", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
}
