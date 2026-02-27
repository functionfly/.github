# FunctionFly Monitoring Guide

FunctionFly includes comprehensive monitoring capabilities using Supabase's built-in tools for real-time observability, alerting, and performance tracking.

## Overview

The monitoring system consists of:

- **Real-time Metrics**: Performance metrics with Server-Sent Events (SSE) streaming
- **Health Monitoring**: Automated health checks for all services and backends
- **Alerting System**: Automatic alert generation based on configurable rules
- **Structured Logging**: Request tracing with correlation IDs
- **Dashboard**: Web-based monitoring interface

## Monitoring Dashboard

Access the monitoring dashboard at: `http://localhost:8080/monitoring/dashboard`

The dashboard provides real-time visibility into:

- System health status
- Service availability
- Active alerts
- Recent activity
- Performance metrics

## API Endpoints

### Health Checks

- `GET /health` - Basic health check
- `GET /health/detailed` - Comprehensive health status
- `GET /health/check?name={service}` - Individual service health check

### Metrics

- `GET /v1/metrics/global` - Global system metrics
- `GET /v1/metrics/stream` - Real-time metrics stream (SSE)

### Monitoring

- `GET /v1/monitoring/metrics` - Performance metrics
- `GET /v1/monitoring/alerts` - Active alerts
- `GET /v1/monitoring/health` - System health status
- `GET /v1/monitoring/events` - Monitoring events
- `GET /v1/monitoring/dashboard` - Monitoring dashboard

## Supabase Integration

FunctionFly leverages Supabase's built-in monitoring tools:

### Real-time Subscriptions

- Automatic broadcasting of alerts and events via PostgreSQL LISTEN/NOTIFY
- Real-time dashboard updates using Server-Sent Events
- Tenant-specific and global monitoring channels

### Audit Logging

- All user actions logged in `audit_events` table
- Request tracing with correlation IDs
- Structured logging with contextual information

### Performance Monitoring

- Metrics stored in dedicated tables with optimized indexing
- Automated metric collection and alerting
- Historical performance data retention

## Alerting Rules

The system includes built-in alerting rules:

### High Error Rate

- **Trigger**: Error rate > 5% for 5+ minutes
- **Severity**: Warning
- **Cooldown**: 10 minutes

### Backend Unavailable

- **Trigger**: Backend unhealthy for 2+ minutes
- **Severity**: Error
- **Cooldown**: 5 minutes

### Circuit Breaker Open

- **Trigger**: Circuit breaker opens due to failures
- **Severity**: Warning
- **Cooldown**: 15 minutes

### High Latency

- **Trigger**: Average response time > 1000ms for 3+ minutes
- **Severity**: Warning
- **Cooldown**: 10 minutes

## Structured Logging

All logs include structured fields for better observability:

```json
{
  "request_id": "req-123456789",
  "tenant_id": "uuid",
  "app_id": "uuid",
  "method": "GET",
  "path": "/api/apps",
  "status_code": 200,
  "duration_ms": 150,
  "user_agent": "Mozilla/5.0...",
  "remote_addr": "192.168.1.1"
}
```

## Configuration

### Environment Variables

- `RATE_LIMIT_REQUESTS` - Rate limit per minute (default: 100)
- `RATE_LIMIT_WINDOW_SECONDS` - Rate limit window (default: 60)
- `API_SHARED_SECRET` - HMAC signing secret

### Database Tables

- `performance_metrics` - Performance measurements
- `alerts` - Alert records
- `system_health_checks` - Health check results
- `monitoring_events` - Real-time events
- `audit_events` - User action audit log

## Development

### Running the Monitor

```bash
make docker-up  # Start database
go run cmd/health-monitor/main.go
```

### Testing Monitoring

```bash
# Health checks
curl http://localhost:8080/health
curl http://localhost:8080/health/detailed

# Metrics
curl http://localhost:8080/v1/metrics/global

# Dashboard
open http://localhost:8080/monitoring/dashboard
```

## Supabase Dashboard

Access Supabase monitoring features:

1. Go to your Supabase project dashboard
2. Navigate to Database > Logs
3. View real-time queries, connections, and performance
4. Monitor database health and usage

## Troubleshooting

### Common Issues

1. **Dashboard not loading**
   - Check API server is running on port 8080
   - Verify CORS settings

2. **Real-time updates not working**
   - Check Server-Sent Events support
   - Verify monitoring service is initialized

3. **Alerts not firing**
   - Check alert rules configuration
   - Verify metric collection is working

4. **Logs not structured**
   - Ensure logging middleware is applied
   - Check request ID generation

## Security Considerations

- All monitoring endpoints include rate limiting
- HMAC signing required for sensitive operations
- Request IDs prevent log correlation attacks
- Audit logs track all administrative actions
