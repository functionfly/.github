---
title: Billing & Subscription
description: Understand how FunctionFly billing, subscriptions, and payments work.
sidebar:
  order: 12
---



This guide explains how FunctionFly billing works, including subscriptions, usage-based charges, invoices, and payment methods.

## Plan Overview

FunctionFly offers tiered plans designed to scale with your needs:

| Plan | Price | Best For |
|------|-------|----------|
| **Free** | $0/mo | Personal demos, learning, side projects |
| **Starter** | $24/mo | Side projects, MVPs, small applications |
| **Professional** | $79/mo | Growing SaaS apps, teams, production apps |
| **Enterprise** | $299/mo | Large-scale apps, compliance needs, dedicated support |
| **Agent Enterprise** | Custom | Unlimited agents, priority support, on-premise |

**Annual Discount:** Starter and Professional save 20% when billed annually. Enterprise saves 30%.

---

## How Billing Works

### Subscription Billing

Subscriptions are billed **monthly or annually** on your billing cycle date.

- Monthly: Billed on the same day each month
- Annual: Billed once per year,，享受 20-30% discount

### Usage-Based Charges

Beyond your base plan, some features use credits or incur overage:

| Feature | Pricing Model |
|---------|--------------|
| Function requests (over plan limit) | Per 1K calls (varies by plan) |
| DNA Mutations | 50 credits each |
| Function Remix | Varies by function |
| StateFabric operations (over limit) | Per 10K operations |
| Agent AI calls (over limit) | Per 100K calls |

### Credits System

FunctionFly uses a **credits system** for micro-transactions:

```
$1 USD = 100 credits
```

Credits are purchased through the Wallet and used for:
- DNA mutation acceptances
- Function remixes
- Premium API calls
- Add-on features

---

## Reading Your Invoice

### Invoice Components

Each invoice includes:

1. **Subscription charges** — Base plan fee
2. **Usage charges** — Overages beyond plan limits
3. **Credits applied** — Credits redeemed against charges
4. **Taxes** — Applicable sales tax or VAT
5. **Credits from referrals** — Referral bonuses applied

### Invoice Example

```
FunctionFly Invoice #FF-2026-0512

Subscription:
  Professional Plan (Monthly)           $79.00
  Additional Functions (2 × $5)         $10.00
  Custom Domain                          $10.00

Usage:
  Function Requests (1.2M over 10M)     $16.00
  StateFabric Operations (150K over 100K) $5.00

Credits:
  Wallet Credits Applied               -$20.00

Subtotal                                $100.00
Tax (8%)                                  $8.00

Total Due                                $108.00
```

### Accessing Invoices

1. Navigate to **Settings → Billing**
2. Click **Invoice History**
3. Download PDF or view online

---

## Payment Methods

### Supported Payment Methods

| Method | Processing Time | Notes |
|--------|----------------|-------|
| Credit/Debit Card | Instant | Visa, Mastercard, Amex |
| PayPal | Instant | Via PayPal account |
| ACH Transfer | 3-5 business days | US bank accounts only |
| Wire Transfer | 3-5 business days | Enterprise only, minimum $500 |

### Adding a Payment Method

1. Go to **Settings → Billing → Payment Methods**
2. Click **Add Payment Method**
3. Enter card details or connect PayPal
4. Set as default if desired

### Failed Payments

If a payment fails:
1. You'll receive an email notification
2. We retry payment after 3 days
3. After 7 days, services may be suspended
4. Add valid payment method to restore service

---

## Subscription Management

### Upgrading Your Plan

Upgrades take effect **immediately** with prorated billing:

1. Go to **Settings → Billing → Plan**
2. Click **Upgrade**
3. Select new plan
4. Confirm prorated charge

**Proration example:** If you upgrade from Starter ($24) to Professional ($79) on day 15 of a 30-day billing cycle:
- Credit: $12 (remaining 15 days of Starter)
- Charge: $39.50 (remaining 15 days of Professional)
- **Due today: $27.50**

### Downgrading Your Plan

Downgrades take effect at the **end of your billing cycle**:

1. Go to **Settings → Billing → Plan**
2. Click **Downgrade**
3. Select new plan
4. Confirm — changes apply at next billing date

### Cancelling Subscription

1. Go to **Settings → Billing → Plan**
2. Click **Cancel Subscription**
3. Choose cancellation date (now or end of cycle)
4. Confirm cancellation

**After cancellation:**
- You retain access until the end of your paid period
- Data is retained for 30 days
- You can reactivate anytime

