FROM ubuntu:24.04 AS stage1
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="/root/.local/bin:$PATH"
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    python3-dev \
    python3-venv \
    pipx \
    && rm -rf /var/lib/apt/lists/*
RUN pipx install vllm
RUN pipx inject vllm torch-c-dlpack-ext

FROM stage1 AS stage2
COPY dist/*.whl /app/dist/
RUN pipx install /app/dist/*.whl --force
RUN mkdir -p /profiler-output
ENTRYPOINT ["InferenceProfiler", "-o", "/profiler-output", "-t", "1000", "-f", "csv"]
CMD ["vllm", "serve", "/app/model/", "--gpu-memory-utilization=0.7", "--max-model-len=2048", "--dtype=bfloat16"]