#!/usr/bin/env bash

set -Eeuo pipefail

# shellcheck disable=SC2155
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2155
readonly PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
readonly SOURCE_ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env}"
readonly BACKEND_UNIT=/etc/systemd/system/ocserv-dashboard.service
readonly OCSERV_OVERRIDE_DIR=/etc/systemd/system/ocserv.service.d

temporary_binary=''
INSTALL_DIR=''
INSTALLED_ENV_FILE=''

log() {
    printf '[systemd-install] %s\n' "$*"
}

die() {
    printf '[systemd-install] ERROR: %s\n' "$*" >&2
    exit 1
}

is_true() {
    case "${1,,}" in
        1|true|yes|on) return 0 ;;
        *) return 1 ;;
    esac
}

cleanup() {
    if [[ -n "${temporary_binary}" && -f "${temporary_binary}" ]]; then
        rm -f -- "${temporary_binary}"
    fi
}

require_root() {
    [[ "${EUID}" -eq 0 ]] || die "run this installer as root: sudo ./scripts/systemd/install.sh"
    [[ -d /run/systemd/system ]] || die "systemd is not running on this host"
    command -v apt-get >/dev/null 2>&1 || die "only Debian/Ubuntu apt-based hosts are supported"
}

load_environment() {
    if [[ ! -f "${SOURCE_ENV_FILE}" ]]; then
        if [[ -f "${PROJECT_ROOT}/.env.sample" ]]; then
            cp "${PROJECT_ROOT}/.env.sample" "${PROJECT_ROOT}/.env"
            chmod 600 "${PROJECT_ROOT}/.env"
            die "created ${PROJECT_ROOT}/.env from .env.sample; set secure values and rerun"
        fi
        die "environment file not found: ${SOURCE_ENV_FILE}"
    fi

    set -a
    # shellcheck source=/dev/null
    source "${SOURCE_ENV_FILE}"
    set +a

    : "${POSTGRES_HOST:?POSTGRES_HOST is required}"
    : "${POSTGRES_PORT:?POSTGRES_PORT is required}"
    : "${POSTGRES_DB:?POSTGRES_DB is required}"
    : "${POSTGRES_USER:?POSTGRES_USER is required}"
    : "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
    : "${SECRET_KEY:?SECRET_KEY is required}"
    : "${JWT_SECRET:?JWT_SECRET is required}"
    : "${SUPERADMIN_USERNAME:?SUPERADMIN_USERNAME is required}"
    : "${SUPERADMIN_PASSWORD:?SUPERADMIN_PASSWORD is required}"

    [[ "${SECRET_KEY}" != "replace-with-a-random-secret" ]] || die "replace the sample SECRET_KEY before installation"
    [[ "${JWT_SECRET}" != "replace-with-a-random-jwt-secret" ]] || die "replace the sample JWT_SECRET before installation"
    [[ "${POSTGRES_PASSWORD}" != "replace-with-a-strong-database-password" ]] || die "replace the sample POSTGRES_PASSWORD before installation"
    [[ "${SUPERADMIN_PASSWORD}" != "replace-with-a-strong-superadmin-password" ]] || die "replace the sample SUPERADMIN_PASSWORD before installation"

    [[ "${POSTGRES_PORT}" =~ ^[0-9]+$ ]] || die "POSTGRES_PORT must be numeric"
    [[ "${POSTGRES_DB}" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]] || die "POSTGRES_DB must be a PostgreSQL identifier"
    [[ "${POSTGRES_USER}" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]] || die "POSTGRES_USER must be a PostgreSQL identifier"

    BACKEND_HOST="${BACKEND_HOST:-0.0.0.0}"
    BACKEND_PORT="${BACKEND_PORT:-8080}"
    INSTALL_POSTGRES="${INSTALL_POSTGRES:-true}"
    INSTALL_DIR="${SYSTEMD_INSTALL_DIR:-/opt/ocserv-dashboard}"
    INSTALLED_ENV_FILE="${INSTALL_DIR}/ocserv-dashboard.env"

    [[ "${BACKEND_HOST}" =~ ^[a-zA-Z0-9.:%_-]+$ ]] || die "BACKEND_HOST contains unsupported characters"
    [[ "${BACKEND_PORT}" =~ ^[0-9]+$ ]] || die "BACKEND_PORT must be numeric"
    [[ "${INSTALL_DIR}" =~ ^/[a-zA-Z0-9._/-]+$ ]] || die "SYSTEMD_INSTALL_DIR must be an absolute path without spaces"
}

