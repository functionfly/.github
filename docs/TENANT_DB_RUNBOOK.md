# Tenant Database Isolation - Operational Runbook

## Overview

This runbook covers the operational procedures for managing per-tenant dedicated databases in FunctionFly's SaaS Starter Pack bundle.

## Architecture

### Components

1. **TenantDBProvisioner** (`internal/storage/tenant_db.go`)
   - Creates/deletes/clones tenant databases
   - Manages per-tenant connection pools
   - Runs tenant schema migrations

2. **TenantPoolManager** (`internal/storage/tenant_pool.go`)
   - Manages connection pools for all tenant databases
   - Health monitoring and auto-isolation
   - Prometheus metrics for observability

3. **TenantDBRegistry** (`internal/storage/tenant_registry.go`)
   - Credential storage with encryption
   - Connection string building
   - Local cache with TTL

4. **TenantDBHealthChecker** (`internal/storage/tenant_health.go`)
   - Background health monitoring (30s intervals)
   - Auto-isolation on repeated failures
   - Health reports and alerting

5. **TenantDBService** (`internal/services/tenant_db_service.go`)
   - Orchestration layer for provisioning/deprovisioning
   - Handles plan changes (upgrades/downgrades)
   - Idempotency guards for concurrent requests

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TENANT_DB_ENABLED` | Enable per-tenant databases | `false` |
| `TENANT_DB_HOST` | PostgreSQL host | `localhost` |
| `TENANT_DB_PORT` | PostgreSQL port | `5432` |
| `TENANT_DB_USER` | PostgreSQL user (must be superuser for CREATE DATABASE) | `postgres` |
| `TENANT_DB_PASSWORD` | PostgreSQL password | (none) |
| `TENANT_DB_TEMPLATE` | Template database name | `functionfly_tenant_template` |
| `TENANT_DB_PREFIX` | Prefix for tenant DB names | `functionfly_tenant_` |
| `TENANT_DB_POOL_MIN` | Min connections per pool | `2` |
| `TENANT_DB_POOL_MAX` | Max connections per pool | `10` |
| `TENANT_DB_MAX_IDLE_TIME` | Max idle time before pool closure | `5m` |
| `TENANT_DB_CONN_MAX_LIFETIME` | Max connection lifetime | `1h` |
| `TENANT_DB_USE_TEMPLATE` | Use template DB for cloning | `false` |
| `TENANT_DB_ENCRYPTION_KEY_ID` | Encryption key reference | `tenant-db-key` |
| `TENANT_DB_RETRY_ATTEMPTS` | Retry attempts for operations | `3` |
| `TENANT_DB_RETRY_DELAY` | Delay between retries | `2s` |

## Provisioning Flow

### Automatic Provisioning (Stripe Webhook)

1. Customer purchases SaaS Starter, Marketplace, or AI App bundle
2. `checkout.session.completed` webhook triggers
3. `provisionBundleResources()` is called in billing handler
4. `provisionDedicatedTenantDB()` provisions the database:
   - Checks if tenant plan qualifies (starter/professional/enterprise)
   - Calls `TenantDBService.ProvisionForTenant()`
   - Database is created via template clone or empty DB
   - Migrations are applied
   - Credentials are stored in registry

### Manual Provisioning (Admin API)

```
POST /admin/tenants/{tenantId}/dedicated-db/provision
```

Requires `PermTenantsWrite` permission and HMAC signature.

## Database Schema

### Tenant Base Schema

Each tenant database gets a copy of `internal/storage/sql/tenant_migrations/20260501142000_tenant_base_schema.up.sql`:

- `tenant_users` - Isolated user accounts
- `tenant_configs` - Tenant settings and feature flags
- `tenant_api_keys` - Programmatic access keys
- `tenant_audit_log` - Compliance audit trail
- `tenant_state_entries` - Isolated state storage
- `tenant_platform_refs` - Cross-DB reference to platform
- `tenant_webhooks` - Event notifications

## Health Monitoring

### Health Check Intervals

- **Pool health checks**: Every 30 seconds
- **Cache cleanup**: Every 1 minute
- **Idle pool cleanup**: Every 5 minutes

### Health Status Types

| Status | Description |
|--------|-------------|
| `healthy` | All checks passing |
| `degraded` | 1-2 consecutive failures |
| `unhealthy` | 3+ consecutive failures |
| `isolated` | Auto-isolated due to failures |
| `suspended` | Manually suspended |
| `closed` | Pool closed, not active |

### Auto-Isolation

When `max_failures` (3) is reached:
1. Pool is closed (no new connections)
2. Status set to `isolated` in registry
3. Alert logged for monitoring

## Admin API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/tenants/{tenantId}/dedicated-db/status` | Get DB status and health |
| POST | `/admin/tenants/{tenantId}/dedicated-db/provision` | Manually provision DB |
| POST | `/admin/tenants/{tenantId}/dedicated-db/suspend` | Suspend DB |
| POST | `/admin/tenants/{tenantId}/dedicated-db/resume` | Resume DB |
| DELETE | `/admin/tenants/{tenantId}/dedicated-db` | Delete DB |
| GET | `/admin/dedicated-dbs` | List all tenant DBs (summary) |

## Troubleshooting

### Database Creation Fails

