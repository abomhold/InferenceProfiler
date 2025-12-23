import logging
import warnings
from typing import Dict, Any, List

from .base import BaseCollector
logger = logging.getLogger(__name__)

try:
    # Suppress FutureWarning from pynvml about the new nvidia-ml-py package
    warnings.filterwarnings("ignore", category=FutureWarning, module="pynvml")
    import pynvml

    HAS_NVML = True
except ImportError:
    HAS_NVML = False


class NvidiaCollector(BaseCollector):
    def __init__(self):
        if HAS_NVML:
            try:
                pynvml.nvmlInit()
            except Exception as e:
                logger.debug(f"NVIDIA NVML not available: {e}")

    @staticmethod
    def collect() -> List[Dict[str, Any]]:
        if not HAS_NVML:
            return []

        gpu_metrics = []
        try:
            device_count = pynvml.nvmlDeviceGetCount()
            for i in range(device_count):
                handle = pynvml.nvmlDeviceGetHandleByIndex(i)
                processes = []
                procs1 = BaseCollector._probe_func(lambda: pynvml.nvmlDeviceGetComputeRunningProcesses(handle))
                procs2 = BaseCollector._probe_func(lambda: pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle))
                for p in procs1 + procs2:
                    processes.append({
                        "pid": p.pid,
                        "name": BaseCollector._probe_file(f'/proc/{p.pid}/comm', default="unknown")[0].strip(),
                        "memory_used_mb": (p.usedGpuMemory or 0) // 1024 // 1024
                    })

                # 2. Collect metrics using the safe probe helper
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
                    lambda: pynvml.nvmlDeviceGetPowerUsage(handle) / 1000.0)
                power_limit_val, t_pwr_lim = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetEnforcedPowerLimit(handle) / 1000.0)
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
                    "power_draw_w": power_draw_val,
                    "t_power_draw_w": t_pwr,
                    "power_limit_w": power_limit_val,
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

        except Exception:
            pass
        return gpu_metrics

    @staticmethod
    def get_static_info() -> Dict[str, Any]:
        if not HAS_NVML:
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
                    lambda: pynvml.nvmlDeviceGetPciInfo(handle).busId,default="unknown"
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
        except Exception:
            BaseCollector.logger.exception("Failed to get static info")
            return {}

    @staticmethod
    def cleanup():
        if HAS_NVML:
            try:
                pynvml.nvmlShutdown()
            except Exception:
                BaseCollector.logger.exception("Failed to shutdown pynvml")
                pass
