import logging
import os
from typing import Dict, Any

from .base import BaseCollector
from .container import ContainerCollector
from .cpu import CpuCollector
from .disk import DiskCollector
from .mem import MemCollector
from .net import NetCollector
from .nvidia import NvidiaCollector
from .proc import ProcCollector
from .vllm import VllmCollector

logger = logging.getLogger(__name__)


class CollectorManager:
    def __init__(self, collect_processes: bool = False):
        """
        Initialize collector manager.
        """
        # Note: Standardizing to instances for all collectors
        # to ensure .collect() and .cleanup() work consistently.
        self.collectors = {
            "cpu": CpuCollector(),
            "disk": DiskCollector(),
            "mem": MemCollector(),
            "net": NetCollector(),
            "container": ContainerCollector(),
            "nvidia": NvidiaCollector(),
            "vllm": VllmCollector(),
        }

        if collect_processes:
            self.collectors["processes"] = ProcCollector()

    def collect_metrics(self) -> Dict[str, Any]:
        """Aggregates metrics from all collectors into a flat dictionary."""
        data = {
            "timestamp": BaseCollector.get_timestamp(),
        }

        for key, collector in self.collectors.items():
            try:
                metrics = collector.collect()
                if isinstance(metrics, dict):
                    # This merges the metrics into the top-level 'data' dict
                    data.update(metrics)
                else:
                    # Fallback if a collector returns a non-dict value
                    data[key] = metrics
            except Exception as e:
                logger.error(f"Error collecting {key} metrics: {e}")

        return data

    def close(self):
        for c in self.collectors.values():
            try:
                c.cleanup()
            except AttributeError:
                pass

    def get_static_info(self, session_uuid: str) -> Dict[str, Any]:
        """Aggregates static info from all collectors."""
        try:
            with open('/proc/stat', 'r') as f:
                for line in f:
                    if line.startswith('btime'):
                        boot_time = int(line.split()[1])
                        break
        except Exception:
            boot_time = 0

        vm_id = self._get_vm_id()

        info = {
            "uuid": session_uuid,
            "vId": vm_id,
            "hostname": os.uname().nodename,
            "kernel": " ".join([x for x in os.uname()]),
            "boot_time": boot_time,
        }

        # Flattens static info into the main dict instead of nesting in "host"
        for key in ["cpu", "mem"]:
            try:
                static_data = self.collectors[key].get_static_info()
                if static_data:
                    info.update(static_data)
            except Exception:
                pass

        try:
            nvidia_static = self.collectors["nvidia"].get_static_info()
            if nvidia_static:
                info["gDriverVersion"] = nvidia_static.get("gDriverVersion")
                info["gCudaVersion"] = nvidia_static.get("gCudaVersion")
                info["nvidia_gpus"] = nvidia_static.get("gpus", [])
        except Exception:
            pass

        return info

    def _get_vm_id(self) -> str:
        """Attempt to get VM/instance ID from various sources."""
        try:
            with open('/sys/class/dmi/id/product_uuid', 'r') as f:
                vm_id = f.read().strip()
                if vm_id and vm_id != "None":
                    return vm_id
        except Exception:
            pass

        try:
            with open('/etc/machine-id', 'r') as f:
                vm_id = f.read().strip()
                if vm_id:
                    return vm_id
        except Exception:
            pass

        return "unavailable"