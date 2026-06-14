package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: billing and revenue (subscriptions, invoices, agent usage, platform fees).

// Billing operations
func (db *PostgresDB) CancelSubscription(ctx context.Context, id uuid.UUID) error {
	return db.billingRepository.CancelSubscription(ctx, id)
}

func (db *PostgresDB) ListAllSubscriptions(ctx context.Context, limit, offset int) ([]*Subscription, error) {
	return db.billingRepository.ListAllSubscriptions(ctx, limit, offset)
}

func (db *PostgresDB) CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error) {
	return db.billingRepository.CreatePricingTier(ctx, tier)
}

func (db *PostgresDB) ListPricingTiers(ctx context.Context) ([]*PricingTier, error) {
	return db.billingRepository.ListPricingTiers(ctx)
}

func (db *PostgresDB) GetPricingTierByID(ctx context.Context, id uuid.UUID) (*PricingTier, error) {
	return db.billingRepository.GetPricingTierByID(ctx, id)
}

func (db *PostgresDB) UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*PricingTier, error) {
	return db.billingRepository.UpdatePricingTier(ctx, id, updates)
}

func (db *PostgresDB) DeletePricingTier(ctx context.Context, id uuid.UUID) error {
	return db.billingRepository.DeletePricingTier(ctx, id)
}

func (db *PostgresDB) CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	return db.billingRepository.CreateSubscription(ctx, sub)
}

func (db *PostgresDB) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	return db.billingRepository.GetSubscriptionByID(ctx, id)
}

func (db *PostgresDB) GetSubscriptionByTenantID(ctx context.Context, tenantID uuid.UUID) (*Subscription, error) {
	return db.billingRepository.GetSubscriptionByTenantID(ctx, tenantID)
}

func (db *PostgresDB) GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error) {
	return db.billingRepository.GetSubscriptionByStripeID(ctx, stripeSubscriptionID)
}

func (db *PostgresDB) UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Subscription, error) {
	return db.billingRepository.UpdateSubscription(ctx, id, updates)
}

func (db *PostgresDB) CreateInvoice(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	return db.billingRepository.CreateInvoice(ctx, invoice)
}

func (db *PostgresDB) CreatePaidInvoiceForStripeCheckoutSession(ctx context.Context, tenantID uuid.UUID, amountCents int, currency, checkoutSessionID, receiptURL string) error {
	return db.billingRepository.CreatePaidInvoiceForStripeCheckoutSession(ctx, tenantID, amountCents, currency, checkoutSessionID, receiptURL)
}

func (db *PostgresDB) ListInvoicesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Invoice, error) {
	return db.billingRepository.ListInvoicesByTenant(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) CountInvoicesByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return db.billingRepository.CountInvoicesByTenant(ctx, tenantID)
}

func (db *PostgresDB) ListAllInvoices(ctx context.Context, limit, offset int) ([]*Invoice, error) {
	return db.billingRepository.ListAllInvoices(ctx, limit, offset)
}

func (db *PostgresDB) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	return db.billingRepository.GetInvoiceByID(ctx, id)
}

func (db *PostgresDB) UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Invoice, error) {
	return db.billingRepository.UpdateInvoice(ctx, id, updates)
}

func (db *PostgresDB) RecordUsageEvent(ctx context.Context, event *UsageEvent) error {
	return db.billingRepository.RecordUsageEvent(ctx, event)
}

func (db *PostgresDB) GetUsageByTenant(ctx context.Context, tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error) {
	return db.billingRepository.GetUsageByTenant(ctx, tenantID, eventType, start, end)
}

func (db *PostgresDB) GetUsageByTenantByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*FunctionUsageRollup, error) {
	return db.billingRepository.GetUsageByTenantByFunction(ctx, tenantID, start, end)
}

func (db *PostgresDB) CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	return db.billingRepository.CreateCoupon(ctx, coupon)
}

func (db *PostgresDB) ListCoupons(ctx context.Context) ([]*Coupon, error) {
	return db.billingRepository.ListCoupons(ctx)
}

func (db *PostgresDB) GetCouponByCode(ctx context.Context, code string) (*Coupon, error) {
	return db.billingRepository.GetCouponByCode(ctx, code)
}

func (db *PostgresDB) RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error) {
	return db.billingRepository.RedeemCoupon(ctx, couponID, tenantID, subscriptionID)
}

