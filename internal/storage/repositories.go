package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository interface implementation methods

// User operations
func (db *PostgresDB) CreateUser(email, passwordHash string, tenantID uuid.UUID) (*User, error) {
	return db.userRepository.CreateUser(email, passwordHash, tenantID)
}

func (db *PostgresDB) CreateUserWithSocialAuth(email string, tenantID uuid.UUID, provider, providerID string, providerData map[string]interface{}) (*User, error) {
	return db.userRepository.CreateUserWithSocialAuth(email, tenantID, provider, providerID, providerData)
}

func (db *PostgresDB) CreateUserWithRole(ctx context.Context, user *User) (*User, error) {
	return db.userRepository.CreateUserWithRole(ctx, user)
}

func (db *PostgresDB) GetUserByEmail(email string) (*User, error) {
	return db.userRepository.GetUserByEmail(email)
}

func (db *PostgresDB) GetUserByID(userID uuid.UUID) (*User, error) {
	return db.userRepository.GetUserByID(userID)
}

func (db *PostgresDB) GetUserByUsername(username string) (*User, error) {
	return db.userRepository.GetUserByUsername(username)
}

// IsUsernameReserved checks if a username is reserved
func (db *PostgresDB) IsUsernameReserved(username string) (bool, error) {
	return db.userRepository.IsUsernameReserved(username)
}

func (db *PostgresDB) GetUserByVerificationToken(token string) (*User, error) {
	return db.userRepository.GetUserByVerificationToken(token)
}

func (db *PostgresDB) GetUserBySocialProvider(provider, providerID string) (*User, error) {
	return db.userRepository.GetUserBySocialProvider(provider, providerID)
}

func (db *PostgresDB) ListUsers() ([]*User, error) {
	return db.userRepository.ListUsers()
}

func (db *PostgresDB) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*User, error) {
	return db.userRepository.UpdateUser(ctx, userID, updates)
}

func (db *PostgresDB) UpdateUserEmailVerification(ctx context.Context, userID uuid.UUID, verified bool, token *string, expiresAt *time.Time) error {
	return db.userRepository.UpdateUserEmailVerification(ctx, userID, verified, token, expiresAt)
}

func (db *PostgresDB) UpdateUserProviderData(userID uuid.UUID, providerData map[string]interface{}) error {
	return db.userRepository.UpdateUserProviderData(userID, providerData)
}

func (db *PostgresDB) UpdateUserSettings(userID uuid.UUID, settings map[string]interface{}) error {
	return db.userRepository.UpdateUserSettings(userID, settings)
}

func (db *PostgresDB) GetUserSettings(userID uuid.UUID) (map[string]interface{}, error) {
	return db.userRepository.GetUserSettings(userID)
}

func (db *PostgresDB) UpdateUserMFA(userID uuid.UUID, secret *string, enabled bool, backupCodes []string, lastUsed *time.Time) error {
	return db.userRepository.UpdateUserMFA(userID, secret, enabled, backupCodes, lastUsed)
}

func (db *PostgresDB) UpdateUserMFAEnabled(userID uuid.UUID, enabled bool) error {
	return db.userRepository.UpdateUserMFAEnabled(userID, enabled)
}

func (db *PostgresDB) UpdateUserMFABackupCodes(userID uuid.UUID, backupCodes []string) error {
	return db.userRepository.UpdateUserMFABackupCodes(userID, backupCodes)
}

func (db *PostgresDB) UpdateUserMFALastUsed(userID uuid.UUID, lastUsed *time.Time) error {
	return db.userRepository.UpdateUserMFALastUsed(userID, lastUsed)
}

func (db *PostgresDB) VerifyPassword(userID uuid.UUID, password string) (bool, error) {
	return db.userRepository.VerifyPassword(userID, password)
}

// OAuth state (CSRF) — persisted for multi-instance OAuth flows
func (db *PostgresDB) StoreOAuthState(ctx context.Context, state string, expiresAt time.Time) error {
	row := &OAuthState{State: state, ExpiresAt: expiresAt}
	return db.GORM.WithContext(ctx).Create(row).Error
}

