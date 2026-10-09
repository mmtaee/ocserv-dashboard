#!/usr/bin/env bash

set -Eeuo pipefail

# shellcheck disable=SC2155
readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2155
readonly PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
readonly BACKEND_DIR="${PROJECT_ROOT}/backend"
readonly SWAG_PACKAGE="github.com/swaggo/swag/cmd/swag"
readonly SWAG_VERSION="${SWAG_VERSION:-v1.16.6}"

die() {
    printf 'swagger: %s\n' "$*" >&2
    exit 1
}

main() {
    command -v go >/dev/null 2>&1 || die "Go is not installed"
    [[ -f "${BACKEND_DIR}/main.go" ]] || die "backend entry point not found: ${BACKEND_DIR}/main.go"

    printf 'swagger: generating backend API documentation\n'
    cd -- "${BACKEND_DIR}"
    go run "${SWAG_PACKAGE}@${SWAG_VERSION}" init \
        --generalInfo main.go \
        --output docs \
        --outputTypes go,json,yaml
    printf 'swagger: updated %s\n' "${BACKEND_DIR}/docs"
}

main "$@"
