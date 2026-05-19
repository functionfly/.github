---
title: Getting Started with Bundles
description: Learn how to deploy and customize Backend-in-a-Box bundles for your application.
sidebar:
  order: 17
---

import { Steps } from '@astrojs/starlight/components';

# Getting Started with Bundles

Deploy a complete backend in minutes with Backend-in-a-Box bundles. This guide walks you through choosing, deploying, and customizing a bundle for your project.

## Choose Your Bundle

| Bundle | Best For | What's Included |
|--------|----------|-----------------|
| **[SaaS Starter](/bundles/saas-starter/)** | Launching a SaaS product | Auth, User DB, Stripe, Email, Analytics |
| **[Marketplace](/bundles/marketplace/)** | Multi-vendor platforms | Sellers, Listings, Payments split, Reviews |
| **[AI App](/bundles/ai-app/)** | AI-powered applications | LLM gateway, Vector DB, Prompt management |

Not sure which bundle fits? See the [full comparison](/bundles/).

<Steps>

1. **Deploy Your Bundle**

   Go to **Dashboard → Bundles** and click **Deploy** on your chosen bundle.

   The deployment typically takes 30-60 seconds. You'll see a progress screen showing:
   - Database provisioning
   - Service configuration
   - API key generation

2. **Connect External Services**

   Depending on your bundle, you'll need to connect:

   - **Stripe** — For payment processing (all bundles)
   - **Email provider** — SendGrid, Postmark, or SMTP (SaaS Starter)
   - **LLM providers** — OpenAI, Anthropic, etc. (AI App)

   Find these settings in **Dashboard → Bundles → [Your Bundle] → Settings**.

3. **Get Your API Keys**

   After deployment, your API keys are available at:

   **Dashboard → Bundles → [Your Bundle] → API Keys**

   Each bundle has its own isolated API namespace:
   ```
   /v1/bundles/saas-starter/users
   /v1/bundles/saas-starter/auth
   /v1/bundles/saas-starter/payments
   ```

4. **Configure Your Domain**

   Point your domain or subdomain to your bundle:

   ```
   bundles.yourapp.com → Your bundle API
   ```

   See [Custom Domains](/guides/custom-domains/) for setup instructions.

5. **Start Building**

   Your bundle is now live. Build your frontend and connect to the bundle API.

</Steps>

## Founder Mode

All bundles include **3 months free** with Founder Mode:

- No credit card required
- Free until you hit **100 users** or **$1K MRR**
- 7-day grace period after hitting limits
- Your data is never deleted

Activate Founder Mode during bundle setup or from **Dashboard → Billing → Founder Mode**.

## Customization

All bundles are designed to be customized:

### Extend the Data Model

Add custom fields to existing schemas:

```javascript
// Example: Add a custom field to users
POST /v1/bundles/saas-starter/schema/extensions
{
  "field": "preferred_language",
  "type": "string",
  "required": false
}
```

### Replace Components

Swap out individual services with your own:

- Replace the email provider with your preferred service
- Use your own Stripe account configuration
- Add custom authentication flows

### Add New Features

Extend bundles with additional FunctionFly features:

- [Functions](/guides/creating-functions/) — Add custom business logic
- [Agents](/guides/creating-agents/) — Add AI-powered automation
- [StateFabric](/guides/statefabric/) — Add durable state
- [Registry](/guides/using-registry/) — Publish reusable components

## Monitoring Your Bundle

Track bundle health and usage from **Dashboard → Bundles → [Your Bundle] → Analytics**.

| Metric | Description |
|--------|-------------|
| **Active Users** | Users who accessed the bundle in the last 30 days |
| **API Calls** | Total calls to bundle endpoints |
| **Revenue** | Payment volume processed (SaaS Starter, Marketplace) |
| **Errors** | Failed requests by endpoint |

Set up alerts at **Dashboard → Bundles → [Your Bundle] → Alerts**.

## Troubleshooting

**Bundle deployment failed**
: Check your Stripe connection and try again. Contact support if the issue persists.

**API keys not working**
: Verify the key has the correct permissions. Regenerate if needed from the dashboard.

**Payments not processing**
: Ensure your Stripe webhook is configured and pointing to your bundle endpoint.

**Bundle feels slow**
: Consider upgrading to a higher plan with dedicated resources. Check [monitoring](/guides/monitoring/) for bottlenecks.

## Next Steps

- [Set up Stripe webhooks](/guides/webhooks/)
- [Configure email templates](/guides/secrets-vault/)
- [Add custom domain](/guides/custom-domains/)
- [Monitor usage and costs](/guides/monitoring/)