import os
from .base import BaseCollector

CGROUP_DIR = '/sys/fs/cgroup'


class ContainerCollector(BaseCollector):
    def collect(self):
        # Basic Cgroup v1 implementation
        if not os.path.exists(CGROUP_DIR):
            return {}

        try:
            cpu_usage = self._read_int(os.path.join(CGROUP_DIR, "cpuacct", "cpuacct.usage"))
            mem_usage = self._read_int(os.path.join(CGROUP_DIR, "memory", "memory.usage_in_bytes"))

            return {
                "cpu_usage_ns": cpu_usage,
                "memory_used_bytes": mem_usage
            }
        except:
            return {}

    def _read_int(self, path):
        with open(path, 'r') as f:
            return int(f.read().strip())