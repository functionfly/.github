package email

import (
	"os"
	"strconv"
	"strings"
)

// NewServiceFromEnv returns the email transport from environment variables.
// Resend is preferred when RESEND_API_KEY is set; otherwise SMTP_* when SMTP_HOST is set.
// Returns (nil, false) when neither is configured (use mock in development).
func NewServiceFromEnv() (Service, bool) {
	resendKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if resendKey != "" {
		fromEmail := strings.TrimSpace(os.Getenv("FROM_EMAIL"))
		if fromEmail == "" {
			fromEmail = strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL"))
		}
		fromName := strings.TrimSpace(os.Getenv("FROM_NAME"))
		if fromName == "" {
			fromName = strings.TrimSpace(os.Getenv("SMTP_FROM_NAME"))
		}
		if fromName == "" {
			fromName = "FunctionFly"
		}
		baseURL := strings.TrimSpace(os.Getenv("BASE_URL"))
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
		authURL := strings.TrimSpace(os.Getenv("AUTH_URL"))
		replyTo := strings.TrimSpace(os.Getenv("RESEND_REPLY_TO"))
		cfg := Config{
			Provider:     "resend",
			ResendAPIKey: resendKey,
			FromEmail:    fromEmail,
			FromName:     fromName,
			BaseURL:      baseURL,
			AuthURL:      authURL,
			ReplyToEmail: replyTo,
		}
		return NewService(cfg), true
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if smtpHost == "" {
		return nil, false
	}

	smtpPort := 587
	if p := strings.TrimSpace(os.Getenv("SMTP_PORT")); p != "" {
		if port, err := strconv.Atoi(p); err == nil && port > 0 {
			smtpPort = port
		}
	}

	fromEmail := strings.TrimSpace(os.Getenv("FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL"))
	}
	if fromEmail == "" {
		fromEmail = "noreply@functionfly.dev"
	}
	fromName := strings.TrimSpace(os.Getenv("FROM_NAME"))
	if fromName == "" {
		fromName = strings.TrimSpace(os.Getenv("SMTP_FROM_NAME"))
	}
	if fromName == "" {
		fromName = "FunctionFly"
	}
	baseURL := strings.TrimSpace(os.Getenv("BASE_URL"))
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	authURL := strings.TrimSpace(os.Getenv("AUTH_URL"))

	cfg := Config{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    fromEmail,
		FromName:     fromName,
		BaseURL:      baseURL,
		AuthURL:      authURL,
	}
	return NewSMTPService(cfg), true
}
