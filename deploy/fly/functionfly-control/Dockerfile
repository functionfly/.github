# Multi-stage build for minimal image
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the orchestrator binary with CGO disabled for Alpine
RUN CGO_ENABLED=0 GOOS=linux GOFLAGS=-toolchain=none go build -o /orchestrator -ldflags="-s -w" ./cmd/orchestrator-api

# Final stage - minimal runtime image
FROM alpine:3.18

# Install CA certificates and necessary runtime packages
RUN apk --no-cache add ca-certificates tzdata curl postgresql-client

WORKDIR /app

# Copy binary from builder
COPY --from=builder /orchestrator .

# Create non-root user for security
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser && \
    chown -R appuser:appgroup /app

USER appuser

# Expose the API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

# Run the orchestrator (migrations are run separately via make migrate-neon against Neon)
CMD ["./orchestrator", "--skip-migrations"]
