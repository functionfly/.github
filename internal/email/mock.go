package email

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// Ensure MockService implements Service interface
var _ Service = (*MockService)(nil)

// MockService implements the Service interface using SMTP (for testing with real email sending)
type MockService struct {
	config Config
}

// NewMockService creates a new mock email service with real SMTP functionality
func NewMockService(config Config) *MockService {
	return &MockService{config: config}
}

// SendVerificationEmail sends an email verification email to a user (with test banner)
func (m *MockService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "[TEST] Verify Your Email Address - FunctionFly"
	verifyBase := baseURL(m.config.AuthURL, m.config.BaseURL)
	verificationURL := buildEmailURL(verifyBase, "token", *user.VerificationToken)

	tpl := VerificationEmailTemplate(verificationURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)

	return m.sendEmail(user.Email, subject, text, html)
}

// SendPasswordResetEmail sends a password reset email (mock/dev version with test banner)
func (m *MockService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "[TEST] Reset Your Password - FunctionFly"
	resetBase := baseURL(m.config.AuthURL, m.config.BaseURL)
	resetURL := buildEmailURL(resetBase, "token", resetToken)

	tpl := PasswordResetEmailTemplate(resetURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)

	return m.sendEmail(user.Email, subject, text, html)
}

// SendBreachNotification sends a GDPR breach notification email (with test banner)
func (m *MockService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "[TEST] URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	breachType := withDefault(extractString(breachDetails, "type"), "Data Breach")
	detectionTime := withDefault(extractString(breachDetails, "detectionTime"), time.Now().Format(time.RFC3339))
	affectedUsers := extractInt(breachDetails, "affectedUsers")
	riskLevel := defaultRiskLevel(extractString(breachDetails, "riskLevel"))

	tpl := BreachNotificationTemplate(breachType, detectionTime, affectedUsers, riskLevel)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)

	return m.sendEmailToMultiple(recipients, subject, text, html)
}

// SendWaitlistConfirmationEmail sends a waitlist confirmation email
func (m *MockService) SendWaitlistConfirmationEmail(email string) error {
	subject := "You're on the list — FunctionFly"
	tpl := WaitlistEmailTemplate()
	return m.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendNewsletterSubscriptionConfirmation sends a newsletter subscription confirmation email
func (m *MockService) SendNewsletterSubscriptionConfirmation(email, name string) error {
	subject := "You're subscribed — FunctionFly"
	tpl := NewsletterSubscriptionTemplate(name)
	return m.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendNewsletterCampaign sends a newsletter campaign email to multiple recipients
func (m *MockService) SendNewsletterCampaign(to []string, subject, previewText, htmlContent string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified for newsletter campaign")
	}
	return m.sendEmailToMultiple(to, subject, previewText, htmlContent)
}

// SendMagicLinkEmail sends a magic link email to an existing user
func (m *MockService) SendMagicLinkEmail(user *storage.User, token string, expiry time.Duration) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Sign in to FunctionFly — Magic Link"
	authBase := baseURL(m.config.AuthURL, m.config.BaseURL)
	magicURL := buildEmailURL(authBase, "token", token)
	expiryMinutes := int(expiry.Minutes())

	tpl := MagicLinkEmailTemplate(magicURL, expiryMinutes)
	return m.sendEmail(user.Email, subject, tpl.Text, tpl.HTML)
}

// SendMagicLinkSignupEmail sends a magic link email for new user signup
func (m *MockService) SendMagicLinkSignupEmail(email string, token string, expiry time.Duration) error {
	subject := "Welcome to FunctionFly — Complete Your Sign Up"
	authBase := baseURL(m.config.AuthURL, m.config.BaseURL)
	magicURL := buildEmailURL(authBase, "token", token)
	expiryMinutes := int(expiry.Minutes())

	tpl := MagicLinkSignupEmailTemplate(magicURL, expiryMinutes)
	return m.sendEmail(email, subject, tpl.Text, tpl.HTML)
}

// SendEmail sends a generic email with the given subject and body
func (m *MockService) SendEmail(to, subject, textBody, htmlBody string) error {
	return m.sendEmail(to, subject, textBody, htmlBody)
}

// ValidateConfiguration checks if the mock email service is properly configured
func (m *MockService) ValidateConfiguration() error {
	if m.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}
	if m.config.SMTPPort == 0 {
		return fmt.Errorf("SMTP port not configured")
	}
	if m.config.FromEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)
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

	if err := client.Hello("functionfly-mock-validation"); err != nil {
		return fmt.Errorf("SMTP hello failed: %w", err)
	}

	if m.config.SMTPUsername != "" && m.config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Quit(); err != nil {
		fmt.Printf("Warning: SMTP QUIT failed: %v\n", err)
	}

	return nil
}

// sendEmail sends an email using SMTP for the mock service
func (m *MockService) sendEmail(to, subject, textBody, htmlBody string) error {
	msg := buildMessage(to, subject, m.config.FromName, m.config.FromEmail, textBody, htmlBody)
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)
	return sendSMTP(addr, m.config.SMTPUsername, m.config.SMTPPassword, m.config.FromEmail, []string{to}, msg)
}

// sendEmailToMultiple sends an email to multiple recipients for the mock service
func (m *MockService) sendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	msg := buildMessageMultiple(to, subject, m.config.FromName, m.config.FromEmail, textBody, htmlBody)
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)
	return sendSMTP(addr, m.config.SMTPUsername, m.config.SMTPPassword, m.config.FromEmail, to, msg)
}

