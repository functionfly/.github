# syntax=docker/dockerfile:1.6
# FunctionFly MicroVM Orchestrator — production image.
# The orchestrator drives Firecracker, which needs /dev/kvm and CAP_NET_ADMIN.
#
# SECURITY REQUIREMENTS:
#   - Run as privileged container OR with --device=/dev/kvm --privileged
#   - Dedicated single-tenant host (not shared multi-tenant environments)
#   - Kubernetes: DaemonSet with KVM passthrough + securityContext.privileged=true
#   - Required capabilities: CAP_NET_ADMIN, CAP_SYS_ADMIN
#   - Required devices: /dev/kvm (character device 10:232)
#   - Required kernel features: KVM support (check /dev/kvm exists)
#
# Example docker run:
#   docker run --device=/dev/kvm --privileged \
#     --cap-add=SYS_ADMIN --cap-add=NET_ADMIN \
#     -v /var/lib/functionfly:/var/lib/functionfly \
#     functionfly-microvm

FROM rust:1.96.1-bookworm AS builder

WORKDIR /build

COPY runtimes/microvm/Cargo.toml runtimes/microvm/Cargo.lock* ./
RUN mkdir -p src && \
    echo "fn main() {}" > src/main.rs && \
    echo "" > src/lib.rs && \
    cargo build --release --bin functionfly-microvm && \
    rm -rf src target/release/deps/functionfly_microvm* target/release/functionfly-microvm

COPY runtimes/microvm/ ./
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/build/target \
    cargo build --release --bin functionfly-microvm

FROM debian:bookworm-slim AS runtime

# Firecracker v1.15.1 — pinned. The orchestrator drives the Firecracker
# HTTP API over a Unix socket.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
        iproute2 \
    && curl -fsSL https://github.com/firecracker-microvm/firecracker/releases/download/v1.15.1/firecracker-v1.15.1-x86_64.tgz \
        | tar -xz -C /tmp \
    && mv /tmp/release-v1.15.1-x86_64/firecracker-v1.15.1-x86_64 /usr/local/bin/firecracker \
    && chmod +x /usr/local/bin/firecracker \
    && rm -rf /tmp/release-v1.15.1-x86_64 \
    && rm -rf /var/lib/apt/lists/*

# The microvm orchestrator creates tap devices and runs Firecracker, which
# needs CAP_NET_ADMIN. It also writes to /var/lib/functionfly/vmimages and
# /var/run/firecracker. We do NOT mark this image non-root because KVM access
# requires the privileged CAP_SETUID/CAP_SETGID — instead the orchestrator
# is launched as a dedicated low-privilege user with only the caps it needs.
RUN groupadd -r -g 10007 microvm && \
    useradd  -r -u 10007 -g microvm -m -d /home/microvm -s /usr/sbin/nologin microvm

RUN mkdir -p /var/lib/functionfly/vmimages /var/run/firecracker && \
    chown -R microvm:microvm /var/lib/functionfly /var/run/firecracker

COPY --from=builder /build/target/release/functionfly-microvm /usr/local/bin/functionfly-microvm

# Healthcheck wrapper: curls /health on the bound port. Microvm binds to
# 0.0.0.0 by default so we explicitly hit loopback inside the container.
RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-9091}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER microvm
EXPOSE 9091

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/functionfly-microvm", \
     "--port", "9091", \
     "--vcpus", "2", \
     "--memory-mb", "512", \
     "--max-vms", "100", \
     "--warm-idle-secs", "600", \
     "--cleanup-interval-secs", "60"]

LABEL org.opencontainers.image.title="functionfly-microvm-orchestrator" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
