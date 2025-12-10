#!/bin/bash

set -euo pipefail
export DEBIAN_FRONTEND=noninteractive

source "$(dirname "$0")/lib.sh"

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