install_packages() {
    local packages=(
        ca-certificates
        gnutls-bin
        iproute2
        iptables
        iptables-persistent
        ocserv
        openssl
        postgresql-client
        procps
    )

    if is_true "${INSTALL_POSTGRES}"; then
        packages+=(postgresql)
    fi

    log "installing OCServ and system dependencies"
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y --no-install-recommends "${packages[@]}"

    command -v go >/dev/null 2>&1 || die "Go is required to build backend/go.mod; install the declared Go version and rerun"
    command -v /usr/sbin/ocserv >/dev/null 2>&1 || die "OCServ installation did not provide /usr/sbin/ocserv"
    systemctl stop ocserv.service 2>/dev/null || true
}

setup_local_postgres() {
    local escaped_password

    if ! is_true "${INSTALL_POSTGRES}"; then
        log "local PostgreSQL installation disabled; using ${POSTGRES_HOST}:${POSTGRES_PORT}"
        return
    fi

    case "${POSTGRES_HOST}" in
        localhost|127.0.0.1|::1) ;;
        *) die "INSTALL_POSTGRES=true requires POSTGRES_HOST to be localhost, 127.0.0.1, or ::1" ;;
    esac

    systemctl enable --now postgresql.service
    escaped_password="${POSTGRES_PASSWORD//\'/\'\'}"

    log "creating/updating PostgreSQL role ${POSTGRES_USER}"
    runuser -u postgres -- psql --set ON_ERROR_STOP=1 postgres <<SQL
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = '${POSTGRES_USER}') THEN
        CREATE ROLE "${POSTGRES_USER}" LOGIN PASSWORD '${escaped_password}';
    ELSE
        ALTER ROLE "${POSTGRES_USER}" WITH LOGIN PASSWORD '${escaped_password}';
    END IF;
END
\$\$;
SQL

    if ! runuser -u postgres -- psql -tAc "SELECT 1 FROM pg_database WHERE datname='${POSTGRES_DB}'" postgres | grep -q 1; then
        log "creating PostgreSQL database ${POSTGRES_DB}"
        runuser -u postgres -- createdb --owner="${POSTGRES_USER}" "${POSTGRES_DB}"
    fi
    runuser -u postgres -- psql --set ON_ERROR_STOP=1 \
        -c "ALTER DATABASE \"${POSTGRES_DB}\" OWNER TO \"${POSTGRES_USER}\"" postgres
}

build_backend() {
    temporary_binary="$(mktemp)"
    log "building unified backend"
    (
        cd "${PROJECT_ROOT}/backend"
        CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "${temporary_binary}" ./main.go
    )

    install -d -m 750 "${INSTALL_DIR}" "${INSTALL_DIR}/cron_journal" "${INSTALL_DIR}/uploads/receipts"
    install -m 755 "${temporary_binary}" "${INSTALL_DIR}/backend"
    if [[ "$(readlink -f "${SOURCE_ENV_FILE}")" != "$(readlink -m "${INSTALLED_ENV_FILE}")" ]]; then
        install -m 600 "${SOURCE_ENV_FILE}" "${INSTALLED_ENV_FILE}"
    else
        chmod 600 "${INSTALLED_ENV_FILE}"
    fi
}

setup_ocserv_host() {
    log "configuring OCServ certificates, authentication, VPN network, and firewall"
    # The shared setup keeps Docker and systemd OCServ layouts identical.
    # shellcheck source=/dev/null
    (
        source "${PROJECT_ROOT}/docker/entrypoint.sh"
        setup_ocserv
    )

    printf 'net.ipv4.ip_forward = 1\n' >/etc/sysctl.d/99-ocserv-dashboard.conf
    sysctl --system >/dev/null
    netfilter-persistent save >/dev/null
}

