import os
import time

from .base import BaseColletor


class ContainerCollector(BaseColletor):
    CGROUP_DIR = '/sys/fs/cgroup'

    @staticmethod
    def is_v2():
        return os.path.exists(os.path.join(ContainerCollector.CGROUP_DIR, "cgroup.controllers"))

    @staticmethod
    def collect():
        if not os.path.exists(ContainerCollector.CGROUP_DIR):
            return {}
        if ContainerCollector.is_v2():
            return ContainerCollector._collect_v2()
        else:
            return ContainerCollector._collect_v1()

    @staticmethod
    def _collect_v1():
        cpu_path = os.path.join(ContainerCollector.CGROUP_DIR, "cpuacct", "cpuacct.usage")
        mem_path = os.path.join(ContainerCollector.CGROUP_DIR, "memory", "memory.usage_in_bytes")
        cpu_usage, t_cpu = BaseColletor._read_int(cpu_path)

        mem_usage, t_mem = BaseColletor._read_int(mem_path)

        return {
            "cgroup_version": 1,
            "cpu_usage_ns": cpu_usage,
            "tv_cpu_usage_ns": t_cpu,
            "memory_used_bytes": mem_usage,
            "tv_memory_used_bytes": t_mem
        }

    @staticmethod
    def _collect_v2():
        cpu_path = os.path.join(ContainerCollector.CGROUP_DIR, "cpu.stat")
        mem_path = os.path.join(ContainerCollector.CGROUP_DIR, "memory.current")
        mem_usage, t_mem = BaseColletor._read_int(mem_path)

        cpu_usage_ns, t_cpu = 0, time.time()
        try:
            with open(cpu_path, 'r') as f:
                t_cpu = time.time()
                for line in f:
                    if line.startswith("usage_usec"):
                        cpu_usage_ns = int(line.split()[1]) * 1000
                        break
        except Exception:
            pass

        return {
            "cgroup_version": 2,
            "cpu_usage_ns": cpu_usage_ns,
            "tv_cpu_usage_ns": t_cpu,
            "memory_used_bytes": mem_usage,
            "tv_memory_used_bytes": t_mem
        }