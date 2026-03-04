# 🚀 FunctionFly Staging Deployment Guide

Deploy a complete staging environment for the FunctionFly platform in minutes. This guide walks you through setting up a production-like environment for testing, demos, and pre-production validation.

---

## 📋 Table of Contents

- [Quick Start](#quick-start) - Get staging running in 5 minutes
- [Prerequisites](#prerequisites) - What you need before starting
- [Step-by-Step Deployment](#step-by-step-deployment) - Detailed instructions
- [Staging Architecture Overview](#staging-architecture-overview) - Understanding the infrastructure
- [Environment Configuration](#environment-configuration) - Configuring key variables
- [Troubleshooting](#troubleshooting) - Common issues and solutions
- [Next Steps](#next-steps) - After staging is running

---

## ⚡ Quick Start

Get your staging environment running in under 5 minutes:

```bash
# 1. Clone and enter the repository
cd functionfly

# 2. Copy the staging environment template
cp .env.staging.example .env.staging

# 3. Edit with your values (see Environment Configuration section)
nano .env.staging  # or use your preferred editor

# 4. Deploy everything
./scripts/deploy-staging.sh

# 5. Verify it's working
curl https://api.staging.functionfly.com/healthz
```

**That's it!** 🎉 Your staging environment should now be accessible at:
- **Dashboard**: https://app.staging.functionfly.com
- **API**: https://api.staging.functionfly.com
- **Edge**: https://edge.staging.functionfly.com

---

## 📦 Prerequisites

Before deploying staging, ensure you have:

### Required Access & Accounts

| Requirement | Purpose | Where to Get |
|------------|---------|--------------|
| **Domain Name** | Host staging subdomains | Cloudflare or any DNS provider |
| **Cloudflare Access** | DNS management + SSL | [cloudflare.com](https://cloudflare.com) |
| **Neon Database** | PostgreSQL staging branch | [neon.tech](https://neon.tech) |
| **Fly.io Account** | Container hosting | [fly.io](https://fly.io) |
| **Docker & Docker Compose** | Run containers locally | [docker.com](https://docker.com) |

### Required Software

```bash
# Check if you have everything installed
docker --version          # Should be 20.10+ 
docker-compose --version  # Should be 2.0+
# OR
docker compose version    # For Docker Compose V2

# Generate secure secrets (macOS/Linux)
openssl rand -base64 48   # For JWT_SECRET
openssl rand -base64 48   # For API_SHARED_SECRET
```

### DNS Records Required

You'll need to create these DNS records in Cloudflare (the setup script can help):

| Subdomain | Record Type | Target | Purpose |
|-----------|-------------|--------|---------|
| `staging.functionfly.com` | CNAME | `functionfly-staging.iad.fly.dev` | Main staging entry |
| `api.staging.functionfly.com` | CNAME | `functionfly-staging.iad.fly.dev` | API endpoint |
| `edge.staging.functionfly.com` | CNAME | `functionfly-staging-edge.iad.fly.dev` | Edge functions |
| `cdn.staging.functionfly.com` | CNAME | `functionfly-staging-cdn.r2.cloudflarestorage.com` | Static assets |
| `app.staging.functionfly.com` | CNAME | `functionfly-staging-dashboard.pages.dev` | Dashboard |

---

## 🛠️ Step-by-Step Deployment

### Step 1: Environment Setup

Create your staging environment file from the template:

```bash
# Copy the template
cp .env.staging.example .env.staging

# Open in your editor
nano .env.staging  # or vim, code, etc.
```

**Required minimum configuration:**

```bash
# Database (Neon staging branch)
DB_HOST=ep-staging-xxxxx.us-east-1.aws.neon.tech
DB_USER=functionfly_owner
DB_PASSWORD=your-secure-password

# Security (generate with: openssl rand -base64 48)
JWT_SECRET=your-64-char-secret-here-change-this
API_SHARED_SECRET=your-64-char-secret-here-too

# URLs (update with your domain)
BASE_URL=https://api.staging.functionfly.com
FRONTEND_URL=https://app.staging.functionfly.com
```

**Verification:**
```bash
# Check file exists
ls -la .env.staging

# Validate syntax (no spaces around =)
grep -E '^[A-Z_]+=.+$' .env.staging | head -5
```

---

### Step 2: DNS Configuration

Use the automated DNS setup script to generate or apply DNS records:

```bash
# View what DNS records will be created (dry run)
./scripts/setup-staging-dns.sh

# Apply DNS records automatically (requires Cloudflare credentials)
./scripts/setup-staging-dns.sh \
  --apply \
  --zone YOUR_CLOUDFLARE_ZONE_ID \
  --token YOUR_CLOUDFLARE_API_TOKEN
```

**Manual DNS Setup** (if not using the script):

1. Log into [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Select your domain (`functionfly.com`)
3. Go to **DNS** → **Records**
4. Add each CNAME record from the table above
5. Ensure **Proxy Status** is enabled (orange cloud)

**Verification:**
```bash
# Check DNS propagation (may take 1-5 minutes)
dig staging.functionfly.com +short
dig api.staging.functionfly.com +short

# Should return your Fly.io targets
# functionfly-staging.iad.fly.dev
```

---

### Step 3: First Deployment

Run the deployment script to start all services:

```bash
# Full deployment (build images, run migrations, start services)
./scripts/deploy-staging.sh

# Or skip building if images already exist
./scripts/deploy-staging.sh --skip-build

# Follow logs after deployment
./scripts/deploy-staging.sh --logs
```

**What the script does:**
1. ✅ Validates your `.env.staging` file
2. ✅ Checks Docker is running
3. ✅ Builds container images
4. ✅ Runs database migrations
5. ✅ Starts all services (Redis, Caddy, API, Health Monitor)
6. ✅ Performs health checks

**Expected output:**
```
ℹ️  Validating environment configuration...
✅ Environment file validated
ℹ️  Checking Docker status...
✅ Docker is running
ℹ️  Building Docker images...
✅ Images built successfully
ℹ️  Starting services...
✅ All services started
ℹ️  Running health checks...
✅ Orchestrator API is healthy
✅ Caddy proxy is responding
✅ Redis is operational

🎉 Staging deployment complete!
   Dashboard: https://app.staging.functionfly.com
   API: https://api.staging.functionfly.com
```

---

### Step 4: Verification Steps

Verify each component is working correctly:

#### 4.1 API Health Check
```bash
# Check API health endpoint
curl https://api.staging.functionfly.com/healthz
# Expected: OK

# Check API status
curl https://api.staging.functionfly.com/v1/status
# Expected: JSON with version and status
```

#### 4.2 Main Staging Domain
```bash
# Check main staging entry
curl https://staging.functionfly.com/health
# Expected: "Staging OK"
```

#### 4.3 Edge Endpoint
```bash
# Test edge subdomain
curl -I https://edge.staging.functionfly.com
# Expected: HTTP/2 200
```

#### 4.4 Service Containers
```bash
# Check all containers are running
docker-compose -f docker-compose.staging.yml ps

# Expected output:
# NAME                           IMAGE                          STATUS
# functionfly-orchestrator-staging   functionfly-orchestrator       Up 2 minutes
# functionfly-caddy-staging          functionfly-caddy              Up 2 minutes
# functionfly-redis-staging          redis:7-alpine                 Up 2 minutes
# functionfly-health-monitor-staging functionfly-health-monitor     Up 2 minutes
```

#### 4.5 Dashboard Access
```bash
# Test dashboard accessibility
curl -I https://app.staging.functionfly.com
# Expected: HTTP/2 200

# Or open in browser
open https://app.staging.functionfly.com  # macOS
xdg-open https://app.staging.functionfly.com  # Linux
```

---

## 🏗️ Staging Architecture Overview

### Subdomain Structure

The staging environment mirrors production with these subdomains:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Staging Domain Structure                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────┐    ┌──────────────────────────────┐  │
│  │ staging.             │    │ api.staging.                 │  │
│  │ functionfly.com      │    │ functionfly.com              │  │
│  ├──────────────────────┤    ├──────────────────────────────┤  │
│  │ • Landing page       │    │ • REST API endpoints         │  │
│  │ • Health checks      │    │ • Authentication             │  │
│  │ • Public routes      │    │ • Function management        │  │
│  │ • /v1/* API routes   │    │ • Metrics endpoint           │  │
│  └──────────────────────┘    └──────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────┐    ┌──────────────────────────────┐  │
│  │ edge.staging.        │    │ cdn.staging.                 │  │
│  │ functionfly.com      │    │ functionfly.com              │  │
│  ├──────────────────────┤    ├──────────────────────────────┤  │
│  │ • Edge execution     │    │ • Static assets              │  │
│  │ • Function runtime   │    │ • SDK files                  │  │
│  │ • High rate limits   │    │ • Documentation              │  │
│  └──────────────────────┘    └──────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ app.staging.functionfly.com                              │  │
│  │ • Dashboard application (React SPA)                      │  │
│  │ • User interface for managing functions                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Service Mapping

| Service | Staging Port | Production Port | Purpose |
|---------|-------------|-----------------|---------|
| Orchestrator API | 8082 | 8080 | Main application API |
| Caddy HTTP | 8083 | 80 | Reverse proxy (HTTP) |
| Caddy HTTPS | 8444 | 443 | Reverse proxy (HTTPS) |
| Redis | 6380 | 6379 | Cache & artifact storage |
| Dashboard | 3001 | 3000 | Development server |

### Data Flow

```
User Request → Cloudflare DNS → Caddy Proxy → Orchestrator API → Neon PostgreSQL
                                                     ↓
                                               Redis Cache
```

1. **Cloudflare** handles DNS resolution and SSL termination
2. **Caddy** routes requests based on subdomain and path
3. **Orchestrator API** processes business logic
4. **Redis** caches artifacts and session data
5. **Neon PostgreSQL** stores persistent data (staging branch)

---

## ⚙️ Environment Configuration

### Critical Variables

These variables **must** be configured for staging to work:

| Variable | Description | How to Generate |
|----------|-------------|-----------------|
| `DB_HOST` | Neon database host | Neon Console → Connection Details |
| `DB_PASSWORD` | Database password | Neon Console → Reset Password |
| `JWT_SECRET` | JWT signing key | `openssl rand -base64 48` |
| `API_SHARED_SECRET` | API authentication | `openssl rand -base64 48` |

### Database Configuration

```bash
# Neon staging branch connection
DB_HOST=ep-staging-xxxxx.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=functionfly_owner
DB_PASSWORD=your-neon-password
DB_NAME=functionfly
DB_SSLMODE=require

# Database encryption (use strong password)
DB_ENCRYPTION_ENABLED=true
DB_MASTER_KEY_PASSWORD=minimum-32-characters-long-key-here
```

### Security Configuration

```bash
# Generate strong secrets (64 characters recommended)
JWT_SECRET=$(openssl rand -base64 48)
API_SHARED_SECRET=$(openssl rand -base64 48)

# Rate limiting (more permissive for staging)
RATE_LIMIT_REQUESTS=200
RATE_LIMIT_WINDOW_SECONDS=60

# CORS origins (include localhost for local development)
CORS_ALLOWED_ORIGINS=https://staging.functionfly.com,https://app.staging.functionfly.com,http://localhost:3000
```

### URL Configuration

```bash
# Staging URLs
BASE_URL=https://api.staging.functionfly.com
FRONTEND_URL=https://app.staging.functionfly.com

# For local development without DNS:
# BASE_URL=http://localhost:8082
# FRONTEND_URL=http://localhost:3000
```

### Optional: OAuth Configuration

For testing authentication flows:

```bash
# Google OAuth (create at https://console.cloud.google.com/)
GOOGLE_CLIENT_ID=your-staging-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-staging-client-secret

# GitHub OAuth (create at https://github.com/settings/developers)
GITHUB_CLIENT_ID=your-staging-github-client-id
GITHUB_CLIENT_SECRET=your-staging-github-client-secret
```

### Optional: Email Configuration

For testing email notifications:

```bash
# SMTP settings (Gmail example)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=staging@functionfly.com
SMTP_PASSWORD=your-app-specific-password
FROM_EMAIL=noreply@staging.functionfly.com
FROM_NAME=FunctionFly Staging
```

---

## 🔧 Troubleshooting

### Common Issues

#### Issue: `.env.staging file not found!`

**Symptom:**
```
❌ .env.staging file not found!
```

**Solution:**
```bash
# Create from template
cp .env.staging.example .env.staging

# Edit the file
nano .env.staging
```

---

#### Issue: Docker daemon is not running

**Symptom:**
```
❌ Docker daemon is not running!
```

**Solution:**
```bash
# Start Docker
sudo systemctl start docker  # Linux
# OR
open -a Docker  # macOS

# Verify it's running
docker info
```

---

#### Issue: Health check fails after deployment

**Symptom:** Services start but health checks timeout.

**Solution:**
```bash
# Check container logs
docker logs functionfly-orchestrator-staging

# Check if database is accessible
docker-compose -f docker-compose.staging.yml exec orchestrator-api \
  sh -c 'echo $DB_HOST'

# Restart services
docker-compose -f docker-compose.staging.yml restart

# Check specific health endpoint
curl -v http://localhost:8082/health
```

---

#### Issue: DNS not resolving

**Symptom:**
```
curl: (6) Could not resolve host: api.staging.functionfly.com
```

**Solution:**
```bash
# Check DNS propagation
dig api.staging.functionfly.com +short

# If empty, check Cloudflare:
# 1. Verify records exist in Cloudflare DNS
# 2. Ensure Proxy Status is enabled (orange cloud)
# 3. Wait 5 minutes for propagation

# Temporary workaround - add to /etc/hosts
sudo echo "127.0.0.1 api.staging.functionfly.com" >> /etc/hosts
```

---

#### Issue: Database connection refused

**Symptom:**
```
Error: connection refused to Neon database
```

**Solution:**
```bash
# Verify connection string
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST/$DB_NAME?sslmode=require"

# Check Neon console:
# 1. Ensure staging branch is active
# 2. Verify IP allowlist includes your server
# 3. Check branch hasn't been deleted

# Test from container
docker-compose -f docker-compose.staging.yml exec orchestrator-api \
  wget -qO- --timeout=10 "https://$DB_HOST" 2>&1 | head
```

---

#### Issue: Caddy reverse proxy returning 502

**Symptom:**
```
HTTP/2 502 Bad Gateway
```

**Solution:**
```bash
# Check Caddy logs
docker logs functionfly-caddy-staging

# Verify orchestrator API is healthy
curl http://localhost:8082/health

# Restart Caddy
docker-compose -f docker-compose.staging.yml restart caddy

# Check Caddy configuration
docker-compose -f docker-compose.staging.yml exec caddy caddy validate --config /etc/caddy/Caddyfile
```

---

#### Issue: Redis connection errors

**Symptom:**
```
Error: connection refused to Redis
```

**Solution:**
```bash
# Check Redis container
docker-compose -f docker-compose.staging.yml ps redis

# Test Redis connection
redis-cli -p 6380 ping

# Restart Redis
docker-compose -f docker-compose.staging.yml restart redis

# Check Redis logs
docker logs functionfly-redis-staging
```

---

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# In .env.staging
LOG_LEVEL=debug

# Then restart services
docker-compose -f docker-compose.staging.yml restart

# Follow logs
docker-compose -f docker-compose.staging.yml logs -f
```

---

## 🚀 Next Steps

After your staging environment is running:

### 1. CI/CD Integration

Automate deployments with GitHub Actions:

```yaml
# .github/workflows/deploy-staging.yml
name: Deploy to Staging

on:
  push:
    branches: [ develop ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy to Staging
        run: |
          echo "${{ secrets.STAGING_ENV }}" > .env.staging
          ./scripts/deploy-staging.sh
```

### 2. Monitoring Setup

Deploy the monitoring stack for staging:

```bash
# Start Prometheus and Grafana for staging
docker-compose -f docker-compose.monitoring.yml up -d

# Access Grafana at http://localhost:3002
# Default credentials: admin/staging-admin
```

### 3. Load Testing

Verify staging performance:

```bash
# Install k6 (https://k6.io/docs/get-started/installation/)

# Run load tests
k6 run load-tests/k6-load-test.js \
  --env BASE_URL=https://api.staging.functionfly.com
```

### 4. Create Test Data

Seed the staging database with test data:

```bash
# Connect to staging database
psql $STAGING_DATABASE_URL

-- Create test tenant and user
INSERT INTO tenants (id, name, slug, status) 
VALUES ('test-tenant', 'Test Tenant', 'test-tenant', 'active');

INSERT INTO users (id, tenant_id, email, role, status)
VALUES ('test-user', 'test-tenant', 'test@staging.functionfly.com', 'admin', 'active');
```

### 5. Documentation

- [ ] Share staging URLs with your team
- [ ] Document any staging-specific configurations
- [ ] Set up staging access controls (IP whitelist if needed)
- [ ] Create runbooks for common operations

---

## 📚 Additional Resources

| Resource | Location | Description |
|----------|----------|-------------|
| Architecture Doc | [`plans/STAGING_DEPLOYMENT_ARCHITECTURE.md`](../plans/STAGING_DEPLOYMENT_ARCHITECTURE.md) | Technical architecture details |
| Environment Template | [`.env.staging.example`](../.env.staging.example) | Complete environment variables |
| Docker Config | [`docker-compose.staging.yml`](../docker-compose.staging.yml) | Service definitions |
| Caddy Config | [`deploy/caddy/staging.Caddyfile`](../deploy/caddy/staging.Caddyfile) | Reverse proxy routing |
| Deploy Script | [`scripts/deploy-staging.sh`](../scripts/deploy-staging.sh) | Deployment automation |
| DNS Script | [`scripts/setup-staging-dns.sh`](../scripts/setup-staging-dns.sh) | DNS configuration |

---

## 🤝 Getting Help

If you encounter issues not covered in this guide:

1. **Check the logs**: `docker-compose -f docker-compose.staging.yml logs -f`
2. **Review the architecture doc**: [`plans/STAGING_DEPLOYMENT_ARCHITECTURE.md`](../plans/STAGING_DEPLOYMENT_ARCHITECTURE.md)
3. **Run diagnostics**: `./scripts/deploy-staging.sh --help`
4. **Contact the team**: Open an issue in the repository

---

## ✅ Deployment Checklist

Use this checklist to ensure a successful deployment:

- [ ] Created `.env.staging` from template
- [ ] Configured database connection (Neon staging branch)
- [ ] Generated JWT and API secrets
- [ ] Set up DNS records in Cloudflare
- [ ] Verified DNS propagation
- [ ] Ran `./scripts/deploy-staging.sh` successfully
- [ ] API health check returns `OK`
- [ ] Main staging domain responds correctly
- [ ] Dashboard is accessible
- [ ] Edge endpoint responds
- [ ] All containers running (`docker-compose -f docker-compose.staging.yml ps`)
- [ ] Tested a simple API request
- [ ] Team has access to staging URLs

---

**Happy deploying! 🎉**

*Last updated: March 4, 2026*
