package compliance

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// SOC2Checker implements SOC2 compliance checking
type SOC2Checker struct {
	db     storage.Repository
	logger *logrus.Logger
}

// NewSOC2Checker creates a new SOC2 compliance checker
func NewSOC2Checker(db storage.Repository, logger *logrus.Logger) *SOC2Checker {
	return &SOC2Checker{
		db:     db,
		logger: logger,
	}
}

// CheckCompliance performs SOC2 compliance checks
func (s *SOC2Checker) CheckCompliance(ctx context.Context) []ComplianceIssue {
	vulnerabilities := []ComplianceIssue{}

	// SOC 2 Trust Services Criteria checks
	checks := []ComplianceCheck{
		{
			Title:       "Access Control Not Properly Configured",
			Description: "Multi-factor authentication not enforced for all administrative access",
			Severity:    "high",
			CheckFunc: func(ctx context.Context) bool {
				// Check if any admin users don't have MFA enabled
				users, err := s.db.ListUsers(ctx)
				if err != nil {
					s.logger.WithError(err).Error("Failed to list users for MFA compliance check")
					return false // Don't report as vulnerability if we can't check
				}

				for _, user := range users {
					// Check if user has admin role
					if user.Role == "admin" || user.Role == "super_admin" {
						// Admin user found without MFA enabled - this is a vulnerability
						if !user.MFAEnabled {
							return true
						}
					}
				}
				return false // All admin users have MFA enabled
			},
		},
		{
			Title:       "Audit Logging Insufficient",
			Description: "Audit logs do not capture all required security events",
			Severity:    "medium",
			CheckFunc: func(ctx context.Context) bool {
				// Check if audit logs capture required security events
				// For SOC 2, we need to log: authentication events, admin actions, user management, permission changes

				// Get recent audit events to check what events are being logged
				events, err := s.db.ListAuditEvents(ctx, 100, 0) // Get last 100 events
				if err != nil {
					s.logger.WithError(err).Error("Failed to list audit events for compliance check")
					return true // Can't verify - assume insufficient logging
				}

				// Check if we have any events at all
				if len(events) == 0 {
					return true // No audit events found
				}

				// Check for required security event types
				hasAdminEvents := false
				hasUserMgmtEvents := false

				for _, event := range events {
					switch event.ResourceType {
					case "tenant", "system", "deployment":
						hasAdminEvents = true
					}

					// Check for user management actions
					if event.ResourceType == "user" && (event.Action == "user.create" || event.Action == "user.update" || event.Action == "user.delete") {
						hasUserMgmtEvents = true
					}
				}

				// For SOC 2 compliance, we should have at least admin events and user management events
				requiredEventsPresent := hasAdminEvents && hasUserMgmtEvents

				// If we don't have the required events, this is a vulnerability
				return !requiredEventsPresent
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
				Component:   "soc2",
				Status:      "open",
				Remediation: "Implement required SOC 2 controls",
				Discovered:  time.Now(),
				Updated:     time.Now(),
			})
		}
	}

	return vulnerabilities
}
