---
title: SaaS Starter Bundle
description: Everything you need to launch a SaaS product — Auth, Payments, Database, Email, and Analytics.
sidebar:
  order: 2
---

The complete backend for SaaS products. Includes everything you need to launch and scale.

## What's Included

### Authentication
- **User management** — Sign up, login, password reset, email verification
- **OAuth providers** — Google, GitHub, Microsoft, Apple
- **Multi-factor authentication** — TOTP and WebAuthn support
- **Session management** — Secure JWT tokens with refresh rotation
- **Role-based access** — Owner, Admin, Member, Viewer roles built in

### Database
- **User schema** — Pre-configured users table with profiles
- **Tenant isolation** — Each organization gets its own data space
- **Migrations** — Safe schema updates with rollback support
- **Backup automation** — Daily backups with point-in-time recovery

### Payments (Stripe)
- **Subscription billing** — Monthly and annual plans with trials
- **Usage-based billing** — Pay for what you use
- **Metered billing** — Track usage and bill at period end
- **Payment webhooks** — Real-time payment events
- **Invoice generation** — Automatic Stripe invoices

### Email Workflows
- **Transactional email** — Welcome, password reset, invoice emails
- **Drip campaigns** — Onboarding sequences
- **Email analytics** — Open rates, click rates, unsubscribes

### Analytics
- **Usage dashboards** — DAU, MAU, retention curves
- **Revenue metrics** — MRR, ARR, churn, LTV
- **Funnel analysis** — Conversion through signup to paid

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Your Frontend                        │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                    SaaS Starter Bundle                      │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │
│  │  Auth   │  │ Payments│  │ Email   │  │Analytics │       │
│  │ Service │  │ (Stripe)│  │ Service │  │  Dashboard│       │
│  └────┬────┘  └────┬────┘  └────┬────┘  └─────────┘       │
│       │            │            │                          │
│  ┌────▼────────────▼────────────▼────┐                     │
│  │         Database (PostgreSQL)       │                    │
│  │  Users │ Subscriptions │ Events    │                    │
│  └────────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

## Pricing

| Plan | Price | Users | Features |
|------|-------|-------|----------|
| **Founder** | Free (3 months) | Up to 100 | All features |
| **Starter** | $29/mo | Up to 1,000 | All features |
| **Growth** | $99/mo | Up to 10,000 | + Advanced analytics |
| **Scale** | $299/mo | Unlimited | + Priority support |

## Getting Started

1. Go to **Dashboard → Bundles → SaaS Starter**
2. Click **Deploy Bundle**
3. Connect your Stripe account
4. Configure your email sender (SendGrid, Postmark, or SMTP)
5. Start building your frontend

## Customization

All components are customizable. You can:
- Modify the user schema to add custom fields
- Replace the email provider with your preferred service
- Add custom Stripe payment flows
- Extend the analytics with your own dashboards

## Next Steps

- [Deploy your first frontend](/guides/custom-domains/)
- [Set up your Stripe webhook](/guides/webhooks/)
- [Configure email templates](/guides/secrets-vault/)