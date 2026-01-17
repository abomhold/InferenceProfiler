PROJECT_NAME  := infprofiler
SRC_DIR       := .
BIN_DIR       := bin
OUTPUT_DIR    := ./output
BINARY_NAME   := infpro
GO_BINARY     := $(BIN_DIR)/$(BINARY_NAME)
GO_MAIN       := $(SRC_DIR)/main.go
DOCKER_IMAGE  := $(PROJECT_NAME)
DOCKER_TAG    := latest
DOCKER_FILE   := Dockerfile

# Model Settings
MODEL_ID      ?= meta-llama/Llama-3.2-1B-Instruct
MODEL_DIR     ?= ./model


.DELETE_ON_ERROR:
.PHONY: all help build clean run snapshot serve test test-v test-cover \
        bench bench-collecting bench-utils bench-manager \
        docker-build docker-run docker-clean refresh get-model test-vllm

all: build

help: ##@ Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?##@ "} /^[a-zA-Z0-9_-]+:.*?##@ / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

refresh: ##@ Tidy modules, format and vet source code
	@echo "--- Refreshing Source Code ---"
	go mod tidy
	go fmt ./...
	go vet ./...

build: ##@ Compile Go binary
	@echo "--- Building $(GO_BINARY) ---"
	@mkdir -p $(BIN_DIR)
	go build -o $(GO_BINARY) $(GO_MAIN)

install: build ##@ Install binary to GOPATH/bin
	go install $(BUILD_FLAGS)

clean: docker-clean ##@ Remove build artifacts
	@echo "--- Cleaning ---"
	rm -rf $(BIN_DIR) $(OUTPUT_DIR) coverage.out coverage.html *.parquet *.jsonl *.prof

run: build ##@ Run continuous profiler (Ctrl+C to stop)
	@mkdir -p $(OUTPUT_DIR)
	./$(GO_BINARY) profiler -dir $(OUTPUT_DIR) -format parquet

test: ##@ Run all unit tests
	@echo "--- Running Tests ---"
	go test ./...

test-v: ##@ Run tests with verbose output
	@echo "--- Running Tests (verbose) ---"
	go test -v ./...

bench: ##@ Run all benchmarks
	@echo "--- Running All Benchmarks ---"
	go test -bench=. -benchmem -run=^$$ ./...

bench,: ##@ Run all benchmarks
	@echo "--- Running All Benchmarks ---"
	go test -bench=. -benchmem -run=^$$ ./... | sed -E ':a;s/([0-9])([0-9]{3})($|[^0-9])/\1,\2\3/;ta'

bench-hf: build ##@ Compare sequential vs concurrent with hyperfine
	@echo "--- Hyperfine Comparison ---"
	hyperfine --warmup 3 \
		'./$(GO_BINARY) snapshot -dynamic -no-process' \
		'./$(GO_BINARY) snapshot -dynamic -no-process -concurrent'

docker-build: ##@ Build Docker image (profile mode)
	@echo "--- Building Docker Image ---"
	docker build --progress=plain \
		-f $(DOCKER_FILE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.

docker-run: docker-build ##@ Run container with GPU support (profile mode)
	@echo "--- Running Docker Container ---"
	@mkdir -p $(OUTPUT_DIR)
	@if [ -d "$(MODEL_DIR)" ]; then \
		echo "Mounting model from $(MODEL_DIR)..."; \
		docker run --rm \
			--gpus all \
			-p "8000:8000" \
			-v $(OUTPUT_DIR):/output \
			-v $(MODEL_DIR):/app/model \
			$(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		echo "No local model at $(MODEL_DIR). Running without mount..."; \
		docker run --rm \
			--gpus all \
			-p "8000:8000" \
			-v $(OUTPUT_DIR):/output \
			$(DOCKER_IMAGE):$(DOCKER_TAG); \
	fi

docker-clean: ##@ Remove Docker images
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	docker rmi $(DOCKER_IMAGE):serve || true
	docker rmi $(DOCKER_IMAGE):run || true
	docker rmi $(DOCKER_IMAGE):sysbench || true

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

test-profiler: ##@ Test profiler HTTP server endpoints
	@echo "--- Testing Profiler Server ---"
	@echo "Static:"
	@curl -s http://localhost:8081/static | head -20
	@echo "\nDynamic:"
	@curl -s http://localhost:8081/dynamic | head -20
