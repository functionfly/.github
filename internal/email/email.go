package email

import (
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// Service defines the interface for email operations
type Service interface {
	// Authentication & Verification
	SendVerificationEmail(user *storage.User, verificationToken string) error
	SendPasswordResetEmail(user *storage.User, resetToken string) error
	SendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error
	SendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error

	// Security & Alerts
	SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error
	SendNewDeviceLoginAlert(email, deviceInfo, location, ipAddress string, loginTime time.Time) error
	SendPasswordChangedConfirmation(email string, changedAt time.Time, deviceInfo string) error
	SendSecurityAlert(email, alertType, description, actionRequired string) error
	SendAgentWalletLowBalance(email string, balance, threshold float64, walletID string) error

	// Marketing & Engagement
	SendWaitlistConfirmationEmail(email string) error
	SendWaitlistInviteEmail(email, inviteCode, signupURL string, expiresAt time.Time) error
	SendNewsletterSubscriptionConfirmation(email, name string) error
	SendNewsletterCampaign(to []string, subject, previewText, htmlContent string) error

	// Platform Operations
	SendFunctionDeploySuccess(email, functionName, version, runtime string, deployTime time.Time) error
	SendFunctionDeployFailure(email, functionName, errorMsg string, retryCount int) error
	SendUsageExportReady(email, exportID, downloadURL string, expiresAt time.Time, sizeBytes int64) error

	// Billing & Payments
	SendPaymentFailed(email string, amount float64, dueDate time.Time, retryURL string) error
	SendInvoiceReady(email, period string, amount float64, invoiceURL string) error

	// Trust & Compliance
	SendTrustRevocationAlert(email, functionName, reason string, revokedAt time.Time) error
	SendDataRequestConfirmation(email, requestType, requestID string, estimatedCompletion time.Time) error
	SendAccountDeletionScheduled(email string, deletionDate time.Time, cancelURL string) error

	// Collaboration
	SendTeamInvite(email, orgName, invitedBy, role, acceptURL string) error
	SendBundleWelcomeEmail(email, bundleName, dashboardURL string) error
	SendInviteEmail(email, inviterName, orgName, role, acceptURL string) error
	SendVaultSecretShared(email, secretName, sharedBy, accessLevel, viewURL string) error

	// Maintenance & Reminders
	SendAPIKeyRotationReminder(email, keyName, keyID string, expiresAt time.Time, rotationURL string) error
	SendMaintenanceNotice(email string, windowStart, windowEnd time.Time, affectedServices []string) error

	// Generic
	SendEmail(to, subject, textBody, htmlBody string) error
	ValidateConfiguration() error
}

// Config holds email service configuration
type Config struct {
	Provider     string // "resend" or "smtp"
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	ResendAPIKey string
	FromEmail    string
	FromName     string
	BaseURL      string
	AuthURL      string // Auth frontend URL (e.g. https://auth.functionfly.com) for verification/reset links
	ReplyToEmail string
}

// NewService creates an email service based on configuration
func NewService(config Config) Service {
	if config.Provider == "resend" && config.ResendAPIKey != "" {
		return NewResendService(ResendConfig{
			APIKey:       config.ResendAPIKey,
			FromEmail:    config.FromEmail,
			FromName:     config.FromName,
			BaseURL:      config.BaseURL,
			AuthURL:      config.AuthURL,
			ReplyToEmail: config.ReplyToEmail,
		})
	}
	// Default to SMTP
	return NewSMTPService(Config{
		SMTPHost:     config.SMTPHost,
		SMTPPort:     config.SMTPPort,
		SMTPUsername: config.SMTPUsername,
		SMTPPassword: config.SMTPPassword,
		FromEmail:    config.FromEmail,
		FromName:     config.FromName,
		BaseURL:      config.BaseURL,
		AuthURL:      config.AuthURL,
	})
}
