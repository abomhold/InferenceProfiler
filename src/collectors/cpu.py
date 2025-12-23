import psutil
import glob
import os
from .base import BaseCollector


class CpuCollector(BaseCollector):
    def __init__(self, correction_factor=100.0):
        self.correction = correction_factor

    def collect(self):
        times = psutil.cpu_times()
        return {
            "user": times.user * self.correction,
            "system": times.system * self.correction,
            "idle": times.idle * self.correction,
            "iowait": getattr(times, 'iowait', 0) * self.correction,
            "load_avg": psutil.getloadavg()[0]
        }

    def get_static_info(self):
        return {
            "num_processors": psutil.cpu_count(),
            "cpu_cache": self._get_cpu_cache()
        }

    def _get_cpu_cache(self):
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
                # Simplified: overwriting key to mimic Go logic for total cache per level
                cache_map[key] = val
            except:
                continue
        return cache_map