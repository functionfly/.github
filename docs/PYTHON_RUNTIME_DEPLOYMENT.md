# Production Deployment

## Python Runtime Service

The Python Runtime Service (`python-runtime`) provides full CPython 3.13 WASI execution using wasmtime. It runs as a separate process to allow the main orchestrator to be built with `CGO_ENABLED=0`.

### Build

```bash
# Build python-runtime with CGO (requires wasmtime CLI)
CGO_ENABLED=1 go build -o python-runtime ./cmd/python-runtime

# Build orchestrator without CGO
CGO_ENABLED=0 go build -o orchestrator-api ./cmd/orchestrator-api
```

### Deployment Architecture

```
┌─────────────────┐     ┌──────────────────┐
│  orchestrator   │     │  python-runtime  │
│  (CGO_ENABLED=0)│────▶│  (CGO_ENABLED=1) │
│                 │     │                  │
│  Port 8080      │     │  Port 8083       │
└─────────────────┘     └──────────────────┘
```

### Configuration

Environment variables for `python-runtime`:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8083` | HTTP server port |
| `POOL_SIZE` | `4` | Number of pooled executors |
| `MAX_MEMORY_MB` | `512` | Max memory per execution (MB) |
| `CPYTHON_WASM_PATH` | `./runtimes/cpython-wasi/python.wasm` | Path to CPython WASM binary |
| `AUTH_TOKEN` | (none) | Token required for `/execute` endpoint |
| `PREWARM` | `false` | Prewarm pool on startup |

Environment variables for `orchestrator`:

| Variable | Default | Description |
|----------|---------|-------------|
| `PYTHON_RUNTIME_URL` | `http://localhost:8083` | Python runtime endpoint |
| `VERIFICATION_ENABLED` | `true` | Enable trust verification |
| `DATA_RETENTION_ENABLED` | `true` | Enable data retention policies |

### Resource Limits

The `python-runtime` enforces the following limits:

| Limit | Value | Description |
|-------|-------|-------------|
| Max code size | 1MB | Maximum Python code size |
| Max body size | 10MB | Maximum request body |
| Max timeout | 300s | Maximum execution timeout |
| Rate limit | 100/min | Per-IP rate limit |
| Connection timeout | 5s | HTTP connection timeout |

### Health Check

The `/health` endpoint performs actual execution testing:

```bash
curl http://localhost:8083/health
```

Returns `"status": "healthy"` only if a test execution succeeds.

### Metrics

Prometheus metrics available at `/metrics`:

- `python_runtime_execution_latency_ms` - Execution latency histogram
- `python_runtime_execution_errors_total` - Error counter
- `python_runtime_up` - Health gauge (1=up, 0=down)
- `python_runtime_request_duration_ms` - HTTP request duration

### Deployment Steps

1. **Build both binaries**:
   ```bash
   CGO_ENABLED=1 go build -o python-runtime ./cmd/python-runtime
   CGO_ENABLED=0 go build -o orchestrator-api ./cmd/orchestrator-api
   ```

2. **Deploy `python-runtime`** on same host as orchestrator:
   ```bash
   ./python-runtime --port 8083 --pool-size 4 --max-memory-mb 512
   ```

3. **Set environment for orchestrator**:
   ```bash
   export PYTHON_RUNTIME_URL=http://localhost:8083
   export AUTH_TOKEN=your_secure_token
   ```

4. **Deploy orchestrator**:
   ```bash
   ./orchestrator-api
   ```

### Security

- **Auth Token**: Set `AUTH_TOKEN` to protect `/execute` endpoint
- **Rate Limiting**: Per-IP limiting (100 requests/minute)
- **Input Validation**: Max code size 1MB, max body 10MB
- **Circuit Breaker**: Connection resilience with 5 failure threshold
- **Retry with Backoff**: 3 attempts with exponential backoff

### Monitoring

Check health:
```bash
curl http://localhost:8083/health | jq
```

Check pool stats:
```bash
curl http://localhost:8083/pool/stats | jq
```

View metrics:
```bash
curl http://localhost:8083/metrics | grep python_runtime
```