---

## Wallet & Credits

### Adding Credits

1. Go to **Wallet** in the dashboard
2. Click **Add Credits**
3. Choose amount or custom amount
4. Complete payment

| Amount | Bonus |
|--------|-------|
| $10+ | None |
| $50+ | 5% bonus |
| $100+ | 10% bonus |
| $500+ | 15% bonus |

### Auto-Recharge

Set up auto-recharge to never run out:

1. Go to **Wallet → Auto-Recharge**
2. Set minimum balance threshold (e.g., $10)
3. Set recharge amount (e.g., $50)
4. Enable auto-recharge

When balance drops below threshold, we automatically recharge.

### Credit Usage

Credits are used automatically for:
- DNA mutation acceptances (50 credits each)
- Function remix purchases
- Premium feature access
- Overages when specified in your plan

---

## Usage Limits & Alerts

### Understanding Limits

Each plan has limits on:

| Resource | Free | Starter | Professional | Enterprise |
|----------|------|---------|--------------|------------|
| Functions | 1 | 5 | 25 | Unlimited |
| Monthly Requests | 100K | 1M | 10M | 100M+ |
| StateFabric Objects | 1 | 5 | 50 | 500+ |
| State Operations/mo | 10K | 100K | 1M | 10M+ |
| Secrets | — | 10 | 50 | 10,000+ |
| Team Members | 1 | 3 | 10 | Unlimited |

### Low Balance Alerts

Set up alerts to avoid service interruptions:

1. Go to **Settings → Billing → Alerts**
2. Enable **Low Balance Alert**
3. Set threshold (default: $5)

Alerts sent via:
- In-app notification
- Email

### Overage Handling

**Free tier:** Hard stop — no overage billing

**Paid tiers:** Overage billing with notice

1. **Warning** at 80% of limit — email notification
2. **Alert** at 95% — email + in-app notification
3. **Overage begins** — billed at overage rate
4. **Hard cap** at 200% — service paused until upgrade or cycle reset

---

## Refunds & Disputes

### Refund Policy

- **Accidental charges:** Full refund within 30 days
- **Service outages:** Pro-rated credit based on uptime SLA
- **Feature not as described:** Case-by-case basis

### Requesting a Refund

1. Go to **Settings → Billing → Invoice**
2. Click **Request Refund** on the invoice
3. Select reason
4. Submit

Refunds typically processed within 5-7 business days.

### Billing Disputes

If you believe a charge is incorrect:

1. Contact support with invoice number
2. Provide explanation of dispute
3. We'll investigate within 5 business days

---

## Tax Information

### Tax Calculation

Tax is calculated based on:
- Your billing address
- Services used
- Applicable tax jurisdiction

### Tax Exemptions

If you're tax-exempt:

1. Go to **Settings → Billing → Tax Information**
2. Upload tax exemption certificate
3. Certificate is reviewed within 2 business days
4. Once approved, tax will be removed from future invoices

### EU VAT

For EU customers:
- VAT ID can be entered for reverse charge
- Businesses with valid VAT ID can self-assess VAT
- Consumer accounts are charged VAT at point of purchase

---

## Referral Program

Earn credits by referring new users:

1. Go to **Settings → Referrals**
2. Get your unique referral link
3. Share with friends

**For each referral who upgrades:**
- You earn $10 in credits
- They get 20% off first month

---

## Enterprise Billing

Enterprise customers have additional options:

### Invoicing

- Net-30 payment terms
- Purchase orders accepted
- Annual invoicing available

### Custom Contracts

- Volume discounts
- Custom SLAs
- Dedicated support terms
- On-premise deployment options

### Getting Started with Enterprise

Contact our sales team at [functionfly.com/contact](https://functionfly.com/contact) or email [enterprise@functionfly.com](mailto:enterprise@functionfly.com)

---

## FAQ

**Can I switch from monthly to annual billing?**
Yes. Go to Settings → Billing → Plan → Switch to Annual. You'll receive a prorated credit for unused monthly billing.

**What happens if I cancel mid-month?**
You keep access until the end of your paid period. No prorated refund, but no further charges.

**Can I get a refund for unused credits?**
Credits are non-refundable. Use them before cancelling.

**How do I update my billing email?**
Go to Settings → Billing → Billing Information. Update the email address for invoices.

**Is my payment information secure?**
Yes. We use Stripe for payments and never store your full card number. All data is PCI-DSS compliant.
