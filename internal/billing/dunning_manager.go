package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/invoice"
)

// DunningManager handles automated retry logic for failed payments
type DunningManager struct {
	dunningRepo     *storage.DunningRepository
	userRepo        storage.Repository
	notificationSvc *notification.Service
	stripeKey       string
	stop            chan struct{}
}

// NewDunningManager creates a new dunning manager
func NewDunningManager(
	dunningRepo *storage.DunningRepository,
	userRepo storage.Repository,
	notificationSvc *notification.Service,
) *DunningManager {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	return &DunningManager{
		dunningRepo:     dunningRepo,
		userRepo:        userRepo,
		notificationSvc: notificationSvc,
		stripeKey:       stripeKey,
		stop:            make(chan struct{}),
	}
}

// IsStripeConfigured returns whether Stripe is configured
func (m *DunningManager) IsStripeConfigured() bool {
	return m.stripeKey != ""
}

// InitiateDunningWorkflow starts the dunning process for a failed payment
func (m *DunningManager) InitiateDunningWorkflow(ctx context.Context, params DunningInitiationParams) (*storage.PaymentRetry, error) {
	// Get retry schedule (default for now, could be based on tenant tier)
	schedule, err := m.dunningRepo.GetRetryScheduleByType(ctx, "default")
	if err != nil {
		logrus.WithError(err).Error("Failed to get retry schedule, using defaults")
		// Use hardcoded defaults if schedule not found
		schedule = &storage.PaymentRetrySchedule{
			MaxRetries:                    4,
			RetryIntervals:                []int{1, 3, 7, 14},
			GracePeriodDays:               14,
			SendCustomerNotifications:     true,
			NotifyAdminOnFinalRetry:       true,
			SuspendServiceAfterFinalRetry: true,
		}
	}

	// Check if a retry already exists for this invoice
	existingRetry, err := m.dunningRepo.GetPaymentRetryByInvoiceID(ctx, params.InvoiceID)
	if err == nil && existingRetry != nil {
		logrus.WithField("invoice_id", params.InvoiceID).Info("Payment retry already exists for invoice")
		return existingRetry, nil
	}

	// Calculate grace period end
	// Use invoice due date if available, otherwise use current time
	// This prevents grace period from overlapping with next billing cycle
	var gracePeriodStart time.Time
	if params.InvoiceDueDate != nil {
		gracePeriodStart = params.InvoiceDueDate.UTC()
	} else {
		gracePeriodStart = time.Now().UTC()
	}
	gracePeriodEndsAt := gracePeriodStart.AddDate(0, 0, schedule.GracePeriodDays)

	// Calculate first retry date
	nextRetryAt := storage.CalculateNextRetryDate(time.Now().UTC(), 0, schedule.RetryIntervals)

	retry := &storage.PaymentRetry{
		TenantID:           params.TenantID,
		SubscriptionID:     params.SubscriptionID,
		InvoiceID:          params.InvoiceID,
		StripeCustomerID:   params.StripeCustomerID,
		ScheduleID:         &schedule.ID,
		CurrentAttempt:     0,
		MaxAttempts:        schedule.MaxRetries,
		Status:             "active",
		AmountDueCents:     params.AmountDueCents,
		Currency:           params.Currency,
		InitialFailureAt:   time.Now().UTC(),
		NextRetryAt:        &nextRetryAt,
		GracePeriodEndsAt:  gracePeriodEndsAt,
		LastFailureCode:    params.FailureCode,
		LastFailureMessage: params.FailureMessage,
		DeclineCode:        params.DeclineCode,
		RetryHistory:       json.RawMessage("[]"),
		Metadata:           params.Metadata,
	}

	if err := m.dunningRepo.CreatePaymentRetry(ctx, retry); err != nil {
		return nil, fmt.Errorf("failed to create payment retry: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"retry_id":              retry.ID,
		"tenant_id":             params.TenantID,
		"invoice_id":            params.InvoiceID,
		"grace_period_ends":     gracePeriodEndsAt,
		"first_retry_scheduled": nextRetryAt,
	}).Info("Initiated dunning workflow")

	// Send initial failure notification
	if schedule.SendCustomerNotifications {
		if err := m.sendInitialFailureNotification(ctx, retry, params.CustomerEmail); err != nil {
			logrus.WithError(err).Warn("Failed to send initial failure notification")
		}
	}

	return retry, nil
}

