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
if [[ -e "${SOURCE_DIR}" ]]; then
    printf '[install] installation source already exists: %s\n' "${SOURCE_DIR}"
    printf '  1. Clear it and download the latest release\n  2. Continue setup with the existing source\n'
    read -r -p 'Selection: ' selection </dev/tty || die 'could not read selection from the terminal'
    case "${selection}" in
        1) "${root[@]}" rm -rf -- "${SOURCE_DIR}" ;;
        2)
            [[ -f "${SOURCE_DIR}/setup.sh" ]] || die "setup script not found: ${SOURCE_DIR}/setup.sh"
            exec "${root[@]}" bash "${SOURCE_DIR}/setup.sh" "$@"
            ;;
        *) die 'invalid selection' ;;
    esac
fi

release="$(git ls-remote --tags --refs --sort=-version:refname "${REPOSITORY}" 'v*' | awk -F/ 'NR == 1 { print $3 }')"
[[ "${release}" =~ ^v[0-9]+(\.[0-9]+)*$ ]] || die 'could not determine the latest release tag'

printf '[install] downloading release %s\n' "${release}"
"${root[@]}" install -d -m 755 "${INSTALL_ROOT}"
"${root[@]}" git clone --depth 1 --branch "${release}" "${REPOSITORY}" "${SOURCE_DIR}"

if [[ -f "${SOURCE_DIR}/setup.sh" ]]; then
    exec "${root[@]}" bash "${SOURCE_DIR}/setup.sh" "$@"
fi
exec "${root[@]}" bash "${SOURCE_DIR}/install.sh" "$@"
