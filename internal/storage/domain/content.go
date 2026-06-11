package domain

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// ContentRepository handles content management (blog, changelog, etc.)
type ContentRepository interface {
	// Changelog
	CreateChangelogEntry(ctx context.Context, entry *storage.ChangelogEntry) (*storage.ChangelogEntry, error)
	GetChangelogEntryByID(id uuid.UUID) (*storage.ChangelogEntry, error)
	GetChangelogEntryByVersion(version string) (*storage.ChangelogEntry, error)
	ListChangelogEntries(limit, offset int, publishedOnly bool) ([]*storage.ChangelogEntry, error)
	UpdateChangelogEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.ChangelogEntry, error)
	DeleteChangelogEntry(ctx context.Context, id uuid.UUID) error
	CreateChangelogChange(ctx context.Context, change *storage.ChangelogChange) (*storage.ChangelogChange, error)
	UpdateChangelogChange(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.ChangelogChange, error)
	DeleteChangelogChange(ctx context.Context, id uuid.UUID) error

	// Blog posts
	CreateBlogPost(ctx context.Context, post *storage.BlogPost) (*storage.BlogPost, error)
	GetBlogPostByID(id uuid.UUID) (*storage.BlogPost, error)
	GetBlogPostBySlug(slug string) (*storage.BlogPost, error)
	ListBlogPosts(limit, offset int, publishedOnly bool, tagFilter []string) ([]*storage.BlogPost, error)
	UpdateBlogPost(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.BlogPost, error)
	DeleteBlogPost(ctx context.Context, id uuid.UUID) error

	// Blog categories
	ListBlogCategories(ctx context.Context) ([]*storage.BlogCategory, error)
	CreateBlogCategory(ctx context.Context, c *storage.BlogCategory) (*storage.BlogCategory, error)
	GetBlogCategoryByID(ctx context.Context, id uuid.UUID) (*storage.BlogCategory, error)
	GetBlogCategoryBySlug(ctx context.Context, slug string) (*storage.BlogCategory, error)
	UpdateBlogCategory(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.BlogCategory, error)
	DeleteBlogCategory(ctx context.Context, id uuid.UUID) error

	// Blog authors
	ListBlogAuthors(ctx context.Context) ([]*storage.BlogAuthor, error)
	CreateBlogAuthor(ctx context.Context, a *storage.BlogAuthor) (*storage.BlogAuthor, error)
	GetBlogAuthorByID(ctx context.Context, id uuid.UUID) (*storage.BlogAuthor, error)
	GetBlogAuthorBySlug(ctx context.Context, slug string) (*storage.BlogAuthor, error)
	UpdateBlogAuthor(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.BlogAuthor, error)
	DeleteBlogAuthor(ctx context.Context, id uuid.UUID) error

	// Blog settings
	GetBlogSettings(ctx context.Context) (*storage.BlogSettings, error)
	UpdateBlogSettings(ctx context.Context, updates map[string]interface{}) (*storage.BlogSettings, error)

	// Blog analytics
	RecordBlogPageView(ctx context.Context, view *storage.BlogPageView) error
	GetBlogAnalyticsSummary(ctx context.Context, days int) (*storage.BlogAnalyticsSummary, error)
	GetBlogViewsTimeSeries(ctx context.Context, days int) ([]storage.BlogViewsTimeSeries, error)
	GetTopBlogPosts(ctx context.Context, days, limit int) ([]storage.TopBlogPost, error)
}

// FeedbackRepository handles user feedback and support tickets
type FeedbackRepository interface {
	CreateFeedback(feedback *storage.Feedback) (*storage.Feedback, error)
	GetFeedbackByID(id uuid.UUID) (*storage.Feedback, error)
	GetFeedbackByUser(userID *uuid.UUID, userEmail *string, limit, offset int) ([]storage.Feedback, error)
	ListFeedback(limit, offset int, statusFilter *string, typeFilter *string) ([]storage.Feedback, error)
	UpdateFeedbackStatus(id uuid.UUID, status string) error
	CreateFeedbackAttachment(attachment *storage.FeedbackAttachment) (*storage.FeedbackAttachment, error)
	GetFeedbackAttachments(feedbackID uuid.UUID) ([]storage.FeedbackAttachment, error)
	GetFeedbackAttachmentByID(attachmentID uuid.UUID) (*storage.FeedbackAttachment, error)
	GetFeedbackStats() (map[string]interface{}, error)
	GetFeedbackAnalytics() (map[string]interface{}, error)
}

