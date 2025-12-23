# Profiler Metrics Reference

## CPU Metrics

| Name | JSON Key | Units | Filepath | Description |
|------|----------|-------|----------|-------------|
| User Mode Time | `vCpuTimeUserMode` | jiffies | `/proc/stat` | Time spent in user mode |
| Kernel Mode Time | `vCpuTimeKernelMode` | jiffies | `/proc/stat` | Time spent in kernel mode |
| Idle Time | `vCpuIdleTime` | jiffies | `/proc/stat` | Time CPU was idle |
| I/O Wait Time | `vCpuTimeIOWait` | jiffies | `/proc/stat` | Time waiting for I/O operations |
| Interrupt Service Time | `vCpuTimeIntSrvc` | jiffies | `/proc/stat` | Time in hardware interrupt handling |
| Soft Interrupt Service Time | `vCpuTimeSoftIntSrvc` | jiffies | `/proc/stat` | Time in software interrupt handling |
| Nice Time | `vCpuNice` | jiffies | `/proc/stat` | Time spent by low-priority processes |
| Steal Time | `vCpuSteal` | jiffies | `/proc/stat` | Time stolen by hypervisor (virtualization) |
| Context Switches | `vCpuContextSwitches` | switches | `/proc/stat` | Total context switches since boot |
| Load Average (1-min) | `vLoadAvg` | load units | `/proc/loadavg` | 1-minute average system load |
| CPU Frequency | `vCpuMhz` | MHz | `/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq` or `/proc/cpuinfo` | Current CPU clock speed |
| CPU Count | `num_processors` | count | system call | Number of logical CPU cores |
| CPU Type | `cpu_type` | string | `/proc/cpuinfo` | CPU model name |
| CPU Cache | `cpu_cache` | bytes | `/sys/devices/system/cpu/cpu*/cache/index*` | L1/L2/L3 cache sizes per level |
| Kernel Info | `kernel_info` | string | `os.uname()` | Kernel version and system info |

## Memory Metrics

| Name | JSON Key | Units | Filepath | Description |
|------|----------|-------|----------|-------------|
| Total Memory | `vMemoryTotal` | bytes | `/proc/meminfo` | Total system memory |
| Free Memory | `vMemoryFree` | bytes | `/proc/meminfo` | Available memory (free + buffers + cached) |
| Used Memory | `vMemoryUsed` | bytes | `/proc/meminfo` | Allocated memory (total - free - buffers - cached) |
| Memory Buffers | `vMemoryBuffers` | bytes | `/proc/meminfo` | Memory used for buffers |
| Memory Cached | `vMemoryCached` | bytes | `/proc/meminfo` | Memory used for page cache and slabs |
| Memory Utilization | `vMemoryPercent` | percent | `/proc/meminfo` | Memory usage percentage |
| Page Faults | `vPgFault` | count | `/proc/vmstat` | Total page faults (minor) |
| Major Page Faults | `vMajorPageFault` | count | `/proc/vmstat` | Major page faults (disk access required) |
| Memory Total (static) | `memory_total_bytes` | bytes | `/proc/meminfo` | Static: total system memory at startup |

## Disk I/O Metrics

| Name | JSON Key | Units | Filepath | Description |
|------|----------|-------|----------|-------------|
| Read Bytes | `read_bytes` | bytes | `/proc/diskstats` | Total bytes read from disk |
| Write Bytes | `write_bytes` | bytes | `/proc/diskstats` | Total bytes written to disk |
| Read Count | `read_count` | count | `/proc/diskstats` | Total number of read operations |
| Write Count | `write_count` | count | `/proc/diskstats` | Total number of write operations |

## Network Metrics

| Name | JSON Key | Units | Filepath | Description |
|------|----------|-------|----------|-------------|
| Bytes Received | `vNetworkBytesRecvd` | bytes | `/proc/net/dev` | Total network bytes received |
| Bytes Sent | `vNetworkBytesSent` | bytes | `/proc/net/dev` | Total network bytes transmitted |

## Container (cgroup) Metrics

