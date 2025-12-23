import logging
from typing import Dict, Any, List

import psutil

from .base_collector import BaseCollector

logger = logging.getLogger(__name__)

try:
    import pynvml

    # Using nvidia-ml-py which is the modern replacement for pynvml
    HAS_NVML = True
except ImportError:
    HAS_NVML = False


class NvidiaCollector(BaseCollector):
    def __init__(self):
        self.available = False
        if HAS_NVML:
            try:
                pynvml.nvmlInit()
                self.available = True
            except (pynvml.NVMLError, OSError):
                # To prevent crashing on non-NVIDIA systems (like your Mac) at startup.
                self.available = False
                logger.debug("NVIDIA NVML not available. GPU metrics will be skipped.")

    def collect(self) -> List[Dict[str, Any]]:
        if not self.available:
            return []

        gpu_metrics = []
        # We assume these calls succeed if initialization passed
        device_count = pynvml.nvmlDeviceGetCount()

        for i in range(device_count):
            handle = pynvml.nvmlDeviceGetHandleByIndex(i)

            # Get running processes
            processes = []
            for p in (
                    pynvml.nvmlDeviceGetComputeRunningProcesses(handle) +
                    pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle)
            ):
                processes.append({
                    "pid": p.pid,
                    "name": self._get_proc_name(p.pid),
                    "memory_used_mb": (p.usedGpuMemory or 0) // 1024 // 1024
                })

            # Build dictionary
            metrics = {
                "index": i,
                "utilization_gpu": pynvml.nvmlDeviceGetUtilizationRates(handle).gpu,
                "utilization_mem": pynvml.nvmlDeviceGetUtilizationRates(handle).memory,
                "memory_total_mb": pynvml.nvmlDeviceGetMemoryInfo(handle).total // 1024 // 1024,
                "memory_used_mb": pynvml.nvmlDeviceGetMemoryInfo(handle).used // 1024 // 1024,
                "memory_free_mb": pynvml.nvmlDeviceGetMemoryInfo(handle).free // 1024 // 1024,
                "bar1_used_mb": pynvml.nvmlDeviceGetBAR1MemoryInfo(handle).bar1Used // 1024 // 1024,
                "bar1_free_mb": pynvml.nvmlDeviceGetBAR1MemoryInfo(handle).bar1Free // 1024 // 1024,
                "temperature_c": (pynvml.nvmlDeviceGetTemperature(handle, pynvml.NVML_TEMPERATURE_GPU)),
                "fan_speed": (pynvml.nvmlDeviceGetFanSpeed(handle)),
                "power_draw_w": f"{pynvml.nvmlDeviceGetPowerUsage(handle) / 1000.0 :.2f}",
                "power_limit_w": f"{pynvml.nvmlDeviceGetEnforcedPowerLimit(handle) / 1000.0 :.2f}",
                "clock_graphics_mhz": (pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS)),
                "clock_sm_mhz": (pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_SM)),
                "clock_mem_mhz": (pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_MEM)),
                "pcie_tx_kbps": (pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_TX_BYTES)),
                "pcie_rx_kbps": (pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_RX_BYTES)),
                "perf_state": f"P{pynvml.nvmlDeviceGetPerformanceState(handle)}",
                "process_count": len([]),
                "processes": []
            }
            gpu_metrics.append(metrics)

        return gpu_metrics

    def get_static_info(self) -> Dict[str, Any]:
        if not self.available:
            return {}

        info = {
            "driver_version": pynvml.nvmlSystemGetDriverVersion(),
            "cuda_version": pynvml.nvmlSystemGetCudaDriverVersion(),
            "gpus": []
        }

        count = pynvml.nvmlDeviceGetCount()
        for i in range(count):
            handle = pynvml.nvmlDeviceGetHandleByIndex(i)
            mem = pynvml.nvmlDeviceGetMemoryInfo(handle)
            pci = pynvml.nvmlDeviceGetPciInfo(handle)

            # Handle bytes vs string for busID safely just for type conversion
            bus_id = pci.busId
            if isinstance(bus_id, bytes):
                bus_id = bus_id.decode('utf-8', errors='ignore')

            gpu_info = {
                "name": pynvml.nvmlDeviceGetName(handle),
                "uuid": pynvml.nvmlDeviceGetUUID(handle),
                "total_memory_mb": mem.total // 1024 // 1024,
                "pci_bus_id": bus_id,
                "max_graphics_clock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS),
                "max_sm_clock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_SM),
                "max_mem_clock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_MEM)
            }
            info["gpus"].append(gpu_info)

        return info

    def _get_proc_name(self, pid: int) -> str:
        try:
            return psutil.Process(pid).name()
        except Exception:
            return "unknown"

    def cleanup(self):
        if self.available:
            pynvml.nvmlShutdown()
