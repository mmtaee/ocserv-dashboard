#!/usr/bin/env bash

set -Eeuo pipefail

: "${OCSERV_PORT:=443}"
: "${OC_NET:=172.16.24.0/24}"
: "${OCSERV_DNS:=1.1.1.1}"
: "${SSL_CN:=ocserv-dashboard}"
: "${SSL_ORG:=ocserv-dashboard}"
: "${SSL_EXPIRE:=3650}"
: "${OCSERV_PRESERVE_CONFIG:=false}"

readonly OCSERV_CONF=/etc/ocserv/ocserv.conf
readonly OCSERV_SSL_DIR=/etc/ocserv/ssl
readonly OCSERV_CERTS_DIR=/etc/ocserv/certs
readonly OCSERV_CA_CERT="${OCSERV_SSL_DIR}/ca-cert.pem"
readonly OCSERV_CA_KEY="${OCSERV_SSL_DIR}/ca-key.pem"
readonly OCSERV_CRL="${OCSERV_SSL_DIR}/crl.pem"

log() {
    printf '[container] %s\n' "$*"
}

is_true() {
    case "${1,,}" in
        1|true|yes|on) return 0 ;;
        *) return 1 ;;
    esac
}

ensure_client_pki() {
    local ca_template="${OCSERV_SSL_DIR}/ca.tmpl"
    local crl_template="${OCSERV_SSL_DIR}/crl.tmpl"

    mkdir -p "${OCSERV_SSL_DIR}/users" "${OCSERV_SSL_DIR}/disabled"
    chmod 700 "${OCSERV_SSL_DIR}" "${OCSERV_SSL_DIR}/users" "${OCSERV_SSL_DIR}/disabled"
    touch "${OCSERV_SSL_DIR}/revoked.pem" "${OCSERV_SSL_DIR}/suspended.pem"
    chmod 600 "${OCSERV_SSL_DIR}/revoked.pem" "${OCSERV_SSL_DIR}/suspended.pem"

    if [[ -f "${OCSERV_CA_CERT}" && ! -f "${OCSERV_CA_KEY}" ]] || \
       [[ ! -f "${OCSERV_CA_CERT}" && -f "${OCSERV_CA_KEY}" ]]; then
        log "client CA is incomplete; both ${OCSERV_CA_CERT} and ${OCSERV_CA_KEY} are required"
        exit 1
    fi

    if [[ ! -f "${OCSERV_CA_CERT}" ]]; then
        log "generating OCServ client certificate authority"
        cat >"${ca_template}" <<EOF
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 1
expiration_days = ${SSL_EXPIRE}
ca
signing_key
cert_signing_key
crl_signing_key
EOF
        certtool --generate-privkey --outfile "${OCSERV_CA_KEY}"
        certtool --generate-self-signed \
            --load-privkey "${OCSERV_CA_KEY}" \
            --template "${ca_template}" \
            --outfile "${OCSERV_CA_CERT}"
        chmod 600 "${OCSERV_CA_KEY}"
        chmod 644 "${OCSERV_CA_CERT}"
    fi

    if [[ ! -f "${OCSERV_CRL}" ]]; then
        cat >"${crl_template}" <<EOF
crl_next_update = 365
crl_number = 1
EOF
        certtool --generate-crl \
            --load-ca-privkey "${OCSERV_CA_KEY}" \
            --load-ca-certificate "${OCSERV_CA_CERT}" \
            --template "${crl_template}" \
            --outfile "${OCSERV_CRL}"
        chmod 644 "${OCSERV_CRL}"
    fi
}

