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

func (db *PostgresDB) ListAllSubscriptions(limit, offset int) ([]*Subscription, error) {
	return db.billingRepository.ListAllSubscriptions(limit, offset)
}

func (db *PostgresDB) CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error) {
	return db.billingRepository.CreatePricingTier(ctx, tier)
}

func (db *PostgresDB) ListPricingTiers() ([]*PricingTier, error) {
	return db.billingRepository.ListPricingTiers()
}

func (db *PostgresDB) GetPricingTierByID(id uuid.UUID) (*PricingTier, error) {
	return db.billingRepository.GetPricingTierByID(id)
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

func (db *PostgresDB) GetSubscriptionByTenantID(tenantID uuid.UUID) (*Subscription, error) {
	return db.billingRepository.GetSubscriptionByTenantID(tenantID)
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

func (db *PostgresDB) ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*Invoice, error) {
	return db.billingRepository.ListInvoicesByTenant(tenantID, limit, offset)
}

func (db *PostgresDB) CountInvoicesByTenant(tenantID uuid.UUID) (int, error) {
	return db.billingRepository.CountInvoicesByTenant(tenantID)
}

func (db *PostgresDB) ListAllInvoices(limit, offset int) ([]*Invoice, error) {
	return db.billingRepository.ListAllInvoices(limit, offset)
}

func (db *PostgresDB) GetInvoiceByID(id uuid.UUID) (*Invoice, error) {
	return db.billingRepository.GetInvoiceByID(id)
}

func (db *PostgresDB) UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Invoice, error) {
	return db.billingRepository.UpdateInvoice(ctx, id, updates)
}

func (db *PostgresDB) RecordUsageEvent(ctx context.Context, event *UsageEvent) error {
	return db.billingRepository.RecordUsageEvent(ctx, event)
}

func (db *PostgresDB) GetUsageByTenant(tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error) {
	return db.billingRepository.GetUsageByTenant(tenantID, eventType, start, end)
}

func (db *PostgresDB) CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error) {
	return db.billingRepository.CreateCoupon(ctx, coupon)
}

func (db *PostgresDB) ListCoupons() ([]*Coupon, error) {
	return db.billingRepository.ListCoupons()
}

func (db *PostgresDB) GetCouponByCode(code string) (*Coupon, error) {
	return db.billingRepository.GetCouponByCode(code)
}

func (db *PostgresDB) RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error) {
	return db.billingRepository.RedeemCoupon(ctx, couponID, tenantID, subscriptionID)
}

// Revenue System Phase 1 - Trust Layer Monetization

// Verification Fees
func (db *PostgresDB) GetVerificationFeeByLevel(level string) (*VerificationFee, error) {
	return db.revenueRepository.GetVerificationFeeByLevel(level)
}

func (db *PostgresDB) ListVerificationFees() ([]*VerificationFee, error) {
	return db.revenueRepository.ListVerificationFees()
}

// Function Verification Payments
func (db *PostgresDB) CreateFunctionVerificationPayment(ctx context.Context, payment *FunctionVerificationPayment) error {
	return db.revenueRepository.CreateFunctionVerificationPayment(ctx, payment)
}

func (db *PostgresDB) GetFunctionVerificationPaymentByID(id uuid.UUID) (*FunctionVerificationPayment, error) {
	return db.revenueRepository.GetFunctionVerificationPaymentByID(id)
}

func (db *PostgresDB) UpdateFunctionVerificationPaymentStatus(ctx context.Context, id uuid.UUID, status string, stripePIID *string) error {
	return db.revenueRepository.UpdateFunctionVerificationPaymentStatus(ctx, id, status, stripePIID)
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

func (db *PostgresDB) GetAgentUsageByAgentID(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*AgentUsage, error) {
	return db.revenueRepository.GetAgentUsageByAgentID(ctx, agentID, limit, offset)
}

func (db *PostgresDB) GetAgentUsageSummary(ctx context.Context, agentID uuid.UUID) (totalCalls, billableCalls, overageCalls, estimatedCost int, err error) {
	return db.revenueRepository.GetAgentUsageSummary(ctx, agentID)
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
func (db *PostgresDB) ListPricingTiersExtended() ([]*PricingTierExtended, error) {
	return db.revenueRepository.ListPricingTiersExtended()
}

func (db *PostgresDB) GetPricingTierExtendedByID(id uuid.UUID) (*PricingTierExtended, error) {
	return db.revenueRepository.GetPricingTierExtendedByID(id)
}
