# 1. Builder
FROM golang:1.25 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o profiler main.go

# 2. Runtime
FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
ARG MODEL_ID="meta-llama/Llama-3.2-1B-Instruct"
ARG MODEL_PATH=/app/model

RUN apt-get update && apt-get install -y --no-install-recommends build-essential curl python3-pip && \
    pip install --no-cache-dir --break-system-packages vllm torch-c-dlpack-ext

COPY --from=builder /app/profiler /usr/local/bin/profiler
RUN echo '#!/bin/sh \n\
vllm --model $MODEL_PATH --port 8000 --gpu-memory-utilization 0.7 & \n\
timeout 60s sh -c "until curl -s localhost:8000/health; do sleep 1; done" \n\
exec "$@"' > /entrypoint.sh && chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["profiler", "--no-procs", "--no-gpu-procs", "-o", "/profiler-output", "--delta", "--", \
     "vllm", "bench", "serve", \
     "--backend", "vllm", \
     "--model", "/app/model", \
     "--dataset-name", "random", \
     "--num-prompts", "100", \
     "--request-rate", "inf", \
     "--endpoint", "http://localhost:8000/v1/completions"]
