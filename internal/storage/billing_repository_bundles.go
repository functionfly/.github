package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreatePricingBundle creates a new Backend-in-a-Box pricing bundle
func (r *BillingRepository) CreatePricingBundle(ctx context.Context, bundle *PricingBundle) (*PricingBundle, error) {
	bundle.ID = uuid.New()
	bundle.CreatedAt = time.Now()
	bundle.UpdatedAt = time.Now()

	featuresJSON, _ := json.Marshal(bundle.FeaturesIncluded)
	limitsJSON, _ := json.Marshal(bundle.FeatureLimits)
	provisioningJSON, _ := json.Marshal(bundle.ProvisioningTemplates)

	query := `
		INSERT INTO pricing_bundles (
			id, slug, name, display_name, description, short_description,
			display_price_cents, billing_interval, stripe_price_id, icon, color,
			features_included, feature_limits, provisioning_templates,
			sort_order, is_active, is_popular, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, slug, name, display_name, description, short_description,
			display_price_cents, billing_interval, stripe_price_id, icon, color,
			features_included, feature_limits, provisioning_templates,
			sort_order, is_active, is_popular, created_at, updated_at`

	var returnedFeatures, returnedLimits, returnedProvisioning []byte
	err := r.db.QueryRow(query,
		bundle.ID, bundle.Slug, bundle.Name, bundle.DisplayName, bundle.Description, bundle.ShortDescription,
		bundle.DisplayPriceCents, bundle.BillingInterval, bundle.StripePriceID, bundle.Icon, bundle.Color,
		featuresJSON, limitsJSON, provisioningJSON,
		bundle.SortOrder, bundle.IsActive, bundle.IsPopular, bundle.CreatedAt, bundle.UpdatedAt,
	).Scan(
		&bundle.ID, &bundle.Slug, &bundle.Name, &bundle.DisplayName, &bundle.Description, &bundle.ShortDescription,
		&bundle.DisplayPriceCents, &bundle.BillingInterval, &bundle.StripePriceID, &bundle.Icon, &bundle.Color,
		&returnedFeatures, &returnedLimits, &returnedProvisioning,
		&bundle.SortOrder, &bundle.IsActive, &bundle.IsPopular, &bundle.CreatedAt, &bundle.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create pricing bundle: %w", err)
	}

	json.Unmarshal(returnedFeatures, &bundle.FeaturesIncluded)
	json.Unmarshal(returnedLimits, &bundle.FeatureLimits)
	json.Unmarshal(returnedProvisioning, &bundle.ProvisioningTemplates)

	return bundle, nil
}

