package config

import (
	"strings"
	"testing"
)

func TestIsProduction(t *testing.T) {
	t.Setenv("DEVELOPMENT", "")
	t.Setenv("PRODUCTION_ENV", "")
	t.Setenv("PRODUCTION", "")
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("APP_ENV", "")

	if IsProduction() {
		t.Fatal("expected non-production with empty env")
	}

	t.Setenv("APP_ENV", "production")
	if !IsProduction() {
		t.Fatal("expected production when APP_ENV=production")
	}

	t.Setenv("DEVELOPMENT", "true")
	if IsProduction() {
		t.Fatal("expected DEVELOPMENT=true to override APP_ENV=production")
	}
}

func TestValidateProductionEnvMissingVars(t *testing.T) {
	t.Setenv("DEVELOPMENT", "")
	t.Setenv("PRODUCTION_ENV", "true")
	for _, key := range []string{
		"API_SHARED_SECRET",
		"REDIS_ADDR",
		"AI_SERVICE_URL",
		"PRIVACY_EXPORT_ACCESS_KEY_ID",
		"PRIVACY_EXPORT_SECRET_ACCESS_KEY",
		"WALLET_ENCRYPTION_KEY",
		"WALLET_AUDIT_HMAC_KEY",
		"MFA_ENCRYPTION_KEY",
		"STRIPE_WEBHOOK_SECRET",
		"RESEND_API_KEY",
		"SMTP_HOST",
		"CORS_ALLOWED_ORIGINS",
		"STORAGE_BACKEND",
	} {
		t.Setenv(key, "")
	}

	err := ValidateProductionEnv()
	if err == nil {
		t.Fatal("expected validation error for missing production vars")
	}
	if !strings.Contains(err.Error(), "AI_SERVICE_URL") {
		t.Fatalf("expected AI_SERVICE_URL in error, got: %v", err)
	}
}

func TestValidateProductionEnvPasses(t *testing.T) {
	t.Setenv("DEVELOPMENT", "")
	t.Setenv("PRODUCTION_ENV", "true")
	t.Setenv("API_SHARED_SECRET", "test-api-shared-secret-value-32chars")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("AI_SERVICE_URL", "http://ai-service:8081")
	t.Setenv("PRIVACY_EXPORT_ACCESS_KEY_ID", "access-key")
	t.Setenv("PRIVACY_EXPORT_SECRET_ACCESS_KEY", "secret-key")
	t.Setenv("WALLET_ENCRYPTION_KEY", "dGVzdC1rZXktdGVzdC1rZXktdGVzdC1rZXktdGVzdA==")
	t.Setenv("WALLET_AUDIT_HMAC_KEY", "audit-hmac-key")
	t.Setenv("MFA_ENCRYPTION_KEY", "mfa-encryption-key")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	t.Setenv("STORAGE_BACKEND", "s3")
	t.Setenv("STORAGE_BUCKET", "functionfly-uploads")

	if err := ValidateProductionEnv(); err != nil {
		t.Fatalf("expected validation to pass, got: %v", err)
	}
}
