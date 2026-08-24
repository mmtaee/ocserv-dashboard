#!/usr/bin/env bash

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

backend_pid=''
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

    log "stopping backend, OCServ, and PostgreSQL"
    [[ -n "${backend_pid}" ]] && kill -TERM "${backend_pid}" 2>/dev/null || true
    [[ -n "${ocserv_pid}" ]] && kill -TERM "${ocserv_pid}" 2>/dev/null || true
    [[ -n "${postgres_pid}" ]] && kill -INT "${postgres_pid}" 2>/dev/null || true
    [[ -n "${backend_pid}" ]] && wait "${backend_pid}" 2>/dev/null || true
    [[ -n "${ocserv_pid}" ]] && wait "${ocserv_pid}" 2>/dev/null || true
    [[ -n "${postgres_pid}" ]] && wait "${postgres_pid}" 2>/dev/null || true
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

    trap handle_signal SIGINT SIGTERM

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

    log "starting OCServ with debug level ${OCSERV_DEBUG}"
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
    wait -n "${backend_pid}" "${ocserv_pid}" "${postgres_pid}"
    exit_code=$?
    set -e

    log "a critical service exited with status ${exit_code}"
    stop_services
    return "${exit_code}"
}

main "$@"
