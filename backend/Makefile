.PHONY: dev up down migrate sqlc build api ws tidy test

# Start all infrastructure services
up:
	docker compose up -d

# Stop all infrastructure services
down:
	docker compose down

# Run DB migrations (requires golang-migrate CLI)
migrate:
	migrate -path db/migrations -database "postgres://orbit:orbit@localhost:5432/orbit?sslmode=disable" up

migrate-down:
	migrate -path db/migrations -database "postgres://orbit:orbit@localhost:5432/orbit?sslmode=disable" down

# Generate sqlc code
sqlc:
	sqlc generate

# Build both binaries
build:
	go build -o bin/api ./cmd/api
	go build -o bin/ws ./cmd/ws

# Run API server (reads .env)
api:
	go run ./cmd/api

# Run WebSocket service
ws:
	go run ./cmd/ws

# Run all unit tests (no infrastructure needed)
test:
	go test ./pkg/... ./internal/...

# Run tests with verbose output
test-v:
	go test -v ./pkg/... ./internal/...

# Tidy dependencies
tidy:
	go mod tidy

# Full local dev start (infra + services)
dev:
	$(MAKE) up
	sleep 3
	$(MAKE) migrate
	$(MAKE) -j2 api ws
