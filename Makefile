PROJECT_NAME  := inference-profiler
DOCKER_IMAGE  := $(PROJECT_NAME)
DOCKER_TAG    := latest
OUTPUT_DIR    := ./output
DIST_DIR      := dist
MODEL         := meta-llama/Llama-3.2-1B-Instruct
MODEL_PATH    := ./local_model

.DELETE_ON_ERROR:
.PHONY: all help dev-refresh format lint build run docker-build docker-run test-vllm clean

all: help

help: ##@ Shows this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?##@ "} /^[a-zA-Z0-9_-]+:.*?##@ / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ##@ Build Python wheel
	@echo "--- Building Wheel ---"
	@uv sync
	@uv build
	@echo "Build artifacts created in $(DIST_DIR)"

run: ##@ Run profiler locally using uv
	@echo "--- Running Profiler Locally ---"
	@mkdir -p $(OUTPUT_DIR)
	@uv sync
	@uv run inference-profiler -o $(OUTPUT_DIR) -t 1000

docker-build: build ##@ Build Docker image
	@echo "--- Building Docker Image ---"
	@docker build --progress=plain \
			   -t $(DOCKER_IMAGE):$(DOCKER_TAG) \
			   .

docker-run: ##@ Run container with GPU support and volume mount
	@echo "--- Running Docker Container ---"
	@mkdir -p $(OUTPUT_DIR)
	@docker run --rm \
	   -p "8000:8000" \
	   -v $(shell pwd)/$(OUTPUT_DIR):/profiler-output \
	   -v $(MODEL_PATH):/app/model \
	   --gpus all \
	   $(DOCKER_IMAGE):$(DOCKER_TAG)

test-vllm: ##@ Test the vllm server (requires container running)
	@echo "--- Testing vllm Server ---"
	@curl http://localhost:8000/v1/chat/completions \
		 -H "Content-Type: application/json" \
		 -d '{ "messages": [{"role": "user", "content": "Explain the difference between TCP and UDP."}], "max_tokens": 1024}'

clean: ##@ Remove build artifacts, cache, and output
	@echo "--- Cleaning Artifacts ---"
	@rm -rf $(OUTPUT_DIR) $(DIST_DIR) .venv
	@find . -type d -name "__pycache__" -exec rm -rf {} +
	@find . -type d -name "*.egg-info" -exec rm -rf {} +
	@rm -rf inference_profiler.egg-info
	@echo "--- Removing Docker Image ---"
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true