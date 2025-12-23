import glob
import os
import time

from .base import BaseColletor


class CpuCollector(BaseColletor):
    @staticmethod
    def collect():
        """Collect all CPU metrics with precise timestamps."""
        stat_metrics, t_stat = CpuCollector._get_proc_stat()
        load_avg, t_load = CpuCollector._get_load_avg()
        cpu_mhz, t_freq = CpuCollector._get_cpu_freq()
        metrics = {
            "vCpuTimeUserMode": stat_metrics['user'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuTimeUserMode": t_stat,
            "vCpuTimeKernelMode": stat_metrics['system'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuTimeKernelMode": t_stat,
            "vCpuIdleTime": stat_metrics['idle'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuIdleTime": t_stat,
            "vCpuTimeIOWait": stat_metrics['iowait'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuTimeIOWait": t_stat,
            "vCpuTimeIntSrvc": stat_metrics['irq'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuTimeIntSrvc": t_stat,
            "vCpuTimeSoftIntSrvc": stat_metrics['softirq'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuTimeSoftIntSrvc": t_stat,
            "vCpuNice": stat_metrics['nice'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuNice": t_stat,
            "vCpuSteal": stat_metrics['steal'] * BaseColletor.JIFFIES_PER_SECOND,
            "tvCpuSteal": t_stat,
            "vCpuContextSwitches": stat_metrics['ctxt'],
            "tvCpuContextSwitches": t_stat,
            "vLoadAvg": load_avg,
            "vCpuMhz": cpu_mhz,
        }

        return metrics

    @staticmethod
    def _get_proc_stat():
        """Reads /proc/stat for CPU times and context switches."""
        metrics = {'user': 0, 'nice': 0, 'system': 0, 'idle': 0, 'iowait': 0, 'irq': 0, 'softirq': 0, 'steal': 0,
                   'ctxt': 0}
        timestamp = time.time()
        try:
            with open('/proc/stat', 'r') as f:
                for line in f:
                    if line.startswith('cpu '):
                        parts = line.split()
                        metrics['user'] = int(parts[1])
                        metrics['nice'] = int(parts[2])
                        metrics['system'] = int(parts[3])
                        metrics['idle'] = int(parts[4])
                        if len(parts) > 5: metrics['iowait'] = int(parts[5])
                        if len(parts) > 6: metrics['irq'] = int(parts[6])
                        if len(parts) > 7: metrics['softirq'] = int(parts[7])
                        if len(parts) > 8: metrics['steal'] = int(parts[8])
                    elif line.startswith('ctxt '):
                        metrics['ctxt'] = int(line.split()[1])
        except Exception:
            pass
        return metrics, timestamp

    @staticmethod
    def _get_load_avg():
        """Reads /proc/loadavg for 1-minute load."""
        timestamp = time.time()
        try:
            with open('/proc/loadavg', 'r') as f:
                return float(f.read().split()[0]), timestamp
        except Exception:
            return 0.0, timestamp

    @staticmethod
    def _get_cpu_freq():
        """Estimates current CPU frequency in MHz."""
        timestamp = time.time()
        freq_mhz = 0.0

        # Method 1: SysFS
        try:
            freqs = []
            files = glob.glob('/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq')
            if files:
                for p in files:
                    try:
                        with open(p, 'r') as f:
                            freqs.append(int(f.read().strip()))
                    except (IOError, OSError):
                        continue
                if freqs:
                    freq_mhz = (sum(freqs) / len(freqs)) / 1000.0
                    return freq_mhz, timestamp
        except Exception:
            pass

        # Method 2: /proc/cpuinfo fallback
        try:
            with open('/proc/cpuinfo', 'r') as f:
                for line in f:
                    if line.startswith('cpu MHz'):
                        freq_mhz = float(line.split(':')[1])
                        break
        except Exception:
            pass

        return freq_mhz, timestamp

    @staticmethod
    def get_static_info():
        """Collect static CPU information."""
        return {
            "num_processors": os.cpu_count() or 0,
            "cpu_type": CpuCollector._get_cpu_type(),
            "cpu_cache": CpuCollector._get_cpu_cache(),
            "kernel_info": CpuCollector._get_kernel_info(),
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
        try:
            with open('/proc/cpuinfo', 'r') as f:
                for line in f:
                    if line.startswith('model name'):
                        return line.split(':', 1)[1].strip()
            return "unknown"
        except Exception:
            return "unknown"

    @staticmethod
    def _get_cpu_cache():
        """Parse CPU cache information from sysfs safely."""
        cache_map = {}
        for index_dir in glob.glob('/sys/devices/system/cpu/cpu*/cache/index*'):
            try:
                with open(os.path.join(index_dir, "level")) as f:
                    level = f.read().strip()
                with open(os.path.join(index_dir, "type")) as f:
                    cache_type = f.read().strip()
                with open(os.path.join(index_dir, "size")) as f:
                    size_str = f.read().strip()
                with open(os.path.join(index_dir, "shared_cpu_map")) as f:
                    shared_cpu_map = f.read().strip()

                suffix = ""
                if cache_type == "Data":
                    suffix = "d"
                elif cache_type == "Instruction":
                    suffix = "i"
                key = f"L{level}{suffix}"

                multiplier = 1
                if size_str.endswith('K'):
                    multiplier = 1024
                    size_val = int(size_str[:-1])
                elif size_str.endswith('M'):
                    multiplier = 1024 * 1024
                    size_val = int(size_str[:-1])
                else:
                    size_val = int(size_str)
                size_bytes = size_val * multiplier

                if key not in cache_map: cache_map[key] = {}
                if shared_cpu_map not in cache_map[key]:
                    cache_map[key][shared_cpu_map] = size_bytes
            except (FileNotFoundError, ValueError, IOError, OSError):
                continue

        result = {}
        for key, cpu_maps in cache_map.items():
            result[key] = sum(cpu_maps.values())

        return result