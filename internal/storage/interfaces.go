package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PgNotification represents a PostgreSQL notification
type PgNotification struct {
	PID     int    `json:"pid"`
	Channel string `json:"channel"`
	Payload string `json:"payload"`
}

// Repository defines the interface for data access
type Repository interface {
	// User operations
	IsUsernameReserved(username string) (bool, error)
	CreateUser(email, passwordHash string, tenantID uuid.UUID) (*User, error)
	CreateUserWithSocialAuth(email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*User, error)
	CreateUserWithRole(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(userID uuid.UUID) (*User, error)
	GetUserByUsername(username string) (*User, error)
	// GetUserForPublicProfile resolves by username (case-insensitive) or by unique email local-part (before @).
	GetUserForPublicProfile(login string) (*User, error)
	SearchUsersByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]UserSearchHit, error)
	GetUserByVerificationToken(token string) (*User, error)
	GetUserBySocialProvider(provider, providerID string) (*User, error)
	ListUsers() ([]*User, error)
	ListUserIDsByTenant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error)
	ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*User, error)
	CountActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*User, error)
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
	IncrementUserTokenVersion(userID uuid.UUID) error

	// OAuth state (CSRF) — persisted for multi-instance OAuth flows
	// redirectURI is optional; when set, callback redirects there with token (e.g. CLI local server).
	// inviteCode is optional; stored for invite-only signup validation on callback (short TTL).
	// codeVerifier is for PKCE (Proof Key for Code Exchange) - required for public clients.
	// loginHint preserves tenant subdomain or email context through the OAuth flow.
	// deviceFingerprint stores a hash of device characteristics for session binding validation.
	StoreOAuthState(ctx context.Context, state string, expiresAt time.Time, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string) error
	ValidateAndConsumeOAuthState(ctx context.Context, state string) (valid bool, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string, err error)
	DeleteExpiredOAuthStates() (int64, error)

	// Signup invite codes (platform invite-only launch)
	CreateSignupInvite(ctx context.Context, label string, maxUses *int, expiresAt *time.Time, createdBy *uuid.UUID) (id uuid.UUID, plainCode string, err error)
	ListSignupInvitesAdmin(ctx context.Context) ([]SignupInviteCodeAdminList, error)
	RevokeSignupInvite(ctx context.Context, id uuid.UUID) error
	ValidateSignupInviteReadOnly(ctx context.Context, plainCode string) error
	ReserveSignupInvite(ctx context.Context, plainCode string) (inviteID uuid.UUID, err error)
	ReleaseSignupInviteReservation(ctx context.Context, inviteID uuid.UUID) error

	// Tenant operations
	CreateTenant(ctx context.Context, name string) (*Tenant, error)
	GetTenantByID(tenantID uuid.UUID) (*Tenant, error)
	GetTenantByStripeCustomerID(stripeCustomerID string) (*Tenant, error)
	ListTenants() ([]*Tenant, error)
	ListTenantsWithStripeCustomerID() ([]*Tenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error)
	UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error
	UpdateTenantTaxSettings(ctx context.Context, tenantID uuid.UUID, settings *TaxSettings) error
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	CountUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error)

	// Tenant membership operations (multi-tenant support)
	IsUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	AddTenantMember(ctx context.Context, userID, tenantID, invitedBy uuid.UUID, role string) error
	AcceptTenantMembership(ctx context.Context, userID, tenantID uuid.UUID) error
	RemoveTenantMember(ctx context.Context, userID, tenantID uuid.UUID) error

	// Team operations
	CreateTeam(team *Team) error
	GetTeamByID(teamID uuid.UUID) (*Team, error)
	GetTeamsByTenantID(tenantID uuid.UUID) ([]*Team, error)
	UpdateTeam(team *Team) error
	DeleteTeam(teamID uuid.UUID) error
	AddTeamMember(membership *TeamMembership) error
	UpdateTeamMember(teamID, userID uuid.UUID, role string) error
	RemoveTeamMember(teamID, userID uuid.UUID) error
	GetTeamMembership(teamID, userID uuid.UUID) (*TeamMembership, error)
	GetUserTeams(userID uuid.UUID) ([]*Team, error)
	GrantTeamPermission(permission *TeamPermission) error
	RevokeTeamPermission(teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error
	GetTeamPermissions(teamID uuid.UUID) ([]*TeamPermission, error)
	GetResourcePermissions(resourceType string, resourceID uuid.UUID) ([]*TeamPermission, error)
	CheckUserResourcePermission(userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error)
	GetUserPermissions(userID uuid.UUID, resourceType string) ([]string, error)
	IsUserTeamOwner(userID, teamID uuid.UUID) (bool, error)
	IsUserTeamAdmin(userID, teamID uuid.UUID) (bool, error)

	// Audit operations
	ListAuditEvents(limit, offset int) ([]*AuditEvent, error)
	LogAuditEvent(ctx context.Context, event *AuditEvent) error
	ListAuditEventsFiltered(limit, offset int, filters map[string]interface{}) ([]*AuditEvent, error)
	GetAuditEventByID(id uuid.UUID) (*AuditEvent, error)

	// Billing operations
	CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error)
	ListPricingTiers() ([]*PricingTier, error)
	GetPricingTierByID(id uuid.UUID) (*PricingTier, error)
	UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*PricingTier, error)
	DeletePricingTier(ctx context.Context, id uuid.UUID) error

	CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetSubscriptionByTenantID(tenantID uuid.UUID) (*Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error)
	ListAllSubscriptions(limit, offset int) ([]*Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Subscription, error)
	CancelSubscription(ctx context.Context, id uuid.UUID) error

	// Stripe Two-Way Sync (Real-time updates from Stripe dashboard)
	CreateStripeSyncEvent(ctx context.Context, event *StripeSyncEvent) (*StripeSyncEvent, error)
	GetStripeSyncEventByEventID(ctx context.Context, stripeEventID string) (*StripeSyncEvent, error)
	UpdateStripeSyncEventStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
	IncrementStripeSyncEventRetryCount(ctx context.Context, id uuid.UUID) error
	ListPendingStripeSyncEvents(ctx context.Context, limit int) ([]*StripeSyncEvent, error)
	UpdateSubscriptionFromStripe(ctx context.Context, stripeSubscriptionID string, stripeData map[string]interface{}) (*Subscription, error)
	UpdateTenantPaymentMethod(ctx context.Context, tenantID uuid.UUID, paymentMethod *PaymentMethodInfoExtended) error
	GetPaymentMethodByStripeID(ctx context.Context, stripePaymentMethodID string) (*PaymentMethodInfoExtended, error)

	CreateInvoice(ctx context.Context, invoice *Invoice) (*Invoice, error)
	// CreatePaidInvoiceForStripeCheckoutSession records a paid one-time Checkout payment (registry wallet, agent credits, etc.). Idempotent on checkoutSessionID (external_reference).
	CreatePaidInvoiceForStripeCheckoutSession(ctx context.Context, tenantID uuid.UUID, amountCents int, currency, checkoutSessionID, receiptURL string) error
	ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*Invoice, error)
	CountInvoicesByTenant(tenantID uuid.UUID) (int, error)
	ListAllInvoices(limit, offset int) ([]*Invoice, error)
	GetInvoiceByID(id uuid.UUID) (*Invoice, error)
	UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Invoice, error)

	// Credit Note operations (for refund accounting and SOX compliance)
	CreateCreditNote(ctx context.Context, creditNote *CreditNote) (*CreditNote, error)
	GetCreditNoteByID(ctx context.Context, id uuid.UUID) (*CreditNote, error)
	GetCreditNoteByReferenceNumber(ctx context.Context, refNumber string) (*CreditNote, error)
	ListCreditNotes(ctx context.Context, filter *CreditNoteFilter, limit, offset int) ([]*CreditNote, int64, error)
	ListCreditNotesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*CreditNote, int64, error)
	ListCreditNotesByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*CreditNote, error)
	UpdateCreditNote(ctx context.Context, creditNote *CreditNote) error
	UpdateCreditNoteStatus(ctx context.Context, id uuid.UUID, status string) error
	VoidCreditNote(ctx context.Context, id uuid.UUID) error
	ApplyCreditNote(ctx context.Context, id uuid.UUID) error
	GetCreditNoteWithRelations(ctx context.Context, id uuid.UUID) (*CreditNote, error)
	GetCreditNoteStats(ctx context.Context, tenantID *uuid.UUID) (*CreditNoteStats, error)
	// Credit Note Line Item operations
	CreateCreditNoteLineItem(ctx context.Context, item *CreditNoteLineItem) error
	GetCreditNoteLineItems(ctx context.Context, creditNoteID uuid.UUID) ([]*CreditNoteLineItem, error)
	DeleteCreditNoteLineItem(ctx context.Context, id uuid.UUID) error
	DeleteCreditNoteLineItems(ctx context.Context, creditNoteID uuid.UUID) error

	RecordUsageEvent(ctx context.Context, event *UsageEvent) error
	GetUsageByTenant(tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error)
	GetUsageByTenantByFunction(tenantID uuid.UUID, start, end time.Time) ([]*FunctionUsageRollup, error)

	CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error)
	ListCoupons() ([]*Coupon, error)
	GetCouponByCode(code string) (*Coupon, error)
	RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error)

	// Affiliate / Referral Commission System
	CreateAffiliateCode(ctx context.Context, code *AffiliateCode) (*AffiliateCode, error)
	GetAffiliateCodeByID(id uuid.UUID) (*AffiliateCode, error)
	GetAffiliateCodeByCode(code string) (*AffiliateCode, error)
	ListAffiliateCodes() ([]*AffiliateCode, error)
	ListAffiliateCodesByPublisher(publisherID uuid.UUID) ([]*AffiliateCode, error)
	UpdateAffiliateCode(ctx context.Context, code *AffiliateCode) error
	CreateAffiliateReferral(ctx context.Context, referral *AffiliateReferral) (*AffiliateReferral, error)
	GetAffiliateReferralByID(id uuid.UUID) (*AffiliateReferral, error)
	GetAffiliateReferralByTenant(tenantID uuid.UUID) (*AffiliateReferral, error)
	ListAffiliateReferralsByCode(codeID uuid.UUID) ([]*AffiliateReferral, error)
	UpdateAffiliateReferralStatus(ctx context.Context, id uuid.UUID, status string) error
	CreateAffiliateCommission(ctx context.Context, commission *AffiliateCommission) (*AffiliateCommission, error)
	GetAffiliateCommissionByID(id uuid.UUID) (*AffiliateCommission, error)
	ListAffiliateCommissionsByCode(codeID uuid.UUID) ([]*AffiliateCommission, error)
	ListAffiliateCommissionsByPublisher(publisherID uuid.UUID) ([]*AffiliateCommission, error)
	UpdateAffiliateCommissionStatus(ctx context.Context, id uuid.UUID, status string) error
	CalculateCommission(commissionType string, commissionValue, baseAmountUSD float64) (commissionCents int64, commissionUSD float64)

	// Revenue System Phase 1 - Trust Layer Monetization
	// Verification Fees
	GetVerificationFeeByLevel(level string) (*VerificationFee, error)
	ListVerificationFees() ([]*VerificationFee, error)

	// Function Verification Payments
	CreateFunctionVerificationPayment(ctx context.Context, payment *FunctionVerificationPayment) error
	GetFunctionVerificationPaymentByID(id uuid.UUID) (*FunctionVerificationPayment, error)
	GetFunctionVerificationPaymentByCheckoutSessionID(ctx context.Context, sessionID string) (*FunctionVerificationPayment, error)
	UpdateFunctionVerificationPaymentStatus(ctx context.Context, id uuid.UUID, status string, stripePIID, stripeCheckoutSessionID *string) error
	UpdateFunctionVerificationPaymentJobID(ctx context.Context, id uuid.UUID, jobID uuid.UUID) error
	GetFunctionVerificationPaymentsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*FunctionVerificationPayment, error)

	// Publisher Earnings
	CreatePublisherEarning(ctx context.Context, earning *PublisherEarning) error
	GetPublisherEarningsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*PublisherEarning, error)
	GetPublisherEarningsSummary(ctx context.Context, tenantID uuid.UUID) (pending, available, withdrawn int, err error)
	GetPublisherEarningsByPeriod(ctx context.Context, tenantID uuid.UUID, year int) ([]*PublisherEarning, error)

	// Agent Subscriptions
	CreateAgentSubscription(ctx context.Context, sub *AgentSubscription) error
	GetAgentSubscriptionByAgentID(ctx context.Context, agentID uuid.UUID) (*AgentSubscription, error)
	GetAgentSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*AgentSubscription, error)
	UpdateAgentSubscriptionStatus(ctx context.Context, id uuid.UUID, status string) error

	// Agent Usage
	CreateAgentUsage(ctx context.Context, usage *AgentUsage) error
	GetAgentUsageByAgentID(ctx context.Context, agentID, tenantID uuid.UUID, limit, offset int) ([]*AgentUsage, error)
	GetAgentUsageSummary(ctx context.Context, agentID, tenantID uuid.UUID) (totalCalls, billableCalls, overageCalls, estimatedCost int, err error)

	// Platform Fees
	CreatePlatformFee(ctx context.Context, fee *PlatformFee) error
	GetPlatformFeesByPeriod(ctx context.Context, year, month int) ([]*PlatformFee, error)
	GetPlatformFeesSummary(ctx context.Context) (totalCollected, totalRefunded, totalPaidOut int, err error)

	// Pricing Tiers Extended (with Moat fields)
	ListPricingTiersExtended() ([]*PricingTierExtended, error)
	GetPricingTierExtendedByID(id uuid.UUID) (*PricingTierExtended, error)

	// Billing Usage Aggregation (registry execution → usage events → rollups)
	AggregateExecutionsForBilling(ctx context.Context, start, end time.Time) ([]*AggregatedBillingUsage, error)
	CreateOrUpdateUsageRollup(ctx context.Context, rollup *UsageRollup) error
	GetInvoiceByPeriod(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*Invoice, error)
	GetLastAggregationTimestamp(ctx context.Context) (time.Time, error)
	SetLastAggregationTimestamp(ctx context.Context, timestamp time.Time) error
	GetLastRollupDate(ctx context.Context) (time.Time, error)
	SetLastRollupDate(ctx context.Context, date time.Time) error

	// Backend-in-a-Box Pricing Bundles
	CreatePricingBundle(ctx context.Context, bundle *PricingBundle) (*PricingBundle, error)
	ListPricingBundles(ctx context.Context, activeOnly bool) ([]*PricingBundle, error)
	GetPricingBundleBySlug(ctx context.Context, slug string) (*PricingBundle, error)
	GetPricingBundleByID(ctx context.Context, id uuid.UUID) (*PricingBundle, error)
	GetPricingBundleByStripePriceID(ctx context.Context, stripePriceID string) (*PricingBundle, error)
	UpdatePricingBundleStripePrice(ctx context.Context, slug, stripePriceID string) error
	CountActiveFounderModeRegistrations(ctx context.Context) (int, error)
	CountRecentSuccessfulDeployments(ctx context.Context) (int, error)

	// Database-Driven Agent Tier Pricing (replaces hardcoded constants)
	GetAgentTierPricingBySlug(ctx context.Context, slug string) (*AgentTierPricing, error)
	ListAgentTierPricing(ctx context.Context, activeOnly bool) ([]*AgentTierPricing, error)
	GetAgentTierPricingForRegion(ctx context.Context, slug string, currencyCode string) (*AgentTierPricing, error)

	// Multi-Currency Support
	SaveCurrencyExchangeRate(ctx context.Context, rate *CurrencyExchangeRate) error
	GetCurrencyExchangeRate(ctx context.Context, baseCurrency, quoteCurrency string, date *time.Time) (*CurrencyExchangeRate, error)
	ConvertCurrency(ctx context.Context, amountCents int, fromCurrency, toCurrency string) (int, error)
	GetSupportedCurrency(ctx context.Context, code string) (*SupportedCurrency, error)
	ListSupportedCurrencies(ctx context.Context) ([]*SupportedCurrency, error)

	// Founder Mode (viral pricing - deferred billing)
	CreateFounderModeRegistration(ctx context.Context, reg *FounderModeRegistration) error
	GetActiveFounderMode(ctx context.Context, tenantID, bundleID uuid.UUID) (*FounderModeRegistration, error)
	ListFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FounderModeRegistration, error)
	ListActiveFounderModesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FounderModeRegistration, error)
	UpdateFounderModeStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateFounderModeProgress(ctx context.Context, id uuid.UUID, users, mrrCents, apiCalls int) error
	ListAllActiveFounderModes(ctx context.Context) ([]*FounderModeRegistration, error)
	StartGracePeriod(ctx context.Context, id uuid.UUID, gracePeriodDays int) error
	GetDeferredBillingConfig(ctx context.Context, bundleID uuid.UUID) (*DeferredBillingConfig, error)

	// Bundle Subscriptions
	CreateBundleSubscription(ctx context.Context, sub *BundleSubscription) error
	UpdateBundleSubscription(ctx context.Context, sub *BundleSubscription) error
	GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*BundleSubscription, error)
	GetBundleSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*BundleSubscription, error)
	ListBundleSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*BundleSubscription, error)

	// App operations
	CreateApp(name, slug string, tenantID uuid.UUID) (*App, error)
	GetAppByID(id uuid.UUID) (*App, error)
	GetAppBySlug(slug string) (*App, error)
	// GetAppBySlugAndTenant returns an app by slug scoped to the tenant (dashboard / tenant APIs).
	GetAppBySlugAndTenant(slug string, tenantID uuid.UUID) (*App, error)
	ListAppsByTenant(tenantID uuid.UUID) ([]*App, error)

	// Backend operations
	CreateBackend(appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error)
	ListBackendsByAppID(appID uuid.UUID) ([]*Backend, error)
	GetBackendByID(id uuid.UUID) (*Backend, error)
	GetAllEnabledBackends() ([]*Backend, error)
	ListAllBackends(ctx context.Context) ([]*Backend, error)
	UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error

	// Platform feature measures (admin security/features page)
	ListFeatureMeasures(ctx context.Context) ([]*FeatureMeasure, error)
	UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error

	// Health check operations
	InsertHealthCheck(backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error
	GetRecentHealthChecks(backendID uuid.UUID, limit int) ([]*HealthCheck, error)

	// Circuit breaker operations
	GetCircuitState(backendID uuid.UUID) (*CircuitState, error)
	UpdateCircuitState(state *CircuitState) error
	UpsertCircuitState(state *CircuitState) error

	// Routing operations
	InsertRoutingEvent(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error
	GetRecentRoutingEvents(limit int, since time.Time) ([]*RoutingEvent, error)

	// Deployment operations
	CreateDeployment(appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*Deployment, error)
	UpdateDeploymentStatus(id uuid.UUID, status, message string, metadata map[string]interface{}) error
	GetDeploymentByID(id uuid.UUID) (*Deployment, error)
	ListDeploymentsByAppID(appID uuid.UUID, limit int) ([]*Deployment, error)
	GetLatestSuccessfulDeployment(appID uuid.UUID, provider string) (*Deployment, error)

	// Status operations
	GetBackendStatusByAppID(appID uuid.UUID) ([]*BackendStatus, error)

	// Artifact operations
	StoreDeploymentArtifact(appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*DeploymentArtifact, error)
	GetDeploymentArtifact(key string) (*DeploymentArtifact, error)

	// Content management operations
	// Changelog operations
	CreateChangelogEntry(ctx context.Context, entry *ChangelogEntry) (*ChangelogEntry, error)
	GetChangelogEntryByID(id uuid.UUID) (*ChangelogEntry, error)
	GetChangelogEntryByVersion(version string) (*ChangelogEntry, error)
	ListChangelogEntries(limit, offset int, publishedOnly bool) ([]*ChangelogEntry, error)
	UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogEntry, error)
	DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error
	CreateChangelogChange(ctx context.Context, change *ChangelogChange) (*ChangelogChange, error)
	UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogChange, error)
	DeleteChangelogChange(ctx context.Context, id uuid.UUID) error

	// Blog operations
	CreateBlogPost(ctx context.Context, post *BlogPost) (*BlogPost, error)
	GetBlogPostByID(id uuid.UUID) (*BlogPost, error)
	GetBlogPostBySlug(slug string) (*BlogPost, error)
	ListBlogPosts(limit, offset int, publishedOnly bool, tagFilter []string) ([]*BlogPost, error)
	UpdateBlogPost(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogPost, error)
	DeleteBlogPost(ctx context.Context, id uuid.UUID) error

	// Blog categories
	ListBlogCategories(ctx context.Context) ([]*BlogCategory, error)
	CreateBlogCategory(ctx context.Context, c *BlogCategory) (*BlogCategory, error)
	GetBlogCategoryByID(ctx context.Context, id uuid.UUID) (*BlogCategory, error)
	GetBlogCategoryBySlug(ctx context.Context, slug string) (*BlogCategory, error)
	UpdateBlogCategory(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogCategory, error)
	DeleteBlogCategory(ctx context.Context, id uuid.UUID) error

	// Blog authors
	ListBlogAuthors(ctx context.Context) ([]*BlogAuthor, error)
	CreateBlogAuthor(ctx context.Context, a *BlogAuthor) (*BlogAuthor, error)
	GetBlogAuthorByID(ctx context.Context, id uuid.UUID) (*BlogAuthor, error)
	GetBlogAuthorBySlug(ctx context.Context, slug string) (*BlogAuthor, error)
	UpdateBlogAuthor(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogAuthor, error)
	DeleteBlogAuthor(ctx context.Context, id uuid.UUID) error

	// Blog settings
	GetBlogSettings(ctx context.Context) (*BlogSettings, error)
	UpdateBlogSettings(ctx context.Context, updates map[string]interface{}) (*BlogSettings, error)

	// Feedback operations
	CreateFeedback(feedback *Feedback) (*Feedback, error)
	GetFeedbackByID(id uuid.UUID) (*Feedback, error)
	GetFeedbackByUser(userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error)
	ListFeedback(limit, offset int, statusFilter *string, typeFilter *string) ([]Feedback, error)
	UpdateFeedbackStatus(id uuid.UUID, status string) error
	CreateFeedbackAttachment(attachment *FeedbackAttachment) (*FeedbackAttachment, error)
	GetFeedbackAttachments(feedbackID uuid.UUID) ([]FeedbackAttachment, error)
	GetFeedbackAttachmentByID(attachmentID uuid.UUID) (*FeedbackAttachment, error)
	GetFeedbackStats() (map[string]interface{}, error)
	GetFeedbackAnalytics() (map[string]interface{}, error)

	// Monitoring operations
	InsertPerformanceMetric(metric *PerformanceMetric) error
	InsertAlert(alert *Alert) error
	InsertSystemHealthCheck(check *SystemHealthCheck) error
	InsertMonitoringEvent(event *MonitoringEvent) error
	QueryMonitoringEvents(eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*MonitoringEvent, error)
	UpdateAlertStatus(alert *Alert) error
	QueryPerformanceMetrics(metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*PerformanceMetric, error)
	QueryActiveAlerts(tenantID *uuid.UUID) ([]*Alert, error)
	QueryLatestSystemHealthChecks() (map[string]*SystemHealthCheck, error)
	GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error)
	StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error
	QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*DatabaseMetric, error)
	PgNotify(channel, payload string) error
	PgListen(ctx context.Context, channel string) error
	PgWaitForNotification(ctx context.Context) (*PgNotification, error)

	// Security operations
	CreateSecurityScan(scan *SecurityScan) (*SecurityScan, error)
	UpdateSecurityScan(scanID uuid.UUID, updates map[string]interface{}) (*SecurityScan, error)
	GetSecurityScan(scanID uuid.UUID) (*SecurityScan, error)
	ListSecurityScans(limit, offset int, filters map[string]interface{}) ([]*SecurityScan, error)
	CreateVulnerability(vuln *Vulnerability) (*Vulnerability, error)
	UpdateVulnerability(vulnID uuid.UUID, updates map[string]interface{}) (*Vulnerability, error)
	GetVulnerabilities(filters map[string]interface{}) ([]*Vulnerability, error)
	GetVulnerabilityByID(vulnID uuid.UUID) (*Vulnerability, error)

	// Session operations
	CreateSession(ctx context.Context, userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error)
	GetSessionByToken(ctx context.Context, sessionToken string) (*Session, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	UpdateSessionMFAStatus(ctx context.Context, sessionToken string, mfaVerified bool) error
	UpdateSessionActivity(ctx context.Context, sessionToken string) error
	DeleteSession(ctx context.Context, sessionToken string) error
	DeleteSessionByID(ctx context.Context, sessionID, userID uuid.UUID) error
	DeleteSessionByIDOnly(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
	ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error)
	CountActiveUserSessions(ctx context.Context, userID uuid.UUID) (int, error)
	ListTenantSessions(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Session, error)

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
	ListUserRefreshTokens(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)

	// Login attempt operations (for account lockout protection)
	CreateLoginAttempt(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*LoginAttempt, error)
	GetRecentFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	GetUserLockoutStatus(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	ClearUserLockout(ctx context.Context, userID uuid.UUID) error
	DeleteOldLoginAttempts(ctx context.Context, before time.Time) (int64, error)

	// Auth event operations (for security auditing)
	LogAuthEvent(ctx context.Context, event *AuthEvent) error
	GetAuthEventsForUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuthEvent, error)
	GetAuthEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*AuthEvent, error)
	GetRecentAuthEvents(ctx context.Context, since time.Time, limit int) ([]*AuthEvent, error)
	DeleteOldAuthEvents(ctx context.Context, before time.Time) (int64, error)

	// Magic link operations (passwordless authentication)
	CreateMagicLink(ctx context.Context, email string, token string, userID *uuid.UUID, ipAddress, userAgent, redirectPath string, expiresAt time.Time) (*MagicLink, error)
	GetMagicLinkByToken(ctx context.Context, token string) (*MagicLink, error)
	MarkMagicLinkUsed(ctx context.Context, id uuid.UUID) error
	GetRecentMagicLinksByEmail(ctx context.Context, email string, since time.Time) ([]*MagicLink, error)
	DeleteExpiredMagicLinks(ctx context.Context) (int64, error)

	// Dashboard configuration operations
	CreateDashboardConfig(ctx context.Context, config *DashboardConfig) (*DashboardConfig, error)
	GetDashboardConfigsByTenant(tenantID uuid.UUID) ([]*DashboardConfig, error)
	GetDashboardConfigsByUser(userID uuid.UUID) ([]*DashboardConfig, error)
	GetDashboardConfigByID(configID uuid.UUID) (*DashboardConfig, error)
	UpdateDashboardConfig(ctx context.Context, configID uuid.UUID, updates map[string]interface{}) (*DashboardConfig, error)
	DeleteDashboardConfig(ctx context.Context, configID uuid.UUID) error

	// Local runtime registry operations
	RegisterLocalRuntime(ctx context.Context, instance *LocalRuntimeInstance) (*LocalRuntimeInstance, error)
	UpdateLocalRuntimeHeartbeat(ctx context.Context, runtimeID string) error
	GetLocalRuntimeByID(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeInstance, error)
	GetLocalRuntimeByRuntimeID(ctx context.Context, runtimeID string) (*LocalRuntimeInstance, error)
	ListActiveLocalRuntimes(ctx context.Context) ([]*LocalRuntimeInstance, error)
	DeregisterLocalRuntime(ctx context.Context, runtimeID string) error
	CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error)

	// Local runtime metrics operations
	RecordLocalRuntimeMetrics(ctx context.Context, metrics *LocalRuntimeMetric) error
	GetLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time, limit int) ([]*LocalRuntimeMetric, error)
	GetLatestLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeMetric, error)
	GetAggregatedLocalRuntimeMetrics(ctx context.Context, since time.Time) (map[string]interface{}, error)

	// Local runtime health operations
	RecordLocalRuntimeHealth(ctx context.Context, health *LocalRuntimeHealth) error
	GetLocalRuntimeHealth(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeHealth, error)

	// Function operations
	CreateFunction(ctx context.Context, function *FunctionConfig) (*FunctionConfig, error)
	GetFunctionByID(ctx context.Context, functionID uuid.UUID) (*FunctionConfig, error)
	ListFunctionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FunctionConfig, error)
	ListAllFunctions(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status *string) ([]*FunctionConfig, int, error)
	UpdateFunction(ctx context.Context, functionID uuid.UUID, updates map[string]interface{}) (*FunctionConfig, error)
	DeleteFunction(ctx context.Context, functionID uuid.UUID) error
	GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*FunctionConfig, error)
	GetActiveDeploymentForFunction(ctx context.Context, functionID uuid.UUID) (*FunctionDeployment, error)

	// Function deployment operations
	CreateFunctionDeployment(ctx context.Context, deployment *FunctionDeployment) (*FunctionDeployment, error)
	GetFunctionDeploymentByID(ctx context.Context, deploymentID uuid.UUID) (*FunctionDeployment, error)
	ListFunctionDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*FunctionDeployment, error)
	UpdateFunctionDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL, errorMessage *string) error

	// Function log operations
	CreateFunctionLog(ctx context.Context, log *FunctionLog) error
	GetFunctionLogs(ctx context.Context, functionID *uuid.UUID, deploymentID *uuid.UUID, limit int, since *time.Time, level *string) ([]*FunctionLog, error)
	DeleteFunctionLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error)

	// Dashboard aggregations (tenant-scoped)
	GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]UsageByDay, error)
	GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]ExecutionRateByHour, error)
	GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]DashboardActivityItem, error)
	GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*DashboardMetrics, error)

	// Incident operations
	CreateIncident(ctx context.Context, incident *Incident) (*Incident, error)
	GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*Incident, error)
	ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*Incident, error)
	ListIncidentsSince(ctx context.Context, since time.Time, limit int) ([]*Incident, error)
	CountIncidentsSince(ctx context.Context, since time.Time) (int, error)
	CountIncidentsGroupedByDay(ctx context.Context, since time.Time) ([]DailyIncidentCount, error)
	UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*Incident, error)
	ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*Incident, error)

	// Provider operations
	CreateProvider(provider *Provider) error
	GetProviderByID(providerID string) (*Provider, error)
	GetProviderByUserAndType(userID uuid.UUID, providerType string) (*Provider, error)
	GetProvidersByUser(userID uuid.UUID) ([]*Provider, error)
	ListAllProviders(ctx context.Context) ([]*Provider, error)
	UpdateProviderStatus(providerID string, status string) error
	UpdateProvider(ctx context.Context, providerID string, updates map[string]interface{}) (*Provider, error)
	UpdateProviderLastUsed(ctx context.Context, providerID string) error
	GetStaleProviders(ctx context.Context, since time.Time) ([]*Provider, error)
	ShareProviderWithTeam(providerID string, teamID string) error
	DeleteProvider(ctx context.Context, providerID string, userID uuid.UUID) error

	// Encryption operations
	EncryptField(value string) (string, error)
	DecryptField(value string) (string, error)

	// Team invite operations
	CreateTeamInvite(invite *TeamInvite) error
	GetTeamInviteByToken(token string) (*TeamInvite, error)
	GetTeamInvitesByTeam(teamID uuid.UUID) ([]*TeamInvite, error)
	UpdateTeamInviteStatus(inviteID uuid.UUID, status string) error
	GetTeamByUserID(userID uuid.UUID) (*Team, error)
	IsTeamAdmin(userID uuid.UUID, teamID string) (bool, error)

	// Follow operations
	// User follows
	FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*UserFollow, error)
	UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error
	IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error)
	GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error)
	GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error)
	GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error)

	// Function follows
	FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*FunctionFollow, error)
	UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error
	IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error)
	GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error)
	GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error)
	GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error)

	// Function favorites
	AddFavorite(ctx context.Context, userID, functionID uuid.UUID, position int) (*FunctionFavorite, error)
	RemoveFavorite(ctx context.Context, userID, functionID uuid.UUID) error
	IsFavorite(ctx context.Context, userID, functionID uuid.UUID) (bool, error)
	GetUserFavorites(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FunctionFavorite, int, error)
	GetFavoriteCount(ctx context.Context, functionID uuid.UUID) (int, error)
	UpdateFavoritePosition(ctx context.Context, userID, functionID uuid.UUID, position int) error
	GetFavoriteByFunction(ctx context.Context, userID, functionID uuid.UUID) (*FunctionFavorite, error)

	// Username change operations (2-per-year limit with early-change fee)
	CreateUsernameChangeHistory(ctx context.Context, history *UsernameChangeHistory) error
	ChangeUsernameWithHistory(ctx context.Context, userID uuid.UUID, newUsername string, history *UsernameChangeHistory) error
	GetUsernameChangeHistory(ctx context.Context, userID uuid.UUID) ([]*UsernameChangeHistory, error)
	CountUsernameChangesInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (int, error)
	GetLastUsernameChange(ctx context.Context, userID uuid.UUID) (*UsernameChangeHistory, error)
	HasUsernameChangedInWindow(ctx context.Context, userID uuid.UUID, windowStart time.Time) (bool, error)

	// Pending username change operations (for payment flow)
	CreatePendingUsernameChange(ctx context.Context, pending *PendingUsernameChange) error
	GetPendingUsernameChangeByID(ctx context.Context, id uuid.UUID) (*PendingUsernameChange, error)
	GetPendingUsernameChangeByCheckoutSession(ctx context.Context, sessionID string) (*PendingUsernameChange, error)
	UpdatePendingUsernameChangeStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteExpiredPendingUsernameChanges(ctx context.Context) (int64, error)
	ListPendingUsernameChangesForUser(ctx context.Context, userID uuid.UUID) ([]*PendingUsernameChange, error)

	// User profile operations
	GetUserSkills(userID uuid.UUID) ([]*UserSkill, error)
	AddUserSkill(skill *UserSkill) error
	RemoveUserSkill(skillID uuid.UUID) error

	// User achievements operations
	GetUserAchievements(userID uuid.UUID) ([]*UserAchievement, error)
	GetAchievementBySlug(slug string) (*Achievement, error)
	ListAchievements() ([]*Achievement, error)
	AwardAchievement(userID, achievementID uuid.UUID, metadata map[string]interface{}) error
	UpdateAchievementProgress(userAchievementID uuid.UUID, progress int, isCompleted bool) error

	// User activity operations
	GetUserActivity(userID uuid.UUID, limit, offset int) ([]*UserActivity, error)
	CreateUserActivity(activity *UserActivity) error
	// GetUserContributionDailyCounts aggregates profile contribution events per UTC day:
	// user_activity rows plus registry function publishes (owner), since the given instant.
	GetUserContributionDailyCounts(userID uuid.UUID, since time.Time) (map[string]int64, error)

	// User analytics operations
	GetUserExecutionStats(userID uuid.UUID) (map[string]interface{}, error)
	// GetUserProfileStats returns authoritative counts for profile UI (not limited by registry list pagination).
	GetUserProfileStats(userID uuid.UUID) (map[string]interface{}, error)
	GetUserTrustBreakdown(userID uuid.UUID) (map[string]interface{}, error)
	GetUserPopularFunctions(userID uuid.UUID, limit int) ([]map[string]interface{}, error)
	GetUserGeographicStats(userID uuid.UUID) (map[string]interface{}, error)
	GetUserDeviceStats(userID uuid.UUID) (map[string]interface{}, error)
	// InitializeTenantAnalytics creates default analytics tracking for a newly provisioned tenant
	InitializeTenantAnalytics(tenantID uuid.UUID) error

	// Email event operations
	CreateEmailEvent(ctx context.Context, event *EmailEvent) error
	GetEmailEvents(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*EmailEvent, error)
	GetPendingBounceReviews(ctx context.Context, limit, offset int) ([]*EmailEvent, error)
	MarkEmailEventReviewed(ctx context.Context, eventID int64, reviewedBy uuid.UUID) error
	GetEmailEventStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error)

	// Email workflow operations (tenant-isolated, auto-provisioned per bundle)
	CreateEmailWorkflowConfig(ctx context.Context, config *EmailWorkflowConfig) error
	GetEmailWorkflowConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]EmailWorkflowConfig, error)
	GetEmailWorkflowConfigsByBundle(ctx context.Context, tenantID uuid.UUID, bundleSlug string) ([]EmailWorkflowConfig, error)
	GetActiveEmailWorkflowConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]EmailWorkflowConfig, error)
	GetEmailWorkflowConfigByID(ctx context.Context, id uuid.UUID) (*EmailWorkflowConfig, error)
	UpdateEmailWorkflowConfig(ctx context.Context, config *EmailWorkflowConfig) error
	DeleteEmailWorkflowConfig(ctx context.Context, id uuid.UUID) error
	CreateEmailWorkflowExecution(ctx context.Context, exec *EmailWorkflowExecution) error
	GetPendingEmailWorkflowExecutions(ctx context.Context, limit int) ([]EmailWorkflowExecution, error)
	GetEmailWorkflowExecutionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]EmailWorkflowExecution, error)
	GetEmailWorkflowExecutionsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]EmailWorkflowExecution, error)
	UpdateEmailWorkflowExecution(ctx context.Context, exec *EmailWorkflowExecution) error
	MarkEmailWorkflowExecutionSent(ctx context.Context, id uuid.UUID) error
	MarkEmailWorkflowExecutionFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
	RetryFailedEmailWorkflowExecutions(ctx context.Context, maxRetries int) ([]EmailWorkflowExecution, error)
	CleanupOldEmailWorkflowExecutions(ctx context.Context, retentionDays int) (int64, error)

	// Waitlist operations
	CreateWaitlistEntry(ctx context.Context, email, name, company, useCase, source, ip, userAgent string) (*WaitlistEntry, error)
	GetWaitlistEntryByEmail(ctx context.Context, email string) (*WaitlistEntry, error)
	ListWaitlistEntries(ctx context.Context, status string, limit, offset int) ([]WaitlistEntryAdminList, int64, error)
	GetWaitlistStats(ctx context.Context) (*WaitlistStats, error)
	UpdateWaitlistEntryStatus(ctx context.Context, id uuid.UUID, status, notes string) error
	IssueInviteToWaitlistEntry(ctx context.Context, entryID, inviteCodeID uuid.UUID) error
	DeleteWaitlistEntry(ctx context.Context, id uuid.UUID) error

	// Newsletter operations
	CreateNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string) (*NewsletterSubscriber, error)
	GetNewsletterSubscriberByEmail(ctx context.Context, email string) (*NewsletterSubscriber, error)
	GetNewsletterSubscriberByID(ctx context.Context, id uuid.UUID) (*NewsletterSubscriber, error)
	ListNewsletterSubscribers(ctx context.Context, status string, limit, offset int) ([]NewsletterSubscriber, int64, error)
	GetActiveNewsletterSubscribers(ctx context.Context) ([]NewsletterSubscriber, error)
	UnsubscribeNewsletterSubscriber(ctx context.Context, email string) error
	DeleteNewsletterSubscriber(ctx context.Context, id uuid.UUID) error
	GetNewsletterStats(ctx context.Context) (map[string]interface{}, error)
	CreateNewsletterCampaign(ctx context.Context, campaign *NewsletterCampaign) (*NewsletterCampaign, error)
	GetNewsletterCampaignByID(ctx context.Context, id uuid.UUID) (*NewsletterCampaign, error)
	ListNewsletterCampaigns(ctx context.Context, status string, limit, offset int) ([]NewsletterCampaign, int64, error)
	UpdateNewsletterCampaign(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*NewsletterCampaign, error)
	CreateNewsletterCampaignEmail(ctx context.Context, campaignEmail *NewsletterCampaignEmail) error
	UpdateNewsletterCampaignEmailStatus(ctx context.Context, id uuid.UUID, status string, emailID string) error
	GetNewsletterCampaignEmailsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]NewsletterCampaignEmail, error)
	UpdateCampaignStats(ctx context.Context, campaignID uuid.UUID) error

	// Usage Forecasting and Alerting operations
	CreateUsageAlert(ctx context.Context, alert *UsageAlert) error
	GetUsageAlertByID(ctx context.Context, id uuid.UUID) (*UsageAlert, error)
	ListUsageAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*UsageAlert, error)
	UpdateUsageAlert(ctx context.Context, alert *UsageAlert) error
	DeleteUsageAlert(ctx context.Context, id uuid.UUID) error
	RecordAlertTrigger(ctx context.Context, history *UsageAlertHistory) error
	GetAlertHistoryByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*UsageAlertHistory, error)

	CreateOrUpdateSpendCap(ctx context.Context, cap *SpendCap) error
	GetSpendCapByTenant(ctx context.Context, tenantID uuid.UUID, periodStart time.Time) (*SpendCap, error)
	UpdateCurrentSpend(ctx context.Context, capID uuid.UUID, spendCents int) error

	SaveUsageForecast(ctx context.Context, forecast *UsageForecast) error
	GetLatestForecast(ctx context.Context, tenantID uuid.UUID, forecastType string) (*UsageForecast, error)
	GetDailyUsageHistory(ctx context.Context, tenantID uuid.UUID, eventType string, days int) ([]*DailyUsagePoint, error)
	GetDailySpendHistory(ctx context.Context, tenantID uuid.UUID, days int) ([]*DailyUsagePoint, error)
	GetCurrentPeriodUsage(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*UsageSummary, error)

	// Cost Allocation operations for detailed cost tracking
	RecordCostAllocationEntry(ctx context.Context, entry *CostAllocationEntry) error
	GetCostAllocationByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*CostAllocationSummary, error)
	GetCostAllocationDailyBreakdown(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*DailyCostBreakdown, error)
	GetTenantCostSummary(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (*TenantCostSummary, error)
	GetAllTenantsCostSummary(ctx context.Context, start, end time.Time) ([]*TenantCostSummary, error)
	GetCostAllocationEntries(ctx context.Context, filter *CostAllocationFilter, limit, offset int) ([]*CostAllocationEntry, int, error)
	GetCostAllocationReport(ctx context.Context, start, end time.Time) (*CostAllocationReport, error)
	GetCostAllocationByRegion(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (map[string]*CostAllocationSummary, error)
	DeleteOldCostAllocationEntries(ctx context.Context, before time.Time) (int64, error)

	// Execution log retention operations
	DeleteOldExecutions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldPublicExecutions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldResourceUsage(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldMEGRecords(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldDriftReports(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteOldExecutionCertificates(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	GetExecutionRetentionStats(ctx context.Context) (map[string]interface{}, error)

	// Execution retention settings operations
	GetExecutionRetentionSettings(ctx context.Context) (*ExecutionRetentionSettings, error)
	UpdateExecutionRetentionSettings(ctx context.Context, updates *ExecutionRetentionSettingsUpdate) (*ExecutionRetentionSettings, error)
	GetOrCreateExecutionRetentionSettings(ctx context.Context) (*ExecutionRetentionSettings, error)
	ResetExecutionRetentionSettingsToDefaults(ctx context.Context, updatedBy *uuid.UUID) (*ExecutionRetentionSettings, error)

	// Usage Export operations
	CreateUsageExportConfiguration(ctx context.Context, config *UsageExportConfiguration) error
	GetUsageExportConfiguration(ctx context.Context, id uuid.UUID) (*UsageExportConfiguration, error)
	UpdateUsageExportConfiguration(ctx context.Context, config *UsageExportConfiguration) error
	DeleteUsageExportConfiguration(ctx context.Context, id uuid.UUID) error
	ListUsageExportConfigurations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UsageExportConfiguration, error)

	CreateUsageExportJob(ctx context.Context, job *UsageExportJob) error
	GetUsageExportJob(ctx context.Context, id uuid.UUID) (*UsageExportJob, error)
	UpdateUsageExportJobStatus(ctx context.Context, id uuid.UUID, status UsageExportStatus, errorMessage string) error
	CompleteUsageExportJob(ctx context.Context, id uuid.UUID, storagePath, storageURL, checksum string, recordCount, fileSize int64) error
	UpdateDeliveryStatus(ctx context.Context, jobID uuid.UUID, status, errorMessage string) error
	UpdateLastExecution(ctx context.Context, configID, jobID uuid.UUID, executedAt time.Time) error
	ListUsageExportJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UsageExportJob, error)
	GetPendingScheduledConfigs(ctx context.Context, now time.Time) ([]*UsageExportConfiguration, error)

	CreateExternalBillingSystem(ctx context.Context, system *ExternalBillingSystem) error
	GetExternalBillingSystem(ctx context.Context, id uuid.UUID) (*ExternalBillingSystem, error)
	UpdateExternalBillingSystem(ctx context.Context, system *ExternalBillingSystem) error
	DeleteExternalBillingSystem(ctx context.Context, id uuid.UUID) error
	ListExternalBillingSystems(ctx context.Context, tenantID uuid.UUID, limit, offset int, activeOnly bool) ([]*ExternalBillingSystem, error)

	CreateBillingIntegrationSync(ctx context.Context, sync *BillingIntegrationSync) error
	GetBillingIntegrationSync(ctx context.Context, id uuid.UUID) (*BillingIntegrationSync, error)
	ListBillingIntegrationSyncs(ctx context.Context, tenantID uuid.UUID, systemID *uuid.UUID, status string, limit, offset int) ([]*BillingIntegrationSync, error)

	CreateUsageExportTemplate(ctx context.Context, template *UsageExportTemplate) error
	GetUsageExportTemplate(ctx context.Context, id uuid.UUID) (*UsageExportTemplate, error)
	ListUsageExportTemplates(ctx context.Context, category string) ([]*UsageExportTemplate, error)

	// Team Memory operations (Team Memory Engine - Shared Brain)
	CreateTeamMemory(ctx context.Context, memory *TeamMemory) (*TeamMemory, error)
	GetTeamMemoryByID(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) (*TeamMemory, error)
	UpdateTeamMemory(ctx context.Context, memory *TeamMemory) (*TeamMemory, error)
	DeleteTeamMemory(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) error
	ListTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, filter TeamMemoryFilter) ([]*TeamMemory, int64, error)
	ListTeamMemoriesByType(ctx context.Context, tenantID, teamID uuid.UUID, memoryType string, limit, offset int) ([]*TeamMemory, error)
	SearchTeamMemories(ctx context.Context, tenantID, teamID uuid.UUID, query string, limit int) ([]*TeamMemorySearchResult, error)
	SearchTeamMemoriesByVector(ctx context.Context, tenantID, teamID uuid.UUID, embedding []float32, limit int) ([]*TeamMemorySearchResult, error)
	ValidateTeamMemory(ctx context.Context, memoryID uuid.UUID, validatedBy uuid.UUID) error
	MarkTeamMemoryAsAccessed(ctx context.Context, memoryID uuid.UUID) error
	CreateEncryptedTeamMemory(ctx context.Context, memory *TeamMemory, encryptedContent, iv, tag []byte) (*TeamMemory, error)
	GetTeamMemoryDecryptionPayload(ctx context.Context, memoryID uuid.UUID) (encryptedContent, iv, tag []byte, err error)

	// Memory extraction queue (auto-update pipeline)
	CreateMemoryExtraction(ctx context.Context, extraction *MemoryExtraction) (*MemoryExtraction, error)
	GetMemoryExtractionsByTeam(ctx context.Context, teamID uuid.UUID, status string, limit int) ([]*MemoryExtraction, error)
	ApproveMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID) (*TeamMemory, error)
	RejectMemoryExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID, reason string) error
	ProcessAutoApplyExtractions(ctx context.Context, batchSize int) (int, error)

	// Memory sharing (cross-team collaboration)
	CreateMemoryShare(ctx context.Context, share *MemoryShare) error
	GetMemoryShareByID(ctx context.Context, shareID uuid.UUID) (*MemoryShare, error)
	GetMemoryShareBetweenTeams(ctx context.Context, memoryID, sourceTeamID, targetTeamID uuid.UUID) (*MemoryShare, error)
	UpdateMemoryShare(ctx context.Context, share *MemoryShare) error
	ListMemorySharesByTargetTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*MemoryShare, error)
	ListMemorySharesBySourceTeam(ctx context.Context, teamID uuid.UUID, status string, limit, offset int) ([]*MemoryShare, error)
	ListMemorySharesByMemoryID(ctx context.Context, memoryID uuid.UUID, status string) ([]*MemoryShare, error)

	// ── Tenant Auth Settings (Backend-in-a-Box) ────────────────────────────────
	CreateAuthSettings(ctx context.Context, settings *TenantAuthSettings) error
	GetAuthSettings(ctx context.Context, tenantID uuid.UUID) (*TenantAuthSettings, error)
	UpdateAuthSettings(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error
	DeleteAuthSettings(ctx context.Context, tenantID uuid.UUID) error

	// Tenant OAuth Providers
	CreateOAuthProvider(ctx context.Context, provider *TenantOAuthProvider) error
	GetOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*TenantOAuthProvider, error)
	ListOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*TenantOAuthProvider, error)
	UpdateOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string, updates map[string]interface{}) (*TenantOAuthProvider, error)
	DeleteOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) error
	GetEnabledOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*TenantOAuthProvider, error)

	// Tenant Invite Codes
	CreateInviteCode(ctx context.Context, invite *TenantInviteCode) error
	GetInviteCode(ctx context.Context, code string) (*TenantInviteCode, error)
	GetInviteCodesByTenant(ctx context.Context, tenantID uuid.UUID, includeUsed bool) ([]*TenantInviteCode, error)
	GetInviteCodeByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*TenantInviteCode, error)
	AcceptInviteCode(ctx context.Context, code string, userID uuid.UUID) error
	RevokeInviteCode(ctx context.Context, code string) error
	IncrementInviteCodeUses(ctx context.Context, code string) error
	DeleteExpiredInviteCodes(ctx context.Context) (int64, error)

	// Tenant Memberships
	CreateMembership(ctx context.Context, membership *TenantMembership) error
	GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*TenantMembership, error)
	ListMemberships(ctx context.Context, tenantID uuid.UUID) ([]*TenantMembership, error)
	ListMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) ([]*TenantMembership, error)
	UpdateMembership(ctx context.Context, tenantID, userID uuid.UUID, updates map[string]interface{}) (*TenantMembership, error)
	DeleteMembership(ctx context.Context, tenantID, userID uuid.UUID) error
	UpdateMembershipLastActive(ctx context.Context, tenantID, userID uuid.UUID) error
	CountMembershipsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) (int, error)

	// Auth Audit Log
	CreateAuthAuditLog(ctx context.Context, log *TenantAuthAuditLog) error
	ListAuthAuditLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int, actions []string, userID *uuid.UUID, since *time.Time) ([]*TenantAuthAuditLog, int, error)
	GetAuthAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*TenantAuthAuditLog, error)
	DeleteOldAuthAuditLogs(ctx context.Context, before time.Time) (int64, error)

	// Tenant Stripe Config (Isolated Payment Processing)
	GetTenantStripeConfig(ctx context.Context, tenantID uuid.UUID) (*TenantStripeConfig, error)
	CreateTenantStripeConfig(ctx context.Context, config *TenantStripeConfig) error
	UpdateTenantStripeConfig(ctx context.Context, config *TenantStripeConfig) error
}
