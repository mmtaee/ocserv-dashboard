#!/bin/bash

if [ -z "$SSL_CN" ]; then
    SSL_CN="End-way-Cisco-VPN"
fi
if [ -z "$SSL_ORG" ]; then
    SSL_ORG="End-way"
fi
if [ -z "$SSL_EXPIRE" ]; then
    SSL_EXPIRE=3650
fi
if [ -z "$OC_NET" ]; then
    OC_NET=172.16.24.0/24
fi


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
  echo "Preparing Ocserv certificate authentication PKI..."

  mkdir -p "${OCSERV_SSL_DIR}" "${OCSERV_SSL_USERS_DIR}" "${OCSERV_SSL_DISABLED_DIR}"
  chmod 700 "${OCSERV_SSL_DIR}" "${OCSERV_SSL_USERS_DIR}" "${OCSERV_SSL_DISABLED_DIR}"

  if [[ ! -f "${OCSERV_CRL_TMPL}" ]]; then
    tee "${OCSERV_CRL_TMPL}" >/dev/null <<EOT
crl_next_update = 365
crl_number = 1
EOT
    chmod 600 "${OCSERV_CRL_TMPL}"
  fi

  touch "${OCSERV_REVOKED_PEM}" "${OCSERV_SUSPENDED_PEM}"
  chmod 600 "${OCSERV_REVOKED_PEM}" "${OCSERV_SUSPENDED_PEM}"

  if [[ -f "${OCSERV_CA_CERT}" && ! -f "${OCSERV_CA_KEY}" ]] || [[ ! -f "${OCSERV_CA_CERT}" && -f "${OCSERV_CA_KEY}" ]]; then
    echo "Incomplete Ocserv client CA. Both ${OCSERV_CA_CERT} and ${OCSERV_CA_KEY} must exist."
    exit 1
  fi

  if [[ ! -f "${OCSERV_CA_CERT}" ]]; then
    local ca_tmpl="${OCSERV_SSL_DIR}/ca.tmpl"

    tee "${ca_tmpl}" >/dev/null <<EOT
cn = "${SSL_CN:-Ocserv Dashboard CA}"
organization = "${SSL_ORG:-Ocserv Dashboard}"
serial = 1
expiration_days = ${SSL_EXPIRE:-3650}
ca
signing_key
cert_signing_key
crl_signing_key
EOT

    certtool --generate-privkey --outfile "${OCSERV_CA_KEY}"
    certtool --generate-self-signed \
      --load-privkey "${OCSERV_CA_KEY}" \
      --template "${ca_tmpl}" \
      --outfile "${OCSERV_CA_CERT}"

    chmod 600 "${OCSERV_CA_KEY}"
    chmod 644 "${OCSERV_CA_CERT}"
  fi

  if [[ ! -f "${OCSERV_CRL}" ]]; then
    certtool --generate-crl \
      --load-ca-privkey "${OCSERV_CA_KEY}" \
      --load-ca-certificate "${OCSERV_CA_CERT}" \
      --template "${OCSERV_CRL_TMPL}" \
      --outfile "${OCSERV_CRL}"
    chmod 644 "${OCSERV_CRL}"
  fi
}

