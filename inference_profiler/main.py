import argparse
import uuid
import time
import signal
import subprocess
import sys
import os
import logging

from inference_profiler.collectors.collector_manager import CollectorManager
from inference_profiler.exporter import Exporter

logger = logging.getLogger(__name__)


def main():
    # --- 0. Logging Config ---
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s',
        datefmt='%H:%M:%S'
    )

    # --- 1. Argument Parsing ---
    parser = argparse.ArgumentParser(description="Resource Profiler with CSV/Parquet Export")
    parser.add_argument("-o", "--output", default="./profiler-output", help="Output directory for logs")
    parser.add_argument("-t", "--interval", type=int, default=1000, help="Sampling interval in milliseconds")
    parser.add_argument("command", nargs=argparse.REMAINDER, help="Optional command to execute and profile")
    args = parser.parse_args()

    # Create output directory immediately
    os.makedirs(args.output, exist_ok=True)

    # --- 2. Initialization ---
    session_uuid = str(uuid.uuid4())
    logger.info("Session UUID: %s", session_uuid)
    logger.info("Output Dir:   %s", args.output)
    logger.info("Interval:     %dms", args.interval)

    collector = CollectorManager()
    exporter = Exporter(args.output, session_uuid)

    # Capture and save static system info once at startup
    logger.info("Capturing static hardware info...")
    static_data = collector.get_static_info(session_uuid)
    exporter.save_static(static_data)

    # --- 3. Process Management (Optional) ---
    proc = None
    if args.command:
        # Handle cases where command is passed after '--'
        cmd_args = args.command[1:] if args.command[0] == '--' else args.command
        logger.info("Starting subprocess: %s", ' '.join(cmd_args))
        try:
            # Start the command non-blocking
            proc = subprocess.Popen(cmd_args)
        except Exception as e:
            logger.critical("Failed to start command: %s", e)
            sys.exit(1)

    # --- 4. Signal Handling ---
    running = True

    def signal_handler(sig, frame):
        nonlocal running
        logger.info("Signal %s received. Stopping profiler...", sig)
        running = False

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # --- 5. Profiling Loop ---
    logger.info("Profiling started. Press Ctrl+C to stop.")

    try:
        while running:
            loop_start = time.time()

            # A. Collect
            metrics = collector.collect_metrics()

            # B. Export Snapshot (JSON)
            exporter.save_snapshot(metrics)

            # C. Check Subprocess Status
            if proc and proc.poll() is not None:
                logger.info("Subprocess finished with exit code %d", proc.returncode)
                running = False
                break

            # D. Sleep Logic (Compensate for collection time)
            elapsed_sec = time.time() - loop_start
            sleep_sec = (args.interval / 1000.0) - elapsed_sec

            if sleep_sec > 0:
                time.sleep(sleep_sec)

    except Exception as e:
        logger.error("Unexpected error in profiling loop: %s", e, exc_info=True)

    finally:
        # --- 6. Cleanup & Post-Processing ---
        logger.info("Shutting down...")

        # Kill subprocess if it's still alive (and we started it)
        if proc and proc.poll() is None:
            logger.info("Terminating subprocess...")
            proc.terminate()
            try:
                proc.wait(timeout=2)
            except subprocess.TimeoutExpired:
                proc.kill()

        # Close collector resources (e.g. NVML)
        collector.close()

        # Convert JSON snapshots to CSV/Parquet
        logger.info("Converting session data...")
        exporter.process_session()
        logger.info("Done.")


if __name__ == "__main__":
    main()