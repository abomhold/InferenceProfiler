PROJECT_NAME  := profiler
SRC_DIR       := src
BIN_DIR       := bin
OUTPUT_DIR    := ./output
BINARY_NAME   := profiler
GO_BINARY     := $(BIN_DIR)/$(BINARY_NAME)
GO_MAIN       := $(SRC_DIR)/main.go
DOCKER_IMAGE  := $(PROJECT_NAME)
DOCKER_TAG    := latest
DOCKER_FILE   := Dockerfile

.DELETE_ON_ERROR:
.PHONY: all help build clean run docker-build docker-run docker-clean refresh test test-v test-cover bench bench-parse bench-collect
all: help

help: ##@ Shows this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?##@ "} /^[a-zA-Z0-9_-]+:.*?##@ / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

refresh: ##@ Install/manage modules, format source code, inspect code
	@echo "--- Refreshing Source Code ---"
	go mod tidy
	go fmt ./...
	go vet ./...

build: refresh ##@ Compile Go binary locally
	@echo "--- Building Binary at $(GO_BINARY) ---"
	mkdir -p $(BIN_DIR)
	go build -o $(GO_BINARY) $(GO_MAIN)
	chmod +x $(GO_BINARY)

run: build ##@ Build and run the profiler locally
	@echo "--- Running Profiler ---"
	mkdir -p $(OUTPUT_DIR)
	./$(GO_BINARY) -o $(OUTPUT_DIR)

test: ##@ Run all unit tests
	@echo "--- Running Tests ---"
	go test ./...

test-v: ##@ Run all unit tests with verbose output
	@echo "--- Running Tests (verbose) ---"
	go test -v ./...

test-cover: ##@ Run tests with coverage report
	@echo "--- Running Tests with Coverage ---"
	go test -cover ./...

bench: ##@ Run all benchmarks
	@echo "--- Running All Benchmarks ---"
	go test -bench=. -benchmem ./...

bench-collect: ##@ Run collection benchmarks (requires /proc)
	@echo "--- Running Collection Benchmarks ---"
	go test -bench=Collect -benchmem ./$(SRC_DIR)/collectors/

bench-output: ##@ Run flatten/aggregate benchmarks
	@echo "--- Running Flatten Benchmarks ---"
	go test -bench=. -benchmem ./$(SRC_DIR)/aggregate/

docker-build: ##@ Build Docker image
	@echo "--- Building Docker Image ($(DOCKER_FILE)) ---"
	docker build --progress=plain \
		-f $(DOCKER_FILE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.

docker-run: docker-build ##@ Run container
	@echo "--- Running Docker Container ---"
	@mkdir -p $(OUTPUT_DIR)
	docker run --rm \
		-p "8000:8000" \
		-v $(OUTPUT_DIR):/profiler-output \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

test-vllm: ##@ Send test request to local vllm server
	@echo "--- Testing vllm Server ---"
	curl http://localhost:8000/v1/chat/completions \
		-H "Content-Type: application/json" \
		-d '{ "messages": [{"role": "user", "content": "Explain the difference between TCP and UDP."}], "max_tokens": 100}'

clean: ##@ Remove all artifacts and docker images
	@echo "--- Cleaning Artifacts ---"
	rm -rf $(BIN_DIR) $(OUTPUT_DIR) coverage.out coverage.html
	@echo "--- Removing Docker Image ---"
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true