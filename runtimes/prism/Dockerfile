# Prism Runtime — multi-stage build
FROM rust:1.82-bookworm AS builder

WORKDIR /build
COPY runtimes/prism/ .
RUN cargo build --release

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r prism && useradd -r -g prism -m prism

COPY --from=builder /build/target/release/prism /usr/local/bin/prism

USER prism
EXPOSE 8084 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD [ "prism", "health" ]

ENTRYPOINT [ "prism" ]
CMD [ "serve", "--host", "0.0.0.0", "--port", "8084" ]