func (db *PostgresDB) ValidateAndConsumeOAuthState(ctx context.Context, state string) (bool, error) {
	var row OAuthState
	err := db.GORM.WithContext(ctx).Where("state = ? AND expires_at > ?", state, time.Now()).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	// Consume: delete so the state can only be used once
	if err := db.GORM.WithContext(ctx).Where("state = ?", state).Delete(&OAuthState{}).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (db *PostgresDB) DeleteExpiredOAuthStates() (int64, error) {
	result := db.GORM.Where("expires_at < ?", time.Now()).Delete(&OAuthState{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// Audit operations
func (db *PostgresDB) ListAuditEvents(limit, offset int) ([]*AuditEvent, error) {
	return db.auditRepository.ListAuditEvents(limit, offset)
}

func (db *PostgresDB) LogAuditEvent(ctx context.Context, event *AuditEvent) error {
	return db.auditRepository.LogAuditEvent(ctx, event)
}

func (db *PostgresDB) ListAuditEventsFiltered(limit, offset int, filters map[string]interface{}) ([]*AuditEvent, error) {
	return db.auditRepository.ListAuditEventsFiltered(limit, offset, filters)
}

// Tenant operations
func (db *PostgresDB) CountRoutingEventsForTenantSince(tenantID uuid.UUID, since time.Time) (int, error) {
	return db.tenantRepository.CountRoutingEventsForTenantSince(tenantID, since)
}

func (db *PostgresDB) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	return db.tenantRepository.CreateTenant(ctx, name)
}

func (db *PostgresDB) GetTenantByID(tenantID uuid.UUID) (*Tenant, error) {
	return db.tenantRepository.GetTenantByID(tenantID)
}

func (db *PostgresDB) ListTenants() ([]*Tenant, error) {
	return db.tenantRepository.ListTenants()
}

func (db *PostgresDB) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error) {
	return db.tenantRepository.UpdateTenant(ctx, tenantID, updates)
}

func (db *PostgresDB) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	return db.tenantRepository.DeleteTenant(ctx, tenantID)
}

// App operations
func (db *PostgresDB) CreateApp(name, slug string, tenantID uuid.UUID) (*App, error) {
	return db.appRepository.CreateApp(name, slug, tenantID)
}

func (db *PostgresDB) GetAppByID(id uuid.UUID) (*App, error) {
	return db.appRepository.GetAppByID(id)
}

func (db *PostgresDB) GetAppBySlug(slug string) (*App, error) {
	return db.appRepository.GetAppBySlug(slug)
}

func (db *PostgresDB) ListAppsByTenant(tenantID uuid.UUID) ([]*App, error) {
	return db.appRepository.ListAppsByTenant(tenantID)
}

// Billing operations
func (db *PostgresDB) CancelSubscription(ctx context.Context, id uuid.UUID) error {
	return db.billingRepository.CancelSubscription(ctx, id)
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

func (db *PostgresDB) ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*Invoice, error) {
	return db.billingRepository.ListInvoicesByTenant(tenantID, limit, offset)
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

// Backend operations
func (db *PostgresDB) CreateBackend(appID uuid.UUID, provider, region, url, sharedSecret string, priority *int) (*Backend, error) {
	return db.backendRepository.CreateBackend(appID, provider, region, url, sharedSecret, priority)
}

func (db *PostgresDB) ListBackendsByAppID(appID uuid.UUID) ([]*Backend, error) {
	return db.backendRepository.ListBackendsByAppID(appID)
}

func (db *PostgresDB) GetBackendByID(id uuid.UUID) (*Backend, error) {
	return db.backendRepository.GetBackendByID(id)
}

func (db *PostgresDB) GetAllEnabledBackends() ([]*Backend, error) {
	return db.backendRepository.GetAllEnabledBackends()
}

func (db *PostgresDB) ListAllBackends(ctx context.Context) ([]*Backend, error) {
	return db.backendRepository.ListAllBackends(ctx)
}

func (db *PostgresDB) UpdateBackendEnabled(ctx context.Context, backendID uuid.UUID, enabled bool) error {
	return db.backendRepository.UpdateBackendEnabled(ctx, backendID, enabled)
}

func (db *PostgresDB) InsertHealthCheck(backendID uuid.UUID, ok bool, statusCode, latencyMs int, errorMessage string) error {
	return db.backendRepository.InsertHealthCheck(backendID, ok, statusCode, latencyMs, errorMessage)
}

func (db *PostgresDB) GetRecentHealthChecks(backendID uuid.UUID, limit int) ([]*HealthCheck, error) {
	return db.backendRepository.GetRecentHealthChecks(backendID, limit)
}

func (db *PostgresDB) GetCircuitState(backendID uuid.UUID) (*CircuitState, error) {
	return db.backendRepository.GetCircuitState(backendID)
}

func (db *PostgresDB) UpdateCircuitState(state *CircuitState) error {
	return db.backendRepository.UpdateCircuitState(state)
}

func (db *PostgresDB) UpsertCircuitState(state *CircuitState) error {
	return db.backendRepository.UpsertCircuitState(state)
}

func (db *PostgresDB) InsertRoutingEvent(appID, backendID uuid.UUID, latencyMs int, outcome, requestID string) error {
	return db.backendRepository.InsertRoutingEvent(appID, backendID, latencyMs, outcome, requestID)
}

func (db *PostgresDB) GetRecentRoutingEvents(limit int, since time.Time) ([]*RoutingEvent, error) {
	return db.backendRepository.GetRecentRoutingEvents(limit, since)
}

func (db *PostgresDB) GetBackendStatusByAppID(appID uuid.UUID) ([]*BackendStatus, error) {
	return db.backendRepository.GetBackendStatusByAppID(appID)
}

// Deployment operations
func (db *PostgresDB) CreateDeployment(appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*Deployment, error) {
	return db.deploymentRepository.CreateDeployment(appID, provider, region, deploymentID, artifactKey, routes)
}

func (db *PostgresDB) UpdateDeploymentStatus(id uuid.UUID, status, message string, metadata map[string]interface{}) error {
	return db.deploymentRepository.UpdateDeploymentStatus(id, status, message, metadata)
}

func (db *PostgresDB) GetDeploymentByID(id uuid.UUID) (*Deployment, error) {
	return db.deploymentRepository.GetDeploymentByID(id)
}

func (db *PostgresDB) ListDeploymentsByAppID(appID uuid.UUID, limit int) ([]*Deployment, error) {
	return db.deploymentRepository.ListDeploymentsByAppID(appID, limit)
}

func (db *PostgresDB) GetLatestSuccessfulDeployment(appID uuid.UUID, provider string) (*Deployment, error) {
	return db.deploymentRepository.GetLatestSuccessfulDeployment(appID, provider)
}

func (db *PostgresDB) StoreDeploymentArtifact(appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*DeploymentArtifact, error) {
	return db.deploymentRepository.StoreDeploymentArtifact(appID, provider, key, contentType, checksum, size)
}

func (db *PostgresDB) GetDeploymentArtifact(key string) (*DeploymentArtifact, error) {
	return db.deploymentRepository.GetDeploymentArtifact(key)
}

// Content operations
func (db *PostgresDB) CreateChangelogEntry(ctx context.Context, entry *ChangelogEntry) (*ChangelogEntry, error) {
	return db.contentRepository.CreateChangelogEntry(ctx, entry)
}

func (db *PostgresDB) GetChangelogEntryByID(id uuid.UUID) (*ChangelogEntry, error) {
	return db.contentRepository.GetChangelogEntryByID(id)
}

func (db *PostgresDB) GetChangelogEntryByVersion(version string) (*ChangelogEntry, error) {
	return db.contentRepository.GetChangelogEntryByVersion(version)
}

func (db *PostgresDB) ListChangelogEntries(limit, offset int, publishedOnly bool) ([]*ChangelogEntry, error) {
	return db.contentRepository.ListChangelogEntries(limit, offset, publishedOnly)
}

func (db *PostgresDB) UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogEntry, error) {
	return db.contentRepository.UpdateChangelogEntry(ctx, id, updates)
}

func (db *PostgresDB) DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteChangelogEntry(ctx, id)
}

