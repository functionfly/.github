# Infisical Setup Guide

This guide explains how to use Infisical for secrets management in the FunctionFly project.

## Overview

FunctionFly now uses Infisical to manage environment variables and secrets across all services. Secrets are stored securely in Infisical and injected at runtime.

## Project Structure

- **Backend**: Go services (orchestrator-api, health-monitor)
- **Frontend**: React/Vite dashboard and Astro site
- **Environments**: dev, staging, prod (configurable)

## Local Development Setup

### 1. Authentication

You've already completed this step - you're logged in as `olyntar@gmail.com`.

### 2. Project Initialization

The project has been initialized as "FunctionFly" in your Infisical organization.

### 3. Running Services Locally

#### Backend Services

Use the updated Makefile commands:

```bash
make dev          # Start development environment with secrets
make api          # Run orchestrator API with secrets
make health-monitor  # Run health monitor with secrets
```

#### Frontend Services

Use the updated npm scripts:

```bash
# Dashboard
cd web/dashboard
npm run dev

# Site
cd web/site
npm run dev
```

## Production Deployment

### 1. Create Service Tokens

For production deployments, create service tokens in the Infisical dashboard:

1. Go to [Infisical Dashboard](https://app.infisical.com)
2. Select your FunctionFly project
3. Go to "Service Tokens" in the sidebar
4. Click "Create Service Token"
5. Configure:
   - Name: `backend-prod` (or appropriate name)
   - Environment: `prod` (or your target environment)
   - Scope: Read access to all secrets (`*`)
   - Expiration: Set as needed (or never expire for long-running services)

### Optional: Admin Registry AI descriptions (Open Router)

To enable "Generate with AI" for function descriptions in Admin → Registry, add this secret in Infisical for your environment (e.g. `dev`):

- **Key:** `OPENROUTER_API_KEY`
- **Value:** Your key from [Open Router](https://openrouter.ai) (e.g. `sk-or-v1-...`)

Leave unset if you don't need AI-generated descriptions.

### 2. Docker Deployment

#### Using Service Tokens

1. Copy `.env.docker` to `.env`:

   ```bash
   cp .env.docker .env
   ```

2. Edit `.env` and add your service token:

   ```bash
   INFISICAL_TOKEN=st-1234567890abcdef...
   ```

3. Run with Docker Compose:

   ```bash
   docker-compose up
   ```

#### Alternative: Using Interactive Login

For development containers, you can mount your Infisical config:

```yaml
# Add to docker-compose.yml service
volumes:
  - ~/.infisical:/root/.infisical:ro
  - .:/app
```

### 3. CI/CD Integration

For CI/CD pipelines, use service tokens as environment variables:

```bash
# GitHub Actions example
env:
  INFISICAL_TOKEN: ${{ secrets.INFISICAL_TOKEN }}
```

## Environment Variables

### Backend Secrets

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `JWT_SECRET`, `API_SHARED_SECRET`
- `PORT`
- Security headers: `RATE_LIMIT_*`, `CORS_*`, `CONTENT_SECURITY_POLICY`, `HSTS_MAX_AGE`

### Frontend Secrets

- **Dashboard**: `VITE_SANITY_*`, `VITE_API_*`, `VITE_METRICS_*`
- **Site**: `SANITY_*`

## Troubleshooting

### Common Issues

1. **"You must be logged in"**
   - Run `infisical login` and complete browser authentication

2. **"Missing credentials for generating plainTextEncryptionKey"**
   - Ensure you're in the correct project directory with `.infisical.json`
   - Try re-initializing: `infisical init`

3. **Service token issues**
   - Verify token has correct permissions for the environment
   - Check token hasn't expired

4. **Secrets not loading**
   - Verify environment name matches (`--env=dev`)
   - Check secret names match exactly

### Getting Help

- [Infisical Documentation](https://infisical.com/docs)
- [Infisical Slack Community](https://infisical.com/slack)

## Migration Notes

- All existing `.env` files have been migrated to Infisical
- Original `.env` files are kept for reference but should not be used in production
- Secrets are now managed centrally and version-controlled through Infisical
