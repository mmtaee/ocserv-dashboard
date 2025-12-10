#!/bin/bash

# ========================================
# Script: setup-ocserv-env.sh
# Description: Interactive script to set up environment variables
#              for an OpenConnect VPN (ocserv) server.
# Author: [Your Name]
# Date: [Date]
# License: MIT (or your preferred license)
# ========================================

# Exit immediately if a command exits with a non-zero status
set -e

# ===============================
# Default Configuration
# ===============================
HOST=$(hostname -I | awk '{print $1}')  # Default host IP (local IP)
SSL_CN="End-way-Cisco-VPN"              # Default SSL common name
SSL_ORG="End-way"                       # Default organization name
SSL_EXPIRE=3650                         # SSL certificate expiration in days
OC_NET="172.16.24.0/24"             # Default VPN subnet
OCSERV_PORT=443                         # Default VPN port
OCSERV_DNS="8.8.8.8"                    # Default DNS server
LANGUAGES=en:English,zh:中文,ru:Русский,fa:فارسی,ar:العربية  # Supported languages
SECRET_KEY=$(openssl rand -hex 32)      # Secret key for app encryption (32 hex chars)
JWT_SECRET=$(openssl rand -hex 32)      # JWT signing secret (32 hex chars)
SSL_C=US
SSL_ST=CA
SSL_L=SanFrancisco

# ===============================
# Functions
# ===============================

# ===============================
# Ensure root or sudo access
# ===============================
ensure_root() {
    if ! command -v sudo >/dev/null 2>&1; then
        print_message error "❌ Error: sudo is not installed on this system."
        exit 1
    fi
}

# ===============================
# Function: print_message
# Description: Print formatted messages with colors
# Parameters:
#   $1 - type: info, success, warn, error, highlight
#   $2 - message string
# ===============================
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
            echo -e "${BLUE}[INFO]${RESET} $message"
            ;;
        success)
            echo -e "${GREEN}[SUCCESS]${RESET} $message"
            ;;
        warn)
            echo -e "${YELLOW}[WARN]${RESET} $message"
            ;;
        error)
            echo -e "${RED}[ERROR]${RESET} $message"
            ;;
        highlight)
            echo -e "${MAGENTA}$message${RESET}"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

# ===============================
# Function: choose_deployment
# Description: Ask user whether to deploy via Docker or systemd
# ===============================
choose_deployment() {
    print_message info "🚀 Deployment options:"
    print_message highlight "   [1] Docker"
    print_message highlight "   [2] systemd service"

    read -rp "Choose deployment method [1-2] (default = 1): " choice
    if [[ -z "$choice" ]]; then
        choice=1
    fi

    case "$choice" in
        1)
            DEPLOY_METHOD="docker"
            ;;
        2)
            DEPLOY_METHOD="systemd"
            ;;
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
# Description: Check if Docker and Docker Compose (plugin) are installed.
#              Show Docker info if installed.
#              If missing, show installation links and exit.
# ===============================
check_docker() {
    local missing=0

    # Check Docker
    if ! command -v sudo docker &> /dev/null; then
        print_message error "❌ Docker is not installed."
        missing=1
    else
        print_message success "✅ Docker is installed."
        # Show Docker version and info
        docker_info=$(sudo docker info --format 'Server Version: {{.ServerVersion}}')
        print_message highlight "🔹 $docker_info"
    fi

    # Check Docker Compose plugin
    if ! sudo docker compose version &> /dev/null; then
        print_message error "❌ Docker Compose (plugin) is not installed."
        missing=1
    else
        print_message success "✅ Docker Compose (plugin) is installed."
        compose_version=$(sudo docker compose version | head -n1)
        print_message highlight "🔹 $compose_version"
    fi

    # If either is missing, show installation instructions and exit
    if [[ $missing -eq 1 ]]; then
        print_message info "🔗 Please follow the official installation guides:"
        print_message highlight "   Docker: https://docs.docker.com/get-docker/"
        print_message highlight "   Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
}

