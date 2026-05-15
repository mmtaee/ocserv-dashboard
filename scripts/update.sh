#!/bin/bash
# ==============================================================
# Script: update.sh
# Description:
#   One-shot in-place upgrade for both systemd and Docker installations
#   of the Ocserv dashboard.
#   - For systemd: Backs up PostgreSQL, pulls latest code, rebuilds Go
#     backend services, rebuilds and redeploys the Vite frontend, and
#     restarts the systemd units.
#   - For Docker: Pulls latest code, rebuilds Docker images, and restarts
#     the Docker Compose stack.
#
# Why this script exists:
#   Running only systemd_backend.sh after a git pull leaves the
#   browser-served frontend bundle stale, so newly added settings
#   fields (e.g. Telegram bot_token, support_username) never
#   appear in the panel even though the API exposes them. This
#   script makes "update" a single, reliable command.
#
# Usage:
#   sudo ./scripts/update.sh [docker|systemd]
#   If no mode specified, tries to auto-detect.
#
# Optional (from .env or environment):
#   DB_BACKUP_DIR   — dump directory (default: /var/backups/ocserv-dashboard)
#   SKIP_DB_BACKUP  — set to 1 to skip pg_dump
# ==============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

# lib.sh applies `set -euo pipefail` and provides log/ok/warn/die helpers.
# shellcheck source=/dev/null
source ./scripts/lib.sh

ensure_root() {
    if [[ "$EUID" -ne 0 ]] && ! sudo -n true 2>/dev/null; then
        die "❌ This script needs sudo. Re-run with: sudo $0"
    fi
}

ensure_root

# Load environment
if [[ -f "${ROOT_DIR}/.env" ]]; then
    log "Loading environment from .env"
    set -o allexport
    # shellcheck disable=SC1091
    source "${ROOT_DIR}/.env"
    set +o allexport
else
    warn ".env not found at ${ROOT_DIR}/.env — proceeding with defaults"
fi

# Detect or get deployment mode
detect_deployment_mode() {
    # Check if user specified mode
    if [[ $# -ge 1 ]]; then
        case "$1" in
            docker)
                echo "docker"
                return
                ;;
            systemd)
                echo "systemd"
                return
                ;;
        esac
    fi

    # Auto-detect
    if systemctl list-unit-files | grep -q "^api.service" && systemctl is-active --quiet api; then
        echo "systemd"
        return
    fi

    if sudo docker compose ps -q 2>/dev/null; then
        echo "docker"
        return
    fi

    die "Could not detect deployment mode. Please specify: sudo ./scripts/update.sh [docker|systemd]"
}

DEPLOY_MODE=$(detect_deployment_mode "$@")
log "Detected deployment mode: ${DEPLOY_MODE}"

backup_database_before_update() {
    if [[ "${DEPLOY_MODE}" != "systemd" ]]; then
        return 0
    fi

    if [[ "${SKIP_DB_BACKUP:-}" == "1" ]]; then
        warn "Skipping database backup (SKIP_DB_BACKUP=1)."
        return 0
    fi
    if [[ -z "${POSTGRES_USER:-}" || -z "${POSTGRES_DB:-}" ]]; then
        warn "POSTGRES_USER or POSTGRES_DB not set; skipping database backup."
        return 0
    fi
    if ! command -v pg_dump >/dev/null 2>&1; then
        die "pg_dump not found. Install postgresql-client (e.g. apt install postgresql-client) or set SKIP_DB_BACKUP=1."
    fi

    local host port backup_dir ts outfile
    host="${POSTGRES_HOST:-localhost}"
    port="${POSTGRES_PORT:-5432}"
    backup_dir="${DB_BACKUP_DIR:-/var/backups/ocserv-dashboard}"
    ts="$(date +%Y%m%d-%H%M%S)"
    outfile="${backup_dir}/${POSTGRES_DB}-${ts}.dump"

    log "Backing up PostgreSQL (${POSTGRES_DB} @ ${host}:${port}) → ${outfile}"
    mkdir -p "${backup_dir}"
    PGPASSWORD="${POSTGRES_PASSWORD:-}" pg_dump \
        -h "${host}" \
        -p "${port}" \
        -U "${POSTGRES_USER}" \
        -d "${POSTGRES_DB}" \
        -Fc \
        -f "${outfile}"
    chmod 600 "${outfile}"
    ok "Database backup complete"
}

backup_database_before_update

# 1) Fetch and check out the latest tag
log "Fetching latest release tag..."
LATEST_TAG=$(get_latest_release_tag)
log "Latest tag: ${LATEST_TAG}"

log "Checking out ${LATEST_TAG}..."
git fetch --tags --quiet
git checkout "${LATEST_TAG}"
ok "Now on ${LATEST_TAG}"

# 2) Update based on deployment mode
if [[ "${DEPLOY_MODE}" == "systemd" ]]; then
    # Systemd update
    log "Rebuilding backend services..."
    ./scripts/systemd/backend.sh
    ok "Backend rebuilt"

    log "Rebuilding frontend..."
    ./scripts/systemd/ui.sh
    ok "Frontend redeployed"

    SERVICES=(api log_stream user_expiry)
    if [[ "${TELEGRAM_BOT_ENABLED:-false}" == "true" ]]; then
        SERVICES+=("telegram_bot")
    fi
    for svc in "${SERVICES[@]}"; do
        if systemctl list-unit-files | grep -q "^${svc}.service"; then
            log "Restarting ${svc}.service"
            sudo systemctl restart "${svc}.service" || warn "failed to restart ${svc}"
        fi
    done

elif [[ "${DEPLOY_MODE}" == "docker" ]]; then
    # Docker update
    log "Shutting down Docker Compose stack..."
    if [[ "${TELEGRAM_BOT_ENABLED:-false}" == "true" ]]; then
        sudo docker compose -f docker-compose.yml -f docker-compose-telegram.yml down
    else
        sudo docker compose down
    fi

    log "Rebuilding Docker images..."
    if [[ "${TELEGRAM_BOT_ENABLED:-false}" == "true" ]]; then
        sudo docker compose -f docker-compose.yml -f docker-compose-telegram.yml build
    else
        sudo docker compose build
    fi

    log "Starting Docker Compose stack..."
    if [[ "${TELEGRAM_BOT_ENABLED:-false}" == "true" ]]; then
        sudo docker compose -f docker-compose.yml -f docker-compose-telegram.yml up -d
    else
        sudo docker compose up -d
    fi

    ok "Docker Compose deployment updated!"
fi

ok "Update completed successfully."
HOST_HINT="${HOST:-<your-host>}"
print_message highlight "🌐 Frontend served at https://${HOST_HINT}:3443"
print_message highlight "💡 Hard-refresh the panel (Ctrl+Shift+R) to drop any cached bundle."
