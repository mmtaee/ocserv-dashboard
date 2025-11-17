#!/bin/bash
set -e

echo "[INFO] Starting User expiry service..."

if [ "$DEBUG" = "1" ]; then
    user_expiry -d
else
    user_expiry
fi