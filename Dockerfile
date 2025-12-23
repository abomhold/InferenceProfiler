FROM ubuntu:24.04
ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /app
RUN mkdir -p /profiler-output
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    python3-dev \
    python3-venv \
    pipx \
    sysbench \
    && rm -rf /var/lib/apt/lists/*
ENV PATH="/root/.local/bin:$PATH"
COPY dist/*.whl /app/dist/
RUN pipx install /app/dist/*.whl --force
ENTRYPOINT ["inference-profiler", "-o", "/profiler-output", "-t", "1000"]

# Default command
# CMD ["vllm", "serve", "/app/model/", "--gpu-memory-utilization=0.7", "--max-model-len=2048", "--dtype=bfloat16"]
CMD ["sysbench", "--test=cpu", "--cpu-max-prime=20000", "--max-requests=4000", "run"]