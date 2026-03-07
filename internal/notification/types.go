package notification

import (
	"context"

	"github.com/functionfly/functionfly/internal/storage"
)

// Channel is the interface for notification channels
type Channel interface {
	Name() string
	Send(ctx context.Context, notification *Notification, user *storage.User) error
	IsConfigured() bool
}

// Trigger is the interface for event triggers
type Trigger interface {
	Name() string
	ShouldTrigger(event interface{}) bool
	BuildNotification(event interface{}) (*Notification, error)
}

// Priority levels
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Status values
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusSent       = "sent"
	StatusFailed     = "failed"
	StatusRead       = "read"
	StatusDelivered  = "delivered"
)

// Channel types
const (
	ChannelEmail   = "email"
	ChannelInApp   = "in_app"
	ChannelWebhook = "webhook"
	ChannelPush    = "push"
)

// Categories
const (
	CategorySystem     = "system"
	CategorySecurity   = "security"
	CategoryBilling    = "billing"
	CategoryDeployment = "deployment"
	CategoryFunction   = "function"
	CategoryTeam       = "team"
	CategoryRegistry   = "registry"
)

// Frequencies
const (
	FrequencyImmediate    = "immediate"
	FrequencyDigestDaily  = "digest_daily"
	FrequencyDigestWeekly = "digest_weekly"
)

// Notification types
const (
	// Deployment notifications
	TypeDeploymentSuccess = "deployment.success"
	TypeDeploymentFailed  = "deployment.failed"
	TypeDeploymentStarted = "deployment.started"

	// Billing notifications
	TypeBillingInvoiceGenerated = "billing.invoice_generated"
	TypeBillingPaymentFailed    = "billing.payment_failed"
	TypeBillingPaymentSuccess   = "billing.payment_success"
	TypeBillingSubscriptionExpiring = "billing.subscription_expiring"

	// Security notifications
	TypeSecurityPasswordChanged  = "security.password_changed"
	TypeSecurityMFAEnabled       = "security.mfa_enabled"
	TypeSecurityNewDeviceLogin   = "security.new_device_login"
	TypeSecuritySuspiciousActivity = "security.suspicious_activity"

	// Team notifications
	TypeTeamInvitation        = "team.invitation"
	TypeTeamMemberAdded       = "team.member_added"
	TypeTeamMemberRemoved     = "team.member_removed"
	TypeTeamRoleChanged       = "team.role_changed"

	// System notifications
	TypeSystemMaintenance     = "system.maintenance"
	TypeSystemUpdateAvailable = "system.update_available"
	TypeSystemAnnouncement    = "system.announcement"
	TypeWelcome               = "system.welcome"

	// Function notifications
	TypeFunctionPublished     = "function.published"
	TypeFunctionUpdated       = "function.updated"
	TypeFunctionDeprecated    = "function.deprecated"
	TypeFunctionExecuted     = "function.executed"
	TypeFunctionError        = "function.error"

	// Registry notifications
	TypeRegistryNewVersion    = "registry.new_version"
	TypeRegistryFunctionRated = "registry.function_rated"

	// Follow notifications
	TypeFollowStarted          = "follow.started"
	TypeFollowFunctionUpdated = "follow.function_updated"
	TypeFollowNewVersion      = "follow.function_new_version"
)

// Analytics status values
const (
	AnalyticsStatusDelivered = "delivered"
	AnalyticsStatusFailed    = "failed"
	AnalyticsStatusBounced   = "bounced"
	AnalyticsStatusOpened    = "opened"
	AnalyticsStatusClicked   = "clicked"
)
