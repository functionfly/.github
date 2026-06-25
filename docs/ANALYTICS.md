# Analytics Documentation

FunctionFly provides a multi-layered analytics system that tracks function generation, execution, billing, and platform-wide metrics.

## Architecture

The analytics system consists of three main layers:

| Layer | Location | Purpose |
|-------|----------|---------|
| **Factory Analytics** | `internal/analytics/` | Tracks function generation metrics (quality, latency, throughput) |
| **Unified Analytics** | `internal/analytics/unified/` | Aggregates data from all sources per tenant |
| **Admin Analytics** | `internal/api/handlers/admin/analytics.go` | Billing/revenue metrics (MRR, ARR, churn, LTV) |

### Data Flow

```
Factory Runs → Factory Metrics → Aggregations → Dashboard Stats
                                       ↓
Billing/State/Agent/Registry → Unified Analytics → Tenant/Platform Summaries
                                       ↓
                              Analytics Rollups (pre-computed)
```

## Database Tables

### Factory Analytics

| Table | Purpose |
|-------|---------|
| `factory_analytics_metrics` | Raw metric data points |
| `factory_analytics_aggregated` | Pre-computed aggregated statistics |

### Unified Analytics

| Table | Purpose |
|-------|---------|
| `analytics_events` | Canonical event store (executions, state ops, billing, agent, registry) |
| `analytics_rollups` | Pre-computed daily/hourly rollups per tenant |

## Metric Types

### Factory Metrics

| Metric Type | Description | Unit |
|-------------|-------------|------|
| `generation_success` | Successful generation runs | count |
| `generation_failure` | Failed generation runs | count |
| `quality_score` | Average quality score | score (0-100) |
| `test_score` | Average test score | score (0-100) |
| `latency_generation` | Generation phase latency | ms |
| `latency_testing` | Testing phase latency | ms |
| `latency_publishing` | Publishing phase latency | ms |
| `latency_total` | Total run latency | ms |
| `opportunity_scanned` | Opportunities scanned | count |
| `function_published` | Functions published | count |
| `review_required` | Functions requiring review | count |

### Unified Metrics (Rollup Names)

| Rollup Metric | Description |
|---------------|-------------|
| `function_executions` | Function execution count |
| `state_read_ops` | State storage read operations |
| `state_write_ops` | State storage write operations |
| `state_storage_bytes` | State storage size in bytes |
| `billing_quantity` | Billing quantity consumed |
| `agent_calls` | Agent execution count |
| `agent_cost_usd` | Agent execution cost in USD |
| `registry_executions` | Registry function executions |

## API Endpoints

### Factory Analytics

All endpoints require authentication. Base path: `/v1/analytics/`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/dashboard` | GET | Dashboard statistics (default agent) |
| `/dashboard/{agent_id}` | GET | Dashboard statistics for specific agent |
| `/timeseries` | GET | Time series data for a metric |
| `/timeseries/{agent_id}` | GET | Time series for specific agent |
| `/metrics` | GET | List raw metrics with filters |
| `/aggregated` | GET | Pre-computed aggregated metrics |
| `/aggregated/hourly` | GET | Hourly statistics (last 24h) |
| `/aggregated/daily` | GET | Daily statistics (last 7d) |
| `/aggregated/weekly` | GET | Weekly statistics (last 4w) |
| `/aggregated/monthly` | GET | Monthly statistics (last 6m) |
| `/runs/{run_id}` | GET | Metrics for specific run |
| `/runs` | GET | Recent runs with metrics |
| `/percentiles` | GET | P50/P95/P99 percentiles |
| `/success-rate` | GET | Success rate percentage |
| `/error-rate` | GET | Error rate percentage |
| `/throughput` | GET | Throughput (functions/hour) |
| `/latency` | GET | Average latency for type |
| `/quality-trend` | GET | Quality trend percentage |

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `period` | string | `24h` | Time period (e.g., `24h`, `7d`, `30d`) |
| `metric_type` | string | - | Metric type to query |
| `start` | RFC3339 | -24h | Start time |
| `end` | RFC3339 | now | End time |
| `agent_id` | string | current | Agent ID for agent-specific queries |
| `limit` | int | 100 | Result limit |
| `offset` | int | 0 | Result offset |

### Unified Analytics

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/v1/analytics/tenants/{tenantId}/summary` | GET | PermTenantsRead | Tenant summary for time range |
| `/v1/analytics/tenants/{tenantId}/timeseries` | GET | PermTenantsRead | Tenant time series by metric kind |
| `/v1/analytics/platform/summary` | GET | PermSystemRead | Platform-wide summary |

