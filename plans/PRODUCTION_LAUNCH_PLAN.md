# FunctionFly Production Launch Plan

**Objective:** Launch FunctionFly to production by tomorrow morning  
**Date:** 2026-03-23  
**Target:** Full platform including API, Dashboard, Edge Functions, Monitoring

---

## Summary of Requirements

| Component | Choice |
|-----------|--------|
| **Database** | Neon PostgreSQL (cloud) |
| **Hosting** | Fly.io |
| **Domain** | functionfly.com (with subdomains) |
| **Scope** | Full platform (API, Dashboard, Edge Functions, Monitoring) |

---

## Phase 1: Pre-Launch Preparation (Tonight)

### 1.1 Environment Configuration

- [ ] Copy `.env.production.example` to `.env.production`
- [ ] Generate secure secrets:

  ```bash
  openssl rand -base64 48  # For JWT_SECRET
  openssl rand -base64 48  # For API_SHARED_SECRET
  openssl rand -hex 32     # For DB_MASTER_KEY_PASSWORD
  ```

- [ ] Configure production database variables:
  - `DB_HOST` = Neon production endpoint
  - `DB_PASSWORD` = Neon database password
  - `DB_SSLMODE=require`
- [ ] Configure domain variables:
  - `BASE_URL=https://api.functionfly.com`
  - `SSL_DOMAIN=functionfly.com`
  - `SSL_EMAIL=your-email@example.com`
- [ ] Configure CORS origins for production domains

### 1.2 Neon Database Setup

- [ ] Create Neon production project at <https://console.neon.tech>
- [ ] Create main branch (or use default)
- [ ] Get connection string (pooled recommended)
- [ ] Set `DATABASE_URL` in Fly.io secrets OR individual `DB_*` variables
- [ ] Test database connectivity before deployment

### 1.3 Fly.io Configuration

- [ ] Ensure `fly.toml` is configured for production:
  - App name: `functionfly-api`
  - Region: `ord` (Chicago) - adjust as needed
  - Min machines: 1 (auto-scaling enabled)
- [ ] Set Fly.io secrets:

  ```bash
  fly secrets set JWT_SECRET=<generated>
  fly secrets set API_SHARED_SECRET=<generated>
  fly secrets set DB_MASTER_KEY_PASSWORD=<generated>
  fly secrets set DATABASE_URL="postgresql://..."
  fly secrets set REDIS_ADDR=your-redis-host:6379
  fly secrets set REDIS_PASSWORD=<generated>
  fly secrets set BASE_URL=https://api.functionfly.com
  ```

### 1.4 DNS Configuration (Cloudflare)

- [ ] Create DNS records for production:

  ```
  api.functionfly.com       CNAME  functionfly-api.fly.dev
  app.functionfly.com       CNAME  functionfly-dashboard.fly.dev  
  dashboard.functionfly.com CNAME  functionfly-dashboard.fly.dev
  ```

- [ ] Enable proxy (orange cloud) for SSL termination at Cloudflare
- [ ] Configure SSL/TLS mode: Full (strict) - Cloudflare origin cert

### 1.5 Generate Production Secrets

```bash
# Generate all required secrets
export JWT_SECRET=$(openssl rand -base64 48)
export API_SHARED_SECRET=$(openssl rand -base64 48)
export DB_MASTER_KEY_PASSWORD=$(openssl rand -hex 32)
export REDIS_PASSWORD=$(openssl rand -base64 32)

echo "JWT_SECRET=$JWT_SECRET"
echo "API_SHARED_SECRET=$API_SHARED_SECRET"
echo "DB_MASTER_KEY_PASSWORD=$DB_MASTER_KEY_PASSWORD"
echo "REDIS_PASSWORD=$REDIS_PASSWORD"
```

---

## Phase 2: Core Platform Deployment

### 2.1 Deploy Orchestrator API to Fly.io

- [ ] Run `fly deploy` from project root
- [ ] Verify deployment:

  ```bash
  fly status
  fly logs -a functionfly-api
  curl https://api.functionfly.com/health
  ```

- [ ] Check metrics endpoint:

  ```bash
  curl https://api.functionfly.com/metrics
  ```

