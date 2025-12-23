import psutil

from .base_collector import BaseCollector

class MemCollector(BaseCollector):
    @staticmethod
    def collect():
        mem = psutil.virtual_memory()
        return {
            "total": mem.total * BaseCollector.JIFFIES_PER_SECOND,
            "free": mem.free * BaseCollector.JIFFIES_PER_SECOND,
            "used": mem.used * BaseCollector.JIFFIES_PER_SECOND,
            "cached": getattr(mem, 'cached', 0) * BaseCollector.JIFFIES_PER_SECOND,
            "buffers": getattr(mem, 'buffers', 0) * BaseCollector.JIFFIES_PER_SECOND
        }


