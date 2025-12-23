import json
import logging
import os

import pandas as pd

logger = logging.getLogger(__name__)


class Exporter:
    def __init__(self, output_dir, session_uuid):
        self.output_dir = output_dir
        os.makedirs(self.output_dir, exist_ok=True)
        self.session_uuid = session_uuid
        self.snapshot_files = []

    def save_static(self, data):
        path = os.path.join(self.output_dir, f"static_{self.session_uuid}.json")
        with open(path, 'w') as f:
            json.dump(data, f, indent=4)
        logger.info("Saved static info: %s", path)

    def save_snapshot(self, metrics):
        timestamp = metrics.get("timestamp", "unknown")
        filename = f"{self.session_uuid}-{timestamp}.json"
        path = os.path.join(self.output_dir, filename)
        with open(path, 'w') as f:
            json.dump(metrics, f)
        self.snapshot_files.append(path)

    def _flatten_metrics(self, data):
        """Flattens nested dictionary for tabular export."""
        flat = {}
        for key, value in data.items():
            if isinstance(value, dict):
                for sub_key, sub_value in value.items():
                    flat[f"{key}_{sub_key}"] = sub_value
            elif isinstance(value, list):
                for i, item in enumerate(value):
                    flat[f"{key}_{i}"] = item
            else:
                flat[key] = value
        return flat

    def process_session(self, export_format='parquet'):
        if not self.snapshot_files:
            logger.warning("No snapshots captured to process.")
            return

        logger.info("Aggregating %d snapshots...", len(self.snapshot_files))
        all_records = []
        self.snapshot_files.sort()

        for file_path in self.snapshot_files:
            try:
                with open(file_path, 'r') as f:
                    data = json.load(f)
                all_records.append(self._flatten_metrics(data))
            except Exception as e:
                logger.warning("Skipping corrupt file %s: %s", file_path, e)

        if not all_records:
            return

        df = pd.DataFrame(all_records)
        base_path = os.path.join(self.output_dir, self.session_uuid)

        if export_format == 'csv':
            target = f"{base_path}.csv"
            df.to_csv(target, index=False)
            logger.info("Exported CSV: %s", target)

        elif export_format == 'tsv':
            target = f"{base_path}.tsv"
            df.to_csv(target, sep='\t', index=False)
            logger.info("Exported TSV: %s", target)

        elif export_format == 'parquet':
            target = f"{base_path}.parquet"
            try:
                # Convert object columns containing JSON strings back to native types if engine supports it,
                # otherwise keep as strings.
                df.to_parquet(target, index=False)
                logger.info("Exported Parquet: %s", target)
            except ImportError:
                fallback = f"{base_path}.csv"
                logger.error("PyArrow/FastParquet not installed. Falling back to CSV.")
                df.to_csv(fallback, index=False)
            except Exception as e:
                logger.error("Parquet export failed: %s", e)
