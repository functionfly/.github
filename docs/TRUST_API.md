# Trust API Documentation

**Version**: 1.1
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
- Manage function attestations and trust policies
- Configure webhooks for real-time notifications
- Stream trust score updates via SSE

---

## Authentication

All Trust API endpoints (except partner registration) require API key authentication.

### API Key Format

API keys are prefixed with `fft_` (Trust API Key) and use a versioned format:

```
fft_v1_<32_hex_characters>_<2_hex_checksum>
```

Example:

```
fft_v1_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6_a1b2
```

The checksum is a CRC8 of the prefix + version + random hex for quick validation.

### Authentication Methods

Include your API key in one of the following:

1. **Authorization Header (Recommended)**

   ```
   Authorization: Bearer fft_abc123def456...
   ```

2. **Direct API Key**

   ```
   Authorization: fft_abc123def456...
   ```

### Internal Endpoints

Some endpoints (partner management, webhook management, revocation management) require JWT authentication. These endpoints are protected by the internal auth middleware and are not accessible via API key alone.

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

## Pricing

The Trust API uses a tiered pricing model with optional overage billing:

| Tier | Monthly Price | Included Requests | Overage |
|------|--------------|-------------------|---------|
| Developer | Free | 50,000/month | Hard stop at limit |
| Startup | $49/month | 500,000/month | $0.005/request |
| Business | $199/month | 2,000,000/month | $0.003/request |
| Enterprise | Custom | Custom | Custom contracts |

**Founder Mode**: New partners can enroll in Founder Mode for free access with 100,000 requests/month for 90 days.

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
    "key_id": "fft_abc123...",
    "key_prefix": "fft_abc1",
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

### Trust Revocation Endpoints

Revocation endpoints allow partners to check if functions have been revoked. Creating revocations requires admin/internal authentication.

#### List Revocations

```http
GET /v1/trust/revoke/revoked?status=active&page=1&page_size=20
Authorization: Bearer {internal_jwt}
```

**Response (200 OK)**:

```json
{
  "revocations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "revocation_id": "rvk_abc123...",
      "function_id": "550e8400-e29b-41d4-a716-446655440002",
      "function_author": "alice",
      "function_name": "data-processor",
      "reason": "security",
      "reason_details": "Critical vulnerability found in dependencies",
      "severity": "critical",
      "status": "active",
      "revocation_type": "full",
      "revoked_at": "2026-03-21T10:00:00Z",
      "revoked_by": "admin-user-id",
      "revoked_by_type": "admin",
      "original_trust_score": 87.5,
      "original_trust_tier": "verified"
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 20
}
```

#### Check Function Revocation Status

```http
GET /v1/trust/revoke/revoked/{function_id}
Authorization: Bearer {api_key}
```

**Response (200 OK)** - Function is revoked:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "is_revoked": true,
  "revocation_id": "rvk_abc123...",
  "reason": "security",
  "severity": "critical",
  "revoked_at": "2026-03-21T10:00:00Z",
  "revocation_type": "full",
  "impact_description": "Function trust score has been reset to 0"
}
```

**Response (200 OK)** - Function is NOT revoked:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "is_revoked": false
}
```

#### Get Revocation Details

```http
GET /v1/trust/revoke/{revocation_id}
Authorization: Bearer {internal_jwt}
```

---

### Attestation Endpoints

Attestations provide cryptographic proof of function properties verified by FunctionFly or trusted partners.

#### List Attestations for Function

```http
GET /v1/trust/attestations?function_id={function_id}&type=verification&status=valid&page=1&page_size=20
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "attestations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440020",
      "attestation_id": "att_abc123...",
      "function_id": "550e8400-e29b-41d4-a716-446655440002",
      "function_version": "1.2.0",
      "function_author": "alice",
      "function_name": "data-processor",
      "type": "verification",
      "status": "valid",
      "title": "Standard Security Verification",
      "description": "Function passed standard security scanning",
      "attester_id": "system",
      "attester_type": "system",
      "attester_name": "FunctionFly Security",
      "verification_level": "standard",
      "proof_hash": "sha256:abc123...",
      "attested_at": "2026-03-21T10:00:00Z",
      "valid_until": "2027-03-21T10:00:00Z"
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 20
}
```

