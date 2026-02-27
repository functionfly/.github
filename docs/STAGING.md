# FunctionFly Staging Environment

This document describes the staging environment setup for FunctionFly, providing an isolated testing environment that mirrors production while using cost-effective Neon database branching.

## Overview

The staging environment is a complete, isolated replica of the production system designed for:
- Testing new features before production deployment
- Bug reproduction and fixing
- Performance testing and validation
- Integration testing with external services
- Pre-deployment validation

## Architecture

### Infrastructure Stack
- **Database**: Neon Postgres (staging branch)
- **Cache**: Redis (separate instance from production)
- **Reverse Proxy**: Caddy (staging configuration)
- **Application**: FunctionFly Orchestrator API (staging build)
- **Monitoring**: Health monitor service

### Network Isolation
- **API Port**: `8082` (vs production `8080`)
- **Web Proxy**: `8083` (vs production `8081`)
- **Redis**: `6380` (vs production `6379`)
- **Database**: Isolated Neon branch with full schema

## Neon Database Branching

### Branch Configuration
- **Branch Name**: `staging`
- **Branch ID**: `br-icy-hall-ai6e37zm`
- **Parent Branch**: `production`
- **Region**: AWS US East 1 (N. Virginia)
- **Compute**: 0.25 CU (shared with production limits)

### Connection Details
```bash
Host: ep-lucky-bird-aie8580h.c-4.us-east-1.aws.neon.tech
Port: 5432
Database: functionfly
User: functionfly_owner
SSL Mode: require
```

### Cost Benefits
- **Zero additional storage costs**: Neon branching uses copy-on-write
- **Instant provisioning**: Branch created in seconds
- **Isolated data**: Changes don't affect production
- **Automatic cleanup**: Easy branch deletion when done

## Configuration Files

### Environment Variables (`.env.staging`)
```bash
# Environment
NODE_ENV=staging
ENVIRONMENT=staging

# Database (Neon staging branch)
DB_HOST=ep-lucky-bird-aie8580h.c-4.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=functionfly_owner
DB_PASSWORD=npg_YzCkZWNy97Dv
DB_NAME=functionfly
DB_SSLMODE=require

# Application
PORT=8082
USE_SUPABASE=false

# Redis (staging instance)
REDIS_ADDR=localhost:6380
REDIS_DB=1
ARTIFACT_TTL=24h

# Security (staging-optimized)
RATE_LIMIT_REQUESTS=200
ADVANCED_SECURITY_SLIDING_WINDOW_LIMIT=200
# ... additional security settings
```

### Docker Compose (`docker-compose.staging.yml`)
- **orchestrator-api**: Main application container
- **redis**: Caching and artifact storage
- **caddy**: Reverse proxy and load balancer
- **health-monitor**: System monitoring service

### Caddy Configuration (`deploy/caddy/staging.Caddyfile`)
- Local development proxy on port 8083
- API routing with `/v1/*` prefix
- Public routing with `/{appSlug}/*` pattern
- Rate limiting (relaxed for testing)
- Health check endpoints

## Database Schema

The staging environment includes the complete production schema:
- **25+ tables** across all domains
- **User management** (tenants, users, authentication)
- **Application registry** (apps, deployments, functions)
- **Billing system** (pricing, invoices, payments)
- **Security auditing** (events, monitoring)
- **Content management** (blogs, documentation)
- **Archive system** (compliance data retention)

### Migration Status
- ✅ All migrations applied successfully
- ✅ Database encryption initialized
- ✅ Foreign key constraints validated
- ✅ Indexes created for performance

## Usage Guide

### Starting the Environment

#### Option 1: Docker Compose (Recommended)
```bash
# Start all staging services
make staging-up

# View logs
make staging-logs

# Check status
make staging-status

# Stop environment
make staging-down
```

#### Option 2: Local Development
```bash
# Run API server directly
make staging-api

# Run health monitor
make staging-health-monitor

# Connect to database
make staging-psql
```

### Database Operations

#### Run Migrations
```bash
make staging-migrate
```

#### Database Connection
```bash
# Direct psql connection
make staging-psql

# Or manually:
psql "postgresql://functionfly_owner:npg_YzCkZWNy97Dv@ep-lucky-bird-aie8580h.c-4.us-east-1.aws.neon.tech/functionfly?sslmode=require"
```

