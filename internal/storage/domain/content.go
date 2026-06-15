package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// ContentRepository handles content management (blog, changelog, etc.)
type ContentRepository interface {
	// Changelog
	CreateChangelogEntry(ctx context.Context, entry *types.ChangelogEntry) (*types.ChangelogEntry, error)
	GetChangelogEntryByID(ctx context.Context, id uuid.UUID) (*types.ChangelogEntry, error)
	GetChangelogEntryByVersion(ctx context.Context, version string) (*types.ChangelogEntry, error)
	ListChangelogEntries(ctx context.Context, limit, offset int, publishedOnly bool) ([]*types.ChangelogEntry, error)
	UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.ChangelogEntry, error)
	DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error
	CreateChangelogChange(ctx context.Context, change *types.ChangelogChange) (*types.ChangelogChange, error)
	UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.ChangelogChange, error)
	DeleteChangelogChange(ctx context.Context, id uuid.UUID) error

	// Blog posts
	CreateBlogPost(ctx context.Context, post *types.BlogPost) (*types.BlogPost, error)
	GetBlogPostByID(ctx context.Context, id uuid.UUID) (*types.BlogPost, error)
	GetBlogPostBySlug(ctx context.Context, slug string) (*types.BlogPost, error)
	ListBlogPosts(ctx context.Context, limit, offset int, publishedOnly bool, tagFilter []string) ([]*types.BlogPost, error)
	UpdateBlogPost(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.BlogPost, error)
	DeleteBlogPost(ctx context.Context, id uuid.UUID) error

	// Blog categories
	ListBlogCategories(ctx context.Context) ([]*types.BlogCategory, error)
	CreateBlogCategory(ctx context.Context, c *types.BlogCategory) (*types.BlogCategory, error)
	GetBlogCategoryByID(ctx context.Context, id uuid.UUID) (*types.BlogCategory, error)
	GetBlogCategoryBySlug(ctx context.Context, slug string) (*types.BlogCategory, error)
	UpdateBlogCategory(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.BlogCategory, error)
	DeleteBlogCategory(ctx context.Context, id uuid.UUID) error

	// Blog authors
	ListBlogAuthors(ctx context.Context) ([]*types.BlogAuthor, error)
	CreateBlogAuthor(ctx context.Context, a *types.BlogAuthor) (*types.BlogAuthor, error)
	GetBlogAuthorByID(ctx context.Context, id uuid.UUID) (*types.BlogAuthor, error)
	GetBlogAuthorBySlug(ctx context.Context, slug string) (*types.BlogAuthor, error)
	UpdateBlogAuthor(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.BlogAuthor, error)
	DeleteBlogAuthor(ctx context.Context, id uuid.UUID) error

	// Blog settings
	GetBlogSettings(ctx context.Context) (*types.BlogSettings, error)
	UpdateBlogSettings(ctx context.Context, updates map[string]interface{}) (*types.BlogSettings, error)

	// Blog analytics
	RecordBlogPageView(ctx context.Context, view *types.BlogPageView) error
	GetBlogAnalyticsSummary(ctx context.Context, days int) (*types.BlogAnalyticsSummary, error)
	GetBlogViewsTimeSeries(ctx context.Context, days int) ([]types.BlogViewsTimeSeries, error)
	GetTopBlogPosts(ctx context.Context, days, limit int) ([]types.TopBlogPost, error)
}

// FeedbackRepository handles user feedback and support tickets
type FeedbackRepository interface {
	CreateFeedback(ctx context.Context, feedback *types.Feedback) (*types.Feedback, error)
	GetFeedbackByID(ctx context.Context, id uuid.UUID) (*types.Feedback, error)
	GetFeedbackByUser(ctx context.Context, userID *uuid.UUID, userEmail *string, limit, offset int) ([]types.Feedback, error)
	ListFeedback(ctx context.Context, limit, offset int, statusFilter *string, typeFilter *string) ([]types.Feedback, error)
	UpdateFeedbackStatus(ctx context.Context, id uuid.UUID, status string) error
	CreateFeedbackAttachment(ctx context.Context, attachment *types.FeedbackAttachment) (*types.FeedbackAttachment, error)
	GetFeedbackAttachments(ctx context.Context, feedbackID uuid.UUID) ([]types.FeedbackAttachment, error)
	GetFeedbackAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*types.FeedbackAttachment, error)
	GetFeedbackStats(ctx context.Context) (map[string]interface{}, error)
	GetFeedbackAnalytics(ctx context.Context) (map[string]interface{}, error)
}

