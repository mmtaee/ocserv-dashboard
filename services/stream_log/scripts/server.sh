#!/bin/bash
set -e

echo "[INFO] Starting Log Stream service..."

if [ "$DEBUG" = "1" ]; then
    stream_log -d
else
    stream_log
fi