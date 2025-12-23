import re
import time

from .base import BaseColletor


class MemCollector(BaseColletor):
    @staticmethod
    def collect():
        mem_info, t_mem_read = MemCollector._get_meminfo()
        total = mem_info.get('MemTotal', 0)
        free_raw = mem_info.get('MemFree', 0)
        buffers = mem_info.get('Buffers', 0)
        cached = mem_info.get('Cached', 0) + mem_info.get('SReclaimable', 0)

        # Calculate 'Available'
        if 'MemAvailable' in mem_info:
            available = mem_info['MemAvailable']
        else:
            available = free_raw + buffers + cached

        # Calculate 'Used' (Total - Free - Buffers - Cached)
        used = total - free_raw - buffers - cached

        # Calculate Percent Used
        percent = 0.0
        if total > 0:
            percent = ((total - available) / total) * 100

        metrics = {
            # Main memory metrics (using t_mem_read)
            "vMemoryTotal": total,
            "tvMemoryTotal": t_mem_read,

            "vMemoryFree": available,
            "tvMemoryFree": t_mem_read,

            "vMemoryUsed": used,
            "tvMemoryUsed": t_mem_read,

            "vMemoryBuffers": buffers,
            "tvMemoryBuffers": t_mem_read,

            "vMemoryCached": cached,
            "tvMemoryCached": t_mem_read,

            "vMemoryPercent": percent,
        }

        # Read Page Fault Stats
        pgfault, pgmajfault, t_vmstat_read = MemCollector._get_page_faults()

        metrics.update({
            "vPgFault": pgfault,
            "tvPgFault": t_vmstat_read,

            "vMajorPageFault": pgmajfault,
            "tvMajorPageFault": t_vmstat_read,
        })

        return metrics

    @staticmethod
    def _get_meminfo():
        """Reads /proc/meminfo and returns (dict, timestamp)."""
        info = {}
        timestamp = time.time_ns()
        try:
            with open('/proc/meminfo', 'r') as f:
                for line in f:
                    parts = line.split()
                    if len(parts) >= 2:
                        key = parts[0].rstrip(':')
                        value = int(parts[1]) * 1024  # Convert kB to bytes
                        info[key] = value
        except FileNotFoundError:
            pass
        return info, timestamp

    @staticmethod
    def _get_page_faults():
        """Reads /proc/vmstat and returns (pgfault, pgmajfault, timestamp)."""
        timestamp = time.time_ns()
        try:
            with open('/proc/vmstat', 'r') as f:
                content = f.read()

            pgfault_match = re.search(r'pgfault\s+(\d+)', content)
            pgmajfault_match = re.search(r'pgmajfault\s+(\d+)', content)

            pgfault = int(pgfault_match.group(1)) if pgfault_match else 0
            pgmajfault = int(pgmajfault_match.group(1)) if pgmajfault_match else 0

            return pgfault, pgmajfault, timestamp
        except Exception:
            return 0, 0, timestamp

    @staticmethod
    def get_static_info():
        """Get static memory information."""
        mem_info, _ = MemCollector._get_meminfo()
        return {
            "memory_total_bytes": mem_info.get('MemTotal', 0),
        }