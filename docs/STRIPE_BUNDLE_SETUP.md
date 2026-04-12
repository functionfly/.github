# Stripe Setup for Backend-in-a-Box Pricing

## Required Stripe Products and Prices

### 1. SaaS Starter Pack - $29/month
```bash
# Create Product
curl https://api.stripe.com/v1/products \
  -u "sk_live_..." \
  -d name="SaaS Starter Pack" \
  -d description="Pre-configured SaaS backend with Auth, Payments, User DB, Email, Analytics"

# Create Price (recurring monthly)
curl https://api.stripe.com/v1/prices \
  -u "sk_live_..." \
  -d product="prod_xxx" \
  -d unit_amount=2900 \
  -d currency=usd \
  -d "recurring[interval]=month"
```

### 2. Marketplace Pack - $49/month
```bash
# Create Product
curl https://api.stripe.com/v1/products \
  -u "sk_live_..." \
  -d name="Marketplace Pack" \
  -d description="Complete marketplace backend with Listings, Payments, Messaging, Notifications"

# Create Price (recurring monthly)
curl https://api.stripe.com/v1/prices \
  -u "sk_live_..." \
  -d product="prod_xxx" \
  -d unit_amount=4900 \
  -d currency=usd \
  -d "recurring[interval]=month"
```

### 3. AI App Pack - $39/month
```bash
# Create Product
curl https://api.stripe.com/v1/products \
  -u "sk_live_..." \
  -d name="AI App Pack" \
  -d description="AI infrastructure with Vector DB, Embeddings, Chat, Memory System"

# Create Price (recurring monthly)
curl https://api.stripe.com/v1/prices \
  -u "sk_live_..." \
  -d product="prod_xxx" \
  -d unit_amount=3900 \
  -d currency=usd \
  -d "recurring[interval]=month"
```

## Environment Variables

Add these to your `.env` file after creating the Stripe prices:

```bash
# Backend-in-a-Box Bundle Pricing
STRIPE_PRICE_SAAS_STARTER=price_xxx    # $29/month
STRIPE_PRICE_MARKETPLACE=price_xxx     # $49/month
STRIPE_PRICE_AI_APP=price_xxx          # $39/month
```

## Webhook Events Required

Ensure your Stripe webhook endpoint handles these events:
- `checkout.session.completed` - User completed bundle subscription
- `invoice.paid` - Recurring payment successful
- `invoice.payment_failed` - Payment failed (grace period)
- `customer.subscription.deleted` - Subscription cancelled

## Founder Mode Flow

For "Build Now, Pay Later" / Founder Mode:
1. User registers without Stripe (deferred billing)
2. System tracks usage toward thresholds (100 users / $1K MRR / 90 days)
3. When threshold hit, create Stripe checkout for conversion
4. After payment, subscription continues normally

## Dashboard Configuration

Update pricing_tiers table with actual Stripe price IDs:

```sql
UPDATE pricing_bundles 
SET stripe_price_id = 'price_xxx' 
WHERE slug = 'saas-starter';

UPDATE pricing_bundles 
SET stripe_price_id = 'price_xxx' 
WHERE slug = 'marketplace';

UPDATE pricing_bundles 
SET stripe_price_id = 'price_xxx' 
WHERE slug = 'ai-app';
```
