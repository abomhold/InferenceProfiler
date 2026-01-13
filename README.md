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

| Flag                  | Default             | Description                                                                            |
|-----------------------|---------------------|----------------------------------------------------------------------------------------|
| `-o <dir>`            | `./profiler-output` | Output directory for logs and exported data                                            |
| `-t <ms>`             | `1`                 | Sampling interval in milliseconds                                                      |
| `-f <format>`         | `jsonl`             | Export format: `jsonl`, `parquet`, `csv`, `tsv`                                        |
| `--stream`            | `false`             | Stream mode: write directly to output file instead of batch processing                 |
| `--no-flatten`        | `false`             | Disable flattening of nested data (GPUs, processes) to columns; stores as JSON strings |
| `--no-cleanup`        | `false`             | Keep intermediary JSON snapshot files after final export (batch mode only)             |
| `--graphs`            | `true`              | Generate HTML graphs after profiling                                                   |
| `--graph-only <file>` | -                   | Generate graphs from existing output file (skips profiling)                            |

### Collector Toggles

All collectors are enabled by default.

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
./profiler --graph-only ./profiler-output/abc123.parquet
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

| Mode   | Flag       | Description                                         |
|--------|------------|-----------------------------------------------------|
| Batch  | (default)  | Writes intermediate JSON files, aggregates at end   |
| Stream | `--stream` | Writes directly to final format, lower memory usage |

### Flatten Modes

The `--no-flatten` flag controls how nested data (GPUs, processes) is represented:

#### Flattened Mode (default)

GPUs and processes are expanded into individual columns with indexed prefixes:

