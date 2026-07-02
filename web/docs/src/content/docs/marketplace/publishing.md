---
title: Publishing
description: List your functions, agents, and extensions in the marketplace
sidebar:
  order: 2
---


Publish your functions, agents, and extensions to reach FunctionFly's user
base and earn revenue.

## Publishing a Function

### Prerequisites

1. Function must exist in the [registry](/registry/) (published via CLI or API)
2. Function must pass trust and security checks

### Create a Listing

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/functions \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "fx_abc123",
    "pricing_model": "per_call",
    "price_per_call": 0.001,
    "description": "Summarize any PDF document in seconds"
  }'
```

### Pricing Models

| Model | Fields | Description |
|-------|--------|-------------|
| `free` | — | No charge |
| `per_call` | `price_per_call` | Fixed price per invocation (USD) |
| `subscription` | `subscription_monthly_usd` | Monthly recurring (USD) |
| `revenue_share` | `revenue_share_percent` | Creator's share of revenue |
| `tiered` | `tiered_pricing` (JSONB) | Volume-based tiers |
| `dynamic` | `min_price`, `max_price`, `demand_factor` | Demand-based pricing |
| `auction` | `start_price`, `reserve_price`, `end_time` | Bidding |

### Rating Blending

Function ratings are a **60/40 blend**:
- 60% — Automated quality score (trust score, determinism, reliability)
- 40% — User ratings (1–5 stars)

This ensures high-quality functions surface even with few user reviews.

## Publishing an Agent

### Prerequisites

1. Agent must be registered in `agent_identities`
2. Agent must have a valid configuration

### Create a Listing

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/agents \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_xyz",
    "listing_type": "worker",
    "pricing_model": "per_call",
    "price_per_call": 0.005,
    "description": "AI code reviewer with deep analysis"
  }'
```

### Listing Types

| Type | Description |
|------|-------------|
| `worker` | Executes tasks on demand |
| `manager` | Orchestrates other agents |
| `infrastructure` | Provides shared services |

### Ranking Algorithm

Agent search results are ranked by a weighted score:

| Factor | Weight |
|--------|--------|
| Trust score | 30% |
| Economic score | 25% |
| Reliability | 20% |
| ROI score | 15% |
| Call volume | 10% |

## Publishing an Extension

### Prerequisites

1. Extension must include a valid manifest
2. Extension must pass security analysis

### Create an Extension

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/extensions \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "GitHub Integration",
    "version": "1.0.0",
    "description": "Connect your functions to GitHub repos",
    "category": "developer-tools",
    "tags": ["github", "ci-cd", "automation"],
    "manifest": {
      "permissions": ["repo:read", "webhook:write"],
      "entry_point": "index.js"
    }
  }'
```

### Extension Lifecycle

1. **Draft** — Created, not visible in marketplace
2. **Published** — Visible, installable by users
3. **Active** — Installed by users, running

### Security Analysis

Every extension undergoes automated security analysis on submission:

- Manifest validation
- Permission scope review
- Sandbox compatibility check
- Code signature verification

Extensions that fail analysis are rejected with a detailed report.

### Installing an Extension

Users install extensions via the marketplace:

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/extensions/{id}/install \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

This creates a plugin in the user's workspace and runs security checks.

## Licensing (Functions)

Functions can specify a license:

| License | Commercial Type | Description |
|---------|----------------|-------------|
| MIT | open | Permissive, no restrictions |
| Apache 2.0 | open | Permissive with patent grant |
| GPL | open | Copyleft, derivative works must be GPL |
| Proprietary | commercial | Custom terms, license key required |
| Custom | varies | Custom SPDX identifier |

### License Keys

Commercial functions can require license keys:

- Keys are hash-based with activation limits
- Support expiration dates
- Can be revoked by the creator
- Activation audit trail maintained

## Next Steps

- [Pricing & Monetization](/marketplace/pricing/) — Revenue, payouts, and creator economy
- [API Reference](/marketplace/api/) — Full endpoint docs
- [Registry Guide](/guides/registry-guide/) — Publishing functions to the registry
