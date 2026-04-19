package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentRetrySchedule defines retry schedules with configurable intervals and grace periods
type PaymentRetrySchedule struct {
	ID                            uuid.UUID `json:"id" db:"id"`
	Name                          string    `json:"name" db:"name"`
	Description                   string    `json:"description" db:"description"`
	MaxRetries                    int       `json:"max_retries" db:"max_retries"`
	RetryIntervals                []int     `json:"retry_intervals" db:"retry_intervals"` // Days between retries
	GracePeriodDays               int       `json:"grace_period_days" db:"grace_period_days"`
	SendCustomerNotifications     bool      `json:"send_customer_notifications" db:"send_customer_notifications"`
	NotifyAdminOnFinalRetry       bool      `json:"notify_admin_on_final_retry" db:"notify_admin_on_final_retry"`
	SuspendServiceAfterFinalRetry bool      `json:"suspend_service_after_final_retry" db:"suspend_service_after_final_retry"`
	ScheduleType                  string    `json:"schedule_type" db:"schedule_type"`
	IsActive                      bool      `json:"is_active" db:"is_active"`
	CreatedAt                     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                     time.Time `json:"updated_at" db:"updated_at"`
}

// PaymentRetry tracks individual payment retry workflows for failed invoices
type PaymentRetry struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	TenantID           uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	SubscriptionID     *uuid.UUID      `json:"subscription_id,omitempty" db:"subscription_id"`
	InvoiceID          string          `json:"invoice_id" db:"invoice_id"`
	StripeCustomerID   string          `json:"stripe_customer_id" db:"stripe_customer_id"`
	ScheduleID         *uuid.UUID      `json:"schedule_id,omitempty" db:"schedule_id"`
	CurrentAttempt     int             `json:"current_attempt" db:"current_attempt"`
	MaxAttempts        int             `json:"max_attempts" db:"max_attempts"`
	Status             string          `json:"status" db:"status"` // 'active', 'paused', 'resolved', 'failed', 'cancelled'
	AmountDueCents     int             `json:"amount_due_cents" db:"amount_due_cents"`
	Currency           string          `json:"currency" db:"currency"`
	InitialFailureAt   time.Time       `json:"initial_failure_at" db:"initial_failure_at"`
	LastRetryAt        *time.Time      `json:"last_retry_at,omitempty" db:"last_retry_at"`
	NextRetryAt        *time.Time      `json:"next_retry_at,omitempty" db:"next_retry_at"`
	GracePeriodEndsAt  time.Time       `json:"grace_period_ends_at" db:"grace_period_ends_at"`
	ResolvedAt         *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionType     string          `json:"resolution_type,omitempty" db:"resolution_type"`
	RetryHistory       json.RawMessage `json:"retry_history" db:"retry_history"`
	LastFailureCode    string          `json:"last_failure_code,omitempty" db:"last_failure_code"`
	LastFailureMessage string          `json:"last_failure_message,omitempty" db:"last_failure_message"`
	DeclineCode        string          `json:"decline_code,omitempty" db:"decline_code"`
	Metadata           json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// RetryAttempt represents a single retry attempt in the history
