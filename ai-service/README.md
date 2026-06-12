# FlyMind AI Service

Intelligent capabilities for FunctionFly: LLM providers (OpenAI, Anthropic, Ollama), embeddings, caching, and health checks.

## Setup

```bash
uv sync
```

Optional: self-hosted content moderation (Detoxify) for toxicity/hate/violence:

```bash
uv sync --extra moderation
```

gRPC server (optional): generate stubs using the project venv (do not use system `pip`/`python`):

```bash
uv sync
uv run python scripts/generate_grpc.py
# then start the app; gRPC listens on 0.0.0.0:50051 by default
```

## Run

From the `ai-service` directory (uses `ai-service/.env`; clear a conflicting repo-root `VIRTUAL_ENV` if `uv` warns):

```bash
unset VIRTUAL_ENV
uv sync
PYTHONPATH=. uv run uvicorn src.main:app --host 127.0.0.1 --port 18081
```

Or use the repo helper (from repo root):

```bash
./scripts/run-ai-service.sh
```

## Configuration

See `.env.example`. Key settings: `DATABASE_URL`, `REDIS_URL`, provider API keys.

## Production Checklist

Before deploying to production, verify the following:

### Security

- [ ] **API Keys**: All API keys are set via environment variables or secrets manager (e.g., Fly.io secrets, AWS Secrets Manager, HashiCorp Vault). Never commit real keys to `.env` files.
- [ ] **Database Credentials**: Use individual connection parameters (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`) instead of `DATABASE_URL` with embedded credentials.
- [ ] **TLS/SSL**:
  - [ ] HTTP service uses TLS (enforced by Fly.io with `force_https = true`)
  - [ ] gRPC uses TLS (`GRPC_USE_TLS=true`) with valid certificates
  - [ ] Database connection uses SSL (`DB_SSLMODE=require`)
- [ ] **CORS**: `CORS_ORIGINS` is set to your production frontend domain(s), NOT `["*"]`
- [ ] **Secrets Rotation**: Document secret rotation procedures for all API keys
- [ ] **gRPC Auth**: `GRPC_USE_AUTH=true` to enable API key authentication for gRPC endpoints
- [ ] **Security Headers**: `SecurityHeadersMiddleware` is active (verified by default)

### Infrastructure

- [ ] **Container Security**:
  - [ ] Dockerfile uses non-root user (`appuser`)
  - [ ] Read-only filesystem configured (`read_only: true` in docker-compose)
  - [ ] No new privileges (`security_opt: no-new-privileges`)
  - [ ] Capabilities dropped (`cap_drop: ALL`)
- [ ] **Resource Limits**: Memory and CPU limits set in `docker-compose.yml` and `fly.toml`
- [ ] **Redis Security**:
  - [ ] Redis bound to localhost only (`127.0.0.1:6379`)
  - [ ] Password protected (`REDIS_PASSWORD` set)
  - [ ] TLS for Redis (if using Upstash or managed Redis with TLS)
- [ ] **Network Isolation**: Services on internal network, only expose necessary ports

### Monitoring & Observability

- [ ] **Health Checks**: Health check endpoint `/health` returns degraded status when dependencies are unavailable
- [ ] **Metrics**: Prometheus metrics endpoint `/metrics` exposed
- [ ] **Tracing**: OpenTelemetry configured with OTLP endpoint (if using Atlas or other observability)
- [ ] **Error Tracking**: Sentry DSN configured (`SENTRY_DSN` environment variable)
- [ ] **Audit Logging**: Audit logs are being written to database

### Dependencies

- [ ] **Lock File**: `uv.lock` is up-to-date and committed
- [ ] **SBOM**: Software Bill of Materials generated and stored
- [ ] **CVE Scanning**: CI pipeline runs regular CVE scans (`pip-audit` or similar)
- [ ] **Dependency Updates**: Regular dependency update process in place

### Configuration

- [ ] **Environment**: `ENVIRONMENT=production` set
- [ ] `DEBUG=false` (never `true` in production)
- [ ] `LOG_LEVEL=INFO` or `WARNING` (not `DEBUG` in production)
- [ ] Rate limits set appropriate for your plan
- [ ] `REQUIRE_REDIS=true` to fail fast if Redis is unavailable
- [ ] Feature flags set appropriately for production load

### Deployment

- [ ] **Secrets**: All secrets set via `flyctl secrets set` or secrets manager
- [ ] **Database Migrations**: Migration scripts tested and idempotent
- [ ] **Rollback Plan**: Know how to rollback to previous version
- [ ] **Deploy Script**: `./deploy.sh` completes successfully
- [ ] **Health Check Passes**: `flyctl health check` passes after deployment

### gRPC (if using)

- [ ] TLS certificates valid and not expired
- [ ] Auth interceptor enabled (`GRPC_USE_AUTH=true`)
- [ ] Client connections use TLS and valid API keys
