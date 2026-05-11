# InferenceProfiler

System metrics collector for ML inference workloads. Collects CPU, memory,
disk, network, container, process, NVIDIA GPU, and vLLM metrics, and emits
them as JSONL.

## Build

```bash
go build -o bin/infpro main.go
```

The NVIDIA collector uses [go-nvml](https://github.com/NVIDIA/go-nvml) and
loads `libnvidia-ml.so` at runtime via dlopen. The binary builds on hosts
without a GPU; collectors that fail to initialize are disabled at startup.

`make build` cross-builds a Linux/amd64 binary in an ephemeral Docker
container — useful when developing on macOS or arm64.

## Usage

```
infpro [command] [flags]
```

### Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `continuous` | `c` (default) | Collect on an interval until Ctrl+C / SIGTERM |
| `snapshot`   | `s`           | Single collection pass, then exit |
| `server`     | `ser`         | HTTP API server for remote control |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-output DIR`        | stdout | Write to `DIR/{uuid}.jsonl` instead of stdout |
| `-uuid ID`           | random | Run identifier |
| `-interval MS`       | 1000   | Collection interval in milliseconds |
| `-flatten`           | false  | Flatten nested structs to top-level keys |
| `-no-vm`             | false  | Disable VM metrics (cpu, mem, disk, net) |
| `-no-container`      | false  | Disable container/cgroup metrics |
| `-no-procs`          | false  | Disable process metrics |
| `-no-nvidia`         | false  | Disable NVIDIA GPU metrics |
| `-no-vllm`           | false  | Disable vLLM metrics |
| `-no-vllm-hist`      | false  | Disable vLLM histogram collection |
| `-vllm-endpoint URL` | `http://localhost:8000/metrics` | vLLM Prometheus endpoint |
| `-disabled LIST`     | (none) | Comma-separated collectors to disable (`vm,container,process,nvidia,vllm,vllm-hist`) |
| `-port PORT`         | 8888   | HTTP port (server mode) |
| `-debug`             | false  | Verbose debug logging to stderr |
| `-poll-stats`        | false  | Show per-collector poller statistics on exit |
| `-pprof ADDR`        | (off)  | Enable pprof server (e.g. `localhost:6060`) |

### Examples

```bash
# Continuous to stdout at 1 s
infpro

# Snapshot to stdout
infpro s

# Continuous to a directory at 100 ms
infpro c -output ./metrics -interval 100

# Skip GPU and vLLM collectors
infpro -no-nvidia -no-vllm

# Same via -disabled
infpro -disabled nvidia,vllm

# Server mode
infpro server -output ./data
infpro ser -port 9090
```

## Output format

JSONL. The first record on every run is the static line — collector
identity, system info, and run UUID:

```json
{"uuid": "...", "timestamp": <ns>, "Vm": {...}, "Nvidia": [...], "Vllm": {...}}
```

Subsequent records are dynamic ticks, one per interval:

```json
{"timestamp": <ns>, "Vm": {...}, "Container": {...}, "Process": [...], "Nvidia": [...], "Vllm": {...}}
```

Section keys depend on which collectors initialized successfully. Dynamic
metric values are `{V, T}` pairs where `T` is a per-field timestamp for
rate calculations.

## HTTP API (server mode)

`infpro server` binds `0.0.0.0:<port>` (port defaults to `8888`).

| Method | Path             | Description |
|--------|------------------|-------------|
| GET    | `/health`        | Health check, returns `ok` |
| GET    | `/snapshot`      | Triggers a fresh parallel poll across collectors and returns `{"static": {...}, "tick": {...}}`. Works whether or not a continuous run is active. |
| GET    | `/collect`       | Current state and run info |
| PUT    | `/collect`       | Start a continuous run (body: `{"uuid": "..."}`, uuid optional — server generates one if omitted) |
| DELETE | `/collect`       | Stop and flush |
| GET    | `/files`         | List output files (optional `?uuid=xxx` prefix filter) |
| GET    | `/files/{uuid}`  | Stream the first file whose name starts with `{uuid}` (supports `Range` for resume) |

A Postman collection covering the full surface is at
[`docs/InferenceProfiler.postman_collection.json`](docs/InferenceProfiler.postman_collection.json).

Quick smoke test against a local server:

```bash
curl -s localhost:8888/health
curl -s localhost:8888/snapshot | jq .
curl -sX PUT  localhost:8888/collect -d '{}'
curl -s        localhost:8888/collect
curl -sX DELETE localhost:8888/collect
curl -s localhost:8888/files
```

## Environment overrides

Every flag has an `INFPRO_<NAME>` env var equivalent. `<NAME>` is the
upper-cased flag with dashes replaced by underscores. Flags on the
command line take precedence over env vars.

Examples:

| Variable | Maps to |
|----------|---------|
| `INFPRO_INTERVAL`      | `-interval` |
| `INFPRO_OUTPUT`        | `-output` |
| `INFPRO_UUID`          | `-uuid` |
| `INFPRO_DEBUG`         | `-debug` |
| `INFPRO_POLL_STATS`    | `-poll-stats` |
| `INFPRO_VLLM_ENDPOINT` | `-vllm-endpoint` |
| `INFPRO_DISABLED`      | `-disabled` |
| `INFPRO_PORT`          | `-port` |
| `INFPRO_PPROF`         | `-pprof` |

## Deployment (AWS)

The repo includes a self-contained OpenTofu + Make deployment for running
a vLLM server alongside `infpro` on EC2 spot instances, with a separate
client node for driving benchmarks.

Layout:

```
main.tofu                     One-file OpenTofu config (server + client + VPC + SG + key)
Makefile                      Build, infra, deploy, restart, pull-results targets
configs/
  default.env                 Committed defaults: project, region, vLLM args, bench params
  cloud-init/
    server.yaml               Server node bootstrap (CUDA, DCGM, vLLM venv)
    client.yaml               Client node bootstrap (vLLM bench CLI)
  systemd/
    benchmark.slice           cgroup slice for vLLM
    profiler.slice            cgroup slice for infpro (pinned CPUs, memory cap)
    vllm.service              vLLM systemd unit
    infpro.service            infpro server systemd unit
```

Secrets and the AWS profile name go in a `secrets.env` **one directory
above the repo** (the Makefile sources it via `-include ../secrets.env`).
At minimum it needs:

```
TF_VAR_PROFILE=your-aws-profile-name
```

Typical end-to-end flow:

```bash
make infra-up            # tofu init + build + apply (creates server + client)
make deploy              # rsync binary, env, scripts, systemd units to nodes
                         # then daemon-reload + restart vllm + infpro on server
make ssh-server          # poke around on the server
make pull-snapshot       # live tick from the server (uses GET /snapshot + jq)
make pull-results        # rsync benchmark results from client
make infra-down          # destroy
```

`make help` lists every target.

The `start_bench` script on the client (installed by `make deploy-client`)
runs `scripts/bench.py`, which drives `vllm bench serve` through a matrix
of input/output lengths and concurrency levels, telling the profiler
server when to start and stop each run via the HTTP API above.

## Metric reference

See [`docs/InferenceProfilerDataDictionary.csv`](docs/InferenceProfilerDataDictionary.csv)
for the full list of collected fields.