func (db *PostgresDB) CreateChangelogChange(ctx context.Context, change *ChangelogChange) (*ChangelogChange, error) {
	return db.contentRepository.CreateChangelogChange(ctx, change)
}

func (db *PostgresDB) UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*ChangelogChange, error) {
	return db.contentRepository.UpdateChangelogChange(ctx, id, updates)
}

func (db *PostgresDB) DeleteChangelogChange(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteChangelogChange(ctx, id)
}

func (db *PostgresDB) CreateBlogPost(ctx context.Context, post *BlogPost) (*BlogPost, error) {
	return db.contentRepository.CreateBlogPost(ctx, post)
}

func (db *PostgresDB) GetBlogPostByID(id uuid.UUID) (*BlogPost, error) {
	return db.contentRepository.GetBlogPostByID(id)
}

func (db *PostgresDB) GetBlogPostBySlug(slug string) (*BlogPost, error) {
	return db.contentRepository.GetBlogPostBySlug(slug)
}

func (db *PostgresDB) ListBlogPosts(limit, offset int, publishedOnly bool, tagFilter []string) ([]*BlogPost, error) {
	return db.contentRepository.ListBlogPosts(limit, offset, publishedOnly, tagFilter)
}

func (db *PostgresDB) UpdateBlogPost(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogPost, error) {
	return db.contentRepository.UpdateBlogPost(ctx, id, updates)
}

func (db *PostgresDB) DeleteBlogPost(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteBlogPost(ctx, id)
}

func (db *PostgresDB) ListBlogCategories(ctx context.Context) ([]*BlogCategory, error) {
	return db.contentRepository.ListBlogCategories(ctx)
}

func (db *PostgresDB) CreateBlogCategory(ctx context.Context, c *BlogCategory) (*BlogCategory, error) {
	return db.contentRepository.CreateBlogCategory(ctx, c)
}

func (db *PostgresDB) GetBlogCategoryByID(ctx context.Context, id uuid.UUID) (*BlogCategory, error) {
	return db.contentRepository.GetBlogCategoryByID(ctx, id)
}

func (db *PostgresDB) GetBlogCategoryBySlug(ctx context.Context, slug string) (*BlogCategory, error) {
	return db.contentRepository.GetBlogCategoryBySlug(ctx, slug)
}

func (db *PostgresDB) UpdateBlogCategory(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogCategory, error) {
	return db.contentRepository.UpdateBlogCategory(ctx, id, updates)
}

func (db *PostgresDB) DeleteBlogCategory(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteBlogCategory(ctx, id)
}

func (db *PostgresDB) ListBlogAuthors(ctx context.Context) ([]*BlogAuthor, error) {
	return db.contentRepository.ListBlogAuthors(ctx)
}

func (db *PostgresDB) CreateBlogAuthor(ctx context.Context, a *BlogAuthor) (*BlogAuthor, error) {
	return db.contentRepository.CreateBlogAuthor(ctx, a)
}

func (db *PostgresDB) GetBlogAuthorByID(ctx context.Context, id uuid.UUID) (*BlogAuthor, error) {
	return db.contentRepository.GetBlogAuthorByID(ctx, id)
}

func (db *PostgresDB) GetBlogAuthorBySlug(ctx context.Context, slug string) (*BlogAuthor, error) {
	return db.contentRepository.GetBlogAuthorBySlug(ctx, slug)
}

func (db *PostgresDB) UpdateBlogAuthor(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*BlogAuthor, error) {
	return db.contentRepository.UpdateBlogAuthor(ctx, id, updates)
}

func (db *PostgresDB) DeleteBlogAuthor(ctx context.Context, id uuid.UUID) error {
	return db.contentRepository.DeleteBlogAuthor(ctx, id)
}

// Feedback operations
func (db *PostgresDB) CreateFeedback(feedback *Feedback) (*Feedback, error) {
	return db.feedbackRepository.CreateFeedback(feedback)
}

func (db *PostgresDB) GetFeedbackByID(id uuid.UUID) (*Feedback, error) {
	return db.feedbackRepository.GetFeedbackByID(id)
}

func (db *PostgresDB) GetFeedbackByUser(userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error) {
	return db.feedbackRepository.GetFeedbackByUser(userID, userEmail, limit, offset)
}

func (db *PostgresDB) ListFeedback(limit, offset int, statusFilter *string) ([]Feedback, error) {
	return db.feedbackRepository.ListFeedback(limit, offset, statusFilter)
}

func (db *PostgresDB) UpdateFeedbackStatus(id uuid.UUID, status string) error {
	return db.feedbackRepository.UpdateFeedbackStatus(id, status)
}

