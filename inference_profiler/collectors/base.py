import logging
import time
from abc import abstractmethod, ABC


class BaseColletor(ABC):
    logger = logging.getLogger(__name__)
    JIFFIES_PER_SECOND = 100

    @staticmethod
    def get_timestamp():
        return time.time()

    @staticmethod
    @abstractmethod
    def collect():
        """Collect metrics. Must be implemented by subclasses."""
        raise NotImplementedError

    @staticmethod
    def get_static_info():
        """Return static hardware info (optional override)."""
        return {}

    @staticmethod
    def cleanup():
        """Cleanup resources (optional override)."""
        pass

    # --- Shared Helper Methods ---
    @staticmethod
    def _read_file(path, default=""):
        """Reads a file and returns (value, timestamp)."""
        timestamp = time.time()
        try:
            with open(path, 'r') as f:
                return f.read()
        except (IOError, OSError):
            return default, timestamp

    @staticmethod
    def _read_int(path):
        """Reads a single integer from a file and returns (value, timestamp)."""
        timestamp = time.time()
        try:
            with open(path, 'r') as f:
                content = f.read().strip()
                return int(content) if content else 0, timestamp
        except Exception:
            return 0, timestamp

    @staticmethod
    def probe(func, default=0):
        """
        Safely executes a callable and returns (result, timestamp).
        Captures the timestamp immediately before execution.
        """
        timestamp = time.time()
        try:
            return func(), timestamp
        except Exception:
            return default, timestamp