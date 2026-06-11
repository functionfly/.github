// Package domain provides domain-specific repository interfaces.
// This splits the monolithic Repository interface into focused, composable interfaces
// following the Interface Segregation Principle.
package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// UserRepository handles user and authentication-related operations
type UserRepository interface {
	IsUsernameReserved(username string) (bool, error)
	CreateUser(email, passwordHash string, tenantID uuid.UUID) (*storage.User, error)
	CreateUserWithSocialAuth(email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*storage.User, error)
	CreateUserWithRole(ctx context.Context, user *storage.User) (*storage.User, error)
	GetUserByEmail(email string) (*storage.User, error)
	GetUserByID(userID uuid.UUID) (*storage.User, error)
	GetUserByUsername(username string) (*storage.User, error)
	GetUserForPublicProfile(login string) (*storage.User, error)
	SearchUsersByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]storage.UserSearchHit, error)
	GetUserByVerificationToken(token string) (*storage.User, error)
	GetUserBySocialProvider(provider, providerID string) (*storage.User, error)
	ListUsers() ([]*storage.User, error)
	ListUserIDsByTenant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error)
	ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.User, error)
	CountActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*storage.User, error)
	UpdateUserEmailVerification(ctx context.Context, userID uuid.UUID, verified bool, token *string, expiresAt *time.Time) error
	UpdateUserProviderData(userID uuid.UUID, providerData map[string]interface{}) error
	UpdateUserSettings(userID uuid.UUID, settings map[string]interface{}) error
	GetUserSettings(userID uuid.UUID) (map[string]interface{}, error)
	DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID) error
	ReactivateUser(ctx context.Context, userID uuid.UUID) error

	// MFA operations
	UpdateUserMFA(userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error
	UpdateUserMFAEnabled(userID uuid.UUID, enabled bool) error
	UpdateUserMFABackupCodes(userID uuid.UUID, backupCodes []string) error
	UpdateUserMFALastUsed(userID uuid.UUID, lastUsed *time.Time) error
	VerifyPassword(userID uuid.UUID, password string) (bool, error)

	// Username change operations (2-per-year limit with early-change fee)
	CreateUsernameChangeHistory(ctx context.Context, history *storage.UsernameChangeHistory) error
	GetUsernameChangeHistory(ctx context.Context, userID uuid.UUID) ([]*storage.UsernameChangeHistory, error)
	CountUsernameChangesInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (int, error)
	GetLastUsernameChange(ctx context.Context, userID uuid.UUID) (*storage.UsernameChangeHistory, error)
	HasUsernameChangedInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (bool, error)

	// Pending username change operations (for payment flow)
	CreatePendingUsernameChange(ctx context.Context, pending *storage.PendingUsernameChange) error
	GetPendingUsernameChangeByID(ctx context.Context, id uuid.UUID) (*storage.PendingUsernameChange, error)
	GetPendingUsernameChangeByCheckoutSession(ctx context.Context, sessionID string) (*storage.PendingUsernameChange, error)
	UpdatePendingUsernameChangeStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteExpiredPendingUsernameChanges(ctx context.Context) (int64, error)
	ListPendingUsernameChangesForUser(ctx context.Context, userID uuid.UUID) ([]*storage.PendingUsernameChange, error)
}

// SessionRepository handles session and token operations
type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*storage.Session, error)
	GetSessionByToken(ctx context.Context, sessionToken string) (*storage.Session, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*storage.Session, error)
	UpdateSessionMFAStatus(ctx context.Context, sessionToken string, mfaVerified bool) error
	UpdateSessionActivity(ctx context.Context, sessionToken string) error
	DeleteSession(ctx context.Context, sessionToken string) error
	DeleteSessionByID(ctx context.Context, sessionID, userID uuid.UUID) error
	DeleteSessionByIDOnly(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
	ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*storage.Session, error)
	CountActiveUserSessions(ctx context.Context, userID uuid.UUID) (int, error)
	ListTenantSessions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*storage.Session, error)

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*storage.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*storage.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
	ListUserRefreshTokens(ctx context.Context, userID uuid.UUID) ([]*storage.RefreshToken, error)

	// Login attempt tracking
	CreateLoginAttempt(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*storage.LoginAttempt, error)
	GetRecentFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	GetUserLockoutStatus(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	ClearUserLockout(ctx context.Context, userID uuid.UUID) error
	DeleteOldLoginAttempts(ctx context.Context, before time.Time) (int64, error)

	// Auth events
	LogAuthEvent(ctx context.Context, event *storage.AuthEvent) error
	GetAuthEventsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*storage.AuthEvent, error)
	GetAuthEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*storage.AuthEvent, error)
	GetRecentAuthEvents(ctx context.Context, since time.Time, limit int) ([]*storage.AuthEvent, error)
	DeleteOldAuthEvents(ctx context.Context, before time.Time) (int64, error)
}

