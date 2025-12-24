# Profiler Metrics Reference

The profiler aggregates metrics from multiple collectors. Below is the complete list of all metrics, their JSON keys,
CSV headers, units, and descriptions.

---

## CPU Metrics

*Source: `/proc/stat`, `/proc/loadavg`, `/sys/devices/system/cpu*`

| Metric Name              | JSON Key / CSV Header | Units   | Description                                   |
|--------------------------|-----------------------|---------|-----------------------------------------------|
| **Time Spent (Jiffies)** |                       |         | *1 Jiffy = 1/100 sec (usually)*               |
| User Time                | `vCpuTimeUserMode`    | Jiffies | Time spent in user mode.                      |
| Kernel Time              | `vCpuTimeKernelMode`  | Jiffies | Time spent in kernel mode.                    |
| Idle Time                | `vCpuIdleTime`        | Jiffies | Time spent in idle loop.                      |
| I/O Wait                 | `vCpuTimeIOWait`      | Jiffies | Time waiting for I/O completion.              |
| Interrupts               | `vCpuTimeIntSrvc`     | Jiffies | Time servicing hardware interrupts.           |
| Soft IRQ                 | `vCpuTimeSoftIntSrvc` | Jiffies | Time servicing software interrupts.           |
| Nice Time                | `vCpuNice`            | Jiffies | Time spent in niced (low priority) user mode. |
| Steal Time               | `vCpuSteal`           | Jiffies | Time stolen by hypervisor (virtualized env).  |
| **Performance**          |                       |         |                                               |
| Context Switches         | `vCpuContextSwitches` | Count   | Total context switches since boot.            |
| Load Average             | `vLoadAvg`            | -       | 1-minute system load average.                 |
| CPU Frequency            | `vCpuMhz`             | MHz     | Current average CPU frequency across cores.   |

---

## Memory Metrics

*Source: `/proc/meminfo`, `/proc/vmstat*`

| Metric Name   | JSON Key / CSV Header | Units | Description                                     |
|---------------|-----------------------|-------|-------------------------------------------------|
| Total Memory  | `vMemoryTotal`        | Bytes | Total physical RAM.                             |
| Free Memory   | `vMemoryFree`         | Bytes | Unused memory (includes buffers/cache).         |
| Used Memory   | `vMemoryUsed`         | Bytes | Active memory (Total - Free - Buffers - Cache). |
| Buffers       | `vMemoryBuffers`      | Bytes | Memory used by kernel buffers.                  |
| Cached        | `vMemoryCached`       | Bytes | Memory used by page cache.                      |
| Usage Percent | `vMemoryPercent`      | %     | Percentage of total RAM in use.                 |
| Page Faults   | `vPgFault`            | Count | Total minor page faults.                        |
| Major Faults  | `vMajorPageFault`     | Count | Total major page faults (requiring disk I/O).   |

---

## Disk I/O Metrics

*Source: `/proc/diskstats*`

| Metric Name | JSON Key / CSV Header | Units | Description                                     |
|-------------|-----------------------|-------|-------------------------------------------------|
| Read Bytes  | `read_bytes`          | Bytes | Cumulative bytes read from all physical disks.  |
| Write Bytes | `write_bytes`         | Bytes | Cumulative bytes written to all physical disks. |
| Read Count  | `read_count`          | Count | Cumulative number of read operations.           |
| Write Count | `write_count`         | Count | Cumulative number of write operations.          |

---

## Network Metrics

*Source: `/proc/net/dev*`

| Metric Name    | JSON Key / CSV Header | Units | Description                                                  |
|----------------|-----------------------|-------|--------------------------------------------------------------|
| Bytes Received | `vNetworkBytesRecvd`  | Bytes | Cumulative bytes received on all non-loopback interfaces.    |
| Bytes Sent     | `vNetworkBytesSent`   | Bytes | Cumulative bytes transmitted on all non-loopback interfaces. |

---

## Container Metrics (cgroups)

*Source: `/sys/fs/cgroup` (Supports v1 and v2)*

| Metric Name    | JSON Key / CSV Header | Units | Description                                             |
|----------------|-----------------------|-------|---------------------------------------------------------|
| CPU Usage      | `cpu_usage_ns`        | ns    | Total CPU time consumed by the container (nanoseconds). |
| Memory Usage   | `memory_used_bytes`   | Bytes | Current memory usage of the container.                  |
| Cgroup Version | `cgroup_version`      | -     | Detected cgroup version (1 or 2).                       |

