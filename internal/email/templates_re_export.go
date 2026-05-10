package email

import (
	"time"

	"github.com/functionfly/functionfly/internal/email/templates"
)

type EmailTemplate = templates.EmailTemplate

func buildEmailURL(base, path, token string) string {
	return templates.BuildEmailURL(base, path, token)
}

func baseURL(authURL, baseURL string) string {
	return templates.BaseURL(authURL, baseURL)
}

func withDefault(value, fallback string) string {
	return templates.WithDefault(value, fallback)
}

func defaultRiskLevel(level string) string {
	return templates.DefaultRiskLevel(level)
}

func extractString(m map[string]interface{}, key string) string {
	return templates.ExtractString(m, key)
}

func extractInt(m map[string]interface{}, key string) int {
	return templates.ExtractInt(m, key)
}

func TestBannerHTML(html string) string {
	return templates.TestBannerHTML(html)
}

func TestBannerText(text string) string {
	return templates.TestBannerText(text)
}

func VerificationEmailTemplate(verifyURL string) EmailTemplate {
	return templates.VerificationEmailTemplate(verifyURL)
}

func PasswordResetEmailTemplate(resetURL string) EmailTemplate {
	return templates.PasswordResetEmailTemplate(resetURL)
}

func MagicLinkEmailTemplate(magicURL string, expiryMinutes int) EmailTemplate {
	return templates.MagicLinkEmailTemplate(magicURL, expiryMinutes)
}

func MagicLinkSignupEmailTemplate(magicURL string, expiryMinutes int) EmailTemplate {
	return templates.MagicLinkSignupEmailTemplate(magicURL, expiryMinutes)
}

func BreachNotificationTemplate(breachType, detectionTime string, affectedUsers int, riskLevel string) EmailTemplate {
	return templates.BreachNotificationTemplate(breachType, detectionTime, affectedUsers, riskLevel)
}

func AgentWalletLowBalanceTemplate(balance float64, threshold float64, walletID string) EmailTemplate {
	return templates.AgentWalletLowBalanceTemplate(balance, threshold, walletID)
}

func NewDeviceLoginTemplate(deviceInfo, location, ipAddress string, loginTime time.Time) EmailTemplate {
	return templates.NewDeviceLoginTemplate(deviceInfo, location, ipAddress, loginTime)
}

func PasswordChangedTemplate(changedAt time.Time, deviceInfo string) EmailTemplate {
	return templates.PasswordChangedTemplate(changedAt, deviceInfo)
}

func SecurityAlertTemplate(alertType, description, actionRequired string) EmailTemplate {
	return templates.SecurityAlertTemplate(alertType, description, actionRequired)
}

func WaitlistEmailTemplate() EmailTemplate {
	return templates.WaitlistEmailTemplate()
}

func NewsletterSubscriptionTemplate(name string) EmailTemplate {
	return templates.NewsletterSubscriptionTemplate(name)
}

func WaitlistInviteTemplate(inviteCode, signupURL string, expiresAt time.Time) EmailTemplate {
	return templates.WaitlistInviteTemplate(inviteCode, signupURL, expiresAt)
}

func FunctionDeploySuccessTemplate(functionName, version, runtime string, deployTime time.Time) EmailTemplate {
	return templates.FunctionDeploySuccessTemplate(functionName, version, runtime, deployTime)
}

func FunctionDeployFailureTemplate(functionName, errorMsg string, retryCount int) EmailTemplate {
	return templates.FunctionDeployFailureTemplate(functionName, errorMsg, retryCount)
}

func UsageExportReadyTemplate(exportID, downloadURL string, expiresAt time.Time, sizeBytes int64) EmailTemplate {
	return templates.UsageExportReadyTemplate(exportID, downloadURL, expiresAt, sizeBytes)
}

func PaymentFailedTemplate(amount float64, dueDate time.Time, retryURL string) EmailTemplate {
	return templates.PaymentFailedTemplate(amount, dueDate, retryURL)
}

func InvoiceReadyTemplate(period string, amount float64, invoiceURL string) EmailTemplate {
	return templates.InvoiceReadyTemplate(period, amount, invoiceURL)
}

