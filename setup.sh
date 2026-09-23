#!/usr/bin/env bash

set -Eeuo pipefail

readonly PROJECT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env}"

deployment=''
node_mode=''
dry_run=false
assume_yes=false
update=false
graphical=false
environment_created=false
detected_host_ip=''
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
Usage: ./setup.sh [--deployment docker|systemd] [--node master|agent] [--update] [--yes] [--dry-run]

Without flags, choose Install or Upgrade. Install then asks for the deployment type and node mode.
--update detects the running deployment and updates it without prompts.
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
            --update)
                update=true
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

choose_action() {
    local selection

    if [[ "${graphical}" == true ]]; then
        selection="$(whiptail --title 'Ocserv Dashboard' --nocancel --menu 'Choose action' 12 60 3 \
            1 'Install' 2 'Upgrade' 3 'Exit' 3>&1 1>&2 2>&3)" || exit 0
        case "${selection}" in
            1) return ;;
            2) update=true; return ;;
            3) exit 0 ;;
        esac
    fi
    printf 'Choose action:\n  1. Install\n  2. Upgrade\n'
    read -r -p 'Selection: ' selection
    case "${selection}" in
        1) ;;
        2) update=true ;;
        *) die "invalid action selection" ;;
    esac
}

choose_install_options() {
    while true; do
        choose_deployment || return 1
        choose_node_mode && return
        deployment=''
    done
}

update_source() {
    local current_release=''
    local latest_release

    command -v git >/dev/null 2>&1 || die "git is required to update"
    latest_release="$(git ls-remote --tags --refs --sort=-version:refname origin 'v*' | awk -F/ 'NR == 1 { print $3 }')"
    [[ "${latest_release}" =~ ^v[0-9]+(\.[0-9]+)*$ ]] || die "could not determine the latest release tag"
    if [[ -f "${PROJECT_ROOT}/.release" ]]; then
        current_release="$(<"${PROJECT_ROOT}/.release")"
    fi
    if [[ "${current_release}" == "${latest_release}" ]]; then
        log "already on release ${latest_release}"
        return
    fi
    git diff --quiet && git diff --cached --quiet || die "source changes must be committed or discarded before updating"
    log "updating ${current_release:-current source} to ${latest_release}"
    git fetch --tags --force origin
    git checkout --detach "${latest_release}"
    printf '%s\n' "${latest_release}" >"${PROJECT_ROOT}/.release"
    exec bash "${PROJECT_ROOT}/setup.sh" --update
}

detect_update_deployment() {
    local detected=''

    if command -v docker >/dev/null 2>&1; then
        if docker info >/dev/null 2>&1; then
            docker_cmd=(docker)
        elif command -v sudo >/dev/null 2>&1 && sudo -n docker info >/dev/null 2>&1; then
            docker_cmd=(sudo docker)
        fi
        if ((${#docker_cmd[@]})) && "${docker_cmd[@]}" container inspect "${CONTAINER_NAME:-ocserv}" >/dev/null 2>&1; then
            detected=docker
        fi
    fi
    if command -v systemctl >/dev/null 2>&1 && \
        { systemctl is-active --quiet ocserv-dashboard.service || systemctl is-enabled --quiet ocserv-dashboard.service || \
            systemctl is-active --quiet ocserv.service || systemctl is-enabled --quiet ocserv.service; }; then
        [[ -z "${detected}" ]] || die "both Docker and systemd Ocserv services are running"
        detected=systemd
    fi
    [[ -n "${detected}" ]] || die "no running Ocserv Docker container or systemd service found"
    deployment="${detected}"
    agent_node="$(normalized_bool "$(env_value AGENT_NODE false)")"
    if [[ "${agent_node}" == true ]]; then
        node_mode=agent
    else
        node_mode=master
    fi
    readonly agent_node
    assume_yes=true
    log "detected ${deployment} ${node_mode} deployment"
}

choose_deployment() {
    if [[ -z "${deployment}" ]]; then
        if [[ "${graphical}" == true ]]; then
            deployment="$(whiptail --title 'Ocserv Dashboard' --cancel-button Back --menu 'Choose deployment' 12 60 2 \
                docker 'Docker' systemd 'Systemd' 3>&1 1>&2 2>&3)" || return 1
        else
        printf 'Choose deployment:\n  1. Docker\n  2. Systemd\n'
        read -r -p 'Selection: ' selection
        case "${selection}" in
            1) deployment=docker ;;
            2) deployment=systemd ;;
            *) die "invalid deployment selection" ;;
        esac
        fi
    fi
    case "${deployment}" in
        docker|systemd) ;;
        *) die "deployment must be docker or systemd" ;;
    esac
}

