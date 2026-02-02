.PHONY: help build run clean test deps setup test-coverage test-verbose

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $1, $2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	go mod download
	go mod tidy

setup: ## Setup project directories and environment
	mkdir -p storage/media/songs storage/media/uploads storage/images/profiles storage/images/covers
	mkdir -p static/css static/js static/images
	cp -n .env.example .env || true
	@echo "Setup complete! Edit .env file with your configuration."

build: ## Build the application
	go build -o bin/tunetudo main.go

run: ## Run the application
	go run main.go

dev: ## Run with hot reload (requires air)
	air

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f tunetudo.db
	rm -f test_*.db
	rm -rf test_storage_*

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-verbose: ## Run tests in verbose mode
	go test -v -race ./...

test-unit: ## Run only unit tests
	go test -v ./services/...

test-integration: ## Run only integration tests
	go test -v -run Integration ./...

test-short: ## Run short tests only
	go test -short ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

docker-build: ## Build Docker image
	docker build -t tunetudo:latest .

docker-run: ## Run Docker container
	docker run -p 2701:2701 -v $(PWD)/storage:/app/storage tunetudo:latest

ci: deps test ## Run CI pipeline (install deps and test)
	@echo "CI pipeline completed successfully!"