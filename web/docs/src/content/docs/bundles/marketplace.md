---
title: Marketplace Bundle
description: Build multi-vendor platforms with seller profiles, listings, payment splits, reviews, and admin tools.
sidebar:
  order: 3
---

# Marketplace Bundle

The complete backend for multi-vendor marketplaces. Handle sellers, listings, payments, reviews, and moderation—all built in.

## What's Included

### Seller Management
- **Seller profiles** — Bio, logo, policies, payout settings
- **Onboarding flow** — Stripe Connect integration for payouts
- **Seller dashboard** — Sales, earnings, order management
- **Multi_currency support** — Accept payments in multiple currencies

### Listings & Products
- **Product catalog** — Categories, variants, inventory tracking
- **Media management** — Images, videos, 360° views
- **Pricing rules** — Quantity discounts, promotional pricing
- **Search & discovery** — Full-text search with filters

### Payments & Splitting
- **Stripe Connect** — Automatic payment splitting to sellers
- **Commission engine** — Configurable platform fees
- **Escrow support** — Hold funds until order completion
- **Refunds & disputes** — Full refund workflow with dispute resolution

### Reviews & Ratings
- **Review submission** — Post-purchase review requests
- **Moderation tools** — Filter inappropriate content
- **Seller responses** — Allow sellers to respond to reviews
- **Rating aggregation** — Overall rating with breakdown

### Admin Tools
- **Dashboard** — Platform-wide metrics and insights
- **User management** — Buyers, sellers, and internal users
- **Content moderation** — Flag and remove listings
- **Payout management** — Manual payouts and adjustments

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Your Frontend                           │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                   Marketplace Bundle                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Seller  │  │ Listigns  │  │ Payments │  │  Reviews  │  │
│  │ Service  │  │  Service  │  │  (Split) │  │  Service  │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       │             │            │            │            │
│  ┌────▼────────────▼────────────▼────────────▼────┐       │
│  │               Database (PostgreSQL)             │       │
│  │ Sellers │ Listings │ Transactions │ Reviews   │       │
│  └─────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

## Pricing

| Plan | Price | Sellers | Features |
|------|-------|---------|----------|
| **Founder** | Free (3 months) | Up to 50 | All features |
| **Starter** | $49/mo | Up to 500 | All features |
| **Growth** | $149/mo | Up to 5,000 | + Advanced analytics |
| **Scale** | $399/mo | Unlimited | + Priority support |

**Transaction fees:** 2.5% + Stripe fees (or 1.5% with annual billing)

## Getting Started

1. Go to **Dashboard → Bundles → Marketplace**
2. Click **Deploy Bundle**
3. Configure your platform fee (default: 10%)
4. Set up Stripe Connect for seller payouts
5. Start building your seller and buyer frontends

## Customization

- Adjust commission rates per category
- Add custom seller onboarding requirements
- Configure escrow release rules
- Extend with your own product types

## Next Steps

- [Set up Stripe Connect](/guides/webhooks/)
- [Configure seller onboarding](/guides/authentication/)
- [Build your first listing type](/guides/creating-functions/)