// ListPricingBundles returns all pricing bundles, optionally filtering by active status
func (r *BillingRepository) ListPricingBundles(ctx context.Context, activeOnly bool) ([]*PricingBundle, error) {
	query := `SELECT id, slug, name, display_name, description, short_description,
		display_price_cents, billing_interval, COALESCE(stripe_price_id, '') as stripe_price_id, icon, color,
		features_included, feature_limits, provisioning_templates,
		sort_order, is_active, is_popular, created_at, updated_at
		FROM pricing_bundles `

	if activeOnly {
		query += `WHERE is_active = true `
	}
	query += `ORDER BY sort_order ASC, display_price_cents ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricing bundles: %w", err)
	}
	defer rows.Close()

	var bundles []*PricingBundle
	for rows.Next() {
		bundle := &PricingBundle{}
		var features, limits, provisioning []byte

		err := rows.Scan(
			&bundle.ID, &bundle.Slug, &bundle.Name, &bundle.DisplayName, &bundle.Description, &bundle.ShortDescription,
			&bundle.DisplayPriceCents, &bundle.BillingInterval, &bundle.StripePriceID, &bundle.Icon, &bundle.Color,
			&features, &limits, &provisioning,
			&bundle.SortOrder, &bundle.IsActive, &bundle.IsPopular, &bundle.CreatedAt, &bundle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricing bundle: %w", err)
		}

		json.Unmarshal(features, &bundle.FeaturesIncluded)
		json.Unmarshal(limits, &bundle.FeatureLimits)
		json.Unmarshal(provisioning, &bundle.ProvisioningTemplates)

		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// GetPricingBundleBySlug retrieves a pricing bundle by its slug
func (r *BillingRepository) GetPricingBundleBySlug(ctx context.Context, slug string) (*PricingBundle, error) {
	query := `SELECT id, slug, name, display_name, description, short_description,
		display_price_cents, billing_interval, COALESCE(stripe_price_id, '') as stripe_price_id, icon, color,
		features_included, feature_limits, provisioning_templates,
		sort_order, is_active, is_popular, created_at, updated_at
		FROM pricing_bundles WHERE slug = $1`

	bundle := &PricingBundle{}
	var features, limits, provisioning []byte

	err := r.db.QueryRow(query, slug).Scan(
		&bundle.ID, &bundle.Slug, &bundle.Name, &bundle.DisplayName, &bundle.Description, &bundle.ShortDescription,
		&bundle.DisplayPriceCents, &bundle.BillingInterval, &bundle.StripePriceID, &bundle.Icon, &bundle.Color,
		&features, &limits, &provisioning,
		&bundle.SortOrder, &bundle.IsActive, &bundle.IsPopular, &bundle.CreatedAt, &bundle.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing bundle: %w", err)
	}

	json.Unmarshal(features, &bundle.FeaturesIncluded)
	json.Unmarshal(limits, &bundle.FeatureLimits)
	json.Unmarshal(provisioning, &bundle.ProvisioningTemplates)

	return bundle, nil
}

// GetPricingBundleByID retrieves a pricing bundle by ID
func (r *BillingRepository) GetPricingBundleByID(ctx context.Context, id uuid.UUID) (*PricingBundle, error) {
	query := `SELECT id, slug, name, display_name, description, short_description,
		display_price_cents, billing_interval, COALESCE(stripe_price_id, '') as stripe_price_id, icon, color,
		features_included, feature_limits, provisioning_templates,
		sort_order, is_active, is_popular, created_at, updated_at
		FROM pricing_bundles WHERE id = $1`

	bundle := &PricingBundle{}
	var features, limits, provisioning []byte

	err := r.db.QueryRow(query, id).Scan(
		&bundle.ID, &bundle.Slug, &bundle.Name, &bundle.DisplayName, &bundle.Description, &bundle.ShortDescription,
		&bundle.DisplayPriceCents, &bundle.BillingInterval, &bundle.StripePriceID, &bundle.Icon, &bundle.Color,
		&features, &limits, &provisioning,
		&bundle.SortOrder, &bundle.IsActive, &bundle.IsPopular, &bundle.CreatedAt, &bundle.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing bundle: %w", err)
	}

	json.Unmarshal(features, &bundle.FeaturesIncluded)
	json.Unmarshal(limits, &bundle.FeatureLimits)
	json.Unmarshal(provisioning, &bundle.ProvisioningTemplates)

	return bundle, nil
}

// UpdatePricingBundleStripePrice updates the Stripe Price ID for a pricing bundle
func (r *BillingRepository) UpdatePricingBundleStripePrice(ctx context.Context, slug, stripePriceID string) error {
	query := `UPDATE pricing_bundles SET stripe_price_id = $1, updated_at = NOW() WHERE slug = $2`
	result, err := r.db.ExecContext(ctx, query, stripePriceID, slug)
	if err != nil {
		return fmt.Errorf("failed to update pricing bundle stripe price: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("pricing bundle not found: %s", slug)
	}

	return nil
}

// GetPricingBundleByStripePriceID retrieves a pricing bundle by its Stripe Price ID
// This is used when processing plan changes from Stripe webhooks
func (r *BillingRepository) GetPricingBundleByStripePriceID(ctx context.Context, stripePriceID string) (*PricingBundle, error) {
	query := `SELECT id, slug, name, display_name, description, short_description,
		display_price_cents, billing_interval, COALESCE(stripe_price_id, '') as stripe_price_id, icon, color,
		features_included, feature_limits, provisioning_templates,
		sort_order, is_active, is_popular, created_at, updated_at
		FROM pricing_bundles WHERE stripe_price_id = $1 AND is_active = true`

	bundle := &PricingBundle{}
	var features, limits, provisioning []byte

	err := r.db.QueryRowContext(ctx, query, stripePriceID).Scan(
		&bundle.ID, &bundle.Slug, &bundle.Name, &bundle.DisplayName, &bundle.Description, &bundle.ShortDescription,
		&bundle.DisplayPriceCents, &bundle.BillingInterval, &bundle.StripePriceID, &bundle.Icon, &bundle.Color,
		&features, &limits, &provisioning,
		&bundle.SortOrder, &bundle.IsActive, &bundle.IsPopular, &bundle.CreatedAt, &bundle.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing bundle by stripe price id: %w", err)
	}

	json.Unmarshal(features, &bundle.FeaturesIncluded)
	json.Unmarshal(limits, &bundle.FeatureLimits)
	json.Unmarshal(provisioning, &bundle.ProvisioningTemplates)

	return bundle, nil
}

// CreateFounderModeRegistration creates a new founder mode registration
func (r *BillingRepository) CreateFounderModeRegistration(ctx context.Context, reg *FounderModeRegistration) error {
	query := `
		INSERT INTO founder_mode_registrations (
			id, tenant_id, bundle_id, mode_type, started_at, ends_at,
			free_days, mrr_threshold_cents, status, max_users_seen,
			max_mrr_seen_cents, max_api_calls_monthly, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.Exec(query,
		reg.ID, reg.TenantID, reg.BundleID, reg.ModeType, reg.StartedAt, reg.EndsAt,
		reg.FreeDays, reg.MRRThresholdCents, reg.Status, reg.MaxUsersSeen,
		reg.MaxMRRSeenCents, reg.MaxAPICallsMonthly, reg.CreatedAt, reg.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create founder mode registration: %w", err)
	}

	return nil
}

// GetActiveFounderMode retrieves the active founder mode registration for a tenant and bundle
func (r *BillingRepository) GetActiveFounderMode(ctx context.Context, tenantID, bundleID uuid.UUID) (*FounderModeRegistration, error) {
	query := `SELECT id, tenant_id, bundle_id, mode_type, started_at, ends_at,
		free_days, mrr_threshold_cents, status, converted_to_bundle_id,
		converted_at, stripe_subscription_id, grace_period_started_at,
		grace_period_ends_at, max_users_seen, max_mrr_seen_cents,
		max_api_calls_monthly, created_at, updated_at
		FROM founder_mode_registrations
		WHERE tenant_id = $1 AND bundle_id = $2 AND status IN ('active', 'grace_period')`

	reg := &FounderModeRegistration{}
	var convertedToBundleID sql.NullString
	var convertedAt, graceStartedAt, graceEndsAt sql.NullTime

	err := r.db.QueryRow(query, tenantID, bundleID).Scan(
		&reg.ID, &reg.TenantID, &reg.BundleID, &reg.ModeType, &reg.StartedAt, &reg.EndsAt,
		&reg.FreeDays, &reg.MRRThresholdCents, &reg.Status, &convertedToBundleID,
		&convertedAt, &reg.StripeSubscriptionID, &graceStartedAt,
		&graceEndsAt, &reg.MaxUsersSeen, &reg.MaxMRRSeenCents,
		&reg.MaxAPICallsMonthly, &reg.CreatedAt, &reg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get founder mode registration: %w", err)
	}

	if convertedToBundleID.Valid {
		id, _ := uuid.Parse(convertedToBundleID.String)
		reg.ConvertedToBundleID = &id
	}
	if convertedAt.Valid {
		reg.ConvertedAt = &convertedAt.Time
	}
	if graceStartedAt.Valid {
		reg.GracePeriodStartedAt = &graceStartedAt.Time
	}
	if graceEndsAt.Valid {
		reg.GracePeriodEndsAt = &graceEndsAt.Time
	}

	return reg, nil
}

// ListFounderModesByTenant retrieves all founder mode registrations for a tenant
func (r *BillingRepository) ListFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FounderModeRegistration, error) {
	query := `SELECT id, tenant_id, bundle_id, mode_type, started_at, ends_at,
		free_days, mrr_threshold_cents, status, converted_to_bundle_id,
		converted_at, stripe_subscription_id, grace_period_started_at,
		grace_period_ends_at, max_users_seen, max_mrr_seen_cents,
		max_api_calls_monthly, created_at, updated_at
		FROM founder_mode_registrations
		WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list founder modes: %w", err)
	}
	defer rows.Close()

	var regs []*FounderModeRegistration
	for rows.Next() {
		reg := &FounderModeRegistration{}
		var convertedToBundleID sql.NullString
		var convertedAt, graceStartedAt, graceEndsAt sql.NullTime

		err := rows.Scan(
			&reg.ID, &reg.TenantID, &reg.BundleID, &reg.ModeType, &reg.StartedAt, &reg.EndsAt,
			&reg.FreeDays, &reg.MRRThresholdCents, &reg.Status, &convertedToBundleID,
			&convertedAt, &reg.StripeSubscriptionID, &graceStartedAt,
			&graceEndsAt, &reg.MaxUsersSeen, &reg.MaxMRRSeenCents,
			&reg.MaxAPICallsMonthly, &reg.CreatedAt, &reg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan founder mode registration: %w", err)
		}

		if convertedToBundleID.Valid {
			id, _ := uuid.Parse(convertedToBundleID.String)
			reg.ConvertedToBundleID = &id
		}
		if convertedAt.Valid {
			reg.ConvertedAt = &convertedAt.Time
		}
		if graceStartedAt.Valid {
			reg.GracePeriodStartedAt = &graceStartedAt.Time
		}
		if graceEndsAt.Valid {
			reg.GracePeriodEndsAt = &graceEndsAt.Time
		}

		regs = append(regs, reg)
	}

	return regs, nil
}

// ListActiveFounderModesByTenant retrieves all active founder mode registrations for a tenant
func (r *BillingRepository) ListActiveFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FounderModeRegistration, error) {
	query := `SELECT id, tenant_id, bundle_id, mode_type, started_at, ends_at,
		free_days, mrr_threshold_cents, status, converted_to_bundle_id,
		converted_at, stripe_subscription_id, grace_period_started_at,
		grace_period_ends_at, max_users_seen, max_mrr_seen_cents,
		max_api_calls_monthly, created_at, updated_at
		FROM founder_mode_registrations
		WHERE tenant_id = $1 AND status IN ('active', 'grace_period') ORDER BY created_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active founder modes: %w", err)
	}
	defer rows.Close()

	var regs []*FounderModeRegistration
	for rows.Next() {
		reg := &FounderModeRegistration{}
		var convertedToBundleID sql.NullString
		var convertedAt, graceStartedAt, graceEndsAt sql.NullTime

		err := rows.Scan(
			&reg.ID, &reg.TenantID, &reg.BundleID, &reg.ModeType, &reg.StartedAt, &reg.EndsAt,
			&reg.FreeDays, &reg.MRRThresholdCents, &reg.Status, &convertedToBundleID,
			&convertedAt, &reg.StripeSubscriptionID, &graceStartedAt,
			&graceEndsAt, &reg.MaxUsersSeen, &reg.MaxMRRSeenCents,
			&reg.MaxAPICallsMonthly, &reg.CreatedAt, &reg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan founder mode registration: %w", err)
		}

		if convertedToBundleID.Valid {
			id, _ := uuid.Parse(convertedToBundleID.String)
			reg.ConvertedToBundleID = &id
		}
		if convertedAt.Valid {
			reg.ConvertedAt = &convertedAt.Time
		}
		if graceStartedAt.Valid {
			reg.GracePeriodStartedAt = &graceStartedAt.Time
		}
		if graceEndsAt.Valid {
			reg.GracePeriodEndsAt = &graceEndsAt.Time
		}

		regs = append(regs, reg)
	}

	return regs, nil
}

