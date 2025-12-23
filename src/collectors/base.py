from abc import ABC, abstractmethod
from typing import Dict, Any

class BaseCollector(ABC):
    @staticmethod
    @abstractmethod
    def collect(self) -> Dict[str, Any]:
        """Return a dictionary of dynamic metrics."""
        pass

    def get_static_info(self) -> Dict[str, Any]:
        """Return a dictionary of static hardware info (optional)."""
        return {}

    def cleanup(self):
        """Optional cleanup method."""
        pass