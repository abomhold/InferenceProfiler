import logging
import warnings
from typing import Dict, Any, List

from .base import BaseColletor

# Suppress FutureWarning from pynvml about the new nvidia-ml-py package
warnings.filterwarnings("ignore", category=FutureWarning, module="pynvml")

logger = logging.getLogger(__name__)

try:
    import pynvml

    HAS_NVML = True
except ImportError:
    HAS_NVML = False


class NvidiaCollector(BaseColletor):
    def __init__(self):
        self.available = False
        if HAS_NVML:
            try:
                pynvml.nvmlInit()
                self.available = True
            except Exception as e:
                self.available = False
                logger.debug(f"NVIDIA NVML not available: {e}")

    def collect(self) -> List[Dict[str, Any]]:
        if not self.available:
            return []

        gpu_metrics = []
        try:
            device_count = pynvml.nvmlDeviceGetCount()
            for i in range(device_count):
                handle = pynvml.nvmlDeviceGetHandleByIndex(i)

                # 1. Get running processes
                processes = []
                try:
                    procs = pynvml.nvmlDeviceGetComputeRunningProcesses(handle) + \
                            pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle)

                    for p in procs:
                        processes.append({
                            "pid": p.pid,
                            "name": self._get_proc_name(p.pid),
                            "memory_used_mb": (p.usedGpuMemory or 0) // 1024 // 1024
                        })
                except Exception:
                    pass

                # 2. Collect metrics using the safe probe helper
                # utilization returns an object, so we wrap extraction in lambdas
                util_gpu, _ = self._probe(lambda: pynvml.nvmlDeviceGetUtilizationRates(handle).gpu)
                util_mem, _ = self._probe(lambda: pynvml.nvmlDeviceGetUtilizationRates(handle).memory)

                # Memory Info
                mem_total, _ = self._probe(lambda: pynvml.nvmlDeviceGetMemoryInfo(handle).total // 1024 // 1024)
                mem_used, _ = self._probe(lambda: pynvml.nvmlDeviceGetMemoryInfo(handle).used // 1024 // 1024)
                mem_free, _ = self._probe(lambda: pynvml.nvmlDeviceGetMemoryInfo(handle).free // 1024 // 1024)

                # BAR1 Memory
                bar1_used, _ = self._probe(lambda: pynvml.nvmlDeviceGetBAR1MemoryInfo(handle).bar1Used // 1024 // 1024)
                bar1_free, _ = self._probe(lambda: pynvml.nvmlDeviceGetBAR1MemoryInfo(handle).bar1Free // 1024 // 1024)

                # Physical sensors
                temp, _ = self._probe(lambda: pynvml.nvmlDeviceGetTemperature(handle, pynvml.NVML_TEMPERATURE_GPU))
                fan_speed, _ = self._probe(lambda: pynvml.nvmlDeviceGetFanSpeed(handle))
                power_draw, _ = self._probe(lambda: pynvml.nvmlDeviceGetPowerUsage(handle) / 1000.0)
                power_limit, _ = self._probe(lambda: pynvml.nvmlDeviceGetEnforcedPowerLimit(handle) / 1000.0)

                # Clocks
                clock_gr, _ = self._probe(lambda: pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS))
                clock_sm, _ = self._probe(lambda: pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_SM))
                clock_mem, _ = self._probe(lambda: pynvml.nvmlDeviceGetClockInfo(handle, pynvml.NVML_CLOCK_MEM))

                # PCIe
                pcie_tx, _ = self._probe(
                    lambda: pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_TX_BYTES))
                pcie_rx, _ = self._probe(
                    lambda: pynvml.nvmlDeviceGetPcieThroughput(handle, pynvml.NVML_PCIE_UTIL_RX_BYTES))

                # State
                perf_state, _ = self._probe(lambda: f"P{pynvml.nvmlDeviceGetPerformanceState(handle)}", default="P?")

                # 3. Build dictionary
                gpu_metrics.append({
                    "index": i,
                    "utilization_gpu": util_gpu,
                    "utilization_mem": util_mem,
                    "memory_total_mb": mem_total,
                    "memory_used_mb": mem_used,
                    "memory_free_mb": mem_free,
                    "bar1_used_mb": bar1_used,
                    "bar1_free_mb": bar1_free,
                    "temperature_c": temp,
                    "fan_speed": fan_speed,
                    "power_draw_w": f"{power_draw:.2f}",
                    "power_limit_w": f"{power_limit:.2f}",
                    "clock_graphics_mhz": clock_gr,
                    "clock_sm_mhz": clock_sm,
                    "clock_mem_mhz": clock_mem,
                    "pcie_tx_kbps": pcie_tx,
                    "pcie_rx_kbps": pcie_rx,
                    "perf_state": perf_state,
                    "process_count": len(processes),
                    "processes": processes
                })

        except Exception:
            pass
        return gpu_metrics

    def get_static_info(self) -> Dict[str, Any]:
        if not self.available:
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

                # Safe bus ID retrieval
                bus_id, _ = self._probe(lambda: pynvml.nvmlDeviceGetPciInfo(handle).busId, default="unknown")
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
            return {}

    @staticmethod
    def _get_proc_name(pid: int) -> str:
        return BaseColletor._read_file(f'/proc/{pid}/comm', default="unknown").strip()

    def cleanup(self):
        if self.available:
            try:
                pynvml.nvmlShutdown()
            except Exception:
                pass