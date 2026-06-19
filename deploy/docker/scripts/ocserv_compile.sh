#!/bin/bash


INSTALL_PREFIX="/usr"
SRC_DIR="/tmp/ocserv"

echo "Preparing source directory..."

rm -rf "$SRC_DIR"
git clone --depth=1 https://gitlab.com/openconnect/ocserv.git "$SRC_DIR"

cd "$SRC_DIR" || exit 1

echo "Configuring build (Meson)..."
meson setup build \
    --prefix="$INSTALL_PREFIX" \
    --sysconfdir=/etc

echo "Compiling..."
meson compile -C build -j"$(nproc)"

echo "Installing..."
meson install -C build

# -------------------------
# Cleanup build artifacts (IMPORTANT for Docker)
# -------------------------
echo "Cleaning build files..."
cd /
rm -rf "$SRC_DIR"

# -------------------------
# Minimal runtime setup
# -------------------------
echo "Creating runtime dirs..."
mkdir -p /etc/ocserv /var/run/ocserv

echo "Adding ocserv user..."
id -u ocserv &>/dev/null || useradd -r -s /usr/sbin/nologin ocserv

echo "Copying default config..."
if [ -f /usr/share/doc/ocserv/examples/sample.config ] || [ -f doc/sample.config ]; then
    cp doc/sample.config /etc/ocserv/ocserv.conf 2>/dev/null || true
fi

# -------------------------
# Optional: shrink binary
# -------------------------
if command -v strip &>/dev/null; then
    echo "Stripping binary..."
    strip /usr/sbin/ocserv || true
fi

echo "Ocserv installed successfully"