#!/bin/bash
set -e

# ===============================
# Script: Systemd Installation for Ocserv Dashboard
# Description:
#   Installs all Ocserv Dashboard services as systemd units.
#
# Requirements:
#   - lib.sh (must exist in same directory)
#
# Exit behavior:
#   Script exits immediately on error (set -e)
# ===============================

# Load shared utilities
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR/lib.sh"

# -----------------------
# Deployment directories
# -----------------------
log "Starting Ocserv Dashboard Systemd Installation..."
INSTALL_DIR="/opt/ocserv_dashboard"
sudo mkdir -p "$INSTALL_DIR"
log "Using installation directory: $INSTALL_DIR"

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
sudo apt install -y gcc curl openssl ca-certificates jq less build-essential libc6-dev pkg-config nginx postgresql-17 postgresql-contrib-17

# -----------------------
# Environment file
# -----------------------
ENV_FILE="$INSTALL_DIR/ocserv_dashboard.env"
if [ -f ".env" ]; then
  log "Copying .env file to $ENV_FILE"
  sudo cp .env "$ENV_FILE"
elif [ -f ".env.sample" ]; then
  log ".env file not found, using .env.sample"
  sudo cp .env.sample "$ENV_FILE"
  warn "IMPORTANT: Please edit $ENV_FILE to configure your settings!"
else
  die "Neither .env nor .env.sample found!"
fi
sudo chmod 600 "$ENV_FILE"

# -----------------------
# Services configuration
# -----------------------
declare -A GO_SERVICES=(
  ["ocserv_admin_api"]="admin_dashboard/api"
  ["ocserv_customer_api"]="customer_dashboard/api"
  ["ocserv_user_manager"]="ocserv_user_manager"
  ["ocserv_log_parser"]="ocserv_log_parser"
  ["ocserv_telegram_bot"]="ocserv_telegram_bot"
)

# -----------------------
# Build Go binaries
# Build Go services
log "Building Go services..."
for service in "${!GO_SERVICES[@]}"; do
  project_dir="${GO_SERVICES[$service]}"
  dest="$INSTALL_DIR/$service"

  log "Building $service from $project_dir..."
  (
    cd "$project_dir" || die "Missing project directory: $project_dir"
    CGO_ENABLED=1 GOOS=linux GOARCH="$ARCH" go build -ldflags="-s -w" -o "$service" main.go
    sudo mv "$service" "$dest"
  )
  sudo chmod +x "$dest"
  ok "Build $service completed"
done

# Run pre-requirements
"$SCRIPT_DIR/pre_requirements.sh"

# Install Postgres if not installed
if ! command -v psql &> /dev/null || ! psql --version | grep -q "17"; then
  echo "PostgreSQL 17 is not installed"
  export POSTGRES_DB POSTGRES_HOST POSTGRES_PORT POSTGRES_USER POSTGRES_PASSWORD
  "$SCRIPT_DIR/postgres.sh"
  ok "✅ PostgreSQL is installed and properly configured."
fi