#### Get Attestation

```http
GET /v1/trust/attestations/{attestation_id}
Authorization: Bearer {api_key}
```

#### Verify Attestation Integrity

```http
GET /v1/trust/attestations/{attestation_id}/verify
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "attestation_id": "att_abc123...",
  "integrity_verified": true,
  "verified_at": "2026-03-21T10:00:00Z"
}
```

#### Get Attestation Chain

```http
GET /v1/trust/attestations/{function_id}/chain
Authorization: Bearer {api_key}
```

**Response (200 OK)**:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "chain_length": 3,
  "attestations": [
    {
      "attestation_id": "att_abc123...",
      "type": "security_scan",
      "title": "Security Scan Passed",
      "status": "valid",
      "attester_type": "system",
      "attested_at": "2026-03-21T10:00:00Z",
      "proof_hash": "sha256:abc123...",
      "previous_hash": "sha256:def456...",
      "integrity_verified": true
    }
  ]
}
```

---

### Trust Policy Endpoints

Trust policies define rules for evaluating function trust. Partners can create custom policies and evaluate functions against them.

#### Create Policy

```http
POST /v1/trust/policies
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "name": "High Security Policy",
  "description": "Require high trust scores and verification for sensitive operations",
  "default_action": "deny",
  "rules": [
    {
      "id": "min_score",
      "type": "min_trust_score",
      "value": 80.0,
      "description": "Minimum trust score of 80"
    },
    {
      "id": "require_verification",
      "type": "verification_required",
      "value": true,
      "description": "Function must be verified"
    },
    {
      "id": "min_tier",
      "type": "tier_minimum",
      "value": "verified",
      "description": "Minimum tier must be verified"
    },
    {
      "id": "no_revocation",
      "type": "no_revocation",
      "value": true,
      "description": "Function must not be revoked"
    }
  ],
  "valid_until": "2027-12-31T23:59:59Z"
}
```

#### List Policies

```http
GET /v1/trust/policies?status=active&page=1&page_size=20
Authorization: Bearer {internal_jwt}
```

#### Get Policy

```http
GET /v1/trust/policies/{policy_id}
Authorization: Bearer {api_key}
```

#### Update Policy

```http
PUT /v1/trust/policies/{policy_id}
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "name": "Updated High Security Policy",
  "rules": [
    {
      "id": "min_score",
      "type": "min_trust_score",
      "value": 85.0
    }
  ]
}
```

#### Delete Policy

```http
DELETE /v1/trust/policies/{policy_id}
Authorization: Bearer {internal_jwt}
```

#### Evaluate Policy Against Function

```http
POST /v1/trust/policies/evaluate
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "policy_id": "pol_abc123..."
}
```

**Response (200 OK)**:

```json
{
  "result": {
    "evaluation_id": "eval_abc123...",
    "policy_id": "pol_abc123...",
    "function_id": "550e8400-e29b-41d4-a716-446655440002",
    "function_author": "alice",
    "function_name": "data-processor",
    "result": "allowed",
    "decision": "policy_rule_passed",
    "reason": "Rule 'min_score' passed",
    "trust_score": 87.5,
    "trust_tier": "verified",
    "is_verified": true,
    "is_revoked": false,
    "rule_results": [
      {
        "rule_id": "min_score",
        "type": "min_trust_score",
        "passed": true,
        "reason": "",
        "actual_value": 87.5,
        "expected_value": 80.0
      },
      {
        "rule_id": "require_verification",
        "type": "verification_required",
        "passed": true,
        "actual_value": true,
        "expected_value": true
      }
    ],
    "evaluated_at": "2026-03-21T10:00:00Z",
    "cache_valid_until": "2026-03-21T10:05:00Z"
  },
  "cached": false
}
```

#### Batch Evaluate Policy

```http
POST /v1/trust/policies/evaluate/batch
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "function_ids": [
    "550e8400-e29b-41d4-a716-446655440002",
    "550e8400-e29b-41d4-a716-446655440003"
  ],
  "policy_id": "pol_abc123..."
}
```

**Response (200 OK)**:

```json
{
  "results": [
    {
      "function_id": "550e8400-e29b-41d4-a716-446655440002",
      "function_author": "alice",
      "function_name": "data-processor",
      "policy_id": "pol_abc123...",
      "result": "allowed",
      "decision": "batch_quick_eval",
      "trust_score": 87.5,
      "trust_tier": "verified",
      "is_verified": true,
      "is_revoked": false,
      "evaluated_at": "2026-03-21T10:00:00Z"
    }
  ],
  "errors": [
    {
      "function_id": "550e8400-e29b-41d4-a716-446655440003",
      "error": "Function not found"
    }
  ],
  "evaluated_at": "2026-03-21T10:00:00Z"
}
```

---

### Webhook Management Endpoints

Partners can configure webhooks to receive real-time notifications about trust events.

#### Create Webhook

```http
POST /v1/webhooks
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "name": "Trust Notifications",
  "description": "Receive notifications for trust score changes",
  "url": "https://partner.example.com/webhooks/trust",
  "events": [
    "trust.score.updated",
    "trust.revocation.created",
    "trust.verification.completed"
  ],
  "function_filter": {
    "authors": ["alice", "bob"],
    "tags": ["production"]
  },
  "max_retries": 3,
  "secret": "whsec_your_secret_here"
}
```

**Response (201 Created)**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440030",
  "webhook_id": "wh_abc123...",
  "name": "Trust Notifications",
  "description": "Receive notifications for trust score changes",
  "url": "https://partner.example.com/webhooks/trust",
  "method": "POST",
  "events": ["trust.score.updated", "trust.revocation.created", "trust.verification.completed"],
  "status": "active",
  "max_retries": 3,
  "created_at": "2026-03-21T10:00:00Z",
  "updated_at": "2026-03-21T10:00:00Z"
}
```

