// Package domain provides domain-specific repository interfaces.
// This splits the monolithic Repository interface into focused, composable interfaces
// following the Interface Segregation Principle.
package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// UserRepository handles user and authentication-related operations
type UserRepository interface {
	IsUsernameReserved(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, email, passwordHash string, tenantID uuid.UUID) (*types.User, error)
	CreateUserWithSocialAuth(ctx context.Context, email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*types.User, error)
	CreateUserWithRole(ctx context.Context, user *types.User) (*types.User, error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*types.User, error)
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	GetUserForPublicProfile(ctx context.Context, login string) (*types.User, error)
	SearchUsersByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]types.UserSearchHit, error)
	GetUserByVerificationToken(ctx context.Context, token string) (*types.User, error)
	GetUserBySocialProvider(ctx context.Context, provider, providerID string) (*types.User, error)
	ListUsers(ctx context.Context) ([]*types.User, error)
	ListUserIDsByTenant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error)
	ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.User, error)
	CountActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*types.User, error)
	UpdateUserEmailVerification(ctx context.Context, userID uuid.UUID, verified bool, token *string, expiresAt *time.Time) error
	UpdateUserProviderData(ctx context.Context, userID uuid.UUID, providerData map[string]interface{}) error
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error
	GetUserSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID) error
	ReactivateUser(ctx context.Context, userID uuid.UUID) error
	IncrementUserTokenVersion(ctx context.Context, userID uuid.UUID) (int, error)

	// MFA operations
	UpdateUserMFA(ctx context.Context, userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error
	UpdateUserMFAEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error
	UpdateUserMFABackupCodes(ctx context.Context, userID uuid.UUID, backupCodes []string) error
	UpdateUserMFALastUsed(ctx context.Context, userID uuid.UUID, lastUsed *time.Time) error
	VerifyPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error)

	// Username change operations (2-per-year limit with early-change fee)
	CreateUsernameChangeHistory(ctx context.Context, history *types.UsernameChangeHistory) error
	ChangeUsernameWithHistory(ctx context.Context, userID uuid.UUID, newUsername string, history *types.UsernameChangeHistory) error
	GetUsernameChangeHistory(ctx context.Context, userID uuid.UUID) ([]*types.UsernameChangeHistory, error)
	CountUsernameChangesInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (int, error)
	GetLastUsernameChange(ctx context.Context, userID uuid.UUID) (*types.UsernameChangeHistory, error)
	HasUsernameChangedInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (bool, error)

	// Pending username change operations (for payment flow)
	CreatePendingUsernameChange(ctx context.Context, pending *types.PendingUsernameChange) error
	GetPendingUsernameChangeByID(ctx context.Context, id uuid.UUID) (*types.PendingUsernameChange, error)
	GetPendingUsernameChangeByCheckoutSession(ctx context.Context, sessionID string) (*types.PendingUsernameChange, error)
	UpdatePendingUsernameChangeStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteExpiredPendingUsernameChanges(ctx context.Context) (int64, error)
	ListPendingUsernameChangesForUser(ctx context.Context, userID uuid.UUID) ([]*types.PendingUsernameChange, error)
}