#### Time Series Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `kind` | MetricKind | `executions`, `state_ops`, `billing`, `agent_calls`, `registry_runs` |
| `granularity` | Granularity | `hour`, `day`, `week`, `month` |
| `start` | RFC3339 | Start of range |
| `end` | RFC3339 | End of range |

### Admin Billing Analytics

| Endpoint | Method | Permission | Description |
|----------|--------|------------|-------------|
| `/admin/analytics/mrr` | GET | PermBillingRead | Monthly Recurring Revenue |
| `/admin/analytics/mrr-series` | GET | PermBillingRead | MRR time series |
| `/admin/analytics/arr` | GET | PermBillingRead | Annual Recurring Revenue |
| `/admin/analytics/churn` | GET | PermBillingRead | Churn metrics |
| `/admin/analytics/churn-series` | GET | PermBillingRead | Churn time series |
| `/admin/analytics/ltv` | GET | PermBillingRead | Lifetime Value metrics |
| `/admin/analytics/financial-report` | GET | PermBillingRead | Financial report |
| `/admin/analytics/tax-jurisdiction` | GET | PermBillingRead | Tax jurisdiction report |

### Admin System Analytics

| Endpoint | Method | Permission | Description |
|----------|--------|------------|-------------|
| `/admin/analytics` | GET | PermSystemRead | Analytics settings |
| `PATCH /admin/analytics` | GET | PermSystemWrite | Update analytics settings |

### Public Analytics

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/users/{username}/analytics` | GET | Public user analytics (rate limited) |

### Embed Analytics

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/v1/registry/functions/{author}/{name}/embed/analytics` | GET | Auth + owner | Embed execution analytics |

## Response Models

### DashboardStats

```json
{
  "total_runs": 100,
  "successful_runs": 95,
  "failed_runs": 5,
  "success_rate": 95.0,
  "avg_quality_score": 87.5,
  "avg_test_score": 92.3,
  "quality_trend": 2.5,
  "functions_generated": 150,
  "functions_published": 120,
  "throughput_per_hour": 12.5,
  "avg_generation_latency": 1500.0,
  "avg_testing_latency": 800.0,
  "avg_publishing_latency": 200.0,
  "avg_total_latency": 2500.0,
  "p95_latency": 4500.0,
  "error_rate": 5.0,
  "pending_reviews": 3,
  "review_rate": 2.5,
  "period_start": "2026-06-23T00:00:00Z",
  "period_end": "2026-06-24T00:00:00Z",
  "last_updated": "2026-06-24T12:00:00Z"
}
```

### TenantSummary

```json
{
  "tenant_id": "uuid",
  "start": "2026-06-01T00:00:00Z",
  "end": "2026-06-30T23:59:59Z",
  "generated_at": "2026-06-24T12:00:00Z",
  "function_executions": 15000,
  "state_storage_bytes": 1073741824,
  "state_read_ops": 50000,
  "state_write_ops": 10000,
  "state_active_states": 150,
  "billing_quantity": 2500,
  "agent_calls": 500,
  "agent_cost_usd": 25.50,
  "agent_success_count": 495,
  "agent_error_count": 5,
  "registry_executions": 3000
}
```

### PlatformSummary

```json
{
  "start": "2026-06-01T00:00:00Z",
  "end": "2026-06-30T23:59:59Z",
  "generated_at": "2026-06-24T12:00:00Z",
  "total_tenants_active": 150,
  "total_function_executions": 500000,
  "total_state_read_ops": 2000000,
  "total_state_write_ops": 400000,
  "total_agent_calls": 25000,
  "total_registry_executions": 100000
}
```