choose_node_mode() {
    if [[ -z "${node_mode}" ]]; then
        if [[ "${graphical}" == true ]]; then
            if whiptail --title 'Ocserv Dashboard' --yes-button Yes --no-button No --cancel-button Back \
                --yesno 'Is this a master node?' 10 60; then
                node_mode=master
            elif [[ "$?" -eq 1 ]]; then
                node_mode=agent
            else
                return 1
            fi
        else
        printf 'Choose node mode:\n  1. Master\n  2. Agent\n'
        read -r -p 'Selection: ' selection
        case "${selection}" in
            1) node_mode=master ;;
            2) node_mode=agent ;;
            *) die "invalid node selection" ;;
        esac
        fi
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

env_value() {
    local key="$1"
    local fallback="$2"
    local value

    value="$(awk -F= -v key="${key}" '$1 == key { sub(/^[^=]*=/, ""); gsub(/^[[:space:]"\047]+|[[:space:]"\047]+$/, ""); print; exit }' "${ENV_FILE}")"
    printf '%s\n' "${value:-${fallback}}"
}

normalized_bool() {
    case "${1,,}" in
        1|true|yes|on) printf 'true\n' ;;
        *) printf 'false\n' ;;
    esac
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

sanitize_environment_file() {
    local temporary

    temporary="$(mktemp "${ENV_FILE}.XXXXXX")"
    LC_ALL=C awk '/^[[:space:]]*($|#)/ || /^[A-Za-z_][A-Za-z0-9_]*=/' "${ENV_FILE}" >"${temporary}"
    chmod 600 "${temporary}"
    mv "${temporary}" "${ENV_FILE}"
}

set_env_default() {
    local key="$1"
    local value="$2"

    awk -F= -v key="${key}" '$1 == key { found = 1 } END { exit !found }' "${ENV_FILE}" || \
        set_env_value "${key}" "${value}"
}