write_ocserv_conf() {
  echo "Writing Ocserv configuration..."
  cat <<EOT >"$OCSERV_CONF"
# ===============================================
# Managed by ocserv-dashboard install.sh
# DO NOT edit or remove this file header
# ===============================================
auth = "certificate"
enable-auth = "plain[passwd=/etc/ocserv/ocpasswd]"
ca-cert = /etc/ocserv/ssl/ca-cert.pem
crl = /etc/ocserv/ssl/crl.pem
cert-user-oid = 2.5.4.3
run-as-user=root
run-as-group=root
socket-file=/var/run/ocserv-socket
isolate-workers=true
max-clients=1024
keepalive=32400
dpd=90
mobile-dpd=1800
switch-to-tcp-timeout=5
try-mtu-discovery=true
server-cert=/etc/ocserv/certs/cert.pem
server-key=/etc/ocserv/certs/cert.key
tls-priorities="NORMAL:%SERVER_PRECEDENCE:%COMPAT:-RSA:-VERS-SSL3.0:-ARCFOUR-128"
auth-timeout=40
min-reauth-time=300
max-ban-score=50
ban-reset-time=300
cookie-timeout=86400
deny-roaming=false
rekey-time=172800
rekey-method=ssl
use-occtl=true
pid-file=/var/run/ocserv.pid
device=vpns
predictable-ips=true
tunnel-all-dns=true
dns=${OCSERV_DNS}
ping-leases=false
mtu=1420
cisco-client-compat=true
dtls-legacy=true
tcp-port=443
udp-port=443
max-same-clients=2
ipv4-network=${OC_NET}
config-per-group=/etc/ocserv/groups/
config-per-user=/etc/ocserv/users/
log-level=3
rate-limit-ms=100
pre-login-banner="$OCSERV_PRE_LOGIN_BANNER"

EOT

OCSERV_BANNER=$(echo "$OCSERV_BANNER" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')
printf 'banner="%s"\n' "$OCSERV_BANNER" >> "$OCSERV_CONF"
}

ensure_ocserv_certificate_auth_config() {
  local tmp

  sed -i \
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

  cp "$tmp" "$OCSERV_CONF"
  rm -f "$tmp"
}

ensure_ocserv_client_pki

# ------------------------------------------------
# Validate existing config
# ------------------------------------------------
if [[ ! -f "$OCSERV_CONF" ]]; then
    echo "📄 ocserv.conf not found, creating new file"
    write_ocserv_conf
elif ! head -n 5 "$OCSERV_CONF" | grep -q "$MANAGED_HEADER"; then
    echo "⚠️ ocserv.conf not managed by dashboard, overwriting"
    write_ocserv_conf
else
    echo "✅ ocserv.conf already managed, no changes needed"
fi

ensure_ocserv_certificate_auth_config

mkdir -p /etc/ocserv/defaults /etc/ocserv/groups /etc/ocserv/users/

# Ensure parent directory exists
GROUP_CONF="/etc/ocserv/defaults/group.conf"
mkdir -p "$(dirname "$GROUP_CONF")"

if [[ ! -f "$GROUP_CONF" ]]; then
    echo "📄 Creating default group configuration"
    touch "${GROUP_CONF}"
else
    echo "✅ Default group configuration already exists"
fi


if [ ! -f /etc/ocserv/certs/cert.pem ]; then
    mkdir -p /etc/ocserv/certs
    cd /etc/ocserv/certs || exit
    touch /etc/ocserv/ocpasswd
    servercert="cert.pem"
    serverkey="key.pem"
    certtool --generate-privkey --outfile ca-key.pem
    cat <<_EOF_ >ca.tmpl
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 1
expiration_days = ${SSL_EXPIRE}
ca
signing_key
cert_signing_key
crl_signing_key
_EOF_
    certtool --generate-self-signed --load-privkey ca-key.pem \
        --template ca.tmpl --outfile ca-cert.pem
    certtool --generate-privkey --outfile ${serverkey}
    cat <<_EOF_ >server.tmpl
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 2
expiration_days = ${SSL_EXPIRE}
signing_key
encryption_key
tls_www_server
_EOF_
    certtool --generate-certificate --load-privkey ${serverkey} \
        --load-ca-certificate ca-cert.pem --load-ca-privkey ca-key.pem \
        --template server.tmpl --outfile ${servercert} >>/tmp/cert.txt 2>&1
    echo "Server Cert pin: $(grep -r 'pin-sha256' /tmp/cert.txt | tr -d '[:space:]')" >>/etc/ocserv/public_key_pin
    echo "Docker Host ip: $(hostname -i)" >>/etc/ocserv/public_key_pin
    rm -rf /tmp/cert.txt
    cp "${servercert}" /etc/ocserv/certs/cert.pem
    cp "${serverkey}" /etc/ocserv/certs/cert.key
fi

# Enable IP forwarding (runtime)
sysctl -w net.ipv4.ip_forward=1

# Detect interface
ETH=$(ip route | awk '/default/ {print $5; exit}')
ETH=${ETH:-eth0}

# NAT only VPN subnet
iptables -t nat -A POSTROUTING -s "$OC_NET" -o "$ETH" -j MASQUERADE

# Allow VPN traffic forwarding
iptables -A FORWARD -s "$OC_NET" -o "$ETH" -j ACCEPT
iptables -A FORWARD -d "$OC_NET" -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow VPN port
iptables -A INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT
iptables -A INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT

# VPN over MTU networks
iptables -A FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

mkdir -p /dev/net               #TUN device
if [ ! -c /dev/net/tun ]; then
    mknod /dev/net/tun c 10 200
fi

chmod 600 /dev/net/tun

cd /usr/local/bin || exit # restore state to app workdir

exec "$@"