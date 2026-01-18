FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o infpro main.go


FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
ARG MODEL_ID="meta-llama/Llama-3.2-1B-Instruct"
ARG MODEL_PATH=/app/model
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    python3-pip \
    python3-dev && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir --break-system-packages vllm torch-c-dlpack-ext
COPY --from=builder /app/infpro /usr/local/bin/infpro
RUN mkdir -p /profiler-output

RUN echo '#!/bin/sh \n\
python3 -m vllm.entrypoints.openai.api_server \
--port 8000 \
--model /app/model \
--gpu-memory-utilization=0.7 \
--max-model-len=2048 \
--dtype=bfloat16 & \n\
timeout 60s sh -c "until curl -s localhost:8000/health; do sleep 1; done" \n\
exec "$@"' > /entrypoint.sh && chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

CMD ["infpro", "profile", "--output", "/profiler-output/delta.jsonl", "--dynamic", "--delta", "--no-json-string", "--", \
     "vllm", "bench", "serve", \
     "--backend", "vllm", \
     "--model", "/app/model", \
     "--dataset-name", "random", \
     "--num-prompts", "1000", \
     "--request-rate", "inf", \
     "--ready-check-timeout-sec", "0", \
     "--endpoint", "http://localhost:8000/v1/completions"]
