# FunctionFly Production Launch Plan - 2 Days

## Overview

This plan outlines the steps to launch FunctionFly to production within 2 days. The infrastructure uses Fly.io for the API, Neon PostgreSQL for the database, Upstash Redis for caching, Cloudflare Pages for the dashboard, and Infisical for secrets management.

**Current Time**: March 19, 2026  
**Target Launch**: March 21, 2026

---

## Infrastructure Summary

| Component | Provider | Notes |
|-----------|----------|-------|
| API | Fly.io | `functionfly-api` app |
| Database | Neon PostgreSQL | Already provisioned |
| Redis | Upstash | Serverless, cheapest option |
| Dashboard | Cloudflare Pages | Separate from API |
| DNS/SSL | Cloudflare | `functionfly.com` configured |
| Secrets | Infisical | Already integrated in codebase |

---

## Day 1: Infrastructure Setup & Core Deployment

### Phase 1: Redis Setup (Upstash)

- [ ] **1.1** Create Upstash Redis account at <https://upstash.com>
- [ ] **1.2** Create a new Redis database (free tier: 10K commands/day)
- [ ] **1.3** Note the REST API URL and token
- [ ] **1.4** Add Upstash Redis credentials to Infisical

```
INFISICAL_TOKEN=<from-infisical>
REDIS_ADDR=<upstash-rest-url>
REDIS_PASSWORD=<upstash-token>
```

### Phase 2: Secrets Configuration (Infisical)

- [ ] **2.1** Ensure all required secrets are in Infisical:
  - `DATABASE_URL` (Neon connection string)
  - `JWT_SECRET` (generate with `openssl rand -base64 48`)
  - `API_SHARED_SECRET` (generate with `openssl rand -base64 48`)
  - `DB_MASTER_KEY_PASSWORD` (32+ character encryption key)
  - `REDIS_ADDR` (from Upstash)
  - `REDIS_PASSWORD` (from Upstash)
  - `BASE_URL=https://api.functionfly.com`
  - `FRONTEND_URL=https://app.functionfly.com`
  - `CORS_ALLOWED_ORIGINS=https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com`
  - `LOG_LEVEL=info`
  - `DEVELOPMENT=false`
- [ ] **2.2** Verify secrets are accessible via Infisical CLI

### Phase 3: DNS Configuration

- [ ] **3.1** Update Cloudflare DNS records for production:

| Record | Type | Name | Target | Proxy |
|--------|------|------|--------|-------|
| API | CNAME | api | `functionfly-api.fly.dev` | Yes |
| Dashboard | CNAME | app | `functionfly-dashboard.pages.dev` | Yes |
| Edge | CNAME | edge | `functionfly-edge.fly.dev` | Yes |
| CDN | CNAME | cdn | `functionfly-cdn.r2.cloudflarestorage.com` | Yes |
| Root | A | @ | Fly.io IP | Yes |

- [ ] **3.2** Verify SSL/TLS mode is "Full (strict)" in Cloudflare
- [ ] **3.3** Test DNS propagation: `dig api.functionfly.com`

### Phase 4: API Deployment to Fly.io

- [ ] **4.1** Build the orchestrator API:

  ```bash
  go build -o bin/orchestrator-api ./cmd/orchestrator-api
  ```

- [ ] **4.2** Deploy to Fly.io:

  ```bash
  fly deploy --app functionfly-api
  ```

- [ ] **4.3** Set secrets via Infisical:

  ```bash
  infisical run -- fly secrets set \
    DATABASE_URL="postgresql://..." \
    JWT_SECRET="..." \
    REDIS_ADDR="..." \
    # ... other secrets
  ```

- [ ] **4.4** Scale machines for production:

  ```bash
  fly scale count 2 --region ord
  ```

- [ ] **4.5** Verify deployment:

  ```bash
  curl https://api.functionfly.com/health
  ```

### Phase 5: Dashboard Deployment to Cloudflare Pages

- [ ] **5.1** Build the dashboard:

  ```bash
  cd web/dashboard
  npm install
  npm run build
  ```

- [ ] **5.2** Deploy to Cloudflare Pages:

  ```bash
  wrangler pages deploy dist --project-name=functionfly-dashboard
  ```

- [ ] **5.3** Configure build settings:
  - Build command: `npm run build`
  - Build output directory: `dist`
  - Environment variables: `VITE_API_URL=https://api.functionfly.com`

- [ ] **5.4** Add custom domain in Cloudflare Pages dashboard:
  - Domain: `app.functionfly.com`
  - Ensure SSL certificate is issued

### Phase 6: Edge Function Deployment (Optional for Day 1)

- [ ] **6.1** Deploy edge target of choice:
  - Cloudflare Workers: `edge-targets/cloudflare-workers/`
  - Fly.io: `edge-targets/fly/`
  - Deno Deploy: `edge-targets/deno-deploy/`

- [ ] **6.2** Configure edge proxy with production `BACKEND_URL=https://api.functionfly.com`

---

## Day 2: Testing, Verification & Launch

### Phase 7: Database Migrations

- [ ] **7.1** Run production migrations:

  ```bash
  # Connect to production database and run migrations
  # Note: Use --skip-migrations if duplicate sequence numbers exist
  fly ssh console -a functionfly-api
  ./orchestrator-api --skip-migrations
  ```