### 2.2 Deploy Dashboard

- [ ] Build dashboard for production:

  ```bash
  cd web/dashboard
  VITE_API_URL=https://api.functionfly.com VITE_APP_ENV=production npm run build
  ```

- [ ] Deploy to Fly.io:

  ```bash
  fly deploy -a functionfly-dashboard --dockerfile deploy/production/Dockerfile.dashboard
  ```

  OR deploy static files to hosting (Cloudflare Pages, Vercel, etc.)
- [ ] Verify dashboard:

  ```bash
  curl https://app.functionfly.com/health
  ```

### 2.3 Database Migrations

- [ ] Run migrations against production Neon database:

  ```bash
  # If using orchestrator-api binary
  ./bin/orchestrator-api migrate --database-url $DATABASE_URL
  
  # Or via docker exec
  docker-compose -f docker-compose.production.yml exec orchestrator-api \
    ./bin/orchestrator-api migrate
  ```

### 2.4 Create Admin Account

- [ ] Create initial admin user:

  ```bash
  # Using the fly CLI
  go run ./cmd/fly/main.go admin create-user \
    --email admin@functionfly.com \
    --password <secure-password> \
    --role admin
  
  # Or via dedicated admin creation tool
  go run ./cmd/create-admin -production
  ```

---

## Phase 3: Edge Functions Infrastructure

### 3.1 Cloudflare Workers Deployment

- [ ] Deploy edge function to Cloudflare Workers:

  ```bash
  cd edge-targets/cloudflare-workers
  wrangler deploy
  ```

- [ ] Configure `EDGE_HEALTH_URL` in orchestrator
- [ ] Test edge health:

  ```bash
  curl https://edge.functionfly.com/healthz
  ```

### 3.2 Deno Deploy (Optional)

- [ ] Deploy to Deno Deploy if configured:

  ```bash
  cd edge-targets/deno-deploy
  deno deploy deploy
  ```

### 3.3 Edge VPS Nodes

- [ ] If using self-hosted edge nodes, configure:
  - `EDGE_NODES=<YOUR_EDGE_IP_1>:Americas/APAC,<YOUR_EDGE_IP_2>:Europe/Africa
- [ ] Verify edge node connectivity

---

## Phase 4: Monitoring Stack

### 4.1 Deploy Monitoring (If Self-Hosted)

- [ ] Deploy Prometheus:

  ```bash
  docker-compose -f docker-compose.monitoring.yml up -d prometheus
  ```

- [ ] Deploy Grafana:

  ```bash
  docker-compose -f docker-compose.monitoring.yml up -d grafana
  ```

- [ ] Deploy Loki and Promtail:

  ```bash
  docker-compose -f docker-compose.monitoring.yml up -d loki promtail
  ```

- [ ] Import dashboards from `deploy/monitoring/grafana/provisioning/dashboards/`

### 4.2 Configure Alerting

- [ ] Set alert email recipients: `ALERT_EMAIL_RECIPIENTS=alerts@functionfly.com`
- [ ] Configure Slack webhook (optional): `ALERT_SLACK_WEBHOOK_URL=`
- [ ] Verify alert rules in `deploy/monitoring/alert_rules.yml`

---

## Phase 5: Verification & Testing

### 5.1 API Verification

- [ ] Test health endpoint:

  ```bash
  curl https://api.functionfly.com/health
  # Expected: OK or {"status":"healthy"}
  ```

- [ ] Test API status:

  ```bash
  curl https://api.functionfly.com/v1/status
  ```

- [ ] Test authentication flow:

  ```bash
  curl -X POST https://api.functionfly.com/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@functionfly.com","password":"..."}'
  ```

### 5.2 Dashboard Verification

- [ ] Access <https://app.functionfly.com>
- [ ] Verify login works
- [ ] Check browser console for errors
- [ ] Test function deployment workflow

### 5.3 End-to-End Testing

- [ ] Deploy a test function:

  ```bash
  fly login
  fly deploy --path ./examples/hello-world
  ```

- [ ] Invoke the function:

  ```bash
  curl -X POST https://api.functionfly.com/v1/functions/hello-world/invoke \
    -H "Authorization: Bearer <token>" \
    -d '{"name":"Test"}'
  ```

- [ ] Check function logs:

  ```bash
  fly logs hello-world
  ```

---

## Phase 6: Production Checklist

### 6.1 Security Hardening

- [ ] Verify SSL/TLS certificates are valid
- [ ] Ensure CORS is properly configured (not `*` in production)
- [ ] Verify rate limiting is enabled
- [ ] Check security headers are present
- [ ] Disable development mode: `DEVELOPMENT=false`
- [ ] Review `ADVANCED_SECURITY_*` settings

### 6.2 Backup Configuration

- [ ] Configure automated database backups (Neon has built-in)
- [ ] Set backup schedule: `BACKUP_SCHEDULE=0 2 * * *`
- [ ] Verify backup retention: `BACKUP_RETENTION_DAYS=30`
- [ ] Test backup restoration procedure

### 6.3 Monitoring Verification

- [ ] Verify Prometheus metrics are being collected
- [ ] Check Grafana dashboards are populated
- [ ] Verify Loki is receiving logs
- [ ] Test alert notifications (send a test alert)

---

## Phase 7: Launch Tomorrow Morning

### 7.1 Pre-Launch Checklist (30 minutes before)

- [ ] Run final health checks on all services
- [ ] Verify no critical errors in logs
- [ ] Check database connection pool health
- [ ] Verify Redis is operational
- [ ] Confirm all DNS records are resolving

### 7.2 Launch Commands

```bash
# 1. Verify API health
curl -f https://api.functionfly.com/health && echo "API: OK"

