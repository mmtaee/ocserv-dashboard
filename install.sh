#!/usr/bin/env bash

set -Eeuo pipefail

readonly REPOSITORY='https://github.com/mmtaee/ocserv-dashboard.git'
readonly INSTALL_DIR="${INSTALL_DIR:-${PWD}/ocserv-dashboard}"

die() {
    printf '[install] ERROR: %s\n' "$*" >&2
    exit 1
}

command -v git >/dev/null 2>&1 || die 'git is required; install it, then rerun this script'
[[ ! -e "${INSTALL_DIR}" ]] || die "installation directory already exists: ${INSTALL_DIR}"

release="$(git ls-remote --tags --refs --sort=-version:refname "${REPOSITORY}" 'v*' | awk -F/ 'NR == 1 { print $3 }')"
[[ "${release}" =~ ^v[0-9]+(\.[0-9]+)*$ ]] || die 'could not determine the latest release tag'

printf '[install] downloading release %s\n' "${release}"
git clone --depth 1 --branch "${release}" "${REPOSITORY}" "${INSTALL_DIR}"
printf '%s\n' "${release}" >"${INSTALL_DIR}/.release"

if [[ -f "${INSTALL_DIR}/setup.sh" ]]; then
    exec bash "${INSTALL_DIR}/setup.sh" "$@"
fi
exec bash "${INSTALL_DIR}/install.sh" "$@"
