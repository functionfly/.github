package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BillingRepository handles billing-related database operations
type BillingRepository struct {
	db *PostgresDB
}

// NewBillingRepository creates a new billing repository
func NewBillingRepository(db *PostgresDB) *BillingRepository {
	return &BillingRepository{db: db}
}

// CreatePricingTier creates a new pricing tier
func (r *BillingRepository) CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error) {
	tier.ID = uuid.New()
	tier.CreatedAt = time.Now()
	tier.UpdatedAt = time.Now()

	query := `
		INSERT INTO pricing_tiers (id, name, description, price_cents, currency, features, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, description, price_cents, currency, features, is_active, created_at, updated_at`

	var features []byte
	if tier.Features != nil {
		features, _ = json.Marshal(tier.Features)
	}

	err := r.db.QueryRow(query, tier.ID, tier.Name, tier.Description, tier.PriceCents,
		tier.Currency, features, tier.IsActive, tier.CreatedAt, tier.UpdatedAt).Scan(
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents,
		&tier.Currency, &features, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create pricing tier: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	return tier, nil
}

// ListPricingTiers lists all active pricing tiers
func (r *BillingRepository) ListPricingTiers() ([]*PricingTier, error) {
	query := `SELECT id, name, description, price_cents, currency, features, is_active, created_at, updated_at
			  FROM pricing_tiers WHERE is_active = true ORDER BY price_cents ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricing tiers: %w", err)
	}
	defer rows.Close()

	var tiers []*PricingTier
	for rows.Next() {
		tier := &PricingTier{}
		var features []byte
		err := rows.Scan(&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents,
			&tier.Currency, &features, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricing tier: %w", err)
		}

		if len(features) > 0 {
			json.Unmarshal(features, &tier.Features)
		}

		tiers = append(tiers, tier)
	}

	return tiers, nil
}

// GetPricingTierByID retrieves a pricing tier by ID
func (r *BillingRepository) GetPricingTierByID(id uuid.UUID) (*PricingTier, error) {
	query := `SELECT id, name, description, price_cents, currency, features, is_active, created_at, updated_at
			  FROM pricing_tiers WHERE id = $1`

	tier := &PricingTier{}
	var features []byte
	err := r.db.QueryRow(query, id).Scan(&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents,
		&tier.Currency, &features, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tier: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	return tier, nil
}

// UpdatePricingTier updates pricing tier fields dynamically
func (r *BillingRepository) UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*PricingTier, error) {
	// Get current tier
	current, err := r.GetPricingTierByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pricing tier: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("pricing tier not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}

	if description, ok := updates["description"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, description)
		argIndex++
	}

	if priceCents, ok := updates["price_cents"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("price_cents = $%d", argIndex))
		args = append(args, priceCents)
		argIndex++
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, isActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return current, nil
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE pricing_tiers SET %s WHERE id = $%d RETURNING id, name, description, price_cents, currency, features, is_active, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	updated := &PricingTier{}
	var features []byte
	err = r.db.QueryRow(query, args...).Scan(&updated.ID, &updated.Name, &updated.Description,
		&updated.PriceCents, &updated.Currency, &features, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update pricing tier: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &updated.Features)
	}

	return updated, nil
}

// DeletePricingTier soft deletes a pricing tier
func (r *BillingRepository) DeletePricingTier(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec("UPDATE pricing_tiers SET is_active = false WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete pricing tier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pricing tier not found")
	}

	return nil
}

// CreateSubscription creates a new subscription
func (r *BillingRepository) CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	sub.Status = "active"

	query := `
		INSERT INTO subscriptions (id, tenant_id, pricing_tier_id, status, stripe_subscription_id, current_period_start, current_period_end, trial_end, cancel_at_period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, tenant_id, pricing_tier_id, status, stripe_subscription_id, current_period_start, current_period_end, trial_end, cancel_at_period_end, created_at, updated_at`

	var stripeID interface{}
	if sub.StripeSubscriptionID != "" {
		stripeID = sub.StripeSubscriptionID
	}
	var stripeReturn *string
	err := r.db.QueryRow(query, sub.ID, sub.TenantID, sub.PricingTierID, sub.Status, stripeID,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.TrialEnd, sub.CancelAtPeriodEnd,
		sub.CreatedAt, sub.UpdatedAt).Scan(&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &stripeReturn,
		&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.TrialEnd, &sub.CancelAtPeriodEnd,
		&sub.CreatedAt, &sub.UpdatedAt)
	if err == nil && stripeReturn != nil {
		sub.StripeSubscriptionID = *stripeReturn
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

// GetSubscriptionByTenantID retrieves a subscription by tenant ID
func (r *BillingRepository) GetSubscriptionByTenantID(tenantID uuid.UUID) (*Subscription, error) {
	query := `
		SELECT s.id, s.tenant_id, s.pricing_tier_id, s.status, s.stripe_subscription_id, s.current_period_start, s.current_period_end,
			   s.trial_end, s.cancel_at_period_end, s.canceled_at, s.created_at, s.updated_at,
			   t.id, t.name, t.description, t.price_cents, t.currency, t.features, t.is_active, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN pricing_tiers t ON s.pricing_tier_id = t.id
		WHERE s.tenant_id = $1 AND s.status = 'active'`

	sub := &Subscription{}
	tier := &PricingTier{}
	var features []byte
	var stripeSubID *string

	err := r.db.QueryRow(query, tenantID).Scan(
		&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &stripeSubID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.TrialEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency, &features,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)
	if err == nil && stripeSubID != nil {
		sub.StripeSubscriptionID = *stripeSubID
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	sub.PricingTier = tier
	return sub, nil
}

// GetSubscriptionByID retrieves a subscription by its ID with pricing tier information
func (r *BillingRepository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	query := `
		SELECT s.id, s.tenant_id, s.pricing_tier_id, s.status, s.stripe_subscription_id, s.current_period_start, s.current_period_end,
			   s.trial_end, s.cancel_at_period_end, s.canceled_at, s.created_at, s.updated_at,
			   t.id, t.name, t.description, t.price_cents, t.currency, t.features, t.is_active, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN pricing_tiers t ON s.pricing_tier_id = t.id
		WHERE s.id = $1`

	sub := &Subscription{}
	tier := &PricingTier{}
	var features []byte
	var stripeSubID *string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &stripeSubID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.TrialEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency, &features,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	if stripeSubID != nil {
		sub.StripeSubscriptionID = *stripeSubID
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	sub.PricingTier = tier
	return sub, nil
}

// UpdateSubscription updates subscription fields
func (r *BillingRepository) UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Subscription, error) {
	// Get current subscription
	current, err := r.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current subscription: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("subscription not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if pricingTierID, ok := updates["pricing_tier_id"].(string); ok {
		if tierUUID, err := uuid.Parse(pricingTierID); err == nil {
			setParts = append(setParts, fmt.Sprintf("pricing_tier_id = $%d", argIndex))
			args = append(args, tierUUID)
			argIndex++
		}
	}

	if currentPeriodStart, ok := updates["current_period_start"].(time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("current_period_start = $%d", argIndex))
		args = append(args, currentPeriodStart)
		argIndex++
	}

	if currentPeriodEnd, ok := updates["current_period_end"].(time.Time); ok {
		setParts = append(setParts, fmt.Sprintf("current_period_end = $%d", argIndex))
		args = append(args, currentPeriodEnd)
		argIndex++
	}

	if trialEnd, ok := updates["trial_end"]; ok {
		if trialEnd == nil {
			setParts = append(setParts, "trial_end = NULL")
		} else if trialEndTime, ok := trialEnd.(time.Time); ok {
			setParts = append(setParts, fmt.Sprintf("trial_end = $%d", argIndex))
			args = append(args, trialEndTime)
			argIndex++
		}
	}

	if cancelAtPeriodEnd, ok := updates["cancel_at_period_end"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("cancel_at_period_end = $%d", argIndex))
		args = append(args, cancelAtPeriodEnd)
		argIndex++
	}

	if canceledAt, ok := updates["canceled_at"]; ok {
		if canceledAt == nil {
			setParts = append(setParts, "canceled_at = NULL")
		} else if canceledAtTime, ok := canceledAt.(time.Time); ok {
			setParts = append(setParts, fmt.Sprintf("canceled_at = $%d", argIndex))
			args = append(args, canceledAtTime)
			argIndex++
		}
	}

	if stripeSubID, ok := updates["stripe_subscription_id"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("stripe_subscription_id = $%d", argIndex))
		var val interface{}
		if stripeSubID != "" {
			val = stripeSubID
		}
		args = append(args, val)
		argIndex++
	}

	if len(setParts) == 0 {
		return current, nil
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE subscriptions SET %s WHERE id = $%d RETURNING id, tenant_id, pricing_tier_id, status, stripe_subscription_id, current_period_start, current_period_end, trial_end, cancel_at_period_end, canceled_at, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	updated := &Subscription{}
	var stripeReturn *string
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&updated.ID, &updated.TenantID, &updated.PricingTierID, &updated.Status, &stripeReturn,
		&updated.CurrentPeriodStart, &updated.CurrentPeriodEnd, &updated.TrialEnd,
		&updated.CancelAtPeriodEnd, &updated.CanceledAt, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	if stripeReturn != nil {
		updated.StripeSubscriptionID = *stripeReturn
	}

	// Get the updated pricing tier information
	tier, err := r.GetPricingTierByID(updated.PricingTierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tier: %w", err)
	}
	updated.PricingTier = tier

	return updated, nil
}

// CancelSubscription cancels a subscription
func (r *BillingRepository) CancelSubscription(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(`
		UPDATE subscriptions
		SET status = 'canceled', canceled_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'active'`, id)

	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found or already canceled")
	}

	return nil
}

// ListAllSubscriptions lists all subscriptions across tenants (for admin).
func (r *BillingRepository) ListAllSubscriptions(limit, offset int) ([]*Subscription, error) {
	query := `
		SELECT id, tenant_id, pricing_tier_id, status, stripe_subscription_id, current_period_start, current_period_end, trial_end, cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()
	var subs []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		var stripeSubID *string
		err := rows.Scan(&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &stripeSubID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
			&sub.TrialEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		if stripeSubID != nil {
			sub.StripeSubscriptionID = *stripeSubID
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// CreateInvoice creates a new invoice
func (r *BillingRepository) CreateInvoice(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	invoice.ID = uuid.New()
	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()
	invoice.Status = "draft"

	query := `
		INSERT INTO invoices (id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency, period_start, period_end, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency, period_start, period_end, due_date, created_at, updated_at`

	err := r.db.QueryRow(query, invoice.ID, invoice.TenantID, invoice.SubscriptionID, invoice.Status,
		invoice.AmountDueCents, invoice.AmountPaidCents, invoice.Currency, invoice.PeriodStart,
		invoice.PeriodEnd, invoice.DueDate, invoice.CreatedAt, invoice.UpdatedAt).Scan(
		&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
		&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
		&invoice.PeriodStart, &invoice.PeriodEnd, &invoice.DueDate, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// ListInvoicesByTenant lists invoices for a tenant
func (r *BillingRepository) ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			   invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
			&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
			&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
			&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// ListAllInvoices lists all invoices across tenants (for admin dashboard)
func (r *BillingRepository) ListAllInvoices(limit, offset int) ([]*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			   invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
			&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
			&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
			&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// GetInvoiceByID retrieves an invoice by ID
func (r *BillingRepository) GetInvoiceByID(id uuid.UUID) (*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			   invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices WHERE id = $1`

	invoice := &Invoice{}
	err := r.db.QueryRow(query, id).Scan(&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
		&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
		&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
		&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// UpdateInvoice updates invoice fields dynamically
func (r *BillingRepository) UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Invoice, error) {
	// Get current invoice
	current, err := r.GetInvoiceByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current invoice: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if amountPaidCents, ok := updates["amount_paid_cents"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("amount_paid_cents = $%d", argIndex))
		args = append(args, amountPaidCents)
		argIndex++
		if amountPaidCents == current.AmountDueCents {
			setParts = append(setParts, "paid_at = NOW()")
		}
	}

	if len(setParts) == 0 {
		return current, nil
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE invoices SET %s WHERE id = $%d RETURNING id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency, invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	updated := &Invoice{}
	err = r.db.QueryRow(query, args...).Scan(&updated.ID, &updated.TenantID, &updated.SubscriptionID, &updated.Status,
		&updated.AmountDueCents, &updated.AmountPaidCents, &updated.Currency,
		&updated.InvoicePdfURL, &updated.HostedInvoiceURL, &updated.PeriodStart,
		&updated.PeriodEnd, &updated.DueDate, &updated.PaidAt, &updated.CreatedAt, &updated.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return updated, nil
}

// RecordUsageEvent records a usage event
func (r *BillingRepository) RecordUsageEvent(ctx context.Context, event *UsageEvent) error {
	event.ID = uuid.New()
	event.Timestamp = time.Now()

	query := `INSERT INTO usage_events (id, tenant_id, event_type, quantity, unit_price_cents, metadata, timestamp)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var metadata []byte
	if event.Metadata != nil {
		metadata, _ = json.Marshal(event.Metadata)
	}

	_, err := r.db.Exec(query, event.ID, event.TenantID, event.EventType, event.Quantity,
		event.UnitPriceCents, metadata, event.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to record usage event: %w", err)
	}

	return nil
}

// GetUsageByTenant gets usage rollups for a tenant
func (r *BillingRepository) GetUsageByTenant(tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error) {
	query := `
		SELECT id, tenant_id, event_type, period_date, total_quantity, created_at, updated_at
		FROM usage_rollups
		WHERE tenant_id = $1 AND event_type = $2 AND period_date >= $3 AND period_date <= $4
		ORDER BY period_date ASC`

	rows, err := r.db.Query(query, tenantID, eventType, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}
	defer rows.Close()

	var rollups []*UsageRollup
	for rows.Next() {
		rollup := &UsageRollup{}
		err := rows.Scan(&rollup.ID, &rollup.TenantID, &rollup.EventType, &rollup.PeriodDate,
			&rollup.TotalQuantity, &rollup.CreatedAt, &rollup.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage rollup: %w", err)
		}
		rollups = append(rollups, rollup)
	}

	return rollups, nil
}

// CreateCoupon creates a new coupon
func (r *BillingRepository) CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	coupon.ID = uuid.New()
	coupon.CreatedAt = time.Now()
	coupon.UpdatedAt = time.Now()
	coupon.TimesRedeemed = 0

	query := `
		INSERT INTO coupons (id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at`

	err := r.db.QueryRow(query, coupon.ID, coupon.Code, coupon.Name, coupon.Description,
		coupon.DiscountType, coupon.DiscountValue, coupon.MaxRedemptions, coupon.TimesRedeemed,
		coupon.ValidFrom, coupon.ValidUntil, coupon.IsActive, coupon.CreatedAt, coupon.UpdatedAt).Scan(
		&coupon.ID, &coupon.Code, &coupon.Name, &coupon.Description, &coupon.DiscountType,
		&coupon.DiscountValue, &coupon.MaxRedemptions, &coupon.TimesRedeemed,
		&coupon.ValidFrom, &coupon.ValidUntil, &coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create coupon: %w", err)
	}

	return coupon, nil
}

// ListCoupons lists all coupons
func (r *BillingRepository) ListCoupons() ([]*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list coupons: %w", err)
	}
	defer rows.Close()

	var coupons []*Coupon
	for rows.Next() {
		coupon := &Coupon{}
		err := rows.Scan(&coupon.ID, &coupon.Code, &coupon.Name, &coupon.Description,
			&coupon.DiscountType, &coupon.DiscountValue, &coupon.MaxRedemptions,
			&coupon.TimesRedeemed, &coupon.ValidFrom, &coupon.ValidUntil,
			&coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coupon: %w", err)
		}
		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// GetCouponByCode retrieves a coupon by code
func (r *BillingRepository) GetCouponByCode(code string) (*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons WHERE UPPER(code) = UPPER($1) AND is_active = true`

	coupon := &Coupon{}
	err := r.db.QueryRow(query, code).Scan(&coupon.ID, &coupon.Code, &coupon.Name,
		&coupon.Description, &coupon.DiscountType, &coupon.DiscountValue, &coupon.MaxRedemptions,
		&coupon.TimesRedeemed, &coupon.ValidFrom, &coupon.ValidUntil, &coupon.IsActive,
		&coupon.CreatedAt, &coupon.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return coupon, nil
}

// GetCouponByID retrieves a coupon by ID
func (r *BillingRepository) GetCouponByID(id uuid.UUID) (*Coupon, error) {
	query := `SELECT id, code, name, description, discount_type, discount_value, max_redemptions, times_redeemed, valid_from, valid_until, is_active, created_at, updated_at
			  FROM coupons WHERE id = $1`

	coupon := &Coupon{}
	err := r.db.QueryRow(query, id).Scan(&coupon.ID, &coupon.Code, &coupon.Name,
		&coupon.Description, &coupon.DiscountType, &coupon.DiscountValue, &coupon.MaxRedemptions,
		&coupon.TimesRedeemed, &coupon.ValidFrom, &coupon.ValidUntil, &coupon.IsActive,
		&coupon.CreatedAt, &coupon.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}

	return coupon, nil
}

// RedeemCoupon redeems a coupon for a tenant
func (r *BillingRepository) RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error) {
	// Check if coupon exists and is valid
	coupon, err := r.GetCouponByID(couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon not found")
	}

	// Check if already redeemed by this tenant
	var count int
	err = r.db.QueryRow("SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND tenant_id = $2", couponID, tenantID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check redemption: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("coupon already redeemed by this tenant")
	}

	// Check redemption limits
	if coupon.MaxRedemptions != nil && coupon.TimesRedeemed >= *coupon.MaxRedemptions {
		return nil, fmt.Errorf("coupon redemption limit exceeded")
	}

	// Create redemption record
	redemption := &CouponRedemption{
		ID:             uuid.New(),
		CouponID:       couponID,
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
		RedeemedAt:     time.Now(),
		Coupon:         coupon,
	}

	query := `INSERT INTO coupon_redemptions (id, coupon_id, tenant_id, subscription_id, redeemed_at)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err = r.db.Exec(query, redemption.ID, redemption.CouponID, redemption.TenantID,
		redemption.SubscriptionID, redemption.RedeemedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create redemption: %w", err)
	}

	// Update coupon redemption count
	_, err = r.db.Exec("UPDATE coupons SET times_redeemed = times_redeemed + 1 WHERE id = $1", couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to update coupon count: %w", err)
	}

	return redemption, nil
}
