# Profiler Metrics Reference

This document provides a comprehensive reference for all metrics collected by the Go profiler.

## Data Structure

All dynamic metrics follow this format:

```json
{
  "metricName": {
    "value": <value>,
    "time": <unix_timestamp_ms>
  }
}
```

- `value`: The metric value (type varies by metric)
- `time`: Collection timestamp in milliseconds since Unix epoch

---

## Table of Contents

1. [Static System Information](#static-system-information)
2. [VM-Level CPU Metrics](#vm-level-cpu-metrics)
3. [VM-Level Memory Metrics](#vm-level-memory-metrics)
4. [VM-Level Disk I/O Metrics](#vm-level-disk-io-metrics)
5. [VM-Level Network Metrics](#vm-level-network-metrics)
6. [Container-Level Metrics](#container-level-metrics)
7. [Process-Level Metrics](#process-level-metrics)
8. [NVIDIA GPU Metrics](#nvidia-gpu-metrics)
9. [vLLM Inference Metrics](#vllm-inference-metrics)

---

## Static System Information

Static information is collected once at profiler startup and saved to `static_{uuid}.json`.

### Root Level

| Key    | Type   | Source                                                        | Description                |
|--------|--------|---------------------------------------------------------------|----------------------------|
| `uuid` | string | Generated UUID v4                                             | Unique session identifier. |
| `vId`  | string | `/sys/class/dmi/id/product_uuid`. Fallback: `/etc/machine-id` | VM/instance identifier.    |

### Host Information (`host` object)

| Key                 | Type             | Source                                                                                        | Description                                   |
|---------------------|------------------|-----------------------------------------------------------------------------------------------|-----------------------------------------------|
| `hostname`          | string           | `syscall.Uname().Nodename`                                                                    | System hostname.                              |
| `bootTime`          | int64            | `/proc/stat` `btime` field                                                                    | System boot time (Unix timestamp seconds).    |
| `vNumProcessors`    | int              | Go `runtime.NumCPU()`                                                                         | Number of logical CPUs/cores.                 |
| `vCpuType`          | string           | `/proc/cpuinfo` `model name` field                                                            | Processor model name string.                  |
| `vCpuCache`         | map[string]int64 | `/sys/devices/system/cpu/cpu*/cache/index*/` files: `level`, `type`, `size`, `shared_cpu_map` | Cache sizes in bytes. Keys: L1d, L1i, L2, L3. |
| `vKernelInfo`       | string           | `syscall.Uname()` (all fields joined)                                                         | Full kernel version string.                   |
| `vMemoryTotalBytes` | int64            | `/proc/meminfo` `MemTotal` × 1024                                                             | Total physical RAM in bytes.                  |
| `vSwapTotalBytes`   | int64            | `/proc/meminfo` `SwapTotal` × 1024                                                            | Total swap space in bytes.                    |

### NVIDIA Static Information

| Key                   | Type   | Source (NVML API)                                                           | Description                   |
|-----------------------|--------|-----------------------------------------------------------------------------|-------------------------------|
| `nvidiaDriverVersion` | string | `nvmlSystemGetDriverVersion()`. Fallback: `/proc/driver/nvidia/version`     | NVIDIA driver version string. |
| `nvidiaCudaVersion`   | string | `nvmlSystemGetCudaDriverVersion()`. Fallback: `/usr/local/cuda/version.txt` | CUDA driver version.          |

### NVIDIA GPU Static Info (`nvidia` array)

Each GPU in the array contains:

| Key                      | Type   | Source (NVML API)                                | Description                    |
|--------------------------|--------|--------------------------------------------------|--------------------------------|
| `nvidiaName`             | string | `nvmlDeviceGetName()`                            | GPU model name.                |
| `nvidiaUuid`             | string | `nvmlDeviceGetUUID()`                            | Unique GPU identifier.         |
| `nvidiaTotalMemoryMb`    | uint64 | `nvmlDeviceGetMemoryInfo().total` / 1024 / 1024  | Total frame buffer in MB.      |
| `nvidiaPciBusId`         | string | `nvmlDeviceGetPciInfo().busId`                   | PCI bus address.               |
| `nvidiaMaxGraphicsClock` | uint32 | `nvmlDeviceGetMaxClockInfo(NVML_CLOCK_GRAPHICS)` | Maximum graphics clock in MHz. |
| `nvidiaMaxSmClock`       | uint32 | `nvmlDeviceGetMaxClockInfo(NVML_CLOCK_SM)`       | Maximum SM clock in MHz.       |
| `nvidiaMaxMemClock`      | uint32 | `nvmlDeviceGetMaxClockInfo(NVML_CLOCK_MEM)`      | Maximum memory clock in MHz.   |

---

## VM-Level CPU Metrics

Prefix: `v` (VM-level)

| Key                   | Type    | Unit         | Source                                                                                                                         | Description                                                          |
|-----------------------|---------|--------------|--------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------|
| `vCpuTime`            | int64   | centiseconds | `/proc/stat` cpu line: sum of `user` + `system` fields                                                                         | Total CPU time consumed (user + kernel mode). 1 cs = 10ms.           |
| `vCpuTimeUserMode`    | int64   | centiseconds | `/proc/stat` cpu line, column 1 (`user`)                                                                                       | Time spent executing user-space processes.                           |
| `vCpuTimeKernelMode`  | int64   | centiseconds | `/proc/stat` cpu line, column 3 (`system`)                                                                                     | Time spent executing kernel code on behalf of processes.             |
| `vCpuIdleTime`        | int64   | centiseconds | `/proc/stat` cpu line, column 4 (`idle`)                                                                                       | Time CPU spent in idle state (no work).                              |
| `vCpuTimeIOWait`      | int64   | centiseconds | `/proc/stat` cpu line, column 5 (`iowait`)                                                                                     | Time CPU spent waiting for I/O operations to complete.               |
| `vCpuTimeIntSrvc`     | int64   | centiseconds | `/proc/stat` cpu line, column 6 (`irq`)                                                                                        | Time spent servicing hardware interrupts.                            |
| `vCpuTimeSoftIntSrvc` | int64   | centiseconds | `/proc/stat` cpu line, column 7 (`softirq`)                                                                                    | Time spent servicing software interrupts (deferred work).            |
| `vCpuNice`            | int64   | centiseconds | `/proc/stat` cpu line, column 2 (`nice`)                                                                                       | Time spent running niced (low priority) user processes.              |
| `vCpuSteal`           | int64   | centiseconds | `/proc/stat` cpu line, column 8 (`steal`)                                                                                      | Time stolen by hypervisor for other VMs (virtualized environments).  |
| `vCpuContextSwitches` | int64   | count        | `/proc/stat` `ctxt` line                                                                                                       | Total context switches across all CPUs since boot.                   |
| `vLoadAvg`            | float64 | ratio        | `/proc/loadavg` column 1                                                                                                       | 1-minute system load average (runnable + uninterruptible processes). |
| `vCpuMhz`             | float64 | MHz          | `/sys/devices/system/cpu/*/cpufreq/scaling_cur_freq` (average of all cores / 1000). Fallback: `/proc/cpuinfo` `cpu MHz` field. | Current average CPU frequency across all cores.                      |

### Notes on CPU Metrics

- All time values are cumulative since system boot
- Values are converted from kernel jiffies (1 jiffy = 1/100 second = 10ms) to centiseconds
- To calculate CPU utilization percentage, compute delta between two samples and divide by elapsed time

---

## VM-Level Memory Metrics

Prefix: `v` (VM-level)

| Key               | Type    | Unit            | Source                                                                                   | Description                                              |
|-------------------|---------|-----------------|------------------------------------------------------------------------------------------|----------------------------------------------------------|
| `vMemoryTotal`    | int64   | bytes           | `/proc/meminfo` `MemTotal` field × 1024                                                  | Total physical RAM installed.                            |
| `vMemoryFree`     | int64   | bytes           | `/proc/meminfo` `MemAvailable` field × 1024. Fallback: `MemFree` + `Buffers` + `Cached`. | Available memory (includes reclaimable buffers/cache).   |
| `vMemoryUsed`     | int64   | bytes           | Derived: `MemTotal` - `MemFree` - `Buffers` - `Cached`                                   | Actively used memory.                                    |
| `vMemoryBuffers`  | int64   | bytes           | `/proc/meminfo` `Buffers` field × 1024                                                   | Memory used by kernel buffers for block device I/O.      |
| `vMemoryCached`   | int64   | bytes           | `/proc/meminfo` `Cached` + `SReclaimable` fields × 1024                                  | Memory used by page cache (file data cache).             |
| `vMemoryPercent`  | float64 | percent (0-100) | Derived: `(MemTotal - MemAvailable) / MemTotal × 100`                                    | Percentage of total RAM currently in use.                |
| `vSwapTotal`      | int64   | bytes           | `/proc/meminfo` `SwapTotal` field × 1024                                                 | Total swap space available.                              |
| `vSwapFree`       | int64   | bytes           | `/proc/meminfo` `SwapFree` field × 1024                                                  | Unused swap space.                                       |
| `vSwapUsed`       | int64   | bytes           | Derived: `SwapTotal` - `SwapFree`                                                        | Currently used swap space.                               |
| `vPgFault`        | int64   | count           | `/proc/vmstat` `pgfault` field                                                           | Total minor page faults (page in memory but not mapped). |
| `vMajorPageFault` | int64   | count           | `/proc/vmstat` `pgmajfault` field                                                        | Major page faults requiring disk I/O to resolve.         |

### Notes on Memory Metrics

- All values from `/proc/meminfo` are converted from kB to bytes (× 1024)
- `vMemoryFree` uses `MemAvailable` when present (kernel 3.14+), which provides a more accurate estimate of available
  memory
- Page faults are cumulative since boot

---

## VM-Level Disk I/O Metrics

Prefix: `v` (VM-level)

| Key                     | Type  | Unit         | Source                                                    | Description                                         |
|-------------------------|-------|--------------|-----------------------------------------------------------|-----------------------------------------------------|
| `vDiskSectorReads`      | int64 | sectors      | `/proc/diskstats` column 5 (summed across physical disks) | Total sectors read. 1 sector = 512 bytes typically. |
| `vDiskSectorWrites`     | int64 | sectors      | `/proc/diskstats` column 9 (summed across physical disks) | Total sectors written.                              |
| `vDiskReadBytes`        | int64 | bytes        | Derived: `vDiskSectorReads` × 512                         | Total bytes read from all physical disks.           |
| `vDiskWriteBytes`       | int64 | bytes        | Derived: `vDiskSectorWrites` × 512                        | Total bytes written to all physical disks.          |
| `vDiskSuccessfulReads`  | int64 | count        | `/proc/diskstats` column 3                                | Number of read operations completed successfully.   |
| `vDiskSuccessfulWrites` | int64 | count        | `/proc/diskstats` column 7                                | Number of write operations completed successfully.  |
| `vDiskMergedReads`      | int64 | count        | `/proc/diskstats` column 4                                | Adjacent read requests merged for efficiency.       |
| `vDiskMergedWrites`     | int64 | count        | `/proc/diskstats` column 8                                | Adjacent write requests merged for efficiency.      |
| `vDiskReadTime`         | int64 | milliseconds | `/proc/diskstats` column 6                                | Total time spent on read operations.                |
| `vDiskWriteTime`        | int64 | milliseconds | `/proc/diskstats` column 10                               | Total time spent on write operations.               |
| `vDiskIOInProgress`     | int64 | count        | `/proc/diskstats` column 11                               | Number of I/O operations currently in flight.       |
| `vDiskIOTime`           | int64 | milliseconds | `/proc/diskstats` column 12                               | Total time spent doing I/O (wall clock).            |
| `vDiskWeightedIOTime`   | int64 | milliseconds | `/proc/diskstats` column 13                               | I/O time weighted by number of pending operations.  |

### Disk Device Filtering

Only physical disk devices are included. The following device name patterns are matched:

- `sd[a-z]+` - SCSI/SATA disks (e.g., sda, sdb)
- `hd[a-z]+` - IDE disks (e.g., hda, hdb)
- `vd[a-z]+` - VirtIO disks (e.g., vda, vdb)
- `xvd[a-z]+` - Xen virtual disks (e.g., xvda)
- `nvme[0-9]+n[0-9]+` - NVMe disks (e.g., nvme0n1)
- `mmcblk[0-9]+` - MMC/SD cards (e.g., mmcblk0)

Partitions (e.g., sda1), loopback devices, and RAM disks are excluded.

---

## VM-Level Network Metrics

Prefix: `v` (VM-level)

| Key                    | Type  | Unit  | Source                                                                               | Description                                     |
|------------------------|-------|-------|--------------------------------------------------------------------------------------|-------------------------------------------------|
| `vNetworkBytesRecvd`   | int64 | bytes | `/proc/net/dev` column 0 (Receive bytes), summed across all non-loopback interfaces  | Cumulative bytes received on all interfaces.    |
| `vNetworkBytesSent`    | int64 | bytes | `/proc/net/dev` column 8 (Transmit bytes), summed across all non-loopback interfaces | Cumulative bytes transmitted on all interfaces. |
| `vNetworkPacketsRecvd` | int64 | count | `/proc/net/dev` column 1                                                             | Total packets received.                         |
| `vNetworkPacketsSent`  | int64 | count | `/proc/net/dev` column 9                                                             | Total packets transmitted.                      |
| `vNetworkErrorsRecvd`  | int64 | count | `/proc/net/dev` column 2                                                             | Receive errors (CRC, framing, etc.).            |
| `vNetworkErrorsSent`   | int64 | count | `/proc/net/dev` column 10                                                            | Transmit errors.                                |
| `vNetworkDropsRecvd`   | int64 | count | `/proc/net/dev` column 3                                                             | Packets dropped on receive (buffer overflow).   |
| `vNetworkDropsSent`    | int64 | count | `/proc/net/dev` column 11                                                            | Packets dropped on transmit.                    |

### Notes on Network Metrics

- Loopback interface (`lo`) is excluded from aggregation
- All counters are cumulative since interface initialization
- Values are aggregated across all network interfaces

---

## Container-Level Metrics

Prefix: `c` (Container-level)

These metrics are collected from Linux cgroups. The collector auto-detects cgroup v1 or v2.

| Key                  | Type   | Unit         | Source (v1)                                                                         | Source (v2)                                     | Description                                                 |
|----------------------|--------|--------------|-------------------------------------------------------------------------------------|-------------------------------------------------|-------------------------------------------------------------|
| `cId`                | string | -            | `/proc/self/cgroup` (parses `/docker/` or `/kubepods/` paths). Fallback: `hostname` | Same                                            | Container ID (Docker/Kubernetes short ID or "unavailable"). |
| `cCgroupVersion`     | int64  | -            | Hardcoded `1`                                                                       | Hardcoded `2`                                   | Detected cgroup version.                                    |
| `cCpuTime`           | int64  | nanoseconds  | `/sys/fs/cgroup/cpuacct/cpuacct.usage`                                              | `/sys/fs/cgroup/cpu.stat` `usage_usec` × 1000   | Total CPU time consumed by all tasks in cgroup.             |
| `cCpuTimeUserMode`   | int64  | centiseconds | `/sys/fs/cgroup/cpuacct/cpuacct.stat` `user` field × 100                            | `/sys/fs/cgroup/cpu.stat` `user_usec` / 10000   | CPU time in user mode.                                      |
| `cCpuTimeKernelMode` | int64  | centiseconds | `/sys/fs/cgroup/cpuacct/cpuacct.stat` `system` field × 100                          | `/sys/fs/cgroup/cpu.stat` `system_usec` / 10000 | CPU time in kernel mode.                                    |
| `cNumProcessors`     | int64  | count        | Go `runtime.NumCPU()`                                                               | Same                                            | Number of CPU processors available to container.            |
| `cMemoryUsed`        | int64  | bytes        | `/sys/fs/cgroup/memory/memory.usage_in_bytes`                                       | `/sys/fs/cgroup/memory.current`                 | Current memory usage by processes in cgroup.                |
| `cMemoryMaxUsed`     | int64  | bytes        | `/sys/fs/cgroup/memory/memory.max_usage_in_bytes`                                   | `/sys/fs/cgroup/memory.peak`                    | Peak memory usage (high watermark).                         |
| `cDiskReadBytes`     | int64  | bytes        | `/sys/fs/cgroup/blkio/blkio.throttle.io_service_bytes` sum of 'Read' operations     | `/sys/fs/cgroup/io.stat` sum of `rbytes`        | Bytes read by cgroup from block devices.                    |
| `cDiskWriteBytes`    | int64  | bytes        | `/sys/fs/cgroup/blkio/blkio.throttle.io_service_bytes` sum of 'Write' operations    | `/sys/fs/cgroup/io.stat` sum of `wbytes`        | Bytes written by cgroup to block devices.                   |
| `cNetworkBytesRecvd` | int64  | bytes        | `/proc/net/dev` (from container's network namespace)                                | Same                                            | Bytes received.                                             |
| `cNetworkBytesSent`  | int64  | bytes        | `/proc/net/dev` (from container's network namespace)                                | Same                                            | Bytes transmitted.                                          |

### Per-CPU Metrics (cgroup v1 only)

| Key Pattern   | Type  | Unit        | Source                                        | Description                            |
|---------------|-------|-------------|-----------------------------------------------|----------------------------------------|
| `cCpu{i}Time` | int64 | nanoseconds | `/sys/fs/cgroup/cpuacct/cpuacct.usage_percpu` | CPU time on processor `i` (0-indexed). |

---

## Process-Level Metrics

Prefix: `p` (Process-level)

These metrics are collected per-process by iterating `/proc/[pid]/*` directories. Enable with `-p` flag.

| Key                            | Type   | Unit         | Source                                                                                        | Description                                    |
|--------------------------------|--------|--------------|-----------------------------------------------------------------------------------------------|------------------------------------------------|
| `pId`                          | int64  | -            | Directory name in `/proc/`                                                                    | Process ID (PID).                              |
| `pName`                        | string | -            | `/proc/[pid]/status` `Name` field. Fallback: `/proc/[pid]/stat` (content between parentheses) | Executable name.                               |
| `pCmdline`                     | string | -            | `/proc/[pid]/cmdline` (null bytes replaced with spaces)                                       | Full command line with arguments.              |
| `pNumThreads`                  | int64  | count        | `/proc/[pid]/stat` field 19 (0-indexed: 17 after state)                                       | Number of threads in this process.             |
| `pCpuTimeUserMode`             | int64  | centiseconds | `/proc/[pid]/stat` field 14 (0-indexed: 11 after state) × 100                                 | CPU time scheduled in user mode.               |
| `pCpuTimeKernelMode`           | int64  | centiseconds | `/proc/[pid]/stat` field 15 (0-indexed: 12 after state) × 100                                 | CPU time scheduled in kernel mode.             |
| `pChildrenUserMode`            | int64  | centiseconds | `/proc/[pid]/stat` field 16 (0-indexed: 13 after state) × 100                                 | Cumulative user time of waited-for children.   |
| `pChildrenKernelMode`          | int64  | centiseconds | `/proc/[pid]/stat` field 17 (0-indexed: 14 after state) × 100                                 | Cumulative kernel time of waited-for children. |
| `pVoluntaryContextSwitches`    | int64  | count        | `/proc/[pid]/status` `voluntary_ctxt_switches` field                                          | Context switches due to blocking (I/O, sleep). |
| `pNonvoluntaryContextSwitches` | int64  | count        | `/proc/[pid]/status` `nonvoluntary_ctxt_switches` field                                       | Context switches due to preemption.            |
| `pBlockIODelays`               | int64  | centiseconds | `/proc/[pid]/stat` field 42 (0-indexed: 39 after state) × 100                                 | Aggregated time waiting for block I/O.         |
| `pVirtualMemoryBytes`          | int64  | bytes        | `/proc/[pid]/stat` field 23 (0-indexed: 20 after state)                                       | Total virtual address space size.              |
| `pResidentSetSize`             | int64  | bytes        | `/proc/[pid]/statm` field 1 (resident pages) × page size (typically 4096)                     | Physical memory currently mapped (RSS).        |

### Notes on Process Metrics

- Process names containing spaces or special characters are handled correctly by finding the last `)` in
  `/proc/[pid]/stat`
- All CPU time values are converted from kernel jiffies to centiseconds
- `pNumProcesses` field in output contains total count of processes collected

---

## NVIDIA GPU Metrics

Prefix: `nvidia`

These metrics are collected via NVML (NVIDIA Management Library) or nvidia-smi fallback.

| Key                      | Type     | Unit            | Source (NVML API)                                                               | Description                                            |
|--------------------------|----------|-----------------|---------------------------------------------------------------------------------|--------------------------------------------------------|
| `nvidiaGpuIndex`         | int64    | -               | Loop index from `nvmlDeviceGetCount()`                                          | Zero-based GPU index.                                  |
| `nvidiaUtilizationGpu`   | int64    | percent (0-100) | `nvmlDeviceGetUtilizationRates().gpu`                                           | Percent of time GPU was executing kernels.             |
| `nvidiaUtilizationMem`   | int64    | percent (0-100) | `nvmlDeviceGetUtilizationRates().memory`                                        | Percent of time memory was being read/written.         |
| `nvidiaMemoryTotalMb`    | int64    | megabytes       | `nvmlDeviceGetMemoryInfo().total` / 1024 / 1024                                 | Total installed frame buffer memory.                   |
| `nvidiaMemoryUsedMb`     | int64    | megabytes       | `nvmlDeviceGetMemoryInfo().used` / 1024 / 1024                                  | Currently allocated frame buffer memory.               |
| `nvidiaMemoryFreeMb`     | int64    | megabytes       | `nvmlDeviceGetMemoryInfo().free` / 1024 / 1024                                  | Available frame buffer memory.                         |
| `nvidiaBar1UsedMb`       | int64    | megabytes       | `nvmlDeviceGetBAR1MemoryInfo().bar1Used` / 1024 / 1024                          | Used PCIe BAR1 memory.                                 |
| `nvidiaBar1FreeMb`       | int64    | megabytes       | `nvmlDeviceGetBAR1MemoryInfo().bar1Free` / 1024 / 1024                          | Free PCIe BAR1 memory.                                 |
| `nvidiaTemperatureC`     | int64    | degrees Celsius | `nvmlDeviceGetTemperature(NVML_TEMPERATURE_GPU)`                                | GPU core temperature.                                  |
| `nvidiaFanSpeed`         | int64    | percent (0-100) | `nvmlDeviceGetFanSpeed()`                                                       | Fan speed percentage.                                  |
| `nvidiaPowerDrawW`       | float64  | watts           | `nvmlDeviceGetPowerUsage()` / 1000                                              | Current power consumption.                             |
| `nvidiaPowerLimitW`      | float64  | watts           | `nvmlDeviceGetEnforcedPowerLimit()` / 1000                                      | Enforced power limit.                                  |
| `nvidiaClockGraphicsMhz` | int64    | MHz             | `nvmlDeviceGetClockInfo(NVML_CLOCK_GRAPHICS)`                                   | Current shader/graphics engine frequency.              |
| `nvidiaClockSmMhz`       | int64    | MHz             | `nvmlDeviceGetClockInfo(NVML_CLOCK_SM)`                                         | Current Streaming Multiprocessor frequency.            |
| `nvidiaClockMemMhz`      | int64    | MHz             | `nvmlDeviceGetClockInfo(NVML_CLOCK_MEM)`                                        | Current memory bus frequency.                          |
| `nvidiaPcieTxKbps`       | int64    | KB/s            | `nvmlDeviceGetPcieThroughput(NVML_PCIE_UTIL_TX_BYTES)`                          | PCIe transmit throughput (Host → GPU).                 |
| `nvidiaPcieRxKbps`       | int64    | KB/s            | `nvmlDeviceGetPcieThroughput(NVML_PCIE_UTIL_RX_BYTES)`                          | PCIe receive throughput (GPU → Host).                  |
| `nvidiaPerfState`        | string   | -               | `nvmlDeviceGetPerformanceState()` formatted as "P{n}"                           | Power/performance state (P0=max performance, P12=min). |
| `nvidiaProcessCount`     | int64    | count           | Length of `nvmlDeviceGetComputeRunningProcesses()`                              | Number of processes using this GPU.                    |
| `nvidiaProcesses`        | []string | -               | Process list from `nvmlDeviceGetComputeRunningProcesses()` + `/proc/[pid]/comm` | List of processes: "PID: name (VRAM MB)".              |

### Multi-GPU Flattening

For systems with multiple GPUs, metrics are stored in an array. When flattened for CSV export:

- `nvidia_0_nvidiaUtilizationGpu_value` - GPU 0 utilization
- `nvidia_1_nvidiaUtilizationGpu_value` - GPU 1 utilization
- etc.

---

## vLLM Inference Metrics

Prefix: `vllm`

These metrics are scraped from vLLM's Prometheus endpoint (default: `http://localhost:8000/metrics`). Configure via
`VLLM_METRICS_URL` environment variable.

### System State Metrics

| Key                          | Type    | Unit  | Source (Prometheus Metric)  | Description                                      |
|------------------------------|---------|-------|-----------------------------|--------------------------------------------------|
| `vllmSystemRequestsRunning`  | float64 | count | `vllm:num_requests_running` | Requests currently being processed by engine.    |
| `vllmSystemRequestsWaiting`  | float64 | count | `vllm:num_requests_waiting` | Requests queued waiting for processing.          |
| `vllmSystemEngineSleepState` | float64 | state | `vllm:engine_sleep_state`   | 1 = Awake (active), 0 = Sleeping (idle).         |
| `vllmSystemPreemptionsTotal` | float64 | count | `vllm:num_preemptions`      | Total requests preempted due to memory pressure. |

### Cache Metrics

| Key                          | Type    | Unit            | Source (Prometheus Metric)  | Description                         |
|------------------------------|---------|-----------------|-----------------------------|-------------------------------------|
| `vllmCacheKvUsagePercent`    | float64 | ratio (0.0-1.0) | `vllm:kv_cache_usage_perc`  | GPU KV-cache utilization.           |
| `vllmCachePrefixHits`        | float64 | count           | `vllm:prefix_cache_hits`    | Successful prefix cache lookups.    |
| `vllmCachePrefixQueries`     | float64 | count           | `vllm:prefix_cache_queries` | Total prefix cache lookup attempts. |
| `vllmCacheMultimodalHits`    | float64 | count           | `vllm:mm_cache_hits`        | Successful multimodal cache hits.   |
| `vllmCacheMultimodalQueries` | float64 | count           | `vllm:mm_cache_queries`     | Total multimodal cache lookups.     |

### Throughput Metrics

| Key                          | Type    | Unit  | Source (Prometheus Metric) | Description                                 |
|------------------------------|---------|-------|----------------------------|---------------------------------------------|
| `vllmRequestsFinishedTotal`  | float64 | count | `vllm:request_success`     | Cumulative successfully completed requests. |
| `vllmRequestsCorruptedTotal` | float64 | count | `vllm:corrupted_requests`  | Cumulative failed/corrupted requests.       |
| `vllmTokensPromptTotal`      | float64 | count | `vllm:prompt_tokens`       | Total prompt (prefill) tokens processed.    |
| `vllmTokensGenerationTotal`  | float64 | count | `vllm:generation_tokens`   | Total generation (decode) tokens produced.  |

### Latency Metrics (Summary Sums)

| Key                          | Type    | Unit    | Source (Prometheus Metric)                | Description                                      |
|------------------------------|---------|---------|-------------------------------------------|--------------------------------------------------|
| `vllmLatencyTtftS_sum`       | float64 | seconds | `vllm:time_to_first_token_seconds_sum`    | Sum of time from request receipt to first token. |
| `vllmLatencyE2eS_sum`        | float64 | seconds | `vllm:e2e_request_latency_seconds_sum`    | Sum of total request latency.                    |
| `vllmLatencyQueueS_sum`      | float64 | seconds | `vllm:request_queue_time_seconds_sum`     | Sum of time spent waiting in queue.              |
| `vllmLatencyInferenceS_sum`  | float64 | seconds | `vllm:request_inference_time_seconds_sum` | Sum of time in active inference.                 |
| `vllmLatencyPrefillS_sum`    | float64 | seconds | `vllm:request_prefill_time_seconds_sum`   | Sum of time processing prompt tokens.            |
| `vllmLatencyDecodeS_sum`     | float64 | seconds | `vllm:request_decode_time_seconds_sum`    | Sum of time generating output tokens.            |
| `vllmLatencyInterTokenS_sum` | float64 | seconds | `vllm:inter_token_latency_seconds_sum`    | Sum of time between successive tokens.           |

### Histogram Metrics

Histograms are exported as JSON objects with bucket boundaries as keys and cumulative counts as values.

| Key Pattern                             | Type               | Source (Prometheus Metric)                   | Description                             |
|-----------------------------------------|--------------------|----------------------------------------------|-----------------------------------------|
| `vllmTokensPerStepHistogram_histogram`  | map[string]float64 | `vllm:iteration_tokens_total_bucket`         | Distribution of tokens per engine step. |
| `vllmReqSizePromptTokens_histogram`     | map[string]float64 | `vllm:request_prompt_tokens_bucket`          | Distribution of prompt lengths.         |
| `vllmReqSizeGenerationTokens_histogram` | map[string]float64 | `vllm:request_generation_tokens_bucket`      | Distribution of generation lengths.     |
| `vllmReqParamsMaxTokens_histogram`      | map[string]float64 | `vllm:request_params_max_tokens_bucket`      | Distribution of `max_tokens` parameter. |
| `vllmReqParamsN_histogram`              | map[string]float64 | `vllm:request_params_n_bucket`               | Distribution of `n` parameter.          |
| `vllmLatencyTtftS_histogram`            | map[string]float64 | `vllm:time_to_first_token_seconds_bucket`    | TTFT histogram buckets.                 |
| `vllmLatencyE2eS_histogram`             | map[string]float64 | `vllm:e2e_request_latency_seconds_bucket`    | E2E latency histogram buckets.          |
| `vllmLatencyQueueS_histogram`           | map[string]float64 | `vllm:request_queue_time_seconds_bucket`     | Queue time histogram buckets.           |
| `vllmLatencyInferenceS_histogram`       | map[string]float64 | `vllm:request_inference_time_seconds_bucket` | Inference time histogram buckets.       |
| `vllmLatencyPrefillS_histogram`         | map[string]float64 | `vllm:request_prefill_time_seconds_bucket`   | Prefill time histogram buckets.         |
| `vllmLatencyDecodeS_histogram`          | map[string]float64 | `vllm:request_decode_time_seconds_bucket`    | Decode time histogram buckets.          |
| `vllmLatencyInterTokenS_histogram`      | map[string]float64 | `vllm:inter_token_latency_seconds_bucket`    | Inter-token latency histogram buckets.  |

### Histogram Format Example

```json
{
  "vllmLatencyTtftS_histogram": {
    "value": {
      "0.001": 0,
      "0.005": 12,
      "0.01": 45,
      "0.025": 120,
      "0.05": 200,
      "0.1": 280,
      "inf": 300
    },
    "time": 1735166000000
  }
}
```

---


## Output Formats

### JSON Snapshot Format

Each collection interval produces a JSON file:

```json
{
  "timestamp": 1735166000000,
  "cpu": {
    "vCpuTime": {
      "value": 12345600,
      "time": 1735166000000
    },
    "vCpuTimeUserMode": {
      "value": 9876500,
      "time": 1735166000000
    }
  },
  "mem": {
    "vMemoryTotal": {
      "value": 17179869184,
      "time": 1735166000001
    },
    "vMemoryUsed": {
      "value": 8589934592,
      "time": 1735166000001
    }
  },
  "disk": {
    ...
  },
  "net": {
    ...
  },
  "container": {
    ...
  },
  "nvidia": [
    {
      ...
    },
    {
      ...
    }
  ],
  "vllm": {
    ...
  },
  "processes": [
    {
      ...
    },
    {
      ...
    }
  ]
}
```

### CSV/TSV Flattened Format

When exported to CSV/TSV, metrics are flattened:

| Column Pattern                  | Description                  |
|---------------------------------|------------------------------|
| `{metricName}_value`            | The metric value             |
| `t{metricName}`                 | The metric timestamp         |
| `nvidia_{i}_{metricName}_value` | GPU `i` metric value         |
| `nvidia_{i}_t{metricName}`      | GPU `i` metric timestamp     |
| `proc_{i}_{metricName}_value`   | Process `i` metric value     |
| `proc_{i}_t{metricName}`        | Process `i` metric timestamp |

---

## Timestamp Fields

All dynamic metrics include timestamps:

| Pattern         | Description                                       |
|-----------------|---------------------------------------------------|
| `t{metricName}` | Timestamp when this specific metric was collected |
| `timestamp`     | Overall collection cycle timestamp                |

Timestamps are in **milliseconds** since Unix epoch (January 1, 1970 00:00:00 UTC).

---

## Environment Variables

| Variable           | Default                         | Description                           |
|--------------------|---------------------------------|---------------------------------------|
| `VLLM_METRICS_URL` | `http://localhost:8000/metrics` | vLLM Prometheus metrics endpoint URL. |

---

## Collection Intervals and Rates

- Default collection interval: 1000ms (1 second)
- Configurable via `-t` flag (in milliseconds)
- vLLM scrape timeout: 500ms
- All counters are cumulative; compute deltas for rates

### Rate Calculation Example

```
CPU Utilization % = (delta_vCpuTime / delta_elapsed_time) * 100 / num_cpus
Network Throughput = delta_vNetworkBytesSent / delta_elapsed_time
Disk IOPS = delta_vDiskSuccessfulReads / delta_elapsed_time
```