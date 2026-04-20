#!/bin/bash

# ==============================================================
# Load shared logging utilities
# (print_message, log, ok, warn, die are defined in lib.sh)
# ==============================================================
source ./scripts/lib.sh

# ==============================================================
# Install build package dependencies
# ==============================================================
sudo apt install -y build-essential meson ninja-build pkg-config git \
                    libgnutls28-dev libev-dev libreadline-dev libtasn1-bin \
                    libpam0g-dev liblz4-dev libseccomp-dev \
                    libnl-route-3-dev libkrb5-dev libradcli-dev \
                    libcurl4-gnutls-dev libcjose-dev libjansson-dev liboath-dev \
                    libprotobuf-c-dev libtalloc-dev libllhttp-dev protobuf-c-compiler \
                    gperf ipcalc


INSTALL_PREFIX="/usr"
SRC_DIR="/tmp/ocserv"

info "Preparing source directory..."

rm -rf "$SRC_DIR"
git clone --depth=1 https://gitlab.com/openconnect/ocserv.git "$SRC_DIR"

cd "$SRC_DIR" || exit 1

info "Configuring build (Meson)..."
meson setup build \
    --prefix="$INSTALL_PREFIX" \
    --sysconfdir=/etc

info "Compiling..."
meson compile -C build -j"$(nproc)"

info "Installing..."
meson install -C build

# -------------------------
# Cleanup build artifacts (IMPORTANT for Docker)
# -------------------------
info "Cleaning build files..."
cd /
rm -rf "$SRC_DIR"

# -------------------------
# Minimal runtime setup
# -------------------------
info "Creating runtime dirs..."
mkdir -p /etc/ocserv /var/run/ocserv

info "Adding ocserv user..."
id -u ocserv &>/dev/null || useradd -r -s /usr/sbin/nologin ocserv

info "Copying default config..."
if [ -f /usr/share/doc/ocserv/examples/sample.config ] || [ -f doc/sample.config ]; then
    cp doc/sample.config /etc/ocserv/ocserv.conf 2>/dev/null || true
fi

# -------------------------
# Optional: shrink binary
# -------------------------
if command -v strip &>/dev/null; then
    info "Stripping binary..."
    strip /usr/sbin/ocserv || true
fi

ok "Ocserv installed successfully"

# -------------------------
# Setup system unit
# -------------------------
cat <<'EOF' | sudo tee /etc/systemd/system/ocserv.service > /dev/null
[Unit]
Description=OpenConnect SSL VPN server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/sbin/ocserv --foreground --config /etc/ocserv/ocserv.conf
ExecReload=/bin/kill -HUP $MAINPID
PIDFile=/var/run/ocserv.pid
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

info "Systemd service created successfully"


