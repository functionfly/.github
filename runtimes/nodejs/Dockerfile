# syntax=docker/dockerfile:1.6
# FunctionFly Node.js Runtime — production image.
# The Node.js runtime executes JS via QuickJS embedded in the binary, so the
# runtime image does NOT need Node.js installed. However, the daemon can
# shell out to node tooling for tooling operations, so we keep a minimal
# Node.js LTS available.

FROM rust:1.96.1-bookworm AS builder

WORKDIR /build

COPY runtimes/nodejs/ ./

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    cargo build --release --bin functionfly-nodejs

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r -g 10006 nodejsruntime && \
    useradd  -r -u 10006 -g nodejsruntime -m -d /home/nodejsruntime -s /usr/sbin/nologin nodejsruntime

COPY --from=builder /build/target/release/functionfly-nodejs /usr/local/bin/functionfly-nodejs

# Healthcheck wrapper: curls /health on the bound port. The runtime binds
# to 127.0.0.1 by default so this works inside the container without
# exposing the port externally.
RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-9091}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER nodejsruntime
EXPOSE 9091

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

ENTRYPOINT ["/usr/bin/tini", "--"]
# Daemon mode listens on --port (default 9091). Add `--network` if you want
# host functions to make outbound requests.
CMD ["/usr/local/bin/functionfly-nodejs", "--daemon", "--port", "9091"]

LABEL org.opencontainers.image.title="functionfly-nodejs-runtime" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
