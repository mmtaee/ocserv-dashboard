#!/bin/bash
# ===============================
# Script: systemd_backend.sh
# Description:
#   Main deployment script for ocserv_dashboard services.
#
#   - Loads logging + helper functions from lib.sh
#   - Detects external interface (ETH) automatically if not provided
#   - Detects OS & architecture
#   - Installs required base packages
#   - Builds Go services (api, log_stream, user_expiry)
#   - Installs binaries into /opt/ocserv_dashboard
#   - Creates & enables systemd services
#   - Reloads and restarts all services
#
# Requirements:
#   - lib.sh must exist and define:
#       print_message(), log(), ok(), warn(), die()
#
# Exit behavior:
#   Script exits immediately on error (set -e)
#   Any error prints the failing line number
# ===============================

# Load logging + helper utilities
source ./scripts/lib.sh

log "Starting Backend Deployment..."

# -----------------------
# Deployment directories
# -----------------------
BIN_DIR="/opt/ocserv_dashboard"
sudo mkdir -p "$BIN_DIR"
log "Using deployment directory: $BIN_DIR"

# -----------------------
# Detect OS and ARCH
# -----------------------
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)   ARCH="amd64" ;;
  i386|i686)ARCH="386" ;;
  aarch64)  ARCH="arm64" ;;
  armv7l)   ARCH="arm" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac
log "Detected OS: $OS, ARCH: $ARCH"

# -----------------------
# Base packages / tools
# -----------------------
log "Installing base packages..."
sudo apt update -y
sudo apt install -y gcc curl openssl ca-certificates jq less build-essential libc6-dev pkg-config

# -----------------------
# Services configuration (collections)
# -----------------------
declare -A SERVICES=(
  ["api"]="./services/api"
  ["log_stream"]="./services/log_stream"
  ["user_expiry"]="./services/user_expiry"
)

# -----------------------
# Build Go binaries
# -----------------------
for service in "${!SERVICES[@]}"; do
  project_dir="${SERVICES[$service]}"
  dest="${BIN_DIR}/${service}"

  log "Building $service from $project_dir ..."
  (
    cd "$project_dir" || die "Missing project directory: $project_dir"
    CGO_ENABLED=1 GOOS=linux GOARCH="${ARCH}" go build -ldflags="-s -w" -o "$service"
    sudo mv "$service" "$dest"
  )
  sudo chmod +x "$dest"
  ok "Build $service completed"
done
ok "All binaries built and deployed into $BIN_DIR"

# -----------------------
# Stop existing services
# -----------------------
log "Stopping existing services (if any)..."
for service in "${!SERVICES[@]}"; do
  sudo systemctl stop "$service" 2>/dev/null || true
done

# -----------------------
# Environment file
# -----------------------
ENV_FILE="${BIN_DIR}/ocserv_dashboard.env"
if [[ -f ".env" ]]; then
  sudo cp .env "$ENV_FILE"
  log "Copied environment file to $ENV_FILE"
else
  warn ".env file not found, skipping environment copy"
fi

# -----------------------
# Create systemd units
# -----------------------
for service in "${!SERVICES[@]}"; do
  unit_file="/etc/systemd/system/${service}.service"
  binary="${BIN_DIR}/${service}"

  case "$service" in
    api)        ARGS="serve --host 127.0.0.1 --port 8080" ;;
    log_stream) ARGS="-h 127.0.0.1 -p 8081 --systemd" ;;
    user_expiry) ARGS="" ;;
    *)          ARGS="" ;;
  esac

  log "Creating systemd unit for $service -> $unit_file"
  sudo tee "$unit_file" >/dev/null <<EOF
[Unit]
Description=$service service
After=network.target

[Service]
Type=simple
EnvironmentFile=${ENV_FILE}
ExecStart=${binary} ${ARGS}
Restart=always
User=root
WorkingDirectory=${BIN_DIR}
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
done

log "Reloading systemd and starting services..."

sudo systemctl daemon-reload

for service in "${!SERVICES[@]}"; do
  sudo systemctl stop "$service"
  sudo systemctl enable "$service"
  sudo systemctl restart "$service"
  ok "Started $service service"
done

ok "Backend services deployed successfully."
