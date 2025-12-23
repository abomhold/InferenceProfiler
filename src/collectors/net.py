import psutil
from .base import BaseCollector

class NetCollector(BaseCollector):
    def collect(self):
        net_io = psutil.net_io_counters()
        if not net_io:
            return {}

        return {
            "bytes_recv": net_io.bytes_recv,
            "bytes_sent": net_io.bytes_sent
        }