# FunctionFly Production Deployment Checklist

This checklist ensures your FunctionFly deployment is ready for production use.

---

## Pre-Deployment

### 1. Infrastructure Setup

- [ ] PostgreSQL 17+ with pgvector extension available
- [ ] Redis 7+ available
- [ ] At least 4GB RAM and 2 CPU cores for the API server
- [ ] At least 20GB disk space for PostgreSQL data
- [ ] Network access to Cloudflare API (for Workers backend)

### 2. Domain & DNS

- [ ] Domain registered and pointing to your server
- [ ] DNS records configured for:
  - [ ] `api.yourdomain.com` → API server
  - [ ] `app.yourdomain.com` → Dashboard (optional)
  - [ ] `cdn.yourdomain.com` → CDN assets (optional)
- [ ] Cloudflare proxy enabled for public endpoints

### 3. SSL/TLS Certificates

- [ ] **PostgreSQL SSL Certificates:**
  ```bash
  # Create secrets directory
  mkdir -p ./secrets && chmod 700 ./secrets

  # Generate self-signed certificate (development/testing only)
  openssl req -new -x509 -days 365 -nodes \
    -out ./secrets/postgresql.crt \
    -keyout ./secrets/postgresql.key \
    -subj "/C=US/ST=State/L=City/O=FunctionFly/CN=postgres"

  # For Let's Encrypt certificates:
  ln -s /etc/letsencrypt/live/$(hostname)/fullchain.pem ./secrets/postgresql.crt
  ln -s /etc/letsencrypt/live/$(hostname)/privkey.pem ./secrets/postgresql.key
  ```

- [ ] **Caddy automatic HTTPS:** Enabled by default (`auto_https on` in Caddyfile)

### 4. Environment Variables

Copy `.env.production.example` to `.env.production` and set:

- [ ] `DB_PASSWORD` - Strong PostgreSQL password (32+ chars)
- [ ] `REDIS_PASSWORD` - Strong Redis password
- [ ] `JWT_SECRET` - Strong JWT signing secret (32+ chars)
- [ ] `API_SHARED_SECRET` - Internal API shared secret
- [ ] `STRIPE_SECRET_KEY` - Stripe live API key (if using billing)
- [ ] `STRIPE_WEBHOOK_SECRET` - Stripe webhook signing secret
- [ ] `RESEND_API_KEY` - Resend API key (for emails)
- [ ] `SSL_EMAIL` - Email for Let's Encrypt certificates
- [ ] `SSL_DOMAIN` - Your domain (e.g., `yourdomain.com`)
- [ ] `MONITORING_BASIC_AUTH_HASH` - Basic auth for monitoring endpoints

Generate monitoring password hash:
```bash
docker run -it caddy:2-alpine caddy hash-password --algorithm bcrypt
```

---

## Database Setup

### 1. PostgreSQL Configuration

- [ ] SSL enabled and configured
- [ ] `listen_addresses = '127.0.0.1'` (or internal network only)
- [ ] Strong password configured (`POSTGRES_PASSWORD`)
- [ ] `scram-sha-256` authentication enabled
- [ ] Replication configured (if using read replica)

### 2. Database Migrations

Run migrations before starting the API:
```bash
# Apply migrations directly (golang-migrate has duplicate sequence issues)
psql -h localhost -U postgres -d functionfly -f migrations/000001_initial_schema.up.sql

# Or use the API with migrations enabled (dev only)
./bin/orchestrator-api --migrations
```

### 3. Initial Admin User

Create production admin account:
```bash
go run ./cmd/create-admin -production
# Set ADMIN_CREATE_PASSWORD environment variable
```

---

## Backup Configuration

### 1. Local Backups

- [ ] Backup directory created: `./backups`
- [ ] Backup retention set: `BACKUP_RETENTION_DAYS=30`
- [ ] Backup schedule configured: `BACKUP_SCHEDULE=0 2 * * *`

### 2. Offsite Backups (Recommended)

- [ ] S3/GCS/R2 bucket created and accessible
- [ ] `OFFSITE_BUCKET` set (e.g., `s3://functionfly-backups`)
- [ ] `OFFSITE_RETENTION_DAYS=90` set
- [ ] `BACKUP_DELETE_ENABLED=true` after testing

Test backup restoration:
```bash
# Create test backup
./scripts/backup-database.sh

# Verify backup file
./scripts/backup-database.sh --verify-only

# Test restore (never on production without verifying)
zcat backups/functionfly_YYYYMMDD_HHMMSS.sql.gz | psql -h localhost -U postgres -d functionfly
```

### 3. Point-In-Time Recovery (PITR)

PostgreSQL supports PITR for granular disaster recovery. This requires:
- Continuous WAL archiving to an offsite location
- A base backup to restore from

#### Setup PITR:

1. **Configure PostgreSQL for WAL archiving:**
```bash
# In postgresql.conf
wal_level = replica
max_wal_senders = 3
wal_keep_size = 1GB
archive_mode = on
archive_command = 'scp %p user@backup-server:/var/backups/wal/%f'
```

2. **Create base backup (before disaster):**
```bash
pg_basebackup -h localhost -U postgres -D /var/backups/base -Ft -z -P
```

3. **Recover to specific point in time:**
```bash
# Stop PostgreSQL
systemctl stop postgresql

# Backup current data
mv /var/lib/postgresql/data /var/lib/postgresql/data.failed

# Restore base backup
tar -xzf /var/backups/base/backup.tar.gz -C /var/lib/postgresql/data

# Create recovery signal
touch /var/lib/postgresql/data/recovery.signal

# Configure recovery.conf (PostgreSQL 12-14) or postgresql.conf (15+)
# In postgresql.conf for PG 15+:
restore_command = 'scp user@backup-server:/var/backups/wal/%f %p'
recovery_target_time = '2024-01-15 14:30:00 UTC'
recovery_target_action = 'promote'

# Start PostgreSQL
systemctl start postgresql
```

