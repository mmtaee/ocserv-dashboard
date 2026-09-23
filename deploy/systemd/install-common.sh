#!/usr/bin/env bash

set -Eeuo pipefail

# shellcheck disable=SC2155
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2155
readonly PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd)"
readonly SOURCE_ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env}"
readonly BACKEND_UNIT=/etc/systemd/system/ocserv-dashboard.service
readonly OCSERV_UNIT=/etc/systemd/system/ocserv.service
readonly NGINX_SITE=/etc/nginx/sites-available/ocserv-dashboard
readonly UI_DIR=/var/www/ocserv-dashboard

temporary_binary=''
INSTALL_DIR=''
INSTALLED_ENV_FILE=''

: "${DEPLOYMENT_AGENT_NODE:?DEPLOYMENT_AGENT_NODE must be set by a master or agent installer}"

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
    [[ "${EUID}" -eq 0 ]] || die "run the selected installer as root through ./install.sh"
    [[ -d /run/systemd/system ]] || die "systemd is not running on this host"
    command -v apt-get >/dev/null 2>&1 || die "only Debian/Ubuntu apt-based hosts are supported"
}

load_environment() {
    if [[ ! -f "${SOURCE_ENV_FILE}" ]]; then
        if [[ -f "${PROJECT_ROOT}/.env.example" ]]; then
            cp "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env"
            chmod 600 "${PROJECT_ROOT}/.env"
            die "created ${PROJECT_ROOT}/.env from .env.example; set secure values and rerun"
        fi
        die "environment file not found: ${SOURCE_ENV_FILE}"
    fi

    set -a
    # shellcheck source=/dev/null
    source "${SOURCE_ENV_FILE}"
    set +a
    AGENT_NODE="${DEPLOYMENT_AGENT_NODE}"
    export AGENT_NODE

    : "${POSTGRES_HOST:?POSTGRES_HOST is required}"
    : "${POSTGRES_PORT:?POSTGRES_PORT is required}"
    : "${POSTGRES_DB:?POSTGRES_DB is required}"
    : "${POSTGRES_USER:?POSTGRES_USER is required}"
	: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
	: "${SECRET_KEY:?SECRET_KEY is required}"
	: "${SUPERADMIN_USERNAME:?SUPERADMIN_USERNAME is required}"
    : "${SUPERADMIN_PASSWORD:?SUPERADMIN_PASSWORD is required}"

	[[ "${SECRET_KEY}" != "replace-with-a-random-secret" ]] || die "replace the sample SECRET_KEY before installation"
    [[ "${POSTGRES_PASSWORD}" != "replace-with-a-strong-database-password" ]] || die "replace the sample POSTGRES_PASSWORD before installation"
    [[ "${SUPERADMIN_PASSWORD}" != "replace-with-a-strong-superadmin-password" ]] || die "replace the sample SUPERADMIN_PASSWORD before installation"

    [[ "${POSTGRES_PORT}" =~ ^[0-9]+$ ]] || die "POSTGRES_PORT must be numeric"
    [[ "${POSTGRES_DB}" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]] || die "POSTGRES_DB must be a PostgreSQL identifier"
    [[ "${POSTGRES_USER}" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]] || die "POSTGRES_USER must be a PostgreSQL identifier"

    BACKEND_HOST="${BACKEND_HOST:-0.0.0.0}"
    BACKEND_PORT="${BACKEND_PORT:-8080}"
    WEB_PORT="${WEB_PORT:-3000}"
    API_PORT="${API_PORT:-8000}"
    CUSTOMER_API_ENABLED="${CUSTOMER_API_ENABLED:-true}"
    INSTALL_POSTGRES="${INSTALL_POSTGRES:-true}"
    GO_VERSION="${GO_VERSION:-1.27.1}"
    INSTALL_DIR="${SYSTEMD_INSTALL_DIR:-/opt/ocserv-dashboard}"
    # Cobra maintenance commands load this file with godotenv when run from INSTALL_DIR.
    INSTALLED_ENV_FILE="${INSTALL_DIR}/.env"

    [[ "${BACKEND_HOST}" =~ ^[a-zA-Z0-9.:%_-]+$ ]] || die "BACKEND_HOST contains unsupported characters"
    [[ "${BACKEND_PORT}" =~ ^[0-9]+$ ]] || die "BACKEND_PORT must be numeric"
    [[ "${HOST:-}" =~ ^[A-Za-z0-9][A-Za-z0-9.-]*$ ]] || die "HOST must be an IPv4 address or hostname; current value: ${HOST:-<empty>}"
    [[ "${WEB_PORT}" =~ ^[0-9]+$ ]] || die "WEB_PORT must be numeric"
    [[ "${API_PORT}" =~ ^[0-9]+$ ]] || die "API_PORT must be numeric"
    [[ "${GO_VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "GO_VERSION must be a semantic version"
    [[ "${INSTALL_DIR}" =~ ^/[a-zA-Z0-9._/-]+$ ]] || die "SYSTEMD_INSTALL_DIR must be an absolute path without spaces"
}

install_packages() {
    local packages=(
        ca-certificates
        build-essential
        curl
        git
        gperf
        gnutls-bin
        iproute2
        iptables
        iptables-persistent
        libcjose-dev
        libcurl4-gnutls-dev
        libev-dev
        libgnutls28-dev
        libjansson-dev
        libkrb5-dev
        libllhttp-dev
        liblz4-dev
        libnl-route-3-dev
        liboath-dev
        libpam0g-dev
        libprotobuf-c-dev
        libreadline-dev
        libseccomp-dev
        libtasn1-bin
        libtalloc-dev
        meson
        ninja-build
        openssl
        pkg-config
        postgresql-client
        protobuf-c-compiler
        procps
    )

    if is_true "${INSTALL_POSTGRES}"; then
        packages+=(postgresql)
    fi
    if ! is_true "${DEPLOYMENT_AGENT_NODE}"; then
        packages+=(nginx nodejs npm)
    fi

    log "installing PostgreSQL, backend, UI, nginx, and Ocserv build dependencies"
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y --no-install-recommends "${packages[@]}"

    install_go
    if is_true "${DEPLOYMENT_AGENT_NODE}"; then
        return
    fi
    command -v corepack >/dev/null 2>&1 || die "Node.js installation did not provide corepack"
    corepack enable
}

install_go() {
    local archive
    local architecture
    local installed_version
    local install_path="/usr/local/lib/go-${GO_VERSION}"
    local version_output
    local work_dir

    if command -v go >/dev/null 2>&1; then
        version_output="$(go version 2>/dev/null || true)"
        if [[ "${version_output}" =~ go([0-9]+\.[0-9]+\.[0-9]+) ]]; then
            installed_version="${BASH_REMATCH[1]}"
            if dpkg --compare-versions "${installed_version}" ge "${GO_VERSION}"; then
                log "using installed Go ${installed_version}"
                return
            fi
        fi
    fi

    case "$(dpkg --print-architecture)" in
        amd64) architecture=amd64 ;;
        arm64) architecture=arm64 ;;
        *) die "Go ${GO_VERSION} is unsupported on this CPU architecture" ;;
    esac
    if [[ ! -x "${install_path}/bin/go" ]]; then
        archive="$(mktemp)"
        work_dir="$(mktemp -d)"
        log "installing Go ${GO_VERSION}"
        curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${architecture}.tar.gz" -o "${archive}" || \
            die "could not download Go ${GO_VERSION}; install Go ${GO_VERSION} or newer and rerun"
        install -d -m 755 /usr/local/lib
        tar -C "${work_dir}" -xzf "${archive}"
        mv "${work_dir}/go" "${install_path}"
        rm -f "${archive}"
        rmdir "${work_dir}"
    fi
    ln -sfn "${install_path}/bin/go" /usr/local/bin/go
    go version | grep -Fq "go${GO_VERSION}" || die "Go ${GO_VERSION} installation failed"
}

