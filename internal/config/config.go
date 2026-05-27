// Package config provides environment-based configuration.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Broker   BrokerConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Env  string
	Port string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL string
}

// BrokerConfig holds message broker settings.
type BrokerConfig struct {
	URL             string
	Exchange        string
	Queue           string
	DeadLetterQueue string
	MaxRetries      int
	RetryBaseDelay  time.Duration
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/go_eda?sslmode=disable"),
		},
		Broker: BrokerConfig{
			URL:             getEnv("BROKER_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:        getEnv("BROKER_EXCHANGE", "events"),
			Queue:           getEnv("BROKER_QUEUE", "order_events"),
			DeadLetterQueue: getEnv("BROKER_DLQ", "order_events_dlq"),
			MaxRetries:      getEnvInt("BROKER_MAX_RETRIES", 3),
			RetryBaseDelay:  time.Duration(getEnvInt("BROKER_RETRY_BASE_DELAY_MS", 1000)) * time.Millisecond,
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
