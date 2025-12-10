#!/bin/bash
# ==============================================================
# Common functions and settings for deployment scripts
# ==============================================================
set -euo pipefail
trap 'echo "❌ Deployment failed at line $LINENO."; exit 1' ERR
export DEBIAN_FRONTEND=noninteractive

# -----------------------
# Logging helpers
# -----------------------
log() { echo -e "ℹ️ $*"; }
ok()  { echo -e "✅ $*"; }
warn(){ echo -e "⚠️ $*"; }
die() { echo -e "❌ $*"; exit 1; }