# ===============================
# Function: check_systemd_os
# Description: Validate OS for systemd deployment.
#              Supported: Ubuntu 20.04/22.04/24.04, Debian 11/12/13
# ===============================
check_systemd_os() {
    # Detect OS
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS_NAME=$ID         # ubuntu or debian
        OS_VERSION=$VERSION_ID
    else
        print_message error "❌ Cannot detect OS. /etc/os-release not found."
        exit 1
    fi

    # Normalize version number (remove quotes if present)
    OS_VERSION="${OS_VERSION//\"/}"

    # Check supported OS
    if [[ "$OS_NAME" == "ubuntu" ]]; then
        if [[ "$OS_VERSION" != "20.04" && "$OS_VERSION" != "22.04" && "$OS_VERSION" != "24.04" ]]; then
            print_message error "❌ Unsupported Ubuntu version: $OS_VERSION"
            print_message info "Supported versions: 20.04, 22.04, 24.04"
            exit 1
        fi
    elif [[ "$OS_NAME" == "debian" ]]; then
        if [[ "$OS_VERSION" != "11" && "$OS_VERSION" != "12" && "$OS_VERSION" != "13" ]]; then
            print_message error "❌ Unsupported Debian version: $OS_VERSION"
            print_message info "Supported versions: 11, 12, 13"
            exit 1
        fi
    else
        print_message error "❌ Unsupported OS: $OS_NAME $OS_VERSION"
        print_message info "Supported: Ubuntu 20.04/22.04/24.04, Debian 11/12/13"
        exit 1
    fi

    print_message success "✅ OS is supported for systemd deployment: $OS_NAME $OS_VERSION"
}

# ===============================
# Function: check_go_version
# Description: Check if Go is installed and meets minimum version requirement
# Parameters:
#   $1 - minimum Go version (default: 1.25)
# ===============================
check_go_version() {
    local required_version="1.25"

    if ! command -v go >/dev/null 2>&1; then
        echo "❌ Go is not installed."
        echo "Install Go from: https://go.dev/doc/install"
        exit 1
    fi

    local current_version
    current_version=$(go version | awk '{print $3}' | sed 's/^go//')  # e.g., 1.25.3
    local current_major_minor="${current_version%.*}"                     # 1.25

    # Compare versions
    if dpkg --compare-versions "$current_major_minor" "lt" "$required_version"; then
        echo "❌ Go version $current_version is less than required $required_version."
        echo "Please upgrade Go: https://go.dev/doc/install"
        exit 1
    fi

    echo "✅ Go version $current_version meets requirement (≥ $required_version)."
}


