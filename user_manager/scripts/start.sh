#!/bin/bash
set -e

echo "[INFO] Starting User expiry service..."

if [ "$DEBUG" = "1" ]; then
    user_expiry -d -docker-mode
else
    user_expiry -docker-mode
fi