// NotificationRepository handles user notifications
type NotificationRepository interface {
	// Waitlist
	CreateWaitlistEntry(ctx context.Context, email, name, company, useCase, source, ip, userAgent string) (*storage.WaitlistEntry, error)
	GetWaitlistEntryByEmail(ctx context.Context, email string) (*storage.WaitlistEntry, error)
	ListWaitlistEntries(ctx context.Context, status string, limit, offset int) ([]storage.WaitlistEntryAdminList, int64, error)
	GetWaitlistStats(ctx context.Context) (*storage.WaitlistStats, error)
	UpdateWaitlistEntryStatus(ctx context.Context, id uuid.UUID, status, notes string) error
	IssueInviteToWaitlistEntry(ctx context.Context, entryID, inviteCodeID uuid.UUID) error
	DeleteWaitlistEntry(ctx context.Context, id uuid.UUID) error

	// Newsletter
	CreateNewsletterSubscriber(ctx context.Context, email, name, source, ipAddress, userAgent string) (*storage.NewsletterSubscriber, error)
	GetNewsletterSubscriberByEmail(ctx context.Context, email string) (*storage.NewsletterSubscriber, error)
	GetNewsletterSubscriberByID(ctx context.Context, id uuid.UUID) (*storage.NewsletterSubscriber, error)
	ListNewsletterSubscribers(ctx context.Context, status string, limit, offset int) ([]storage.NewsletterSubscriber, int64, error)
	GetActiveNewsletterSubscribers(ctx context.Context) ([]storage.NewsletterSubscriber, error)
	UnsubscribeNewsletterSubscriber(ctx context.Context, email string) error
	DeleteNewsletterSubscriber(ctx context.Context, id uuid.UUID) error
	GetNewsletterStats(ctx context.Context) (map[string]interface{}, error)

	// Campaigns
	CreateNewsletterCampaign(ctx context.Context, campaign *storage.NewsletterCampaign) (*storage.NewsletterCampaign, error)
	GetNewsletterCampaignByID(ctx context.Context, id uuid.UUID) (*storage.NewsletterCampaign, error)
	ListNewsletterCampaigns(ctx context.Context, status string, limit, offset int) ([]storage.NewsletterCampaign, int64, error)
	UpdateNewsletterCampaign(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*storage.NewsletterCampaign, error)
	CreateNewsletterCampaignEmail(ctx context.Context, campaignEmail *storage.NewsletterCampaignEmail) error
	UpdateNewsletterCampaignEmailStatus(ctx context.Context, id uuid.UUID, status string, emailID string) error
	GetNewsletterCampaignEmailsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]storage.NewsletterCampaignEmail, error)
	UpdateCampaignStats(ctx context.Context, campaignID uuid.UUID) error
}

// EmailRepository handles email event tracking
type EmailRepository interface {
	CreateEmailEvent(ctx context.Context, event *storage.EmailEvent) error
	GetEmailEvents(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*storage.EmailEvent, error)
	GetPendingBounceReviews(ctx context.Context, limit, offset int) ([]*storage.EmailEvent, error)
	MarkEmailEventReviewed(ctx context.Context, eventID int64, reviewedBy uuid.UUID) error
	GetEmailEventStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error)
}

// AuditRepository handles audit logging
type AuditRepository interface {
	ListAuditEvents(limit, offset int) ([]*storage.AuditEvent, error)
	LogAuditEvent(ctx context.Context, event *storage.AuditEvent) error
	ListAuditEventsFiltered(limit, offset int, filters map[string]interface{}) ([]*storage.AuditEvent, error)
}

// MonitoringRepository handles system monitoring and metrics
type MonitoringRepository interface {
	InsertPerformanceMetric(metric *storage.PerformanceMetric) error
	InsertAlert(alert *storage.Alert) error
	InsertSystemHealthCheck(check *storage.SystemHealthCheck) error
	InsertMonitoringEvent(event *storage.MonitoringEvent) error
	QueryMonitoringEvents(eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*storage.MonitoringEvent, error)
	UpdateAlertStatus(alert *storage.Alert) error
	QueryPerformanceMetrics(metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*storage.PerformanceMetric, error)
	QueryActiveAlerts(tenantID *uuid.UUID) ([]*storage.Alert, error)
	QueryLatestSystemHealthChecks() (map[string]*storage.SystemHealthCheck, error)
	GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error)
	StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error
	QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*storage.DatabaseMetric, error)
}

// SecurityRepository handles security scanning and vulnerability tracking
type SecurityRepository interface {
	CreateSecurityScan(scan *storage.SecurityScan) (*storage.SecurityScan, error)
	UpdateSecurityScan(scanID uuid.UUID, updates map[string]interface{}) (*storage.SecurityScan, error)
	GetSecurityScan(scanID uuid.UUID) (*storage.SecurityScan, error)
	ListSecurityScans(limit, offset int, filters map[string]interface{}) ([]*storage.SecurityScan, error)
	CreateVulnerability(vuln *storage.Vulnerability) (*storage.Vulnerability, error)
	UpdateVulnerability(vulnID uuid.UUID, updates map[string]interface{}) (*storage.Vulnerability, error)
	GetVulnerabilities(filters map[string]interface{}) ([]*storage.Vulnerability, error)
	GetVulnerabilityByID(vulnID uuid.UUID) (*storage.Vulnerability, error)
}
