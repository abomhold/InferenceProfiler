# Makefile for inference-profiler (Go)

PROJECT_NAME := inference-profiler
BINARY       := profiler
OUTPUT_DIR   := ./output
TESTDATA_DIR := ./testdata

.PHONY: all build test clean run testdata help

all: build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the profiler binary
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) .

test: testdata ## Run tests
	go test -v ./...

testdata: ## Generate test data files
	@./scripts/generate_testdata.sh $(TESTDATA_DIR)

run: build ## Run profiler locally
	@mkdir -p $(OUTPUT_DIR)
	./$(BINARY) -o $(OUTPUT_DIR) -t 1000

clean: ## Remove build artifacts
	rm -rf $(BINARY) $(OUTPUT_DIR) $(TESTDATA_DIR)
	go clean

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run

deps: ## Download dependencies
	go mod download
	go mod tidy
