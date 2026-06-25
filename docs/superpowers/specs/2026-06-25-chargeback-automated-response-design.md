# Chargeback Automated Response - Production Design

**Date:** 2026-06-25
**Status:** Approved for Implementation

---

## Overview

Automated response system for Stripe chargebacks/disputes that compiles evidence, optionally auto-refunds small disputes, and notifies all relevant parties.

---

## Architecture

```
Stripe Webhooks                     DisputeResponseManager
─────────────────                   ──────────────────────
dispute.created ──────────────────► handleDisputeCreated()
dispute.updated ──────────────────► handleDisputeUpdated()
dispute.closed  ──────────────────► handleDisputeClosed()

Components:
  ├─ EvidenceCompiler       (compiles invoice, logs, comms, subscription data)
  ├─ AutoRefundProcessor    (threshold-based auto-refund decision)
  └─ NotificationWorkflow   (customer + admin notifications)
```

---

## Components

### 1. DisputeResponseManager

**File:** `internal/billing/dispute_response_manager.go`

- Background processor (goroutine-based, like DunningManager)
- Handles webhook events: `dispute.created`, `dispute.updated`, `dispute.closed`
- Orchestrates evidence compilation, refund decisions, notifications
- Idempotent via existing `UpsertDispute()` pattern
- Graceful shutdown support

### 2. EvidenceCompiler

**File:** `internal/billing/evidence_compiler.go`

- `CompileDisputeEvidence(disputeID) (*stripe.DisputeEvidenceParams, error)`
- Sources compiled:
  - Invoice/receipt data from `invoices` table
  - Execution/usage logs from `execution_logs` table
  - Customer communication history from `support_tickets` / `messages` tables
  - Subscription details from `subscriptions` table (plan, terms, signup timestamp)
- Returns Stripe-formatted evidence struct ready for submission

### 3. AutoRefundProcessor

**File:** `internal/billing/dispute_response_manager.go` (integrated)

- Threshold: $50 USD (5000 cents)
- Logic:
  - If `amount_cents < 5000` AND `reason IN ('duplicate', 'product_not_received')` → auto-refund
  - If `reason == 'fraud'` → NEVER auto-refund
- Configurable via `CHARGEBACK_AUTO_REFUND_THRESHOLD` env var
- Logs all decisions to `dispute_automation_log` table
- Supports manual override via API

### 4. NotificationWorkflow

**Extends:** `internal/notification/service.go`

**Customer notifications:**
- Dispute filed: "A chargeback was filed. We're investigating."
- Evidence submitted: "We've submitted evidence on your behalf."
- Dispute won: "The dispute was resolved in your favor."
- Dispute lost: "The dispute was not resolved. Contact support."

**Admin notifications:**
- New dispute alert (all details)
- Evidence due soon (X days remaining)
- Auto-refund executed
- Dispute resolved (won/lost with amount)

---

## Database Changes

### Table: `dispute_automation_log`

```sql
CREATE TABLE IF NOT EXISTS dispute_automation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES payment_disputes(id),
    action VARCHAR(50) NOT NULL,
    outcome VARCHAR(50) NOT NULL,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT now()
);
```

### Table: `dispute_automation_config`