// NotificationRepository handles user notifications
type NotificationRepository interface {
	// Waitlist
	CreateWaitlistEntry(ctx context.Context, email, name, company, useCase, source, ip, userAgent string) (*types.WaitlistEntry, error)
	GetWaitlistEntryByEmail(ctx context.Context, email string) (*types.WaitlistEntry, error)
	ListWaitlistEntries(ctx context.Context, status string, limit, offset int) ([]types.WaitlistEntryAdminList, int64, error)
	GetWaitlistStats(ctx context.Context) (*types.WaitlistStats, error)
	UpdateWaitlistEntryStatus(ctx context.Context, id uuid.UUID, status, notes string) error
	IssueInviteToWaitlistEntry(ctx context.Context, entryID, inviteCodeID uuid.UUID) error
	DeleteWaitlistEntry(ctx context.Context, id uuid.UUID) error

	// Newsletter
	CreateNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string) (*types.NewsletterSubscriber, error)
	CreatePendingNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string, confirmationToken string) (*types.NewsletterSubscriber, error)
	GetNewsletterSubscriberByEmail(ctx context.Context, email string) (*types.NewsletterSubscriber, error)
	GetNewsletterSubscriberByID(ctx context.Context, id uuid.UUID) (*types.NewsletterSubscriber, error)
	ListNewsletterSubscribers(ctx context.Context, status string, limit, offset int) ([]types.NewsletterSubscriber, int64, error)
	GetActiveNewsletterSubscribers(ctx context.Context) ([]types.NewsletterSubscriber, error)
	UnsubscribeNewsletterSubscriber(ctx context.Context, email string) error
	MarkNewsletterSubscriberBounced(ctx context.Context, email string) error
	ConfirmNewsletterSubscription(ctx context.Context, email string) error
	DeleteNewsletterSubscriber(ctx context.Context, id uuid.UUID) error
	GetNewsletterStats(ctx context.Context) (map[string]interface{}, error)

	// Campaigns
	CreateNewsletterCampaign(ctx context.Context, campaign *types.NewsletterCampaign) (*types.NewsletterCampaign, error)
	GetNewsletterCampaignByID(ctx context.Context, id uuid.UUID) (*types.NewsletterCampaign, error)
	ListNewsletterCampaigns(ctx context.Context, status string, limit, offset int) ([]types.NewsletterCampaign, int64, error)
	UpdateNewsletterCampaign(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*types.NewsletterCampaign, error)
	CreateNewsletterCampaignEmail(ctx context.Context, campaignEmail *types.NewsletterCampaignEmail) error
	UpdateNewsletterCampaignEmailStatus(ctx context.Context, id uuid.UUID, status string, emailID string) error
	GetNewsletterCampaignEmailsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]types.NewsletterCampaignEmail, error)
	UpdateCampaignStats(ctx context.Context, campaignID uuid.UUID) error
}

// EmailRepository handles email event tracking
type EmailRepository interface {
	CreateEmailEvent(ctx context.Context, event *types.EmailEvent) error
	GetEmailEvents(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*types.EmailEvent, error)
	GetPendingBounceReviews(ctx context.Context, limit, offset int) ([]*types.EmailEvent, error)
	MarkEmailEventReviewed(ctx context.Context, eventID int64, reviewedBy uuid.UUID) error
	GetEmailEventStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error)
}

// AuditRepository handles audit logging
type AuditRepository interface {
	ListAuditEvents(ctx context.Context, limit, offset int) ([]*types.AuditEvent, error)
	LogAuditEvent(ctx context.Context, event *types.AuditEvent) error
	ListAuditEventsFiltered(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*types.AuditEvent, error)
	GetAuditEventByID(ctx context.Context, id uuid.UUID) (*types.AuditEvent, error)
}

