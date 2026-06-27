#!/bin/bash

# ==============================================================
# Load shared logging utilities
# (print_message, log, ok, warn, die are defined in lib.sh)
# ==============================================================
source ./scripts/lib.sh

# ===============================
# Function: check_systemd_os
# Description:
#   Validate that the host OS is supported for systemd deployment.
#   Supported OS: Ubuntu 20.04/22.04/24.04, Debian 11/12/13
# ===============================
check_systemd_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS_NAME=$ID
        OS_VERSION="${VERSION_ID//\"/}"
    else
        die "Cannot detect OS. /etc/os-release not found."
    fi

    if [[ "$OS_NAME" == "ubuntu" ]]; then
        [[ "$OS_VERSION" =~ ^(20.04|22.04|24.04)$ ]] || \
            die "Unsupported Ubuntu version: $OS_VERSION"
    elif [[ "$OS_NAME" == "debian" ]]; then
        [[ "$OS_VERSION" =~ ^(11|12|13)$ ]] || \
            die "Unsupported Debian version: $OS_VERSION"
    else
        die "Unsupported OS: $OS_NAME $OS_VERSION"
    fi

    print_message success "✅ OS supported for systemd deployment: $OS_NAME $OS_VERSION"
}

# ===============================
# Function: check_go_version
# Description:
#   Verify that Go is installed and meets minimum version requirement.
# Parameters:
#   $1 - minimum Go version (default: 1.25)
# ===============================
check_go_version() {
    local go_mod_file="services/api/go.mod"

    if [[ ! -f "$go_mod_file" ]]; then
        die "❌ go.mod not found at $go_mod_file"
    fi

    # Extract Go version from the go.mod file (e.g., "1.25")
    local required_version
    required_version=$(grep '^go ' "$go_mod_file" | awk '{print $2}')
    [[ -n "$required_version" ]] || die "❌ Could not read Go version from $go_mod_file"

    # Normalize required_version to include patch if missing
    if [[ ! "$required_version" =~ \.[0-9]+$ ]]; then
        required_version="${required_version}.0"
    fi

    if ! command -v go >/dev/null 2>&1; then
        die "Go is not installed. Install from: https://go.dev/doc/install"
    fi

    # Get current Go version (e.g., 1.25.5)
    local current_version
    current_version=$(go version | awk '{print $3}' | sed 's/^go//')

    # Ensure current_version includes patch number for comparison
    if [[ ! "$current_version" =~ \.[0-9]+$ ]]; then
        current_version="${current_version}.0"
    fi

    # Compare versions
    if dpkg --compare-versions "$current_version" "lt" "$required_version"; then
        die "Go version $current_version < required $required_version. Upgrade at https://go.dev/doc/install"
    fi

    print_message success "✅ Go version $current_version meets requirement (≥ $required_version)"
}


# ==========================================
# Function: ensure_node
# Description:
#   Ensures Node.js v23.x or higher exists.
#   Installs Node.js via NodeSource if missing or outdated.
#   Installs npm if missing.
#   Installs Yarn globally.
# ==========================================
ensure_node() {
    log "Checking Node.js..."

    REQUIRED_NODE_MAJOR="24"

    if command -v node >/dev/null 2>&1; then
        CURRENT_NODE_VERSION=$(node -v | sed 's/^v//')
    else
        CURRENT_NODE_VERSION=""
    fi

    CURRENT_NODE_MAJOR="${CURRENT_NODE_VERSION%%.*}"

    if [[ -z "$CURRENT_NODE_VERSION" || "$CURRENT_NODE_MAJOR" -lt "$REQUIRED_NODE_MAJOR" ]]; then
        warn "Node.js missing or outdated"

        # only purge if apt package actually exists
        if dpkg -l | grep -q "^ii  nodejs"; then
            sudo apt-get purge -y nodejs
        fi

        sudo rm -f /etc/apt/sources.list.d/nodesource.list
        sudo rm -f /usr/share/keyrings/nodesource.gpg

        curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key \
            | sudo gpg --dearmor -o /usr/share/keyrings/nodesource.gpg

        echo "deb [signed-by=/usr/share/keyrings/nodesource.gpg] https://deb.nodesource.com/node_${REQUIRED_NODE_MAJOR}.x nodistro main" \
            | sudo tee /etc/apt/sources.list.d/nodesource.list >/dev/null

        sudo apt-get update
        sudo apt-get install -y nodejs

        CURRENT_NODE_VERSION=$(node -v | sed 's/^v//')
        ok "Node.js installed: v$CURRENT_NODE_VERSION"
    else
        ok "Node.js is already installed: v$CURRENT_NODE_VERSION"
    fi

    if ! command -v npm >/dev/null 2>&1; then
        warn "npm not found. Installing..."
        sudo apt-get install -y npm
    fi

    if ! command -v yarn >/dev/null 2>&1; then
        sudo npm install -g yarn
        ok "Yarn installed"
    else
        ok "Yarn already installed"
    fi
}

check_systemd_os
check_go_version
ensure_node