// ProcessScheduledRetries processes all scheduled retries
func (m *DunningManager) ProcessScheduledRetries(ctx context.Context) error {
	retries, err := m.dunningRepo.GetActivePaymentRetries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get scheduled retries: %w", err)
	}

	logrus.WithField("count", len(retries)).Info("Processing scheduled payment retries")

	for _, retry := range retries {
		if err := m.processRetry(ctx, &retry); err != nil {
			logrus.WithError(err).WithField("retry_id", retry.ID).Error("Failed to process retry")
			// Continue processing other retries
		}
	}

	return nil
}

// ProcessGracePeriodExpirations processes retries where grace period has ended
func (m *DunningManager) ProcessGracePeriodExpirations(ctx context.Context) error {
	retries, err := m.dunningRepo.GetOverdueGracePeriodRetries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get overdue grace period retries: %w", err)
	}

	for _, retry := range retries {
		logrus.WithFields(logrus.Fields{
			"retry_id":  retry.ID,
			"tenant_id": retry.TenantID,
		}).Info("Grace period ended, suspending service")

		if err := m.SuspendService(ctx, retry.ID, retry.TenantID, "grace_period_expired"); err != nil {
			logrus.WithError(err).WithField("retry_id", retry.ID).Error("Failed to suspend service")
		}
	}

	return nil
}

// processRetry handles a single retry attempt
func (m *DunningManager) processRetry(ctx context.Context, retry *storage.PaymentRetry) error {
	if !m.IsStripeConfigured() {
		return fmt.Errorf("Stripe not configured")
	}

	// Get schedule for this retry
	schedule, err := m.dunningRepo.GetRetryScheduleByType(ctx, "default")
	if err != nil {
		return err
	}

	// Attempt the retry via Stripe
	stripeInvoice, err := m.attemptStripeRetry(retry)

	attempt := storage.RetryAttempt{
		AttemptNumber: retry.CurrentAttempt + 1,
		AttemptedAt:   time.Now().UTC(),
		Status:        "failed",
	}

	if err != nil {
		attempt.ErrorMessage = err.Error()
		logrus.WithError(err).WithField("retry_id", retry.ID).Error("Retry attempt failed")
	} else if stripeInvoice != nil && stripeInvoice.Status == stripe.InvoiceStatusPaid {
		attempt.Status = "success"
		attempt.StripeInvoiceID = stripeInvoice.ID
	}

	// Record the attempt
	if err := m.dunningRepo.RecordRetryAttempt(ctx, retry.ID, attempt); err != nil {
		logrus.WithError(err).Warn("Failed to record retry attempt")
	}

	// Update retry state
	retry.CurrentAttempt++
	retry.LastRetryAt = &attempt.AttemptedAt

	if attempt.Status == "success" {
		// Payment succeeded - resolve the retry
		now := time.Now().UTC()
		retry.ResolvedAt = &now
		retry.ResolutionType = "payment_success"
		retry.Status = "resolved"
		retry.NextRetryAt = nil

		// Send success notification
		if schedule.SendCustomerNotifications {
			m.sendPaymentRecoveredNotification(ctx, retry)
		}
	} else {
		// Payment still failed
		if retry.CurrentAttempt >= retry.MaxAttempts {
			// Max retries exceeded
			retry.Status = "failed"
			retry.NextRetryAt = nil

			// Final escalation notifications
			if schedule.NotifyAdminOnFinalRetry {
				m.sendAdminFinalFailureNotification(ctx, retry)
			}
			if schedule.SendCustomerNotifications {
				m.sendFinalFailureNotification(ctx, retry)
			}

			// Suspend service if configured
			if schedule.SuspendServiceAfterFinalRetry {
				if err := m.SuspendService(ctx, retry.ID, retry.TenantID, "max_retries_exceeded"); err != nil {
					logrus.WithError(err).Error("Failed to suspend service after max retries")
				}
			}
		} else {
			// Schedule next retry
			nextRetry := storage.CalculateNextRetryDate(retry.InitialFailureAt, retry.CurrentAttempt, schedule.RetryIntervals)
			retry.NextRetryAt = &nextRetry

			// Send retry reminder notification
			if schedule.SendCustomerNotifications {
				m.sendRetryReminderNotification(ctx, retry, retry.CurrentAttempt)
			}
		}
	}

	return m.dunningRepo.UpdatePaymentRetry(ctx, retry)
}

