#!/bin/bash
set -e

echo "[INFO] Starting Telegram bot service..."

if [ "${DEBUG:-}" = "1" ]; then
    exec telegram_bot -d
fi

exec telegram_bot
