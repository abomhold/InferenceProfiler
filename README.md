# Profiler Metrics Reference

The profiler aggregates metrics from multiple collectors. Below is the complete list of all metrics organized by device type, with their JSON keys, units, sources, and descriptions.

All dynamic metrics include a timestamp field prefixed with `t` (e.g., `tvCpuTimeUserMode` is the timestamp for `vCpuTimeUserMode`).

---

## VM-Level Metrics

### CPU Metrics

*Source: `/proc/stat`, `/proc/loadavg`, `/sys/devices/system/cpu/*`

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **CPU Time (Aggregate)** |  |  |  |  |
| Total CPU Time | `vCpuTime` | cs | File: `/proc/stat`Logic: Sum of `user` + `system` fields on `cpu` line. | Total CPU time (user + kernel mode). 1 cs = 10ms. |
| User Mode Time | `vCpuTimeUserMode` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 1 (`user`). | Time spent executing user-space processes. |
| Kernel Mode Time | `vCpuTimeKernelMode` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 3 (`system`). | Time spent executing kernel code on behalf of processes. |
| Idle Time | `vCpuIdleTime` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 4 (`idle`). | Time CPU spent in idle state (no work). |
| I/O Wait Time | `vCpuTimeIOWait` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 5 (`iowait`). | Time CPU spent waiting for I/O operations to complete. |
| IRQ Time | `vCpuTimeIntSrvc` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 6 (`irq`). | Time spent servicing hardware interrupts. |
| Soft IRQ Time | `vCpuTimeSoftIntSrvc` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 7 (`softirq`). | Time spent servicing software interrupts (deferred work). |
| Nice Time | `vCpuNice` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 2 (`nice`). | Time spent running niced (low priority) user processes. |
| Steal Time | `vCpuSteal` | cs | File: `/proc/stat`<br>

<br>Field: `cpu` line, col 8 (`steal`). | Time stolen by hypervisor for other VMs (virtualized environments). |
| **Performance Counters** |  |  |  |  |
| Context Switches | `vCpuContextSwitches` | count | File: `/proc/stat`<br>

<br>Field: `ctxt` line. | Total context switches across all CPUs since boot. |
| Load Average | `vLoadAvg` | ratio | File: `/proc/loadavg`<br>

<br>Field: 1st column (1-min avg). | 1-minute system load average (runnable + uninterruptible processes). |
| CPU Frequency | `vCpuMhz` | MHz | File: `/sys/devices/system/cpu/*/cpufreq/scaling_cur_freq`<br>

<br>Logic: Average of all cores / 1000.<br>

<br>Fallback: `/proc/cpuinfo` (`cpu MHz`). | Current average CPU frequency across all cores. |

**Static Info:**

| Field | JSON Key | Detailed Source | Description |
| --- | --- | --- | --- |
| Processor Count | `vNumProcessors` | Python: `os.cpu_count()` | Number of logical CPUs/cores. |
| CPU Model | `vCpuType` | File: `/proc/cpuinfo`<br>

<br>Field: `model name`. | Processor model name string. |
| Cache Sizes | `vCpuCache` | File: `/sys/devices/system/cpu/cpu*/cache/index*`<br>

<br>Files used: `level`, `type`, `size`, `shared_cpu_map`. | Object with L1d, L1i, L2, L3 cache sizes in bytes. |
| Kernel Info | `vKernelInfo` | Python: `os.uname()` | Full kernel version string. |

---

### Memory Metrics

*Source: `/proc/meminfo`, `/proc/vmstat*`

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **Physical Memory** |  |  |  |  |
| Total Memory | `vMemoryTotal` | bytes | File: `/proc/meminfo`<br>

<br>Field: `MemTotal` (converted kB to B). | Total physical RAM installed. |
| Free Memory | `vMemoryFree` | bytes | File: `/proc/meminfo`<br>

<br>Field: `MemAvailable` (fallback: `MemFree` + `Buffers` + `Cached`). | Available memory (includes reclaimable buffers/cache). |
| Used Memory | `vMemoryUsed` | bytes | Derived: `vMemoryTotal` - `vMemoryFree` - `vMemoryBuffers` - `vMemoryCached`. | Actively used memory (Total - Free - Buffers - Cached). |
| Buffer Memory | `vMemoryBuffers` | bytes | File: `/proc/meminfo`<br>