// attemptStripeRetry attempts to retry a payment via Stripe
func (m *DunningManager) attemptStripeRetry(retry *storage.PaymentRetry) (*stripe.Invoice, error) {
	if retry.InvoiceID == "" {
		return nil, fmt.Errorf("no invoice ID available for retry")
	}

	// Get the invoice from Stripe
	params := &stripe.InvoiceParams{}
	stripeInvoice, err := invoice.Get(retry.InvoiceID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice from Stripe: %w", err)
	}

	// Check if already paid
	if stripeInvoice.Status == stripe.InvoiceStatusPaid {
		return stripeInvoice, nil
	}

	// Finalize the invoice if needed
	if stripeInvoice.Status == stripe.InvoiceStatusDraft {
		stripeInvoice, err = invoice.FinalizeInvoice(retry.InvoiceID, &stripe.InvoiceFinalizeInvoiceParams{})
		if err != nil {
			return nil, fmt.Errorf("failed to finalize invoice: %w", err)
		}
	}

	// Pay the invoice (this triggers the payment attempt)
	payParams := &stripe.InvoicePayParams{
		PaymentMethod: stripe.String(""), // Use customer's default payment method
	}
	stripeInvoice, err = invoice.Pay(retry.InvoiceID, payParams)
	if err != nil {
		return nil, fmt.Errorf("payment attempt failed: %w", err)
	}

	return stripeInvoice, nil
}

// SuspendService suspends service for a tenant due to failed payment
// This updates both the service_suspensions table and the tenant status
func (m *DunningManager) SuspendService(ctx context.Context, retryID, tenantID uuid.UUID, reason string) error {
	// Check if already suspended
	isSuspended, err := m.dunningRepo.IsTenantSuspended(ctx, tenantID)
	if err != nil {
		return err
	}
	if isSuspended {
		return nil // Already suspended
	}

	suspension := &storage.ServiceSuspension{
		TenantID:          tenantID,
		PaymentRetryID:    retryID,
		SuspendedBy:       "system",
		Reason:            reason,
		SuspendedFeatures: json.RawMessage(`["api_access","function_publishing","new_deployments"]`),
	}

	if err := m.dunningRepo.CreateServiceSuspension(ctx, suspension); err != nil {
		return fmt.Errorf("failed to create service suspension: %w", err)
	}

	// CRITICAL: Update tenant status to "suspended" so auth middleware blocks access
	if err := m.userRepo.UpdateTenantStatus(ctx, tenantID, "suspended"); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to update tenant status to suspended")
		// Don't fail the suspension, but log the error
	}

	// Send suspension notification
	m.sendServiceSuspendedNotification(ctx, suspension)

	logrus.WithFields(logrus.Fields{
		"tenant_id":     tenantID,
		"retry_id":      retryID,
		"suspension_id": suspension.ID,
		"reason":        reason,
	}).Info("Service suspended due to failed payment recovery")

	return nil
}

// RestoreService restores service for a tenant after payment is resolved
// This updates both the service_suspensions table and the tenant status
func (m *DunningManager) RestoreService(ctx context.Context, tenantID uuid.UUID, restoredBy, reason string) error {
	suspension, err := m.dunningRepo.GetActiveSuspensionByTenant(ctx, tenantID)
	if err != nil {
		return err // No active suspension
	}

	if err := m.dunningRepo.RestoreService(ctx, suspension.ID, restoredBy, reason); err != nil {
		return fmt.Errorf("failed to restore service: %w", err)
	}

	// CRITICAL: Restore tenant status to "active"
	if err := m.userRepo.UpdateTenantStatus(ctx, tenantID, "active"); err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Error("Failed to restore tenant status to active")
		// Don't fail the restoration, but log the error
	}

	// Send restoration notification
	m.sendServiceRestoredNotification(ctx, tenantID, suspension)

	logrus.WithFields(logrus.Fields{
		"tenant_id":   tenantID,
		"restored_by": restoredBy,
		"reason":      reason,
	}).Info("Service restored after payment resolution")

	return nil
}

// IsTenantSuspended checks if a tenant's service is currently suspended
func (m *DunningManager) IsTenantSuspended(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return m.dunningRepo.IsTenantSuspended(ctx, tenantID)
}

// DunningInitiationParams parameters for initiating dunning workflow
type DunningInitiationParams struct {
	TenantID         uuid.UUID
	SubscriptionID   *uuid.UUID
	InvoiceID        string
	StripeCustomerID string
	CustomerEmail    string
	AmountDueCents   int
	Currency         string
	FailureCode      string
	FailureMessage   string
	DeclineCode      string
	Metadata         json.RawMessage
	InvoiceDueDate   *time.Time // Optional: invoice due date for grace period calculation
}

// Notification helpers

