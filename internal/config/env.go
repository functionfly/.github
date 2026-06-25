package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetFrontendURL() string {
	return getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
}

func GetBaseURL() string {
	return getEnvOrDefault("BASE_URL", "http://localhost:8080")
}

// RequiredEnvVars lists environment variables that must be set for the server to start.
var RequiredEnvVars = []struct {
	Name    string
	EnvVar  string
	Example string
}{
	{"JWT Secret", "JWT_SECRET", "openssl rand -base64 32"},
	{"Privacy Salt", "PRIVACY_SALT", "openssl rand -base64 32"},
	{"Database Host", "DB_HOST", "localhost"},
	{"Database User", "DB_USER", "postgres"},
	{"Database Name", "DB_NAME", "functionfly"},
}

// OptionalEnvVars lists environment variables that should be set for full functionality.
var OptionalEnvVars = []struct {
	Name         string
	EnvVar       string
	DefaultValue string
}{
	{"Database Port", "DB_PORT", ""},
	{"Database SSL Mode", "DB_SSLMODE", ""},
	{"Redis Address", "REDIS_ADDR", ""},
	{"Base URL", "BASE_URL", ""},
	{"Shutdown Timeout", "SHUTDOWN_TIMEOUT", ""},
	{"Enable Vector Search", "ENABLE_VECTOR_SEARCH", "false"},
}

// GitHubEnvVars lists GitHub-specific environment variables.
// These are conditionally required when GitHub integration is enabled.
var GitHubEnvVars = []struct {
	Name    string
	EnvVar  string
	Example string
}{
	{"GitHub OAuth Client ID", "GITHUB_CLIENT_ID", "Create at github.com/settings/developers"},
	{"GitHub OAuth Client Secret", "GITHUB_CLIENT_SECRET", "Create at github.com/settings/developers"},
}

// GitHubOptionalEnvVars lists optional GitHub environment variables with defaults.
var GitHubOptionalEnvVars = []struct {
	Name         string
	EnvVar       string
	DefaultValue string
}{
	{"GitHub Vault Key", "GITHUB_VAULT_KEY", ""},
	{"GitHub Redirect URL", "GITHUB_REDIRECT_URL", ""},
	{"Frontend URL", "FRONTEND_URL", ""},
}

// ValidateEnv checks that all required environment variables are set.
// It returns an error listing all missing variables rather than failing on the
// first one, so operators can fix everything in one pass.
func ValidateEnv() error {
	var missing []string
	isDev := os.Getenv("DEVELOPMENT") == "true"

	for _, v := range RequiredEnvVars {
		if strings.TrimSpace(os.Getenv(v.EnvVar)) == "" {
			missing = append(missing, fmt.Sprintf("  - %s (%s) — e.g. %s", v.Name, v.EnvVar, v.Example))
		}
	}

	// In production, DB_PASSWORD is also required
	if !isDev && strings.TrimSpace(os.Getenv("DB_PASSWORD")) == "" {
		missing = append(missing, "  - Database Password (DB_PASSWORD) — required in production")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n%s", strings.Join(missing, "\n"))
	}

	// Warn about optional vars that are missing but have defaults
	for _, v := range OptionalEnvVars {
		if strings.TrimSpace(os.Getenv(v.EnvVar)) == "" {
			logrus.Warnf("ENV: %s (%s) not set, defaulting to %q", v.Name, v.EnvVar, v.DefaultValue)
		}
	}

	// Validate GitHub integration env vars (warn if partially configured)
	ghClientID := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID"))
	ghClientSecret := strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET"))
	if ghClientID != "" || ghClientSecret != "" {
		// At least one is set — warn if the other is missing
		for _, v := range GitHubEnvVars {
			if strings.TrimSpace(os.Getenv(v.EnvVar)) == "" {
				logrus.Warnf("ENV: GitHub integration partially configured — %s (%s) is missing. OAuth will fail without both.", v.Name, v.EnvVar)
			}
		}
	} else {
		logrus.Info("ENV: GitHub integration not configured (GITHUB_CLIENT_ID / GITHUB_CLIENT_SECRET not set)")
	}
	for _, v := range GitHubOptionalEnvVars {
		if strings.TrimSpace(os.Getenv(v.EnvVar)) == "" {
			logrus.Debugf("ENV: %s (%s) not set, defaulting to %q", v.Name, v.EnvVar, v.DefaultValue)
		}
	}

	// Warn if CORS is not explicitly configured
	if strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")) == "" {
		isDev := os.Getenv("DEVELOPMENT") == "true"
		if isDev {
			logrus.Warn("ENV: CORS_ALLOWED_ORIGINS not set — allowing all origins in development mode")
		} else {
			logrus.Warn("ENV: CORS_ALLOWED_ORIGINS not set in production — cross-origin requests will be denied")
		}
	}

	// Run production-specific validation (includes Stripe webhook secret check)
	if err := ValidateProductionEnv(); err != nil {
		return err
	}

	logrus.Info("Environment validation passed")
	return nil
}