// Affiliate / Referral Commission System
func (db *PostgresDB) CreateAffiliateCode(ctx context.Context, code *AffiliateCode) (*AffiliateCode, error) {
	return db.billingRepository.CreateAffiliateCode(ctx, code)
}

func (db *PostgresDB) GetAffiliateCodeByID(ctx context.Context, id uuid.UUID) (*AffiliateCode, error) {
	return db.billingRepository.GetAffiliateCodeByID(ctx, id)
}

func (db *PostgresDB) GetAffiliateCodeByCode(ctx context.Context, code string) (*AffiliateCode, error) {
	return db.billingRepository.GetAffiliateCodeByCode(ctx, code)
}

func (db *PostgresDB) ListAffiliateCodes(ctx context.Context) ([]*AffiliateCode, error) {
	return db.billingRepository.ListAffiliateCodes(ctx)
}

func (db *PostgresDB) ListAffiliateCodesByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*AffiliateCode, error) {
	return db.billingRepository.ListAffiliateCodesByPublisher(ctx, publisherID)
}

func (db *PostgresDB) UpdateAffiliateCode(ctx context.Context, code *AffiliateCode) error {
	return db.billingRepository.UpdateAffiliateCode(ctx, code)
}

func (db *PostgresDB) CreateAffiliateReferral(ctx context.Context, referral *AffiliateReferral) (*AffiliateReferral, error) {
	return db.billingRepository.CreateAffiliateReferral(ctx, referral)
}

func (db *PostgresDB) GetAffiliateReferralByID(ctx context.Context, id uuid.UUID) (*AffiliateReferral, error) {
	return db.billingRepository.GetAffiliateReferralByID(ctx, id)
}

func (db *PostgresDB) GetAffiliateReferralByTenant(ctx context.Context, tenantID uuid.UUID) (*AffiliateReferral, error) {
	return db.billingRepository.GetAffiliateReferralByTenant(ctx, tenantID)
}

func (db *PostgresDB) ListAffiliateReferralsByCode(ctx context.Context, codeID uuid.UUID) ([]*AffiliateReferral, error) {
	return db.billingRepository.ListAffiliateReferralsByCode(ctx, codeID)
}

func (db *PostgresDB) UpdateAffiliateReferralStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.billingRepository.UpdateAffiliateReferralStatus(ctx, id, status)
}

func (db *PostgresDB) CreateAffiliateCommission(ctx context.Context, commission *AffiliateCommission) (*AffiliateCommission, error) {
	return db.billingRepository.CreateAffiliateCommission(ctx, commission)
}

func (db *PostgresDB) GetAffiliateCommissionByID(ctx context.Context, id uuid.UUID) (*AffiliateCommission, error) {
	return db.billingRepository.GetAffiliateCommissionByID(ctx, id)
}

func (db *PostgresDB) ListAffiliateCommissionsByCode(ctx context.Context, codeID uuid.UUID) ([]*AffiliateCommission, error) {
	return db.billingRepository.ListAffiliateCommissionsByCode(ctx, codeID)
}

func (db *PostgresDB) ListAffiliateCommissionsByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*AffiliateCommission, error) {
	return db.billingRepository.ListAffiliateCommissionsByPublisher(ctx, publisherID)
}

func (db *PostgresDB) UpdateAffiliateCommissionStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.billingRepository.UpdateAffiliateCommissionStatus(ctx, id, status)
}

func (db *PostgresDB) CalculateCommission(ctx context.Context, commissionType string, commissionValue, baseAmountUSD float64) (commissionCents int64, commissionUSD float64) {
	return db.billingRepository.CalculateCommission(ctx, commissionType, commissionValue, baseAmountUSD)
}

// Revenue System Phase 1 - Trust Layer Monetization

// Verification Fees
func (db *PostgresDB) GetVerificationFeeByLevel(ctx context.Context, level string) (*VerificationFee, error) {
	return db.revenueRepository.GetVerificationFeeByLevel(ctx, level)
}

func (db *PostgresDB) ListVerificationFees(ctx context.Context) ([]*VerificationFee, error) {
	return db.revenueRepository.ListVerificationFees(ctx)
}

// Function Verification Payments
func (db *PostgresDB) CreateFunctionVerificationPayment(ctx context.Context, payment *FunctionVerificationPayment) error {
	return db.revenueRepository.CreateFunctionVerificationPayment(ctx, payment)
}

func (db *PostgresDB) GetFunctionVerificationPaymentByID(ctx context.Context, id uuid.UUID) (*FunctionVerificationPayment, error) {
	return db.revenueRepository.GetFunctionVerificationPaymentByID(ctx, id)
}