| Name | JSON Key | Units | Filepath | Description |
|------|----------|-------|----------|-------------|
| CPU Usage (v1) | `cpu_usage_ns` | nanoseconds | `/sys/fs/cgroup/cpuacct/cpuacct.usage` | Container CPU time (cgroup v1) |
| Memory Used (v1) | `memory_used_bytes` | bytes | `/sys/fs/cgroup/memory/memory.usage_in_bytes` | Container memory (cgroup v1) |
| CPU Usage (v2) | `cpu_usage_ns` | nanoseconds | `/sys/fs/cgroup/cpu.stat` | Container CPU time (cgroup v2) |
| Memory Used (v2) | `memory_used_bytes` | bytes | `/sys/fs/cgroup/memory.current` | Container memory (cgroup v2) |
| cgroup Version | `cgroup_version` | integer | `/sys/fs/cgroup/` | Detected cgroup version (1 or 2) |

## GPU (NVIDIA) Metrics

| Name | JSON Key | Units | Filepath/Source | Description |
|------|----------|-------|----------|-------------|
| GPU Utilization | `utilization_gpu` | percent | `nvmlDeviceGetUtilizationRates()` | GPU compute utilization |
| Memory Utilization | `utilization_mem` | percent | `nvmlDeviceGetUtilizationRates()` | GPU memory utilization |
| Total Memory | `memory_total_mb` | MB | `nvmlDeviceGetMemoryInfo()` | Total GPU VRAM |
| Used Memory | `memory_used_mb` | MB | `nvmlDeviceGetMemoryInfo()` | Allocated GPU VRAM |
| Free Memory | `memory_free_mb` | MB | `nvmlDeviceGetMemoryInfo()` | Available GPU VRAM |
| BAR1 Used | `bar1_used_mb` | MB | `nvmlDeviceGetBAR1MemoryInfo()` | PCIe BAR1 memory used |
| BAR1 Free | `bar1_free_mb` | MB | `nvmlDeviceGetBAR1MemoryInfo()` | PCIe BAR1 memory free |
| Temperature | `temperature_c` | °C | `nvmlDeviceGetTemperature()` | GPU die temperature |
| Fan Speed | `fan_speed` | percent | `nvmlDeviceGetFanSpeed()` | GPU cooling fan speed |
| Power Draw | `power_draw_w` | watts | `nvmlDeviceGetPowerUsage()` | Current GPU power consumption |
| Power Limit | `power_limit_w` | watts | `nvmlDeviceGetEnforcedPowerLimit()` | GPU power limit |
| Graphics Clock | `clock_graphics_mhz` | MHz | `nvmlDeviceGetClockInfo()` | GPU graphics engine clock |
| SM Clock | `clock_sm_mhz` | MHz | `nvmlDeviceGetClockInfo()` | Streaming Multiprocessor clock |
| Memory Clock | `clock_mem_mhz` | MHz | `nvmlDeviceGetClockInfo()` | GPU memory clock |
| PCIe TX | `pcie_tx_kbps` | KB/s | `nvmlDeviceGetPcieThroughput()` | PCIe transmit throughput |
| PCIe RX | `pcie_rx_kbps` | KB/s | `nvmlDeviceGetPcieThroughput()` | PCIe receive throughput |
| Performance State | `perf_state` | string | `nvmlDeviceGetPerformanceState()` | Performance state (P0-P15) |
| Process Count | `process_count` | count | `nvmlDeviceGetComputeRunningProcesses()` | Number of processes using GPU |
| GPU Index | `index` | integer | device enumeration | GPU device index |
| GPU Name (static) | `name` | string | `nvmlDeviceGetName()` | GPU model name |
| GPU UUID (static) | `uuid` | string | `nvmlDeviceGetUUID()` | Globally unique GPU identifier |
| Total Memory (static) | `total_memory_mb` | MB | `nvmlDeviceGetMemoryInfo()` | Static: total GPU VRAM |
| PCI Bus ID (static) | `pci_bus_id` | string | `nvmlDeviceGetPciInfo()` | GPU PCIe bus location |
| Max Graphics Clock (static) | `max_graphics_clock` | MHz | `nvmlDeviceGetMaxClockInfo()` | Maximum graphics clock supported |
| Max SM Clock (static) | `max_sm_clock` | MHz | `nvmlDeviceGetMaxClockInfo()` | Maximum SM clock supported |
| Max Memory Clock (static) | `max_mem_clock` | MHz | `nvmlDeviceGetMaxClockInfo()` | Maximum memory clock supported |