// TenantRepository handles tenant/organization operations
type TenantRepository interface {
	CreateTenant(ctx context.Context, name string) (*storage.Tenant, error)
	GetTenantByID(tenantID uuid.UUID) (*storage.Tenant, error)
	GetTenantByStripeCustomerID(stripeCustomerID string) (*storage.Tenant, error)
	ListTenants() ([]*storage.Tenant, error)
	ListTenantsWithStripeCustomerID() ([]*storage.Tenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*storage.Tenant, error)
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	CountUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error)
}

// TeamRepository handles team membership and permissions
type TeamRepository interface {
	CreateTeam(team *storage.Team) error
	GetTeamByID(teamID uuid.UUID) (*storage.Team, error)
	GetTeamsByTenantID(tenantID uuid.UUID) ([]*storage.Team, error)
	UpdateTeam(team *storage.Team) error
	DeleteTeam(teamID uuid.UUID) error
	AddTeamMember(membership *storage.TeamMembership) error
	UpdateTeamMember(teamID, userID uuid.UUID, role string) error
	RemoveTeamMember(teamID, userID uuid.UUID) error
	GetTeamMembership(teamID, userID uuid.UUID) (*storage.TeamMembership, error)
	GetUserTeams(userID uuid.UUID) ([]*storage.Team, error)
	GrantTeamPermission(permission *storage.TeamPermission) error
	RevokeTeamPermission(teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error
	GetTeamPermissions(teamID uuid.UUID) ([]*storage.TeamPermission, error)
	GetResourcePermissions(resourceType string, resourceID uuid.UUID) ([]*storage.TeamPermission, error)
	CheckUserResourcePermission(userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error)
	GetUserPermissions(userID uuid.UUID, resourceType string) ([]string, error)
	IsUserTeamOwner(userID, teamID uuid.UUID) (bool, error)
	IsUserTeamAdmin(userID, teamID uuid.UUID) (bool, error)
}

// BillingRepository handles billing, subscriptions and payments
type BillingRepository interface {
	// Pricing tiers
	CreatePricingTier(ctx context.Context, tier *storage.PricingTier) (*storage.PricingTier, error)
	ListPricingTiers() ([]*storage.PricingTier, error)
	GetPricingTierByID(id uuid.UUID) (*storage.PricingTier, error)
	UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.PricingTier, error)
	DeletePricingTier(ctx context.Context, id uuid.UUID) error

	// Subscriptions
	CreateSubscription(ctx context.Context, sub *storage.Subscription) (*storage.Subscription, error)
	GetSubscriptionByTenantID(tenantID uuid.UUID) (*storage.Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*storage.Subscription, error)
	ListAllSubscriptions(limit, offset int) ([]*storage.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.Subscription, error)
	CancelSubscription(ctx context.Context, id uuid.UUID) error

	// Invoices
	CreateInvoice(ctx context.Context, invoice *storage.Invoice) (*storage.Invoice, error)
	CreatePaidInvoiceForStripeCheckoutSession(ctx context.Context, tenantID uuid.UUID, amountCents int, currency, checkoutSessionID, receiptURL string) error
	ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*storage.Invoice, error)
	CountInvoicesByTenant(tenantID uuid.UUID) (int, error)
	ListAllInvoices(limit, offset int) ([]*storage.Invoice, error)
	GetInvoiceByID(id uuid.UUID) (*storage.Invoice, error)
	GetInvoiceByPeriod(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*storage.Invoice, error)
	UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.Invoice, error)

	// Usage and billing events
	RecordUsageEvent(ctx context.Context, event *storage.UsageEvent) error
	GetUsageByTenant(tenantID uuid.UUID, eventType string, start, end time.Time) ([]*storage.UsageRollup, error)
	AggregateExecutionsForBilling(ctx context.Context, start, end time.Time) ([]*storage.AggregatedBillingUsage, error)
	CreateOrUpdateUsageRollup(ctx context.Context, rollup *storage.UsageRollup) error
	GetLastAggregationTimestamp(ctx context.Context) (time.Time, error)
	SetLastAggregationTimestamp(ctx context.Context, timestamp time.Time) error
	GetLastRollupDate(ctx context.Context) (time.Time, error)
	SetLastRollupDate(ctx context.Context, date time.Time) error

	// Coupons
	CreateCoupon(ctx context.Context, coupon *storage.Coupon) (*storage.Coupon, error)
	ListCoupons() ([]*storage.Coupon, error)
	GetCouponByCode(code string) (*storage.Coupon, error)
	RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*storage.CouponRedemption, error)

	// Extended tiers
	ListPricingTiersExtended() ([]*storage.PricingTierExtended, error)
	GetPricingTierExtendedByID(id uuid.UUID) (*storage.PricingTierExtended, error)

	// Stripe Two-Way Sync
	CreateStripeSyncEvent(ctx context.Context, event *storage.StripeSyncEvent) (*storage.StripeSyncEvent, error)
	GetStripeSyncEventByEventID(ctx context.Context, stripeEventID string) (*storage.StripeSyncEvent, error)
	UpdateStripeSyncEventStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
	IncrementStripeSyncEventRetryCount(ctx context.Context, id uuid.UUID) error
	ListPendingStripeSyncEvents(ctx context.Context, limit int) ([]*storage.StripeSyncEvent, error)
	UpdateSubscriptionFromStripe(ctx context.Context, stripeSubscriptionID string, stripeData map[string]interface{}) (*storage.Subscription, error)
	GetTenantByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*storage.Tenant, error)
	UpdateTenantPaymentMethod(ctx context.Context, tenantID uuid.UUID, paymentMethod *storage.PaymentMethodInfoExtended) error
	GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*storage.PaymentMethodInfoExtended, error)
}

