#!/usr/bin/env bash

set -Eeuo pipefail

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
readonly DOCKERFILE="${PROJECT_ROOT}/Dockerfile-Dev"
readonly IMAGE_NAME="${DEV_IMAGE_NAME:-ocserv-dashboard-dev:latest}"
readonly CONTAINER_NAME=ocserv
readonly DATA_ROOT="${DEV_DATA_ROOT:-${PROJECT_ROOT}/.volume}"
readonly POSTGRES_DATA_DIR="${DATA_ROOT}/postgresql18"
readonly GO_VERSION="${GO_VERSION:-1.26.0}"
readonly VPN_PORT="${DEV_VPN_PORT:-443}"
readonly API_PORT="${DEV_API_PORT:-8080}"
readonly POSTGRES_PORT="${DEV_POSTGRES_PORT:-5435}"
readonly OCSERV_DEBUG_LEVEL="${OCSERV_DEBUG:-999}"
readonly DEVELOPMENT_LABEL="io.ocserv-dashboard.environment=development"

declare -a docker_cmd

log() {
    printf '\033[1;34m[dev]\033[0m %s\n' "$*"
}

success() {
    printf '\033[1;32m[dev]\033[0m %s\n' "$*"
}

warn() {
    printf '\033[1;33m[dev]\033[0m %s\n' "$*" >&2
}

die() {
    printf '\033[1;31m[dev] ERROR:\033[0m %s\n' "$*" >&2
    exit 1
}