#### List Webhooks

```http
GET /v1/webhooks?status=active&page=1&page_size=20
Authorization: Bearer {internal_jwt}
```

#### Get Webhook

```http
GET /v1/webhooks/{webhook_id}
Authorization: Bearer {internal_jwt}
```

#### Update Webhook

```http
PUT /v1/webhooks/{webhook_id}
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "name": "Updated Trust Notifications",
  "events": ["trust.score.updated", "trust.revocation.created"],
  "status": "active"
}
```

#### Delete Webhook

```http
DELETE /v1/webhooks/{webhook_id}
Authorization: Bearer {internal_jwt}
```

#### Test Webhook

```http
POST /v1/webhooks/{webhook_id}/test
Authorization: Bearer {internal_jwt}
Content-Type: application/json

{
  "event_type": "trust.score.updated",
  "test_data": {
    "message": "This is a test webhook"
  }
}
```

#### List Webhook Deliveries

```http
GET /v1/webhooks/{webhook_id}/deliveries?status=success&page=1&page_size=20
Authorization: Bearer {internal_jwt}
```

#### Get Webhook Delivery Stats

```http
GET /v1/webhooks/{webhook_id}/stats
Authorization: Bearer {internal_jwt}
```

---

### Real-time Streaming Endpoints (SSE)

Partners can stream trust score updates in real-time using Server-Sent Events (SSE).

#### Stream All Watched Function Updates

```http
GET /v1/trust/stream/sse
Authorization: Bearer {internal_jwt}
```

#### Stream Specific Function Updates

```http
GET /v1/trust/stream/functions/{function_id}/sse
Authorization: Bearer {internal_jwt}
```

**SSE Event Format**:

```
event: trust_score_update
data: {"function_id":"550e8400-e29b-41d4-a716-446655440002","trust_score":87.5,"trust_tier":"verified","is_verified":true,"updated_at":"2026-03-21T10:00:00Z"}

event: revocation_created
data: {"function_id":"550e8400-e29b-41d4-a716-446655440002","revocation_id":"rvk_abc123...","reason":"security","severity":"critical"}
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

## Webhook Events

Partners can subscribe to the following webhook events:

| Event | Description |
|-------|-------------|
| `trust.score.updated` | Trust score changed for a function |
| `trust.revocation.created` | New trust revocation created |
| `trust.revocation.lifted` | Trust revocation was lifted |
| `trust.verification.completed` | Verification request completed |
| `trust.report.submitted` | New trust report submitted |
| `policy.evaluation.created` | Policy evaluation was performed |

### Webhook Payload Format

```json
{
  "event": "trust.revocation.created",
  "timestamp": "2026-03-21T10:00:00Z",
  "webhook_id": "wh_abc123...",
  "data": {
    "revocation_id": "rvk_abc123...",
    "function_id": "550e8400-e29b-41d4-a716-446655440002",
    "function_author": "alice",
    "function_name": "data-processor",
    "reason": "security",
    "severity": "critical",
    "revoked_by": "admin-user-id"
  }
}
```

---

## Trust Tiers

| Tier | Description |
|------|-------------|
| `untrusted` | Function has no trust score or has been explicitly untrusted |
| `trusted` | Function has a basic trust score |
| `verified` | Function has passed verification |
| `highly_trusted` | Function has an excellent trust score and multiple attestations |

---

## Policy Rule Types

| Rule Type | Description | Value Type |
|-----------|-------------|------------|
| `min_trust_score` | Minimum trust score required | float64 (0-100) |
| `verification_required` | Function must be verified | bool |
| `tier_minimum` | Minimum trust tier required | string (untrusted/trusted/verified/highly_trusted) |
| `no_revocation` | Function must not be revoked | bool |
| `min_success_rate` | Minimum success rate required | float64 (0-1) |

---

## Attestation Types

| Type | Description |
|------|-------------|
| `verification` | Function passed FunctionFly verification |
| `security_scan` | Function passed security scanning |
| `code_review` | Function code was reviewed |
| `execution` | Function was tested in execution environment |
| `compliance` | Function meets compliance requirements |
| `signature` | Function has cryptographic signature from author |

---

## Revocation Reasons

| Reason | Description |
|--------|-------------|
| `security` | Security vulnerability discovered |
| `malware` | Function contains malware |
| `abuse` | Function is being abused |
| `policy_violation` | Function violates FunctionFly policies |
| `reported` | Function was reported by a partner |
| `deprecated` | Function has been deprecated |
| `other` | Other reason |

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
| 400 | `invalid_partner_id` | Invalid partner ID format |
| 401 | `missing_auth` | No Authorization header |
| 401 | `invalid_api_key` | API key is invalid or revoked |
| 403 | `partner_inactive` | Partner account not active |
| 403 | `ip_not_allowed` | IP not in allowlist |
| 403 | `insufficient_scope` | Missing required scope |
| 403 | `forbidden` | Not authorized to access resource |
| 404 | `partner_not_found` | Partner doesn't exist |
| 404 | `trust_not_found` | Trust score not found |
| 404 | `policy_not_found` | Policy not found |
| 404 | `revocation_not_found` | Revocation not found |
| 404 | `attestation_not_found` | Attestation not found |
| 404 | `webhook_not_found` | Webhook not found |
| 409 | `already_revoked` | Function already has active revocation |
| 409 | `slug_conflict` | Partner slug already in use |
| 409 | `email_conflict` | Email already registered |
| 429 | `rate_limit_exceeded` | Rate limit exceeded |
| 429 | `quota_exceeded` | Monthly quota exceeded |
| 500 | `internal_error` | Internal server error |

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