func (db *PostgresDB) CreateFeedbackAttachment(attachment *FeedbackAttachment) (*FeedbackAttachment, error) {
	return db.feedbackRepository.CreateFeedbackAttachment(attachment)
}

func (db *PostgresDB) GetFeedbackAttachments(feedbackID uuid.UUID) ([]FeedbackAttachment, error) {
	return db.feedbackRepository.GetFeedbackAttachments(feedbackID)
}

func (db *PostgresDB) GetFeedbackStats() (map[string]interface{}, error) {
	return db.feedbackRepository.GetFeedbackStats()
}

func (db *PostgresDB) GetFeedbackAnalytics() (map[string]interface{}, error) {
	return db.feedbackRepository.GetFeedbackAnalytics()
}

// Monitoring operations
func (db *PostgresDB) InsertPerformanceMetric(metric *PerformanceMetric) error {
	return db.monitoringRepository.InsertPerformanceMetric(metric)
}

func (db *PostgresDB) InsertAlert(alert *Alert) error {
	return db.monitoringRepository.InsertAlert(alert)
}

func (db *PostgresDB) InsertSystemHealthCheck(check *SystemHealthCheck) error {
	return db.monitoringRepository.InsertSystemHealthCheck(check)
}

func (db *PostgresDB) InsertMonitoringEvent(event *MonitoringEvent) error {
	return db.monitoringRepository.InsertMonitoringEvent(event)
}

func (db *PostgresDB) QueryMonitoringEvents(eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*MonitoringEvent, error) {
	return db.monitoringRepository.QueryMonitoringEvents(eventType, tenantID, since, limit)
}

func (db *PostgresDB) UpdateAlertStatus(alert *Alert) error {
	return db.monitoringRepository.UpdateAlertStatus(alert)
}

func (db *PostgresDB) QueryPerformanceMetrics(metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*PerformanceMetric, error) {
	return db.monitoringRepository.QueryPerformanceMetrics(metricType, tenantID, since, limit)
}

func (db *PostgresDB) QueryActiveAlerts(tenantID *uuid.UUID) ([]*Alert, error) {
	return db.monitoringRepository.QueryActiveAlerts(tenantID)
}

func (db *PostgresDB) QueryLatestSystemHealthChecks() (map[string]*SystemHealthCheck, error) {
	return db.monitoringRepository.QueryLatestSystemHealthChecks()
}

func (db *PostgresDB) PgNotify(channel, payload string) error {
	return db.monitoringRepository.PgNotify(channel, payload)
}

func (db *PostgresDB) PgListen(ctx context.Context, channel string) error {
	return db.monitoringRepository.PgListen(ctx, channel)
}

func (db *PostgresDB) PgWaitForNotification(ctx context.Context) (*PgNotification, error) {
	return db.monitoringRepository.PgWaitForNotification(ctx)
}

func (db *PostgresDB) GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error) {
	return db.monitoringRepository.GetDatabaseHealthMetrics(ctx)
}

func (db *PostgresDB) StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error {
	return db.monitoringRepository.StoreDatabaseMetrics(ctx, metrics)
}

func (db *PostgresDB) QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*DatabaseMetric, error) {
	return db.monitoringRepository.QueryDatabaseMetrics(ctx, metricType, since, limit)
}

// Session operations
func (db *PostgresDB) CreateSession(userID uuid.UUID, sessionToken string, ipAddress, userAgent string, expiresAt time.Time) (*Session, error) {
	return db.sessionRepository.CreateSession(userID, sessionToken, ipAddress, userAgent, expiresAt)
}

func (db *PostgresDB) GetSessionByToken(sessionToken string) (*Session, error) {
	return db.sessionRepository.GetSessionByToken(sessionToken)
}

func (db *PostgresDB) GetSessionByID(sessionID uuid.UUID) (*Session, error) {
	return db.sessionRepository.GetSessionByID(sessionID)
}

func (db *PostgresDB) UpdateSessionMFAStatus(sessionToken string, mfaVerified bool) error {
	return db.sessionRepository.UpdateSessionMFAStatus(sessionToken, mfaVerified)
}

func (db *PostgresDB) UpdateSessionActivity(sessionToken string) error {
	return db.sessionRepository.UpdateSessionActivity(sessionToken)
}

func (db *PostgresDB) DeleteSession(sessionToken string) error {
	return db.sessionRepository.DeleteSession(sessionToken)
}

func (db *PostgresDB) DeleteExpiredSessions() (int64, error) {
	return db.sessionRepository.DeleteExpiredSessions()
}

func (db *PostgresDB) DeleteUserSessions(userID uuid.UUID) error {
	return db.sessionRepository.DeleteUserSessions(userID)
}

func (db *PostgresDB) DeleteSessionByID(sessionID, userID uuid.UUID) error {
	return db.sessionRepository.DeleteSessionByID(sessionID, userID)
}

func (db *PostgresDB) ListUserSessions(userID uuid.UUID) ([]*Session, error) {
	return db.sessionRepository.ListUserSessions(userID)
}

// Refresh token operations
func (db *PostgresDB) CreateRefreshToken(userID uuid.UUID, tokenHash string, ipAddress, userAgent string, expiresAt time.Time) (*RefreshToken, error) {
	return db.refreshTokenRepository.CreateRefreshToken(userID, tokenHash, ipAddress, userAgent, expiresAt)
}

func (db *PostgresDB) GetRefreshTokenByHash(tokenHash string) (*RefreshToken, error) {
	return db.refreshTokenRepository.GetRefreshTokenByHash(tokenHash)
}

func (db *PostgresDB) RevokeRefreshToken(tokenID uuid.UUID) error {
	return db.refreshTokenRepository.RevokeRefreshToken(tokenID)
}

