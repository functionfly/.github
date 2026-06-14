package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// BreachNotificationService handles GDPR Article 33 breach notifications
type BreachNotificationService struct {
	emailSvc email.Service
	repo     storage.Repository
	logger   *logrus.Logger
}

// BreachDetails contains information about a data breach
type BreachDetails struct {
	Type           string    `json:"type"`
	DetectionTime  time.Time `json:"detectionTime"`
	AffectedUsers  int       `json:"affectedUsers"`
	RiskLevel      string    `json:"riskLevel"` // "low", "medium", "high"
	Description    string    `json:"description"`
	AffectedData   []string  `json:"affectedData"` // types of personal data affected
	RemedialAction string    `json:"remedialAction"`
}

// NewBreachNotificationService creates a new breach notification service
func NewBreachNotificationService(emailSvc email.Service, repo storage.Repository, logger *logrus.Logger) *BreachNotificationService {
	return &BreachNotificationService{
		emailSvc: emailSvc,
		repo:     repo,
		logger:   logger,
	}
}

// NotifyBreach sends breach notifications to required recipients
// GDPR Article 33 requires notification within 72 hours of becoming aware of the breach
func (b *BreachNotificationService) NotifyBreach(ctx context.Context, breach BreachDetails) error {
	b.logger.WithFields(logrus.Fields{
		"breach_type":    breach.Type,
		"affected_users": breach.AffectedUsers,
		"risk_level":     breach.RiskLevel,
	}).Warn("Initiating GDPR Article 33 breach notification process")

	// Get notification recipients
	recipients, err := b.getNotificationRecipients(ctx)
	if err != nil {
		b.logger.WithError(err).Error("Failed to get breach notification recipients")
		return fmt.Errorf("failed to get notification recipients: %w", err)
	}

	if len(recipients) == 0 {
		b.logger.Error("No breach notification recipients configured")
		return fmt.Errorf("no breach notification recipients configured")
	}

	// Prepare breach details for email
	breachDetails := map[string]interface{}{
		"type":          breach.Type,
		"detectionTime": breach.DetectionTime.Format(time.RFC3339),
		"affectedUsers": breach.AffectedUsers,
		"riskLevel":     breach.RiskLevel,
		"description":   breach.Description,
		"affectedData":  breach.AffectedData,
		"remedialAction": breach.RemedialAction,
	}

	// Send notification emails
	if err := b.emailSvc.SendBreachNotification(recipients, breachDetails); err != nil {
		b.logger.WithError(err).Error("Failed to send breach notification emails")
		return fmt.Errorf("failed to send breach notification emails: %w", err)
	}

	// Log the notification in audit trail
	if err := b.logBreachNotification(breach, recipients); err != nil {
		b.logger.WithError(err).Warn("Failed to log breach notification in audit trail")
		// Don't return error here as the notification was sent successfully
	}

	b.logger.WithFields(logrus.Fields{
		"recipients_count": len(recipients),
		"breach_type":      breach.Type,
	}).Info("GDPR Article 33 breach notification sent successfully")

	return nil
}

// getNotificationRecipients returns the list of email addresses that should receive breach notifications
func (b *BreachNotificationService) getNotificationRecipients(ctx context.Context) ([]string, error) {
	var recipients []string

	// Primary security contact from environment
	if securityEmail := os.Getenv("SECURITY_EMAIL"); securityEmail != "" {
		recipients = append(recipients, securityEmail)
	}

	// Data Protection Officer
	if dpoEmail := os.Getenv("DPO_EMAIL"); dpoEmail != "" {
		recipients = append(recipients, dpoEmail)
	}

	// Compliance team
	if complianceEmail := os.Getenv("COMPLIANCE_EMAIL"); complianceEmail != "" {
		recipients = append(recipients, complianceEmail)
	}

	// Get admin users from database
	admins, err := b.repo.ListUsers(ctx)
	if err != nil {
		b.logger.WithError(err).Warn("Failed to get admin users for breach notification")
	} else {
		for _, user := range admins {
			// Check if user has admin role (admin, super_admin, or administrator)
			if user.Email != "" && b.isAdminUser(user) {
				recipients = append(recipients, user.Email)
			}
		}
	}

	// Remove duplicates
	recipients = b.removeDuplicates(recipients)

	if len(recipients) == 0 {
		return nil, fmt.Errorf("no breach notification recipients configured")
	}

	return recipients, nil
}

// isAdminUser checks if a user has admin privileges
func (b *BreachNotificationService) isAdminUser(user *storage.User) bool {
	if user == nil || user.Role == "" {
		return false
	}

	// Check for admin roles based on the User.Role field
	// Using strings.ToLower for case-insensitive comparison to handle potential data inconsistencies
	role := strings.ToLower(user.Role)
	switch role {
	case "admin", "super_admin", "administrator":
		return true
	default:
		return false
	}
}

// removeDuplicates removes duplicate email addresses from the slice
func (b *BreachNotificationService) removeDuplicates(emails []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, email := range emails {
		if !seen[email] {
			seen[email] = true
			result = append(result, email)
		}
	}

	return result
}

// logBreachNotification logs the breach notification in the audit trail
func (b *BreachNotificationService) logBreachNotification(breach BreachDetails, recipients []string) error {
	// Create audit event for breach notification
	auditEvent := &storage.AuditEvent{
		ActorEmail:   "system@functionfly.com",
		Action:       "gdpr.breach_notification",
		ResourceType: "compliance",
		RequestID:    "breach-" + time.Now().Format("20060102-150405"),
		IPAddress:    "system",
		UserAgent:    "breach-notification-service",
		BeforeState: map[string]interface{}{
			"breach_type":     breach.Type,
			"detection_time":  breach.DetectionTime,
			"affected_users":  breach.AffectedUsers,
			"risk_level":      breach.RiskLevel,
			"recipients":      recipients,
			"recipients_count": len(recipients),
			"notification_time": time.Now(),
		},
		Timestamp: time.Now(),
		Success:   true,
	}

	return b.repo.LogAuditEvent(nil, auditEvent)
}

// ValidateNotificationSetup checks if breach notification infrastructure is properly configured
func (b *BreachNotificationService) ValidateNotificationSetup() []string {
	var issues []string

	// Check if security email is configured
	if os.Getenv("SECURITY_EMAIL") == "" {
		issues = append(issues, "SECURITY_EMAIL environment variable not configured")
	}

	// Check if DPO email is configured
	if os.Getenv("DPO_EMAIL") == "" {
		issues = append(issues, "DPO_EMAIL environment variable not configured")
	}

	// Check if compliance email is configured
	if os.Getenv("COMPLIANCE_EMAIL") == "" {
		issues = append(issues, "COMPLIANCE_EMAIL environment variable not configured")
	}

	// Check if email service is available
	if b.emailSvc == nil {
		issues = append(issues, "Email service not configured")
	}

	// Check if repository is available
	if b.repo == nil {
		issues = append(issues, "Repository not available for audit logging")
	}

	// Check if logger is available
	if b.logger == nil {
		issues = append(issues, "Logger not configured")
	}

	return issues
}

// IsInitialized checks if the breach notification service is properly initialized
func (b *BreachNotificationService) IsInitialized() bool {
	// Check if all required components are available
	return b.emailSvc != nil && b.repo != nil && b.logger != nil
}