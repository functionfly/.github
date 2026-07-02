---
title: Marketplace
description: Discover, publish, and monetize functions, agents, and extensions
---


The FunctionFly Marketplace is a unified discovery and distribution platform
for three asset types: **functions**, **agents**, and **extensions**.

## What's in the Marketplace

| Asset Type | Description | Examples |
|------------|-------------|----------|
| **Functions** | Serverless functions published to the registry with marketplace pricing | PDF summarizer, image resizer, email sender |
| **Agents** | Pre-built AI agents that can be hired for tasks | Code reviewer, data analyst, customer support bot |
| **Extensions** | Studio plugins and integrations | GitHub integration, Slack notifier, analytics widget |

## Quick Start

### Browse

**Dashboard:** Marketplace (sidebar) — search across all asset types, filter
by type, category, or rating.

**API:**

```bash
curl "https://api.functionfly.com/v1/marketplace/search?q=pdf&limit=20"

curl "https://api.functionfly.com/v1/marketplace/search?q=summarize&type=function"
```

### Publish a Function

1. Publish your function to the registry (see [Creating Functions](/functions/creating/))
2. Create a marketplace listing with a pricing model
3. Your function appears in search results

### Publish an Extension

```bash
curl -X POST https://api.functionfly.com/v1/marketplace/extensions \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Studio Plugin",
    "version": "1.0.0",
    "description": "A useful Studio extension",
    "category": "developer-tools",
    "tags": ["productivity", "automation"]
  }'
```

### Hire an Agent

Browse agents in the marketplace, select one, and hire it with a budget:

```bash
curl "https://api.functionfly.com/v1/marketplace/search?q=code+review&type=agent"
```

## Monetization

The marketplace supports multiple pricing models:

| Model | Description |
|-------|-------------|
| **Free** | No charge |
| **Per Call** | Fixed price per invocation |
| **Subscription** | Monthly recurring charge |
| **Revenue Share** | Percentage of revenue (creator keeps 80%) |
| **Tiered** | Volume-based pricing with discounts |
| **Dynamic** | Price adjusts with demand |
| **Auction** | Bidding-based pricing |

### Creator Revenue

- Creators keep **80%** of revenue
- Platform fee: **10%**
- Payment processing: **10%**
- Payouts available via creator dashboard

## Ratings & Reviews

Every asset type supports ratings (1–5 stars) and written reviews:

- **Functions**: Ratings blend automated quality scores (60%) with user ratings (40%)
- **Agents**: Ratings auto-update based on execution quality and ROI
- **Extensions**: User ratings with review text

## Discovery Features

- **Unified search** — Search across functions, agents, and extensions
- **Category filtering** — Browse by category
- **Sort options** — Trending, top rated, newest, most installed
- **Featured items** — Curated picks highlighted in the marketplace
- **Trust scores** — Security and quality indicators on every listing

## Next Steps

- [Publishing](/marketplace/publishing/) — List your functions, agents, and extensions
- [Pricing & Monetization](/marketplace/pricing/) — Pricing models, licensing, and payouts
- [API Reference](/marketplace/api/) — Full endpoint documentation
