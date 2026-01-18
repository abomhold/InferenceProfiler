FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o infpro main.go

FROM vllm/vllm-openai:latest
RUN mkdir -p /profiler-output
RUN cat <<'EOF' > /entrypoint.sh
#!/bin/sh
python3 -m vllm.entrypoints.openai.api_server \
  --port 8000 \
  --model /app/model \
  --gpu-memory-utilization=0.7 \
  --max-model-len=2048 \
  --dtype=bfloat16 &
timeout 60s sh -c 'until curl -s 127.0.0.1:8000/health; do sleep 1; done'
export CUDA_VISIBLE_DEVICES=""
infpro profile --output /profiler-output/delta.jsonl --dynamic --delta --no-json-string -- \
  vllm bench serve --backend vllm --model /app/model --dataset-name random --num-prompts 100 \
  --request-rate 20 --random-input-len 128 --random-output-len 64 \
  --ready-check-timeout-sec 1
EOF
RUN chmod +x /entrypoint.sh
COPY --from=builder /app/infpro /usr/local/bin/infpro
ENTRYPOINT ["/entrypoint.sh"]
