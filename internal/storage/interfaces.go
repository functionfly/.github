package storage

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// PgNotification is re-exported from the types package.
type PgNotification = types.PgNotification

// Type aliases for the types package. The canonical definitions live in
// internal/types; these aliases preserve the public API of the storage package
// (e.g. `storage.User`, `storage.Repository`, etc.) so existing callers continue
// to compile unchanged.
type (
	APIKeyBudget                      = types.APIKeyBudget
	APIKeyCostSummary                 = types.APIKeyCostSummary
	ARRMetrics                        = types.ARRMetrics
	Achievement                       = types.Achievement
	AgentFunction                     = types.AgentFunction
	AgentFunctionCategory             = types.AgentFunctionCategory
	AgentFunctionDefinition           = types.AgentFunctionDefinition
	AgentFunctionExecution            = types.AgentFunctionExecution
	AgentFunctionExecutionsFilter     = types.AgentFunctionExecutionsFilter
	AgentFunctionPolicy               = types.AgentFunctionPolicy
	AgentMCPServer                    = types.AgentMCPServer
	AgentMemory                       = types.AgentMemory
	AgentMemoryIndex                  = types.AgentMemoryIndex
	AgentSubscription                 = types.AgentSubscription
	AgentTierPricing                  = types.AgentTierPricing
	AgentUsage                        = types.AgentUsage
	AggregatedBillingUsage            = types.AggregatedBillingUsage
	AffiliateCode                     = types.AffiliateCode
	AffiliateCommission               = types.AffiliateCommission
	AffiliateReferral                 = types.AffiliateReferral
	Alert                             = types.Alert
	AnalyticsEvent                    = types.AnalyticsEvent
	App                               = types.App
	AppAnalyticsResponse              = types.AppAnalyticsResponse
	AppAnalyticsSummary               = types.AppAnalyticsSummary
	AppBackendBreakdown               = types.AppBackendBreakdown
	AppErrorBreakdown                 = types.AppErrorBreakdown
	AppLatencyTimeseriesPoint         = types.AppLatencyTimeseriesPoint
	AppRequestTimeseriesPoint         = types.AppRequestTimeseriesPoint
	AuditEvent                        = types.AuditEvent
	AuthEvent                         = types.AuthEvent
	Backend                           = types.Backend
	BackendStatus                     = types.BackendStatus
	BillingIntegrationSync            = types.BillingIntegrationSync
	BillingSystemType                 = types.BillingSystemType
	BlogDailyAnalytics                = types.BlogDailyAnalytics
	BlogAnalyticsSummary              = types.BlogAnalyticsSummary
	BlogAuthor                        = types.BlogAuthor
	BlogCategory                      = types.BlogCategory
	BlogPageView                      = types.BlogPageView
	BlogPost                          = types.BlogPost
	BlogSettings                      = types.BlogSettings
	BlogViewsTimeSeries               = types.BlogViewsTimeSeries
	BundleSubscription                = types.BundleSubscription
	ChurnMetrics                      = types.ChurnMetrics
	CohortRetention                   = types.CohortRetention
	CostAllocationChargeback          = types.CostAllocationChargeback
	CostAllocationEntry               = types.CostAllocationEntry
	CostAllocationFilter              = types.CostAllocationFilter
	CostAllocationReport              = types.CostAllocationReport
	CostAnomaly                       = types.CostAnomaly
	Coupon                            = types.Coupon
	CouponRedemption                  = types.CouponRedemption
	CreditNote                        = types.CreditNote
	CreditNoteFilter                  = types.CreditNoteFilter
	CreditNoteLineItem                = types.CreditNoteLineItem
	CreditNoteStats                   = types.CreditNoteStats
	CurrencyExchangeRate              = types.CurrencyExchangeRate
	DashboardActivityItem             = types.DashboardActivityItem
	DashboardMetrics                  = types.DashboardMetrics
	DeploymentArtifact                = types.DeploymentArtifact
	DeployKeyCreateRequest            = types.DeployKeyCreateRequest
	DeployKeyListResponse             = types.DeployKeyListResponse
	DeployKeyResponse                 = types.DeployKeyResponse
	EmailEvent                        = types.EmailEvent
	ExecutionRateByHour               = types.ExecutionRateByHour
	ExecutionRetentionSettings        = types.ExecutionRetentionSettings
	ExecutionRetentionSettingsUpdate  = types.ExecutionRetentionSettingsUpdate
	FailedPaymentMetrics              = types.FailedPaymentMetrics
	FinancialReport                   = types.FinancialReport
	FunctionUsageRollup               = types.FunctionUsageRollup
	FunctionWebhookCreateRequest      = types.FunctionWebhookCreateRequest
	FunctionWebhookDeliveryListResponse = types.FunctionWebhookDeliveryListResponse
	FunctionWebhookDeliveryResponse   = types.FunctionWebhookDeliveryResponse
	FunctionWebhookListResponse       = types.FunctionWebhookListResponse
	FunctionWebhookResponse           = types.FunctionWebhookResponse
	FunctionWebhookTestRequest        = types.FunctionWebhookTestRequest
	FunctionWebhookUpdateRequest      = types.FunctionWebhookUpdateRequest
	GradingConfig                     = types.GradingConfig
	HealthCheck                       = types.HealthCheck
	Invoice                           = types.Invoice
	JSONMap                           = types.JSONMap
	LTVMetrics                        = types.LTVMetrics
	LinkConnectorRequest              = types.LinkConnectorRequest
	LoginAttempt                      = types.LoginAttempt
	MRRCohortData                     = types.MRRCohortData
	MRRMetrics                        = types.MRRMetrics
	MagicLink                         = types.MagicLink
	MemoryStats                       = types.MemoryStats
	NewsletterCampaign                = types.NewsletterCampaign
	ScheduleConfig                    = types.ScheduleConfig
	SyncTriggerResponse               = types.SyncTriggerResponse
	TransformRule                     = types.TransformRule
	NewsletterCampaignEmail           = types.NewsletterCampaignEmail
	NewsletterSubscriber              = types.NewsletterSubscriber
	OAuthState                        = types.OAuthState
	PasswordPolicy                    = types.PasswordPolicy
	PaymentMethodInfoExtended         = types.PaymentMethodInfoExtended
	PendingUsernameChange             = types.PendingUsernameChange
	PricingModel                      = types.PricingModel
	PricingTier                       = types.PricingTier
	Provider                          = types.Provider
	ProviderSettings                  = types.ProviderSettings
	RefreshToken                      = types.RefreshToken
	RevenueRecognitionEntry           = types.RevenueRecognitionEntry
	RoutingEvent                      = types.RoutingEvent
	Session                           = types.Session
	SignupInviteCode                  = types.SignupInviteCode
	SignupInviteCodeAdminList         = types.SignupInviteCodeAdminList
	SocialLinks                       = types.SocialLinks
	StateCheck                        = types.StateCheck
	StoredWebhookPayload              = types.StoredWebhookPayload
	StripeSyncEvent                   = types.StripeSyncEvent
	Subscription                      = types.Subscription
	SubscriptionChurnEvent            = types.SubscriptionChurnEvent
	SubscriptionLifecycleMetrics      = types.SubscriptionLifecycleMetrics
	SupportedCurrency                 = types.SupportedCurrency
	TaxCalculationRequest             = types.TaxCalculationRequest
	TaxCalculationResult              = types.TaxCalculationResult
	TaxExemptionCertificate           = types.TaxExemptionCertificate
	TaxIDType                         = types.TaxIDType
	TaxIDValidationLog                = types.TaxIDValidationLog
	TaxRate                           = types.TaxRate
	TaxSettings                       = types.TaxSettings
	TeamAuditLog                      = types.TeamAuditLog
	TeamInvite                        = types.TeamInvite
	TeamMemory                        = types.TeamMemory
	TeamMemoryFilter                  = types.TeamMemoryFilter
	TeamMemorySearchResult            = types.TeamMemorySearchResult
	TeamMembership                    = types.TeamMembership
	TeamPermission                    = types.TeamPermission
	TeamQuota                         = types.TeamQuota
	Tenant                            = types.Tenant
	TenantAuthAuditLog                = types.TenantAuthAuditLog
	TenantAuthSettings                = types.TenantAuthSettings
	TenantInviteCode                  = types.TenantInviteCode
	TenantMembership                  = types.TenantMembership
	TenantOAuthProvider               = types.TenantOAuthProvider
	User                              = types.User
	UserAchievement                   = types.UserAchievement
	UserActivity                      = types.UserActivity
	UserSearchHit                     = types.UserSearchHit
	UserSkill                         = types.UserSkill
	UsernameChangeHistory             = types.UsernameChangeHistory
	UsageByDay                        = types.UsageByDay
	UsageEvent                        = types.UsageEvent
	UsageForecastConfig               = types.UsageForecastConfig
	UsageRollup                       = types.UsageRollup
	UsageTrend                        = types.UsageTrend
	VerificationFee                   = types.VerificationFee
	FunctionVerificationPayment       = types.FunctionVerificationPayment
	PublisherEarning                  = types.PublisherEarning
	PlatformFee                       = types.PlatformFee
	PricingBundle                     = types.PricingBundle
	PricingTierExtended               = types.PricingTierExtended
	FounderModeRegistration           = types.FounderModeRegistration
	DeferredBillingConfig             = types.DeferredBillingConfig
	FeatureMeasure                    = types.FeatureMeasure
	CircuitState                      = types.CircuitState
	Deployment                        = types.Deployment
	EmailWorkflowConfig               = types.EmailWorkflowConfig
	EmailWorkflowExecution            = types.EmailWorkflowExecution
	EnvironmentVariable               = types.EnvironmentVariable
	MemoryExtraction                  = types.MemoryExtraction
	ChangelogEntry                    = types.ChangelogEntry
	ChangelogChange                   = types.ChangelogChange
	FeedbackType                      = types.FeedbackType
	FeedbackStatus                    = types.FeedbackStatus
	FeedbackPriority                  = types.FeedbackPriority
	DeploymentStatus                  = types.DeploymentStatus
	DeploymentStatusType              = types.DeploymentStatusType
	HealthCheckType                   = types.HealthCheckType
	Team                              = types.Team
	FeedbackAttachment = types.FeedbackAttachment
	MonitoringEvent = types.MonitoringEvent
	PerformanceMetric = types.PerformanceMetric
	SystemHealthCheck = types.SystemHealthCheck
	DatabaseMetric = types.DatabaseMetric
	SecurityScan = types.SecurityScan
	Vulnerability = types.Vulnerability
	DashboardConfig = types.DashboardConfig
	LocalRuntimeInstance = types.LocalRuntimeInstance
	FunctionConfig = types.FunctionConfig
	LocalRuntimeHealth = types.LocalRuntimeHealth
	LocalRuntimeMetric = types.LocalRuntimeMetric
	FunctionDeployment = types.FunctionDeployment
	FunctionLog = types.FunctionLog
	Incident = types.Incident
	FunctionFavorite = types.FunctionFavorite
	FunctionFollow = types.FunctionFollow
	UserFollow = types.UserFollow
	UsageAlert = types.UsageAlert
	WaitlistEntry = types.WaitlistEntry
	WaitlistEntryAdminList = types.WaitlistEntryAdminList
	WaitlistStats = types.WaitlistStats
	DailyUsagePoint = types.DailyUsagePoint
	SpendCap = types.SpendCap
	UsageAlertHistory = types.UsageAlertHistory
	UsageForecast = types.UsageForecast
	CostAllocationSummary = types.CostAllocationSummary
	DailyCostBreakdown = types.DailyCostBreakdown
	TenantCostSummary = types.TenantCostSummary
	UsageExportConfiguration = types.UsageExportConfiguration
	ExternalBillingSystem = types.ExternalBillingSystem
	UsageExportJob = types.UsageExportJob
	UsageExportStatus = types.UsageExportStatus
	MemoryShare = types.MemoryShare
	UsageExportTemplate = types.UsageExportTemplate
	EUVATValidation = types.EUVATValidation
	TaxJurisdictionReport = types.TaxJurisdictionReport
	WebhookReplayRequest = types.WebhookReplayRequest
	DepartmentBudget = types.DepartmentBudget
	TeamCostAllocation = types.TeamCostAllocation
	TeamCostBreakdown = types.TeamCostBreakdown
	BrainSignal = types.BrainSignal
	BudgetAlert = types.BudgetAlert
	SignalFilter = types.SignalFilter
	SignalListParams = types.SignalListParams
	BrainComposer = types.BrainComposer
	BrainFeedbackRequest = types.BrainFeedbackRequest
	BrainSearchResult = types.BrainSearchResult
	BrainStats = types.BrainStats
	BrainTrigger = types.BrainTrigger
	CertQuestion = types.CertQuestion
	CertQuestionPublic = types.CertQuestionPublic
	CertTier = types.CertTier
	CertExam = types.CertExam
	CertPracticalChallenge = types.CertPracticalChallenge
	CertCredential = types.CertCredential
	CertGradingQueueItem = types.CertGradingQueueItem
	Connector = types.Connector
	UserConnector = types.UserConnector
	TopBlogPost = types.TopBlogPost
	Feedback = types.Feedback
	UsageSummary = types.UsageSummary
	DeployKey = types.DeployKey
	FunctionWebhookDelivery = types.FunctionWebhookDelivery
	FunctionWebhookSubscription = types.FunctionWebhookSubscription
	FunctionWebhookTestResponse = types.FunctionWebhookTestResponse
	GitHubConnection = types.GitHubConnection
	GitHubRepo = types.GitHubRepo
	ListReposParams = types.ListReposParams
	GitHubImport = types.GitHubImport
	GitHubSyncLog = types.GitHubSyncLog
	GitHubWebhook = types.GitHubWebhook
	ListImportsParams = types.ListImportsParams
	ListSyncLogsParams = types.ListSyncLogsParams
	GitHubImportTemplate = types.GitHubImportTemplate
	PCIAuditEvent = types.PCIAuditEvent
	ContractAsset = types.ContractAsset
	DeferredRevenueSummary = types.DeferredRevenueSummary
	PerformanceObligation = types.PerformanceObligation
	RevenueRecognitionEvent = types.RevenueRecognitionEvent
	RevenueRecognitionSchedule = types.RevenueRecognitionSchedule
	RecognizedRevenueSummary = types.RecognizedRevenueSummary
	UsageExportFormat = types.UsageExportFormat
	RevenueRecognitionReport = types.RevenueRecognitionReport

)

