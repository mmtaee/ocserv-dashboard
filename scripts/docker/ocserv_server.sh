#!/bin/bash
set -e

DEBUG=${DEBUG:-0}  # Default to 0 if not set

# -----------------------------
# Forward signals to child processes
# -----------------------------
# shellcheck disable=SC2064
trap "echo '[INFO] Caught SIGTERM, stopping...'; kill -TERM \$OCSERV_PID \$API_PID 2>/dev/null" SIGTERM SIGINT

# -----------------------------
# Start API service as non-root user
# -----------------------------
echo "[INFO] Starting API service..."
if [ "$DEBUG" = "1" ]; then
    api serve -d &
else
    api serve &
fi
API_PID=$!

# -----------------------------
# Start ocserv as root
# -----------------------------
echo "[INFO] Starting ocserv..."
/usr/sbin/ocserv --foreground --debug=999 --config=/etc/ocserv/ocserv.conf &
OCSERV_PID=$!

# -----------------------------
# Wait for any process to exit
# -----------------------------
wait -n

# -----------------------------
# If one process exits, terminate the other
# -----------------------------
echo "[INFO] One of the processes exited, stopping all services..."
kill -TERM $OCSERV_PID $API_PID 2>/dev/null || true

# -----------------------------
# Wait for all processes to clean up
# -----------------------------
wait
echo "[INFO] All services stopped."