type RetryAttempt struct {
	AttemptNumber   int       `json:"attempt_number"`
	AttemptedAt     time.Time `json:"attempted_at"`
	Status          string    `json:"status"` // 'success', 'failed', 'skipped'
	ErrorCode       string    `json:"error_code,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	StripeInvoiceID string    `json:"stripe_invoice_id,omitempty"`
}

// DunningNotification tracks notifications sent during the dunning process
type DunningNotification struct {
	ID                     uuid.UUID       `json:"id" db:"id"`
	PaymentRetryID         uuid.UUID       `json:"payment_retry_id" db:"payment_retry_id"`
	NotificationType       string          `json:"notification_type" db:"notification_type"`
	AttemptNumber          *int            `json:"attempt_number,omitempty" db:"attempt_number"`
	RecipientEmail         string          `json:"recipient_email" db:"recipient_email"`
	RecipientUserID        *uuid.UUID      `json:"recipient_user_id,omitempty" db:"recipient_user_id"`
	Subject                string          `json:"subject" db:"subject"`
	Body                   string          `json:"body,omitempty" db:"body"`
	SentAt                 time.Time       `json:"sent_at" db:"sent_at"`
	DeliveredAt            *time.Time      `json:"delivered_at,omitempty" db:"delivered_at"`
	OpenedAt               *time.Time      `json:"opened_at,omitempty" db:"opened_at"`
	ClickedAt              *time.Time      `json:"clicked_at,omitempty" db:"clicked_at"`
	EmailProviderMessageID string          `json:"email_provider_message_id,omitempty" db:"email_provider_message_id"`
	Metadata               json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
}

// ServiceSuspension tracks service suspensions due to failed payments
type ServiceSuspension struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	TenantID          uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	PaymentRetryID    uuid.UUID       `json:"payment_retry_id" db:"payment_retry_id"`
	SuspendedAt       time.Time       `json:"suspended_at" db:"suspended_at"`
	SuspendedBy       string          `json:"suspended_by" db:"suspended_by"`
	Reason            string          `json:"reason" db:"reason"`
	RestoredAt        *time.Time      `json:"restored_at,omitempty" db:"restored_at"`
	RestoredBy        string          `json:"restored_by,omitempty" db:"restored_by"`
	RestorationReason string          `json:"restoration_reason,omitempty" db:"restoration_reason"`
	SuspendedFeatures json.RawMessage `json:"suspended_features" db:"suspended_features"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// DunningRepository handles dunning management operations
type DunningRepository struct {
	db *PostgresDB
}

// NewDunningRepository creates a new dunning repository
func NewDunningRepository(db *PostgresDB) *DunningRepository {
	return &DunningRepository{db: db}
}

// GetRetryScheduleByType retrieves a retry schedule by its type
func (r *DunningRepository) GetRetryScheduleByType(ctx context.Context, scheduleType string) (*PaymentRetrySchedule, error) {
	var schedule PaymentRetrySchedule
	query := `
		SELECT id, name, description, max_retries, retry_intervals, grace_period_days,
		       send_customer_notifications, notify_admin_on_final_retry, suspend_service_after_final_retry,
		       schedule_type, is_active, created_at, updated_at
		FROM payment_retry_schedules
		WHERE schedule_type = $1 AND is_active = true
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, scheduleType)
	err := row.Scan(
		&schedule.ID, &schedule.Name, &schedule.Description, &schedule.MaxRetries, &schedule.RetryIntervals,
		&schedule.GracePeriodDays, &schedule.SendCustomerNotifications, &schedule.NotifyAdminOnFinalRetry,
		&schedule.SuspendServiceAfterFinalRetry, &schedule.ScheduleType, &schedule.IsActive,
		&schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default schedule if specific type not found
			return r.GetRetryScheduleByType(ctx, "default")
		}
		return nil, err
	}
	return &schedule, nil
}

// CreatePaymentRetry creates a new payment retry record
func (r *DunningRepository) CreatePaymentRetry(ctx context.Context, retry *PaymentRetry) error {
	query := `
		INSERT INTO payment_retries (
			id, tenant_id, subscription_id, invoice_id, stripe_customer_id, schedule_id,
			current_attempt, max_attempts, status, amount_due_cents, currency,
			initial_failure_at, last_retry_at, next_retry_at, grace_period_ends_at,
			retry_history, last_failure_code, last_failure_message, decline_code, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`
	if retry.ID == uuid.Nil {
		retry.ID = uuid.New()
	}
	now := time.Now().UTC()
	retry.CreatedAt = now
	retry.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		retry.ID, retry.TenantID, retry.SubscriptionID, retry.InvoiceID, retry.StripeCustomerID, retry.ScheduleID,
		retry.CurrentAttempt, retry.MaxAttempts, retry.Status, retry.AmountDueCents, retry.Currency,
		retry.InitialFailureAt, retry.LastRetryAt, retry.NextRetryAt, retry.GracePeriodEndsAt,
		retry.RetryHistory, retry.LastFailureCode, retry.LastFailureMessage, retry.DeclineCode, retry.Metadata,
	)
	return err
}

// GetPaymentRetryByInvoiceID retrieves a payment retry by Stripe invoice ID
func (r *DunningRepository) GetPaymentRetryByInvoiceID(ctx context.Context, invoiceID string) (*PaymentRetry, error) {
	query := `
		SELECT id, tenant_id, subscription_id, invoice_id, stripe_customer_id, schedule_id,
		       current_attempt, max_attempts, status, amount_due_cents, currency,
		       initial_failure_at, last_retry_at, next_retry_at, grace_period_ends_at,
		       resolved_at, resolution_type, retry_history, last_failure_code, 
		       last_failure_message, decline_code, metadata, created_at, updated_at
		FROM payment_retries
		WHERE invoice_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, invoiceID)
	return scanPaymentRetry(row)
}

// GetActivePaymentRetries retrieves all active payment retries that need processing
func (r *DunningRepository) GetActivePaymentRetries(ctx context.Context) ([]PaymentRetry, error) {
	query := `
		SELECT id, tenant_id, subscription_id, invoice_id, stripe_customer_id, schedule_id,
		       current_attempt, max_attempts, status, amount_due_cents, currency,
		       initial_failure_at, last_retry_at, next_retry_at, grace_period_ends_at,
		       resolved_at, resolution_type, retry_history, last_failure_code, 
		       last_failure_message, decline_code, metadata, created_at, updated_at
		FROM payment_retries
		WHERE status = 'active'
		  AND (next_retry_at IS NULL OR next_retry_at <= $1)
		ORDER BY next_retry_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPaymentRetryRows(rows)
}

// GetOverdueGracePeriodRetries retrieves retries where grace period has ended
func (r *DunningRepository) GetOverdueGracePeriodRetries(ctx context.Context) ([]PaymentRetry, error) {
	query := `
		SELECT id, tenant_id, subscription_id, invoice_id, stripe_customer_id, schedule_id,
		       current_attempt, max_attempts, status, amount_due_cents, currency,
		       initial_failure_at, last_retry_at, next_retry_at, grace_period_ends_at,
		       resolved_at, resolution_type, retry_history, last_failure_code, 
		       last_failure_message, decline_code, metadata, created_at, updated_at
		FROM payment_retries
		WHERE status = 'active'
		  AND grace_period_ends_at <= $1
		ORDER BY grace_period_ends_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPaymentRetryRows(rows)
}

// UpdatePaymentRetry updates a payment retry record
func (r *DunningRepository) UpdatePaymentRetry(ctx context.Context, retry *PaymentRetry) error {
	query := `
		UPDATE payment_retries SET
			current_attempt = $1,
			status = $2,
			last_retry_at = $3,
			next_retry_at = $4,
			resolved_at = $5,
			resolution_type = $6,
			retry_history = $7,
			last_failure_code = $8,
			last_failure_message = $9,
			decline_code = $10,
			metadata = $11,
			updated_at = $12
		WHERE id = $13
	`
	retry.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		retry.CurrentAttempt, retry.Status, retry.LastRetryAt, retry.NextRetryAt,
		retry.ResolvedAt, retry.ResolutionType, retry.RetryHistory,
		retry.LastFailureCode, retry.LastFailureMessage, retry.DeclineCode,
		retry.Metadata, retry.UpdatedAt, retry.ID,
	)
	return err
}

