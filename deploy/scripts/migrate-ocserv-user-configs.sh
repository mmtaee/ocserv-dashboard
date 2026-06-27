#!/bin/bash
# ==============================================================
# Script: migrate-ocserv-user-configs.sh
# Description:
#   One-time, idempotent migration for existing installations that
#   may have empty per-user ocserv config files under /etc/ocserv/users.
#
#   Empty per-user config files block config-per-group inheritance.
#   This script replaces empty per-user files with symlinks to the
#   corresponding group config file.
#
#   Non-empty custom per-user config files are preserved.
#
# Optional environment:
#   OCSERV_CONFIG_DIR - ocserv config root, default: /etc/ocserv
# ==============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}" || exit 1

# shellcheck source=/dev/null
source ./scripts/lib.sh

if [[ -f "${ROOT_DIR}/.env" ]]; then
  set -o allexport
  # shellcheck disable=SC1091
  source "${ROOT_DIR}/.env"
  set +o allexport
fi

POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-ocserv}"
POSTGRES_USER="${POSTGRES_USER:-ocserv}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"

OCSERV_CONFIG_DIR="${OCSERV_CONFIG_DIR:-/etc/ocserv}"
OCSERV_USERS_DIR="${OCSERV_CONFIG_DIR}/users"
OCSERV_GROUPS_DIR="${OCSERV_CONFIG_DIR}/groups"

if ! command -v psql >/dev/null 2>&1; then
  die "psql not found. Install PostgreSQL client tools before running this migration."
fi

if [[ ! -d "${OCSERV_GROUPS_DIR}" ]]; then
  warn "Ocserv groups directory not found: ${OCSERV_GROUPS_DIR}; skipping user config migration."
  exit 0
fi

mkdir -p "${OCSERV_USERS_DIR}"

log "Migrating ocserv per-user config files"
log "Database: ${POSTGRES_DB} @ ${POSTGRES_HOST}:${POSTGRES_PORT}"
log "Ocserv config directory: ${OCSERV_CONFIG_DIR}"

migrated=0
kept=0
skipped=0

while IFS=$'\t' read -r username group; do
  if [[ -z "${username}" || -z "${group}" ]]; then
    continue
  fi

  if [[ "${username}" == */* || "${group}" == */* ]]; then
    warn "Skipping unsafe username/group: ${username} -> ${group}"
    skipped=$((skipped + 1))
    continue
  fi

  group_file="${OCSERV_GROUPS_DIR}/${group}"
  user_file="${OCSERV_USERS_DIR}/${username}"

  if [[ ! -f "${group_file}" ]]; then
    warn "Skipping ${username}: missing group file ${group_file}"
    skipped=$((skipped + 1))
    continue
  fi

  if [[ -d "${user_file}" && ! -L "${user_file}" ]]; then
    warn "Keeping ${username}: ${user_file} is a directory"
    kept=$((kept + 1))
    continue
  fi

  if [[ -e "${user_file}" && ! -L "${user_file}" && -s "${user_file}" ]]; then
    log "Keeping ${username}: non-empty custom user config"
    kept=$((kept + 1))
    continue
  fi

  rm -f "${user_file}"
  ln -s "${group_file}" "${user_file}"

  log "${username} -> ${group}"
  migrated=$((migrated + 1))
done < <(
  PGPASSWORD="${POSTGRES_PASSWORD}" psql \
    -h "${POSTGRES_HOST}" \
    -p "${POSTGRES_PORT}" \
    -U "${POSTGRES_USER}" \
    -d "${POSTGRES_DB}" \
    -AtF $'\t' \
    -c "SELECT username, \"group\"
        FROM ocserv_users
        WHERE \"group\" IS NOT NULL
          AND \"group\" <> ''
          AND \"group\" <> 'defaults'
          AND \"group\" <> '*';"
)

ok "Ocserv user config migration complete: migrated=${migrated}, kept=${kept}, skipped=${skipped}"

if command -v occtl >/dev/null 2>&1; then
  if occtl reload; then
    ok "ocserv configuration reloaded"
  else
    warn "occtl reload failed; ocserv may need to be reloaded manually"
  fi
else
  warn "occtl not found; ocserv may need to be reloaded manually"
fi
