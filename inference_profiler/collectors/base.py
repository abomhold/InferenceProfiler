import logging
import time
from abc import abstractmethod, ABC
from typing import Any


class BaseCollector(ABC):
    logger = logging.getLogger(__name__)
    JIFFIES_PER_SECOND = 100

    @staticmethod
    @abstractmethod
    def collect() -> dict[str, Any]:
        raise NotImplementedError

    @staticmethod
    def get_static_info() -> dict[str, Any]:
        return {}

    @staticmethod
    def cleanup():
        pass

    # --- Shared Helper Methods ---
    @staticmethod
    def get_timestamp():
        """Returns current time in milliseconds since epoch."""
        return int(time.time() * 1000)

    @staticmethod
    def get_epoch_second():
        """Returns current time in seconds since epoch (for currentTime field)."""
        return int(time.time())

    @staticmethod
    def _probe_file(file, default=None):
        """
        Reads a file safely.
        Returns: (content_string, timestamp)
        """
        timestamp = BaseCollector.get_timestamp()
        try:
            with open(file, 'r') as f:
                content = f.read().strip()
                return content, timestamp
        except Exception:
            BaseCollector.logger.debug(f"Failed to read file: {file}")
            return default, timestamp

    @staticmethod
    def _probe_func(func, default=0):
        """
        Executes a function safely.
        Returns: (result, timestamp)
        """
        timestamp = BaseCollector.get_timestamp()
        try:
            return func(), timestamp
        except Exception:
            BaseCollector.logger.exception(f"Failed to call func: {func}")
            return default, timestamp

    @staticmethod
    def _read_int(file, default=0):
        """
        Reads a file and converts content to int.
        Returns: (int, timestamp)
        """
        content, timestamp = BaseCollector._probe_file(file)
        try:
            return int(content) if content else default, timestamp
        except (ValueError, TypeError):
            BaseCollector.logger.warning(f"Failed to parse int: {content}")
            return default, timestamp

    @staticmethod
    def _get_file_lines(file):
        """
        Reads a file safely and returns a list of lines.
        Returns: (lines_list, timestamp)
        """
        content, timestamp = BaseCollector._probe_file(file)
        if content:
            return content.splitlines(), timestamp
        return [], timestamp

    @staticmethod
    def _parse_proc_kv(file, separator=':'):
        """
        Parses a Key-Value file (like /proc/meminfo or /proc/cpuinfo).
        Returns: (dict, timestamp)
        """
        lines, timestamp = BaseCollector._get_file_lines(file)
        data = {}
        for line in lines:
            if separator in line:
                try:
                    key, val = line.split(separator, 1)
                    data[key.strip()] = val.strip()
                except ValueError:
                    BaseCollector.logger.warning(f"Failed to parse key-value pair: {line}")
                    continue
        return data, timestamp