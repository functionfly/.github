package compliance

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// EmailServiceValidator interface for validating email service configuration
type EmailServiceValidator interface {
	ValidateConfiguration() error
}

// EncryptionManagerValidator interface for validating encryption configuration
type EncryptionManagerValidator interface {
	IsEncryptionEnabled() bool
	GetEncryptionStatus() map[string]interface{}
}

// BreachNotificationServiceValidator interface for validating breach notification service
type BreachNotificationServiceValidator interface {
	ValidateNotificationSetup() []string
	IsInitialized() bool
}

// SLAMonitorValidator interface for validating SLA monitoring
type SLAMonitorValidator interface {
	IsSLAMonitoringEnabled() bool
	GetSLAMetrics() map[string]interface{}
}

// GDPRChecker implements GDPR compliance checking
type GDPRChecker struct {
	db             storage.Repository
	logger         *logrus.Logger
	emailSvc       EmailServiceValidator         // Optional email service validator
	encryptMgr     EncryptionManagerValidator    // Optional encryption manager validator
	breachSvc      BreachNotificationServiceValidator // Optional breach notification service validator
	slaMonitor     SLAMonitorValidator           // Optional SLA monitor validator
}

// NewGDPRChecker creates a new GDPR compliance checker
func NewGDPRChecker(db storage.Repository, logger *logrus.Logger) *GDPRChecker {
	return &GDPRChecker{
		db:         db,
		logger:     logger,
		emailSvc:   nil, // Optional - can be set later if email validation is needed
		encryptMgr: nil, // Optional - can be set later if encryption validation is needed
		breachSvc:  nil, // Optional - can be set later if breach notification validation is needed
		slaMonitor: nil, // Optional - can be set later if SLA monitoring validation is needed
	}
}

// NewGDPRCheckerWithEmail creates a new GDPR compliance checker with email service validation
func NewGDPRCheckerWithEmail(db storage.Repository, logger *logrus.Logger, emailSvc EmailServiceValidator) *GDPRChecker {
	return &GDPRChecker{
		db:         db,
		logger:     logger,
		emailSvc:   emailSvc,
		encryptMgr: nil, // Optional
		breachSvc:  nil, // Optional
		slaMonitor: nil, // Optional
	}
}

// NewGDPRCheckerWithValidators creates a new GDPR compliance checker with all validators
func NewGDPRCheckerWithValidators(db storage.Repository, logger *logrus.Logger, emailSvc EmailServiceValidator, encryptMgr EncryptionManagerValidator) *GDPRChecker {
	return &GDPRChecker{
		db:         db,
		logger:     logger,
		emailSvc:   emailSvc,
		encryptMgr: encryptMgr,
		breachSvc:  nil, // Optional
		slaMonitor: nil, // Optional
	}
}

// NewGDPRCheckerWithAllValidators creates a new GDPR compliance checker with all available validators
func NewGDPRCheckerWithAllValidators(db storage.Repository, logger *logrus.Logger, emailSvc EmailServiceValidator, encryptMgr EncryptionManagerValidator, breachSvc BreachNotificationServiceValidator, slaMonitor SLAMonitorValidator) *GDPRChecker {
	return &GDPRChecker{
		db:         db,
		logger:     logger,
		emailSvc:   emailSvc,
		encryptMgr: encryptMgr,
		breachSvc:  breachSvc,
		slaMonitor: slaMonitor,
	}
}