```json
{
  "timestamp": 1735166000000,
  "vCpuTime": 123456,
  "nvidiaGpuCount": 2,
  "nvidia0UtilizationGpu": 85,
  "nvidia0MemoryUsedBytes": 8589934592,
  "nvidia1UtilizationGpu": 72,
  "processCount": 150,
  "process0Pid": 1,
  "process0Name": "systemd"
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

All dynamic metrics include per-field timestamps with a `T` suffix:

| Field           | Description                                                       |
|-----------------|-------------------------------------------------------------------|
| `timestamp`     | Overall collection cycle timestamp (nanoseconds since Unix epoch) |
| `{metricName}`  | The metric value                                                  |
| `{metricName}T` | Timestamp when this specific metric was collected (nanoseconds)   |

### Output Files

Each profiling session produces:

| File                      | Description                                               |
|---------------------------|-----------------------------------------------------------|
| `{uuid}.json`             | Static system information (collected once at startup)     |
| `{uuid}-{timestamp}.json` | Individual snapshot files (batch mode only, intermediate) |
| `{uuid}.{ext}`            | Final aggregated output (jsonl/parquet/csv/tsv)           |
| `{uuid}-graphs.html`      | Interactive HTML graphs (if `--graphs` enabled)           |

---

## Static System Information

Static information is collected once at profiler startup and saved to `{uuid}.json`.

### Session & Host Identification

| Key         | Type   | Source                                                | Description                               |
|-------------|--------|-------------------------------------------------------|-------------------------------------------|
| `uuid`      | string | Generated UUID v4                                     | Unique session identifier                 |
| `vId`       | string | `/sys/class/dmi/id/product_uuid` or `/etc/machine-id` | VM/instance identifier                    |
| `vHostname` | string | `syscall.Uname().Nodename`                            | System hostname                           |
| `vBootTime` | int64  | `/proc/stat` `btime`                                  | System boot time (Unix timestamp seconds) |

### CPU Static Info

| Key              | Type   | Source                                | Description                                         |
|------------------|--------|---------------------------------------|-----------------------------------------------------|
| `vNumProcessors` | int    | `runtime.NumCPU()`                    | Number of logical CPUs/cores                        |
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

| Key                 | Type  | Source                             | Description                 |
|---------------------|-------|------------------------------------|-----------------------------|
| `vMemoryTotalBytes` | int64 | `/proc/meminfo` `MemTotal` × 1024  | Total physical RAM in bytes |
| `vSwapTotalBytes`   | int64 | `/proc/meminfo` `SwapTotal` × 1024 | Total swap space in bytes   |

### Container Static Info

| Key              | Type   | Source              | Description              |
|------------------|--------|---------------------|--------------------------|
| `cId`            | string | `/proc/self/cgroup` | Container ID or hostname |
| `cNumProcessors` | int64  | `runtime.NumCPU()`  | Available processors     |
| `cCgroupVersion` | int64  | Auto-detected       | Cgroup version (1 or 2)  |

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

| Key                   | Type    | Unit         | Source                               | Description                       |
|-----------------------|---------|--------------|--------------------------------------|-----------------------------------|
| `vCpuTime`            | int64   | centiseconds | `/proc/stat`                         | Total CPU time (user + kernel)    |
| `vCpuTimeUserMode`    | int64   | centiseconds | `/proc/stat` column 1                | User-space execution time         |
| `vCpuTimeKernelMode`  | int64   | centiseconds | `/proc/stat` column 3                | Kernel-mode execution time        |
| `vCpuIdleTime`        | int64   | centiseconds | `/proc/stat` column 4                | Idle time                         |
| `vCpuTimeIOWait`      | int64   | centiseconds | `/proc/stat` column 5                | I/O wait time                     |
| `vCpuTimeIntSrvc`     | int64   | centiseconds | `/proc/stat` column 6                | Hardware interrupt time           |
| `vCpuTimeSoftIntSrvc` | int64   | centiseconds | `/proc/stat` column 7                | Software interrupt time           |
| `vCpuNice`            | int64   | centiseconds | `/proc/stat` column 2                | Nice (low priority) time          |
| `vCpuSteal`           | int64   | centiseconds | `/proc/stat` column 8                | Hypervisor stolen time            |
| `vCpuContextSwitches` | int64   | count        | `/proc/stat` `ctxt`                  | Total context switches since boot |
| `vLoadAvg`            | float64 | ratio        | `/proc/loadavg`                      | 1-minute load average             |
| `vCpuMhz`             | float64 | MHz          | `/sys/devices/system/cpu/*/cpufreq/` | Current average CPU frequency     |

**Notes:**

- All time values are cumulative since system boot
- 1 centisecond = 10 milliseconds
- To calculate utilization: `(delta_busy / delta_total) × 100`

---

## VM-Level Memory Metrics

Prefix: `v` (VM-level). Disable with `--no-memory`.

| Key                     | Type    | Unit    | Source                                    | Description                  |
|-------------------------|---------|---------|-------------------------------------------|------------------------------|
| `vMemoryTotal`          | int64   | bytes   | `/proc/meminfo` `MemTotal`                | Total physical RAM           |
| `vMemoryFree`           | int64   | bytes   | `/proc/meminfo` `MemAvailable`            | Available memory             |
| `vMemoryUsed`           | int64   | bytes   | Derived                                   | Actively used memory         |
| `vMemoryBuffers`        | int64   | bytes   | `/proc/meminfo` `Buffers`                 | Kernel buffer memory         |
| `vMemoryCached`         | int64   | bytes   | `/proc/meminfo` `Cached` + `SReclaimable` | Page cache memory            |
| `vMemoryPercent`        | float64 | percent | Derived                                   | RAM usage percentage         |
| `vMemorySwapTotal`      | int64   | bytes   | `/proc/meminfo` `SwapTotal`               | Total swap space             |
| `vMemorySwapFree`       | int64   | bytes   | `/proc/meminfo` `SwapFree`                | Free swap space              |
| `vMemorySwapUsed`       | int64   | bytes   | Derived                                   | Used swap space              |
| `vMemoryPgFault`        | int64   | count   | `/proc/vmstat` `pgfault`                  | Minor page faults            |
| `vMemoryMajorPageFault` | int64   | count   | `/proc/vmstat` `pgmajfault`               | Major page faults (disk I/O) |

---

## VM-Level Disk I/O Metrics

Prefix: `v` (VM-level). Disable with `--no-disk`.

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

**Device Filtering:** Only physical disks are included: `sd*`, `hd*`, `vd*`, `xvd*`, `nvme*n*`, `mmcblk*`. Partitions
and loopback devices are excluded.

---

## VM-Level Network Metrics

Prefix: `v` (VM-level). Disable with `--no-network`.

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

Prefix: `c` (Container-level). Disable with `--no-container`.

Auto-detects cgroup v1 or v2.

| Key                  | Type   | Unit         | Source                                        | Description                 |
|----------------------|--------|--------------|-----------------------------------------------|-----------------------------|
| `cCpuTime`           | int64  | nanoseconds  | `cpuacct.usage` / `cpu.stat usage_usec`       | Total CPU time              |
| `cCpuTimeUserMode`   | int64  | centiseconds | `cpuacct.stat` / `cpu.stat user_usec`         | User mode CPU time          |
| `cCpuTimeKernelMode` | int64  | centiseconds | `cpuacct.stat` / `cpu.stat system_usec`       | Kernel mode CPU time        |
| `cMemoryUsed`        | int64  | bytes        | `memory.usage_in_bytes` / `memory.current`    | Current memory usage        |
| `cMemoryMaxUsed`     | int64  | bytes        | `memory.max_usage_in_bytes` / `memory.peak`   | Peak memory usage           |
| `cDiskReadBytes`     | int64  | bytes        | `blkio.throttle.io_service_bytes` / `io.stat` | Bytes read                  |
| `cDiskWriteBytes`    | int64  | bytes        | `blkio.throttle.io_service_bytes` / `io.stat` | Bytes written               |
| `cDiskSectorIO`      | int64  | sectors      | `blkio.sectors` (v1 only)                     | Total sector I/O            |
| `cNetworkBytesRecvd` | int64  | bytes        | `/proc/net/dev`                               | Network bytes received      |
| `cNetworkBytesSent`  | int64  | bytes        | `/proc/net/dev`                               | Network bytes sent          |
| `cPgFault`           | int64  | count        | `memory.stat pgfault`                         | Page faults                 |
| `cMajorPgFault`      | int64  | count        | `memory.stat pgmajfault`                      | Major page faults           |
| `cNumProcesses`      | int64  | count        | `pids.current` / tasks file                   | Process count in cgroup     |
| `cCpuPerCpuJson`     | string | -            | `cpuacct.usage_percpu` (v1 only)              | Per-CPU times as JSON array |

---

## Process-Level Metrics

Prefix: `process{i}` (flattened) or `processesJson` (JSON mode). Disable with `--no-procs`.

| Key                            | Type   | Unit         | Source                          | Description                  |
|--------------------------------|--------|--------------|---------------------------------|------------------------------|
| `pId`                          | int64  | -            | `/proc/[pid]`                   | Process ID                   |
| `pName`                        | string | -            | `/proc/[pid]/status`            | Executable name              |
| `pCmdline`                     | string | -            | `/proc/[pid]/cmdline`           | Full command line            |
| `pNumThreads`                  | int64  | count        | `/proc/[pid]/stat` field 19     | Thread count                 |
| `pCpuTimeUserMode`             | int64  | centiseconds | `/proc/[pid]/stat` field 14     | User CPU time                |
| `pCpuTimeKernelMode`           | int64  | centiseconds | `/proc/[pid]/stat` field 15     | Kernel CPU time              |
| `pChildrenUserMode`            | int64  | centiseconds | `/proc/[pid]/stat` field 16     | Children user time           |
| `pChildrenKernelMode`          | int64  | centiseconds | `/proc/[pid]/stat` field 17     | Children kernel time         |
| `pVoluntaryContextSwitches`    | int64  | count        | `/proc/[pid]/status`            | Voluntary context switches   |
| `pNonvoluntaryContextSwitches` | int64  | count        | `/proc/[pid]/status`            | Involuntary context switches |
| `pBlockIODelays`               | int64  | centiseconds | `/proc/[pid]/stat` field 42     | Block I/O wait time          |
| `pVirtualMemoryBytes`          | int64  | bytes        | `/proc/[pid]/stat` field 23     | Virtual memory size          |
| `pResidentSetSize`             | int64  | bytes        | `/proc/[pid]/statm` × page size | Physical memory (RSS)        |

### Flattened Output Example

```json
{
  "processCount": 150,
  "process0Pid": 1,
  "process0PidT": 1735166000001,
  "process0Name": "systemd",
  "process0ResidentSetSize": 8388608,
  "process1Pid": 2,
  "process1Name": "kthreadd"
}
```

---

## NVIDIA GPU Metrics

Prefix: `nvidia{i}` (flattened) or `nvidiaGpusJson` (JSON mode). Disable with `--no-nvidia`. Disable process enumeration
only with `--no-gpu-procs`.

### Utilization

| Key                       | Type  | Unit    | Source                                   | Description                        |
|---------------------------|-------|---------|------------------------------------------|------------------------------------|
| `utilizationGpu`          | int64 | percent | `nvmlDeviceGetUtilizationRates().gpu`    | GPU compute utilization            |
| `utilizationMemory`       | int64 | percent | `nvmlDeviceGetUtilizationRates().memory` | Memory controller utilization      |
| `utilizationEncoder`      | int64 | percent | `nvmlDeviceGetEncoderUtilization()`      | Video encoder utilization          |
| `utilizationDecoder`      | int64 | percent | `nvmlDeviceGetDecoderUtilization()`      | Video decoder utilization          |
| `utilizationJpeg`         | int64 | percent | `nvmlDeviceGetJpgUtilization()`          | JPEG engine utilization (Turing+)  |
| `utilizationOfa`          | int64 | percent | `nvmlDeviceGetOfaUtilization()`          | Optical Flow Accelerator (Turing+) |
| `encoderSamplingPeriodUs` | int64 | μs      | NVML                                     | Encoder sampling period            |
| `decoderSamplingPeriodUs` | int64 | μs      | NVML                                     | Decoder sampling period            |

### Memory

| Key                   | Type  | Unit  | Source                            | Description        |
|-----------------------|-------|-------|-----------------------------------|--------------------|
| `memoryUsedBytes`     | int64 | bytes | `nvmlDeviceGetMemoryInfo().used`  | Used frame buffer  |
| `memoryFreeBytes`     | int64 | bytes | `nvmlDeviceGetMemoryInfo().free`  | Free frame buffer  |
| `memoryTotalBytes`    | int64 | bytes | `nvmlDeviceGetMemoryInfo().total` | Total frame buffer |
| `memoryReservedBytes` | int64 | bytes | `nvmlDeviceGetMemoryInfo_v2()`    | Reserved memory    |
| `bar1UsedBytes`       | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()`   | Used BAR1 memory   |
| `bar1FreeBytes`       | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()`   | Free BAR1 memory   |
| `bar1TotalBytes`      | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()`   | Total BAR1 memory  |

### Temperature

| Key                  | Type  | Unit | Source                             | Description                       |
|----------------------|-------|------|------------------------------------|-----------------------------------|
| `temperatureGpuC`    | int64 | °C   | `nvmlDeviceGetTemperature(GPU)`    | GPU core temperature              |
| `temperatureMemoryC` | int64 | °C   | `nvmlDeviceGetTemperature(MEMORY)` | Memory temperature (HBM, Ampere+) |

### Fan

| Key               | Type   | Unit    | Source                       | Description                 |
|-------------------|--------|---------|------------------------------|-----------------------------|
| `fanSpeedPercent` | int64  | percent | `nvmlDeviceGetFanSpeed()`    | Fan speed                   |
| `fanSpeedsJson`   | string | -       | `nvmlDeviceGetFanSpeed_v2()` | Per-fan speeds (JSON array) |

### Clocks

| Key                   | Type  | Unit | Source                             | Description                |
|-----------------------|-------|------|------------------------------------|----------------------------|
| `clockGraphicsMhz`    | int64 | MHz  | `nvmlDeviceGetClockInfo(GRAPHICS)` | Graphics clock             |
| `clockSmMhz`          | int64 | MHz  | `nvmlDeviceGetClockInfo(SM)`       | SM clock                   |
| `clockMemoryMhz`      | int64 | MHz  | `nvmlDeviceGetClockInfo(MEM)`      | Memory clock               |
| `clockVideoMhz`       | int64 | MHz  | `nvmlDeviceGetClockInfo(VIDEO)`    | Video clock                |
| `appClockGraphicsMhz` | int64 | MHz  | `nvmlDeviceGetApplicationsClock()` | Application graphics clock |
| `appClockMemoryMhz`   | int64 | MHz  | `nvmlDeviceGetApplicationsClock()` | Application memory clock   |

### Performance State

| Key                | Type | Unit | Source                            | Description             |
|--------------------|------|------|-----------------------------------|-------------------------|
| `performanceState` | int  | -    | `nvmlDeviceGetPerformanceState()` | P-state (0=max, 15=min) |

### Power

| Key                    | Type  | Unit        | Source                                  | Description                    |
|------------------------|-------|-------------|-----------------------------------------|--------------------------------|
| `powerUsageMw`         | int64 | milliwatts  | `nvmlDeviceGetPowerUsage()`             | Current power draw             |
| `powerLimitMw`         | int64 | milliwatts  | `nvmlDeviceGetPowerManagementLimit()`   | Power limit                    |
| `powerEnforcedLimitMw` | int64 | milliwatts  | `nvmlDeviceGetEnforcedPowerLimit()`     | Enforced power limit           |
| `energyConsumptionMj`  | int64 | millijoules | `nvmlDeviceGetTotalEnergyConsumption()` | Total energy since driver load |

### PCIe

| Key                    | Type  | Unit      | Source                                  | Description             |
|------------------------|-------|-----------|-----------------------------------------|-------------------------|
| `pcieTxBytesPerSec`    | int64 | bytes/sec | `nvmlDeviceGetPcieThroughput(TX)`       | PCIe TX throughput      |
| `pcieRxBytesPerSec`    | int64 | bytes/sec | `nvmlDeviceGetPcieThroughput(RX)`       | PCIe RX throughput      |
| `pcieCurrentLinkGen`   | int   | -         | `nvmlDeviceGetCurrPcieLinkGeneration()` | Current PCIe generation |
| `pcieCurrentLinkWidth` | int   | -         | `nvmlDeviceGetCurrPcieLinkWidth()`      | Current PCIe link width |
| `pcieReplayCounter`    | int64 | count     | `nvmlDeviceGetPcieReplayCounter()`      | PCIe replay errors      |

### Throttling

| Key                         | Type   | Unit        | Source                                      | Description                          |
|-----------------------------|--------|-------------|---------------------------------------------|--------------------------------------|
| `clocksEventReasons`        | uint64 | bitmask     | `nvmlDeviceGetCurrentClocksEventReasons()`  | Raw throttle bitmask                 |
| `throttleReasonsActiveJson` | string | -           | Decoded                                     | Active throttle reasons (JSON array) |
| `violationPowerNs`          | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(POWER)`       | Time throttled due to power          |
| `violationThermalNs`        | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(THERMAL)`     | Time throttled due to thermal        |
| `violationReliabilityNs`    | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(RELIABILITY)` | Time throttled for reliability       |
| `violationBoardLimitNs`     | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(BOARD_LIMIT)` | Time at board limit                  |
| `violationLowUtilNs`        | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(LOW_UTIL)`    | Time at low utilization              |
| `violationSyncBoostNs`      | int64  | nanoseconds | `nvmlDeviceGetViolationStatus(SYNC_BOOST)`  | Time in sync boost                   |

### ECC Errors

| Key                         | Type  | Unit  | Source                                                | Description                          |
|-----------------------------|-------|-------|-------------------------------------------------------|--------------------------------------|
| `eccVolatileSbe`            | int64 | count | `nvmlDeviceGetTotalEccErrors(CORRECTED, VOLATILE)`    | Corrected errors since driver load   |
| `eccVolatileDbe`            | int64 | count | `nvmlDeviceGetTotalEccErrors(UNCORRECTED, VOLATILE)`  | Uncorrected errors since driver load |
| `eccAggregateSbe`           | int64 | count | `nvmlDeviceGetTotalEccErrors(CORRECTED, AGGREGATE)`   | Lifetime corrected errors            |
| `eccAggregateDbe`           | int64 | count | `nvmlDeviceGetTotalEccErrors(UNCORRECTED, AGGREGATE)` | Lifetime uncorrected errors          |
| `retiredPagesSbe`           | int64 | count | `nvmlDeviceGetRetiredPages(SBE)`                      | Pages retired due to SBE             |
| `retiredPagesDbe`           | int64 | count | `nvmlDeviceGetRetiredPages(DBE)`                      | Pages retired due to DBE             |
| `retiredPending`            | bool  | -     | `nvmlDeviceGetRetiredPagesPendingStatus()`            | Retirement pending reboot            |
| `remappedRowsCorrectable`   | int64 | count | `nvmlDeviceGetRemappedRows()`                         | Correctable remapped rows            |
| `remappedRowsUncorrectable` | int64 | count | `nvmlDeviceGetRemappedRows()`                         | Uncorrectable remapped rows          |
| `remappedRowsPending`       | bool  | -     | `nvmlDeviceGetRemappedRows()`                         | Remapping pending                    |
| `remappedRowsFailure`       | bool  | -     | `nvmlDeviceGetRemappedRows()`                         | Remapping failure                    |

### Encoder/Decoder Stats

| Key                   | Type | Unit  | Source                        | Description             |
|-----------------------|------|-------|-------------------------------|-------------------------|
| `encoderSessionCount` | int  | count | `nvmlDeviceGetEncoderStats()` | Active encoder sessions |
| `encoderAvgFps`       | int  | fps   | `nvmlDeviceGetEncoderStats()` | Average encoder FPS     |
| `encoderAvgLatencyUs` | int  | μs    | `nvmlDeviceGetEncoderStats()` | Average encoder latency |
| `fbcSessionCount`     | int  | count | `nvmlDeviceGetFBCStats()`     | Active FBC sessions     |
| `fbcAvgFps`           | int  | fps   | `nvmlDeviceGetFBCStats()`     | Average FBC FPS         |
| `fbcAvgLatencyUs`     | int  | μs    | `nvmlDeviceGetFBCStats()`     | Average FBC latency     |

### NVLink

| Key                   | Type   | Description                        |
|-----------------------|--------|------------------------------------|
| `nvlinkBandwidthJson` | string | Per-link TX/RX bytes (JSON array)  |
| `nvlinkErrorsJson`    | string | Per-link error counts (JSON array) |

### GPU Processes

| Key                      | Type   | Description                                  |
|--------------------------|--------|----------------------------------------------|
| `processCount`           | int64  | Number of processes using this GPU           |
| `processesJson`          | string | Process details (JSON array)                 |
| `processUtilizationJson` | string | Per-process utilization samples (JSON array) |

### Flattened Output Example

```json
{
  "nvidiaGpuCount": 2,
  "nvidia0UtilizationGpu": 85,
  "nvidia0UtilizationGpuT": 1735166000002,
  "nvidia0MemoryUsedBytes": 8589934592,
  "nvidia0PowerUsageMw": 245500,
  "nvidia0TemperatureGpuC": 65,
  "nvidia0PerformanceState": 0,
  "nvidia0ProcessesJson": "[{\"pid\":1234,\"name\":\"python\",\"usedMemoryBytes\":4294967296,\"type\":\"compute\"}]",
  "nvidia1UtilizationGpu": 72,
  "nvidia1MemoryUsedBytes": 6442450944
}
```

---

## vLLM Inference Metrics

Prefix: `vllm`. Disable with `--no-vllm`. Configure endpoint via `VLLM_METRICS_URL` environment variable (default:
`http://localhost:8000/metrics`).

### Availability & Timing

| Key             | Type  | Unit        | Source        | Description                        |
|-----------------|-------|-------------|---------------|------------------------------------|
| `vllmAvailable` | bool  | -           | HTTP response | Whether vLLM endpoint is reachable |
| `vllmTimestamp` | int64 | nanoseconds | Scrape time   | When metrics were collected        |

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
  "vllmHistogramsJson": "{\"latencyTtft\":{\"0.001\":0,\"0.005\":12,\"0.01\":45,\"inf\":300},...}"
}
```

Available histogram fields:

- `latencyTtft`, `latencyE2e`, `latencyQueue`, `latencyInference`, `latencyPrefill`, `latencyDecode`,
  `latencyInterToken`
- `reqSizePromptTokens`, `reqSizeGenerationTokens`
- `tokensPerStep`, `reqParamsMaxTokens`, `reqParamsN`

---

## Environment Variables

| Variable           | Default                         | Description                      |
|--------------------|---------------------------------|----------------------------------|
| `VLLM_METRICS_URL` | `http://localhost:8000/metrics` | vLLM Prometheus metrics endpoint |