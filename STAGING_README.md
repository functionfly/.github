# FunctionFly Staging Environment

This document describes how to set up and run the FunctionFly staging environment.

## Overview

The staging environment is a complete deployment setup that mirrors production but uses staging-specific configurations, databases, and resources. It's designed for:

- Pre-production testing
- Integration testing
- User acceptance testing
- Performance testing

## Architecture

The staging environment consists of:

- **Orchestrator API** (Port 8082) - Main API service
- **Caddy Reverse Proxy** (Port 8083) - Load balancer and routing
- **Redis** (Port 6380) - Artifact storage and caching
- **Health Monitor** - Background service monitoring

## Database

Staging uses a **Neon Postgres** staging branch with the following configuration:

```bash
DB_HOST=ep-lucky-bird-aie8580h.c-4.us-east-1.aws.neon.tech
DB_NAME=neondb
DB_USER=neondb_owner
DB_SSLMODE=require
```

> **⚠️ Security Note**: Update the database credentials in `.env.staging` before deploying!

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Setup

1. **Clone the repository** (if not already done):
   ```bash
   git clone https://github.com/olyntar/functionfly-web.git
   cd functionfly-web
   ```

2. **Configure environment**:
   ```bash
   # Copy staging environment file
   cp .env.staging .env.staging.local

   # Edit with your staging credentials
   nano .env.staging.local
   ```

3. **Update staging configuration**:
   - Database credentials (Neon staging branch)
   - JWT secrets
   - API shared secrets
   - OAuth provider credentials
   - Email SMTP settings

### Running

Start the staging environment:

```bash
./scripts/run-staging.sh
```

Or manually:

```bash
# Load environment variables
export $(grep -v '^#' .env.staging | xargs)

# Start services
docker-compose -f docker-compose.staging.yml up --build -d
```

### Verification

Check that services are running:

```bash
# Health check
curl http://localhost:8082/health

# View logs
docker-compose -f docker-compose.staging.yml logs -f

# Check container status
docker-compose -f docker-compose.staging.yml ps
```

## Configuration

### Environment Variables

Key staging-specific settings in `.env.staging`:

```bash
# Environment
NODE_ENV=staging
ENVIRONMENT=staging

# Database (Neon staging branch)
DB_HOST=ep-lucky-bird-aie8580h.c-4.us-east-1.aws.neon.tech
DB_NAME=neondb

# Ports (different from production)
PORT=8082
REDIS_ADDR=localhost:6380

# Security (staging-specific secrets)
JWT_SECRET=your-staging-jwt-secret-key-change-this-in-production
API_SHARED_SECRET=your-staging-api-shared-secret-change-this-in-production

# Rate limiting (more permissive)
RATE_LIMIT_REQUESTS=200
RATE_LIMIT_WINDOW_SECONDS=60
```

### Docker Services

The staging environment runs these services:

- **orchestrator-api**: Main API server (Go)
- **redis**: In-memory database for artifacts
- **caddy**: Reverse proxy and load balancer
- **health-monitor**: Background monitoring service

## URLs and Ports

- **API Endpoint**: http://localhost:8082
- **Caddy Proxy**: http://localhost:8083
- **Health Check**: http://localhost:8082/health
- **Redis**: localhost:6380 (internal)

## Differences from Production

Staging environment differences:

- **Database**: Uses Neon staging branch instead of production
- **Ports**: Different ports to avoid conflicts
- **Rate Limiting**: More permissive limits for testing
- **Archive Storage**: Disabled to reduce costs
- **Security**: Some security features enabled for testing
- **OAuth**: Uses test/sandbox accounts
- **Email**: Uses external SMTP (Gmail) instead of production service

## Monitoring

### Health Checks

The staging environment includes comprehensive health monitoring:

- Container health checks
- Application health endpoints
- Database connectivity monitoring
- Redis availability checks

### Logs

View logs for all services:

```bash
docker-compose -f docker-compose.staging.yml logs -f
```

View logs for specific service:

```bash
docker-compose -f docker-compose.staging.yml logs -f orchestrator-api
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8082, 8083, 6380 are available
2. **Database connection**: Verify Neon staging credentials
3. **Environment variables**: Ensure `.env.staging` is properly configured
4. **Docker issues**: Try `docker system prune` to clean up

### Reset Environment

To completely reset the staging environment:

```bash
# Stop and remove containers
docker-compose -f docker-compose.staging.yml down -v

# Rebuild from scratch
docker-compose -f docker-compose.staging.yml up --build --force-recreate
```

## Deployment

For production deployment, use `docker-compose.production.yml` instead of the staging configuration.

## Security Considerations

- **Never commit secrets**: `.env.staging` contains sensitive data
- **Use strong passwords**: Generate unique secrets for staging
- **Network isolation**: Staging runs on separate network
- **Access control**: Limit access to staging environment

## Support

For issues with the staging environment:

1. Check the logs: `docker-compose -f docker-compose.staging.yml logs`
2. Verify configuration in `.env.staging`
3. Test individual services: `docker-compose -f docker-compose.staging.yml up <service-name>`
4. Check health endpoints