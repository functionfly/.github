# Stripe webhooks (FunctionFly API)

The orchestrator exposes a **public** Stripe webhook endpoint:

- **URL path:** `POST /v1/webhooks/stripe`
- **Full URL (example):** `https://api.example.com/v1/webhooks/stripe`

Set **`STRIPE_WEBHOOK_SECRET`** on the API server to the signing secret for that endpoint (Dashboard or `stripe listen` output).

## Required events

Subscribe this endpoint to at least:

- `checkout.session.completed`
- `customer.subscription.updated`
- `customer.subscription.deleted`

These are used for:

- Checkout completion (e.g. credits, registry wallet, **State Fabric add-on** entitlements)
- Subscription lifecycle sync for **State Fabric add-on** rows

## Stripe Dashboard (production)

1. [Stripe Dashboard](https://dashboard.stripe.com/) → **Developers** → **Webhooks** → **Add endpoint**
2. **Endpoint URL:** `https://<your-api-host>/v1/webhooks/stripe`
3. **Events to send:** select the three events above (or "Select events" and add them)
4. Copy the endpoint **Signing secret** (starts with `whsec_`) into **`STRIPE_WEBHOOK_SECRET`** on the API

## Local development with Stripe CLI

Install the [Stripe CLI](https://stripe.com/docs/stripe-cli), then run the listener so Stripe sends signed webhooks to your local API (default orchestrator port **8080**):

```bash
stripe listen \
  --events checkout.session.completed,customer.subscription.updated,customer.subscription.deleted \
  --forward-to http://localhost:8080/v1/webhooks/stripe
```

The CLI prints a **webhook signing secret** (`whsec_...`). Use it while testing:

```bash
export STRIPE_WEBHOOK_SECRET="whsec_..."   # paste from stripe listen output
```

Keep **`STRIPE_SECRET_KEY`** set to your test secret (`sk_test_...`) so Checkout and subscriptions match the same Stripe account as the CLI.

### Optional: trigger test events

With the listener running in another terminal:

```bash
stripe trigger checkout.session.completed
stripe trigger customer.subscription.updated
stripe trigger customer.subscription.deleted
```

Note: synthetic payloads may not include your app's `metadata`; real flows (Checkout with metadata) are what populate State Fabric add-on entitlement rows.

---

## Tenant-Isolated Payment Webhooks

For tenants with isolated payment flow enabled, FunctionFly supports per-tenant webhook endpoints. Each tenant with isolated payments gets their own webhook URL and signing secret.

### Endpoint

- **URL path:** `POST /v1/billing/tenants/{tenant_id}/webhook`
- **Purpose:** Receives Stripe webhooks for tenant-isolated payment events

### How it works

1. When a tenant is provisioned with isolated payments, a unique webhook URL is generated:
   ```
   https://api.functionfly.io/v1/billing/tenants/{tenant_id}/webhook
   ```

2. A signing secret is generated for signature verification (stored in `tenant_stripe_configs.metadata.webhook_secret`)

3. The tenant's Stripe account is configured to send webhooks to this URL

### Supported Events

- `checkout.session.completed` - Payment completed (bundle subscriptions, wallet credits, agent credits)
- `invoice.payment_succeeded` - Recurring payment succeeded
- `invoice.payment_failed` - Recurring payment failed
- `customer.subscription.updated` - Subscription modified
- `customer.subscription.deleted` - Subscription cancelled

### Stripe Dashboard Configuration

For each tenant with isolated payments, add a separate webhook endpoint in Stripe:
1. **Dashboard** → **Developers** → **Webhooks** → **Add endpoint**
2. **Endpoint URL:** `https://api.functionfly.io/v1/billing/tenants/{tenant_id}/webhook`
3. **Events:** Select the events above
4. Copy the signing secret and store it in `tenant_stripe_configs.metadata.webhook_secret`

### Security

- Each tenant has a unique signing secret
- Signatures are verified using HMAC-SHA256 with 5-minute timestamp tolerance
- Rate limiting: 100 calls per minute per tenant

---

## Environment Variables for Payment Isolation

| Variable | Purpose | Default |
|----------|---------|---------|
| `USE_ISOLATED_PAYMENT_FLOW` | Enable isolated checkout flow | `false` |
| `STRIPE_SECRET_KEY` | Stripe API key | Required |
| `STRIPE_WEBHOOK_SECRET` | Platform webhook secret | Required in production |
| `APP_URL` | Base URL for webhook URLs | `https://api.functionfly.io` |

### Enabling Isolated Payments for SaaS Starter

To enable isolated payment processing for SaaS Starter bundle tenants:

```bash
# In production, set this environment variable
export USE_ISOLATED_PAYMENT_FLOW=true

# The platform will automatically:
# 1. Create tenant-specific Stripe customer
# 2. Route payments through isolated flow
# 3. Configure webhook endpoint for the tenant
```

### Payment Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `platform` | Default platform Stripe account | Most tenants |
| `isolated` | Tenant's own Stripe customer, platform-managed | SaaS Starter, AI App bundles |
| `connect` | Full Stripe Connect with tenant's own account | Marketplace bundles |