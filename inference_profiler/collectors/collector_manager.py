import os
import time
from typing import Dict, Any

import psutil

from .container import ContainerCollector
from .cpu import CpuCollector
from .disk import DiskCollector
from .mem import MemCollector
from .net import NetCollector
from .nvidia import NvidiaCollector


class CollectorManager:
    def __init__(self):
        self.collectors = {
            "cpu": CpuCollector(),
            "mem": MemCollector(),
            "disk": DiskCollector(),
            "net": NetCollector(),
            "containers": ContainerCollector(),
            "nvidia": NvidiaCollector(),
        }

    def collect_metrics(self) -> Dict[str, Any]:
        """Aggregates dynamic metrics from all collectors."""
        data = {
            "timestamp": time.time_ns(),
        }
        for key, collector in self.collectors.items():
            data[key] = collector.collect()
        return data

    def close(self):
        for c in self.collectors.values():
            c.cleanup()

    def get_static_info(self, session_uuid: str) -> Dict[str, Any]:
        """Aggregates static info from all collectors."""
        info = {
            "uuid": session_uuid,
            "host": {
                "hostname": os.uname().nodename,
                "kernel": " ".join([x for x in os.uname()]),
                "boot_time": psutil.boot_time(),
            }
        }

        cpu_static = self.collectors["cpu"].get_static_info()
        info["host"].update(cpu_static)

        nvidia_static = self.collectors["nvidia"].get_static_info()
        if nvidia_static:
            info["nvidia_driver"] = nvidia_static.get("driver_version")
            info["cuda_version"] = nvidia_static.get("cuda_version")
            info["nvidia"] = nvidia_static.get("gpus", [])
        else:
            info["nvidia"] = []

        return info
