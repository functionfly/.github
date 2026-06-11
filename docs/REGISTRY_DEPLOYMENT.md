# FunctionFly Registry Deployment Guide

This guide covers deploying the FunctionFly Registry to Fly.io with Neon Postgres, including configuration, monitoring, and maintenance.

## Overview

- **Platform**: [Fly.io](https://fly.io)
- **Database**: [Neon Postgres](https://neon.tech) (already in use)
- **Storage**: Cloudflare R2 (backups and artifacts)
- **Caching**: Upstash Redis
- **CDN**: Cloudflare (DNS)

## Prerequisites

### Required Tools

```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Verify installation
fly version
```

### Required Accounts

- Fly.io account with billing configured
- Neon Postgres project (already configured)
- Cloudflare R2 bucket
- Upstash Redis database

### Required Secrets

Set these via `fly secrets set`:

| Secret | Description | Source |
|--------|-------------|--------|
| `DATABASE_URL` | Neon connection string with PgBouncer | Neon Dashboard |
| `REDIS_URL` | Upstash Redis URL | Upstash Console |
| `R2_ACCESS_KEY_ID` | R2 API key | Cloudflare Dashboard |
| `R2_SECRET_ACCESS_KEY` | R2 API secret | Cloudflare Dashboard |
| `R2_ENDPOINT` | R2 endpoint URL | Cloudflare Dashboard |
| `JWT_SECRET` | JWT signing secret | Generate with `openssl rand -hex 32` |
| `SLACK_WEBHOOK_URL` | Slack notifications (optional) | Slack App |
| `INFISICAL_TOKEN` | Secret sync (optional) | Infisical |

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Fly.io                              │
│  ┌─────────────────┐      ┌──────────────────────────────┐ │
│  │ Orchestrator    │◄────►│  Upstash Redis               │ │
│  │ (Go API)        │      │  (Sessions, Cache, Rate Limit)│ │
│  │  Port: 8080     │      └──────────────────────────────┘ │
│  └────────┬────────┘                                       │
│           │                                                │
│           │      ┌──────────────────────────────┐          │
│           └─────►│  Neon Postgres               │          │
│                  │  (Registry Data + WAL)        │          │
│                  └──────────────────────────────┘          │
│                           │                                │
│                           ▼                                │
│                  ┌──────────────────────────────┐          │
│                  │  R2 Backups                  │          │
│                  │  (Daily dumps + WAL)         │          │
│                  └──────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

## Initial Setup

### 1. Create Fly Apps

```bash
# Create production app
fly apps create functionfly-orchestrator

# Create staging app (optional but recommended)
fly apps create functionfly-orchestrator-staging
```

### 2. Configure Secrets

```bash
# Production secrets
fly secrets set --app functionfly-orchestrator \
  DATABASE_URL="postgresql://...neon.tech/functionfly?pgbouncer=true" \
  REDIS_URL="rediss://...upstash.io" \
  R2_ACCESS_KEY_ID="your_key" \
  R2_SECRET_ACCESS_KEY="your_secret" \
  R2_ENDPOINT="https://...r2.cloudflarestorage.com" \
  R2_BUCKET="functionfly-prod" \
  JWT_SECRET="$(openssl rand -hex 32)"

# Staging secrets (use different Neon database)
fly secrets set --app functionfly-orchestrator-staging \
  DATABASE_URL="postgresql://...neon.tech/functionfly_staging?pgbouncer=true" \
  R2_BUCKET="functionfly-staging" \
  # ... other secrets
```

### 3. Create R2 Buckets

```bash
# Using AWS CLI with R2 endpoint
aws s3 mb s3://functionfly-prod \
  --endpoint-url https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com

aws s3 mb s3://functionfly-staging \
  --endpoint-url https://YOUR_ACCOUNT_ID.r2.cloudflarestorage.com
```

### 4. Create Fly Volumes (for temporary function storage)

```bash
# Production volume (10GB)
fly volumes create functionfly_data \
  --app functionfly-orchestrator \
  --region iad \
  --size 10

# Staging volume (5GB)
fly volumes create functionfly_staging_data \
  --app functionfly-orchestrator-staging \
  --region iad \
  --size 5
```

## Deployment

### Manual Deployment

```bash
# Deploy to staging first
./scripts/deploy-fly-production.sh

# Or deploy directly to production (skips staging)
./scripts/deploy-fly-production.sh --skip-staging

# Force deploy without confirmations
./scripts/deploy-fly-production.sh --force
```

### CI/CD Deployment

Deployments are automatically triggered on pushes to `main` branch:

```yaml
# .github/workflows/deploy-fly.yml
```

The workflow:
1. Builds and tests the application
2. Runs security scans
3. Deploys to staging
4. Runs health checks on staging
5. Deploys to production
6. Runs smoke tests
7. Notifies via Slack

### Rolling Back

If deployment fails, automatic rollback occurs. To manually rollback:

```bash
# View available releases
fly releases list --app functionfly-orchestrator

# Rollback to previous version
fly deploy --app functionfly-orchestrator \
  --image $(fly releases list --app functionfly-orchestrator | tail -2 | head -1 | awk '{print $2}')
```

## Monitoring

### Health Checks

The orchestrator exposes several health endpoints:

| Endpoint | Description |
|----------|-------------|
| `/health` | Basic health check (liveness) |
| `/health/detailed` | Detailed health with component status |
| `/metrics` | Prometheus metrics |

### Fly.io Dashboard

Monitor via:
- [Fly.io Dashboard](https://fly.io/dashboard)
- Metrics: Built-in dashboard + Prometheus at `:9090/metrics`
- Logs: `fly logs --app functionfly-orchestrator`

### Alerting

Configure alerts in the workflow file or via Fly's built-in alerting.

## Scaling

### Vertical Scaling

Edit `fly.toml`:

```toml
[[vm]]
  cpu_kind = "shared"
  cpus = 4      # Increase for more CPU
  memory_mb = 4096  # Increase for more memory
```

Then deploy:
```bash
fly deploy --app functionfly-orchestrator
```

### Horizontal Scaling

```bash
# Scale to 3 instances
fly scale count 3 --app functionfly-orchestrator

# Scale per region
fly scale count 2 --region iad
fly scale count 1 --region ams
```

### Multi-Region Deployment

```bash
# Add EU region
fly regions add ams --app functionfly-orchestrator

# Scale in multiple regions
fly scale count 2 --region iad
fly scale count 1 --region ams
```

## Backup & Recovery

### Automated Backups

Daily backups run automatically via GitHub Actions:
- **Schedule**: Daily at 02:00 UTC
- **Storage**: R2 bucket (`functionfly-prod/backups/`)
- **Retention**: 30 days
- **Verification**: Automatic after each backup

### Manual Backup

```bash
# Run backup manually
./scripts/backup-registry-fly.sh

# Verify latest backup
./scripts/backup-registry-fly.sh --verify

# List available backups
./scripts/backup-registry-fly.sh --list
```

### Restore from Backup

```bash
# Download backup from R2
aws s3 cp s3://functionfly-prod/backups/registry_YYYYMMDD_HHMMSS_functions.sql.gz \
  backup.sql.gz --endpoint-url https://...r2.cloudflarestorage.com

# Extract and restore
gunzip backup.sql.gz
psql "${DATABASE_URL}" < backup.sql
```

See [REGISTRY_DISASTER_RECOVERY.md](./REGISTRY_DISASTER_RECOVERY.md) for full DR procedures.

## Troubleshooting

### Common Issues

#### App won't start

```bash
# Check logs
fly logs --app functionfly-orchestrator

# Check status
fly status --app functionfly-orchestrator

# SSH into app for debugging
fly ssh console --app functionfly-orchestrator
```

#### Database connection issues

```bash
# Verify connection string
fly secrets get DATABASE_URL --app functionfly-orchestrator

# Test connectivity from within Fly
fly ssh console --app functionfly-orchestrator --command "pg_isready -d ${DATABASE_URL}"
```

#### High memory usage

```bash
# Check metrics
fly metrics --app functionfly-orchestrator

# Scale up memory
fly scale memory 4096 --app functionfly-orchestrator
```

### Performance Tuning

#### Database

```bash
# Monitor slow queries (from local with psql)
psql "${DATABASE_URL}" -c "SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Check connection pool status
psql "${DATABASE_URL}" -c "SHOW max_connections;"
```

#### Application

- Enable Go profiling: Set `PPROF_ENABLED=true`
- Access pprof: `fly ssh console --app functionfly-orchestrator --command "curl localhost:8080/debug/pprof/heap"`

## Maintenance

### Regular Tasks

| Task | Frequency | Command |
|------|-----------|---------|
| Update Fly CLI | Weekly | `fly version update` |
| Rotate secrets | Quarterly | `fly secrets set` |
| Review logs | Daily | `fly logs --app functionfly-orchestrator` |
| Check backups | Daily | GitHub Actions status |
| Update Go version | As needed | Update `fly.toml` and redeploy |

### Security Updates

```bash
# Update base image (if using Dockerfile)
fly deploy --app functionfly-orchestrator --dockerfile deploy/docker/Dockerfile.orchestrator

# Review security advisories
go list -m -versions all | grep -E "(critical|high)"
```

## Cost Optimization

### Current Estimated Costs (per month)

| Component | Service | Cost |
|-----------|---------|------|
| Orchestrator | Fly.io (2x shared-cpu-2x) | ~$30 |
| Redis | Upstash (10GB) | ~$80 |
| Database | Neon (Pro plan) | ~$50 |
| Storage | R2 (500GB) | ~$5 |
| **Total** | | **~$165/month** |

### Cost Reduction Options

1. **Use Fly.io Redis** instead of Upstash (saves ~$60/month)
2. **Scale down VMs** during low-traffic periods
3. **Enable auto-stop** for staging environment
4. **Reduce R2 storage** with aggressive cleanup policies

## Support

- **Fly.io Docs**: https://fly.io/docs/
- **Neon Docs**: https://neon.tech/docs/
- **Internal Runbooks**: See `docs/REGISTRY_DISASTER_RECOVERY.md`
