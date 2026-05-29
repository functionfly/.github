package studio

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PlanSubscriptionSyncInput is populated from Stripe subscription webhooks.
type PlanSubscriptionSyncInput struct {
	CreatorTenantID   string
	PlanID            string
	PlanName          string
	StripeSubID       string
	SubscriberTenant  string
	SubscriberUser    string
	SubscriberName    string
	SubscriberEmail   string
	Status            string
	Amount            float64
	BillingCycle      string
	PeriodStart       time.Time
	PeriodEnd         time.Time
	CancelAtPeriodEnd bool
}

// UpsertPlanSubscriptionFromStripe records or updates a marketplace plan subscription.
func (r *MarketplaceRepository) UpsertPlanSubscriptionFromStripe(ctx context.Context, in PlanSubscriptionSyncInput) error {
	status := normalizeSubscriptionStatus(in.Status)
	billingCycle := in.BillingCycle
	if billingCycle == "" {
		billingCycle = "monthly"
	}

	var planID interface{}
	if in.PlanID != "" {
		planID = in.PlanID
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO marketplace_plan_subscriptions (
			id, plan_id, creator_tenant_id, subscriber_tenant_id, subscriber_user_id,
			subscriber_name, subscriber_email, plan_name, status, amount, currency,
			billing_cycle, current_period_start, current_period_end, cancel_at_period_end,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1::uuid, $2::uuid, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid,
			$5, $6, $7, $8, $9, 'USD', $10, $11, $12, $13, NOW(), NOW()
		)
		ON CONFLICT DO NOTHING`,
		planID, in.CreatorTenantID, in.SubscriberTenant, in.SubscriberUser,
		in.SubscriberName, in.SubscriberEmail, in.PlanName, status, in.Amount, billingCycle,
		in.PeriodStart, in.PeriodEnd, in.CancelAtPeriodEnd,
	)
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		// Table may lack unique constraint on stripe id — fall through to update by metadata match.
		_ = err
	}

	// Update existing row matched by subscriber + creator + plan name when possible.
	_, err = r.db.ExecContext(ctx, `
		UPDATE marketplace_plan_subscriptions
		SET status = $8,
		    amount = $9,
		    billing_cycle = $10,
		    current_period_start = $11,
		    current_period_end = $12,
		    cancel_at_period_end = $13,
		    updated_at = NOW()
		WHERE creator_tenant_id = $2::uuid
		  AND plan_name = $7
		  AND subscriber_email = $6`,
		planID, in.CreatorTenantID, in.SubscriberTenant, in.SubscriberUser,
		in.SubscriberName, in.SubscriberEmail, in.PlanName, status, in.Amount, billingCycle,
		in.PeriodStart, in.PeriodEnd, in.CancelAtPeriodEnd,
	)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return fmt.Errorf("upsert plan subscription: %w", err)
	}
	return nil
}

func normalizeSubscriptionStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "trialing", "past_due", "cancelled", "canceled":
		if status == "canceled" {
			return "cancelled"
		}
		return strings.ToLower(status)
	default:
		return "active"
	}
}

// GetPlanByID loads a subscription plan owned by the creator tenant.
func (r *MarketplaceRepository) GetPlanByID(ctx context.Context, tenantID, planID string) (*SubscriptionPlan, error) {
	var plan SubscriptionPlan
	var featuresRaw []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, tenant_id::text, name, price, billing_cycle, features, created_at, updated_at
		FROM marketplace_subscription_plans
		WHERE id = $1::uuid AND tenant_id = $2::uuid AND active = true`,
		planID, tenantID,
	).Scan(&plan.ID, &plan.TenantID, &plan.Name, &plan.Price, &plan.BillingCycle, &featuresRaw, &plan.CreatedAt, &plan.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan not found")
	}
	if err != nil {
		return nil, err
	}
	_ = featuresRaw
	return &plan, nil
}

// RecordPlanSubscriptionCheckout creates a pending subscription row after checkout completes.
func (r *MarketplaceRepository) RecordPlanSubscriptionCheckout(
	ctx context.Context,
	creatorTenantID, planID, subscriberTenantID, subscriberUserID, subscriberName, subscriberEmail string,
	planName string,
	amount float64,
	billingCycle string,
) error {
	if billingCycle == "" {
		billingCycle = "monthly"
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO marketplace_plan_subscriptions (
			id, plan_id, creator_tenant_id, subscriber_tenant_id, subscriber_user_id,
			subscriber_name, subscriber_email, plan_name, status, amount, currency,
			billing_cycle, current_period_start, current_period_end, cancel_at_period_end
		) VALUES (
			$1::uuid, NULLIF($2, '')::uuid, $3::uuid, NULLIF($4, '')::uuid, NULLIF($5, '')::uuid,
			$6, $7, $8, 'active', $9, 'USD', $10, NOW(), NOW() + INTERVAL '30 days', false
		)`,
		uuid.New().String(), planID, creatorTenantID, subscriberTenantID, subscriberUserID,
		subscriberName, subscriberEmail, planName, amount, billingCycle,
	)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return nil
	}
	return err
}
