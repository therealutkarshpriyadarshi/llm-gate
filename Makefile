.PHONY: help build test run clean lint docker-build docker-up docker-down install-deps tidy

# Variables
BINARY_NAME=gateway
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-deps: ## Install Go dependencies
	go mod download
	go mod verify

tidy: ## Tidy Go modules
	go mod tidy

build: ## Build the gateway binary
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/gateway/main.go
	@echo "Build complete: bin/${BINARY_NAME}"

run: ## Run the gateway locally
	go run cmd/gateway/main.go

test: ## Run unit tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out

integration-test: ## Run integration tests
	go test -v -tags=integration ./tests/integration/...

e2e-test: ## Run end-to-end tests
	go test -v -tags=e2e ./tests/e2e/...

lint: ## Run linter
	golangci-lint run --timeout=5m

fmt: ## Format code
	gofmt -w -s .
	goimports -w .

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out

docker-build: ## Build Docker image
	docker build -t llmgate:${VERSION} -f deployments/docker/Dockerfile .

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f

docker-clean: ## Clean Docker resources
	docker-compose down -v
	docker system prune -f

check: lint test ## Run linter and tests

all: clean install-deps lint test build ## Run all steps

.DEFAULT_GOAL := help