func (db *PostgresDB) GetFunctionVerificationPaymentByCheckoutSessionID(ctx context.Context, sessionID string) (*FunctionVerificationPayment, error) {
	return db.revenueRepository.GetFunctionVerificationPaymentByCheckoutSessionID(ctx, sessionID)
}

func (db *PostgresDB) UpdateFunctionVerificationPaymentStatus(ctx context.Context, id uuid.UUID, status string, stripePIID, stripeCheckoutSessionID *string) error {
	return db.revenueRepository.UpdateFunctionVerificationPaymentStatus(ctx, id, status, stripePIID, stripeCheckoutSessionID)
}

func (db *PostgresDB) UpdateFunctionVerificationPaymentJobID(ctx context.Context, id uuid.UUID, jobID uuid.UUID) error {
	return db.revenueRepository.UpdateFunctionVerificationPaymentJobID(ctx, id, jobID)
}

func (db *PostgresDB) GetFunctionVerificationPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*FunctionVerificationPayment, error) {
	return db.revenueRepository.GetFunctionVerificationPaymentsByTenant(ctx, tenantID, limit, offset)
}

// Publisher Earnings
func (db *PostgresDB) CreatePublisherEarning(ctx context.Context, earning *PublisherEarning) error {
	return db.revenueRepository.CreatePublisherEarning(ctx, earning)
}

func (db *PostgresDB) GetPublisherEarningsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*PublisherEarning, error) {
	return db.revenueRepository.GetPublisherEarningsByTenant(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) GetPublisherEarningsSummary(ctx context.Context, tenantID uuid.UUID) (pending, available, withdrawn int, err error) {
	return db.revenueRepository.GetPublisherEarningsSummary(ctx, tenantID)
}

func (db *PostgresDB) GetPublisherEarningsByPeriod(ctx context.Context, tenantID uuid.UUID, year int) ([]*PublisherEarning, error) {
	return db.revenueRepository.GetPublisherEarningsByPeriod(ctx, tenantID, year)
}

// Agent Subscriptions
func (db *PostgresDB) CreateAgentSubscription(ctx context.Context, sub *AgentSubscription) error {
	return db.revenueRepository.CreateAgentSubscription(ctx, sub)
}

func (db *PostgresDB) GetAgentSubscriptionByAgentID(ctx context.Context, agentID uuid.UUID) (*AgentSubscription, error) {
	return db.revenueRepository.GetAgentSubscriptionByAgentID(ctx, agentID)
}

func (db *PostgresDB) GetAgentSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*AgentSubscription, error) {
	return db.revenueRepository.GetAgentSubscriptionsByTenant(ctx, tenantID)
}

func (db *PostgresDB) UpdateAgentSubscriptionStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.revenueRepository.UpdateAgentSubscriptionStatus(ctx, id, status)
}

// Agent Usage
func (db *PostgresDB) CreateAgentUsage(ctx context.Context, usage *AgentUsage) error {
	return db.revenueRepository.CreateAgentUsage(ctx, usage)
}

func (db *PostgresDB) GetAgentUsageByAgentID(ctx context.Context, agentID, tenantID uuid.UUID, limit, offset int) ([]*AgentUsage, error) {
	return db.revenueRepository.GetAgentUsageByAgentID(ctx, agentID, tenantID, limit, offset)
}

func (db *PostgresDB) GetAgentUsageSummary(ctx context.Context, agentID, tenantID uuid.UUID) (totalCalls, billableCalls, overageCalls, estimatedCost int, err error) {
	return db.revenueRepository.GetAgentUsageSummary(ctx, agentID, tenantID)
}

// Platform Fees
func (db *PostgresDB) CreatePlatformFee(ctx context.Context, fee *PlatformFee) error {
	return db.revenueRepository.CreatePlatformFee(ctx, fee)
}

func (db *PostgresDB) GetPlatformFeesByPeriod(ctx context.Context, year, month int) ([]*PlatformFee, error) {
	return db.revenueRepository.GetPlatformFeesByPeriod(ctx, year, month)
}

func (db *PostgresDB) GetPlatformFeesSummary(ctx context.Context) (totalCollected, totalRefunded, totalPaidOut int, err error) {
	return db.revenueRepository.GetPlatformFeesSummary(ctx)
}

