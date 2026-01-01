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

ENTRYPOINT ["/usr/local/bin/profiler", "-o", "/profiler-output", "-t", "100", "-f", "parquet","--stream", "--no-cleanup"]

CMD ["python3", "-c", \
     "from vllm import LLM, SamplingParams; \
      llm = LLM(model='/app/model/', \
                gpu_memory_utilization=0.8, \
                max_model_len=2048, \
                dtype='bfloat16'); \
      sampling_params = SamplingParams(max_tokens=512); \
      prompts = ['Explain the difference between TCP and UDP.'] * 20; \
      print(llm.generate(prompts, sampling_params))"]