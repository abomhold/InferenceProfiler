# Profiler Metrics Reference

This document provides a comprehensive reference for all metrics collected.
---

## Table of Contents

1. [Command-Line Options](#command-line-options)
2. [Output Formats](#output-formats)
3. [Static System Information](#static-system-information)
4. [VM-Level CPU Metrics](#vm-level-cpu-metrics)
5. [VM-Level Memory Metrics](#vm-level-memory-metrics)
6. [VM-Level Disk I/O Metrics](#vm-level-disk-io-metrics)
7. [VM-Level Network Metrics](#vm-level-network-metrics)
8. [Container-Level Metrics](#container-level-metrics)
9. [Process-Level Metrics](#process-level-metrics)
10. [NVIDIA GPU Metrics](#nvidia-gpu-metrics)
11. [vLLM Inference Metrics](#vllm-inference-metrics)

---

## Command-Line Options

### Output Options

| Flag          | Default              | Description                                                                                     |
|---------------|----------------------|-------------------------------------------------------------------------------------------------|
| `-o <dir>`    | `./profiler-aggregate` | Output directory for logs and exported data                                                   |
| `-t <ms>`     | `1000`               | Sampling interval in milliseconds                                                               |
| `-f <format>` | `jsonl`              | Export format: `jsonl`, `parquet`, `csv`, `tsv`                                                 |
| `-no-flatten` | `false`              | When set, disables flattening of nested data (GPUs, processes) to columns; stores as JSON strings instead |
| `-no-cleanup` | `false`              | When set, keeps the intermediary JSON snapshot files after final export                         |


### Collector Toggles

All collectors are enabled by default.

| Flag            | Default | Description                                                     |
|-----------------|---------|-----------------------------------------------------------------|
| `-no-cpu`       | `false` | Disable CPU metrics collection                                  |
| `-no-memory`    | `false` | Disable memory metrics collection                               |
| `-no-disk`      | `false` | Disable disk I/O metrics collection                             |
| `-no-network`   | `false` | Disable network metrics collection                              |
| `-no-container` | `false` | Disable container/cgroup metrics collection                     |
| `-no-procs`     | `false` | Disable per-process metrics collection                          |
| `-no-nvidia`    | `false` | Disable all NVIDIA GPU metrics collection                       |
| `-no-gpu-procs` | `false` | Disable GPU process enumeration (still collects GPU metrics)    |
| `-no-vllm`      | `false` | Disable vLLM metrics collection                                 |

### Example Usage

```bash
# Minimal: just CPU/memory, TSV output, no flattening
./profiler -no-disk -no-network -no-container -no-procs -no-nvidia -no-vllm -f tsv -no-flatten

# Full GPU monitoring without process overhead
./profiler -no-procs -no-gpu-procs -f parquet

# CSV with no processes, flattened (default)
./profiler -no-procs -f csv

# Profile a subprocess
./profiler -f jsonl -- python train.py --epochs 10
```

---

## Output Formats

### Export Formats

| Format    | Extension  | Description                                                        |
|-----------|------------|--------------------------------------------------------------------|
| `jsonl`   | `.jsonl`   | JSON Lines - one JSON object per line, good for streaming          |
| `parquet` | `.parquet` | Columnar format - efficient for analytics, supports dynamic schema |
| `csv`     | `.csv`     | Comma-separated values - universal compatibility                   |
| `tsv`     | `.tsv`     | Tab-separated values - handles commas in data better               |

### Flatten Modes

The `-no-flatten` flag controls how nested data (GPUs, processes) is represented:

#### Flattened Mode (default, `-no-flatten` not set)

GPUs and processes are expanded into individual columns with indexed prefixes:

```json
{
  "timestamp": 1735166000000,
  "vCpuTime": 123456,
  "vCpuTimeT": 1735166000001,
  "nvidiaGpuCount": 2,
  "nvidia0UtilizationGpu": 85,
  "nvidia0UtilizationGpuT": 1735166000002,
  "nvidia0MemoryUsedMb": 8192,
  "nvidia1UtilizationGpu": 72,
  "nvidia1MemoryUsedMb": 6144,
  "processCount": 150,
  "process0Pid": 1,
  "process0Name": "systemd",
  "process0ResidentSetSize": 8388608,
  "process1Pid": 2,
  "process1Name": "kthreadd"
}
```

#### JSON String Mode (`-no-flatten` set)

GPUs and processes are serialized as JSON strings, resulting in fewer columns:

```json
{
  "timestamp": 1735166000000,
  "vCpuTime": 123456,
  "vCpuTimeT": 1735166000001,
  "nvidiaGpuCount": 2,
  "nvidiaGpusJson": "[{\"index\":0,\"utilizationGpu\":85,...},{\"index\":1,...}]",
  "processCount": 150,
  "processesJson": "[{\"pId\":1,\"pName\":\"systemd\",...},...]"
}
```

### Timestamp Convention

All dynamic metrics include per-field timestamps with a `T` suffix:

| Field           | Description                                                        |
|-----------------|--------------------------------------------------------------------|
| `timestamp`     | Overall collection cycle timestamp (nanoseconds since Unix epoch)  |
| `{metricName}`  | The metric value                                                   |
| `{metricName}T` | Timestamp when this specific metric was collected (nanoseconds)    |

Example:

```json
{
  "vCpuTime": 123456,
  "vCpuTimeT": 1735166000001,
  "vMemoryUsed": 8589934592,
  "vMemoryUsedT": 1735166000003
}
```

### Output Files

Each profiling session produces:

| File                      | Description                                           |
|---------------------------|-------------------------------------------------------|
| `{uuid}.json`             | Static system information (collected once at startup) |
| `{uuid}-{timestamp}.json` | Individual snapshot files (intermediate)              |
| `{uuid}.{ext}`            | Final aggregated output (jsonl/parquet/csv/tsv)       |

---

## Static System Information

Static information is collected once at profiler startup and saved to `{uuid}.json`.

### Session & Host Identification

| Key            | Type   | Source                                                | Description               |
|----------------|--------|-------------------------------------------------------|---------------------------|
| `uuid`         | string | Generated UUID v4                                     | Unique session identifier |
| `vId`          | string | `/sys/class/dmi/id/product_uuid` or `/etc/machine-id` | VM/instance identifier    |
| `vHostname`    | string | `syscall.Uname().Nodename`                            | System hostname           |
| `vBootTime`    | int64  | `/proc/stat` `btime` field                            | System boot time (Unix timestamp seconds) |

### CPU Static Info

| Key              | Type   | Source                                | Description                                    |
|------------------|--------|---------------------------------------|------------------------------------------------|
| `vNumProcessors` | int    | `runtime.NumCPU()`                    | Number of logical CPUs/cores                   |
| `vCpuType`       | string | `/proc/cpuinfo` `model name`          | Processor model name                           |
| `vCpuCache`      | string | `/sys/devices/system/cpu/cpu*/cache/` | Cache sizes formatted as "L1d:32K L1i:32K L2:256K L3:8M" |
| `vKernelInfo`    | string | `syscall.Uname()`                     | Full kernel version string                     |

### Memory Static Info

| Key                 | Type  | Source                             | Description                |
|---------------------|-------|------------------------------------|----------------------------|
| `vMemoryTotalBytes` | int64 | `/proc/meminfo` `MemTotal` × 1024  | Total physical RAM in bytes |
| `vSwapTotalBytes`   | int64 | `/proc/meminfo` `SwapTotal` × 1024 | Total swap space in bytes   |

### Container Static Info

| Key              | Type   | Source              | Description                    |
|------------------|--------|---------------------|--------------------------------|
| `cId`            | string | `/proc/self/cgroup` | Container ID or hostname       |
| `cNumProcessors` | int64  | `runtime.NumCPU()`  | Available processors           |
| `cCgroupVersion` | int64  | Auto-detected       | Cgroup version (1 or 2)        |

### Network Static Info

| Key                 | Type   | Source             | Description                        |
|---------------------|--------|--------------------|------------------------------------|
| `networkInterfaces` | string | `/sys/class/net/`  | JSON array of network interface info |

Each network interface object contains:

| Field       | Type   | Description                              |
|-------------|--------|------------------------------------------|
| `name`      | string | Interface name (e.g., "eth0")            |
| `mac`       | string | MAC address                              |
| `state`     | string | Operational state (up/down)              |
| `mtu`       | int64  | Maximum transmission unit                |
| `speedMbps` | int64  | Link speed in Mbps (if available)        |

### Disk Static Info

| Key     | Type   | Source              | Description                |
|---------|--------|---------------------|----------------------------|
| `disks` | string | `/sys/class/block/` | JSON array of disk info    |

Each disk object contains:

| Field       | Type   | Description           |
|-------------|--------|-----------------------|
| `name`      | string | Device name (e.g., "sda", "nvme0n1") |
| `model`     | string | Disk model            |
| `vendor`    | string | Disk vendor           |
| `sizeBytes` | int64  | Total size in bytes   |

### NVIDIA Static Info

| Key                   | Type   | Source                             | Description           |
|-----------------------|--------|------------------------------------|-----------------------|
| `nvidiaDriverVersion` | string | `nvmlSystemGetDriverVersion()`     | NVIDIA driver version |
| `nvidiaCudaVersion`   | string | `nvmlSystemGetCudaDriverVersion()` | CUDA driver version   |
| `nvidiaGpus`          | string | NVML device enumeration            | JSON array of GPU info |

Each GPU object in `nvidiaGpus` contains:

| Field           | Type   | Source                            | Description              |
|-----------------|--------|-----------------------------------|--------------------------|
| `index`         | int    | Loop index                        | Zero-based GPU index     |
| `name`          | string | `nvmlDeviceGetName()`             | GPU model name           |
| `uuid`          | string | `nvmlDeviceGetUUID()`             | Unique GPU identifier    |
| `totalMemoryMb` | int64  | `nvmlDeviceGetMemoryInfo().total` | Total frame buffer in MB |

---

## VM-Level CPU Metrics

Prefix: `v` (VM-level). Disable with `-no-cpu`.

| Key                   | Type    | Unit         | Source                                   | Description                       |
|-----------------------|---------|--------------|------------------------------------------|-----------------------------------|
| `vCpuTime`            | int64   | centiseconds | `/proc/stat` cpu line: `user` + `system` | Total CPU time (user + kernel)    |
| `vCpuTimeUserMode`    | int64   | centiseconds | `/proc/stat` column 1                    | User-space execution time         |
| `vCpuTimeKernelMode`  | int64   | centiseconds | `/proc/stat` column 3                    | Kernel-mode execution time        |
| `vCpuIdleTime`        | int64   | centiseconds | `/proc/stat` column 4                    | Idle time                         |
| `vCpuTimeIOWait`      | int64   | centiseconds | `/proc/stat` column 5                    | I/O wait time                     |
| `vCpuTimeIntSrvc`     | int64   | centiseconds | `/proc/stat` column 6                    | Hardware interrupt time           |
| `vCpuTimeSoftIntSrvc` | int64   | centiseconds | `/proc/stat` column 7                    | Software interrupt time           |
| `vCpuNice`            | int64   | centiseconds | `/proc/stat` column 2                    | Nice (low priority) time          |
| `vCpuSteal`           | int64   | centiseconds | `/proc/stat` column 8                    | Hypervisor stolen time            |
| `vCpuContextSwitches` | int64   | count        | `/proc/stat` `ctxt` line                 | Total context switches since boot |
| `vLoadAvg`            | float64 | ratio        | `/proc/loadavg` column 1                 | 1-minute load average             |
| `vCpuMhz`             | float64 | MHz          | `/sys/devices/system/cpu/*/cpufreq/`     | Current average CPU frequency     |

**Notes:**

- All time values are cumulative since system boot
- 1 centisecond = 10 milliseconds
- To calculate utilization: `(delta_busy / delta_total) × 100`

---

## VM-Level Memory Metrics

Prefix: `v` (VM-level). Disable with `-no-memory`.

| Key                      | Type    | Unit    | Source                                    | Description                  |
|--------------------------|---------|---------|-------------------------------------------|------------------------------|
| `vMemoryTotal`           | int64   | bytes   | `/proc/meminfo` `MemTotal`                | Total physical RAM           |
| `vMemoryFree`            | int64   | bytes   | `/proc/meminfo` `MemAvailable`            | Available memory             |
| `vMemoryUsed`            | int64   | bytes   | Derived                                   | Actively used memory         |
| `vMemoryBuffers`         | int64   | bytes   | `/proc/meminfo` `Buffers`                 | Kernel buffer memory         |
| `vMemoryCached`          | int64   | bytes   | `/proc/meminfo` `Cached` + `SReclaimable` | Page cache memory            |
| `vMemoryPercent`         | float64 | percent | Derived                                   | RAM usage percentage         |
| `vMemorySwapTotal`       | int64   | bytes   | `/proc/meminfo` `SwapTotal`               | Total swap space             |
| `vMemorySwapFree`        | int64   | bytes   | `/proc/meminfo` `SwapFree`                | Free swap space              |
| `vMemorySwapUsed`        | int64   | bytes   | Derived                                   | Used swap space              |
| `vMemoryPgFault`         | int64   | count   | `/proc/vmstat` `pgfault`                  | Minor page faults            |
| `vMemoryMajorPageFault`  | int64   | count   | `/proc/vmstat` `pgmajfault`               | Major page faults (disk I/O) |

---

## VM-Level Disk I/O Metrics

Prefix: `v` (VM-level). Disable with `-no-disk`.

| Key                     | Type  | Unit    | Source                   | Description                    |
|-------------------------|-------|---------|--------------------------|--------------------------------|
| `vDiskSectorReads`      | int64 | sectors | `/proc/diskstats` col 5  | Sectors read (1 sector = 512B) |
| `vDiskSectorWrites`     | int64 | sectors | `/proc/diskstats` col 9  | Sectors written                |
| `vDiskReadBytes`        | int64 | bytes   | Derived                  | Total bytes read               |
| `vDiskWriteBytes`       | int64 | bytes   | Derived                  | Total bytes written            |
| `vDiskSuccessfulReads`  | int64 | count   | `/proc/diskstats` col 3  | Completed read operations      |
| `vDiskSuccessfulWrites` | int64 | count   | `/proc/diskstats` col 7  | Completed write operations     |
| `vDiskMergedReads`      | int64 | count   | `/proc/diskstats` col 4  | Merged read requests           |
| `vDiskMergedWrites`     | int64 | count   | `/proc/diskstats` col 8  | Merged write requests          |
| `vDiskReadTime`         | int64 | ms      | `/proc/diskstats` col 6  | Time spent reading             |
| `vDiskWriteTime`        | int64 | ms      | `/proc/diskstats` col 10 | Time spent writing             |
| `vDiskIOInProgress`     | int64 | count   | `/proc/diskstats` col 11 | I/O operations in flight       |
| `vDiskIOTime`           | int64 | ms      | `/proc/diskstats` col 12 | Total I/O time                 |
| `vDiskWeightedIOTime`   | int64 | ms      | `/proc/diskstats` col 13 | Weighted I/O time              |

**Device Filtering:** Only physical disks are included: `sd*`, `hd*`, `vd*`, `xvd*`, `nvme*n*`, `mmcblk*`. Partitions and loopback devices are excluded.

---

## VM-Level Network Metrics

Prefix: `v` (VM-level). Disable with `-no-network`.

| Key                    | Type  | Unit  | Source                 | Description         |
|------------------------|-------|-------|------------------------|---------------------|
| `vNetworkBytesRecvd`   | int64 | bytes | `/proc/net/dev` col 0  | Bytes received      |
| `vNetworkBytesSent`    | int64 | bytes | `/proc/net/dev` col 8  | Bytes transmitted   |
| `vNetworkPacketsRecvd` | int64 | count | `/proc/net/dev` col 1  | Packets received    |
| `vNetworkPacketsSent`  | int64 | count | `/proc/net/dev` col 9  | Packets transmitted |
| `vNetworkErrorsRecvd`  | int64 | count | `/proc/net/dev` col 2  | Receive errors      |
| `vNetworkErrorsSent`   | int64 | count | `/proc/net/dev` col 10 | Transmit errors     |
| `vNetworkDropsRecvd`   | int64 | count | `/proc/net/dev` col 3  | Dropped on receive  |
| `vNetworkDropsSent`    | int64 | count | `/proc/net/dev` col 11 | Dropped on transmit |

**Note:** Loopback interface (`lo`) is excluded. Values are aggregated across all interfaces.

---

## Container-Level Metrics

Prefix: `c` (Container-level). Disable with `-no-container`.

Auto-detects cgroup v1 or v2.

### Dynamic Container Metrics

| Key                  | Type   | Unit         | Source                                        | Description              |
|----------------------|--------|--------------|-----------------------------------------------|--------------------------|
| `cCpuTime`           | int64  | nanoseconds  | `cpuacct.usage` / `cpu.stat usage_usec`       | Total CPU time           |
| `cCpuTimeUserMode`   | int64  | centiseconds | `cpuacct.stat` / `cpu.stat user_usec`         | User mode CPU time       |
| `cCpuTimeKernelMode` | int64  | centiseconds | `cpuacct.stat` / `cpu.stat system_usec`       | Kernel mode CPU time     |
| `cMemoryUsed`        | int64  | bytes        | `memory.usage_in_bytes` / `memory.current`    | Current memory usage     |
| `cMemoryMaxUsed`     | int64  | bytes        | `memory.max_usage_in_bytes` / `memory.peak`   | Peak memory usage        |
| `cDiskReadBytes`     | int64  | bytes        | `blkio.throttle.io_service_bytes` / `io.stat` | Bytes read               |
| `cDiskWriteBytes`    | int64  | bytes        | `blkio.throttle.io_service_bytes` / `io.stat` | Bytes written            |
| `cDiskSectorIO`      | int64  | sectors      | `blkio.sectors` (v1 only)                     | Total sector I/O         |
| `cNetworkBytesRecvd` | int64  | bytes        | `/proc/net/dev`                               | Network bytes received   |
| `cNetworkBytesSent`  | int64  | bytes        | `/proc/net/dev`                               | Network bytes sent       |
| `cPgFault`           | int64  | count        | `memory.stat pgfault`                         | Page faults              |
| `cMajorPgFault`      | int64  | count        | `memory.stat pgmajfault`                      | Major page faults        |
| `cNumProcesses`      | int64  | count        | `pids.current` / tasks file                   | Process count in cgroup  |
| `cCpuPerCpuJson`     | string | -            | `cpuacct.usage_percpu` (v1 only)              | Per-CPU times as JSON array |

---

## Process-Level Metrics

Prefix: `process{i}` (flattened) or `processesJson` (JSON mode). Disable with `-no-procs`.

| Key                          | Type   | Unit         | Source                          | Description                  |
|------------------------------|--------|--------------|---------------------------------|------------------------------|
| `pId`                        | int64  | -            | `/proc/[pid]`                   | Process ID                   |
| `pName`                      | string | -            | `/proc/[pid]/status`            | Executable name              |
| `pCmdline`                   | string | -            | `/proc/[pid]/cmdline`           | Full command line            |
| `pNumThreads`                | int64  | count        | `/proc/[pid]/stat` field 19     | Thread count                 |
| `pCpuTimeUserMode`           | int64  | centiseconds | `/proc/[pid]/stat` field 14     | User CPU time                |
| `pCpuTimeKernelMode`         | int64  | centiseconds | `/proc/[pid]/stat` field 15     | Kernel CPU time              |
| `pChildrenUserMode`          | int64  | centiseconds | `/proc/[pid]/stat` field 16     | Children user time           |
| `pChildrenKernelMode`        | int64  | centiseconds | `/proc/[pid]/stat` field 17     | Children kernel time         |
| `pVoluntaryContextSwitches`  | int64  | count        | `/proc/[pid]/status`            | Voluntary context switches   |
| `pNonvoluntaryContextSwitches` | int64 | count       | `/proc/[pid]/status`            | Involuntary context switches |
| `pBlockIODelays`             | int64  | centiseconds | `/proc/[pid]/stat` field 42     | Block I/O wait time          |
| `pVirtualMemoryBytes`        | int64  | bytes        | `/proc/[pid]/stat` field 23     | Virtual memory size          |
| `pResidentSetSize`           | int64  | bytes        | `/proc/[pid]/statm` × page size | Physical memory (RSS)        |

### Flattened Output Example

```json
{
  "processCount": 150,
  "process0Pid": 1,
  "process0PidT": 1735166000001,
  "process0Name": "systemd",
  "process0NameT": 1735166000001,
  "process0ResidentSetSize": 8388608,
  "process1Pid": 2,
  "process1Name": "kthreadd"
}
```

---

## NVIDIA GPU Metrics

Prefix: `nvidia{i}` (flattened) or `nvidiaGpusJson` (JSON mode). Disable with `-no-nvidia`. Disable process enumeration only with `-no-gpu-procs`.

| Key                | Type    | Unit    | Source                                   | Description                   |
|--------------------|---------|---------|------------------------------------------|-------------------------------|
| `utilizationGpu`   | int64   | percent | `nvmlDeviceGetUtilizationRates().gpu`    | GPU compute utilization       |
| `utilizationMem`   | int64   | percent | `nvmlDeviceGetUtilizationRates().memory` | Memory controller utilization |
| `memoryUsedMb`     | int64   | MB      | `nvmlDeviceGetMemoryInfo().used`         | Used frame buffer             |
| `memoryFreeMb`     | int64   | MB      | `nvmlDeviceGetMemoryInfo().free`         | Free frame buffer             |
| `bar1UsedMb`       | int64   | MB      | `nvmlDeviceGetBAR1MemoryInfo()`          | Used BAR1 memory              |
| `temperatureC`     | int64   | °C      | `nvmlDeviceGetTemperature()`             | GPU temperature               |
| `fanSpeed`         | int64   | percent | `nvmlDeviceGetFanSpeed()`                | Fan speed                     |
| `clockGraphicsMhz` | int64   | MHz     | `nvmlDeviceGetClockInfo(GRAPHICS)`       | Graphics clock                |
| `clockSmMhz`       | int64   | MHz     | `nvmlDeviceGetClockInfo(SM)`             | SM clock                      |
| `clockMemMhz`      | int64   | MHz     | `nvmlDeviceGetClockInfo(MEM)`            | Memory clock                  |
| `pcieTxKbps`       | int64   | KB/s    | `nvmlDeviceGetPcieThroughput(TX)`        | PCIe TX throughput            |
| `pcieRxKbps`       | int64   | KB/s    | `nvmlDeviceGetPcieThroughput(RX)`        | PCIe RX throughput            |
| `powerDrawW`       | float64 | watts   | `nvmlDeviceGetPowerUsage()`              | Power consumption             |
| `perfState`        | string  | -       | `nvmlDeviceGetPerformanceState()`        | Performance state (P0-P12)    |
| `processCount`     | int64   | count   | Process enumeration                      | GPU processes count           |
| `processesJson`    | string  | -       | Process enumeration                      | GPU processes (JSON)          |

### Flattened Output Example

```json
{
  "nvidiaGpuCount": 2,
  "nvidia0UtilizationGpu": 85,
  "nvidia0UtilizationGpuT": 1735166000002,
  "nvidia0MemoryUsedMb": 8192,
  "nvidia0PowerDrawW": 245.5,
  "nvidia0ProcessesJson": "[{\"pid\":1234,\"name\":\"python\",\"usedMemoryMb\":4096}]",
  "nvidia1UtilizationGpu": 72,
  "nvidia1MemoryUsedMb": 6144
}
```

---

## vLLM Inference Metrics

Prefix: `vllm`. Disable with `-no-vllm`. Configure endpoint via `VLLM_METRICS_URL` environment variable (default: `http://localhost:8000/metrics`).

### Availability & Timing

| Key               | Type  | Unit        | Source         | Description                    |
|-------------------|-------|-------------|----------------|--------------------------------|
| `vllmAvailable`   | bool  | -           | HTTP response  | Whether vLLM endpoint is reachable |
| `vllmTimestamp`   | int64 | nanoseconds | Scrape time    | When metrics were collected    |

### System State

| Key                    | Type    | Unit  | Source                      | Description        |
|------------------------|---------|-------|-----------------------------|--------------------|
| `vllmRequestsRunning`  | float64 | count | `vllm:num_requests_running` | Active requests    |
| `vllmRequestsWaiting`  | float64 | count | `vllm:num_requests_waiting` | Queued requests    |
| `vllmEngineSleepState` | float64 | -     | `vllm:engine_sleep_state`   | Engine sleep state |
| `vllmPreemptionsTotal` | float64 | count | `vllm:num_preemptions`      | Preempted requests |

### Cache Metrics

| Key                       | Type    | Unit  | Source                      | Description          |
|---------------------------|---------|-------|-----------------------------|----------------------|
| `vllmKvCacheUsagePercent` | float64 | ratio | `vllm:kv_cache_usage_perc`  | KV-cache utilization |
| `vllmPrefixCacheHits`     | float64 | count | `vllm:prefix_cache_hits`    | Prefix cache hits    |
| `vllmPrefixCacheQueries`  | float64 | count | `vllm:prefix_cache_queries` | Prefix cache queries |

### Throughput

| Key                          | Type    | Unit  | Source                    | Description             |
|------------------------------|---------|-------|---------------------------|-------------------------|
| `vllmRequestsFinishedTotal`  | float64 | count | `vllm:request_success`    | Completed requests      |
| `vllmRequestsCorruptedTotal` | float64 | count | `vllm:corrupted_requests` | Failed requests         |
| `vllmTokensPromptTotal`      | float64 | count | `vllm:prompt_tokens`      | Prompt tokens processed |
| `vllmTokensGenerationTotal`  | float64 | count | `vllm:generation_tokens`  | Generated tokens        |

### Latency (Sums and Counts)

| Key                         | Type    | Unit    | Source                                      | Description        |
|-----------------------------|---------|---------|---------------------------------------------|--------------------|
| `vllmLatencyTtftSum`        | float64 | seconds | `vllm:time_to_first_token_seconds_sum`      | TTFT sum           |
| `vllmLatencyTtftCount`      | float64 | count   | `vllm:time_to_first_token_seconds_count`    | TTFT count         |
| `vllmLatencyE2eSum`         | float64 | seconds | `vllm:e2e_request_latency_seconds_sum`      | E2E latency sum    |
| `vllmLatencyE2eCount`       | float64 | count   | `vllm:e2e_request_latency_seconds_count`    | E2E count          |
| `vllmLatencyQueueSum`       | float64 | seconds | `vllm:request_queue_time_seconds_sum`       | Queue time sum     |
| `vllmLatencyQueueCount`     | float64 | count   | `vllm:request_queue_time_seconds_count`     | Queue count        |
| `vllmLatencyInferenceSum`   | float64 | seconds | `vllm:request_inference_time_seconds_sum`   | Inference time sum |
| `vllmLatencyInferenceCount` | float64 | count   | `vllm:request_inference_time_seconds_count` | Inference count    |
| `vllmLatencyPrefillSum`     | float64 | seconds | `vllm:request_prefill_time_seconds_sum`     | Prefill time sum   |
| `vllmLatencyPrefillCount`   | float64 | count   | `vllm:request_prefill_time_seconds_count`   | Prefill count      |
| `vllmLatencyDecodeSum`      | float64 | seconds | `vllm:request_decode_time_seconds_sum`      | Decode time sum    |
| `vllmLatencyDecodeCount`    | float64 | count   | `vllm:request_decode_time_seconds_count`    | Decode count       |

### Histograms

Histogram bucket data is stored as JSON in `vllmHistogramsJson`:

```json
{
  "vllmHistogramsJson": "{\"latencyTtft\":{\"0.001\":0,\"0.005\":12,\"0.01\":45,\"inf\":300},\"latencyE2e\":{...},...}"
}
```

Available histogram fields:
- `latencyTtft`, `latencyE2e`, `latencyQueue`, `latencyInference`, `latencyPrefill`, `latencyDecode`, `latencyInterToken`
- `reqSizePromptTokens`, `reqSizeGenerationTokens`
- `tokensPerStep`, `reqParamsMaxTokens`, `reqParamsN`

---

## Environment Variables

| Variable           | Default                         | Description                      |
|--------------------|---------------------------------|----------------------------------|
| `VLLM_METRICS_URL` | `http://localhost:8000/metrics` | vLLM Prometheus metrics endpoint |
