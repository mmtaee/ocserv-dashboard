#!/bin/bash

set -euo pipefail
export DEBIAN_FRONTEND=noninteractive

source "$(dirname "$0")/lib.sh"

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