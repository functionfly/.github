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
	PriorityLow     = "low"
	PriorityNormal  = "normal"
	PriorityHigh    = "high"
	PriorityUrgent  = "urgent"
	PriorityCritical = "critical"
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
	ChannelSlack   = "slack"
)

// Categories
const (
	CategorySystem        = "system"
	CategorySecurity      = "security"
	CategoryBilling       = "billing"
	CategoryDeployment    = "deployment"
	CategoryFunction      = "function"
	CategoryTeam          = "team"
	CategoryMessages      = "messages"
	CategoryRegistry      = "registry"
	CategoryFailover      = "failover"
	CategoryProvider      = "provider"
	CategoryConsciousness = "consciousness"
	CategoryPayout        = "payout"
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
	TypeBillingInvoiceGenerated      = "billing.invoice_generated"
	TypeBillingPaymentFailed         = "billing.payment_failed"
	TypeBillingPaymentSuccess        = "billing.payment_success"
	TypeBillingSubscriptionExpiring  = "billing.subscription_expiring"
	TypeBillingSubscriptionCreated   = "billing.subscription_created"
	TypeBillingWalletToppedUp        = "billing.wallet_topped_up"
	TypeBillingWalletLowBalance      = "billing.wallet_low_balance"
	TypeBillingAlert                 = "billing.alert"
	TypeBillingSpendCapWarning       = "billing.spend_cap_warning"
	TypeBillingForecastExceeded      = "billing.forecast_exceeded"
	TypeBillingUsageSpike            = "billing.usage_spike"

	// Payout notifications
	TypePayoutCompleted      = "payout.completed"
	TypePayoutFailed         = "payout.failed"
	TypePayoutCancelled      = "payout.cancelled"
	TypePayoutReversed       = "payout.reversed"
	TypePayoutScheduled      = "payout.scheduled"
	TypePayoutApprovalNeeded = "payout.approval_needed"

	// Chargeback and refund notifications
	TypeBillingDisputeCreated           = "billing.dispute_created"
	TypeBillingDisputeUpdated           = "billing.dispute_updated"
	TypeBillingDisputeResolved          = "billing.dispute_resolved"
	TypeBillingDisputeEvidenceDue       = "billing.dispute_evidence_due"
	TypeBillingRefundProcessed          = "billing.refund_processed"
	TypeBillingChargebackFundsWithdrawn = "billing.chargeback_funds_withdrawn"

	// Founder Mode (Backend-in-a-Box) notifications
	TypeFounderModeThresholdWarning  = "founder_mode.threshold_warning"
	TypeFounderModeThresholdReached  = "founder_mode.threshold_reached"
	TypeFounderModeGracePeriodEnding = "founder_mode.grace_period_ending"
	TypeFounderModeConverted         = "founder_mode.converted"

	// Security notifications
	TypeSecurityPasswordChanged    = "security.password_changed"
	TypeSecurityMFAEnabled         = "security.mfa_enabled"
	TypeSecurityNewDeviceLogin     = "security.new_device_login"
	TypeSecuritySuspiciousActivity = "security.suspicious_activity"
	TypeSecurityUsernameChanged    = "security.username_changed"

	// Team notifications
	TypeTeamInvitation     = "team.invitation"
	TypeTeamMemberAdded    = "team.member_added"
	TypeTeamMemberRemoved  = "team.member_removed"
	TypeTeamRoleChanged    = "team.role_changed"
	TypeTeamDirectMessage  = "team.direct_message"
	TypeTeamCreated        = "team.created"
	TypeTeamDeleted        = "team.deleted"
	TypeTeamInviteSent     = "team.invite_sent"
	TypeTeamInviteAccepted = "team.invite_accepted"

	// System notifications
	TypeSystemMaintenance     = "system.maintenance"
	TypeSystemUpdateAvailable = "system.update_available"
	TypeSystemAnnouncement    = "system.announcement"
	TypeWelcome               = "system.welcome"

	// Function notifications
	TypeFunctionPublished  = "function.published"
	TypeFunctionUpdated    = "function.updated"
	TypeFunctionDeprecated = "function.deprecated"
	TypeFunctionExecuted   = "function.executed"
	TypeFunctionError      = "function.error"

	// Registry notifications
	TypeRegistryNewVersion    = "registry.new_version"
	TypeRegistryFunctionRated = "registry.function_rated"

	// Failover notifications
	TypeFailoverTriggered = "failover.triggered"
	TypeFailoverResolved   = "failover.resolved"

	// Provider notifications
	TypeProviderOffline = "provider.offline"
	TypeProviderOnline  = "provider.online"
	TypeProviderDegraded = "provider.degraded"

	// Follow notifications
	TypeFollowStarted         = "follow.started"
	TypeFollowFunctionUpdated = "follow.function_updated"
	TypeFollowNewVersion      = "follow.function_new_version"

	// Function Consciousness notifications
	TypeConsciousnessInsight     = "consciousness.insight"
	TypeConsciousnessDigest      = "consciousness.digest"
	TypeConsciousnessCritical    = "consciousness.critical"
	TypeConsciousnessAutoApplied = "consciousness.auto_applied"
	TypeConsciousnessScoreChanged = "consciousness.score_changed"
)

// Analytics status values
const (
	AnalyticsStatusDelivered = "delivered"
	AnalyticsStatusFailed    = "failed"
	AnalyticsStatusBounced   = "bounced"
	AnalyticsStatusOpened    = "opened"
	AnalyticsStatusClicked   = "clicked"
)
