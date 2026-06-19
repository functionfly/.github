# Production Readiness Checklist

This document tracks production readiness status for FunctionFly deployment.

## Last Updated: 2026-06-18

---

## 1. Database Migrations

### Status: ⚠️ Known Issues

**Validation Script:** `./scripts/validate-migrations.sh`
- **Result:** FAIL - Duplicate migration versions detected
- **Root Cause:** Historical mixed naming conventions (6-digit sequential + 14-digit timestamp)
- **Impact:** golang-migrate may fail on duplicate sequence numbers

**Files with Duplicates (partial list):**
- `20251101000015_security_audit_tables` / `20251101000218_security_audit_tables`
- `20251101000016_create_sessions_table` / `20251101000219_create_sessions_table`
- `20251101000005_add_tenant_status` / `20251101000209_add_tenant_status`
- `20251101000006_complete_billing_system` / `20251101000210_complete_billing_system`
- Many more (see `migrations/` directory)

**Workaround (already in use):**
- Production deployments use `--skip-migrations` flag
- Schema applied via direct SQL or DBA process
- See [MIGRATIONS.md](MIGRATIONS.md) for migration policy

### Required Actions:
- [ ] Validate migrations on fresh database before deployment
- [ ] Consider running `scripts/fix-duplicate-migrations.sh` to renumber orphan migrations
- [ ] Document exact schema version/commit deployed to production
- [ ] Use timestamp format (YYYYMMDDHHMMSS) for all new migrations

---

## 2. Required Environment Variables

### Status: ✅ Documented

**Reference:** [`.env.example`](../.env.example), [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md)

### Critical Variables (MUST be set):
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=functionfly
DB_USER=postgres
DB_PASSWORD=<secure-password>
DB_SSLMODE=require  # or 'disable' for local

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=<secure-password>

# Security (generate with: openssl rand -hex 32)
JWT_SECRET=<32+ char secret>
API_SHARED_SECRET=<32+ char secret>
PRODUCTION_ENV=true

# Application
PORT=8080
LOG_LEVEL=info
```

### Optional but Recommended:
```bash
# Stripe (required for billing)
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Email (Resend recommended)
RESEND_API_KEY=re_...
FROM_EMAIL=noreply@yourdomain.com

# Storage
STORAGE_BACKEND=s3  # or 'r2' for Cloudflare R2
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
R2_ACCOUNT_ID=...  # for R2
```

### Secrets Management:
- **Canonical store:** Infisical (see [INFISICAL_SETUP.md](INFISICAL_SETUP.md))
- **Local development:** `.env` file (gitignored)
- **Fly.io:** `scripts/sync-infisical-to-fly.sh`

---

## 3. Code Health

### Status: ✅ Fixed (2026-06-18)

**Build Command:** `go build ./cmd/... ./internal/...`
**Result:** All packages build successfully

### Issues Fixed (2026-06-18):
1. **circuitbreaker duplicate declaration**: Removed duplicate `ErrCircuitOpen` in `external_services.go`
2. **type mismatch**: Changed `*logrus.Logger` to `*logrus.Entry` in `ExternalCircuitBreakerManager`
3. **wrong function call**: Changed `NewPrometheusMetrics(cfg)` to `NewWithMetrics(cfg, nil)`
4. **wasmpool stub adapters removed**: Removed incomplete adapter files that referenced non-existent wasmpool interfaces
5. **scripts missing main()**: Added `main()` function to `scripts/gen_dre_node_key.go`
6. **wasmpool load test TenantId**: Fixed field name to `TenantID` in `tests/load/wasm-pool/main.go`
7. **SQL injection filter**: Added URL decoding to `Detect()` method to handle encoded payloads
8. **auth middleware tests**: Updated weak JWT secrets to use strong 32-byte secrets

### Test Status (2026-06-18):
- `internal/api/middleware` - ✅ All tests passing
- `internal/api/middleware/advanced_security` - ✅ All tests passing
- Remaining test failures are pre-existing issues in test files (wrong arguments to repo methods)

### Required Actions:
- [x] Fix build errors - **DONE**
- [ ] Run `go mod tidy` to ensure dependencies are correct (optional)
- [ ] Apply database schema before running integration tests
- [ ] Set up proper WASM runtime environment for WASM tests

---

## 4. External Dependencies

### Status: ✅ All Configured

| Service | Default Port | Docker Compose | Notes |
|---------|-------------|---------------|-------|
| PostgreSQL | 5432 | `docker-compose.production.yml` | PG 17+ required |
| Redis | 6379 | `docker-compose.production.yml` | Redis 7+ |
| Caddy | 8083/443 | `docker-compose.production.yml` | TLS termination |
| Prometheus | 9090 | `docker-compose.monitoring.yml` | Optional |
| Grafana | 3000 | `docker-compose.monitoring.yml` | Optional |
| Loki | 3100 | `docker-compose.monitoring.yml` | Optional |

### Required for Full Functionality:
- [ ] PostgreSQL 17+ with pgvector extension
- [ ] Redis 7+ for sessions/caching
- [ ] SMTP server (Resend recommended) for transactional email

---

## 5. Deployment Checklist

### Pre-Deployment:
- [ ] **Database:** Ensure PostgreSQL 17+ is running with correct credentials
- [ ] **Redis:** Ensure Redis 7+ is running with correct credentials
- [ ] **Schema:** Apply schema via direct SQL (do NOT rely on `--skip-migrations`)
- [ ] **Secrets:** Load secrets from Infisical or `.env.production`
- [ ] **Build:** Run `make build` successfully
- [ ] **Tests:** Run `make test-fast` with >90% passing
- [ ] **Migrations:** Run `./scripts/validate-migrations.sh` (expect warnings)

### Deployment Steps:
```bash
# 1. Validate environment
./scripts/test-production-readiness.sh

