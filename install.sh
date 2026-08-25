#!/usr/bin/env bash

set -Eeuo pipefail

readonly PROJECT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env}"

deployment=''
node_mode=''
dry_run=false
assume_yes=false
declare -a docker_cmd

log() {
    printf '[install] %s\n' "$*"
}

die() {
    printf '[install] ERROR: %s\n' "$*" >&2
    exit 1
}

usage() {
    cat <<'EOF'
Usage: ./install.sh [--deployment docker|systemd] [--node master|agent] [--yes] [--dry-run]

Without flags, the installer asks for the deployment type first and node mode second.
EOF
}

parse_arguments() {
    while (($# > 0)); do
        case "$1" in
            --deployment)
                (($# >= 2)) || die "--deployment requires a value"
                deployment="$2"
                shift 2
                ;;
            --node)
                (($# >= 2)) || die "--node requires a value"
                node_mode="$2"
                shift 2
                ;;
            --yes)
                assume_yes=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *) die "unknown argument: $1" ;;
        esac
    done
}

choose_deployment() {
    if [[ -z "${deployment}" ]]; then
        printf 'Choose deployment:\n  1. Docker\n  2. Systemd\n'
        read -r -p 'Selection: ' selection
        case "${selection}" in
            1) deployment=docker ;;
            2) deployment=systemd ;;
            *) die "invalid deployment selection" ;;
        esac
    fi
    case "${deployment}" in
        docker|systemd) ;;
        *) die "deployment must be docker or systemd" ;;
    esac
}

choose_node_mode() {
    if [[ -z "${node_mode}" ]]; then
        printf 'Choose node mode:\n  1. Master\n  2. Agent\n'
        read -r -p 'Selection: ' selection
        case "${selection}" in
            1) node_mode=master ;;
            2) node_mode=agent ;;
            *) die "invalid node selection" ;;
        esac
    fi
    case "${node_mode}" in
        master) agent_node=false ;;
        agent) agent_node=true ;;
        *) die "node must be master or agent" ;;
    esac
    readonly agent_node
}

confirm() {
    local prompt="$1"
    if [[ "${assume_yes}" == true ]]; then
        return 0
    fi
    read -r -p "${prompt} [y/N] " answer
    [[ "${answer,,}" == y || "${answer,,}" == yes ]]
}

set_env_value() {
    local key="$1"
    local value="$2"
    local temporary
    temporary="$(mktemp "${ENV_FILE}.XXXXXX")"
    awk -v key="${key}" -v value="${value}" '
        BEGIN { replaced = 0 }
        $0 ~ "^[[:space:]]*" key "=" {
            if (!replaced) print key "=" value
            replaced = 1
            next
        }
        { print }
        END { if (!replaced) print key "=" value }
    ' "${ENV_FILE}" >"${temporary}"
    chmod 600 "${temporary}"
    mv "${temporary}" "${ENV_FILE}"
}

prepare_environment() {
    local current_mode
    if [[ ! -f "${ENV_FILE}" ]]; then
        [[ -f "${PROJECT_ROOT}/.env.sample" ]] || die ".env.sample is missing"
        command -v openssl >/dev/null 2>&1 || die "openssl is required to generate initial secrets"
        install -m 600 "${PROJECT_ROOT}/.env.sample" "${ENV_FILE}"
        set_env_value SECRET_KEY "\"$(openssl rand -hex 32)\""
        set_env_value POSTGRES_PASSWORD "\"$(openssl rand -hex 32)\""
        set_env_value SUPERADMIN_PASSWORD "\"$(openssl rand -base64 24 | tr -d '\n')\""
        set_env_value AGENT_NODE "${agent_node}"
        log "created ${ENV_FILE} with generated secrets; review it before exposing the service"
        return
    fi

    current_mode="$(awk -F= '/^[[:space:]]*AGENT_NODE=/{gsub(/[[:space:]\"\047]/, "", $2); print tolower($2); exit}' "${ENV_FILE}")"
    if [[ -n "${current_mode}" && "${current_mode}" != "${agent_node}" ]]; then
        confirm "${ENV_FILE} has AGENT_NODE=${current_mode}; change it to ${agent_node}?" || \
            die "existing environment was not changed"
    fi
    set_env_value AGENT_NODE "${agent_node}"
    chmod 600 "${ENV_FILE}"
}

