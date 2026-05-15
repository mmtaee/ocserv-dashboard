#!/bin/bash
# ==============================================================
# Library Script: lib.sh
# Description:
#   Shared helper functions and environment setup for deployment
#   scripts in the ocserv_user_management system.
#
#   Provides:
#     - Strict Bash safety flags (set -euo pipefail, traps)
#     - Colorized logging and message functions
#     - Exit helpers
#     - Version/tag handling functions
#
# Usage:
#   source ./script/lib.sh
# ==============================================================

# ==============================================================
# Bash Safety Settings
# ==============================================================
# - Exit on error
# - Treat unset variables as errors
# - Fail pipeline if any command fails
# - Set non-interactive frontend for apt
set -euo pipefail
trap 'echo "❌ Deployment failed at line $LINENO."; exit 1' ERR
export DEBIAN_FRONTEND=noninteractive

# ==============================================================
# Function: print_message
# Description:
#   Print formatted messages with colors
# Parameters:
#   $1 - type: info, success, warn, error, highlight
#   $2 - message string
# Usage:
#   print_message info "Starting deployment..."
# ==============================================================
print_message() {
    local type="$1"
    local message="$2"

    local RED="\e[31m"
    local GREEN="\e[32m"
    local YELLOW="\e[33m"
    local BLUE="\e[34m"
    local MAGENTA="\e[35m"
    local RESET="\e[0m"

    case "$type" in
        info)
            echo -e "${BLUE}[INFO]$message ${RESET} "
            ;;
        success)
            echo -e "${GREEN}[SUCCESS]$message ${RESET} "
            ;;
        warn)
            echo -e "${YELLOW}[WARN]$message ${RESET} "
            ;;
        error)
            echo -e "${RED}[ERROR]$message ${RESET} "
            ;;
        highlight)
            echo -e "${MAGENTA}$message${RESET}"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# ==============================================================
# Logging and Exit Helper Functions
# Description:
#   Convenience wrappers around print_message for common log levels
# Usage:
#   log "Informational message"
#   ok  "Operation completed successfully"
#   warn "This is a warning"
#   die  "Fatal error occurred"
# ==============================================================
log()  { print_message info    "$*"; }
info() { print_message info    "$*"; }
ok()   { print_message success "$*"; }
warn() { print_message warn    "$*"; }
die()  { print_message error   "$*"; exit 1; }

# ==============================================================
# Function: is_valid_version
# Description:
#   Validates if a string is a valid version tag (starts with v)
# Parameters:
#   $1 - version string to validate
# Returns:
#   0 if valid, 1 otherwise
# ==============================================================
is_valid_version() {
    [[ "$1" =~ ^v[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]
}

# ==============================================================
# Function: get_latest_release_tag
# Description:
#   Fetches the latest release tag from GitHub
# Returns:
#   stdout - latest tag
#   exit 0 - success
#   exit 1 - failure
# ==============================================================
get_latest_release_tag() {
    local latest_tag
    latest_tag=$(
        curl -fsSL https://api.github.com/repos/mmtaee/ocserv-dashboard/releases/latest \
        | grep '"tag_name":' \
        | sed -E 's/.*"([^"]+)".*/\1/'
    )

    if ! is_valid_version "$latest_tag"; then
        echo "Failed to get valid latest release" >&2
        return 1
    fi

    echo "$latest_tag"
}

# ==============================================================
# Notes:
#   - All deployment scripts should source this file at the top:
#       source ./script/lib.sh
#   - Avoid duplicating safety flags in each script.
# ==============================================================
