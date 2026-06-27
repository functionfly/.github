# Chargeback Automated Response System - Design

**Date**: 2026-06-27
**Status**: Approved for Implementation
**Type**: Production Feature

---

## Overview

Implement a full production Chargeback Automated Response system with dispute response workflow. The system automates handling of Stripe chargebacks while protecting revenue through intelligent decision-making and human oversight for high-value disputes.

---

## Architecture

### Components

1. **DisputeResponseManager** (`internal/billing/dispute_response_manager.go`)
   - Existing, needs enhancement and integration
   - Handles automation decisions, evidence compilation, Stripe API calls

2. **DisputeScheduler** (`internal/scheduler/dispute_scheduler.go`) - NEW
   - Monitors evidence deadlines
   - Escalates approaching deadlines
   - Processes pending evidence submissions

3. **Webhook Integration** (`internal/api/handlers/webhooks/stripe.go`)
   - Wire DisputeResponseManager into webhook flow
   - Trigger automation on dispute events

4. **Admin API Handlers** (`internal/api/handlers/billing/disputes.go`) - NEW
   - List/view disputes with filtering
   - Preview compiled evidence
   - Approve/reject evidence submission
   - Manual override actions

5. **NotificationService** (`internal/notification/service.go`)
   - Enhance customer notifications (currently stubs)
   - Admin escalation notifications

---

## Decision Matrix

| Dispute Amount | Reason | Decision | Action |
|---------------|--------|----------|--------|
| < $25 | Any | Auto-refund | Immediate refund, notify admin |
| $25 - $150 | duplicate, product_not_received | Auto-contest | Compile evidence, submit automatically |
| $25 - $150 | fraudulent | Manual review | Flag for admin, await decision |
| $25 - $150 | subscription_canceled | Manual review | Flag for admin, await decision |
| > $150 | Any | Escalation | Admin notification with recommendation |
| Any | High fraud score | Escalation | Flag for fraud team |

---

## Evidence Package

### Components

1. **Receipt/Invoice**
   - Fetch from Stripe API using `dispute.Charge.ID`
   - Include payment amount, date, description

2. **Platform Activity Log (30 days)**
   - API calls, function executions
   - Logins and session data
   - Feature usage patterns

3. **Support Communications**
   - Support tickets related to the charge
   - Email correspondence about the dispute reason

4. **Subscription Terms**
   - Plan details at time of purchase
   - Terms agreed at signup

5. **Customer Info**
   - Name, email, billing address
   - Account age and status

---

## Automation Workflow

### On `charge.dispute.created`

1. Parse dispute from Stripe webhook
2. Lookup tenant/customer from charge metadata
3. Create/update `payment_disputes` record
4. Log event in `dispute_automation_log`
5. Apply decision matrix:
   - **Auto-refund**: Execute refund immediately
   - **Auto-contest**: Compile and submit evidence
   - **Manual review**: Notify admins, set `pending_review` status
6. Send customer notification (for auto-refund)

### On `charge.dispute.updated`

1. Update dispute status in database
2. Log status change
3. Notify customer of status change
4. If `needs_review` status, re-escalate to admin

### On `charge.dispute.closed`

1. Update final status and outcome
2. If **won**: Log victory, notify customer
3. If **lost**:
   - Notify customer with explanation
   - Record in analytics
   - If high-value: flag for fraud analysis
   - If repeat offender (2+ in 6 months): flag account

---

## Scheduler Jobs

### EvidenceDeadlineMonitor
- **Schedule**: Every 4 hours
- **Action**: Query disputes with `evidence_due_by` within 48 hours
- **Result**: Send escalation notification to admins

### PendingReviewChecker
- **Schedule**: Every hour
- **Action**: Find disputes pending review > 24 hours
- **Result**: Send reminder to admins

### FraudPatternAnalyzer
- **Schedule**: Daily at 6 AM
- **Action**: Analyze disputes in last 90 days for fraud patterns
- **Result**: Flag accounts with 2+ disputes, mark for fraud review

---

## API Endpoints

### Admin Endpoints (Protected, admin role required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/disputes` | List disputes with filters |
| GET | `/api/admin/disputes/{id}` | Get dispute details |
| GET | `/api/admin/disputes/{id}/evidence` | Preview compiled evidence |
| POST | `/api/admin/disputes/{id}/submit` | Submit evidence to Stripe |
| POST | `/api/admin/disputes/{id}/refund` | Issue manual refund |
| POST | `/api/admin/disputes/{id}/skip` | Skip automated response |
| GET | `/api/admin/disputes/stats` | Dispute statistics |

### Internal Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/internal/disputes/automation-log/{id}` | Get automation log |

---

## Database Changes

### New Table: `dispute_automation_config`

