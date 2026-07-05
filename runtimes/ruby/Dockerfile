# syntax=docker/dockerfile:1.6
# FunctionFly Ruby Runtime — production image.

FROM rust:1.96.1-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        ruby-dev \
        libruby \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY runtimes/ruby/ ./

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    cargo build --release --bin functionfly-ruby

FROM ruby:3.3-slim-bookworm AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        tini \
        libruby \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r -g 10004 rubyruntime && \
    useradd  -r -u 10004 -g rubyruntime -m -d /home/rubyruntime -s /usr/sbin/nologin rubyruntime

COPY --from=builder /build/target/release/functionfly-ruby /usr/local/bin/functionfly-ruby

RUN printf '#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:${HEALTH_PORT:-8094}/health\n' \
        > /usr/local/bin/healthcheck && \
    chmod +x /usr/local/bin/healthcheck

USER rubyruntime
EXPOSE 8094

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/functionfly-ruby", "--port", "8094"]

LABEL org.opencontainers.image.title="functionfly-ruby-runtime" \
      org.opencontainers.image.source="https://github.com/functionfly/functionfly" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.vendor="FunctionFly"
