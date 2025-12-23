import logging
import os
import time
from typing import Dict, Any

from .base import BaseColletor
from .container import ContainerCollector
from .cpu import CpuCollector
from .disk import DiskCollector
from .mem import MemCollector
from .net import NetCollector
from .nvidia import NvidiaCollector
from .proc import ProcCollector

logger = logging.getLogger(__name__)

class CollectorManager:
    def __init__(self):
        self.collectors = {
            "cpu": CpuCollector(),
            "mem": MemCollector(),
            "disk": DiskCollector(),
            "net": NetCollector(),
            "containers": ContainerCollector(),
            "nvidia": NvidiaCollector(),
            "processes": ProcCollector(),
        }

    def collect_metrics(self) -> Dict[str, Any]:
        """Aggregates dynamic metrics from all collectors."""
        data = {
            "timestamp": BaseColletor.get_timestamp(),
        }
        for key, collector in self.collectors.items():
            try:
                data[key] = collector.collect()
            except Exception as e:
                logger.error(e)
                data[key] = {}
        return data

    def close(self):
        for c in self.collectors.values():
            c.cleanup()

    def get_static_info(self, session_uuid: str) -> Dict[str, Any]:
        """Aggregates static info from all collectors."""
        # Calculate boot time without psutil
        try:
            with open('/proc/stat', 'r') as f:
                for line in f:
                    if line.startswith('btime'):
                        boot_time = int(line.split()[1])
                        break
        except Exception:
            boot_time = 0

        info = {
            "uuid": session_uuid,
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
            nvidia_static = self.collectors["nvidia"].get_static_info()
            if nvidia_static:
                info["nvidia_driver"] = nvidia_static.get("driver_version")
                info["cuda_version"] = nvidia_static.get("cuda_version")
                info["nvidia"] = nvidia_static.get("gpus", [])
            else:
                info["nvidia"] = []
        except Exception:
            info["nvidia"] = []

        return info