func (db *PostgresDB) RevokeUserRefreshTokens(userID uuid.UUID) error {
	return db.refreshTokenRepository.RevokeUserRefreshTokens(userID)
}

func (db *PostgresDB) DeleteExpiredRefreshTokens() (int64, error) {
	return db.refreshTokenRepository.DeleteExpiredRefreshTokens()
}

func (db *PostgresDB) ListUserRefreshTokens(userID uuid.UUID) ([]*RefreshToken, error) {
	return db.refreshTokenRepository.ListUserRefreshTokens(userID)
}

// Login attempt operations
func (db *PostgresDB) CreateLoginAttempt(userID uuid.UUID, ipAddress, userAgent string, successful bool, lockoutUntil *time.Time) (*LoginAttempt, error) {
	return db.loginAttemptRepository.CreateLoginAttempt(userID, ipAddress, userAgent, successful, lockoutUntil)
}

func (db *PostgresDB) GetRecentFailedLoginAttempts(userID uuid.UUID, since time.Time) (int, error) {
	return db.loginAttemptRepository.GetRecentFailedLoginAttempts(userID, since)
}

func (db *PostgresDB) GetUserLockoutStatus(userID uuid.UUID) (*time.Time, error) {
	return db.loginAttemptRepository.GetUserLockoutStatus(userID)
}

func (db *PostgresDB) ClearUserLockout(userID uuid.UUID) error {
	return db.loginAttemptRepository.ClearUserLockout(userID)
}

func (db *PostgresDB) DeleteOldLoginAttempts(before time.Time) (int64, error) {
	return db.loginAttemptRepository.DeleteOldLoginAttempts(before)
}

// Auth event operations
func (db *PostgresDB) LogAuthEvent(event *AuthEvent) error {
	return db.authEventRepository.LogAuthEvent(event)
}

func (db *PostgresDB) GetAuthEventsForUser(userID uuid.UUID, limit, offset int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetAuthEventsForUser(userID, limit, offset)
}

func (db *PostgresDB) GetAuthEventsByType(eventType string, limit, offset int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetAuthEventsByType(eventType, limit, offset)
}

func (db *PostgresDB) GetRecentAuthEvents(since time.Time, limit int) ([]*AuthEvent, error) {
	return db.authEventRepository.GetRecentAuthEvents(since, limit)
}

func (db *PostgresDB) DeleteOldAuthEvents(before time.Time) (int64, error) {
	return db.authEventRepository.DeleteOldAuthEvents(before)
}

