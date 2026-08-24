# OCServ Dashboard Backend

This repository provides one backend application that runs the Admin API, Customer API, Worker, and optional Telegram Bot. The Docker images also run OCServ and PostgreSQL 18 in the same container.

## Runtime sequence

The Docker container starts services in this order:

```text
prepare OCServ configuration and networking
→ initialize and start PostgreSQL 18
→ wait for PostgreSQL readiness
→ run database migrations
→ start the unified backend
→ start OCServ
→ supervise all critical processes
```

If PostgreSQL, OCServ, or the backend exits unexpectedly, the remaining processes are stopped. SIGINT and SIGTERM trigger graceful shutdown.

## Requirements

- Linux host with Docker and Docker Compose
- `/dev/net/tun` available on the host
- Permission to add the `NET_ADMIN` capability
- Permission to read `/var/run/docker.sock`
- Ports `443/tcp`, `443/udp`, and `8080/tcp` available

Create persistent host directories:

```bash
sudo mkdir -p \
  /opt/ocserv_dashboard/docker_volumes/pg_db \
  /opt/ocserv_dashboard/docker_volumes/ocserv \
  /opt/ocserv_dashboard/docker_volumes/cron_journal \
  /opt/ocserv_dashboard/docker_volumes/telegram_receipts
```

The container must be named `ocserv`. The Worker reads OCServ logs through the Docker API using that container name.

## Production Docker deployment

Copy and configure the environment file:

```bash
cp .env.sample .env
```

Replace at least these values:

```env
SECRET_KEY="generate-a-random-value"
JWT_SECRET="generate-a-different-random-value"
POSTGRES_PASSWORD="set-a-strong-database-password"
```

Random secrets can be generated with:

```bash
openssl rand -hex 32
```

Use the following Compose configuration with the production `Dockerfile`:

```yaml
services:
  ocserv:
    container_name: ocserv
    image: ocserv-dashboard:latest
    build:
      context: .
      dockerfile: Dockerfile
      args:
        GO_VERSION: ${GO_VERSION:-1.26.0}
    env_file:
      - ./.env
    environment:
      HOST_PROC: /host/proc
      HOST_SYS: /host/sys
      POSTGRES_HOST: 127.0.0.1
      TELEGRAM_RECEIPTS_DIR: /opt/ocserv_dashboard/uploads/receipts
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun:/dev/net/tun
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /opt/ocserv_dashboard/docker_volumes/ocserv:/etc/ocserv
      - /opt/ocserv_dashboard/docker_volumes/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts
      - /opt/ocserv_dashboard/docker_volumes/pg_db:/var/lib/postgresql
      - /opt/ocserv_dashboard/docker_volumes/cron_journal:/app/cron_journal
    ports:
      - "${OCSERV_PORT:-443}:${OCSERV_PORT:-443}/tcp"
      - "${OCSERV_PORT:-443}:${OCSERV_PORT:-443}/udp"
      - "8080:8080"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://127.0.0.1:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
```

Save it as `compose.production.yml`, then run:

```bash
sudo docker compose -f compose.production.yml up --build -d
```

Check startup and migration logs:

```bash
sudo docker compose -f compose.production.yml logs -f ocserv
```

### Standalone production container

Build the image without Compose:

```bash
sudo docker build \
  --build-arg GO_VERSION=1.26.0 \
  -f Dockerfile \
  -t ocserv-dashboard:latest \
  .
```

Run the complete production backend stack:

```bash
sudo docker run -d \
  --name ocserv \
  --restart unless-stopped \
  --env-file ./.env \
  --env HOST_PROC=/host/proc \
  --env HOST_SYS=/host/sys \
  --env POSTGRES_HOST=127.0.0.1 \
  --env TELEGRAM_RECEIPTS_DIR=/opt/ocserv_dashboard/uploads/receipts \
  --cap-add NET_ADMIN \
  --device /dev/net/tun:/dev/net/tun \
  --volume /var/run/docker.sock:/var/run/docker.sock:ro \
  --volume /proc:/host/proc:ro \
  --volume /sys:/host/sys:ro \
  --volume /opt/ocserv_dashboard/docker_volumes/ocserv:/etc/ocserv \
  --volume /opt/ocserv_dashboard/docker_volumes/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts \
  --volume /opt/ocserv_dashboard/docker_volumes/pg_db:/var/lib/postgresql \
  --volume /opt/ocserv_dashboard/docker_volumes/cron_journal:/app/cron_journal \
  --publish 443:443/tcp \
  --publish 443:443/udp \
  --publish 8080:8080 \
  ocserv-dashboard:latest
```

Follow its logs with:

```bash
sudo docker logs -f ocserv
```

## UI development Docker deployment

`Docker-Dev` contains backend services only. It enables backend debug mode, permits local UI origins on ports `3000` and `5173`, disables Telegram by default, and supplies development-only database credentials.