func PaymentSuccessTemplate(amount float64, description string, chargedAt time.Time, receiptURL string) EmailTemplate {
	return templates.PaymentSuccessTemplate(amount, description, chargedAt, receiptURL)
}

func TrialExpiringTemplate(daysRemaining int, upgradeURL string) EmailTemplate {
	return templates.TrialExpiringTemplate(daysRemaining, upgradeURL)
}

func SubscriptionChangeTemplate(changeType string, oldPlan, newPlan string, effectiveDate time.Time, manageURL string) EmailTemplate {
	return templates.SubscriptionChangeTemplate(changeType, oldPlan, newPlan, effectiveDate, manageURL)
}

func UsageAlertTemplate(usageType string, currentUsage, limit int64, percentageUsed int, resetDate string, upgradeURL string) EmailTemplate {
	return templates.UsageAlertTemplate(usageType, currentUsage, limit, percentageUsed, resetDate, upgradeURL)
}

func ExecutionFailedTemplate(functionName, version, errorMsg string, failedAt time.Time) EmailTemplate {
	return templates.ExecutionFailedTemplate(functionName, version, errorMsg, failedAt)
}

func RateLimitExceededTemplate(limitType string, currentUsage, limit int64, windowDescription string, upgradeURL string) EmailTemplate {
	return templates.RateLimitExceededTemplate(limitType, currentUsage, limit, windowDescription, upgradeURL)
}

func FunctionDeletedTemplate(functionName string, deletedAt time.Time, restoreURL string) EmailTemplate {
	return templates.FunctionDeletedTemplate(functionName, deletedAt, restoreURL)
}

func ReferralInviteTemplate(referrerName string, inviteURL string, rewardDescription string, expiresAt time.Time) EmailTemplate {
	return templates.ReferralInviteTemplate(referrerName, inviteURL, rewardDescription, expiresAt)
}

func ReferralRewardTemplate(rewardType string, rewardValue string, claimURL string, expiryDate time.Time) EmailTemplate {
	return templates.ReferralRewardTemplate(rewardType, rewardValue, claimURL, expiryDate)
}

func TrustRevocationTemplate(functionName, reason string, revokedAt time.Time) EmailTemplate {
	return templates.TrustRevocationTemplate(functionName, reason, revokedAt)
}

func DataRequestConfirmationTemplate(requestType, requestID string, estimatedCompletion time.Time) EmailTemplate {
	return templates.DataRequestConfirmationTemplate(requestType, requestID, estimatedCompletion)
}

func AccountDeletionScheduledTemplate(deletionDate time.Time, cancelURL string) EmailTemplate {
	return templates.AccountDeletionScheduledTemplate(deletionDate, cancelURL)
}

func TeamInviteTemplate(orgName, invitedBy, role string, acceptURL string) EmailTemplate {
	return templates.TeamInviteTemplate(orgName, invitedBy, role, acceptURL)
}

func VaultSecretSharedTemplate(secretName, sharedBy, accessLevel string, viewURL string) EmailTemplate {
	return templates.VaultSecretSharedTemplate(secretName, sharedBy, accessLevel, viewURL)
}

func InviteEmailTemplate(inviterName, orgName, role, acceptURL string) EmailTemplate {
	return templates.InviteEmailTemplate(inviterName, orgName, role, acceptURL)
}

func KeyRotationReminderTemplate(keyName, keyID string, expiresAt time.Time, rotationURL string) EmailTemplate {
	return templates.KeyRotationReminderTemplate(keyName, keyID, expiresAt, rotationURL)
}

func MaintenanceNoticeTemplate(windowStart, windowEnd time.Time, affectedServices []string) EmailTemplate {
	return templates.MaintenanceNoticeTemplate(windowStart, windowEnd, affectedServices)
}

func BundleWelcomeTemplate(bundleName, dashboardURL string) EmailTemplate {
	return templates.BundleWelcomeTemplate(bundleName, dashboardURL)
}

func BundleEmailWorkflows(bundleSlug string) []templates.EmailWorkflow {
	return templates.BundleEmailWorkflows(bundleSlug)
}