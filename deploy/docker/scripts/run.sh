#!/bin/bash
set -e

# Load environment variables from .env if it exists
if [ -f /.env ]; then
    export $(cat /.env | grep -v '^#' | xargs)
fi

# -----------------------------
# Helper function to start a service in background
# -----------------------------
start_service() {
    local service_name=$1
    shift
    echo "[INFO] Starting $service_name..."
    "$@" &
    eval "${service_name}_PID=\$!"
    echo "[INFO] $service_name started with PID ${!service_name_PID}"
}

# -----------------------------
# Stop all services on exit
# -----------------------------
trap 'echo "[INFO] Stopping all services..."; \
      kill -TERM $POSTGRES_PID $OCSERV_PID $ADMIN_API_PID $CUSTOMER_API_PID $LOG_PARSER_PID $USER_MANAGER_PID $TELEGRAM_BOT_PID $NGINX_PID 2>/dev/null || true; \
      wait' SIGTERM SIGINT

# -----------------------------
# 1. Start Postgres
# -----------------------------
echo "[INFO] Starting PostgreSQL..."
su - postgres -c "/usr/lib/postgresql/17/bin/pg_ctl -D /var/lib/postgresql/17/main -l /var/log/postgresql/postgresql-17-main.log start" &
POSTGRES_PID=$!

# Wait for Postgres to be ready
echo "[INFO] Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if su - postgres -c "pg_isready -q"; then
        echo "[INFO] PostgreSQL is ready!"
        break
    fi
    sleep 1
    if [ $i -eq 30 ]; then
        echo "[ERROR] PostgreSQL failed to start in time!"
        exit 1
    fi
done

# -----------------------------
# 2. Initialize Ocserv (runs the entrypoint script)
# -----------------------------
echo "[INFO] Initializing Ocserv..."
/usr/local/bin/ocserv_entrypoint.sh

# -----------------------------
# 3. Start Admin Dashboard API
# -----------------------------
start_service "ADMIN_API" /usr/local/bin/admin_api migrate
sleep 2
start_service "ADMIN_API" /usr/local/bin/admin_api serve ${ADMIN_API_DEBUG:+ -d}

# -----------------------------
# 4. Start Customer Dashboard API
# -----------------------------
start_service "CUSTOMER_API" /usr/local/bin/customer_api serve ${CUSTOMER_API_DEBUG:+ -d}

# -----------------------------
# 5. Start Ocserv Log Parser
# -----------------------------
start_service "LOG_PARSER" /usr/local/bin/ocserv_log_parser serve ${LOG_PARSER_DEBUG:+ -d}

# -----------------------------
# 6. Start Ocserv User Manager
# -----------------------------
start_service "USER_MANAGER" /usr/local/bin/ocserv_user_manager serve ${USER_MANAGER_DEBUG:+ -d}

# -----------------------------
# 7. Start Telegram Bot (if enabled)
# -----------------------------
if [ "$TELEGRAM_BOT_ENABLED" = "true" ]; then
    start_service "TELEGRAM_BOT" /usr/local/bin/ocserv_telegram_bot serve ${TELEGRAM_BOT_DEBUG:+ -d}
fi

# -----------------------------
# 8. Start Ocserv
# -----------------------------
start_service "OCSERV" /usr/local/sbin/ocserv --foreground --debug=3 --config=/etc/ocserv/ocserv.conf

# -----------------------------
# 9. Start Nginx
# -----------------------------
echo "[INFO] Starting Nginx..."
nginx &
NGINX_PID=$!

# -----------------------------
# Wait for any service to exit
# -----------------------------
echo "[INFO] All services started successfully"
wait -n

# -----------------------------
# Cleanup
# -----------------------------
echo "[INFO] One of the services exited. Stopping all services..."
kill -TERM $POSTGRES_PID $OCSERV_PID $ADMIN_API_PID $CUSTOMER_API_PID $LOG_PARSER_PID $USER_MANAGER_PID $TELEGRAM_BOT_PID $NGINX_PID 2>/dev/null || true

# Stop Postgres gracefully
su - postgres -c "/usr/lib/postgresql/17/bin/pg_ctl -D /var/lib/postgresql/17/main stop" 2>/dev/null || true

echo "[INFO] All services stopped"
