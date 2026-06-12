package email

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// SMTPService implements the Service interface using SMTP
type SMTPService struct {
	config Config
}

// NewSMTPService creates a new SMTP email service
func NewSMTPService(config Config) *SMTPService {
	return &SMTPService{config: config}
}

// SendVerificationEmail sends an email verification email to a user
func (s *SMTPService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "Verify Your Email Address - FunctionFly"
	verifyBase := baseURL(s.config.AuthURL, s.config.BaseURL)
	verificationURL := buildEmailURL(verifyBase, "token", *user.VerificationToken)

	tpl := VerificationEmailTemplate(verificationURL)
	return s.sendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

// SendPasswordResetEmail sends a password reset email to a user
func (s *SMTPService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Reset Your Password - FunctionFly"
	resetBase := baseURL(s.config.AuthURL, s.config.BaseURL)
	resetURL := buildEmailURL(resetBase, "token", resetToken)

	tpl := PasswordResetEmailTemplate(resetURL)
	return s.sendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

// SendBreachNotification sends a GDPR breach notification email
func (s *SMTPService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	breachType := withDefault(extractString(breachDetails, "type"), "Data Breach")
	detectionTime := withDefault(extractString(breachDetails, "detectionTime"), time.Now().Format(time.RFC3339))
	affectedUsers := extractInt(breachDetails, "affectedUsers")
	riskLevel := defaultRiskLevel(extractString(breachDetails, "riskLevel"))

	tpl := BreachNotificationTemplate(breachType, detectionTime, affectedUsers, riskLevel)
	return s.sendEmailToMultiple(recipients, subject, tpl.Text, tpl.HTML)
}

// SendWaitlistConfirmationEmail sends a waitlist confirmation email
func (s *SMTPService) SendWaitlistConfirmationEmail(email string) error {
	subject := "You're on the list — FunctionFly"
	tpl := WaitlistEmailTemplate()
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendNewsletterSubscriptionConfirmation sends a newsletter subscription confirmation email
func (s *SMTPService) SendNewsletterSubscriptionConfirmation(email, name string) error {
	subject := "You're subscribed — FunctionFly"
	tpl := NewsletterSubscriptionTemplate(name)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendNewsletterConfirmationEmail sends a newsletter confirmation email with confirmation link
func (s *SMTPService) SendNewsletterConfirmationEmail(email, name, confirmationURL string) error {
	subject := "Confirm Your Newsletter Subscription — FunctionFly"
	tpl := NewsletterConfirmationTemplate(name, confirmationURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendNewsletterCampaign sends a newsletter campaign email to multiple recipients
func (s *SMTPService) SendNewsletterCampaign(to []string, subject, previewText, htmlContent string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified for newsletter campaign")
	}
	return s.sendEmailToMultiple(to, subject, previewText, htmlContent)
}

// SendMagicLinkEmail sends a magic link email to an existing user
func (s *SMTPService) SendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Sign in to FunctionFly — Magic Link"
	authBase := baseURL(s.config.AuthURL, s.config.BaseURL)
	magicURL := buildEmailURL(authBase, "token", token)
	expiryMinutes := int(expiry.Minutes())

	tpl := MagicLinkEmailTemplate(magicURL, expiryMinutes)
	return s.sendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

// SendMagicLinkSignupEmail sends a magic link email for new user signup
func (s *SMTPService) SendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error {
	subject := "Welcome to FunctionFly — Complete Your Sign Up"
	authBase := baseURL(s.config.AuthURL, s.config.BaseURL)
	magicURL := buildEmailURL(authBase, "token", token)
	expiryMinutes := int(expiry.Minutes())

	tpl := MagicLinkSignupEmailTemplate(magicURL, expiryMinutes)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendEmail sends a generic email with the given subject and body
func (s *SMTPService) SendEmail(to, subject, textBody, htmlBody string) error {
	return s.sendEmail(to, subject, textBody, htmlBody)
}

// ValidateConfiguration checks if the SMTP email service is properly configured
func (s *SMTPService) ValidateConfiguration() error {
	if s.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	if s.config.SMTPPort == 0 {
		return fmt.Errorf("SMTP port not configured")
	}
	if s.config.FromEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("SMTP server unreachable: %w", err)
	}
	conn.Close()

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	defer client.Close()

	if err := client.Hello("functionfly-validation"); err != nil {
		return fmt.Errorf("SMTP hello failed: %w", err)
	}

	if s.config.SMTPUsername != "" && s.config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Quit(); err != nil {
		fmt.Printf("Warning: SMTP QUIT failed: %v\n", err)
	}

	return nil
}

// sendEmail sends an email using SMTP
func (s *SMTPService) sendEmail(to, subject, textBody, htmlBody string) error {
	msg := buildMessage(to, subject, s.config.FromName, s.config.FromEmail, textBody, htmlBody)
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return sendSMTP(addr, s.config.SMTPUsername, s.config.SMTPPassword, s.config.FromEmail, []string{to}, msg)
}

// sendEmailToMultiple sends an email to multiple recipients
func (s *SMTPService) sendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	msg := buildMessageMultiple(to, subject, s.config.FromName, s.config.FromEmail, textBody, htmlBody)
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return sendSMTP(addr, s.config.SMTPUsername, s.config.SMTPPassword, s.config.FromEmail, to, msg)
}

// ═══════════════════════════════════════════════════════════════════════════════
// NEW TEMPLATE METHODS (Velocity Orange Brand)
// ═══════════════════════════════════════════════════════════════════════════════

// Security & Alerts
func (s *SMTPService) SendNewDeviceLoginAlert(email, deviceInfo, location, ipAddress string, loginTime time.Time) error {
	subject := "New Device Login Detected — FunctionFly"
	tpl := NewDeviceLoginTemplate(deviceInfo, location, ipAddress, loginTime)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendPasswordChangedConfirmation(email string, changedAt time.Time, deviceInfo string) error {
	subject := "Password Changed — FunctionFly"
	tpl := PasswordChangedTemplate(changedAt, deviceInfo)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendSecurityAlert(email, alertType, description, actionRequired string) error {
	subject := fmt.Sprintf("Security Alert: %s — FunctionFly", alertType)
	tpl := SecurityAlertTemplate(alertType, description, actionRequired)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendAgentWalletLowBalance(email string, balance, threshold float64, walletID string) error {
	subject := "Wallet Balance Low — FunctionFly"
	tpl := AgentWalletLowBalanceTemplate(balance, threshold, walletID)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Marketing & Engagement
func (s *SMTPService) SendWaitlistInviteEmail(email, inviteCode, signupURL string, expiresAt time.Time) error {
	subject := "You're In! FunctionFly Early Access"
	tpl := WaitlistInviteTemplate(inviteCode, signupURL, expiresAt)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Platform Operations
func (s *SMTPService) SendFunctionDeploySuccess(email, functionName, version, runtime string, deployTime time.Time) error {
	subject := fmt.Sprintf("Deployment Successful: %s — FunctionFly", functionName)
	tpl := FunctionDeploySuccessTemplate(functionName, version, runtime, deployTime)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendFunctionDeployFailure(email, functionName, errorMsg string, retryCount int) error {
	subject := fmt.Sprintf("Deployment Failed: %s — FunctionFly", functionName)
	tpl := FunctionDeployFailureTemplate(functionName, errorMsg, retryCount)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendUsageExportReady(email, exportID, downloadURL string, expiresAt time.Time, sizeBytes int64) error {
	subject := "Your Export is Ready — FunctionFly"
	tpl := UsageExportReadyTemplate(exportID, downloadURL, expiresAt, sizeBytes)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendExecutionFailed(email, functionName, version, errorMsg string, failedAt time.Time) error {
	subject := fmt.Sprintf("Execution Failed: %s — FunctionFly", functionName)
	tpl := ExecutionFailedTemplate(functionName, version, errorMsg, failedAt)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendRateLimitExceeded(email string, limitType string, currentUsage, limit int64, windowDescription string, upgradeURL string) error {
	subject := "Rate Limit Exceeded — FunctionFly"
	tpl := RateLimitExceededTemplate(limitType, currentUsage, limit, windowDescription, upgradeURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendFunctionDeleted(email, functionName string, deletedAt time.Time, restoreURL string) error {
	subject := fmt.Sprintf("Function Deleted: %s — FunctionFly", functionName)
	tpl := FunctionDeletedTemplate(functionName, deletedAt, restoreURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Billing & Payments
func (s *SMTPService) SendPaymentFailed(email string, amount float64, dueDate time.Time, retryURL string) error {
	subject := "Payment Failed — FunctionFly"
	tpl := PaymentFailedTemplate(amount, dueDate, retryURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendInvoiceReady(email, period string, amount float64, invoiceURL string) error {
	subject := fmt.Sprintf("Invoice Ready: %s — FunctionFly", period)
	tpl := InvoiceReadyTemplate(period, amount, invoiceURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendPaymentSuccess(email string, amount float64, description string, chargedAt time.Time, receiptURL string) error {
	subject := "Payment Successful — FunctionFly"
	tpl := PaymentSuccessTemplate(amount, description, chargedAt, receiptURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendTrialExpiring(email string, daysRemaining int, upgradeURL string) error {
	subject := "Trial Ending Soon — FunctionFly"
	tpl := TrialExpiringTemplate(daysRemaining, upgradeURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendSubscriptionChange(email, changeType, oldPlan, newPlan string, effectiveDate time.Time, manageURL string) error {
	subject := "Subscription Changed — FunctionFly"
	tpl := SubscriptionChangeTemplate(changeType, oldPlan, newPlan, effectiveDate, manageURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendUsageAlert(email string, usageType string, currentUsage, limit int64, percentageUsed int, resetDate string, upgradeURL string) error {
	subject := fmt.Sprintf("Usage Alert: %d%% — FunctionFly", percentageUsed)
	tpl := UsageAlertTemplate(usageType, currentUsage, limit, percentageUsed, resetDate, upgradeURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Trust & Compliance
func (s *SMTPService) SendTrustRevocationAlert(email, functionName, reason string, revokedAt time.Time) error {
	subject := fmt.Sprintf("Trust Status Revoked: %s — FunctionFly", functionName)
	tpl := TrustRevocationTemplate(functionName, reason, revokedAt)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendDataRequestConfirmation(email, requestType, requestID string, estimatedCompletion time.Time) error {
	subject := "Data Request Received — FunctionFly"
	tpl := DataRequestConfirmationTemplate(requestType, requestID, estimatedCompletion)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendAccountDeletionScheduled(email string, deletionDate time.Time, cancelURL string) error {
	subject := "Account Deletion Scheduled — FunctionFly"
	tpl := AccountDeletionScheduledTemplate(deletionDate, cancelURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Collaboration
func (s *SMTPService) SendTeamInvite(email, orgName, invitedBy, role, acceptURL string) error {
	subject := fmt.Sprintf("You're Invited to Join %s — FunctionFly", orgName)
	tpl := TeamInviteTemplate(orgName, invitedBy, role, acceptURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendBundleWelcomeEmail(email, bundleName, dashboardURL string) error {
	subject := fmt.Sprintf("Welcome to %s — FunctionFly", bundleName)
	tpl := BundleWelcomeTemplate(bundleName, dashboardURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendInviteEmail(email, inviterName, orgName, role, acceptURL string) error {
	subject := fmt.Sprintf("You're Invited to Join %s — FunctionFly", orgName)
	tpl := InviteEmailTemplate(inviterName, orgName, role, acceptURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendVaultSecretShared(email, secretName, sharedBy, accessLevel, viewURL string) error {
	subject := "Secret Shared with You — FunctionFly Vault"
	tpl := VaultSecretSharedTemplate(secretName, sharedBy, accessLevel, viewURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Maintenance & Reminders
func (s *SMTPService) SendAPIKeyRotationReminder(email, keyName, keyID string, expiresAt time.Time, rotationURL string) error {
	subject := "API Key Expiring Soon — FunctionFly"
	tpl := KeyRotationReminderTemplate(keyName, keyID, expiresAt, rotationURL)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendMaintenanceNotice(email string, windowStart, windowEnd time.Time, affectedServices []string) error {
	subject := "Scheduled Maintenance Notice — FunctionFly"
	tpl := MaintenanceNoticeTemplate(windowStart, windowEnd, affectedServices)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendReferralInvite(email, referrerName, inviteURL, rewardDescription string, expiresAt time.Time) error {
	subject := "You've Been Invited! — FunctionFly"
	tpl := ReferralInviteTemplate(referrerName, inviteURL, rewardDescription, expiresAt)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *SMTPService) SendReferralReward(email, rewardType, rewardValue, claimURL string, expiryDate time.Time) error {
	subject := "You Earned a Reward! — FunctionFly"
	tpl := ReferralRewardTemplate(rewardType, rewardValue, claimURL, expiryDate)
	return s.sendEmail(email, subject, tpl.Text, tpl.HTML)
}