// MonitoringRepository handles system monitoring and metrics
type MonitoringRepository interface {
	InsertPerformanceMetric(ctx context.Context, metric *types.PerformanceMetric) error
	InsertAlert(ctx context.Context, alert *types.Alert) error
	InsertSystemHealthCheck(ctx context.Context, check *types.SystemHealthCheck) error
	InsertMonitoringEvent(ctx context.Context, event *types.MonitoringEvent) error
	QueryMonitoringEvents(ctx context.Context, eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*types.MonitoringEvent, error)
	UpdateAlertStatus(ctx context.Context, alert *types.Alert) error
	QueryPerformanceMetrics(ctx context.Context, metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*types.PerformanceMetric, error)
	QueryActiveAlerts(ctx context.Context, tenantID *uuid.UUID) ([]*types.Alert, error)
	QueryLatestSystemHealthChecks(ctx context.Context) (map[string]*types.SystemHealthCheck, error)
	GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error)
	StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error
	QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*types.DatabaseMetric, error)
	PgNotify(channel, payload string) error
	PgListen(ctx context.Context, channel string) error
	PgWaitForNotification(ctx context.Context) (*types.PgNotification, error)
}

// SecurityRepository handles security scanning and vulnerability tracking
type SecurityRepository interface {
	CreateSecurityScan(ctx context.Context, scan *types.SecurityScan) (*types.SecurityScan, error)
	UpdateSecurityScan(ctx context.Context, scanID uuid.UUID, updates map[string]interface{}) (*types.SecurityScan, error)
	GetSecurityScan(ctx context.Context, scanID uuid.UUID) (*types.SecurityScan, error)
	ListSecurityScans(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*types.SecurityScan, error)
	CreateVulnerability(ctx context.Context, vuln *types.Vulnerability) (*types.Vulnerability, error)
	UpdateVulnerability(ctx context.Context, vulnID uuid.UUID, updates map[string]interface{}) (*types.Vulnerability, error)
	GetVulnerabilities(ctx context.Context, filters map[string]interface{}) ([]*types.Vulnerability, error)
	GetVulnerabilityByID(ctx context.Context, vulnID uuid.UUID) (*types.Vulnerability, error)
}

// MagicLinkRepository handles passwordless authentication via magic links
type MagicLinkRepository interface {
	CreateMagicLink(ctx context.Context, email string, token string, userID *uuid.UUID, ipAddress, userAgent, redirectPath string, expiresAt time.Time) (*types.MagicLink, error)
	GetMagicLinkByToken(ctx context.Context, token string) (*types.MagicLink, error)
	MarkMagicLinkUsed(ctx context.Context, id uuid.UUID) error
	GetRecentMagicLinksByEmail(ctx context.Context, email string, since time.Time) ([]*types.MagicLink, error)
	DeleteExpiredMagicLinks(ctx context.Context) (int64, error)
}

// EmailWorkflowRepository handles tenant-isolated email workflow configurations
type EmailWorkflowRepository interface {
	CreateEmailWorkflowConfig(ctx context.Context, config *types.EmailWorkflowConfig) error
	GetEmailWorkflowConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]types.EmailWorkflowConfig, error)
	GetEmailWorkflowConfigsByBundle(ctx context.Context, tenantID uuid.UUID, bundleSlug string) ([]types.EmailWorkflowConfig, error)
	GetActiveEmailWorkflowConfigsByTenant(ctx context.Context, tenantID uuid.UUID) ([]types.EmailWorkflowConfig, error)
	GetEmailWorkflowConfigByID(ctx context.Context, id uuid.UUID) (*types.EmailWorkflowConfig, error)
	UpdateEmailWorkflowConfig(ctx context.Context, config *types.EmailWorkflowConfig) error
	DeleteEmailWorkflowConfig(ctx context.Context, id uuid.UUID) error
	CreateEmailWorkflowExecution(ctx context.Context, exec *types.EmailWorkflowExecution) error
	GetPendingEmailWorkflowExecutions(ctx context.Context, limit int) ([]types.EmailWorkflowExecution, error)
	GetEmailWorkflowExecutionsByWorkflow(ctx context.Context, workflowID uuid.UUID, limit int) ([]types.EmailWorkflowExecution, error)
	GetEmailWorkflowExecutionsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]types.EmailWorkflowExecution, error)
	UpdateEmailWorkflowExecution(ctx context.Context, exec *types.EmailWorkflowExecution) error
	MarkEmailWorkflowExecutionSent(ctx context.Context, id uuid.UUID) error
	MarkEmailWorkflowExecutionFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
	RetryFailedEmailWorkflowExecutions(ctx context.Context, maxRetries int) ([]types.EmailWorkflowExecution, error)
	CleanupOldEmailWorkflowExecutions(ctx context.Context, retentionDays int) (int64, error)
}