// RecordRetryAttempt adds a retry attempt to the history
func (r *DunningRepository) RecordRetryAttempt(ctx context.Context, retryID uuid.UUID, attempt RetryAttempt) error {
	// Get current history
	var currentHistory []RetryAttempt
	query := `SELECT retry_history FROM payment_retries WHERE id = $1`
	var historyJSON []byte
	err := r.db.QueryRowContext(ctx, query, retryID).Scan(&historyJSON)
	if err != nil {
		return err
	}

	if len(historyJSON) > 0 {
		json.Unmarshal(historyJSON, &currentHistory)
	}

	// Append new attempt
	currentHistory = append(currentHistory, attempt)

	// Update
	historyJSON, _ = json.Marshal(currentHistory)
	updateQuery := `UPDATE payment_retries SET retry_history = $1, updated_at = $2 WHERE id = $3`
	_, err = r.db.ExecContext(ctx, updateQuery, historyJSON, time.Now().UTC(), retryID)
	return err
}

// CreateDunningNotification creates a dunning notification record
func (r *DunningRepository) CreateDunningNotification(ctx context.Context, notification *DunningNotification) error {
	query := `
		INSERT INTO dunning_notifications (
			id, payment_retry_id, notification_type, attempt_number, recipient_email,
			recipient_user_id, subject, body, sent_at, email_provider_message_id, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}
	notification.CreatedAt = time.Now().UTC()
	if notification.SentAt.IsZero() {
		notification.SentAt = notification.CreatedAt
	}

	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.PaymentRetryID, notification.NotificationType,
		notification.AttemptNumber, notification.RecipientEmail, notification.RecipientUserID,
		notification.Subject, notification.Body, notification.SentAt,
		notification.EmailProviderMessageID, notification.Metadata,
	)
	return err
}

// GetDunningNotificationsByRetry retrieves notifications for a payment retry
func (r *DunningRepository) GetDunningNotificationsByRetry(ctx context.Context, retryID uuid.UUID) ([]DunningNotification, error) {
	query := `
		SELECT id, payment_retry_id, notification_type, attempt_number, recipient_email,
		       recipient_user_id, subject, body, sent_at, delivered_at, opened_at,
		       clicked_at, email_provider_message_id, metadata, created_at
		FROM dunning_notifications
		WHERE payment_retry_id = $1
		ORDER BY sent_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, retryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDunningNotificationRows(rows)
}

