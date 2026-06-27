#!/bin/bash

# ==============================================================
# Script: ui.sh
# Description:
#   Handles admin and customer frontend build + Nginx TLS reverse proxy deployment
#   for the ocserv_dashboard systemd-based installation.
#
#   Responsibilities:
#     - Load shared logging helpers (from lib.sh)
#     - Build admin and customer frontends (Vite)
#     - Install & configure Nginx
#     - Reverse proxy Admin API (:8080) and Customer API (:8081)
#     - Deploy compiled frontends into /var/www/admin and /var/www/customer
#
# Prerequisites:
#   - Must be executed from deploy/systemd directory
#   - `lib.sh` must exist at ./lib.sh
#   - Requires root or sudo privileges
#
# Usage:
#   ./ui.sh
#
# ==============================================================

# ==========================================
# Load shared logging helpers
# ==========================================
source ./lib.sh

log "Starting frontend deployment..."

# ==========================================
# Function: build_frontend
# Description:
#   Builds the Vite-based frontend.
#   Parameters:
#     $1 - Frontend directory (../../admin_dashboard/ui or ../../customer_dashboard/ui)
#     $2 - Destination directory (/var/www/admin or /var/www/customer)
#   Output:
#     - Compiled frontend at destination
# ==========================================
build_frontend() {
  local frontend_dir=$1
  local dest_dir=$2

  cd "$frontend_dir" || exit 1

  log "Building frontend in $frontend_dir ..."

  npm install
  npm run build

  [[ -d dist ]] || die "dist folder not found after npm run build"
  ok "Frontend build completed"

  log "Deploying frontend to $dest_dir ..."
  sudo mkdir -p "$dest_dir"
  sudo cp -r dist/* "$dest_dir"
  sudo chown -R www-data:www-data "$dest_dir"

  cd - >/dev/null || exit 1
}

build_frontend "../../admin_dashboard/ui" "/var/www/admin"
build_frontend "../../customer_dashboard/ui" "/var/www/customer"

# ==========================================
# Function: setup_nginx
# Description:
#   Installs Nginx and configures:
#     - TLS using self-signed certificate
#     - Static serving of /var/www/admin and /var/www/customer
#     - Reverse proxy to:
#         * Admin API backend (127.0.0.1:8080)
#         * Customer API backend (127.0.0.1:8081)
#
#   Also deploys compiled frontend assets.
# ==========================================
setup_nginx() {
  log "Installing Nginx..."
  sudo apt-get install -y nginx
  sudo rm -rf /etc/nginx/sites-enabled/default 2>/dev/null || true

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
  sudo tee /etc/nginx/conf.d/ocserv-dashboard.conf >/dev/null <<'EOF'
upstream admin_api_backend { server 127.0.0.1:8080; }
upstream customer_api_backend { server 127.0.0.1:8081; }

server {
    listen 80;
    server_name _;

    # Admin UI at /admin
    location /admin {
        alias /var/www/admin;
        index index.html;
        try_files $uri $uri/ /admin/index.html;
    }

    # Customer UI at root
    location / {
        alias /var/www/customer;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # Admin API
    location ~ ^/admin/api {
        rewrite ^/admin(/api.*)$ $1 break;
        proxy_pass http://admin_api_backend;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Customer API
    location ~ ^/customer/api {
        rewrite ^/customer(/api.*)$ $1 break;
        proxy_pass http://customer_api_backend;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 443 ssl;
    server_name _;

    ssl_certificate     /etc/nginx/certs/cert.pem;
    ssl_certificate_key /etc/nginx/certs/cert.key;

    # Admin UI at /admin
    location /admin {
        alias /var/www/admin;
        index index.html;
        try_files $uri $uri/ /admin/index.html;
    }

    # Customer UI at root
    location / {
        alias /var/www/customer;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # Admin API
    location ~ ^/admin/api {
        rewrite ^/admin(/api.*)$ $1 break;
        proxy_pass http://admin_api_backend;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Customer API
    location ~ ^/customer/api {
        rewrite ^/customer(/api.*)$ $1 break;
        proxy_pass http://customer_api_backend;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Host $http_host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

  # Validate Nginx
  log "Testing Nginx configuration..."
  sudo nginx -t

  # Restart Nginx
  sudo systemctl daemon-reload
  sudo systemctl enable --now nginx.service
  sudo systemctl restart nginx.service

  if sudo systemctl is-active --quiet nginx; then
      ok "Nginx is running."
  else
      die "Nginx failed to start."
  fi
}

setup_nginx

ok "Frontend deployment completed successfully."
