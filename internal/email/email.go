package email

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// Service defines the interface for email operations
type Service interface {
	SendVerificationEmail(user *storage.User, verificationToken string) error
	SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error
	ValidateConfiguration() error
}

// Config holds email service configuration
type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string
}

// SMTPService implements the Service interface using SMTP
type SMTPService struct {
	config Config
}

// NewSMTPService creates a new SMTP email service
func NewSMTPService(config Config) *SMTPService {
	return &SMTPService{
		config: config,
	}
}

// SendVerificationEmail sends an email verification email to a user
func (s *SMTPService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "Verify Your Email Address - FunctionFly"
	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.config.BaseURL, *user.VerificationToken)

	// HTML email template
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .button { display: inline-block; background-color: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>FunctionFly</h1>
        </div>
        <div class="content">
            <h2>Welcome to FunctionFly!</h2>
            <p>Thank you for signing up. Please verify your email address to complete your registration.</p>
            <p>Click the button below to verify your email:</p>
            <a href="%s" class="button">Verify Email Address</a>
            <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
            <p><a href="%s">%s</a></p>
            <p>This verification link will expire in 24 hours.</p>
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 FunctionFly. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, verificationURL, verificationURL, verificationURL)

	// Plain text version
	textBody := fmt.Sprintf(`Welcome to FunctionFly!

Thank you for signing up. Please verify your email address to complete your registration.

Click this link to verify your email: %s

This verification link will expire in 24 hours.

If you didn't create an account, please ignore this email.

--
FunctionFly Team
`, verificationURL)

	return s.sendEmail(user.Email, subject, textBody, htmlBody)
}