run_backend_migrations() {
    log "running database migrations"
    (
        cd "${INSTALL_DIR}"
        ./backend migrate
    )
}

migrate_user_config_links() {
    local username group group_file user_file
    local user_rows
    local migrated=0
    local kept=0
    local skipped=0

    log "migrating legacy empty per-user OCServ configs"
    user_rows="$(
        PGPASSWORD="${POSTGRES_PASSWORD}" PGSSLMODE="${POSTGRES_SSLMODE:-disable}" psql \
            -h "${POSTGRES_HOST}" \
            -p "${POSTGRES_PORT}" \
            -U "${POSTGRES_USER}" \
            -d "${POSTGRES_DB}" \
            -AtF $'\t' \
            -c "SELECT username, \"group\" FROM ocserv_users WHERE \"group\" IS NOT NULL AND \"group\" <> '' AND \"group\" <> 'defaults' AND \"group\" <> '*';"
    )" || die "failed to read OCServ users for filesystem migration"

    while IFS=$'\t' read -r username group; do
        [[ -n "${username}" && -n "${group}" ]] || continue
        if [[ "${username}" == */* || "${group}" == */* ]]; then
            log "skipping unsafe username/group: ${username} -> ${group}"
            ((skipped += 1))
            continue
        fi

        group_file="/etc/ocserv/groups/${group}"
        user_file="/etc/ocserv/users/${username}"
        if [[ ! -f "${group_file}" ]]; then
            ((skipped += 1))
            continue
        fi
        if [[ -d "${user_file}" && ! -L "${user_file}" ]]; then
            ((kept += 1))
            continue
        fi
        if [[ -e "${user_file}" && ! -L "${user_file}" && -s "${user_file}" ]]; then
            ((kept += 1))
            continue
        fi

        rm -f -- "${user_file}"
        ln -s "${group_file}" "${user_file}"
        ((migrated += 1))
    done <<<"${user_rows}"
    log "user config migration: migrated=${migrated}, kept=${kept}, skipped=${skipped}"
}

write_systemd_units() {
    log "writing systemd service definitions"
    install -d -m 755 "${OCSERV_OVERRIDE_DIR}"
    cat >"${OCSERV_OVERRIDE_DIR}/override.conf" <<'EOF'
[Unit]
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=
ExecStart=/usr/sbin/ocserv --foreground --config=/etc/ocserv/ocserv.conf
Restart=on-failure
RestartSec=5s
EOF

    cat >"${BACKEND_UNIT}" <<EOF
[Unit]
Description=OCServ Dashboard unified backend
After=network-online.target ocserv.service
Wants=network-online.target
Requires=ocserv.service

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
EnvironmentFile=${INSTALLED_ENV_FILE}
Environment=SYSTEMD=true
ExecStartPre=${INSTALL_DIR}/backend migrate
ExecStartPre=${INSTALL_DIR}/backend create-superadmin
ExecStart=${INSTALL_DIR}/backend serve --host ${BACKEND_HOST} --port ${BACKEND_PORT}
Restart=on-failure
RestartSec=5s
KillSignal=SIGTERM
TimeoutStopSec=30s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable ocserv.service ocserv-dashboard.service
    systemctl restart ocserv.service
    systemctl restart ocserv-dashboard.service
}

verify_services() {
    systemctl is-active --quiet ocserv.service || {
        journalctl -u ocserv.service -n 50 --no-pager >&2 || true
        die "OCServ failed to start"
    }
    systemctl is-active --quiet ocserv-dashboard.service || {
        journalctl -u ocserv-dashboard.service -n 50 --no-pager >&2 || true
        die "backend failed to start"
    }

    log "installation complete"
    log "backend: http://${BACKEND_HOST}:${BACKEND_PORT}"
    log "status: systemctl status ocserv ocserv-dashboard"
}

main() {
    trap cleanup EXIT
    require_root
    load_environment
    install_packages
    setup_local_postgres
    build_backend
    setup_ocserv_host
    run_backend_migrations
    migrate_user_config_links
    write_systemd_units
    verify_services
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