// Pricing Tiers Extended
func (db *PostgresDB) ListPricingTiersExtended(ctx context.Context) ([]*PricingTierExtended, error) {
	return db.revenueRepository.ListPricingTiersExtended(ctx)
}

func (db *PostgresDB) GetPricingTierExtendedByID(ctx context.Context, id uuid.UUID) (*PricingTierExtended, error) {
	return db.revenueRepository.GetPricingTierExtendedByID(ctx, id)
}

// AggregateExecutionsForBilling aggregates function executions for billing over a time period
func (db *PostgresDB) AggregateExecutionsForBilling(ctx context.Context, start, end time.Time) ([]*AggregatedBillingUsage, error) {
	return db.billingRepository.AggregateExecutionsForBilling(ctx, start, end)
}

// CreateOrUpdateUsageRollup creates or updates a usage rollup for a tenant/event_type/date
func (db *PostgresDB) CreateOrUpdateUsageRollup(ctx context.Context, rollup *UsageRollup) error {
	return db.billingRepository.CreateOrUpdateUsageRollup(ctx, rollup)
}

// GetInvoiceByPeriod retrieves an invoice by tenant and period
func (db *PostgresDB) GetInvoiceByPeriod(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*Invoice, error) {
	return db.billingRepository.GetInvoiceByPeriod(ctx, tenantID, periodStart, periodEnd)
}

// GetLastAggregationTimestamp gets the last timestamp that was aggregated
func (db *PostgresDB) GetLastAggregationTimestamp(ctx context.Context) (time.Time, error) {
	return db.billingRepository.GetLastAggregationTimestamp(ctx)
}

// SetLastAggregationTimestamp updates the last aggregation timestamp
func (db *PostgresDB) SetLastAggregationTimestamp(ctx context.Context, timestamp time.Time) error {
	return db.billingRepository.SetLastAggregationTimestamp(ctx, timestamp)
}

// GetLastRollupDate gets the last date that was rolled up
func (db *PostgresDB) GetLastRollupDate(ctx context.Context) (time.Time, error) {
	return db.billingRepository.GetLastRollupDate(ctx)
}

// SetLastRollupDate updates the last rollup date
func (db *PostgresDB) SetLastRollupDate(ctx context.Context, date time.Time) error {
	return db.billingRepository.SetLastRollupDate(ctx, date)
}

// Usage Forecasting and Alerting operations - delegated to UsageAlertRepository
func (db *PostgresDB) CreateUsageAlert(ctx context.Context, alert *UsageAlert) error {
	return db.usageAlertRepository.CreateUsageAlert(ctx, alert)
}

func (db *PostgresDB) GetUsageAlertByID(ctx context.Context, id uuid.UUID) (*UsageAlert, error) {
	return db.usageAlertRepository.GetUsageAlertByID(ctx, id)
}

func (db *PostgresDB) ListUsageAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*UsageAlert, error) {
	return db.usageAlertRepository.ListUsageAlertsByTenant(ctx, tenantID)
}

func (db *PostgresDB) UpdateUsageAlert(ctx context.Context, alert *UsageAlert) error {
	return db.usageAlertRepository.UpdateUsageAlert(ctx, alert)
}

func (db *PostgresDB) DeleteUsageAlert(ctx context.Context, id uuid.UUID) error {
	return db.usageAlertRepository.DeleteUsageAlert(ctx, id)
}

func (db *PostgresDB) RecordAlertTrigger(ctx context.Context, history *UsageAlertHistory) error {
	return db.usageAlertRepository.RecordAlertTrigger(ctx, history)
}

func (db *PostgresDB) GetAlertHistoryByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*UsageAlertHistory, error) {
	return db.usageAlertRepository.GetAlertHistoryByTenant(ctx, tenantID, limit)
}

func (db *PostgresDB) CreateOrUpdateSpendCap(ctx context.Context, cap *SpendCap) error {
	return db.usageAlertRepository.CreateOrUpdateSpendCap(ctx, cap)
}

func (db *PostgresDB) GetSpendCapByTenant(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*SpendCap, error) {
	return db.usageAlertRepository.GetSpendCapByTenant(ctx, tenantID, periodStart)
}

func (db *PostgresDB) UpdateCurrentSpend(ctx context.Context, capID uuid.UUID, spendCents int) error {
	return db.usageAlertRepository.UpdateCurrentSpend(ctx, capID, spendCents)
}

func (db *PostgresDB) SaveUsageForecast(ctx context.Context, forecast *UsageForecast) error {
	return db.usageAlertRepository.SaveUsageForecast(ctx, forecast)
}

