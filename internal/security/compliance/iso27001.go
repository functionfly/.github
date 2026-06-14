package compliance

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// ISO27001Checker implements ISO 27001 compliance checking
type ISO27001Checker struct {
	db     storage.Repository
	logger *logrus.Logger
}

// NewISO27001Checker creates a new ISO 27001 compliance checker
func NewISO27001Checker(db storage.Repository, logger *logrus.Logger) *ISO27001Checker {
	return &ISO27001Checker{
		db:     db,
		logger: logger,
	}
}

// CheckCompliance performs ISO 27001 compliance checks
func (i *ISO27001Checker) CheckCompliance(ctx context.Context) []ComplianceIssue {
	vulnerabilities := []ComplianceIssue{}

	// ISO 27001 Information Security Management System checks
	checks := []ComplianceCheck{
		{
			Title:       "Information Security Policy Not Defined",
			Description: "No documented information security policy found",
			Severity:    "high",
			CheckFunc: func(ctx context.Context) bool {
				// Check if security policy files exist
				// Look for common security policy document names and locations
				policyFiles := []string{
					"SECURITY.md",
					"security-policy.md",
					"SECURITY_POLICY.md",
					"docs/security-policy.md",
					"docs/SECURITY.md",
					"security/README.md",
					"compliance/security-policy.md",
					"policies/security-policy.md",
				}

				// Check if any of these policy files exist
				for _, policyFile := range policyFiles {
					if _, err := os.Stat(policyFile); err == nil {
						// Policy file exists
						return false // No vulnerability - policy is documented
					}
				}

				// Check if there's a security directory with policy content
				if info, err := os.Stat("security"); err == nil && info.IsDir() {
					entries, err := os.ReadDir("security")
					if err == nil {
						for _, entry := range entries {
							if !entry.IsDir() && (strings.Contains(strings.ToLower(entry.Name()), "policy") ||
								strings.Contains(strings.ToLower(entry.Name()), "security")) {
								return false // Found security policy document
							}
						}
					}
				}

				// Check if there's a docs/policies directory
				if info, err := os.Stat("docs/policies"); err == nil && info.IsDir() {
					entries, err := os.ReadDir("docs/policies")
					if err == nil {
						for _, entry := range entries {
							if !entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "security") {
								return false // Found security policy in policies directory
							}
						}
					}
				}

				return true // No security policy documentation found - vulnerability
			},
		},
		{
			Title:       "Access Control Policy Not Implemented",
			Description: "Access control policies are not properly implemented across all systems",
			Severity:    "high",
			CheckFunc: func(ctx context.Context) bool {
				// Check if role-based access control is properly configured
				users, err := i.db.ListUsers(ctx)
				if err != nil {
					return true // Can't verify - assume vulnerability
				}

				// Check for users with undefined or excessive roles
				for _, user := range users {
					if user.Role == "" {
						return true // User without defined role
					}
					if user.Role != "admin" && user.Role != "super_admin" && user.Role != "user" {
						return true // Unexpected role - potential security issue
					}
				}
				return false
			},
		},
		{
			Title:       "Cryptographic Controls Not Implemented",
			Description: "Encryption is not properly implemented for sensitive data",
			Severity:    "high",
			CheckFunc: func(ctx context.Context) bool {
				// Check if encryption is enabled for sensitive data
				// This includes database encryption, password hashing, and API encryption

				// 1. Check password hashing implementation
				users, err := i.db.ListUsers(ctx)
				if err != nil {
					i.logger.WithError(err).Error("Failed to check user password encryption")
					return true // Can't verify - assume vulnerability
				}

				// Check that all users have properly hashed passwords (not plain text)
				for _, user := range users {
					if user.PasswordHash == "" {
						return true // User without password hash
					}
					// Check for common weak hashing patterns
					if strings.HasPrefix(user.PasswordHash, "$2a$") ||
					   strings.HasPrefix(user.PasswordHash, "$2b$") ||
					   strings.HasPrefix(user.PasswordHash, "$2y$") {
						// bcrypt is good - continue checking
					} else if strings.HasPrefix(user.PasswordHash, "$argon2") {
						// argon2 is good - continue checking
					} else if len(user.PasswordHash) < 20 {
						// Very short hash likely indicates weak/no hashing
						return true
					}
				}

				// 2. Check if database encryption configuration exists
				// Look for database encryption settings or environment variables
				encryptionIndicators := []string{
					"DATABASE_ENCRYPTION",
					"DB_ENCRYPTION",
					"ENCRYPTION_KEY",
					"DATABASE_SSL",
					"DB_SSL",
				}

				hasEncryption := false
				for _, env := range os.Environ() {
					for _, indicator := range encryptionIndicators {
						if strings.Contains(strings.ToUpper(env), indicator) {
							hasEncryption = true
							break
						}
					}
					if hasEncryption {
						break
					}
				}

				// 3. Check for TLS/HTTPS configuration files
				tlsFiles := []string{
					"tls/cert.pem",
					"tls/key.pem",
					"certs/server.crt",
					"certs/server.key",
					"ssl/cert.pem",
					"ssl/private.key",
				}

				hasTLS := false
				for _, tlsFile := range tlsFiles {
					if _, err := os.Stat(tlsFile); err == nil {
						hasTLS = true
						break
					}
				}

				// For ISO 27001 compliance, we need proper password hashing
				// Database encryption and TLS are additional protections
				passwordsSecure := true // We already checked this above

				// If passwords are secure and we have some form of encryption/TLS, we're good
				return !(passwordsSecure && (hasEncryption || hasTLS))
			},
		},
	}

	for _, check := range checks {
		if check.CheckFunc(ctx) {
			vulnerabilities = append(vulnerabilities, ComplianceIssue{
				ID:          generateVulnID(),
				Title:       check.Title,
				Description: check.Description,
				Severity:    check.Severity,
				Category:    "compliance",
				Component:   "iso27001",
				Status:      "open",
				Remediation: "Implement required ISO 27001 information security controls",
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	return vulnerabilities
}