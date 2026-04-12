package wallet

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// ProductionSecurityValidation validates all security-related environment variables
// and configuration for production deployment
func ProductionSecurityValidation() error {
	isProd := os.Getenv("PRODUCTION") == "true" || os.Getenv("ENVIRONMENT") == "production"

	if !isProd {
		logrus.Info("Running in non-production mode - security validations relaxed")
		return nil
	}

	logrus.Info("Running production security validations...")

	errors := []string{}

	// 1. API_SHARED_SECRET is required for HMAC protection in production
	if os.Getenv("API_SHARED_SECRET") == "" {
		errors = append(errors, "API_SHARED_SECRET is required in production for HMAC signature verification")
	}

	// 2. WALLET_ENCRYPTION_KEY is required for encryption at rest
	enc := GetWalletEncryption()
	if !enc.IsEnabled() {
		errors = append(errors, "WALLET_ENCRYPTION_KEY must be set in production for wallet encryption at rest")
	}

	// 3. STRIPE_WEBHOOK_SECRET is required for webhook security
	if os.Getenv("STRIPE_WEBHOOK_SECRET") == "" {
		logrus.Warn("STRIPE_WEBHOOK_SECRET not set - Stripe webhooks will not be verified (security risk)")
		errors = append(errors, "STRIPE_WEBHOOK_SECRET is required in production for webhook signature verification")
	}

	// 4. JWT_SECRET should be strong (not default)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" || jwtSecret == "functionfly-jwt-secret-key-2026" {
		errors = append(errors, "JWT_SECRET must be set to a strong random value in production (not the default)")
	}

	// 5. Database encryption should be configured
	if os.Getenv("DATABASE_URL") != "" {
		// Check if using SSL
		dbURL := os.Getenv("DATABASE_URL")
		if !contains(dbURL, "sslmode=require") && !contains(dbURL, "sslmode=verify") {
			logrus.Warn("Database connection may not be using SSL - consider adding sslmode=require")
		}
	}

	// 6. Redis should be configured with authentication in production
	if os.Getenv("REDIS_PASSWORD") == "" && os.Getenv("REDIS_URL") != "" {
		// Check if Redis URL contains password
		redisURL := os.Getenv("REDIS_URL")
		if !contains(redisURL, ":@") && !contains(redisURL, ":password") {
			logrus.Warn("Redis may not be configured with authentication")
		}
	}

	// 7. CORS_ALLOWED_ORIGINS should be set (not wildcard in production)
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" || corsOrigins == "*" {
		errors = append(errors, "CORS_ALLOWED_ORIGINS must be set to specific origins in production (not wildcard)")
	}

	// 8. Content Security Policy should be configured
	if os.Getenv("CONTENT_SECURITY_POLICY") == "" {
		logrus.Info("Using default Content Security Policy - consider customizing for production")
	}

	// 9. Admin adjustment limits should be reviewed
	limits := GetAdminAdjustmentLimits()
	if limits.SingleOperationMax > 10000 {
		logrus.Warnf("WALLET_ADMIN_ADJUSTMENT_SINGLE_MAX is high ($%.2f) - consider lowering for production", limits.SingleOperationMax)
	}

	// 10. Log security configuration
	logrus.WithFields(logrus.Fields{
		"hmac_enabled":                     os.Getenv("API_SHARED_SECRET") != "",
		"wallet_encryption_enabled":        enc.IsEnabled(),
		"stripe_webhook_verification":      os.Getenv("STRIPE_WEBHOOK_SECRET") != "",
		"admin_adjustment_single_max":      limits.SingleOperationMax,
		"admin_adjustment_daily_max":       limits.DailyMax,
		"admin_adjustment_secondary_above": limits.RequiresSecondaryApprovalAbove,
	}).Info("Security configuration summary")

	if len(errors) > 0 {
		for _, err := range errors {
			logrus.Errorf("Security validation error: %s", err)
		}
		return fmt.Errorf("production security validation failed with %d errors", len(errors))
	}

	logrus.Info("Production security validation passed")
	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SecurityChecklist returns a checklist of security items and their status
func SecurityChecklist() map[string]bool {
	return map[string]bool{
		"API_SHARED_SECRET_set":     os.Getenv("API_SHARED_SECRET") != "",
		"WALLET_ENCRYPTION_KEY_set": GetWalletEncryption().IsEnabled(),
		"STRIPE_WEBHOOK_SECRET_set": os.Getenv("STRIPE_WEBHOOK_SECRET") != "",
		"JWT_SECRET_set":            os.Getenv("JWT_SECRET") != "" && os.Getenv("JWT_SECRET") != "functionfly-jwt-secret-key-2026",
		"DATABASE_SSL_configured":   contains(os.Getenv("DATABASE_URL"), "sslmode=require") || contains(os.Getenv("DATABASE_URL"), "sslmode=verify"),
		"CORS_ORIGINS_set":          os.Getenv("CORS_ALLOWED_ORIGINS") != "" && os.Getenv("CORS_ALLOWED_ORIGINS") != "*",
		"PRODUCTION_flag_set":       os.Getenv("PRODUCTION") == "true" || os.Getenv("ENVIRONMENT") == "production",
	}
}
