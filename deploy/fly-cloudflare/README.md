# FunctionFly Deployment - Fly.io + Cloudflare + Vercel Stack

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Cloudflare CDN                               │
│              (DDoS Protection, WAF, Edge Caching)                      │
└──────────────────────┬───────────────────────────────────────────────┘
                       │
          ┌────────────┴────────────┐
          │                         │
   ┌──────▼──────┐          ┌───────▼────────┐
   │  Vercel     │          │   Fly.io       │
   │  (Frontend) │          │  (API + Graph  │
   │             │          │   Execution)   │
   └─────────────┘          └───────┬────────┘
                                    │
                   ┌────────────────┴────────────────┐
                   │                                 │
          ┌────────▼────────┐            ┌──────────▼───────┐
          │     Neon        │            │   Upstash      │
          │  (PostgreSQL +  │            │    Redis       │
          │   pgvector)     │            │  (Cache/Queue) │
          └─────────────────┘            └────────────────┘
```

## Services

| Service | Provider | Purpose |
|---------|----------|---------|
| **Dashboard** | Vercel | React SPA (Next.js static export) |
| **API** | Fly.io | Go orchestrator API + graph execution |
| **Marketing Site** | Vercel | Astro static site |
| **Database** | Neon | PostgreSQL 15 + pgvector + extensions |
| **Cache** | Upstash | Redis 7 with persistence |
| **Edge Functions** | Cloudflare Workers | Webhook receivers, triggers |
| **Storage** | Cloudflare R2 | File uploads, logs, backups |

## Quick Deploy

### 1. Prerequisites

```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Install Wrangler (Cloudflare)
npm install -g wrangler

# Vercel CLI
npm i -g vercel
```

### 2. Environment Setup

```bash
# Copy environment template
cp deploy/fly-cloudflare/.env.example .env

# Fill in your secrets:
# - NEON_DATABASE_URL (from Neon dashboard)
# - UPSTASH_REDIS_URL (from Upstash dashboard)
# - CLOUDFLARE_API_TOKEN
# - FLY_API_TOKEN
```

### 3. Deploy Database (Neon)

Neon is already serverless - just get your connection string:

```bash
# Neon creates branches automatically
# Use the connection string in your Fly secrets
fly secrets set DATABASE_URL="postgresql://..." -a functionfly-api
```

### 4. Deploy Cache (Upstash)

Upstash Redis is serverless:

```bash
# Get your REST API URL from Upstash dashboard
fly secrets set REDIS_URL="rediss://..." -a functionfly-api
fly secrets set UPSTASH_REDIS_REST_URL="https://..." -a functionfly-api
```

### 5. Deploy API (Fly.io)

```bash
cd deploy/fly-cloudflare

# Create Fly app (first time)
fly apps create functionfly-api
fly apps create functionfly-worker

# Deploy API server
fly deploy -c fly-api.toml

# Deploy worker (scales independently)
fly deploy -c fly-worker.toml

# Scale workers based on queue depth
fly autoscale set min=2 max=10 -a functionfly-worker
```

### 6. Deploy Frontend (Vercel)

```bash
cd web/dashboard

# Deploy to Vercel
vercel --prod

# Or use the GitHub integration
```

### 7. Setup Cloudflare

```bash
cd deploy/fly-cloudflare

# Deploy Workers
wrangler deploy workers/webhook-receiver.js

# Configure R2 bucket
wrangler r2 bucket create functionfly-storage

# Setup D1 for edge metadata (optional)
wrangler d1 create functionfly-metadata
```

## Configuration Files

### Fly.io - API Server (`fly-api.toml`)

```toml
app = "functionfly-api"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile.api"

[env]
  PORT = "8080"
  LOG_LEVEL = "info"
  SKIP_MIGRATION_VALIDATION = "true"
  VERIFICATION_ENABLED = "false"
  METRICS_ENABLED = "true"
  AI_SERVICE_URL = "https://ai-service.internal"
  # Neon and Upstash set via secrets

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 2
  processes = ["app"]

  [http_service.concurrency]
    type = "connections"
    hard_limit = 1000
    soft_limit = 500

[checks]
  [checks.alive]
    type = "http"
    port = 8080
    path = "/health/live"
    interval = "10s"
    timeout = "2s"

  [checks.ready]
    type = "http"
    port = 8080
    path = "/health/ready"
    interval = "30s"
    timeout = "5s"

[[vm]]
  memory = "512mb"
  cpu_kind = "shared"
  cpus = 1

# Scale based on CPU/memory
[metrics]
  port = 9090
  path = "/metrics"
```

### Fly.io - Worker (`fly-worker.toml`)

```toml
app = "functionfly-worker"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile.worker"

[env]
  LOG_LEVEL = "info"
  WORKER_QUEUE = "graph_execution"
  # Neon and Upstash via secrets

# Workers don't expose HTTP
[processes]
  app = "./bin/worker --queue=graph_execution"

[[vm]]
  memory = "1gb"
  cpu_kind = "shared"
  cpus = 2

# Auto-scale based on queue depth (requires custom metrics)
[metrics]
  port = 9090
  path = "/metrics"
