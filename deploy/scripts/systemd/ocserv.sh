#!/bin/bash
# ======================================================================
# Script: systemd_ocserv.sh
# Description:
#   Installs and configures Ocserv (OpenConnect VPN server),
#   generates SSL certificates if missing, configures iptables NAT,
#   enables persistent kernel forwarding, and activates ocserv.service.
#
# Environment variables (optional):
#   LANGUAGES   → List of frontend languages (default: "en")
#   SSL_*       → Certificate metadata for Nginx self-signed cert
#   OCSERV_PORT → For reference when redirecting traffic (not required)
#   ETH         → External interface (auto-detected if missing)
#
# Usage:
#   sudo ./systemd_ocserv.sh
#
# Requirements:
#   - Must be run as root or with sudo
#   - Debian / Ubuntu system with apt package manager
# ======================================================================

# ==============================================================
# Load shared logging utilities
# (print_message, log, ok, warn, die are defined in lib.sh)
# ==============================================================
source ./scripts/lib.sh

# Sensible defaults (can be overridden via environment)
OCSERV_PORT="${OCSERV_PORT:-443}"              # ocserv TCP/UDP port; 443 is typical
OC_NET="${OC_NET:-172.16.24.0/24}"             # VPN IPv4 subnet
OCSERV_DNS="${OCSERV_DNS:-1.1.1.1}"           # DNS pushed to clients
ETH="${ETH:-}"                                 # External interface (auto-detect if empty)


# ==========================================
# Function: auto_detect_interface
# Description:
#   Detect the primary outbound interface automatically.
#   Fails if detection is impossible.
# ==========================================
auto_detect_interface() {
  if [[ -z "${ETH}" ]]; then
    ETH="$(ip -o -4 route show to default 2>/dev/null | awk '{print $5}' | head -n1 || true)"
    [[ -n "${ETH}" ]] || die "Could not detect external interface. Set ETH manually (e.g., ETH=eth0)."
    log "Auto-detected external interface: ${ETH}"
  fi
}
auto_detect_interface

# ==============================================================
# 1. Install Ocserv + Required Tools
# ==============================================================
log "Installing Ocserv..."

compile_and_install_ocserv() {
    # Install build package dependencies
    sudo apt install -y build-essential meson ninja-build pkg-config git \
                        libgnutls28-dev libev-dev libreadline-dev libtasn1-bin \
                        libpam0g-dev liblz4-dev libseccomp-dev \
                        libnl-route-3-dev libkrb5-dev libradcli-dev \
                        libcurl4-gnutls-dev libcjose-dev libjansson-dev liboath-dev \
                        libprotobuf-c-dev libtalloc-dev libllhttp-dev protobuf-c-compiler \
                        gperf ipcalc

    INSTALL_PREFIX="/usr"
    SRC_DIR="/tmp/ocserv"

    info "Preparing source directory..."

    rm -rf "$SRC_DIR"
    git clone --depth=1 https://gitlab.com/openconnect/ocserv.git "$SRC_DIR"

    cd "$SRC_DIR" || exit 1

    info "Configuring build (Meson)..."
    meson setup build \
        --prefix="$INSTALL_PREFIX" \
        --sysconfdir=/etc

    info "Compiling..."
    meson compile -C build -j"$(nproc)"

    info "Installing..."
    meson install -C build

    # Cleanup build artifacts
    info "Cleaning build files..."
    cd /
    rm -rf "$SRC_DIR"

    # Minimal runtime setup
    info "Creating runtime dirs..."
    sudo mkdir -p /etc/ocserv /var/run/ocserv

    info "Adding ocserv user..."
    id -u ocserv &>/dev/null || sudo useradd -r -s /usr/sbin/nologin ocserv

    info "Copying default config..."
    if [ -f /usr/share/doc/ocserv/examples/sample.config ] || [ -f doc/sample.config ]; then
        sudo cp doc/sample.config /etc/ocserv/ocserv.conf 2>/dev/null || true
    fi

    # Optional: shrink binary
    if command -v strip &>/dev/null; then
        info "Stripping binary..."
        sudo strip /usr/sbin/ocserv || true
    fi

    # Setup system unit
    cat <<'EOF' | sudo tee /etc/systemd/system/ocserv.service > /dev/null
[Unit]
Description=OpenConnect SSL VPN server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/sbin/ocserv --foreground --config /etc/ocserv/ocserv.conf
ExecReload=/bin/kill -HUP $MAINPID
PIDFile=/var/run/ocserv.pid
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    info "Systemd service created successfully"
    ok "Ocserv installed successfully"
}

if compile_and_install_ocserv; then
    log "ocserv installed successfully from source."
else
    sudo apt install -y ocserv
fi

log "Installing dependencies..."
sudo apt-get install -y gnutls-bin openssl iptables iptables-persistent