check_docker() {
    command -v docker >/dev/null 2>&1 || die "Docker is required"
    if docker info >/dev/null 2>&1; then
        docker_cmd=(docker)
    elif command -v sudo >/dev/null 2>&1 && sudo docker info >/dev/null 2>&1; then
        docker_cmd=(sudo docker)
    else
        die "cannot connect to the Docker daemon"
    fi
    [[ -c /dev/net/tun ]] || die "/dev/net/tun is unavailable; load the tun kernel module"
}

install_docker() {
    local image="ocserv-dashboard:${node_mode}"
    local container="${CONTAINER_NAME:-ocserv}"
    local data_root="${DEPLOY_DATA_ROOT:-/opt/ocserv_dashboard/docker_volumes}"
    local dockerfile="${PROJECT_ROOT}/deploy/docker/Dockerfile.${node_mode}"

    check_docker
    [[ -f "${dockerfile}" ]] || die "Dockerfile not found: ${dockerfile}"
    if "${docker_cmd[@]}" container inspect "${container}" >/dev/null 2>&1; then
        die "container ${container} already exists; remove it explicitly before reinstalling"
    fi
    if ! mkdir -p "${data_root}/postgresql18" "${data_root}/ocserv" \
        "${data_root}/cron_journal" "${data_root}/telegram_receipts"; then
        command -v sudo >/dev/null 2>&1 || die "cannot create deployment volumes under ${data_root}"
        sudo mkdir -p "${data_root}/postgresql18" "${data_root}/ocserv" \
            "${data_root}/cron_journal" "${data_root}/telegram_receipts"
    fi

    "${docker_cmd[@]}" build --file "${dockerfile}" --tag "${image}" "${PROJECT_ROOT}"
    "${docker_cmd[@]}" run --detach \
        --name "${container}" \
        --restart unless-stopped \
        --env-file "${ENV_FILE}" \
        --env "AGENT_NODE=${agent_node}" \
        --env HOST_PROC=/host/proc \
        --env HOST_SYS=/host/sys \
        --cap-add NET_ADMIN \
        --sysctl net.ipv4.ip_forward=1 \
        --device /dev/net/tun:/dev/net/tun \
        --volume /var/run/docker.sock:/var/run/docker.sock:ro \
        --volume /proc:/host/proc:ro \
        --volume /sys:/host/sys:ro \
        --volume "${data_root}/ocserv:/etc/ocserv" \
        --volume "${data_root}/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts" \
        --volume "${data_root}/postgresql18:/var/lib/postgresql" \
        --volume "${data_root}/cron_journal:/app/cron_journal" \
        --publish 443:443/tcp \
        --publish 443:443/udp \
        --publish 8080:8080/tcp \
        "${image}"
    log "started Docker ${node_mode} node in container ${container}"
}

install_systemd() {
    local installer="${PROJECT_ROOT}/deploy/systemd/${node_mode}/install.sh"
    [[ -x "${installer}" ]] || die "systemd installer not found or not executable: ${installer}"
    if [[ "${EUID}" -eq 0 ]]; then
        ENV_FILE="${ENV_FILE}" "${installer}"
        return
    fi
    command -v sudo >/dev/null 2>&1 || die "sudo is required for systemd installation"
    sudo env "ENV_FILE=${ENV_FILE}" "${installer}"
}

main() {
    parse_arguments "$@"
    choose_deployment
    choose_node_mode

    log "deployment=${deployment} node=${node_mode} AGENT_NODE=${agent_node}"
    if [[ "${dry_run}" == true ]]; then
        if [[ "${deployment}" == docker ]]; then
            log "dry run: would use deploy/docker/Dockerfile.${node_mode}"
        else
            log "dry run: would use deploy/systemd/${node_mode}/install.sh"
        fi
        return
    fi

    prepare_environment
    case "${deployment}" in
        docker) install_docker ;;
        systemd) install_systemd ;;
    esac
}

main "$@"
