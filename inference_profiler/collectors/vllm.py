import os
import re
import urllib.error
import urllib.request
from typing import Dict, Any, List, Tuple, Union

from .base import BaseCollector

logger = BaseCollector.logger


class VllmCollector(BaseCollector):
    METRICS_URL = os.getenv("VLLM_METRICS_URL", "http://localhost:8000/metrics")

    # Regex to parse Prometheus lines
    _PROMETHEUS_REGEX = re.compile(r"^([a-zA-Z0-9_:]+)(?:\{(.+)\})?\s+([0-9\.eE\+\-]+|nan|inf|NaN|Inf)$")

    # Labels to completely ignore to reduce noise
    IGNORED_LABELS = {'model_name', 'model', 'engine_id', 'engine', 'handler', 'method'}

    # Renaming map for cleaner JSON output
    METRIC_ALIASES = {
        # --- System / State ---
        "num_requests_running": "system_requests_running",
        "num_requests_waiting": "system_requests_waiting",
        "engine_sleep_state": "system_engine_sleep_state",
        "num_preemptions": "system_preemptions_total",
        "cache_config_info": "config_cache",

        # --- Cache ---
        "kv_cache_usage_perc": "cache_kv_usage_percent",
        "prefix_cache_hits": "cache_prefix_hits",
        "prefix_cache_queries": "cache_prefix_queries",
        "mm_cache_hits": "cache_multimodal_hits",
        "mm_cache_queries": "cache_multimodal_queries",

        # --- Requests ---
        "request_success": "requests_finished_total",
        "corrupted_requests": "requests_corrupted_total",

        # --- Tokens (Throughput) ---
        "prompt_tokens": "tokens_prompt_total",
        "generation_tokens": "tokens_generation_total",
        "iteration_tokens_total": "tokens_per_step_histogram",

        # --- Latency (Timing) ---
        "time_to_first_token_seconds": "latency_ttft_s",
        "e2e_request_latency_seconds": "latency_e2e_s",
        "request_queue_time_seconds": "latency_queue_s",
        "request_inference_time_seconds": "latency_inference_s",
        "request_prefill_time_seconds": "latency_prefill_s",
        "request_decode_time_seconds": "latency_decode_s",
        "inter_token_latency_seconds": "latency_inter_token_s",

        # --- Request Details ---
        "request_prompt_tokens": "req_size_prompt_tokens",
        "request_generation_tokens": "req_size_generation_tokens",
        "request_params_max_tokens": "req_params_max_tokens",
        "request_params_n": "req_params_n"
    }

    @staticmethod
    def collect() -> Dict[str, Any]:
        metrics = {}
        scrape_time = BaseCollector.get_timestamp()

        try:
            with urllib.request.urlopen(VllmCollector.METRICS_URL, timeout=0.5) as response:
                body = response.read().decode('utf-8')

            parsed_data = VllmCollector._parse_prometheus(body)
            metrics.update(parsed_data)
            if metrics:
                metrics['timestamp'] = scrape_time

        except (urllib.error.URLError, ConnectionResetError, ConnectionRefusedError):
            # vLLM server likely not up yet; return empty (no timestamp implies no data)
            pass
        except Exception as e:
            logger.debug(f"vLLM metric collection failed: {e}")

        return metrics

    @staticmethod
    def _get_clean_name(raw_name: str) -> str:
        lookup_name = raw_name.replace("vllm:", "")
        if lookup_name in VllmCollector.METRIC_ALIASES:
            return VllmCollector.METRIC_ALIASES[lookup_name]
        return raw_name.replace(':', '_')

    @staticmethod
    def _try_parse_number(s: str) -> Union[float, int, str]:
        """Attempts to convert string to int, then float, else returns string."""
        try:
            return int(s)
        except ValueError:
            try:
                return float(s)
            except ValueError:
                return s

    @staticmethod
    def _parse_prometheus(text: str) -> Dict[str, Any]:
        data = {}

        # Temporary storage for histograms
        histograms: Dict[Tuple[str, tuple], List[Tuple[float, float]]] = {}

        for line in text.splitlines():
            line = line.strip()
            if not line or line.startswith('#'):
                continue

            match = VllmCollector._PROMETHEUS_REGEX.match(line)
            if not match:
                continue

            name, label_str, value_str = match.groups()

            # Parse Value
            try:
                if 'nan' in value_str.lower():
                    val = 0.0
                elif 'inf' in value_str.lower():
                    val = float('inf')
                else:
                    val = float(value_str)
            except ValueError:
                continue

            # Parse & Filter Labels
            labels = {}
            if label_str:
                parts = label_str.split(',')
                for p in parts:
                    if '=' in p:
                        k, v = p.split('=', 1)
                        k = k.strip()
                        if k not in VllmCollector.IGNORED_LABELS:
                            labels[k] = v.strip().strip('"')

            # Rename Logic
            is_bucket = name.endswith('_bucket')
            is_sum = name.endswith('_sum')
            is_count = name.endswith('_count')
            is_info = name.endswith('_info')

            base_lookup = name
            suffix = ""

            if is_bucket:
                base_lookup = name[:-7]
                suffix = "_bucket"
            elif is_sum:
                base_lookup = name[:-4]
                suffix = "_sum"
            elif is_count:
                base_lookup = name[:-6]
                suffix = "_count"
            elif is_info:
                pass

            clean_base = VllmCollector._get_clean_name(base_lookup)

            if is_info and labels:
                for k, v in labels.items():
                    info_key = f"{clean_base}_{k}"
                    data[info_key] = VllmCollector._try_parse_number(v)
                continue

            # Handle Histograms (Buckets)
            if is_bucket and 'le' in labels:
                le_val_str = labels.pop('le')
                try:
                    le_val = float('inf') if '+Inf' in le_val_str or 'inf' in le_val_str.lower() else float(le_val_str)
                except ValueError:
                    continue

                other_labels = tuple(sorted(labels.items()))
                histo_key = (clean_base, other_labels)

                if histo_key not in histograms:
                    histograms[histo_key] = []

                histograms[histo_key].append((le_val, val))

            else:
                # Handle Standard Metrics
                final_key = f"{clean_base}{suffix}"

                if labels:
                    lbl_suffix = ""
                    for k, v in sorted(labels.items()):
                        lbl_suffix += f"_{k}_{v}"
                    final_key += lbl_suffix

                data[final_key] = val

        # Flush Histograms into Dictionary Format
        for (base_name, label_tuple), buckets in histograms.items():
            lbl_suffix = ""
            for k, v in label_tuple:
                lbl_suffix += f"_{k}_{v}"

            full_key = f"{base_name}{lbl_suffix}_histogram"
            buckets.sort(key=lambda x: x[0])

            histo_dict = {}
            for le, count in buckets:
                le_key = "inf" if le == float('inf') else str(le)
                histo_dict[le_key] = count

            data[full_key] = histo_dict

        return data

    @staticmethod
    def get_static_info():
        return {}

    @staticmethod
    def cleanup():
        pass