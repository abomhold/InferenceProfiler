from .base import BaseCollector


class NetCollector(BaseCollector):
    @staticmethod
    def collect():
        total_recv = 0
        total_sent = 0

        lines, timestamp = BaseCollector._get_file_lines('/proc/net/dev')

        # Skip header lines (usually first 2)
        for line in lines[2:]:
            if ':' in line:
                try:
                    _, stats = line.split(':', 1)
                    parts = stats.split()
                    if len(parts) >= 9:
                        total_recv += int(parts[0])
                        total_sent += int(parts[8])
                except (ValueError, IndexError):
                    continue

        return {
            "vNetworkBytesRecvd": total_recv,
            "tvNetworkBytesRecvd": timestamp,
            "vNetworkBytesSent": total_sent,
            "tvNetworkBytesSent": timestamp,
        }