from abc import ABC, abstractmethod
from typing import Dict, Any

# this mainly exists for two purposes:
# 1. Shared constants
# 2. Prevent complicated conditional calls (e.g. `if HAS_NVML: nvml.nvmlInit()`) outside of the collectors
class BaseCollector(ABC):
    JIFFIES_PER_SECOND = 100

    @abstractmethod
    def collect(self) -> Dict[str, Any]:
        raise NotImplementedError

    @staticmethod
    def get_static_info() -> Dict[str, Any]:
        pass

    @staticmethod
    def cleanup() -> None:
        pass

