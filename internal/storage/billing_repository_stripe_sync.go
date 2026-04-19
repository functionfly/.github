package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// CreateStripeSyncEvent creates a record of a Stripe webhook event
func (r *BillingRepository) CreateStripeSyncEvent(ctx context.Context, event *StripeSyncEvent) (*StripeSyncEvent, error) {
	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	var tenantID interface{}
	if event.TenantID != nil {
		tenantID = *event.TenantID
	}

	var errorMsg interface{}
	if event.ErrorMessage != nil {
		errorMsg = *event.ErrorMessage
	}

	var idempotencyKey interface{}
	if event.IdempotencyKey != nil {
		idempotencyKey = *event.IdempotencyKey
	}

	query := `
		INSERT INTO stripe_sync_events (id, stripe_event_id, stripe_object_id, event_type, event_data, tenant_id, status, error_message, retry_count, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, stripe_event_id, stripe_object_id, event_type, event_data, tenant_id, status, error_message, retry_count, idempotency_key, created_at, updated_at`

	var returnedEventData datatypes.JSON
	var returnedTenantID *uuid.UUID
	var returnedErrorMsg *string
	var returnedIdempotencyKey *string

	err := r.db.QueryRowContext(ctx, query,
		event.ID, event.StripeEventID, event.StripeObjectID, event.EventType,
		datatypes.JSON(event.EventData), tenantID, event.Status, errorMsg,
		event.RetryCount, idempotencyKey, event.CreatedAt, event.UpdatedAt,
	).Scan(
		&event.ID, &event.StripeEventID, &event.StripeObjectID, &event.EventType,
		&returnedEventData, &returnedTenantID, &event.Status,
		&returnedErrorMsg, &event.RetryCount, &returnedIdempotencyKey,
		&event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe sync event: %w", err)
	}

	event.EventData = json.RawMessage(returnedEventData)
	if returnedTenantID != nil {
		event.TenantID = returnedTenantID
	}
	if returnedErrorMsg != nil {
		event.ErrorMessage = returnedErrorMsg
	}
	if returnedIdempotencyKey != nil {
		event.IdempotencyKey = returnedIdempotencyKey
	}

	return event, nil
}

// GetStripeSyncEventByEventID retrieves a sync event by Stripe event ID (idempotency check)
func (r *BillingRepository) GetStripeSyncEventByEventID(ctx context.Context, stripeEventID string) (*StripeSyncEvent, error) {
	query := `
		SELECT id, stripe_event_id, stripe_object_id, event_type, event_data, tenant_id, status, error_message, processed_at, retry_count, idempotency_key, created_at, updated_at
		FROM stripe_sync_events
		WHERE stripe_event_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	event := &StripeSyncEvent{}
	var eventData datatypes.JSON
	var tenantID *uuid.UUID
	var processedAt *time.Time
	var errorMsg *string
	var idempotencyKey *string

	err := r.db.QueryRowContext(ctx, query, stripeEventID).Scan(
		&event.ID, &event.StripeEventID, &event.StripeObjectID, &event.EventType,
		&eventData, &tenantID, &event.Status, &errorMsg,
		&processedAt, &event.RetryCount, &idempotencyKey,
		&event.CreatedAt, &event.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe sync event: %w", err)
	}

	event.EventData = json.RawMessage(eventData)
	if tenantID != nil {
		event.TenantID = tenantID
	}
	if processedAt != nil {
		event.ProcessedAt = processedAt
	}
	if errorMsg != nil {
		event.ErrorMessage = errorMsg
	}
	if idempotencyKey != nil {
		event.IdempotencyKey = idempotencyKey
	}

	return event, nil
}

// UpdateStripeSyncEventStatus updates the status and error message of a sync event
func (r *BillingRepository) UpdateStripeSyncEventStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	query := `
		UPDATE stripe_sync_events
		SET status = $1, error_message = $2, processed_at = $3, updated_at = NOW()
		WHERE id = $4`

	var errMsg interface{}
	if errorMsg != nil {
		errMsg = *errorMsg
	}

	var processedAt time.Time
	if status == StripeSyncStatusProcessed || status == StripeSyncStatusFailed {
		processedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query, status, errMsg, processedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update stripe sync event status: %w", err)
	}
	return nil
}

// IncrementStripeSyncEventRetryCount increments the retry count
func (r *BillingRepository) IncrementStripeSyncEventRetryCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE stripe_sync_events
		SET retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}
	return nil
}

// ListPendingStripeSyncEvents lists sync events that need processing
func (r *BillingRepository) ListPendingStripeSyncEvents(ctx context.Context, limit int) ([]*StripeSyncEvent, error) {
	query := `
		SELECT id, stripe_event_id, stripe_object_id, event_type, event_data, tenant_id, status, error_message, retry_count, idempotency_key, created_at, updated_at
		FROM stripe_sync_events
		WHERE status = 'pending' OR (status = 'failed' AND retry_count < 5)
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending stripe sync events: %w", err)
	}
	defer rows.Close()

	var events []*StripeSyncEvent
	for rows.Next() {
		event := &StripeSyncEvent{}
		var eventData datatypes.JSON
		var tenantID *uuid.UUID
		var errorMsg *string
		var idempotencyKey *string

		err := rows.Scan(
			&event.ID, &event.StripeEventID, &event.StripeObjectID, &event.EventType,
			&eventData, &tenantID, &event.Status, &errorMsg,
			&event.RetryCount, &idempotencyKey,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stripe sync event: %w", err)
		}

		event.EventData = json.RawMessage(eventData)
		if tenantID != nil {
			event.TenantID = tenantID
		}
		if errorMsg != nil {
			event.ErrorMessage = errorMsg
		}
		if idempotencyKey != nil {
			event.IdempotencyKey = idempotencyKey
		}
		events = append(events, event)
	}

	return events, nil
}