Use this Compose configuration:

```yaml
services:
  ocserv:
    container_name: ocserv
    image: ocserv-dashboard-dev:latest
    build:
      context: .
      dockerfile: Docker-Dev
      args:
        GO_VERSION: ${GO_VERSION:-1.26.0}
    environment:
      HOST_PROC: /host/proc
      HOST_SYS: /host/sys
      POSTGRES_HOST: 127.0.0.1
      TELEGRAM_RECEIPTS_DIR: /opt/ocserv_dashboard/uploads/receipts
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun:/dev/net/tun
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /opt/ocserv_dashboard/docker_volumes/ocserv:/etc/ocserv
      - /opt/ocserv_dashboard/docker_volumes/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts
      - /opt/ocserv_dashboard/docker_volumes/pg_db:/var/lib/postgresql
      - /opt/ocserv_dashboard/docker_volumes/cron_journal:/app/cron_journal
    ports:
      - "443:443/tcp"
      - "443:443/udp"
      - "8080:8080"
      - "127.0.0.1:5435:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://127.0.0.1:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
```

Save it as `compose.development.yml`, then run:

```bash
sudo docker compose -f compose.development.yml up --build
```

The UI can access the API at:

```text
http://localhost:8080
```

The development PostgreSQL server is available locally at `127.0.0.1:5435`.

### Standalone development container

Build the UI-development backend image without Compose:

```bash
sudo docker build \
  --build-arg GO_VERSION=1.26.0 \
  -f Docker-Dev \
  -t ocserv-dashboard-dev:latest \
  .
```

Run all development backend services:

```bash
sudo docker run -d \
  --name ocserv \
  --restart unless-stopped \
  --env HOST_PROC=/host/proc \
  --env HOST_SYS=/host/sys \
  --env POSTGRES_HOST=127.0.0.1 \
  --env TELEGRAM_RECEIPTS_DIR=/opt/ocserv_dashboard/uploads/receipts \
  --cap-add NET_ADMIN \
  --device /dev/net/tun:/dev/net/tun \
  --volume /var/run/docker.sock:/var/run/docker.sock:ro \
  --volume /proc:/host/proc:ro \
  --volume /sys:/host/sys:ro \
  --volume /opt/ocserv_dashboard/docker_volumes/ocserv:/etc/ocserv \
  --volume /opt/ocserv_dashboard/docker_volumes/telegram_receipts:/opt/ocserv_dashboard/uploads/receipts \
  --volume /opt/ocserv_dashboard/docker_volumes/pg_db:/var/lib/postgresql \
  --volume /opt/ocserv_dashboard/docker_volumes/cron_journal:/app/cron_journal \
  --publish 443:443/tcp \
  --publish 443:443/udp \
  --publish 8080:8080 \
  --publish 127.0.0.1:5435:5432 \
  ocserv-dashboard-dev:latest
```

The development and production examples use the same host data directories and container name. Do not run them simultaneously. Change the development host paths if isolated data is required.

## Telegram Bot

Telegram settings and bot accounts are stored through the dashboard. Enable the in-process Telegram service with:

```env
TELEGRAM_BOT_ENABLED=true
```

Receipts are persisted at:

```text
/opt/ocserv_dashboard/docker_volumes/telegram_receipts
```

No separate Telegram executable or container is required.

## Native systemd installation

For a native Debian or Ubuntu deployment:

```bash
cp .env.sample .env
sudo ./scripts/systemd/install.sh
```

The installer can provision local PostgreSQL when `INSTALL_POSTGRES=true`. Set it to `false` and configure the PostgreSQL connection variables when using an existing server.

Check native service status:

```bash
systemctl status ocserv ocserv-dashboard
journalctl -fu ocserv-dashboard
```

## Common operations

Health check:

```bash
curl -fsS http://127.0.0.1:8080/health
```

Stop the production stack:

```bash
docker compose -f compose.production.yml down
```

Rebuild after backend changes:

```bash
docker compose -f compose.development.yml up --build
```

Reset development data only when it is safe to permanently remove the selected host directories. PostgreSQL 18 data must be mounted at `/var/lib/postgresql`; `/var/lib/postgresql/data` is the pre-18 layout and must not be used for this image.

## Troubleshooting

If OCServ cannot initialize networking, confirm the TUN device and capability:

```bash
ls -l /dev/net/tun
docker inspect ocserv --format '{{json .HostConfig.CapAdd}}'
```

If the Worker stops because it cannot read OCServ logs, confirm:

- the container is named `ocserv`;
- `/var/run/docker.sock` is mounted;
- the Docker daemon socket is readable by the container process.

If startup stops before the backend is launched, inspect the container logs. PostgreSQL readiness and migration failures are reported before OCServ and the backend are started.