install_ocserv_binary() {
    log "building and installing the current Ocserv binary"
    bash "${PROJECT_ROOT}/deploy/common/ocserv_binary_install.sh"
    command -v /usr/sbin/ocserv >/dev/null 2>&1 || die "Ocserv binary installation failed"
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

build_ui() {
    if is_true "${DEPLOYMENT_AGENT_NODE}"; then
        return
    fi

    log "building admin UI"
    corepack yarn --cwd "${PROJECT_ROOT}/web/admin" install --immutable
    VITE_API_BASE_URL=/api \
        VITE_API_TIMEOUT_MS=15000 \
        VITE_I18N_LANGUAGES="en:English,it:Italiano,zh-cn:中文(简体),zh-tw:中文(繁體),ru:Русский,fa:فارسی,ar:العربية" \
        corepack yarn --cwd "${PROJECT_ROOT}/web/admin" build
    install -d -m 755 "${UI_DIR}"
    rm -rf -- "${UI_DIR:?}"/*
    cp -a "${PROJECT_ROOT}/web/admin/dist/." "${UI_DIR}/"

    if ! is_true "${CUSTOMER_API_ENABLED}"; then
        return
    fi

    log "building customer UI"
    corepack yarn --cwd "${PROJECT_ROOT}/web/customer" install --immutable
    VITE_API_BASE_URL=/api \
        VITE_API_TIMEOUT_MS=15000 \
        VITE_USE_MOCKS=false \
        VITE_I18N_LANGUAGES="en:English,it:Italiano,zh-cn:中文(简体),zh-tw:中文(繁體),ru:Русский,fa:فارسی,ar:العربية" \
        corepack yarn --cwd "${PROJECT_ROOT}/web/customer" build --base=/customer/
    install -d -m 755 "${UI_DIR}/customer"
    cp -a "${PROJECT_ROOT}/web/customer/dist/." "${UI_DIR}/customer/"
}

setup_ocserv_host() {
    log "configuring Ocserv certificates, authentication, VPN network, and firewall"
    # The shared setup keeps Docker and systemd Ocserv layouts identical.
    # shellcheck source=/dev/null
    (
        source "${PROJECT_ROOT}/deploy/docker/entrypoint.sh"
        setup_ocserv
    )

    printf 'net.ipv4.ip_forward = 1\n' >/etc/sysctl.d/99-ocserv-dashboard.conf
    sysctl --system >/dev/null
    netfilter-persistent save >/dev/null
}

configure_nginx() {
    if is_true "${DEPLOYMENT_AGENT_NODE}"; then
        systemctl disable --now nginx.service 2>/dev/null || true
        return
    fi

    log "configuring nginx for the admin UI and API"
    rm -f /etc/nginx/sites-enabled/default
    {
        cat <<EOF
server {
    listen ${WEB_PORT} ssl default_server;
    listen [::]:${WEB_PORT} ssl default_server;
    server_name ${HOST};
    root ${UI_DIR};

    ssl_certificate /etc/ocserv/certs/cert.pem;
    ssl_certificate_key /etc/ocserv/certs/cert.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    error_page 497 =301 https://\$host:${WEB_PORT}\$request_uri;

    location /api/ {
        proxy_pass http://127.0.0.1:${BACKEND_PORT};
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location = /health {
        proxy_pass http://127.0.0.1:${BACKEND_PORT}/health;
    }

    location = /swagger { return 404; }
    location ^~ /swagger/ { return 404; }
EOF
        if is_true "${CUSTOMER_API_ENABLED}"; then
            cat <<'EOF'

    location = /customer { return 301 /customer/; }
    location /customer/ { try_files $uri $uri/ /customer/index.html; }
EOF
        else
            cat <<'EOF'

    location ^~ /customer { return 404; }
EOF
        fi
        cat <<EOF

    location / { try_files \$uri \$uri/ /index.html; }
}

server {
    listen ${API_PORT} ssl;
    listen [::]:${API_PORT} ssl;
    server_name ${HOST};

    ssl_certificate /etc/ocserv/certs/cert.pem;
    ssl_certificate_key /etc/ocserv/certs/cert.key;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        proxy_pass http://127.0.0.1:${BACKEND_PORT};
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
    } >"${NGINX_SITE}"
    ln -sf "${NGINX_SITE}" /etc/nginx/sites-enabled/ocserv-dashboard
    nginx -t
    systemctl enable --now nginx.service
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

    log "migrating legacy empty per-user Ocserv configs"
    user_rows="$(
        PGPASSWORD="${POSTGRES_PASSWORD}" PGSSLMODE="${POSTGRES_SSLMODE:-disable}" psql \
            -h "${POSTGRES_HOST}" \
            -p "${POSTGRES_PORT}" \
            -U "${POSTGRES_USER}" \
            -d "${POSTGRES_DB}" \
            -AtF $'\t' \
            -c "SELECT username, \"group\" FROM ocserv_users WHERE \"group\" IS NOT NULL AND \"group\" <> '' AND \"group\" <> 'defaults' AND \"group\" <> '*';"
    )" || die "failed to read Ocserv users for filesystem migration"

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
    rm -f -- /etc/systemd/system/ocserv.service.d/override.conf
    cat >"${OCSERV_UNIT}" <<'EOF'
[Unit]
Description=OpenConnect VPN server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/sbin/ocserv --foreground --config=/etc/ocserv/ocserv.conf
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    cat >"${BACKEND_UNIT}" <<EOF
[Unit]
Description=Ocserv Dashboard unified backend
After=network-online.target ocserv.service
Wants=network-online.target
Requires=ocserv.service

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
EnvironmentFile=${INSTALLED_ENV_FILE}
Environment=SYSTEMD=true
Environment=AGENT_NODE=${DEPLOYMENT_AGENT_NODE}
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
        die "Ocserv failed to start"
    }
    systemctl is-active --quiet ocserv-dashboard.service || {
        journalctl -u ocserv-dashboard.service -n 50 --no-pager >&2 || true
        die "backend failed to start"
    }
    if ! is_true "${DEPLOYMENT_AGENT_NODE}"; then
        systemctl is-active --quiet nginx.service || {
            journalctl -u nginx.service -n 50 --no-pager >&2 || true
            die "nginx failed to start"
        }
    fi

    log "installation complete"
    log "backend: http://${BACKEND_HOST}:${BACKEND_PORT}"
    if ! is_true "${DEPLOYMENT_AGENT_NODE}"; then
        log "UI: https://${HOST}:${WEB_PORT}"
    fi
    log "status: systemctl status ocserv ocserv-dashboard"
}

main() {
    trap cleanup EXIT
    require_root
    load_environment
    install_packages
    install_ocserv_binary
    setup_local_postgres
    build_backend
    build_ui
    setup_ocserv_host
    run_backend_migrations
    migrate_user_config_links
    write_systemd_units
    configure_nginx
    verify_services
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