// RevenueRepository handles revenue-related operations (trust layer monetization)
type RevenueRepository interface {
	// Verification fees
	GetVerificationFeeByLevel(level string) (*storage.VerificationFee, error)
	ListVerificationFees() ([]*storage.VerificationFee, error)

	// Function verification payments
	CreateFunctionVerificationPayment(ctx context.Context, payment *storage.FunctionVerificationPayment) error
	GetFunctionVerificationPaymentByID(id uuid.UUID) (*storage.FunctionVerificationPayment, error)
	GetFunctionVerificationPaymentByCheckoutSessionID(ctx context.Context, sessionID string) (*storage.FunctionVerificationPayment, error)
	UpdateFunctionVerificationPaymentStatus(ctx context.Context, id uuid.UUID, status string, stripePIID, stripeCheckoutSessionID *string) error
	UpdateFunctionVerificationPaymentJobID(ctx context.Context, id uuid.UUID, jobID uuid.UUID) error
	GetFunctionVerificationPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*storage.FunctionVerificationPayment, error)

	// Publisher earnings
	CreatePublisherEarning(ctx context.Context, earning *storage.PublisherEarning) error
	GetPublisherEarningsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*storage.PublisherEarning, error)
	GetPublisherEarningsSummary(ctx context.Context, tenantID uuid.UUID) (pending, available, withdrawn int, err error)
	GetPublisherEarningsByPeriod(ctx context.Context, tenantID uuid.UUID, year int) ([]*storage.PublisherEarning, error)

	// Agent subscriptions and usage
	CreateAgentSubscription(ctx context.Context, sub *storage.AgentSubscription) error
	GetAgentSubscriptionByAgentID(ctx context.Context, agentID uuid.UUID) (*storage.AgentSubscription, error)
	GetAgentSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.AgentSubscription, error)
	UpdateAgentSubscriptionStatus(ctx context.Context, id uuid.UUID, status string) error

	CreateAgentUsage(ctx context.Context, usage *storage.AgentUsage) error
	GetAgentUsageByAgentID(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*storage.AgentUsage, error)
	GetAgentUsageSummary(ctx context.Context, agentID uuid.UUID) (totalCalls, billableCalls, overageCalls, estimatedCost int, err error)

	// Platform fees
	CreatePlatformFee(ctx context.Context, fee *storage.PlatformFee) error
	GetPlatformFeesByPeriod(ctx context.Context, year, month int) ([]*storage.PlatformFee, error)
	GetPlatformFeesSummary(ctx context.Context) (totalCollected, totalRefunded, totalPaidOut int, err error)

	// Database-Driven Agent Tier Pricing (replaces hardcoded constants)
	GetAgentTierPricingBySlug(ctx context.Context, slug string) (*storage.AgentTierPricing, error)
	ListAgentTierPricing(ctx context.Context, activeOnly bool) ([]*storage.AgentTierPricing, error)
	GetAgentTierPricingForRegion(ctx context.Context, slug string, currencyCode string) (*storage.AgentTierPricing, error)

	// Multi-Currency Support
	SaveCurrencyExchangeRate(ctx context.Context, rate *storage.CurrencyExchangeRate) error
	GetCurrencyExchangeRate(ctx context.Context, baseCurrency, quoteCurrency string, date *time.Time) (*storage.CurrencyExchangeRate, error)
	ConvertCurrency(ctx context.Context, amountCents int, fromCurrency, toCurrency string) (int, error)
	GetSupportedCurrency(ctx context.Context, code string) (*storage.SupportedCurrency, error)
	ListSupportedCurrencies(ctx context.Context) ([]*storage.SupportedCurrency, error)
}

// OAuthRepository handles OAuth state and invite codes
type OAuthRepository interface {
	StoreOAuthState(ctx context.Context, state string, expiresAt time.Time, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string) error
	ValidateAndConsumeOAuthState(ctx context.Context, state string) (valid bool, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string, err error)
	DeleteExpiredOAuthStates() (int64, error)

	// Signup invite codes
	CreateSignupInvite(ctx context.Context, label string, maxUses *int, expiresAt *time.Time, createdBy *uuid.UUID) (id uuid.UUID, plainCode string, err error)
	ListSignupInvitesAdmin(ctx context.Context) ([]storage.SignupInviteCodeAdminList, error)
	RevokeSignupInvite(ctx context.Context, id uuid.UUID) error
	ValidateSignupInviteReadOnly(ctx context.Context, plainCode string) error
	ReserveSignupInvite(ctx context.Context, plainCode string) (inviteID uuid.UUID, err error)
	ReleaseSignupInviteReservation(ctx context.Context, inviteID uuid.UUID) error
}