```sql
CREATE TABLE dispute_automation_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auto_refund_enabled BOOLEAN DEFAULT true,
    auto_refund_threshold_cents INT DEFAULT 2500,
    auto_refund_allowed_reasons TEXT[] DEFAULT ARRAY['duplicate', 'product_not_received'],
    evidence_auto_submit BOOLEAN DEFAULT false,
    evidence_auto_submit_threshold_cents INT DEFAULT 15000,
    manual_review_threshold_cents INT DEFAULT 15000,
    customer_notification_enabled BOOLEAN DEFAULT true,
    admin_escalation_enabled BOOLEAN DEFAULT true,
    admin_escalation_threshold_cents INT DEFAULT 15000,
    fraud_detection_enabled BOOLEAN DEFAULT true,
    repeat_offender_window_days INT DEFAULT 180,
    repeat_offender_threshold INT DEFAULT 2,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### New Table: `dispute_customer_notifications`

```sql
CREATE TABLE dispute_customer_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES payment_disputes(id),
    notification_type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    content JSONB,
    success BOOLEAN DEFAULT true,
    error_message TEXT
);
```

### Indexes

```sql
CREATE INDEX idx_payment_disputes_evidence_due_by ON payment_disputes(evidence_due_by) WHERE evidence_due_by IS NOT NULL;
CREATE INDEX idx_payment_disputes_status_pending ON payment_disputes(status) WHERE status = 'pending_review';
CREATE INDEX idx_dispute_automation_log_dispute_id ON dispute_automation_log(dispute_id);
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CHARGEBACK_AUTO_REFUND_ENABLED` | `true` | Enable auto-refund |
| `CHARGEBACK_AUTO_REFUND_THRESHOLD_CENTS` | `2500` | Auto-refund threshold ($25) |
| `CHARGEBACK_AUTO_REFUND_ALLOWED_REASONS` | `duplicate,product_not_received` | Reasons to auto-refund |
| `CHARGEBACK_EVIDENCE_AUTO_SUBMIT` | `false` | Auto-submit evidence |
| `CHARGEBACK_EVIDENCE_AUTO_SUBMIT_THRESHOLD_CENTS` | `15000` | Threshold for auto-submit ($150) |
| `CHARGEBACK_MANUAL_REVIEW_THRESHOLD_CENTS` | `15000` | Manual review threshold ($150) |
| `CHARGEBACK_ADMIN_ESCALATION_ENABLED` | `true` | Enable admin escalation |
| `CHARGEBACK_ADMIN_ESCALATION_THRESHOLD_CENTS` | `15000` | Escalation threshold ($150) |
| `CHARGEBACK_CUSTOMER_NOTIFICATION_ENABLED` | `true` | Send customer notifications |
| `CHARGEBACK_FRAUD_DETECTION_ENABLED` | `true` | Enable fraud pattern detection |
| `CHARGEBACK_REPEAT_OFFENDER_WINDOW_DAYS` | `180` | Lookback window for repeat detection |
| `CHARGEBACK_REPEAT_OFFENDER_THRESHOLD` | `2` | Disputes to flag as repeat offender |

---

## Error Handling

### Retry Logic
- Evidence submission failure: Retry 3 times with exponential backoff
- Refund failure: Log error, notify admin immediately
- Webhook processing failure: Return error to Stripe for retry

### Dead Letter Queue
- Failed automation actions logged to `dispute_automation_log` with `failed` outcome
- Admin notification on repeated failures

---

## Monitoring & Alerting

### Metrics
- `chargeback_auto_refund_total` - Counter by reason
- `chargeback_auto_contest_total` - Counter
- `chargeback_manual_review_pending` - Gauge
- `chargeback_evidence_deadline_approaching` - Gauge
- `chargeback_lost_total` - Counter
- `chargeback_won_total` - Counter

### Alerts
- Manual review pending > 48 hours
- Evidence deadline < 24 hours with no action
- Lost disputes > threshold in 24 hours
- Fraud pattern detected

---

## Implementation Phases

### Phase 1: Core Integration
- Wire DisputeResponseManager into webhook handler
- Implement evidence compilation enhancement
- Add customer notifications
- Basic admin API handlers

### Phase 2: Scheduler & Escalation
- Dispute scheduler for deadline monitoring
- Admin escalation notifications
- Fraud pattern detection

### Phase 3: Admin Dashboard UI
- Dispute list and detail views
- Evidence preview and submission UI
- Statistics dashboard

---

## Security Considerations

- All admin endpoints require authentication and admin role
- Stripe webhook signature verification is mandatory
- Evidence data contains PII - handle according to data policy
- Audit logging for all manual overrides
- Rate limiting on admin API endpoints

---

## Testing Strategy

- Unit tests for decision matrix logic
- Integration tests for Stripe webhook handling
- Mock Stripe API for evidence submission tests
- Admin API endpoint tests with authentication
- Scheduler job tests with time-based triggers
