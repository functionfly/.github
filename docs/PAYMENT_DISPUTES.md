# Payment Disputes and Refunds

This document describes the chargeback and refund handling system for FunctionFly.

## Overview

FunctionFly now tracks Stripe chargebacks (disputes) and refunds through webhooks and provides admin management tools to monitor and respond to them. This prevents revenue leakage and accounting inconsistencies when customers dispute charges.

## Features

### Webhook Events Handled

The system processes the following Stripe webhook events:

- `charge.dispute.created` - When a customer files a chargeback
- `charge.dispute.updated` - When a dispute status changes
- `charge.dispute.closed` - When a dispute is resolved
- `charge.dispute.funds_withdrawn` - When funds are withdrawn after losing a dispute
- `charge.refunded` - When a charge is refunded

### Database Tables

#### `payment_disputes`

Stores all chargeback disputes from Stripe:

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Internal dispute ID |
| `stripe_dispute_id` | string | Stripe's dispute identifier |
| `stripe_payment_id` | string | Associated payment intent |
| `stripe_charge_id` | string | Associated charge ID |
| `tenant_id` | UUID | Associated tenant (if known) |
| `user_id` | UUID | Associated user (if known) |
| `amount_cents` | integer | Disputed amount |
| `currency` | string | Currency code (e.g., USD) |
| `reason` | string | Dispute reason from Stripe |
| `status` | string | Current status (needs_response, won, lost, etc.) |
| `evidence_due_by` | timestamp | Deadline for evidence submission |
| `evidence_submitted` | boolean | Whether evidence was submitted |
| `evidence_data` | JSONB | Structured evidence details |
| `outcome` | string | Final outcome (won, lost) |
| `outcome_reason` | string | Reason for outcome |

#### `payment_refunds`

Stores all refunds from Stripe:

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Internal refund ID |
| `stripe_refund_id` | string | Stripe's refund identifier |
| `stripe_payment_id` | string | Associated payment intent |
| `stripe_charge_id` | string | Associated charge ID |
| `tenant_id` | UUID | Associated tenant (if known) |
| `user_id` | UUID | Associated user (if known) |
| `amount_cents` | integer | Refund amount |
| `currency` | string | Currency code |
| `status` | string | Refund status (succeeded, failed, pending) |
| `reason` | string | Refund reason (duplicate, fraud, requested_by_customer) |
| `receipt_number` | string | Stripe receipt number |
| `failure_reason` | string | If refund failed |

## Admin API Endpoints

### Dispute Management

#### List Disputes
```
GET /v1/admin/billing/disputes
```

Query Parameters:
- `limit` - Maximum results (default 50, max 500)
- `offset` - Pagination offset
- `status` - Filter by status
- `reason` - Filter by reason
- `is_open` - Filter to open disputes only (true/false)
- `requires_action` - Filter to disputes needing evidence
- `tenant_id` - Filter by tenant
- `start_date` / `end_date` - Date range filter (RFC3339)

#### Get Open Disputes
```
GET /v1/admin/billing/disputes/open
```

Returns only disputes requiring action with a count of items needing evidence.

#### Get Dispute Details
```
GET /v1/admin/billing/disputes/{disputeId}
```

Returns full dispute details including tenant and user information.

#### Update Dispute Status
```
PATCH /v1/admin/billing/disputes/{disputeId}/status
```

Request Body:
```json
{
  "status": "won",
  "outcome": "won",
  "outcome_reason": "Customer withdrew dispute"
}
```

Valid statuses:
- `needs_response` - Needs evidence submission
- `warning_needs_response` - Pre-chargeback warning
- `under_review` - Under review by bank
- `won` - Dispute won
- `lost` - Dispute lost
- `closed` - Closed

#### Submit Evidence
```
POST /v1/admin/billing/disputes/{disputeId}/evidence
```

Request Body (EvidenceDetails):
```json
{
  "product_description": "FunctionFly subscription - Professional Plan",
  "customer_email": "customer@example.com",
  "customer_name": "John Doe",
  "customer_purchase_ip": "203.0.113.1",
  "billing_address": "123 Main St, City, Country",
  "receipt_url": "https://receipt.stripe.com/rcpt_xxx",
  "service_date": "2026-04-01",
  "refund_policy_url": "https://functionfly.com/refund-policy",
  "refund_policy_disclosed": true,
  "customer_communication": "Customer acknowledged service usage in email on 2026-04-01",
  "access_activity_log": "User accessed dashboard 25 times during subscription period"
}
```

Note: Evidence must be submitted through Stripe's dashboard or API. This endpoint records the submitted evidence for tracking purposes.

#### Get Dispute Statistics
```
GET /v1/admin/billing/disputes/stats
```

Returns:
```json
{
  "total_disputes": 50,
  "open_disputes": 3,
  "won_disputes": 35,
  "lost_disputes": 12,
  "total_disputed_cents": 250000,
  "by_status": {
    "needs_response": 2,
    "won": 35,
    "lost": 12,
    "under_review": 1
  }
}
```

### Refund Management

