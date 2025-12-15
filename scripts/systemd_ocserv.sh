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
log "Installing Ocserv and dependencies..."
sudo apt-get install -y ocserv gnutls-bin iptables iptables-persistent

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
log "Writing Ocserv configuration..."

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

# ==============================================================
# 4. Firewall Rules / NAT
# ==============================================================
log "Configuring iptables firewall and NAT..."

# Allow Ocserv ports
sudo iptables -I INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT
sudo iptables -I INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT

# NAT for VPN subnet
sudo iptables -t nat -A POSTROUTING -s "${OC_NET}" -o "${ETH}" -j MASQUERADE

# Forward outbound + inbound VPN connections
sudo iptables -A FORWARD -s "${OC_NET}" -o "${ETH}" -j ACCEPT
sudo iptables -A FORWARD -d "${OC_NET}" -m state --state ESTABLISHED,RELATED -j ACCEPT

# Save firewall rules
sudo debconf-set-selections <<EOF
iptables-persistent iptables-persistent/autosave_v4 boolean true
iptables-persistent iptables-persistent/autosave_v6 boolean true
EOF

sudo sh -c "iptables-save > /etc/iptables/rules.v4"
sudo sh -c "ip6tables-save > /etc/iptables/rules.v6"
sudo netfilter-persistent save || true

# ==============================================================
# 5. Enable Kernel Forwarding
# ==============================================================
log "Enabling IP forwarding..."

sudo sysctl -w net.ipv4.ip_forward=1
# Persist safely via /etc/sysctl.d
echo "net.ipv4.ip_forward = 1" | sudo tee /etc/sysctl.d/99-ocserv.conf >/dev/null
sudo sysctl --system

# ==============================================================
# 6. Start & Enable Ocserv Service
# ==============================================================
log "Starting Ocserv..."

sudo systemctl daemon-reload
sudo systemctl enable ocserv.service
sudo systemctl restart ocserv.service

if systemctl is-active --quiet ocserv; then
  ok "Ocserv is running."
else
  die "Ocserv failed to start."
fi

ok "Ocserv VPN deployment completed successfully!"

# ==============================================================
# 7. Cleanup
# ==============================================================
log "Cleaning apt caches..."
sudo apt autoremove -y
sudo apt autoclean -y

ok "Cleanup completed."
