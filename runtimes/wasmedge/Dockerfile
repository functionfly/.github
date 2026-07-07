# syntax=docker/dockerfile:1.6
# FunctionFly WasmEdge Runtime — production image.

FROM rust:1.96.1-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gnupg \
        wget \
    && wget -qO - https://apt.wasmedge.org/wasmedge.pub | apt-key add - \
    && echo "deb [arch=amd64] https://apt.wasmedge.org bookworm latest" > /etc/apt/sources.list.d/wasmedge.list \
    && apt-get update \
    && apt-get install -y wasmedge_0.14.0 wasmedge-dev_0.14.0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY runtimes/wasmedge/ ./

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    cargo build --release --bin functionfly-wasmedge-runtime

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
        gnupg \
    && curl -fsSL https://github.com/WasmEdge/WasmEdge/releases/download/0.14.0/WasmEdge-0.14.0.debian_bookworm_amd64.deb -o /tmp/wasmedge.deb \
    && dpkg -i /tmp/wasmedge.deb \
    && rm /tmp/wasmedge.deb \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r -g 10005 wasmedgeruntime && \
    useradd  -r -u 10005 -g wasmedgeruntime -m -d /home/wasmedgeruntime -s /usr/sbin/nologin wasmedgeruntime

COPY --from=builder /build/target/release/functionfly-wasmedge-runtime /usr/local/bin/functionfly-wasmedge-runtime

RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-8093}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER wasmedgeruntime
EXPOSE 8093

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/functionfly-wasmedge-runtime", "--port", "8093"]

LABEL org.opencontainers.image.title="functionfly-wasmedge-runtime" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
