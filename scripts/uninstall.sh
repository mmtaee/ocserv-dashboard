#!/bin/bash

# ==============================================================
# Script: uninstall.sh
# Description: Uninstall ocserv_user_management application
#              and clean up all related files, services, and Docker containers.
# Usage:
#   sudo ./uninstall.sh
# ==============================================================

# Load logging helpers
source ./scripts/lib.sh

# ===============================
# Function: uninstall_docker
# Description: Stop and remove Docker Compose stack and images
# ===============================
uninstall_docker() {
    log "🛑 Stopping and removing Docker Compose stack..."
    sudo docker compose down 2>/dev/null || true

    read -rp "Do you want to remove pulled Docker images? [y/N]: " remove_images
    remove_images=${remove_images:-N}

    if [[ "$remove_images" =~ ^[Yy]$ ]]; then
        log "🗑️ Removing pulled Docker images..."
        sudo docker rmi golang:1.25 debian:trixie-slim nginx:alpine -f 2>/dev/null || true
        ok "✅ Docker images removed."
    else
        log "⏭️ Skipping removal of Docker images."
    fi

    local volume_dir="/opt/ocserv_dashboard/docker_volumes"
    if [[ -d "$volume_dir" ]]; then
        warn "📦 Docker Volume found in $volume_dir"
        read -rp "Do you want to remove Docker volumes under $volume_dir? [y/N]: " remove_vol
        remove_vol=${remove_vol:-N}
        if [[ "$remove_vol" =~ ^[Yy]$ ]]; then
            log "🗑️ Removing Docker volumes in $volume_dir..."
            sudo rm -rf $volume_dir
            ok "✅ Docker volumes $volume_dir removed."
        else
            log "⏭️ Skipping removal of Docker volumes."
        fi
    fi

    log "🧹 Docker environment and installation cleaned."
}

# ===============================
# Function: uninstall_systemd
# Description: Stop systemd services and remove binaries and configs
# ===============================
uninstall_systemd() {
    local services=("api" "log_stream" "user_expiry" "ocserv")
    local bin_dir="/opt/ocserv_dashboard"

    log "🛑 Stopping systemd services..."
    for service in "${services[@]}"; do
        sudo systemctl stop "$service" 2>/dev/null || true
        sudo systemctl disable "$service" 2>/dev/null || true
        sudo rm -f "/etc/systemd/system/${service}.service"
    done
    sudo systemctl daemon-reload

    # Ask user whether to keep or remove all data in /opt/ocserv_dashboard
    if [[ -d "$bin_dir" ]]; then
        warn "📂 found data in $bin_dir"
        read -rp "Do you want to purge all data in $bin_dir? [y/N]: " purge_data
        purge_data=${purge_data:-N}
        if [[ "$purge_data" =~ ^[Yy]$ ]]; then
            log "🗑️ Purging $bin_dir ..."
            sudo rm -rf "$bin_dir"
            ok "✅ All data removed."
        else
            log "📂 Keeping existing data in $bin_dir"
        fi
    fi

    warn "⚠️ You are about to uninstall PostgreSQL."
    read -rp "Do you want to proceed? [y/N]: " confirm
    confirm=${confirm:-N}
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        log "🗑️ Uninstalling PostgreSQL..."
        sudo apt remove postgresql-17 -y
        ok "✅ PostgreSQL removed."
    else
        log "⏭️ Skipping PostgreSQL removal."
    fi

    log "🌐 Removing Nginx frontend files..."
    sudo rm -rf /var/www/site
    sudo rm -f /etc/nginx/conf.d/site.conf

    log "🔄 Restarting Nginx service..."
    sudo systemctl restart nginx 2>/dev/null || true


    warn "⚠️ You are about to uninstall Nginx."
    read -rp "Do you want to proceed? [y/N]: " confirm
    confirm=${confirm:-N}
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        log "🗑️ Uninstalling Nginx..."
        sudo apt remove nginx -y
        ok "✅ Nginx frontend and service removed."
    else
        log "⏭️ Skipping Nginx removal."
    fi

    warn "⚠️ You are about to remove Ocserv configuration and VPN NAT rules."
    read -rp "Do you want to proceed? [y/N]: " confirm
    confirm=${confirm:-N}
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        log "🛑 Removing ocserv configuration..."
        sudo systemctl stop ocserv 2>/dev/null || true
        sudo systemctl disable ocserv 2>/dev/null || true
        sudo rm -rf /etc/ocserv
        sudo apt remove ocserv -y

        log "🔥 Cleaning iptables NAT/forwarding rules..."
        sudo iptables -t nat -D POSTROUTING -s "${OC_NET:-172.16.24.0/24}" -o "${ETH:-eth0}" -j MASQUERADE 2>/dev/null || true
        sudo iptables -D FORWARD -s "${OC_NET:-172.16.24.0/24}" -o "${ETH:-eth0}" -j ACCEPT 2>/dev/null || true
        sudo iptables -D FORWARD -d "${OC_NET:-172.16.24.0/24}" -m state --state ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || true
        sudo netfilter-persistent save 2>/dev/null || true

        ok "✅ Ocserv and VPN NAT rules removed."
    else
        ok "✅ Skipping Ocserv removal."
    fi

    log "🧹 Systemd environment cleaned."
}


# ===============================
# Main Execution
# ===============================
main() {
    info "🚀 Uninstallation started"
    uninstall_docker
    uninstall_systemd
    ok "🎉 Uninstallation completed successfully!"
}

main