## Aggregation Periods

| Period | Retention | Description |
|--------|-----------|-------------|
| Hourly | 24 hours | Aggregated per hour |
| Daily | 7 days | Aggregated per day |
| Weekly | 4 weeks | Aggregated per week |
| Monthly | 6 months | Aggregated per month |

## Recording Metrics

### Buffered Recording (Recommended)

For high-volume metric recording, use buffered recording:

```go
analyticsSvc.RecordMetricBuffered(MetricRecord{
    RunID:       &runID,
    AgentID:     agentID,
    MetricType:  MetricTypeQualityScore,
    MetricValue: 85.5,
    Labels:      map[string]any{"source": "factory"},
})
```

The service buffers metrics and flushes in batches (default: 100 metrics or 5 seconds).

### Immediate Recording

For critical metrics that must not be lost:

```go
analyticsSvc.RecordMetric(ctx, MetricRecord{
    RunID:       &runID,
    AgentID:     agentID,
    MetricType:  MetricTypeGenerationFailure,
    MetricValue: 1,
    Labels:      map[string]any{"error": err.Error()},
})
```

### Recording Factory Runs

Use `RecordFactoryRun` to record a complete run with all metrics:

```go
metrics := RunMetrics{
    Success:              true,
    TotalLatencyMs:       2500,
    GenerationLatencyMs:  1500,
    TestingLatencyMs:     800,
    PublishingLatencyMs:  200,
    AvgQualityScore:      87.5,
    AvgTestScore:         92.3,
    OpportunitiesScanned: 50,
    FunctionsPublished:   3,
    ReviewRequired:       false,
}
analyticsSvc.RecordFactoryRun(ctx, runID, agentID, metrics)
```

## Background Jobs

### Aggregation Job

Runs hourly/daily/weekly/monthly aggregation:

```go
analyticsSvc.RunAggregationJob(ctx)
```

### Cleanup Job

Removes metrics older than retention period:

```go
deleted, err := analyticsSvc.CleanupOldMetrics(ctx, 90) // 90 days retention
```

## Unified Analytics Sync Job

The unified analytics sync job (`internal/api/server.go`) periodically populates `analytics_rollups` from source tables (function_logs, state_usage_metrics, agent_execution_records, etc.).

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ANALYTICS_SYNC_ENABLED` | `true` | Enable sync job |
| `ANALYTICS_SYNC_INTERVAL` | `1h` | Sync interval |

### Service Configuration

```go
unifiedSvc := unified.NewService(db, usageAgg, unified.ServiceConfig{
    UseRollups:  true,
    EventStore:  eventStore,
})
```

When `UseRollups` is enabled, reads first check rollups table for performance.

## Service Initialization

```go
// Create service
analyticsSvc := analytics.NewService(db, analytics.DefaultServiceConfig(factoryConfig.AgentID))

// Run migrations
analyticsSvc.AutoMigrate(ctx)

// Start background workers
// ... server runs ...

// Stop gracefully
analyticsSvc.Stop()
```

## Example: Dashboard API Call

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/v1/analytics/dashboard?period=24h"
```

## Example: Time Series API Call

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/v1/analytics/timeseries?metric_type=quality_score&period=daily"
```

## Example: Tenant Summary API Call

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/v1/analytics/tenants/$TENANT_ID/summary?start=2026-06-01T00:00:00Z&end=2026-06-30T23:59:59Z"
```

## Example: Tenant Time Series API Call

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/v1/analytics/tenants/$TENANT_ID/timeseries?kind=executions&granularity=day"
```

## Related Documentation

- [API.md](./API.md) - General API documentation
- [MONITORING.md](./MONITORING.md) - Infrastructure monitoring
- [FUNCTION_DNA_ARCHITECTURE.md](./FUNCTION_DNA_ARCHITECTURE.md) - Factory/DNA system
