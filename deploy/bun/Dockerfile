FROM oven/bun:1-alpine

WORKDIR /app

# Copy compiled Go runtime executor
COPY bin/runtime-bun /usr/local/bin/
RUN chmod +x /usr/local/bin/runtime-bun

# Copy function execution runtime
COPY deploy/bun/runtime/ /app/runtime/

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run the Bun runtime
CMD ["runtime-bun"]