#### List Refunds
```
GET /v1/admin/billing/refunds
```

Query Parameters:
- `limit` - Maximum results (default 50, max 500)
- `offset` - Pagination offset
- `status` - Filter by status
- `reason` - Filter by reason
- `tenant_id` - Filter by tenant
- `start_date` / `end_date` - Date range filter

#### Get Refund Details
```
GET /v1/admin/billing/refunds/{refundId}
```

#### Get Refund Statistics
```
GET /v1/admin/billing/refunds/stats
```

Returns:
```json
{
  "total_refunds": 100,
  "total_refunded_cents": 500000,
  "recent_refund_count": 15,
  "by_status": {
    "succeeded": 95,
    "failed": 3,
    "pending": 2
  },
  "by_reason": {
    "requested_by_customer": 80,
    "duplicate": 10,
    "fraudulent": 10
  }
}
```

### Chargeback Reconciliation

#### Get Financial Impact
```
GET /v1/admin/billing/chargebacks/reconciliation
```

Query Parameters:
- `start_date` / `end_date` - Date range for reconciliation

Returns:
```json
{
  "total_disputed_cents": 250000,
  "total_refunded_cents": 0,
  "net_chargeback_impact_cents": 265000,
  "dispute_fee_cents": 15000,
  "recovery_rate": 70.0
}
```

Note: Stripe charges a $15 fee per dispute, included in the calculation.

## Notifications

The system sends notifications for the following events:

### Dispute Created
- **Type**: `billing.dispute_created`
- **Priority**: Urgent
- **Channels**: In-app, Email, Webhook
- **Recipients**: Admin users
- **Content**: New dispute filed, amount, deadline for evidence

### Evidence Due Soon
- **Type**: `billing.dispute_evidence_due`
- **Priority**: Urgent
- **Channels**: In-app, Email, Webhook
- **Content**: Warning that evidence deadline is approaching (24h, 72h)

### Dispute Resolved
- **Type**: `billing.dispute_resolved`
- **Priority**: High
- **Channels**: In-app, Email
- **Content**: Outcome (won/lost), amount

### Refund Processed
- **Type**: `billing.refund_processed`
- **Priority**: Normal
- **Channels**: In-app, Webhook
- **Content**: Refund details, reason

### Funds Withdrawn
- **Type**: `billing.chargeback_funds_withdrawn`
- **Priority**: High
- **Channels**: In-app, Email, Webhook
- **Content**: Amount withdrawn, dispute fee charged

## Best Practices

### Monitoring

1. **Set up alerts** for `charge.dispute.created` events to respond quickly
2. **Monitor the open disputes endpoint** daily for items requiring action
3. **Review dispute statistics weekly** to identify trends

### Responding to Disputes

1. **Submit evidence promptly** - Most disputes allow 7-21 days for response
2. **Include comprehensive evidence**:
   - Customer communication showing service usage
   - Billing agreement or terms of service acceptance
   - Delivery/usage logs
   - Refund policy disclosure
3. **Track submission** using the evidence endpoint for audit purposes

### Preventing Disputes

1. **Clear billing descriptors** - Ensure customers recognize charges
2. **Receipts and confirmations** - Send immediately after payment
3. **Clear refund policy** - Make it easily accessible
4. **Proactive customer service** - Resolve issues before they become disputes

## Accounting Integration

The dispute and refund data can be used for:

- Monthly revenue reconciliation
- Identifying high-risk customers or payment patterns
- Calculating true customer lifetime value (including chargebacks)
- Financial reporting and forecasting

### Example: Monthly Reconciliation

```sql
-- Total disputed amount this month
SELECT 
    SUM(amount_cents) / 100.0 as disputed_usd,
    COUNT(*) as dispute_count
FROM payment_disputes
WHERE created_at >= DATE_TRUNC('month', NOW())
  AND created_at < DATE_TRUNC('month', NOW() + INTERVAL '1 month');

-- Won vs Lost disputes
SELECT 
    outcome,
    COUNT(*),
    SUM(amount_cents) / 100.0 as amount_usd
FROM payment_disputes
WHERE created_at >= DATE_TRUNC('month', NOW())
GROUP BY outcome;

-- Total refunds
SELECT 
    SUM(amount_cents) / 100.0 as refunded_usd,
    COUNT(*) as refund_count
FROM payment_refunds
WHERE created_at >= DATE_TRUNC('month', NOW());
```

## Troubleshooting

### Webhook Not Received

1. Check Stripe dashboard webhook settings
2. Verify `STRIPE_WEBHOOK_SECRET` is configured
3. Check application logs for webhook processing errors
4. Verify the webhook endpoint is publicly accessible

### Missing Tenant/User Association

Disputes and refunds are linked to tenants/users via charge metadata. If the metadata is missing:
1. The records are still stored
2. Use the admin API to manually associate if needed
3. Review payment flow to ensure metadata is consistently included

### Notification Failures

Notifications depend on the notification service. If notifications aren't sent:
1. Check notification service configuration
2. Verify admin users exist with proper permissions
3. Check notification service logs
