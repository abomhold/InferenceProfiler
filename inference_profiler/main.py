import argparse
import logging
import signal
import subprocess
import sys
import time
import uuid

from .collectors.collector_manager import CollectorManager
from .exporter import Exporter

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
    parser.add_argument(
        "-o",
        "--output",
        default="./profiler-output",
        help="Output directory for logs"
    )
    parser.add_argument("-t", "--interval", default=1000,
                        type=int, help="Sampling interval in milliseconds")
    parser.add_argument("-f", "--format", default='parquet',
                        choices=['parquet', 'csv', 'tsv'], help="Final export format (default: parquet)")
    parser.add_argument("command",
                        nargs=argparse.REMAINDER, help="Optional command to execute and profile")
    args = parser.parse_args()

    # --- 2. Initialization ---
    session_uuid = str(uuid.uuid4())
    logger.info("Session UUID: %s", session_uuid)
    logger.info("Output Dir:   %s", args.output)
    logger.info("Interval:     %dms", args.interval)
    logger.info("Format:       %s", args.format)

    collector = CollectorManager()
    exporter = Exporter(args.output, session_uuid)

    # Capture and save static system info once at startup
    logger.info("Capturing static hardware info...")
    static_data = collector.get_static_info(session_uuid)
    exporter.save_static(static_data)

    # --- 3. Process Management ---
    proc = None
    if args.command:
        cmd_args = args.command[1:] if args.command and args.command[0] == '--' else args.command
        if not cmd_args:
            logger.error("No command provided after flags.")
            sys.exit(1)

        logger.info("Starting subprocess: %s", ' '.join(cmd_args))
        try:
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

            # A. Collect & Export JSON Snapshot
            metrics = collector.collect_metrics()
            exporter.save_snapshot(metrics)

            # B. Check Subprocess Status
            if proc and proc.poll() is not None:
                logger.info("Subprocess finished with exit code %d", proc.returncode)
                running = False
                break

            # C. Sleep Logic
            elapsed_sec = time.time() - loop_start
            sleep_sec = (args.interval / 1000.0) - elapsed_sec
            if sleep_sec > 0:
                time.sleep(sleep_sec)

    except Exception as e:
        logger.error("Unexpected error in profiling loop: %s", e, exc_info=True)

    finally:
        # --- 6. Cleanup & Post-Processing ---
        logger.info("Shutting down...")

        if proc and proc.poll() is None:
            logger.info("Terminating subprocess...")
            proc.terminate()
            try:
                proc.wait(timeout=2)
            except subprocess.TimeoutExpired:
                proc.kill()

        collector.close()

        logger.info("Converting session data to %s...", args.format.upper())
        exporter.process_session(export_format=args.format)
        logger.info("Done.")


if __name__ == "__main__":
    main()