apply_static_defaults() {
    local release='unknown'

    [[ -f "${PROJECT_ROOT}/.release" ]] && release="$(<"${PROJECT_ROOT}/.release")"
    set_env_default ALLOW_ORIGINS "\"https://$(env_value HOST ''):3000\""
    set_env_default CURRENT_RELEASE "\"${release}\""
    if [[ -z "$(env_value SECRET_KEY '')" ]]; then
        command -v openssl >/dev/null 2>&1 || die "openssl is required to generate the application secret"
        set_env_value SECRET_KEY "\"$(openssl rand -hex 32)\""
    fi
    set_env_default SUPERADMIN_USERNAME "\"admin\""
    if [[ -z "$(env_value SUPERADMIN_PASSWORD '')" ]]; then
        command -v openssl >/dev/null 2>&1 || die "openssl is required to generate the administrator password"
        set_env_value SUPERADMIN_PASSWORD "\"$(openssl rand -base64 24 | tr -d '\\n')\""
    fi
    set_env_default OCSERV_PORT 443
    set_env_default OC_NET "\"172.16.24.0/24\""
    set_env_default OCSERV_DNS "\"1.1.1.1\""
    set_env_default OCSERV_BANNER "\"Welcome to the VPN service\""
    set_env_default OCSERV_PRE_LOGIN_BANNER "\"Welcome\""
    set_env_default SSL_CN "\"$(env_value HOST '')\""
    set_env_default SSL_ORG "\"ocserv-dashboard\""
    set_env_default SSL_EXPIRE 3650
    set_env_default ETH "\"\""
    set_env_default TELEGRAM_BOT_ENABLED false
    set_env_default CUSTOMER_API_ENABLED true
    set_env_default BACKEND_HOST "\"0.0.0.0\""
    set_env_default BACKEND_PORT 8080
    set_env_default WEB_PORT 3000
    set_env_default API_PORT 8000
    set_env_default SYSTEMD false
    set_env_default POSTGRES_HOST "\"127.0.0.1\""
    set_env_default POSTGRES_PORT 5432
    set_env_default POSTGRES_DB "\"ocserv_db\""
    set_env_default POSTGRES_USER "\"ocserv\""
    if [[ -z "$(env_value POSTGRES_PASSWORD '')" ]]; then
        command -v openssl >/dev/null 2>&1 || die "openssl is required to generate the database password"
        set_env_value POSTGRES_PASSWORD "\"$(openssl rand -hex 32)\""
    fi
    set_env_default POSTGRES_SSLMODE "\"disable\""
    set_env_default PGDATA "\"/var/lib/postgresql/18/docker\""
    set_env_default INSTALL_POSTGRES true
    set_env_default SYSTEMD_INSTALL_DIR "\"/opt/ocserv-dashboard\""
    set_env_default OCSERV_PRESERVE_CONFIG false
    set_env_default OCSERV_REGENERATE_CONFIG false
    set_env_default TELEGRAM_RECEIPTS_DIR "\"\""
    set_env_default TELEGRAM_I18N_PATH "\"\""
    set_env_default TELEGRAM_BOT_I18N_PATH "\"\""
    set_env_default TELEGRAM_BOT_METADATA_LOCALES_PATH "\"\""
    set_env_default GO_VERSION "\"1.27.1\""
    set_env_default NODE_VERSION "\"24\""
    set_env_default RUN_MIGRATIONS true
    set_env_default MIGRATION_MAX_ATTEMPTS 30
    set_env_default MIGRATION_RETRY_SECONDS 2
    set_env_default POSTGRES_READY_MAX_ATTEMPTS 60
    set_env_default POSTGRES_READY_RETRY_SECONDS 1
    set_env_default DEBUG 0
    set_env_default OCSERV_DEBUG 999
}

detect_host_ip() {
    local host_ip

    if [[ "${graphical}" == true ]]; then
        whiptail --infobox 'DETECTING SERVER IPV4 ADDRESS...' 8 60 || true
    fi
    host_ip="$(ip -4 route get 1.1.1.1 2>/dev/null | awk '/src/ { for (i = 1; i <= NF; i++) if ($i == "src") { print $(i + 1); exit } }')"
    if [[ -z "${host_ip}" ]]; then
        host_ip="$(hostname -I 2>/dev/null | awk '{ print $1 }')"
    fi
    if [[ ! "${host_ip}" =~ ^[0-9]{1,3}(\.[0-9]{1,3}){3}$ ]]; then
        if [[ "${graphical}" == false ]]; then
            die "could not determine a host IPv4 address; set HOST in ${ENV_FILE}"
        fi
        whiptail --title 'SERVER ADDRESS' --msgbox \
            'AUTOMATIC IPV4 DETECTION FAILED. YOU WILL BE ASKED TO ENTER THE SERVER ADDRESS.' \
            10 75 || true
        while true; do
            host_ip="$(whiptail --title 'SERVER ADDRESS' --nocancel --inputbox \
                'COULD NOT DETECT AN IPV4 ADDRESS. ENTER YOUR SERVER IPV4 ADDRESS OR DOMAIN NAME.' \
                12 75 3>&1 1>&2 2>&3)"
            [[ "${host_ip}" =~ ^[A-Za-z0-9][A-Za-z0-9.-]*$ ]] && break
            whiptail --msgbox 'ENTER A PLAIN IPV4 ADDRESS OR DOMAIN NAME WITHOUT HTTP:// OR HTTPS://.' 8 75 || true
        done
    fi
    if [[ "${graphical}" == true ]]; then
        whiptail --msgbox "DETECTED SERVER IPV4 ADDRESS: ${host_ip}" 8 60 || true
    fi
    detected_host_ip="${host_ip}"
}

