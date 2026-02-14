# FunctionFly

A virtual edge layer for routing requests to the best available customer-provided edge backend.

## Overview

FunctionFly is a control plane that routes incoming traffic to edge computing targets across multiple providers (Cloudflare Workers, Vercel, Fly.io, Deno Deploy) with health-aware selection, latency-informed scoring, and circuit breaker protection.

## Architecture

- **Control Plane**: Go services for configuration, routing decisions, and health monitoring
- **Data Plane**: Customer-deployed edge targets that implement a uniform HTTP contract
- **Edge Entry**: Caddy reverse proxy for TLS termination and initial routing

## Quick Start

1. **Start the database**:
   ```bash
   make docker-up
   ```

2. **Run the API server**:
   ```bash
   make dev
   ```

3. **Test the API**:
   ```bash
   curl http://localhost:8080/health
   ```

## Development

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Commands

- `make build` - Build all services
- `make test` - Run tests
- `make dev` - Start development environment
- `make docker-up` - Start Docker services
- `make docker-down` - Stop Docker services

### Project Structure

```
├── cmd/                    # Application entry points
│   ├── orchestrator-api/   # Main API service
│   └── health-monitor/     # Health monitoring service
├── internal/               # Private application code
│   ├── api/                # HTTP API handlers
│   ├── routing/            # Routing logic
│   ├── storage/            # Database layer
│   └── auth/               # Authentication
├── migrations/             # Database migrations
├── deploy/                 # Deployment configurations
└── plans/                  # Project specifications
```

## MVP1 Scope

FunctionFly MVP1 includes:

- ✅ JWT authentication for dashboard
- ✅ App-scoped API keys for programmatic access
- ✅ Health-aware backend selection
- ✅ Latency-informed routing with EWMA scoring
- ✅ Circuit breaker pattern (closed/open/half-open)
- ✅ Fast failover for idempotent methods
- ⏳ Provider adapters (Cloudflare Workers, Vercel, Fly.io, Deno Deploy)
- ⏳ Basic dashboard UI
- ⏳ HMAC request signing
- ⏳ Rate limiting

## License

MIT