```sql
CREATE TABLE IF NOT EXISTS dispute_automation_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auto_refund_enabled BOOLEAN DEFAULT true,
    auto_refund_threshold_cents INTEGER DEFAULT 5000,
    auto_refund_allowed_reasons TEXT[] DEFAULT ARRAY['duplicate', 'product_not_received'],
    evidence_auto_submit BOOLEAN DEFAULT true,
    customer_notification_enabled BOOLEAN DEFAULT true,
    admin_escalation_enabled BOOLEAN DEFAULT true,
    admin_escalation_threshold_cents INTEGER DEFAULT 50000,
    updated_at TIMESTAMP DEFAULT now()
);
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/admin/billing/disputes/automated-config` | Get automation settings |
| PATCH | `/v1/admin/billing/disputes/automated-config` | Update thresholds/rules |
| POST | `/v1/admin/billing/disputes/{id}/skip-auto-refund` | Skip auto-refund for dispute |
| POST | `/v1/admin/billing/disputes/{id}/force-refund` | Force refund (override threshold) |
| GET | `/v1/admin/billing/disputes/{id}/evidence-preview` | Preview compiled evidence |
| GET | `/v1/admin/billing/disputes/{id}/automation-log` | Get automation log for dispute |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CHARGEBACK_AUTO_REFUND_THRESHOLD` | `50.00` USD | Max amount to auto-refund |
| `CHARGEBACK_AUTO_REFUND_ENABLED` | `true` | Enable/disable auto-refund |
| `CHARGEBACK_EVIDENCE_AUTO_SUBMIT` | `true` | Auto-submit evidence to Stripe |
| `CHARGEBACK_ADMIN_ESCALATION_ENABLED` | `true` | Escalate large disputes to admin |

---

## Workflow State Machine

```
dispute.created
    │
    ├─► [auto_refund eligible] ──► AutoRefundProcessor
    │                                    │
    │                                    ▼
    │                             stripe.Refund()
    │                                    │
    │                                    ▼
    │                             Log → Notify(customer, admin)
    │
    ├─► [evidence_auto_submit] ──► EvidenceCompiler
    │                                    │
    │                                    ▼
    │                             stripe.Dispute.SubmitEvidence()
    │                                    │
    │                                    ▼
    │                             Log → Notify(admins)
    │
    └─► Notify(admins: new dispute, all details)

dispute.updated (status change)
    │
    └─► Log status → Notify(customer, admins)

dispute.closed
    │
    ├─► won  ──► Log → Notify(customer(congrats), admins)
    └─► lost ──► Log → Notify(customer(apology), admins)
```

---

## Error Handling

| Scenario | Action |
|----------|--------|
| Evidence compilation fails | Log error, alert admin, DO NOT auto-submit, keep as `needs_response` |
| Stripe API call fails | Retry 3x with exponential backoff, then alert admin |
| Auto-refund fails | Mark dispute for manual review, alert admin |
| Duplicate webhook events | Idempotent via `UpsertDispute` |

---

## Implementation Phases

### Phase 1: Core Infrastructure
- `DisputeResponseManager` skeleton with background processor
- Webhook handler registration in Stripe webhook controller
- Basic logging

### Phase 2: Evidence Compilation
- `EvidenceCompiler` implementation
- Integration with existing invoice, execution logs, support tables
- Evidence submission to Stripe

### Phase 3: Auto-Refund
- `AutoRefundProcessor` with threshold logic
- Stripe refund API integration
- Decision logging

### Phase 4: Notifications
- Customer notification templates
- Admin escalation alerts
- Resolution notifications

### Phase 5: API & Config
- Admin config endpoints
- Manual override endpoints
- Evidence preview endpoint

### Phase 6: Database & Polish
- Migration scripts
- Config management
- Testing & error handling

---

## Files to Create/Modify

### New Files
- `internal/billing/dispute_response_manager.go`
- `internal/billing/evidence_compiler.go`
- `internal/api/handlers/admin/dispute_automation_handlers.go`
- `migrations/20260625150000_dispute_automation_tables.up.sql`
- `migrations/20260625150000_dispute_automation_tables.down.sql`

### Files to Modify
- `internal/api/routes_admin.go` - Add new routes
- `internal/api/handlers/webhooks/stripe.go` - Register dispute response manager
- `internal/notification/service.go` - Add dispute notification methods
- `internal/billing/dunning_manager.go` - Follow same pattern for structure
- `internal/storage/dispute_repository.go` - Add automation log queries

---

## Dependencies

- Existing `DisputeRepository` for dispute CRUD
- Existing `NotificationService` for sending alerts
- Stripe Go SDK (`github.com/stripe/stripe-go/v76`) - already in use
- Existing `invoice_repository`, `execution_log_repository` for evidence data
