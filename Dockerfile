# Build
FROM golang:1.25 AS builder
WORKDIR /app
COPY ../src/main.go .
RUN go mod init InferenceProfiler && go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o profiler main.go

# Setup
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
RUN mkdir -p /profiler-output
COPY --from=builder /app/profiler /usr/local/bin/profiler
ENTRYPOINT ["/usr/local/bin/profiler", "-o", "/profiler-output", "-t", "5000"]

# User
RUN apt-get update && \
    apt-get install -y python3-dev python3-venv pipx build-essential && \
    rm -rf /var/lib/apt/lists/*
RUN pipx install vllm huggingface_hub
RUN pipx inject vllm torch-c-dlpack-ext

ENV PATH="/root/.local/bin:$PATH"
ENV HF_HOME=/app/models
ARG MODEL="meta-llama/Llama-3.2-1B-Instruct"

RUN hf download $MODEL --local-dir /app/model/ --token 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

CMD ["vllm", "serve", "/app/model/", "--gpu-memory-utilization=0.7", "--max-model-len=2048", "--dtype=bfloat16"]