// SessionRepository handles session and token operations
type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*types.Session, error)
	GetSessionByToken(ctx context.Context, sessionToken string) (*types.Session, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*types.Session, error)
	UpdateSessionMFAStatus(ctx context.Context, sessionToken string, mfaVerified bool) error
	UpdateSessionActivity(ctx context.Context, sessionToken string) error
	DeleteSession(ctx context.Context, sessionToken string) error
	DeleteSessionByID(ctx context.Context, sessionID, userID uuid.UUID) error
	DeleteSessionByIDOnly(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
	ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*types.Session, error)
	CountActiveUserSessions(ctx context.Context, userID uuid.UUID) (int, error)
	ListTenantSessions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*types.Session, error)

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*types.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*types.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
	ListUserRefreshTokens(ctx context.Context, userID uuid.UUID) ([]*types.RefreshToken, error)

	// Login attempt tracking
	CreateLoginAttempt(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*types.LoginAttempt, error)
	GetRecentFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	GetUserLockoutStatus(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	ClearUserLockout(ctx context.Context, userID uuid.UUID) error
	DeleteOldLoginAttempts(ctx context.Context, before time.Time) (int64, error)

	// Auth events
	LogAuthEvent(ctx context.Context, event *types.AuthEvent) error
	GetAuthEventsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*types.AuthEvent, error)
	GetAuthEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*types.AuthEvent, error)
	GetRecentAuthEvents(ctx context.Context, since time.Time, limit int) ([]*types.AuthEvent, error)
	DeleteOldAuthEvents(ctx context.Context, before time.Time) (int64, error)
}

// TenantRepository handles tenant/organization operations
type TenantRepository interface {
	CreateTenant(ctx context.Context, name string) (*types.Tenant, error)
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*types.Tenant, error)
	GetTenantByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*types.Tenant, error)
	ListTenants(ctx context.Context) ([]*types.Tenant, error)
	ListTenantsWithStripeCustomerID(ctx context.Context) ([]*types.Tenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*types.Tenant, error)
	UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	CountUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountRoutingEventsForTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)

	// Tenant membership (multi-tenant support)
	IsUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	AddTenantMember(ctx context.Context, userID, tenantID, invitedBy uuid.UUID, role string) error
	AcceptTenantMembership(ctx context.Context, userID, tenantID uuid.UUID) error
	RemoveTenantMember(ctx context.Context, userID, tenantID uuid.UUID) error
}

// TeamRepository handles team membership and permissions
type TeamRepository interface {
	CreateTeam(ctx context.Context, team *types.Team) error
	GetTeamByID(ctx context.Context, teamID uuid.UUID) (*types.Team, error)
	GetTeamsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*types.Team, error)
	UpdateTeam(ctx context.Context, team *types.Team) error
	DeleteTeam(ctx context.Context, teamID uuid.UUID) error
	AddTeamMember(ctx context.Context, membership *types.TeamMembership) error
	UpdateTeamMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
	RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error
	GetTeamMembership(ctx context.Context, teamID, userID uuid.UUID) (*types.TeamMembership, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*types.Team, error)
	GrantTeamPermission(ctx context.Context, permission *types.TeamPermission) error
	RevokeTeamPermission(ctx context.Context, teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error
	GetTeamPermissions(ctx context.Context, teamID uuid.UUID) ([]*types.TeamPermission, error)
	GetResourcePermissions(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*types.TeamPermission, error)
	CheckUserResourcePermission(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID, resourceType string) ([]string, error)
	IsUserTeamOwner(ctx context.Context, userID, teamID uuid.UUID) (bool, error)
	IsUserTeamAdmin(ctx context.Context, userID, teamID uuid.UUID) (bool, error)
}

