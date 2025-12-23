import os
import time
import psutil
from typing import Dict, Any

from .collectors.cpu import CpuCollector
from .collectors.mem import MemCollector
from .collectors.disk import DiskCollector
from .collectors.net import NetCollector
from .collectors.nvidia import NvidiaCollector
from .collectors.container import ContainerCollector


class Collector:
    def __init__(self):
        # Initialize all sub-collectors
        self.collectors = {
            "cpu": CpuCollector(),
            "mem": MemCollector(),
            "disk": DiskCollector(),
            "net": NetCollector(),
            "nvidia": NvidiaCollector(),
            "containers": ContainerCollector()
        }

    def close(self):
        for c in self.collectors.values():
            c.cleanup()

    def get_static_info(self, session_uuid: str) -> Dict[str, Any]:
        """Aggregates static info from all collectors."""
        info = {
            "uuid": session_uuid,
            "host": {
                "hostname": os.uname().nodename,
                "kernel": f"{os.uname().sysname} {os.uname().release}",
                "boot_time": psutil.boot_time(),
            }
        }

        # Merge static info from sub-collectors
        # We manually map them to keep the output JSON structure clean
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

    def collect_metrics(self) -> Dict[str, Any]:
        """Aggregates dynamic metrics from all collectors."""
        data = {
            "timestamp": time.time_ns(),
        }

        # Iterate through collectors and populate the data dict
        # keys in self.collectors match keys in output json
        for key, collector in self.collectors.items():
            data[key] = collector.collect()

        return data