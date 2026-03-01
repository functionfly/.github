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
	GetUserByVerificationToken(token string) (*User, error)
	GetUserBySocialProvider(provider, providerID string) (*User, error)
	ListUsers() ([]*User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*User, error)
	UpdateUserEmailVerification(ctx context.Context, userID uuid.UUID, verified bool, token *string, expiresAt *time.Time) error
	UpdateUserProviderData(userID uuid.UUID, providerData map[string]interface{}) error
	// MFA operations
	UpdateUserMFA(userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error
	UpdateUserMFAEnabled(userID uuid.UUID, enabled bool) error
	UpdateUserMFABackupCodes(userID uuid.UUID, backupCodes []string) error
	UpdateUserMFALastUsed(userID uuid.UUID, lastUsed *time.Time) error
	VerifyPassword(userID uuid.UUID, password string) (bool, error)

	// Tenant operations
	CreateTenant(ctx context.Context, name string) (*Tenant, error)
	GetTenantByID(tenantID uuid.UUID) (*Tenant, error)
	ListTenants() ([]*Tenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error)
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error)

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

	// Billing operations
	CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error)
	ListPricingTiers() ([]*PricingTier, error)
	GetPricingTierByID(id uuid.UUID) (*PricingTier, error)
	UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*PricingTier, error)
	DeletePricingTier(ctx context.Context, id uuid.UUID) error

	CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetSubscriptionByTenantID(tenantID uuid.UUID) (*Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Subscription, error)
	CancelSubscription(ctx context.Context, id uuid.UUID) error

	CreateInvoice(ctx context.Context, invoice *Invoice) (*Invoice, error)
	ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*Invoice, error)
	GetInvoiceByID(id uuid.UUID) (*Invoice, error)
	UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Invoice, error)

	RecordUsageEvent(ctx context.Context, event *UsageEvent) error
	GetUsageByTenant(tenantID uuid.UUID, eventType string, start, end time.Time) ([]*UsageRollup, error)

	CreateCoupon(ctx context.Context, coupon *Coupon) (*Coupon, error)
	ListCoupons() ([]*Coupon, error)
	GetCouponByCode(code string) (*Coupon, error)
	RedeemCoupon(ctx context.Context, couponID, tenantID uuid.UUID, subscriptionID *uuid.UUID) (*CouponRedemption, error)

	// App operations
	CreateApp(name, slug string, tenantID uuid.UUID) (*App, error)
	GetAppByID(id uuid.UUID) (*App, error)
	GetAppBySlug(slug string) (*App, error)
	ListAppsByTenant(tenantID uuid.UUID) ([]*App, error)

	// Backend operations
	CreateBackend(appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error)
	ListBackendsByAppID(appID uuid.UUID) ([]*Backend, error)
	GetBackendByID(id uuid.UUID) (*Backend, error)
	GetAllEnabledBackends() ([]*Backend, error)

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

	// Feedback operations
	CreateFeedback(feedback *Feedback) (*Feedback, error)
	GetFeedbackByID(id uuid.UUID) (*Feedback, error)
	GetFeedbackByUser(userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error)
	ListFeedback(limit, offset int, statusFilter *string) ([]Feedback, error)
	UpdateFeedbackStatus(id uuid.UUID, status string) error
	CreateFeedbackAttachment(attachment *FeedbackAttachment) (*FeedbackAttachment, error)
	GetFeedbackAttachments(feedbackID uuid.UUID) ([]FeedbackAttachment, error)
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
	CreateSession(userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error)
	GetSessionByToken(sessionToken string) (*Session, error)
	GetSessionByID(sessionID uuid.UUID) (*Session, error)
	UpdateSessionMFAStatus(sessionToken string, mfaVerified bool) error
	UpdateSessionActivity(sessionToken string) error
	DeleteSession(sessionToken string) error
	DeleteSessionByID(sessionID, userID uuid.UUID) error
	DeleteExpiredSessions() (int64, error)
	DeleteUserSessions(userID uuid.UUID) error
	ListUserSessions(userID uuid.UUID) ([]*Session, error)

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

	// Dashboard aggregations (tenant-scoped)
	GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]UsageByDay, error)
	GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]ExecutionRateByHour, error)
	GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]DashboardActivityItem, error)

	// Incident operations
	CreateIncident(ctx context.Context, incident *Incident) (*Incident, error)
	GetIncidentByID(ctx context.Context, incidentID uuid.UUID) (*Incident, error)
	ListIncidents(ctx context.Context, limit int, offset int, status *string) ([]*Incident, error)
	UpdateIncident(ctx context.Context, incidentID uuid.UUID, updates map[string]interface{}) (*Incident, error)
	ResolveIncident(ctx context.Context, incidentID uuid.UUID) (*Incident, error)

	// Provider operations
	CreateProvider(provider *Provider) error
	GetProviderByID(providerID string) (*Provider, error)
	GetProviderByUserAndType(userID uuid.UUID, providerType string) (*Provider, error)
	GetProvidersByUser(userID uuid.UUID) ([]*Provider, error)
	UpdateProviderStatus(providerID string, status string) error
	ShareProviderWithTeam(providerID string, teamID string) error

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
}
