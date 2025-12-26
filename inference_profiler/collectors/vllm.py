import os
import re
import urllib.error
import urllib.request
from collections import defaultdict
from typing import Dict, Any

from .base import BaseCollector

logger = BaseCollector.logger


class VllmCollector(BaseCollector):
    METRICS_URL = os.getenv("VLLM_METRICS_URL", "http://localhost:8000/metrics")
    IGNORED_LABELS = {'model_name', 'model', 'engine_id', 'engine', 'handler', 'method'}
    _LINE_RE = re.compile(r'^([a-zA-Z_:][a-zA-Z0-9_:]*)(?:\{(.+?)\})?\s+(.+)$')

    @staticmethod
    def collect() -> Dict[str, Any]:
        try:
            with urllib.request.urlopen(VllmCollector.METRICS_URL, timeout=0.5) as response:
                metrics = VllmCollector._parse(response.read().decode('utf-8'))
                if metrics:
                    metrics['timestamp'] = BaseCollector.get_timestamp()
                return metrics
        except (urllib.error.URLError, ConnectionResetError, ConnectionRefusedError):
            return {}
        except Exception as e:
            logger.debug(f"vLLM metric collection failed: {e}")
            return {}

    @staticmethod
    def _parse_value(s: str) -> float:
        s = s.lower()
        if 'nan' in s:
            return 0.0
        if 'inf' in s:
            return float('inf') if s[0] != '-' else float('-inf')
        return float(s)

    @staticmethod
    def _parse_labels(s: str) -> Dict[str, str]:
        labels = {}
        for part in s.split(','):
            if '=' in part:
                k, v = part.split('=', 1)
                k = k.strip()
                if k not in VllmCollector.IGNORED_LABELS:
                    labels[k] = v.strip().strip('"')
        return labels

    @staticmethod
    def _label_suffix(labels: Dict[str, str]) -> str:
        return "".join(f"_{k}_{v}" for k, v in sorted(labels.items()))

    @staticmethod
    def _parse(text: str) -> Dict[str, Any]:
        data = {}
        histograms = defaultdict(dict)

        for line in text.splitlines():
            line = line.strip()
            if not line or line.startswith('#'):
                continue

            if not (match := VllmCollector._LINE_RE.match(line)):
                continue

            name, label_str, value_str = match.groups()

            try:
                value = VllmCollector._parse_value(value_str)
            except ValueError:
                continue

            labels = VllmCollector._parse_labels(label_str) if label_str else {}
            clean_name = name.replace(':', '_')

            if name.endswith('_bucket') and 'le' in labels:
                le = labels.pop('le')
                key = f"{clean_name[:-7]}{VllmCollector._label_suffix(labels)}_histogram"
                histograms[key][le] = value
            elif name.endswith('_info') and labels:
                for k, v in labels.items():
                    try:
                        data[f"{clean_name}_{k}"] = int(v)
                    except ValueError:
                        try:
                            data[f"{clean_name}_{k}"] = float(v)
                        except ValueError:
                            data[f"{clean_name}_{k}"] = v
            else:
                data[f"{clean_name}{VllmCollector._label_suffix(labels)}"] = value

        data.update(histograms)
        return data

    @staticmethod
    def get_static_info():
        return {}

    @staticmethod
    def cleanup():
        pass