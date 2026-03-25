# Trust API Documentation

**Version**: 1.0  
**Status**: Active  
**Base URL**: `https://api.functionfly.com/v1`

The Trust API allows external platform partners to access FunctionFly's verification and trust infrastructure. Other platforms can pay to integrate FunctionFly's trust scoring system into their own applications.

---

## Overview

The Trust API is a B2B2B revenue stream that allows other platforms to:

- Query trust scores for functions
- Submit functions for verification
- Report trust issues with functions
- Track API usage

## Authentication

All Trust API endpoints (except partner registration) require API key authentication.

### API Key Format

API keys are prefixed with `tak_` (Trust API Key):

```
tak_abc123def456...
```

### Authentication Methods

Include your API key in one of the following:

1. **Authorization Header (Recommended)**

   ```
   Authorization: Bearer tak_abc123def456...
   ```

2. **Direct API Key**

   ```
   Authorization: tak_abc123def456...
   ```

---

## Rate Limits

Rate limits are enforced per-partner based on tier:

| Tier | Requests/Minute | Requests/Day | Monthly Limit |
|------|-----------------|--------------|---------------|
| Developer | 60 | 10,000 | 50,000 |
| Startup | 300 | 100,000 | 500,000 |
| Business | 1,000 | 500,000 | 2,000,000 |
| Enterprise | 10,000 | 10,000,000 | 100,000,000 |

Rate limit headers are included in all responses:

- `X-RateLimit-Limit`: Your rate limit
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when the limit resets

---

## Endpoints

### Partner Management

#### Register Partner

```http
POST /v1/partners
Content-Type: application/json

{
  "name": "ACME Corp",
  "slug": "acme-corp",
  "contact_email": "api@acme.com",
  "contact_name": "John Doe",
  "website_url": "https://acme.com",
  "tier": "developer"
}
```

**Response (201 Created)**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "ACME Corp",
  "slug": "acme-corp",
  "contact_email": "api@acme.com",
  "contact_name": "John Doe",
  "website_url": "https://acme.com",
  "tier": "developer",
  "rate_limit_per_minute": 60,
  "rate_limit_per_day": 10000,
  "monthly_request_limit": 50000,
  "current_month_usage": 0,
  "status": "pending",
  "created_at": "2026-03-21T10:00:00Z"
}
```

#### List Partners

```http
GET /v1/partners?status=active&tier=business&page=1&page_size=20
```

#### Get Partner

```http
GET /v1/partners/{partner_id}
```

#### Update Partner

```http
PATCH /v1/partners/{partner_id}
Content-Type: application/json

{
  "name": "ACME Corporation",
  "tier": "business"
}
```

---

### API Key Management

#### Create API Key

```http
POST /v1/partners/{partner_id}/api-keys
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "name": "Production Key",
  "description": "Key for production environment",
  "scopes": ["trust:read", "trust:write", "verification:request"],
  "allowed_ips": ["192.168.1.0/24", "10.0.0.1"],
  "expires_at": "2027-01-01T00:00:00Z"
}
```

**Response (201 Created)** - Note: `key` is only shown once:

```json
{
  "api_key": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "key_id": "tak_abc123...",
    "key_prefix": "tak_abc1",
    "name": "Production Key",
    "description": "Key for production environment",
    "scopes": ["trust:read", "trust:write", "verification:request"],
    "allowed_ips": ["192.168.1.0/24", "10.0.0.1"],
    "expires_at": "2027-01-01T00:00:00Z",
    "is_revoked": false,
    "use_count": 0,
    "created_at": "2026-03-21T10:00:00Z"
  },
  "partner": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "ACME Corp",
    "slug": "acme-corp",
    "tier": "developer",
    "status": "active"
  },
  "message": "Save the API key securely. It will not be shown again."
}
```

#### List API Keys

```http
GET /v1/partners/{partner_id}/api-keys
Authorization: Bearer {api_key}
```

#### Revoke API Key

```http
DELETE /v1/partners/{partner_id}/api-keys/{key_id}
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "reason": "Compromised key rotation"
}
```

---

### Trust Score Endpoints

#### Get Trust Score

```http
GET /v1/trust/score/{function_id}
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "trust_score": 87.5,
  "trust_tier": "verified",
  "is_verified": true,
  "verification_level": "standard",
  "last_updated": "2026-03-21T09:30:00Z",
  "components": {
    "reliability": 92.0,
    "latency": 88.5,
    "error_rate": 85.0,
    "user_rating": 90.0,
    "verification": 100.0
  },
  "metrics": {
    "total_calls": 15420,
    "success_rate": 0.985,
    "p50_latency_ms": 45.2,
    "p95_latency_ms": 120.5,
    "p99_latency_ms": 250.0,
    "error_rate": 0.015,
    "timeout_rate": 0.005
  }
}
```

#### Batch Trust Score

```http
POST /v1/trust/batch
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "function_ids": [
    "550e8400-e29b-41d4-a716-446655440002",
    "550e8400-e29b-41d4-a716-446655440003",
    "550e8400-e29b-41d4-a716-446655440004"
  ]
}
```

**Response (200 OK)**:

```json
{
  "scores": [
    {
      "function_id": "550e8400-e29b-41d4-a716-446655440002",
      "trust_score": 87.5,
      "trust_tier": "verified",
      "is_verified": true,
      "verification_level": "standard",
      "last_updated": "2026-03-21T09:30:00Z",
      "components": {...},
      "metrics": {...}
    }
  ],
  "errors": [
    {
      "function_id": "550e8400-e29b-41d4-a716-446655440003",
      "error": "Trust score not found"
    }
  ]
}
```

#### Get Trust History

```http
GET /v1/trust/history/{function_id}?page=1&page_size=20
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "history": [
    {
      "trust_score": 87.5,
      "trust_tier": "verified",
      "is_verified": true,
      "calculated_at": "2026-03-21T09:00:00Z",
      "window_start": "2026-03-20T09:00:00Z",
      "window_end": "2026-03-21T09:00:00Z"
    },
    {
      "trust_score": 85.0,
      "trust_tier": "verified",
      "is_verified": true,
      "calculated_at": "2026-03-20T09:00:00Z",
      "window_start": "2026-03-19T09:00:00Z",
      "window_end": "2026-03-20T09:00:00Z"
    }
  ],
  "total_count": 30,
  "page": 1,
  "page_size": 20
}
```

---

### Verification Endpoints

#### Submit Verification Request

**Scope Required**: `verification:request`

```http
POST /v1/trust/verify
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "metadata": {
    "use_case": "data_processing",
    "expected_traffic": "high"
  }
}
```

**Response (201 Created)**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "verification_id": "vfy_abc123def456...",
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_author": "alice",
  "function_name": "data-processor",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "status": "pending",
  "created_at": "2026-03-21T10:00:00Z"
}
```