// UpdateFounderModeStatus updates the status of a founder mode registration
func (r *BillingRepository) UpdateFounderModeStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE founder_mode_registrations SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update founder mode status: %w", err)
	}
	return nil
}

// UpdateFounderModeProgress updates the usage metrics for founder mode tracking
func (r *BillingRepository) UpdateFounderModeProgress(ctx context.Context, id uuid.UUID, users, mrrCents, apiCalls int) error {
	query := `UPDATE founder_mode_registrations SET
		max_users_seen = GREATEST(max_users_seen, $1),
		max_mrr_seen_cents = GREATEST(max_mrr_seen_cents, $2),
		max_api_calls_monthly = GREATEST(max_api_calls_monthly, $3),
		updated_at = NOW()
		WHERE id = $4`
	_, err := r.db.Exec(query, users, mrrCents, apiCalls, id)
	if err != nil {
		return fmt.Errorf("failed to update founder mode progress: %w", err)
	}
	return nil
}

// CreateBundleSubscription creates a new bundle subscription
func (r *BillingRepository) CreateBundleSubscription(ctx context.Context, sub *BundleSubscription) error {
	query := `
		INSERT INTO bundle_subscriptions (
			id, tenant_id, bundle_id, founder_mode_id, converted_from_founder_mode,
			status, stripe_subscription_id, current_period_start, current_period_end,
			cancel_at_period_end, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(query,
		sub.ID, sub.TenantID, sub.BundleID, sub.FounderModeID, sub.ConvertedFromFounderMode,
		sub.Status, sub.StripeSubscriptionID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.CreatedAt, sub.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create bundle subscription: %w", err)
	}

	return nil
}

// GetBundleSubscriptionByTenant retrieves the active bundle subscription for a tenant
func (r *BillingRepository) GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*BundleSubscription, error) {
	query := `SELECT id, tenant_id, bundle_id, founder_mode_id, converted_from_founder_mode,
		status, stripe_subscription_id, default_app_id, current_period_start, current_period_end,
		cancel_at_period_end, canceled_at, created_at, updated_at
		FROM bundle_subscriptions
		WHERE tenant_id = $1 AND status IN ('active', 'deferred')`

	sub := &BundleSubscription{}
	var founderModeID sql.NullString
	var canceledAt sql.NullTime
	var defaultAppID sql.NullString

	err := r.db.QueryRow(query, tenantID).Scan(
		&sub.ID, &sub.TenantID, &sub.BundleID, &founderModeID, &sub.ConvertedFromFounderMode,
		&sub.Status, &sub.StripeSubscriptionID, &defaultAppID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &canceledAt, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle subscription: %w", err)
	}

	if founderModeID.Valid {
		id, _ := uuid.Parse(founderModeID.String)
		sub.FounderModeID = &id
	}
	if defaultAppID.Valid {
		id, _ := uuid.Parse(defaultAppID.String)
		sub.DefaultAppID = &id
	}
	if canceledAt.Valid {
		sub.CanceledAt = &canceledAt.Time
	}

	return sub, nil
}

// GetBundleSubscriptionByStripeID retrieves a bundle subscription by Stripe subscription ID
func (r *BillingRepository) GetBundleSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*BundleSubscription, error) {
	query := `SELECT id, tenant_id, bundle_id, founder_mode_id, converted_from_founder_mode,
		status, stripe_subscription_id, default_app_id, current_period_start, current_period_end,
		cancel_at_period_end, canceled_at, created_at, updated_at
		FROM bundle_subscriptions
		WHERE stripe_subscription_id = $1 AND status IN ('active', 'deferred')`

	sub := &BundleSubscription{}
	var founderModeID sql.NullString
	var canceledAt sql.NullTime
	var defaultAppID sql.NullString

	err := r.db.QueryRow(query, stripeSubID).Scan(
		&sub.ID, &sub.TenantID, &sub.BundleID, &founderModeID, &sub.ConvertedFromFounderMode,
		&sub.Status, &sub.StripeSubscriptionID, &defaultAppID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd, &canceledAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle subscription by stripe id: %w", err)
	}

	if founderModeID.Valid {
		id, _ := uuid.Parse(founderModeID.String)
		sub.FounderModeID = &id
	}
	if defaultAppID.Valid {
		id, _ := uuid.Parse(defaultAppID.String)
		sub.DefaultAppID = &id
	}
	if canceledAt.Valid {
		sub.CanceledAt = &canceledAt.Time
	}

	return sub, nil
}

// UpdateBundleSubscription updates an existing bundle subscription
func (r *BillingRepository) UpdateBundleSubscription(ctx context.Context, sub *BundleSubscription) error {
	query := `
		UPDATE bundle_subscriptions SET
			bundle_id = $2,
			founder_mode_id = $3,
			converted_from_founder_mode = $4,
			status = $5,
			stripe_subscription_id = $6,
			default_app_id = $7,
			current_period_start = $8,
			current_period_end = $9,
			cancel_at_period_end = $10,
			canceled_at = $11,
			updated_at = $12
		WHERE id = $1`

	result, err := r.db.Exec(query,
		sub.ID, sub.BundleID, sub.FounderModeID, sub.ConvertedFromFounderMode,
		sub.Status, sub.StripeSubscriptionID, sub.DefaultAppID,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.CanceledAt, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update bundle subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("bundle subscription not found")
	}

	return nil
}

// ListBundleSubscriptionsByTenant retrieves all bundle subscriptions for a tenant
func (r *BillingRepository) ListBundleSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*BundleSubscription, error) {
	query := `SELECT id, tenant_id, bundle_id, founder_mode_id, converted_from_founder_mode,
		status, stripe_subscription_id, default_app_id, current_period_start, current_period_end,
		cancel_at_period_end, canceled_at, created_at, updated_at
		FROM bundle_subscriptions WHERE tenant_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list bundle subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*BundleSubscription
	for rows.Next() {
		sub := &BundleSubscription{}
		var founderModeID sql.NullString
		var canceledAt sql.NullTime
		var defaultAppID sql.NullString

		err := rows.Scan(
			&sub.ID, &sub.TenantID, &sub.BundleID, &founderModeID, &sub.ConvertedFromFounderMode,
			&sub.Status, &sub.StripeSubscriptionID, &defaultAppID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
			&sub.CancelAtPeriodEnd, &canceledAt, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bundle subscription: %w", err)
		}

		if founderModeID.Valid {
			id, _ := uuid.Parse(founderModeID.String)
			sub.FounderModeID = &id
		}
		if defaultAppID.Valid {
			id, _ := uuid.Parse(defaultAppID.String)
			sub.DefaultAppID = &id
		}
		if canceledAt.Valid {
			sub.CanceledAt = &canceledAt.Time
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// ListAllActiveFounderModes retrieves all active founder mode registrations across all tenants
// This is used by the background deferred billing checker
func (r *BillingRepository) ListAllActiveFounderModes(ctx context.Context) ([]*FounderModeRegistration, error) {
	query := `SELECT id, tenant_id, bundle_id, mode_type, started_at, ends_at,
		free_days, mrr_threshold_cents, status, converted_to_bundle_id,
		converted_at, stripe_subscription_id, grace_period_started_at,
		grace_period_ends_at, max_users_seen, max_mrr_seen_cents,
		max_api_calls_monthly, created_at, updated_at
		FROM founder_mode_registrations
		WHERE status IN ('active') ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all active founder modes: %w", err)
	}
	defer rows.Close()

	var regs []*FounderModeRegistration
	for rows.Next() {
		reg := &FounderModeRegistration{}
		var convertedToBundleID sql.NullString
		var convertedAt, graceStartedAt, graceEndsAt sql.NullTime

		err := rows.Scan(
			&reg.ID, &reg.TenantID, &reg.BundleID, &reg.ModeType, &reg.StartedAt, &reg.EndsAt,
			&reg.FreeDays, &reg.MRRThresholdCents, &reg.Status, &convertedToBundleID,
			&convertedAt, &reg.StripeSubscriptionID, &graceStartedAt,
			&graceEndsAt, &reg.MaxUsersSeen, &reg.MaxMRRSeenCents,
			&reg.MaxAPICallsMonthly, &reg.CreatedAt, &reg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan founder mode registration: %w", err)
		}

		if convertedToBundleID.Valid {
			id, _ := uuid.Parse(convertedToBundleID.String)
			reg.ConvertedToBundleID = &id
		}
		if convertedAt.Valid {
			reg.ConvertedAt = &convertedAt.Time
		}
		if graceStartedAt.Valid {
			reg.GracePeriodStartedAt = &graceStartedAt.Time
		}
		if graceEndsAt.Valid {
			reg.GracePeriodEndsAt = &graceEndsAt.Time
		}

		regs = append(regs, reg)
	}

	return regs, nil
}

// StartGracePeriod marks a founder mode registration as entering the grace period
func (r *BillingRepository) StartGracePeriod(ctx context.Context, id uuid.UUID, gracePeriodDays int) error {
	graceStartedAt := time.Now().UTC()
	graceEndsAt := graceStartedAt.AddDate(0, 0, gracePeriodDays)

	query := `UPDATE founder_mode_registrations SET
		status = 'grace_period',
		grace_period_started_at = $1,
		grace_period_ends_at = $2,
		updated_at = NOW()
		WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, graceStartedAt, graceEndsAt, id)
	if err != nil {
		return fmt.Errorf("failed to start grace period: %w", err)
	}
	return nil
}

// GetDeferredBillingConfig retrieves the deferred billing configuration for a bundle
func (r *BillingRepository) GetDeferredBillingConfig(ctx context.Context, bundleID uuid.UUID) (*DeferredBillingConfig, error) {
	query := `SELECT id, bundle_id, is_default, trigger_user_count, trigger_revenue_cents,
		trigger_api_calls, trigger_days_elapsed, grace_period_days,
		warning_email_template, trigger_email_template, conversion_email_template,
		created_at, updated_at
		FROM deferred_billing_configs
		WHERE bundle_id = $1 AND is_default = true`

	config := &DeferredBillingConfig{}
	var userCount, revenueCents, apiCalls, daysElapsed *int

	err := r.db.QueryRowContext(ctx, query, bundleID).Scan(
		&config.ID, &config.BundleID, &config.IsDefault, &userCount, &revenueCents,
		&apiCalls, &daysElapsed, &config.GracePeriodDays, &config.WarningEmailTemplate,
		&config.TriggerEmailTemplate, &config.ConversionEmailTemplate,
		&config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deferred billing config: %w", err)
	}

	config.TriggerUserCount = userCount
	config.TriggerRevenueCents = revenueCents
	config.TriggerAPICalls = apiCalls
	config.TriggerDaysElapsed = daysElapsed

	return config, nil
}

// CountActiveFounderModeRegistrations counts founder mode registrations with active or grace_period status
func (r *BillingRepository) CountActiveFounderModeRegistrations(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM founder_mode_registrations WHERE status IN ('active', 'grace_period')`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active founder mode registrations: %w", err)
	}
	return count, nil
}

// CountRecentSuccessfulDeployments counts deployments from the past week with success status
func (r *BillingRepository) CountRecentSuccessfulDeployments(ctx context.Context) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM deployments
		WHERE status = 'success'
		AND created_at >= NOW() - INTERVAL '7 days'
	`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent deployments: %w", err)
	}
	return count, nil
}
