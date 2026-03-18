# Fly.io Deployment Guide

This document describes how to set up automatic deployments to Fly.io for the **orchestrator API** (backend).

For **domain setup** (api.functionfly.com + functionfly.com) and **coming-soon mode** (launch page only), see [DOMAIN_AND_COMING_SOON_SETUP.md](DOMAIN_AND_COMING_SOON_SETUP.md).

## Overview

The project has two GitHub Actions workflows for deploying to Fly.io:

1. **Production** ([`fly-deploy.yml`](../.github/workflows/fly-deploy.yml)) - Deploys on push to `main`/`master`
2. **Staging** ([`fly-deploy-staging.yml`](../.github/workflows/fly-deploy-staging.yml)) - Deploys on push to `develop`

## Prerequisites

### 1. Install Flyctl

```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh

# Windows
powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
```

### 2. Login to Fly.io

```bash
flyctl auth login
```

### 3. Create Fly.io Apps

If the apps don't exist yet, create them:

```bash
# Production app
flyctl apps create functionfly-control

# Staging app (if separate)
flyctl apps create functionfly-control-staging
```

### 4. Set Secrets

You need to add the Fly.io API token as a GitHub secret.

#### Option A: Using GitHub Web UI

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add the following secrets:

| Secret Name | Value |
|-------------|-------|
| `FLY_API_TOKEN` | Your Fly.io API token (production) |
| `FLY_API_TOKEN_STAGING` | Your Fly.io API token (staging) - optional |

#### Option B: Using GitHub CLI

```bash
# Get your Fly.io token
flyctl auth token

# Add to GitHub secrets
gh secret set FLY_API_TOKEN --body "$(flyctl auth token)"
```

## Deployment Triggers

### Production Deployment
- **Trigger**: Push to `main` or `master` branch
- **Paths that trigger deployment**:
  - `cmd/orchestrator-api/**`
  - `internal/api/**`
  - `internal/storage/**`
  - `internal/deployment/**`
  - `deploy/fly/**`
  - `go.mod`
  - `go.sum`

### Staging Deployment
- **Trigger**: Push to `develop` branch
- **Same path filters as production**

### Manual Deployment
You can also trigger deployments manually from the GitHub Actions UI:
1. Go to **Actions** → **Deploy to Fly.io**
2. Click **Run workflow**
3. Select the branch and click **Run workflow**

## Manual Deployment (CLI)

**Run from the repository root** so the Docker build context includes `go.mod` and the full app:

```bash
cd /path/to/functionfly   # repo root

brew install flyctl   # if needed
flyctl auth login
flyctl deploy --config deploy/fly/functionfly-control/fly.toml --remote-only
flyctl status --app functionfly-control
flyctl logs --app functionfly-control
```

## Required Secrets (Fly.io)

Set in **Fly Dashboard → functionfly-control → Secrets** (or via `flyctl secrets set`) so the API can reach DB, Redis, etc.:

| Secret | Example | Required |
|--------|---------|----------|
| `DB_HOST` | Neon host or IP | Yes |
| `DB_PORT` | `5432` | Yes |
| `DB_USER` | DB user | Yes |
| `DB_PASSWORD` | DB password | Yes |
| `DB_NAME` | `functionfly` | Yes |
| `DB_SSLMODE` | `require` (Neon) or `disable` | Yes |
| `REDIS_ADDR` | `host:6379` | Yes (or use Upstash) |
| `JWT_SECRET` | Random string | Yes |
| `API_SHARED_SECRET` | Random string | Yes |
| `CORS_ALLOWED_ORIGINS` | `https://functionfly.com,https://www.functionfly.com` | Yes |
| `BASE_URL` | `https://api.functionfly.com` | Yes |
| `FRONTEND_URL` | `https://functionfly.com` | Yes |

Optional: `REDIS_PASSWORD`, `DB_MASTER_KEY_PASSWORD`, `GOOGLE_CLIENT_*`, `GITHUB_CLIENT_*`, etc. See `.env.example`.

Example (replace values):

