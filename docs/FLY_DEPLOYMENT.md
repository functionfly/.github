# Fly.io Deployment Guide

This document describes how to set up automatic deployments to Fly.io.

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

If you need to deploy manually without GitHub Actions:

```bash
# Install flyctl if not already installed
brew install flyctl

# Login
flyctl auth login

# Deploy production
flyctl deploy --config deploy/fly/functionfly-control/fly.toml

# Check status
flyctl status --app functionfly-control

# View logs
flyctl logs --app functionfly-control
```

## Environment Variables

The deployment uses the following environment variables (configured in `fly.toml`):

| Variable | Value | Description |
|----------|-------|-------------|
| `REGION` | `iad` | Primary region (Virginia) |
| `PRIMARY_REGION` | `iad` | Primary region for data |

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
