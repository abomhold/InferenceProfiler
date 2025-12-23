import psutil
from typing import Dict, Any, List
from .base import BaseCollector

try:
    import pynvml

    HAS_NVML = True
except ImportError:
    HAS_NVML = False


class NvidiaCollector(BaseCollector):
    def __init__(self):
        self.available = False
        if HAS_NVML:
            pynvml.nvmlInit()
            self.available = True

    def collect(self) -> List[Dict[str, Any]]:
        if not self.available:
            return []

        gpu_metrics = []
        device_count = pynvml.nvmlDeviceGetCount()

        for i in range(device_count):
            handle = pynvml.nvmlDeviceGetHandleByIndex(i)

            # Fetch struct objects once
            util = pynvml.nvmlDeviceGetUtilizationRates(handle)
            mem = pynvml.nvmlDeviceGetMemoryInfo(handle)
            bar1 = pynvml.nvmlDeviceGetBAR1MemoryInfo(handle)

            # Aggregate processes (Compute + Graphics)
            nv_procs = (
                    pynvml.nvmlDeviceGetComputeRunningProcesses(handle) +
                    pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle)
            )

            # Build process list
            processes = [
                {
                    "pid": p.pid,
                    "name": self._get_proc_name(p.pid),
                    "memory_used_mb": (p.usedGpuMemory or 0) // 1024 // 1024
                }
                for p in nv_procs
            ]

            # Create dictionary
            metrics = {
                "index": i,
                "utilization_gpu": util.gpu,
                "utilization_mem": util.memory,
                "memory_total_mb": mem.total // 1024 // 1024,
                "memory_used_mb": mem.used // 1024 // 1024,
                "memory_free_mb": mem.free // 1024 // 1024,
                "bar1_used_mb": bar1.bar1Used // 1024 // 1024,
                "bar1_free_mb": bar1.bar1Free // 1024 // 1024,
                "temperature_c": pynvml.nvmlDeviceGetTemperature(handle, pynvml.NVML_TEMPERATURE_GPU),
                "fan_speed": pynvml.nvmlDeviceGetFanSpeed(handle),
                "power_draw_w": f"{pynvml.nvmlDeviceGetPowerUsage(handle) / 1000.0:.2f}",
                "power_limit_w": f"{pynvml.nvmlDeviceGetEnforcedPowerLimit(handle) / 1000.0:.2f}",
                "clock_graphics_mhz": pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS),
                "clock_sm_mhz": pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_SM),
                "clock_mem_mhz": pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_MEM),
                "pcie_tx_kbps": pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_TX_BYTES),
                "pcie_rx_kbps": pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_RX_BYTES),
                "perf_state": f"P{pynvml.nvmlDeviceGetPerformanceState(handle)}",
                "compute_mode": str(pynvml.nvmlDeviceGetComputeMode(handle)),
                "process_count": len(processes),
                "processes": processes
            }
            gpu_metrics.append(metrics)

        return gpu_metrics

    def get_static_info(self) -> Dict[str, Any]:
        if not self.available:
            return {}

        # List comprehension to build GPU info list
        gpus = []
        for i in range(pynvml.nvmlDeviceGetCount()):
            h = pynvml.nvmlDeviceGetHandleByIndex(i)
            mem = pynvml.nvmlDeviceGetMemoryInfo(h)
            pci = pynvml.nvmlDeviceGetPciInfo(h)

            gpus.append({
                "name": pynvml.nvmlDeviceGetName(h),
                "uuid": pynvml.nvmlDeviceGetUUID(h),
                "total_memory_mb": mem.total // 1024 // 1024,
                "pci_bus_id": pci.busId.decode('utf-8', errors='ignore') if isinstance(pci.busId, bytes) else pci.busId,
                "max_graphics_clock": pynvml.nvmlDeviceGetMaxClockInfo(h, pynvml.NVML_CLOCK_GRAPHICS),
                "max_sm_clock": pynvml.nvmlDeviceGetMaxClockInfo(h, pynvml.NVML_CLOCK_SM),
                "max_mem_clock": pynvml.nvmlDeviceGetMaxClockInfo(h, pynvml.NVML_CLOCK_MEM),
            })

        return {
            "driver_version": pynvml.nvmlSystemGetDriverVersion(),
            "cuda_version": pynvml.nvmlSystemGetCudaDriverVersion(),
            "gpus": gpus
        }

    def _get_proc_name(self, pid: int) -> str:
        try:
            return psutil.Process(pid).name()
        except (psutil.NoSuchProcess, psutil.AccessDenied):
            return "unknown"

    def cleanup(self):
        if self.available:
            pynvml.nvmlShutdown()