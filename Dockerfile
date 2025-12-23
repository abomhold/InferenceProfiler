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
ENTRYPOINT ["InferenceProfiler", "-o", "/profiler-output", "-t", "1000"]
CMD ["/root/.local/share/pipx/venvs/vllm/bin/python", "-c", \
     "from vllm import LLM; \
      llm = LLM(model='/app/model/', \
                gpu_memory_utilization=0.8, \
                max_model_len=2048, \
                dtype='bfloat16'); \
      print(llm.generate('Explain the difference between TCP and UDP.')[0].outputs[0].text)"]