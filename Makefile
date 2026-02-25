.PHONY: build dev run test test-integration up down migrate-up migrate-down migrate-create lint clean proto docker-build vet fmt fmt-check security check frontend-install frontend-dev frontend-build

# Variables
APP_NAME=aiox-api
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations
DB_URL=postgres://aiox:aiox_secret@localhost:5433/aiox?sslmode=disable

# Build
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api

# Run in development
dev:
	go run ./cmd/api

# Docker
up:
	docker compose up -d --build

down:
	docker compose down

down-v:
	docker compose down -v

docker-build:
	docker build -t $(APP_NAME):latest .

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

# Code quality
vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:" && gofmt -l . && exit 1)

# Lint
lint:
	golangci-lint run ./...

# Security scan (requires govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest)
security:
	govulncheck ./...

# Run all checks: format, vet, test
check: fmt-check vet test

# Proto (generate Go + Python code from .proto files)
proto:
	protoc --proto_path=proto/worker/v1 \
		--go_out=internal/worker/workerpb --go_opt=paths=source_relative \
		--go-grpc_out=internal/worker/workerpb --go-grpc_opt=paths=source_relative \
		worker.proto
	worker/.venv/bin/python -m grpc_tools.protoc \
		-I proto/worker/v1 \
		--python_out=worker/worker \
		--grpc_python_out=worker/worker \
		proto/worker/v1/worker.proto
	sed -i 's/^import worker_pb2/from . import worker_pb2/' worker/worker/worker_pb2_grpc.py

# Frontend
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# Clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
