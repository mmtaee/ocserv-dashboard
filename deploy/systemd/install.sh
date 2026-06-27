#!/bin/bash

# ==============================================================
# Script: install.sh
# Description:
#   Full installation for ocserv-dashboard using systemd.
#   Runs all necessary scripts in order.
#
# Usage:
#   ./install.sh
#
# Prerequisites:
#   - Must be run as root or with sudo
#   - Debian/Ubuntu system
# ==============================================================

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

# Source lib.sh for logging and helpers
source ./lib.sh

log "Starting Ocserv Dashboard installation..."

# Load environment variables from project root .env
if [[ -f "../../.env" ]]; then
    log "Loading environment variables from ../../.env"
    set -o allexport
    source "../../.env"
    set +o allexport
else
    warn "../../.env not found — proceeding with defaults"
fi

# Run pre-requirements check
./pre_requirements.sh

# Install PostgreSQL
./postgres.sh

# Install Ocserv
./ocserv.sh

# Install backend services
./backend.sh

# Install UI and Nginx
./ui.sh

ok "Installation complete!"
info "Access admin UI at http://your-server/admin"
info "Access customer UI at http://your-server/"
