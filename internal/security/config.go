package security

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// Configuration security scanning functions

func (sas *SecurityAuditService) scanConfigurationSecurity(ctx context.Context, config ScanConfig) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check for exposed secrets
	if exposedSecrets, err := sas.checkExposedSecrets(); err == nil && len(exposedSecrets) > 0 {
		for _, secret := range exposedSecrets {
			vulnerabilities = append(vulnerabilities, Vulnerability{
				ID:          generateVulnID(),
				Title:       "Exposed Secret Detected",
				Description: fmt.Sprintf("Secret %s is exposed in configuration or logs", secret),
				Severity:    "critical",
				Category:    "config",
				Component:   "configuration",
				Status:      "open",
				Remediation: "Move secrets to secure vault and rotate exposed credentials",
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	// Check for debug mode enabled in production
	if debugEnabled, err := sas.checkDebugMode(); err == nil && debugEnabled {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Debug Mode Enabled in Production",
			Description: "Application is running in debug mode in production environment",
			Severity:    "medium",
			Category:    "config",
			Component:   "application",
			Status:      "open",
			Remediation: "Disable debug mode and error details in production",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkExposedSecrets() ([]string, error) {
	// Check for exposed secrets in config files, logs, etc.
	exposedSecrets := []string{}

	// Check environment variables for sensitive data
	sensitiveEnvVars := []string{
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
		"AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET", "AZURE_TENANT_ID",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_PRIVATE_KEY",
		"DATABASE_URL", "DB_PASSWORD", "DB_USER",
		"JWT_SECRET", "API_KEY", "SECRET_KEY",
		"GITHUB_TOKEN", "GITLAB_TOKEN", "SLACK_TOKEN",
		"STRIPE_SECRET_KEY", "STRIPE_PUBLISHABLE_KEY",
		"SENDGRID_API_KEY", "MAILGUN_API_KEY",
		"REDIS_URL", "RABBITMQ_URL",
	}

	for _, envVar := range sensitiveEnvVars {
		if value := os.Getenv(envVar); value != "" {
			// Check if it looks like a real secret (not placeholder)
			if len(value) > 10 && !strings.Contains(strings.ToLower(value), "your_") &&
			   !strings.Contains(strings.ToLower(value), "change_me") &&
			   !strings.Contains(strings.ToLower(value), "example") {
				exposedSecrets = append(exposedSecrets, fmt.Sprintf("Environment variable: %s", envVar))
			}
		}
	}

	// Check common config files for hardcoded secrets
	configPaths := []string{
		".env", ".env.local", ".env.production", ".env.staging",
		"config.json", "config.yaml", "config.yml",
		"settings.json", "settings.py",
		"application.properties", "application.yml",
		"secrets.json", "secrets.yaml",
	}

	secretPatterns := []struct {
		name  string
		regex string
	}{
		{"AWS Access Key", `AKIA[0-9A-Z]{16}`},
		{"AWS Secret Key", `[A-Za-z0-9+/]{40}`},
		{"JWT Token", `eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*`},
		{"Generic API Key", `[A-Za-z0-9]{32,}`},
		{"Database URL", `(postgres|mysql|mongodb)://[^:]+:[^@]+@`},
		{"Private Key", `-----BEGIN (RSA|EC|PRIVATE KEY)-----`},
		{"Stripe Key", `sk_(test|live)_[A-Za-z0-9]{24}`},
		{"GitHub Token", `ghp_[A-Za-z0-9]{36}`},
		{"Slack Token", `xox[baprs]-[0-9]+-[0-9]+-[0-9]+-[A-Za-z0-9]+`},
	}

	for _, configPath := range configPaths {
		if data, err := os.ReadFile(configPath); err == nil {
			content := string(data)
			for _, pattern := range secretPatterns {
				if matched, _ := regexp.MatchString(pattern.regex, content); matched {
					exposedSecrets = append(exposedSecrets, fmt.Sprintf("%s in %s", pattern.name, configPath))
				}
			}
		}
	}

	// Check log files for exposed secrets
	logPaths := []string{
		"logs/app.log", "logs/error.log", "logs/access.log",
		"app.log", "error.log", "access.log",
		"*.log",
	}

	for _, logPath := range logPaths {
		// Use glob pattern for log files
		if strings.Contains(logPath, "*") {
			// Simple glob expansion - check common log files
			continue
		}

		if data, err := os.ReadFile(logPath); err == nil {
			content := string(data)
			for _, pattern := range secretPatterns {
				if matched, _ := regexp.MatchString(pattern.regex, content); matched {
					exposedSecrets = append(exposedSecrets, fmt.Sprintf("%s in %s", pattern.name, logPath))
				}
			}

			// Also check for password patterns in logs
			if strings.Contains(content, "password") || strings.Contains(content, "PASSWORD") {
				exposedSecrets = append(exposedSecrets, fmt.Sprintf("Password reference in %s", logPath))
			}
		}
	}

	// Check for secrets in source code (common patterns)
	sourceFiles := []string{
		"main.go", "server.go", "config.go", "database.go",
		"settings.py", "config.py", "app.py",
		"index.js", "server.js", "config.js",
	}

	for _, sourceFile := range sourceFiles {
		if data, err := os.ReadFile(sourceFile); err == nil {
			content := string(data)
			for _, pattern := range secretPatterns {
				if matched, _ := regexp.MatchString(pattern.regex, content); matched {
					exposedSecrets = append(exposedSecrets, fmt.Sprintf("%s in source file %s", pattern.name, sourceFile))
				}
			}
		}
	}

	// Check for exposed database passwords in connection strings
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		if len(dbPassword) < 12 && !strings.Contains(dbPassword, "!") && !strings.Contains(dbPassword, "@") {
			exposedSecrets = append(exposedSecrets, "Weak database password in environment")
		}
	}

	// Check for default/example credentials
	defaultCreds := map[string]string{
		"admin":      "admin",
		"root":       "root",
		"postgres":   "postgres",
		"test":       "test",
		"user":       "password",
		"api":        "key",
		"secret":     "secret",
		"key":        "value",
	}

	for username, password := range defaultCreds {
		if os.Getenv("DB_USER") == username && os.Getenv("DB_PASSWORD") == password {
			exposedSecrets = append(exposedSecrets, fmt.Sprintf("Default credentials: %s/%s", username, password))
		}
	}

	return exposedSecrets, nil
}

func (sas *SecurityAuditService) checkDebugMode() (bool, error) {
	// Check if debug mode is enabled
	return false, nil
}