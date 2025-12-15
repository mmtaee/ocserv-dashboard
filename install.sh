#!/bin/bash
# ==============================================================
# Script: install.sh
# Description:
#   Interactive installer for the OpenConnect VPN dashboard
#   and backend services (ocserv + web UI). Provides options
#   for Docker-based or systemd-based deployments.
#
# Features:
#   - Detects system public IP and network interface
#   - Configures environment variables (.env)
#   - Checks prerequisites (Docker, Docker Compose, Go, OS)
#   - Builds frontend (Vite/React/Vue) and backend Go services
#   - Deploys via Docker Compose or systemd services
#   - Supports SSL certificate creation and VPN configuration
#
# Usage:
#   sudo ./install.sh
#
# Dependencies:
#   - Bash 5+
#   - sudo, curl, openssl
#   - Docker / Docker Compose (if Docker deployment)
#   - Go >= 1.25 (if systemd deployment)
# ==============================================================

# Load shared helpers
source ./scripts/lib.sh

# ===============================
# Default Configuration
# ===============================
HOST=$(hostname -I | awk '{print $1}')                        # Default host IP (local)
SSL_CN="End-way-Cisco-VPN"                                    # Default SSL common name
SSL_ORG="End-way"                                             # Default organization name
SSL_EXPIRE=3650                                               # SSL certificate expiration in days
OC_NET="172.16.24.0/24"                                       # Default VPN subnet
OCSERV_PORT=443                                               # Default VPN port
OCSERV_DNS="8.8.8.8"                                          # Default DNS server
LANGUAGES="en:English,zh:中文,ru:Русский,fa:فارسی,ar:العربية"  # Supported languages
SECRET_KEY=$(openssl rand -hex 32)                            # Secret key for app encryption (32 hex chars)
JWT_SECRET=$(openssl rand -hex 32)                            # JWT signing secret (32 hex chars)
SSL_C=US                                                      # SSL Country iso2
SSL_ST=CA                                                     # SSL State name
SSL_L=SanFrancisco                                            # SSl City name

# ===============================
# Function: ensure_root
# Description:
#   Ensure script is run with root privileges.
#   Exits if sudo is not installed or accessible.
# ===============================
ensure_root() {
    if ! command -v sudo >/dev/null 2>&1; then
        print_message error "❌ Error: sudo is not installed."
        exit 1
    fi
}

# ===============================
# Function: choose_deployment
# Description:
#   Prompt user to select deployment method:
#     1) Docker
#     2) Systemd service
#     3) Standalone systemd dashboard
#   Sets the global variable DEPLOY_METHOD
# ===============================
choose_deployment() {
    print_message info "🚀 Deployment Options:"
    print_message highlight "   [1] Docker"
    print_message highlight "   [2] Systemd Full (Ocserv + Dashboard)"
    print_message highlight "   [3] Systemd Dashboard (Standalone Setup/Upgrade)"
    print_message highlight "   [4] Uninstall"

    read -rp "Choose deployment method [1-4] (default = 1): " choice
    choice=${choice:-1}

    case "$choice" in
        1) DEPLOY_METHOD="docker" ;;
        2) DEPLOY_METHOD="systemd" ;;
        3) DEPLOY_METHOD="standalone" ;;
        4) DEPLOY_METHOD="uninstall" ;;
        *)
            print_message warn "Invalid choice, defaulting to Docker."
            DEPLOY_METHOD="docker"
            ;;
    esac

    print_message highlight "✅ Selected deployment method: ${DEPLOY_METHOD}"
    printf "\n"
}

# ===============================
# Function: check_docker
# Description:
#   Verify that Docker and Docker Compose plugin are installed.
#   Show version info. If missing, display installation links and exit.
# ===============================
check_docker() {
    local missing=0

    if ! command -v sudo docker &> /dev/null; then
        print_message error "❌ Docker is not installed."
        missing=1
    else
        print_message success "✅ Docker is installed."
        docker_info=$(sudo docker info --format 'Server Version: {{.ServerVersion}}')
        print_message highlight "🔹 $docker_info"
    fi

    if ! sudo docker compose version &> /dev/null; then
        print_message error "❌ Docker Compose (plugin) is not installed."
        missing=1
    else
        print_message success "✅ Docker Compose (plugin) is installed."
        compose_version=$(sudo docker compose version | head -n1)
        print_message highlight "🔹 $compose_version"
    fi

    if [[ $missing -eq 1 ]]; then
        print_message info "🔗 Installation guides:"
        print_message highlight "   Docker: https://docs.docker.com/get-docker/"
        print_message highlight "   Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
}

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


