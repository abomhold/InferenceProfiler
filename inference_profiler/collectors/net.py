import time

from .base import BaseColletor


class NetCollector(BaseColletor):
    @staticmethod
    def collect():
        total_recv = 0
        total_sent = 0
        read_time = time.time()
        try:
            with open('/proc/net/dev', 'r') as f:
                lines = f.readlines()
                for line in lines[2:]:
                    if ':' in line:
                        _, stats = line.split(':', 1)
                        parts = stats.split()

                        if len(parts) >= 9:
                            total_recv += int(parts[0])
                            total_sent += int(parts[8])
        except Exception:
            pass

        return {
            "vNetworkBytesRecvd": total_recv,
            "tvNetworkBytesRecvd": read_time,

            "vNetworkBytesSent": total_sent,
            "tvNetworkBytesSent": read_time,
        }