configure_host() {
    local host
    local value
    local key

    host="$(env_value HOST '')"
    if [[ -z "${host}" ]]; then
        detect_host_ip
        host="${detected_host_ip}"
        set_env_value HOST "${host}"
        log "set HOST=${host}"
    fi
    while [[ ! "${host}" =~ ^[A-Za-z0-9][A-Za-z0-9.-]*$ ]]; do
        [[ "${graphical}" == true ]] || die "HOST must be an IPv4 address or hostname"
        host="$(whiptail --title 'SERVER ADDRESS' --nocancel --inputbox \
            'HOST IS INVALID. ENTER A PLAIN IPV4 ADDRESS OR DOMAIN NAME.' 10 75 3>&1 1>&2 2>&3)"
        set_env_value HOST "${host}"
    done

    for key in ALLOW_ORIGINS SSL_CN; do
        value="$(env_value "${key}" '')"
        if [[ "${value}" == *'{HOST}'* ]]; then
            value="${value//\{HOST\}/${host}}"
            set_env_value "${key}" "\"${value}\""
        fi
    done

    if [[ "$(env_value ALLOW_ORIGINS '')" == "https://127.0.0.1:3000,https://localhost:3000" ]]; then
        set_env_value ALLOW_ORIGINS "\"https://${host}:3000\""
    fi
    if [[ "$(env_value SSL_CN '')" == "ocserv-dashboard" ]]; then
        set_env_value SSL_CN "\"${host}\""
    fi
}

install_apt_package() {
    local package="$1"

    command -v apt-get >/dev/null 2>&1 || die "${package} is required; install it, then rerun the installer"
    command -v sudo >/dev/null 2>&1 || [[ "${EUID}" -eq 0 ]] || die "sudo is required to install ${package}"
    log "installing ${package}"
    if [[ "${EUID}" -eq 0 ]]; then
        apt-get update
        DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends "${package}"
    else
        sudo apt-get update
        sudo env DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends "${package}"
    fi
}

ensure_whiptail() {
    command -v whiptail >/dev/null 2>&1 || install_apt_package whiptail
}

ensure_nano() {
    command -v nano >/dev/null 2>&1 || install_apt_package nano
}

ensure_openssl() {
    command -v openssl >/dev/null 2>&1 || install_apt_package openssl
}

add_nano_help() {
    local temporary
    temporary="$(mktemp "${ENV_FILE}.XXXXXX")"
    {
        printf '# Nano help: save with Ctrl+O, then Enter; exit with Ctrl+X.\n'
        printf '# To exit without saving: Ctrl+X, then N.\n\n'
        cat "${ENV_FILE}"
    } >"${temporary}"
    chmod 600 "${temporary}"
    mv "${temporary}" "${ENV_FILE}"
}

edit_environment_value() {
    local key="$1"
    local label="$2"
    local value

    value="$(env_value "${key}" '')"
    if [[ "${key}" == SUPERADMIN_PASSWORD || "${key}" == SECRET_KEY ]]; then
        value="$(whiptail --title 'Edit .env' --passwordbox "${label}" 10 70 3>&1 1>&2 2>&3)" || return 0
        [[ -n "${value}" ]] || return
    else
        value="$(whiptail --title 'Edit .env' --inputbox "${label}" 10 70 "${value}" 3>&1 1>&2 2>&3)" || return 0
    fi
    [[ "${value}" != *'"'* && ! "${value}" =~ [[:cntrl:]] ]] || {
        whiptail --msgbox 'CONTROL CHARACTERS, DOUBLE QUOTES, AND NEW LINES ARE NOT SUPPORTED HERE.' 8 70 || true
        return
    }
    case "${key}" in
        HOST) [[ "${value}" =~ ^[A-Za-z0-9][A-Za-z0-9.-]*$ ]] || {
            whiptail --msgbox 'HOST MUST BE AN IPV4 ADDRESS OR DOMAIN NAME WITHOUT HTTP:// OR HTTPS://.' 8 70 || true
            return
        } ;;
        WEB_PORT|OCSERV_PORT) [[ "${value}" =~ ^[0-9]+$ ]] || {
            whiptail --msgbox 'Port must be numeric.' 8 60 || true
            return
        } ;;
    esac
    set_env_value "${key}" "\"${value}\""
    if [[ "${key}" == HOST ]]; then
        set_env_value ALLOW_ORIGINS "\"https://${value}:3000\""
        set_env_value SSL_CN "\"${value}\""
    fi
}

