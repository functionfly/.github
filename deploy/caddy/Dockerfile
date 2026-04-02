# FunctionFly Caddy Dockerfile with rate_limit plugin and auth site
# This Dockerfile builds Caddy with the necessary plugins for rate limiting
# and serves the static auth site at /var/www/auth

# Builder stage
FROM caddy:2.7-builder-alpine AS builder

# Build Caddy with rate_limit plugin
RUN xcaddy build \
    --with github.com/mholt/caddy-ratelimit

# Auth site build stage
FROM oven/bun:1-alpine AS auth-builder

# Copy and build auth site
COPY web/auth /src/auth
WORKDIR /src/auth
RUN bun install && bun run build

# Final stage
FROM caddy:2.7-alpine

# Copy the custom-built Caddy binary
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

# Copy appropriate Caddyfile based on environment
ARG CADDY_CONFIG=staging
COPY deploy/caddy/${CADDY_CONFIG}.Caddyfile /etc/caddy/Caddyfile

# Copy built auth site
COPY --from=auth-builder /src/auth/dist /var/www/auth

# Create data directory for Caddy (for certificates, etc.)
RUN mkdir -p /data/caddy

# Expose ports
EXPOSE 80 443

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Default command (inherited from base image)
