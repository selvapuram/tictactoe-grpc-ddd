# Makefile
.PHONY: build run test test-unit test-integration proto clean docker help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build parameters
BINARY_NAME=tictactoe-server
BINARY_PATH=./cmd/server
DOCKER_IMAGE=tictactoe:latest

# Proto parameters
PROTO_DIR=./proto
PROTO_FILES=$(PROTO_DIR)/*.proto

help: ## Display this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

build: deps ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)

run: build ## Run the application
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	$(GOTEST) -v -race -coverprofile=coverage-unit.out ./internal/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -race -coverprofile=coverage-integration.out ./test/integration/...

test-coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage-unit.out -o coverage-unit.html
	$(GOCMD) tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Coverage reports generated: coverage-unit.html, coverage-integration.html"

benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage*.out coverage*.html

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Run in Docker
	docker run -p 8080:8080 $(DOCKER_IMAGE)

format: ## Format code
	$(GOCMD) fmt ./...
	$(GOCMD) mod tidy

install-tools: ## Install development tools
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

dev: ## Development setup
	make install-tools
	make proto
	make deps

# Default target
all: clean proto build test