// UpdateSubscriptionFromStripe updates a subscription with data from Stripe
// This is the core two-way sync method that handles changes from Stripe dashboard
func (r *BillingRepository) UpdateSubscriptionFromStripe(ctx context.Context, stripeSubscriptionID string, stripeData map[string]interface{}) (*Subscription, error) {
	// Get current subscription by Stripe ID
	sub, err := r.GetSubscriptionByStripeID(ctx, stripeSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription by stripe id: %w", err)
	}
	if sub == nil {
		return nil, fmt.Errorf("subscription not found for stripe id: %s", stripeSubscriptionID)
	}

	// Build updates from Stripe data
	updates := map[string]interface{}{
		"stripe_subscription_id": stripeSubscriptionID,
	}

	// Map status if provided
	if status, ok := stripeData["status"].(string); ok {
		updates["status"] = mapStripeStatusToInternal(status)
	}

	// Handle period dates
	if currentPeriodStart, ok := stripeData["current_period_start"].(int64); ok {
		updates["current_period_start"] = time.Unix(currentPeriodStart, 0)
	}
	if currentPeriodEnd, ok := stripeData["current_period_end"].(int64); ok {
		updates["current_period_end"] = time.Unix(currentPeriodEnd, 0)
	}

	// Handle trial
	if trialEnd, ok := stripeData["trial_end"].(int64); ok && trialEnd > 0 {
		t := time.Unix(trialEnd, 0)
		updates["trial_end"] = &t
	} else if trialEnd == 0 {
		updates["trial_end"] = nil
	}

	// Handle cancellation
	if canceledAt, ok := stripeData["canceled_at"].(int64); ok && canceledAt > 0 {
		t := time.Unix(canceledAt, 0)
		updates["canceled_at"] = &t
	}
	if cancelAtPeriodEnd, ok := stripeData["cancel_at_period_end"].(bool); ok {
		updates["cancel_at_period_end"] = cancelAtPeriodEnd
	}

	// Handle quantity changes (for usage-based subscriptions)
	if quantity, ok := stripeData["quantity"].(int); ok {
		updates["quantity"] = quantity
	}

	// Apply the updates
	updated, err := r.UpdateSubscription(ctx, sub.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription from stripe: %w", err)
	}

	return updated, nil
}