func (m *DunningManager) sendInitialFailureNotification(ctx context.Context, retry *storage.PaymentRetry, customerEmail string) error {
	if m.notificationSvc == nil || customerEmail == "" {
		return nil
	}

	amountDue := float64(retry.AmountDueCents) / 100.0

	return m.notificationSvc.SendBillingAlert(ctx, customerEmail, "payment_failed", map[string]interface{}{
		"amount_due":       amountDue,
		"currency":         retry.Currency,
		"attempt_count":    1,
		"next_retry_at":    retry.NextRetryAt,
		"grace_period_end": retry.GracePeriodEndsAt,
		"retry_id":         retry.ID,
	})
}

func (m *DunningManager) sendRetryReminderNotification(ctx context.Context, retry *storage.PaymentRetry, attemptNumber int) error {
	if m.notificationSvc == nil {
		return nil
	}

	// Get tenant users to notify
	users, err := m.userRepo.ListActiveUsersByTenant(ctx, retry.TenantID)
	if err != nil || len(users) == 0 {
		return err
	}

	amountDue := float64(retry.AmountDueCents) / 100.0

	for _, user := range users {
		if user.Email == "" {
			continue
		}

		// Create custom notification for retry reminder
		data := map[string]interface{}{
			"amount_due":       amountDue,
			"currency":         retry.Currency,
			"attempt_number":   attemptNumber,
			"max_attempts":     retry.MaxAttempts,
			"next_retry_at":    retry.NextRetryAt,
			"grace_period_end": retry.GracePeriodEndsAt,
			"retry_id":         retry.ID,
		}

		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    fmt.Sprintf("Payment Retry Attempt %d of %d", attemptNumber, retry.MaxAttempts),
			Body:     fmt.Sprintf("Your payment of %.2f %s is scheduled for retry. Please update your payment method to avoid service interruption.", amountDue, retry.Currency),
			Data:     data,
			Channels: []string{notification.ChannelEmail},
			Priority: notification.PriorityHigh,
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", user.ID).Warn("Failed to send retry reminder")
		}
	}

	return nil
}

func (m *DunningManager) sendFinalFailureNotification(ctx context.Context, retry *storage.PaymentRetry) error {
	if m.notificationSvc == nil {
		return nil
	}

	users, err := m.userRepo.ListActiveUsersByTenant(ctx, retry.TenantID)
	if err != nil || len(users) == 0 {
		return err
	}

	amountDue := float64(retry.AmountDueCents) / 100.0

	for _, user := range users {
		if user.Email == "" {
			continue
		}

		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    "Final Payment Notice - Service Will Be Suspended",
			Body:     fmt.Sprintf("All payment retry attempts failed for your invoice of %.2f %s. Your service will be suspended if payment is not received immediately.", amountDue, retry.Currency),
			Data: map[string]interface{}{
				"amount_due": amountDue,
				"currency":   retry.Currency,
				"retry_id":   retry.ID,
				"is_final":   true,
			},
			Channels: []string{notification.ChannelEmail, notification.ChannelInApp},
			Priority: notification.PriorityUrgent,
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", user.ID).Warn("Failed to send final failure notification")
		}
	}

	return nil
}

func (m *DunningManager) sendAdminFinalFailureNotification(ctx context.Context, retry *storage.PaymentRetry) error {
	if m.notificationSvc == nil {
		return nil
	}

	amountDue := float64(retry.AmountDueCents) / 100.0

	// Get tenant information for the notification
	tenant, err := m.userRepo.GetTenantByID(retry.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", retry.TenantID).Warn("Failed to get tenant for admin notification")
		// Continue anyway - we can still notify with the tenant ID
	}

	tenantName := "Unknown"
	if tenant != nil {
		tenantName = tenant.Name
	}

	// Get admin emails from system configuration
	// First try environment variable, then fallback to getting users with admin role
	adminEmails := m.getAdminEmails(ctx, retry.TenantID)
	if len(adminEmails) == 0 {
		logrus.WithField("tenant_id", retry.TenantID).Warn("No admin emails configured for final failure notification")
		return nil
	}

	// Send urgent alert to all configured admin emails
	for _, email := range adminEmails {
		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   retry.TenantID, // Use tenant ID as the recipient identifier
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    "URGENT: Payment Recovery Failed - Immediate Action Required",
			Body: fmt.Sprintf("Final payment retry attempt failed for tenant %s (ID: %s). Amount due: %.2f %s. "+
				"Service suspension will occur if not resolved immediately.",
				tenantName, retry.TenantID.String(), amountDue, retry.Currency),
			Data: map[string]interface{}{
				"tenant_id":       retry.TenantID.String(),
				"tenant_name":     tenantName,
				"retry_id":        retry.ID.String(),
				"amount_due":      amountDue,
				"currency":        retry.Currency,
				"max_attempts":    retry.MaxAttempts,
				"failure_code":    retry.LastFailureCode,
				"is_final":        true,
				"alert_level":     "critical",
				"requires_action": true,
			},
			Channels: []string{notification.ChannelEmail, notification.ChannelInApp},
			Priority: notification.PriorityUrgent,
		})
		if err != nil {
			logrus.WithError(err).WithField("email", email).Warn("Failed to send admin final failure notification")
		} else {
			logrus.WithField("email", email).Info("Admin final failure notification sent")
		}
	}

	return nil
}