<br>Field: `Buffers`. | Memory used by kernel buffers for block device I/O. |
| Cached Memory | `vMemoryCached` | bytes | File: `/proc/meminfo`<br>

<br>Logic: `Cached` + `SReclaimable`. | Memory used by page cache (file data cache). |
| Memory Percent | `vMemoryPercent` | % | Derived: `(Total - Available) / Total * 100`. | Percentage of total RAM currently in use. |
| **Swap Space** |  |  |  |  |
| Swap Total | `vSwapTotal` | bytes | File: `/proc/meminfo`<br>

<br>Field: `SwapTotal`. | Total swap space available. |
| Swap Free | `vSwapFree` | bytes | File: `/proc/meminfo`<br>

<br>Field: `SwapFree`. | Unused swap space. |
| Swap Used | `vSwapUsed` | bytes | Derived: `SwapTotal` - `SwapFree`. | Currently used swap space. |
| **Page Faults** |  |  |  |  |
| Page Faults | `vPgFault` | count | File: `/proc/vmstat`<br>

<br>Field: `pgfault`. | Total minor page faults (page in memory but not mapped). |
| Major Page Faults | `vMajorPageFault` | count | File: `/proc/vmstat`<br>

<br>Field: `pgmajfault`. | Major page faults requiring disk I/O to resolve. |

**Static Info:**

| Field | JSON Key | Detailed Source | Description |
| --- | --- | --- | --- |
| Memory Total | `vMemoryTotalBytes` | File: `/proc/meminfo`<br>

<br>Field: `MemTotal`. | Total physical RAM in bytes. |
| Swap Total | `vSwapTotalBytes` | File: `/proc/meminfo`<br>

<br>Field: `SwapTotal`. | Total swap space in bytes. |

---

### Disk I/O Metrics

*Source: `/proc/diskstats` (aggregated across physical disks)*

**Aggregation Logic:** Iterates over `/proc/diskstats`. Includes lines matching regex `^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$`. Sums values across all matched devices.

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **Sector Counts** |  |  |  |  |
| Sectors Read | `vDiskSectorReads` | sectors | File: `/proc/diskstats`<br>

<br>Column: 5. | Total sectors read (1 sector = 512 bytes typically). |
| Sectors Written | `vDiskSectorWrites` | sectors | File: `/proc/diskstats`<br>

<br>Column: 9. | Total sectors written. |
| **Byte Counts** |  |  |  |  |
| Bytes Read | `vDiskReadBytes` | bytes | Derived: `vDiskSectorReads` * 512. | Total bytes read from all physical disks. |
| Bytes Written | `vDiskWriteBytes` | bytes | Derived: `vDiskSectorWrites` * 512. | Total bytes written to all physical disks. |
| **Operation Counts** |  |  |  |  |
| Successful Reads | `vDiskSuccessfulReads` | count | File: `/proc/diskstats`<br>

<br>Column: 3. | Number of read operations completed successfully. |
| Successful Writes | `vDiskSuccessfulWrites` | count | File: `/proc/diskstats`<br>

<br>Column: 7. | Number of write operations completed successfully. |
| Merged Reads | `vDiskMergedReads` | count | File: `/proc/diskstats`<br>

<br>Column: 4. | Adjacent read requests merged for efficiency. |
| Merged Writes | `vDiskMergedWrites` | count | File: `/proc/diskstats`<br>

<br>Column: 8. | Adjacent write requests merged for efficiency. |
| **Timing** |  |  |  |  |
| Read Time | `vDiskReadTime` | ms | File: `/proc/diskstats`<br>

<br>Column: 6. | Total time spent on read operations. |
| Write Time | `vDiskWriteTime` | ms | File: `/proc/diskstats`<br>

<br>Column: 10. | Total time spent on write operations. |
| I/O Time | `vDiskIOTime` | ms | File: `/proc/diskstats`<br>

<br>Column: 12. | Total time spent doing I/O (wall clock). |
| Weighted I/O Time | `vDiskWeightedIOTime` | ms | File: `/proc/diskstats`<br>