func (db *PostgresDB) GetLatestForecast(ctx context.Context, tenantID uuid.UUID, forecastType string) (*UsageForecast, error) {
	return db.usageAlertRepository.GetLatestForecast(ctx, tenantID, forecastType)
}

func (db *PostgresDB) GetDailyUsageHistory(ctx context.Context, tenantID uuid.UUID, eventType string, days int) ([]*DailyUsagePoint, error) {
	return db.usageAlertRepository.GetDailyUsageHistory(ctx, tenantID, eventType, days)
}

func (db *PostgresDB) GetDailySpendHistory(ctx context.Context, tenantID uuid.UUID, days int) ([]*DailyUsagePoint, error) {
	return db.usageAlertRepository.GetDailySpendHistory(ctx, tenantID, days)
}

func (db *PostgresDB) GetCurrentPeriodUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*UsageSummary, error) {
	return db.usageAlertRepository.GetCurrentPeriodUsage(ctx, tenantID, periodStart, periodEnd)
}

// Cost Allocation Operations - delegated to BillingRepository
func (db *PostgresDB) RecordCostAllocationEntry(ctx context.Context, entry *CostAllocationEntry) error {
	return db.billingRepository.RecordCostAllocationEntry(ctx, entry)
}

func (db *PostgresDB) GetCostAllocationByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*CostAllocationSummary, error) {
	return db.billingRepository.GetCostAllocationByFunction(ctx, tenantID, start, end)
}

func (db *PostgresDB) GetCostAllocationDailyBreakdown(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*DailyCostBreakdown, error) {
	return db.billingRepository.GetCostAllocationDailyBreakdown(ctx, tenantID, start, end)
}

func (db *PostgresDB) GetTenantCostSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*TenantCostSummary, error) {
	return db.billingRepository.GetTenantCostSummary(ctx, tenantID, start, end)
}

func (db *PostgresDB) GetAllTenantsCostSummary(ctx context.Context, start, end time.Time) ([]*TenantCostSummary, error) {
	return db.billingRepository.GetAllTenantsCostSummary(ctx, start, end)
}

func (db *PostgresDB) GetCostAllocationEntries(ctx context.Context, filter *CostAllocationFilter, limit, offset int) ([]*CostAllocationEntry, int, error) {
	return db.billingRepository.GetCostAllocationEntries(ctx, filter, limit, offset)
}

func (db *PostgresDB) GetCostAllocationReport(ctx context.Context, start, end time.Time) (*CostAllocationReport, error) {
	return db.billingRepository.GetCostAllocationReport(ctx, start, end)
}

func (db *PostgresDB) GetCostAllocationByRegion(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (map[string]*CostAllocationSummary, error) {
	return db.billingRepository.GetCostAllocationByRegion(ctx, tenantID, start, end)
}

func (db *PostgresDB) DeleteOldCostAllocationEntries(ctx context.Context, before time.Time) (int64, error) {
	return db.billingRepository.DeleteOldCostAllocationEntries(ctx, before)
}

// Backend-in-a-Box Pricing Bundles - delegated to BillingRepository
func (db *PostgresDB) CreatePricingBundle(ctx context.Context, bundle *PricingBundle) (*PricingBundle, error) {
	return db.billingRepository.CreatePricingBundle(ctx, bundle)
}

// Stripe Two-Way Sync Operations - delegated to BillingRepository
func (db *PostgresDB) CreateStripeSyncEvent(ctx context.Context, event *StripeSyncEvent) (*StripeSyncEvent, error) {
	return db.billingRepository.CreateStripeSyncEvent(ctx, event)
}

func (db *PostgresDB) GetStripeSyncEventByEventID(ctx context.Context, stripeEventID string) (*StripeSyncEvent, error) {
	return db.billingRepository.GetStripeSyncEventByEventID(ctx, stripeEventID)
}

func (db *PostgresDB) UpdateStripeSyncEventStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	return db.billingRepository.UpdateStripeSyncEventStatus(ctx, id, status, errorMsg)
}

func (db *PostgresDB) IncrementStripeSyncEventRetryCount(ctx context.Context, id uuid.UUID) error {
	return db.billingRepository.IncrementStripeSyncEventRetryCount(ctx, id)
}

func (db *PostgresDB) ListPendingStripeSyncEvents(ctx context.Context, limit int) ([]*StripeSyncEvent, error) {
	return db.billingRepository.ListPendingStripeSyncEvents(ctx, limit)
}

