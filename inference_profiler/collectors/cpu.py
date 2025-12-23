import glob
import os

import psutil

from .base_collector import BaseCollector


class CpuCollector(BaseCollector):
    @staticmethod
    def collect():
        times = psutil.cpu_times()
        return {
            "user": times.user * BaseCollector.JIFFIES_PER_SECOND,
            "system": times.system * BaseCollector.JIFFIES_PER_SECOND,
            "idle": times.idle * BaseCollector.JIFFIES_PER_SECOND,
            "iowait": getattr(times, 'iowait', 0) * BaseCollector.JIFFIES_PER_SECOND,
            "load_avg": psutil.getloadavg()[0]
        }

    @staticmethod
    def get_static_info():
        return {
            "num_processors": psutil.cpu_count(),
            "cpu_cache": CpuCollector._get_cpu_cache()
        }

    @staticmethod
    def _get_cpu_cache():
        cache_map = {}
        for index_dir in glob.glob('/sys/devices/system/cpu/cpu*/cache/index*'):
            try:
                with open(os.path.join(index_dir, "level")) as f:
                    lvl = f.read().strip()
                with open(os.path.join(index_dir, "type")) as f:
                    typ = f.read().strip()
                with open(os.path.join(index_dir, "size")) as f:
                    size_str = f.read().strip()

                suffix = "d" if typ == "Data" else "i" if typ == "Instruction" else ""
                key = f"L{lvl}{suffix}"

                multiplier = 1
                if size_str.endswith('K'):
                    multiplier = 1024
                elif size_str.endswith('M'):
                    multiplier = 1024 * 1024

                val = int(size_str.rstrip('KM')) * multiplier
                cache_map[key] = val
            except Exception:
                continue
        return cache_map