<br>Column: 13. | I/O time weighted by number of pending operations. |
| **Queue** |  |  |  |  |
| I/Os in Progress | `vDiskIOInProgress` | count | File: `/proc/diskstats`<br>

<br>Column: 11. | Number of I/O operations currently in flight. |

---

### Network Metrics

*Source: `/proc/net/dev` (aggregated across all non-loopback interfaces)*

**Aggregation Logic:** Iterates over `/proc/net/dev`. Skips first 2 header lines. Skips interface `lo`. Sums columns for all other interfaces.

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **Byte Counters** |  |  |  |  |
| Bytes Received | `vNetworkBytesRecvd` | bytes | File: `/proc/net/dev`<br>

<br>Column: 0 (Receive bytes). | Cumulative bytes received on all interfaces. |
| Bytes Sent | `vNetworkBytesSent` | bytes | File: `/proc/net/dev`<br>

<br>Column: 8 (Transmit bytes). | Cumulative bytes transmitted on all interfaces. |
| **Packet Counters** |  |  |  |  |
| Packets Received | `vNetworkPacketsRecvd` | count | File: `/proc/net/dev`<br>

<br>Column: 1. | Total packets received. |
| Packets Sent | `vNetworkPacketsSent` | count | File: `/proc/net/dev`<br>

<br>Column: 9. | Total packets transmitted. |
| **Error Counters** |  |  |  |  |
| Receive Errors | `vNetworkErrorsRecvd` | count | File: `/proc/net/dev`<br>

<br>Column: 2. | Receive errors (CRC, framing, etc.). |
| Send Errors | `vNetworkErrorsSent` | count | File: `/proc/net/dev`<br>

<br>Column: 10. | Transmit errors. |
| **Drop Counters** |  |  |  |  |
| Receive Drops | `vNetworkDropsRecvd` | count | File: `/proc/net/dev`<br>

<br>Column: 3. | Packets dropped on receive (buffer overflow). |
| Send Drops | `vNetworkDropsSent` | count | File: `/proc/net/dev`<br>

<br>Column: 11. | Packets dropped on transmit. |

---

### VM Identification

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| VM ID | `vId` | string | 1. `/sys/class/dmi/id/product_uuid`<br>

<br>2. Fallback: `/etc/machine-id` | VM/instance identifier. |
| Current Time | `currentTime` | seconds | Python: `int(time.time())` | Unix epoch timestamp (seconds since 1970-01-01 00:00:00 UTC). |

---

## Container-Level Metrics (cgroups)

*Source: `/sys/fs/cgroup` (Auto-detects v1 or v2)*

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **Identification** |  |  |  |  |
| Container ID | `cId` | string | File: `/proc/self/cgroup` (parses for `/docker/` or `/kubepods/`)<br>

<br>Fallback: `os.uname().nodename` | Container ID (Docker/Kubernetes short ID or "unavailable"). |
| Cgroup Version | `cCgroupVersion` | int | Logic: Check existence of `/sys/fs/cgroup/cgroup.controllers`. | Detected cgroup version (1 or 2). |
| **CPU Metrics** |  |  |  |  |
| Total CPU Time | `cCpuTime` | ns | v1: `cpuacct/cpuacct.usage`<br>

<br>v2: `cpu.stat` (key `usage_usec` * 1000) | Total CPU time consumed by all tasks in cgroup. |
| User Mode Time | `cCpuTimeUserMode` | cs | v1: `cpuacct/cpuacct.stat` (key `user`)<br>

<br>v2: `cpu.stat` (key `user_usec` / 10000) | CPU time in user mode. |
| Kernel Mode Time | `cCpuTimeKernelMode` | cs | v1: `cpuacct/cpuacct.stat` (key `system`)<br>

<br>v2: `cpu.stat` (key `system_usec` / 10000) | CPU time in kernel mode. |
| Processor Count | `cNumProcessors` | count | Python: `os.cpu_count()` | Number of CPU processors available to container. |
| Per-CPU Time | `cCpu{i}Time` | ns | v1: `cpuacct/cpuacct.usage_percpu`<br>

<br>v2: Not supported | CPU time on processor i (cgroup v1 only). |
| **Memory Metrics** |  |  |  |  |
| Memory Used | `cMemoryUsed` | bytes | v1: `memory/memory.usage_in_bytes`<br>

