# FunctionFly SAR - Stateful Agent Runtime

A production-ready Rust-based runtime for executing AI agent workflows with event-driven architecture, WASM sandboxing, and enterprise-grade security.

## Features

- **Event-Driven Architecture**: NATS-based publish/subscribe for agent lifecycle events
- **WASM Sandboxing**: Secure execution of AI agent logic in isolated WASM environments
- **Graph-Based Execution**: DAG-based workflow engine with retry policies and topological ordering
- **Multi-Tier Memory**: Hot/Warm/Cold memory tiers with LRU eviction
- **Observability**: Prometheus metrics, OpenTelemetry tracing, cost attribution
- **Production Security**: API key auth, rate limiting, TLS support, input validation

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        SAR Runtime                                │
├─────────────────────────────────────────────────────────────────┤
│  HTTP API (Axum)  │  gRPC API (Tonic)  │  Metrics (Prometheus) │
├───────────────────┼────────────────────┼────────────────────────┤
│  Agent Registry   │  Lifecycle Manager  │  Scheduler              │
├───────────────────┴────────────────────┴────────────────────────┤
│                    WASM Sandbox Engine                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Cell Pool   │  │ Memory     │  │ Fuel Meter  │             │
│  │             │  │ Limiter    │  │             │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
├─────────────────────────────────────────────────────────────────┤
│  NATS Event Bus  │  Persistence (PostgreSQL/SQLite)            │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Rust 1.70+
- NATS server (optional, for event-driven features)
- PostgreSQL 15+ (optional, for persistence)

### Build

```bash
cargo build --release
```

### Run

```bash
# With environment variables
export SAR_API_KEY="your-secure-api-key"
export NATS_URL="nats://localhost:4222"
export DATABASE_URL="postgres://user:pass@localhost/sar"

./target/release/functionfly-sar --api-port 8082

# Or use environment file
cp .env.example .env
# Edit .env with your configuration
source .env && ./target/release/functionfly-sar
```

### Docker

```bash
docker build -t functionfly-sar:latest -f Dockerfile .
docker run -p 8082:8082 \
  -e SAR_API_KEY="your-key" \
  -e NATS_URL="nats://host.docker.internal:4222" \
  -e DATABASE_URL="postgres://user:pass@host.docker.internal/sar" \
  functionfly-sar:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NATS_URL` | NATS server URL | `nats://localhost:4222` |
| `DATABASE_URL` | PostgreSQL connection string | In-memory store |
| `REDIS_URL` | Redis URL for caching | None |
| `SAR_API_KEY` | API key for authentication | None (dev mode) |
| `SAR_ADMIN_API_KEY` | Admin API key | None |
| `SAR_REQUIRE_AUTH` | Require authentication | `false` |
| `SAR_TLS_CERT` | TLS certificate path | None |
| `SAR_TLS_KEY` | TLS key path | None |
| `SAR_NATS_TLS_ENABLED` | Enable TLS for NATS | `false` |

### Command Line Options

```bash
functionfly-sar --help
```

```
FunctionFly Stateful Agent Runtime - Production Ready

Usage: functionfly-sar [OPTIONS]

Options:
  -p, --api-port <PORT>       API port [default: 8082]
  -m, --max-concurrent <N>     Max concurrent agents [default: 10000]
  -r, --rate-limit-rps <RPS>  Rate limit RPS [default: 100]
  -b, --rate-limit-burst <N>  Rate limit burst [default: 20]
  --require-auth              Require API key authentication
  -h, --help                  Show help
  -V, --version               Show version
```

## API Reference

### Health Check

```bash
curl http://localhost:8082/health
```

Response:
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "uptime_seconds": 3600
}
```

### Register Agent

```bash
curl -X POST http://localhost:8082/api/agents \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "name": "my-agent",
    "priority": 2,
    "max_concurrent_cells": 100,
    "isolation_enabled": true,
    "metadata": {
      "team": "ai-research"
    }
  }'
```

### List Agents

```bash
curl http://localhost:8082/api/agents \
  -H "X-API-Key: your-api-key"
```

### Unregister Agent

```bash
curl -X DELETE http://localhost:8082/api/agents/{agent_id} \
  -H "X-API-Key: your-api-key"
```

### Agent Heartbeat

```bash
curl -X POST http://localhost:8082/api/agents/{agent_id}/heartbeat \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "status": "running",
    "state_snapshot": {
      "current_task": "processing",
      "progress": "50%"
    }
  }'
```

### Graceful Shutdown

```bash
curl -X POST http://localhost:8082/api/agents/{agent_id}/shutdown \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"grace_period_seconds": 60}'
```

### Lifecycle Stats

```bash
curl http://localhost:8082/api/lifecycle/stats \
  -H "X-API-Key: your-api-key"
```

## Security

### Authentication

API keys are configured via environment variables:
- `SAR_API_KEY`: Standard API key
- `SAR_ADMIN_API_KEY`: Admin-level API key (all permissions)

Health and metrics endpoints are publicly accessible.

### Rate Limiting

Token bucket algorithm with configurable RPS and burst:
- Global rate limit: 100 RPS / 20 burst (configurable)
- Per-agent rate limiting: 100 RPS / 20 burst

### Input Validation

- Agent name: max 256 characters
- Metadata: max 64 entries, 1024 bytes per value
- Graph: max 1000 nodes, 5000 edges
- Input: max 100 entries, 64KB per value, 1MB total

### WASM Sandboxing

- Memory limit: configurable per cell (default 64MB)
- Fuel limit: configurable (default 1M operations)
- Compute timeout: configurable (default 5s)
- Allowed env vars: PATH, HOME, PWD, TMPDIR (configurable)

## Monitoring

### Prometheus Metrics

```
curl http://localhost:8082/metrics
```

Exposed metrics:
- `sar_agents_total` - Total registered agents
- `sar_agents_running` - Currently running agents
- `sar_executions_completed` - Total completed executions
- `sar_executions_failed` - Total failed executions
- `sar_queue_depth` - Current scheduler queue depth

## Development

### Build

```bash
cargo build --all-features
```

### Test

```bash
cargo test --all-features --lib
```

### Lint

```bash
cargo fmt
cargo clippy --all-targets --all-features -- -D warnings
```

### Feature Flags

| Feature | Description |
|---------|-------------|
| `nats-events` | Enable NATS event bus |
| `wasm-sandbox` | Enable WASM sandbox execution |
| `multi-memory` | Enable multi-tier memory (Redis + LRU) |

## Production Checklist

- [ ] Set `SAR_API_KEY` or `SAR_ADMIN_API_KEY`
- [ ] Enable `SAR_REQUIRE_AUTH` in production
- [ ] Configure TLS with `SAR_TLS_CERT` and `SAR_TLS_KEY`
- [ ] Set up PostgreSQL with `DATABASE_URL`
- [ ] Configure NATS with TLS for production
- [ ] Set appropriate rate limits
- [ ] Configure monitoring/alerting
- [ ] Review and adjust WASM sandbox limits
- [ ] Set up log aggregation

## License

MIT