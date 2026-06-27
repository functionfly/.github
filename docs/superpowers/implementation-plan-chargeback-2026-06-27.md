# Chargeback Automated Response - Implementation Plan

**Date**: 2026-06-27
**Design Doc**: `docs/superpowers/specs/2026-06-27-chargeback-automated-response-design.md`

---

## Implementation Overview

### Files to Create
- `internal/scheduler/dispute_scheduler.go` - Scheduler for deadline monitoring
- `internal/api/handlers/billing/disputes.go` - Admin API handlers
- `migrations/YYYYMMDDHHMMSS_enhance_dispute_automation.sql` - Database changes

### Files to Modify
- `internal/api/handlers/webhooks/stripe.go` - Wire DisputeResponseManager
- `internal/billing/dispute_response_manager.go` - Enhance evidence, implement notifications
- `internal/notification/service.go` - Add dispute customer notification methods
- `internal/api/routes_admin.go` - Register dispute admin routes
- `internal/api/routes.go` - Register routes if needed

---

## Phase 1: Core Integration

### Task 1.1: Wire DisputeResponseManager into Webhook Handler

**File**: `internal/api/handlers/webhooks/stripe.go`

1. Add `disputeResponseManager *billingpkg.DisputeResponseManager` field to `StripeWebhookHandler`
2. Add `SetDisputeResponseManager()` setter method
3. In `handleChargeDisputeCreated`:
   - Call `h.disputeResponseManager.HandleDisputeCreated(ctx, &dispute, paymentDispute)`
4. In `handleChargeDisputeUpdated`:
   - Call `h.disputeResponseManager.HandleDisputeUpdated(ctx, &dispute, paymentDispute)`
5. In `handleChargeDisputeClosed`:
   - Call `h.disputeResponseManager.HandleDisputeClosed(ctx, &dispute, paymentDispute)`

**File**: `internal/api/routes.go` or wherever webhook handler is wired

1. After creating `disputeResponseManager`, call `webhookHandler.SetDisputeResponseManager(disputeResponseManager)`

### Task 1.2: Enhance Evidence Compilation

**File**: `internal/billing/dispute_response_manager.go`

Update `compileEvidenceData()` to pull real data:

```go
func (m *DisputeResponseManager) compileEvidenceData(ctx context.Context, paymentDispute *storage.PaymentDispute) (*CompiledEvidence, error) {
    evidence := &CompiledEvidence{}

    // 1. Get tenant info for billing address and name
    if paymentDispute.TenantID != nil {
        tenant, err := m.getTenantInfo(ctx, *paymentDispute.TenantID)
        if err == nil && tenant != nil {
            evidence.CustomerName = tenant.Name
            evidence.BillingAddress = tenant.BillingAddress
        }
    }

    // 2. Get user info for email, IP, name
    if paymentDispute.UserID != nil {
        user, err := m.getUserInfo(ctx, *paymentDispute.UserID)
        if err == nil && user != nil {
            evidence.CustomerEmail = user.Email
            if evidence.CustomerName == "" {
                evidence.CustomerName = user.Name
            }
            evidence.CustomerPurchaseIP = user.LastLoginIP
        }
    }

    // 3. Get receipt URL from Stripe or local invoice
    evidence.ReceiptURL = m.getReceiptURL(ctx, paymentDispute)

    // 4. Get subscription details
    subscriptionInfo := m.getSubscriptionInfo(ctx, paymentDispute)
    evidence.ProductDescription = fmt.Sprintf("FunctionFly AI Platform - %s - %s",
        paymentDispute.Reason, subscriptionInfo.PlanName)

    // 5. Get 30-day activity log
    evidence.AccessActivityLog = m.getAccessActivityLog(ctx, paymentDispute)

    // 6. Get support ticket communications
    evidence.CustomerCommunication = m.getSupportCommunications(ctx, paymentDispute)

    // 7. Set service date to first of month
    evidence.ServiceDate = time.Now().Format("2006-01-02")

    // 8. Standard refund policy info
    evidence.RefundPolicyURL = "https://functionfly.com/refund-policy"
    evidence.RefundPolicyDisclosed = true
    evidence.RefundPolicyDisclosedText = "Customers are shown our refund policy at checkout."

    return evidence, nil
}
```