// CheckCompliance performs GDPR compliance checks
func (g *GDPRChecker) CheckCompliance(ctx context.Context) []ComplianceIssue {
	vulnerabilities := []ComplianceIssue{}

	// GDPR (General Data Protection Regulation) compliance checks
	checks := []ComplianceCheck{
		{
			Title:       "Data Protection Officer Not Designated",
			Description: "No Data Protection Officer (DPO) designated for GDPR compliance",
			Severity:    "medium",
			CheckFunc: func() bool {
				// Check if there's a designated DPO (Data Protection Officer)
				// GDPR requires DPO designation for certain organizations

				// 1. Check for users with DPO role
				users, err := g.db.ListUsers()
				if err != nil {
					g.logger.WithError(err).Error("Failed to check for DPO designation in users")
					return true // Can't verify - assume vulnerability
				}

				dpoRoles := []string{"dpo", "data-protection-officer", "data_protection_officer", "gdpr-officer"}
				hasDPORole := false

				for _, user := range users {
					userRoleLower := strings.ToLower(user.Role)
					for _, dpoRole := range dpoRoles {
						if strings.Contains(userRoleLower, dpoRole) {
							hasDPORole = true
							break
						}
					}
					if hasDPORole {
						break
					}
				}

				if hasDPORole {
					return false // DPO designated via user role
				}

				// 2. Check for DPO designation in configuration/documentation
				dpoConfigFiles := []string{
					"config/dpo.json",
					"config/gdpr.json",
					"docs/dpo.md",
					"docs/gdpr/dpo.md",
					"compliance/dpo.md",
					"policies/dpo.md",
					"GDPR_DPO.md",
					"DPO.md",
				}

				for _, configFile := range dpoConfigFiles {
					if _, err := os.Stat(configFile); err == nil {
						return false // DPO configuration/documentation found
					}
				}

				// 3. Check environment variables for DPO designation
				dpoEnvVars := []string{
					"DPO_EMAIL",
					"DPO_NAME",
					"GDPR_DPO_EMAIL",
					"DATA_PROTECTION_OFFICER",
				}

				for _, env := range os.Environ() {
					envUpper := strings.ToUpper(env)
					for _, dpoVar := range dpoEnvVars {
						if strings.HasPrefix(envUpper, dpoVar+"=") {
							return false // DPO designated via environment variable
						}
					}
				}

				// 4. Check audit events for DPO-related activities
				events, err := g.db.ListAuditEvents(50, 0)
				if err == nil {
					dpoActivities := 0
					for _, event := range events {
						actionLower := strings.ToLower(event.Action)
						if strings.Contains(actionLower, "dpo") ||
							strings.Contains(actionLower, "data.protection") ||
							strings.Contains(actionLower, "gdpr") {
							dpoActivities++
						}
					}
					if dpoActivities > 0 {
						return false // DPO activities detected in audit log
					}
				}

				// No DPO designation found - this may be a GDPR violation
				// Note: Not all organizations require a DPO, but if they process personal data at scale, they do
				return true
			},
		},
		{
			Title:       "Data Processing Records Not Maintained",
			Description: "Records of processing activities are not properly maintained",
			Severity:    "medium",
			CheckFunc: func() bool {
				// Check if data processing records are maintained
				// GDPR requires documentation of processing activities

				// 1. Check for data processing documentation files
				processingDocFiles := []string{
					"docs/data-processing.md",
					"docs/processing-activities.md",
					"compliance/data-processing.md",
					"policies/data-processing.md",
					"GDPR_PROCESSING.md",
					"DATA_PROCESSING_RECORDS.md",
					"privacy/processing-records.md",
					"gdpr/processing-activities.md",
				}

				for _, docFile := range processingDocFiles {
					if _, err := os.Stat(docFile); err == nil {
						return false // Processing records documentation found
					}
				}

				// 2. Check for processing records in configuration
				processingConfigFiles := []string{
					"config/processing-activities.json",
					"config/data-processing.json",
					"privacy/processing.json",
					"gdpr/processing.json",
				}

				for _, configFile := range processingConfigFiles {
					if _, err := os.Stat(configFile); err == nil {
						return false // Processing records configuration found
					}
				}

				// 3. Check audit events for data processing activities
				events, err := g.db.ListAuditEvents(100, 0)
				if err != nil {
					g.logger.WithError(err).Error("Failed to check audit events for processing records")
					return true // Can't verify - assume vulnerability
				}

				// Look for data processing related activities
				processingActivities := 0
				dataProcessingTerms := []string{
					"process", "processing", "personal.data", "user.data", "gdpr",
					"data.collection", "data.storage", "data.analysis",
				}

				for _, event := range events {
					actionLower := strings.ToLower(event.Action)
					resourceLower := strings.ToLower(event.ResourceType)

					for _, term := range dataProcessingTerms {
						if strings.Contains(actionLower, term) ||
							strings.Contains(resourceLower, term) {
							processingActivities++
							break
						}
					}
				}

				// If we have processing activities in audit logs, records might be maintained
				if processingActivities >= 5 { // Reasonable threshold for active processing
					return false // Processing activities detected in audit trail
				}

				// 4. Check for data processing inventory/log files
				processingLogFiles := []string{
					"logs/processing-activities.log",
					"logs/data-processing.log",
					"audit/processing-records.log",
					"privacy/processing-log.json",
				}

				for _, logFile := range processingLogFiles {
					if _, err := os.Stat(logFile); err == nil {
						return false // Processing records log found
					}
				}

				// 5. Check environment variables for processing configuration
				processingEnvVars := []string{
					"DATA_PROCESSING_RECORDS",
					"PROCESSING_ACTIVITIES",
					"GDPR_PROCESSING_LOG",
					"PRIVACY_PROCESSING",
				}

				for _, env := range os.Environ() {
					envUpper := strings.ToUpper(env)
					for _, procVar := range processingEnvVars {
						if strings.HasPrefix(envUpper, procVar+"=") {
							return false // Processing records configured via environment
						}
					}
				}

				// No evidence of data processing records maintenance found
				return true
			},
		},
		{
			Title:       "Data Subject Rights Not Implemented",
			Description: "Data subject rights (access, rectification, erasure) are not properly implemented",
			Severity:    "high",
			CheckFunc: func() bool {
				// Check if data subject rights are implemented
				// GDPR requires implementation of data subject rights

				// 1. Check for GDPR rights API endpoints in routes
				gdprRightsEndpoints := []string{
					"/api/gdpr/access",
					"/api/gdpr/rectification",
					"/api/gdpr/erasure",
					"/api/gdpr/portability",
					"/api/privacy/access",
					"/api/privacy/delete",
					"/api/privacy/export",
					"/api/user/data-access",
					"/api/user/delete",
					"/api/user/export",
				}

				// Check if routes file contains GDPR rights endpoints
				if routesData, err := os.ReadFile("internal/api/routes.go"); err == nil {
					routesContent := string(routesData)
					gdprEndpointsFound := 0

					for _, endpoint := range gdprRightsEndpoints {
						if strings.Contains(routesContent, endpoint) {
							gdprEndpointsFound++
						}
					}

					if gdprEndpointsFound >= 3 { // At least access, rectification, and erasure
						return false // GDPR rights endpoints found
					}
				}

				// 2. Check for GDPR rights handler files
				gdprHandlerFiles := []string{
					"internal/api/handlers/gdpr/",
					"internal/api/handlers/privacy/",
					"internal/api/handlers/user/data_access.go",
					"internal/api/handlers/user/delete.go",
					"internal/api/handlers/gdpr/access.go",
					"internal/api/handlers/gdpr/delete.go",
					"internal/api/handlers/privacy/access.go",
					"internal/api/handlers/privacy/delete.go",
				}

				for _, handlerFile := range gdprHandlerFiles {
					if _, err := os.Stat(handlerFile); err == nil {
						return false // GDPR rights handlers found
					}
				}

				// 3. Check for GDPR rights in documentation
				gdprRightsDocs := []string{
					"docs/gdpr/rights.md",
					"docs/privacy/rights.md",
					"privacy/rights.md",
					"gdpr/data-subject-rights.md",
					"GDPR_RIGHTS.md",
					"PRIVACY_RIGHTS.md",
				}

				for _, docFile := range gdprRightsDocs {
					if _, err := os.Stat(docFile); err == nil {
						return false // GDPR rights documentation found
					}
				}

				// 4. Check audit events for GDPR rights usage
				events, err := g.db.ListAuditEvents(50, 0)
				if err == nil {
					gdprRightsActivities := 0
					gdprRightsTerms := []string{
						"gdpr.access", "gdpr.delete", "gdpr.export", "privacy.access",
						"data.access", "data.delete", "data.export", "user.delete",
						"right.access", "right.erasure", "right.portability",
					}

					for _, event := range events {
						actionLower := strings.ToLower(event.Action)
						for _, term := range gdprRightsTerms {
							if strings.Contains(actionLower, term) {
								gdprRightsActivities++
								break
							}
						}
					}

					if gdprRightsActivities > 0 {
						return false // GDPR rights activities detected in audit trail
					}
				}

				// 5. Check for GDPR rights configuration
				gdprRightsConfig := []string{
					"config/gdpr-rights.json",
					"config/privacy-rights.json",
					"privacy/config.json",
					"gdpr/config.json",
				}

				for _, configFile := range gdprRightsConfig {
					if _, err := os.Stat(configFile); err == nil {
						return false // GDPR rights configuration found
					}
				}

				// No evidence of GDPR data subject rights implementation found
				return true
			},
		},
		{
			Title:       "Personal Data Not Properly Protected",
			Description: "Personal data is not adequately protected against unauthorized access",
			Severity:    "high",
			CheckFunc: func() bool {
				// Check data protection measures
				users, err := g.db.ListUsers()
				if err != nil {
					return true // Can't verify - assume vulnerability
				}

				// Check if user data is properly protected
				for _, user := range users {
					// Check for unencrypted sensitive data (basic check)
					if user.PasswordHash == "" {
						g.logger.WithField("user_id", user.ID).Warn("User found without password hash")
						return true // User without password hash
					}

					// Verify encryption is enabled and working
					if g.encryptMgr != nil {
						// Use the encryption manager for direct validation
						if !g.encryptMgr.IsEncryptionEnabled() {
							g.logger.Warn("Database encryption not enabled according to encryption manager")
							return true // Encryption not enabled
						}

						// Check encryption status for additional validation
						status := g.encryptMgr.GetEncryptionStatus()
						if shouldRotate, ok := status["should_rotate"].(bool); ok && shouldRotate {
							g.logger.Warn("Encryption keys require rotation")
							// This could be a warning rather than failure, depending on policy
						}
					} else {
						// Fallback to environment variable check
						if os.Getenv("DB_ENCRYPTION_ENABLED") != "true" {
							g.logger.Warn("Database encryption not enabled (checked via environment)")
							return true // Encryption not enabled
						}
					}

					// Check for proper password hash format (bcrypt/scrypt indicators)
					if !strings.HasPrefix(user.PasswordHash, "$2a$") &&
					   !strings.HasPrefix(user.PasswordHash, "$2b$") &&
					   !strings.HasPrefix(user.PasswordHash, "$2y$") {
						g.logger.WithField("user_id", user.ID).Warn("User password hash does not appear to be properly hashed")
						return true // Password not properly hashed
					}

					// Additional encryption verification could include:
					// - Checking if sensitive fields are encrypted at rest
					// - Verifying encryption key rotation schedule
					// - Testing encryption/decryption functionality
					// - Validating backup encryption
				}
				return false
			},
		},
		{
			Title:       "Data Breach Notification Process Not Established",
			Description: "No established process for notifying data breaches within 72 hours",
			Severity:    "medium",
			CheckFunc: func() bool {
				// Check if breach notification process is established
				// GDPR Article 33 requires notification within 72 hours

				// Check if security email is configured
				if os.Getenv("SECURITY_EMAIL") == "" {
					return true // Non-compliant - no security contact
				}

				// Check if DPO email is configured
				if os.Getenv("DPO_EMAIL") == "" {
					return true // Non-compliant - no DPO contact
				}

				// Check if compliance email is configured
				if os.Getenv("COMPLIANCE_EMAIL") == "" {
					return true // Non-compliant - no compliance contact
				}

				// Check if email service is properly configured
				if g.emailSvc != nil {
					// If we have access to the email service, validate it directly
					if err := g.emailSvc.ValidateConfiguration(); err != nil {
						g.logger.WithError(err).Warn("Email service configuration validation failed")
						return true // Non-compliant - email service misconfigured
					}
				} else {
					// Fallback to environment variable checks
					if os.Getenv("SMTP_HOST") == "" || os.Getenv("SMTP_PORT") == "" {
						return true // Non-compliant - email service not configured
					}

					// Check if from email is configured for breach notifications
					if os.Getenv("FROM_EMAIL") == "" {
						return true // Non-compliant - no sender email configured
					}
				}

				// Production-grade validation checks

				// 1. Breach notification service is properly initialized
				if g.breachSvc != nil {
					if !g.breachSvc.IsInitialized() {
						g.logger.Warn("Breach notification service is not properly initialized")
						return true // Non-compliant - service not initialized
					}

					// Check for setup issues
					setupIssues := g.breachSvc.ValidateNotificationSetup()
					if len(setupIssues) > 0 {
						g.logger.WithField("issues", setupIssues).Warn("Breach notification setup issues found")
						return true // Non-compliant - setup issues exist
					}
				} else {
					g.logger.Warn("Breach notification service validator not available")
				}

				// 2. Email templates are configured
				// Check if breach notification templates exist (basic file check)
				if _, err := os.Stat("templates/breach_notification.html"); os.IsNotExist(err) {
					if _, err := os.Stat("internal/email/templates/breach_notification.html"); os.IsNotExist(err) {
						g.logger.Warn("Breach notification email templates not found")
						return true // Non-compliant - templates missing
					}
				}

				// 3. Escalation procedures are documented
				// Check for incident response documentation
				docPaths := []string{
					"docs/incident-response.md",
					"RUNBOOK.md",
					"docs/security/incident-response.md",
					"SECURITY.md",
				}
				hasIncidentResponseDoc := false
				for _, path := range docPaths {
					if _, err := os.Stat(path); err == nil {
						hasIncidentResponseDoc = true
						break
					}
				}
				if !hasIncidentResponseDoc {
					g.logger.Warn("Incident response documentation not found")
					return true // Non-compliant - documentation missing
				}

				// 4. 72-hour SLA monitoring is in place
				if g.slaMonitor != nil {
					if !g.slaMonitor.IsSLAMonitoringEnabled() {
						g.logger.Warn("72-hour SLA monitoring for breach notifications is not enabled")
						return true // Non-compliant - SLA monitoring missing
					}

					// Check SLA metrics
					slaMetrics := g.slaMonitor.GetSLAMetrics()
					if avgResponseTime, ok := slaMetrics["avg_response_time_hours"].(float64); ok && avgResponseTime > 72 {
						g.logger.WithField("avg_response_hours", avgResponseTime).Warn("Average breach response time exceeds 72-hour GDPR requirement")
						return true // Non-compliant - SLA not met
					}
				} else {
					// Fallback: Check for SLA monitoring environment variables
					if os.Getenv("SLA_MONITORING_ENABLED") != "true" {
						g.logger.Warn("SLA monitoring not configured")
						return true // Non-compliant - SLA monitoring not configured
					}
				}

				// 5. SMTP authentication credentials if required
				// This is already partially checked above, but let's be more thorough
				smtpHost := os.Getenv("SMTP_HOST")
				smtpPort := os.Getenv("SMTP_PORT")
				if smtpHost != "" && smtpPort != "" {
					// If SMTP is configured, ensure authentication is set up if required
					// Most SMTP servers require auth, but some local/dev servers don't
					smtpUser := os.Getenv("SMTP_USERNAME")
					smtpPass := os.Getenv("SMTP_PASSWORD")

					// Check if this looks like a production SMTP server that requires auth
					productionSMTPHosts := []string{
						"smtp.gmail.com",
						"smtp.outlook.com",
						"smtp.office365.com",
						"smtp.sendgrid.com",
						"smtp.mailgun.org",
						"smtp.amazon.com",
					}

					requiresAuth := false
					for _, host := range productionSMTPHosts {
						if strings.Contains(smtpHost, host) {
							requiresAuth = true
							break
						}
					}

					// Also check common port numbers that typically require auth
					if port, err := strconv.Atoi(smtpPort); err == nil {
						if port == 587 || port == 465 || port == 25 {
							requiresAuth = true
						}
					}

					if requiresAuth && (smtpUser == "" || smtpPass == "") {
						g.logger.Warn("SMTP server appears to require authentication but credentials not configured")
						return true // Non-compliant - auth credentials missing for production SMTP
					}
				}

				return false // Compliant - all breach notification requirements met
			},
		},
	}

	for _, check := range checks {
		if check.CheckFunc() {
			vulnerabilities = append(vulnerabilities, ComplianceIssue{
				ID:          generateVulnID(),
				Title:       check.Title,
				Description: check.Description,
				Severity:    check.Severity,
				Category:    "compliance",
				Component:   "gdpr",
				Status:      "open",
				Remediation: "Implement required GDPR data protection measures",
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	return vulnerabilities
}
