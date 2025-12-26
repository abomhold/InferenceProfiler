import warnings
from typing import Dict, Any, List

from .base import BaseCollector

logger = BaseCollector.logger

# Suppress FutureWarning from pynvml about the new nvidia-ml-py package
warnings.filterwarnings("ignore", category=FutureWarning, module="pynvml")

try:
    import pynvml

    NVML_AVAILABLE = True
except ImportError:
    NVML_AVAILABLE = False
    pynvml = None


class NvidiaCollector(BaseCollector):
    _initialized = False

    def __init__(self):
        NvidiaCollector._ensure_init()

    @classmethod
    def _ensure_init(cls):
        """Idempotent initialization of NVML."""
        if not NVML_AVAILABLE:
            return

        if not cls._initialized:
            try:
                pynvml.nvmlInit()
                cls._initialized = True
                logger.info("NVIDIA NVML initialized successfully.")
            except pynvml.NVMLError as e:
                logger.error(f"Failed to initialize NVIDIA NVML: {e}")

    @staticmethod
    def collect() -> List[Dict[str, Any]]:
        NvidiaCollector._ensure_init()

        if not NVML_AVAILABLE or not NvidiaCollector._initialized:
            return []

        gpu_metrics = []
        try:
            device_count = pynvml.nvmlDeviceGetCount()
            for i in range(device_count):
                handle = pynvml.nvmlDeviceGetHandleByIndex(i)

                # --- Process Collection ---
                process_strings = []
                active_procs = []

                try:
                    active_procs.extend(pynvml.nvmlDeviceGetComputeRunningProcesses(handle))
                except pynvml.NVMLError:
                    pass

                try:
                    active_procs.extend(pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle))
                except pynvml.NVMLError:
                    pass

                # Deduplicate based on PID
                seen_pids = set()
                unique_procs = []
                for p in active_procs:
                    if p.pid not in seen_pids:
                        seen_pids.add(p.pid)
                        unique_procs.append(p)

                for p in unique_procs:
                    proc_name = "unknown"
                    try:
                        with open(f'/proc/{p.pid}/comm', 'r') as f:
                            proc_name = f.read().strip()
                    except Exception:
                        pass

                    mem_used = (p.usedGpuMemory or 0) // 1024 // 1024
                    process_strings.append(f"{p.pid}: {proc_name} ({mem_used} MB)")

                # --- Metrics Collection ---
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

                power_draw_mw, t_pwr = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetPowerUsage(handle))

                power_limit_mw, t_pwr_lim = BaseCollector._probe_func(
                    lambda: pynvml.nvmlDeviceGetEnforcedPowerLimit(handle))

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

                # ecc_single, t_ecc_s = BaseCollector._probe_func(
                #     lambda: pynvml.nvmlDeviceGetTotalEccErrors(
                #         handle, pynvml.NVML_SINGLE_BIT_ECC, pynvml.NVML_VOLATILE_ECC), default=0)
                #
                # ecc_double, t_ecc_d = BaseCollector._probe_func(
                #     lambda: pynvml.nvmlDeviceGetTotalEccErrors(
                #         handle, pynvml.NVML_DOUBLE_BIT_ECC, pynvml.NVML_VOLATILE_ECC), default=0)

                gpu_metrics.append({
                    "gGpuIndex": i,
                    # Utilization
                    "gUtilizationGpu": util_gpu,
                    "tgUtilizationGpu": t_util_gpu,
                    "gUtilizationMem": util_mem,
                    "tgUtilizationMem": t_util_mem,
                    # Memory
                    "gMemoryTotalMb": mem_total,
                    "tgMemoryTotalMb": t_mem_total,
                    "gMemoryUsedMb": mem_used,
                    "tgMemoryUsedMb": t_mem_used,
                    "gMemoryFreeMb": mem_free,
                    "tgMemoryFreeMb": t_mem_free,
                    "gBar1UsedMb": bar1_used,
                    "tgBar1UsedMb": t_bar1_used,
                    "gBar1FreeMb": bar1_free,
                    "tgBar1FreeMb": t_bar1_free,
                    # Thermal
                    "gTemperatureC": temp,
                    "tgTemperatureC": t_temp,
                    "gFanSpeed": fan_speed,
                    "tgFanSpeed": t_fan,
                    # Power (convert milliwatts to watts)
                    "gPowerDrawW": power_draw_mw / 1000.0 if power_draw_mw else 0,
                    "tgPowerDrawW": t_pwr,
                    "gPowerLimitW": power_limit_mw / 1000.0 if power_limit_mw else 0,
                    "tgPowerLimitW": t_pwr_lim,
                    # Clocks
                    "gClockGraphicsMhz": clock_gr,
                    "tgClockGraphicsMhz": t_clk_gr,
                    "gClockSmMhz": clock_sm,
                    "tgClockSmMhz": t_clk_sm,
                    "gClockMemMhz": clock_mem,
                    "tgClockMemMhz": t_clk_mem,
                    # PCIe
                    "gPcieTxKbps": pcie_tx,
                    "tgPcieTxKbps": t_pcie_tx,
                    "gPcieRxKbps": pcie_rx,
                    "tgPcieRxKbps": t_pcie_rx,
                    # Performance state
                    "gPerfState": perf_state,
                    "tgPerfState": t_perf,
                    # ECC errors
                   # "gEccSingleBitErrors": ecc_single,
                    # "tgEccSingleBitErrors": t_ecc_s,
                    # "gEccDoubleBitErrors": ecc_double,
                    # "tgEccDoubleBitErrors": t_ecc_d,
                    # Processes
                    "gProcessCount": len(process_strings),
                    "gProcesses": process_strings
                })

        except Exception as e:
            logger.error(f"Error collecting NVIDIA metrics: {e}", exc_info=True)

        return gpu_metrics

    @staticmethod
    def get_static_info() -> Dict[str, Any]:
        NvidiaCollector._ensure_init()

        if not NVML_AVAILABLE or not NvidiaCollector._initialized:
            return {}

        try:
            info = {
                "gDriverVersion": pynvml.nvmlSystemGetDriverVersion(),
                "gCudaVersion": pynvml.nvmlSystemGetCudaDriverVersion(),
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
                    "gName": pynvml.nvmlDeviceGetName(handle),
                    "gUuid": pynvml.nvmlDeviceGetUUID(handle),
                    "gTotalMemoryMb": mem.total // 1024 // 1024,
                    "gPciBusId": bus_id,
                    "gMaxGraphicsClock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_GRAPHICS),
                    "gMaxSmClock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_SM),
                    "gMaxMemClock": pynvml.nvmlDeviceGetMaxClockInfo(handle, pynvml.NVML_CLOCK_MEM)
                })
            return info
        except Exception as e:
            logger.error(f"Failed to get static NVIDIA info: {e}")
            return {}

    @staticmethod
    def cleanup():
        if NVML_AVAILABLE and NvidiaCollector._initialized:
            try:
                pynvml.nvmlShutdown()
                NvidiaCollector._initialized = False
            except Exception:
                pass