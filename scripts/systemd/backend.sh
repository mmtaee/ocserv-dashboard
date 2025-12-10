#!/bin/bash

set -euo pipefail
export DEBIAN_FRONTEND=noninteractive

source "$(dirname "$0")/lib.sh"

# Sensible defaults (can be overridden via environment)
OCSERV_PORT="${OCSERV_PORT:-443}"              # ocserv TCP/UDP port; 443 is typical
OC_NET="${OC_NET:-172.16.24.0/24}"             # VPN IPv4 subnet
OCSERV_DNS="${OCSERV_DNS:-1.1.1.1}"           # DNS pushed to clients
ETH="${ETH:-}"                                 # External interface (auto-detect if empty)

# Auto-detect external interface if not set
if [[ -z "${ETH}" ]]; then
  ETH="$(ip -o -4 route show to default 2>/dev/null | awk '{print $5}' | head -n1 || true)"
  [[ -n "${ETH}" ]] || die "Could not auto-detect external interface. Set ETH env var (e.g. ETH=eth0)."
  log "Auto-detected external interface: ${ETH}"
fi

log "Starting deployment..."

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
    cd "$project_dir" || die "Missing project dir: $project_dir"
    go mod tidy
    go mod download
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
