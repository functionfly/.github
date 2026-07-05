# syntax=docker/dockerfile:1.6
# FunctionFly Deno Runtime — production image.

FROM rust:1.96.1-bookworm AS builder

WORKDIR /build

COPY runtimes/deno/ ./

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    cargo build --release --bin functionfly-deno-runtime

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r -g 10002 denoruntime && \
    useradd  -r -u 10002 -g denoruntime -m -d /home/denoruntime -s /usr/sbin/nologin denoruntime

COPY --from=builder /build/target/release/functionfly-deno-runtime /usr/local/bin/functionfly-deno-runtime

RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-8090}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER denoruntime
EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/functionfly-deno-runtime", "--port", "8090"]

LABEL org.opencontainers.image.title="functionfly-deno-runtime" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
