# FunctionFly Production Deployment Guide

This guide covers deploying the **FunctionFly** platform (orchestrator API, dashboard, data stores, and optional observability stack) with Docker Compose on your own infrastructure.

For Fly.io + Neon + Cloudflare Pages-style deployment, see [FLY_DEPLOYMENT.md](FLY_DEPLOYMENT.md) and [DOMAIN_AND_COMING_SOON_SETUP.md](DOMAIN_AND_COMING_SOON_SETUP.md).

## Secrets (single workflow)

- **Canonical edits:** [Infisical](https://infisical.com) per environment (`dev` / `staging` / `prod`). See [INFISICAL_SETUP.md](INFISICAL_SETUP.md).
- **Fly.io runtime:** sync with [`scripts/sync-infisical-to-fly.sh`](../scripts/sync-infisical-to-fly.sh) after changes, or use the helpers in [`.fly/README.md`](../.fly/README.md).
- **Local laptop:** gitignored `.env` from [`.env.example`](../.env.example), or `infisical run` / `make dev` — do not commit secrets.

## Prerequisites

- Docker and Docker Compose installed
- PostgreSQL 17+ database
- Redis 7+ cache
- Domain name with SSL certificate
- Environment variables configured

## Architecture

The production deployment consists of:

1. **Orchestrator API** - Main API server (Go)
2. **Dashboard** - React frontend
3. **PostgreSQL** - Primary database
4. **Redis** - Cache and session store
5. **Caddy** - Reverse proxy and TLS termination
6. **Prometheus** - Metrics collection
7. **Grafana** - Metrics visualization
8. **Loki** - Log aggregation
9. **Promtail** - Log collection

## Environment Configuration

1. Copy the production environment template:

```bash
cp .env.production .env
```

1. Configure the following required variables:

```bash
# Database
DB_PASSWORD=your_secure_password
DB_HOST=postgres
DB_PORT=5432
DB_NAME=functionfly
DB_USER=functionfly

# Redis
REDIS_PASSWORD=your_secure_password
REDIS_HOST=redis
REDIS_PORT=6379

# Security
JWT_SECRET=your_jwt_secret_min_32_chars
API_SHARED_SECRET=your_api_secret_min_32_chars

# Domain
DOMAIN=yourdomain.com
BASE_URL=https://yourdomain.com

# Email (optional)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_email@example.com
SMTP_PASSWORD=your_email_password
```

## Deployment Steps

### 1. Build and Start Services

```bash
# Build all services
docker-compose -f docker-compose.production.yml build

# Start services in detached mode
docker-compose -f docker-compose.production.yml up -d
```

### 2. Initialize Database

**Schema / migrations:** The repository’s `migrations/` tree has historically contained duplicate sequence numbers, which breaks tooling such as golang-migrate in some setups. Many environments run the API with `--skip-migrations` during development; **production** must still apply a known-good schema.

Choose one approach and document it for your team:

1. **Apply SQL from scratch** using your DBA process (Neon/console, `psql`, or a single curated migration bundle you maintain).
2. **Use orchestrator migrate** only if you have validated that your migration set applies cleanly end-to-end on an empty database (recommended: rehearse on a staging branch first).

```bash
# If your build supports it and migrations are validated:
docker-compose -f docker-compose.production.yml exec orchestrator-api ./bin/orchestrator-api migrate

# Seed initial data (optional)
docker-compose -f docker-compose.production.yml exec orchestrator-api ./bin/orchestrator-api seed
```

Do **not** assume `migrate` works without verification—confirm against a disposable database before touching production.

### 3. Verify Deployment

```bash
# Check service status
docker-compose -f docker-compose.production.yml ps

# Check logs
docker-compose -f docker-compose.production.yml logs -f

# Test API health
curl https://yourdomain.com/health
```

## Service URLs

- **API**: <https://yourdomain.com/api>
- **Dashboard**: <https://yourdomain.com>
- **Grafana**: <https://yourdomain.com:3000>
- **Prometheus**: <https://yourdomain.com:9090>

## Monitoring

### Grafana Dashboards

Access Grafana at <https://yourdomain.com:3000> with credentials:

- Username: admin
- Password: (set via GRAFANA_ADMIN_PASSWORD environment variable)

Pre-configured dashboards:

- FunctionFly API Metrics
- Database Performance
- Redis Metrics
- System Resources

### Prometheus Metrics

Access Prometheus at <https://yourdomain.com:9090>

Key metrics to monitor:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `database_connections_active` - Active database connections
- `redis_connected_clients` - Redis connections

### Log Aggregation

Logs are collected by Promtail and stored in Loki.

Access logs via Grafana:

1. Go to Grafana
2. Click "Explore"
3. Select "Loki" datasource
4. Query: `{container="functionfly-orchestrator"}`

## Backup and Recovery

This repo has two common Postgres layouts:

| Compose file | DB user / name (defaults) |
|--------------|-----------------------------|
| [`docker-compose.production.yml`](../docker-compose.production.yml) (repo root) | `functionfly` / `functionfly` (override with `DB_USER`, `DB_NAME` in `.env`) |
| [`deploy/production/docker-compose.yml`](../deploy/production/docker-compose.yml) | `functionfly_prod` / `functionfly_prod` |

Use the **same** user and database name in `pg_dump` / `psql` as your running stack. The script [`scripts/backup-database.sh`](../scripts/backup-database.sh) is written for the **`deploy/production`** stack (`functionfly_prod`); if you use the root `docker-compose.production.yml`, prefer a manual `pg_dump` (below) or align the script with your `DB_*` values.

### Database backup (manual, plain SQL)

From the repo root, with `.env` loaded (so `DB_USER` / `DB_NAME` / `DB_PASSWORD` match compose):

```bash
set -a && [ -f .env ] && . ./.env && set +a
mkdir -p ./backups
STAMP=$(date +%Y%m%d_%H%M%S)
docker-compose -f docker-compose.production.yml exec -T postgres \
  pg_dump -U "${DB_USER:-functionfly}" "${DB_NAME:-functionfly}" \
  | gzip > "./backups/functionfly_${STAMP}.sql.gz"
gunzip -t "./backups/functionfly_${STAMP}.sql.gz" && echo "OK: gzip integrity"
```

### Database restore (replace existing DB — destructive)

Only when you intend to overwrite the live database with a backup file:

```bash
gunzip -c ./backups/functionfly_YYYYMMDD_HHMMSS.sql.gz | \
  docker-compose -f docker-compose.production.yml exec -T postgres \
  psql -U "${DB_USER:-functionfly}" -d "${DB_NAME:-functionfly}"
```

### Testing backup restoration (recommended procedure)

**Always test against a separate database first**, not production.

1. **Create a backup** using the manual `pg_dump` flow above (or your cron job output under `./backups/`).

2. **Create an empty database** for the test (name is arbitrary; example uses `functionfly_restore_test`):

   ```bash
   docker-compose -f docker-compose.production.yml exec postgres \
     psql -U "${DB_USER:-functionfly}" -d postgres -v ON_ERROR_STOP=1 -c \
     "CREATE DATABASE functionfly_restore_test OWNER \"${DB_USER:-functionfly}\";"
   ```

3. **Restore the backup into that database** (pipe plain SQL; no `pg_restore` unless you used `pg_dump -Fc`):

   ```bash
   gunzip -c ./backups/functionfly_YYYYMMDD_HHMMSS.sql.gz | \
     docker-compose -f docker-compose.production.yml exec -T postgres \
     psql -U "${DB_USER:-functionfly}" -d functionfly_restore_test -v ON_ERROR_STOP=1
   ```

4. **Verify schema and data**:

   ```bash
   docker-compose -f docker-compose.production.yml exec postgres \
     psql -U "${DB_USER:-functionfly}" -d functionfly_restore_test -c \
     "SELECT COUNT(*) AS users FROM users;"
   docker-compose -f docker-compose.production.yml exec postgres \
     psql -U "${DB_USER:-functionfly}" -d functionfly_restore_test -c \
     "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 3;"
   ```

5. **Optional:** point a **staging** orchestrator at `functionfly_restore_test` (or a clone host) and smoke-test login and one critical API path.

6. **Drop the test database** when finished:

   ```bash
   docker-compose -f docker-compose.production.yml exec postgres \
     psql -U "${DB_USER:-functionfly}" -d postgres -c \
     "DROP DATABASE IF EXISTS functionfly_restore_test;"
   ```

7. **Record RTO/RPO:** note wall-clock time from “backup file in hand” to “queries pass on restored DB” (RTO) and how old the backup was (RPO).

#### Schedule

- Run the full test **before launch** and after any change to backup path, retention, or compose DB settings.
- Repeat **quarterly** (or per your compliance policy), after major schema changes, and after infrastructure moves.

### Neon (managed Postgres)

Use the **Neon CLI** for connection strings; do **not** paste passwords into shell history when you can avoid it—prefer `neon auth` or `NEON_API_KEY`.

#### Logical backup (Neon CLI + `pg_dump`)

From the repo root (requires `neon`, `pg_dump`, and `gzip`; install CLI with `make neon-install`):

```bash
make backup-neon
# same as:
# NEON_BRANCH=production NEON_DATABASE_NAME=functionfly ./scripts/backup-neon.sh
```

Optional: `NEON_PROJECT_ID=<uuid>` if context is not set (`neon set-context` / [NEON.md](NEON.md)).

This writes `./backups/functionfly_YYYYMMDD_HHMMSS.sql.gz` and verifies gzip integrity. The CLI returns a **direct** (non-pooled) URL, which is appropriate for `pg_dump` ([NEON.md](NEON.md)).

One-liner equivalent:

```bash
mkdir -p ./backups
STAMP=$(date +%Y%m%d_%H%M%S)
pg_dump "$(neon connection-string production --database-name functionfly)" \
  | gzip > "./backups/functionfly_${STAMP}.sql.gz"
gunzip -t "./backups/functionfly_${STAMP}.sql.gz"
```

#### Restore drill (logical dump)

Restore only into a **non-production** database that is **empty** (or you have dropped objects first). Practical options:

1. **Local Postgres (simplest):** create an empty database locally, then:

   ```bash
   gunzip -c ./backups/functionfly_YYYYMMDD_HHMMSS.sql.gz | \
     psql -h localhost -U postgres -d functionfly_restore_test
   ```

2. **Neon CLI for the target URL:** use `neon connection-string <branch> --database-name functionfly` as the `psql` target when the database is empty (e.g. new empty DB created in the Neon console on a dev branch).

**Native Neon drill (no `pg_dump` file):** to validate Neon’s storage without a logical restore, create a throwaway branch: `neon branches create --name backup-drill --parent production`, run sanity queries via `neon connection-string backup-drill --database-name functionfly`, then `neon branches delete backup-drill`. That exercises branching / PITR-style workflows; see [NEON.md](NEON.md).

#### Other Neon tools

- **Point-in-time / parent restore:** `neon branches restore` — see `neon branches restore --help` and [DISASTER_RECOVERY_RUNBOOK.md](DISASTER_RECOVERY_RUNBOOK.md).

See also [NEON.md](NEON.md).

## Security Considerations

1. **SSL/TLS**: Caddy automatically provisions SSL certificates via Let's Encrypt
2. **Firewall**: Only expose ports 80, 443, and 3000 (Grafana)
3. **Secrets**: Use strong passwords and secrets (minimum 32 characters)
4. **Updates**: Regularly update Docker images and dependencies
5. **Backups**: Maintain regular database backups

## Scaling

### Horizontal Scaling

```bash
# Scale orchestrator API
docker-compose -f docker-compose.production.yml up -d --scale orchestrator-api=3

# Scale dashboard
docker-compose -f docker-compose.production.yml up -d --scale dashboard=3
```

### Database Scaling

For high availability, consider:

- PostgreSQL replication (primary + replicas)
- Connection pooling (PgBouncer)
- Read replicas for analytics

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose -f docker-compose.production.yml logs <service-name>

# Check resource usage
docker stats
```

### Database Connection Issues

```bash
# Test database connectivity
docker-compose -f docker-compose.production.yml exec postgres pg_isready -U functionfly

# Check database logs
docker-compose -f docker-compose.production.yml logs postgres
```

### High Memory Usage

```bash
# Check memory usage
docker stats --no-stream

# Restart services
docker-compose -f docker-compose.production.yml restart
```

## Maintenance

### Regular Tasks

1. **Daily**: Check logs for errors
2. **Weekly**: Review metrics and performance
3. **Monthly**: Update dependencies and security patches
4. **Quarterly**: Review and rotate secrets

### Updates

```bash
# Pull latest images
docker-compose -f docker-compose.production.yml pull

# Restart with new images
docker-compose -f docker-compose.production.yml up -d
```

## Support

For issues and questions:

- GitHub Issues: <https://github.com/functionfly/functionfly/issues>
- Documentation: <https://docs.functionfly.com>
- Community: <https://discord.gg/functionfly>