func (db *PostgresDB) UpdateSubscriptionFromStripe(ctx context.Context, stripeSubscriptionID string, stripeData map[string]interface{}) (*Subscription, error) {
	return db.billingRepository.UpdateSubscriptionFromStripe(ctx, stripeSubscriptionID, stripeData)
}

func (db *PostgresDB) UpdateTenantPaymentMethod(ctx context.Context, tenantID uuid.UUID, paymentMethod *PaymentMethodInfoExtended) error {
	return db.billingRepository.UpdateTenantPaymentMethod(ctx, tenantID, paymentMethod)
}

func (db *PostgresDB) GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*PaymentMethodInfoExtended, error) {
	return db.billingRepository.GetPaymentMethodByStripeID(ctx, stripePaymentMethodID)
}

func (db *PostgresDB) ListPricingBundles(ctx context.Context, activeOnly bool) ([]*PricingBundle, error) {
	return db.billingRepository.ListPricingBundles(ctx, activeOnly)
}

func (db *PostgresDB) GetPricingBundleBySlug(ctx context.Context, slug string) (*PricingBundle, error) {
	return db.billingRepository.GetPricingBundleBySlug(ctx, slug)
}

func (db *PostgresDB) GetPricingBundleByID(ctx context.Context, id uuid.UUID) (*PricingBundle, error) {
	return db.billingRepository.GetPricingBundleByID(ctx, id)
}

func (db *PostgresDB) GetPricingBundleByStripePriceID(ctx context.Context, stripePriceID string) (*PricingBundle, error) {
	return db.billingRepository.GetPricingBundleByStripePriceID(ctx, stripePriceID)
}

func (db *PostgresDB) UpdatePricingBundleStripePrice(ctx context.Context, slug, stripePriceID string) error {
	return db.billingRepository.UpdatePricingBundleStripePrice(ctx, slug, stripePriceID)
}

// Founder Mode (viral pricing - deferred billing) - delegated to BillingRepository
func (db *PostgresDB) CreateFounderModeRegistration(ctx context.Context, reg *FounderModeRegistration) error {
	return db.billingRepository.CreateFounderModeRegistration(ctx, reg)
}

func (db *PostgresDB) GetActiveFounderMode(ctx context.Context, tenantID, bundleID uuid.UUID) (*FounderModeRegistration, error) {
	return db.billingRepository.GetActiveFounderMode(ctx, tenantID, bundleID)
}

func (db *PostgresDB) ListFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FounderModeRegistration, error) {
	return db.billingRepository.ListFounderModesByTenant(ctx, tenantID)
}

func (db *PostgresDB) ListActiveFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FounderModeRegistration, error) {
	return db.billingRepository.ListActiveFounderModesByTenant(ctx, tenantID)
}

func (db *PostgresDB) UpdateFounderModeStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.billingRepository.UpdateFounderModeStatus(ctx, id, status)
}

func (db *PostgresDB) UpdateFounderModeProgress(ctx context.Context, id uuid.UUID, users, mrrCents, apiCalls int) error {
	return db.billingRepository.UpdateFounderModeProgress(ctx, id, users, mrrCents, apiCalls)
}

func (db *PostgresDB) ListAllActiveFounderModes(ctx context.Context) ([]*FounderModeRegistration, error) {
	return db.billingRepository.ListAllActiveFounderModes(ctx)
}

func (db *PostgresDB) StartGracePeriod(ctx context.Context, id uuid.UUID, gracePeriodDays int) error {
	return db.billingRepository.StartGracePeriod(ctx, id, gracePeriodDays)
}

func (db *PostgresDB) GetDeferredBillingConfig(ctx context.Context, bundleID uuid.UUID) (*DeferredBillingConfig, error) {
	return db.billingRepository.GetDeferredBillingConfig(ctx, bundleID)
}

// Bundle Subscriptions - delegated to BillingRepository
func (db *PostgresDB) CreateBundleSubscription(ctx context.Context, sub *BundleSubscription) error {
	return db.billingRepository.CreateBundleSubscription(ctx, sub)
}

func (db *PostgresDB) GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*BundleSubscription, error) {
	return db.billingRepository.GetBundleSubscriptionByTenant(ctx, tenantID)
}

func (db *PostgresDB) GetBundleSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*BundleSubscription, error) {
	return db.billingRepository.GetBundleSubscriptionByStripeID(ctx, stripeSubID)
}

