# syntax=docker/dockerfile:1.6
# FunctionFly Bun Runtime — production image.

FROM rust:1.96.1-bookworm AS builder

WORKDIR /build

COPY runtimes/bun/ ./

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    cargo build --release --bin functionfly-bun-runtime

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r -g 10001 bunruntime && \
    useradd  -r -u 10001 -g bunruntime -m -d /home/bunruntime -s /usr/sbin/nologin bunruntime

COPY --from=builder /build/target/release/functionfly-bun-runtime /usr/local/bin/functionfly-bun-runtime

RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-8091}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER bunruntime
EXPOSE 8091

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/functionfly-bun-runtime", "--port", "8091"]

LABEL org.opencontainers.image.title="functionfly-bun-runtime" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