show_failure_logs() {
    local exit_code=$?

    trap - ERR
    warn "command failed at line ${BASH_LINENO[0]} with status ${exit_code}"
    if [[ ${#docker_cmd[@]} -gt 0 ]] && "${docker_cmd[@]}" container inspect "${CONTAINER_NAME}" >/dev/null 2>&1; then
        warn "last ${DEV_ERROR_LOG_LINES:-100} container log lines:"
        "${docker_cmd[@]}" logs --tail "${DEV_ERROR_LOG_LINES:-100}" "${CONTAINER_NAME}" >&2 || true
    fi
    exit "${exit_code}"
}

trap show_failure_logs ERR

configure_commands() {
    command -v docker >/dev/null 2>&1 || die "Docker is not installed"

    if docker info >/dev/null 2>&1; then
        docker_cmd=(docker)
    elif command -v sudo >/dev/null 2>&1 && sudo docker info >/dev/null 2>&1; then
        docker_cmd=(sudo docker)
    else
        die "cannot connect to the Docker daemon"
    fi

    log "Docker command: ${docker_cmd[*]}"
}

validate_host() {
    [[ -f "${DOCKERFILE}" ]] || die "development Dockerfile not found: ${DOCKERFILE}"
    [[ -f "${PROJECT_ROOT}/docker/entrypoint.sh" ]] || die "docker/entrypoint.sh not found"
    [[ -f "${PROJECT_ROOT}/docker/server.sh" ]] || die "docker/server.sh not found"
    [[ -S /var/run/docker.sock ]] || die "/var/run/docker.sock is unavailable"
    [[ -c /dev/net/tun ]] || die "/dev/net/tun is unavailable; load the tun module before running this script"

    [[ "${VPN_PORT}" =~ ^[0-9]+$ ]] || die "DEV_VPN_PORT must be numeric"
    [[ "${API_PORT}" =~ ^[0-9]+$ ]] || die "DEV_API_PORT must be numeric"
    [[ "${POSTGRES_PORT}" =~ ^[0-9]+$ ]] || die "DEV_POSTGRES_PORT must be numeric"

    log "creating persistent development directories under ${DATA_ROOT}"
    if ! mkdir -p \
        "${POSTGRES_DATA_DIR}" \
        "${DATA_ROOT}/ocserv" \
        "${DATA_ROOT}/cron_journal" \
        "${DATA_ROOT}/telegram_receipts"; then
        command -v sudo >/dev/null 2>&1 || die "cannot create persistent directories under ${DATA_ROOT}"
        warn "creating ${DATA_ROOT} with sudo because it is not writable by the current user"
        sudo mkdir -p \
            "${POSTGRES_DATA_DIR}" \
            "${DATA_ROOT}/ocserv" \
            "${DATA_ROOT}/cron_journal" \
            "${DATA_ROOT}/telegram_receipts"
    fi
}

remove_previous_container() {
    local environment_label
    local existing_image

    if ! "${docker_cmd[@]}" container inspect "${CONTAINER_NAME}" >/dev/null 2>&1; then
        log "container name ${CONTAINER_NAME} is available"
        return
    fi

    environment_label="$(
        "${docker_cmd[@]}" container inspect \
            --format '{{index .Config.Labels "io.ocserv-dashboard.environment"}}' \
            "${CONTAINER_NAME}" 2>/dev/null || true
    )"
    existing_image="$(
        "${docker_cmd[@]}" container inspect \
            --format '{{.Config.Image}}' \
            "${CONTAINER_NAME}" 2>/dev/null || true
    )"

    if [[ "${environment_label}" != development && "${existing_image}" != "${IMAGE_NAME}" && "${REPLACE_CONTAINER:-false}" != true ]]; then
        die "container ${CONTAINER_NAME} is not marked as development; set REPLACE_CONTAINER=true to replace it explicitly"
    fi

    warn "removing existing container ${CONTAINER_NAME}; persistent data is preserved"
    "${docker_cmd[@]}" rm --force "${CONTAINER_NAME}" >/dev/null
}

build_image() {
    local cache_args=()

    if [[ "${NO_CACHE:-false}" == true ]]; then
        cache_args+=(--no-cache)
    fi

    log "building ${IMAGE_NAME} from ${DOCKERFILE}"
    log "BuildKit plain progress is enabled for detailed build logs"
    "${docker_cmd[@]}" build \
        --progress=plain \
        "${cache_args[@]}" \
        --build-arg "GO_VERSION=${GO_VERSION}" \
        --file "${DOCKERFILE}" \
        --tag "${IMAGE_NAME}" \
        "${PROJECT_ROOT}"
    success "image built: ${IMAGE_NAME}"
}

start_container() {
    log "starting ${CONTAINER_NAME}"
    "${docker_cmd[@]}" run --detach \
        --name "${CONTAINER_NAME}" \
        --label "${DEVELOPMENT_LABEL}" \
        --restart unless-stopped \
        --env HOST_PROC=/host/proc \
        --env HOST_SYS=/host/sys \
        --env POSTGRES_HOST=127.0.0.1 \
        --env TELEGRAM_RECEIPTS_DIR=/opt/ocserv_dashboard/uploads/receipts \
        --env "OCSERV_DEBUG=${OCSERV_DEBUG_LEVEL}" \
        --cap-add NET_ADMIN \
        --sysctl net.ipv4.ip_forward=1 \
        --device /dev/net/tun:/dev/net/tun \
        --volume /var/run/docker.sock:/var/run/docker.sock:ro \
        --volume /proc:/host/proc:ro \
        --volume /sys:/host/sys:ro \
        --volume "${DATA_ROOT}/ocserv:/etc/ocserv" \
        --volume "${DATA_ROOT}/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts" \
        --volume "${POSTGRES_DATA_DIR}:/var/lib/postgresql" \
        --volume "${DATA_ROOT}/cron_journal:/app/cron_journal" \
        --publish "${VPN_PORT}:443/tcp" \
        --publish "${VPN_PORT}:443/udp" \
        --publish "${API_PORT}:8080" \
        --publish "127.0.0.1:${POSTGRES_PORT}:5432" \
        "${IMAGE_NAME}" >/dev/null

    success "development container started"
    log "API:        http://127.0.0.1:${API_PORT}"
    log "Health:     http://127.0.0.1:${API_PORT}/health"
    log "PostgreSQL: 127.0.0.1:${POSTGRES_PORT}"
    log "VPN:        tcp/udp ${VPN_PORT}"
    log "OCServ log: debug level ${OCSERV_DEBUG_LEVEL}"
}

follow_logs() {
    if [[ "${FOLLOW_LOGS:-true}" != true ]]; then
        log "log streaming disabled; run: ${docker_cmd[*]} logs -f ${CONTAINER_NAME}"
        return
    fi

    log "following container logs; Ctrl-C stops log viewing but leaves the container running"
    if ! "${docker_cmd[@]}" logs --follow --timestamps "${CONTAINER_NAME}"; then
        warn "log streaming stopped; container ${CONTAINER_NAME} remains managed by Docker"
    fi
}

main() {
    configure_commands
    validate_host
    build_image
    remove_previous_container
    start_container
    follow_logs
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
