import re
from .base import BaseCollector


class DiskCollector(BaseCollector):
    SECTOR_SIZE = 512

    @staticmethod
    def collect():
        stats, t_read = DiskCollector._get_disk_io()

        return {
            # Sector-based metrics
            "vDiskSectorReads": stats['sector_reads'],
            "tvDiskSectorReads": t_read,
            "vDiskSectorWrites": stats['sector_writes'],
            "tvDiskSectorWrites": t_read,
            # Byte-based metrics (derived from sectors)
            "vDiskReadBytes": stats['sector_reads'] * DiskCollector.SECTOR_SIZE,
            "tvDiskReadBytes": t_read,
            "vDiskWriteBytes": stats['sector_writes'] * DiskCollector.SECTOR_SIZE,
            "tvDiskWriteBytes": t_read,
            # Operation counts
            "vDiskSuccessfulReads": stats['read_count'],
            "tvDiskSuccessfulReads": t_read,
            "vDiskSuccessfulWrites": stats['write_count'],
            "tvDiskSuccessfulWrites": t_read,
            # Merged operations (adjacent I/O merged for efficiency)
            "vDiskMergedReads": stats['merged_reads'],
            "tvDiskMergedReads": t_read,
            "vDiskMergedWrites": stats['merged_writes'],
            "tvDiskMergedWrites": t_read,
            # Time spent on I/O (in milliseconds)
            "vDiskReadTime": stats['read_time_ms'],
            "tvDiskReadTime": t_read,
            "vDiskWriteTime": stats['write_time_ms'],
            "tvDiskWriteTime": t_read,
            # I/O in progress and total I/O time
            "vDiskIOInProgress": stats['io_in_progress'],
            "tvDiskIOInProgress": t_read,
            "vDiskIOTime": stats['io_time_ms'],
            "tvDiskIOTime": t_read,
            "vDiskWeightedIOTime": stats['weighted_io_time_ms'],
            "tvDiskWeightedIOTime": t_read,
        }

    @staticmethod
    def _get_disk_io():
        metrics = {
            'read_count': 0,
            'merged_reads': 0,
            'sector_reads': 0,
            'read_time_ms': 0,
            'write_count': 0,
            'merged_writes': 0,
            'sector_writes': 0,
            'write_time_ms': 0,
            'io_in_progress': 0,
            'io_time_ms': 0,
            'weighted_io_time_ms': 0,
        }

        # Regex to match physical devices (sda, vda, nvme0n1, mmcblk0)
        # Excludes partitions (sda1), loopbacks (loop0), and ram (ram0)
        disk_pattern = re.compile(r'^(sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme\d+n\d+|mmcblk\d+)$')

        lines, timestamp = BaseCollector._get_file_lines('/proc/diskstats')

        for line in lines:
            parts = line.split()
            if len(parts) < 14:
                continue

            dev_name = parts[2]
            if disk_pattern.match(dev_name):
                try:
                    # Field positions per /proc/diskstats documentation:
                    # 3: reads completed
                    # 4: reads merged
                    # 5: sectors read
                    # 6: time reading (ms)
                    # 7: writes completed
                    # 8: writes merged
                    # 9: sectors written
                    # 10: time writing (ms)
                    # 11: I/Os in progress
                    # 12: time doing I/Os (ms)
                    # 13: weighted time doing I/Os (ms)
                    metrics['read_count'] += int(parts[3])
                    metrics['merged_reads'] += int(parts[4])
                    metrics['sector_reads'] += int(parts[5])
                    metrics['read_time_ms'] += int(parts[6])
                    metrics['write_count'] += int(parts[7])
                    metrics['merged_writes'] += int(parts[8])
                    metrics['sector_writes'] += int(parts[9])
                    metrics['write_time_ms'] += int(parts[10])
                    metrics['io_in_progress'] += int(parts[11])
                    metrics['io_time_ms'] += int(parts[12])
                    metrics['weighted_io_time_ms'] += int(parts[13])
                except (ValueError, IndexError):
                    continue

        return metrics, timestamp