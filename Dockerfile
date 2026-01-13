# Build
FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o profiler src/main.go

# Profile
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    python3-dev \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --break-system-packages vllm torch-c-dlpack-ext


COPY --from=builder /app/profiler /usr/local/bin/profiler
RUN mkdir -p /profiler-output
ENTRYPOINT ["/usr/local/bin/profiler","--no-procs", "--no-gpu-procs", "-o", "/profiler-output", "-t", "100", "-f", "parquet", "--stream", "--"]

# Workload: 'vllm' is now in the global PATH
CMD ["vllm", "serve", "/app/model/", "--gpu-memory-utilization=0.7", "--max-model-len=2048", "--dtype=bfloat16"]
