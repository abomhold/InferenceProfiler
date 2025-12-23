import psutil
from .base import BaseCollector

class MemCollector(BaseCollector):
    def __init__(self, correction_factor=0.001):
        self.correction = correction_factor

    def collect(self):
        mem = psutil.virtual_memory()
        return {
            "total": mem.total * self.correction,
            "free": mem.free * self.correction,
            "used": mem.used * self.correction,
            "cached": getattr(mem, 'cached', 0) * self.correction,
            "buffers": getattr(mem, 'buffers', 0) * self.correction
        }