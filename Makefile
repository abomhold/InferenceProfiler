PROJECT_NAME  := infprofiler
SRC_DIR       := .
BIN_DIR       := bin
OUTPUT_DIR    := ./output
BINARY_NAME   := infprofiler
GO_BINARY     := $(BIN_DIR)/$(BINARY_NAME)
GO_MAIN       := $(SRC_DIR)/main.go
DOCKER_IMAGE  := $(PROJECT_NAME)
DOCKER_TAG    := latest
DOCKER_FILE   := Dockerfile

# Model Settings
MODEL_ID      ?= meta-llama/Llama-3.2-1B-Instruct
MODEL_DIR     ?= ./model

# Build flags
LDFLAGS       := -s -w
BUILD_FLAGS   := -ldflags "$(LDFLAGS)"

.DELETE_ON_ERROR:
.PHONY: all help build clean run snapshot serve test test-v test-cover \
        bench bench-collectors bench-probing bench-vm bench-manager \
        docker-build docker-run docker-clean refresh get-model test-vllm

all: build

help: ##@ Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?##@ "} /^[a-zA-Z0-9_-]+:.*?##@ / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================================
# Development
# ============================================================================

refresh: ##@ Tidy modules, format and vet source code
	@echo "--- Refreshing Source Code ---"
	go mod tidy
	go fmt ./...
	go vet ./...

build: ##@ Compile Go binary
	@echo "--- Building $(GO_BINARY) ---"
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(GO_BINARY) $(GO_MAIN)

build-dev: ##@ Build without optimizations (faster compile)
	@mkdir -p $(BIN_DIR)
	go build -o $(GO_BINARY) $(GO_MAIN)

install: build ##@ Install binary to GOPATH/bin
	go install $(BUILD_FLAGS)

clean: ##@ Remove build artifacts
	@echo "--- Cleaning ---"
	rm -rf $(BIN_DIR) $(OUTPUT_DIR) coverage.out coverage.html *.parquet *.jsonl

# ============================================================================
# Run Commands
# ============================================================================

run: build ##@ Run continuous profiler (Ctrl+C to stop)
	@mkdir -p $(OUTPUT_DIR)
	./$(GO_BINARY) profiler -o $(OUTPUT_DIR) --format parquet

snapshot: build ##@ Capture single metrics snapshot
	./$(GO_BINARY) snapshot

snapshot-json: build ##@ Capture snapshot and save to JSON
	@mkdir -p $(OUTPUT_DIR)
	./$(GO_BINARY) snapshot -o $(OUTPUT_DIR)/snapshot.json

serve: build ##@ Run HTTP metrics server on :8080
	./$(GO_BINARY) serve --addr :8080

profile-cmd: build ##@ Profile a command (usage: make profile-cmd CMD="sleep 5")
	@mkdir -p $(OUTPUT_DIR)
	./$(GO_BINARY) profile -o $(OUTPUT_DIR) -- $(CMD)

# ============================================================================
# Testing
# ============================================================================

test: ##@ Run all unit tests
	@echo "--- Running Tests ---"
	go test ./...

test-v: ##@ Run tests with verbose output
	@echo "--- Running Tests (verbose) ---"
	go test -v ./...

test-cover: ##@ Run tests with coverage report
	@echo "--- Running Tests with Coverage ---"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race: ##@ Run tests with race detector
	go test -race ./...

# ============================================================================
# Benchmarks
# ============================================================================

bench: ##@ Run all benchmarks
	@echo "--- Running All Benchmarks ---"
	go test -bench=. -benchmem -run=^$$ ./...

bench-short: ##@ Run benchmarks (shorter, 1s each)
	go test -bench=. -benchmem -benchtime=1s -run=^$$ ./...

bench-collectors: ##@ Run collector benchmarks
	@echo "--- Collector Benchmarks ---"
	go test -bench=. -benchmem -run=^$$ ./pkg/collectors/

bench-probing: ##@ Run probing/file I/O benchmarks
	@echo "--- Probing Benchmarks ---"
	go test -bench=. -benchmem -run=^$$ ./pkg/probing/

bench-vm: ##@ Run VM collector benchmarks only
	go test -bench=BenchmarkIsolated -benchmem -run=^$$ ./pkg/collectors/

bench-manager: ##@ Run manager benchmarks
	go test -bench=BenchmarkManager -benchmem -run=^$$ ./pkg/collectors/

bench-cpu: ##@ Profile CPU during benchmark
	go test -bench=BenchmarkManager_CollectDynamic_All -benchmem -cpuprofile=cpu.prof -run=^$$ ./pkg/collectors/
	@echo "View with: go tool pprof cpu.prof"

bench-mem: ##@ Profile memory during benchmark
	go test -bench=BenchmarkManager_CollectDynamic_All -benchmem -memprofile=mem.prof -run=^$$ ./pkg/collectors/
	@echo "View with: go tool pprof mem.prof"

# ============================================================================
# Docker
# ============================================================================

docker-build: ##@ Build Docker image
	@echo "--- Building Docker Image ---"
	docker build --progress=plain \
		-f $(DOCKER_FILE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.

docker-run: docker-build ##@ Run container with GPU support
	@echo "--- Running Docker Container ---"
	@mkdir -p $(OUTPUT_DIR)
	@if [ -d "$(MODEL_DIR)" ]; then \
		echo "Mounting model from $(MODEL_DIR)..."; \
		docker run --rm \
			--gpus all \
			-p "8000:8000" \
			-v $(OUTPUT_DIR):/profiler-output \
			-v $(MODEL_DIR):/app/model \
			$(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		echo "No local model at $(MODEL_DIR). Running without mount..."; \
		docker run --rm \
			--gpus all \
			-p "8000:8000" \
			-v $(OUTPUT_DIR):/profiler-output \
			$(DOCKER_IMAGE):$(DOCKER_TAG); \
	fi

docker-clean: ##@ Remove Docker image
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

# ============================================================================
# Utilities
# ============================================================================

get-model: ##@ Download HuggingFace model
	@echo "--- Downloading Model: $(MODEL_ID) ---"
	@mkdir -p $(MODEL_DIR)
	huggingface-cli download $(MODEL_ID) --local-dir $(MODEL_DIR)
	@echo "--- Download Complete ---"

test-vllm: ##@ Send test requests to local vLLM server
	@echo "--- Testing vLLM Server ---"
	@while true; do \
		curl -sN http://localhost:8000/v1/chat/completions \
			-H "Content-Type: application/json" \
			-d '{"messages": [{"role": "user", "content": "Explain TCP vs UDP briefly."}], "max_tokens": 128, "stream": true}'; \
		echo; \
		sleep 0.5; \
	done

loc: ##@ Count lines of code
	@find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1

deps: ##@ Show module dependencies
	go mod graph | head -20