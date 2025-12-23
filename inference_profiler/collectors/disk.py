import psutil


class DiskCollector:
    @staticmethod
    def collect():
        disk_io = psutil.disk_io_counters()
        if not disk_io:
            return {}

        return {
            "read_bytes": disk_io.read_bytes,
            "write_bytes": disk_io.write_bytes,
            "read_count": disk_io.read_count,
            "write_count": disk_io.write_count
        }
