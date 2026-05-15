#!/bin/bash

set -e

DEBUG=${DEBUG:-0}  # Default to 0 if not set

# -----------------------------
# Forward signals to child processes
# -----------------------------
# shellcheck disable=SC2064
trap "echo '[INFO] Caught SIGTERM, stopping...'; kill -TERM \$OCSERV_PID \$API_PID \$WEBHOOK_PID 2>/dev/null" SIGTERM SIGINT

# -----------------------------
# preload sqlite to postgreSQL database
# -----------------------------
#DB_DIR="/usr/local/bin/db"
#DB_FILE="${DB_DIR}/ocserv.db"
#
#if [ -f "$DB_FILE" ]; then
#    if api db-loader; then
#        mv "${DB_FILE}" \
#           "${DB_DIR}/loaded_to_postgres_ocserv.db"
#    else
#        exit 128
#    fi
#fi

# -----------------------------
# migrating database
# -----------------------------
echo "[INFO] Starting migrating database schemas..."
api migrate


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
# Start Webhook service as non-root user
# -----------------------------
echo "[INFO] Starting Webhook service..."
webhook &
WEBHOOK_PID=$!

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
# If one process exits, terminate the others
# -----------------------------
echo "[INFO] One of the processes exited, stopping all services..."
kill -TERM $OCSERV_PID $API_PID $WEBHOOK_PID 2>/dev/null || true

# -----------------------------
# Wait for all processes to clean up
# -----------------------------
wait
echo "[INFO] All services stopped."
