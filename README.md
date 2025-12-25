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

## Project Structure

```
.
├── cmd/
│   └── profiler/
│       └── main.go           # Entry point
├── internal/
│   ├── collector/
│   │   ├── base.go           # BaseCollector with I/O helpers
│   │   ├── manager.go        # Orchestrates all collectors
│   │   ├── cpu.go            # CPU metrics
│   │   ├── mem.go            # Memory metrics
│   │   ├── disk.go           # Disk I/O metrics
│   │   ├── net.go            # Network metrics
│   │   ├── container.go      # Container/cgroup metrics
│   │   ├── proc.go           # Per-process metrics
│   │   ├── nvidia.go         # GPU metrics (NVML)
│   │   └── vllm.go           # vLLM inference metrics
│   └── exporter/
│       └── exporter.go       # JSON/CSV output
├── scripts/
│   └── generate_testdata.sh  # Test data generator
├── go.mod
├── Makefile
└── README.md
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
| `vDiskSectorReads` | Timed[int64] | sectors | Sectors read |
| `vDiskSectorWrites` | Timed[int64] | sectors | Sectors written |
| `vDiskReadBytes` | Timed[int64] | bytes | Bytes read |
| `vDiskWriteBytes` | Timed[int64] | bytes | Bytes written |
| `vDiskSuccessfulReads` | Timed[int64] | count | Read operations |
| `vDiskSuccessfulWrites` | Timed[int64] | count | Write operations |
| `vDiskMergedReads` | Timed[int64] | count | Merged reads |
| `vDiskMergedWrites` | Timed[int64] | count | Merged writes |
| `vDiskReadTime` | Timed[int64] | ms | Time reading |
| `vDiskWriteTime` | Timed[int64] | ms | Time writing |
| `vDiskIOTime` | Timed[int64] | ms | Total I/O time |
| `vDiskWeightedIOTime` | Timed[int64] | ms | Weighted I/O time |
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
| `cNumProcessors` | int | - | CPU count |
| `cMemoryUsed` | Timed[int64] | bytes | Current memory |
| `cMemoryMaxUsed` | Timed[int64] | bytes | Peak memory |
| `cDiskReadBytes` | Timed[int64] | bytes | Bytes read |
| `cDiskWriteBytes` | Timed[int64] | bytes | Bytes written |
| `cNetworkBytesRecvd` | Timed[int64] | bytes | Network received |
| `cNetworkBytesSent` | Timed[int64] | bytes | Network sent |

### Process (`processes`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `pId` | int | - | Process ID |
| `pName` | string | - | Process name |
| `pCmdline` | string | - | Command line |
| `pNumThreads` | Timed[int64] | count | Thread count |
| `pCpuTimeUserMode` | Timed[int64] | cs | User CPU time |
| `pCpuTimeKernelMode` | Timed[int64] | cs | Kernel CPU time |
| `pChildrenUserMode` | Timed[int64] | cs | Children user time |
| `pChildrenKernelMode` | Timed[int64] | cs | Children kernel time |
| `pVoluntaryContextSwitches` | Timed[int64] | count | Voluntary ctx switches |
| `pNonvoluntaryContextSwitches` | Timed[int64] | count | Involuntary ctx switches |
| `pBlockIODelays` | Timed[int64] | cs | Block I/O delays |
| `pVirtualMemoryBytes` | Timed[int64] | bytes | Virtual memory |
| `pResidentSetSize` | Timed[int64] | bytes | RSS memory |

### GPU (`nvidia[]`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `gGpuIndex` | int | - | GPU index |
| `gUtilizationGpu` | Timed[int] | % | GPU utilization |
| `gUtilizationMem` | Timed[int] | % | Memory utilization |
| `gMemoryTotalMb` | Timed[int64] | MB | Total VRAM |
| `gMemoryUsedMb` | Timed[int64] | MB | Used VRAM |
| `gMemoryFreeMb` | Timed[int64] | MB | Free VRAM |
| `gBar1UsedMb` | Timed[int64] | MB | BAR1 used |
| `gBar1FreeMb` | Timed[int64] | MB | BAR1 free |
| `gTemperatureC` | Timed[int] | °C | Temperature |
| `gFanSpeed` | Timed[int] | % | Fan speed |
| `gPowerDrawW` | Timed[float64] | W | Power draw |
| `gPowerLimitW` | Timed[float64] | W | Power limit |
| `gClockGraphicsMhz` | Timed[int] | MHz | Graphics clock |
| `gClockSmMhz` | Timed[int] | MHz | SM clock |
| `gClockMemMhz` | Timed[int] | MHz | Memory clock |
| `gPcieTxKbps` | Timed[int] | KB/s | PCIe TX |
| `gPcieRxKbps` | Timed[int] | KB/s | PCIe RX |
| `gPerfState` | Timed[string] | - | Performance state |
| `gEccSingleBitErrors` | Timed[int64] | count | ECC single-bit |
| `gEccDoubleBitErrors` | Timed[int64] | count | ECC double-bit |

### vLLM (`vllm`)

| Field | Type | Unit | Description |
|-------|------|------|-------------|
| `system_requests_running` | float64 | count | Active requests |
| `system_requests_waiting` | float64 | count | Queued requests |
| `system_engine_sleep_state` | float64 | - | Engine sleep state |
| `system_preemptions_total` | float64 | count | Preemptions |
| `cache_kv_usage_percent` | float64 | ratio | KV cache usage |
| `cache_prefix_hits` | float64 | count | Prefix cache hits |
| `cache_prefix_queries` | float64 | count | Prefix cache queries |
| `requests_finished_total` | float64 | count | Completed requests |
| `requests_corrupted_total` | float64 | count | Corrupted requests |
| `tokens_prompt_total` | float64 | count | Prompt tokens |
| `tokens_generation_total` | float64 | count | Generated tokens |
| `latency_ttft_s_sum` | float64 | s | Time to first token (sum) |
| `latency_e2e_s_sum` | float64 | s | End-to-end latency (sum) |
| `latency_queue_s_sum` | float64 | s | Queue latency (sum) |
| `latency_inference_s_sum` | float64 | s | Inference latency (sum) |
| `latency_prefill_s_sum` | float64 | s | Prefill latency (sum) |
| `latency_decode_s_sum` | float64 | s | Decode latency (sum) |
| `latency_inter_token_s_sum` | float64 | s | Inter-token latency (sum) |

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
| Process | `/proc/[pid]/stat`, `/proc/[pid]/status`, `/proc/[pid]/statm` |
| GPU | NVML (via go-nvml) |
| vLLM | HTTP `/metrics` (Prometheus format) |

## License

MIT
