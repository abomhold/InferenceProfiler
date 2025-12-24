import os
from .base import BaseCollector


class ContainerCollector(BaseCollector):
    CGROUP_DIR = '/sys/fs/cgroup'

    @staticmethod
    def is_v2():
        return os.path.exists(os.path.join(ContainerCollector.CGROUP_DIR, "cgroup.controllers"))

    @staticmethod
    def collect():
        if not os.path.exists(ContainerCollector.CGROUP_DIR):
            return {}
        if ContainerCollector.is_v2():
            return ContainerCollector._collect_v2()
        else:
            return ContainerCollector._collect_v1()

    @staticmethod
    def _get_container_id():
        """Attempt to detect container ID from various sources."""
        # Try /proc/self/cgroup
        try:
            with open('/proc/self/cgroup', 'r') as f:
                for line in f:
                    parts = line.strip().split(':')
                    if len(parts) >= 3:
                        path = parts[2]
                        # Docker format: /docker/<container_id>
                        if '/docker/' in path:
                            return path.split('/docker/')[-1][:12]
                        # Kubernetes format: various patterns
                        if '/kubepods/' in path:
                            segments = path.split('/')
                            if segments:
                                return segments[-1][:12]
        except Exception:
            pass

        # Try hostname as fallback (often set to container ID)
        try:
            hostname = os.uname().nodename
            if len(hostname) == 12 and all(c in '0123456789abcdef' for c in hostname.lower()):
                return hostname
        except Exception:
            pass

        return "unavailable"

    @staticmethod
    def _collect_v1():
        cpu_path = os.path.join(ContainerCollector.CGROUP_DIR, "cpuacct", "cpuacct.usage")
        cpu_stat_path = os.path.join(ContainerCollector.CGROUP_DIR, "cpuacct", "cpuacct.stat")
        mem_path = os.path.join(ContainerCollector.CGROUP_DIR, "memory", "memory.usage_in_bytes")
        mem_max_path = os.path.join(ContainerCollector.CGROUP_DIR, "memory", "memory.max_usage_in_bytes")
        blkio_path = os.path.join(ContainerCollector.CGROUP_DIR, "blkio", "blkio.throttle.io_service_bytes")

        # CPU metrics
        cpu_usage, t_cpu = BaseCollector._read_int(cpu_path)

        # CPU user/kernel split from cpuacct.stat
        cpu_stat, t_cpu_stat = BaseCollector._parse_proc_kv(cpu_stat_path, separator=' ')
        cpu_user_jiffies = int(cpu_stat.get('user', 0))
        cpu_kernel_jiffies = int(cpu_stat.get('system', 0))

        # Memory metrics
        mem_usage, t_mem = BaseCollector._read_int(mem_path)
        mem_max, t_mem_max = BaseCollector._read_int(mem_max_path)

        # Block I/O metrics
        disk_read, disk_write, t_blkio = ContainerCollector._parse_blkio_v1(blkio_path)

        # Per-CPU stats
        per_cpu = ContainerCollector._get_per_cpu_v1()

        # Network stats (from container's network namespace)
        net_recv, net_sent, t_net = ContainerCollector._get_container_net_stats()

        result = {
            "cId": ContainerCollector._get_container_id(),
            "cCgroupVersion": 1,
            # Total CPU time in nanoseconds
            "cCpuTime": cpu_usage,
            "tcCpuTime": t_cpu,
            # CPU time split by mode (in centiseconds)
            "cCpuTimeUserMode": cpu_user_jiffies * BaseCollector.JIFFIES_PER_SECOND,
            "tcCpuTimeUserMode": t_cpu_stat,
            "cCpuTimeKernelMode": cpu_kernel_jiffies * BaseCollector.JIFFIES_PER_SECOND,
            "tcCpuTimeKernelMode": t_cpu_stat,
            # Number of processors
            "cNumProcessors": os.cpu_count() or 0,
            # Memory metrics
            "cMemoryUsed": mem_usage,
            "tcMemoryUsed": t_mem,
            "cMemoryMaxUsed": mem_max,
            "tcMemoryMaxUsed": t_mem_max,
            # Disk I/O
            "cDiskReadBytes": disk_read,
            "tcDiskReadBytes": t_blkio,
            "cDiskWriteBytes": disk_write,
            "tcDiskWriteBytes": t_blkio,
            # Network
            "cNetworkBytesRecvd": net_recv,
            "tcNetworkBytesRecvd": t_net,
            "cNetworkBytesSent": net_sent,
            "tcNetworkBytesSent": t_net,
        }

        # Add per-CPU stats
        result.update(per_cpu)

        return result

    @staticmethod
    def _collect_v2():
        cpu_stat_path = os.path.join(ContainerCollector.CGROUP_DIR, "cpu.stat")
        mem_path = os.path.join(ContainerCollector.CGROUP_DIR, "memory.current")
        mem_peak_path = os.path.join(ContainerCollector.CGROUP_DIR, "memory.peak")
        io_stat_path = os.path.join(ContainerCollector.CGROUP_DIR, "io.stat")

        # Memory metrics
        mem_usage, t_mem = BaseCollector._read_int(mem_path)
        mem_peak, t_mem_peak = BaseCollector._read_int(mem_peak_path)

        # Parse cpu.stat key-values
        cpu_stats, t_cpu = BaseCollector._parse_proc_kv(cpu_stat_path, separator=' ')
        cpu_usage_micros = int(cpu_stats.get('usage_usec', 0))
        cpu_user_micros = int(cpu_stats.get('user_usec', 0))
        cpu_system_micros = int(cpu_stats.get('system_usec', 0))

        # I/O stats
        disk_read, disk_write, t_io = ContainerCollector._parse_io_stat_v2(io_stat_path)

        # Network stats
        net_recv, net_sent, t_net = ContainerCollector._get_container_net_stats()

        return {
            "cId": ContainerCollector._get_container_id(),
            "cCgroupVersion": 2,
            # Total CPU time in nanoseconds (converted from microseconds)
            "cCpuTime": cpu_usage_micros * 1000,
            "tcCpuTime": t_cpu,
            # CPU time split by mode (in centiseconds, converted from microseconds)
            "cCpuTimeUserMode": cpu_user_micros // 10000,  # usec to centisec
            "tcCpuTimeUserMode": t_cpu,
            "cCpuTimeKernelMode": cpu_system_micros // 10000,
            "tcCpuTimeKernelMode": t_cpu,
            # Number of processors
            "cNumProcessors": os.cpu_count() or 0,
            # Memory metrics
            "cMemoryUsed": mem_usage,
            "tcMemoryUsed": t_mem,
            "cMemoryMaxUsed": mem_peak,
            "tcMemoryMaxUsed": t_mem_peak,
            # Disk I/O
            "cDiskReadBytes": disk_read,
            "tcDiskReadBytes": t_io,
            "cDiskWriteBytes": disk_write,
            "tcDiskWriteBytes": t_io,
            # Network
            "cNetworkBytesRecvd": net_recv,
            "tcNetworkBytesRecvd": t_net,
            "cNetworkBytesSent": net_sent,
            "tcNetworkBytesSent": t_net,
        }

    @staticmethod
    def _get_per_cpu_v1():
        """Get per-CPU usage for cgroup v1."""
        result = {}
        percpu_path = os.path.join(ContainerCollector.CGROUP_DIR, "cpuacct", "cpuacct.usage_percpu")
        content, timestamp = BaseCollector._probe_file(percpu_path)

        if content:
            values = content.split()
            for i, val in enumerate(values):
                try:
                    result[f"cCpu{i}Time"] = int(val)
                    result[f"tcCpu{i}Time"] = timestamp
                except ValueError:
                    continue

        return result

    @staticmethod
    def _parse_blkio_v1(path):
        """Parse blkio.throttle.io_service_bytes for v1."""
        read_bytes = 0
        write_bytes = 0

        lines, timestamp = BaseCollector._get_file_lines(path)
        for line in lines:
            parts = line.split()
            if len(parts) >= 3:
                try:
                    op = parts[1].lower()
                    value = int(parts[2])
                    if op == 'read':
                        read_bytes += value
                    elif op == 'write':
                        write_bytes += value
                except (ValueError, IndexError):
                    continue

        return read_bytes, write_bytes, timestamp

    @staticmethod
    def _parse_io_stat_v2(path):
        """Parse io.stat for cgroup v2."""
        read_bytes = 0
        write_bytes = 0

        lines, timestamp = BaseCollector._get_file_lines(path)
        for line in lines:
            # Format: "major:minor rbytes=X wbytes=Y rios=Z wios=W ..."
            parts = line.split()
            for part in parts:
                if part.startswith('rbytes='):
                    try:
                        read_bytes += int(part.split('=')[1])
                    except (ValueError, IndexError):
                        pass
                elif part.startswith('wbytes='):
                    try:
                        write_bytes += int(part.split('=')[1])
                    except (ValueError, IndexError):
                        pass

        return read_bytes, write_bytes, timestamp

    @staticmethod
    def _get_container_net_stats():
        """Get network stats from container's perspective."""
        recv = 0
        sent = 0

        lines, timestamp = BaseCollector._get_file_lines('/proc/net/dev')

        for line in lines[2:]:  # Skip headers
            if ':' in line:
                try:
                    iface, data = line.split(':', 1)
                    iface = iface.strip()

                    # Skip loopback
                    if iface == 'lo':
                        continue

                    parts = data.split()
                    if len(parts) >= 9:
                        recv += int(parts[0])
                        sent += int(parts[8])
                except (ValueError, IndexError):
                    continue

        return recv, sent, timestamp