#!/bin/bash

# ==============================================================
# Script: systemd_ui.sh
# Description:
#   Handles frontend build + Nginx TLS reverse proxy deployment
#   for the ocserv_user_management systemd-based installation.
#
#   Responsibilities:
#     - Load shared logging helpers (from lib.sh)
#     - Ensure Node.js + Yarn are installed
#     - Build the frontend (Vite)
#     - Install & configure Nginx (HTTPS on port 3443)
#     - Reverse proxy API (:8080) and log-stream SSE (:8081)
#     - Deploy compiled frontend into /var/www/site
#
# Prerequisites:
#   - Must be executed from the project root
#   - `lib.sh` must exist at ./script/lib.sh
#   - Node.js and Yarn will auto-install if missing
#   - Requires root or sudo privileges
#
# Usage:
#   ./script/systemd_ui.sh
#
# ==============================================================

# ==========================================
# Load shared logging helpers (print_message)
# ==========================================
source ./scripts/lib.sh

log "Starting frontend deployment..."

# ==========================================
# Function: ensure_node
# Description:
#   Ensures Node.js v23.x or higher exists.
#   Installs Node.js via NodeSource if missing or outdated.
#   Installs npm if missing.
#   Installs Yarn globally.
# ==========================================
ensure_node() {
  log "Checking Node.js..."
  REQUIRED_NODE_MAJOR="23"

  if command -v node >/dev/null 2>&1; then
      CURRENT_NODE_VERSION=$(node -v | sed 's/^v//')
  else
      CURRENT_NODE_VERSION=""
  fi

  CURRENT_NODE_MAJOR="${CURRENT_NODE_VERSION%%.*}"

  if [[ -z "$CURRENT_NODE_VERSION" || "$CURRENT_NODE_MAJOR" -lt "$REQUIRED_NODE_MAJOR" ]]; then
      warn "Node.js missing or outdated (current: ${CURRENT_NODE_VERSION:-none}). Installing Node.js ${REQUIRED_NODE_MAJOR}.x..."
      curl -fsSL https://deb.nodesource.com/setup_23.x | sudo -E bash -
      sudo apt-get install -y nodejs
      CURRENT_NODE_VERSION=$(node -v | sed 's/^v//')
      ok "Node.js installed: v$CURRENT_NODE_VERSION"
  else
      ok "Node.js is already installed: v$CURRENT_NODE_VERSION"
  fi

  if ! command -v npm >/dev/null 2>&1; then
      warn "npm not found. Installing..."
      sudo apt-get install -y npm
  fi

  sudo npm install -g yarn
  ok "Yarn installed"
}

ensure_node

# ==========================================
# Function: build_frontend
# Description:
#   Builds the Vite-based frontend.
#   Steps:
#     - Clean Yarn cache
#     - Install dependencies
#     - Build with environment variables
#   Output:
#     - ./web/dist directory
# ==========================================
build_frontend() {
  cd ./web || exit 1

  log "Cleaning yarn cache..."
  yarn cache clean

  log "Installing dependencies..."
  yarn install

  log "Building frontend..."
  NODE_ENV=production \
  VITE_I18N_LANGUAGES="${LANGUAGES:-en}" \
  yarn run build

  [[ -d dist ]] || die "dist folder not found after yarn build"
  ok "Frontend build completed"

  cd - >/dev/null || exit 1
}

build_frontend

# ==========================================
# Function: setup_nginx
# Description:
#   Installs Nginx and configures:
#     - HTTP → HTTPS redirect (3000 → 3443)
#     - TLS using self-signed certificate
#     - Static serving of /var/www/site
#     - Reverse proxy to:
#         * API backend (127.0.0.1:8080)
#         * Log stream SSE backend (127.0.0.1:8081)
#
#   Also deploys compiled frontend assets.
# ==========================================
setup_nginx() {
  log "Installing Nginx..."
  sudo apt-get install -y nginx
  sudo rm -rf /etc/nginx/sites-enabled/default

  CERT_DIR="/etc/nginx/certs"
  CERT_KEY="${CERT_DIR}/cert.key"
  CERT_PEM="${CERT_DIR}/cert.pem"
  sudo mkdir -p "$CERT_DIR"

  # Create cert if missing
  if [[ ! -f "$CERT_KEY" || ! -f "$CERT_PEM" ]]; then
    log "Generating self-signed SSL certificate..."
    sudo openssl req -x509 -nodes -days "${SSL_EXPIRE:-365}" -newkey rsa:2048 \
      -keyout "$CERT_KEY" -out "$CERT_PEM" \
      -subj "/C=${SSL_C:-US}/ST=${SSL_ST:-State}/L=${SSL_L:-City}/O=${SSL_ORG:-Org}/OU=${SSL_OU:-Unit}/CN=${SSL_CN:-localhost}"
  fi

  # Write Nginx config
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
        proxy_set_header Connection '';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
    }
}
EOF

  # Deploy frontend assets
  sudo mkdir -p /var/www/site
  sudo cp -r web/dist/* /var/www/site
  sudo chown -R www-data:www-data /var/www/site

  # Validate Nginx
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
}

setup_nginx

ok "Frontend deployment completed successfully."
