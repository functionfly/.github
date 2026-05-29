package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// IsProduction returns true when the process is configured for a production deployment.
// DEVELOPMENT=true explicitly opts out of production mode (local dev), even if APP_ENV=production
// is set by mistake in the environment.
func IsProduction() bool {
	if os.Getenv("DEVELOPMENT") == "true" {
		return false
	}

	switch {
	case os.Getenv("PRODUCTION_ENV") == "true":
		return true
	case os.Getenv("PRODUCTION") == "true":
		return true
	case strings.EqualFold(os.Getenv("ENVIRONMENT"), "production"):
		return true
	case strings.EqualFold(os.Getenv("APP_ENV"), "production"):
		return true
	default:
		return false
	}
}

// ValidateProductionEnv checks production-only requirements and fails fast when any are missing.
func ValidateProductionEnv() error {
	if !IsProduction() {
		return nil
	}

	logrus.Info("Running production environment validation...")

	var missing []string

	prodRequired := []struct {
		Name   string
		EnvVar string
		Hint   string
	}{
		{"API Shared Secret", "API_SHARED_SECRET", "openssl rand -base64 32"},
		{"Redis Address", "REDIS_ADDR", "redis:6379"},
		{"AI Service URL", "AI_SERVICE_URL", "https://ai.yourdomain.com"},
		{"Privacy Export Access Key", "PRIVACY_EXPORT_ACCESS_KEY_ID", "S3/R2 access key ID"},
		{"Privacy Export Secret Key", "PRIVACY_EXPORT_SECRET_ACCESS_KEY", "S3/R2 secret access key"},
		{"Wallet Encryption Key", "WALLET_ENCRYPTION_KEY", "openssl rand -base64 32"},
		{"Wallet Audit HMAC Key", "WALLET_AUDIT_HMAC_KEY", "openssl rand -base64 32"},
		{"MFA Encryption Key", "MFA_ENCRYPTION_KEY", "openssl rand -base64 32"},
		{"Stripe Webhook Secret", "STRIPE_WEBHOOK_SECRET", "whsec_... from Stripe dashboard"},
	}

	for _, v := range prodRequired {
		if strings.TrimSpace(os.Getenv(v.EnvVar)) == "" {
			missing = append(missing, fmt.Sprintf("  - %s (%s) — e.g. %s", v.Name, v.EnvVar, v.Hint))
		}
	}

	if strings.TrimSpace(os.Getenv("RESEND_API_KEY")) == "" && strings.TrimSpace(os.Getenv("SMTP_HOST")) == "" {
		missing = append(missing, "  - Email delivery (RESEND_API_KEY or SMTP_HOST) — required for auth and transactional email")
	}

	cors := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if cors == "" || cors == "*" {
		missing = append(missing, "  - CORS Allowed Origins (CORS_ALLOWED_ORIGINS) — must list specific origins, not wildcard")
	}

	backend := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_BACKEND")))
	if backend == "" {
		backend = strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_TYPE")))
	}
	if backend == "" || backend == "local" {
		missing = append(missing, "  - Object Storage (STORAGE_BACKEND) — must be s3 or r2 in production, not local")
	}

	if backend == "s3" || backend == "r2" {
		if strings.TrimSpace(os.Getenv("STORAGE_BUCKET")) == "" && strings.TrimSpace(os.Getenv("S3_BUCKET")) == "" {
			missing = append(missing, "  - Storage Bucket (STORAGE_BUCKET) — required when STORAGE_BACKEND is s3 or r2")
		}
	}

	if os.Getenv("PRODUCTION_ENV") != "true" {
		logrus.Warn("ENV: PRODUCTION_ENV=true not set — advanced security middleware (DDoS, geo-blocking, rate limits) stays disabled until set")
	}

	if strings.TrimSpace(os.Getenv("MAXMIND_LICENSE_KEY")) == "" {
		logrus.Warn("ENV: MAXMIND_LICENSE_KEY not set — using simplified region detection. Recommended for accurate GDPR geo processing (free at https://www.maxmind.com/en/geolite2/signup)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("production environment validation failed:\n%s", strings.Join(missing, "\n"))
	}

	logrus.Info("Production environment validation passed")
	return nil
}