## Process Metrics

| Name | JSON Key | Units | Filepath | Description |
|------|----------|-------|----------|-------------|
| Process ID | `pId` | integer | `/proc/[pid]/` | Process identifier |
| Command Line | `pCmdline` | string | `/proc/[pid]/cmdline` | Process command and arguments |
| Process Name | `pName` | string | `/proc/[pid]/status` | Process executable name |
| Thread Count | `pNumThreads` | count | `/proc/[pid]/stat` | Number of threads in process |
| User Mode CPU Time | `pCpuTimeUserMode` | jiffies | `/proc/[pid]/stat` | CPU time in user mode |
| Kernel Mode CPU Time | `pCpuTimeKernelMode` | jiffies | `/proc/[pid]/stat` | CPU time in kernel mode |
| Children User Mode CPU Time | `pChildrenUserMode` | jiffies | `/proc/[pid]/stat` | CPU time of child processes (user mode) |
| Children Kernel Mode CPU Time | `pChildrenKernelMode` | jiffies | `/proc/[pid]/stat` | CPU time of child processes (kernel mode) |
| Voluntary Context Switches | `pVoluntaryContextSwitches` | count | `/proc/[pid]/status` | Context switches initiated by process |
| Involuntary Context Switches | `pInvoluntaryContextSwitches` | count | `/proc/[pid]/status` | Context switches forced by scheduler |
| Block I/O Delays | `pBlockIODelays` | jiffies | `/proc/[pid]/stat` | Time blocked on I/O |
| Virtual Memory | `pVirtualMemoryBytes` | bytes | `/proc/[pid]/stat` | Virtual memory address space size |

## System Static Metadata

| Name | JSON Key | Units | Source | Description |
|------|----------|-------|--------|-------------|
| Session UUID | `uuid` | string | generated | Unique profiling session identifier |
| Hostname | `hostname` | string | `os.uname()` | System hostname |
| Kernel | `kernel` | string | `os.uname()` | Full kernel information string |
| Boot Time | `boot_time` | unix timestamp | `/proc/stat` | System boot timestamp |
| Driver Version | `driver_version` | string | `nvmlSystemGetDriverVersion()` | NVIDIA driver version |
| CUDA Version | `cuda_version` | string | `nvmlSystemGetCudaDriverVersion()` | CUDA runtime version |
| Collection Timestamp | `timestamp` | float | `time.time()` | Snapshot collection timestamp |
| Metric Timestamp (tv_*) | `tv_<metric>` | float | `time.time()` | Precise timestamp for individual metric |

## Notes

- **Jiffies**: Linux kernel timer ticks (100 per second by default, configurable via `JIFFIES_PER_SECOND`)
- **Timestamps**: Each metric includes both the value and a corresponding `tv_` timestamp for synchronization
- **GPU Metrics**: Only available if NVIDIA drivers and pynvml are installed
- **Container Metrics**: Only available inside a container; detects cgroup v1 or v2
- **Process Metrics**: Collected for all running processes; can be memory-intensive on systems with many processes

---

## Coverage Analysis vs. Container Profiler Paper (Hoang et al., 2023)

### What the Paper Mentions but Code Doesn't Implement

**Host-Level Gaps (Table 1 in paper):**
- ❌ `vDiskReadTime`, `vDiskWriteTime` - time spent reading/writing disk (code collects sector counts but not time)

**Container-Level Gaps (Table 2 in paper):**
- ❌ `cDiskSectorIO`, `cDiskReadBytes`, `cDiskWriteBytes` - container-level disk I/O via cgroup blkio subsystem
- ❌ `cNetworkBytesRecvd`, `cNetworkBytesSent` - per-interface network stats for containers

**Process-Level Gaps (Table 3 in paper):**
- ❌ `pResidentSetSize` (RSS) - resident set size in physical memory

**Static Metadata Gaps:**
- ❌ Host VM ID
- ❌ Container ID