// BillingRepository handles billing, subscriptions and payments
type BillingRepository interface {
	// Pricing tiers
	CreatePricingTier(ctx context.Context, tier *types.PricingTier) (*types.PricingTier, error)
	ListPricingTiers(ctx context.Context) ([]*types.PricingTier, error)
	GetPricingTierByID(ctx context.Context, id uuid.UUID) (*types.PricingTier, error)
	UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.PricingTier, error)
	DeletePricingTier(ctx context.Context, id uuid.UUID) error

	// Subscriptions
	CreateSubscription(ctx context.Context, sub *types.Subscription) (*types.Subscription, error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*types.Subscription, error)
	GetSubscriptionByTenantID(ctx context.Context, tenantID uuid.UUID) (*types.Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*types.Subscription, error)
	ListAllSubscriptions(ctx context.Context, limit, offset int) ([]*types.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.Subscription, error)
	CancelSubscription(ctx context.Context, id uuid.UUID) error

	// Invoices
	CreateInvoice(ctx context.Context, invoice *types.Invoice) (*types.Invoice, error)
	CreatePaidInvoiceForStripeCheckoutSession(ctx context.Context, tenantID uuid.UUID, amountCents int, currency, checkoutSessionID, receiptURL string) error
	ListInvoicesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*types.Invoice, error)
	CountInvoicesByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	ListAllInvoices(ctx context.Context, limit, offset int) ([]*types.Invoice, error)
	GetInvoiceByID(ctx context.Context, id uuid.UUID) (*types.Invoice, error)
	GetInvoiceByPeriod(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*types.Invoice, error)
	UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.Invoice, error)

	// Usage and billing events
	RecordUsageEvent(ctx context.Context, event *types.UsageEvent) error
	GetUsageByTenant(ctx context.Context, tenantID uuid.UUID, eventType string, start, end time.Time) ([]*types.UsageRollup, error)
	GetUsageByTenantByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*types.FunctionUsageRollup, error)
	AggregateExecutionsForBilling(ctx context.Context, start, end time.Time) ([]*types.AggregatedBillingUsage, error)
	CreateOrUpdateUsageRollup(ctx context.Context, rollup *types.UsageRollup) error
	GetLastAggregationTimestamp(ctx context.Context) (time.Time, error)
	SetLastAggregationTimestamp(ctx context.Context, timestamp time.Time) error
	GetLastRollupDate(ctx context.Context) (time.Time, error)
	SetLastRollupDate(ctx context.Context, date time.Time) error

	// Coupons
	CreateCoupon(ctx context.Context, coupon *types.Coupon) (*types.Coupon, error)
	ListCoupons(ctx context.Context) ([]*types.Coupon, error)
	GetCouponByCode(ctx context.Context, code string) (*types.Coupon, error)
	RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*types.CouponRedemption, error)

	// Extended tiers
	ListPricingTiersExtended(ctx context.Context) ([]*types.PricingTierExtended, error)
	GetPricingTierExtendedByID(ctx context.Context, id uuid.UUID) (*types.PricingTierExtended, error)

	// Stripe Two-Way Sync
	CreateStripeSyncEvent(ctx context.Context, event *types.StripeSyncEvent) (*types.StripeSyncEvent, error)
	GetStripeSyncEventByEventID(ctx context.Context, stripeEventID string) (*types.StripeSyncEvent, error)
	UpdateStripeSyncEventStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
	IncrementStripeSyncEventRetryCount(ctx context.Context, id uuid.UUID) error
	ListPendingStripeSyncEvents(ctx context.Context, limit int) ([]*types.StripeSyncEvent, error)
	UpdateSubscriptionFromStripe(ctx context.Context, stripeSubscriptionID string, stripeData map[string]interface{}) (*types.Subscription, error)
	GetTenantByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*types.Tenant, error)
	UpdateTenantPaymentMethod(ctx context.Context, tenantID uuid.UUID, paymentMethod *types.PaymentMethodInfoExtended) error
	GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*types.PaymentMethodInfoExtended, error)
}

