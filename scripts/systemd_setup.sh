#!/bin/bash

# ==============================================================
# Deployment Script: ocserv_user_management + Web + Nginx
# ==============================================================
#
# Description:
#  - Builds Go services (api, log_stream) and installs them into /opt/ocserv_user_management
#  - Creates and enables systemd service units
#  - Builds frontend (npm + Vite) and deploys to /var/www/site
#  - Installs and configures Nginx with self-signed SSL and reverse proxy
#  - Installs ocserv and configures VPN + iptables NAT/persistence
#
# Usage:
#   sudo ./deploy.sh
#
# Prerequisites:
#  - Debian 12+ (or Ubuntu) with apt
#  - Git repo with `api`, `log_stream`, and `web/` directories
#  - Optional environment variables:
#       SSL_EXPIRE, SSL_C, SSL_ST, SSL_L, SSL_ORG, SSL_OU, SSL_CN
#       OC_NET, OCSERV_PORT, OCSERV_DNS, ETH
# ==============================================================

set -euo pipefail
trap 'echo "❌ Deployment failed at line $LINENO."; exit 1' ERR
export DEBIAN_FRONTEND=noninteractive

# -----------------------
# Helpers & Defaults
# -----------------------
log() { echo -e "ℹ️ $*"; }
ok()  { echo -e "✅ $*"; }
warn(){ echo -e "⚠️ $*"; }
die() { echo -e "❌ $*"; exit 1; }

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
## Go toolchain (if missing)
#if ! command -v go >/dev/null 2>&1; then
#  warn "Go not found, installing golang..."
#  sudo apt-get install -y golang
#  ok "Installed Go: $(go version)"
#else
#  ok "Go is present: $(go version)"
#fi

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

ok "Go services deployed successfully."

# -----------------------
# Node.js check & install
# -----------------------
log "Checking Node.js..."
# Prefer major-version check (23.x), not an exact pin
REQUIRED_NODE_MAJOR="23"

# Get current Node.js version (without leading 'v'), empty if not installed
if command -v node >/dev/null 2>&1; then
    CURRENT_NODE_VERSION=$(node -v | sed 's/^v//')
else
    CURRENT_NODE_VERSION=""
fi

CURRENT_NODE_MAJOR="${CURRENT_NODE_VERSION%%.*}"

if [[ -z "$CURRENT_NODE_VERSION" || "$CURRENT_NODE_MAJOR" -lt "$REQUIRED_NODE_MAJOR" ]]; then
    warn "Node.js not found or older than ${REQUIRED_NODE_MAJOR}.x (current: ${CURRENT_NODE_VERSION:-none}). Installing Node.js 23.x..."
    curl -fsSL https://deb.nodesource.com/setup_23.x | sudo -E bash -
    sudo apt-get install -y nodejs

    CURRENT_NODE_VERSION=$(node -v | sed 's/^v//')
    ok "Node.js installed: v$CURRENT_NODE_VERSION"
else
    ok "Node.js is already installed: v$CURRENT_NODE_VERSION"
fi

# Ensure npm exists
if ! command -v npm >/dev/null 2>&1; then
    warn "npm not found. Installing npm..."
    sudo apt-get install -y npm
fi

sudo npm install -g yarn

# -----------------------
# Build frontend
# -----------------------
cd ./web

log "Cleaning yarn caches..."
yarn cache clean

log "Installing yarn dependencies..."
yarn install

log "Building frontend..."
NODE_ENV=production VITE_I18N_LANGUAGES="${LANGUAGES:-en}" yarn run build
[[ -d dist ]] || die "dist folder not found after build"

# -----------------------
# Install and configure Nginx
# -----------------------
cd - >/dev/null
log "Installing Nginx..."
sudo apt-get install -y nginx
sudo rm -rf /etc/nginx/sites-enabled/default

