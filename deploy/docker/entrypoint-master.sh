#!/usr/bin/env bash

set -Eeuo pipefail

export AGENT_NODE=false
exec /usr/local/bin/container-entrypoint-common "$@"