**Symptom**: `provisionDedicatedTenantDB` returns error

**Causes**:
1. PostgreSQL user lacks `CREATEDB` privilege
2. Template database doesn't exist
3. Connection to PostgreSQL fails

**Resolution**:
1. Verify PostgreSQL user has `CREATEDB`: `psql -c "SHOW grantee"`
2. Run bootstrap script: `./scripts/bootstrap-tenant-template.sh`
3. Check connection: `psql -h $TENANT_DB_HOST -p $TENANT_DB_PORT -U $TENANT_DB_USER -d postgres`

### Connection Pool Exhausted

**Symptom**: `pgx pool timeout` errors

**Causes**:
1. Too many connections to tenant DB
2. Long-running queries blocking connections
3. Network issues

**Resolution**:
1. Check pool stats: `GET /admin/dedicated-dbs`
2. Increase `TENANT_DB_POOL_MAX` (default 10)
3. Check for idle connections: PostgreSQL `pg_stat_activity`

### Encryption/Decryption Failures

**Symptom**: `failed to decrypt: cipher: message authentication failed`

**Causes**:
1. Password was encrypted with different key
2. Database corruption
3. Legacy "encrypted:" prefix vs `ENC:` prefix mismatch

**Resolution**:
1. Check `tenant_database_configs` table for password format
2. Verify `db_master_key_password` env var is consistent
3. Re-provision if needed (data loss warning)

### Tenant DB Suspended Unexpectedly

**Symptom**: Tenant reports access issues

**Causes**:
1. Health checker detected failures and auto-isolated
2. Admin manually suspended
3. Payment failure triggered suspension

**Resolution**:
1. Check health status: `GET /admin/tenants/{tenantId}/dedicated-db/status`
2. Check tenant suspension reason in billing logs
3. Resume if appropriate: `POST /admin/tenants/{tenantId}/dedicated-db/resume`

## Prometheus Metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `functionfly_tenant_pool_connections` | `tenant_id`, `pool_status` | Connection count per pool |
| `functionfly_tenant_pool_errors_total` | `tenant_id`, `error_type` | Pool errors |
| `functionfly_tenant_pool_health_check_duration_seconds` | `tenant_id`, `status` | Health check latency |

## Template Database Bootstrap

### Initial Setup

```bash
# Run as superuser (postgres)
sudo -u postgres ./scripts/bootstrap-tenant-template.sh
```

### Recreate Template (WARNING: affects all future clones)

```bash
sudo -u postgres ./scripts/bootstrap-tenant-template.sh --recreate
```

### Verify Template

```sql
-- Connect to template and verify
psql -h localhost -U postgres -d functionfly_tenant_template -c "\dt"
```

Expected tables:
- tenant_users
- tenant_configs
- tenant_api_keys
- tenant_audit_log
- tenant_state_entries
- tenant_platform_refs
- tenant_webhooks

## Graceful Shutdown

When the orchestrator API shuts down:

1. **TenantDBProvisioner.Close()** is called
2. All tenant pools are closed (with 5s timeout each)
3. Template pool is closed
4. Waits up to 10s for all pools to close

To ensure clean shutdown, the API should:
1. Call `TenantDBService.Stop()` to stop health checker
2. Call `TenantDBProvisioner.Close()` to close all pools
3. Call `TenantPoolManager.Close()` for pool manager

## Rollback Procedures

### If Provisioning Fails Mid-Way

1. Check `tenant_database_configs` for partial state
2. Delete any partially created database: `DROP DATABASE IF EXISTS functionfly_tenant_<id>`
3. Clean up registry: `DELETE FROM tenant_database_configs WHERE tenant_id = '<id>'`
4. Retry provisioning

### If Migration Fails

1. Check `tenant_db_migrations` table for failure reason
2. Fix migration file if needed
3. Retry: tenant DB will pick up from last successful migration

## Performance Tuning

### For High-Tenant Environments

```bash
# Increase pool size for high-traffic tenants
TENANT_DB_POOL_MAX=20

# Reduce idle timeout to free connections faster
TENANT_DB_MAX_IDLE_TIME=2m

# Enable template DB cloning (faster provisioning)
TENANT_DB_USE_TEMPLATE=true
```

### For Low-Traffic Environments

```bash
# Reduce pool size to save resources
TENANT_DB_POOL_MIN=1
TENANT_DB_POOL_MAX=5

# Increase idle timeout to avoid connection churn
TENANT_DB_MAX_IDLE_TIME=10m
```

## Security Considerations

1. **Encryption**: Passwords are encrypted with AES-256-GCM using scrypt-derived key
2. **Zero-Knowledge**: Server cannot decrypt tenant data without proper key
3. **Audit Logging**: All operations are logged in `tenant_audit_log`
4. **Network Isolation**: Tenant DBs should be in isolated network segment
5. **Access Control**: Only platform DB user should have access to tenant DBs

## Alerting Recommendations

| Alert | Condition | Severity |
|-------|-----------|----------|
| `TenantDBUnhealthy` | Any tenant DB in unhealthy state > 5min | Warning |
| `TenantDBIsolated` | Any tenant DB isolated | Critical |
| `TenantPoolExhausted` | Pool at max connections for > 5min | Warning |
| `ProvisioningFailures` | > 3 provisioning failures in 10min | Warning |