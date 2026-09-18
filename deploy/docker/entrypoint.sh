#!/usr/bin/env bash

# Shared Ocserv setup used by master, agent, and systemd deployments.

set -Eeuo pipefail

strip_surrounding_quotes() {
    local value="$1"

    if [[ "${#value}" -ge 2 ]] && \
       { [[ "${value:0:1}" == '"' && "${value: -1}" == '"' ]] || \
         [[ "${value:0:1}" == "'" && "${value: -1}" == "'" ]]; }; then
        value="${value:1:-1}"
    fi
    printf '%s\n' "${value}"
}

normalize_container_environment() {
    local variable
    local value
    local -a variables=(
        AGENT_NODE ALLOW_ORIGINS API_PORT BACKEND_HOST BACKEND_PORT CURRENT_RELEASE DEBUG HOST
        CUSTOMER_API_ENABLED ETH HTTP_PORT NGINX_ENABLED OCSERV_BANNER OCSERV_DEBUG
        OCSERV_PRE_LOGIN_BANNER OCSERV_REGENERATE_CONFIG RUN_MIGRATIONS SECRET_KEY
        SUPERADMIN_PASSWORD SUPERADMIN_USERNAME SYSTEMD TELEGRAM_BOT_ENABLED WEB_PORT
        TELEGRAM_RECEIPTS_DIR
    )

    for variable in "${variables[@]}"; do
        [[ -v "${variable}" ]] || continue
        value="$(strip_surrounding_quotes "${!variable}")"
        printf -v "${variable}" '%s' "${value}"
        export "${variable}"
    done
}

: "${OCSERV_PORT:=443}"
: "${OC_NET:=172.16.24.0/24}"
: "${OCSERV_DNS:=1.1.1.1}"
: "${SSL_CN:=ocserv-dashboard}"
: "${SSL_ORG:=ocserv-dashboard}"
: "${SSL_EXPIRE:=3650}"
: "${OCSERV_PRESERVE_CONFIG:=false}"
: "${PGDATA:=/var/lib/postgresql/18/docker}"
: "${POSTGRES_HOST:=127.0.0.1}"
: "${POSTGRES_PORT:=5432}"
: "${POSTGRES_USER:=ocserv}"
: "${POSTGRES_DB:=ocserv_db}"
: "${POSTGRES_SSLMODE:=disable}"
: "${POSTGRES_PASSWORD:=}"

OCSERV_PORT="$(strip_surrounding_quotes "${OCSERV_PORT}")"
OC_NET="$(strip_surrounding_quotes "${OC_NET}")"
OCSERV_DNS="$(strip_surrounding_quotes "${OCSERV_DNS}")"
SSL_CN="$(strip_surrounding_quotes "${SSL_CN}")"
SSL_ORG="$(strip_surrounding_quotes "${SSL_ORG}")"
SSL_EXPIRE="$(strip_surrounding_quotes "${SSL_EXPIRE}")"
OCSERV_PRESERVE_CONFIG="$(strip_surrounding_quotes "${OCSERV_PRESERVE_CONFIG}")"
PGDATA="$(strip_surrounding_quotes "${PGDATA}")"
POSTGRES_HOST="$(strip_surrounding_quotes "${POSTGRES_HOST}")"
POSTGRES_PORT="$(strip_surrounding_quotes "${POSTGRES_PORT}")"
POSTGRES_USER="$(strip_surrounding_quotes "${POSTGRES_USER}")"
POSTGRES_DB="$(strip_surrounding_quotes "${POSTGRES_DB}")"
POSTGRES_SSLMODE="$(strip_surrounding_quotes "${POSTGRES_SSLMODE}")"
POSTGRES_PASSWORD="$(strip_surrounding_quotes "${POSTGRES_PASSWORD}")"

export PGDATA POSTGRES_HOST POSTGRES_PORT POSTGRES_USER POSTGRES_DB POSTGRES_SSLMODE POSTGRES_PASSWORD
normalize_container_environment

