#!/bin/bash
# ===============================
# Script: systemd_postgres.sh
# ===============================

set -e
trap 'die "Error on line $LINENO"' ERR

# Load helpers
source ./scripts/lib.sh

# ===============================
# Load environment variables
# ===============================
ENV_FILE=".env"

if [ ! -f "$ENV_FILE" ]; then
    die ".env file not found"
fi

export $(grep -v '^#' "$ENV_FILE" | xargs)

[ -z "$POSTGRES_DB" ] && die "POSTGRES_DB is not set"
[ -z "$POSTGRES_USER" ] && die "POSTGRES_USER is not set"
[ -z "$POSTGRES_PASSWORD" ] && die "POSTGRES_PASSWORD is not set"

ok "Environment loaded"

# ===============================
# Install PostgreSQL 17
# ===============================
ok "Installing PostgreSQL 17..."

sudo rm -f /etc/apt/sources.list.d/pgdg.list
sudo rm -f /usr/share/keyrings/postgresql.gpg

sudo apt update -y
sudo apt install -y wget gnupg lsb-release curl

sudo mkdir -p /usr/share/keyrings

# Add PostgreSQL repo
if [ ! -f /etc/apt/sources.list.d/pgdg.list ]; then
    log "Adding PostgreSQL repository..."

    curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
        | sudo gpg --dearmor -o /usr/share/keyrings/postgresql.gpg

    echo "deb [signed-by=/usr/share/keyrings/postgresql.gpg] \
http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" \
        | sudo tee /etc/apt/sources.list.d/pgdg.list
fi

sudo apt update -y
sudo apt install -y postgresql-17

ok "PostgreSQL installed"

# ===============================
# Start service
# ===============================
ok "Starting PostgreSQL..."

sudo systemctl enable postgresql
sudo systemctl restart postgresql

ok "PostgreSQL is running"

# ===============================
# Create USER (safe)
# ===============================
ok "Creating user..."

sudo -u postgres psql <<EOF
DO \$\$
BEGIN
   IF NOT EXISTS (
      SELECT 1 FROM pg_roles WHERE rolname = '$POSTGRES_USER'
   ) THEN
      CREATE USER $POSTGRES_USER WITH PASSWORD '$POSTGRES_PASSWORD';
   END IF;
END
\$\$;
EOF

# ===============================
# Create DATABASE (must be outside DO)
# ===============================
ok "Creating database..."

if ! sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='$POSTGRES_DB'" | grep -q 1; then
    sudo -u postgres psql -c "CREATE DATABASE $POSTGRES_DB OWNER $POSTGRES_USER"
fi

# ===============================
# Grant privileges
# ===============================
ok "Granting privileges..."

sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $POSTGRES_DB TO $POSTGRES_USER;"

ok "Database and user configured"

# ===============================
# Final output
# ===============================
ok "PostgreSQL setup complete"

echo "--------------------------------------"
echo "Database : $POSTGRES_DB"
echo "User     : $POSTGRES_USER"
echo "Host     : localhost"
echo "Port     : 5432"
echo "--------------------------------------"

ok "PostgreSQL configured ✅"