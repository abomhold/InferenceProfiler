# Profiler Metrics Reference

This document provides a comprehensive reference for all metrics collected. The profiler collects resource utilization
metrics at three levels of granularity—VM/host, container, and process—enabling characterization of resource use with
increasing isolation. This multi-level approach allows identification of performance bottlenecks and attribution of
resource consumption to specific workloads.

> **Reference**: Metric definitions and profiling methodology are based on the Container Profiler research (Hoang et
> al., GigaScience, 2023), which demonstrated that fine-grained resource utilization profiling can help identify
> bottlenecks and inform optimal cloud deployment decisions.

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

| Flag           | Default             | Description                                                                            |
|----------------|---------------------|----------------------------------------------------------------------------------------|
| `-o <dir>`     | `./profiler-output` | Output directory for logs and exported data                                            |
| `-t <ms>`      | `100`               | Sampling interval in milliseconds                                                      |
| `-f <format>`  | `jsonl`             | Export format: `jsonl`, `parquet`, `csv`, `tsv`                                        |
| `--stream`     | `false`             | Stream mode: write directly to output file instead of batch processing                 |
| `--no-flatten` | `false`             | Disable flattening of nested data (GPUs, processes) to columns; stores as JSON strings |
| `--no-cleanup` | `false`             | Keep intermediary JSON snapshot files after final export (batch mode only)             |
| `--graphs`     | `true`              | Generate HTML graphs after profiling                                                   |
| `-i <file>`    | -                   | Input file for graph-only mode                                                         |
| `-g`           | `false`             | Graph-only mode (requires `-i`)                                                        |

### Collector Toggles

All collectors are enabled by default. Disabling unnecessary collectors reduces profiling overhead.

| Flag             | Default | Description                                                  |
|------------------|---------|--------------------------------------------------------------|
| `--no-cpu`       | `false` | Disable CPU metrics collection                               |
| `--no-memory`    | `false` | Disable memory metrics collection                            |
| `--no-disk`      | `false` | Disable disk I/O metrics collection                          |
| `--no-network`   | `false` | Disable network metrics collection                           |
| `--no-container` | `false` | Disable container/cgroup metrics collection                  |
| `--no-procs`     | `false` | Disable per-process metrics collection                       |
| `--no-nvidia`    | `false` | Disable all NVIDIA GPU metrics collection                    |
| `--no-gpu-procs` | `false` | Disable GPU process enumeration (still collects GPU metrics) |
| `--no-vllm`      | `false` | Disable vLLM metrics collection                              |

### Example Usage