readonly OCSERV_CONF=/etc/ocserv/ocserv.conf
readonly OCSERV_SSL_DIR=/etc/ocserv/ssl
readonly OCSERV_CERTS_DIR=/etc/ocserv/certs
readonly OCSERV_CA_CERT="${OCSERV_SSL_DIR}/ca-cert.pem"
readonly OCSERV_CA_KEY="${OCSERV_SSL_DIR}/ca-key.pem"
readonly OCSERV_CRL="${OCSERV_SSL_DIR}/crl.pem"
readonly POSTGRES_ENTRYPOINT=/usr/local/bin/docker-entrypoint.sh

log() {
    printf '[container] %s\n' "$*"
}

die() {
    printf '[container] ERROR: %s\n' "$*" >&2
    exit 1
}

is_true() {
    case "${1,,}" in
        1|true|yes|on) return 0 ;;
        *) return 1 ;;
    esac
}

validate_postgres_runtime() {
    [[ -x "${POSTGRES_ENTRYPOINT}" ]] || \
        die "official PostgreSQL entrypoint is missing: ${POSTGRES_ENTRYPOINT}"
    command -v postgres >/dev/null 2>&1 || die "PostgreSQL server binary is missing"
    command -v pg_isready >/dev/null 2>&1 || die "pg_isready is missing"
    command -v gosu >/dev/null 2>&1 || die "gosu is missing"

    if [[ "${PGDATA}" != /var/lib/postgresql/* ]]; then
        log "WARNING: PGDATA=${PGDATA} is outside /var/lib/postgresql; make sure it is persisted explicitly"
    fi
}

ensure_client_pki() {
    local ca_template="${OCSERV_SSL_DIR}/ca.tmpl"
    local crl_template="${OCSERV_SSL_DIR}/crl.tmpl"

    mkdir -p "${OCSERV_SSL_DIR}/users" "${OCSERV_SSL_DIR}/disabled"
    chmod 700 "${OCSERV_SSL_DIR}" "${OCSERV_SSL_DIR}/users" "${OCSERV_SSL_DIR}/disabled"
    touch "${OCSERV_SSL_DIR}/revoked.pem" "${OCSERV_SSL_DIR}/suspended.pem"
    chmod 600 "${OCSERV_SSL_DIR}/revoked.pem" "${OCSERV_SSL_DIR}/suspended.pem"

    if [[ -f "${OCSERV_CA_CERT}" && -f "${OCSERV_CA_KEY}" ]] && \
       ! certtool --certificate-info --infile "${OCSERV_CA_CERT}" >/dev/null 2>&1; then
        log "client CA certificate is invalid; regenerating the client PKI"
        rm -f "${OCSERV_CA_CERT}" "${OCSERV_CA_KEY}" "${OCSERV_CRL}"
    elif [[ -f "${OCSERV_CA_CERT}" && ! -f "${OCSERV_CA_KEY}" ]] || \
         [[ ! -f "${OCSERV_CA_CERT}" && -f "${OCSERV_CA_KEY}" ]]; then
        log "client CA is incomplete; regenerating the client PKI"
        rm -f "${OCSERV_CA_CERT}" "${OCSERV_CA_KEY}" "${OCSERV_CRL}"
    fi

    if [[ ! -f "${OCSERV_CA_CERT}" ]]; then
        log "generating Ocserv client certificate authority"
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

    log "generating Ocserv server certificate"
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

    log "writing Ocserv configuration"
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
    local ip_forward

    ip_forward="$(cat /proc/sys/net/ipv4/ip_forward 2>/dev/null || true)"
    if [[ "${ip_forward}" != 1 ]]; then
        if ! sysctl -w net.ipv4.ip_forward=1 >/dev/null; then
            log "cannot enable IP forwarding; create the container with --sysctl net.ipv4.ip_forward=1"
            exit 1
        fi
    else
        log "IPv4 forwarding is enabled"
    fi

    external_interface="${ETH:-}"
    if [[ -z "${external_interface}" ]]; then
        external_interface="$(ip route | awk '/default/ {print $5; exit}')"
    fi
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

setup_ocserv() {
    ensure_ocserv_layout
    ensure_networking
}

main() {
    (( $# > 0 )) || die "no container command supplied"

    # Keep this wrapper running as root for Ocserv/network setup. PostgreSQL is
    # started later by container-server through the official Postgres entrypoint,
    # which drops only the PostgreSQL child process to the postgres user.
    validate_postgres_runtime
    setup_ocserv
    exec "$@"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