// ═══════════════════════════════════════════════════════════════════════════════
// NEW TEMPLATE METHODS (Velocity Orange Brand)
// ═══════════════════════════════════════════════════════════════════════════════

// Security & Alerts
func (m *MockService) SendNewDeviceLoginAlert(email, deviceInfo, location, ipAddress string, loginTime time.Time) error {
	subject := "[TEST] New Device Login Detected — FunctionFly"
	tpl := NewDeviceLoginTemplate(deviceInfo, location, ipAddress, loginTime)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendPasswordChangedConfirmation(email string, changedAt time.Time, deviceInfo string) error {
	subject := "[TEST] Password Changed — FunctionFly"
	tpl := PasswordChangedTemplate(changedAt, deviceInfo)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendSecurityAlert(email, alertType, description, actionRequired string) error {
	subject := fmt.Sprintf("[TEST] Security Alert: %s — FunctionFly", alertType)
	tpl := SecurityAlertTemplate(alertType, description, actionRequired)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendAgentWalletLowBalance(email string, balance, threshold float64, walletID string) error {
	subject := "[TEST] Wallet Balance Low — FunctionFly"
	tpl := AgentWalletLowBalanceTemplate(balance, threshold, walletID)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

// Marketing & Engagement
func (m *MockService) SendWaitlistInviteEmail(email, inviteCode, signupURL string, expiresAt time.Time) error {
	subject := "[TEST] You're In! FunctionFly Early Access"
	tpl := WaitlistInviteTemplate(inviteCode, signupURL, expiresAt)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

// Platform Operations
func (m *MockService) SendFunctionDeploySuccess(email, functionName, version, runtime string, deployTime time.Time) error {
	subject := fmt.Sprintf("[TEST] Deployment Successful: %s — FunctionFly", functionName)
	tpl := FunctionDeploySuccessTemplate(functionName, version, runtime, deployTime)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendFunctionDeployFailure(email, functionName, errorMsg string, retryCount int) error {
	subject := fmt.Sprintf("[TEST] Deployment Failed: %s — FunctionFly", functionName)
	tpl := FunctionDeployFailureTemplate(functionName, errorMsg, retryCount)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendUsageExportReady(email, exportID, downloadURL string, expiresAt time.Time, sizeBytes int64) error {
	subject := "[TEST] Your Export is Ready — FunctionFly"
	tpl := UsageExportReadyTemplate(exportID, downloadURL, expiresAt, sizeBytes)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

// Billing & Payments
func (m *MockService) SendPaymentFailed(email string, amount float64, dueDate time.Time, retryURL string) error {
	subject := "[TEST] Payment Failed — FunctionFly"
	tpl := PaymentFailedTemplate(amount, dueDate, retryURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendInvoiceReady(email, period string, amount float64, invoiceURL string) error {
	subject := fmt.Sprintf("[TEST] Invoice Ready: %s — FunctionFly", period)
	tpl := InvoiceReadyTemplate(period, amount, invoiceURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

// Trust & Compliance
func (m *MockService) SendTrustRevocationAlert(email, functionName, reason string, revokedAt time.Time) error {
	subject := fmt.Sprintf("[TEST] Trust Status Revoked: %s — FunctionFly", functionName)
	tpl := TrustRevocationTemplate(functionName, reason, revokedAt)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendDataRequestConfirmation(email, requestType, requestID string, estimatedCompletion time.Time) error {
	subject := "[TEST] Data Request Received — FunctionFly"
	tpl := DataRequestConfirmationTemplate(requestType, requestID, estimatedCompletion)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendAccountDeletionScheduled(email string, deletionDate time.Time, cancelURL string) error {
	subject := "[TEST] Account Deletion Scheduled — FunctionFly"
	tpl := AccountDeletionScheduledTemplate(deletionDate, cancelURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

// Collaboration
func (m *MockService) SendTeamInvite(email, orgName, invitedBy, role, acceptURL string) error {
	subject := fmt.Sprintf("[TEST] You're Invited to Join %s — FunctionFly", orgName)
	tpl := TeamInviteTemplate(orgName, invitedBy, role, acceptURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendBundleWelcomeEmail(email, bundleName, dashboardURL string) error {
	subject := fmt.Sprintf("[TEST] Welcome to %s — FunctionFly", bundleName)
	tpl := BundleWelcomeTemplate(bundleName, dashboardURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendInviteEmail(email, inviterName, orgName, role, acceptURL string) error {
	subject := fmt.Sprintf("[TEST] You're Invited to Join %s — FunctionFly", orgName)
	tpl := InviteEmailTemplate(inviterName, orgName, role, acceptURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendVaultSecretShared(email, secretName, sharedBy, accessLevel, viewURL string) error {
	subject := "[TEST] Secret Shared with You — FunctionFly Vault"
	tpl := VaultSecretSharedTemplate(secretName, sharedBy, accessLevel, viewURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

// Maintenance & Reminders
func (m *MockService) SendAPIKeyRotationReminder(email, keyName, keyID string, expiresAt time.Time, rotationURL string) error {
	subject := "[TEST] API Key Expiring Soon — FunctionFly"
	tpl := KeyRotationReminderTemplate(keyName, keyID, expiresAt, rotationURL)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}

func (m *MockService) SendMaintenanceNotice(email string, windowStart, windowEnd time.Time, affectedServices []string) error {
	subject := "[TEST] Scheduled Maintenance Notice — FunctionFly"
	tpl := MaintenanceNoticeTemplate(windowStart, windowEnd, affectedServices)
	html := TestBannerHTML(tpl.HTML)
	text := TestBannerText(tpl.Text)
	return m.sendEmail(email, subject, text, html)
}
