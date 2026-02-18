.PHONY: build dev run test test-integration up down migrate-up migrate-down migrate-create lint clean

# Variables
APP_NAME=aiox-api
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations
DB_URL=postgres://aiox:aiox_secret@localhost:5432/aiox?sslmode=disable

# Build
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

# Run in development
dev:
	go run ./cmd/api

# Docker
up:
	docker compose up -d

down:
	docker compose down

down-v:
	docker compose down -v

# Migrations
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down

migrate-down-1:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

# Tests
test:
	go test ./internal/... -v -race -count=1

test-integration:
	go test ./tests/... -v -race -count=1 -tags=integration

test-coverage:
	go test ./... -coverprofile=coverage.out -race -count=1
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
