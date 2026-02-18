#!/usr/bin/env bash

source ./scripts/lib.sh

export PATH="/usr/sbin:$PATH"

PROJECT_ID=473862
API_URL="https://gitlab.com/api/v4/projects/${PROJECT_ID}/releases"

# Repo URL for downloading tarball
REPO_URL="https://gitlab.com/openconnect/ocserv"

# -------------------------
# Get requested version
# -------------------------
OCSERV_VERSION="${1:-}"

if [ -z "$OCSERV_VERSION" ]; then
    info "No version specified. Fetching latest release..."
    OCSERV_VERSION=$(curl -fsSL "$API_URL" \
        | grep '"tag_name"' | head -n1 | cut -d'"' -f4) \
        || die "Failed to fetch latest version from GitLab API"
fi

ok "Installing ocserv version: $OCSERV_VERSION"

# -------------------------
# Install dependencies
# -------------------------
info "Installing build dependencies..."
apt update --allow-releaseinfo-change -y || die "apt update failed"

apt install -y \
  build-essential autoconf automake libtool pkg-config \
  libgnutls28-dev libev-dev libseccomp-dev \
  libnl-3-dev libnl-route-3-dev gperf ipcalc\
  libpam0g-dev liblz4-dev libprotobuf-c-dev protobuf-c-compiler \
  libreadline-dev libtalloc-dev libhttp-parser-dev \
  liboath-dev gettext curl ca-certificates \
  || die "Dependency installation failed"

ok "Dependencies installed"

# -------------------------
# Download source using curl
# -------------------------
TARBALL="ocserv-${OCSERV_VERSION}.tar.gz"
DOWNLOAD_URL="${REPO_URL}/-/archive/${OCSERV_VERSION}/${TARBALL}"

info "API URL (for fetching releases): $API_URL"
info "Downloading ${DOWNLOAD_URL}"

# Use curl with retry and progress
curl -fSL --retry 3 --retry-delay 2 -o "${TARBALL}" "${DOWNLOAD_URL}" || die "Download failed"

info "Extracting source"
tar xf "${TARBALL}" || die "Extraction failed"

cd "ocserv-${OCSERV_VERSION}" || die "Source directory not found"

# -------------------------
# Build
# -------------------------
info "Preparing build system"
autoreconf -fi || die "autoreconf failed"

info "Configuring build system"

./configure --with-systemdsystemunitdir=/lib/systemd/system --prefix=/usr --bindir=/usr/bin || die "configure failed"

info "Compiling"
make -j"$(nproc)" || die "Build failed"

info "Installing"
make install || die "Install failed"

ok "Build and installation complete"

# -------------------------
# Setup directories
# -------------------------
info "Creating required directories"
mkdir -p /etc/ocserv /var/lib/ocserv

# -------------------------
# Setup system unit
# -------------------------
cat <<'EOF' | sudo tee /etc/systemd/system/ocserv.service > /dev/null
[Unit]
Description=OpenConnect SSL VPN server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/sbin/ocserv --foreground --config /etc/ocserv/ocserv.conf
ExecReload=/bin/kill -HUP $MAINPID
PIDFile=/var/run/ocserv.pid
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF


# -------------------------
# Enable service
# -------------------------
info "Enabling and starting systemd service"
systemctl daemon-reload
systemctl enable ocserv || warn "Could not enable ocserv service"
systemctl restart ocserv || die "Failed to start ocserv"

ok "ocserv ${OCSERV_VERSION} installed successfully!"
info "Binary: /usr/local/sbin/ocserv"
info "Config: /etc/ocserv/ocserv.conf"

systemctl --no-pager status ocserv || warn "Service status check failed"

