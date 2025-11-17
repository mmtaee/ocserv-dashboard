#!/bin/bash
set -e

echo "[INFO] Starting Log Stream service..."

if [ "$DEBUG" = "1" ]; then
    log_stream -d
else
    log_stream
fi