// mapStripeStatusToInternal maps Stripe subscription status to internal status
func mapStripeStatusToInternal(stripeStatus string) string {
	switch stripeStatus {
	case "active":
		return "active"
	case "canceled":
		return "cancelled"
	case "incomplete":
		return "incomplete"
	case "incomplete_expired":
		return "expired"
	case "past_due":
		return "past_due"
	case "paused":
		return "paused"
	case "trialing":
		return "trialing"
	case "unpaid":
		return "unpaid"
	default:
		return stripeStatus
	}
}

// GetTenantByStripeCustomerID retrieves a tenant by their Stripe customer ID
func (r *BillingRepository) GetTenantByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*Tenant, error) {
	query := `
		SELECT id, name, status, plan, stripe_customer_id, created_at, updated_at
		FROM tenants
		WHERE stripe_customer_id = $1`

	tenant := &Tenant{}
	var stripeCID sql.NullString
	err := r.db.QueryRowContext(ctx, query, stripeCustomerID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Status, &tenant.Plan,
		&stripeCID, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by stripe customer id: %w", err)
	}
	if stripeCID.Valid && stripeCID.String != "" {
		tenant.StripeCustomerID = &stripeCID.String
	}
	return tenant, nil
}

// UpdateTenantPaymentMethod updates the tenant's default payment method info from Stripe
func (r *BillingRepository) UpdateTenantPaymentMethod(ctx context.Context, tenantID uuid.UUID, paymentMethod *PaymentMethodInfoExtended) error {
	// First, mark any existing payment methods as non-default
	_, err := r.db.ExecContext(ctx, `
		UPDATE tenant_payment_methods
		SET is_default = false, updated_at = NOW()
		WHERE tenant_id = $1`,
		tenantID)
	if err != nil {
		return fmt.Errorf("failed to reset default payment methods: %w", err)
	}

	// Upsert the payment method
	query := `
		INSERT INTO tenant_payment_methods (
			id, tenant_id, stripe_payment_method_id, brand, last4, exp_month, exp_year, is_default, billing_details, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, true, $8, NOW(), NOW()
		)
		ON CONFLICT (tenant_id, stripe_payment_method_id)
		DO UPDATE SET
			brand = EXCLUDED.brand,
			last4 = EXCLUDED.last4,
			exp_month = EXCLUDED.exp_month,
			exp_year = EXCLUDED.exp_year,
			is_default = true,
			billing_details = EXCLUDED.billing_details,
			updated_at = NOW()
		RETURNING id`

	var id uuid.UUID
	err = r.db.QueryRowContext(ctx, query,
		uuid.New(), tenantID, paymentMethod.StripePaymentMethodID,
		paymentMethod.Brand, paymentMethod.Last4, paymentMethod.ExpMonth,
		paymentMethod.ExpYear, datatypes.JSON(paymentMethod.BillingDetails),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to upsert payment method: %w", err)
	}

	return nil
}

// GetPaymentMethodByStripeID retrieves a payment method by Stripe ID
func (r *BillingRepository) GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*PaymentMethodInfoExtended, error) {
	query := `
		SELECT id, tenant_id, stripe_payment_method_id, brand, last4, exp_month, exp_year, is_default, billing_details, created_at, updated_at
		FROM tenant_payment_methods
		WHERE stripe_payment_method_id = $1`

	pm := &PaymentMethodInfoExtended{}
	var billingDetails datatypes.JSON
	var tenantID uuid.UUID

	err := r.db.QueryRowContext(ctx, query, stripePaymentMethodID).Scan(
		&pm.ID, &tenantID, &pm.StripePaymentMethodID, &pm.Brand, &pm.Last4,
		&pm.ExpMonth, &pm.ExpYear, &pm.IsDefault, &billingDetails,
		&pm.CreatedAt, &pm.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment method: %w", err)
	}

	pm.TenantID = tenantID
	pm.BillingDetails = json.RawMessage(billingDetails)
	return pm, nil
}
