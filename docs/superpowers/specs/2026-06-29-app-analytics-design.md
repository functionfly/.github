# App Analytics Design Spec

## Overview

Replace the "Coming Soon" placeholder in `AppDetailPage` AnalyticsTab with a production analytics dashboard showing request counts, error rates, latency charts, backend comparison, and function execution metrics.

## API

**Endpoint:** `GET /v1/apps/{appId}/analytics?days=7`

**Response:**
```json
{
  "summary": {
    "totalRequests": 1234,
    "avgLatencyMs": 45.2,
    "p95LatencyMs": 120,
    "p99LatencyMs": 250,
    "errorRate": 0.02,
    "successRate": 0.98,
    "totalExecutions": 5678,
    "totalCostCents": 432
  },
  "requestsOverTime": [
    { "timestamp": "2026-06-29T00:00:00Z", "total": 100, "success": 98, "errors": 2 }
  ],
  "latencyOverTime": [
    { "timestamp": "2026-06-29T00:00:00Z", "avgMs": 45, "p50Ms": 40, "p95Ms": 120, "p99Ms": 250 }
  ],
  "topErrors": [
    { "statusCode": 500, "count": 15 }
  ],
  "backendBreakdown": [
    { "backendId": "...", "provider": "workers", "requests": 800, "avgLatencyMs": 40, "errorRate": 0.01 }
  ]
}
```

## Data Sources

1. **routing_events** (direct app_id) — request counts, latency, outcomes
2. **execution_metrics** (via functions.app_id) — function-level invocations
3. **cost_allocation_entries** (via functions.app_id) — cost data

## Frontend Components

- Stat cards: Total Requests, Avg Latency, Error Rate, Success Rate
- Requests Over Time: AreaChart (success/error stacked)
- Latency Distribution: LineChart (p50/p95/p99)
- Error Breakdown: BarChart by status code
- Backend Comparison: BarChart
- Time range toggle: 24h / 7d / 30d

## Files Changed

### Backend
- `internal/api/handlers/apps/apps.go` — new HandleGetAppAnalytics method
- `internal/api/types/types.go` — new analytics response types
- `internal/api/routes_platform.go` — register new route
- `internal/storage/interfaces.go` — new analytics storage methods
- `internal/storage/backend_repository.go` — implement routing_events queries
- `internal/storage/app_analytics_repository.go` — new file for app analytics queries

### Frontend
- `web/dashboard/src/api/apps.ts` — new getAnalytics method
- `web/dashboard/src/hooks/useApps.ts` — new useAppAnalytics hook
- `web/dashboard/src/pages/AppDetailPage/index.tsx` — replace AnalyticsTab
- `web/dashboard/src/types/index.ts` — new analytics types
