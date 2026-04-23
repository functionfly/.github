package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

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

// OptionalEnvVars lists environment variables with recommended defaults.
var OptionalEnvVars = []struct {
	Name         string
	EnvVar       string
	DefaultValue string
}{
	{"Database Port", "DB_PORT", "5432"},
	{"Database SSL Mode", "DB_SSLMODE", "disable"},
	{"Redis Address", "REDIS_ADDR", "localhost:6379"},
	{"Base URL", "BASE_URL", "http://localhost:8080"},
	{"Shutdown Timeout", "SHUTDOWN_TIMEOUT", "30s"},
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

	// Warn if CORS is not explicitly configured
	if strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")) == "" {
		isDev := os.Getenv("DEVELOPMENT") == "true"
		if isDev {
			logrus.Warn("ENV: CORS_ALLOWED_ORIGINS not set — allowing all origins in development mode")
		} else {
			logrus.Warn("ENV: CORS_ALLOWED_ORIGINS not set in production — cross-origin requests will be denied")
		}
	}

	logrus.Info("Environment validation passed")
	return nil
}
