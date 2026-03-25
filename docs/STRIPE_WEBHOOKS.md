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
3. **Events to send:** select the three events above (or “Select events” and add them)
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

Note: synthetic payloads may not include your app’s `metadata`; real flows (Checkout with metadata) are what populate State Fabric add-on entitlement rows.
