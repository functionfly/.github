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
func (r *BillingRepository) GetSubscriptionByTenantID(ctx context.Context, tenantID uuid.UUID) (*Subscription, error) {
	query := `
		SELECT s.id, s.tenant_id, s.pricing_tier_id, s.status, s.current_period_start, s.current_period_end,
			   s.trial_end, s.cancel_at_period_end, s.canceled_at, s.created_at, s.updated_at,
			   t.id, t.name, t.description, t.price_cents, t.currency, t.features, t.is_active, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN pricing_tiers t ON s.pricing_tier_id = t.id
		WHERE s.tenant_id = $1 AND s.status = 'active'`

	sub := &Subscription{}
	tier := &PricingTier{}
	var features []byte

	err := r.db.QueryRow(query, tenantID).Scan(
		&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.TrialEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency, &features,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

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
		SELECT s.id, s.tenant_id, s.pricing_tier_id, s.status, s.current_period_start, s.current_period_end,
			   s.trial_end, s.cancel_at_period_end, s.canceled_at, s.created_at, s.updated_at,
			   t.id, t.name, t.description, t.price_cents, t.currency, t.features, t.is_active, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN pricing_tiers t ON s.pricing_tier_id = t.id
		WHERE s.id = $1`

	sub := &Subscription{}
	tier := &PricingTier{}
	var features []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.TrialEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency, &features,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

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
	tier, err := r.GetPricingTierByID(ctx, updated.PricingTierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tier: %w", err)
	}
	updated.PricingTier = tier

	return updated, nil
}

// GetSubscriptionByStripeID retrieves a subscription by its Stripe subscription ID
func (r *BillingRepository) GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error) {
	query := `
		SELECT s.id, s.tenant_id, s.pricing_tier_id, s.status, s.current_period_start, s.current_period_end,
			   s.trial_end, s.cancel_at_period_end, s.canceled_at, s.created_at, s.updated_at,
			   t.id, t.name, t.description, t.price_cents, t.currency, t.features, t.is_active, t.created_at, t.updated_at
		FROM subscriptions s
		JOIN pricing_tiers t ON s.pricing_tier_id = t.id
		WHERE s.stripe_subscription_id = $1`

	sub := &Subscription{}
	tier := &PricingTier{}
	var features []byte

	err := r.db.QueryRowContext(ctx, query, stripeSubscriptionID).Scan(
		&sub.ID, &sub.TenantID, &sub.PricingTierID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.TrialEnd, &sub.CancelAtPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency, &features,
		&tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription by stripe id: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	sub.PricingTier = tier
	return sub, nil
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
func (r *BillingRepository) ListAllSubscriptions(ctx context.Context, limit, offset int) ([]*Subscription, error) {
	query := `
		SELECT id, tenant_id, pricing_tier_id, status, stripe_subscription_id, current_period_start, current_period_end, trial_end, cancel_at_period_end, canceled_at, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
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
