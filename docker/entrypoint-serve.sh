#!/bin/sh
set -e

# Start profiler HTTP server in background
echo "Starting profiler server on :8081..."
infprofiler serve -addr :8081 &
PROFILER_PID=$!

# Give it time to start
sleep 1

# Check if profiler is running
if kill -0 $PROFILER_PID 2>/dev/null; then
    echo "Profiler server started (PID: $PROFILER_PID)"
else
    echo "Warning: Profiler server failed to start"
fi

# Execute the main command (e.g., vllm serve)
echo "Starting: $@"
exec "$@"