# 2. Verify Dashboard
curl -f https://app.functionfly.com && echo "Dashboard: OK"

# 3. Verify Edge
curl -f https://edge.functionfly.com/healthz && echo "Edge: OK"

# 4. Check all Fly.io machines
fly status -a functionfly-api
fly status -a functionfly-dashboard

# 5. View recent logs
fly logs -a functionfly-api --recent
```

---

## Rollback Plan

If issues occur during launch:

### Quick Rollback Commands

```bash
# Rollback to previous deployment
fly deploy -a functionfly-api --image <previous-image>

# Rollback dashboard
fly deploy -a functionfly-dashboard --image <previous-image>

# If database issues: restore from Neon backup
# Neon Console → Backups → Select backup → Restore
```

### Emergency Contacts

| Component | Contact |
|-----------|---------|
| Fly.io Support | <https://fly.io/support> |
| Neon Support | <https://neon.tech/support> |
| Cloudflare Support | <https://cloudflare.com/support> |

---

## File Locations Reference

| Component | File/Location |
|-----------|---------------|
| Production env | `.env.production` |
| Fly.io config | `fly.toml` |
| Docker Compose | `docker-compose.production.yml` |
| Production deploy docs | `docs/PRODUCTION_DEPLOYMENT.md` |
| Dashboard build | `web/dashboard/` |
| Edge workers | `edge-targets/cloudflare-workers/` |
| Monitoring config | `deploy/production/monitoring/` |
| Database migrations | `migrations/` |

---

## Notes

1. **Neon Database**: Neon has built-in continuous backup and point-in-time restore. No external backup solution needed for database.

2. **Fly.io Scaling**: The `fly.toml` is configured with `min_machines_running = 1`. Adjust based on expected load.

3. **SSL Certificates**: Fly.io provides automatic TLS for `*.fly.dev` domains. For custom domains (functionfly.com), Cloudflare handles SSL with origin cert.

4. **Redis**: For production, consider using Fly.io's managed Redis (Upstash) or self-hosted on a VPS.

5. **Monitoring**: The docker-compose.monitoring.yml can be deployed separately if self-hosting monitoring.

6. **Edge Functions**: Cloudflare Workers are the primary edge target. Configure `EDGE_HEALTH_URL` to monitor edge health.

---

## Next Steps After Launch

1. Set up proper domain routing (www to non-www redirect)
2. Configure proper CDN for static assets
3. Set up log aggregation and alerting
4. Configure performance monitoring
5. Set up cost monitoring and alerts
6. Plan for horizontal scaling