// sendEmail sends an email using SMTP
func (s *SMTPService) sendEmail(to, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if s.config.SMTPUsername != "" || s.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		err = smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, s.config.FromEmail, []string{to}, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MockService implements the Service interface using SMTP (for testing with real email sending)
type MockService struct {
	config Config
}

// NewMockService creates a new mock email service with real SMTP functionality
func NewMockService(config Config) *MockService {
	return &MockService{
		config: config,
	}
}

// SendBreachNotification sends a GDPR breach notification email
func (s *SMTPService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	// Extract breach details with defaults
	breachType := breachDetails["type"].(string)
	if breachType == "" {
		breachType = "Data Breach"
	}
	detectionTime := breachDetails["detectionTime"].(string)
	if detectionTime == "" {
		detectionTime = time.Now().Format(time.RFC3339)
	}
	affectedUsers := breachDetails["affectedUsers"].(int)
	riskLevel := breachDetails["riskLevel"].(string)
	if riskLevel == "" {
		riskLevel = "high"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Data Breach Notification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .urgent { color: #dc3545; font-weight: bold; font-size: 18px; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 20px; margin: 20px 0; border-left: 4px solid #dc3545; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>DATA BREACH NOTIFICATION</h1>
            <p class="urgent">GDPR Article 33 Compliance</p>
        </div>
        <div class="content">
            <h2>Urgent Security Incident</h2>
            <p>This is an official notification regarding a data breach that affects personal data processing activities.</p>

            <div class="details">
                <h3>Breach Details:</h3>
                <ul>
                    <li><strong>Type:</strong> %s</li>
                    <li><strong>Detection Time:</strong> %s</li>
                    <li><strong>Estimated Affected Users:</strong> %d</li>
                    <li><strong>Risk Level:</strong> %s</li>
                </ul>
            </div>

            <p><strong>Next Steps:</strong></p>
            <ul>
                <li>Notification to supervisory authority within 72 hours</li>
                <li>Communication to affected individuals (if high risk)</li>
                <li>Implementation of remedial measures</li>
                <li>Documentation of the incident</li>
            </ul>

            <p>For additional information or concerns, please contact our Data Protection Officer.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 FunctionFly. All rights reserved.</p>
            <p>This notification is sent in compliance with GDPR Article 33.</p>
        </div>
    </div>
</body>
</html>
`, breachType, detectionTime, affectedUsers, riskLevel)

	textBody := fmt.Sprintf(`DATA BREACH NOTIFICATION - GDPR Article 33 Compliance

Urgent Security Incident

This is an official notification regarding a data breach that affects personal data processing activities.

Breach Details:
- Type: %s
- Detection Time: %s
- Estimated Affected Users: %d
- Risk Level: %s

Next Steps:
- Notification to supervisory authority within 72 hours
- Communication to affected individuals (if high risk)
- Implementation of remedial measures
- Documentation of the incident

For additional information or concerns, please contact our Data Protection Officer.

--
FunctionFly Security Team
This notification is sent in compliance with GDPR Article 33.
`, breachType, detectionTime, affectedUsers, riskLevel)

	return s.sendEmailToMultiple(recipients, subject, textBody, htmlBody)
}

// ValidateConfiguration checks if the email service is properly configured
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

	// Test SMTP connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Construct SMTP server address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Test basic TCP connection first
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("SMTP server unreachable: %w", err)
	}
	conn.Close()

	// Test SMTP handshake
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	defer client.Close()

	// Get server greeting
	if err := client.Hello("functionfly-validation"); err != nil {
		return fmt.Errorf("SMTP hello failed: %w", err)
	}

	// Test authentication if credentials are provided
	if s.config.SMTPUsername != "" && s.config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Send QUIT to gracefully close connection
	if err := client.Quit(); err != nil {
		// This is not a critical error, just log it
		fmt.Printf("Warning: SMTP QUIT failed: %v\n", err)
	}

	return nil
}

// sendEmailToMultiple sends an email to multiple recipients
func (s *SMTPService) sendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if s.config.SMTPUsername != "" || s.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		err = smtp.SendMail(addr, auth, s.config.FromEmail, to, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, s.config.FromEmail, to, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send breach notification email: %w", err)
	}

	return nil
}

// SendVerificationEmail sends an email verification email to a user (with real SMTP for testing)
func (m *MockService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "[TEST] Verify Your Email Address - FunctionFly"
	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", m.config.BaseURL, *user.VerificationToken)

	// HTML email template
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .test-notice { background-color: #f39c12; color: white; padding: 10px; text-align: center; font-weight: bold; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .button { display: inline-block; background-color: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="test-notice">
            ⚠️ TEST EMAIL - FunctionFly Development Environment
        </div>
        <div class="header">
            <h1>FunctionFly</h1>
        </div>
        <div class="content">
            <h2>Welcome to FunctionFly!</h2>
            <p>Thank you for signing up. Please verify your email address to complete your registration.</p>
            <p>Click the button below to verify your email:</p>
            <a href="%s" class="button">Verify Email Address</a>
            <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
            <p><a href="%s">%s</a></p>
            <p>This verification link will expire in 24 hours.</p>
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 FunctionFly. All rights reserved.</p>
            <p><em>This is a test email from the development environment.</em></p>
        </div>
    </div>
</body>
</html>
`, verificationURL, verificationURL, verificationURL)

	// Plain text version
	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

Welcome to FunctionFly!

Thank you for signing up. Please verify your email address to complete your registration.

Click this link to verify your email: %s

This verification link will expire in 24 hours.

If you didn't create an account, please ignore this email.

--
FunctionFly Team

* This is a test email from the development environment.
`, verificationURL)

	return m.sendEmail(user.Email, subject, textBody, htmlBody)
}

// SendBreachNotification sends a GDPR breach notification email (with real SMTP for testing)
func (m *MockService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "[TEST] URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	// Extract breach details with defaults
	breachType := "Data Breach"
	if bt, ok := breachDetails["type"].(string); ok && bt != "" {
		breachType = bt
	}
	detectionTime := time.Now().Format(time.RFC3339)
	if dt, ok := breachDetails["detectionTime"].(string); ok && dt != "" {
		detectionTime = dt
	}
	affectedUsers := 0
	if au, ok := breachDetails["affectedUsers"].(int); ok {
		affectedUsers = au
	}
	riskLevel := "high"
	if rl, ok := breachDetails["riskLevel"].(string); ok && rl != "" {
		riskLevel = rl
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Data Breach Notification</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .test-notice { background-color: #f39c12; color: white; padding: 10px; text-align: center; font-weight: bold; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .urgent { color: #dc3545; font-weight: bold; font-size: 18px; }
        .content { padding: 30px 20px; background-color: #f9f9f9; }
        .details { background-color: white; padding: 20px; margin: 20px 0; border-left: 4px solid #dc3545; }
        .footer { text-align: center; font-size: 12px; color: #666; padding: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="test-notice">
            ⚠️ TEST EMAIL - FunctionFly Development Environment
        </div>
        <div class="header">
            <h1>DATA BREACH NOTIFICATION</h1>
            <p class="urgent">GDPR Article 33 Compliance</p>
        </div>
        <div class="content">
            <h2>Urgent Security Incident</h2>
            <p>This is an official notification regarding a data breach that affects personal data processing activities.</p>

            <div class="details">
                <h3>Breach Details:</h3>
                <ul>
                    <li><strong>Type:</strong> %s</li>
                    <li><strong>Detection Time:</strong> %s</li>
                    <li><strong>Estimated Affected Users:</strong> %d</li>
                    <li><strong>Risk Level:</strong> %s</li>
                </ul>
            </div>

            <p><strong>Next Steps:</strong></p>
            <ul>
                <li>Notification to supervisory authority within 72 hours</li>
                <li>Communication to affected individuals (if high risk)</li>
                <li>Implementation of remedial measures</li>
                <li>Documentation of the incident</li>
            </ul>

            <p>For additional information or concerns, please contact our Data Protection Officer.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 FunctionFly. All rights reserved.</p>
            <p>This notification is sent in compliance with GDPR Article 33.</p>
            <p><em>This is a test email from the development environment.</em></p>
        </div>
    </div>
</body>
</html>
`, breachType, detectionTime, affectedUsers, riskLevel)

	textBody := fmt.Sprintf(`[TEST EMAIL - FunctionFly Development Environment]

DATA BREACH NOTIFICATION - GDPR Article 33 Compliance

Urgent Security Incident

This is an official notification regarding a data breach that affects personal data processing activities.

Breach Details:
- Type: %s
- Detection Time: %s
- Estimated Affected Users: %d
- Risk Level: %s

Next Steps:
- Notification to supervisory authority within 72 hours
- Communication to affected individuals (if high risk)
- Implementation of remedial measures
- Documentation of the incident

For additional information or concerns, please contact our Data Protection Officer.

--
FunctionFly Security Team
This notification is sent in compliance with GDPR Article 33.

* This is a test email from the development environment.
`, breachType, detectionTime, affectedUsers, riskLevel)

	return m.sendEmailToMultiple(recipients, subject, textBody, htmlBody)
}

// sendEmail sends an email using SMTP for the mock service
func (m *MockService) sendEmail(to, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", m.config.FromName, m.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if m.config.SMTPUsername != "" || m.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		err = smtp.SendMail(addr, auth, m.config.FromEmail, []string{to}, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, m.config.FromEmail, []string{to}, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// sendEmailToMultiple sends an email to multiple recipients for the mock service
func (m *MockService) sendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	// Create message
	var message bytes.Buffer

	// Headers
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", m.config.FromName, m.config.FromEmail))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=boundary123\r\n")
	message.WriteString("\r\n")

	// Plain text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)

	// Send email (with or without authentication)
	var err error
	if m.config.SMTPUsername != "" || m.config.SMTPPassword != "" {
		// Use authentication when credentials are provided
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		err = smtp.SendMail(addr, auth, m.config.FromEmail, to, message.Bytes())
	} else {
		// Send without authentication (for Mailpit and other local/dev servers)
		err = smtp.SendMail(addr, nil, m.config.FromEmail, to, message.Bytes())
	}
	if err != nil {
		return fmt.Errorf("failed to send breach notification email: %w", err)
	}

	return nil
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

	// Test SMTP connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Construct SMTP server address
	addr := fmt.Sprintf("%s:%d", m.config.SMTPHost, m.config.SMTPPort)

	// Test basic TCP connection first
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("SMTP server unreachable: %w", err)
	}
	conn.Close()

	// Test SMTP handshake
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %w", err)
	}
	defer client.Close()

	// Get server greeting
	if err := client.Hello("functionfly-mock-validation"); err != nil {
		return fmt.Errorf("SMTP hello failed: %w", err)
	}

	// Test authentication if credentials are provided
	if m.config.SMTPUsername != "" && m.config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", m.config.SMTPUsername, m.config.SMTPPassword, m.config.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Send QUIT to gracefully close connection
	if err := client.Quit(); err != nil {
		// This is not a critical error, just log it
		fmt.Printf("Warning: SMTP QUIT failed: %v\n", err)
	}

	return nil
}