#!/usr/bin/env bash

# Shared all-in-one process supervisor for master and agent images.

set -Eeuo pipefail

: "${RUN_MIGRATIONS:=true}"
: "${MIGRATION_MAX_ATTEMPTS:=30}"
: "${MIGRATION_RETRY_SECONDS:=2}"
: "${POSTGRES_PORT:=5432}"
: "${POSTGRES_USER:=ocserv}"
: "${POSTGRES_DB:=ocserv_db}"
: "${POSTGRES_READY_MAX_ATTEMPTS:=60}"
: "${POSTGRES_READY_RETRY_SECONDS:=1}"
: "${DEBUG:=0}"
: "${OCSERV_DEBUG:=999}"
: "${AGENT_NODE:=false}"
: "${CUSTOMER_API_ENABLED:=true}"
: "${TELEGRAM_BOT_ENABLED:=false}"
: "${NGINX_ENABLED:=true}"
: "${WEB_PORT:=3000}"
: "${API_PORT:=8000}"
: "${HOST:=127.0.0.1}"

backend_pid=''
nginx_pid=''
ocserv_pid=''
postgres_pid=''
stopping=false

log() {
    printf '[server] %s\n' "$*"
}

is_true() {
    case "${1,,}" in
        1|true|yes|on) return 0 ;;
        *) return 1 ;;
    esac
}

load_postgres_password() {
    if [[ -n "${POSTGRES_PASSWORD:-}" && -n "${POSTGRES_PASSWORD_FILE:-}" ]]; then
        log "POSTGRES_PASSWORD and POSTGRES_PASSWORD_FILE are mutually exclusive"
        return 1
    fi
    if [[ -z "${POSTGRES_PASSWORD:-}" && -n "${POSTGRES_PASSWORD_FILE:-}" ]]; then
        [[ -r "${POSTGRES_PASSWORD_FILE}" ]] || {
            log "cannot read POSTGRES_PASSWORD_FILE=${POSTGRES_PASSWORD_FILE}"
            return 1
        }
        POSTGRES_PASSWORD="$(<"${POSTGRES_PASSWORD_FILE}")"
        export POSTGRES_PASSWORD
        unset POSTGRES_PASSWORD_FILE
    fi
    if [[ -z "${POSTGRES_PASSWORD:-}" ]]; then
        log "POSTGRES_PASSWORD is required for the bundled PostgreSQL server"
        return 1
    fi
    if [[ "${POSTGRES_PASSWORD}" == "replace-with-a-strong-database-password" ]]; then
        log "replace the sample POSTGRES_PASSWORD before starting the container"
        return 1
    fi
}

run_migrations() {
    local attempt=1

    if ! is_true "${RUN_MIGRATIONS}"; then
        log "database migrations disabled"
        return
    fi

    while (( attempt <= MIGRATION_MAX_ATTEMPTS )); do
        log "running database migrations (attempt ${attempt}/${MIGRATION_MAX_ATTEMPTS})"
        if /usr/local/bin/backend migrate; then
            log "database migrations completed"
            return
        fi

        if (( attempt == MIGRATION_MAX_ATTEMPTS )); then
            log "database migrations failed after ${MIGRATION_MAX_ATTEMPTS} attempts"
            return 1
        fi

        sleep "${MIGRATION_RETRY_SECONDS}"
        ((attempt += 1))
    done
}

ensure_superadmin() {
    if [[ -z "${SUPERADMIN_USERNAME:-}" || -z "${SUPERADMIN_PASSWORD:-}" ]]; then
        log "SUPERADMIN_USERNAME and SUPERADMIN_PASSWORD are required"
        return 1
    fi
    if [[ "${SUPERADMIN_PASSWORD}" == "replace-with-a-strong-superadmin-password" ]]; then
        log "replace the sample SUPERADMIN_PASSWORD before starting the container"
        return 1
    fi
    log "ensuring initial superadmin ${SUPERADMIN_USERNAME}"
    /usr/local/bin/backend create-superadmin
}

start_postgres() {
    local attempt=1

    # This image is intentionally all-in-one; backend always uses its local PostgreSQL.
    export POSTGRES_HOST=127.0.0.1
    export PGPORT="${POSTGRES_PORT}"

    log "starting PostgreSQL 18"
    /usr/local/bin/docker-entrypoint.sh postgres -p "${POSTGRES_PORT}" &
    postgres_pid=$!

    while (( attempt <= POSTGRES_READY_MAX_ATTEMPTS )); do
        if pg_isready \
            --host="${POSTGRES_HOST}" \
            --port="${POSTGRES_PORT}" \
            --username="${POSTGRES_USER}" \
            --dbname="${POSTGRES_DB}" >/dev/null 2>&1; then
            log "PostgreSQL is ready"
            return
        fi

        if ! kill -0 "${postgres_pid}" 2>/dev/null; then
            set +e
            wait "${postgres_pid}"
            local exit_code=$?
            set -e
            log "PostgreSQL exited during startup with status ${exit_code}"
            if (( exit_code == 0 )); then
                return 1
            fi
            return "${exit_code}"
        fi

        sleep "${POSTGRES_READY_RETRY_SECONDS}"
        ((attempt += 1))
    done

    log "PostgreSQL did not become ready after ${POSTGRES_READY_MAX_ATTEMPTS} attempts"
    return 1
}