# ==============================================================
# 2. Generate Ocserv Certificates (If Missing)
# ==============================================================
# Function: generate_ocserv_certs
# Description:
#   Creates CA + server certificates needed by Ocserv.
#   Only runs if cert.pem does NOT exist.
# ==============================================================
generate_ocserv_certs() {
  log "Generating SSL certificates for Ocserv..."

  sudo mkdir -p /etc/ocserv/certs
  sudo touch /etc/ocserv/ocpasswd

  SERVER_CERT="cert.pem"
  SERVER_KEY="key.pem"

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

  sudo certtool --generate-self-signed \
    --load-privkey ca-key.pem \
    --template ca.tmpl \
    --outfile ca-cert.pem

  sudo certtool --generate-privkey --outfile "${SERVER_KEY}"

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
    --load-privkey "${SERVER_KEY}" \
    --load-ca-certificate ca-cert.pem \
    --load-ca-privkey ca-key.pem \
    --template server.tmpl \
    --outfile "${SERVER_CERT}"

  sudo cp "${SERVER_CERT}" /etc/ocserv/certs/cert.pem
  sudo cp "${SERVER_KEY}" /etc/ocserv/certs/cert.key
}

# Only generate if missing
if [[ ! -f /etc/ocserv/certs/cert.pem ]]; then
  generate_ocserv_certs
fi

# ==============================================================
# 3. Ocserv Main Configuration
# ==============================================================
OCSERV_CONF="/etc/ocserv/ocserv.conf"
MANAGED_HEADER="# Managed by ocserv-dashboard install.sh"

OCSERV_SSL_DIR="/etc/ocserv/ssl"
OCSERV_SSL_USERS_DIR="${OCSERV_SSL_DIR}/users"
OCSERV_SSL_DISABLED_DIR="${OCSERV_SSL_DIR}/disabled"
OCSERV_CA_CERT="${OCSERV_SSL_DIR}/ca-cert.pem"
OCSERV_CA_KEY="${OCSERV_SSL_DIR}/ca-key.pem"
OCSERV_CRL="${OCSERV_SSL_DIR}/crl.pem"
OCSERV_CRL_TMPL="${OCSERV_SSL_DIR}/crl.tmpl"
OCSERV_REVOKED_PEM="${OCSERV_SSL_DIR}/revoked.pem"
OCSERV_SUSPENDED_PEM="${OCSERV_SSL_DIR}/suspended.pem"

ensure_ocserv_client_pki() {
  log "Preparing Ocserv certificate authentication PKI..."

  sudo mkdir -p "${OCSERV_SSL_DIR}" "${OCSERV_SSL_USERS_DIR}" "${OCSERV_SSL_DISABLED_DIR}"
  sudo chmod 700 "${OCSERV_SSL_DIR}" "${OCSERV_SSL_USERS_DIR}" "${OCSERV_SSL_DISABLED_DIR}"

  if [[ ! -f "${OCSERV_CRL_TMPL}" ]]; then
    sudo tee "${OCSERV_CRL_TMPL}" >/dev/null <<EOT
crl_next_update = 365
crl_number = 1
EOT
    sudo chmod 600 "${OCSERV_CRL_TMPL}"
  fi

  sudo touch "${OCSERV_REVOKED_PEM}" "${OCSERV_SUSPENDED_PEM}"
  sudo chmod 600 "${OCSERV_REVOKED_PEM}" "${OCSERV_SUSPENDED_PEM}"

  if [[ -f "${OCSERV_CA_CERT}" && ! -f "${OCSERV_CA_KEY}" ]] || [[ ! -f "${OCSERV_CA_CERT}" && -f "${OCSERV_CA_KEY}" ]]; then
    die "Incomplete Ocserv client CA. Both ${OCSERV_CA_CERT} and ${OCSERV_CA_KEY} must exist."
  fi

  if [[ ! -f "${OCSERV_CA_CERT}" ]]; then
    local ca_tmpl="${OCSERV_SSL_DIR}/ca.tmpl"

    sudo tee "${ca_tmpl}" >/dev/null <<EOT
cn = "${SSL_CN:-Ocserv Dashboard CA}"
organization = "${SSL_ORG:-Ocserv Dashboard}"
serial = 1
expiration_days = ${SSL_EXPIRE:-3650}
ca
signing_key
cert_signing_key
crl_signing_key
EOT

    sudo certtool --generate-privkey --outfile "${OCSERV_CA_KEY}"
    sudo certtool --generate-self-signed \
      --load-privkey "${OCSERV_CA_KEY}" \
      --template "${ca_tmpl}" \
      --outfile "${OCSERV_CA_CERT}"

    sudo chmod 600 "${OCSERV_CA_KEY}"
    sudo chmod 644 "${OCSERV_CA_CERT}"
  fi

  if [[ ! -f "${OCSERV_CRL}" ]]; then
    sudo certtool --generate-crl \
      --load-ca-privkey "${OCSERV_CA_KEY}" \
      --load-ca-certificate "${OCSERV_CA_CERT}" \
      --template "${OCSERV_CRL_TMPL}" \
      --outfile "${OCSERV_CRL}"
    sudo chmod 644 "${OCSERV_CRL}"
  fi
}

