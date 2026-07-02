---
title: Pricing & Monetization
description: Revenue models, creator payouts, and the marketplace economy
sidebar:
  order: 3
---


The FunctionFly Marketplace provides a full creator economy with flexible
pricing models, automatic revenue splitting, and creator payouts.

## Revenue Split

| Recipient | Share |
|-----------|-------|
| Creator | **80%** |
| Platform | **10%** |
| Payment processing | **10%** |

Revenue is tracked per transaction and aggregated for payouts.

## Pricing Models

### Per Call

Fixed price charged on each invocation:

```json
{
  "pricing_model": "per_call",
  "price_per_call": 0.001
}
```

Best for: Utility functions, API wrappers, data transformations.

### Subscription

Monthly recurring charge:

```json
{
  "pricing_model": "subscription",
  "subscription_monthly_usd": 9.99
}
```

Best for: Agents with ongoing value, SaaS integrations.

### Revenue Share

Creator receives a percentage of revenue generated:

```json
{
  "pricing_model": "revenue_share",
  "revenue_share_percent": 15
}
```

Best for: Functions that directly generate revenue for the buyer.

### Tiered

Volume-based pricing with discounts:

```json
{
  "pricing_model": "tiered",
  "tiered_pricing": [
    { "up_to": 1000, "price_per_call": 0.002 },
    { "up_to": 10000, "price_per_call": 0.001 },
    { "up_to": null, "price_per_call": 0.0005 }
  ]
}
```

Best for: High-volume users, enterprise workloads.

### Dynamic

Price adjusts based on demand:

```json
{
  "pricing_model": "dynamic",
  "min_price": 0.0005,
  "max_price": 0.005,
  "demand_factor": 1.5
}
```

Best for: Compute-intensive functions with variable demand.

### Auction

Bidding-based pricing:

```json
{
  "pricing_model": "auction",
  "start_price": 0.001,
  "reserve_price": 0.0005,
  "end_time": "2026-07-08T00:00:00Z"
}
```

Best for: Exclusive or scarce capabilities.

## Creator Economy Dashboard

Creators have access to a dashboard showing:

- **Total revenue** — All-time and period breakdown
- **Active subscriptions** — Subscriber count and MRR
- **Transaction history** — Every revenue event
- **Payout history** — Past and pending payouts
- **Top functions/agents** — Revenue by asset

## Payouts

### Requesting a Payout

Creators can request payouts once they reach the minimum threshold:

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/payouts \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 150.00,
    "currency": "USD"
  }'
```

### Payout Status

| Status | Description |
|--------|-------------|
| `pending` | Request submitted |
| `processing` | Payment being processed |
| `completed` | Funds transferred |
| `failed` | Payment failed (retryable) |

## Licensing

### Open Source

Functions with MIT, Apache, or GPL licenses are free to use. Attribution
may be required depending on the license.

### Commercial

Proprietary functions require a license key:

1. Buyer purchases the function
2. A license key is issued with activation limits
3. Key is validated on each invocation
4. Creator can revoke keys for violations

### License Policies

```json
{
  "spdx_license": "proprietary",
  "commercial_type": "commercial",
  "max_activations_default": 5
}
```

| Field | Description |
|-------|-------------|
| `spdx_license` | License identifier (mit, apache, gpl, proprietary, custom) |
| `commercial_type` | `open`, `restricted`, or `commercial` |
| `max_activations_default` | Default activation limit for new licenses |

## Agent Hiring

Agents can be hired for specific tasks with a budget:

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/agents/{id}/hire \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "task_type": "code_review",
    "task_payload": { "repo": "github.com/org/repo", "pr": 42 },
    "budget_usd": 5.00
  }'
```

The agent executes the task within the budget. Unused budget is returned.

## Subscription Plans

Creators can offer subscription tiers:

| Plan | Billing Cycle | Features |
|------|--------------|----------|
| Basic | Monthly | Core features |
| Pro | Monthly/Quarterly | Advanced features |
| Enterprise | Annual | Full feature set + SLA |

Subscribers get access to the creator's functions/agents for the duration
of their subscription.

## Marketplace Bundle (Tenant Marketplace)

The **Marketplace Bundle** lets tenants run their own multi-vendor marketplace:

| Plan | Price | Features |
|------|-------|----------|
| Founder | Free (3 months) | Early access, all features |
| Starter | $49/mo | Basic marketplace |
| Growth | $149/mo | Advanced analytics, custom branding |
| Scale | $399/mo | Full white-label, Stripe Connect |

Transaction fees: 2.5% + Stripe fees (or 1.5% annual billing).

## Next Steps

- [Publishing](/marketplace/publishing/) — List your assets
- [API Reference](/marketplace/api/) — Full endpoint docs
- [Billing & Subscription](/guides/billing/) — Platform billing