#### PITR Best Practices:

- Test PITR quarterly on a staging environment
- Store WAL archives in a different region/cloud from base backups
- Use `recovery_target_timeline = 'latest'` for timeline continuity
- Document the exact target time before beginning recovery
- After recovery, run `ANALYZE` to update statistics

#### Verify PITR Readiness:

```bash
# Check WAL archiving is working
psql -h localhost -U postgres -c "SELECT * FROM pg_stat_archiver;"

# Should show non-zero counts for archived_count
# archived_count | last_archived_wal | last_archived_time
# ---------------+-------------------+------------------------
# 12345         | 0000000100000ABC  | 2024-01-15 10:30:00
```

---

## Monitoring Setup

### 1. Prometheus

- [ ] Prometheus accessible at `http://localhost:9091`
- [ ] Targets visible at Status → Targets
- [ ] Metrics being collected for:
  - [ ] Orchestrator API (`orchestrator-api:9090`)
  - [ ] PostgreSQL (`postgres:5432`)
  - [ ] Redis (`redis:6379`)
  - [ ] Caddy (`caddy:80`)

### 2. Grafana

- [ ] Grafana accessible at `http://localhost:3001`
- [ ] Default admin password changed
- [ ] Data sources configured (Prometheus, Loki)
- [ ] Dashboards imported

### 3. Loki (Log Aggregation)

- [ ] Loki accessible at `http://localhost:3100`
- [ ] Authentication disabled for internal use (`auth_enabled: false`)
- [ ] Log retention configured (720h = 30 days)

### 4. DataDog (Optional)

To enable DataDog remote write:
```bash
# Set API key
echo -n "your-datadog-api-key" > /etc/secrets/datadog_api_key
chmod 600 /etc/secrets/datadog_api_key

# Enable in prometheus.yml
DATADOG_ENABLED=true
curl -X POST http://localhost:9090/-/reload
```

---

## Security Checklist

### 1. Network Security

- [ ] PostgreSQL port 5432 not exposed publicly
- [ ] Redis port 6379 not exposed publicly
- [ ] Only necessary ports exposed (80, 443, 8080)
- [ ] Internal network traffic encrypted (SSL/TLS)

### 2. Application Security

- [ ] `DEBUG=false` in production
- [ ] `APP_ENV=production` set
- [ ] `LOG_FORMAT=json` for production logging
- [ ] Strong secrets configured (not defaults)
- [ ] CORS origins restricted to your domains

### 3. Secrets Management

- [ ] No secrets in version control
- [ ] Secrets stored in environment variables or secrets manager
- [ ] Secrets rotated periodically
- [ ] Backup of secrets securely stored

---

## Health Checks

### 1. API Health Endpoints

```bash
# Liveness check
curl http://localhost:8080/health/live

# Readiness check (includes dependencies)
curl http://localhost:8080/health/ready

# Full health status
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ready",
  "checks": {
    "database": {"status": "up", "latency_ms": 5},
    "redis": {"status": "up", "latency_ms": 2}
  }
}
```

### 2. Caddy Health

```bash
curl http://localhost:80/health
curl http://localhost/healthz
```

### 3. PostgreSQL Health

```bash
pg_isready -h localhost -p 5432 -U postgres
```

### 4. Redis Health

```bash
redis-cli -a YOUR_PASSWORD ping
# Expected: PONG
```

---

## Deployment Verification

### 1. Start Services

```bash
# Start all services
cd deploy/production
docker-compose up -d

# Or for full production with PgBouncer:
docker-compose -f docker-compose.yml up -d
```

### 2. Verify Services

```bash
# Check all services are running
docker-compose ps

# Check logs
docker-compose logs -f orchestrator-api

# Verify health
curl http://localhost:8080/health/ready
```

### 3. Test End-to-End

- [ ] API responds at `https://api.yourdomain.com/healthz`
- [ ] Dashboard loads at `https://app.yourdomain.com`
- [ ] User registration/login works
- [ ] Can create and deploy a test function
- [ ] Function executes and returns result

---

## Troubleshooting

### Common Issues

**PostgreSQL connection refused:**
- Check PostgreSQL is running: `pg_isready -h localhost -p 5432`
- Check SSL certificates exist and are valid
- Verify `DB_HOST` and `DB_PORT` in environment

**Redis connection refused:**
- Check Redis is running: `redis-cli ping`
- Verify `REDIS_PASSWORD` is set correctly

**Caddy certificate errors:**
- Check DNS is pointing to server
- Check ports 80 and 443 are open
- View Caddy logs: `docker-compose logs -f caddy`

**Backup failures:**
- Check AWS/GCS credentials are valid
- Verify bucket exists and is accessible
- Check disk space: `df -h`

---

## Rollback Procedure

If deployment fails:

```bash
# Stop current services
docker-compose down

# Restore database from backup
zcat backups/functionfly_YYYYMMDD_HHMMSS.sql.gz | psql -h localhost -U postgres -d functionfly

# Start previous version
# (update image tag to previous version)
docker-compose up -d
```

---

## Post-Deployment

- [ ] Monitor error rates for 24 hours
- [ ] Verify backups are running successfully
- [ ] Set up alerting for critical issues
- [ ] Document any custom configurations
- [ ] Train operations team on troubleshooting
