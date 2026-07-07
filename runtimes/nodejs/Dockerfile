# syntax=docker/dockerfile:1.6
# FunctionFly Node.js Runtime — production image.

FROM rust:1.96.1-bookworm AS builder
WORKDIR /build
RUN apt-get update && apt-get install -y --no-install-recommends musl-tools && rm -rf /var/lib/apt/lists/*
COPY runtimes/nodejs/ ./
RUN rustup target add x86_64-unknown-linux-musl
RUN cargo build --release --target x86_64-unknown-linux-musl --bin functionfly-nodejs

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates curl tini
RUN addgroup -g 10006 -S nodejsruntime && adduser -u 10006 -S nodejsruntime -G nodejsruntime
COPY --from=builder /build/target/x86_64-unknown-linux-musl/release/functionfly-nodejs /usr/local/bin/functionfly-nodejs
RUN printf "#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:9091/health" > /usr/local/bin/healthcheck && chmod +x /usr/local/bin/healthcheck
USER nodejsruntime
EXPOSE 9091
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/usr/local/bin/healthcheck"]
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/functionfly-nodejs", "--daemon", "--port", "9091"]