func (db *PostgresDB) ListBundleSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*BundleSubscription, error) {
	return db.billingRepository.ListBundleSubscriptionsByTenant(ctx, tenantID)
}

func (db *PostgresDB) UpdateBundleSubscription(ctx context.Context, sub *BundleSubscription) error {
	return db.billingRepository.UpdateBundleSubscription(ctx, sub)
}

// ==================== Analytics Repository Delegation ====================

// CalculateMRR delegates to analytics repository
func (db *PostgresDB) CalculateMRR(ctx context.Context, year, month int) (*MRRMetrics, error) {
	return db.analyticsRepository.CalculateMRR(ctx, year, month)
}

// CalculateARR delegates to analytics repository
func (db *PostgresDB) CalculateARR(ctx context.Context, year, month int) (*ARRMetrics, error) {
	return db.analyticsRepository.CalculateARR(ctx, year, month)
}

// GetMRRTimeseries delegates to analytics repository
func (db *PostgresDB) GetMRRTimeseries(ctx context.Context, months int) ([]MRRMetrics, error) {
	return db.analyticsRepository.GetMRRTimeseries(ctx, months)
}

// CalculateChurnMetrics delegates to analytics repository
func (db *PostgresDB) CalculateChurnMetrics(ctx context.Context, year, month int) (*ChurnMetrics, error) {
	return db.analyticsRepository.CalculateChurnMetrics(ctx, year, month)
}

// RecordSubscriptionChurnEvent delegates to analytics repository
func (db *PostgresDB) RecordSubscriptionChurnEvent(ctx context.Context, event *SubscriptionChurnEvent) error {
	return db.analyticsRepository.RecordSubscriptionChurnEvent(ctx, event)
}

// GetChurnMetricsTimeseries delegates to analytics repository
func (db *PostgresDB) GetChurnMetricsTimeseries(ctx context.Context, months int) ([]ChurnMetrics, error) {
	return db.analyticsRepository.GetChurnMetricsTimeseries(ctx, months)
}

// GenerateFinancialReport delegates to analytics repository
func (db *PostgresDB) GenerateFinancialReport(ctx context.Context, reportType string, periodStart, periodEnd time.Time) (*FinancialReport, error) {
	return db.analyticsRepository.GenerateFinancialReport(ctx, reportType, periodStart, periodEnd)
}

// GetTaxJurisdictionReport delegates to analytics repository
func (db *PostgresDB) GetTaxJurisdictionReport(ctx context.Context, periodMonth string) ([]TaxJurisdictionReport, error) {
	return db.analyticsRepository.GetTaxJurisdictionReport(ctx, periodMonth)
}

// GetLTVMetrics delegates to analytics repository
func (db *PostgresDB) GetLTVMetrics(ctx context.Context) (*LTVMetrics, error) {
	return db.analyticsRepository.GetLTVMetrics(ctx)
}

// =============================================================================
// Database-Driven Agent Tier Pricing (replaces hardcoded constants)
// =============================================================================

func (db *PostgresDB) GetAgentTierPricingBySlug(ctx context.Context, slug string) (*AgentTierPricing, error) {
	return db.revenueRepository.GetAgentTierPricingBySlug(ctx, slug)
}

func (db *PostgresDB) ListAgentTierPricing(ctx context.Context, activeOnly bool) ([]*AgentTierPricing, error) {
	return db.revenueRepository.ListAgentTierPricing(ctx, activeOnly)
}

func (db *PostgresDB) GetAgentTierPricingForRegion(ctx context.Context, slug string, currencyCode string) (*AgentTierPricing, error) {
	return db.revenueRepository.GetAgentTierPricingForRegion(ctx, slug, currencyCode)
}

// =============================================================================
// Multi-Currency Support
// =============================================================================

func (db *PostgresDB) SaveCurrencyExchangeRate(ctx context.Context, rate *CurrencyExchangeRate) error {
	return db.revenueRepository.SaveCurrencyExchangeRate(ctx, rate)
}

func (db *PostgresDB) GetCurrencyExchangeRate(ctx context.Context, baseCurrency, quoteCurrency string, date *time.Time) (*CurrencyExchangeRate, error) {
	return db.revenueRepository.GetCurrencyExchangeRate(ctx, baseCurrency, quoteCurrency, date)
}

