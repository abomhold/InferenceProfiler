import psutil

from .base_collector import BaseCollector


class NetCollector(BaseCollector):
    @staticmethod
    def collect():
        net_io = psutil.net_io_counters()
        return {
            "bytes_recv": net_io.bytes_recv,
            "bytes_sent": net_io.bytes_sent
        }