- [ ] **7.2** Verify database schema is correct:

  ```bash
  psql "$DATABASE_URL" -c "\dt"
  ```

### Phase 8: Functional Testing

- [ ] **8.1** Test authentication endpoints:

  ```bash
  curl -X POST https://api.functionfly.com/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"test123","username":"testuser"}'
  ```

- [ ] **8.2** Test function execution:

  ```bash
  # Upload and execute a test function
  curl -X POST https://api.functionfly.com/v1/functions \
    -H "Authorization: Bearer <token>"
  ```

- [ ] **8.3** Test API key creation and usage

- [ ] **8.4** Test State Fabric operations (stores, events, snapshots)

- [ ] **8.5** Test Agent Swarm endpoints

- [ ] **8.6** Test Flywheel Network operations

### Phase 9: Dashboard Testing

- [ ] **9.1** Verify login flow at <https://app.functionfly.com>
- [ ] **9.2** Test function creation and deployment UI
- [ ] **9.3** Test analytics dashboard
- [ ] **9.4** Test State Fabric UI components
- [ ] **9.5** Verify all API calls proxy correctly to `api.functionfly.com`

### Phase 10: Monitoring Setup

- [ ] **10.1** Configure Prometheus metrics endpoint:

  ```bash
  curl https://api.functionfly.com/metrics
  ```

- [ ] **10.2** Set up Grafana dashboards (if not already configured)

- [ ] **10.3** Configure health check monitoring:
  - Primary: `https://api.functionfly.com/health`
  - Edge: `https://edge.functionfly.com/healthz`

- [ ] **10.4** Set up alerts for:
  - High error rate (>1%)
  - High latency (>500ms p95)
  - Machine restarts

### Phase 11: Production Checklist

- [ ] **11.1** Verify all 12-factor config is correct
- [ ] **11.2** Confirm `DEVELOPMENT=false`
- [ ] **11.3** Verify CORS is restricted to production domains
- [ ] **11.4** Confirm rate limiting is appropriate for production
- [ ] **11.5** Test backup and restore procedures
- [ ] **11.6** Document incident response procedures
- [ ] **11.7** Verify admin account access: `admin@functionfly.com` / `admin123`

### Phase 12: Launch

- [ ] **12.1** Update DNS to point production traffic
- [ ] **12.2** Enable maintenance mode during final deployment (if needed)
- [ ] **12.3** Final smoke tests
- [ ] **12.4** Announce launch
- [ ] **12.5** Monitor error rates and performance for 24 hours

---

## Production URLs

| Service | URL |
|---------|-----|
| API | <https://api.functionfly.com> |
| Dashboard | <https://app.functionfly.com> |
| Edge | <https://edge.functionfly.com> |
| CDN | <https://cdn.functionfly.com> |
| Health | <https://api.functionfly.com/health> |
| Metrics | <https://api.functionfly.com/metrics> |

---

## Environment Variables Reference

### Required Secrets (Infisical)

```
DATABASE_URL=postgresql://user:pass@ep-xxx.pooler.region.aws.neon.tech/functionfly?sslmode=require
JWT_SECRET=<64-char-random-string>
API_SHARED_SECRET=<64-char-random-string>
DB_MASTER_KEY_PASSWORD=<32+ char password>
REDIS_ADDR=<upstash-rest-url>
REDIS_PASSWORD=<upstash-token>
BASE_URL=https://api.functionfly.com
FRONTEND_URL=https://app.functionfly.com
CORS_ALLOWED_ORIGINS=https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://auth.functionfly.com,https://admin.functionfly.com
LOG_LEVEL=info
DEVELOPMENT=false
```

### Optional (Set via Fly Secrets)

```
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=<sendgrid-api-key>
FROM_EMAIL=noreply@functionfly.com
STRIPE_PUBLISHABLE_KEY=pk_live_...
STRIPE_SECRET_KEY=sk_live_...
```

---

## Troubleshooting

### API Won't Start

```bash
# Check logs
fly logs -a functionfly-api

# SSH into machine
fly ssh console -a functionfly-api

# Check secrets
fly secrets list -a functionfly-api
```

### Database Connection Issues

```bash
# Test Neon connection
psql "$DATABASE_URL" -c "SELECT 1"

# Check SSL mode
# Ensure DB_SSLMODE=require
```

### Dashboard API Calls Failing

```bash
# Verify CORS configuration
curl -I -X OPTIONS https://api.functionfly.com/v1/auth/login \
  -H "Origin: https://app.functionfly.com"

# Check VITE_API_URL in dashboard build
```

---

## Success Criteria

The production launch is considered successful when:

1. ✅ API responds at `https://api.functionfly.com/health` with 200
2. ✅ Dashboard loads at `https://app.functionfly.com`
3. ✅ User can sign up and log in
4. ✅ Function can be published and executed
5. ✅ API keys can be created and used
6. ✅ Error rate < 1% for 1 hour
7. ✅ p95 latency < 500ms

---

## Notes

- **Migrations**: Use `--skip-migrations` flag if duplicate sequence numbers exist in `migrations/`
- **Admin Account**: `admin@functionfly.com` / `admin123` (change immediately after first login)
- **Backups**: Neon provides automatic daily backups; verify backup retention settings
- **Rate Limits**: Default production limits are aggressive; adjust via `RATE_LIMIT_REQUESTS`