# ===============================
# Function: get_ip
# Description: Detects the public IP of the VPS
#              and allows user to confirm or override.
# ===============================
get_ip() {
    print_message info "🔍 Detecting public IP ..."
#    set -x

    local detected_ip
    detected_ip=$(curl -s --max-time 5 https://api.ipify.org || \
                  curl -s --max-time 5 https://ifconfig.me || \
                  curl -s --max-time 5 https://checkip.amazonaws.com)
#    set +x
    print_message info "Detected IP: $detected_ip"

    if [[ -n "$detected_ip" && "$detected_ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_message highlight "✅ Detected public IP: ${detected_ip}"
        read -rp "👉 Do you want to use this IP? [Y/n]: " choice
        case "$choice" in
            [Nn]*)
                read -rp "🌐 Enter your VPS host or IP: " manual_ip
                HOST=${manual_ip:-$(hostname -I | awk '{print $1}')}
                ;;
            *)
                HOST=$detected_ip
                ;;
        esac
    else
        print_message warn "⚠️ Failed to detect public IP automatically."
        read -rp "🌐 Enter your VPS host or IP: " manual_ip
        HOST=${manual_ip:-$(hostname -I | awk '{print $1}')}
    fi

    print_message highlight "🔧 Using host IP: ${HOST}"
    printf "\n"
}

# ===============================
# Function: get_envs
# Description: Prompt user for environment configurations
#              and set default values if left blank.
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
# Description: Allow user to select the preferred site language
# ===============================
get_site_lang() {
    print_message info "🌐 Available languages:"
    IFS=',' read -ra langs <<< "$LANGUAGES"
    local i=1
    for entry in "${langs[@]}"; do
        code="${entry%%:*}"
        name="${entry#*:}"
        print_message highlight "   [$i] $name ($code)"
        i=$((i+1))
    done

    printf "\n"
    read -rp "Choose a language [1-${#langs[@]}] (default = all): " choice
    printf "\n"

    if [[ -z "$choice" ]]; then
        print_message highlight "✅ Using all available languages: $LANGUAGES"
        printf "\n"
    else
        if [[ ! "$choice" =~ ^[0-9]+$ || "$choice" -lt 1 || "$choice" -gt ${#langs[@]} ]]; then
            choice=1
        fi
        selected="${langs[$((choice-1))]}"
        LANGUAGES=$selected
        print_message highlight "✅ Selected site language: ${selected#*:} (${selected%%:*})"
        printf "\n"
    fi
}

# ===============================
# Function: set_environment
# Description: Generate a .env file with the configured environment variables
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
    print_message info "🔧 Environments written to .env:"
    print_message highlight "   HOST             = ${HOST}"
    print_message highlight "   SECRET_KEY       = ${SECRET_KEY:0:8}..."
    print_message highlight "   JWT_SECRET       = ${JWT_SECRET:0:8}..."
    print_message highlight "   SSL_CN           = ${SSL_CN}"
    print_message highlight "   SSL_ORG          = ${SSL_ORG}"
    print_message highlight "   SSL_C            = ${SSL_C}"
    print_message highlight "   SSL_ST           = ${SSL_ST}"
    print_message highlight "   SSL_L            = ${SSL_L}"
    print_message highlight "   OC_NET           = ${OC_NET}"
    print_message highlight "   SSL_EXPIRE       = ${SSL_EXPIRE}"
    print_message highlight "   OCSERV_PORT      = ${OCSERV_PORT}"
    print_message highlight "   OCSERV_DNS       = ${OCSERV_DNS}"
    print_message highlight "   LANGUAGES        = ${LANGUAGES}"
    print_message highlight "   ALLOW_ORIGINS    = https://${HOST}:3443"
    printf "\n"
}

# ===============================
# Function: setup_docker
# Description: Pull required Docker images and start Docker Compose stack
# ===============================
setup_docker() {
    print_message info "🚀 Pulling required Docker images..."
    sudo docker pull golang:1.25.0
    sudo docker pull debian:trixie-slim
    sudo docker pull nginx:alpine
    print_message success "🎉 All Docker images pulled successfully!"

    print_message info "🛠 Starting Docker Compose..."
    sleep 3
    sudo docker compose up --build -d
    print_message success "✅ Docker Compose deployment completed!"
}


# ===============================
# Function: GetInterface
# Description: List physical network interfaces and let user select one.
#              If only one interface exists, automatically select it.
# Sets:
#   ETH - selected network interface
# ===============================
get_interface() {
    printf "\n"

    # Get all physical interfaces (exclude lo, docker bridges, veth, tun, br-*, vethe*)
    interface_list=$(ip -o link show | awk '{print $2}' | tr -d ':' | grep -Ev '^(lo|docker|br-|veth|tun|vethe)')

    if [[ -z "$interface_list" ]]; then
        print_message error "❌ No physical network interfaces found!"
        exit 1
    fi

    # Convert to array
    numbered_interfaces=()
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
    i=1
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
        GetInterface
    fi
}


# ===============================
# Function: setup_systemd
# Description: Setup systemd deployment and check Go environment.
# ===============================
setup_systemd() {
    print_message info "⚙️ Setting up systemd service..."

    # Let user select physical interface
    get_interface

    # Export environment variables for child script
    export ETH
    export OCSERV_PORT
    export SSL_OC_NET
    export SSL_CN
    export SSL_ORG
    export SSL_EXPIRE
    export OCSERV_DNS

    # Run the systemd ocserv setup script
    ./scripts/systemd/full_setup.sh
}


# ===============================
# Function: deploy
# Description: Deploy the application based on the chosen deployment method.
#              - If Docker is selected:
#                  1. Pulls required Docker images (golang, debian, nginx)
#                  2. Builds and starts the Docker Compose stack
#              - If systemd is selected:
#                  Calls the systemd setup function (placeholder for now)
# Parameters:
#   DEPLOY_METHOD - must be either "docker" or "systemd"
# ===============================
deploy(){
    if [[ "$DEPLOY_METHOD" == "docker" ]]; then
        setup_docker
    else
        setup_systemd
    fi
}


# ===============================
# Main Execution
# ===============================
main() {
    # Ensure script is running as root or sudo
    ensure_root "$@"

    # Deployment choice: docker or systemd
    choose_deployment

    # install curl
    sudo apt install -y curl

    # Check prerequisites based on deployment method
    if [[ "$DEPLOY_METHOD" == "docker" ]]; then
        check_docker
    else
        check_systemd_os
        check_go_version
    fi

    # Load existing .env or run interactive setup
    ENV_FILE=".env"
    if [[ -f "$ENV_FILE" ]]; then
        print_message info "✅ .env file detected. Loading environment variables..."
        set -o allexport
        # shellcheck disable=SC1090
        source "${ENV_FILE}"
        set +o allexport
        print_message success "✅ Environment loaded from ${ENV_FILE}"
    else
        print_message info "⚡ No .env file found. Running interactive setup..."
        get_ip
        get_envs
        get_site_lang
        set_environment
    fi

    # Deploy application
    deploy

    # Show service URL
    print_message highlight "🌐 Web service is up and running!"
    print_message highlight "🔗 Access it at: https://${HOST}:3443 or http://${HOST}:3000"
    print_message highlight "⚡ Tip: Make sure your firewall allows ports 3000 and 3443"
    exit 0
}

main
