#!/bin/bash
set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR/lib.sh"

log "Starting Ocserv Dashboard Systemd Uninstallation..."

# -----------------------
# Services to uninstall
# -----------------------
declare -a SERVICES=(
  "ocserv_admin_api"
  "ocserv_customer_api"
  "ocserv_user_manager"
  "ocserv_log_parser"
  "ocserv_telegram_bot"
)

# -----------------------
# Stop and disable services
# -----------------------
log "Stopping and disabling services..."
for service in "${SERVICES[@]}"; do
  if systemctl list-unit-files | grep -q "^${service}.service"; then
    sudo systemctl stop "$service" || true
    sudo systemctl disable "$service" || true
    sudo rm -f "/etc/systemd/system/${service}.service"
    log "Removed $service"
  fi
done

# -----------------------
# Remove Nginx config
# -----------------------
sudo rm -f "/etc/nginx/sites-available/ocserv_dashboard"
sudo rm -f "/etc/nginx/sites-enabled/ocserv_dashboard"

# -----------------------
# Reload systemd
# -----------------------
sudo systemctl daemon-reload

# -----------------------
# Remove files (optional: keep data)
# -----------------------
INSTALL_DIR="/opt/ocserv_dashboard"
read -p "Remove installation directory ($INSTALL_DIR)? This will delete all binaries. [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  sudo rm -rf "$INSTALL_DIR"
  ok "Removed $INSTALL_DIR"
fi

read -p "Remove UI files from /var/www? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  sudo rm -rf /var/www/ocserv_admin /var/www/ocserv_customer
  ok "Removed UI files"
fi

ok "Uninstallation complete!"