---

## Process Metrics

*Source: `/proc/[pid]/stat`, `/proc/[pid]/status*`
*Note: This collector returns a list of objects, one per process.*

| Metric Name      | JSON Key / CSV Header         | Units   | Description                                |
|------------------|-------------------------------|---------|--------------------------------------------|
| PID              | `pId`                         | -       | Process ID.                                |
| Name             | `pName`                       | String  | Process executable name.                   |
| Command          | `pCmdline`                    | String  | Full command line arguments.               |
| Threads          | `pNumThreads`                 | Count   | Number of threads in the process.          |
| User Time        | `pCpuTimeUserMode`            | Jiffies | CPU time in user mode.                     |
| Kernel Time      | `pCpuTimeKernelMode`          | Jiffies | CPU time in kernel mode.                   |
| Virtual Mem      | `pVirtualMemoryBytes`         | Bytes   | Total virtual memory size.                 |
| Context Switches | `pVoluntaryContextSwitches`   | Count   | Voluntary context switches.                |
| Involuntary CS   | `pInvoluntaryContextSwitches` | Count   | Involuntary context switches (preemption). |
| Block I/O Delay  | `pBlockIODelays`              | Jiffies | Time spent waiting for block I/O.          |

---

## NVIDIA GPU Metrics

*Source: `nvidia-ml-py` (NVML)*

**Note:** If multiple GPUs are present, the JSON output is a list `[{}, {}]`. For CSV/Parquet, rows are usually
flattened or repeated per GPU.

| Metric Name         | JSON Key / CSV Header | Units  | Description                                              |
|---------------------|-----------------------|--------|----------------------------------------------------------|
| **Utilization**     |                       |        |                                                          |
| GPU Utilization     | `utilization_gpu`     | %      | % time one or more kernels were executing on the GPU.    |
| Memory Utilization  | `utilization_mem`     | %      | % time global (device) memory was being read or written. |
| **Memory**          |                       |        |                                                          |
| Total Memory        | `memory_total_mb`     | MB     | Total installed frame buffer memory.                     |
| Used Memory         | `memory_used_mb`      | MB     | Total allocated frame buffer memory.                     |
| Free Memory         | `memory_free_mb`      | MB     | Total free frame buffer memory.                          |
| BAR1 Used           | `bar1_used_mb`        | MB     | Used PCIe BAR1 memory.                                   |
| BAR1 Free           | `bar1_free_mb`        | MB     | Free PCIe BAR1 memory.                                   |
| **Power & Thermal** |                       |        |                                                          |
| Temperature         | `temperature_c`       | °C     | GPU Core temperature.                                    |
| Power Draw          | `power_draw_w`        | W      | Current power usage.                                     |
| Power Limit         | `power_limit_w`       | W      | Enforced power limit.                                    |
| Fan Speed           | `fan_speed`           | %      | Fan speed percentage (0-100).                            |
| Performance State   | `perf_state`          | String | Power state (P0=Max, P15=Min).                           |
| **Clocks**          |                       |        |                                                          |
| Graphics Clock      | `clock_graphics_mhz`  | MHz    | Current frequency of graphics (shader) engine.           |
| SM Clock            | `clock_sm_mhz`        | MHz    | Current frequency of Streaming Multiprocessors.          |
| Memory Clock        | `clock_mem_mhz`       | MHz    | Current frequency of memory bus.                         |
| **PCIe Throughput** |                       |        |                                                          |
| PCIe TX             | `pcie_tx_kbps`        | KB/s   | Transmit throughput (Host to GPU).                       |
| PCIe RX             | `pcie_rx_kbps`        | KB/s   | Receive throughput (GPU to Host).                        |
| **Processes**       |                       |        |                                                          |
| Process Count       | `process_count`       | Count  | Number of processes currently computing on this GPU.     |
| Process List        | `processes`           | List   | Detailed list of processes (PID, Name, VRAM used).       |

---

## vLLM Metrics

*Source: `http://localhost:8000/metrics` (Prometheus endpoint)*

These metrics are collected from the vLLM inference engine.
**Note:** Histogram metrics (suffix `_histogram`) are exported as complex objects (JSON) or flattened columns (
CSV/Parquet).

