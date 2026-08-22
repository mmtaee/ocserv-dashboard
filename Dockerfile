# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.0

FROM golang:${GO_VERSION}-bookworm AS backend-builder

ENV CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN go build -trimpath -ldflags="-s -w" -o /out/backend ./main.go

FROM debian:trixie-slim

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gnutls-bin \
        iproute2 \
        iptables \
        ocserv \
        procps \
        tini \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=backend-builder /out/backend /usr/local/bin/backend
COPY --chmod=755 docker/entrypoint.sh /usr/local/bin/container-entrypoint
COPY --chmod=755 docker/server.sh /usr/local/bin/container-server

ENV OCSERV_PORT=443 \
    OC_NET=172.16.24.0/24 \
    OCSERV_DNS=1.1.1.1 \
    SSL_CN=ocserv-dashboard \
    SSL_ORG=ocserv-dashboard \
    SSL_EXPIRE=3650 \
    OCSERV_PRESERVE_CONFIG=false \
    RUN_MIGRATIONS=true \
    MIGRATION_MAX_ATTEMPTS=30 \
    MIGRATION_RETRY_SECONDS=2 \
    TELEGRAM_RECEIPTS_DIR=/app/uploads/receipts

RUN mkdir -p \
        /app/cron_journal \
        /app/uploads/receipts \
        /etc/ocserv/defaults \
        /etc/ocserv/groups \
        /etc/ocserv/users

EXPOSE 443/tcp 443/udp 8080/tcp

VOLUME ["/etc/ocserv", "/app/cron_journal", "/app/uploads/receipts"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD curl --fail --silent http://127.0.0.1:8080/health >/dev/null || exit 1

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/container-entrypoint"]
CMD ["/usr/local/bin/container-server"]
