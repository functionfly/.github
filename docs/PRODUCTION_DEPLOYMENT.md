# FunctionFly Production Deployment Guide

This guide covers deploying the **FunctionFly** platform (orchestrator API, dashboard, data stores, and optional observability stack) with Docker Compose on your own infrastructure.

For Fly.io + Neon + Cloudflare Pages-style deployment, see [FLY_DEPLOYMENT.md](FLY_DEPLOYMENT.md) and [DOMAIN_AND_COMING_SOON_SETUP.md](DOMAIN_AND_COMING_SOON_SETUP.md).

## Prerequisites

- Docker and Docker Compose installed
- PostgreSQL 17+ database
- Redis 7+ cache
- Domain name with SSL certificate
- Environment variables configured

## Architecture

The production deployment consists of:

1. **Orchestrator API** - Main API server (Go)
2. **Dashboard** - React frontend
3. **PostgreSQL** - Primary database
4. **Redis** - Cache and session store
5. **Caddy** - Reverse proxy and TLS termination
6. **Prometheus** - Metrics collection
7. **Grafana** - Metrics visualization
8. **Loki** - Log aggregation
9. **Promtail** - Log collection

## Environment Configuration

1. Copy the production environment template:

```bash
cp .env.production .env
```

1. Configure the following required variables:

```bash
# Database
DB_PASSWORD=your_secure_password
DB_HOST=postgres
DB_PORT=5432
DB_NAME=functionfly
DB_USER=functionfly

# Redis
REDIS_PASSWORD=your_secure_password
REDIS_HOST=redis
REDIS_PORT=6379

# Security
JWT_SECRET=your_jwt_secret_min_32_chars
API_SHARED_SECRET=your_api_secret_min_32_chars

# Domain
DOMAIN=yourdomain.com
BASE_URL=https://yourdomain.com

# Email (optional)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_email@example.com
SMTP_PASSWORD=your_email_password
```

## Deployment Steps

### 1. Build and Start Services

```bash
# Build all services
docker-compose -f docker-compose.production.yml build

# Start services in detached mode
docker-compose -f docker-compose.production.yml up -d
```

### 2. Initialize Database

**Schema / migrations:** The repository’s `migrations/` tree has historically contained duplicate sequence numbers, which breaks tooling such as golang-migrate in some setups. Many environments run the API with `--skip-migrations` during development; **production** must still apply a known-good schema.

Choose one approach and document it for your team:

1. **Apply SQL from scratch** using your DBA process (Neon/console, `psql`, or a single curated migration bundle you maintain).
2. **Use orchestrator migrate** only if you have validated that your migration set applies cleanly end-to-end on an empty database (recommended: rehearse on a staging branch first).

```bash
# If your build supports it and migrations are validated:
docker-compose -f docker-compose.production.yml exec orchestrator-api ./bin/orchestrator-api migrate

# Seed initial data (optional)
docker-compose -f docker-compose.production.yml exec orchestrator-api ./bin/orchestrator-api seed
```

Do **not** assume `migrate` works without verification—confirm against a disposable database before touching production.

### 3. Verify Deployment

```bash
# Check service status
docker-compose -f docker-compose.production.yml ps

# Check logs
docker-compose -f docker-compose.production.yml logs -f

# Test API health
curl https://yourdomain.com/health
```

## Service URLs

- **API**: <https://yourdomain.com/api>
- **Dashboard**: <https://yourdomain.com>
- **Grafana**: <https://yourdomain.com:3000>
- **Prometheus**: <https://yourdomain.com:9090>

## Monitoring

### Grafana Dashboards

Access Grafana at <https://yourdomain.com:3000> with credentials:

- Username: admin
- Password: (set via GRAFANA_ADMIN_PASSWORD environment variable)

Pre-configured dashboards:

- FunctionFly API Metrics
- Database Performance
- Redis Metrics
- System Resources

### Prometheus Metrics

Access Prometheus at <https://yourdomain.com:9090>

Key metrics to monitor:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `database_connections_active` - Active database connections
- `redis_connected_clients` - Redis connections

### Log Aggregation

Logs are collected by Promtail and stored in Loki.

Access logs via Grafana:

1. Go to Grafana
2. Click "Explore"
3. Select "Loki" datasource
4. Query: `{container="functionfly-orchestrator"}`

## Backup and Recovery

### Database Backup

```bash
# Manual backup
docker-compose -f docker-compose.production.yml exec postgres pg_dump -U functionfly functionfly > backup.sql

# Automated backups are configured via cron
# Backups are stored in /var/backups/postgres/
```

### Database Restore

```bash
# Restore from backup
docker-compose -f docker-compose.production.yml exec -T postgres psql -U functionfly functionfly < backup.sql
```

## Security Considerations

1. **SSL/TLS**: Caddy automatically provisions SSL certificates via Let's Encrypt
2. **Firewall**: Only expose ports 80, 443, and 3000 (Grafana)
3. **Secrets**: Use strong passwords and secrets (minimum 32 characters)
4. **Updates**: Regularly update Docker images and dependencies
5. **Backups**: Maintain regular database backups

## Scaling

### Horizontal Scaling

```bash
# Scale orchestrator API
docker-compose -f docker-compose.production.yml up -d --scale orchestrator-api=3

# Scale dashboard
docker-compose -f docker-compose.production.yml up -d --scale dashboard=3
```

### Database Scaling

For high availability, consider:

- PostgreSQL replication (primary + replicas)
- Connection pooling (PgBouncer)
- Read replicas for analytics

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose -f docker-compose.production.yml logs <service-name>

# Check resource usage
docker stats
```

### Database Connection Issues

```bash
# Test database connectivity
docker-compose -f docker-compose.production.yml exec postgres pg_isready -U functionfly

# Check database logs
docker-compose -f docker-compose.production.yml logs postgres
```

### High Memory Usage

```bash
# Check memory usage
docker stats --no-stream

# Restart services
docker-compose -f docker-compose.production.yml restart
```

## Maintenance

### Regular Tasks

1. **Daily**: Check logs for errors
2. **Weekly**: Review metrics and performance
3. **Monthly**: Update dependencies and security patches
4. **Quarterly**: Review and rotate secrets

### Updates

```bash
# Pull latest images
docker-compose -f docker-compose.production.yml pull

# Restart with new images
docker-compose -f docker-compose.production.yml up -d
```

## Support

For issues and questions:

- GitHub Issues: <https://github.com/functionfly/functionfly/issues>
- Documentation: <https://docs.functionfly.com>
- Community: <https://discord.gg/functionfly>
