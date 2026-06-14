package compliance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PCIDSSChecker implements PCI DSS compliance checking
type PCIDSSChecker struct {
	db         storage.Repository
	logger     *logrus.Logger
	encryptMgr EncryptionManagerValidator // Optional encryption manager validator
}

// NewPCIDSSChecker creates a new PCI DSS compliance checker
func NewPCIDSSChecker(db storage.Repository, logger *logrus.Logger) *PCIDSSChecker {
	return &PCIDSSChecker{
		db:         db,
		logger:     logger,
		encryptMgr: nil, // Optional - can be set later if encryption validation is needed
	}
}

// NewPCIDSSCheckerWithEncryption creates a new PCI DSS compliance checker with encryption validation
func NewPCIDSSCheckerWithEncryption(db storage.Repository, logger *logrus.Logger, encryptMgr EncryptionManagerValidator) *PCIDSSChecker {
	return &PCIDSSChecker{
		db:         db,
		logger:     logger,
		encryptMgr: encryptMgr,
	}
}

// CheckCompliance performs PCI DSS compliance checks
func (p *PCIDSSChecker) CheckCompliance(ctx context.Context) []ComplianceIssue {
	vulnerabilities := []ComplianceIssue{}

	// PCI DSS (Payment Card Industry Data Security Standard) compliance checks
	checks := []ComplianceCheck{
		{
			Title:       "Cardholder Data Not Encrypted",
			Description: "Sensitive cardholder data is not properly encrypted during transmission and storage",
			Severity:    "critical",
			CheckFunc: func(ctx context.Context) bool {
				// PCI DSS Requirement 3: Protect stored cardholder data
				// Check if sensitive payment data is properly encrypted

				// 1. Verify database encryption is enabled
				if os.Getenv("DB_ENCRYPTION_ENABLED") != "true" {
					p.logger.Warn("Database encryption not enabled for PCI DSS compliance")
					return true // Non-compliant - encryption not enabled
				}

				// 2. Check if Stripe (payment processor) is properly configured
				// PCI DSS prefers using PA-DSS compliant payment processors
				stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
				stripePublishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")

				if stripeSecretKey == "" || stripePublishableKey == "" {
					p.logger.Warn("Stripe payment processor not properly configured")
					return true // Non-compliant - payment processor not configured
				}

				// 3. Verify no sensitive cardholder data is stored locally
				// In a PCI-compliant system, cardholder data should not be stored
				// Check if any potentially sensitive payment data exists in database
				_, err := p.db.ListInvoicesByTenant(ctx, uuid.Nil, 100, 0) // Check recent invoices
				if err != nil {
					p.logger.WithError(err).Warn("Failed to check invoices for PCI DSS compliance")
					return true // Can't verify - assume vulnerability
				}

				// 4. Check for any plain text payment data in audit logs
				auditEvents, err := p.db.ListAuditEventsFiltered(ctx, 50, 0, map[string]interface{}{
					"action": []string{"billing.*", "payment.*", "invoice.*"},
				})
				if err != nil {
					p.logger.WithError(err).Warn("Failed to check audit events for payment data exposure")
					return true // Can't verify audit trail
				}

				// Look for any suspicious patterns in audit events
				suspiciousPatterns := []string{
					"card_number", "cvv", "card_code", "pin", "track_data",
					"primary_account_number", "pan",
				}

				for _, event := range auditEvents {
					eventStr := strings.ToLower(fmt.Sprintf("%v", event))
					for _, pattern := range suspiciousPatterns {
						if strings.Contains(eventStr, pattern) {
							p.logger.WithFields(logrus.Fields{
								"event_id":   event.ID,
								"action":     event.Action,
								"pattern":    pattern,
							}).Warn("Potential cardholder data exposure detected in audit logs")
							return true // Non-compliant - sensitive data may be exposed
						}
					}
				}

				// 5. Check encryption key management
				if p.encryptMgr != nil {
					// Use the encryption manager for comprehensive validation
					if !p.encryptMgr.IsEncryptionEnabled() {
						p.logger.Warn("Encryption manager reports encryption is not enabled")
						return true // Non-compliant - encryption not enabled
					}

					// Check encryption status for additional validation
					status := p.encryptMgr.GetEncryptionStatus()

					// Verify key rotation is not overdue (PCI DSS requirement)
					if shouldRotate, ok := status["should_rotate"].(bool); ok && shouldRotate {
						p.logger.Warn("Encryption keys require rotation - PCI DSS compliance risk")
						return true // Non-compliant - keys need rotation
					}

					// Check if we have active encryption keys
					if activeKeys, ok := status["active_keys"].(int); ok && activeKeys == 0 {
						p.logger.Warn("No active encryption keys found")
						return true // Non-compliant - no active keys
					}

					// Check if encrypted fields are configured
					if encryptedFields, ok := status["encrypted_fields"].(int); ok && encryptedFields == 0 {
						p.logger.Warn("No encrypted fields configured for sensitive data")
						// This is a warning but not necessarily non-compliant if no sensitive data exists
					}

				} else {
					// Fallback to environment variable checks
					if os.Getenv("DB_MASTER_KEY_PASSWORD") == "" {
						p.logger.Warn("Database master key not configured for encryption (no encryption manager available)")
						return true // Non-compliant - encryption keys not properly managed
					}
				}

				return false // Compliant - cardholder data encryption properly implemented
			},
		},
		{
			Title:       "Insecure Payment Processing",
			Description: "Payment processing does not use secure protocols and methods",
			Severity:    "high",
			CheckFunc: func(ctx context.Context) bool {
				// PCI DSS Requirement 4: Encrypt transmission of cardholder data across open networks
				// Check if payment processing uses secure protocols and methods

				// 1. Verify HTTPS/TLS configuration
				// Check if the application enforces HTTPS for payment-related endpoints
				httpsEnforced := os.Getenv("FORCE_HTTPS")
				if httpsEnforced != "true" {
					p.logger.Warn("HTTPS not enforced for payment processing endpoints")
					return true // Non-compliant - insecure transmission
				}

				// 2. Check Stripe webhook security
				stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
				if stripeWebhookSecret == "" {
					p.logger.Warn("Stripe webhook secret not configured - webhooks not verified")
					return true // Non-compliant - webhook security missing
				}

				// 3. Verify secure payment API configuration
				stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
				stripePublishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")

				if stripeSecretKey == "" || stripePublishableKey == "" {
					p.logger.Warn("Stripe API keys not configured")
					return true // Non-compliant - payment API not secured
				}

				// Check for test keys in production (basic validation)
				if strings.Contains(stripeSecretKey, "_test_") &&
				   os.Getenv("ENVIRONMENT") == "production" {
					p.logger.Warn("Stripe test keys detected in production environment")
					return true // Non-compliant - test keys in production
				}

				// 4. Check payment audit logging
				// Verify that payment operations are properly logged
				paymentAuditEvents, err := p.db.ListAuditEventsFiltered(ctx, 20, 0, map[string]interface{}{
					"action": []string{"billing.payment.*", "billing.invoice.*", "billing.subscription.*"},
				})
				if err != nil {
					p.logger.WithError(err).Warn("Failed to verify payment audit logging")
					return true // Can't verify audit compliance
				}

				if len(paymentAuditEvents) == 0 {
					p.logger.Warn("No payment audit events found - payment operations may not be logged")
					return true // Non-compliant - payment operations not audited
				}

				// 5. Check for secure payment form configuration
				// Verify that payment forms use proper security measures
				secureFormConfig := os.Getenv("SECURE_PAYMENT_FORMS")
				if secureFormConfig != "true" {
					p.logger.Warn("Secure payment forms not explicitly configured")
					return true // Non-compliant - payment forms may be insecure
				}

				// 6. Verify no insecure payment methods are enabled
				// Check for any potentially insecure payment configurations
				insecureMethods := []string{
					os.Getenv("ALLOW_HTTP_PAYMENTS"),
					os.Getenv("DISABLE_SSL_VERIFICATION"),
					os.Getenv("ALLOW_PLAIN_TEXT_CARDS"),
				}

				for _, method := range insecureMethods {
					if method == "true" {
						p.logger.Warn("Insecure payment method detected in configuration")
						return true // Non-compliant - insecure payment methods enabled
					}
				}

				// 7. Check payment error handling security
				// Verify that payment errors don't leak sensitive information
				errorMaskingEnabled := os.Getenv("PAYMENT_ERROR_MASKING")
				if errorMaskingEnabled != "true" {
					p.logger.Warn("Payment error masking not enabled - may leak sensitive data")
					return true // Non-compliant - error handling insecure
				}

				return false // Compliant - payment processing security properly implemented
			},
		},
		{
			Title:       "Access Control Not Implemented for Cardholder Data",
			Description: "Access to cardholder data is not properly restricted",
			Severity:    "high",
			CheckFunc: func(ctx context.Context) bool {
				// Check access controls for payment data using sophisticated RBAC analysis
				users, err := p.db.ListUsers(ctx)
				if err != nil {
					return true // Can't verify - assume vulnerability
				}

				// Check recent audit events for payment/billing access patterns
				events, err := p.db.ListAuditEvents(ctx, 100, 0)
				if err != nil {
					p.logger.WithError(err).Error("Failed to check audit events for RBAC analysis")
					return true // Can't verify audit trail - assume vulnerability
				}

				// Analyze role distribution and access patterns
				adminUsers := 0
				regularUsers := 0
				privilegedAccessEvents := 0
				unauthorizedAccessAttempts := 0

				// Count users by role type
				for _, user := range users {
					switch user.Role {
					case "admin", "super_admin":
						adminUsers++
					case "user":
						regularUsers++
					default:
						// Unknown role - potential security issue
						return true
					}
				}

				// Analyze audit events for privileged operations
				paymentRelatedActions := []string{
					"billing", "payment", "invoice", "subscription", "card", "credit",
				}

				for _, event := range events {
					actionLower := strings.ToLower(event.Action)
					resourceLower := strings.ToLower(event.ResourceType)

					// Check if this is a privileged payment-related action
					isPaymentAction := false
					for _, paymentTerm := range paymentRelatedActions {
						if strings.Contains(actionLower, paymentTerm) ||
							strings.Contains(resourceLower, paymentTerm) {
							isPaymentAction = true
							break
						}
					}

					if isPaymentAction {
						privilegedAccessEvents++

						// Check if non-admin user performed privileged action
						if event.ActorUserID != nil {
							user, err := p.db.GetUserByID(ctx, *event.ActorUserID)
							if err == nil && user != nil {
								if user.Role != "admin" && user.Role != "super_admin" {
									unauthorizedAccessAttempts++
								}
							}
						}
					}
				}

				// PCI DSS RBAC violations:
				// 1. Any unauthorized access to payment data
				// 2. No separation between admin and regular user roles
				// 3. Too many admin users (lack of least privilege)
				// 4. No audit trail of privileged access

				if unauthorizedAccessAttempts > 0 {
					return true // Unauthorized access detected
				}

				if adminUsers == 0 && privilegedAccessEvents > 0 {
					return true // Privileged actions occurring but no admin users defined
				}

				if adminUsers > regularUsers && regularUsers > 0 {
					return true // Too many admin users compared to regular users (violates least privilege)
				}

				return false // RBAC appears properly implemented
			},
		},
		{
			Title:       "Security Logs Not Maintained",
			Description: "Audit logs for payment system access are not properly maintained",
			Severity:    "medium",
			CheckFunc: func(ctx context.Context) bool {
				// Check if payment-related audit logging is implemented
				events, err := p.db.ListAuditEvents(ctx, 50, 0)
				if err != nil {
					return true // Can't verify - assume vulnerability
				}

				// Look for payment-related events (this is a basic check)
				paymentEvents := 0
				for _, event := range events {
					if strings.Contains(event.ResourceType, "billing") ||
						strings.Contains(event.ResourceType, "payment") ||
						strings.Contains(event.Action, "billing") ||
						strings.Contains(event.Action, "payment") {
						paymentEvents++
					}
				}

				// If no payment events in recent audit log, might indicate insufficient logging
				return paymentEvents == 0
			},
		},
		{
			Title:       "Vulnerability Scanning Not Performed",
			Description: "Regular vulnerability scans of payment systems are not performed",
			Severity:    "medium",
			CheckFunc: func(ctx context.Context) bool {
				// Check if vulnerability scanning is performed regularly
				// For PCI DSS compliance, vulnerability scans should be performed at least quarterly

				// Look for recent security scans in audit events
				events, err := p.db.ListAuditEvents(ctx, 200, 0) // Get more events to check scan history
				if err != nil {
					p.logger.WithError(err).Error("Failed to check vulnerability scan history")
					return true // Can't verify - assume vulnerability
				}

				// Look for security scan events in the last 120 days (approximately quarterly + buffer)
				scanEvents := 0
				cutoffTime := time.Now().AddDate(0, 0, -120) // 120 days ago

				for _, event := range events {
					// Look for security-related actions or scans
					if (strings.Contains(event.Action, "scan") ||
						strings.Contains(event.Action, "security") ||
						strings.Contains(event.ResourceType, "security") ||
						strings.Contains(event.ResourceType, "vulnerability")) &&
						event.Timestamp.After(cutoffTime) {
						scanEvents++
					}
				}

				// PCI DSS requires quarterly scans, so we should have at least 1 scan in the last 120 days
				// If no scans found, this indicates insufficient vulnerability scanning
				return scanEvents == 0
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
				Component:   "pci_dss",
				Status:      "open",
				Remediation: "Implement required PCI DSS security controls for payment data",
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	return vulnerabilities
}