#### Get Verification Status

```http
GET /v1/trust/verify/{verification_id}
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440005",
  "verification_id": "vfy_abc123def456...",
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_author": "alice",
  "function_name": "data-processor",
  "function_version": "1.2.0",
  "verification_level": "standard",
  "status": "completed",
  "trust_score": 92.5,
  "trust_tier": "highly_trusted",
  "verification_badge_url": "https://functionfly.com/badges/vfy_abc123...",
  "created_at": "2026-03-21T10:00:00Z",
  "completed_at": "2026-03-21T12:30:00Z"
}
```

---

### Report Endpoints

#### Submit Trust Report

**Scope Required**: `reports:submit`

```http
POST /v1/trust/report
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "report_type": "malware",
  "severity": "high",
  "title": "Suspicious network behavior",
  "description": "The function appears to be making unexpected outbound connections to unknown servers...",
  "evidence": {
    "outbound_ips": ["203.0.113.50", "203.0.113.51"],
    "timestamps": ["2026-03-21T08:15:00Z"]
  }
}
```

**Response (201 Created)**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440006",
  "report_id": "rpt_abc123def456...",
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "function_author": "alice",
  "function_name": "data-processor",
  "report_type": "malware",
  "severity": "high",
  "title": "Suspicious network behavior",
  "description": "The function appears to be making unexpected outbound connections...",
  "status": "pending",
  "created_at": "2026-03-21T10:00:00Z"
}
```

#### Get Report Status

```http
GET /v1/trust/report/{report_id}
Authorization: Bearer {api_key}
```

---

### Usage Endpoints

#### Get Partner Usage

```http
GET /v1/partners/{partner_id}/usage?start_date=2026-03-01T00:00:00Z&end_date=2026-03-21T23:59:59Z
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "partner_id": "550e8400-e29b-41d4-a716-446655440000",
  "period_start": "2026-03-01T00:00:00Z",
  "period_end": "2026-03-21T23:59:59Z",
  "total_requests": 15234,
  "successful_requests": 14890,
  "failed_requests": 344,
  "average_latency_ms": 52.3,
  "rate_limit_hits": 12,
  "top_endpoints": [
    {"endpoint": "/v1/trust/score", "count": 12500},
    {"endpoint": "/v1/trust/batch", "count": 2100},
    {"endpoint": "/v1/trust/history", "count": 634}
  ]
}
```

---

## Scopes

| Scope | Description |
|-------|-------------|
| `trust:read` | Read trust scores and history |
| `trust:write` | Submit trust reports |
| `verification:request` | Request function verification |
| `reports:submit` | Submit trust issue reports |
| `partners:manage` | Manage partner account (admin only) |

---

## Error Responses

All errors follow a consistent format:

```json
{
  "error": "Human-readable error message",
  "code": "machine_readable_code"
}
```

### Common Error Codes

| HTTP Status | Code | Description |
|-------------|------|-------------|
| 400 | `invalid_request` | Malformed request |
| 400 | `invalid_function_id` | Invalid function ID format |
| 401 | `missing_auth` | No Authorization header |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `partner_inactive` | Partner account not active |
| 403 | `ip_not_allowed` | IP not in allowlist |
| 403 | `insufficient_scope` | Missing required scope |
| 404 | `partner_not_found` | Partner doesn't exist |
| 404 | `trust_not_found` | Trust score not found |
| 429 | `rate_limit_exceeded` | Rate limit exceeded |
| 429 | `quota_exceeded` | Monthly quota exceeded |

---

## Webhooks

Partners can configure a webhook URL to receive notifications about:

- Usage threshold alerts (80%, 90%, 100% of quota)
- Rate limit exceedances
- Verification request completions

Configure webhook URL in partner settings:

```json
{
  "webhook_url": "https://partner.example.com/trust-api-webhook",
  "webhook_secret": "whsec_..."
}
```

---

## SDKs

Official SDKs are available for:

- [JavaScript/TypeScript](https://github.com/functionfly/sdk-js)
- [Python](https://github.com/functionfly/sdk-python)
- [Go](https://github.com/functionfly/sdk-go)

---

## Support

- **Email**: <api-support@functionfly.com>
- **Documentation**: <https://docs.functionfly.com/trust-api>
- **Status Page**: <https://status.functionfly.com>
