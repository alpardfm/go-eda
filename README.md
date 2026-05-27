# Go EDA

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.13-FF6600?style=flat-square&logo=rabbitmq&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white)

Event-Driven Architecture showcase in Go — demonstrating publish/subscribe, dead letter queues, retry with exponential backoff, and separate producer/consumer services.

---

## Architecture

```
┌──────────────┐         ┌─────────────┐         ┌──────────────┐
│   HTTP API   │─publish─▶│  RabbitMQ   │─consume─▶│   Consumer   │
│  (Producer)  │         │  (Broker)   │         │   (Worker)   │
└──────────────┘         └──────┬──────┘         └──────────────┘
       │                        │ on failure
       │                        ▼
       │                 ┌─────────────┐
       │                 │  Dead Letter │
       │                 │    Queue     │
       │                 └─────────────┘
       ▼
┌──────────────┐
│  PostgreSQL  │
└──────────────┘
```

### Event Flow

1. API receives HTTP request (e.g., create order)
2. Order is persisted to PostgreSQL
3. Domain event is published to RabbitMQ (e.g., `order.created`)
4. Consumer picks up the event and processes it
5. On failure: retry with exponential backoff (max 3 retries)
6. After max retries: event moves to Dead Letter Queue for manual inspection

### Structured Event Format

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "order.created",
  "timestamp": "2026-05-27T10:00:00Z",
  "payload": {
    "order_id": "...",
    "customer": "John Doe",
    "amount": 99.99
  },
  "metadata": {
    "source": "api",
    "retry_count": 0
  }
}
```

---

## Features

- **Separate binaries** — API (producer) and Consumer (worker) run independently
- **Broker-agnostic interfaces** — `event.Publisher` and `event.Consumer` can be swapped (RabbitMQ → Kafka → NATS)
- **Dead Letter Queue** — failed events preserved for debugging
- **Retry with exponential backoff** — 1s → 2s → 4s before DLQ
- **Structured events** — typed payloads with metadata (ID, timestamp, retry count)
- **Graceful shutdown** — both services handle SIGINT/SIGTERM cleanly
- **Topic-based routing** — events routed by type (`order.created`, `order.paid`, `order.cancelled`)
- **PostgreSQL persistence** — orders stored with proper indexes

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.24 |
| HTTP | stdlib `net/http` (Go 1.22+ routing) |
| Broker | RabbitMQ via `rabbitmq/amqp091-go` |
| Database | PostgreSQL via `pgx/v5` |
| Container | Docker Compose |

---

## Project Structure

```
go-eda/
├── cmd/
│   ├── api/main.go           # HTTP API server (producer)
│   └── consumer/main.go      # Event consumer (worker)
├── internal/
│   ├── config/               # Environment-based configuration
│   ├── event/                # Core event types + broker interfaces
│   ├── handler/              # HTTP handlers
│   ├── order/                # Order domain (entity, service, repository interface)
│   └── platform/
│       ├── broker/           # RabbitMQ implementation
│       └── database/         # PostgreSQL implementation
├── migrations/               # SQL schema
├── docker-compose.yml        # PostgreSQL + RabbitMQ
├── Makefile
└── .env.example
```

---

## API Endpoints

| Method | Path | Description | Event Published |
|--------|------|-------------|-----------------|
| `POST` | `/orders` | Create order | `order.created` |
| `GET` | `/orders` | List all orders | — |
| `GET` | `/orders/{id}` | Get order by ID | — |
| `POST` | `/orders/{id}/pay` | Mark order as paid | `order.paid` |
| `POST` | `/orders/{id}/cancel` | Cancel order | `order.cancelled` |
| `GET` | `/health` | Health check | — |

---

## Quick Start

### Prerequisites
- Go 1.24+
- Docker & Docker Compose

### 1. Start Infrastructure
```bash
make infra-up
```
This starts PostgreSQL (port 5432) and RabbitMQ (port 5672, management UI at 15672).

### 2. Run API Server
```bash
make run-api
```
API available at `http://localhost:8080`.

### 3. Run Consumer (separate terminal)
```bash
make run-consumer
```

### 4. Test the Flow
```bash
# Create an order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer": "John Doe", "amount": 99.99}'

# Pay the order
curl -X POST http://localhost:8080/orders/{id}/pay

# Check consumer logs — you'll see the events being processed
```

### 5. RabbitMQ Management
Open `http://localhost:15672` (guest/guest) to inspect queues, exchanges, and DLQ.

---

## Design Decisions

### Why separate binaries?
Producer and consumer scale independently. In production, you might have 1 API instance but 5 consumer instances for high-throughput event processing.

### Why interface-based broker?
```go
type Publisher interface {
    Publish(ctx context.Context, event Event) error
    Close() error
}
```
Swap RabbitMQ for Kafka or NATS without changing business logic. Also enables in-memory implementation for testing.

### Why exponential backoff?
Transient failures (network blip, DB timeout) often resolve themselves. Immediate retry floods the system. Backoff gives time to recover: 1s → 2s → 4s → DLQ.

### Why DLQ instead of infinite retry?
Poison messages (malformed data, business logic errors) will never succeed. DLQ preserves them for manual inspection without blocking the queue.

---

## Available Commands

```bash
make run-api        # Start API server
make run-consumer   # Start event consumer
make infra-up       # Start PostgreSQL + RabbitMQ
make infra-down     # Stop infrastructure
make test           # Run tests
make lint           # Run golangci-lint
make build          # Build both binaries
```

---

## License

MIT