# 2. Build
make build

# 3. Apply database schema (choose one)
# Option A: Direct SQL
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/YYYYMMDDHHMMSS_initial_schema.up.sql

# Option B: If migrations are validated
./bin/orchestrator-api migrate

# 4. Start services
docker-compose -f docker-compose.production.yml up -d

# 5. Verify
curl https://yourdomain.com/health
curl https://yourdomain.com/api/v1/health
```

### Post-Deployment:
- [ ] Verify `/health` endpoint returns 200
- [ ] Verify `/api/v1/health` returns 200
- [ ] Check logs for errors: `docker-compose logs -f`
- [ ] Verify Prometheus metrics accessible
- [ ] Test basic auth flow (login/logout)
- [ ] Test Stripe webhook endpoint

---

## 6. Monitoring Setup

### Status: ✅ Available

**Reference:** [MONITORING.md](MONITORING.md), [docker-compose.monitoring.yml](../docker-compose.monitoring.yml)

### Key Metrics to Monitor:
```promql
# API Health
http_requests_total
http_request_duration_seconds_bucket

# Database
database_connections_active
database_connections_idle

# Redis
redis_connected_clients
redis_memory_used_bytes

# Business
active_subscriptions_total
registry_functions_total
```

### Alerting Recommendations:
- [ ] Set up alerts for `http_request_duration_seconds > 2s`
- [ ] Set up alerts for `database_connections_active > 80%`
- [ ] Set up alerts for `redis_memory_used_bytes > 80%`
- [ ] Set up alerts for failed payment webhooks

---

## 7. Backup & Recovery

### Status: ✅ Documented

**Reference:** [DISASTER_RECOVERY_RUNBOOK.md](DISASTER_RECOVERY_RUNBOOK.md), [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md)

### Required Backups:
- [ ] Daily pg_dump of PostgreSQL
- [ ] Test restore procedure quarterly
- [ ] Redis RDB persistence enabled
- [ ] R2/ S3 backup of function artifacts

### RTO/ RPO Targets:
- **RTO:** 4 hours (time to restore service)
- **RPO:** 24 hours (maximum data loss)

---

## 8. Security Hardening

### Status: ⚠️ Partial

**Reference:** [SECURITY.md](SECURITY.md), [ADMIN_CLOUDFLARE_SECURITY.md](ADMIN_CLOUDFLARE_SECURITY.md)

### Production Must-Haves:
- [ ] `PRODUCTION_ENV=true` set
- [ ] `JWT_SECRET` minimum 32 characters
- [ ] `CORS_ALLOWED_ORIGINS` set to specific domains (not `*`)
- [ ] `STRICT_IP_VALIDATION=true`
- [ ] SSL/ TLS via Caddy or load balancer
- [ ] Database SSL enabled (`DB_SSLMODE=require`)
- [ ] Redis password set
- [ ] Cloudflare WAF rules configured
- [ ] Rate limiting enabled

### Optional but Recommended:
- [ ] WebAuthn/ FIDO2 for admin accounts
- [ ] SAML SSO for enterprise
- [ ] IP allowlisting for admin routes
- [ ] Audit logging to external SIEM

---

## 9. Go Module Health

### Status: ⚠️ Needs Verification

**Command:** `go mod verify && go mod tidy`

### Current State:
- Some indirectly required modules need explicit import
- node_modules has incorrect version specifiers (non-blocking)

### Required Actions:
- [ ] Run `go mod tidy` and commit any changes
- [ ] Verify `go.mod` / `go.sum` are in sync
- [ ] Consider running `go mod verify` in CI

---

## Summary

| Category | Status | Notes |
|----------|--------|-------|
| Database Migrations | ⚠️ Known Issues | Use --skip-migrations, apply schema directly |
| Environment Variables | ✅ Complete | See .env.example |
| Code Health | ✅ Fixed | Build errors resolved (2026-06-18) |
| External Dependencies | ✅ Configured | Docker Compose available |
| Deployment Process | ✅ Documented | See PRODUCTION_DEPLOYMENT.md |
| Monitoring | ✅ Available | docker-compose.monitoring.yml |
| Backup & Recovery | ✅ Documented | See DISASTER_RECOVERY_RUNBOOK.md |
| Security | ⚠️ Partial | Enable PRODUCTION_ENV=true |

### Critical Path to Production:
1. ~~Fix `internal/storage/config.go` build errors~~ - **RESOLVED**
2. ~~Verify all packages build with `go build ./...`~~ - **RESOLVED**
3. Validate migrations on fresh database
4. Configure all required environment variables
5. Deploy with `--skip-migrations` or validated schema
6. Run `./scripts/test-production-readiness.sh`

---

## References

- [PRODUCTION_DEPLOYMENT.md](PRODUCTION_DEPLOYMENT.md) - Full deployment guide
- [DISASTER_RECOVERY_RUNBOOK.md](DISASTER_RECOVERY_RUNBOOK.md) - Recovery procedures
- [MIGRATIONS.md](MIGRATIONS.md) - Migration policy
- [INFISICAL_SETUP.md](INFISICAL_SETUP.md) - Secrets management
- [SECURITY.md](SECURITY.md) - Security guidelines