stop_services() {
    if [[ "${stopping}" == true ]]; then
        return
    fi
    stopping=true

    log "stopping backend, nginx, Ocserv, and PostgreSQL"
    [[ -n "${backend_pid}" ]] && kill -TERM "${backend_pid}" 2>/dev/null || true
    [[ -n "${nginx_pid}" ]] && kill -QUIT "${nginx_pid}" 2>/dev/null || true
    [[ -n "${ocserv_pid}" ]] && kill -TERM "${ocserv_pid}" 2>/dev/null || true
    [[ -n "${postgres_pid}" ]] && kill -INT "${postgres_pid}" 2>/dev/null || true
    [[ -n "${backend_pid}" ]] && wait "${backend_pid}" 2>/dev/null || true
    [[ -n "${nginx_pid}" ]] && wait "${nginx_pid}" 2>/dev/null || true
    [[ -n "${ocserv_pid}" ]] && wait "${ocserv_pid}" 2>/dev/null || true
    [[ -n "${postgres_pid}" ]] && wait "${postgres_pid}" 2>/dev/null || true
}

validate_image_mode() {
    local image_agent=false
    local runtime_agent=false

    is_true "${AGENT_NODE}" && runtime_agent=true
    is_true "${IMAGE_AGENT_NODE:-${AGENT_NODE}}" && image_agent=true
    if [[ "${runtime_agent}" != "${image_agent}" ]]; then
        log "AGENT_NODE=${AGENT_NODE} does not match image build mode ${IMAGE_AGENT_NODE}"
        return 1
    fi
    if ! is_true "${AGENT_NODE}" && is_true "${CUSTOMER_API_ENABLED}" && \
       ! is_true "${IMAGE_CUSTOMER_API_ENABLED:-true}"; then
        log "customer API is enabled but this image was built without the customer UI"
        return 1
    fi
}

configure_nginx() {
    rm -f /etc/nginx/sites-enabled/default
    if is_true "${AGENT_NODE}" || ! is_true "${NGINX_ENABLED}"; then
        log "nginx and UI routes disabled"
        return
    fi
    [[ "${WEB_PORT}" =~ ^[0-9]+$ ]] || { log "WEB_PORT must be numeric"; return 1; }
    [[ "${API_PORT}" =~ ^[0-9]+$ ]] || { log "API_PORT must be numeric"; return 1; }

    {
        cat <<EOF
server {
    listen ${WEB_PORT} ssl default_server;
    listen [::]:${WEB_PORT} ssl default_server;
    server_name ${HOST};
    root /usr/share/nginx/html;

    ssl_certificate /etc/ocserv/certs/cert.pem;
    ssl_certificate_key /etc/ocserv/certs/cert.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    error_page 497 =301 https://\$host:${WEB_PORT}\$request_uri;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location = /health {
        proxy_pass http://127.0.0.1:8080/health;
    }

    location = /swagger {
        return 404;
    }

    location ^~ /swagger/ {
        return 404;
    }
EOF
        if is_true "${CUSTOMER_API_ENABLED}"; then
            cat <<'EOF'

    location = /customer {
        return 301 /customer/;
    }

    location /customer/ {
        try_files $uri $uri/ /customer/index.html;
    }
EOF
        else
            cat <<'EOF'

    location ^~ /customer {
        return 404;
    }
EOF
        fi
        cat <<'EOF'

    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF
        cat <<EOF

server {
    listen ${API_PORT} ssl;
    listen [::]:${API_PORT} ssl;
    server_name ${HOST};

    ssl_certificate /etc/ocserv/certs/cert.pem;
    ssl_certificate_key /etc/ocserv/certs/cert.key;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
    } >/etc/nginx/sites-enabled/ocserv-dashboard.conf

    nginx -t
}

start_nginx() {
    if is_true "${AGENT_NODE}" || ! is_true "${NGINX_ENABLED}"; then
        return
    fi
    log "starting nginx for admin UI, API, and customer UI"
    nginx -g 'daemon off;' &
    nginx_pid=$!
}

handle_signal() {
    trap - SIGINT SIGTERM
    log "received shutdown signal"
    stop_services
    exit 0
}

main() {
    local backend_args=(serve --docker-mode)
    local exit_code
    local -a service_pids

    trap handle_signal SIGINT SIGTERM

    validate_image_mode
    configure_nginx
    load_postgres_password
    if ! start_postgres; then
        stop_services
        return 1
    fi
    if ! run_migrations; then
        stop_services
        return 1
    fi
    if ! ensure_superadmin; then
        stop_services
        return 1
    fi

    if is_true "${DEBUG}"; then
        backend_args+=(--debug)
    fi

    log "starting backend"
    /usr/local/bin/backend "${backend_args[@]}" &
    backend_pid=$!
    start_nginx

    log "starting Ocserv with debug level ${OCSERV_DEBUG}"
    if [[ "${OCSERV_DEBUG}" == 0 ]]; then
        /usr/sbin/ocserv \
            --foreground \
            --config=/etc/ocserv/ocserv.conf &
    else
        /usr/sbin/ocserv \
            --foreground \
            --debug="${OCSERV_DEBUG}" \
            --config=/etc/ocserv/ocserv.conf &
    fi
    ocserv_pid=$!

    set +e
    service_pids=("${backend_pid}" "${ocserv_pid}" "${postgres_pid}")
    if [[ -n "${nginx_pid}" ]]; then
        service_pids+=("${nginx_pid}")
    fi
    wait -n "${service_pids[@]}"
    exit_code=$?
    set -e

    log "a critical service exited with status ${exit_code}"
    stop_services
    return "${exit_code}"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
