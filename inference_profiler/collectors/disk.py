import re
from .base import BaseCollector


class DiskCollector(BaseCollector):
    SECTOR_SIZE = 512

    @staticmethod
    def collect():
        stats, t_read = DiskCollector._get_disk_io()

        return {
            "read_bytes": stats['read_bytes'],
            "tv_read_bytes": t_read,
            "write_bytes": stats['write_bytes'],
            "tv_write_bytes": t_read,
            "read_count": stats['read_count'],
            "tv_read_count": t_read,
            "write_count": stats['write_count'],
            "tv_write_count": t_read
        }

    @staticmethod
    def _get_disk_io():
        metrics = {'read_count': 0, 'write_count': 0, 'read_bytes': 0, 'write_bytes': 0}

        # Regex to match physical devices (sda, vda, nvme0n1, mmcblk0)
        # Excludes partitions (sda1), loopbacks (loop0), and ram (ram0)
        # 1. sd/hd/vd + letters (standard disks)
        # 2. nvme/mmc patterns (modern flash storage)
        # 3. xvd + letters (Xen disks)
        disk_pattern = re.compile(r'^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$')

        lines, timestamp = BaseCollector._get_file_lines('/proc/diskstats')

        for line in lines:
            parts = line.split()
            if len(parts) < 14:
                continue

            dev_name = parts[2]
            if disk_pattern.match(dev_name):
                try:
                    metrics['read_count'] += int(parts[3])
                    metrics['read_bytes'] += int(parts[5]) * DiskCollector.SECTOR_SIZE
                    metrics['write_count'] += int(parts[7])
                    metrics['write_bytes'] += int(parts[9]) * DiskCollector.SECTOR_SIZE
                except (ValueError, IndexError):
                    continue

        return metrics, timestamp