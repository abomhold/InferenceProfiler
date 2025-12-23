import psutil

JIFFIES_PER_SECOND = 100


class MemCollector:
    @staticmethod
    def collect():
        mem = psutil.virtual_memory()
        return {
            "total": mem.total * JIFFIES_PER_SECOND,
            "free": mem.free * JIFFIES_PER_SECOND,
            "used": mem.used * JIFFIES_PER_SECOND,
            "cached": getattr(mem, 'cached', 0) * JIFFIES_PER_SECOND,
            "buffers": getattr(mem, 'buffers', 0) * JIFFIES_PER_SECOND
        }
