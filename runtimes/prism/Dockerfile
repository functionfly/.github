# syntax=docker/dockerfile:1.6
# FunctionFly Prism Runtime — production image.
# Multi-stage build with non-root runtime user.

FROM rust:1.96.1-bookworm AS builder

WORKDIR /build

# Pre-cache dependencies by copying manifests first. This layer is only
# invalidated when Cargo.toml/Cargo.lock changes, not on source edits.
COPY runtimes/prism/Cargo.toml runtimes/prism/Cargo.lock* ./
RUN mkdir -p src && \
    echo "fn main() {}" > src/main.rs && \
    echo "" > src/lib.rs && \
    cargo build --release --bin prism && \
    rm -rf src target/release/deps/functionfly_prism* target/release/prism

COPY runtimes/prism/ ./

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/build/target \
    cargo build --release --bin prism

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
    && rm -rf /var/lib/apt/lists/*

# Prism runs as a non-root user. UID 10009 chosen to avoid collisions
# with system users and the other runtime UIDs (10001-10008).
RUN groupadd -r -g 10009 prismruntime && \
    useradd -r -u 10009 -g prismruntime -m -d /home/prismruntime -s /usr/sbin/nologin prismruntime

COPY --from=builder /build/target/release/prism /usr/local/bin/prism

# Healthcheck wrapper: curls /health on the bound port. The runtime binds
# to 127.0.0.1 by default so this works inside the container without
# exposing the port externally. The CMD override below defaults to loopback
# but operators can rebind to 0.0.0.0 if they put prism behind a reverse
# proxy in a network-isolated subnet.
RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-8080}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER prismruntime
EXPOSE 8080 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

# Use tini as PID 1 so SIGTERM is properly forwarded to the runtime and
# zombie children are reaped.
ENTRYPOINT ["/usr/bin/tini", "--"]

# Default args: bind to loopback (127.0.0.1) by default — secure-by-default.
# Override with --host 0.0.0.0 if you put prism behind a reverse proxy in a
# network-isolated subnet.
# The runtime also requires NATS_URL (defaults to nats://localhost:4222) and
# optionally RUNTIME_API_TOKEN (required if ENVIRONMENT=production).
CMD ["prism", "serve", "--host", "127.0.0.1", "--port", "8080"]

LABEL org.opencontainers.image.title="functionfly-prism-runtime" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
