import warnings
from typing import Dict, Any, List
from .base import BaseCollector

logger = BaseCollector.logger

# Suppress FutureWarning from pynvml about the new nvidia-ml-py package
warnings.filterwarnings("ignore", category=FutureWarning, module="pynvml")

import pynvml


class NvidiaCollector(BaseCollector):
    _initialized = False

    def __init__(self):
        NvidiaCollector._ensure_init()

    @classmethod
    def _ensure_init(cls):
        """Idempotent initialization of NVML."""
        if not cls._initialized:
            try:
                pynvml.nvmlInit()
                cls._initialized = True
                logger.info("NVIDIA NVML initialized successfully.")
            except pynvml.NVMLError as e:
                logger.error(f"Failed to initialize NVIDIA NVML: {e}")
                # We do not set _initialized to True, so we can retry or fail gracefully later

    @staticmethod
    def collect() -> List[Dict[str, Any]]:
        # Fix 1: Ensure NVML is initialized before trying to collect
        NvidiaCollector._ensure_init()

        if not NvidiaCollector._initialized:
            return []

        gpu_metrics = []
        try:
            device_count = pynvml.nvmlDeviceGetCount()
            for i in range(device_count):
                handle = pynvml.nvmlDeviceGetHandleByIndex(i)

                # --- Process Collection ---
                # Combined compute + graphics processes
                processes = []
                active_procs = []

                try:
                    active_procs.extend(pynvml.nvmlDeviceGetComputeRunningProcesses(handle))
                except pynvml.NVMLError:
                    pass  # No compute processes or not supported

                try:
                    active_procs.extend(pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle))
                except pynvml.NVMLError:
                    pass  # No graphics processes

                # Deduplicate based on PID
                seen_pids = set()
                unique_procs = []
                for p in active_procs:
                    if p.pid not in seen_pids:
                        seen_pids.add(p.pid)
                        unique_procs.append(p)

                for p in unique_procs:
                    # Fix 2: Resolve process name via /proc for Docker compatibility
                    proc_name = "unknown"
                    try:
                        with open(f'/proc/{p.pid}/comm', 'r') as f:
                            proc_name = f.read().strip()
                    except Exception:
                        pass  # Fallback to unknown if permission denied or pid gone

                    mem_used = (p.usedGpuMemory or 0) // 1024 // 1024
                    processes.append({
                        "pid": p.pid,
                        "name": proc_name,
                        "memory_used_mb": mem_used
                    })

                # --- Metrics Collection ---
                # We use _probe_func helper from BaseCollector to capture (value, timestamp)

                util_gpu, t_util_gpu = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetUtilizationRates(handle).gpu)

                util_mem, t_util_mem = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetUtilizationRates(handle).memory)

                mem_total, t_mem_total = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetMemoryInfo(handle).total // 1024 // 1024)

                mem_used, t_mem_used = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetMemoryInfo(handle).used // 1024 // 1024)

                mem_free, t_mem_free = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetMemoryInfo(handle).free // 1024 // 1024)

                bar1_used, t_bar1_used = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetBAR1MemoryInfo(handle).bar1Used // 1024 // 1024)

                bar1_free, t_bar1_free = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetBAR1MemoryInfo(handle).bar1Free // 1024 // 1024)

                temp, t_temp = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetTemperature(handle, pynvml.NVML_TEMPERATURE_GPU))

                fan_speed, t_fan = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetFanSpeed(handle))

                power_draw_val, t_pwr = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetPowerUsage(
                        handle) // 1000) # Convert to watts

                power_limit_val, t_pwr_lim = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetEnforcedPowerLimit(handle) // 1000) # Convert to watts

                clock_gr, t_clk_gr = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS))

                clock_sm, t_clk_sm = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_SM))

                clock_mem, t_clk_mem = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_MEM))

                pcie_tx, t_pcie_tx = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_TX_BYTES))

                pcie_rx, t_pcie_rx = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_RX_BYTES))

                perf_state, t_perf = BaseCollector._probe_func(
                    lambda: f"P{pynvml.nvmlDeviceGetPerformanceState(handle)}", default="P?")

                gpu_metrics.append({
                    "utilization_gpu": util_gpu,
                    "t_utilization_gpu": t_util_gpu,
                    "utilization_mem": util_mem,
                    "t_utilization_mem": t_util_mem,
                    "memory_total_mb": mem_total,
                    "t_memory_total_mb": t_mem_total,
                    "memory_used_mb": mem_used,
                    "t_memory_used_mb": t_mem_used,
                    "memory_free_mb": mem_free,
                    "t_memory_free_mb": t_mem_free,
                    "bar1_used_mb": bar1_used,
                    "t_bar1_used_mb": t_bar1_used,
                    "bar1_free_mb": bar1_free,
                    "t_bar1_free_mb": t_bar1_free,
                    "temperature_c": temp,
                    "t_temperature_c": t_temp,
                    "fan_speed": fan_speed,
                    "t_fan_speed": t_fan,
                    "power_draw_w": power_draw_val / 1000.0 if power_draw_val else 0,
                    "t_power_draw_w": t_pwr,
                    "power_limit_w": power_limit_val / 1000.0 if power_limit_val else 0,
                    "t_power_limit_w": t_pwr_lim,
                    "clock_graphics_mhz": clock_gr,
                    "t_clock_graphics_mhz": t_clk_gr,
                    "clock_sm_mhz": clock_sm,
                    "t_clock_sm_mhz": t_clk_sm,
                    "clock_mem_mhz": clock_mem,
                    "t_clock_mem_mhz": t_clk_mem,
                    "pcie_tx_kbps": pcie_tx,
                    "t_pcie_tx_kbps": t_pcie_tx,
                    "pcie_rx_kbps": pcie_rx,
                    "t_pcie_rx_kbps": t_pcie_rx,
                    "perf_state": perf_state,
                    "t_perf_state": t_perf,
                    "process_count": len(processes),
                    "processes": processes
                })

        except Exception as e:
            # Fix 3: Log errors instead of swallowing them
            logger.error(f"Error collecting NVIDIA metrics: {e}", exc_info=True)

        return gpu_metrics

    @staticmethod
    def get_static_info() -> Dict[str, Any]:
        NvidiaCollector._ensure_init()

        if not NvidiaCollector._initialized:
            return {}

        try:
            info = {
                "driver_version": pynvml.nvmlSystemGetDriverVersion(),
                "cuda_version": pynvml.nvmlSystemGetCudaDriverVersion(),
                "gpus": []
            }
            count = pynvml.nvmlDeviceGetCount()
            for i in range(count):
                handle = pynvml.nvmlDeviceGetHandleByIndex(i)
                mem = pynvml.nvmlDeviceGetMemoryInfo(handle)

                bus_id, _ = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetPciInfo(handle).busId, default="unknown"
                )
                if isinstance(bus_id, bytes):
                    bus_id = bus_id.decode('utf-8', errors='ignore')

                info["gpus"].append({
                    "name": pynvml.nvmlDeviceGetName(handle),
                    "uuid": pynvml.nvmlDeviceGetUUID(handle),
                    "total_memory_mb": mem.total // 1024 // 1024,
                    "pci_bus_id": bus_id,
                    "max_graphics_clock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS),
                    "max_sm_clock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_SM),
                    "max_mem_clock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_MEM)
                })
            return info
        except Exception as e:
            logger.error(f"Failed to get static NVIDIA info: {e}")
            return {}

    @staticmethod
    def cleanup():
        if NvidiaCollector._initialized:
            try:
                pynvml.nvmlShutdown()
                NvidiaCollector._initialized = False
            except Exception:
                pass