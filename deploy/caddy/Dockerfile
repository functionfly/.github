# FunctionFly Caddy Dockerfile with rate_limit plugin
# This Dockerfile builds Caddy with the necessary plugins for rate limiting

# Builder stage
FROM caddy:2.7-builder-alpine AS builder

# Build Caddy with rate_limit plugin
RUN xcaddy build \
    --with github.com/mholt/caddy-ratelimit

# Final stage
FROM caddy:2.7-alpine

# Copy the custom-built Caddy binary
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

# Copy appropriate Caddyfile based on environment
ARG CADDY_CONFIG=staging
COPY deploy/caddy/${CADDY_CONFIG}.Caddyfile /etc/caddy/Caddyfile

# Create data directory for Caddy (for certificates, etc.)
RUN mkdir -p /data/caddy

# Expose ports
EXPOSE 80 443

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Default command (inherited from base image)
