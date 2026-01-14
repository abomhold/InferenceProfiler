# Profiler Metrics Reference

This document provides a comprehensive reference for all metrics collected by the system profiler.

---

## Table of Contents

1. [Command-Line Options](#command-line-options)
2. [Output Formats](#output-formats)
3. [Static System Information](#static-system-information)
4. [VM-Level Metrics](#vm-level-metrics)
5. [Container-Level Metrics](#container-level-metrics)
6. [Process-Level Metrics](#process-level-metrics)
7. [NVIDIA GPU Metrics](#nvidia-gpu-metrics)
8. [vLLM Inference Metrics](#vllm-inference-metrics)

---

## Command-Line Options

### Configuration Flags

| Category | Flag | Default | Description |
|----------|------|---------|-------------|
| **Output** | `-o <dir>` | `./profiler-output` | Output directory for logs and exported data |
| | `-t <ms>` | `100` | Sampling interval in milliseconds |
| | `-f <format>` | `jsonl` | Export format: `jsonl`, `parquet`, `csv`, `tsv` |
| | `--stream` | `false` | Stream mode: write directly to output file (lower memory usage) |
| | `--no-flatten` | `false` | Disable flattening of nested data; stores GPUs/processes as JSON strings |
| | `--no-cleanup` | `false` | Keep intermediate JSON snapshot files after export (batch mode only) |
| | `--graphs` | `true` | Generate interactive HTML visualization after profiling |
| | `-i <file>` | - | Input file for graph generation |
| | `-g` | `false` | Graph-only mode: generate graphs from `-i` file without profiling |
| **Collectors** | `--no-cpu` | `false` | Disable CPU metrics collection |
| | `--no-memory` | `false` | Disable memory metrics collection |
| | `--no-disk` | `false` | Disable disk I/O metrics collection |
| | `--no-network` | `false` | Disable network metrics collection |
| | `--no-container` | `false` | Disable container/cgroup metrics collection |
| | `--no-procs` | `false` | Disable per-process metrics collection |
| | `--no-nvidia` | `false` | Disable all NVIDIA GPU metrics collection |
| | `--no-gpu-procs` | `false` | Disable GPU process enumeration (still collects GPU metrics) |
| | `--no-vllm` | `false` | Disable vLLM metrics collection |

### Example Usage

```bash
# Minimal: CPU/memory only, TSV output
./profiler --no-disk --no-network --no-container --no-procs --no-nvidia --no-vllm -f tsv

# Full GPU monitoring without process overhead, streaming to Parquet
./profiler --no-procs --no-gpu-procs -f parquet --stream

# Profile a subprocess with all collectors enabled
./profiler -f jsonl -- python train.py --epochs 10

# Generate graphs from existing output without profiling
./profiler -g -i ./profiler-output/abc123.parquet
```

---

## Output Formats

### Export Formats

| Format | Extension | Description |
|--------|-----------|-------------|
| `jsonl` | `.jsonl` | JSON Lines - one JSON object per line, human-readable and streamable |
| `parquet` | `.parquet` | Columnar format - efficient compression and fast analytics queries |
| `csv` | `.csv` | Comma-separated values - universal compatibility with Excel/spreadsheets |
| `tsv` | `.tsv` | Tab-separated values - better handling of text with commas |

**Streaming vs Batch Mode:** Use `--stream` to write directly to the final format with lower memory usage. Default batch mode writes intermediate JSON snapshots then aggregates at the end.

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

| Field | Description |
|-------|-------------|
| `timestamp` | Overall collection cycle start time (nanoseconds since Unix epoch) |
| `{metricName}` | The metric value |
| `{metricName}T` | Timestamp when this specific metric was collected (nanoseconds since Unix epoch) |

**Note:** Individual metric timestamps (`{metricName}T`) allow precise correlation of measurements collected at slightly different times within a single collection cycle.

### Output Files

Each profiling session produces:

| File | Description |
|------|-------------|
| `{uuid}.json` | Static system information (collected once at startup) |
| `{uuid}-{timestamp}.json` | Individual snapshot files (batch mode only, intermediate) |
| `{uuid}.{ext}` | Final aggregated metrics (jsonl/parquet/csv/tsv) |
| `{uuid}.html` | Interactive HTML graphs with time series and heatmaps (if `--graphs` enabled) |

---

## Static System Information

Static information is collected once at profiler startup and saved to `{uuid}.json`. These values do not change during profiling.

### System Identification and Configuration

| Category | Key | Type | Unit | Source | Description |
|----------|-----|------|------|--------|-------------|
| **Session** | `uuid` | string | - | Generated UUID v4 | Unique session identifier for this profiling run |
| | `vId` | string | - | `/sys/class/dmi/id/product_uuid` or `/etc/machine-id` | VM/instance unique identifier |
| | `vBootTime` | int64 | seconds | `/proc/stat` `btime` | System boot time (Unix timestamp seconds) |
| **CPU** | `vNumProcessors` | int | count | `runtime.NumCPU()` | Number of logical CPUs/cores available to the OS |
| | `vCpuType` | string | - | `/proc/cpuinfo` `model name` | Processor model name (e.g., "Intel(R) Xeon(R) CPU E5-2680") |
| | `vCpuCache` | string | - | `/sys/devices/system/cpu/cpu*/cache/` | Cache hierarchy (e.g., "L1d:32K L1i:32K L2:256K L3:8M") |
| | `vKernelInfo` | string | - | `syscall.Uname()` | Full kernel version string including architecture |
| **Time Sync** | `vTimeSynced` | bool | - | `adjtimex()` syscall | Whether kernel clock is synchronized with time source (NTP/PTP) |
| | `vTimeOffsetSeconds` | float64 | seconds | `adjtimex()` syscall | Current time offset from reference clock |
| | `vTimeMaxErrorSeconds` | float64 | seconds | `adjtimex()` syscall | Maximum estimated time error bound |
| **Memory** | `vMemoryTotalBytes` | int64 | bytes | `/proc/meminfo` `MemTotal` × 1024 | Total installed physical RAM |
| | `vSwapTotalBytes` | int64 | bytes | `/proc/meminfo` `SwapTotal` × 1024 | Total configured swap space |
| **Container** | `cId` | string | - | `/proc/self/cgroup` | Container ID (from Docker/cgroup) or hostname if not containerized |
| | `cNumProcessors` | int64 | count | `runtime.NumCPU()` | Number of processors visible to container |
| | `cCgroupVersion` | int64 | - | Auto-detected | Cgroup version in use (1 or 2) |

### Network Interfaces

| Key | Type | Description |
|-----|------|-------------|
| `networkInterfaces` | string | JSON array of network interface objects (excludes loopback) |

Each network interface object contains:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Interface name (e.g., "eth0", "ens5") |
| `mac` | string | Hardware MAC address |
| `state` | string | Operational state ("up", "down", "unknown") |
| `mtu` | int64 | Maximum transmission unit (typically 1500 or 9000 for jumbo frames) |
| `speedMbps` | int64 | Link speed in megabits per second (-1 if unavailable/down) |

### Disk Devices

| Key | Type | Description |
|-----|------|-------------|
| `disks` | string | JSON array of disk objects (physical disks only, excludes partitions) |

Each disk object contains:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Device name (e.g., "sda", "nvme0n1", "vda") |
| `model` | string | Disk model identifier |
| `vendor` | string | Disk manufacturer |
| `sizeBytes` | int64 | Total disk capacity in bytes |

### NVIDIA GPU Static Information

| Category | Key | Type | Description |
|----------|-----|------|-------------|
| **Driver** | `nvidiaDriverVersion` | string | NVIDIA driver version (e.g., "525.125.06") |
| | `nvidiaCudaVersion` | string | CUDA driver version (e.g., "12.0") |
| | `nvmlVersion` | string | NVML library version |
| | `nvidiaGpuCount` | int | Number of GPUs detected |
| **GPU Array** | `nvidiaGpus` | string | JSON array of GPU static info objects |

Each GPU object in `nvidiaGpus` contains:

| Category | Field | Type | Description |
|----------|-------|------|-------------|
| **Identity** | `index` | int | Zero-based GPU index (0, 1, 2...) |
| | `name` | string | GPU model name (e.g., "NVIDIA A100-SXM4-80GB") |
| | `uuid` | string | Globally unique GPU identifier (persistent across reboots) |
| | `serial` | string | GPU board serial number |
| | `boardPartNumber` | string | Board part number for RMA/identification |
| | `brand` | string | Product brand (Tesla, GeForce, Quadro, RTX, etc.) |
| | `architecture` | string | GPU architecture (Ampere, Hopper, Ada, Turing, etc.) |
| **Compute** | `cudaCapabilityMajor` | int | CUDA compute capability major version |
| | `cudaCapabilityMinor` | int | CUDA compute capability minor version |
| | `numCores` | int | Number of CUDA cores |
| | `computeMode` | string | Compute mode (Default, ExclusiveProcess, Prohibited, ExclusiveThread) |
| **Memory** | `memoryTotalBytes` | int64 | Total frame buffer capacity in bytes |
| | `bar1TotalBytes` | int64 | Total BAR1 memory (CPU-visible GPU memory) in bytes |
| | `memoryBusWidthBits` | int | Memory interface width in bits (e.g., 5120 for HBM2e) |
| **Clocks** | `maxClockGraphicsMhz` | int | Maximum graphics engine clock in MHz |
| | `maxClockMemoryMhz` | int | Maximum memory clock in MHz |
| | `maxClockSmMhz` | int | Maximum SM (streaming multiprocessor) clock in MHz |
| | `maxClockVideoMhz` | int | Maximum video engine clock in MHz |
| **PCIe** | `pciBusId` | string | PCI bus ID (e.g., "0000:00:1e.0") |
| | `pciDeviceId` | uint32 | PCI device identifier |
| | `pciSubsystemId` | uint32 | PCI subsystem identifier |
| | `pcieMaxLinkGen` | int | Maximum PCIe generation supported (3, 4, 5) |
| | `pcieMaxLinkWidth` | int | Maximum PCIe lane width (x1, x4, x8, x16) |
| **Power** | `powerDefaultLimitMw` | int | Factory default power limit in milliwatts |
| | `powerMinLimitMw` | int | Minimum configurable power limit in milliwatts |
| | `powerMaxLimitMw` | int | Maximum configurable power limit in milliwatts |
| **Firmware** | `vbiosVersion` | string | Video BIOS version |
| | `inforomImageVersion` | string | InfoROM image version |
| | `inforomOemVersion` | string | InfoROM OEM version |
| **Thermal** | `numFans` | int | Number of cooling fans |
| | `tempShutdownC` | int | Critical shutdown temperature threshold in °C |
| | `tempSlowdownC` | int | Thermal throttling start temperature in °C |
| | `tempMaxOperatingC` | int | Maximum safe operating temperature in °C |
| | `tempTargetC` | int | Target temperature for thermal management in °C |
| **Features** | `eccModeEnabled` | bool | ECC memory error correction enabled status |
| | `persistenceModeOn` | bool | Persistence mode enabled (keeps driver loaded) |
| | `isMultiGpuBoard` | bool | Whether this is a multi-GPU board (e.g., Tesla K80) |
| | `displayModeEnabled` | bool | Display output capability enabled |
| | `displayActive` | bool | Display currently connected and active |
| | `migModeEnabled` | bool | Multi-Instance GPU mode enabled |
| | `encoderCapacityH264` | int | H.264 encoder capacity percentage (0-100) |
| | `encoderCapacityHevc` | int | HEVC (H.265) encoder capacity percentage (0-100) |
| | `encoderCapacityAv1` | int | AV1 encoder capacity percentage (0-100) |
| | `nvlinkCount` | int | Number of active NVLink connections |

---

## VM-Level Metrics

All VM-level metrics use the `v` prefix and represent system-wide measurements. Individual collectors can be disabled with corresponding flags.

| Category | Key | Type | Unit | Source | Description |
|----------|-----|------|------|--------|-------------|
| **CPU** (`--no-cpu`) | `vCpuTime` | int64 | centiseconds | `/proc/stat` | Total CPU time spent in all modes (user + system) |
| | `vCpuTimeUserMode` | int64 | centiseconds | `/proc/stat` column 1 | CPU time executing user-space code |
| | `vCpuTimeKernelMode` | int64 | centiseconds | `/proc/stat` column 3 | CPU time executing kernel code |
| | `vCpuIdleTime` | int64 | centiseconds | `/proc/stat` column 4 | CPU time spent idle |
| | `vCpuTimeIOWait` | int64 | centiseconds | `/proc/stat` column 5 | CPU time waiting for I/O operations to complete |
| | `vCpuTimeIntSrvc` | int64 | centiseconds | `/proc/stat` column 6 | CPU time servicing hardware interrupts |
| | `vCpuTimeSoftIntSrvc` | int64 | centiseconds | `/proc/stat` column 7 | CPU time servicing software interrupts (softirqs) |
| | `vCpuNice` | int64 | centiseconds | `/proc/stat` column 2 | CPU time executing nice (low priority) processes |
| | `vCpuSteal` | int64 | centiseconds | `/proc/stat` column 8 | CPU time stolen by hypervisor for other VMs |
| | `vCpuContextSwitches` | int64 | count | `/proc/stat` `ctxt` | Cumulative context switches since boot |
| | `vLoadAvg` | float64 | ratio | `/proc/loadavg` | 1-minute load average (runnable + uninterruptible processes) |
| | `vCpuMhz` | float64 | MHz | `/sys/devices/system/cpu/*/cpufreq/` | Current average CPU frequency across all cores |
| **Memory** (`--no-memory`) | `vMemoryTotal` | int64 | bytes | `/proc/meminfo` `MemTotal` | Total physical RAM |
| | `vMemoryFree` | int64 | bytes | `/proc/meminfo` `MemAvailable` | Memory available for allocation (includes reclaimable cache) |
| | `vMemoryUsed` | int64 | bytes | Derived | Actively used memory (excluding buffers/cache) |
| | `vMemoryBuffers` | int64 | bytes | `/proc/meminfo` `Buffers` | Memory used for kernel buffers |
| | `vMemoryCached` | int64 | bytes | `/proc/meminfo` `Cached` + `SReclaimable` | Memory used for page cache (can be reclaimed under pressure) |
| | `vMemoryPercent` | float64 | percent | Derived | Memory utilization percentage |
| | `vMemorySwapTotal` | int64 | bytes | `/proc/meminfo` `SwapTotal` | Total swap space configured |
| | `vMemorySwapFree` | int64 | bytes | `/proc/meminfo` `SwapFree` | Swap space available |
| | `vMemorySwapUsed` | int64 | bytes | Derived | Swap space in use |
| | `vMemoryPgFault` | int64 | count | `/proc/vmstat` `pgfault` | Cumulative minor page faults (resolved from cache) |
| | `vMemoryMajorPageFault` | int64 | count | `/proc/vmstat` `pgmajfault` | Cumulative major page faults (required disk I/O) |
| **Disk I/O** (`--no-disk`) | `vDiskSectorReads` | int64 | sectors | `/proc/diskstats` col 5 | Cumulative sectors read (1 sector = 512 bytes) |
| | `vDiskSectorWrites` | int64 | sectors | `/proc/diskstats` col 9 | Cumulative sectors written |
| | `vDiskReadBytes` | int64 | bytes | Derived | Cumulative bytes read from disk |
| | `vDiskWriteBytes` | int64 | bytes | Derived | Cumulative bytes written to disk |
| | `vDiskSuccessfulReads` | int64 | count | `/proc/diskstats` col 3 | Cumulative completed read operations |
| | `vDiskSuccessfulWrites` | int64 | count | `/proc/diskstats` col 7 | Cumulative completed write operations |
| | `vDiskMergedReads` | int64 | count | `/proc/diskstats` col 4 | Cumulative read requests merged by I/O scheduler |
| | `vDiskMergedWrites` | int64 | count | `/proc/diskstats` col 8 | Cumulative write requests merged by I/O scheduler |
| | `vDiskReadTime` | int64 | ms | `/proc/diskstats` col 6 | Cumulative time spent reading |
| | `vDiskWriteTime` | int64 | ms | `/proc/diskstats` col 10 | Cumulative time spent writing |
| | `vDiskIOInProgress` | int64 | count | `/proc/diskstats` col 11 | I/O operations currently in flight |
| | `vDiskIOTime` | int64 | ms | `/proc/diskstats` col 12 | Cumulative time with active I/O |
| | `vDiskWeightedIOTime` | int64 | ms | `/proc/diskstats` col 13 | Cumulative weighted I/O time (accounts for queue depth) |
| **Network** (`--no-network`) | `vNetworkBytesRecvd` | int64 | bytes | `/proc/net/dev` col 0 | Cumulative bytes received across all interfaces |
| | `vNetworkBytesSent` | int64 | bytes | `/proc/net/dev` col 8 | Cumulative bytes transmitted across all interfaces |
| | `vNetworkPacketsRecvd` | int64 | count | `/proc/net/dev` col 1 | Cumulative packets received |
| | `vNetworkPacketsSent` | int64 | count | `/proc/net/dev` col 9 | Cumulative packets transmitted |
| | `vNetworkErrorsRecvd` | int64 | count | `/proc/net/dev` col 2 | Cumulative receive errors |
| | `vNetworkErrorsSent` | int64 | count | `/proc/net/dev` col 10 | Cumulative transmit errors |
| | `vNetworkDropsRecvd` | int64 | count | `/proc/net/dev` col 3 | Cumulative packets dropped on receive |
| | `vNetworkDropsSent` | int64 | count | `/proc/net/dev` col 11 | Cumulative packets dropped on transmit |

**Notes:**
- CPU time values are cumulative since system boot; 1 centisecond = 10 milliseconds
- To calculate utilization over an interval: `(delta_busy / delta_total) × 100`
- Disk filtering: Only physical disks included (`sd*`, `hd*`, `vd*`, `xvd*`, `nvme*n*`, `mmcblk*`)
- Network: Loopback interface (`lo`) excluded; values aggregated across all physical/virtual interfaces

---

## Container-Level Metrics

Prefix: `c` (Container-level). Disable with `--no-container`. Auto-detects cgroup v1 or v2.

| Category | Key | Type | Unit | Source | Description |
|----------|-----|------|------|--------|-------------|
| **CPU** | `cCpuTime` | int64 | nanoseconds | `cpuacct.usage` / `cpu.stat usage_usec` | Total CPU time consumed by all processes in container |
| | `cCpuTimeUserMode` | int64 | centiseconds | `cpuacct.stat` / `cpu.stat user_usec` | CPU time spent in user mode within container |
| | `cCpuTimeKernelMode` | int64 | centiseconds | `cpuacct.stat` / `cpu.stat system_usec` | CPU time spent in kernel mode within container |
| | `cCpuPerCpuJson` | string | - | `cpuacct.usage_percpu` (v1 only) | Per-CPU time breakdown as JSON array (cgroup v1 only) |
| **Memory** | `cMemoryUsed` | int64 | bytes | `memory.usage_in_bytes` / `memory.current` | Current memory usage by container |
| | `cMemoryMaxUsed` | int64 | bytes | `memory.max_usage_in_bytes` / `memory.peak` | Peak memory usage since container start |
| | `cPgFault` | int64 | count | `memory.stat pgfault` | Cumulative page faults within container |
| | `cMajorPgFault` | int64 | count | `memory.stat pgmajfault` | Cumulative major page faults (required I/O) |
| **Disk** | `cDiskReadBytes` | int64 | bytes | `blkio.throttle.io_service_bytes` / `io.stat` | Cumulative bytes read by container |
| | `cDiskWriteBytes` | int64 | bytes | `blkio.throttle.io_service_bytes` / `io.stat` | Cumulative bytes written by container |
| | `cDiskSectorIO` | int64 | sectors | `blkio.sectors` (v1 only) | Cumulative sector I/O (cgroup v1 only) |
| **Network** | `cNetworkBytesRecvd` | int64 | bytes | `/proc/net/dev` | Cumulative network bytes received by container |
| | `cNetworkBytesSent` | int64 | bytes | `/proc/net/dev` | Cumulative network bytes sent by container |
| **Process** | `cNumProcesses` | int64 | count | `pids.current` / tasks file | Number of processes currently running in container |

---

## Process-Level Metrics

Prefix: `process{i}` (flattened) or `processesJson` (JSON mode). Disable with `--no-procs`.

| Category | Key | Type | Unit | Source | Description |
|----------|-----|------|------|--------|-------------|
| **Identity** | `pId` | int64 | - | `/proc/[pid]` | Process identifier |
| | `pName` | string | - | `/proc/[pid]/status` | Executable name (command name only) |
| | `pCmdline` | string | - | `/proc/[pid]/cmdline` | Full command line with arguments |
| | `pNumThreads` | int64 | count | `/proc/[pid]/stat` field 19 | Number of threads in this process |
| **CPU Time** | `pCpuTimeUserMode` | int64 | centiseconds | `/proc/[pid]/stat` field 14 | CPU time spent executing user code |
| | `pCpuTimeKernelMode` | int64 | centiseconds | `/proc/[pid]/stat` field 15 | CPU time spent executing kernel code |
| | `pChildrenUserMode` | int64 | centiseconds | `/proc/[pid]/stat` field 16 | CPU time spent in user mode by dead children |
| | `pChildrenKernelMode` | int64 | centiseconds | `/proc/[pid]/stat` field 17 | CPU time spent in kernel mode by dead children |
| **Context** | `pVoluntaryContextSwitches` | int64 | count | `/proc/[pid]/status` | Cumulative voluntary context switches (process yielded CPU) |
| | `pNonvoluntaryContextSwitches` | int64 | count | `/proc/[pid]/status` | Cumulative involuntary context switches (preempted) |
| | `pBlockIODelays` | int64 | centiseconds | `/proc/[pid]/stat` field 42 | Cumulative time waiting for block I/O |
| **Memory** | `pVirtualMemoryBytes` | int64 | bytes | `/proc/[pid]/stat` field 23 | Virtual memory size (address space reserved) |
| | `pResidentSetSize` | int64 | bytes | `/proc/[pid]/statm` × page size | Physical memory (RAM) currently used by process |

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

Prefix: `nvidia{i}` (flattened) or `nvidiaGpusJson` (JSON mode). Disable with `--no-nvidia`. Disable process enumeration only with `--no-gpu-procs`.

| Category | Key | Type | Unit | Source | Description |
|----------|-----|------|------|--------|-------------|
| **Utilization** | `utilizationGpu` | int64 | percent | `nvmlDeviceGetUtilizationRates().gpu` | GPU compute engine utilization (0-100) |
| | `utilizationMemory` | int64 | percent | `nvmlDeviceGetUtilizationRates().memory` | Memory controller utilization (0-100) |
| | `utilizationEncoder` | int64 | percent | `nvmlDeviceGetEncoderUtilization()` | Video encoder utilization (0-100) |
| | `utilizationDecoder` | int64 | percent | `nvmlDeviceGetDecoderUtilization()` | Video decoder utilization (0-100) |
| | `utilizationJpeg` | int64 | percent | `nvmlDeviceGetJpgUtilization()` | JPEG engine utilization (Turing+ only) |
| | `utilizationOfa` | int64 | percent | `nvmlDeviceGetOfaUtilization()` | Optical Flow Accelerator utilization (Turing+ only) |
| | `encoderSamplingPeriodUs` | int64 | μs | NVML | Encoder utilization sampling period |
| | `decoderSamplingPeriodUs` | int64 | μs | NVML | Decoder utilization sampling period |
| **Memory** | `memoryUsedBytes` | int64 | bytes | `nvmlDeviceGetMemoryInfo().used` | Frame buffer memory currently in use |
| | `memoryFreeBytes` | int64 | bytes | `nvmlDeviceGetMemoryInfo().free` | Frame buffer memory available for allocation |
| | `memoryTotalBytes` | int64 | bytes | `nvmlDeviceGetMemoryInfo().total` | Total frame buffer capacity |
| | `memoryReservedBytes` | int64 | bytes | `nvmlDeviceGetMemoryInfo_v2()` | Memory reserved by driver/firmware (unavailable to applications) |
| | `bar1UsedBytes` | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()` | BAR1 memory in use (GPU-visible system memory for DMA) |
| | `bar1FreeBytes` | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()` | BAR1 memory available |
| | `bar1TotalBytes` | int64 | bytes | `nvmlDeviceGetBAR1MemoryInfo()` | Total BAR1 memory capacity |
| **Temperature** | `temperatureGpuC` | int64 | °C | `nvmlDeviceGetTemperature(GPU)` | GPU die temperature |
| | `temperatureMemoryC` | int64 | °C | `nvmlDeviceGetTemperature(MEMORY)` | HBM memory junction temperature (Ampere+ only) |
| **Fan** | `fanSpeedPercent` | int64 | percent | `nvmlDeviceGetFanSpeed()` | Fan speed (single fan or average of multiple fans) |
| | `fanSpeedsJson` | string | - | `nvmlDeviceGetFanSpeed_v2()` | Individual fan speeds as JSON array (multi-fan GPUs only) |
| **Performance** | `performanceState` | int | - | `nvmlDeviceGetPerformanceState()` | Current P-state (0=maximum performance, 15=minimum) |
| **Clocks** | `clockGraphicsMhz` | int64 | MHz | `nvmlDeviceGetClockInfo(GRAPHICS)` | Current graphics engine clock frequency |
| | `clockSmMhz` | int64 | MHz | `nvmlDeviceGetClockInfo(SM)` | Current SM (streaming multiprocessor) clock frequency |
| | `clockMemoryMhz` | int64 | MHz | `nvmlDeviceGetClockInfo(MEM)` | Current memory interface clock frequency |
| | `clockVideoMhz` | int64 | MHz | `nvmlDeviceGetClockInfo(VIDEO)` | Current video engine clock frequency |
| **Power** | `powerUsageMw` | int64 | milliwatts | `nvmlDeviceGetPowerUsage()` | Instantaneous power consumption |
| | `powerLimitMw` | int64 | milliwatts | `nvmlDeviceGetPowerManagementLimit()` | Current power management limit (configurable) |
| | `powerEnforcedLimitMw` | int64 | milliwatts | `nvmlDeviceGetEnforcedPowerLimit()` | Hardware-enforced power ceiling (cannot be changed) |
| | `energyConsumptionMj` | int64 | millijoules | `nvmlDeviceGetTotalEnergyConsumption()` | Cumulative energy consumed since driver load |
| **PCIe** | `pcieTxBytesPerSec` | int64 | bytes/sec | `nvmlDeviceGetPcieThroughput(TX)` | PCIe transmit throughput (GPU to CPU) |
| | `pcieRxBytesPerSec` | int64 | bytes/sec | `nvmlDeviceGetPcieThroughput(RX)` | PCIe receive throughput (CPU to GPU) |
| | `pcieCurrentLinkGen` | int | - | `nvmlDeviceGetCurrPcieLinkGeneration()` | Active PCIe generation (1-5 for Gen1-Gen5) |
| | `pcieCurrentLinkWidth` | int | - | `nvmlDeviceGetCurrPcieLinkWidth()` | Active PCIe lane count (x1, x4, x8, x16) |
| | `pcieReplayCounter` | int64 | count | `nvmlDeviceGetPcieReplayCounter()` | Cumulative PCIe packet replay errors (link quality indicator) |
| **Throttling** | `clocksEventReasons` | uint64 | bitmask | `nvmlDeviceGetCurrentClocksEventReasons()` | Active clock throttle reasons as NVML bitmask |
| | `violationPowerNs` | int64 | nanoseconds | `nvmlDeviceGetViolationStatus(POWER)` | Cumulative time throttled by power limit since driver load |
| | `violationThermalNs` | int64 | nanoseconds | `nvmlDeviceGetViolationStatus(THERMAL)` | Cumulative time throttled by temperature since driver load |
| | `violationReliabilityNs` | int64 | nanoseconds | `nvmlDeviceGetViolationStatus(RELIABILITY)` | Cumulative time throttled for reliability reasons since driver load |
| | `violationBoardLimitNs` | int64 | nanoseconds | `nvmlDeviceGetViolationStatus(BOARD_LIMIT)` | Cumulative time at board power limit since driver load |
| | `violationLowUtilNs` | int64 | nanoseconds | `nvmlDeviceGetViolationStatus(LOW_UTIL)` | Cumulative time with clocks reduced due to low utilization |
| | `violationSyncBoostNs` | int64 | nanoseconds | `nvmlDeviceGetViolationStatus(SYNC_BOOST)` | Cumulative time in sync boost mode |
| **ECC Errors** | `eccAggregateSbe` | int64 | count | `nvmlDeviceGetTotalEccErrors(CORRECTED, AGGREGATE)` | Lifetime correctable ECC errors (persists across reboots) |
| | `eccAggregateDbe` | int64 | count | `nvmlDeviceGetTotalEccErrors(UNCORRECTED, AGGREGATE)` | Lifetime uncorrectable ECC errors (persists across reboots) |
| **Retired Pages** | `retiredPagesSbe` | int64 | count | `nvmlDeviceGetRetiredPages(SBE)` | Memory pages retired due to excessive correctable errors |
| | `retiredPagesDbe` | int64 | count | `nvmlDeviceGetRetiredPages(DBE)` | Memory pages retired due to uncorrectable errors |
| | `retiredPending` | bool | - | `nvmlDeviceGetRetiredPagesPendingStatus()` | Page retirement pending system reboot |
| **Remapped Rows** | `remappedRowsCorrectable` | int64 | count | `nvmlDeviceGetRemappedRows()` | Memory rows remapped due to correctable errors |
| | `remappedRowsUncorrectable` | int64 | count | `nvmlDeviceGetRemappedRows()` | Memory rows remapped due to uncorrectable errors |
| | `remappedRowsPending` | bool | - | `nvmlDeviceGetRemappedRows()` | Row remapping operation pending |
| | `remappedRowsFailure` | bool | - | `nvmlDeviceGetRemappedRows()` | Row remapping operation failed |
| **NVLink** | `nvlinkBandwidthJson` | string | - | NVML | Per-link TX/RX bandwidth as JSON array of `{link, txBytes, rxBytes}` |
| | `nvlinkErrorsJson` | string | - | NVML | Per-link error counters as JSON array of `{link, crcErrors, eccErrors, replayErrors, recoveryCount}` |

### GPU Processes

| Key | Type | Description |
|-----|------|-------------|
| `processCount` | int64 | Number of processes actively using this GPU (compute/graphics/MPS) |
| `processesJson` | string | Process details as JSON array: `[{pid, name, usedMemoryBytes, type}]` |
| `processUtilizationJson` | string | Per-process utilization samples: `[{pid, smUtil, memUtil, encUtil, decUtil, timestampUs}]` |

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

Prefix: `vllm`. Disable with `--no-vllm`. Configure endpoint via `VLLM_METRICS_URL` environment variable (default: `http://localhost:8000/metrics`).

### System State and Cache

| Category | Key | Type | Unit | Source | Description |
|----------|-----|------|------|--------|-------------|
| **Availability** | `vllmAvailable` | bool | - | HTTP response | Whether vLLM Prometheus endpoint is reachable |
| | `vllmTimestamp` | int64 | nanoseconds | Scrape time | When metrics were collected from vLLM |
| **State** | `vllmRequestsRunning` | float64 | count | `vllm:num_requests_running` | Requests currently being processed |
| | `vllmRequestsWaiting` | float64 | count | `vllm:num_requests_waiting` | Requests waiting in queue |
| | `vllmEngineSleepState` | float64 | - | `vllm:engine_sleep_state` | Engine sleep state (0=active, 1=sleeping) |
| | `vllmPreemptionsTotal` | float64 | count | `vllm:num_preemptions` | Cumulative requests preempted to free memory |
| **Cache** | `vllmKvCacheUsagePercent` | float64 | ratio | `vllm:kv_cache_usage_perc` | KV-cache utilization (0.0-1.0) |
| | `vllmPrefixCacheHits` | float64 | count | `vllm:prefix_cache_hits` | Cumulative prefix cache hits |
| | `vllmPrefixCacheQueries` | float64 | count | `vllm:prefix_cache_queries` | Cumulative prefix cache lookups |
| **Requests** | `vllmRequestsFinishedTotal` | float64 | count | `vllm:request_success` | Cumulative successfully completed requests |
| | `vllmRequestsCorruptedTotal` | float64 | count | `vllm:corrupted_requests` | Cumulative failed/corrupted requests |
| **Tokens** | `vllmTokensPromptTotal` | float64 | count | `vllm:prompt_tokens` | Cumulative prompt tokens processed |
| | `vllmTokensGenerationTotal` | float64 | count | `vllm:generation_tokens` | Cumulative tokens generated |

### Latency Metrics (Sums and Counts)

| Metric | Sum Key | Count Key | Source Prefix | Description |
|--------|---------|-----------|---------------|-------------|
| **TTFT** | `vllmLatencyTtftSum` | `vllmLatencyTtftCount` | `vllm:time_to_first_token_seconds` | Time to first token (prompt processing + first token generation) |
| **E2E** | `vllmLatencyE2eSum` | `vllmLatencyE2eCount` | `vllm:e2e_request_latency_seconds` | End-to-end request latency (queue + inference) |
| **Queue** | `vllmLatencyQueueSum` | `vllmLatencyQueueCount` | `vllm:request_queue_time_seconds` | Time spent waiting in queue |
| **Inference** | `vllmLatencyInferenceSum` | `vllmLatencyInferenceCount` | `vllm:request_inference_time_seconds` | Total inference time (prefill + decode) |
| **Prefill** | `vllmLatencyPrefillSum` | `vllmLatencyPrefillCount` | `vllm:request_prefill_time_seconds` | Prompt processing time |
| **Decode** | `vllmLatencyDecodeSum` | `vllmLatencyDecodeCount` | `vllm:request_decode_time_seconds` | Token generation time |

All latency sums are in `seconds` (float64), counts are dimensionless (float64). Calculate average as: `sum / count`.

### Histograms

Histogram bucket data is stored as JSON in `vllmHistogramsJson`:

```json
{
  "vllmHistogramsJson": "{\"latencyTtft\":{\"0.001\":0,\"0.005\":12,\"0.01\":45,\"inf\":300},...}"
}
```

Available histogram fields:
- **Latency**: `latencyTtft`, `latencyE2e`, `latencyQueue`, `latencyInference`, `latencyPrefill`, `latencyDecode`, `latencyInterToken`
- **Request Size**: `reqSizePromptTokens`, `reqSizeGenerationTokens`
- **Throughput**: `tokensPerStep`, `reqParamsMaxTokens`, `reqParamsN`

Histogram buckets represent cumulative counts at each threshold. Subtract consecutive buckets to get per-bucket counts.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VLLM_METRICS_URL` | `http://localhost:8000/metrics` | vLLM Prometheus metrics endpoint URL |
