package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/resend/resend-go/v2"
	"github.com/sirupsen/logrus"
)

const (
	maxRetries    = 3
	retryBaseWait = 1 * time.Second
)

// ResendConfig holds configuration for the Resend email service
type ResendConfig struct {
	APIKey       string
	FromEmail    string
	FromName     string
	BaseURL      string
	AuthURL      string
	ReplyToEmail string
}

// ResendService implements the Service interface using the Resend API
type ResendService struct {
	client *resend.Client
	config ResendConfig
}

// NewResendService creates a new Resend email service
func NewResendService(config ResendConfig) *ResendService {
	return &ResendService{
		client: resend.NewClient(config.APIKey),
		config: config,
	}
}

// isRetryableError reports whether the error is a transient one (e.g. 429, 5xx) that callers may retry.
func (s *ResendService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, resend.ErrRateLimit) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "500") || strings.Contains(msg, "503")
}

func (s *ResendService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "Verify Your Email Address - FunctionFly"
	verifyBase := s.config.AuthURL
	if verifyBase == "" {
		verifyBase = s.config.BaseURL
	}
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", verifyBase, *user.VerificationToken)

	tpl := VerificationEmailTemplate(verificationURL)
	return s.SendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Reset Your Password - FunctionFly"
	resetBase := s.config.AuthURL
	if resetBase == "" {
		resetBase = s.config.BaseURL
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", resetBase, resetToken)

	tpl := PasswordResetEmailTemplate(resetURL)
	return s.SendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	breachType := withDefault(extractString(breachDetails, "type"), "Data Breach")
	detectionTime := withDefault(extractString(breachDetails, "detectionTime"), time.Now().Format(time.RFC3339))
	affectedUsers := extractInt(breachDetails, "affectedUsers")
	riskLevel := defaultRiskLevel(extractString(breachDetails, "riskLevel"))

	tpl := BreachNotificationTemplate(breachType, detectionTime, affectedUsers, riskLevel)
	return s.SendEmailToMultiple(recipients, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendEmail(to, subject, textBody, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = s.config.ReplyToEmail
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.sendWithRetry(ctx, params)
}

func (s *ResendService) SendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      to,
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = s.config.ReplyToEmail
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return s.sendWithRetry(ctx, params)
}

// sendWithRetry sends an email with exponential backoff retry for transient errors.
func (s *ResendService) sendWithRetry(ctx context.Context, params *resend.SendEmailRequest) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := s.client.Emails.SendWithContext(ctx, params)
		if err == nil {
			return nil
		}

		lastErr = err

		// Log detailed error for debugging
		logrus.WithError(err).
			WithField("attempt", attempt).
			WithField("to", params.To).
			WithField("from", params.From).
			WithField("subject", params.Subject).
			Warn("Email send attempt failed")

		if !s.isRetryableError(err) {
			return fmt.Errorf("email send failed: %w", err)
		}

		if attempt < maxRetries {
			wait := retryBaseWait * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	logrus.WithError(lastErr).
		WithField("to", params.To).
		WithField("from", params.From).
		WithField("subject", params.Subject).
		WithField("attempts", maxRetries).
		Error("Email send failed after all retries")
	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, lastErr)
}

func (s *ResendService) SendWaitlistConfirmationEmail(email string) error {
	subject := "You're on the list — FunctionFly"
	tpl := WaitlistEmailTemplate()
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendNewsletterSubscriptionConfirmation(email, name string) error {
	subject := "You're subscribed — FunctionFly"
	tpl := NewsletterSubscriptionTemplate(name)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendNewsletterCampaign(to []string, subject, previewText, htmlContent string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified for newsletter campaign")
	}

	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      to,
		Subject: subject,
		Html:    htmlContent,
		Text:    previewText,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = s.config.ReplyToEmail
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return s.sendWithRetry(ctx, params)
}

func (s *ResendService) SendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Sign in to FunctionFly — Magic Link"
	authBase := baseURL(s.config.AuthURL, s.config.BaseURL)
	magicURL := buildEmailURL(authBase, "token", token)
	expiryMinutes := int(expiry.Minutes())

	tpl := MagicLinkEmailTemplate(magicURL, expiryMinutes)
	return s.SendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error {
	subject := "Welcome to FunctionFly — Complete Your Sign Up"
	authBase := baseURL(s.config.AuthURL, s.config.BaseURL)
	magicURL := buildEmailURL(authBase, "token", token)
	expiryMinutes := int(expiry.Minutes())

	tpl := MagicLinkSignupEmailTemplate(magicURL, expiryMinutes)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// ValidateConfiguration checks if the Resend API key is configured and valid
func (s *ResendService) ValidateConfiguration() error {
	if s.config.APIKey == "" {
		return fmt.Errorf("Resend API key not configured")
	}
	if s.config.FromEmail == "" {
		return fmt.Errorf("from email not configured - set FROM_EMAIL environment variable with a domain verified in Resend")
	}
	if s.config.AuthURL == "" && s.config.BaseURL == "" {
		return fmt.Errorf("auth URL not configured - set AUTH_URL environment variable for magic links")
	}

	// Test API key by fetching domains (lightweight check)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	domains, err := s.client.Domains.ListWithContext(ctx)
	if err != nil {
		return fmt.Errorf("Resend API key validation failed: %w - check RESEND_API_KEY is correct", err)
	}

	logrus.WithField("domain_count", len(domains.Data)).
		WithField("from_email", s.config.FromEmail).
		Info("Resend API validated successfully")

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// NEW TEMPLATE METHODS (Velocity Orange Brand)
// ═══════════════════════════════════════════════════════════════════════════════

// Security & Alerts
func (s *ResendService) SendNewDeviceLoginAlert(email, deviceInfo, location, ipAddress string, loginTime time.Time) error {
	subject := "New Device Login Detected — FunctionFly"
	tpl := NewDeviceLoginTemplate(deviceInfo, location, ipAddress, loginTime)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendPasswordChangedConfirmation(email string, changedAt time.Time, deviceInfo string) error {
	subject := "Password Changed — FunctionFly"
	tpl := PasswordChangedTemplate(changedAt, deviceInfo)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendSecurityAlert(email, alertType, description, actionRequired string) error {
	subject := fmt.Sprintf("Security Alert: %s — FunctionFly", alertType)
	tpl := SecurityAlertTemplate(alertType, description, actionRequired)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendAgentWalletLowBalance(email string, balance, threshold float64, walletID string) error {
	subject := "Wallet Balance Low — FunctionFly"
	tpl := AgentWalletLowBalanceTemplate(balance, threshold, walletID)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Marketing & Engagement
func (s *ResendService) SendWaitlistInviteEmail(email, inviteCode, signupURL string, expiresAt time.Time) error {
	subject := "You're In! FunctionFly Early Access"
	tpl := WaitlistInviteTemplate(inviteCode, signupURL, expiresAt)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Platform Operations
func (s *ResendService) SendFunctionDeploySuccess(email, functionName, version, runtime string, deployTime time.Time) error {
	subject := fmt.Sprintf("Deployment Successful: %s — FunctionFly", functionName)
	tpl := FunctionDeploySuccessTemplate(functionName, version, runtime, deployTime)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendFunctionDeployFailure(email, functionName, errorMsg string, retryCount int) error {
	subject := fmt.Sprintf("Deployment Failed: %s — FunctionFly", functionName)
	tpl := FunctionDeployFailureTemplate(functionName, errorMsg, retryCount)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendUsageExportReady(email, exportID, downloadURL string, expiresAt time.Time, sizeBytes int64) error {
	subject := "Your Export is Ready — FunctionFly"
	tpl := UsageExportReadyTemplate(exportID, downloadURL, expiresAt, sizeBytes)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Billing & Payments
func (s *ResendService) SendPaymentFailed(email string, amount float64, dueDate time.Time, retryURL string) error {
	subject := "Payment Failed — FunctionFly"
	tpl := PaymentFailedTemplate(amount, dueDate, retryURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendInvoiceReady(email, period string, amount float64, invoiceURL string) error {
	subject := fmt.Sprintf("Invoice Ready: %s — FunctionFly", period)
	tpl := InvoiceReadyTemplate(period, amount, invoiceURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Trust & Compliance
func (s *ResendService) SendTrustRevocationAlert(email, functionName, reason string, revokedAt time.Time) error {
	subject := fmt.Sprintf("Trust Status Revoked: %s — FunctionFly", functionName)
	tpl := TrustRevocationTemplate(functionName, reason, revokedAt)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendDataRequestConfirmation(email, requestType, requestID string, estimatedCompletion time.Time) error {
	subject := "Data Request Received — FunctionFly"
	tpl := DataRequestConfirmationTemplate(requestType, requestID, estimatedCompletion)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendAccountDeletionScheduled(email string, deletionDate time.Time, cancelURL string) error {
	subject := "Account Deletion Scheduled — FunctionFly"
	tpl := AccountDeletionScheduledTemplate(deletionDate, cancelURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Collaboration
func (s *ResendService) SendTeamInvite(email, orgName, invitedBy, role, acceptURL string) error {
	subject := fmt.Sprintf("You're Invited to Join %s — FunctionFly", orgName)
	tpl := TeamInviteTemplate(orgName, invitedBy, role, acceptURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendBundleWelcomeEmail(email, bundleName, dashboardURL string) error {
	subject := fmt.Sprintf("Welcome to %s — FunctionFly", bundleName)
	tpl := BundleWelcomeTemplate(bundleName, dashboardURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendInviteEmail(email, inviterName, orgName, role, acceptURL string) error {
	subject := fmt.Sprintf("You're Invited to Join %s — FunctionFly", orgName)
	tpl := InviteEmailTemplate(inviterName, orgName, role, acceptURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendVaultSecretShared(email, secretName, sharedBy, accessLevel, viewURL string) error {
	subject := "Secret Shared with You — FunctionFly Vault"
	tpl := VaultSecretSharedTemplate(secretName, sharedBy, accessLevel, viewURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

// Maintenance & Reminders
func (s *ResendService) SendAPIKeyRotationReminder(email, keyName, keyID string, expiresAt time.Time, rotationURL string) error {
	subject := "API Key Expiring Soon — FunctionFly"
	tpl := KeyRotationReminderTemplate(keyName, keyID, expiresAt, rotationURL)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}

func (s *ResendService) SendMaintenanceNotice(email string, windowStart, windowEnd time.Time, affectedServices []string) error {
	subject := "Scheduled Maintenance Notice — FunctionFly"
	tpl := MaintenanceNoticeTemplate(windowStart, windowEnd, affectedServices)
	return s.SendEmail(email, subject, tpl.Text, tpl.HTML)
}
