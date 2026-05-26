FROM debian:trixie-slim

ARG DEBIAN_MIRROR
ARG DEBIAN_SECURITY_MIRROR

RUN cat /etc/apt/sources.list.d/debian.sources
RUN echo "---"

RUN if [ -n "${DEBIAN_MIRROR}" ] || [ -n "${DEBIAN_SECURITY_MIRROR}" ]; then \
        debian_mirror="${DEBIAN_MIRROR:-http://deb.debian.org/debian}" && \
        debian_security_mirror="${DEBIAN_SECURITY_MIRROR:-http://deb.debian.org/debian-security}" && \
        echo "deb ${debian_mirror} trixie main contrib non-free non-free-firmware" > /etc/apt/sources.list && \
        echo "deb ${debian_mirror} trixie-updates main contrib non-free non-free-firmware" >> /etc/apt/sources.list && \
        echo "deb ${debian_security_mirror} trixie-security main non-free-firmware" >> /etc/apt/sources.list \
    ; fi

RUN echo "---"
RUN cat /etc/apt/sources.list.d/debian.sources


# sudo docker build -f test-ocserv.Dockerfile .

# sudo docker build --build-arg DEBIAN_MIRROR=https://linux-mirror.example.com/repository/debian \
#    --build-arg DEBIAN_SECURITY_MIRROR=https://linux-mirror.example.com/repository/debian-security \
#    -f test-mirror.Dockerfile .