# -----------------------
# Build and install UIs
# -----------------------
log "Building and installing UIs..."
# Build Admin UI
cd admin_dashboard/ui || die "Missing admin_dashboard/ui"
npm ci
npm run build
sudo mkdir -p /var/www/ocserv_admin
sudo cp -r dist/* /var/www/ocserv_admin/
cd - > /dev/null

# Build Customer UI
cd customer_dashboard/ui || die "Missing customer_dashboard/ui"
npm ci
npm run build
sudo mkdir -p /var/www/ocserv_customer
sudo cp -r dist/* /var/www/ocserv_customer/
cd - > /dev/null

# -----------------------
# Stop existing services
# -----------------------
log "Stopping existing services..."
for service in "${!GO_SERVICES[@]}"; do
  sudo systemctl stop "$service" 2>/dev/null || true
done
sudo systemctl stop nginx 2>/dev/null || true
sudo systemctl stop postgresql 2>/dev/null || true

# -----------------------
# Create required directories
# -----------------------
RECEIPTS_DIR="$INSTALL_DIR/uploads/receipts"
sudo mkdir -p "$RECEIPTS_DIR" /etc/ocserv /etc/ocserv/ssl /etc/ocserv/certs
sudo chmod 750 "$RECEIPTS_DIR"

# -----------------------
# Initialize Postgres (if not already initialized)
# -----------------------
if [ ! -d "/var/lib/postgresql/17/main" ]; then
  log "Initializing PostgreSQL database..."
  sudo su - postgres -c "/usr/lib/postgresql/17/bin/initdb -D /var/lib/postgresql/17/main -A md5"
  sudo systemctl start postgresql
  sleep 3
fi

# -----------------------
# Database Migration
# -----------------------
log "Running database migrations..."
"$INSTALL_DIR/ocserv_admin_api" migrate

# -----------------------
# Create Nginx config
# -----------------------
NGINX_CONF="/etc/nginx/sites-available/ocserv_dashboard"
sudo tee "$NGINX_CONF" > /dev/null <<EOF
upstream admin_api_backend { server 127.0.0.1:8080; }
upstream customer_api_backend { server 127.0.0.1:8081; }

server {
    listen 3000;
    return 302 https://\$host:3443\$request_uri;
}

server {
    listen 3443 ssl;
    server_name _;

    ssl_certificate     /etc/ocserv/certs/cert.pem;
    ssl_certificate_key /etc/ocserv/certs/cert.key;

    # Admin Dashboard UI
    location / {
        root /var/www/ocserv_admin;
        index index.html;
        try_files \$uri \$uri/ /index.html;
    }

    # Customer Dashboard UI
    location /customer {
        alias /var/www/ocserv_customer;
        index index.html;
        try_files \$uri \$uri/ /customer/index.html;
    }

    # Admin API
    location ~ ^/(api) {
        proxy_pass http://admin_api_backend;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Customer API
    location ~ ^/(customer-api) {
        proxy_pass http://customer_api_backend;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable Nginx config
sudo ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# -----------------------
# Create systemd units for Go services
# -----------------------
for service in "${!GO_SERVICES[@]}"; do
  unit_file="/etc/systemd/system/${service}.service"
  binary="$INSTALL_DIR/$service"

  case "$service" in
    ocserv_admin_api)
      ARGS="serve --host \${ADMIN_API_HOST:-0.0.0.0} --port \${ADMIN_API_PORT:-8080}"
      ;;
    ocserv_customer_api)
      ARGS="serve --host \${CUSTOMER_API_HOST:-0.0.0.0} --port \${CUSTOMER_API_PORT:-8081}"
      ;;
    ocserv_user_manager)
      ARGS="serve"
      ;;
    ocserv_log_parser)
      ARGS="serve --host \${LOG_PARSER_HOST:-0.0.0.0} --port \${LOG_PARSER_PORT:-8082}"
      ;;
    ocserv_telegram_bot)
      ARGS="serve"
      ;;
  esac

  log "Creating systemd unit for $service"
  sudo tee "$unit_file" > /dev/null <<EOF
[Unit]
Description=$service service for Ocserv Dashboard
After=network.target postgresql.service

[Service]
Type=simple
EnvironmentFile=$ENV_FILE
ExecStart=$binary $ARGS
Restart=always
User=root
WorkingDirectory=$INSTALL_DIR
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

  sudo chmod 644 "$unit_file"
done

# -----------------------
# Reload systemd and start services
# -----------------------
log "Reloading systemd and starting services..."
sudo systemctl daemon-reload

# Start Postgres
sudo systemctl enable --now postgresql
sleep 3

# Start Go services
for service in "${!GO_SERVICES[@]}"; do
  sudo systemctl enable --now "$service"
  ok "Started $service"
done

# Start Nginx
sudo systemctl enable --now nginx

# -----------------------
# Install & Setup Ocserv
# -----------------------
log "Setting up Ocserv..."

# Function to get network interface
get_interface() {
  printf "\n"
  local interface_list
  interface_list=$(ip -o link show | awk '{print $2}' | tr -d ':' | grep -Ev '^(lo|docker|br-|veth|tun|vethe)')
  if [[ -z "$interface_list" ]]; then
    die "❌ No physical network interfaces found!"
  fi
  
  local numbered_interfaces=()
  for iface in $interface_list; do
    numbered_interfaces+=("$iface")
  done

  if [[ ${#numbered_interfaces[@]} -eq 1 ]]; then
    ETH="${numbered_interfaces[0]}"
    print_message highlight "✅ Only one physical interface found. Auto-selected: $ETH"
    return
  fi

  print_message highlight "Available physical network interfaces:"
  local i=1
  for iface in "${numbered_interfaces[@]}"; do
    print_message highlight "$(printf "%4d: %s" "$i" "$iface")"
    ((i++))
  done

  read -rp "Enter the number corresponding to the desired network interface: " interface_number
  if [[ "$interface_number" =~ ^[0-9]+$ ]] && (( interface_number >= 1 && interface_number <= ${#numbered_interfaces[@]} )); then
    ETH="${numbered_interfaces[$((interface_number-1))]}"
    print_message highlight "✅ Selected interface: $ETH"
    printf "\n"
  else
    print_message error "❌ Invalid selection: $interface_number. Please try again."
    printf "\n"
    get_interface
  fi
}

get_interface

export OCSERV_PORT SSL_CN SSL_ORG SSL_EXPIRE OCSERV_DNS ETH OCSERV_BANNER OCSERV_PRE_LOGIN_BANNER OC_NET
"$SCRIPT_DIR/ocserv.sh"

log "Cleaning apt caches..."
sudo apt autoremove -y
sudo apt autoclean -y

ok "Cleanup completed."

log "Installation complete!"
echo
echo "Next steps:"
echo " 1. Edit $ENV_FILE to configure your settings"
echo " 2. Restart systemd services with: sudo systemctl restart ocserv_*"
