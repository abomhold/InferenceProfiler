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

bench,: ##@ Run all benchmarks with comma formatting
	@echo "--- Running All Benchmarks ---"
	go test -bench=. -benchmem -run=^$$ ./... | python3 -c "import re,sys;[print(re.sub(r'(\d)(?=(\d{3})+(?!\d))',r'\1,',l),end='')for l in sys.stdin]"

bench-hf: build ## Benchmark & Plot (Numbered, Tall, Safe Legend)
	$(eval RESULTS := $(shell mktemp))
	$(eval PLOT := $(shell mktemp))
	@echo "--- Running Benchmarks ---"
	@hyperfine --style full --export-csv $(RESULTS) --warmup 10 \
	   './$(GO_BINARY) ss' \
	   './$(GO_BINARY) ss --concurrent' \
	   './$(GO_BINARY) ss --static' \
	   './$(GO_BINARY) ss --static --concurrent' \
	   './$(GO_BINARY) ss --dynamic' \
	   './$(GO_BINARY) ss --dynamic --concurrent' \
	   './$(GO_BINARY) ss --dynamic --no-procs'	\
	   './$(GO_BINARY) ss --dynamic --no-procs --concurrent' \
	   './$(GO_BINARY) ss --dynamic --no-procs --no-gpu-procs' \
	   './$(GO_BINARY) ss --dynamic --no-procs --no-gpu-procs --concurrent' \
	   './$(GO_BINARY) ss --dynamic --no-procs --no-nvidia' \
	   './$(GO_BINARY) ss --dynamic --no-procs --no-nvidia --concurrent'
	@sed -i 's|\./$(GO_BINARY) ||g' $(RESULTS)
	@echo 'set terminal dumb size 120, 45 ansi256' > $(PLOT)
	@echo 'set datafile separator ","' >> $(PLOT)
	@echo 'set title "Benchmark Results"' >> $(PLOT)
	@echo 'set xlabel "Time (ms)"' >> $(PLOT)
	@echo 'set ylabel "ID"' >> $(PLOT)
	@echo 'set tmargin 2' >> $(PLOT)
	@echo 'set lmargin 6' >> $(PLOT)
	@echo 'set offsets 0, 0, 0.5, 0.5' >> $(PLOT)
	@echo 'set yrange [0:*]' >> $(PLOT)
	@echo 'set mxtics 3' >> $(PLOT)
	@echo 'unset key' >> $(PLOT)
	@echo 'plot "$(RESULTS)" every ::1 using ($$2*1000):0:($$7*1000):($$8*1000):ytic(sprintf("%.0f", $$0+1)) with xerrorbars lc rgb "cyan", \\' >> $(PLOT)
	@echo '     ""           every ::1 using ($$2*1000):0:("*") with labels tc rgb "green" offset 0,0' >> $(PLOT)
	@gnuplot $(PLOT)
	@echo ""
	@echo "Legend:"
	@sed '1d' $(RESULTS) | cut -d, -f1 | nl -w2 -s ' : '
	@rm $(RESULTS) $(PLOT)

docker-build: ##@ Build Docker image (profile mode)
	@echo "--- Building Docker Image ---"
	docker build --progress=plain \
		-f $(DOCKER_FILE) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.

docker-run: docker-build get-model ##@ Run container with GPU support (profile mode)
	@echo "--- Running Docker Container ---"
	@mkdir -p $(OUTPUT_DIR)
	docker run --gpus all \
		-v $(OUTPUT_DIR):/profiler-output \
		-v $(MODEL_DIR):/app/model \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean: ##@ Remove Docker images
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

get-model: ##@ Download HuggingFace model
	@echo "--- Downloading Model: $(MODEL_ID) ---"
	@mkdir -p $(MODEL_DIR)
	hf download $(MODEL_ID) --local-dir $(MODEL_DIR)
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