edit_environment_boolean() {
    local key="$1"
    local prompt="$2"

    if whiptail --title 'Edit .env' --yes-button Yes --no-button No --defaultno --yesno "${prompt}" 10 70; then
        set_env_value "${key}" true
    elif [[ "$?" -eq 1 ]]; then
        set_env_value "${key}" false
    fi
}

select_network_interface() {
    local interface
    local selection
    local -a choices=(none 'No interface')

    whiptail --infobox 'DETECTING NETWORK INTERFACES...' 8 60 || true
    command -v ip >/dev/null 2>&1 || {
        whiptail --msgbox 'The ip command is unavailable; ETH was not changed.' 8 60 || true
        return
    }
    while IFS= read -r interface; do
        interface="${interface#*: }"
        interface="${interface%%:*}"
        interface="${interface%@*}"
        [[ "${interface}" == lo ]] || choices+=("${interface}" "${interface}")
    done < <(ip -o link show)
    if ((${#choices[@]} == 2)); then
        whiptail --msgbox 'No non-loopback network interfaces were found.' 8 60 || true
        return
    fi
    selection="$(whiptail --title 'Network interface' --menu 'Choose the interface for Ocserv' 18 70 10 \
        "${choices[@]}" 3>&1 1>&2 2>&3)" || return 0
    if [[ "${selection}" == none ]]; then
        set_env_value ETH '""'
    else
        set_env_value ETH "\"${selection}\""
    fi
}

edit_environment_graphically() {
    local host
    local selection

    host="$(env_value HOST '')"
    if ! whiptail --title 'Server address' --yes-button Use --no-button Change \
        --yesno "Use HOST=${host}?" 10 70; then
        edit_environment_value HOST 'Enter an IPv4 address or domain name'
    fi
    whiptail --title 'Edit .env' --yesno 'Edit the main environment settings before installation?' 10 70 || return 0
    while true; do
        selection="$(whiptail --title 'Edit .env' --notags --menu 'Choose a setting to edit' 24 80 14 \
            HOST 'HOST' \
            SECRET_KEY 'SECRET KEY' \
            SUPERADMIN_USERNAME 'SUPERADMIN USERNAME' \
            SUPERADMIN_PASSWORD 'SUPERADMIN PASSWORD' \
            OCSERV_PORT 'OCSERV PORT' \
            OC_NET 'OC NET' \
            OCSERV_DNS 'OCSERV DNS' \
            OCSERV_BANNER 'OCSERV BANNER' \
            OCSERV_PRE_LOGIN_BANNER 'OCSERV PRE LOGIN BANNER' \
            SSL_CN 'SSL CN' \
            SSL_ORG 'SSL ORG' \
            SSL_EXPIRE 'SSL EXPIRE' \
            ETH 'ETH' \
            telegram 'TELEGRAM BOT ENABLED' \
            customer 'CUSTOMER API ENABLED' \
            done 'CONTINUE INSTALLATION' 3>&1 1>&2 2>&3)" || return 0
        [[ "${selection}" == done ]] && return
        case "${selection}" in
            telegram) edit_environment_boolean TELEGRAM_BOT_ENABLED 'TELEGRAM BOT ENABLED?' ;;
            customer)
                if [[ "${agent_node}" == true ]]; then
                    whiptail --msgbox 'Customer API is disabled on agent nodes.' 8 60 || true
                else
                    edit_environment_boolean CUSTOMER_API_ENABLED 'CUSTOMER API ENABLED?'
                fi
                ;;
            ETH) select_network_interface ;;
            *) edit_environment_value "${selection}" "ENTER ${selection//_/ }" ;;
        esac
    done
}