# ===============================
# Function: get_ip
# Description:
#   Detects public IP and prompts user to confirm or override
# ===============================
get_ip() {
    print_message info "🔍 Detecting public IP ..."
    local detected_ip
    detected_ip=$(curl -s --max-time 5 https://api.ipify.org || \
                  curl -s --max-time 5 https://ifconfig.me || \
                  curl -s --max-time 5 https://checkip.amazonaws.com)

    print_message info "Detected IP: $detected_ip"

    if [[ "$detected_ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        read -rp "Use this IP? [Y/n]: " choice
        HOST=${detected_ip}
        [[ "$choice" =~ [Nn] ]] && read -rp "Enter your VPS host or IP: " HOST
    else
        read -rp "Enter your VPS host or IP: " HOST
    fi

    print_message highlight "🔧 Using host IP: ${HOST}"
    printf "\n"
}

# ===============================
# Function: get_envs
# Description:
#   Prompt user for ocserv and SSL environment configurations.
# ===============================
get_envs(){
    # ocserv port
    read -rp "Enter your ocserv port or leave blank to use ${OCSERV_PORT}: " port
    [[ -n "$port" ]] && OCSERV_PORT=$port
    print_message highlight "✅ Using port: ${OCSERV_PORT}"
    printf "\n"

    # Company Name
    read -rp "Enter your company name or leave blank to use '${SSL_CN}': " cn
    [[ -n "$cn" ]] && SSL_CN=$cn
    print_message highlight "✅ Using company name: ${SSL_CN}"
    printf "\n"

    # Organization Name
    read -rp "Enter your organization name or leave blank to use '${SSL_ORG}': " org
    [[ -n "$org" ]] && SSL_ORG=$org
    print_message highlight "✅ Using organization name: ${SSL_ORG}"
    printf "\n"

    # Country
    read -rp "Enter your country code (2 letters) or leave blank to use '${SSL_C}': " country
    [[ -n "$country" ]] && SSL_C=$country
    print_message highlight "✅ Using country: ${SSL_C}"
    printf "\n"

    # State / Province
    read -rp "Enter your state or leave blank to use '${SSL_ST}': " state
    [[ -n "$state" ]] && SSL_ST=$state
    print_message highlight "✅ Using state: ${SSL_ST}"
    printf "\n"

    # Locality / City
    read -rp "Enter your city or leave blank to use '${SSL_L}': " locality
    [[ -n "$locality" ]] && SSL_L=$locality
    print_message highlight "✅ Using city: ${SSL_L}"
    printf "\n"

    # SSL Expiration Days
    read -rp "Enter SSL expire days or leave blank to use ${SSL_EXPIRE} days: " expire
    [[ -n "$expire" ]] && SSL_EXPIRE=$expire
    print_message highlight "✅ Using SSL expiration days: ${SSL_EXPIRE}"
    printf "\n"

    # ocserv IPv4 Network
    read -rp "Enter ocserv IPv4 network or leave blank to use ${OC_NET}: " oc_net
    [[ -n "$oc_net" ]] && OC_NET=$oc_net
    print_message highlight "✅ Using ocserv IPv4 network: ${OC_NET}"
    printf "\n"

    # ocserv DNS
    read -rp "Enter your DNS server or leave blank to use default (${OCSERV_DNS}): " dns
    [[ -n "$dns" ]] && OCSERV_DNS="$dns"
    print_message highlight "✅ Using ocserv DNS: ${OCSERV_DNS}"
    printf "\n"
}

# ===============================
# Function: get_site_lang
# Description:
#   Prompt user to select the preferred site language from LANGUAGES
# ===============================
get_site_lang() {
    print_message info "🌐 Available languages:"
    IFS=',' read -ra langs <<< "$LANGUAGES"
    local i=1
    for entry in "${langs[@]}"; do
        print_message highlight "   [$i] ${entry#*:} (${entry%%:*})"
        ((i++))
    done

    read -rp "Choose a language [1-${#langs[@]}] (default = all): " choice
    if [[ -z "$choice" ]]; then
        print_message highlight "✅ Using all languages: $LANGUAGES"
    else
        [[ "$choice" -lt 1 || "$choice" -gt ${#langs[@]} ]] && choice=1
        LANGUAGES="${langs[$((choice-1))]}"
        print_message highlight "✅ Selected language: ${LANGUAGES#*:} (${LANGUAGES%%:*})"
    fi
    printf "\n"
}

# ===============================
# Function: set_environment
# Description:
#   Create .env file containing all environment variables
# ===============================
set_environment() {
    ENV_FILE=".env"
    print_message info "Creating environment file at $ENV_FILE ..."
    cat > "$ENV_FILE" <<EOL
HOST=${HOST}
SECRET_KEY=${SECRET_KEY}
JWT_SECRET=${JWT_SECRET}
SSL_CN=${SSL_CN}
SSL_ORG=${SSL_ORG}
OC_NET=${OC_NET}
SSL_C=${SSL_C}
SSL_ST=${SSL_ST}
SSL_L=${SSL_L}
SSL_EXPIRE=${SSL_EXPIRE}
OCSERV_PORT=${OCSERV_PORT}
OCSERV_DNS=${OCSERV_DNS}
LANGUAGES="${LANGUAGES}"
ALLOW_ORIGINS=https://${HOST}:3443
EOL
    print_message success "✅ Environment file created successfully."
}

# ===============================
# Function: get_interface
# Description:
#   Lists physical network interfaces and lets user select one.
#   Automatically selects if only one exists.
# Sets:
#   ETH - selected network interface
# ===============================
get_interface() {
    printf "\n"

    # Get all physical interfaces (exclude lo, docker bridges, veth, tun, br-*, vethe*)
    local interface_list
    interface_list=$(ip -o link show | awk '{print $2}' | tr -d ':' | grep -Ev '^(lo|docker|br-|veth|tun|vethe)')

    if [[ -z "$interface_list" ]]; then
        die "❌ No physical network interfaces found!"
    fi

    # Convert to array
    local numbered_interfaces=()
    for iface in $interface_list; do
        numbered_interfaces+=("$iface")
    done

    # If only one interface exists, auto-select it
    if [[ ${#numbered_interfaces[@]} -eq 1 ]]; then
        ETH="${numbered_interfaces[0]}"
        print_message highlight "✅ Only one physical interface found. Auto-selected: $ETH"
        return
    fi

    # Multiple interfaces: show numbered list
    print_message highlight "Available physical network interfaces:"
    local i=1
    for iface in "${numbered_interfaces[@]}"; do
        print_message highlight "$(printf "%4d: %s" "$i" "$iface")"
        ((i++))
    done

    # Prompt user for selection
    read -rp "Enter the number corresponding to the desired network interface: " interface_number
    if [[ "$interface_number" =~ ^[0-9]+$ ]] && (( interface_number >= 1 && interface_number <= ${#numbered_interfaces[@]} )); then
        ETH="${numbered_interfaces[$((interface_number-1))]}"
        print_message highlight "✅ Selected interface: $ETH"
        printf "\n"
    else
        print_message error "❌ Invalid selection: $interface_number. Please try again."
        printf "\n"
        get_interface
    fi
}

# ===============================
# Function: setup_docker
# Description:
#   Pull required Docker images and start Docker Compose stack
# ===============================
setup_docker() {
    print_message info "🚀 Pulling required Docker images..."
    sudo docker pull golang:1.25.0
    sudo docker pull debian:trixie-slim
    sudo docker pull nginx:alpine
    print_message success "🎉 All Docker images pulled successfully!"
    print_message info "🛠 Starting Docker Compose..."
    sudo docker compose up --build -d
    print_message success "✅ Docker Compose deployment completed!"
}

# ===============================
# Function: setup_systemd
# Description:
#   Sets up systemd deployment for backend, UI, and optionally ocserv VPN.
#   - full_setup = true  : deploy backend + UI + VPN (ocserv)
#   - full_setup = false : deploy backend + UI only, but ensures ocserv is configured
# ===============================
setup_systemd() {
    local full_setup="$1"

    export OCSERV_PORT SSL_CN SSL_ORG SSL_EXPIRE OCSERV_DNS

    # If not full setup, ensure ocserv is installed and configured
    if [[ "$full_setup" != true ]]; then
        if ! command -v ocserv >/dev/null 2>&1; then
            warn "⚠️ Ocserv is not installed. Skipping VPN setup."
        elif [[ ! -f /etc/ocserv/ocserv.conf ]]; then
            warn "⚠️ /etc/ocserv/ocserv.conf not found. Skipping VPN setup."
        else
            # Check if auth line exists
            if ! grep -q '^auth\s*=\s*"plain\[passwd=/etc/ocserv/ocpasswd\]"' /etc/ocserv/ocserv.conf; then
                warn "⚠️ Ocserv config found, but auth line is missing. Skipping VPN setup."
            else
                ok "✅ Ocserv is installed and properly configured."
            fi
        fi
    fi

    # Deploy backend and UI systemd services
    ./scripts/systemd_backend.sh
    ./scripts/systemd_ui.sh

    # Deploy VPN (ocserv) only if full setup requested
    if [[ "$full_setup" == true ]]; then
          # Select network interface for NAT/firewall
          get_interface

          export ETH

        ./scripts/systemd_ocserv.sh
    fi
}

# ===============================
# Function: uninstall
# Description:
#   Runs the uninstall script to remove all deployed components.
#   Handles both Docker and systemd deployments, including:
#     - Stopping services
#     - Removing binaries
#     - Cleaning Nginx and SSL
#     - Cleaning iptables rules
#     - Optionally purging /opt/ocserv_dashboard or Docker volumes
# ===============================
uninstall() {
    ./scripts/uninstall.sh
}

# ===============================
# Function: deploy
# Description:
#   Deploys the application based on the selected DEPLOY_METHOD.
#   Supported methods:
#     - docker: pulls Docker images and runs Docker Compose stack
#     - systemd: sets up systemd services (full mode)
#     - standalone: sets up systemd dashboard only (no VPN)
#     - uninstall: removes all deployed components
# ===============================
deploy() {
    case "$DEPLOY_METHOD" in
        docker) setup_docker ;;
        systemd) setup_systemd true ;;
        standalone) setup_systemd false ;;
    esac
}

# ===============================
# Function: main
# Description:
#   Entry point for the deployment script.
#   1. Ensures root privileges.
#   2. Prompts the user to choose a deployment method:
#        - docker: deploy via Docker Compose
#        - systemd: deploy systemd services (full mode)
#        - standalone: deploy systemd dashboard only (no VPN)
#        - uninstall: remove all deployed components
#   3. Skips all setup logic if 'uninstall' is selected.
#   4. Installs prerequisites (curl), checks environment.
#   5. Loads or interactively generates .env environment file.
#   6. Calls the deployment function based on the selected method.
#   7. Prints final service access information.
# ===============================
main() {
    # Ensure script is running with root privileges
    ensure_root

    # Let user choose deployment method
    choose_deployment

    # If uninstall mode, run uninstall and exit immediately
    if [[ "$DEPLOY_METHOD" == "uninstall" ]]; then
        print_message info "⚠️ Uninstall mode selected. Removing deployed components..."
        uninstall
        exit 0
    fi

    # Install required tools
    sudo apt install -y curl

    # Check prerequisites based on deployment method
    if [[ "$DEPLOY_METHOD" == "docker" ]]; then
        check_docker
    else
        check_systemd_os
        check_go_version
    fi

    # Load existing environment file or run interactive setup
    ENV_FILE=".env"
    if [[ -f "$ENV_FILE" ]]; then
        print_message info "✅ Loading environment from $ENV_FILE"
        set -o allexport
        # shellcheck disable=SC1090
        source "$ENV_FILE"
        set +o allexport
        print_message success "✅ Environment loaded"
    else
        print_message info "⚡ No .env found. Running interactive setup..."
        get_ip
        get_envs
        get_site_lang
        set_environment
    fi

    # Deploy based on selected method
    deploy

    # Show final access information
    print_message highlight "🌐 Web service is running at:"
    print_message highlight "   https://${HOST}:3443 or http://${HOST}:3000"
    print_message highlight "⚡ Ensure firewall allows ports 3000 and 3443"

    exit 0
}

# ===============================
# Run the main function
# ===============================
main