New helper methods needed:
- `getReceiptURL(ctx, paymentDispute)` - Look up invoice by stripe_charge_id
- `getSubscriptionInfo(ctx, paymentDispute)` - Get plan name, term details
- `getSupportCommunications(ctx, paymentDispute)` - Get recent support tickets

### Task 1.3: Implement Customer Notifications

**File**: `internal/notification/service.go`

Add new methods:

```go
func (s *Service) SendDisputeCustomerUpdate(ctx context.Context, userID uuid.UUID, disputeID, status, message string) error
func (s *Service) SendDisputeCustomerResolved(ctx context.Context, userID uuid.UUID, disputeID string, won bool, amountUSD float64) error
func (s *Service) SendDisputeCustomerEvidenceReceived(ctx context.Context, userID uuid.UUID, disputeID string) error
```

**File**: `internal/billing/dispute_response_manager.go`

Update stub implementations:

```go
func (m *DisputeResponseManager) notifyCustomerStatusChange(ctx context.Context, paymentDispute *storage.PaymentDispute, newStatus string) {
    if m.notificationSvc == nil || paymentDispute.UserID == nil {
        return
    }

    // Get user email
    user, err := m.getUserInfo(ctx, *paymentDispute.UserID)
    if err != nil || user == nil || user.Email == "" {
        logrus.WithField("user_id", paymentDispute.UserID).Warn("Cannot notify customer: user not found")
        return
    }

    message := getCustomerStatusMessage(newStatus, paymentDispute.Reason)
    err = m.notificationSvc.SendDisputeCustomerUpdate(ctx, *paymentDispute.UserID, paymentDispute.StripeDisputeID, newStatus, message)
    if err != nil {
        logrus.WithError(err).WithField("dispute_id", paymentDispute.StripeDisputeID).Warn("Failed to notify customer")
    }
}
```

Add helper `getCustomerStatusMessage(status, reason string) string`

### Task 1.4: Create Admin API Handlers

**File**: `internal/api/handlers/billing/disputes.go` (new file)

```go
package billing

import (
    "github.com/functionfly/functionfly/internal/billing"
    "github.com/functionfly/functionfly/internal/storage"
    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"
)

type DisputeHandler struct {
    disputeResponseManager *billingpkg.DisputeResponseManager
    disputeRepo           *storage.DisputeRepository
}

func NewDisputeHandler(drm *billingpkg.DisputeResponseManager, dr *storage.DisputeRepository) *DisputeHandler {
    return &DisputeHandler{
        disputeResponseManager: drm,
        disputeRepo:            dr,
    }
}

func (h *DisputeHandler) RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/admin/disputes", h.ListDisputes).Methods("GET")
    r.HandleFunc("/admin/disputes/open", h.ListOpenDisputes).Methods("GET")
    r.HandleFunc("/admin/disputes/stats", h.GetStats).Methods("GET")
    r.HandleFunc("/admin/disputes/{disputeId}", h.GetDispute).Methods("GET")
    r.HandleFunc("/admin/disputes/{disputeId}/evidence", h.GetEvidence).Methods("GET")
    r.HandleFunc("/admin/disputes/{disputeId}/submit", h.SubmitEvidence).Methods("POST")
    r.HandleFunc("/admin/disputes/{disputeId}/refund", h.IssueRefund).Methods("POST")
    r.HandleFunc("/admin/disputes/{disputeId}/skip", h.SkipAutoResponse).Methods("POST")
    r.HandleFunc("/admin/disputes/{disputeId}/automation-log", h.GetAutomationLog).Methods("GET")
}
```

Implement all handler methods with proper auth checks.

**File**: `internal/api/routes_admin.go`

Add route registration:
```go
disputeHandler := billing.NewDisputeHandler(disputeResponseManager, disputeRepo)
disputeHandler.RegisterRoutes(adminRouter)
```

