package security

import (
	"context"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// Database security scanning functions

func (sas *SecurityAuditService) scanDatabaseSecurity(ctx context.Context, config ScanConfig) ([]Vulnerability, error) {
	vulnerabilities := []Vulnerability{}

	// Check database encryption
	if noEncryption, err := sas.checkDatabaseEncryption(); err == nil && noEncryption {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Database Encryption Not Configured",
			Description: "Database data is not encrypted at rest",
			Severity:    "high",
			Category:    "crypto",
			Component:   "database",
			Status:      "open",
			Remediation: "Enable database encryption at rest using AES-256",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	// Check for weak passwords in database
	if weakPasswords, err := sas.checkDatabasePasswords(); err == nil && weakPasswords {
		vulnerabilities = append(vulnerabilities, Vulnerability{
			ID:          generateVulnID(),
			Title:       "Weak Database Passwords Detected",
			Description: "Database contains user accounts with weak passwords",
			Severity:    "high",
			Category:    "auth",
			Component:   "database",
			Status:      "open",
			Remediation: "Enforce strong password policies and hash passwords properly",
			Discovered:  time.Now(),
			Updated:     time.Now(),
		})
	}

	return vulnerabilities, nil
}

func (sas *SecurityAuditService) checkDatabaseEncryption() (bool, error) {
	// Check if database encryption is configured
	// Return true if encryption is NOT configured (vulnerability exists)

	// Check environment variables for SSL configuration
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" || sslMode == "disable" {
		return true, nil // SSL disabled - encryption vulnerability
	}

	// Try to cast the repository to PostgresDB to access encryption methods
	if postgresDB, ok := sas.db.(*storage.PostgresDB); ok {
		// Check if field-level encryption is enabled
		if !postgresDB.IsEncryptionEnabled() {
			return true, nil // Field encryption not enabled
		}
	} else {
		// If we can't cast, check via environment variable
		encryptionKey := os.Getenv("DB_ENCRYPTION_KEY")
		if encryptionKey == "" {
			return true, nil // No encryption key configured
		}
	}

	// Check for Transparent Data Encryption (TDE) or similar via database queries
	if postgresDB, ok := sas.db.(*storage.PostgresDB); ok {
		// Check SSL connection status using pg_stat_ssl
		var sslActive bool
		err := postgresDB.QueryRow(`
			SELECT COALESCE(bool_or(ssl), false)
			FROM pg_stat_ssl
			WHERE pid = pg_backend_pid()
		`).Scan(&sslActive)
		if err == nil && !sslActive {
			return true, nil // SSL not active on current connection
		}

		// Check if pgcrypto extension is available (for encryption functions)
		var cryptoAvailable bool
		err = postgresDB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto'
			)
		`).Scan(&cryptoAvailable)
		if err == nil && !cryptoAvailable {
			// pgcrypto not available - check if any encrypted fields exist
			var encryptedFieldCount int
			err = postgresDB.QueryRow(`
				SELECT COUNT(*) FROM information_schema.columns
				WHERE table_schema = 'public'
				AND column_name LIKE '%encrypted%' OR column_name LIKE '%cipher%'
			`).Scan(&encryptedFieldCount)
			if err == nil && encryptedFieldCount == 0 {
				return true, nil // No encrypted fields and no pgcrypto
			}
		}

		// Check for database-level encryption settings
		var encryptionMethod string
		err = postgresDB.QueryRow(`
			SELECT setting FROM pg_settings
			WHERE name = 'ssl' AND setting = 'on'
		`).Scan(&encryptionMethod)
		if err != nil {
			// SSL setting not found or not 'on'
			return true, nil // SSL not enabled at database level
		}
	}

	// Additional checks: look for encryption-related environment variables
	encryptionVars := []string{
		"DB_ENCRYPTION_ENABLED",
		"DATABASE_ENCRYPTION",
		"PG_ENCRYPTION_KEY",
	}

	for _, envVar := range encryptionVars {
		if value := os.Getenv(envVar); value != "" {
			if strings.ToLower(value) == "true" || value == "1" || value == "yes" {
				return false, nil // Encryption appears to be enabled
			}
		}
	}

	// If we reach here, encryption status is uncertain but SSL might be enabled
	// Return false (no vulnerability) since SSL provides transport encryption
	if sslMode == "require" || sslMode == "verify-ca" || sslMode == "verify-full" {
		return false, nil // SSL is properly configured
	}

	// Default: assume encryption is not configured
	return true, nil
}

func (sas *SecurityAuditService) checkDatabasePasswords() (bool, error) {
	// Check for weak passwords in database
	// Return true if weak passwords are detected (vulnerability exists)

	// Get all users from the database
	users, err := sas.db.ListUsers()
	if err != nil {
		return false, err // Can't check if we can't access users
	}

	weakPasswordCount := 0
	totalUsers := len(users)

	// Common weak password hashes (MD5, SHA1 of common passwords)
	weakHashes := map[string]bool{
		// MD5 hashes of common passwords
		"5f4dcc3b5aa765d61d8327deb882cf99": true, // password
		"e99a18c428cb38d5f260853678922e03": true, // abc123
		"25d55ad283aa400af464c76d713c07ad": true, // 123456
		"7c6a180b36896a0a8c02787eeafb0e4c": true, // 123456789
		"6cb75f652a9b52798eb6cf2201057c73": true, // 12345
		"46f94c8de14fb36680850768ff1b7f2a": true, // 1234567890
		"c21f969b5f03d33d43e04f8f136e7682": true, // admin
		"21232f297a57a5a743894a0e4a801fc3": true, // admin
		"d41d8cd98f00b204e9800998ecf8427e": true, // empty string
	}

	for _, user := range users {
		passwordHash := user.PasswordHash

		// Check for empty or null passwords
		if passwordHash == "" || passwordHash == "NULL" {
			// Check if user has social auth (might be acceptable)
			if user.Provider == nil || *user.Provider == "" {
				weakPasswordCount++
				continue
			}
		}

		// Check for very short hashes (indicates weak hashing)
		if len(passwordHash) < 32 {
			weakPasswordCount++
			continue
		}

		// Check for known weak password hashes
		hashLower := strings.ToLower(passwordHash)
		if weakHashes[hashLower] {
			weakPasswordCount++
			continue
		}

		// Check for unsalted MD5 hashes (32 chars, starts with common patterns)
		if len(passwordHash) == 32 && !strings.Contains(passwordHash, "$") {
			// Additional check: look for MD5 characteristics
			if matched, _ := regexp.MatchString("^[a-f0-9]{32}$", passwordHash); matched {
				weakPasswordCount++
				continue
			}
		}

		// Check for unsalted SHA1 hashes (40 chars)
		if len(passwordHash) == 40 && !strings.Contains(passwordHash, "$") {
			if matched, _ := regexp.MatchString("^[a-f0-9]{40}$", passwordHash); matched {
				weakPasswordCount++
				continue
			}
		}
	}

	// If more than 10% of users have weak passwords, or any admin users have weak passwords
	weakPercentage := float64(weakPasswordCount) / float64(totalUsers) * 100

	if weakPercentage >= 10 {
		return true, nil // Significant number of weak passwords
	}

	// Check specifically for admin users with weak passwords
	for _, user := range users {
		if user.Role == "admin" || user.Role == "super_admin" {
			if user.PasswordHash == "" ||
			   len(user.PasswordHash) < 32 ||
			   weakHashes[strings.ToLower(user.PasswordHash)] {
				return true, nil // Admin with weak password is critical
			}
		}
	}

	return false, nil // No significant weak password issues detected
}