<br>v2: `memory.current` | Current memory usage by processes in cgroup. |
| Memory Max Used | `cMemoryMaxUsed` | bytes | v1: `memory/memory.max_usage_in_bytes`<br>

<br>v2: `memory.peak` | Peak memory usage (high watermark). |
| **Disk I/O** |  |  |  |  |
| Disk Read Bytes | `cDiskReadBytes` | bytes | v1: `blkio/blkio.throttle.io_service_bytes` (sum 'Read')<br>

<br>v2: `io.stat` (sum `rbytes`) | Bytes read by cgroup from block devices. |
| Disk Write Bytes | `cDiskWriteBytes` | bytes | v1: `blkio/blkio.throttle.io_service_bytes` (sum 'Write')<br>

<br>v2: `io.stat` (sum `wbytes`) | Bytes written by cgroup to block devices. |
| **Network** |  |  |  |  |
| Network Bytes Received | `cNetworkBytesRecvd` | bytes | File: `/proc/net/dev`<br>

<br>(Read from inside container namespace) | Bytes received (from container's network namespace). |
| Network Bytes Sent | `cNetworkBytesSent` | bytes | File: `/proc/net/dev`<br>

<br>(Read from inside container namespace) | Bytes transmitted (from container's network namespace). |

---

## Process-Level Metrics

*Source: `/proc/[pid]/stat`, `/proc/[pid]/status`, `/proc/[pid]/statm*`

**Note:** The profiler iterates through all numeric directories in `/proc/` (`glob.glob('/proc/[0-9]*')`).

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **Identification** |  |  |  |  |
| Process ID | `pId` | int | Directory name in `/proc/` | Process ID (PID). |
| Process Name | `pName` | string | File: `/proc/[pid]/status` (key `Name`) or `/proc/[pid]/stat` (parens) | Executable name. |
| Command Line | `pCmdline` | string | File: `/proc/[pid]/cmdline` | Full command line with arguments. |
| Process Count | `pNumProcesses` | count | Counter in `ProcCollector.collect` | Total number of processes collected. |
| **Threading** |  |  |  |  |
| Thread Count | `pNumThreads` | count | File: `/proc/[pid]/stat`<br>

<br>Field index: 19 (0-indexed base: 18) | Number of threads in this process. |
| **CPU Time** |  |  |  |  |
| User Mode Time | `pCpuTimeUserMode` | cs | File: `/proc/[pid]/stat`<br>

<br>Field index: 13 | CPU time scheduled in user mode. |
| Kernel Mode Time | `pCpuTimeKernelMode` | cs | File: `/proc/[pid]/stat`<br>

<br>Field index: 14 | CPU time scheduled in kernel mode. |
| Children User Time | `pChildrenUserMode` | cs | File: `/proc/[pid]/stat`<br>

<br>Field index: 15 | Cumulative user time of waited-for children. |
| Children Kernel Time | `pChildrenKernelMode` | cs | File: `/proc/[pid]/stat`<br>

<br>Field index: 16 | Cumulative kernel time of waited-for children. |
| **Context Switches** |  |  |  |  |
| Voluntary Switches | `pVoluntaryContextSwitches` | count | File: `/proc/[pid]/status`<br>

<br>Key: `voluntary_ctxt_switches` | Context switches due to blocking (I/O, sleep). |
| Involuntary Switches | `pNonvoluntaryContextSwitches` | count | File: `/proc/[pid]/status`<br>

<br>Key: `nonvoluntary_ctxt_switches` | Context switches due to preemption. |
| **I/O** |  |  |  |  |
| Block I/O Delays | `pBlockIODelays` | cs | File: `/proc/[pid]/stat`<br>

<br>Field index: 41 | Aggregated time waiting for block I/O. |
| **Memory** |  |  |  |  |
| Virtual Memory | `pVirtualMemoryBytes` | bytes | File: `/proc/[pid]/stat`<br>

<br>Field index: 22 | Total virtual address space size. |
| Resident Set Size | `pResidentSetSize` | bytes | File: `/proc/[pid]/statm`<br>

<br>Field: 1 (pages) * Page Size (4096) | Physical memory currently mapped (RSS). |

---

## GPU Metrics (NVIDIA)

*Source: NVML (via `pynvml` library)*

| Metric Name | JSON Key | Units | Detailed Source | Description |
| --- | --- | --- | --- | --- |
| **Identification** |  |  |  |  |
| GPU Index | `gGpuIndex` | int | Loop index in `pynvml.nvmlDeviceGetCount()` | Zero-based GPU index. |
| **Utilization** |  |  |  |  |
| GPU Utilization | `gUtilizationGpu` | % | API: `nvmlDeviceGetUtilizationRates().gpu` | Percent of time GPU was executing kernels. |
| Memory Utilization | `gUtilizationMem` | % | API: `nvmlDeviceGetUtilizationRates().memory` | Percent of time memory was being read/written. |
| **Memory** |  |  |  |  |
| Total Memory | `gMemoryTotalMb` | MB | API: `nvmlDeviceGetMemoryInfo().total` | Total installed frame buffer memory. |
| Used Memory | `gMemoryUsedMb` | MB | API: `nvmlDeviceGetMemoryInfo().used` | Currently allocated frame buffer memory. |
| Free Memory | `gMemoryFreeMb` | MB | API: `nvmlDeviceGetMemoryInfo().free` | Available frame buffer memory. |
| BAR1 Used | `gBar1UsedMb` | MB | API: `nvmlDeviceGetBAR1MemoryInfo().bar1Used` | Used PCIe BAR1 memory. |
| BAR1 Free | `gBar1FreeMb` | MB | API: `nvmlDeviceGetBAR1MemoryInfo().bar1Free` | Free PCIe BAR1 memory. |
| **Thermal & Power** |  |  |  |  |
| Temperature | `gTemperatureC` | °C | API: `nvmlDeviceGetTemperature` | GPU core temperature. |
| Fan Speed | `gFanSpeed` | % | API: `nvmlDeviceGetFanSpeed` | Fan speed percentage. |
| Power Draw | `gPowerDrawW` | W | API: `nvmlDeviceGetPowerUsage` (mW / 1000) | Current power consumption. |
| Power Limit | `gPowerLimitW` | W | API: `nvmlDeviceGetEnforcedPowerLimit` (mW / 1000) | Enforced power limit. |
| **Clock Speeds** |  |  |  |  |
| Graphics Clock | `gClockGraphicsMhz` | MHz | API: `nvmlDeviceGetClockInfo(NVML_CLOCK_GRAPHICS)` | Current shader/graphics engine frequency. |
| SM Clock | `gClockSmMhz` | MHz | API: `nvmlDeviceGetClockInfo(NVML_CLOCK_SM)` | Current Streaming Multiprocessor frequency. |
| Memory Clock | `gClockMemMhz` | MHz | API: `nvmlDeviceGetClockInfo(NVML_CLOCK_MEM)` | Current memory bus frequency. |
| **PCIe Throughput** |  |  |  |  |
| PCIe TX | `gPcieTxKbps` | KB/s | API: `nvmlDeviceGetPcieThroughput(TX_BYTES)` | PCIe transmit throughput (Host → GPU). |
| PCIe RX | `gPcieRxKbps` | KB/s | API: `nvmlDeviceGetPcieThroughput(RX_BYTES)` | PCIe receive throughput (GPU → Host). |
| **State & Errors** |  |  |  |  |
| Performance State | `gPerfState` | string | API: `nvmlDeviceGetPerformanceState` | Power state (P0=max, P12=min). |
| ECC Single-Bit | `gEccSingleBitErrors` | count | API: `nvmlDeviceGetTotalEccErrors(SINGLE_BIT, VOLATILE)` | Correctable single-bit ECC errors. |
| ECC Double-Bit | `gEccDoubleBitErrors` | count | API: `nvmlDeviceGetTotalEccErrors(DOUBLE_BIT, VOLATILE)` | Uncorrectable double-bit ECC errors. |
| **Processes** |  |  |  |  |
| Process Count | `gProcessCount` | count | List len from `nvmlDeviceGetComputeRunningProcesses` | Number of processes using this GPU. |
| Process List | `gProcesses` | list | List content + `/proc/[pid]/comm` | List of processes: "PID: name (VRAM MB)". |

**Static Info:**

| Field | JSON Key | Detailed Source | Description |
| --- | --- | --- | --- |
| Driver Version | `gDriverVersion` | API: `nvmlSystemGetDriverVersion` | NVIDIA driver version string. |
| CUDA Version | `gCudaVersion` | API: `nvmlSystemGetCudaDriverVersion` | CUDA driver version. |
| GPU Name | `gName` | API: `nvmlDeviceGetName` | GPU model name. |
| GPU UUID | `gUuid` | API: `nvmlDeviceGetUUID` | Unique GPU identifier. |
| Total Memory | `gTotalMemoryMb` | API: `nvmlDeviceGetMemoryInfo().total` | Total frame buffer in MB. |
| PCI Bus ID | `gPciBusId` | API: `nvmlDeviceGetPciInfo().busId` | PCI bus address. |
| Max Graphics Clock | `gMaxGraphicsClock` | API: `nvmlDeviceGetMaxClockInfo(GRAPHICS)` | Maximum graphics clock in MHz. |
| Max SM Clock | `gMaxSmClock` | API: `nvmlDeviceGetMaxClockInfo(SM)` | Maximum SM clock in MHz. |
| Max Memory Clock | `gMaxMemClock` | API: `nvmlDeviceGetMaxClockInfo(MEM)` | Maximum memory clock in MHz. |

---

## vLLM Inference Engine Metrics

*Source: HTTP GET to `VLLM_METRICS_URL` (default: `http://localhost:8000/metrics`)*

The collector parses the Prometheus text response. It uses a regex `^([a-zA-Z0-9_:]+)(?:\{(.+)\})?\s+([0-9\.eE\+\-]+|nan|inf|NaN|Inf)$` to extract values.

| Metric Name | JSON Key | Units | Detailed Source (Prometheus Metric Name) | Description |
| --- | --- | --- | --- | --- |
| **System State** |  |  |  |  |
| Requests Running | `system_requests_running` | count | `vllm:num_requests_running` | Requests currently being processed by engine. |
| Requests Waiting | `system_requests_waiting` | count | `vllm:num_requests_waiting` | Requests queued waiting for processing. |
| Engine Sleep State | `system_engine_sleep_state` | state | `vllm:engine_sleep_state` | 1 = Awake (active), 0 = Sleeping (idle). |
| Preemptions | `system_preemptions_total` | count | `vllm:num_preemptions` | Total requests preempted due to memory pressure. |
| **Cache Utilization** |  |  |  |  |
| KV Cache Usage | `cache_kv_usage_percent` | ratio | `vllm:kv_cache_usage_perc` | GPU KV-cache utilization (0.0-1.0). |
| Prefix Cache Hits | `cache_prefix_hits` | count | `vllm:prefix_cache_hits` | Successful prefix cache lookups. |
| Prefix Cache Queries | `cache_prefix_queries` | count | `vllm:prefix_cache_queries` | Total prefix cache lookup attempts. |
| Multimodal Cache Hits | `cache_multimodal_hits` | count | `vllm:mm_cache_hits` | Successful multimodal cache hits. |
| Multimodal Cache Queries | `cache_multimodal_queries` | count | `vllm:mm_cache_queries` | Total multimodal cache lookups. |
| **Throughput Counters** |  |  |  |  |
| Finished Requests | `requests_finished_total` | count | `vllm:request_success` | Cumulative successfully completed requests. |
| Corrupted Requests | `requests_corrupted_total` | count | `vllm:corrupted_requests` | Cumulative failed/corrupted requests. |
| Prompt Tokens | `tokens_prompt_total` | count | `vllm:prompt_tokens` | Total prompt (prefill) tokens processed. |
| Generation Tokens | `tokens_generation_total` | count | `vllm:generation_tokens` | Total generation (decode) tokens produced. |
| **Latency (Summary)** |  |  |  |  |
| Time to First Token | `latency_ttft_s` | seconds | `vllm:time_to_first_token_seconds_sum` | Time from request receipt to first token generated. |
| End-to-End Latency | `latency_e2e_s` | seconds | `vllm:e2e_request_latency_seconds_sum` | Total request latency from receipt to completion. |
| Queue Latency | `latency_queue_s` | seconds | `vllm:request_queue_time_seconds_sum` | Time spent waiting in queue before processing. |
| Inference Time | `latency_inference_s` | seconds | `vllm:request_inference_time_seconds_sum` | Time spent in active inference (prefill + decode). |
| Prefill Time | `latency_prefill_s` | seconds | `vllm:request_prefill_time_seconds_sum` | Time spent processing prompt tokens. |
| Decode Time | `latency_decode_s` | seconds | `vllm:request_decode_time_seconds_sum` | Time spent generating output tokens. |
| Inter-Token Latency | `latency_inter_token_s` | seconds | `vllm:inter_token_latency_seconds_sum` | Average time between successive output tokens. |
| **Histograms** |  |  |  |  |
| Tokens per Step | `tokens_per_step_histogram` | map | `vllm:iteration_tokens_total_bucket` | Distribution of tokens generated per engine step. |
| Prompt Size | `req_size_prompt_tokens_histogram` | map | `vllm:request_prompt_tokens_bucket` | Distribution of prompt lengths in tokens. |
| Generation Size | `req_size_generation_tokens_histogram` | map | `vllm:request_generation_tokens_bucket` | Distribution of generation lengths in tokens. |
| Max Tokens Param | `req_params_max_tokens_histogram` | map | `vllm:request_params_max_tokens_bucket` | Distribution of `max_tokens` parameter values. |
| N Param | `req_params_n_histogram` | map | `vllm:request_params_n_bucket` | Distribution of `n` (num_sequences) parameter values. |
| TTFT Distribution | `latency_ttft_s_histogram` | map | `vllm:time_to_first_token_seconds_bucket` | Histogram buckets for time-to-first-token. |
| E2E Distribution | `latency_e2e_s_histogram` | map | `vllm:e2e_request_latency_seconds_bucket` | Histogram buckets for end-to-end latency. |
| Queue Distribution | `latency_queue_s_histogram` | map | `vllm:request_queue_time_seconds_bucket` | Histogram buckets for queue wait time. |
| Inference Distribution | `latency_inference_s_histogram` | map | `vllm:request_inference_time_seconds_bucket` | Histogram buckets for inference time. |
| Prefill Distribution | `latency_prefill_s_histogram` | map | `vllm:request_prefill_time_seconds_bucket` | Histogram buckets for prefill time. |
| Decode Distribution | `latency_decode_s_histogram` | map | `vllm:request_decode_time_seconds_bucket` | Histogram buckets for decode time. |
| Inter-Token Distribution | `latency_inter_token_s_histogram` | map | `vllm:inter_token_latency_seconds_bucket` | Histogram buckets for inter-token latency. |

**Histogram Format:** Histograms are exported as JSON objects with bucket boundaries as keys and cumulative counts as values:

```json
{
  "0.001": 0,
  "0.005": 12,
  "0.01": 45,
  "0.025": 120,
  "inf": 150
}

```

---

## Timestamp Fields

All dynamic metrics include corresponding timestamp fields with a `t` prefix:

| Pattern | Example | Detailed Source | Description |
| --- | --- | --- | --- |
| `tv{MetricName}` | `tvCpuTimeUserMode` | Python: `time.time() * 1000` (on file read) | Timestamp for VM-level metric. |
| `tc{MetricName}` | `tcCpuTime` | Python: `time.time() * 1000` (on file read) | Timestamp for container-level metric. |
| `tp{MetricName}` | `tpCpuTimeUserMode` | Python: `time.time() * 1000` (on file read) | Timestamp for process-level metric. |
| `tg{MetricName}` | `tgUtilizationGpu` | Python: `time.time() * 1000` (on API call) | Timestamp for GPU-level metric. |

Timestamps are in milliseconds since Unix epoch (January 1, 1970 00:00:00 UTC).

---

## Output Formats

The profiler supports three export formats:

| Format | Extension | Use Case |
| --- | --- | --- |
| Parquet | `.parquet` | Default. Columnar format, excellent compression, fast analytics. |
| CSV | `.csv` | Universal compatibility, human-readable. |
| TSV | `.tsv` | Tab-separated, useful for certain tools. |

Nested structures (GPU list, process list, histograms) are flattened for tabular export:

* GPU metrics: `nvidia_0_gUtilizationGpu`, `nvidia_1_gUtilizationGpu`, etc.
* Histograms: Stored as JSON strings in single columns.