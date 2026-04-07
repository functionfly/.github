# Fly.io Deployment Guide for FunctionFly Orchestrator API

This directory contains the configuration and scripts for deploying the FunctionFly Orchestrator API to Fly.io.

## Prerequisites

1. **Fly CLI installed**: `curl -L https://fly.io/install.sh | sh`
2. **Fly account**: Sign up at <https://fly.io>
3. **Neon PostgreSQL**: Create a project at <https://neon.tech>
4. **Redis**: Use Fly.io's managed Redis or an external provider

## Quick Start

### 1. Create the Fly.io App

```bash
# Login to Fly
fly auth login

# Create the app (if not already created)
fly apps create functionfly-api
```

### 2. Set Secrets (upload via CLI)

**Option A – Infisical → Fly (recommended)**  
Secrets are maintained in [Infisical](https://infisical.com); sync them to Fly with:

```bash
INFISICAL_ENV=prod FLY_APP=functionfly-control ./scripts/sync-infisical-to-fly.sh
```

See [docs/INFISICAL_SETUP.md](../docs/INFISICAL_SETUP.md) (canonical store, allowlist, EU region) and [docs/FLY_SECRETS_FROM_ENV.md](../docs/FLY_SECRETS_FROM_ENV.md).

**Option B – Script with inline variables**  
Edit the variables at the top of `.fly/set-secrets.sh`, then run:

```bash
# Use app from current directory (fly.toml)
./.fly/set-secrets.sh production

# Or specify the Fly app name explicitly
./.fly/set-secrets.sh production functionfly-control
# Or: FLY_APP=functionfly-control ./.fly/set-secrets.sh production
```

**Option C – From your local `.env` or manual import**  
See [docs/FLY_SECRETS_FROM_ENV.md](../docs/FLY_SECRETS_FROM_ENV.md) for `fly secrets set`, `fly secrets import`, and what not to upload.

**Option D – Manual `fly secrets set`**  
Replace values and run (use `--app YOUR_APP` if not in the app directory):

```bash
# Database (Neon PostgreSQL)
fly secrets set DB_HOST=your-neon-host.neon.tech --app functionfly-control
fly secrets set DB_PORT=5432 --app functionfly-control
fly secrets set DB_USER=your-username --app functionfly-control
fly secrets set DB_PASSWORD=your-password --app functionfly-control
fly secrets set DB_NAME=functionfly --app functionfly-control
fly secrets set DB_SSLMODE=require --app functionfly-control

# Redis
fly secrets set REDIS_ADDR=your-redis-host:6379 --app functionfly-control
fly secrets set REDIS_PASSWORD=your-redis-password --app functionfly-control

# Authentication & Security (generate with: openssl rand -hex 32)
fly secrets set JWT_SECRET=your-jwt-secret-minimum-32-chars --app functionfly-control
fly secrets set API_SHARED_SECRET=your-api-shared-secret --app functionfly-control
fly secrets set DB_MASTER_KEY_PASSWORD=your-secure-master-key --app functionfly-control

# Application (required for CORS and coming-soon frontend)
fly secrets set BASE_URL=https://api.functionfly.com --app functionfly-control
fly secrets set CORS_ALLOWED_ORIGINS=https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com,https://status.functionfly.com --app functionfly-control
fly secrets set FRONTEND_URL=https://functionfly.com --app functionfly-control
```

**List or unset:**

```bash
fly secrets list --app functionfly-control
fly secrets unset SECRET_NAME --app functionfly-control
```

### 3. Deploy

```bash
# Deploy the application
fly deploy

# Or use the deployment script
.fly/deploy.sh production
```

## Configuration

### fly.toml

The main configuration file includes:

- **App name**: `functionfly-api`
- **Region**: `ord` (Chicago) - can be changed to your preferred region
- **Port**: 8080 (HTTP)
- **VM Size**: shared-cpu-2x with 512MB RAM
- **Health Check**: `/health` endpoint checked every 30 seconds
- **Auto-scaling**: Enabled with minimum 1 machine running

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DB_HOST` | Yes | PostgreSQL host (Neon endpoint) |
| `DB_PORT` | Yes | PostgreSQL port (5432) |
| `DB_USER` | Yes | PostgreSQL username |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_NAME` | Yes | Database name |
| `DB_SSLMODE` | Yes | SSL mode (require for Neon) |
| `REDIS_ADDR` | Yes | Redis address (host:port) |
| `REDIS_PASSWORD` | No | Redis password (if required) |
| `JWT_SECRET` | Yes | Secret key for JWT tokens |
| `BASE_URL` | Yes | Base URL of the API |
| `API_SHARED_SECRET` | Yes | Shared secret for API authentication |
| `DB_MASTER_KEY_PASSWORD` | Yes | Master key for database encryption |
| `LOG_LEVEL` | No | Log level (info, debug) |
| `DEVELOPMENT` | No | Set to "true" for dev mode |

## Using Neon with Fly.io

### Connection String Format

Neon provides two types of connection endpoints:

1. **Direct connection** (for migrations):

   ```
   ep-xxx.us-east-1.aws.neon.tech
   ```

2. **Pooled connection** (recommended for the app):

   ```
   ep-xxx-pooler.us-east-1.aws.neon.tech
   ```

The pooled connection uses PgBouncer for better connection management and is recommended for the orchestrator API.

### Setting Neon Secrets

```bash
# Using DATABASE_URL (recommended - single variable)
fly secrets set DATABASE_URL="postgresql://username:password@ep-xxx-pooler.us-east-1.aws.neon.tech/functionfly?sslmode=require"

# Or using individual DB_* variables
fly secrets set DB_HOST=ep-xxx-pooler.us-east-1.aws.neon.tech
fly secrets set DB_PORT=5432
fly secrets set DB_USER=your-username
fly secrets set DB_PASSWORD=your-password
fly secrets set DB_NAME=functionfly
fly secrets set DB_SSLMODE=require
```

If both `DATABASE_URL` and `DB_*` are set, `DATABASE_URL` takes precedence.

### Environment Variables Precedence

| Priority | Variable | Description |
|----------|----------|-------------|
| 1 | DATABASE_URL | Full connection string (recommended) |
| 2 | DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE | Individual variables |

## Using Fly.io Managed Redis

```bash
# Create a Redis instance
fly redis create

# Get the connection details
fly redis list
```

## Health Checks

The application exposes a `/health` endpoint that Fly.io uses for health checks. The endpoint returns:

- `200 OK` when healthy
- `500 Internal Server Error` when unhealthy

## Monitoring

```bash
# View logs
fly logs

# Check status
fly status

# Open monitoring dashboard
fly dashboard
```

## Scaling

### Vertical Scaling

Edit `fly.toml` and change the VM size:

```toml
[[vm]]
  size = "shared-cpu-4x"  # or "performance-2x" for dedicated CPU
  memory = "1024"
```

### Horizontal Scaling

The configuration already enables:

- `auto_start_machines = true`
- `min_machines_running = 1`

To add more machines, use the Fly.io dashboard or CLI.

## Troubleshooting

### Check Logs

```bash
fly logs -n 100
```

### SSH into Machine

```bash
fly ssh console
```

### Check Health

```bash
fly health check
```

### View Secrets

```bash
fly secrets list
```

### Rollback

```bash
fly releases
fly rollback <release-version>
```

## CI/CD Integration

For automated deployments, set up a GitHub Actions workflow:

```yaml
name: Deploy to Fly.io
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: superfly/flyctl-actions/setup-flyctl@v1
      - run: flyctl deploy --token ${{ secrets.FLY_API_TOKEN }}
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

Generate a token at: <https://fly.io/docs/flyctl/access-tokens/>

## Files in This Directory

| File | Description |
|------|-------------|
| `deploy.sh` | Deployment script |
| `set-secrets.sh` | Interactive script to set all required secrets |
| `secrets.example` | Example secrets configuration |
| `README.md` | This file |

## Quick Reference

```bash
# 1. Set all secrets at once
.fly/set-secrets.sh production

# 2. Deploy the application
.fly/deploy.sh production

# 3. Check status
fly status
fly logs
```
