import glob
import os

from .base import BaseCollector


class CpuCollector(BaseCollector):
    @staticmethod
    def collect():
        stat_metrics, t_stat = CpuCollector._get_proc_stat()
        load_avg, t_load = CpuCollector._get_load_avg()
        cpu_mhz, t_freq = CpuCollector._get_cpu_freq()

        metrics = {
            # CPU time metrics (in centiseconds, 1 jiffy = 1/100 sec)
            "vCpuTimeUserMode": stat_metrics['user'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuTimeUserMode": t_stat,
            "vCpuTimeKernelMode": stat_metrics['system'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuTimeKernelMode": t_stat,
            "vCpuIdleTime": stat_metrics['idle'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuIdleTime": t_stat,
            "vCpuTimeIOWait": stat_metrics['iowait'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuTimeIOWait": t_stat,
            "vCpuTimeIntSrvc": stat_metrics['irq'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuTimeIntSrvc": t_stat,
            "vCpuTimeSoftIntSrvc": stat_metrics['softirq'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuTimeSoftIntSrvc": t_stat,
            "vCpuNice": stat_metrics['nice'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuNice": t_stat,
            "vCpuSteal": stat_metrics['steal'] * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuSteal": t_stat,
            # Total CPU time (user + kernel) in centiseconds
            "vCpuTime": (stat_metrics['user'] + stat_metrics['system']) * BaseCollector.JIFFIES_PER_SECOND,
            "tvCpuTime": t_stat,
            # Context switches
            "vCpuContextSwitches": stat_metrics['ctxt'],
            "tvCpuContextSwitches": t_stat,
            # Load average and frequency
            "vLoadAvg": load_avg,
            "tvLoadAvg": t_load,
            "vCpuMhz": cpu_mhz,
            "tvCpuMhz": t_freq,
        }
        return metrics

    @staticmethod
    def _get_proc_stat():
        metrics = {
            'user': 0, 'nice': 0, 'system': 0, 'idle': 0,
            'iowait': 0, 'irq': 0, 'softirq': 0, 'steal': 0, 'ctxt': 0
        }
        lines, timestamp = BaseCollector._get_file_lines('/proc/stat')

        for line in lines:
            parts = line.split()
            if not parts:
                continue

            if parts[0] == 'cpu':
                metrics['user'] = int(parts[1])
                metrics['nice'] = int(parts[2])
                metrics['system'] = int(parts[3])
                metrics['idle'] = int(parts[4])
                if len(parts) > 5:
                    metrics['iowait'] = int(parts[5])
                if len(parts) > 6:
                    metrics['irq'] = int(parts[6])
                if len(parts) > 7:
                    metrics['softirq'] = int(parts[7])
                if len(parts) > 8:
                    metrics['steal'] = int(parts[8])
            elif parts[0] == 'ctxt':
                metrics['ctxt'] = int(parts[1])

        return metrics, timestamp

    @staticmethod
    def _get_load_avg():
        content, timestamp = BaseCollector._probe_file('/proc/loadavg')
        try:
            return float(content.split()[0]), timestamp
        except (AttributeError, ValueError, IndexError):
            return 0.0, timestamp

    @staticmethod
    def _get_cpu_freq():
        timestamp = BaseCollector.get_timestamp()
        freq_mhz = 0.0

        # Method 1: SysFS
        freqs = []
        files = glob.glob('/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq')
        for p in files:
            val, _ = BaseCollector._read_int(p)
            if val > 0:
                freqs.append(val)

        if freqs:
            freq_mhz = (sum(freqs) / len(freqs)) / 1000.0
            return freq_mhz, timestamp

        # Method 2: /proc/cpuinfo fallback
        lines, _ = BaseCollector._get_file_lines('/proc/cpuinfo')
        for line in lines:
            if line.startswith('cpu MHz'):
                try:
                    freq_mhz = float(line.split(':')[1])
                    break
                except ValueError:
                    pass

        return freq_mhz, timestamp

    @staticmethod
    def get_static_info():
        return {
            "vNumProcessors": os.cpu_count() or 0,
            "vCpuType": CpuCollector._get_cpu_type(),
            "vCpuCache": CpuCollector._get_cpu_cache(),
            "vKernelInfo": CpuCollector._get_kernel_info(),
        }

    @staticmethod
    def _get_kernel_info():
        try:
            u = os.uname()
            return f"{u.sysname} {u.nodename} {u.release} {u.version} {u.machine}"
        except Exception:
            return "unknown"

    @staticmethod
    def _get_cpu_type():
        lines, _ = BaseCollector._get_file_lines('/proc/cpuinfo')
        for line in lines:
            if line.startswith('model name'):
                return line.split(':', 1)[1].strip()
        return "unknown"

    @staticmethod
    def _get_cpu_cache():
        cache_map = {}
        for index_dir in glob.glob('/sys/devices/system/cpu/cpu*/cache/index*'):
            level, _ = BaseCollector._probe_file(os.path.join(index_dir, "level"))
            cache_type, _ = BaseCollector._probe_file(os.path.join(index_dir, "type"))
            size_str, _ = BaseCollector._probe_file(os.path.join(index_dir, "size"))
            shared_cpu_map, _ = BaseCollector._probe_file(os.path.join(index_dir, "shared_cpu_map"))

            if not (level and size_str):
                continue

            suffix = ""
            if cache_type == "Data":
                suffix = "d"
            elif cache_type == "Instruction":
                suffix = "i"

            key = f"L{level}{suffix}"

            # Convert size to bytes
            multiplier = 1
            if size_str.endswith('K'):
                multiplier = 1024
                size_str = size_str[:-1]
            elif size_str.endswith('M'):
                multiplier = 1024 * 1024
                size_str = size_str[:-1]

            try:
                size_bytes = int(size_str) * multiplier
            except ValueError:
                continue

            if key not in cache_map:
                cache_map[key] = {}
            # Deduplicate based on shared cpu map
            if shared_cpu_map not in cache_map[key]:
                cache_map[key][shared_cpu_map] = size_bytes

        result = {}
        for key, cpu_maps in cache_map.items():
            result[key] = sum(cpu_maps.values())

        return result