write_ocserv_conf_systemd() {
  log "Writing Ocserv configuration..."
  sudo tee "$OCSERV_CONF" >/dev/null <<EOT
# ===============================================
# Managed by ocserv-dashboard install.sh
# DO NOT edit or remove this file header
# ===============================================

auth = "certificate"
enable-auth = "plain[passwd=/etc/ocserv/ocpasswd]"
ca-cert = /etc/ocserv/ssl/ca-cert.pem
crl = /etc/ocserv/ssl/crl.pem
cert-user-oid = 2.5.4.3
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

pre-login-banner="${OCSERV_PRE_LOGIN_BANNER}"
EOT

OCSERV_BANNER=$(echo "$OCSERV_BANNER" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')
printf 'banner = "%s"\n' "$OCSERV_BANNER" | sudo tee -a "$OCSERV_CONF" > /dev/null
}

ensure_ocserv_certificate_auth_config() {
  local tmp

  sudo sed -i \
    -e '/^\s*auth\s*=/d' \
    -e '/^\s*enable-auth\s*=/d' \
    -e '/^\s*ca-cert\s*=/d' \
    -e '/^\s*crl\s*=/d' \
    -e '/^\s*cert-user-oid\s*=/d' \
    "$OCSERV_CONF"

  tmp="$(mktemp)"
  awk 'NR == 6 {
    print "auth = \"certificate\""
    print "enable-auth = \"plain[passwd=/etc/ocserv/ocpasswd]\""
    print "ca-cert = /etc/ocserv/ssl/ca-cert.pem"
    print "crl = /etc/ocserv/ssl/crl.pem"
    print "cert-user-oid = 2.5.4.3"
  }
  { print }' "$OCSERV_CONF" > "$tmp"

  sudo cp "$tmp" "$OCSERV_CONF"
  rm -f "$tmp"
}

ensure_ocserv_client_pki

if [[ ! -f "$OCSERV_CONF" ]]; then
    info "📄 ocserv.conf not found, creating new systemd config"
    write_ocserv_conf_systemd
elif ! head -n 5 "$OCSERV_CONF" | grep -q "$MANAGED_HEADER"; then
    warn "⚠️ ocserv.conf not managed by dashboard, overwriting"
    write_ocserv_conf_systemd
else
    ok "✅ ocserv.conf already managed (systemd mode)"
fi

ensure_ocserv_certificate_auth_config

sudo mkdir -p /etc/ocserv/defaults /etc/ocserv/groups /etc/ocserv/users

# Ensure parent directory exists
GROUP_CONF="/etc/ocserv/defaults/group.conf"
sudo mkdir -p "$(dirname "$GROUP_CONF")"

if [[ ! -f "$GROUP_CONF" ]]; then
    info "📄 Creating default group configuration"
    sudo touch "${GROUP_CONF}"
else
    ok "✅ Default group configuration already exists"
fi

# ==============================================================
# 4. Enable Kernel Forwarding
# ==============================================================
log "Enabling IP forwarding..."

sudo sysctl -w net.ipv4.ip_forward=1
# Persist safely via /etc/sysctl.d
echo "net.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-ocserv.conf >/dev/null
sudo sysctl --system

# ==============================================================
# 5. Firewall Rules / NAT
# ==============================================================
log "Configuring firewall..."

# Allow VPN ports
sudo iptables -C INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT 2>/dev/null || \
sudo iptables -A INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT

sudo iptables -C INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT 2>/dev/null || \
sudo iptables -A INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT

# NAT VPN subnet
sudo iptables -t nat -C POSTROUTING -s "${OC_NET}" -o "${ETH}" -j MASQUERADE 2>/dev/null || \
sudo iptables -t nat -A POSTROUTING -s "${OC_NET}" -o "${ETH}" -j MASQUERADE

# Forward rules
sudo iptables -C FORWARD -s "${OC_NET}" -o "${ETH}" -j ACCEPT 2>/dev/null || \
sudo iptables -A FORWARD -s "${OC_NET}" -o "${ETH}" -j ACCEPT

sudo iptables -C FORWARD -d "${OC_NET}" -m state --state ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || \
sudo iptables -A FORWARD -d "${OC_NET}" -m state --state ESTABLISHED,RELATED -j ACCEPT

sudo iptables -A FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

log "Saving firewall rules..."
#iptables-save | sudo tee /etc/iptables/rules.v4 > /dev/null
sudo sh -c "iptables-save > /etc/iptables/rules.v4"

#sudo netfilter-persistent save || true

# ==============================================================
# 6. Start & Enable Ocserv Service
# ==============================================================
info "Enabling and starting systemd service"

export PATH="/usr/sbin:$PATH"

sudo systemctl daemon-reload
sudo systemctl enable ocserv.service
sudo systemctl restart ocserv.service

OCSERV_VERSION=$(ocserv --version | head -n 1)

info "ocserv ${OCSERV_VERSION} installed successfully!"
info "Binary: /usr/local/sbin/ocserv"
info "Config: /etc/ocserv/ocserv.conf"

if systemctl is-active --quiet ocserv; then
  ok "Ocserv is running."
else
  die "Ocserv failed to start."
fi

ok "Ocserv VPN deployment completed successfully!"

