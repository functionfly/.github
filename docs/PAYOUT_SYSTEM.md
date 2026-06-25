# Agent Payouts System

## Overview

The Agent Payouts system enables agents to receive automatic payouts for their earnings via Stripe Connect Express accounts.

## Quick Start

```bash
# Required environment variables
export STRIPE_SECRET_KEY=sk_live_...        # Stripe live/test secret key
export STRIPE_WEBHOOK_SECRET=whsec_...       # Stripe webhook signing secret

# Optional: Enable manual approval workflow (default: automatic)
export PAYOUT_MANUAL_APPROVAL_ENABLED=false

# Start the API
./bin/orchestrator-api --skip-migrations
```

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `STRIPE_SECRET_KEY` | Stripe API secret key (sk_live_... or sk_test_...) | `sk_live_...` |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret | `whsec_...` |

### Optional - Payout Processing

| Variable | Default | Description |
|----------|---------|-------------|
| `PAYOUT_MANUAL_APPROVAL_ENABLED` | `false` | Enable manual approval workflow for high-value payouts |
| `PAYOUT_APPROVAL_THRESHOLD_USD` | `1000.0` | Payout amount (USD) requiring first approval |
| `PAYOUT_SECOND_APPROVAL_THRESHOLD_USD` | `10000.0` | Payout amount (USD) requiring second approval |
| `PAYOUT_MAX_PER_DAY` | `5` | Maximum number of payouts per user per 24 hours |
| `PAYOUT_MAX_AMOUNT_PER_DAY_CENTS` | `1000000` ($10,000) | Maximum total payout amount per user per 24 hours |
| `PAYOUT_MIN_INTERVAL_MINUTES` | `60` | Minimum minutes between payout requests |

### Optional - Auto-Payout Scheduler

| Variable | Default | Description |
|----------|---------|-------------|
| `PAYOUT_SCHEDULER_ENABLED` | `false` | Enable automatic scheduled payouts |
| `PAYOUT_SCHEDULER_CRON` | `0 2 * * 1` | Cron expression (default: Monday 2 AM UTC) |
| `PAYOUT_SCHEDULER_TIMEZONE` | `UTC` | Timezone for scheduler |

## Processing Modes

### Automatic (Default)

When `PAYOUT_MANUAL_APPROVAL_ENABLED=false`:
1. User requests payout
2. System validates balance, velocity limits, Stripe Connect status
3. Stripe Transfer created immediately
4. Payout marked as `completed` or `failed`

### Manual Approval (Opt-in)

When `PAYOUT_MANUAL_APPROVAL_ENABLED=true`:
1. User requests payout
2. System evaluates approval rules
3. If approval required: payout status = `pending_approval`
4. Admin reviews and approves (may require 2 levels for high-value)
5. After final approval: Stripe Transfer created
6. Payout marked as `completed`

## Approval Rules

The system supports configurable approval thresholds:

| Threshold | Approval Required |
|-----------|------------------|
| $0 - $999.99 | None (automatic) |
| $1,000 - $9,999.99 | 1 approval |
| $10,000+ | 2 approvals |

Default approver roles: `finance`, `admin`

## API Endpoints

### User Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/payouts/connect-account` | Get Connect account status |
| POST | `/v1/payouts/connect-account` | Start Stripe onboarding |
| POST | `/v1/payouts/connect-account/refresh` | Refresh account status |
| GET | `/v1/payouts/balance` | Get payout balance |
| POST | `/v1/payouts/request` | Request payout |
| POST | `/v1/payouts/{id}/cancel` | Cancel payout |
| GET | `/v1/payouts/requests` | List payout requests |
| GET | `/v1/payouts/history` | Enhanced history with fees |
| GET | `/v1/payouts/schedule` | Get schedule preference |
| PUT | `/v1/payouts/schedule` | Update schedule preference |

### Admin Endpoints

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/v1/admin/payouts/pending` | `billing:read` | List pending approvals |
| POST | `/v1/admin/payouts/{id}/approve` | `billing:write` | Approve payout |
| POST | `/v1/admin/payouts/{id}/reject` | `billing:write` | Reject payout |
| GET | `/v1/admin/payouts/approval-rules` | `billing:read` | List approval rules |
| POST | `/v1/admin/payouts/approval-rules` | `billing:write` | Create approval rule |

## Webhook Events

The system handles these Stripe webhook events:

| Event | Action |
|-------|--------|
| `payout.paid` | Mark payout as completed |
| `payout.failed` | Mark payout as failed, refresh account |
| `transfer.reversed` | Reverse ledger debit |
| `account.updated` | Update Connect account status |

## Database Schema

### Key Tables

- `stripe_connect_accounts` - Per-user Stripe Express accounts
- `payout_requests` - Withdrawal requests with status tracking
- `payout_ledger` - Immutable audit trail of fund movements
- `payout_approval_records` - Approval workflow records
- `payout_approval_audit` - Approval action audit log
- `payout_schedule_preferences` - Auto-payout configuration
- `payout_velocity_tracking` - Fraud prevention tracking
- `payout_fee_deductions` - Fee deduction records

## Security

- All user endpoints require JWT authentication
- Admin endpoints require `billing:read` or `billing:write` permission
- Sensitive admin operations require HMAC signature
- Idempotency keys prevent duplicate payouts
- Velocity limits prevent fraud
- Immutable ledger provides audit trail

## Troubleshooting

### Payout stuck in "pending_approval"

1. Check `PAYOUT_MANUAL_APPROVAL_ENABLED` is set correctly
2. Verify approval service is initialized (check logs)
3. Use admin UI to approve/reject the payout

### "No connected account found"

1. User must complete Stripe Connect onboarding first
2. Check `stripe_connect_accounts` table for user's account
3. Verify `details_submitted` and `payouts_enabled` are true

### Stripe transfer failed

1. Check Stripe Dashboard for transfer status
2. Verify webhook is receiving events
3. Check `STRIPE_WEBHOOK_SECRET` is correct