```

### Cloudflare Worker - Webhook Receiver

```javascript
// workers/webhook-receiver.js
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;
    
    // Route to Fly.io API
    const apiUrl = `https://functionfly-api.fly.dev/webhook${path}`;
    
    // Forward with Cloudflare's edge caching
    const response = await fetch(apiUrl, {
      method: request.method,
      headers: request.headers,
      body: request.body,
    });
    
    // Add edge headers
    const newHeaders = new Headers(response.headers);
    newHeaders.set('CF-Cache-Status', 'DYNAMIC');
    
    return new Response(response.body, {
      status: response.status,
      headers: newHeaders,
    });
  }
};
```

### Vercel - Dashboard (`vercel.json`)

```json
{
  "version": 2,
  "builds": [
    {
      "src": "web/dashboard/package.json",
      "use": "@vercel/static-build",
      "config": {
        "distDir": "dist"
      }
    }
  ],
  "routes": [
    {
      "src": "/api/(.*)",
      "dest": "https://functionfly-api.fly.dev/api/$1"
    },
    {
      "src": "/gx/(.*)",
      "dest": "https://functionfly-api.fly.dev/gx/$1"
    },
    {
      "src": "/(.*)",
      "dest": "/index.html"
    }
  ],
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        {
          "key": "Access-Control-Allow-Origin",
          "value": "*"
        }
      ]
    }
  ]
}
```

## Environment Variables

### Required Secrets (Fly.io)

```bash
# Database
fly secrets set DATABASE_URL="postgresql://...neon.tech/..." -a functionfly-api

# Redis
fly secrets set REDIS_URL="rediss://...upstash.io:6379" -a functionfly-api

# JWT/Security
fly secrets set JWT_SECRET="your-256-bit-secret" -a functionfly-api
fly secrets set API_SHARED_SECRET="internal-api-secret" -a functionfly-api

# AI Service
fly secrets set OPENAI_API_KEY="sk-..." -a functionfly-api
fly secrets set ANTHROPIC_API_KEY="sk-ant-..." -a functionfly-api

# DRE (optional)
fly secrets set DRE_NODE_KEY="base64-encoded-ed25519-key" -a functionfly-api

# Stripe (if using payments)
fly secrets set STRIPE_SECRET_KEY="sk_live_..." -a functionfly-api
fly secrets set STRIPE_WEBHOOK_SECRET="whsec_..." -a functionfly-api
```

## Scaling Strategy

### Fly.io Auto-scaling

```bash
# API servers - scale based on concurrent connections
fly autoscale set min=2 max=20 -a functionfly-api

# Workers - scale based on queue depth (custom metric)
# Add to fly-worker.toml:
# [metrics]
#   [[metrics.scaling]]
#     metric = "queue_depth"
#     min = 2
#     max = 50
```

### Neon

Neon auto-scales compute - no configuration needed. Just monitor:

```bash
# Check usage in Neon dashboard
# Upgrade to Pro for more compute units if needed
```

### Upstash

```bash
# Redis auto-scales
# Monitor in Upstash dashboard
# Upgrade tier if hitting limits
```

## Monitoring

### Fly.io Metrics

```bash
# View logs
fly logs -a functionfly-api

# Metrics dashboard
fly metrics dashboard -a functionfly-api

# Custom metrics via Prometheus
flyctl status --json | jq '.allocations[]'
```

### Cloudflare Analytics

```bash
# View in Cloudflare dashboard:
# - Request volume
# - Cache hit ratio
# - Error rates
# - Edge latency
```

### Vercel Analytics

```bash
# View in Vercel dashboard:
# - Core Web Vitals
# - Real User Monitoring
```

## Backup Strategy

### Neon Backups

Neon provides automatic backups:
- 7 days for Free tier
- 30 days for Pro
- Point-in-time restore via dashboard or API

### Manual Backup Script

```bash
#!/bin/bash
# backup.sh

date=$(date +%Y%m%d_%H%M%S)
filename="backup_${date}.sql"

# Neon supports pg_dump
pg_dump $DATABASE_URL > "/backups/$filename"

# Upload to R2
wrangler r2 object put functionfly-backups/$filename --file=/backups/$filename
```

## Security

### Cloudflare WAF Rules

```bash
# Block suspicious patterns
# - Rate limit: 100 req/min per IP
# - Challenge: 1000 req/10min per IP
# - Block: > 10000 req/hr per IP
```

### Fly.io mTLS (internal)

```bash
# Enable mTLS between API and Workers
fly certs create functionfly-api.internal
```

### Secrets Rotation

```bash
# Rotate database credentials via Neon dashboard
# Update Fly secret
fly secrets set DATABASE_URL="new-connection-string" -a functionfly-api

# Rolling deploy
fly deploy -c fly-api.toml --strategy rolling
```

## Cost Optimization

| Tier | Approx Monthly Cost |
|------|---------------------|
| **Development** | $0-20 |
| Neon Free + Upstash Free + Fly Hobby | |
| **Production** | $50-200 |
| Neon Pro ($15) + Upstash Pro ($20) + Fly Standard ($50-100) | |
| **Scale** | $200-1000+ |
| Neon Enterprise + Upstash Enterprise + Fly multi-region | |

## Troubleshooting

### Connection Issues

```bash
# Test Neon connection
psql $DATABASE_URL -c "SELECT version();"

# Test Upstash
redis-cli -u $REDIS_URL PING

# Test Fly internal
fly ssh console -a functionfly-api
# then: curl localhost:8080/health
```

### Performance

```bash
# Check query performance in Neon dashboard
# Use connection pooling (PgBouncer not needed with Neon's pooler)

# Monitor Fly CPU/memory
fly status -a functionfly-api --json | jq '.allocations[].resources'
```

### Cold Starts

Fly.io machines sleep after inactivity (configurable):

```toml
# fly-api.toml - prevent sleeping for production
[http_service]
  auto_stop_machines = false
  min_machines_running = 2
```

## Migration from Self-Hosted

```bash
# 1. Export from current PostgreSQL
pg_dump -h localhost -U postgres functionfly > backup.sql

# 2. Import to Neon
psql $NEON_DATABASE_URL < backup.sql

# 3. Update connection strings
fly secrets set DATABASE_URL="$NEON_DATABASE_URL" -a functionfly-api

# 4. Deploy
fly deploy -c fly-api.toml
```