// Dashboard configuration operations
func (db *PostgresDB) CreateDashboardConfig(ctx context.Context, config *DashboardConfig) (*DashboardConfig, error) {
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	query := `
		INSERT INTO dashboard_configs (id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := db.ExecContext(ctx, query,
		config.ID, config.TenantID, config.UserID, config.ConfigType,
		config.Name, config.Config, config.IsActive, config.CreatedAt, config.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard config: %w", err)
	}

	return config, nil
}

func (db *PostgresDB) GetDashboardConfigsByTenant(tenantID uuid.UUID) ([]*DashboardConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at
		FROM dashboard_configs
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(context.Background(), query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dashboard configs: %w", err)
	}
	defer rows.Close()

	var configs []*DashboardConfig
	for rows.Next() {
		config := &DashboardConfig{}
		err := rows.Scan(
			&config.ID, &config.TenantID, &config.UserID, &config.ConfigType,
			&config.Name, &config.Config, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dashboard config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (db *PostgresDB) GetDashboardConfigsByUser(userID uuid.UUID) ([]*DashboardConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at
		FROM dashboard_configs
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user dashboard configs: %w", err)
	}
	defer rows.Close()

	var configs []*DashboardConfig
	for rows.Next() {
		config := &DashboardConfig{}
		err := rows.Scan(
			&config.ID, &config.TenantID, &config.UserID, &config.ConfigType,
			&config.Name, &config.Config, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dashboard config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (db *PostgresDB) GetDashboardConfigByID(configID uuid.UUID) (*DashboardConfig, error) {
	query := `
		SELECT id, tenant_id, user_id, config_type, name, config, is_active, created_at, updated_at
		FROM dashboard_configs
		WHERE id = $1
	`

	config := &DashboardConfig{}
	err := db.QueryRowContext(context.Background(), query, configID).Scan(
		&config.ID, &config.TenantID, &config.UserID, &config.ConfigType,
		&config.Name, &config.Config, &config.IsActive, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dashboard config not found")
		}
		return nil, fmt.Errorf("failed to get dashboard config: %w", err)
	}

	return config, nil
}

func (db *PostgresDB) UpdateDashboardConfig(ctx context.Context, configID uuid.UUID, updates map[string]interface{}) (*DashboardConfig, error) {
	// Get current config
	config, err := db.GetDashboardConfigByID(configID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		config.Name = name
	}
	if configData, ok := updates["config"].(map[string]interface{}); ok {
		config.Config = configData
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		config.IsActive = isActive
	}

	config.UpdatedAt = time.Now()

	// Update in database
	query := `
		UPDATE dashboard_configs
		SET name = $2, config = $3, is_active = $4, updated_at = $5
		WHERE id = $1
	`

	_, err = db.ExecContext(ctx, query,
		configID, config.Name, config.Config, config.IsActive, config.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update dashboard config: %w", err)
	}

	return config, nil
}

func (db *PostgresDB) DeleteDashboardConfig(ctx context.Context, configID uuid.UUID) error {
	query := `DELETE FROM dashboard_configs WHERE id = $1`

	result, err := db.ExecContext(ctx, query, configID)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dashboard config not found")
	}

	return nil
}

// Local runtime registry operations
func (db *PostgresDB) RegisterLocalRuntime(ctx context.Context, instance *LocalRuntimeInstance) (*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.RegisterLocalRuntime(ctx, instance)
}

func (db *PostgresDB) UpdateLocalRuntimeHeartbeat(ctx context.Context, runtimeID string) error {
	return db.localRuntimeRepository.UpdateLocalRuntimeHeartbeat(ctx, runtimeID)
}

func (db *PostgresDB) GetLocalRuntimeByID(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.GetLocalRuntimeByID(ctx, instanceID)
}

func (db *PostgresDB) GetLocalRuntimeByRuntimeID(ctx context.Context, runtimeID string) (*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.GetLocalRuntimeByRuntimeID(ctx, runtimeID)
}

func (db *PostgresDB) ListActiveLocalRuntimes(ctx context.Context) ([]*LocalRuntimeInstance, error) {
	return db.localRuntimeRepository.ListActiveLocalRuntimes(ctx)
}

func (db *PostgresDB) DeregisterLocalRuntime(ctx context.Context, runtimeID string) error {
	return db.localRuntimeRepository.DeregisterLocalRuntime(ctx, runtimeID)
}

func (db *PostgresDB) CleanupStaleLocalRuntimes(ctx context.Context, maxAge time.Duration) (int64, error) {
	return db.localRuntimeRepository.CleanupStaleLocalRuntimes(ctx, maxAge)
}

// Local runtime metrics operations
func (db *PostgresDB) RecordLocalRuntimeMetrics(ctx context.Context, metrics *LocalRuntimeMetric) error {
	return db.localRuntimeRepository.RecordLocalRuntimeMetrics(ctx, metrics)
}

func (db *PostgresDB) GetLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID, since time.Time, limit int) ([]*LocalRuntimeMetric, error) {
	return db.localRuntimeRepository.GetLocalRuntimeMetrics(ctx, instanceID, since, limit)
}

func (db *PostgresDB) GetLatestLocalRuntimeMetrics(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeMetric, error) {
	return db.localRuntimeRepository.GetLatestLocalRuntimeMetrics(ctx, instanceID)
}

func (db *PostgresDB) GetAggregatedLocalRuntimeMetrics(ctx context.Context, since time.Time) (map[string]interface{}, error) {
	return db.localRuntimeRepository.GetAggregatedLocalRuntimeMetrics(ctx, since)
}

// Local runtime health operations
func (db *PostgresDB) RecordLocalRuntimeHealth(ctx context.Context, health *LocalRuntimeHealth) error {
	return db.localRuntimeRepository.RecordLocalRuntimeHealth(ctx, health)
}

func (db *PostgresDB) GetLocalRuntimeHealth(ctx context.Context, instanceID uuid.UUID) (*LocalRuntimeHealth, error) {
	return db.localRuntimeRepository.GetLocalRuntimeHealth(ctx, instanceID)
}

// Function operations
func (db *PostgresDB) CreateFunction(ctx context.Context, function *FunctionConfig) (*FunctionConfig, error) {
	return db.functionRepository.CreateFunction(ctx, function)
}

func (db *PostgresDB) GetFunctionByID(ctx context.Context, functionID uuid.UUID) (*FunctionConfig, error) {
	return db.functionRepository.GetFunctionByID(ctx, functionID)
}

func (db *PostgresDB) ListFunctionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FunctionConfig, error) {
	return db.functionRepository.ListFunctionsByTenant(ctx, tenantID)
}

func (db *PostgresDB) ListAllFunctions(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status *string) ([]*FunctionConfig, int, error) {
	return db.functionRepository.ListAllFunctions(ctx, limit, offset, tenantID, status)
}

func (db *PostgresDB) UpdateFunction(ctx context.Context, functionID uuid.UUID, updates map[string]interface{}) (*FunctionConfig, error) {
	return db.functionRepository.UpdateFunction(ctx, functionID, updates)
}

func (db *PostgresDB) DeleteFunction(ctx context.Context, functionID uuid.UUID) error {
	return db.functionRepository.DeleteFunction(ctx, functionID)
}

func (db *PostgresDB) GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*FunctionConfig, error) {
	return db.functionRepository.GetFunctionByAppIDAndName(ctx, appID, name)
}

func (db *PostgresDB) GetActiveDeploymentForFunction(ctx context.Context, functionID uuid.UUID) (*FunctionDeployment, error) {
	return db.functionRepository.GetActiveDeploymentForFunction(ctx, functionID)
}

// Function deployment operations
func (db *PostgresDB) CreateFunctionDeployment(ctx context.Context, deployment *FunctionDeployment) (*FunctionDeployment, error) {
	return db.functionRepository.CreateFunctionDeployment(ctx, deployment)
}

func (db *PostgresDB) GetFunctionDeploymentByID(ctx context.Context, deploymentID uuid.UUID) (*FunctionDeployment, error) {
	return db.functionRepository.GetFunctionDeploymentByID(ctx, deploymentID)
}

func (db *PostgresDB) ListFunctionDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*FunctionDeployment, error) {
	return db.functionRepository.ListFunctionDeployments(ctx, functionID, limit)
}

func (db *PostgresDB) UpdateFunctionDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL, errorMessage *string) error {
	return db.functionRepository.UpdateFunctionDeploymentStatus(ctx, deploymentID, status, deployedURL, errorMessage)
}

// Function log operations
func (db *PostgresDB) CreateFunctionLog(ctx context.Context, log *FunctionLog) error {
	return db.functionRepository.CreateFunctionLog(ctx, log)
}

func (db *PostgresDB) GetFunctionLogs(ctx context.Context, functionID *uuid.UUID, deploymentID *uuid.UUID, limit int, since *time.Time, level *string) ([]*FunctionLog, error) {
	return db.functionRepository.GetFunctionLogs(ctx, functionID, deploymentID, limit, since, level)
}

// Dashboard aggregations
func (db *PostgresDB) GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]UsageByDay, error) {
	return db.functionRepository.GetUsageByDay(ctx, tenantID, days)
}

func (db *PostgresDB) GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]ExecutionRateByHour, error) {
	return db.functionRepository.GetExecutionRateByHour(ctx, tenantID, hours)
}

func (db *PostgresDB) GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]DashboardActivityItem, error) {
	return db.functionRepository.GetRecentActivityForTenant(ctx, tenantID, limit)
}

// Team operations
func (db *PostgresDB) CreateTeam(team *Team) error {
	return db.teamRepository.CreateTeam(team)
}

func (db *PostgresDB) GetTeamByID(teamID uuid.UUID) (*Team, error) {
	return db.teamRepository.GetTeamByID(teamID)
}

func (db *PostgresDB) GetTeamsByTenantID(tenantID uuid.UUID) ([]*Team, error) {
	return db.teamRepository.GetTeamsByTenantID(tenantID)
}

func (db *PostgresDB) UpdateTeam(team *Team) error {
	return db.teamRepository.UpdateTeam(team)
}

func (db *PostgresDB) DeleteTeam(teamID uuid.UUID) error {
	return db.teamRepository.DeleteTeam(teamID)
}

func (db *PostgresDB) AddTeamMember(membership *TeamMembership) error {
	return db.teamRepository.AddTeamMember(membership)
}

func (db *PostgresDB) UpdateTeamMember(teamID, userID uuid.UUID, role string) error {
	return db.teamRepository.UpdateTeamMember(teamID, userID, role)
}

func (db *PostgresDB) RemoveTeamMember(teamID, userID uuid.UUID) error {
	return db.teamRepository.RemoveTeamMember(teamID, userID)
}

func (db *PostgresDB) GetTeamMembership(teamID, userID uuid.UUID) (*TeamMembership, error) {
	return db.teamRepository.GetTeamMembership(teamID, userID)
}

func (db *PostgresDB) GetUserTeams(userID uuid.UUID) ([]*Team, error) {
	return db.teamRepository.GetUserTeams(userID)
}

func (db *PostgresDB) GrantTeamPermission(permission *TeamPermission) error {
	return db.teamRepository.GrantTeamPermission(permission)
}

func (db *PostgresDB) RevokeTeamPermission(teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	return db.teamRepository.RevokeTeamPermission(teamID, resourceType, resourceID)
}

func (db *PostgresDB) GetTeamPermissions(teamID uuid.UUID) ([]*TeamPermission, error) {
	return db.teamRepository.GetTeamPermissions(teamID)
}

func (db *PostgresDB) GetResourcePermissions(resourceType string, resourceID uuid.UUID) ([]*TeamPermission, error) {
	return db.teamRepository.GetResourcePermissions(resourceType, resourceID)
}

func (db *PostgresDB) CheckUserResourcePermission(userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error) {
	return db.teamRepository.CheckUserResourcePermission(userID, resourceType, resourceID, requiredPerm)
}

func (db *PostgresDB) GetUserPermissions(userID uuid.UUID, resourceType string) ([]string, error) {
	return db.teamRepository.GetUserPermissions(userID, resourceType)
}

func (db *PostgresDB) IsUserTeamOwner(userID, teamID uuid.UUID) (bool, error) {
	return db.teamRepository.IsUserTeamOwner(userID, teamID)
}

func (db *PostgresDB) IsUserTeamAdmin(userID, teamID uuid.UUID) (bool, error) {
	return db.teamRepository.IsUserTeamAdmin(userID, teamID)
}

// Provider operations
func (db *PostgresDB) CreateProvider(provider *Provider) error {
	// Encrypt the token before storing
	if provider.Token != "" {
		encryptedToken, err := db.EncryptField(provider.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt provider token: %w", err)
		}
		provider.Token = encryptedToken
	}
	return db.GORM.Create(provider).Error
}

func (db *PostgresDB) GetProviderByID(providerID string) (*Provider, error) {
	var provider Provider
	err := db.GORM.Where("id = ?", providerID).First(&provider).Error
	if err != nil {
		return nil, err
	}

	// Decrypt the token
	if provider.Token != "" {
		decryptedToken, err := db.DecryptField(provider.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt provider token: %w", err)
		}
		provider.Token = decryptedToken
	}

	return &provider, nil
}

func (db *PostgresDB) GetProviderByUserAndType(userID uuid.UUID, providerType string) (*Provider, error) {
	var provider Provider
	err := db.GORM.Where("user_id = ? AND provider = ? AND status = 'active'", userID, providerType).First(&provider).Error
	if err != nil {
		return nil, err
	}

	// Decrypt the token
	if provider.Token != "" {
		decryptedToken, err := db.DecryptField(provider.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt provider token: %w", err)
		}
		provider.Token = decryptedToken
	}

	return &provider, nil
}

func (db *PostgresDB) GetProvidersByUser(userID uuid.UUID) ([]*Provider, error) {
	var providers []*Provider
	err := db.GORM.Where("user_id = ? AND status = 'active'", userID).Find(&providers).Error
	if err != nil {
		return nil, err
	}

	// Decrypt tokens for all providers
	for _, provider := range providers {
		if provider.Token != "" {
			decryptedToken, err := db.DecryptField(provider.Token)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt provider token for %s: %w", provider.ID, err)
			}
			provider.Token = decryptedToken
		}
	}

	return providers, nil
}

func (db *PostgresDB) UpdateProviderStatus(providerID string, status string) error {
	return db.GORM.Model(&Provider{}).Where("id = ?", providerID).Update("status", status).Error
}

func (db *PostgresDB) ListAllProviders(ctx context.Context) ([]*Provider, error) {
	var providers []*Provider
	err := db.GORM.WithContext(ctx).Find(&providers).Error
	return providers, err
}

func (db *PostgresDB) UpdateProvider(ctx context.Context, providerID string, updates map[string]interface{}) (*Provider, error) {
	// Get current provider
	var provider Provider
	if err := db.GORM.WithContext(ctx).Where("id = ?", providerID).First(&provider).Error; err != nil {
		return nil, err
	}

	// Apply updates
	if status, ok := updates["status"].(string); ok {
		provider.Status = status
	}
	if isShared, ok := updates["is_shared"].(bool); ok {
		provider.IsShared = isShared
	}
	if teamID, ok := updates["team_id"].(string); ok {
		provider.TeamID = &teamID
	}

	// Update in database
	if err := db.GORM.WithContext(ctx).Save(&provider).Error; err != nil {
		return nil, err
	}

	return &provider, nil
}

func (db *PostgresDB) ShareProviderWithTeam(providerID string, teamID string) error {
	return db.GORM.Model(&Provider{}).Where("id = ?", providerID).Updates(map[string]interface{}{
		"is_shared": true,
		"team_id":   teamID,
	}).Error
}

// Team invite operations
func (db *PostgresDB) CreateTeamInvite(invite *TeamInvite) error {
	return db.GORM.Create(invite).Error
}

func (db *PostgresDB) GetTeamInviteByToken(token string) (*TeamInvite, error) {
	var invite TeamInvite
	err := db.GORM.Where("token = ? AND status = 'pending' AND expires_at > ?", token, time.Now()).First(&invite).Error
	return &invite, err
}

func (db *PostgresDB) GetTeamInvitesByTeam(teamID uuid.UUID) ([]*TeamInvite, error) {
	var invites []*TeamInvite
	err := db.GORM.Where("team_id = ?", teamID).Find(&invites).Error
	return invites, err
}

func (db *PostgresDB) UpdateTeamInviteStatus(inviteID uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == "accepted" {
		now := time.Now()
		updates["accepted_at"] = now
	}
	return db.GORM.Model(&TeamInvite{}).Where("id = ?", inviteID).Updates(updates).Error
}

func (db *PostgresDB) GetTeamByUserID(userID uuid.UUID) (*Team, error) {
	var membership TeamMembership
	err := db.GORM.Where("user_id = ?", userID).First(&membership).Error
	if err != nil {
		return nil, err
	}

	var team Team
	err = db.GORM.Where("id = ?", membership.TeamID).First(&team).Error
	return &team, err
}

func (db *PostgresDB) IsTeamAdmin(userID uuid.UUID, teamID string) (bool, error) {
	var membership TeamMembership
	err := db.GORM.Where("user_id = ? AND team_id = ? AND role = 'admin'", userID, teamID).First(&membership).Error
	return err == nil, err
}

// Follow operations

// User follows
func (db *PostgresDB) FollowUser(ctx context.Context, followerID, followedUserID uuid.UUID, reason *string, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion bool) (*UserFollow, error) {
	return db.followRepository.FollowUser(ctx, followerID, followedUserID, reason, notifyOnNewFunction, notifyOnFunctionUpdate, notifyOnNewVersion)
}

func (db *PostgresDB) UnfollowUser(ctx context.Context, followerID, followedUserID uuid.UUID) error {
	return db.followRepository.UnfollowUser(ctx, followerID, followedUserID)
}

func (db *PostgresDB) IsFollowingUser(ctx context.Context, followerID, followedUserID uuid.UUID) (bool, error) {
	return db.followRepository.IsFollowingUser(ctx, followerID, followedUserID)
}

func (db *PostgresDB) GetUserFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error) {
	return db.followRepository.GetUserFollowers(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetUserFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserFollow, int, error) {
	return db.followRepository.GetUserFollowing(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetUserFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return db.followRepository.GetUserFollowerCount(ctx, userID)
}

func (db *PostgresDB) GetUserFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return db.followRepository.GetUserFollowingCount(ctx, userID)
}

// Function follows
func (db *PostgresDB) FollowFunction(ctx context.Context, userID, functionID uuid.UUID, reason *string, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification bool) (*FunctionFollow, error) {
	return db.followRepository.FollowFunction(ctx, userID, functionID, reason, notifyOnNewVersion, notifyOnRatingChange, notifyOnTrustChange, notifyOnVerification)
}

func (db *PostgresDB) UnfollowFunction(ctx context.Context, userID, functionID uuid.UUID) error {
	return db.followRepository.UnfollowFunction(ctx, userID, functionID)
}

func (db *PostgresDB) IsFollowingFunction(ctx context.Context, userID, functionID uuid.UUID) (bool, error) {
	return db.followRepository.IsFollowingFunction(ctx, userID, functionID)
}

func (db *PostgresDB) GetFunctionFollowers(ctx context.Context, functionID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error) {
	return db.followRepository.GetFunctionFollowers(ctx, functionID, limit, offset)
}

func (db *PostgresDB) GetUserFunctionFollows(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FunctionFollow, int, error) {
	return db.followRepository.GetUserFunctionFollows(ctx, userID, limit, offset)
}

func (db *PostgresDB) GetFunctionFollowerCount(ctx context.Context, functionID uuid.UUID) (int, error) {
	return db.followRepository.GetFunctionFollowerCount(ctx, functionID)
}
