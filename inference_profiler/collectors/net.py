from .base import BaseCollector


class NetCollector(BaseCollector):
    @staticmethod
    def collect():
        stats, timestamp = NetCollector._get_network_stats()

        return {
            # Byte counters
            "vNetworkBytesRecvd": stats['bytes_recv'],
            "tvNetworkBytesRecvd": timestamp,
            "vNetworkBytesSent": stats['bytes_sent'],
            "tvNetworkBytesSent": timestamp,
            # Packet counters
            "vNetworkPacketsRecvd": stats['packets_recv'],
            "tvNetworkPacketsRecvd": timestamp,
            "vNetworkPacketsSent": stats['packets_sent'],
            "tvNetworkPacketsSent": timestamp,
            # Error counters
            "vNetworkErrorsRecvd": stats['errs_recv'],
            "tvNetworkErrorsRecvd": timestamp,
            "vNetworkErrorsSent": stats['errs_sent'],
            "tvNetworkErrorsSent": timestamp,
            # Drop counters
            "vNetworkDropsRecvd": stats['drops_recv'],
            "tvNetworkDropsRecvd": timestamp,
            "vNetworkDropsSent": stats['drops_sent'],
            "tvNetworkDropsSent": timestamp,
        }

    @staticmethod
    def _get_network_stats():
        stats = {
            'bytes_recv': 0,
            'packets_recv': 0,
            'errs_recv': 0,
            'drops_recv': 0,
            'bytes_sent': 0,
            'packets_sent': 0,
            'errs_sent': 0,
            'drops_sent': 0,
        }

        lines, timestamp = BaseCollector._get_file_lines('/proc/net/dev')

        # Skip header lines (usually first 2)
        for line in lines[2:]:
            if ':' in line:
                try:
                    iface, data = line.split(':', 1)
                    iface = iface.strip()

                    # Skip loopback interface
                    if iface == 'lo':
                        continue

                    parts = data.split()
                    if len(parts) >= 16:
                        # Receive stats (columns 0-7)
                        stats['bytes_recv'] += int(parts[0])
                        stats['packets_recv'] += int(parts[1])
                        stats['errs_recv'] += int(parts[2])
                        stats['drops_recv'] += int(parts[3])
                        # Transmit stats (columns 8-15)
                        stats['bytes_sent'] += int(parts[8])
                        stats['packets_sent'] += int(parts[9])
                        stats['errs_sent'] += int(parts[10])
                        stats['drops_sent'] += int(parts[11])
                except (ValueError, IndexError):
                    continue

        return stats, timestamp