// CreateServiceSuspension creates a service suspension record
func (r *DunningRepository) CreateServiceSuspension(ctx context.Context, suspension *ServiceSuspension) error {
	query := `
		INSERT INTO service_suspensions (
			id, tenant_id, payment_retry_id, suspended_at, suspended_by, reason,
			suspended_features, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if suspension.ID == uuid.Nil {
		suspension.ID = uuid.New()
	}
	now := time.Now().UTC()
	suspension.CreatedAt = now
	suspension.UpdatedAt = now
	if suspension.SuspendedAt.IsZero() {
		suspension.SuspendedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		suspension.ID, suspension.TenantID, suspension.PaymentRetryID,
		suspension.SuspendedAt, suspension.SuspendedBy, suspension.Reason,
		suspension.SuspendedFeatures, suspension.CreatedAt, suspension.UpdatedAt,
	)
	return err
}

// GetActiveSuspensionByTenant retrieves the active suspension for a tenant
func (r *DunningRepository) GetActiveSuspensionByTenant(ctx context.Context, tenantID uuid.UUID) (*ServiceSuspension, error) {
	query := `
		SELECT id, tenant_id, payment_retry_id, suspended_at, suspended_by, reason,
		       restored_at, restored_by, restoration_reason, suspended_features,
		       created_at, updated_at
		FROM service_suspensions
		WHERE tenant_id = $1 AND restored_at IS NULL
		ORDER BY suspended_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, tenantID)
	var s ServiceSuspension
	err := row.Scan(
		&s.ID, &s.TenantID, &s.PaymentRetryID, &s.SuspendedAt, &s.SuspendedBy, &s.Reason,
		&s.RestoredAt, &s.RestoredBy, &s.RestorationReason, &s.SuspendedFeatures,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// RestoreService removes a service suspension
func (r *DunningRepository) RestoreService(ctx context.Context, suspensionID uuid.UUID, restoredBy, reason string) error {
	query := `
		UPDATE service_suspensions SET
			restored_at = $1,
			restored_by = $2,
			restoration_reason = $3,
			updated_at = $1
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), restoredBy, reason, suspensionID)
	return err
}

// IsTenantSuspended checks if a tenant has an active service suspension
func (r *DunningRepository) IsTenantSuspended(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM service_suspensions
		WHERE tenant_id = $1 AND restored_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, tenantID)
	err := row.Scan(&count)
	return count > 0, err
}

// CalculateNextRetryDate calculates the next retry date based on the schedule
func CalculateNextRetryDate(initialFailure time.Time, currentAttempt int, retryIntervals []int) time.Time {
	if currentAttempt >= len(retryIntervals) {
		// Use the last interval for subsequent retries
		return initialFailure.AddDate(0, 0, retryIntervals[len(retryIntervals)-1])
	}
	return initialFailure.AddDate(0, 0, retryIntervals[currentAttempt])
}

// scanPaymentRetry scans a PaymentRetry from a database row
func scanPaymentRetry(row *sql.Row) (*PaymentRetry, error) {
	var retry PaymentRetry
	err := row.Scan(
		&retry.ID, &retry.TenantID, &retry.SubscriptionID, &retry.InvoiceID, &retry.StripeCustomerID,
		&retry.ScheduleID, &retry.CurrentAttempt, &retry.MaxAttempts, &retry.Status, &retry.AmountDueCents,
		&retry.Currency, &retry.InitialFailureAt, &retry.LastRetryAt, &retry.NextRetryAt,
		&retry.GracePeriodEndsAt, &retry.ResolvedAt, &retry.ResolutionType, &retry.RetryHistory,
		&retry.LastFailureCode, &retry.LastFailureMessage, &retry.DeclineCode, &retry.Metadata,
		&retry.CreatedAt, &retry.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &retry, nil
}

// scanPaymentRetryRows scans PaymentRetry rows from a query result
func scanPaymentRetryRows(rows *sql.Rows) ([]PaymentRetry, error) {
	var retries []PaymentRetry
	for rows.Next() {
		var retry PaymentRetry
		err := rows.Scan(
			&retry.ID, &retry.TenantID, &retry.SubscriptionID, &retry.InvoiceID, &retry.StripeCustomerID,
			&retry.ScheduleID, &retry.CurrentAttempt, &retry.MaxAttempts, &retry.Status, &retry.AmountDueCents,
			&retry.Currency, &retry.InitialFailureAt, &retry.LastRetryAt, &retry.NextRetryAt,
			&retry.GracePeriodEndsAt, &retry.ResolvedAt, &retry.ResolutionType, &retry.RetryHistory,
			&retry.LastFailureCode, &retry.LastFailureMessage, &retry.DeclineCode, &retry.Metadata,
			&retry.CreatedAt, &retry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		retries = append(retries, retry)
	}
	return retries, rows.Err()
}

// scanDunningNotificationRows scans DunningNotification rows from a query result
func scanDunningNotificationRows(rows *sql.Rows) ([]DunningNotification, error) {
	var notifications []DunningNotification
	for rows.Next() {
		var n DunningNotification
		err := rows.Scan(
			&n.ID, &n.PaymentRetryID, &n.NotificationType, &n.AttemptNumber, &n.RecipientEmail,
			&n.RecipientUserID, &n.Subject, &n.Body, &n.SentAt, &n.DeliveredAt, &n.OpenedAt,
			&n.ClickedAt, &n.EmailProviderMessageID, &n.Metadata, &n.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}