func (db *PostgresDB) ConvertCurrency(ctx context.Context, amountCents int, fromCurrency, toCurrency string) (int, error) {
	return db.revenueRepository.ConvertCurrency(ctx, amountCents, fromCurrency, toCurrency)
}

func (db *PostgresDB) GetSupportedCurrency(ctx context.Context, code string) (*SupportedCurrency, error) {
	return db.revenueRepository.GetSupportedCurrency(ctx, code)
}

func (db *PostgresDB) ListSupportedCurrencies(ctx context.Context) ([]*SupportedCurrency, error) {
	return db.revenueRepository.ListSupportedCurrencies(ctx)
}

// =============================================================================
// Credit Note operations (for refund accounting / SOX compliance)
// =============================================================================

func (db *PostgresDB) CreateCreditNote(ctx context.Context, creditNote *CreditNote) (*CreditNote, error) {
	return db.creditNoteRepository.Create(ctx, creditNote)
}

func (db *PostgresDB) GetCreditNoteByID(ctx context.Context, id uuid.UUID) (*CreditNote, error) {
	return db.creditNoteRepository.GetByID(ctx, id)
}

func (db *PostgresDB) GetCreditNoteByReferenceNumber(ctx context.Context, refNumber string) (*CreditNote, error) {
	return db.creditNoteRepository.GetByReferenceNumber(ctx, refNumber)
}

func (db *PostgresDB) ListCreditNotes(ctx context.Context, filter *CreditNoteFilter, limit, offset int) ([]*CreditNote, int64, error) {
	return db.creditNoteRepository.List(ctx, filter, limit, offset)
}

func (db *PostgresDB) ListCreditNotesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*CreditNote, int64, error) {
	return db.creditNoteRepository.ListByTenant(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) ListCreditNotesByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*CreditNote, error) {
	return db.creditNoteRepository.ListByInvoice(ctx, invoiceID)
}

func (db *PostgresDB) UpdateCreditNote(ctx context.Context, creditNote *CreditNote) error {
	return db.creditNoteRepository.Update(ctx, creditNote)
}

func (db *PostgresDB) UpdateCreditNoteStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.creditNoteRepository.UpdateStatus(ctx, id, status)
}

func (db *PostgresDB) VoidCreditNote(ctx context.Context, id uuid.UUID) error {
	return db.creditNoteRepository.Void(ctx, id)
}

func (db *PostgresDB) ApplyCreditNote(ctx context.Context, id uuid.UUID) error {
	return db.creditNoteRepository.Apply(ctx, id)
}

func (db *PostgresDB) GetCreditNoteWithRelations(ctx context.Context, id uuid.UUID) (*CreditNote, error) {
	return db.creditNoteRepository.GetWithRelations(ctx, id)
}

func (db *PostgresDB) GetCreditNoteStats(ctx context.Context, tenantID *uuid.UUID) (*CreditNoteStats, error) {
	return db.creditNoteRepository.GetCreditNoteStats(ctx, tenantID)
}

// Credit Note Line Item operations
func (db *PostgresDB) CreateCreditNoteLineItem(ctx context.Context, item *CreditNoteLineItem) error {
	return db.creditNoteRepository.CreateLineItem(ctx, item)
}

func (db *PostgresDB) GetCreditNoteLineItems(ctx context.Context, creditNoteID uuid.UUID) ([]*CreditNoteLineItem, error) {
	return db.creditNoteRepository.GetLineItems(ctx, creditNoteID)
}

func (db *PostgresDB) DeleteCreditNoteLineItem(ctx context.Context, id uuid.UUID) error {
	return db.creditNoteRepository.DeleteLineItem(ctx, id)
}

func (db *PostgresDB) DeleteCreditNoteLineItems(ctx context.Context, creditNoteID uuid.UUID) error {
	return db.creditNoteRepository.DeleteLineItems(ctx, creditNoteID)
}

// Tenant Stripe Config (Isolated Payment Processing)
func (db *PostgresDB) GetTenantStripeConfig(ctx context.Context, tenantID uuid.UUID) (*TenantStripeConfig, error) {
	return db.tenantStripeConfigRepository.GetByTenantID(ctx, tenantID)
}

func (db *PostgresDB) CreateTenantStripeConfig(ctx context.Context, config *TenantStripeConfig) error {
	return db.tenantStripeConfigRepository.Create(ctx, config)
}

func (db *PostgresDB) UpdateTenantStripeConfig(ctx context.Context, config *TenantStripeConfig) error {
	return db.tenantStripeConfigRepository.Update(ctx, config)
}
