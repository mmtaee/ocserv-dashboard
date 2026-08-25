#!/usr/bin/env bash

set -Eeuo pipefail

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
export DEPLOYMENT_AGENT_NODE=true
exec "${SCRIPT_DIR}/../install-common.sh" "$@"
