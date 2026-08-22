#!/usr/bin/env bash

set -Eeuo pipefail

: "${RUN_MIGRATIONS:=true}"
: "${MIGRATION_MAX_ATTEMPTS:=30}"
: "${MIGRATION_RETRY_SECONDS:=2}"
: "${DEBUG:=0}"
: "${OCSERV_DEBUG:=0}"

backend_pid=''
ocserv_pid=''
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

stop_services() {
    if [[ "${stopping}" == true ]]; then
        return
    fi
    stopping=true

    log "stopping backend and OCServ"
    [[ -n "${backend_pid}" ]] && kill -TERM "${backend_pid}" 2>/dev/null || true
    [[ -n "${ocserv_pid}" ]] && kill -TERM "${ocserv_pid}" 2>/dev/null || true
    [[ -n "${backend_pid}" ]] && wait "${backend_pid}" 2>/dev/null || true
    [[ -n "${ocserv_pid}" ]] && wait "${ocserv_pid}" 2>/dev/null || true
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

    run_migrations
    trap handle_signal SIGINT SIGTERM

    if is_true "${DEBUG}"; then
        backend_args+=(--debug)
    fi

    log "starting backend"
    /usr/local/bin/backend "${backend_args[@]}" &
    backend_pid=$!

    log "starting OCServ"
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
    wait -n "${backend_pid}" "${ocserv_pid}"
    exit_code=$?
    set -e

    log "a critical service exited with status ${exit_code}"
    stop_services
    return "${exit_code}"
}

main "$@"
