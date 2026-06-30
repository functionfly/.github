package trustapi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillingRepository handles billing operations for Trust API
type BillingRepository struct {
	db *gorm.DB
}

// NewBillingRepository creates a new billing repository
func NewBillingRepository(db *gorm.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

// ============================================
// Tier Pricing Operations
// ============================================

// UpsertTierPricing creates or updates tier pricing
func (r *BillingRepository) UpsertTierPricing(ctx context.Context, pricing *PartnerTierPricing) error {
	result := r.db.WithContext(ctx).Where("tier = ?", pricing.Tier).Assign(pricing).FirstOrCreate(pricing)
	return result.Error
}

// GetTierPricing retrieves pricing for a specific tier
func (r *BillingRepository) GetTierPricing(ctx context.Context, tier string) (*PartnerTierPricing, error) {
	var pricing PartnerTierPricing
	if err := r.db.WithContext(ctx).Where("tier = ? AND is_active = ?", tier, true).First(&pricing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tier pricing not found: %s", tier)
		}
		return nil, err
	}
	return &pricing, nil
}

// ListTierPricing lists all active tier pricing
func (r *BillingRepository) ListTierPricing(ctx context.Context) ([]PartnerTierPricing, error) {
	var pricing []PartnerTierPricing
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("monthly_price_cents ASC").Find(&pricing).Error; err != nil {
		return nil, err
	}
	return pricing, nil
}

// UpdateTierPricingStripePrice updates the Stripe price ID for a tier
func (r *BillingRepository) UpdateTierPricingStripePrice(ctx context.Context, tier, stripePriceID string) error {
	return r.db.WithContext(ctx).Model(&PartnerTierPricing{}).
		Where("tier = ?", tier).
		Update("stripe_price_id", stripePriceID).Error
}

// ============================================
// Partner Billing Usage Operations
// ============================================

