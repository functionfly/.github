package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RevenueRepository handles database operations for revenue system
type RevenueRepository struct {
	db *sql.DB
}

// NewRevenueRepository creates a new revenue repository
func NewRevenueRepository(db *sql.DB) *RevenueRepository {
	return &RevenueRepository{db: db}
}

// =============================================================================
// Verification Fees
// =============================================================================

// GetVerificationFeeByLevel retrieves the verification fee for a given level
func (r *RevenueRepository) GetVerificationFeeByLevel(level string) (*VerificationFee, error) {
	query := `
		SELECT id, level, price_cents, currency, is_active, min_plan, description, created_at, updated_at
		FROM verification_fees
		WHERE level = $1 AND is_active = true`

	fee := &VerificationFee{}
	var minPlan sql.NullString
	err := r.db.QueryRow(query, level).Scan(
		&fee.ID, &fee.Level, &fee.PriceCents, &fee.Currency, &fee.IsActive,
		&minPlan, &fee.Description, &fee.CreatedAt, &fee.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if minPlan.Valid {
		fee.MinPlan = &minPlan.String
	}
	return fee, nil
}

// ListVerificationFees lists all active verification fees
func (r *RevenueRepository) ListVerificationFees() ([]*VerificationFee, error) {
	query := `
		SELECT id, level, price_cents, currency, is_active, min_plan, description, created_at, updated_at
		FROM verification_fees
		WHERE is_active = true
		ORDER BY price_cents ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fees []*VerificationFee
	for rows.Next() {
		fee := &VerificationFee{}
		var minPlan sql.NullString
		err := rows.Scan(
			&fee.ID, &fee.Level, &fee.PriceCents, &fee.Currency, &fee.IsActive,
			&minPlan, &fee.Description, &fee.CreatedAt, &fee.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if minPlan.Valid {
			fee.MinPlan = &minPlan.String
		}
		fees = append(fees, fee)
	}
	return fees, nil
}

// =============================================================================
// Function Verification Payments
// =============================================================================

// CreateFunctionVerificationPayment creates a new verification payment record
func (r *RevenueRepository) CreateFunctionVerificationPayment(ctx context.Context, payment *FunctionVerificationPayment) error {
	query := `
		INSERT INTO function_verification_payments
		(id, function_id, verification_level, amount_cents, currency, status,
		 stripe_payment_intent_id, stripe_checkout_session_id, tenant_id, paid_by,
		 verification_job_id, paid_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	now := time.Now()
	payment.ID = uuid.New()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.FunctionID, payment.VerificationLevel, payment.AmountCents,
		payment.Currency, payment.Status, payment.StripePaymentIntentID,
		payment.StripeCheckoutSessionID, payment.TenantID, payment.PaidBy,
		payment.VerificationJobID, payment.PaidAt, payment.CreatedAt, payment.UpdatedAt,
	)
	return err
}

// GetFunctionVerificationPaymentByID retrieves a verification payment by ID
func (r *RevenueRepository) GetFunctionVerificationPaymentByID(id uuid.UUID) (*FunctionVerificationPayment, error) {
	query := `
		SELECT id, function_id, verification_level, amount_cents, currency, status,
		       stripe_payment_intent_id, stripe_checkout_session_id, tenant_id, paid_by,
		       verification_job_id, paid_at, created_at, updated_at
		FROM function_verification_payments
		WHERE id = $1`

	payment := &FunctionVerificationPayment{}
	err := r.db.QueryRow(query, id).Scan(
		&payment.ID, &payment.FunctionID, &payment.VerificationLevel, &payment.AmountCents,
		&payment.Currency, &payment.Status, &payment.StripePaymentIntentID,
		&payment.StripeCheckoutSessionID, &payment.TenantID, &payment.PaidBy,
		&payment.VerificationJobID, &payment.PaidAt, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// GetFunctionVerificationPaymentByCheckoutSessionID retrieves a verification payment by Stripe checkout session ID
func (r *RevenueRepository) GetFunctionVerificationPaymentByCheckoutSessionID(ctx context.Context, sessionID string) (*FunctionVerificationPayment, error) {
	query := `
		SELECT id, function_id, verification_level, amount_cents, currency, status,
		       stripe_payment_intent_id, stripe_checkout_session_id, tenant_id, paid_by,
		       verification_job_id, paid_at, created_at, updated_at
		FROM function_verification_payments
		WHERE stripe_checkout_session_id = $1`

	payment := &FunctionVerificationPayment{}
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&payment.ID, &payment.FunctionID, &payment.VerificationLevel, &payment.AmountCents,
		&payment.Currency, &payment.Status, &payment.StripePaymentIntentID,
		&payment.StripeCheckoutSessionID, &payment.TenantID, &payment.PaidBy,
		&payment.VerificationJobID, &payment.PaidAt, &payment.CreatedAt, &payment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// UpdateFunctionVerificationPaymentStatus updates the status of a verification payment
func (r *RevenueRepository) UpdateFunctionVerificationPaymentStatus(ctx context.Context, id uuid.UUID, status string, stripePIID, stripeCheckoutSessionID *string) error {
	query := `
		UPDATE function_verification_payments
		SET status = $2,
		    stripe_payment_intent_id = COALESCE($3, stripe_payment_intent_id),
		    stripe_checkout_session_id = COALESCE($4, stripe_checkout_session_id),
		    updated_at = NOW(),
		    paid_at = CASE WHEN $2 = 'paid' THEN NOW() ELSE paid_at END
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, status, stripePIID, stripeCheckoutSessionID)
	return err
}

// GetFunctionVerificationPaymentsByTenant retrieves all verification payments for a tenant
func (r *RevenueRepository) GetFunctionVerificationPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*FunctionVerificationPayment, error) {
	query := `
		SELECT id, function_id, verification_level, amount_cents, currency, status,
		       stripe_payment_intent_id, stripe_checkout_session_id, tenant_id, paid_by,
		       verification_job_id, paid_at, created_at, updated_at
		FROM function_verification_payments
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*FunctionVerificationPayment
	for rows.Next() {
		payment := &FunctionVerificationPayment{}
		err := rows.Scan(
			&payment.ID, &payment.FunctionID, &payment.VerificationLevel, &payment.AmountCents,
			&payment.Currency, &payment.Status, &payment.StripePaymentIntentID,
			&payment.StripeCheckoutSessionID, &payment.TenantID, &payment.PaidBy,
			&payment.VerificationJobID, &payment.PaidAt, &payment.CreatedAt, &payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, nil
}

// =============================================================================
// Publisher Earnings
// =============================================================================

// CreatePublisherEarning creates a new publisher earning record
func (r *RevenueRepository) CreatePublisherEarning(ctx context.Context, earning *PublisherEarning) error {
	query := `
		INSERT INTO publisher_earnings
		(id, tenant_id, publisher_user_id, function_id, function_name, transaction_type,
		 amount_cents, currency, gross_amount_cents, platform_fee_cents, net_amount_cents,
		 platform_fee_percent, status, stripe_payout_id, period_month, period_year, metadata,
		 earned_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	now := time.Now()
	earning.ID = uuid.New()
	earning.CreatedAt = now
	earning.UpdatedAt = now
	if earning.EarnedAt.IsZero() {
		earning.EarnedAt = now
	}

	metadata, _ := json.Marshal(earning.Metadata)
	_, err := r.db.ExecContext(ctx, query,
		earning.ID, earning.TenantID, earning.PublisherUserID, earning.FunctionID,
		earning.FunctionName, earning.TransactionType, earning.AmountCents, earning.Currency,
		earning.GrossAmountCents, earning.PlatformFeeCents, earning.NetAmountCents,
		earning.PlatformFeePercent, earning.Status, earning.StripePayoutID,
		earning.PeriodMonth, earning.PeriodYear, metadata, earning.EarnedAt,
		earning.CreatedAt, earning.UpdatedAt,
	)
	return err
}

// GetPublisherEarningsByTenant retrieves earnings for a publisher tenant
func (r *RevenueRepository) GetPublisherEarningsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*PublisherEarning, error) {
	query := `
		SELECT id, tenant_id, publisher_user_id, function_id, function_name, transaction_type,
		       amount_cents, currency, gross_amount_cents, platform_fee_cents, net_amount_cents,
		       platform_fee_percent, status, stripe_payout_id, period_month, period_year,
		       metadata, earned_at, created_at, updated_at
		FROM publisher_earnings
		WHERE tenant_id = $1
		ORDER BY earned_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var earnings []*PublisherEarning
	for rows.Next() {
		earning := &PublisherEarning{}
		var metadata []byte
		err := rows.Scan(
			&earning.ID, &earning.TenantID, &earning.PublisherUserID, &earning.FunctionID,
			&earning.FunctionName, &earning.TransactionType, &earning.AmountCents, &earning.Currency,
			&earning.GrossAmountCents, &earning.PlatformFeeCents, &earning.NetAmountCents,
			&earning.PlatformFeePercent, &earning.Status, &earning.StripePayoutID,
			&earning.PeriodMonth, &earning.PeriodYear, &metadata, &earning.EarnedAt,
			&earning.CreatedAt, &earning.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if metadata != nil {
			earning.Metadata = metadata
		}
		earnings = append(earnings, earning)
	}
	return earnings, nil
}

// GetPublisherEarningsSummary returns aggregated earnings summary for a tenant
func (r *RevenueRepository) GetPublisherEarningsSummary(ctx context.Context, tenantID uuid.UUID) (pending, available, withdrawn int, err error) {
	query := `
		SELECT status, COALESCE(SUM(net_amount_cents), 0) as total
		FROM publisher_earnings
		WHERE tenant_id = $1
		GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var total int
		if err := rows.Scan(&status, &total); err != nil {
			return 0, 0, 0, err
		}
		switch status {
		case "pending":
			pending = total
		case "available":
			available = total
		case "withdrawn":
			withdrawn = total
		}
	}
	return pending, available, withdrawn, nil
}

// GetPublisherEarningsByPeriod returns earnings grouped by period
func (r *RevenueRepository) GetPublisherEarningsByPeriod(ctx context.Context, tenantID uuid.UUID, year int) ([]*PublisherEarning, error) {
	query := `
		SELECT id, tenant_id, publisher_user_id, function_id, function_name, transaction_type,
		       amount_cents, currency, gross_amount_cents, platform_fee_cents, net_amount_cents,
		       platform_fee_percent, status, stripe_payout_id, period_month, period_year,
		       metadata, earned_at, created_at, updated_at
		FROM publisher_earnings
		WHERE tenant_id = $1 AND period_year = $2
		ORDER BY period_year DESC, period_month DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var earnings []*PublisherEarning
	for rows.Next() {
		earning := &PublisherEarning{}
		var metadata []byte
		err := rows.Scan(
			&earning.ID, &earning.TenantID, &earning.PublisherUserID, &earning.FunctionID,
			&earning.FunctionName, &earning.TransactionType, &earning.AmountCents, &earning.Currency,
			&earning.GrossAmountCents, &earning.PlatformFeeCents, &earning.NetAmountCents,
			&earning.PlatformFeePercent, &earning.Status, &earning.StripePayoutID,
			&earning.PeriodMonth, &earning.PeriodYear, &metadata, &earning.EarnedAt,
			&earning.CreatedAt, &earning.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if metadata != nil {
			earning.Metadata = metadata
		}
		earnings = append(earnings, earning)
	}
	return earnings, nil
}

// =============================================================================
// Agent Subscriptions
// =============================================================================

// CreateAgentSubscription creates a new agent subscription
func (r *RevenueRepository) CreateAgentSubscription(ctx context.Context, sub *AgentSubscription) error {
	query := `
		INSERT INTO agent_subscriptions
		(id, agent_id, tenant_id, plan_name, price_per_agent_cents, currency, max_agents,
		 status, current_period_start, current_period_end, stripe_subscription_id,
		 stripe_customer_id, last_payment_status, last_payment_at, cancel_at_period_end,
		 cancelled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	now := time.Now()
	sub.ID = uuid.New()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.AgentID, sub.TenantID, sub.PlanName, sub.PricePerAgentCents,
		sub.Currency, sub.MaxAgents, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.StripeSubscriptionID, sub.StripeCustomerID, sub.LastPaymentStatus, sub.LastPaymentAt,
		sub.CancelAtPeriodEnd, sub.CancelledAt, sub.CreatedAt, sub.UpdatedAt,
	)
	return err
}

// GetAgentSubscriptionByAgentID retrieves an agent subscription by agent ID
func (r *RevenueRepository) GetAgentSubscriptionByAgentID(ctx context.Context, agentID uuid.UUID) (*AgentSubscription, error) {
	query := `
		SELECT id, agent_id, tenant_id, plan_name, price_per_agent_cents, currency, max_agents,
		       status, current_period_start, current_period_end, stripe_subscription_id,
		       stripe_customer_id, last_payment_status, last_payment_at, cancel_at_period_end,
		       cancelled_at, created_at, updated_at
		FROM agent_subscriptions
		WHERE agent_id = $1 AND status = 'active'`

	sub := &AgentSubscription{}
	err := r.db.QueryRowContext(ctx, query, agentID).Scan(
		&sub.ID, &sub.AgentID, &sub.TenantID, &sub.PlanName, &sub.PricePerAgentCents,
		&sub.Currency, &sub.MaxAgents, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.StripeSubscriptionID, &sub.StripeCustomerID, &sub.LastPaymentStatus, &sub.LastPaymentAt,
		&sub.CancelAtPeriodEnd, &sub.CancelledAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// GetAgentSubscriptionsByTenant retrieves all agent subscriptions for a tenant
func (r *RevenueRepository) GetAgentSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*AgentSubscription, error) {
	query := `
		SELECT id, agent_id, tenant_id, plan_name, price_per_agent_cents, currency, max_agents,
		       status, current_period_start, current_period_end, stripe_subscription_id,
		       stripe_customer_id, last_payment_status, last_payment_at, cancel_at_period_end,
		       cancelled_at, created_at, updated_at
		FROM agent_subscriptions
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*AgentSubscription
	for rows.Next() {
		sub := &AgentSubscription{}
		err := rows.Scan(
			&sub.ID, &sub.AgentID, &sub.TenantID, &sub.PlanName, &sub.PricePerAgentCents,
			&sub.Currency, &sub.MaxAgents, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
			&sub.StripeSubscriptionID, &sub.StripeCustomerID, &sub.LastPaymentStatus, &sub.LastPaymentAt,
			&sub.CancelAtPeriodEnd, &sub.CancelledAt, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// UpdateAgentSubscriptionStatus updates the status of an agent subscription
func (r *RevenueRepository) UpdateAgentSubscriptionStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE agent_subscriptions
		SET status = $2, updated_at = NOW(),
		    cancelled_at = CASE WHEN $2 = 'cancelled' THEN NOW() ELSE cancelled_at END
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

// =============================================================================
// Agent Usage
// =============================================================================

// CreateAgentUsage creates a new agent usage record
func (r *RevenueRepository) CreateAgentUsage(ctx context.Context, usage *AgentUsage) error {
	query := `
		INSERT INTO agent_usage
		(id, agent_id, tenant_id, subscription_id, period_start, period_end,
		 total_calls, total_executions, total_errors, total_latency_ms,
		 billable_calls, overage_calls, estimated_cost_cents, status,
		 stripe_invoice_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	now := time.Now()
	usage.ID = uuid.New()
	usage.CreatedAt = now
	usage.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		usage.ID, usage.AgentID, usage.TenantID, usage.SubscriptionID,
		usage.PeriodStart, usage.PeriodEnd, usage.TotalCalls, usage.TotalExecutions,
		usage.TotalErrors, usage.TotalLatencyMs, usage.BillableCalls, usage.OverageCalls,
		usage.EstimatedCostCents, usage.Status, usage.StripeInvoiceID,
		usage.CreatedAt, usage.UpdatedAt,
	)
	return err
}

// GetAgentUsageByAgentID retrieves agent usage records by agent ID
func (r *RevenueRepository) GetAgentUsageByAgentID(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*AgentUsage, error) {
	query := `
		SELECT id, agent_id, tenant_id, subscription_id, period_start, period_end,
		       total_calls, total_executions, total_errors, total_latency_ms,
		       billable_calls, overage_calls, estimated_cost_cents, status,
		       stripe_invoice_id, created_at, updated_at
		FROM agent_usage
		WHERE agent_id = $1
		ORDER BY period_start DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, agentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []*AgentUsage
	for rows.Next() {
		usage := &AgentUsage{}
		err := rows.Scan(
			&usage.ID, &usage.AgentID, &usage.TenantID, &usage.SubscriptionID,
			&usage.PeriodStart, &usage.PeriodEnd, &usage.TotalCalls, &usage.TotalExecutions,
			&usage.TotalErrors, &usage.TotalLatencyMs, &usage.BillableCalls, &usage.OverageCalls,
			&usage.EstimatedCostCents, &usage.Status, &usage.StripeInvoiceID,
			&usage.CreatedAt, &usage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		usages = append(usages, usage)
	}
	return usages, nil
}

// GetAgentUsageSummary returns aggregated usage summary for an agent
func (r *RevenueRepository) GetAgentUsageSummary(ctx context.Context, agentID uuid.UUID) (totalCalls, billableCalls, overageCalls, estimatedCost int, err error) {
	query := `
		SELECT COALESCE(SUM(total_calls), 0), COALESCE(SUM(billable_calls), 0),
		       COALESCE(SUM(overage_calls), 0), COALESCE(SUM(estimated_cost_cents), 0)
		FROM agent_usage
		WHERE agent_id = $1 AND status = 'active'`

	err = r.db.QueryRowContext(ctx, query, agentID).Scan(&totalCalls, &billableCalls, &overageCalls, &estimatedCost)
	return totalCalls, billableCalls, overageCalls, estimatedCost, err
}

// =============================================================================
// Platform Fees
// =============================================================================

// CreatePlatformFee creates a new platform fee record
func (r *RevenueRepository) CreatePlatformFee(ctx context.Context, fee *PlatformFee) error {
	query := `
		INSERT INTO platform_fees
		(id, fee_type, source_transaction_id, source_type, gross_amount_cents,
		 platform_fee_cents, net_amount_cents, platform_fee_percent, currency,
		 tenant_id, user_id, function_id, agent_id, status, stripe_transfer_id,
		 paid_out_at, period_month, period_year, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	fee.ID = uuid.New()
	fee.CreatedAt = time.Now()

	metadata, _ := json.Marshal(fee.Metadata)
	_, err := r.db.ExecContext(ctx, query,
		fee.ID, fee.FeeType, fee.SourceTransactionID, fee.SourceType,
		fee.GrossAmountCents, fee.PlatformFeeCents, fee.NetAmountCents,
		fee.PlatformFeePercent, fee.Currency, fee.TenantID, fee.UserID,
		fee.FunctionID, fee.AgentID, fee.Status, fee.StripeTransferID,
		fee.PaidOutAt, fee.PeriodMonth, fee.PeriodYear, metadata, fee.CreatedAt,
	)
	return err
}

// GetPlatformFeesByPeriod returns platform fees grouped by period
func (r *RevenueRepository) GetPlatformFeesByPeriod(ctx context.Context, year, month int) ([]*PlatformFee, error) {
	query := `
		SELECT id, fee_type, source_transaction_id, source_type, gross_amount_cents,
		       platform_fee_cents, net_amount_cents, platform_fee_percent, currency,
		       tenant_id, user_id, function_id, agent_id, status, stripe_transfer_id,
		       paid_out_at, period_month, period_year, metadata, created_at
		FROM platform_fees
		WHERE period_year = $1 AND period_month = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fees []*PlatformFee
	for rows.Next() {
		fee := &PlatformFee{}
		var metadata []byte
		err := rows.Scan(
			&fee.ID, &fee.FeeType, &fee.SourceTransactionID, &fee.SourceType,
			&fee.GrossAmountCents, &fee.PlatformFeeCents, &fee.NetAmountCents,
			&fee.PlatformFeePercent, &fee.Currency, &fee.TenantID, &fee.UserID,
			&fee.FunctionID, &fee.AgentID, &fee.Status, &fee.StripeTransferID,
			&fee.PaidOutAt, &fee.PeriodMonth, &fee.PeriodYear, &metadata, &fee.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if metadata != nil {
			fee.Metadata = metadata
		}
		fees = append(fees, fee)
	}
	return fees, nil
}

// GetPlatformFeesSummary returns aggregated platform fees summary
func (r *RevenueRepository) GetPlatformFeesSummary(ctx context.Context) (totalCollected, totalRefunded, totalPaidOut int, err error) {
	query := `
		SELECT status, COALESCE(SUM(platform_fee_cents), 0) as total
		FROM platform_fees
		GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var total int
		if err := rows.Scan(&status, &total); err != nil {
			return 0, 0, 0, err
		}
		switch status {
		case "collected":
			totalCollected = total
		case "refunded":
			totalRefunded = total
		case "paid_out":
			totalPaidOut = total
		}
	}
	return totalCollected, totalRefunded, totalPaidOut, nil
}

// =============================================================================
// Pricing Tier Extended (with Moat fields)
// =============================================================================

// ListPricingTiersExtended lists all active pricing tiers with extended fields
func (r *RevenueRepository) ListPricingTiersExtended() ([]*PricingTierExtended, error) {
	query := `
		SELECT id, name, description, price_cents, currency, features, is_active,
		       COALESCE(tier_type, 'subscription') as tier_type,
		       stripe_price_id, trial_days, max_agents, max_functions, max_executions_per_month,
		       created_at, updated_at
		FROM pricing_tiers
		WHERE is_active = true
		ORDER BY price_cents ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []*PricingTierExtended
	for rows.Next() {
		tier := &PricingTierExtended{}
		var features []byte
		var stripePriceID sql.NullString
		var tierType sql.NullString
		var trialDays, maxAgents, maxFunctions, maxExecutions sql.NullInt64

		err := rows.Scan(
			&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency,
			&features, &tier.IsActive, &tierType, &stripePriceID, &trialDays,
			&maxAgents, &maxFunctions, &maxExecutions, &tier.CreatedAt, &tier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if features != nil {
			tier.Features = features
		}
		if stripePriceID.Valid {
			tier.StripePriceID = &stripePriceID.String
		}
		if tierType.Valid {
			tier.TierType = tierType.String
		}
		if trialDays.Valid {
			tier.TrialDays = int(trialDays.Int64)
		}
		if maxAgents.Valid {
			tier.MaxAgents = int(maxAgents.Int64)
		}
		if maxFunctions.Valid {
			tier.MaxFunctions = int(maxFunctions.Int64)
		}
		if maxExecutions.Valid {
			tier.MaxExecutionsPerMonth = int(maxExecutions.Int64)
		}

		tiers = append(tiers, tier)
	}
	return tiers, nil
}

// GetPricingTierExtendedByID retrieves a pricing tier by ID with extended fields
func (r *RevenueRepository) GetPricingTierExtendedByID(id uuid.UUID) (*PricingTierExtended, error) {
	query := `
		SELECT id, name, description, price_cents, currency, features, is_active,
		       COALESCE(tier_type, 'subscription') as tier_type,
		       stripe_price_id, trial_days, max_agents, max_functions, max_executions_per_month,
		       created_at, updated_at
		FROM pricing_tiers
		WHERE id = $1`

	tier := &PricingTierExtended{}
	var features []byte
	var stripePriceID sql.NullString
	var tierType sql.NullString
	var trialDays, maxAgents, maxFunctions, maxExecutions sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents, &tier.Currency,
		&features, &tier.IsActive, &tierType, &stripePriceID, &trialDays,
		&maxAgents, &maxFunctions, &maxExecutions, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if features != nil {
		tier.Features = features
	}
	if stripePriceID.Valid {
		tier.StripePriceID = &stripePriceID.String
	}
	if tierType.Valid {
		tier.TierType = tierType.String
	}
	if trialDays.Valid {
		tier.TrialDays = int(trialDays.Int64)
	}
	if maxAgents.Valid {
		tier.MaxAgents = int(maxAgents.Int64)
	}
	if maxFunctions.Valid {
		tier.MaxFunctions = int(maxFunctions.Int64)
	}
	if maxExecutions.Valid {
		tier.MaxExecutionsPerMonth = int(maxExecutions.Int64)
	}

	return tier, nil
}

// =============================================================================
// Helper: Calculate platform fee
// =============================================================================

// CalculatePlatformFee calculates the platform fee for a given gross amount
func CalculatePlatformFee(grossAmountCents int, feePercent float64) (platformFeeCents, netAmountCents int) {
	platformFeeCents = int(float64(grossAmountCents) * feePercent / 100)
	netAmountCents = grossAmountCents - platformFeeCents
	return platformFeeCents, netAmountCents
}

// DefaultPlatformFeePercent is the default platform fee percentage (15%)
const DefaultPlatformFeePercent = 15.0

// Platform fee calculation using default
func CalculateDefaultPlatformFee(grossAmountCents int) (platformFeeCents, netAmountCents int) {
	return CalculatePlatformFee(grossAmountCents, DefaultPlatformFeePercent)
}

// VerifyFunctionCost returns the cost for verifying a function at a given level
func VerifyFunctionCost(fee *VerificationFee, tenantPlan *string) (int, error) {
	if fee == nil {
		return 0, fmt.Errorf("verification fee not found")
	}

	// Check if tenant plan meets minimum requirement
	if fee.MinPlan != nil && tenantPlan != nil {
		planHierarchy := map[string]int{"free": 0, "starter": 1, "pro": 2, "scale": 3, "enterprise": 4}
		requiredPlan := *fee.MinPlan
		tenantPlanRank, tenantOk := planHierarchy[*tenantPlan]
		requiredRank, requiredOk := planHierarchy[requiredPlan]
		if tenantOk && requiredOk && tenantPlanRank < requiredRank {
			return 0, fmt.Errorf("plan %s required for %s verification", requiredPlan, fee.Level)
		}
	}

	return fee.PriceCents, nil
}
