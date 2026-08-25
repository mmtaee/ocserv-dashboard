#!/usr/bin/env bash

set -Eeuo pipefail

export AGENT_NODE=true
exec /usr/local/bin/container-entrypoint-common "$@"