```bash
# Minimal: just CPU/memory, TSV output, no flattening
./profiler --no-disk --no-network --no-container --no-procs --no-nvidia --no-vllm -f tsv --no-flatten

# Full GPU monitoring without process overhead, streaming to parquet
./profiler --no-procs --no-gpu-procs -f parquet --stream

# CSV with no processes, flattened (default)
./profiler --no-procs -f csv

# Profile a subprocess
./profiler -f jsonl -- python train.py --epochs 10

# Generate graphs from existing output
./profiler -g -i ./profiler-output/metrics_abc123.parquet
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

### Streaming vs Batch Mode

| Mode   | Flag       | Description                                                                          |
|--------|------------|--------------------------------------------------------------------------------------|
| Batch  | (default)  | Writes intermediate JSON files, aggregates at end; supports delta calculations       |
| Stream | `--stream` | Writes directly to final format, lower memory usage; ideal for long-running profiles |

### Flatten Modes

The `--no-flatten` flag controls how nested data (GPUs, processes) is represented:

#### Flattened Mode (default)

GPUs and processes are expanded into individual columns with indexed prefixes:

```json
{
  "timestamp": 1735166000000,
  "vCpuTime": 123456,
  "nvidiaGpuCount": 2,
  "gpuIndex": 0,
  "gpuUtilizationGpu": 85,
  "gpuMemoryUsedBytes": 8589934592,
  "processCount": 150,
  "proc0Pid": 1,
  "proc0Name": "systemd"
}
```

#### JSON String Mode (`--no-flatten`)

GPUs and processes are serialized as JSON strings:

```json
{
  "timestamp": 1735166000000,
  "vCpuTime": 123456,
  "nvidiaGpuCount": 2,
  "nvidiaGpusJson": "[{\"index\":0,\"utilizationGpu\":85,...}]",
  "processCount": 150,
  "processesJson": "[{\"pId\":1,\"pName\":\"systemd\",...}]"
}
```

### Timestamp Convention

All dynamic metrics include per-field timestamps with a `T` suffix, enabling precise calculation of the time between
samples for rate calculations:

| Field           | Description                                                       |
|-----------------|-------------------------------------------------------------------|
| `timestamp`     | Overall collection cycle timestamp (nanoseconds since Unix epoch) |
| `{metricName}`  | The metric value                                                  |
| `{metricName}T` | Timestamp when this specific metric was collected (nanoseconds)   |

### Output Files

Each profiling session produces:

| File                             | Description                                               |
|----------------------------------|-----------------------------------------------------------|
| `static_{uuid}.json`             | Static system information (collected once at startup)     |
| `snapshots/snapshot_NNNNNN.json` | Individual snapshot files (batch mode only, intermediate) |
| `metrics_{uuid}.{ext}`           | Final aggregated output (jsonl/parquet/csv/tsv)           |
| `report_{uuid}.html`             | Interactive HTML graphs (if `--graphs` enabled)           |

---

## Static System Information

Static information is collected once at profiler startup and saved to `static_{uuid}.json`. These metrics typically
describe hardware characteristics and system configuration.

### Session & Host Identification

| Key         | Type   | Source                                                | Description                               |
|-------------|--------|-------------------------------------------------------|-------------------------------------------|
| `uuid`      | string | Generated UUID v4                                     | Unique session identifier                 |
| `vId`       | string | `/sys/class/dmi/id/product_uuid` or `/etc/machine-id` | VM/instance identifier for cloud tracking |
| `vHostname` | string | `syscall.Uname().Nodename`                            | System hostname                           |
| `vBootTime` | int64  | `/proc/stat` `btime`                                  | System boot time (Unix timestamp seconds) |

### CPU Static Info

| Key              | Type   | Source                                | Description                                         |
|------------------|--------|---------------------------------------|-----------------------------------------------------|
| `vNumProcessors` | int    | `runtime.NumCPU()`                    | Number of logical CPUs/cores available              |
| `vCpuType`       | string | `/proc/cpuinfo` `model name`          | Processor model name                                |
| `vCpuCache`      | string | `/sys/devices/system/cpu/cpu*/cache/` | Cache sizes (e.g., "L1d:32K L1i:32K L2:256K L3:8M") |
| `vKernelInfo`    | string | `syscall.Uname()`                     | Full kernel version string                          |

### Time Synchronization

| Key                    | Type    | Source               | Description                          |
|------------------------|---------|----------------------|--------------------------------------|
| `vTimeSynced`          | bool    | `adjtimex()` syscall | Whether kernel clock is synchronized |
| `vTimeOffsetSeconds`   | float64 | `adjtimex()` syscall | Time offset from reference (seconds) |
| `vTimeMaxErrorSeconds` | float64 | `adjtimex()` syscall | Maximum time error (seconds)         |

### Memory Static Info

| Key                 | Type  | Unit  | Source                             | Description                 |
|---------------------|-------|-------|------------------------------------|-----------------------------|
| `vMemoryTotalBytes` | int64 | bytes | `/proc/meminfo` `MemTotal` × 1024  | Total physical RAM in bytes |
| `vSwapTotalBytes`   | int64 | bytes | `/proc/meminfo` `SwapTotal` × 1024 | Total swap space in bytes   |

### Container Static Info

| Key              | Type   | Source              | Description                        |
|------------------|--------|---------------------|------------------------------------|
| `cId`            | string | `/proc/self/cgroup` | Container ID or hostname           |
| `cNumProcessors` | int64  | `runtime.NumCPU()`  | Number of CPUs available to cgroup |
| `cCgroupVersion` | int64  | Auto-detected       | Cgroup version (1 or 2)            |

### Network Static Info

| Key                 | Type   | Description                             |
|---------------------|--------|-----------------------------------------|
| `networkInterfaces` | string | JSON array of network interface objects |

Each network interface object:

| Field       | Type   | Description                       |
|-------------|--------|-----------------------------------|
| `name`      | string | Interface name (e.g., "eth0")     |
| `mac`       | string | MAC address                       |
| `state`     | string | Operational state (up/down)       |
| `mtu`       | int64  | Maximum transmission unit         |
| `speedMbps` | int64  | Link speed in Mbps (if available) |

### Disk Static Info

| Key     | Type   | Description                |
|---------|--------|----------------------------|
| `disks` | string | JSON array of disk objects |

Each disk object:

| Field       | Type   | Description                          |
|-------------|--------|--------------------------------------|
| `name`      | string | Device name (e.g., "sda", "nvme0n1") |
| `model`     | string | Disk model                           |
| `vendor`    | string | Disk vendor                          |
| `sizeBytes` | int64  | Total size in bytes                  |

### NVIDIA Static Info

| Key                   | Type   | Source                             | Description                   |
|-----------------------|--------|------------------------------------|-------------------------------|
| `nvidiaDriverVersion` | string | `nvmlSystemGetDriverVersion()`     | NVIDIA driver version         |
| `nvidiaCudaVersion`   | string | `nvmlSystemGetCudaDriverVersion()` | CUDA driver version           |
| `nvmlVersion`         | string | `nvmlSystemGetNVMLVersion()`       | NVML library version          |
| `nvidiaGpuCount`      | int    | Device enumeration                 | Number of GPUs                |
| `nvidiaGpus`          | string | NVML device enumeration            | JSON array of GPU static info |

Each GPU object in `nvidiaGpus`:

| Field                 | Type   | Description                                    |
|-----------------------|--------|------------------------------------------------|
| `index`               | int    | Zero-based GPU index                           |
| `name`                | string | GPU model name                                 |
| `uuid`                | string | Unique GPU identifier                          |
| `serial`              | string | Serial number                                  |
| `boardPartNumber`     | string | Board part number                              |
| `brand`               | string | Brand (GeForce, Tesla, Quadro, etc.)           |
| `architecture`        | string | Architecture (Ampere, Hopper, Ada, etc.)       |
| `cudaCapabilityMajor` | int    | CUDA compute capability major                  |
| `cudaCapabilityMinor` | int    | CUDA compute capability minor                  |
| `memoryTotalBytes`    | int64  | Total frame buffer in bytes                    |
| `bar1TotalBytes`      | int64  | Total BAR1 memory in bytes                     |
| `memoryBusWidthBits`  | int    | Memory bus width in bits                       |
| `numCores`            | int    | Number of CUDA cores                           |
| `maxClockGraphicsMhz` | int    | Max graphics clock (MHz)                       |
| `maxClockMemoryMhz`   | int    | Max memory clock (MHz)                         |
| `maxClockSmMhz`       | int    | Max SM clock (MHz)                             |
| `maxClockVideoMhz`    | int    | Max video clock (MHz)                          |
| `pciBusId`            | string | PCI bus ID                                     |
| `pciDeviceId`         | uint32 | PCI device ID                                  |
| `pciSubsystemId`      | uint32 | PCI subsystem ID                               |
| `pcieMaxLinkGen`      | int    | Max PCIe generation                            |
| `pcieMaxLinkWidth`    | int    | Max PCIe link width                            |
| `powerDefaultLimitMw` | int    | Default power limit (milliwatts)               |
| `powerMinLimitMw`     | int    | Min power limit (milliwatts)                   |
| `powerMaxLimitMw`     | int    | Max power limit (milliwatts)                   |
| `vbiosVersion`        | string | VBIOS version                                  |
| `inforomImageVersion` | string | InfoROM image version                          |
| `inforomOemVersion`   | string | InfoROM OEM version                            |
| `numFans`             | int    | Number of fans                                 |
| `tempShutdownC`       | int    | Shutdown temperature (°C)                      |
| `tempSlowdownC`       | int    | Slowdown temperature (°C)                      |
| `tempMaxOperatingC`   | int    | Max operating temperature (°C)                 |
| `tempTargetC`         | int    | Target temperature (°C)                        |
| `eccModeEnabled`      | bool   | ECC mode status                                |
| `persistenceModeOn`   | bool   | Persistence mode status                        |
| `computeMode`         | string | Compute mode (Default, ExclusiveProcess, etc.) |
| `isMultiGpuBoard`     | bool   | Multi-GPU board flag                           |
| `displayModeEnabled`  | bool   | Display mode enabled                           |
| `displayActive`       | bool   | Display currently active                       |
| `migModeEnabled`      | bool   | MIG mode status                                |
| `encoderCapacityH264` | int    | H.264 encoder capacity                         |
| `encoderCapacityHevc` | int    | HEVC encoder capacity                          |
| `encoderCapacityAv1`  | int    | AV1 encoder capacity                           |
| `nvlinkCount`         | int    | Number of NVLinks                              |

---

## VM-Level CPU Metrics

Prefix: `v` (VM-level). Disable with `--no-cpu`.

Host/VM-level resource utilization metrics are obtained from the Linux `/proc` virtual filesystem. The `/proc`
filesystem consists of dynamically generated files produced on demand by the Linux kernel, providing data regarding the
state of the system. These metrics capture total system resource utilization, including background processes external to
any container.

| Key                   | Type    | Unit         | Source                               | Description                                                                                                                                                                             |
|-----------------------|---------|--------------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `vCpuTime`            | int64   | centiseconds | `/proc/stat`                         | Total CPU time (user + kernel). Represents cumulative processor time consumed by all processes since boot.                                                                              |
| `vCpuTimeUserMode`    | int64   | centiseconds | `/proc/stat` column 1                | Time the CPU spent executing in user mode. Captures time spent executing application source code. High values during computational stages indicate CPU-bound workloads.                 |
| `vCpuTimeKernelMode`  | int64   | centiseconds | `/proc/stat` column 3                | Time the CPU spent executing in kernel mode. The kernel is typically invoked for privileged operations like disk and network I/O.                                                       |
| `vCpuIdleTime`        | int64   | centiseconds | `/proc/stat` column 4                | Time the CPU was idle. High idle time during computational stages indicates potential for performance optimization through better parallelization of code.                              |
| `vCpuTimeIOWait`      | int64   | centiseconds | `/proc/stat` column 5                | Time the CPU spent waiting for I/O (disk or network) to complete. High values indicate I/O-bound workloads where storage or network bandwidth limits performance.                       |
| `vCpuTimeIntSrvc`     | int64   | centiseconds | `/proc/stat` column 6                | Time spent handling hardware interrupts (IRQs). Hardware interrupts occur when devices need CPU attention.                                                                              |
| `vCpuTimeSoftIntSrvc` | int64   | centiseconds | `/proc/stat` column 7                | Time spent handling software interrupts (soft IRQs). Soft interrupts commonly occur with network I/O processing.                                                                        |
| `vCpuNice`            | int64   | centiseconds | `/proc/stat` column 2                | Time spent on low-priority (nice) processes. Processes with positive nice values run at reduced priority.                                                                               |
| `vCpuSteal`           | int64   | centiseconds | `/proc/stat` column 8                | Time stolen by the hypervisor for other virtual machines. Non-zero values indicate resource contention in virtualized environments; relevant for cloud VM deployments.                  |
| `vCpuContextSwitches` | int64   | count        | `/proc/stat` `ctxt`                  | Total number of context switches across all CPU cores since boot. High context switch rates can indicate thread contention, especially when more threads are used than available cores. |
| `vLoadAvg`            | float64 | ratio        | `/proc/loadavg`                      | 1-minute load average. Represents the average number of processes in the run queue. Values consistently above the number of CPUs indicate system overload.                              |
| `vCpuMhz`             | float64 | MHz          | `/sys/devices/system/cpu/*/cpufreq/` | Current average CPU frequency across all cores. Useful for detecting frequency throttling due to thermal or power limits.                                                               |

**Performance Analysis Notes:**

- **Compute-bound workloads**: High `vCpuTimeUserMode` with low `vCpuIdleTime`
- **I/O-bound workloads**: High `vCpuTimeIOWait` and `vCpuIdleTime` with low `vCpuTimeUserMode`
- **Optimization opportunities**: High `vCpuIdleTime` during expected compute-bound stages suggests room for better
  parallelization
- **Thread contention**: High `vCpuContextSwitches` rate combined with high `vCpuTimeKernelMode` indicates excessive
  context switching

**Notes:**

- All time values are cumulative since system boot
- 1 centisecond = 10 milliseconds (100 jiffies per second)
- To calculate utilization: `(delta_busy / delta_total) × 100`

---

## VM-Level Memory Metrics

Prefix: `v` (VM-level). Disable with `--no-memory`.

Memory metrics help identify whether workloads are memory-constrained or if memory allocation patterns could benefit
from optimization.

| Key                     | Type    | Unit    | Source                                    | Description                                                                                                                                    |
|-------------------------|---------|---------|-------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `vMemoryTotal`          | int64   | bytes   | `/proc/meminfo` `MemTotal`                | Total physical RAM available to the system.                                                                                                    |
| `vMemoryFree`           | int64   | bytes   | `/proc/meminfo` `MemAvailable`            | Memory available for allocation without swapping. Includes free memory plus reclaimable caches.                                                |
| `vMemoryUsed`           | int64   | bytes   | Derived                                   | Actively used memory (Total - Free - Buffers - Cached). High sustained values may indicate memory pressure.                                    |
| `vMemoryBuffers`        | int64   | bytes   | `/proc/meminfo` `Buffers`                 | Memory used by kernel buffers for block device I/O.                                                                                            |
| `vMemoryCached`         | int64   | bytes   | `/proc/meminfo` `Cached` + `SReclaimable` | Memory used for page cache and reclaimable slab memory. The kernel uses available memory for caching to improve I/O performance.               |
| `vMemoryPercent`        | float64 | percent | Derived                                   | Percentage of RAM in use: `(Total - Available) / Total × 100`.                                                                                 |
| `vMemorySwapTotal`      | int64   | bytes   | `/proc/meminfo` `SwapTotal`               | Total swap space configured.                                                                                                                   |
| `vMemorySwapFree`       | int64   | bytes   | `/proc/meminfo` `SwapFree`                | Free swap space. Decreasing values indicate memory pressure.                                                                                   |
| `vMemorySwapUsed`       | int64   | bytes   | Derived                                   | Used swap space. Non-zero values indicate the system has experienced memory pressure.                                                          |
| `vMemoryPgFault`        | int64   | count   | `/proc/vmstat` `pgfault`                  | Minor page faults (pages not in memory but not requiring disk I/O). Cumulative since boot.                                                     |
| `vMemoryMajorPageFault` | int64   | count   | `/proc/vmstat` `pgmajfault`               | Major page faults requiring disk I/O to resolve. High rates indicate insufficient physical memory, causing swapping and significant slowdowns. |

**Performance Analysis Notes:**

- High `vMemoryUsed` across all pipeline stages may indicate greedy allocation by executables
- Increasing `vMemorySwapUsed` during execution indicates memory exhaustion
- High `vMemoryMajorPageFault` rates correlate with severe performance degradation

---

## VM-Level Disk I/O Metrics

Prefix: `v` (VM-level). Disable with `--no-disk`.

Disk I/O metrics are aggregated across all physical storage devices and help identify storage bottlenecks. Different
pipeline stages often exhibit distinct disk I/O patterns—download stages may be network-limited while split stages are
often disk-write intensive.

| Key                     | Type  | Unit    | Source                   | Description                                                                                                              |
|-------------------------|-------|---------|--------------------------|--------------------------------------------------------------------------------------------------------------------------|
| `vDiskSectorReads`      | int64 | sectors | `/proc/diskstats` col 5  | Total sectors read across all disks. 1 sector = 512 bytes.                                                               |
| `vDiskSectorWrites`     | int64 | sectors | `/proc/diskstats` col 9  | Total sectors written across all disks.                                                                                  |
| `vDiskReadBytes`        | int64 | bytes   | Derived                  | Total bytes read (`sectors × 512`).                                                                                      |
| `vDiskWriteBytes`       | int64 | bytes   | Derived                  | Total bytes written (`sectors × 512`).                                                                                   |
| `vDiskSuccessfulReads`  | int64 | count   | `/proc/diskstats` col 3  | Number of completed read operations.                                                                                     |
| `vDiskSuccessfulWrites` | int64 | count   | `/proc/diskstats` col 7  | Number of completed write operations.                                                                                    |
| `vDiskMergedReads`      | int64 | count   | `/proc/diskstats` col 4  | Adjacent read requests merged by the I/O scheduler. High merge rates indicate sequential access patterns.                |
| `vDiskMergedWrites`     | int64 | count   | `/proc/diskstats` col 8  | Adjacent write requests merged by the I/O scheduler.                                                                     |
| `vDiskReadTime`         | int64 | ms      | `/proc/diskstats` col 6  | Total time spent reading (milliseconds). Includes queue wait time.                                                       |
| `vDiskWriteTime`        | int64 | ms      | `/proc/diskstats` col 10 | Total time spent writing (milliseconds).                                                                                 |
| `vDiskIOInProgress`     | int64 | count   | `/proc/diskstats` col 11 | Number of I/O operations currently in flight. Consistently high values indicate disk saturation.                         |
| `vDiskIOTime`           | int64 | ms      | `/proc/diskstats` col 12 | Total time the disk was busy with I/O. Used to calculate disk utilization percentage.                                    |
| `vDiskWeightedIOTime`   | int64 | ms      | `/proc/diskstats` col 13 | Weighted I/O time accounting for queue depth. High values relative to `vDiskIOTime` indicate I/O queuing and contention. |

**Device Filtering:** Only physical disks are included: `sd*`, `hd*`, `vd*`, `xvd*`, `nvme*n*`, `mmcblk*`. Partitions (
e.g., `sda1`) and loopback devices are excluded to avoid double-counting.

**Performance Analysis Notes:**

- Split/demultiplexing stages that write many small files are often limited by disk write speed
- High `vDiskIOInProgress` with high `vCpuTimeIOWait` indicates disk saturation
- Compare `vDiskWeightedIOTime` to `vDiskIOTime` to assess I/O queue depth

---

## VM-Level Network Metrics

Prefix: `v` (VM-level). Disable with `--no-network`.

Network metrics help identify network-bound stages such as data download phases. Values are aggregated across all
interfaces except loopback.

| Key                    | Type  | Unit  | Source                 | Description                                                                                  |
|------------------------|-------|-------|------------------------|----------------------------------------------------------------------------------------------|
| `vNetworkBytesRecvd`   | int64 | bytes | `/proc/net/dev` col 0  | Total bytes received across all non-loopback interfaces.                                     |
| `vNetworkBytesSent`    | int64 | bytes | `/proc/net/dev` col 8  | Total bytes transmitted across all non-loopback interfaces.                                  |
| `vNetworkPacketsRecvd` | int64 | count | `/proc/net/dev` col 1  | Total packets received.                                                                      |
| `vNetworkPacketsSent`  | int64 | count | `/proc/net/dev` col 9  | Total packets transmitted.                                                                   |
| `vNetworkErrorsRecvd`  | int64 | count | `/proc/net/dev` col 2  | Receive errors (CRC errors, framing errors, etc.).                                           |
| `vNetworkErrorsSent`   | int64 | count | `/proc/net/dev` col 10 | Transmit errors.                                                                             |
| `vNetworkDropsRecvd`   | int64 | count | `/proc/net/dev` col 3  | Packets dropped on receive (buffer overflow, etc.). High values indicate network congestion. |
| `vNetworkDropsSent`    | int64 | count | `/proc/net/dev` col 11 | Packets dropped on transmit.                                                                 |

**Note:** Loopback interface (`lo`) is excluded from aggregation. Values are cumulative since boot.

**Performance Analysis Notes:**

- Download/data transfer stages are typically limited by network bandwidth
- High `vNetworkDropsRecvd` indicates the system cannot process incoming data fast enough

---

## Container-Level Metrics

Prefix: `c` (Container-level). Disable with `--no-container`.

Container-level metrics isolate resource utilization to only the processes within the container, excluding background
processes on the host. This isolation is crucial for accurate profiling when multiple workloads share a host or when
background system processes are present. Docker leverages Linux cgroups for resource management, and these metrics are
collected from the `/sys/fs/cgroup` virtual filesystem.

The profiler auto-detects cgroup v1 or v2 and collects from the appropriate paths.

| Key                  | Type   | Unit         | Source (v1 / v2)                                           | Description                                                                                                                        |
|----------------------|--------|--------------|------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|
| `cCpuTime`           | int64  | nanoseconds  | `cpuacct.usage` / `cpu.stat usage_usec × 1000`             | Total CPU time consumed by all tasks in the container.                                                                             |
| `cCpuTimeUserMode`   | int64  | centiseconds | `cpuacct.stat user` / `cpu.stat user_usec ÷ 10000`         | CPU time consumed by container tasks in user mode. Represents time spent executing application code within the container.          |
| `cCpuTimeKernelMode` | int64  | centiseconds | `cpuacct.stat system` / `cpu.stat system_usec ÷ 10000`     | CPU time consumed by container tasks in kernel mode. Includes time for I/O operations and system calls.                            |
| `cMemoryUsed`        | int64  | bytes        | `memory.usage_in_bytes` / `memory.current`                 | Current memory usage by the container. Does not include host-level caching external to the container.                              |
| `cMemoryMaxUsed`     | int64  | bytes        | `memory.max_usage_in_bytes` / `memory.peak`                | Peak (high watermark) memory usage by the container since it started.                                                              |
| `cDiskReadBytes`     | int64  | bytes        | `blkio.throttle.io_service_bytes Read` / `io.stat rbytes`  | Number of bytes read from block devices by the container.                                                                          |
| `cDiskWriteBytes`    | int64  | bytes        | `blkio.throttle.io_service_bytes Write` / `io.stat wbytes` | Number of bytes written to block devices by the container. May differ from host metrics due to Docker's internal mount management. |
| `cDiskSectorIO`      | int64  | sectors      | `blkio.sectors` (v1 only)                                  | Total sectors transferred to/from specific devices (cgroup v1 only).                                                               |
| `cNetworkBytesRecvd` | int64  | bytes        | `/proc/net/dev` (container namespace)                      | Network bytes received by the container.                                                                                           |
| `cNetworkBytesSent`  | int64  | bytes        | `/proc/net/dev` (container namespace)                      | Network bytes sent by the container.                                                                                               |
| `cPgFault`           | int64  | count        | `memory.stat pgfault`                                      | Page faults within the container.                                                                                                  |
| `cMajorPgFault`      | int64  | count        | `memory.stat pgmajfault`                                   | Major page faults within the container requiring disk I/O.                                                                         |
| `cNumProcesses`      | int64  | count        | `pids.current` / tasks file                                | Number of processes currently running in the container.                                                                            |
| `cCpuPerCpuJson`     | string | -            | `cpuacct.usage_percpu` (v1 only)                           | Per-CPU time breakdown as JSON array (cgroup v1 only). Useful for detecting CPU affinity issues.                                   |

**Why Container Metrics Differ from Host Metrics:**

Container metrics may differ from host-level metrics because:

1. Host metrics include background processes outside the container
2. Docker manages disk writes through internal mount points; caching and management overhead appears in host metrics but
   not container metrics
3. Pipe management and context switching between processes are handled by the OS and reflected in host metrics

These differences help assess resource contention effects when multiple jobs run on the same physical host.

---

## Process-Level Metrics

Prefix: `p` (Process-level) or `proc{i}` when flattened. Disable with `--no-procs`.

Process-level metrics provide the finest granularity, enabling attribution of resource bottlenecks to specific tasks or
executables. This is particularly valuable for complex parallel pipelines where multiple processes run concurrently.

**Note:** Process-level profiling has the highest overhead and collection time varies with the number of processes. In
testing with up to 85 concurrent processes, collection remained under 100ms.

| Key                            | Type   | Unit         | Source                          | Description                                                                                                                         |
|--------------------------------|--------|--------------|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------|
| `pId`                          | int64  | -            | `/proc/[pid]`                   | Process ID.                                                                                                                         |
| `pName`                        | string | -            | `/proc/[pid]/status` Name       | Executable name (truncated to 15 characters by kernel).                                                                             |
| `pCmdline`                     | string | -            | `/proc/[pid]/cmdline`           | Full command line with arguments.                                                                                                   |
| `pNumThreads`                  | int64  | count        | `/proc/[pid]/stat` field 19     | Number of threads in the process.                                                                                                   |
| `pCpuTimeUserMode`             | int64  | centiseconds | `/proc/[pid]/stat` field 14     | CPU time this process has been scheduled in user mode. Time spent executing the process's own code.                                 |
| `pCpuTimeKernelMode`           | int64  | centiseconds | `/proc/[pid]/stat` field 15     | CPU time this process has been scheduled in kernel mode. Time spent in system calls.                                                |
| `pChildrenUserMode`            | int64  | centiseconds | `/proc/[pid]/stat` field 16     | CPU time waited-for children have been scheduled in user mode.                                                                      |
| `pChildrenKernelMode`          | int64  | centiseconds | `/proc/[pid]/stat` field 17     | CPU time waited-for children have been scheduled in kernel mode.                                                                    |
| `pVoluntaryContextSwitches`    | int64  | count        | `/proc/[pid]/status`            | Number of voluntary context switches (process yielded CPU, e.g., waiting for I/O).                                                  |
| `pNonvoluntaryContextSwitches` | int64  | count        | `/proc/[pid]/status`            | Number of involuntary context switches (process preempted by scheduler). High values indicate CPU contention.                       |
| `pBlockIODelays`               | int64  | centiseconds | `/proc/[pid]/stat` field 42     | Aggregated block I/O delays. Time the process spent waiting for block I/O to complete.                                              |
| `pVirtualMemoryBytes`          | int64  | bytes        | `/proc/[pid]/stat` field 23     | Virtual memory size. Total address space allocated (may exceed physical memory).                                                    |
| `pResidentSetSize`             | int64  | bytes        | `/proc/[pid]/statm` × page size | Physical memory (RSS). Number of pages the process has in real memory, multiplied by page size. Actual physical memory consumption. |

### Flattened Output Example

When flattening is enabled, processes are limited to the first 10 (`MaxFlattenedProcesses`) to control output size:

```json
{
  "processCount": 150,
  "proc0Pid": 1,
  "proc0PidT": 1735166000001,
  "proc0Name": "systemd",
  "proc0ResidentSetSize": 8388608,
  "proc1Pid": 2,
  "proc1Name": "kthreadd"
}
```

**Performance Analysis Notes:**

- Identify "straggler" processes by comparing `pCpuTimeUserMode` across parallel workers
- High `pNonvoluntaryContextSwitches` indicates the process is competing for CPU time
- High `pBlockIODelays` identifies processes waiting on slow storage
- Compare `pResidentSetSize` to find memory-heavy processes

---

## NVIDIA GPU Metrics

Prefix: `gpu` when flattened (with `gpuIndex` field) or fields stored in `NvidiaGPUs` array. Disable with `--no-nvidia`.
Disable process enumeration only with `--no-gpu-procs`.

GPU metrics are collected via the NVIDIA Management Library (NVML). All metrics are collected per-GPU and include
per-field timestamps.

### Utilization

| Key                       | Type  | Unit    | Source                                   | Description                                                                                                      |
|---------------------------|-------|---------|------------------------------------------|------------------------------------------------------------------------------------------------------------------|
| `utilizationGpu`          | int64 | percent | `nvmlDeviceGetUtilizationRates().gpu`    | GPU compute engine utilization. Percentage of time over the sampling period that compute kernels were executing. |
| `utilizationMemory`       | int64 | percent | `nvmlDeviceGetUtilizationRates().memory` | Memory controller utilization. Percentage of time memory was being read or written.                              |
| `utilizationEncoder`      | int64 | percent | `nvmlDeviceGetEncoderUtilization()`      | Video encoder engine utilization.                                                                                |
| `utilizationDecoder`      | int64 | percent | `nvmlDeviceGetDecoderUtilization()`      | Video decoder engine utilization.                                                                                |
| `utilizationJpeg`         | int64 | percent | `nvmlDeviceGetJpgUtilization()`          | JPEG engine utilization (Turing+ architectures).                                                                 |
| `utilizationOfa`          | int64 | percent | `nvmlDeviceGetOfaUtilization()`          | Optical Flow Accelerator utilization (Turing+ architectures).                                                    |
| `encoderSamplingPeriodUs` | int64 | μs      | NVML                                     | Encoder sampling period.                                                                                         |
| `decoderSamplingPeriodUs` | int64 | μs      | NVML                                     | Decoder sampling period.                                                                                         |

### Memory

| Key                   | Type  | Unit  | Source                            | Description                               |
|-----------------------|-------|-------|-----------------------------------|-------------------------------------------|
| `memoryUsedBytes`     | int64 | bytes | `nvmlDeviceGetMemoryInfo().used`  | GPU frame buffer memory currently in use. |
| `memoryFreeBytes`     | int64 | bytes | `nvmlDeviceGetMemoryInfo().free`  | GPU frame buffer memory available.        |
| `memoryTotalBytes`    | int64 | bytes | `nvmlDeviceGetMemoryInfo().total` | Total GPU frame buffer memory.            |
| `memoryReservedBytes` | int64 | bytes | `nvmlDeviceGetMemoryInfo_v2()`    | Memory reserved by driver/system.         |
| `bar1UsedBytes`       | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()`   | Used BAR1 aperture memory.                |
| `bar1FreeBytes`       | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()`   | Free BAR1 aperture memory.                |
| `bar1TotalBytes`      | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()`   | Total BAR1 aperture size.                 |

### Temperature

| Key                  | Type  | Unit | Source                             | Description                                       |
|----------------------|-------|------|------------------------------------|---------------------------------------------------|
| `temperatureGpuC`    | int64 | °C   | `nvmlDeviceGetTemperature(GPU)`    | GPU die temperature.                              |
| `temperatureMemoryC` | int64 | °C   | `nvmlDeviceGetTemperature(MEMORY)` | HBM memory temperature (Ampere+ with HBM memory). |

### Fan

| Key               | Type   | Unit    | Source                       | Description                     |
|-------------------|--------|---------|------------------------------|---------------------------------|
| `fanSpeedPercent` | int64  | percent | `nvmlDeviceGetFanSpeed()`    | Fan speed as percentage of max. |
| `fanSpeedsJson`   | string | -       | `nvmlDeviceGetFanSpeed_v2()` | Per-fan speeds as JSON array.   |

### Clocks

| Key                   | Type  | Unit | Source                             | Description                             |
|-----------------------|-------|------|------------------------------------|-----------------------------------------|
| `clockGraphicsMhz`    | int64 | MHz  | `nvmlDeviceGetClockInfo(GRAPHICS)` | Current graphics engine clock.          |
| `clockSmMhz`          | int64 | MHz  | `nvmlDeviceGetClockInfo(SM)`       | Current streaming multiprocessor clock. |
| `clockMemoryMhz`      | int64 | MHz  | `nvmlDeviceGetClockInfo(MEM)`      | Current memory clock.                   |
| `clockVideoMhz`       | int64 | MHz  | `nvmlDeviceGetClockInfo(VIDEO)`    | Current video encoder/decoder clock.    |
| `appClockGraphicsMhz` | int64 | MHz  | `nvmlDeviceGetApplicationsClock()` | Application-requested graphics clock.   |
| `appClockMemoryMhz`   | int64 | MHz  | `nvmlDeviceGetApplicationsClock()` | Application-requested memory clock.     |

### Performance State

| Key                | Type | Unit | Source                            | Description                                      |
|--------------------|------|------|-----------------------------------|--------------------------------------------------|
| `performanceState` | int  | -    | `nvmlDeviceGetPerformanceState()` | Current P-state (0=max performance, 15=minimum). |

### Power

| Key                    | Type  | Unit        | Source                                  | Description                                                  |
|------------------------|-------|-------------|-----------------------------------------|--------------------------------------------------------------|
| `powerUsageMw`         | int64 | milliwatts  | `nvmlDeviceGetPowerUsage()`             | Current power draw of the GPU.                               |
| `powerLimitMw`         | int64 | milliwatts  | `nvmlDeviceGetPowerManagementLimit()`   | Current power limit setting.                                 |
| `powerEnforcedLimitMw` | int64 | milliwatts  | `nvmlDeviceGetEnforcedPowerLimit()`     | Actually enforced power limit (may differ from set limit).   |
| `energyConsumptionMj`  | int64 | millijoules | `nvmlDeviceGetTotalEnergyConsumption()` | Total energy consumed since driver load. Cumulative counter. |

### PCIe

| Key                    | Type  | Unit      | Source                                  | Description                                        |
|------------------------|-------|-----------|-----------------------------------------|----------------------------------------------------|
| `pcieTxBytesPerSec`    | int64 | bytes/sec | `nvmlDeviceGetPcieThroughput(TX)`       | PCIe transmit throughput (GPU to host).            |
| `pcieRxBytesPerSec`    | int64 | bytes/sec | `nvmlDeviceGetPcieThroughput(RX)`       | PCIe receive throughput (host to GPU).             |
| `pcieCurrentLinkGen`   | int   | -         | `nvmlDeviceGetCurrPcieLinkGeneration()` | Current PCIe link generation (1-5).                |
| `pcieCurrentLinkWidth` | int   | -         | `nvmlDeviceGetCurrPcieLinkWidth()`      | Current PCIe link width (lanes).                   |
| `pcieReplayCounter`    | int64 | count     | `nvmlDeviceGetPcieReplayCounter()`      | PCIe replay errors. Non-zero indicates bus issues. |

### Throttling

| Key                      | Type   | Unit        | Source                                      | Description                                             |
|--------------------------|--------|-------------|---------------------------------------------|---------------------------------------------------------|
| `clocksEventReasons`     | uint64 | bitmask     | `nvmlDeviceGetCurrentClocksEventReasons()`  | Raw throttle reason bitmask.                            |
| `throttleReasonsActive`  | array  | -           | Decoded                                     | Human-readable active throttle reasons as string array. |
| `violationPowerNs`       | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(POWER)`       | Cumulative time throttled due to power limit.           |
| `violationThermalNs`     | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(THERMAL)`     | Cumulative time throttled due to temperature.           |
| `violationReliabilityNs` | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(RELIABILITY)` | Cumulative time throttled for reliability.              |
| `violationBoardLimitNs`  | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(BOARD_LIMIT)` | Cumulative time at board power limit.                   |
| `violationLowUtilNs`     | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(LOW_UTIL)`    | Cumulative time at low utilization clocks.              |
| `violationSyncBoostNs`   | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(SYNC_BOOST)`  | Cumulative time in sync boost state.                    |

### ECC Errors

| Key                         | Type  | Unit  | Source                                                | Description                                        |
|-----------------------------|-------|-------|-------------------------------------------------------|----------------------------------------------------|
| `eccVolatileSbe`            | int64 | count | `nvmlDeviceGetTotalEccErrors(CORRECTED, VOLATILE)`    | Correctable single-bit errors since driver load.   |
| `eccVolatileDbe`            | int64 | count | `nvmlDeviceGetTotalEccErrors(UNCORRECTED, VOLATILE)`  | Uncorrectable double-bit errors since driver load. |
| `eccAggregateSbe`           | int64 | count | `nvmlDeviceGetTotalEccErrors(CORRECTED, AGGREGATE)`   | Lifetime correctable errors.                       |
| `eccAggregateDbe`           | int64 | count | `nvmlDeviceGetTotalEccErrors(UNCORRECTED, AGGREGATE)` | Lifetime uncorrectable errors.                     |
| `retiredPagesSbe`           | int64 | count | `nvmlDeviceGetRetiredPages(SBE)`                      | Memory pages retired due to single-bit errors.     |
| `retiredPagesDbe`           | int64 | count | `nvmlDeviceGetRetiredPages(DBE)`                      | Memory pages retired due to double-bit errors.     |
| `retiredPending`            | bool  | -     | `nvmlDeviceGetRetiredPagesPendingStatus()`            | Whether page retirement is pending reboot.         |
| `remappedRowsCorrectable`   | int64 | count | `nvmlDeviceGetRemappedRows()`                         | Memory rows remapped for correctable errors.       |
| `remappedRowsUncorrectable` | int64 | count | `nvmlDeviceGetRemappedRows()`                         | Memory rows remapped for uncorrectable errors.     |
| `remappedRowsPending`       | bool  | -     | `nvmlDeviceGetRemappedRows()`                         | Row remapping pending.                             |
| `remappedRowsFailure`       | bool  | -     | `nvmlDeviceGetRemappedRows()`                         | Row remapping failure occurred.                    |

### Encoder/Decoder Stats

| Key                   | Type | Unit  | Source                        | Description                           |
|-----------------------|------|-------|-------------------------------|---------------------------------------|
| `encoderSessionCount` | int  | count | `nvmlDeviceGetEncoderStats()` | Active encoder sessions.              |
| `encoderAvgFps`       | int  | fps   | `nvmlDeviceGetEncoderStats()` | Average encoder FPS.                  |
| `encoderAvgLatencyUs` | int  | μs    | `nvmlDeviceGetEncoderStats()` | Average encoder latency.              |
| `fbcSessionCount`     | int  | count | `nvmlDeviceGetFBCStats()`     | Active frame buffer capture sessions. |
| `fbcAvgFps`           | int  | fps   | `nvmlDeviceGetFBCStats()`     | Average FBC FPS.                      |
| `fbcAvgLatencyUs`     | int  | μs    | `nvmlDeviceGetFBCStats()`     | Average FBC latency.                  |

### NVLink

| Key                   | Type   | Description                          |
|-----------------------|--------|--------------------------------------|
| `nvlinkBandwidthJson` | string | Per-link TX/RX bytes as JSON array.  |
| `nvlinkErrorsJson`    | string | Per-link error counts as JSON array. |

NVLink bandwidth object fields: `link`, `txBytes`, `rxBytes`, `throughputTx`, `throughputRx`
NVLink errors object fields: `link`, `crcErrors`, `eccErrors`, `replayErrors`, `recoveryCount`

### GPU Processes

| Key                      | Type   | Description                                    |
|--------------------------|--------|------------------------------------------------|
| `processCount`           | int64  | Number of processes using this GPU.            |
| `processesJson`          | string | Process details as JSON array.                 |
| `processUtilizationJson` | string | Per-process utilization samples as JSON array. |

GPU process object fields: `pid`, `name`, `usedMemoryBytes`, `type` (compute/graphics/mps)

### Flattened Output Example

When flattened, each GPU generates a separate record with a `gpuIndex` field:

```json
{
  "timestamp": 1735166000000,
  "gpuIndex": 0,
  "gpuUtilizationGpu": 85,
  "gpuUtilizationGpuT": 1735166000002,
  "gpuMemoryUsedBytes": 8589934592,
  "gpuPowerUsageMw": 245500,
  "gpuTemperatureGpuC": 65,
  "gpuPerformanceState": 0
}
```

---

## vLLM Inference Metrics

Prefix: `vllm`. Disable with `--no-vllm`. Configure endpoint via `VLLM_METRICS_URL` environment variable (default:
`http://localhost:8000/metrics`).

These metrics are scraped from vLLM's Prometheus-format metrics endpoint and provide insight into inference engine
performance.

### Availability & Timing

| Key             | Type  | Unit        | Source        | Description                         |
|-----------------|-------|-------------|---------------|-------------------------------------|
| `vllmAvailable` | bool  | -           | HTTP response | Whether vLLM endpoint is reachable. |
| `vllmTimestamp` | int64 | nanoseconds | Scrape time   | When metrics were collected.        |

### System State

| Key                    | Type    | Unit  | Source                      | Description                               |
|------------------------|---------|-------|-----------------------------|-------------------------------------------|
| `vllmRequestsRunning`  | float64 | count | `vllm:num_requests_running` | Number of requests currently processing.  |
| `vllmRequestsWaiting`  | float64 | count | `vllm:num_requests_waiting` | Number of requests queued for processing. |
| `vllmEngineSleepState` | float64 | -     | `vllm:engine_sleep_state`   | Engine sleep state indicator.             |
| `vllmPreemptionsTotal` | float64 | count | `vllm:num_preemptions`      | Total preempted requests.                 |

### Cache Metrics

| Key                       | Type    | Unit  | Source                      | Description                     |
|---------------------------|---------|-------|-----------------------------|---------------------------------|
| `vllmKvCacheUsagePercent` | float64 | ratio | `vllm:kv_cache_usage_perc`  | KV-cache utilization (0.0-1.0). |
| `vllmPrefixCacheHits`     | float64 | count | `vllm:prefix_cache_hits`    | Prefix cache hit count.         |
| `vllmPrefixCacheQueries`  | float64 | count | `vllm:prefix_cache_queries` | Prefix cache query count.       |

### Throughput

| Key                          | Type    | Unit  | Source                    | Description                      |
|------------------------------|---------|-------|---------------------------|----------------------------------|
| `vllmRequestsFinishedTotal`  | float64 | count | `vllm:request_success`    | Total completed requests.        |
| `vllmRequestsCorruptedTotal` | float64 | count | `vllm:corrupted_requests` | Total failed/corrupted requests. |
| `vllmTokensPromptTotal`      | float64 | count | `vllm:prompt_tokens`      | Total prompt tokens processed.   |
| `vllmTokensGenerationTotal`  | float64 | count | `vllm:generation_tokens`  | Total tokens generated.          |

### Latency (Sums and Counts)

These histogram sum/count metrics enable calculation of averages: `average = sum / count`

| Key                         | Type    | Unit    | Source                                      | Description                       |
|-----------------------------|---------|---------|---------------------------------------------|-----------------------------------|
| `vllmLatencyTtftSum`        | float64 | seconds | `vllm:time_to_first_token_seconds_sum`      | Time-to-first-token sum.          |
| `vllmLatencyTtftCount`      | float64 | count   | `vllm:time_to_first_token_seconds_count`    | Time-to-first-token count.        |
| `vllmLatencyE2eSum`         | float64 | seconds | `vllm:e2e_request_latency_seconds_sum`      | End-to-end request latency sum.   |
| `vllmLatencyE2eCount`       | float64 | count   | `vllm:e2e_request_latency_seconds_count`    | End-to-end request latency count. |
| `vllmLatencyQueueSum`       | float64 | seconds | `vllm:request_queue_time_seconds_sum`       | Queue wait time sum.              |
| `vllmLatencyQueueCount`     | float64 | count   | `vllm:request_queue_time_seconds_count`     | Queue wait time count.            |
| `vllmLatencyInferenceSum`   | float64 | seconds | `vllm:request_inference_time_seconds_sum`   | Inference time sum.               |
| `vllmLatencyInferenceCount` | float64 | count   | `vllm:request_inference_time_seconds_count` | Inference time count.             |
| `vllmLatencyPrefillSum`     | float64 | seconds | `vllm:request_prefill_time_seconds_sum`     | Prefill phase time sum.           |
| `vllmLatencyPrefillCount`   | float64 | count   | `vllm:request_prefill_time_seconds_count`   | Prefill phase time count.         |
| `vllmLatencyDecodeSum`      | float64 | seconds | `vllm:request_decode_time_seconds_sum`      | Decode phase time sum.            |
| `vllmLatencyDecodeCount`    | float64 | count   | `vllm:request_decode_time_seconds_count`    | Decode phase time count.          |

### Histograms

Histogram bucket data is stored as JSON in `vllmHistogramsJson`:

```json
{
  "vllmHistogramsJson": "{\"latencyTtft\":{\"0.001\":0,\"0.005\":12,\"0.01\":45,\"inf\":300},...}"
}
```

Available histogram fields:

- **Latency**: `latencyTtft`, `latencyE2e`, `latencyQueue`, `latencyInference`, `latencyPrefill`, `latencyDecode`,
  `latencyInterToken`
- **Request Sizes**: `reqSizePromptTokens`, `reqSizeGenerationTokens`
- **Iteration**: `tokensPerStep`
- **Parameters**: `reqParamsMaxTokens`, `reqParamsN`

---

## Environment Variables

| Variable           | Default                         | Description                       |
|--------------------|---------------------------------|-----------------------------------|
| `VLLM_METRICS_URL` | `http://localhost:8000/metrics` | vLLM Prometheus metrics endpoint. |

---