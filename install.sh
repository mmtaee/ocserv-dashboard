#!/usr/bin/env bash

set -Eeuo pipefail

readonly REPOSITORY='https://github.com/mmtaee/ocserv-dashboard.git'
readonly INSTALL_ROOT='/opt/ocserv-dashboard'
readonly SOURCE_DIR="${INSTALL_ROOT}/source"
declare -a root

die() {
    printf '[install] ERROR: %s\n' "$*" >&2
    exit 1
}

command -v git >/dev/null 2>&1 || die 'git is required; install it, then rerun this script'
if [[ "${EUID}" -ne 0 ]]; then
    command -v sudo >/dev/null 2>&1 || die 'sudo is required to install under /opt'
    root=(sudo)
else
    root=()
fi
[[ ! -e "${INSTALL_ROOT}" ]] || die "installation directory already exists: ${INSTALL_ROOT}"

release="$(git ls-remote --tags --refs --sort=-version:refname "${REPOSITORY}" 'v*' | awk -F/ 'NR == 1 { print $3 }')"
[[ "${release}" =~ ^v[0-9]+(\.[0-9]+)*$ ]] || die 'could not determine the latest release tag'

printf '[install] downloading release %s\n' "${release}"
"${root[@]}" install -d -m 755 "${INSTALL_ROOT}"
"${root[@]}" git clone --depth 1 --branch "${release}" "${REPOSITORY}" "${SOURCE_DIR}"

if [[ -f "${SOURCE_DIR}/setup.sh" ]]; then
    exec "${root[@]}" bash "${SOURCE_DIR}/setup.sh" "$@"
fi
exec "${root[@]}" bash "${SOURCE_DIR}/install.sh" "$@"
