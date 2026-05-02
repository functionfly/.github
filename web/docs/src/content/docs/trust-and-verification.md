---
title: "Trust and Verification"
description: "How to use FunctionFly's trust system to verify AI tools, models, and functions"
---

FunctionFly's trust and verification system helps you evaluate the reliability, security, and quality of AI tools, models, and functions - whether you're building them yourself or using third-party ones.

## What is the Trust System?

The trust system provides:

- **Trust Scores** (0-100) - A composite score evaluating functions across multiple dimensions
- **Trust Tiers** - Categorical rankings (Highly Trusted → Untrusted) for quick assessment
- **Verification** - Security checks including malware scanning, signature verification, and determinism testing
- **Attestations** - Cryptographic proof of function properties
- **Revocation Tracking** - Stay informed when functions are no longer trustworthy

## Who Should Use This?

- **Developers building AI tools** - Get your functions verified to build trust with users
- **AI agents** - Evaluate which functions to use based on trust scores
- **Platform operators** - Enforce trust requirements for function execution
- **End users** - Understand the reliability of AI tool integrations

## Trust Score Components

Functions are evaluated across five dimensions:

| Component | Weight | Description |
|-----------|--------|-------------|
| **Reliability** | 30% | Historical execution success rate |
| **Latency** | 20% | Performance consistency and speed |
| **Error Rate** | 20% | Frequency of execution errors |
| **User Rating** | 15% | Aggregate user satisfaction scores |
| **Verification** | 15% | Security verification bonus (0-15 points) |

## Trust Tiers

| Tier | Score Range | Description |
|------|-------------|-------------|
| **Highly Trusted** | 90-100 | Excellent score + verified |
| **Verified** | 70-89 | Verified or high score |
| **Trusted** | 50-69 | Basic trust established |
| **Untrusted** | <50 | Low or no trust; use with caution |

## Verification Levels

Higher verification levels provide greater trust bonuses:

| Level | Bonus | Turnaround | What's Checked |
|-------|-------|------------|----------------|
| **None** | 0 | - | No verification |
| **Basic** | 3 | Immediate | Automated security scan |
| **Standard** | 7 | < 1 hour | Basic + deterministic execution test |
| **Advanced** | 11 | 1-3 days | Standard + manual code review |
| **Enterprise** | 15 | 1-4 weeks | Advanced + compliance audit |

## Getting Started

### Check a Function's Trust Score

```bash
curl -X GET "https://api.functionfly.com/v1/trust/score/{function_id}" \
  -H "Authorization: Bearer {api_key}"
```

Response:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "trust_score": 87.5,
  "trust_tier": "verified",
  "is_verified": true,
  "verification_level": "standard",
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
    "p50_latency_ms": 45.2
  }
}
```

### Batch Check Multiple Functions

```bash
curl -X POST "https://api.functionfly.com/v1/trust/batch" \
  -H "Authorization: Bearer {api_key}" \
  -H "Content-Type: application/json" \
  -d '{"function_ids": ["id1", "id2", "id3"]}'
```

### View Trust History

Track how a function's trust has changed over time:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/history/{function_id}?page=1&page_size=20" \
  -H "Authorization: Bearer {api_key}"
```

## Submitting Functions for Verification

If you're a function author, submit your function for verification:

```bash
curl -X POST "https://api.functionfly.com/v1/trust/verify" \
  -H "Authorization: Bearer {api_key}" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "550e8400-e29b-41d4-a716-446655440002",
    "function_version": "1.2.0",
    "verification_level": "standard"
  }'
```

## Trust Policies

Define rules to enforce minimum trust requirements:

```json
{
  "name": "High Security Policy",
  "default_action": "deny",
  "rules": [
    {"type": "min_trust_score", "value": 80.0},
    {"type": "verification_required", "value": true},
    {"type": "no_revocation", "value": true}
  ]
}
```

### Policy Rule Types

| Rule | Description |
|------|-------------|
| `min_trust_score` | Minimum trust score (0-100) |
| `verification_required` | Function must be verified |
| `tier_minimum` | Minimum tier (untrusted/trusted/verified/highly_trusted) |
| `no_revocation` | Function must not be revoked |
| `min_success_rate` | Minimum success rate (0-1) |

## Attestations

Attestations provide cryptographic proof of function properties:

| Type | Description |
|------|-------------|
| `verification` | Passed FunctionFly verification |
| `security_scan` | Passed security scanning |
| `code_review` | Code was reviewed |
| `execution` | Tested in execution environment |
| `compliance` | Meets compliance requirements |
| `signature` | Cryptographic signature from author |

### Verify Attestations

```bash
# Verify attestation integrity
curl -X GET "https://api.functionfly.com/v1/trust/attestations/{attestation_id}/verify" \
  -H "Authorization: Bearer {api_key}"

# Get full attestation chain
curl -X GET "https://api.functionfly.com/v1/trust/attestations/{function_id}/chain" \
  -H "Authorization: Bearer {api_key}"
```

## Revocation

Check if a function has been revoked:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/revoke/revoked/{function_id}" \
  -H "Authorization: Bearer {api_key}"
```

### Revocation Reasons

| Reason | Description |
|--------|-------------|
| `security` | Security vulnerability discovered |
| `malware` | Function contains malware |
| `abuse` | Function is being abused |
| `policy_violation` | Violates FunctionFly policies |
| `reported` | Reported by a partner |
| `deprecated` | Function deprecated |

## FXCERT - Function Execution Certificate

Verified functions receive an FXCERT - a legal-grade artifact proving execution properties:

```json
{
  "certificate_id": "vfy_abc123...",
  "function_id": "550e8400-e29b-41d4-a716-446655440002",
  "level": "standard",
  "trust": {
    "trust_score": 87.5,
    "determinism_score": 0.95,
    "reproducibility": true
  },
  "signature": "base64_signature"
}
```

## AI Agent Integration

When an AI agent selects a function:

1. **Query** registry with trust thresholds
2. **Evaluate** trust scores using weighted formula
3. **Filter** by verification level if required
4. **Select** highest-scoring function meeting requirements
5. **Monitor** execution and update trust

### Fallback Protocol

When invocation fails:

1. Check if failure is trust-related (`TRUST_VIOLATION` vs `EXECUTION_ERROR`)
2. Find fallback candidates: `?capability={cap}&trust_score_min={score}`
3. Evaluate fallback trust scores
4. Select best fallback (higher trust, consider latency/cost)

## Webhooks

Subscribe to real-time trust events:

```json
{
  "name": "Trust Notifications",
  "url": "https://your-app.com/webhooks/trust",
  "events": [
    "trust.score.updated",
    "trust.revocation.created",
    "trust.verification.completed"
  ]
}
```

### Available Events

| Event | Description |
|-------|-------------|
| `trust.score.updated` | Trust score changed |
| `trust.revocation.created` | New revocation |
| `trust.revocation.lifted` | Revocation lifted |
| `trust.verification.completed` | Verification completed |
| `trust.report.submitted` | New report |

## SDKs

- [JavaScript/TypeScript SDK](/docs/sdks/javascript)
- [Python SDK](/docs/sdks/python)
- [Go SDK](/docs/sdks/go)
- [All SDKs](/docs/sdks)

## Related Documentation

- [Trust API Reference](/docs/trust-api) - Full API endpoint documentation
- [Trust Protocol Specification](/docs/trust-protocol-spec) - Technical protocol details
- [Security Features](/docs/security) - Security implementation
