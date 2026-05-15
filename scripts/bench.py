#!/opt/vllm/bin/python
"""
Run vLLM benchmarks with InferenceProfiler telemetry.

Usage: ./bench.py [env_file]
"""

import json
import logging
import os
import requests
import socket
import subprocess
import sys
import time
import uuid
from dotenv import load_dotenv
from pathlib import Path
from urllib.parse import urlparse

log = logging.getLogger("bench")
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
os.environ.setdefault("VLLM_TARGET_DEVICE", "cpu")


def env(key):
    v = os.getenv(key)
    if v is None:
        raise EnvironmentError(f"Missing: {key}")
    return v


def ints(raw):
    return [int(x.strip()) for x in raw.split(",")]


def median(vals):
    return sorted(vals)[(len(vals) - 1) // 2]


def main():
    env_file = Path(
        sys.argv[1] if len(sys.argv) > 1 else os.getenv("ENV_FILE", "experiment.env")
    )
    if not env_file.exists():
        log.critical("%s not found", env_file)
        sys.exit(1)
    load_dotenv(env_file)

    experiment = env("EXPERIMENT")
    model = env("MODEL_ID")
    backend = env("VLLM_BACKEND")
    endpoint = env("VLLM_ENDPOINT")
    in_lens = ints(env("INPUT_LENS"))
    out_lens = ints(env("OUTPUT_LENS"))
    concurs = ints(env("CONCURRENCY"))
    iterations = int(env("ITERATIONS"))
    warmup = env("WARMUP_PROMPTS")
    rate = env("REQUEST_RATE")
    num_prompts = env("NUM_PROMPTS")
    burstiness = env("BURSTINESS")
    range_ratio = os.getenv("RANDOM_RANGE_RATIO", "0.0")
    dataset_path = os.getenv("DATASET_PATH")
    dataset_name = os.getenv("DATASET_NAME", "")
    out_dir = os.path.join(env("OUTPUT_DIR"), f"{experiment}.{time.time()}")
    vllm = f"http://{env('VLLM_HOST')}:{env('VLLM_PORT')}"
    profiler = f"http://{env('INFPRO_HOST')}:{env('INFPRO_PORT')}"

    base_in = median(in_lens)
    base_out = median(out_lens)
    base_concur = median(concurs)

    os.makedirs(out_dir, exist_ok=True)
    with open(os.path.join(out_dir, "exp.env"), "w") as f:
        for key, value in os.environ.items():
            f.write(f"{key}={value}\n")

    configs = set()
    if dataset_path:
        for c in concurs:
            configs.add((0, 0, c))
    else:
        for i in in_lens:
            configs.add((i, base_out, base_concur))
        for o in out_lens:
            configs.add((base_in, o, base_concur))
        for c in concurs:
            configs.add((base_in, base_out, c))
    configs = sorted(configs)

    log.info(
        "%s: %d configs x %d iters = %d runs",
        experiment,
        len(configs),
        iterations,
        len(configs) * iterations,
    )

    for i in range(iterations):
        for in_len, out_len, concur in configs:
            uid = str(uuid.uuid4())
            run_path = os.path.join(out_dir, uid)
            os.makedirs(run_path)
            log.info(
                "in=%s out=%s concur=%s iter=%s uuid=%s",
                in_len,
                out_len,
                concur,
                i,
                uid,
            )

            requests.get(f"{vllm}/health", timeout=5).raise_for_status()
            try:
                requests.put(f"{profiler}/collect", json={"uuid": uid}, timeout=5)
            except requests.RequestException:
                log.warning("profiler unreachable")

            start = time.time_ns()
            offset = time.time_ns() - time.monotonic_ns()

            cmd = [
                "vllm",
                "bench",
                "serve",
                "--base-url",
                vllm,
                "--model",
                model,
                "--backend",
                backend,
                "--endpoint",
                endpoint,
                "--dataset-name",
                dataset_name,
                "--request-rate",
                rate,
                "--burstiness",
                burstiness,
                "--num-warmups",
                warmup,
                "--max-concurrency",
                str(concur),
                "--num-prompts",
                num_prompts,
                "--result-dir",
                run_path,
                "--save-result",
                "--save-detailed",
                "--result-filename",
                "results.json",
                "--skip-chat-template",
                "--temperature",
                "0.0",
                "--seed",
                "42"
            ]
            if dataset_path:
                cmd += ["--dataset-path", dataset_path]
            else:
                cmd += [
                    "--random-input-len",
                    str(in_len),
                    "--random-output-len",
                    str(out_len),
                    "--random-range-ratio",
                    range_ratio,
                ]
            proc = subprocess.Popen(cmd)
            try:
                proc.wait()
                if proc.returncode != 0:
                    raise subprocess.CalledProcessError(proc.returncode, cmd)
            except Exception:
                if proc.poll() is None:
                    proc.kill()
                    proc.wait()
                raise

            end = time.time_ns()
            time.sleep(2)

            try:
                requests.delete(f"{profiler}/collect", timeout=5)
            except requests.RequestException as e:
                log.error("stop profiler: %s", e)

            try:
                resp = requests.get(f"{profiler}/files/{uid}.jsonl", timeout=30)
                resp.raise_for_status()
                host = urlparse(profiler).hostname
                Path(run_path, f"{host}.jsonl").write_bytes(resp.content)
            except requests.RequestException as e:
                log.error("download profiler data: %s", e)

            with open(os.path.join(out_dir, "run_log.jsonl"), "a") as f:
                f.write(
                    json.dumps(
                        {
                            "experiment": experiment,
                            "client": socket.gethostname(),
                            "uuid": uid,
                            "num_prompts": num_prompts,
                            "input_len": in_len,
                            "output_len": out_len,
                            "concurrency": concur,
                            "iteration": i,
                            "start_ns": start,
                            "offset_ns": offset,
                            "end_ns": end,
                        }
                    )
                    + "\n"
                )

    log.info("Done. %d runs in %s", len(configs) * iterations, out_dir)


if __name__ == "__main__":
    main()
