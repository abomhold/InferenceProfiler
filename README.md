# Inference Profiler

System resource profiler for monitoring VM, container, process, and GPU metrics during ML inference workloads. Exports to JSON, CSV, or Parquet.

## Quick Start

```bash
# Build
go build -o profiler ./cmd/profiler

# Run standalone (Ctrl+C to stop)
./profiler -o ./output -t 1000

# Profile a command
./profiler -o ./output -- python serve.py
```

## Features

- **VM Metrics**: CPU, memory, disk I/O, network
- **Container Metrics**: cgroup v1/v2, per-container resource usage
- **Process Metrics**: Per-process CPU, memory, threads, context switches
- **GPU Metrics**: NVIDIA GPU utilization, memory, power, temperature
- **vLLM Metrics**: Inference engine statistics from Prometheus endpoint

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `./profiler-output` | Output directory |
| `-t, --interval` | `1000` | Sampling interval (ms) |
| `-f, --format` | `parquet` | Export format: csv, tsv, parquet |
| `-p, --processes` | `false` | Enable per-process metrics |

## Output Structure

```
output/
├── static_<uuid>.json      # Hardware info (once at start)
├── <uuid>-<timestamp>.json # Per-sample snapshots
└── <uuid>.csv              # Aggregated time series
```

## Metrics Reference

### CPU (`cpu`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `vCpuTime` | Timed[int64] | cs | Total CPU time (user + kernel) |
| `vCpuTimeUserMode` | Timed[int64] | cs | User space execution time |
| `vCpuTimeKernelMode` | Timed[int64] | cs | Kernel execution time |
| `vCpuIdleTime` | Timed[int64] | cs | Idle time |
| `vCpuTimeIOWait` | Timed[int64] | cs | I/O wait time |
| `vCpuTimeIntSrvc` | Timed[int64] | cs | Hardware interrupt time |
| `vCpuTimeSoftIntSrvc` | Timed[int64] | cs | Software interrupt time |
| `vCpuNice` | Timed[int64] | cs | Nice process time |
| `vCpuSteal` | Timed[int64] | cs | Hypervisor stolen time |
| `vCpuContextSwitches` | Timed[int64] | count | Context switches |
| `vLoadAvg` | Timed[float64] | ratio | 1-min load average |
| `vCpuMhz` | Timed[float64] | MHz | Current CPU frequency |

### Memory (`mem`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `vMemoryTotal` | Timed[int64] | bytes | Total RAM |
| `vMemoryFree` | Timed[int64] | bytes | Available memory |
| `vMemoryUsed` | Timed[int64] | bytes | Used memory |
| `vMemoryBuffers` | Timed[int64] | bytes | Kernel buffers |
| `vMemoryCached` | Timed[int64] | bytes | Page cache |
| `vMemoryPercent` | Timed[float64] | % | Usage percentage |
| `vSwapTotal` | Timed[int64] | bytes | Total swap |
| `vSwapFree` | Timed[int64] | bytes | Free swap |
| `vSwapUsed` | Timed[int64] | bytes | Used swap |
| `vPgFault` | Timed[int64] | count | Minor page faults |
| `vMajorPageFault` | Timed[int64] | count | Major page faults |

### Disk (`disk`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `vDiskReadBytes` | Timed[int64] | bytes | Bytes read |
| `vDiskWriteBytes` | Timed[int64] | bytes | Bytes written |
| `vDiskSuccessfulReads` | Timed[int64] | count | Read operations |
| `vDiskSuccessfulWrites` | Timed[int64] | count | Write operations |
| `vDiskReadTime` | Timed[int64] | ms | Time reading |
| `vDiskWriteTime` | Timed[int64] | ms | Time writing |
| `vDiskIOTime` | Timed[int64] | ms | Total I/O time |
| `vDiskIOInProgress` | Timed[int64] | count | In-flight I/O |

### Network (`net`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `vNetworkBytesRecvd` | Timed[int64] | bytes | Bytes received |
| `vNetworkBytesSent` | Timed[int64] | bytes | Bytes sent |
| `vNetworkPacketsRecvd` | Timed[int64] | count | Packets received |
| `vNetworkPacketsSent` | Timed[int64] | count | Packets sent |
| `vNetworkErrorsRecvd` | Timed[int64] | count | Receive errors |
| `vNetworkErrorsSent` | Timed[int64] | count | Send errors |
| `vNetworkDropsRecvd` | Timed[int64] | count | Receive drops |
| `vNetworkDropsSent` | Timed[int64] | count | Send drops |

### Container (`containers`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `cId` | string | - | Container ID |
| `cCgroupVersion` | int | - | Cgroup version (1 or 2) |
| `cCpuTime` | Timed[int64] | ns | Total CPU time |
| `cCpuTimeUserMode` | Timed[int64] | cs | User mode time |
| `cCpuTimeKernelMode` | Timed[int64] | cs | Kernel mode time |
| `cMemoryUsed` | Timed[int64] | bytes | Current memory |
| `cMemoryMaxUsed` | Timed[int64] | bytes | Peak memory |
| `cDiskReadBytes` | Timed[int64] | bytes | Bytes read |
| `cDiskWriteBytes` | Timed[int64] | bytes | Bytes written |
| `cNetworkBytesRecvd` | Timed[int64] | bytes | Network received |
| `cNetworkBytesSent` | Timed[int64] | bytes | Network sent |

### GPU (`nvidia[]`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `gGpuIndex` | int | - | GPU index |
| `gUtilizationGpu` | Timed[int] | % | GPU utilization |
| `gUtilizationMem` | Timed[int] | % | Memory utilization |
| `gMemoryUsedMb` | Timed[int64] | MB | Used VRAM |
| `gMemoryFreeMb` | Timed[int64] | MB | Free VRAM |
| `gTemperatureC` | Timed[int] | °C | Temperature |
| `gPowerDrawW` | Timed[float64] | W | Power draw |
| `gClockGraphicsMhz` | Timed[int] | MHz | Graphics clock |
| `gPerfState` | Timed[string] | - | Performance state |

### vLLM (`vllm`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `system_requests_running` | float64 | count | Active requests |
| `system_requests_waiting` | float64 | count | Queued requests |
| `cache_kv_usage_percent` | float64 | ratio | KV cache usage |
| `requests_finished_total` | float64 | count | Completed requests |
| `tokens_prompt_total` | float64 | count | Prompt tokens |
| `tokens_generation_total` | float64 | count | Generated tokens |
| `latency_ttft_s_sum` | float64 | s | Time to first token (sum) |
| `latency_e2e_s_sum` | float64 | s | End-to-end latency (sum) |

## Timed Values

All metrics that change over time include timestamps for accurate rate calculations:

```json
{
  "vCpuTime": {
    "value": 12345678,
    "time": 1703462400000
  }
}
```

The `time` field is Unix milliseconds.

## Testing

Generate mock procfs/sysfs data:

```bash
./scripts/generate_testdata.sh testdata
```

Run tests:

```bash
go test -v ./...
```

## Data Sources

| Metric Type | Source |
|-------------|--------|
| CPU | `/proc/stat`, `/proc/loadavg`, sysfs cpufreq |
| Memory | `/proc/meminfo`, `/proc/vmstat` |
| Disk | `/proc/diskstats` |
| Network | `/proc/net/dev` |
| Container | `/sys/fs/cgroup` (v1 and v2) |
| GPU | NVML (via go-nvml) |
| vLLM | HTTP `/metrics` (Prometheus format) |

## License

MIT