```bash
flyctl secrets set DB_HOST=ep-xxx-pooler.us-east-1.aws.neon.tech --app functionfly-control
flyctl secrets set DB_PORT=5432 --app functionfly-control
flyctl secrets set DB_USER=functionfly_owner --app functionfly-control
flyctl secrets set DB_PASSWORD=xxx --app functionfly-control
flyctl secrets set DB_NAME=functionfly --app functionfly-control
flyctl secrets set DB_SSLMODE=require --app functionfly-control
flyctl secrets set REDIS_ADDR=xxx.upstash.io:6379 --app functionfly-control
flyctl secrets set JWT_SECRET=xxx --app functionfly-control
flyctl secrets set API_SHARED_SECRET=xxx --app functionfly-control
flyctl secrets set CORS_ALLOWED_ORIGINS=https://functionfly.com,https://www.functionfly.com --app functionfly-control
flyctl secrets set BASE_URL=https://api.functionfly.com --app functionfly-control
flyctl secrets set FRONTEND_URL=https://functionfly.com --app functionfly-control
```

## Custom Domain (api.functionfly.com)

1. Add the certificate:

   ```bash
   flyctl certs add api.functionfly.com --app functionfly-control
   ```

2. Get DNS instructions:

   ```bash
   flyctl certs show api.functionfly.com --app functionfly-control
   ```

3. In your DNS (e.g. Cloudflare for functionfly.com):
   - Add **CNAME**: `api` → `functionfly-control.fly.dev.`
   - If using Cloudflare proxy (orange cloud), also add the **TXT** record `_fly-ownership` shown by `fly certs show`.

4. Verify:

   ```bash
   flyctl certs check api.functionfly.com --app functionfly-control
   curl -s https://api.functionfly.com/healthz
   ```

Then set the frontend’s API URL to `https://api.functionfly.com` (e.g. `VITE_API_URL`).

## Deploying functions to Fly.io (adapter)

When using the **Fly deployment adapter** (provider `fly`) to deploy functions/apps from the orchestrator or CLI:

- **API base URL**: The adapter uses the [Fly Machines API](https://fly.io/docs/machines/api/working-with-machines-api/). Set `FLY_API_HOSTNAME` to override the base URL (default: `https://api.machines.dev`). For app secrets/env vars that use the legacy API, you can set `FLY_API_HOSTNAME=https://api.fly.io/v1`.
- **Provider config** (per deployment):
  - `api_token` (required): Fly API token (e.g. `fly tokens deploy`).
  - `app_name` (required): Fly app name.
  - `org_slug` (required when creating a new app): Organization slug (e.g. `personal`). Omit only if the app already exists.
  - `image` (optional): Full image reference (e.g. `registry.fly.io/myapp:v1`). If omitted, defaults to `registry.fly.io/<app_name>:<version>` (image must be pre-pushed via `flyctl deploy` or CI).
- **Rollback**: Implemented via the Machines API by updating each machine’s image to the previous version tag you provide.

## Environment Variables

The deployment uses these in `fly.toml` and/or Fly secrets:

| Variable | Value | Description |
|----------|-------|-------------|
| `REGION` | `iad` | Primary region (Virginia) |
| `PRIMARY_REGION` | `iad` | Primary region for data |
| `FLY_API_HOSTNAME` | `https://api.machines.dev` | Optional. Fly API base URL for the deployment adapter (Machines API default). |

## Auto-scaling

The Fly.io app is configured with auto-scaling:
- **Minimum**: 1 instance
- **Maximum**: 3 instances
- **Balance regions**: false (all traffic goes to primary region)

## Monitoring

### Check Deployment Status
```bash
flyctl status --app functionfly-control
```

### View Logs
```bash
flyctl logs --app functionfly-control
```

### View Metrics
```bash
flyctl metrics --app functionfly-control
```

### SSH into Container
```bash
flyctl ssh console --app functionfly-control
```

## Troubleshooting

### Deployment Fails

1. Check the GitHub Actions logs
2. Verify `FLY_API_TOKEN` secret is set correctly
3. Ensure the Fly.io app exists: `flyctl apps list`

### Health Check Failures

```bash
# Check health status
flyctl health check --app functionfly-control
```

### Rollback

If you need to rollback to a previous version:

```bash
# List releases
flyctl releases --app functionfly-control

# Rollback to previous release
flyctl rollback --app functionfly-control
```

## Cost Estimate

- **Single region (iad)**: ~$5-15/month
- **With standby regions**: ~$10-30/month
- **Auto-scaling**: Additional costs based on usage