// getAdminEmails retrieves admin email addresses for notifications
// Checks environment config first, then falls back to active tenant users
func (m *DunningManager) getAdminEmails(ctx context.Context, tenantID uuid.UUID) []string {
	var emails []string

	// First priority: Environment variable for billing alerts
	if billingAlertsEmail := os.Getenv("BILLING_ALERTS_EMAIL"); billingAlertsEmail != "" {
		emails = append(emails, billingAlertsEmail)
	}

	// Second priority: Finance team email
	if financeEmail := os.Getenv("FINANCE_TEAM_EMAIL"); financeEmail != "" {
		emails = append(emails, financeEmail)
	}

	// Third priority: Get users with owner/admin role from the tenant
	users, err := m.userRepo.ListActiveUsersByTenant(ctx, tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to list users for admin notification")
	} else {
		for _, user := range users {
			if user.Email != "" && (user.Role == "owner" || user.Role == "admin") {
				emails = append(emails, user.Email)
			}
		}
	}

	// Deduplicate emails
	seen := make(map[string]bool)
	unique := []string{}
	for _, email := range emails {
		if !seen[email] {
			seen[email] = true
			unique = append(unique, email)
		}
	}

	return unique
}

func (m *DunningManager) sendPaymentRecoveredNotification(ctx context.Context, retry *storage.PaymentRetry) error {
	if m.notificationSvc == nil {
		return nil
	}

	users, err := m.userRepo.ListActiveUsersByTenant(ctx, retry.TenantID)
	if err != nil || len(users) == 0 {
		return err
	}

	for _, user := range users {
		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    "Payment Successful - Thank You!",
			Body:     "Your payment has been successfully processed. Thank you for your continued business.",
			Data: map[string]interface{}{
				"retry_id": retry.ID,
				"resolved": true,
			},
			Channels: []string{notification.ChannelInApp},
			Priority: notification.PriorityNormal,
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", user.ID).Warn("Failed to send payment recovered notification")
		}
	}

	// Also restore service if it was suspended
	m.RestoreService(ctx, retry.TenantID, "system", "payment_successful")

	return nil
}

func (m *DunningManager) sendServiceSuspendedNotification(ctx context.Context, suspension *storage.ServiceSuspension) error {
	if m.notificationSvc == nil {
		return nil
	}

	users, err := m.userRepo.ListActiveUsersByTenant(ctx, suspension.TenantID)
	if err != nil || len(users) == 0 {
		return err
	}

	for _, user := range users {
		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    "Service Suspended - Payment Required",
			Body:     "Your service has been suspended due to non-payment. Please update your payment method to restore service.",
			Data: map[string]interface{}{
				"suspension_id": suspension.ID,
				"suspended_at":  suspension.SuspendedAt,
			},
			Channels: []string{notification.ChannelEmail, notification.ChannelInApp},
			Priority: notification.PriorityUrgent,
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", user.ID).Warn("Failed to send suspension notification")
		}
	}

	return nil
}

func (m *DunningManager) sendServiceRestoredNotification(ctx context.Context, tenantID uuid.UUID, suspension *storage.ServiceSuspension) error {
	if m.notificationSvc == nil {
		return nil
	}

	users, err := m.userRepo.ListActiveUsersByTenant(ctx, tenantID)
	if err != nil || len(users) == 0 {
		return err
	}

	for _, user := range users {
		_, err := m.notificationSvc.Send(ctx, notification.SendRequest{
			UserID:   user.ID,
			Type:     notification.TypeBillingAlert,
			Category: notification.CategoryBilling,
			Title:    "Service Restored",
			Body:     "Your service has been restored. Thank you for updating your payment information.",
			Data: map[string]interface{}{
				"restored": true,
			},
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
			Priority: notification.PriorityNormal,
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", user.ID).Warn("Failed to send restoration notification")
		}
	}

	return nil
}

// Stop signals the dunning processors to stop for graceful shutdown
func (m *DunningManager) Stop() {
	close(m.stop)
}

// StopChan returns the stop channel for the dunning processors
func (m *DunningManager) StopChan() <-chan struct{} {
	return m.stop
}