### Task 1.5: Database Migration

**File**: `migrations/YYYYMMDDHHMMSS_enhance_dispute_automation.up.sql`

```sql
-- dispute_automation_config table
CREATE TABLE IF NOT EXISTS dispute_automation_config (
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

-- dispute_customer_notifications table
CREATE TABLE IF NOT EXISTS dispute_customer_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES payment_disputes(id),
    notification_type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    content JSONB,
    success BOOLEAN DEFAULT true,
    error_message TEXT
);

-- Insert default config
INSERT INTO dispute_automation_config (id) VALUES (gen_random_uuid()) ON CONFLICT DO NOTHING;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payment_disputes_evidence_due_by ON payment_disputes(evidence_due_by) WHERE evidence_due_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_disputes_status_pending ON payment_disputes(status) WHERE status = 'pending_review';
CREATE INDEX IF NOT EXISTS idx_dispute_automation_log_dispute_id ON dispute_automation_log(dispute_id);
CREATE INDEX IF NOT EXISTS idx_dispute_customer_notifications_dispute_id ON dispute_customer_notifications(dispute_id);
```

**Down migration** (`YYYYMMDDHHMMSS_enhance_dispute_automation.down.sql`):
```sql
DROP TABLE IF EXISTS dispute_customer_notifications;
DROP TABLE IF EXISTS dispute_automation_config;
DROP INDEX IF EXISTS idx_payment_disputes_evidence_due_by;
DROP INDEX IF EXISTS idx_payment_disputes_status_pending;
DROP INDEX IF EXISTS idx_dispute_automation_log_dispute_id;
DROP INDEX IF EXISTS idx_dispute_customer_notifications_dispute_id;
```

---

## Phase 2: Scheduler & Escalation

### Task 2.1: Create Dispute Scheduler

**File**: `internal/scheduler/dispute_scheduler.go`

```go
package scheduler

import (
    "context"
    "time"

    "github.com/functionfly/functionfly/internal/billing"
    "github.com/functionfly/functionfly/internal/notification"
    "github.com/functionfly/functionfly/internal/storage"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
)

type DisputeScheduler struct {
    disputeRepo        *storage.DisputeRepository
    disputeRespManager *billing.DisputeResponseManager
    notificationSvc    *notification.Service
    stop               chan struct{}
}

func NewDisputeScheduler(dr *storage.DisputeRepository, drm *billing.DisputeResponseManager, ns *notification.Service) *DisputeScheduler {
    return &DisputeScheduler{
        disputeRepo:        dr,
        disputeRespManager: drm,
        notificationSvc:    ns,
        stop:               make(chan struct{}),
    }
}

func (s *DisputeScheduler) Start() {
    // EvidenceDeadlineMonitor - every 4 hours
    go s.runPeriodic(s.checkEvidenceDeadlines, 4*time.Hour, "evidence_deadline_check")

    // PendingReviewChecker - every hour
    go s.runPeriodic(s.checkPendingReviews, 1*time.Hour, "pending_review_check")

    // FraudPatternAnalyzer - daily at 6 AM
    go s.runDaily(6, 0, s.analyzeFraudPatterns, "fraud_pattern_analysis")
}

func (s *DisputeScheduler) checkEvidenceDeadlines(ctx context.Context) {
    // Find disputes with evidence due in next 48 hours
    disputes, err := s.disputeRepo.GetDisputesWithApproachingDeadline(ctx, 48*time.Hour)
    if err != nil {
        logrus.WithError(err).Error("Failed to check evidence deadlines")
        return
    }

    for _, dispute := range disputes {
        adminUsers := s.getAdminUsers(ctx)
        if len(adminUsers) > 0 {
            daysRemaining := time.Until(*dispute.EvidenceDueBy) / (24 * time.Hour)
            s.notificationSvc.SendDisputeEvidenceDueSoon(ctx, adminUsers, dispute.StripeDisputeID, int(daysRemaining))
        }
    }
}

func (s *DisputeScheduler) checkPendingReviews(ctx context.Context) {
    // Find disputes pending review > 24 hours
    disputes, err := s.disputeRepo.GetStalePendingDisputes(ctx, 24*time.Hour)
    if err != nil {
        logrus.WithError(err).Error("Failed to check pending reviews")
        return
    }

    for _, dispute := range disputes {
        adminUsers := s.getAdminUsers(ctx)
        if len(adminUsers) > 0 {
            s.notificationSvc.SendDisputePendingReminder(ctx, adminUsers, dispute.StripeDisputeID)
        }
    }
}

func (s *DisputeScheduler) analyzeFraudPatterns(ctx context.Context) {
    // Get disputes from last 180 days
    window := 180 * 24 * time.Hour
    disputes, err := s.disputeRepo.GetRecentDisputes(ctx, window)
    if err != nil {
        logrus.WithError(err).Error("Failed to analyze fraud patterns")
        return
    }

    // Group by tenant/user, find repeat offenders
    offenderMap := make(map[string][]*storage.PaymentDispute)
    for _, d := range disputes {
        key := d.TenantID.String() // Simplified
        offenderMap[key] = append(offenderMap[key], d)
    }

    for tenantID, disputeList := range offenderMap {
        if len(disputeList) >= 2 {
            // Flag as repeat offender
            logrus.WithField("tenant_id", tenantID).Warn("Repeat chargeback offender detected")
            // Could update a fraud risk score on the tenant record
        }
    }
}
```

