#!/usr/bin/env bash

set -Eeuo pipefail

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
export DEPLOYMENT_AGENT_NODE=false
exec "${SCRIPT_DIR}/../install-common.sh" "$@"