// Repository defines the interface for data access
type Repository interface {
	// User operations
	IsUsernameReserved(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, email, passwordHash string, tenantID uuid.UUID) (*User, error)
	CreateUserWithSocialAuth(ctx context.Context, email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*User, error)
	CreateUserWithRole(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	// GetUserForPublicProfile resolves by username (case-insensitive) or by unique email local-part (before @).
	GetUserForPublicProfile(ctx context.Context, login string) (*User, error)
	SearchUsersByUsernamePrefix(ctx context.Context, prefix string, limit int) ([]UserSearchHit, error)
	GetUserByVerificationToken(ctx context.Context, token string) (*User, error)
	GetUserBySocialProvider(ctx context.Context, provider, providerID string) (*User, error)
	ListUsers(ctx context.Context, ) ([]*User, error)
	ListUserIDsByTenant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error)
	ListActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*User, error)
	CountActiveUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*User, error)
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
	// OAuth state (CSRF) — persisted for multi-instance OAuth flows
	// redirectURI is optional; when set, callback redirects there with token (e.g. CLI local server).
	// inviteCode is optional; stored for invite-only signup validation on callback (short TTL).
	// codeVerifier is for PKCE (Proof Key for Code Exchange) - required for public clients.
	// loginHint preserves tenant subdomain or email context through the OAuth flow.
	// deviceFingerprint stores a hash of device characteristics for session binding validation.
	StoreOAuthState(ctx context.Context, state string, expiresAt time.Time, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string) error
	ValidateAndConsumeOAuthState(ctx context.Context, state string) (valid bool, redirectURI, inviteCode, codeVerifier, loginHint, deviceFingerprint string, err error)
	DeleteExpiredOAuthStates(ctx context.Context, ) (int64, error)

	// Signup invite codes (platform invite-only launch)
	CreateSignupInvite(ctx context.Context, label string, maxUses *int, expiresAt *time.Time, createdBy *uuid.UUID) (id uuid.UUID, plainCode string, err error)
	ListSignupInvitesAdmin(ctx context.Context) ([]SignupInviteCodeAdminList, error)
	RevokeSignupInvite(ctx context.Context, id uuid.UUID) error
	ValidateSignupInviteReadOnly(ctx context.Context, plainCode string) error
	ReserveSignupInvite(ctx context.Context, plainCode string) (inviteID uuid.UUID, err error)
	ReleaseSignupInviteReservation(ctx context.Context, inviteID uuid.UUID) error

	// Tenant operations
	CreateTenant(ctx context.Context, name string) (*Tenant, error)
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error)
	GetTenantByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*Tenant, error)
	ListTenants(ctx context.Context, ) ([]*Tenant, error)
	ListTenantsWithStripeCustomerID(ctx context.Context, ) ([]*Tenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error)
	UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error
	UpdateTenantTaxSettings(ctx context.Context, tenantID uuid.UUID, settings *TaxSettings) error
	SetTenantDegradedMode(ctx context.Context, tenantID uuid.UUID, degraded bool, reason string) error
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	CountUsersByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountRoutingEventsForTenantSince(ctx context.Context, tenantID uuid.UUID, since time.Time) (int, error)

	// Tenant membership operations (multi-tenant support)
	IsUserInTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	AddTenantMember(ctx context.Context, userID, tenantID, invitedBy uuid.UUID, role string) error
	AcceptTenantMembership(ctx context.Context, userID, tenantID uuid.UUID) error
	RemoveTenantMember(ctx context.Context, userID, tenantID uuid.UUID) error

	// Team operations
	CreateTeam(ctx context.Context, team *Team) error
	GetTeamByID(ctx context.Context, teamID uuid.UUID) (*Team, error)
	GetTeamsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Team, error)
	UpdateTeam(ctx context.Context, team *Team) error
	DeleteTeam(ctx context.Context, teamID uuid.UUID) error
	AddTeamMember(ctx context.Context, membership *TeamMembership) error
	UpdateTeamMember(ctx context.Context, teamID, userID uuid.UUID, role string) error
	RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error
	GetTeamMembership(ctx context.Context, teamID, userID uuid.UUID) (*TeamMembership, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	GrantTeamPermission(ctx context.Context, permission *TeamPermission) error
	RevokeTeamPermission(ctx context.Context, teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error
	GetTeamPermissions(ctx context.Context, teamID uuid.UUID) ([]*TeamPermission, error)
	GetResourcePermissions(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*TeamPermission, error)
	CheckUserResourcePermission(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID, resourceType string) ([]string, error)
	IsUserTeamOwner(ctx context.Context, userID, teamID uuid.UUID) (bool, error)
	IsUserTeamAdmin(ctx context.Context, userID, teamID uuid.UUID) (bool, error)
	TransferTeamOwnership(ctx context.Context, teamID, fromUserID, toUserID uuid.UUID) error
	LeaveTeam(ctx context.Context, teamID, userID uuid.UUID) error
	CreateTeamAuditLog(ctx context.Context, log *TeamAuditLog) error
	GetTeamAuditLogs(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]*TeamAuditLog, error)
	GetTeamQuotas(ctx context.Context, teamID uuid.UUID) ([]*TeamQuota, error)
	UpdateTeamQuota(ctx context.Context, teamID uuid.UUID, resourceType string, delta int) error

	// Audit operations
	ListAuditEvents(ctx context.Context, limit, offset int) ([]*AuditEvent, error)
	LogAuditEvent(ctx context.Context, event *AuditEvent) error
	ListAuditEventsFiltered(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*AuditEvent, error)
	GetAuditEventByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error)

	// Billing operations
	CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error)
	ListPricingTiers(ctx context.Context, ) ([]*PricingTier, error)
	GetPricingTierByID(ctx context.Context, id uuid.UUID) (*PricingTier, error)
	UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*PricingTier, error)
	DeletePricingTier(ctx context.Context, id uuid.UUID) error

	CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetSubscriptionByTenantID(ctx context.Context, tenantID uuid.UUID) (*Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error)
	ListAllSubscriptions(ctx context.Context, limit, offset int) ([]*Subscription, error)
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
	ListInvoicesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Invoice, error)
	CountInvoicesByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	ListAllInvoices(ctx context.Context, limit, offset int) ([]*Invoice, error)
	GetInvoiceByID(ctx context.Context, id uuid.UUID) (*Invoice, error)
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
	GetUsageByTenant(ctx context.Context, tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error)
	GetUsageByTenantByFunction(ctx context.Context, tenantID uuid.UUID, start, end time.Time) ([]*FunctionUsageRollup, error)

	CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error)
	ListCoupons(ctx context.Context, ) ([]*Coupon, error)
	GetCouponByCode(ctx context.Context, code string) (*Coupon, error)
	RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error)

	// Affiliate / Referral Commission System
	CreateAffiliateCode(ctx context.Context, code *AffiliateCode) (*AffiliateCode, error)
	GetAffiliateCodeByID(ctx context.Context, id uuid.UUID) (*AffiliateCode, error)
	GetAffiliateCodeByCode(ctx context.Context, code string) (*AffiliateCode, error)
	ListAffiliateCodes(ctx context.Context, ) ([]*AffiliateCode, error)
	ListAffiliateCodesByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*AffiliateCode, error)
	UpdateAffiliateCode(ctx context.Context, code *AffiliateCode) error
	CreateAffiliateReferral(ctx context.Context, referral *AffiliateReferral) (*AffiliateReferral, error)
	GetAffiliateReferralByID(ctx context.Context, id uuid.UUID) (*AffiliateReferral, error)
	GetAffiliateReferralByTenant(ctx context.Context, tenantID uuid.UUID) (*AffiliateReferral, error)
	ListAffiliateReferralsByCode(ctx context.Context, codeID uuid.UUID) ([]*AffiliateReferral, error)
	UpdateAffiliateReferralStatus(ctx context.Context, id uuid.UUID, status string) error
	CreateAffiliateCommission(ctx context.Context, commission *AffiliateCommission) (*AffiliateCommission, error)
	GetAffiliateCommissionByID(ctx context.Context, id uuid.UUID) (*AffiliateCommission, error)
	ListAffiliateCommissionsByCode(ctx context.Context, codeID uuid.UUID) ([]*AffiliateCommission, error)
	ListAffiliateCommissionsByPublisher(ctx context.Context, publisherID uuid.UUID) ([]*AffiliateCommission, error)
	UpdateAffiliateCommissionStatus(ctx context.Context, id uuid.UUID, status string) error
	CalculateCommission(ctx context.Context, commissionType string, commissionValue, baseAmountUSD float64) (commissionCents int64, commissionUSD float64)

	// Revenue System Phase 1 - Trust Layer Monetization
	// Verification Fees
	GetVerificationFeeByLevel(ctx context.Context, level string) (*VerificationFee, error)
	ListVerificationFees(ctx context.Context, ) ([]*VerificationFee, error)

	// Function Verification Payments
	CreateFunctionVerificationPayment(ctx context.Context, payment *FunctionVerificationPayment) error
	GetFunctionVerificationPaymentByID(ctx context.Context, id uuid.UUID) (*FunctionVerificationPayment, error)
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
	ListPricingTiersExtended(ctx context.Context, ) ([]*PricingTierExtended, error)
	GetPricingTierExtendedByID(ctx context.Context, id uuid.UUID) (*PricingTierExtended, error)

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
	GetFounderModeAnalytics(ctx context.Context) (*FounderModeAnalytics, error)

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
	CalculateAndUpdateFounderTiers(ctx context.Context) error

	// Bundle Subscriptions
	CreateBundleSubscription(ctx context.Context, sub *BundleSubscription) error
	UpdateBundleSubscription(ctx context.Context, sub *BundleSubscription) error
	GetBundleSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*BundleSubscription, error)
	GetBundleSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*BundleSubscription, error)
	ListBundleSubscriptionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*BundleSubscription, error)
	ListBundleTemplates(ctx context.Context, bundleSlug string) ([]*BundleFunctionTemplate, error)
	ListPendingDeployments(ctx context.Context) ([]*BundleSubscription, error)
	ListAwaitingProvider(ctx context.Context, tenantID uuid.UUID) ([]*BundleSubscription, error)

	// App operations
	CreateApp(ctx context.Context, name, slug string, tenantID uuid.UUID) (*App, error)
	GetAppByID(ctx context.Context, id uuid.UUID) (*App, error)
	GetAppBySlug(ctx context.Context, slug string) (*App, error)
	// GetAppBySlugAndTenant returns an app by slug scoped to the tenant (dashboard / tenant APIs).
	GetAppBySlugAndTenant(ctx context.Context, slug string, tenantID uuid.UUID) (*App, error)
	ListAppsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*App, error)
	UpdateApp(ctx context.Context, id uuid.UUID, name string) (*App, error)
	DeleteApp(ctx context.Context, id uuid.UUID) error

	// Backend operations
	CreateBackend(ctx context.Context, appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error)
	ListBackendsByAppID(ctx context.Context, appID uuid.UUID) ([]*Backend, error)
	GetBackendByID(ctx context.Context, id uuid.UUID) (*Backend, error)
	GetAllEnabledBackends(ctx context.Context, ) ([]*Backend, error)
	ListAllBackends(ctx context.Context) ([]*Backend, error)
	UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error
	DeleteBackend(ctx context.Context, backendID uuid.UUID) error

	// Platform feature measures (admin security/features page)
	ListFeatureMeasures(ctx context.Context) ([]*FeatureMeasure, error)
	UpdateFeatureMeasureEnabled(ctx context.Context, id uuid.UUID, enabled bool) error

	// Health check operations
	InsertHealthCheck(ctx context.Context, backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error
	GetRecentHealthChecks(ctx context.Context, backendID uuid.UUID, limit int) ([]*HealthCheck, error)
	DeleteHealthChecksBefore(ctx context.Context, before time.Time) (int64, error)

	// Circuit breaker operations
	GetCircuitState(ctx context.Context, backendID uuid.UUID) (*CircuitState, error)
	UpdateCircuitState(ctx context.Context, state *CircuitState) error
	UpsertCircuitState(ctx context.Context, state *CircuitState) error

	// Routing operations
	InsertRoutingEvent(ctx context.Context, appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error
	GetRecentRoutingEvents(ctx context.Context, limit int, since time.Time) ([]*RoutingEvent, error)
	GetRecentRoutingEventsByBackend(ctx context.Context, backendID uuid.UUID, limit int) ([]*RoutingEvent, error)

	// App analytics operations
	GetAppAnalyticsSummary(ctx context.Context, appID uuid.UUID, since time.Time) (*AppAnalyticsSummary, error)
	GetAppRequestTimeseries(ctx context.Context, appID uuid.UUID, since time.Time, interval string) ([]*AppRequestTimeseriesPoint, error)
	GetAppLatencyTimeseries(ctx context.Context, appID uuid.UUID, since time.Time, interval string) ([]*AppLatencyTimeseriesPoint, error)
	GetAppTopErrors(ctx context.Context, appID uuid.UUID, since time.Time) ([]*AppErrorBreakdown, error)
	GetAppBackendBreakdown(ctx context.Context, appID uuid.UUID, since time.Time) ([]*AppBackendBreakdown, error)

	// Deployment operations
	CreateDeployment(ctx context.Context, appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status, message string, metadata map[string]interface{}) error
	GetDeploymentByID(ctx context.Context, id uuid.UUID) (*Deployment, error)
	ListDeploymentsByAppID(ctx context.Context, appID uuid.UUID, limit int) ([]*Deployment, error)
	GetLatestSuccessfulDeployment(ctx context.Context, appID uuid.UUID, provider string) (*Deployment, error)

	// Status operations
	GetBackendStatusByAppID(ctx context.Context, appID uuid.UUID) ([]*BackendStatus, error)

	// Artifact operations
	StoreDeploymentArtifact(ctx context.Context, appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*DeploymentArtifact, error)
	GetDeploymentArtifact(ctx context.Context, key string) (*DeploymentArtifact, error)

	// Content management operations
	// Changelog operations
	CreateChangelogEntry(ctx context.Context, entry *ChangelogEntry) (*ChangelogEntry, error)
	GetChangelogEntryByID(ctx context.Context, id uuid.UUID) (*ChangelogEntry, error)
	GetChangelogEntryByVersion(ctx context.Context, version string) (*ChangelogEntry, error)
	ListChangelogEntries(ctx context.Context, limit, offset int, publishedOnly bool) ([]*ChangelogEntry, error)
	UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogEntry, error)
	DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error
	CreateChangelogChange(ctx context.Context, change *ChangelogChange) (*ChangelogChange, error)
	UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogChange, error)
	DeleteChangelogChange(ctx context.Context, id uuid.UUID) error

	// Blog operations
	CreateBlogPost(ctx context.Context, post *BlogPost) (*BlogPost, error)
	GetBlogPostByID(ctx context.Context, id uuid.UUID) (*BlogPost, error)
	GetBlogPostBySlug(ctx context.Context, slug string) (*BlogPost, error)
	ListBlogPosts(ctx context.Context, limit, offset int, publishedOnly bool, tagFilter []string) ([]*BlogPost, error)
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
	CreateFeedback(ctx context.Context, feedback *Feedback) (*Feedback, error)
	GetFeedbackByID(ctx context.Context, id uuid.UUID) (*Feedback, error)
	GetFeedbackByUser(ctx context.Context, userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error)
	ListFeedback(ctx context.Context, limit, offset int, statusFilter *string, typeFilter *string) ([]Feedback, error)
	UpdateFeedbackStatus(id uuid.UUID, status string) error
	CreateFeedbackAttachment(ctx context.Context, attachment *FeedbackAttachment) (*FeedbackAttachment, error)
	GetFeedbackAttachments(ctx context.Context, feedbackID uuid.UUID) ([]FeedbackAttachment, error)
	GetFeedbackAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*FeedbackAttachment, error)
	GetFeedbackStats(ctx context.Context, ) (map[string]interface{}, error)
	GetFeedbackAnalytics(ctx context.Context, ) (map[string]interface{}, error)

	// Monitoring operations
	InsertPerformanceMetric(metric *PerformanceMetric) error
	InsertAlert(alert *Alert) error
	InsertSystemHealthCheck(check *SystemHealthCheck) error
	InsertMonitoringEvent(event *MonitoringEvent) error
	QueryMonitoringEvents(ctx context.Context, eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*MonitoringEvent, error)
	UpdateAlertStatus(alert *Alert) error
	QueryPerformanceMetrics(ctx context.Context, metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*PerformanceMetric, error)
	QueryActiveAlerts(ctx context.Context, tenantID *uuid.UUID) ([]*Alert, error)
	QueryLatestSystemHealthChecks(ctx context.Context, ) (map[string]*SystemHealthCheck, error)
	GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error)
	StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error
	QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*DatabaseMetric, error)
	PgNotify(channel, payload string) error
	PgListen(ctx context.Context, channel string) error
	PgWaitForNotification(ctx context.Context) (*PgNotification, error)

	// Security operations
	CreateSecurityScan(ctx context.Context, scan *SecurityScan) (*SecurityScan, error)
	UpdateSecurityScan(ctx context.Context, scanID uuid.UUID, updates map[string]interface{}) (*SecurityScan, error)
	GetSecurityScan(ctx context.Context, scanID uuid.UUID) (*SecurityScan, error)
	ListSecurityScans(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*SecurityScan, error)
	CreateVulnerability(ctx context.Context, vuln *Vulnerability) (*Vulnerability, error)
	UpdateVulnerability(ctx context.Context, vulnID uuid.UUID, updates map[string]interface{}) (*Vulnerability, error)
	GetVulnerabilities(ctx context.Context, filters map[string]interface{}) ([]*Vulnerability, error)
	GetVulnerabilityByID(ctx context.Context, vulnID uuid.UUID) (*Vulnerability, error)

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
	CreateLoginHistory(ctx context.Context, userID uuid.UUID, eventType, ipAddress, userAgent, device, loginMethod string, mfaUsed bool, sessionID *uuid.UUID) (*LoginHistory, error)
	ListUserLoginHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*LoginHistory, error)
	CountUserLoginHistory(ctx context.Context, userID uuid.UUID) (int, error)

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
	GetDashboardConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*DashboardConfig, error)
	GetDashboardConfigsByUser(ctx context.Context, userID uuid.UUID) ([]*DashboardConfig, error)
	GetDashboardConfigByID(ctx context.Context, configID uuid.UUID) (*DashboardConfig, error)
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
	GetTotalDowntimeMinutesSince(ctx context.Context, since time.Time) (int, error)
	UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*Incident, error)
	ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*Incident, error)

	// Provider operations
	CreateProvider(ctx context.Context, provider *Provider) error
	GetProviderByID(ctx context.Context, providerID string) (*Provider, error)
	GetProviderByUserAndType(ctx context.Context, userID uuid.UUID, providerType string) (*Provider, error)
	GetProvidersByUser(ctx context.Context, userID uuid.UUID) ([]*Provider, error)
	ListAllProviders(ctx context.Context) ([]*Provider, error)
	ListProviderSettings(ctx context.Context) ([]*ProviderSettings, error)
	GetProviderSettings(ctx context.Context, provider string) (*ProviderSettings, error)
	SetProviderDisabled(ctx context.Context, provider string, disabled bool, reason, disabledBy string) error
	UpdateProviderStatus(ctx context.Context, providerID string, status string) error
	UpdateProvider(ctx context.Context, providerID string, updates map[string]interface{}) (*Provider, error)
	UpdateProviderLastUsed(ctx context.Context, providerID string) error
	GetStaleProviders(ctx context.Context, since time.Time) ([]*Provider, error)
	ShareProviderWithTeam(ctx context.Context, providerID string, teamID string) error
	DeleteProvider(ctx context.Context, providerID string, userID uuid.UUID) error

	// Encryption operations
	EncryptField(ctx context.Context, value string) (string, error)
	DecryptField(ctx context.Context, value string) (string, error)

	// Team invite operations
	CreateTeamInvite(invite *TeamInvite) error
	GetTeamInviteByToken(ctx context.Context, token string) (*TeamInvite, error)
	GetTeamInvitesByTeam(ctx context.Context, teamID uuid.UUID) ([]*TeamInvite, error)
	UpdateTeamInviteStatus(ctx context.Context, inviteID uuid.UUID, status string) error
	GetTeamByUserID(ctx context.Context, userID uuid.UUID) (*Team, error)
	IsTeamAdmin(ctx context.Context, userID uuid.UUID, teamID string) (bool, error)

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
	GetUserSkills(ctx context.Context, userID uuid.UUID) ([]*UserSkill, error)
	AddUserSkill(skill *UserSkill) error
	RemoveUserSkill(skillID uuid.UUID) error

	// User achievements operations
	GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]*UserAchievement, error)
	GetAchievementBySlug(ctx context.Context, slug string) (*Achievement, error)
	ListAchievements(ctx context.Context, ) ([]*Achievement, error)
	AwardAchievement(userID, achievementID uuid.UUID, metadata map[string]interface{}) error
	UpdateAchievementProgress(userAchievementID uuid.UUID, progress int, isCompleted bool) error

	// User activity operations
	GetUserActivity(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserActivity, error)
	CreateUserActivity(ctx context.Context, activity *UserActivity) error
	// GetUserContributionDailyCounts aggregates profile contribution events per UTC day:
	// user_activity rows plus registry function publishes (owner), since the given instant.
	GetUserContributionDailyCounts(ctx context.Context, userID uuid.UUID, since time.Time) (map[string]int64, error)

	// User analytics operations
	GetUserExecutionStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	// GetUserProfileStats returns authoritative counts for profile UI (not limited by registry list pagination).
	GetUserProfileStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserTrustBreakdown(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserPopularFunctions(ctx context.Context, userID uuid.UUID, limit int) ([]map[string]interface{}, error)
	GetUserGeographicStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	GetUserDeviceStats(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
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
	CreatePendingNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string, confirmationToken string) (*NewsletterSubscriber, error)
	GetNewsletterSubscriberByEmail(ctx context.Context, email string) (*NewsletterSubscriber, error)
	GetNewsletterSubscriberByID(ctx context.Context, id uuid.UUID) (*NewsletterSubscriber, error)
	ListNewsletterSubscribers(ctx context.Context, status string, limit, offset int) ([]NewsletterSubscriber, int64, error)
	GetActiveNewsletterSubscribers(ctx context.Context) ([]NewsletterSubscriber, error)
	UnsubscribeNewsletterSubscriber(ctx context.Context, email string) error
	MarkNewsletterSubscriberBounced(ctx context.Context, email string) error
	ConfirmNewsletterSubscription(ctx context.Context, email string) error
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
	SearchTeamMemoriesFallback(ctx context.Context, tenantID, teamID uuid.UUID, query string, memoryType, category string, limit int) ([]*TeamMemorySearchResult, error)
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

	// AI Chat operations
	CreateChatSession(ctx context.Context, sess *AIChatSession) (*AIChatSession, error)
	GetChatSessionByID(ctx context.Context, id uuid.UUID) (*AIChatSession, error)
	ListChatSessions(ctx context.Context, userID uuid.UUID, opts ListAIChatSessionsOpts) ([]*AIChatSession, int, error)
	ListChatMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*AIChatMessage, error)
	CreateChatMessage(ctx context.Context, msg *AIChatMessage) (*AIChatMessage, error)
	DeleteChatSession(ctx context.Context, id uuid.UUID) error

	// Achievement operations
	ListAchievementDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*AchievementDefinition, error)
	GetAchievementProgress(ctx context.Context, employeeID uuid.UUID) ([]*AchievementProgress, error)
	CheckAndAwardAchievements(ctx context.Context, employeeID uuid.UUID) error
	CreateAchievementDefinition(ctx context.Context, ach *AchievementDefinition) (*AchievementDefinition, error)
	SeedAchievementDefinitions(ctx context.Context, tenantID uuid.UUID) error

	// Employee operations
	GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*Employee, error)
	GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	GetEmployeeByFFID(ctx context.Context, ffid string) (*Employee, error)
	GetDepartmentByID(ctx context.Context, id int64) (*Department, error)
	ListEmployees(ctx context.Context, tenantID uuid.UUID, opts ListEmployeesOpts) ([]*Employee, int, error)
	CreateEmployee(ctx context.Context, emp *Employee) (*Employee, error)
	UpdateEmployee(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	UpdateClearanceLevel(ctx context.Context, employeeID uuid.UUID, level int) error
	GetEmployeeSkills(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeSkill, error)
	AddEmployeeSkill(ctx context.Context, skill *EmployeeSkill) (*EmployeeSkill, error)
	GetEmployeeCertifications(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeCertification, error)
	AddEmployeeCertification(ctx context.Context, cert *EmployeeCertification) (*EmployeeCertification, error)
	GetEmployeeAchievements(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeAchievement, error)
	ListProjects(ctx context.Context, tenantID uuid.UUID, opts ListProjectsOpts) ([]*Project, int, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*Project, error)
	CreateProject(ctx context.Context, proj *Project) (*Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	ListTasks(ctx context.Context, tenantID uuid.UUID, opts ListTasksOpts) ([]*Task, int, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error)
	CreateTask(ctx context.Context, task *Task) (*Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*TaskComment, error)
	CreateTaskComment(ctx context.Context, comment *TaskComment) (*TaskComment, error)
	ListTimeEntries(ctx context.Context, employeeID uuid.UUID, opts ListTimeEntriesOpts) ([]*TimeEntry, int, error)
	CreateTimeEntry(ctx context.Context, entry *TimeEntry) (*TimeEntry, error)
	GetTimeEntryByID(ctx context.Context, id uuid.UUID) (*TimeEntry, error)
	UpdateTimeEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteTimeEntry(ctx context.Context, id uuid.UUID) error
	ListPTORequests(ctx context.Context, employeeID uuid.UUID, opts ListPTORequestsOpts) ([]*PTORequest, int, error)
	CreatePTORequest(ctx context.Context, req *PTORequest) (*PTORequest, error)
	GetPTORequestByID(ctx context.Context, id uuid.UUID) (*PTORequest, error)
	UpdatePTORequestStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID, notes *string) error
	GetGoalTree(ctx context.Context, tenantID uuid.UUID) ([]*PerformanceGoal, error)
	GetIdentityCard(ctx context.Context, employeeID uuid.UUID) (*IdentityCard, error)
	UpdateGoalCascade(ctx context.Context, id uuid.UUID, parentGoalID *uuid.UUID, goalLevel, cascadeVisibility string) error

	// Email account operations
	CreateEmailAccount(ctx context.Context, ea *EmailAccount) (*EmailAccount, error)
	ListEmailAccounts(ctx context.Context, tenantID uuid.UUID, opts ListEmailAccountsOpts) ([]*EmailAccount, int, error)
	UpdateEmailAccount(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// FWOS event operations
	CreateFWOSEvent(ctx context.Context, ev *FWOSEvent) (*FWOSEvent, error)
	ListFWOSEvents(ctx context.Context, tenantID uuid.UUID, opts ListFWOSEventsOpts) ([]*FWOSEvent, int, error)

	// Feature flag operations
	CreateFeatureFlag(ctx context.Context, ff *FeatureFlag) (*FeatureFlag, error)
	GetFeatureFlagByKey(ctx context.Context, tenantID uuid.UUID, key string) (*FeatureFlag, error)
	ListFeatureFlags(ctx context.Context, tenantID uuid.UUID, opts ListFeatureFlagsOpts) ([]*FeatureFlag, int, error)
	UpdateFeatureFlag(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// Feedback round operations
	CreateFeedbackRound(ctx context.Context, fr *FeedbackRound) (*FeedbackRound, error)
	GetFeedbackRoundByID(ctx context.Context, id uuid.UUID) (*FeedbackRound, error)
	ListFeedbackRounds(ctx context.Context, tenantID uuid.UUID, opts ListFeedbackRoundsOpts) ([]*FeedbackRound, int, error)
	UpdateFeedbackRound(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreateFeedbackRoundResponse(ctx context.Context, r *FeedbackRoundResponse) (*FeedbackRoundResponse, error)
	GetFeedbackRoundResults(ctx context.Context, roundID uuid.UUID) ([]map[string]interface{}, error)

	// Incident operations
	CreateFWOSIncident(ctx context.Context, inc *FWOSIncident) (*FWOSIncident, error)
	GetFWOSIncidentByID(ctx context.Context, id uuid.UUID) (*FWOSIncident, error)
	ListFWOSIncidents(ctx context.Context, tenantID uuid.UUID, opts ListIncidentsOpts) ([]*FWOSIncident, int, error)
	UpdateFWOSIncident(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreateIncidentEvent(ctx context.Context, ev *IncidentEvent) (*IncidentEvent, error)
	ListIncidentEvents(ctx context.Context, incidentID uuid.UUID, limit, offset int) ([]*IncidentEvent, error)
	AddIncidentResponder(ctx context.Context, resp *IncidentResponder) (*IncidentResponder, error)
	ListIncidentResponders(ctx context.Context, incidentID uuid.UUID) ([]*IncidentResponder, error)
	CreatePostmortem(ctx context.Context, pm *Postmortem) (*Postmortem, error)
	GetPostmortemByIncident(ctx context.Context, incidentID uuid.UUID) (*Postmortem, error)

	// Innovation grant operations
	CreateInnovationGrant(ctx context.Context, grant *InnovationGrant) (*InnovationGrant, error)
	GetInnovationGrantByID(ctx context.Context, id uuid.UUID) (*InnovationGrant, error)
	ListInnovationGrants(ctx context.Context, tenantID uuid.UUID, opts ListInnovationGrantsOpts) ([]*InnovationGrant, int, error)
	UpdateInnovationGrant(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	GetInnovationGrantVoteByVoter(ctx context.Context, grantID, voterID uuid.UUID) (*InnovationGrantVote, error)
	CreateInnovationGrantVote(ctx context.Context, vote *InnovationGrantVote) (*InnovationGrantVote, error)

	// Knowledge article operations
	CreateKnowledgeArticle(ctx context.Context, article *KnowledgeArticle) (*KnowledgeArticle, error)
	GetKnowledgeArticleBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*KnowledgeArticle, error)
	ListKnowledgeArticles(ctx context.Context, tenantID uuid.UUID, opts ListKnowledgeOpts) ([]*KnowledgeArticle, int, error)
	UpdateKnowledgeArticle(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	SearchKnowledgeArticles(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*KnowledgeArticle, error)

	// Learning operations
	CreateLearningCourse(ctx context.Context, course *LearningCourse) (*LearningCourse, error)
	GetLearningCourseByID(ctx context.Context, id uuid.UUID) (*LearningCourse, error)
	ListLearningCourses(ctx context.Context, tenantID uuid.UUID) ([]*LearningCourse, error)
	GetEmployeeLearning(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeLearning, error)
	EnrollCourse(ctx context.Context, el *EmployeeLearning) (*EmployeeLearning, error)
	UpdateLearningProgress(ctx context.Context, id int64, updates map[string]interface{}) error

	// Lifecycle operations
	CreateLifecycleEvent(ctx context.Context, ev *LifecycleEvent) (*LifecycleEvent, error)
	ListLifecycleEvents(ctx context.Context, employeeID uuid.UUID, opts ListLifecycleEventsOpts) ([]*LifecycleEvent, int, error)
	CreateLifecycleWorkflow(ctx context.Context, wf *LifecycleWorkflow) (*LifecycleWorkflow, error)
	ListLifecycleWorkflows(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*LifecycleWorkflow, int, error)
	GetLifecycleWorkflowByID(ctx context.Context, id uuid.UUID) (*LifecycleWorkflow, error)
	GetLifecycleWorkflowInstance(ctx context.Context, id uuid.UUID) (*LifecycleWorkflowInstance, error)
	UpdateLifecycleWorkflowInstance(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// Living memory operations
	CreateLivingMemoryEntry(ctx context.Context, e *LivingMemoryEntry) (*LivingMemoryEntry, error)
	GetLivingMemoryEntry(ctx context.Context, id uuid.UUID) (*LivingMemoryEntry, error)
	ListLivingMemoryEntries(ctx context.Context, tenantID uuid.UUID, opts ListLivingMemoryOpts) ([]*LivingMemoryEntry, int, error)
	SearchLivingMemory(ctx context.Context, tenantID uuid.UUID, opts SearchLivingMemoryOpts) ([]*LivingMemoryEntry, error)
	IncrementLivingMemoryViewCount(ctx context.Context, id uuid.UUID) error

	// Skills graph operations
	CreateSkillsGraphEntry(ctx context.Context, s *SkillsGraph) (*SkillsGraph, error)
	GetSkillsGraph(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*SkillsGraph, int, error)
	GetSkillGaps(ctx context.Context, tenantID uuid.UUID, limit int) ([]*SkillsGraph, error)
	CalculateSkillsGraph(ctx context.Context, tenantID uuid.UUID) ([]*SkillsGraph, error)

	// Team health operations
	CreateTeamHealthMetric(ctx context.Context, m *TeamHealthMetric) (*TeamHealthMetric, error)
	GetTeamHealthMetrics(ctx context.Context, tenantID uuid.UUID, opts ListTeamHealthOpts) ([]*TeamHealthMetric, int, error)
	CalculateTeamHealth(ctx context.Context, tenantID uuid.UUID, departmentID int64) (*TeamHealthMetric, error)
	GetLatestTeamHealth(ctx context.Context, tenantID uuid.UUID, departmentID int64) (*TeamHealthMetric, error)

	// SSO provisioning operations
	CreateSSOProvisioningConfig(ctx context.Context, cfg *SSOProvisioningConfig) (*SSOProvisioningConfig, error)
	GetSSOProvisioningConfigByID(ctx context.Context, id uuid.UUID) (*SSOProvisioningConfig, error)
	ListSSOProvisioningConfigs(ctx context.Context, tenantID uuid.UUID, opts ListSSOProvisioningConfigsOpts) ([]*SSOProvisioningConfig, int, error)
	UpdateSSOProvisioningConfig(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreateSSOProvisioningLog(ctx context.Context, log *SSOProvisioningLog) (*SSOProvisioningLog, error)
	ListSSOProvisioningLogs(ctx context.Context, configID uuid.UUID, opts ListSSOProvisioningLogsOpts) ([]*SSOProvisioningLog, int, error)

	// Marketplace operations
	CreateMarketplaceOpportunity(ctx context.Context, opp *MarketplaceOpportunity) (*MarketplaceOpportunity, error)
	GetMarketplaceOpportunityByID(ctx context.Context, id uuid.UUID) (*MarketplaceOpportunity, error)
	ListMarketplaceOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListMarketplaceOpportunitiesOpts) ([]*MarketplaceOpportunity, int, error)
	UpdateMarketplaceOpportunity(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreateMarketplaceApplication(ctx context.Context, app *MarketplaceApplication) (*MarketplaceApplication, error)
	GetMarketplaceApplicationByID(ctx context.Context, id uuid.UUID) (*MarketplaceApplication, error)
	UpdateMarketplaceApplicationStatus(ctx context.Context, id uuid.UUID, status string) error

	// Mentorship operations
	CreateMentorshipMatch(ctx context.Context, match *MentorshipMatch) (*MentorshipMatch, error)
	GetMentorshipMatchByID(ctx context.Context, id uuid.UUID) (*MentorshipMatch, error)
	ListMentorshipMatches(ctx context.Context, employeeID uuid.UUID, opts ListMentorshipMatchesOpts) ([]*MentorshipMatch, int, error)
	UpdateMentorshipMatch(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// Mission control operations
	CreateMissionControlSnapshot(ctx context.Context, s *MissionControlSnapshot) (*MissionControlSnapshot, error)
	GetLatestMissionControlSnapshot(ctx context.Context, tenantID uuid.UUID) (*MissionControlSnapshot, error)
	ListMissionControlSnapshots(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*MissionControlSnapshot, int, error)
	GenerateMissionControlSnapshot(ctx context.Context, tenantID uuid.UUID) (*MissionControlSnapshot, error)

	// Notification operations
	ListNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]*FWOSNotification, int, error)
	CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int, error)
	MarkNotificationRead(ctx context.Context, id uuid.UUID) error
	MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error

	// Org chart operations
	GetOrgChart(ctx context.Context, tenantID uuid.UUID) ([]*Employee, error)
	GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]*Employee, error)

	// Org chart import operations
	CreateOrgChartImport(ctx context.Context, imp *OrgChartImport) (*OrgChartImport, error)
	GetOrgChartImportByID(ctx context.Context, id uuid.UUID) (*OrgChartImport, error)

	// Package registry operations
	GetPackageByID(ctx context.Context, id uuid.UUID) (*PackageRegistry, error)
	ListPackages(ctx context.Context, tenantID uuid.UUID, opts ListPackageRegistryOpts) ([]*PackageRegistry, int, error)
	CreatePackage(ctx context.Context, pkg *PackageRegistry) (*PackageRegistry, error)
	ListPackageVersions(ctx context.Context, packageID uuid.UUID, opts ListPackageVersionsOpts) ([]*PackageVersion, int, error)

	// Performance operations
	CreatePerformanceGoal(ctx context.Context, goal *PerformanceGoal) (*PerformanceGoal, error)
	GetPerformanceGoalByID(ctx context.Context, id uuid.UUID) (*PerformanceGoal, error)
	ListPerformanceGoals(ctx context.Context, employeeID uuid.UUID, opts ListPerformanceGoalsOpts) ([]*PerformanceGoal, int, error)
	UpdatePerformanceGoal(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreatePerformanceReview(ctx context.Context, rev *PerformanceReview) (*PerformanceReview, error)
	GetPerformanceReviewByID(ctx context.Context, id uuid.UUID) (*PerformanceReview, error)
	ListPerformanceReviews(ctx context.Context, employeeID uuid.UUID, opts ListPerformanceReviewsOpts) ([]*PerformanceReview, int, error)
	UpdatePerformanceReview(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreatePeerFeedback(ctx context.Context, fb *PeerFeedback) (*PeerFeedback, error)
	ListPeerFeedback(ctx context.Context, toEmployeeID uuid.UUID, limit, offset int) ([]*PeerFeedback, error)

	// Reputation operations
	CreateReputationScore(ctx context.Context, s *ReputationScore) (*ReputationScore, error)
	CalculateReputation(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID) ([]*ReputationScore, error)
	GetReputationScores(ctx context.Context, employeeID uuid.UUID) ([]*ReputationScore, error)
	GetReputationLeaderboard(ctx context.Context, tenantID uuid.UUID, category string, limit, offset int) ([]*ReputationScore, int, error)
	GetReputationHistory(ctx context.Context, employeeID uuid.UUID, category string) ([]*ReputationHistory, error)

	// Push notification operations
	CreatePushSubscription(ctx context.Context, ps *PushSubscription) (*PushSubscription, error)
	UpsertNotificationPreference(ctx context.Context, pref *NotificationPreference) (*NotificationPreference, error)
	ListNotificationPreferences(ctx context.Context, userID uuid.UUID, opts ListNotificationPreferencesOpts) ([]*NotificationPreference, int, error)

	// Wallet pass operations
	CreateWalletPass(ctx context.Context, wp *WalletPass) (*WalletPass, error)
	GetWalletPassByQRToken(ctx context.Context, qrToken string) (*WalletPass, error)
	ListWalletPasses(ctx context.Context, employeeID uuid.UUID, opts ListWalletPassesOpts) ([]*WalletPass, int, error)
	UpdateWalletPass(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	CreateWalletPassTemplate(ctx context.Context, t *WalletPassTemplate) (*WalletPassTemplate, error)
	ListWalletPassTemplates(ctx context.Context, tenantID uuid.UUID, opts ListWalletPassTemplatesOpts) ([]*WalletPassTemplate, int, error)

	// Badge operations
	CreateDigitalBadge(ctx context.Context, b *DigitalBadge) (*DigitalBadge, error)
	ListDigitalBadges(ctx context.Context, tenantID uuid.UUID, opts ListBadgesOpts) ([]*DigitalBadge, int, error)
	GetDigitalBadgeBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*DigitalBadge, error)
	AwardEmployeeBadge(ctx context.Context, eb *EmployeeBadge) (*EmployeeBadge, error)
	GetEmployeeBadges(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeBadge, error)
	RevokeEmployeeBadge(ctx context.Context, employeeID, badgeID uuid.UUID) error

	// Certificate operations
	CreateCertificateKey(ctx context.Context, ck *CertificateKey) (*CertificateKey, error)
	GetCertificateKeysByCertID(ctx context.Context, certificateID uuid.UUID) ([]*CertificateKey, error)
	ListCertificates(ctx context.Context, employeeID uuid.UUID, opts ListCertificatesOpts) ([]*EmployeeCertificate, int, error)
	IssueCertificate(ctx context.Context, cert *EmployeeCertificate) (*EmployeeCertificate, error)
	RevokeCertificate(ctx context.Context, id uuid.UUID, reason string) error
	GetCertificateBySerial(ctx context.Context, serial string) (*EmployeeCertificate, error)

	// Compensation operations
	GetActiveCompensation(ctx context.Context, employeeID uuid.UUID) (*CompensationRecord, error)
	CreateCompensationRecord(ctx context.Context, rec *CompensationRecord) (*CompensationRecord, error)
	LogCompensationAccess(ctx context.Context, log *CompensationAccessLog) error
	ListEquityGrants(ctx context.Context, employeeID uuid.UUID) ([]*EquityGrant, error)

	// Data classification operations
	CreateDataClassification(ctx context.Context, dc *DataClassification) (*DataClassification, error)
	GetDataClassification(ctx context.Context, tenantID uuid.UUID, resourceType string, resourceID uuid.UUID) (*DataClassification, error)
	ListDataClassifications(ctx context.Context, tenantID uuid.UUID, opts ListDataClassificationsOpts) ([]*DataClassification, int, error)

	// Device operations
	ListDevices(ctx context.Context, tenantID uuid.UUID, opts ListDevicesOpts) ([]*Device, int, error)
	CreateDevice(ctx context.Context, d *Device) (*Device, error)
	UpdateDevice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	GetDeviceByID(ctx context.Context, id uuid.UUID) (*Device, error)

	// Document operations
	ListDocuments(ctx context.Context, tenantID uuid.UUID, opts ListDocumentsOpts) ([]*Document, int, error)
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*Document, error)
	CreateDocument(ctx context.Context, doc *Document) (*Document, error)
	UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	IncrementDocumentViewCount(ctx context.Context, id uuid.UUID) error
	CreateDocumentShare(ctx context.Context, share *DocumentShare) (*DocumentShare, error)
	ListDocumentShares(ctx context.Context, documentID uuid.UUID) ([]*DocumentShare, error)
	CreateDocumentSignature(ctx context.Context, ds *DocumentSignature) (*DocumentSignature, error)
	GetDocumentSignatureByID(ctx context.Context, id uuid.UUID) (*DocumentSignature, error)
	UpdateDocumentSignature(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// Career operations
	ListCareerPaths(ctx context.Context, tenantID uuid.UUID, opts ListCareerPathsOpts) ([]*CareerPath, int, error)
	GetCareerPathByID(ctx context.Context, id uuid.UUID) (*CareerPath, error)
	GetEmployeeCareerProgressByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeCareerProgress, error)
	CreateEmployeeCareerProgress(ctx context.Context, prog *EmployeeCareerProgress) (*EmployeeCareerProgress, error)
	GetCareerTimeline(ctx context.Context, employeeID uuid.UUID) ([]*CareerTimelineEvent, error)
	CreateCareerTimelineEvent(ctx context.Context, ev *CareerTimelineEvent) (*CareerTimelineEvent, error)

	// Founder operations
	GetFounderCount(ctx context.Context) (int, error)
	GetFounderRank(ctx context.Context, userID uuid.UUID) (int, error)
	AssignFounderStatus(ctx context.Context, userID uuid.UUID) (int, error)
	GetFoundersLeaderboard(ctx context.Context, limit int) ([]*User, int, error)
	GetFounderEarlyAccessFeatures(ctx context.Context) ([]*FounderEarlyAccessFeature, error)
	GetFounderEarlyAccessFeatureBySlug(ctx context.Context, slug string) (*FounderEarlyAccessFeature, error)
	GetUserFounderEarlyAccess(ctx context.Context, userID uuid.UUID) ([]*FounderEarlyAccess, error)
	HasUserClaimedEarlyAccess(ctx context.Context, userID uuid.UUID, slug string) (bool, error)
	ClaimFounderEarlyAccess(ctx context.Context, userID uuid.UUID, feature *FounderEarlyAccessFeature) error
	ListActiveFounderVotes(ctx context.Context) ([]*FounderVote, error)
	ListFounderVotes(ctx context.Context) ([]*FounderVote, error)
	GetFounderVote(ctx context.Context, voteID uuid.UUID) (*FounderVote, error)
	GetFounderVoteResponse(ctx context.Context, voteID, userID uuid.UUID) (*FounderVoteResponse, error)
	GetFounderVoteResults(ctx context.Context, voteID uuid.UUID) (map[string]int, int, error)
	CastFounderVote(ctx context.Context, voteID, userID uuid.UUID, optionID string) error
	CreateFounderVote(ctx context.Context, vote *FounderVote) error
	UpdateFounderVote(ctx context.Context, voteID uuid.UUID, updates map[string]interface{}) error
	DeleteFounderVote(ctx context.Context, voteID uuid.UUID) error
}

// Untyped re-exports of typed string constants from the types package.
// These allow callers in other packages to use `storage.ExportStatusPending`
// as a bare identifier (matching the pre-refactor API).
const (
	ExportStatusPending    = types.ExportStatusPending
	ExportStatusProcessing = types.ExportStatusProcessing
	ExportStatusCompleted  = types.ExportStatusCompleted
	ExportStatusFailed     = types.ExportStatusFailed
	ExportStatusExpired    = types.ExportStatusExpired

	ExportFormatCSV     = types.ExportFormatCSV
	ExportFormatJSON    = types.ExportFormatJSON
	ExportFormatParquet = types.ExportFormatParquet
	ExportFormatExcel   = types.ExportFormatExcel
)

const (
	Plan           = "plan"
	PlanFree       = "free"
	PlanPro        = "pro"
	PlanTeam       = "team"
	PlanEnterprise = "enterprise"
)

const (
	RecognitionMethodPointInTime = "point_in_time"
	RecognitionMethodOverTime     = "over_time"
	ContractAssetType             = "contract_asset"
)

const (
	BillingSystemQuickBooks = types.BillingSystemQuickBooks
	BillingSystemXero       = types.BillingSystemXero

	MFAModeOptional = types.MFAModeOptional
	MFAModeRequired = types.MFAModeRequired
	MFAModeEnforced = types.MFAModeEnforced

	SSOProviderNone = types.SSOProviderNone
	SSOProviderSAML = types.SSOProviderSAML
	SSOProviderOIDC = types.SSOProviderOIDC

	RoleTeamOwner  = types.RoleTeamOwner
	RoleTeamAdmin  = types.RoleTeamAdmin
	RoleTeamMember = types.RoleTeamMember
	RoleTeamViewer = types.RoleTeamViewer

	CertCredentialStatusActive    = types.CertCredentialStatusActive
	CertCredentialStatusExpired   = types.CertCredentialStatusExpired
	CertCredentialStatusRevoked   = types.CertCredentialStatusRevoked
	CertCredentialStatusSuspended = types.CertCredentialStatusSuspended

	CertExamStatusDraft     = types.CertExamStatusDraft
	CertExamStatusScheduled = types.CertExamStatusScheduled
	CertExamStatusSubmitted = types.CertExamStatusSubmitted
	CertExamStatusGrading   = types.CertExamStatusGrading

	ReferralStatusPending   = types.ReferralStatusPending
	ReferralStatusConverted = types.ReferralStatusConverted
	ReferralStatusQualified = types.ReferralStatusQualified
	ReferralStatusCanceled  = types.ReferralStatusCanceled

	CommissionStatusPending  = types.CommissionStatusPending
	CommissionStatusApproved = types.CommissionStatusApproved

	StripeSyncStatusPending   = types.StripeSyncStatusPending
	StripeSyncStatusIgnored   = types.StripeSyncStatusIgnored

	MembershipStatusActive    = types.MembershipStatusActive
	MembershipStatusSuspended = types.MembershipStatusSuspended
	MembershipStatusInvited   = types.MembershipStatusInvited
)

var (
	IsValidOAuthProvider = types.IsValidOAuthProvider
	IsValidRole          = types.IsValidRole
)
