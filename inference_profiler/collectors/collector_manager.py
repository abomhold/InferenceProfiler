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

        Args:
            collect_processes: Whether to collect per-process metrics (can be expensive)
        """
        self.collectors = {
            "cpu": CpuCollector,
            "disk": DiskCollector,
            "mem": MemCollector,
            "net": NetCollector,
            "containers": ContainerCollector,
            "nvidia": NvidiaCollector(),
            "vllm": VllmCollector(),
        }

        if collect_processes:
            self.collectors["processes"] = ProcCollector

    def collect_metrics(self) -> Dict[str, Any]:
        """Aggregates dynamic metrics from all collectors."""
        data = {
            "timestamp": BaseCollector.get_timestamp(),
        }
        for key, collector in self.collectors.items():
            try:
                data[key] = collector.collect()
            except Exception as e:
                logger.error(f"Error collecting {key} metrics: {e}")
                data[key] = {}
        return data

    def close(self):
        for c in self.collectors.values():
            c.cleanup()

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

        # Get VM ID from various sources
        vm_id = self._get_vm_id()

        info = {
            "uuid": session_uuid,
            "vId": vm_id,
            "host": {
                "hostname": os.uname().nodename,
                "kernel": " ".join([x for x in os.uname()]),
                "boot_time": boot_time,
            }
        }

        # Safe static collection
        try:
            cpu_static = self.collectors["cpu"].get_static_info()
            info["host"].update(cpu_static)
        except Exception:
            pass

        try:
            mem_static = self.collectors["mem"].get_static_info()
            info["host"].update(mem_static)
        except Exception:
            pass

        try:
            nvidia_static = self.collectors["nvidia"].get_static_info()
            if nvidia_static:
                info["gDriverVersion"] = nvidia_static.get("gDriverVersion")
                info["gCudaVersion"] = nvidia_static.get("gCudaVersion")
                info["nvidia"] = nvidia_static.get("gpus", [])
            else:
                info["nvidia"] = []
        except Exception:
            info["nvidia"] = []

        return info

    def _get_vm_id(self) -> str:
        """Attempt to get VM/instance ID from various sources."""
        # Try cloud provider metadata
        vm_id = "unavailable"

        # DMI product UUID
        try:
            with open('/sys/class/dmi/id/product_uuid', 'r') as f:
                vm_id = f.read().strip()
                if vm_id and vm_id != "None":
                    return vm_id
        except Exception:
            pass

        # Machine ID as fallback
        try:
            with open('/etc/machine-id', 'r') as f:
                vm_id = f.read().strip()
                if vm_id:
                    return vm_id
        except Exception:
            pass

        return vm_id