import re

from .base import BaseCollector


class MemCollector(BaseCollector):
    @staticmethod
    def collect():
        mem_info, t_mem_read = MemCollector._get_meminfo()
        pgfault, pgmajfault, t_vmstat_read = MemCollector._get_page_faults()

        #todo: don't collect total everytime
        total = mem_info.get('MemTotal', 0)
        free_raw = mem_info.get('MemFree', 0)
        buffers = mem_info.get('Buffers', 0)
        cached = mem_info.get('Cached', 0) + mem_info.get('SReclaimable', 0)

        if 'MemAvailable' in mem_info:
            available = mem_info['MemAvailable']
        else:
            available = free_raw + buffers + cached

        used = total - free_raw - buffers - cached

        percent = 0.0
        if total > 0:
            percent = ((total - available) / total) * 100

        metrics = {
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
            "vPgFault": pgfault,
            "tvPgFault": t_vmstat_read,
            "vMajorPageFault": pgmajfault,
            "tvMajorPageFault": t_vmstat_read,
        }

        return metrics

    @staticmethod
    def _get_meminfo():
        raw_info, timestamp = BaseCollector._parse_proc_kv('/proc/meminfo')
        processed = {}
        for k, v in raw_info.items():
            try:
                processed[k] = int(v.replace(' kB', '')) * 1024
            except ValueError:
                continue
        return processed, timestamp

    @staticmethod
    def _get_page_faults():
        content, timestamp = BaseCollector._probe_file('/proc/vmstat')
        if not content:
            return 0, 0, timestamp

        pgfault = 0
        pgmajfault = 0

        pgfault_match = re.search(r'pgfault\s+(\d+)', content)
        pgmajfault_match = re.search(r'pgmajfault\s+(\d+)', content)

        if pgfault_match: pgfault = int(pgfault_match.group(1))
        if pgmajfault_match: pgmajfault = int(pgmajfault_match.group(1))

        return pgfault, pgmajfault, timestamp

    @staticmethod
    def get_static_info():
        mem_info, _ = MemCollector._get_meminfo()
        return {
            "memory_total_bytes": mem_info.get('MemTotal', 0),
        }