Add new repository methods:
- `GetDisputesWithApproachingDeadline(ctx, within time.Duration)`
- `GetStalePendingDisputes(ctx, olderThan time.Duration)`
- `GetRecentDisputes(ctx, since time.Duration)`

### Task 2.2: Register Scheduler

In `cmd/orchestrator-api` or main server setup:
```go
disputeScheduler := scheduler.NewDisputeScheduler(disputeRepo, disputeResponseManager, notificationSvc)
disputeScheduler.Start()
```

---

## Phase 3: Admin Dashboard UI

### Task 3.1: Frontend Components

**New files in `web/dashboard/src/`**:

- `pages/admin/disputes/` - Disputes list and detail pages
- `components/admin/disputes/` - Dispute components

Key components:
- `DisputeList.tsx` - Filterable list with status badges
- `DisputeDetail.tsx` - Full dispute info, timeline, actions
- `EvidencePreview.tsx` - Preview of compiled evidence
- `DisputeStats.tsx` - Statistics dashboard
- `DisputeTimeline.tsx` - Automation log timeline

---

## Implementation Order

1. **Migration** - Run database changes first
2. **Task 1.1** - Wire DisputeResponseManager into webhooks
3. **Task 1.2** - Enhance evidence compilation
4. **Task 1.3** - Implement customer notifications
5. **Task 1.4** - Create admin API handlers
6. **Task 1.5** - (Already done in migration step)
7. **Task 2.1** - Create dispute scheduler
8. **Task 2.2** - Register scheduler in main
9. **Task 3.1** - Admin dashboard UI

---

## Testing Requirements

1. **Unit tests**:
   - Decision matrix logic in DisputeResponseManager
   - Evidence compilation with mock data
   - Scheduler job logic

2. **Integration tests**:
   - Webhook → DisputeResponseManager flow
   - Admin API endpoints
   - Scheduler jobs with time mocking

3. **Manual testing**:
   - Create test dispute via Stripe dashboard
   - Verify webhook triggers automation
   - Check evidence compilation
   - Test admin API endpoints

---

## Environment Variables for Testing

```bash
# Enable full automation in development
CHARGEBACK_AUTO_REFUND_ENABLED=true
CHARGEBACK_AUTO_REFUND_THRESHOLD_CENTS=2500
CHARGEBACK_EVIDENCE_AUTO_SUBMIT=false  # Keep false for manual approval
CHARGEBACK_MANUAL_REVIEW_THRESHOLD_CENTS=15000
CHARGEBACK_ADMIN_ESCALATION_ENABLED=true
CHARGEBACK_CUSTOMER_NOTIFICATION_ENABLED=true
```
