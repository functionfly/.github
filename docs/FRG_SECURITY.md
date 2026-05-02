# FRG Security Guide

The Function Registry + Live Runtime Graph (FRG) system provides versioned, composable function graphs with streaming execution. This document outlines the security features and requirements for FRG endpoints.

## Overview

FRG security implements defense-in-depth with multiple layers:

1. **Authentication & Authorization** - JWT-based auth with tenant isolation
2. **Rate Limiting** - Per-tenant rate limiting via AdvancedSecurityMiddleware
3. **Quota Enforcement** - Real-time usage tracking with Redis-backed counters
4. **Input Validation** - Size limits and sanitization
5. **Webhook Security** - HMAC-SHA256 signature verification
6. **Audit Logging** - Prometheus metrics for all operations

## Endpoints

### Public (Read-only)
- `GET /frg/graphs` - List graphs
- `GET /frg/graphs/{author}/{name}` - Get graph details
- `GET /frg/graphs/{author}/{name}/instances` - List instances
- `GET /frg/discover` - Semantic search
- `GET /frg/graphs/{author}/{name}/optimizations` - Get optimizations

### Protected (Auth + Rate Limiting)
- `POST /frg/graphs` - Create graph
- `PUT /frg/graphs/{author}/{name}` - Update graph
- `DELETE /frg/graphs/{author}/{name}` - Delete graph
- `POST /frg/graphs/{author}/{name}/publish` - Publish version
- `POST /frg/graphs/{author}/{name}/remix` - Fork graph
- `POST /gx/{author}/{name}` - Execute graph
- `POST /gx/{author}/{name}@{version}` - Execute specific version
- `POST /frg/instances/{instance_id}/stop` - Stop instance
- `POST /frg/compose` - AI compose
- `POST /frg/functions/generate` - AI generate function

### Webhooks (Public with Signature Verification)
- `POST /webhook/{path:.*}` - Dynamic webhooks
- `POST /api/webhooks/graph/{graph_id}` - Fixed path webhooks

## Authentication

FRG endpoints require JWT authentication via `AuthMiddleware`:

```bash
Authorization: Bearer <jwt_token>
```

The JWT must contain:
- `user_id` - User identifier
- `tenant_id` - Tenant identifier
- `username` - Username for authorship

## Tenant Isolation

Private graphs can only be accessed by their owners or tenant members:

```go
if def.Visibility == "private" {
    if def.OwnerUserID != nil && *def.OwnerUserID != user.UserID {
        if def.TenantID != nil && *def.TenantID != user.TenantID {
            return HTTP 403 Forbidden
        }
    }
}
```

## Rate Limiting

All write and execution endpoints use `AdvancedRateLimit` middleware:

```go
advancedSecurityMiddleware.AdvancedRateLimit(authMiddleware.RequireAuth(frgHandler.ExecuteGraph))
```

Rate limits are per-tenant with IP fallback for unauthenticated requests.

## Quota Enforcement

FRG graph executions are subject to real-time quota enforcement via `RealtimeUsageTracker`:

### Quota Check Flow

1. Before execution, `UsageTracker.RecordExecution()` is called
2. If quota exceeded, returns HTTP 402 with `QUOTA_EXCEEDED` error
3. Quota headers are added to responses

### Quota Response Headers

- `X-Quota-Executions-Used` - Current executions used
- `X-Quota-Executions-Limit` - Execution limit
- `X-Quota-Executions-Percent` - Usage percentage
- `X-Quota-Status` - ok/warning/critical/exceeded

### Quota Exceeded Response (HTTP 402)

```json
{
  "error": {
    "code": "QUOTA_EXCEEDED",
    "message": "Execution quota exceeded: 1000 of 1000 used",
    "type": "quota_exceeded"
  },
  "quota_status": {
    "tenant_id": "uuid",
    "executions_used": 1000,
    "executions_limit": 1000,
    "executions_percent": 100.0,
    "status": "exceeded"
  },
  "upgrade_url": "/settings/billing"
}
```

## Input Size Limits

To prevent memory exhaustion and abuse:

| Limit | Value | Purpose |
|-------|-------|---------|
| Max nodes per graph | 100 | Prevents complex graphs |
| Max edges per graph | 500 | Prevents excessive connections |
| Max graph name length | 100 chars | Prevents long names |
| Max graph JSON size | ~1MB | Memory protection |

## Execution Timeouts

Graph execution has configurable timeouts:

- **Default**: 5 minutes (`defaultExecutionTimeout`)
- **Override**: `?timeout=30s` parameter (capped at default)

```go
const defaultExecutionTimeout = 5 * time.Minute
execTimeout := defaultExecutionTimeout
if timeoutStr := r.URL.Query().Get("timeout"); timeoutStr != "" {
    if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
        if parsedTimeout < defaultExecutionTimeout {
            execTimeout = parsedTimeout
        }
    }
}
execCtx, execCancel := context.WithTimeout(ctx, execTimeout)
```

## Webhook Signature Verification

Webhook endpoints verify HMAC-SHA256 signatures:

```go
// Signature format: HMAC-SHA256(secret, timestamp.body)
signature := hmac.New(sha256.New, []byte(secret))
signature.Write([]byte(timestamp + "." + body))
expected := hex.EncodeToString(signature.Sum(nil))
```

The `FRG_WEBHOOK_SECRET` environment variable configures the secret.

## XSS Sanitization

AI-generated content is sanitized before storage:

```go
func sanitizeString(s string) string {
    s = strings.ReplaceAll(s, "<", "&lt;")
    s = strings.ReplaceAll(s, ">", "&gt;")
    s = strings.ReplaceAll(s, "\"", "&quot;")
    s = strings.ReplaceAll(s, "'", "&#39;")
    return s
}
```

Applied to AI-composed graph descriptions and metadata.

## Audit Metrics

FRG operations are recorded via Prometheus metrics:

### Counters
- `functionfly_frg_graph_executions_total` - Execution count by tenant, graph, operation, status
- `functionfly_frg_quota_exceeded_total` - Quota blocks by tenant
- `functionfly_frg_webhook_signature_failures_total` - Webhook verification failures
- `functionfly_frg_graph_creation_total` - Graph creations by tenant, visibility, status

### Histograms
- `functionfly_frg_graph_execution_duration_seconds` - Execution duration

### Gauges
- `functionfly_frg_graph_active_count` - Active executions by tenant
- `functionfly_frg_quota_usage_percent` - Current quota usage percentage
- `functionfly_frg_graph_nodes_total` - Total nodes across graphs

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `FRG_WEBHOOK_SECRET` | Webhook HMAC secret | (required for webhooks) |
| `AI_SERVICE_URL` | AI composition service URL | `http://localhost:18081` |
| `AI_SERVICE_API_KEY` | AI service authentication | - |

## Security Checklist

Before deploying FRG to production:

- [ ] Set `FRG_WEBHOOK_SECRET` to a secure random value (64 hex chars)
- [ ] Configure rate limiting limits appropriate for your usage
- [ ] Verify quota enforcement is active (`UsageTracker.IsEnabled()`)
- [ ] Review input size limits for your use case
- [ ] Ensure AI service is behind proper authentication
- [ ] Monitor `functionfly_frg_quota_exceeded_total` for quota issues
- [ ] Monitor `functionfly_frg_webhook_signature_failures_total` for attack attempts

## Tenant ID Protection

The FRG system intentionally removes `TenantID` from requests sent to external AI services to prevent information leakage:

```go
// CompositionRequest sent to AI service
type CompositionRequest struct {
    Prompt           string   `json:"prompt"`
    Requirements     []string `json:"requirements,omitempty"`
    PreferredRuntime string   `json:"preferred_runtime"`
    // TenantID intentionally omitted - do not leak tenant identifiers
}
```

## Related Documentation

- [REALTIME_FEATURES_README.md](REALTIME_FEATURES_README.md) - Quota and usage tracking
- [MONITORING.md](MONITORING.md) - Metrics and alerting
- [SECURITY.md](SECURITY.md) - General security features
