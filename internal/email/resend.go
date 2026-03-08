package email

import (
	"context"
	"fmt"
	"time"

	"github.com/resend/resend-go/v2"
	"github.com/functionfly/functionfly/internal/storage"
)

type ResendConfig struct {
	APIKey      string
	FromEmail   string
	FromName    string
	BaseURL     string
	ReplyToEmail string
}

type ResendService struct {
	client   *resend.Client
	config   ResendConfig
}

func NewResendService(config ResendConfig) *ResendService {
	return &ResendService{
		client: resend.NewClient(config.APIKey),
		config: config,
	}
}

func (s *ResendService) SendVerificationEmail(user *storage.User, verificationToken string) error {
	if user == nil || user.VerificationToken == nil {
		return fmt.Errorf("user or verification token is nil")
	}

	subject := "Verify Your Email Address - FunctionFly"
	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.config.BaseURL, *user.VerificationToken)

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
            <p>%s</p>
            <p>This verification link will expire in 24 hours.</p>
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 FunctionFly. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, verificationURL, verificationURL)

	textBody := fmt.Sprintf(`Welcome to FunctionFly!

Thank you for signing up. Please verify your email address to complete your registration.

Click this link to verify your email: %s

This verification link will expire in 24 hours.

If you didn't create an account, please ignore this email.

--
FunctionFly Team
`, verificationURL)

	return s.SendEmail(user.Email, subject, textBody, htmlBody)
}

func (s *ResendService) SendPasswordResetEmail(user *storage.User, resetToken string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	subject := "Reset Your Password - FunctionFly"
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.config.BaseURL, resetToken)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Reset Your Password</title>
<style>body{font-family:Arial,sans-serif;line-height:1.6;color:#333}.container{max-width:600px;margin:0 auto;padding:20px}.header{background-color:#4F46E5;color:white;padding:20px;text-align:center}.content{padding:30px 20px;background-color:#f9f9f9}.button{display:inline-block;background-color:#4F46E5;color:white;padding:12px 24px;text-decoration:none;border-radius:4px;margin:20px 0}.footer{text-align:center;font-size:12px;color:#666;padding:20px}</style>
</head>
<body><div class="container">
<div class="header"><h1>FunctionFly</h1></div>
<div class="content">
<h2>Reset Your Password</h2>
<p>We received a request to reset the password for your account.</p>
<p>Click the button below to reset your password. This link expires in 1 hour.</p>
<a href="%s" class="button">Reset Password</a>
<p>If you didn't request a password reset, you can safely ignore this email.</p>
</div>
<div class="footer"><p>&copy; FunctionFly. All rights reserved.</p></div>
</div></body></html>`, resetURL)

	textBody := fmt.Sprintf("Reset your FunctionFly password by visiting: %s\n\nThis link expires in 1 hour.\n\nIf you didn't request this, ignore this email.", resetURL)

	return s.SendEmail(user.Email, subject, textBody, htmlBody)
}

func (s *ResendService) SendBreachNotification(recipients []string, breachDetails map[string]interface{}) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified for breach notification")
	}

	subject := "URGENT: Data Breach Notification - GDPR Article 33 Compliance"

	breachType := "Data Breach"
	if v, ok := breachDetails["type"].(string); ok && v != "" {
		breachType = v
	}

	detectionTime := time.Now().Format(time.RFC3339)
	if v, ok := breachDetails["detectionTime"].(string); ok && v != "" {
		detectionTime = v
	}

	affectedUsers := 0
	if v, ok := breachDetails["affectedUsers"].(int); ok {
		affectedUsers = v
	}

	riskLevel := "high"
	if v, ok := breachDetails["riskLevel"].(string); ok && v != "" {
		riskLevel = v
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

	return s.SendEmailToMultiple(recipients, subject, textBody, htmlBody)
}

func (s *ResendService) SendEmail(to, subject, textBody, htmlBody string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = &s.config.ReplyToEmail
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *ResendService) SendEmailToMultiple(to []string, subject, textBody, htmlBody string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := &resend.SendEmailRequest{
		From:    s.config.FromName + " <" + s.config.FromEmail + ">",
		To:      to,
		Subject: subject,
		Html:    htmlBody,
	}

	if s.config.ReplyToEmail != "" {
		params.ReplyTo = &s.config.ReplyToEmail
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *ResendService) ValidateConfiguration() error {
	if s.config.APIKey == "" {
		return fmt.Errorf("Resend API key not configured")
	}
	if s.config.FromEmail == "" {
		return fmt.Errorf("from email not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.Keys.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate Resend API key: %w", err)
	}

	return nil
}
