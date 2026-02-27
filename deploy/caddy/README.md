# FunctionFly Caddy Configuration

This directory contains the Caddy configuration for FunctionFly's edge proxy layer.

## Overview

Caddy serves as the edge-facing reverse proxy that handles:

- **TLS Termination**: Automatic HTTPS with Let's Encrypt certificates
- **Public Routing**: Routes `/{appSlug}/*` requests to the orchestrator for backend selection
- **Rate Limiting**: Per-app and global rate limiting
- **Health Checks**: Circuit breaker equivalent with fail-fast on unhealthy backends
- **Proxying**: Forwards requests to selected backends with retry logic

## Configuration

### Configuration Files

There are three Caddyfile configurations:

- **`Caddyfile`**: Production configuration with automatic HTTPS for `functionfly.com`
- **`staging.Caddyfile`**: Staging configuration with relaxed rate limits
- **`local.Caddyfile`**: Local development configuration without HTTPS

### Environment Variables

- **`CADDY_CONFIG`**: Controls which Caddyfile to use (`local`, `staging`, or `production`)
  - Default: `local` (for development)
  - Set to `staging` for staging deployments
  - Set to `production` for production (uses main Caddyfile)

### Routing Rules

All configurations define:

- **Rate Limits**:
  - Global: 100 requests/minute per IP (production/staging)
  - Per App: 1000-2000 requests/minute per app slug
  - API: 100 requests/minute per IP for `/v1/*` routes
- **Health Checks**: 30s interval, 10s timeout, 30s fail duration
- **Retries**: 2 retries on 502/503/504 or network errors

### Environment Variables

The setup expects these services to be available:

- `orchestrator-api:8080` - The FunctionFly orchestrator API service

## Deployment

### Using Docker Compose

The Caddy service is included in the main `docker-compose.yml`. To run:

```bash
# Start all services including Caddy (uses local config by default)
docker-compose up -d

# Or start just the Caddy service
docker-compose up -d caddy

# Use staging configuration
CADDY_CONFIG=staging docker-compose up -d caddy

# Use production configuration
CADDY_CONFIG=production docker-compose up -d caddy
```

### Manual Deployment

Build and run the Caddy container:

```bash
# Build with local config (default)
docker build -t functionfly-caddy ./deploy/caddy

# Build with staging config
docker build --build-arg CADDY_CONFIG=staging -t functionfly-caddy ./deploy/caddy

# Build with production config
docker build --build-arg CADDY_CONFIG=Caddyfile -t functionfly-caddy ./deploy/caddy

# Run the container
docker run -d \
  --name functionfly-caddy \
  -p 8083:8083 \
  -v caddy_data:/data \
  --link orchestrator-api \
  functionfly-caddy
```

## DNS Configuration

Point your domain (`your-domain.com` in the Caddyfile) to the server running Caddy.

## Request Flow

1. User requests `https://your-domain.com/my-app/api/endpoint`
2. Caddy extracts `my-app` as the app slug
3. Applies rate limiting
4. Forwards to `orchestrator-api:8080` with `X-App-Slug: my-app`
5. Orchestrator selects the best backend
6. Caddy proxies the request to the selected backend
7. On failure, retries to failover backends (for idempotent methods)

## Health Checks

- Caddy health: `http://localhost/health`
- Container health check runs every 30s

## Monitoring

Caddy logs are output in JSON format to stdout. Monitor for:

- Rate limit hits: Look for `rate_limit` events
- Upstream errors: `reverse_proxy` error events
- TLS certificate renewals: `tls` events

## Customization

### Adjusting Rate Limits

Edit the `rate_limit` blocks in `Caddyfile`:

```caddyfile
rate_limit {
    zone per_app {
        key {path.0}
        window 1m
        events 2000  # Increase to 2000 requests/minute
    }
}
```

### Adding Domains

Add additional domains by duplicating the site block:

```caddyfile
another-domain.com {
    # Same configuration as your-domain.com
    # ...
}
```

### Custom TLS Certificates

For production, consider using custom certificates:

```caddyfile
your-domain.com {
    tls /path/to/cert.pem /path/to/key.pem
    # ... rest of config
}
```

## Troubleshooting

### Common Issues

1. **Caddy won't start**: Check if ports 80/443 are available
2. **TLS certificate issues**: Ensure DNS is properly configured
3. **Connection to orchestrator fails**: Verify `orchestrator-api` service is running and accessible

### Logs

View Caddy logs:

```bash
docker-compose logs -f caddy
```

### Testing

Test the setup:

```bash
# Health check
curl http://localhost/health

# API endpoint (if orchestrator is running)
curl http://localhost/v1/health
```