### Testing Endpoints

#### API Access
```bash
# Health check
curl http://localhost:8082/health

# API endpoints
curl http://localhost:8082/v1/apps
```

#### Web Interface
```bash
# Caddy proxy
curl http://localhost:8083/health
```

## Environment Comparison

| Component | Development | Staging | Production |
|-----------|-------------|---------|------------|
| Database | Local Postgres | Neon staging branch | Neon production branch |
| Port (API) | 8080 | 8082 | 8080 |
| Port (Web) | 8081 | 8083 | 8443 |
| Redis DB | 0 | 1 | 0 |
| Rate Limits | 100/min | 200/min | 100/min |
| SSL | Disabled | Disabled | Required |
| Monitoring | Basic | Full | Full |

## Security Configuration

### Authentication & Authorization
- JWT-based authentication (staging-specific secrets)
- Advanced security middleware enabled
- Rate limiting (relaxed for testing)
- SQL injection protection
- XSS filtering
- Path traversal protection

### Data Protection
- Database encryption with master/data keys
- SSL/TLS encryption for database connections
- Secure password hashing
- Audit logging enabled

## Monitoring & Observability

### Health Checks
- Application health endpoint: `/health`
- Database connectivity checks
- Redis connectivity validation
- Container health monitoring

### Logging
- Structured JSON logging
- Docker container logs
- Application-specific log levels
- Error tracking and alerting

### Metrics
- Database connection pooling metrics
- Request/response metrics
- Security event monitoring
- Performance monitoring

## Deployment Workflow

### Typical Staging Workflow
1. **Feature Development**: Develop on local development environment
2. **Staging Deployment**: Deploy to staging for integration testing
3. **Validation**: Run automated tests and manual QA
4. **Production Deployment**: Promote validated changes to production

### Branch Management
```bash
# Create feature branch from staging
neon branches create --name feature-x --parent staging

# Promote staging to production (when ready)
neon branches promote staging

# Cleanup old branches
neon branches delete feature-x
```

## Troubleshooting

### Common Issues

#### Database Connection Failed
```bash
# Check Neon branch status
neon branches list --project-id billowing-rice-16808528

# Test connection manually
make staging-psql
```

#### Port Conflicts
```bash
# Check what's using ports
lsof -i :8082
lsof -i :8083

# Kill conflicting processes or change ports in .env.staging
```

#### Migration Errors
```bash
# Check migration status
make staging-migrate

# Reset if needed (CAUTION: destroys data)
neon branches reset staging
```

### Logs and Debugging
```bash
# View all logs
make staging-logs

# View specific service logs
docker-compose -f docker-compose.staging.yml logs orchestrator-api

# Debug database queries
make staging-psql
```

## Cost Optimization

### Neon Branching Benefits
- **Storage**: ~3x more cost-effective than separate databases
- **Compute**: Shared compute limits with production
- **Backup**: Automatic point-in-time recovery included
- **Scaling**: Auto-scaling within branch limits

### Resource Limits
- **Compute Hours**: 0/100 CU-hrs (shared pool)
- **Storage**: 0.5 GB limit
- **Network**: 5 GB transfer limit
- **Retention**: 6 hours history

## Future Enhancements

### Planned Improvements
- [ ] Automated staging deployments (CI/CD)
- [ ] Staging data seeding scripts
- [ ] Performance benchmarking suite
- [ ] Integration test automation
- [ ] Staging environment monitoring dashboard

### Scaling Considerations
- [ ] Multiple staging environments per feature
- [ ] Geographic region testing
- [ ] Load testing capabilities
- [ ] Disaster recovery testing

---

## Quick Reference

```bash
# Start staging
make staging-up

# Run API locally
make staging-api

# Database connection
make staging-psql

# View logs
make staging-logs

# Stop staging
make staging-down
```

**Staging Database**: `postgresql://functionfly_owner:npg_YzCkZWNy97Dv@ep-lucky-bird-aie8580h.c-4.us-east-1.aws.neon.tech/functionfly?sslmode=require`

**API Endpoint**: `http://localhost:8082`

**Web Interface**: `http://localhost:8083`