// GetOrCreateBillingUsage gets or creates a billing usage record for a partner
func (r *BillingRepository) GetOrCreateBillingUsage(ctx context.Context, partnerID uuid.UUID) (*PartnerBillingUsage, error) {
	var usage PartnerBillingUsage

	// Try to find existing
	err := r.db.WithContext(ctx).Where("partner_id = ?", partnerID).First(&usage).Error
	if err == nil {
		return &usage, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new record with billing period
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	usage = PartnerBillingUsage{
		PartnerID:          partnerID,
		BillingPeriodStart: periodStart,
		BillingPeriodEnd:   periodEnd,
		RequestsThisPeriod: 0,
		OveragesThisPeriod: 0,
	}

	if err := r.db.WithContext(ctx).Create(&usage).Error; err != nil {
		return nil, err
	}

	return &usage, nil
}

// UpdateBillingUsage updates the billing usage record
func (r *BillingRepository) UpdateBillingUsage(ctx context.Context, usage *PartnerBillingUsage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

// ResetPartnerBillingUsage resets usage for a new billing period
func (r *BillingRepository) ResetPartnerBillingUsage(ctx context.Context, partnerID uuid.UUID) error {
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	return r.db.WithContext(ctx).Model(&PartnerBillingUsage{}).
		Where("partner_id = ?", partnerID).
		Updates(map[string]interface{}{
			"billing_period_start": periodStart,
			"billing_period_end":   periodEnd,
			"requests_this_period": 0,
			"overages_this_period": 0,
			"last_reported_at":     nil,
			"last_reported_usage":  0,
		}).Error
}

// GetBillingUsageByPartner retrieves billing usage for a partner
func (r *BillingRepository) GetBillingUsageByPartner(ctx context.Context, partnerID uuid.UUID) (*PartnerBillingUsage, error) {
	var usage PartnerBillingUsage
	if err := r.db.WithContext(ctx).Where("partner_id = ?", partnerID).First(&usage).Error; err != nil {
		return nil, err
	}
	return &usage, nil
}

// ============================================
// Billing Records Operations
// ============================================

// CreateBillingRecord creates a new billing record
func (r *BillingRepository) CreateBillingRecord(ctx context.Context, record *TrustAPIBillingRecord) error {
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(record).Error
}

// GetBillingRecordByID retrieves a billing record by ID
func (r *BillingRepository) GetBillingRecordByID(ctx context.Context, id uuid.UUID) (*TrustAPIBillingRecord, error) {
	var record TrustAPIBillingRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// ListBillingRecordsByPartner lists billing records for a partner
func (r *BillingRepository) ListBillingRecordsByPartner(ctx context.Context, partnerID uuid.UUID, limit, offset int) ([]TrustAPIBillingRecord, int64, error) {
	var records []TrustAPIBillingRecord
	var total int64

	query := r.db.WithContext(ctx).Where("partner_id = ?", partnerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("period_end DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// UpdateBillingRecordStripeInvoice updates the Stripe invoice ID for a billing record
func (r *BillingRepository) UpdateBillingRecordStripeInvoice(ctx context.Context, recordID uuid.UUID, stripeInvoiceID, status string) error {
	updates := map[string]interface{}{
		"stripe_invoice_id":     stripeInvoiceID,
		"stripe_payment_status": status,
	}
	return r.db.WithContext(ctx).Model(&TrustAPIBillingRecord{}).
		Where("id = ?", recordID).
		Updates(updates).Error
}

// UpdateBillingRecordStatus updates the status of a billing record
func (r *BillingRepository) UpdateBillingRecordStatus(ctx context.Context, recordID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&TrustAPIBillingRecord{}).
		Where("id = ?", recordID).
		Update("status", status).Error
}

// GetBillingRecordsByStripeInvoice retrieves billing records by Stripe invoice ID
func (r *BillingRepository) GetBillingRecordsByStripeInvoice(ctx context.Context, stripeInvoiceID string) ([]TrustAPIBillingRecord, error) {
	var records []TrustAPIBillingRecord
	if err := r.db.WithContext(ctx).Where("stripe_invoice_id = ?", stripeInvoiceID).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// ============================================
// Partner Operations (Billing Related)
// ============================================

// UpdatePartnerBillingStatus updates the billing status of a partner
func (r *BillingRepository) UpdatePartnerBillingStatus(ctx context.Context, partnerID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&TrustAPIPartner{}).
		Where("id = ?", partnerID).
		Update("billing_status", status).Error
}

// UpdatePartnerStripeInfo updates Stripe-related fields for a partner
func (r *BillingRepository) UpdatePartnerStripeInfo(ctx context.Context, partnerID uuid.UUID, customerID, subscriptionID, priceID string) error {
	updates := map[string]interface{}{
		"stripe_customer_id":     customerID,
		"stripe_subscription_id": subscriptionID,
		"stripe_price_id":        priceID,
	}
	return r.db.WithContext(ctx).Model(&TrustAPIPartner{}).
		Where("id = ?", partnerID).
		Updates(updates).Error
}

// ListPartnersWithActiveBilling lists partners with active billing status
func (r *BillingRepository) ListPartnersWithActiveBilling(ctx context.Context) ([]TrustAPIPartner, error) {
	var partners []TrustAPIPartner
	if err := r.db.WithContext(ctx).
		Where("billing_status IN ?", []string{"active", "trial", "founder"}).
		Where("status = ?", "active").
		Find(&partners).Error; err != nil {
		return nil, err
	}
	return partners, nil
}

// ListPartnersInFounderMode lists partners currently in founder mode
func (r *BillingRepository) ListPartnersInFounderMode(ctx context.Context) ([]TrustAPIPartner, error) {
	var partners []TrustAPIPartner
	if err := r.db.WithContext(ctx).
		Where("is_founder_mode = ?", true).
		Where("status = ?", "active").
		Find(&partners).Error; err != nil {
		return nil, err
	}
	return partners, nil
}

// ============================================
// Monthly Billing Cycle
// ============================================

// GenerateMonthlyBilling generates billing records for all active partners
func (r *BillingRepository) GenerateMonthlyBilling(ctx context.Context) ([]TrustAPIBillingRecord, error) {
	partners, err := r.ListPartnersWithActiveBilling(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list partners: %w", err)
	}

	var records []TrustAPIBillingRecord
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	for _, partner := range partners {
		// Get tier pricing
		var pricing PartnerTierPricing
		if err := r.db.WithContext(ctx).Where("tier = ?", partner.Tier).First(&pricing).Error; err != nil {
			continue // Skip if pricing not found
		}

		// Get current usage
		var usage PartnerBillingUsage
		if err := r.db.WithContext(ctx).Where("partner_id = ?", partner.ID).First(&usage).Error; err != nil {
			continue // Skip if no usage record
		}

		// Skip if founder mode
		if partner.IsFounderMode {
			continue
		}

		// Calculate charges
		baseCharge := pricing.MonthlyPriceCents
		var overageCharge int
		var overageRequests int

		if pricing.HasOverageBilling && usage.OveragesThisPeriod > 0 {
			overageRequests = usage.OveragesThisPeriod
			units := (overageRequests + 999) / 1000
			overageCharge = units * pricing.OveragePricePer1000
		}

		record := TrustAPIBillingRecord{
			PartnerID:          partner.ID,
			PeriodStart:        periodStart,
			PeriodEnd:          periodEnd,
			BaseRequests:       usage.RequestsThisPeriod - overageRequests,
			OverageRequests:    overageRequests,
			TotalRequests:      usage.RequestsThisPeriod,
			BaseChargeCents:    baseCharge,
			OverageChargeCents: overageCharge,
			TotalChargeCents:   baseCharge + overageCharge,
			Status:             "draft",
		}

		if err := r.CreateBillingRecord(ctx, &record); err != nil {
			continue // Log but don't fail
		}

		records = append(records, record)

		// Reset usage for new period
		r.ResetPartnerBillingUsage(ctx, partner.ID)
	}

	return records, nil
}

// CleanupOldBillingUsage removes old billing usage records (for privacy)
func (r *BillingRepository) CleanupOldBillingUsage(ctx context.Context, olderThan time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("billing_period_end < ?", olderThan).Delete(&PartnerBillingUsage{})
	return result.RowsAffected, result.Error
}

// ============================================
// Helper Methods
// ============================================

// GetUsageSummaryForPartner gets a summary of usage for a partner
func (r *BillingRepository) GetUsageSummaryForPartner(ctx context.Context, partnerID uuid.UUID) (*PartnerUsageReport, error) {
	var partner TrustAPIPartner
	if err := r.db.WithContext(ctx).Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return nil, err
	}

	// Get tier pricing
	var pricing PartnerTierPricing
	if err := r.db.WithContext(ctx).Where("tier = ?", partner.Tier).First(&pricing).Error; err != nil {
		return nil, err
	}

	// Get current usage
	var usage PartnerBillingUsage
	if err := r.db.WithContext(ctx).Where("partner_id = ?", partnerID).First(&usage).Error; err != nil {
		return nil, err
	}

	// Get endpoint breakdown
	var endpointUsage []EndpointBillingUsage
	r.db.WithContext(ctx).Model(&TrustAPIUsage{}).
		Select("endpoint, COUNT(*) as request_count, 0 as cached_count").
		Where("partner_id = ?", partnerID).
		Where("created_at >= ? AND created_at <= ?", usage.BillingPeriodStart, usage.BillingPeriodEnd).
		Group("endpoint").
		Order("request_count DESC").
		Scan(&endpointUsage)

	// Calculate charges
	baseCharge := float64(pricing.MonthlyPriceCents) / 100.0
	var overageCharge float64
	if pricing.HasOverageBilling && usage.OveragesThisPeriod > 0 {
		units := (usage.OveragesThisPeriod + 999) / 1000
		overageCharge = float64(units*pricing.OveragePricePer1000) / 100.0
	}

	return &PartnerUsageReport{
		PartnerID:        partnerID,
		PeriodStart:      usage.BillingPeriodStart,
		PeriodEnd:        usage.BillingPeriodEnd,
		TotalRequests:    int64(usage.RequestsThisPeriod),
		IncludedRequests: int64(pricing.IncludedRequests),
		OverageRequests:  int64(usage.OveragesThisPeriod),
		BaseChargeUSD:    baseCharge,
		OverageChargeUSD: overageCharge,
		TotalChargeUSD:   baseCharge + overageCharge,
		EndpointUsage:    endpointUsage,
	}, nil
}

// UpdatePartner updates a partner record
func (r *BillingRepository) UpdatePartner(partner *TrustAPIPartner) error {
	return r.db.Save(partner).Error
}

// GetPartnerByID retrieves a partner by ID
func (r *BillingRepository) GetPartnerByID(id uuid.UUID) (*TrustAPIPartner, error) {
	var partner TrustAPIPartner
	if err := r.db.Where("id = ?", id).First(&partner).Error; err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetPartnerByStripeCustomerID retrieves a partner by Stripe customer ID
func (r *BillingRepository) GetPartnerByStripeCustomerID(ctx context.Context, customerID string) (*TrustAPIPartner, error) {
	var partner TrustAPIPartner
	if err := r.db.WithContext(ctx).Where("stripe_customer_id = ?", customerID).First(&partner).Error; err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetPartnerByContactEmail retrieves a partner by contact email (for JWT user resolution)
func (r *BillingRepository) GetPartnerByContactEmail(ctx context.Context, email string) (*TrustAPIPartner, error) {
	var partner TrustAPIPartner
	if err := r.db.WithContext(ctx).Where("contact_email = ?", email).First(&partner).Error; err != nil {
		return nil, err
	}
	return &partner, nil
}

// ListPartnersWithFilters lists partners with optional filters
func (r *BillingRepository) ListPartnersWithFilters(status, tier string, limit, offset int) ([]TrustAPIPartner, int64, error) {
	var partners []TrustAPIPartner
	var total int64

	query := r.db.Model(&TrustAPIPartner{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if tier != "" {
		query = query.Where("tier = ?", tier)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&partners).Error; err != nil {
		return nil, 0, err
	}

	return partners, total, nil
}

// Helper function for JSON marshaling
func mustMarshalJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return data
}
