.PHONY: run-api run-consumer infra-up infra-down test lint build

run-api:
	go run ./cmd/api

run-consumer:
	go run ./cmd/consumer

infra-up:
	docker compose up -d

infra-down:
	docker compose down

test:
	go test ./...

lint:
	golangci-lint run ./...

build:
	go build ./cmd/api
	go build ./cmd/consumer