prepare_environment() {
    local current_mode
    local postgres_password

    ensure_openssl
    if [[ ! -f "${ENV_FILE}" ]]; then
        [[ -f "${PROJECT_ROOT}/.env.example" ]] || die ".env.example is missing"
        install -m 600 "${PROJECT_ROOT}/.env.example" "${ENV_FILE}"
        set_env_value SECRET_KEY "\"$(openssl rand -hex 32)\""
        set_env_value POSTGRES_PASSWORD "\"$(openssl rand -hex 32)\""
        set_env_value SUPERADMIN_PASSWORD "\"$(openssl rand -base64 24 | tr -d '\n')\""
        set_env_value AGENT_NODE "${agent_node}"
        [[ "${agent_node}" == true ]] && set_env_value CUSTOMER_API_ENABLED false
        configure_host
        apply_static_defaults
        log "created ${ENV_FILE} with generated secrets; review it before exposing the service"
        environment_created=true
        return
    fi

    sanitize_environment_file
    current_mode="$(awk -F= '/^[[:space:]]*AGENT_NODE=/{gsub(/[[:space:]\"\047]/, "", $2); print tolower($2); exit}' "${ENV_FILE}")"
    if [[ -n "${current_mode}" && "${current_mode}" != "${agent_node}" ]]; then
        confirm "${ENV_FILE} has AGENT_NODE=${current_mode}; change it to ${agent_node}?" || \
            die "existing environment was not changed"
    fi
    set_env_value AGENT_NODE "${agent_node}"
    postgres_password="$(env_value POSTGRES_PASSWORD '')"
    if [[ "${postgres_password}" == "replace-with-a-strong-database-password" ]]; then
        command -v openssl >/dev/null 2>&1 || die "openssl is required to replace the sample PostgreSQL password"
        set_env_value POSTGRES_PASSWORD "\"$(openssl rand -hex 32)\""
        log "replaced the sample POSTGRES_PASSWORD with a generated secret"
    fi
    configure_host
    apply_static_defaults
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

create_docker_volumes() {
    local data_root="$1"
    local existing_parent="${data_root}"
    local -a volume_dirs=(
        "${data_root}/postgresql18"
        "${data_root}/ocserv"
        "${data_root}/cron_journal"
        "${data_root}/telegram_receipts"
    )

    while [[ ! -e "${existing_parent}" ]]; do
        existing_parent="$(dirname -- "${existing_parent}")"
    done
    [[ -d "${existing_parent}" ]] || die "deployment volume parent is not a directory: ${existing_parent}"

    # The default data root is below /opt, which is normally writable only by
    # root. Select sudo before attempting mkdir so a successful install does
    # not first print spurious permission-denied errors. For a custom path,
    # inspect its nearest existing parent so a new user-owned path needs no sudo.
    if [[ "${EUID}" -eq 0 || ( -w "${existing_parent}" && -x "${existing_parent}" ) ]]; then
        mkdir -p "${volume_dirs[@]}"
        return
    fi

    command -v sudo >/dev/null 2>&1 || die "cannot create deployment volumes under ${data_root}; sudo is required"
    sudo mkdir -p "${volume_dirs[@]}"
}

install_docker() {
    local image="ocserv-dashboard:${node_mode}"
    local container="${CONTAINER_NAME:-ocserv}"
    local data_root="${DEPLOY_DATA_ROOT:-/opt/ocserv_dashboard/docker_volumes}"
    local dockerfile="${PROJECT_ROOT}/deploy/docker/Dockerfile"
    local customer_api_enabled
    local telegram_bot_enabled
    local web_port
    local api_port
    local ocserv_port
    local host
    local -a service_publish_args

    customer_api_enabled="$(normalized_bool "$(env_value CUSTOMER_API_ENABLED true)")"
    telegram_bot_enabled="$(normalized_bool "$(env_value TELEGRAM_BOT_ENABLED false)")"
    web_port="$(env_value WEB_PORT 3000)"
    api_port="$(env_value API_PORT 8000)"
    ocserv_port="$(env_value OCSERV_PORT 443)"
    host="$(env_value HOST '')"
    [[ "${web_port}" =~ ^[0-9]+$ ]] || die "WEB_PORT must be numeric"
    [[ "${api_port}" =~ ^[0-9]+$ ]] || die "API_PORT must be numeric"
    [[ "${ocserv_port}" =~ ^[0-9]+$ ]] || die "OCSERV_PORT must be numeric"
    if [[ "${agent_node}" == true ]]; then
        service_publish_args=(--publish 8080:8080/tcp)
    else
        service_publish_args=(
            --publish "${web_port}:${web_port}/tcp"
            --publish "${api_port}:${api_port}/tcp"
        )
    fi

    check_docker
    [[ -f "${dockerfile}" ]] || die "Dockerfile not found: ${dockerfile}"
    if "${docker_cmd[@]}" container inspect "${container}" >/dev/null 2>&1; then
        log "removing existing container ${container} before reinstalling"
        "${docker_cmd[@]}" container rm --force "${container}"
    fi
    create_docker_volumes "${data_root}"

    "${docker_cmd[@]}" build \
        --build-arg "GO_VERSION=$(env_value GO_VERSION 1.27.1)" \
        --build-arg "NODE_VERSION=$(env_value NODE_VERSION 24)" \
        --build-arg "AGENT_NODE=${agent_node}" \
        --build-arg "CUSTOMER_API_ENABLED=${customer_api_enabled}" \
        --build-arg "TELEGRAM_BOT_ENABLED=${telegram_bot_enabled}" \
        --file "${dockerfile}" \
        --tag "${image}" \
        "${PROJECT_ROOT}"
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
        --publish "${ocserv_port}:${ocserv_port}/tcp" \
        --publish "${ocserv_port}:${ocserv_port}/udp" \
        "${service_publish_args[@]}" \
        "${image}"
    log "started Docker ${node_mode} node in container ${container}"
    log "Ocserv: ${host}:${ocserv_port} (TCP/UDP)"
    if [[ "${agent_node}" == false ]]; then
        log "UI: https://${host}:${web_port}"
    fi

    # log "following container logs; press Ctrl-C to stop following"
    log "following container logs, run ${docker_cmd[*]} logs -f ${container}"
    # "${docker_cmd[@]}" logs -f "${container}"
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
    local argument_count=$#

    parse_arguments "$@"
    if ((argument_count == 0)); then
        graphical=true
        ensure_whiptail
        while true; do
            choose_action
            if [[ "${update}" == true ]] || choose_install_options; then
                break
            fi
        done
    fi
    if [[ "${update}" == true ]]; then
        update_source
        detect_update_deployment
    else
        if [[ "${graphical}" == false ]]; then
            choose_deployment
            choose_node_mode
        fi
    fi

    log "deployment=${deployment} node=${node_mode} AGENT_NODE=${agent_node}"
    if [[ "${dry_run}" == true ]]; then
        if [[ "${deployment}" == docker ]]; then
            log "dry run: would use deploy/docker/Dockerfile with AGENT_NODE=${agent_node}"
        else
            log "dry run: would use deploy/systemd/${node_mode}/install.sh"
        fi
        return
    fi

    prepare_environment
    if [[ "${graphical}" == true && "${update}" == false ]]; then
        edit_environment_graphically
        configure_host
    elif [[ "${environment_created}" == true ]]; then
        ensure_nano
        add_nano_help
        log "opening ${ENV_FILE} in nano in 2 seconds; save and exit to continue"
        sleep 2
        nano "${ENV_FILE}"
    fi
    case "${deployment}" in
        docker)
            set_env_value SYSTEMD false
            install_docker
            ;;
        systemd)
            set_env_value SYSTEMD true
            install_systemd
            ;;
    esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    main "$@"
fi