// RevenueRepository handles revenue-related operations (trust layer monetization)
type RevenueRepository interface {
	// Verification fees
	GetVerificationFeeByLevel(ctx context.Context, level string) (*types.VerificationFee, error)
	ListVerificationFees(ctx context.Context) ([]*types.VerificationFee, error)

	// Function verification payments
	CreateFunctionVerificationPayment(ctx context.Context, payment *types.FunctionVerificationPayment) error
	GetFunctionVerificationPaymentByID(ctx context.Context, id uuid.UUID) (*types.FunctionVerificationPayment, error)
	GetFunctionVerificationPaymentByCheckoutSessionID(ctx context.Context, sessionID string) (*types.FunctionVerificationPayment, error)
	UpdateFunctionVerificationPaymentStatus(ctx context.Context, id uuid.UUID, status string, stripePIID, stripeCheckoutSessionID *string) error
	UpdateFunctionVerificationPaymentJobID(ctx context.Context, id uuid.UUID, jobID uuid.UUID) error
	GetFunctionVerificationPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*types.FunctionVerificationPayment, error)

	// Publisher earnings
	CreatePublisherEarning(ctx context.Context, earning *types.PublisherEarning) error
	GetPublisherEarningsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*types.PublisherEarning, error)
	GetPublisherEarningsSummary(ctx context.Context, tenantID uuid.UUID) (pending, available, withdrawn int, err error)
	GetPublisherEarningsByPeriod(ctx context.Context, tenantID uuid.UUID, year int) ([]*types.PublisherEarning, error)

	// Agent subscriptions and usage
	CreateAgentSubscription(ctx context.Context, sub *types.AgentSubscription) error
	GetAgentSubscriptionByAgentID(ctx context.Context, agentID uuid.UUID) (*types.AgentSubscription, error)
	GetAgentSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.AgentSubscription, error)
	UpdateAgentSubscriptionStatus(ctx context.Context, id uuid.UUID, status string) error

	CreateAgentUsage(ctx context.Context, usage *types.AgentUsage) error
	GetAgentUsageByAgentID(ctx context.Context, agentID, tenantID uuid.UUID, limit, offset int) ([]*types.AgentUsage, error)
	GetAgentUsageSummary(ctx context.Context, agentID, tenantID uuid.UUID) (totalCalls, billableCalls, overageCalls, estimatedCost int, err error)

	// Platform fees
	CreatePlatformFee(ctx context.Context, fee *types.PlatformFee) error
	GetPlatformFeesByPeriod(ctx context.Context, year, month int) ([]*types.PlatformFee, error)
	GetPlatformFeesSummary(ctx context.Context) (totalCollected, totalRefunded, totalPaidOut int, err error)

	// Database-Driven Agent Tier Pricing (replaces hardcoded constants)
	GetAgentTierPricingBySlug(ctx context.Context, slug string) (*types.AgentTierPricing, error)
	ListAgentTierPricing(ctx context.Context, activeOnly bool) ([]*types.AgentTierPricing, error)
	GetAgentTierPricingForRegion(ctx context.Context, slug string, currencyCode string) (*types.AgentTierPricing, error)

	// Multi-Currency Support
	SaveCurrencyExchangeRate(ctx context.Context, rate *types.CurrencyExchangeRate) error
	GetCurrencyExchangeRate(ctx context.Context, baseCurrency, quoteCurrency string, date *time.Time) (*types.CurrencyExchangeRate, error)
	ConvertCurrency(ctx context.Context, amountCents int, fromCurrency, toCurrency string) (int, error)
	GetSupportedCurrency(ctx context.Context, code string) (*types.SupportedCurrency, error)
	ListSupportedCurrencies(ctx context.Context) ([]*types.SupportedCurrency, error)
}

// OAuthRepository handles OAuth state and invite codes
type OAuthRepository interface {
	StoreOAuthState(ctx context.Context, state string, expiresAt time.Time, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string) error
	ValidateAndConsumeOAuthState(ctx context.Context, state string) (valid bool, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string, err error)
	DeleteExpiredOAuthStates(ctx context.Context) (int64, error)

	// Signup invite codes
	CreateSignupInvite(ctx context.Context, label string, maxUses *int, expiresAt *time.Time, createdBy *uuid.UUID) (id uuid.UUID, plainCode string, err error)
	ListSignupInvitesAdmin(ctx context.Context) ([]types.SignupInviteCodeAdminList, error)
	RevokeSignupInvite(ctx context.Context, id uuid.UUID) error
	ValidateSignupInviteReadOnly(ctx context.Context, plainCode string) error
	ReserveSignupInvite(ctx context.Context, plainCode string) (inviteID uuid.UUID, err error)
	ReleaseSignupInviteReservation(ctx context.Context, inviteID uuid.UUID) error
}