| Metric Name                    | JSON Key / CSV Header                  | Units   | Description                                                     |
|--------------------------------|----------------------------------------|---------|-----------------------------------------------------------------|
| **System State**               |                                        |         |                                                                 |
| Requests Running               | `system_requests_running`              | Count   | Number of requests currently being processed by the engine.     |
| Requests Waiting               | `system_requests_waiting`              | Count   | Number of requests waiting in the queue.                        |
| Engine Sleep State             | `system_engine_sleep_state`            | State   | 1 = Awake, 0 = Sleeping (Idle).                                 |
| Preemptions                    | `system_preemptions_total`             | Count   | Total number of requests preempted due to resource constraints. |
| **Cache Usage**                |                                        |         |                                                                 |
| KV Cache Usage                 | `cache_kv_usage_percent`               | %       | Percentage of GPU KV-cache currently used (0.0 to 1.0).         |
| Prefix Hits                    | `cache_prefix_hits`                    | Count   | Total number of successful prefix cache hits (shared prefixes). |
| Prefix Queries                 | `cache_prefix_queries`                 | Count   | Total number of prefix cache lookups.                           |
| Multi-modal Hits               | `cache_multimodal_hits`                | Count   | Total hits in the multi-modal cache (e.g., images).             |
| Multi-modal Queries            | `cache_multimodal_queries`             | Count   | Total queries to the multi-modal cache.                         |
| **Throughput**                 |                                        |         |                                                                 |
| Finished Requests              | `requests_finished_total`              | Count   | Cumulative count of successfully completed requests.            |
| Corrupted Requests             | `requests_corrupted_total`             | Count   | Cumulative count of corrupted or failed requests.               |
| Prompt Tokens                  | `tokens_prompt_total`                  | Count   | Total number of prompt (prefill) tokens processed.              |
| Gen. Tokens                    | `tokens_generation_total`              | Count   | Total number of generation (decode) tokens produced.            |
| **Latency (Summaries)**        |                                        |         |                                                                 |
| TTFT                           | `latency_ttft_s`                       | Seconds | Time to First Token (Summary).                                  |
| E2E Latency                    | `latency_e2e_s`                        | Seconds | End-to-End Request Latency (Summary).                           |
| Inter-Token Latency            | `latency_inter_token_s`                | Seconds | Time between generation of subsequent tokens.                   |
| Queue Latency                  | `latency_queue_s`                      | Seconds | Time spent waiting in the queue.                                |
| Inference Time                 | `latency_inference_s`                  | Seconds | Time spent actively running inference.                          |
| Prefill Time                   | `latency_prefill_s`                    | Seconds | Time spent in the prefill phase.                                |
| Decode Time                    | `latency_decode_s`                     | Seconds | Time spent in the decode phase.                                 |
| **Distributions (Histograms)** |                                        |         | *Key-Value pairs (e.g., "0.5": 10) representing bucket counts.* |
| Step Tokens Dist.              | `tokens_per_step_histogram`            | Map     | Distribution of tokens generated per engine step.               |
| Prompt Size Dist.              | `req_size_prompt_tokens_histogram`     | Map     | Distribution of prompt lengths (tokens).                        |
| Gen. Size Dist.                | `req_size_generation_tokens_histogram` | Map     | Distribution of generation lengths (tokens).                    |
| Max Tokens Dist.               | `req_params_max_tokens_histogram`      | Map     | Distribution of `max_tokens` parameter values requested.        |
| N Params Dist.                 | `req_params_n_histogram`               | Map     | Distribution of `n` (num sequences) parameter values.           |
| TTFT Dist.                     | `latency_ttft_s_histogram`             | Map     | Distribution of Time To First Token.                            |
| E2E Dist.                      | `latency_e2e_s_histogram`              | Map     | Distribution of End-to-End Latency.                             |
| Queue Dist.                    | `latency_queue_s_histogram`            | Map     | Distribution of Queue Latency.                                  |
| Inference Dist.                | `latency_inference_s_histogram`        | Map     | Distribution of Inference Time.                                 |
| Prefill Dist.                  | `latency_prefill_s_histogram`          | Map     | Distribution of Prefill Time.                                   |
| Decode Dist.                   | `latency_decode_s_histogram`           | Map     | Distribution of Decode Time.                                    |
| Inter-Token Dist.              | `latency_inter_token_s_histogram`      | Map     | Distribution of Inter-Token Latency.                            |
