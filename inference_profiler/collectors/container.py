import os

from .base_collector import BaseCollector

CGROUP_DIR = '/sys/fs/cgroup'

class ContainerCollector(BaseCollector):
    @staticmethod
    def collect():
        if not os.path.exists(CGROUP_DIR):
            return {}

        return {
            "cpu_usage_ns": ContainerCollector._read_int(
                os.path.join(CGROUP_DIR, "cpuacct", "cpuacct.usage")),
            "memory_used_bytes": ContainerCollector._read_int(
                os.path.join(CGROUP_DIR, "memory", "memory.usage_in_bytes"))
        }

    @staticmethod
    def _read_int(path) -> int:
        try:
            with open(path, 'r') as f:
                return int(f.read().strip())
        except Exception:
            return 0
