# Production Environment Variables

## Tenant Isolated Payment Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `USE_ISOLATED_PAYMENT_FLOW` | **Yes** | Set to `true` in production to enable isolated payment flow |
| `APP_URL` | **Yes** | Your public API URL (e.g., `https://api.functionfly.com`) - used for webhook URLs |
| `STRIPE_SECRET_KEY` | **Yes** | Your live Stripe secret key (`sk_live_...`) |
| `STRIPE_WEBHOOK_SECRET` | **Yes** | Platform webhook signing secret from Stripe Dashboard |
| `INTERNAL_WEBHOOK_SECRET` | **Yes** | Internal auth for webhook callbacks |

## Existing Required Variables

These should already be set but must be verified for production:

| Variable | Current | Required for Isolated Payments |
|----------|---------|-------------------------------|
| `STRIPE_SECRET_KEY` | `sk_test_...` | ❌ Must switch to live `sk_live_...` |
| `STRIPE_WEBHOOK_SECRET` | Set | ✅ Should remain same |
| `APP_URL` | Not set | ❌ Must be set to public API URL |

## Production Checklist

### 1. Environment Variables (.env or secrets manager)

```bash
# Enable isolated payment flow
USE_ISOLATED_PAYMENT_FLOW=true

# Set to your public API URL (required for tenant webhook URLs)
APP_URL=https://api.functionfly.com

# Switch to Stripe live keys
STRIPE_SECRET_KEY=sk_live_...
```

### 2. Stripe Dashboard Configuration

For each tenant with isolated payments, add webhook endpoints:

1. **Dashboard** → **Developers** → **Webhooks** → **Add endpoint**
2. **Endpoint URL:** `https://api.functionfly.com/v1/billing/tenants/{tenant_id}/webhook`
3. **Events to listen for:**
   - `checkout.session.completed`
   - `invoice.payment_succeeded`
   - `invoice.payment_failed`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
4. Copy the **Signing secret** and store it in `tenant_stripe_configs.metadata.webhook_secret`

### 3. Database Migration

The migration `20260501133244` should already be applied (verified earlier). If not:

```bash
migrate -path migrations -database $DATABASE_URL up 20260501133244
```

### 4. Application Restart

After setting env vars, restart the orchestrator API:

```bash
./bin/orchestrator-api --skip-migrations
```

## Payment Modes Summary

| Mode | When Active | Stripe Customer | Webhook |
|------|-------------|-----------------|---------|
| `platform` | Default (no config) | Platform's customer | `/v1/billing/subscription/webhook` |
| `isolated` | `USE_ISOLATED_PAYMENT_FLOW=true` | Tenant's own customer | `/v1/billing/tenants/{id}/webhook` |
| `connect` | Marketplace bundle | Tenant's Connect account | `/v1/billing/tenants/{id}/webhook` |

## Bundle Payment Configuration

| Bundle | Payment Mode | Payment Methods |
|--------|--------------|-----------------|
| `saas-starter` | `isolated` | card only |
| `marketplace` | `connect` | card, us_bank_account, eu_bank_transfer |
| `ai-app` | `isolated` | card, us_bank_account |
| Default | `platform` | card only |