import json
import os
import logging
import pandas as pd
from typing import Dict, Any

logger = logging.getLogger(__name__)


class Exporter:
    def __init__(self, output_dir: str, session_uuid: str):
        self.output_dir = output_dir
        self.session_uuid = session_uuid
        self.snapshot_files = []

        # Ensure output directory exists
        os.makedirs(output_dir, exist_ok=True)

    def save_static(self, data: Dict[str, Any]):
        """Saves static hardware info to {uuid}.json"""
        filename = f"{self.session_uuid}.json"
        path = os.path.join(self.output_dir, filename)

        try:
            with open(path, 'w') as f:
                json.dump(data, f, indent=4)
            logger.info("Static info saved: %s", path)
        except IOError as e:
            logger.error("Error saving static info: %s", e)

    def save_snapshot(self, data: Dict[str, Any]):
        """Saves a single metric interval to {uuid}-{timestamp}.json"""
        ts = data.get("timestamp", 0)
        filename = f"{self.session_uuid}-{ts}.json"
        path = os.path.join(self.output_dir, filename)

        try:
            with open(path, 'w') as f:
                json.dump(data, f, indent=4)
            self.snapshot_files.append(path)
        except IOError as e:
            logger.error("Error saving snapshot: %s", e)

    def process_session(self):
        """
        Reads all captured JSON snapshots, flattens them into a tabular format,
        and exports to CSV, TSV, and Parquet.
        """
        if not self.snapshot_files:
            logger.warning("No snapshots captured to process.")
            return

        logger.info("Aggregating %d snapshots...", len(self.snapshot_files))

        all_records = []

        # Sort files to ensure chronological order in the DataFrame
        self.snapshot_files.sort()

        for file_path in self.snapshot_files:
            try:
                with open(file_path, 'r') as f:
                    data = json.load(f)

                # Flatten the complex nested dictionary into a single row
                flat_record = self._flatten_metrics(data)
                all_records.append(flat_record)

            except Exception as e:
                logger.warning("Skipping corrupt file %s: %s", file_path, e)

        if not all_records:
            return

        # Create DataFrame
        df = pd.DataFrame(all_records)

        # Define base path for exports
        base_name = os.path.join(self.output_dir, self.session_uuid)

        # 1. Export CSV
        csv_path = f"{base_name}.csv"
        df.to_csv(csv_path, index=False)
        logger.info("Exported CSV:     %s", csv_path)

        # 2. Export TSV
        tsv_path = f"{base_name}.tsv"
        df.to_csv(tsv_path, sep='\t', index=False)
        logger.info("Exported TSV:     %s", tsv_path)

        # 3. Export Parquet (requires pyarrow)
        try:
            parquet_path = f"{base_name}.parquet"
            df.to_parquet(parquet_path, index=False)
            logger.info("Exported Parquet: %s", parquet_path)
        except ImportError:
            logger.warning("PyArrow not found. Skipping Parquet export.")
        except Exception as e:
            logger.error("Parquet export failed: %s", e)

    def _flatten_metrics(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Flattens the nested metric dictionary (CPU, Mem, Nvidia List) into
        a flat dictionary suitable for a CSV row.
        """
        flat = {
            "timestamp": data.get("timestamp"),
            "uuid": self.session_uuid
        }

        # Categories that are simple dictionaries: data['cpu'] = {'user': 10, ...}
        simple_categories = ["cpu", "mem", "disk", "net", "containers"]

        for cat in simple_categories:
            if cat in data and isinstance(data[cat], dict):
                for k, v in data[cat].items():
                    flat[f"{cat}_{k}"] = v

        # Handle NVIDIA GPUs (List of dictionaries)
        nvidia_data = data.get("nvidia", [])
        if isinstance(nvidia_data, list):
            for gpu in nvidia_data:
                idx = gpu.get("index", "unknown")
                prefix = f"nvidia_{idx}"

                for k, v in gpu.items():
                    # Skip the index itself as it's part of the prefix
                    if k == "index":
                        continue

                    # Handle the 'processes' list inside GPU data specially
                    if k == "processes" and isinstance(v, list):
                        flat[f"{prefix}_process_count_detailed"] = len(v)
                    else:
                        flat[f"{prefix}_{k}"] = v

        return flat