ensure_server_certificate() {
    local ca_template="${OCSERV_CERTS_DIR}/ca.tmpl"
    local server_template="${OCSERV_CERTS_DIR}/server.tmpl"

    mkdir -p "${OCSERV_CERTS_DIR}"
    if [[ -f "${OCSERV_CERTS_DIR}/cert.pem" && -f "${OCSERV_CERTS_DIR}/cert.key" ]]; then
        return
    fi

    log "generating OCServ server certificate"
    cat >"${ca_template}" <<EOF
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 1
expiration_days = ${SSL_EXPIRE}
ca
signing_key
cert_signing_key
crl_signing_key
EOF
    certtool --generate-privkey --outfile "${OCSERV_CERTS_DIR}/ca-key.pem"
    certtool --generate-self-signed \
        --load-privkey "${OCSERV_CERTS_DIR}/ca-key.pem" \
        --template "${ca_template}" \
        --outfile "${OCSERV_CERTS_DIR}/ca-cert.pem"

    cat >"${server_template}" <<EOF
cn = "${SSL_CN}"
organization = "${SSL_ORG}"
serial = 2
expiration_days = ${SSL_EXPIRE}
signing_key
encryption_key
tls_www_server
EOF
    certtool --generate-privkey --outfile "${OCSERV_CERTS_DIR}/cert.key"
    certtool --generate-certificate \
        --load-privkey "${OCSERV_CERTS_DIR}/cert.key" \
        --load-ca-certificate "${OCSERV_CERTS_DIR}/ca-cert.pem" \
        --load-ca-privkey "${OCSERV_CERTS_DIR}/ca-key.pem" \
        --template "${server_template}" \
        --outfile "${OCSERV_CERTS_DIR}/cert.pem"
    chmod 600 "${OCSERV_CERTS_DIR}/ca-key.pem" "${OCSERV_CERTS_DIR}/cert.key"
    chmod 644 "${OCSERV_CERTS_DIR}/ca-cert.pem" "${OCSERV_CERTS_DIR}/cert.pem"
}

write_ocserv_config() {
    local banner="${OCSERV_BANNER:-}"
    local pre_login_banner="${OCSERV_PRE_LOGIN_BANNER:-}"

    banner="${banner//\\/\\\\}"
    banner="${banner//\"/\\\"}"
    banner="${banner//$'\n'/\\n}"
    pre_login_banner="${pre_login_banner//\\/\\\\}"
    pre_login_banner="${pre_login_banner//\"/\\\"}"
    pre_login_banner="${pre_login_banner//$'\n'/\\n}"

    log "writing OCServ configuration"
    cat >"${OCSERV_CONF}" <<EOF
# Managed by the ocserv-dashboard container
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
server-key = /etc/ocserv/certs/cert.key
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
config-per-user = /etc/ocserv/users/
log-level = 3
rate-limit-ms = 100
pre-login-banner = "${pre_login_banner}"
banner = "${banner}"
EOF
}

ensure_ocserv_layout() {
    mkdir -p /etc/ocserv/defaults /etc/ocserv/groups /etc/ocserv/users
    touch /etc/ocserv/defaults/group.conf /etc/ocserv/ocpasswd
    chmod 600 /etc/ocserv/ocpasswd

    ensure_client_pki
    ensure_server_certificate

    if [[ ! -f "${OCSERV_CONF}" ]] || \
       is_true "${OCSERV_REGENERATE_CONFIG:-false}" || \
       { ! grep -q '^# Managed by the ocserv-dashboard container$' "${OCSERV_CONF}" && \
         ! is_true "${OCSERV_PRESERVE_CONFIG}"; }; then
        write_ocserv_config
    else
        log "preserving existing ${OCSERV_CONF}"
    fi
}

ensure_networking() {
    local external_interface

    if ! sysctl -w net.ipv4.ip_forward=1 >/dev/null; then
        log "cannot enable IP forwarding; run the container with NET_ADMIN/privileged permissions"
        exit 1
    fi

    external_interface="$(ip route | awk '/default/ {print $5; exit}')"
    external_interface="${external_interface:-eth0}"

    iptables -t nat -C POSTROUTING -s "${OC_NET}" -o "${external_interface}" -j MASQUERADE 2>/dev/null || \
        iptables -t nat -A POSTROUTING -s "${OC_NET}" -o "${external_interface}" -j MASQUERADE
    iptables -C FORWARD -s "${OC_NET}" -o "${external_interface}" -j ACCEPT 2>/dev/null || \
        iptables -A FORWARD -s "${OC_NET}" -o "${external_interface}" -j ACCEPT
    iptables -C FORWARD -d "${OC_NET}" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || \
        iptables -A FORWARD -d "${OC_NET}" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
    iptables -C INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT 2>/dev/null || \
        iptables -A INPUT -p tcp --dport "${OCSERV_PORT}" -j ACCEPT
    iptables -C INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT 2>/dev/null || \
        iptables -A INPUT -p udp --dport "${OCSERV_PORT}" -j ACCEPT
    iptables -C FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu 2>/dev/null || \
        iptables -A FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

    mkdir -p /dev/net
    if [[ ! -c /dev/net/tun ]]; then
        mknod /dev/net/tun c 10 200
    fi
    chmod 600 /dev/net/tun
}

main() {
    ensure_ocserv_layout
    ensure_networking
    exec "$@"
}

main "$@"
