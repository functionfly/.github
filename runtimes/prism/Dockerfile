# FunctionFly Prism Runtime

FROM rust:1.96.1-bookworm AS builder
WORKDIR /build
RUN apt-get update && apt-get install -y --no-install-recommends musl-tools clang protobuf-compiler libprotobuf-dev libseccomp-dev && rm -rf /var/lib/apt/lists/*
COPY runtimes/prism/ ./
RUN rustup target add x86_64-unknown-linux-musl
RUN PROTOC=/usr/bin/protoc cargo build --release --target x86_64-unknown-linux-musl --bin prism

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates curl tini
RUN addgroup -g 10007 -S prismruntime && adduser -u 10007 -S prismruntime -G prismruntime
COPY --from=builder /build/target/x86_64-unknown-linux-musl/release/prism /usr/local/bin/prism
RUN printf "#!/bin/sh\nset -e\ncurl -fsS http://127.0.0.1:8084/health" > /usr/local/bin/healthcheck && chmod +x /usr/local/bin/healthcheck
USER prismruntime
EXPOSE 8084
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/usr/local/bin/healthcheck"]
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["prism", "start", "--address", "127.0.0.1:8084"]