CERT_DIR="/etc/nginx/certs"
CERT_KEY="${CERT_DIR}/cert.key"
CERT_PEM="${CERT_DIR}/cert.pem"
sudo mkdir -p "$CERT_DIR"

if [[ ! -f "$CERT_KEY" || ! -f "$CERT_PEM" ]]; then
  log "Generating self-signed SSL certificate for Nginx..."
  sudo openssl req -x509 -nodes -days "${SSL_EXPIRE:-365}" -newkey rsa:2048 \
    -keyout "$CERT_KEY" -out "$CERT_PEM" \
    -subj "/C=${SSL_C:-US}/ST=${SSL_ST:-State}/L=${SSL_L:-City}/O=${SSL_ORG:-Org}/OU=${SSL_OU:-Unit}/CN=${SSL_CN:-localhost}"
fi

# Nginx reverse proxy (HTTP redirect :3000 -> :3443; TLS on :3443)
sudo tee /etc/nginx/conf.d/site.conf >/dev/null <<'EOF'
upstream api_backend { server 127.0.0.1:8080; }
upstream log_stream_backend { server 127.0.0.1:8081; }

server {
    listen 3000;
    return 302 https://$host:3443$request_uri;
}

server {
    listen 3443 ssl;
    server_name _;

    ssl_certificate     /etc/nginx/certs/cert.pem;
    ssl_certificate_key /etc/nginx/certs/cert.key;

    location / {
        root /var/www/site;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    location ~ ^/(api) {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /ws/ {
        proxy_pass http://log_stream_backend/;
        proxy_http_version 1.1;
    
        # Keep the connection open for SSE
        proxy_set_header Connection '';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_buffering off;   # Important for SSE (no buffering)
        proxy_cache off;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;

        # Nginx will automatically forward text/event-stream responses
    }
}
EOF

# Deploy frontend
sudo mkdir -p /var/www/site
sudo cp -r web/dist/* /var/www/site
sudo chown -R www-data:www-data /var/www/site

# Test & start Nginx
log "Testing Nginx configuration..."
sudo systemctl daemon-reload
sudo systemctl enable --now nginx.service
sudo systemctl restart nginx.service
sudo nginx -t

if sudo systemctl is-active --quiet nginx; then
    ok "Nginx is running."
else
    die "Nginx failed to start."
fi

# -----------------------
# Install Ocserv (VPN)
# -----------------------
log "Installing Ocserv..."
sudo apt-get install -y ocserv gnutls-bin iptables iptables-persistent

# Generate ocserv certs if missing
if [[ ! -f /etc/ocserv/certs/cert.pem ]]; then
  log "Generating SSL certificates for Ocserv..."
  sudo mkdir -p /etc/ocserv/certs
  sudo touch /etc/ocserv/ocpasswd

  servercert="cert.pem"
  serverkey="key.pem"

  SSL_CN="${SSL_CN:-End-way-Cisco-VPN}"
  SSL_ORG="${SSL_ORG:-End-way}"
  SSL_EXPIRE="${SSL_EXPIRE:-3650}"

  sudo certtool --generate-privkey --outfile ca-key.pem
  cat <<_EOF_ | sudo tee ca.tmpl >/dev/null
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 1
expiration_days = ${SSL_EXPIRE}
ca
signing_key
cert_signing_key
crl_signing_key
_EOF_
  sudo certtool --generate-self-signed --load-privkey ca-key.pem --template ca.tmpl --outfile ca-cert.pem
  sudo certtool --generate-privkey --outfile "${serverkey}"

  cat <<_EOF_ | sudo tee server.tmpl >/dev/null
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 2
expiration_days = ${SSL_EXPIRE}
signing_key
encryption_key
tls_www_server
_EOF_
  sudo certtool --generate-certificate \
    --load-privkey "${serverkey}" \
    --load-ca-certificate ca-cert.pem \
    --load-ca-privkey ca-key.pem \
    --template server.tmpl \
    --outfile "${servercert}"

  sudo cp "${servercert}" /etc/ocserv/certs/cert.pem
  sudo cp "${serverkey}" /etc/ocserv/certs/cert.key
fi

# Configure ocserv
log "Configuring Ocserv..."
sudo tee /etc/ocserv/ocserv.conf >/dev/null <<EOT
# -----------------------
# Ocserv Configuration
# -----------------------
auth = "plain[passwd=/etc/ocserv/ocpasswd]"
run-as-user = root
run-as-group = root

socket-file = /var/run/ocserv-socket
isolate-workers = true
max-clients = 1024

keepalive = 32400
dpd = 90
mobile-dpd = 1800
switch-to-tcp-timeout = 5
try-mtu-discovery = true

server-cert = /etc/ocserv/certs/cert.pem
server-key  = /etc/ocserv/certs/cert.key
tls-priorities = "NORMAL:%SERVER_PRECEDENCE:%COMPAT:-RSA:-VERS-SSL3.0:-ARCFOUR-128"

auth-timeout = 40
min-reauth-time = 300
max-ban-score = 50
ban-reset-time = 300
cookie-timeout = 86400
deny-roaming = false
rekey-time = 172800
rekey-method = ssl

use-occtl = true
pid-file = /var/run/ocserv.pid
log-level = 3
rate-limit-ms = 100

device = vpns
predictable-ips = true
tunnel-all-dns = true
dns = ${OCSERV_DNS}
ping-leases = false
mtu = 1420
cisco-client-compat = true
dtls-legacy = true

tcp-port = ${OCSERV_PORT}
udp-port = ${OCSERV_PORT}

max-same-clients = 2
ipv4-network = ${OC_NET}

config-per-group = /etc/ocserv/groups/
config-per-user  = /etc/ocserv/users/
EOT

sudo mkdir -p /etc/ocserv/defaults /etc/ocserv/groups /etc/ocserv/users
sudo touch /etc/ocserv/defaults/group.conf

# -----------------------
# Firewall / NAT (iptables)
# -----------------------
log "Configuring iptables for VPN NAT/forwarding..."

# Open VPN port for TCP and UDP
sudo iptables -I INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT
sudo iptables -I INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT

# NAT for VPN subnet via external interface
sudo iptables -t nat -A POSTROUTING -s "${OC_NET}" -o "${ETH}" -j MASQUERADE

# Forward VPN traffic out via $ETH, allow return traffic (stateful)
sudo iptables -A FORWARD -s "${OC_NET}" -o "${ETH}" -j ACCEPT
sudo iptables -A FORWARD -d "${OC_NET}" -m state --state ESTABLISHED,RELATED -j ACCEPT

# Preseed persistence prompts and ensure persistence
sudo debconf-set-selections <<EOF
iptables-persistent iptables-persistent/autosave_v4 boolean true
iptables-persistent iptables-persistent/autosave_v6 boolean true
EOF
# (package was installed earlier with ocserv); still save explicitly:
sudo sh -c "iptables-save > /etc/iptables/rules.v4"
sudo sh -c "ip6tables-save > /etc/iptables/rules.v6"
sudo netfilter-persistent save || true

# -----------------------
# Enable IP Forwarding (persistent)
# -----------------------
log "Enabling IP forwarding..."
sudo sysctl -w net.ipv4.ip_forward=1
# Persist safely via /etc/sysctl.d
echo "net.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-ocserv.conf >/dev/null
sudo sysctl --system

# -----------------------
# Start Ocserv
# -----------------------
sudo systemctl daemon-reload
sudo systemctl enable ocserv.service
sudo systemctl restart ocserv.service
if systemctl is-active --quiet ocserv; then
  ok "Ocserv is running."
else
  die "Ocserv failed to start."
fi

ok "Deployment completed successfully!"

# -----------------------
# Cleaning
# -----------------------
log "Cleaning unused packages..."

sudo apt autoremove -y
sudo apt autoclean -y

ok "Cleaning completed."
