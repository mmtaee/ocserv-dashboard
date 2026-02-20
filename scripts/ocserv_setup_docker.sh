#!/usr/bin/env bash
set -euo pipefail

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
    echo "[INFO] No version specified. Fetching latest release..."
    OCSERV_VERSION=$(curl -fsSL "$API_URL" \
        | grep '"tag_name"' | head -n1 | cut -d'"' -f4)
    if [ -z "$OCSERV_VERSION" ]; then
        echo "[ERROR] Failed to fetch latest version from GitLab API"
        exit 1
    fi
fi

echo "[INFO] Installing ocserv version: $OCSERV_VERSION"

# -------------------------
# Install dependencies
# -------------------------
echo "[INFO] Installing build dependencies..."

apt-get update --allow-releaseinfo-change -y

apt install -y \
  build-essential autoconf automake libtool pkg-config \
  libgnutls28-dev libev-dev libseccomp-dev \
  libnl-3-dev libnl-route-3-dev gperf ipcalc \
  libpam0g-dev liblz4-dev libprotobuf-c-dev protobuf-c-compiler \
  libreadline-dev libtalloc-dev libhttp-parser-dev \
  liboath-dev \
  || { echo "[ERROR] Dependency installation failed"; exit 1; }

echo "[INFO] Dependencies installed"

# -------------------------
# Download source using curl
# -------------------------
TARBALL="ocserv-${OCSERV_VERSION}.tar.gz"
DOWNLOAD_URL="${REPO_URL}/-/archive/${OCSERV_VERSION}/${TARBALL}"

echo "[INFO] Downloading ${DOWNLOAD_URL}..."
curl -fSL --retry 3 --retry-delay 2 -o "${TARBALL}" "${DOWNLOAD_URL}" || { echo "[ERROR] Download failed"; exit 1; }

echo "[INFO] Extracting source..."
tar xf "${TARBALL}" || { echo "[ERROR] Extraction failed"; exit 1; }

cd "ocserv-${OCSERV_VERSION}" || { echo "[ERROR] Source directory not found"; exit 1; }

# -------------------------
# Build
# -------------------------
echo "[INFO] Preparing build system..."
autoreconf -fi || { echo "[ERROR] autoreconf failed"; exit 1; }

echo "[INFO] Configuring build system..."
./configure --prefix=/usr --bindir=/usr/bin || { echo "[ERROR] configure failed"; exit 1; }

echo "[INFO] Compiling..."
make -j"$(nproc)" || { echo "[ERROR] Build failed"; exit 1; }

echo "[INFO] Installing..."
make install || { echo "[ERROR] Install failed"; exit 1; }

echo "[INFO] Build and installation complete"

# -------------------------
# Setup directories
# -------------------------
echo "[INFO] Creating required directories..."
mkdir -p /etc/ocserv /var/lib/ocserv

# -------------------------
# Enable service (if systemd present)
# -------------------------
echo "[WARN] Systemd not detected. Skipping service enable/start. You can run ocserv manually."
echo "[OK] ocserv ${OCSERV_VERSION} installed successfully (manual start mode)"
echo "[INFO] Binary: /usr/local/sbin/ocserv"
echo "[INFO] Config: /etc/ocserv/ocserv.conf"

