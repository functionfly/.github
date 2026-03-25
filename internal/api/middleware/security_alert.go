package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/metrics"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// SecurityAlertMiddleware handles security alert checking and triggering
type SecurityAlertMiddleware struct {
	db           *sql.DB
	redisClient  *redis.Client
	alertHandler SecurityAlerter
}

// SecurityAlerter interface for triggering alerts
type SecurityAlerter interface {
	TriggerAlert(ctx context.Context, alertType, severity, ipAddress string, details map[string]interface{}) error
}

// SecurityAlertRule represents a security alert rule from the database
type SecurityAlertRule struct {
	ID                   string
	Name                 string
	AlertType            string
	Threshold            int
	WindowSeconds        int
	Severity             string
	IsEnabled            bool
	NotificationChannels []string
}

// AlertEvent represents an alert event to be logged/triggered
type AlertEvent struct {
	AlertRuleID   string                 `json:"alert_rule_id"`
	AlertType     string                 `json:"alert_type"`
	Severity      string                 `json:"severity"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Email         string                 `json:"email,omitempty"`
	Threshold     int                    `json:"threshold"`
	CurrentCount  int                    `json:"current_count"`
	WindowSeconds int                    `json:"window_seconds"`
	Details       map[string]interface{} `json:"details,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// NewSecurityAlertMiddleware creates a new security alert middleware
func NewSecurityAlertMiddleware(db *sql.DB, redisClient *redis.Client, alertHandler SecurityAlerter) *SecurityAlertMiddleware {
	return &SecurityAlertMiddleware{
		db:           db,
		redisClient:  redisClient,
		alertHandler: alertHandler,
	}
}

// RequireSecurityAlertCheck middleware that checks requests against alert rules
func (m *SecurityAlertMiddleware) RequireSecurityAlertCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract client information
		clientIP := extractClientIPFromRequest(r)
		userID := ""
		email := ""

		// Try to get user info from context
		if claims := GetUserFromContext(r); claims != nil {
			userID = claims.UserID.String()
			email = claims.Email
		}

		// Check failed login threshold
		if err := m.checkFailedLoginThreshold(ctx, clientIP, userID, email); err != nil {
			logrus.WithError(err).Warn("Failed to check failed login threshold")
		}

		// Check rate limit exceeded
		if err := m.checkRateLimitExceeded(ctx, clientIP); err != nil {
			logrus.WithError(err).Warn("Failed to check rate limit exceeded")
		}

		// Check IP blocked
		if err := m.checkIPBlocked(ctx, clientIP); err != nil {
			logrus.WithError(err).Warn("Failed to check IP blocked")
		}

		// Check suspicious activity
		if err := m.checkSuspiciousActivity(ctx, clientIP, userID); err != nil {
			logrus.WithError(err).Warn("Failed to check suspicious activity")
		}

		// Check session anomaly
		if userID != "" {
			if err := m.checkSessionAnomaly(ctx, clientIP, userID); err != nil {
				logrus.WithError(err).Warn("Failed to check session anomaly")
			}
		}

		next.ServeHTTP(w, r)
	}
}

// checkFailedLoginThreshold checks if failed login threshold is exceeded
func (m *SecurityAlertMiddleware) checkFailedLoginThreshold(ctx context.Context, ipAddress, userID, email string) error {
	metrics.RecordSecurityAlertCheck("failed_login_threshold")

	// Get enabled alert rules for failed_login_threshold
	rules, err := m.getEnabledAlertRules(ctx, "failed_login_threshold")
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	// Check Redis for failed login count
	for _, rule := range rules {
		key := fmt.Sprintf("security:failed_login:%s", ipAddress)
		count, err := m.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get failed login count: %w", err)
		}

		if count >= rule.Threshold {
			// Trigger alert
			details := map[string]interface{}{
				"ip_address":     ipAddress,
				"failed_attempts": count,
				"window_seconds":  rule.WindowSeconds,
			}

			if err := m.triggerAlert(ctx, rule, ipAddress, userID, email, details); err != nil {
				logrus.WithError(err).Error("Failed to trigger failed login threshold alert")
			}
		}
	}

	return nil
}

// checkRateLimitExceeded checks if rate limit is exceeded
func (m *SecurityAlertMiddleware) checkRateLimitExceeded(ctx context.Context, ipAddress string) error {
	metrics.RecordSecurityAlertCheck("rate_limit_exceeded")

	rules, err := m.getEnabledAlertRules(ctx, "rate_limit_exceeded")
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	// Check Redis for rate limit hits
	for _, rule := range rules {
		key := fmt.Sprintf("security:rate_limit:%s", ipAddress)
		count, err := m.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get rate limit count: %w", err)
		}

		if count >= rule.Threshold {
			details := map[string]interface{}{
				"ip_address":     ipAddress,
				"rate_limit_hits": count,
				"window_seconds":  rule.WindowSeconds,
			}

			if err := m.triggerAlert(ctx, rule, ipAddress, "", "", details); err != nil {
				logrus.WithError(err).Error("Failed to trigger rate limit exceeded alert")
			}
		}
	}

	return nil
}

// checkIPBlocked checks if IP has been blocked
func (m *SecurityAlertMiddleware) checkIPBlocked(ctx context.Context, ipAddress string) error {
	metrics.RecordSecurityAlertCheck("ip_blocked")

	rules, err := m.getEnabledAlertRules(ctx, "ip_blocked")
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	// Check Redis for blocked IP
	for _, rule := range rules {
		key := fmt.Sprintf("security:blocked:%s", ipAddress)
		blocked, err := m.redisClient.Exists(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("failed to check blocked IP: %w", err)
		}

		if blocked > 0 {
			details := map[string]interface{}{
				"ip_address":    ipAddress,
				"window_seconds": rule.WindowSeconds,
			}

			if err := m.triggerAlert(ctx, rule, ipAddress, "", "", details); err != nil {
				logrus.WithError(err).Error("Failed to trigger IP blocked alert")
			}
		}
	}

	return nil
}

// checkSuspiciousActivity checks for suspicious activity patterns
func (m *SecurityAlertMiddleware) checkSuspiciousActivity(ctx context.Context, ipAddress, userID string) error {
	metrics.RecordSecurityAlertCheck("suspicious_activity")

	rules, err := m.getEnabledAlertRules(ctx, "suspicious_activity")
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	// Check for suspicious patterns (multiple user agents, rapid requests, etc.)
	for _, rule := range rules {
		key := fmt.Sprintf("security:suspicious:%s:%s", ipAddress, userID)
		count, err := m.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get suspicious activity count: %w", err)
		}

		if count >= rule.Threshold {
			details := map[string]interface{}{
				"ip_address":      ipAddress,
				"suspicious_hits": count,
				"window_seconds":  rule.WindowSeconds,
			}

			if err := m.triggerAlert(ctx, rule, ipAddress, userID, "", details); err != nil {
				logrus.WithError(err).Error("Failed to trigger suspicious activity alert")
			}
		}
	}

	return nil
}

// checkSessionAnomaly checks for session anomalies
func (m *SecurityAlertMiddleware) checkSessionAnomaly(ctx context.Context, ipAddress, userID string) error {
	metrics.RecordSecurityAlertCheck("session_anomaly")

	rules, err := m.getEnabledAlertRules(ctx, "session_anomaly")
	if err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	if len(rules) == 0 {
		return nil
	}

	// Check for session anomalies (multiple IPs, unusual patterns)
	for _, rule := range rules {
		key := fmt.Sprintf("security:session_anomaly:%s", userID)
		count, err := m.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get session anomaly count: %w", err)
		}

		if count >= rule.Threshold {
			details := map[string]interface{}{
				"ip_address":    ipAddress,
				"user_id":       userID,
				"anomaly_count": count,
				"window_seconds": rule.WindowSeconds,
			}

			if err := m.triggerAlert(ctx, rule, ipAddress, userID, "", details); err != nil {
				logrus.WithError(err).Error("Failed to trigger session anomaly alert")
			}
		}
	}

	return nil
}

// getEnabledAlertRules retrieves enabled alert rules for a specific type
func (m *SecurityAlertMiddleware) getEnabledAlertRules(ctx context.Context, alertType string) ([]SecurityAlertRule, error) {
	query := `
		SELECT id, name, alert_type, threshold, window_seconds, severity,
		       is_enabled, notification_channels
		FROM security_alert_rules
		WHERE alert_type = $1 AND is_enabled = TRUE`

	rows, err := m.db.QueryContext(ctx, query, alertType)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []SecurityAlertRule
	for rows.Next() {
		var rule SecurityAlertRule
		var notificationChannels []byte

		if err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.AlertType,
			&rule.Threshold,
			&rule.WindowSeconds,
			&rule.Severity,
			&rule.IsEnabled,
			&notificationChannels,
		); err != nil {
			logrus.WithError(err).Warn("Failed to scan alert rule")
			continue
		}

		if notificationChannels != nil {
			if err := json.Unmarshal(notificationChannels, &rule.NotificationChannels); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal notification channels")
				rule.NotificationChannels = []string{}
			}
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// triggerAlert triggers an alert and logs it to the security_events table
func (m *SecurityAlertMiddleware) triggerAlert(ctx context.Context, rule SecurityAlertRule, ipAddress, userID, email string, details map[string]interface{}) error {
	metrics.RecordSecurityAlert(rule.AlertType, rule.Severity)

	// Log event to security_events table
	eventQuery := `
		INSERT INTO security_events (event_type, ip_address, user_id, email, details, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	var userIDPtr, emailPtr *string
	if userID != "" {
		userIDPtr = &userID
	}
	if email != "" {
		emailPtr = &email
	}

	_, err = m.db.ExecContext(ctx, eventQuery, rule.AlertType, ipAddress, userIDPtr, emailPtr, detailsJSON)
	if err != nil {
		logrus.WithError(err).Error("Failed to log security event")
	}

	// Trigger notification if handler supports it
	if m.alertHandler != nil {
		if err := m.alertHandler.TriggerAlert(ctx, rule.AlertType, rule.Severity, ipAddress, details); err != nil {
			logrus.WithError(err).Error("Failed to trigger alert notification")
		}
	}

	// Send notifications based on channels
	for _, channel := range rule.NotificationChannels {
		if err := m.sendNotification(ctx, channel, rule, ipAddress, userID, email, details); err != nil {
			logrus.WithError(err).WithField("channel", channel).Error("Failed to send notification")
		}
	}

	logrus.WithFields(logrus.Fields{
		"alert_rule_id": rule.ID,
		"alert_type":    rule.AlertType,
		"severity":      rule.Severity,
		"ip_address":    ipAddress,
		"user_id":       userID,
	}).Warn("Security alert triggered")

	return nil
}

// sendNotification sends notifications via the specified channel
func (m *SecurityAlertMiddleware) sendNotification(ctx context.Context, channel string, rule SecurityAlertRule, ipAddress, userID, email string, details map[string]interface{}) error {
	switch strings.ToLower(channel) {
	case "email":
		return m.sendEmailNotification(ctx, rule, ipAddress, userID, email, details)
	case "slack":
		return m.sendSlackNotification(ctx, rule, ipAddress, userID, email, details)
	case "pagerduty":
		return m.sendPagerDutyNotification(ctx, rule, ipAddress, userID, email, details)
	default:
		logrus.WithField("channel", channel).Warn("Unknown notification channel")
		return nil
	}
}

// sendEmailNotification sends an email notification
func (m *SecurityAlertMiddleware) sendEmailNotification(ctx context.Context, rule SecurityAlertRule, ipAddress, userID, email string, details map[string]interface{}) error {
	// TODO: Implement email notification using existing email service
	logrus.WithFields(logrus.Fields{
		"channel":    "email",
		"alert_type": rule.AlertType,
		"severity":   rule.Severity,
		"ip_address": ipAddress,
	}).Info("Would send email notification for security alert")
	return nil
}

// sendSlackNotification sends a Slack notification
func (m *SecurityAlertMiddleware) sendSlackNotification(ctx context.Context, rule SecurityAlertRule, ipAddress, userID, email string, details map[string]interface{}) error {
	// TODO: Implement Slack notification using webhook
	logrus.WithFields(logrus.Fields{
		"channel":    "slack",
		"alert_type": rule.AlertType,
		"severity":   rule.Severity,
		"ip_address": ipAddress,
	}).Info("Would send Slack notification for security alert")
	return nil
}

// sendPagerDutyNotification sends a PagerDuty notification
func (m *SecurityAlertMiddleware) sendPagerDutyNotification(ctx context.Context, rule SecurityAlertRule, ipAddress, userID, email string, details map[string]interface{}) error {
	// TODO: Implement PagerDuty notification
	logrus.WithFields(logrus.Fields{
		"channel":    "pagerduty",
		"alert_type": rule.AlertType,
		"severity":   rule.Severity,
		"ip_address": ipAddress,
	}).Info("Would send PagerDuty notification for security alert")
	return nil
}

// RecordFailedLogin records a failed login attempt for alerting
func (m *SecurityAlertMiddleware) RecordFailedLogin(ctx context.Context, ipAddress string) error {
	key := fmt.Sprintf("security:failed_login:%s", ipAddress)
	window := 15 * time.Minute // Default window

	// Increment counter
	count, err := m.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment failed login counter: %w", err)
	}

	// Set expiry on first increment
	if count == 1 {
		if err := m.redisClient.Expire(ctx, key, window).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to set expiry on failed login counter")
		}
	}

	// Record metric
	metrics.AdminLoginAttempts.WithLabelValues("failure").Inc()

	return nil
}

// RecordSuccessfulLogin records a successful login for metrics
func (m *SecurityAlertMiddleware) RecordSuccessfulLogin(ctx context.Context, ipAddress string) error {
	metrics.AdminLoginAttempts.WithLabelValues("success").Inc()
	return nil
}

// RecordRateLimitHit records a rate limit hit
func (m *SecurityAlertMiddleware) RecordRateLimitHit(ctx context.Context, ipAddress string) error {
	key := fmt.Sprintf("security:rate_limit:%s", ipAddress)
	window := 1 * time.Minute // Default window

	count, err := m.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	if count == 1 {
		if err := m.redisClient.Expire(ctx, key, window).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to set expiry on rate limit counter")
		}
	}

	metrics.AdminRateLimitHits.Inc()

	return nil
}

// RecordIPBlocked records an IP block event
func (m *SecurityAlertMiddleware) RecordIPBlocked(ctx context.Context, ipAddress string) error {
	key := fmt.Sprintf("security:blocked:%s", ipAddress)
	window := 1 * time.Hour // Default window

	if err := m.redisClient.Set(ctx, key, "1", window).Err(); err != nil {
		return fmt.Errorf("failed to set blocked IP: %w", err)
	}

	metrics.AdminIPBlocks.Inc()

	return nil
}

// RecordCSRFViolation records a CSRF violation
func (m *SecurityAlertMiddleware) RecordCSRFViolation(ctx context.Context, ipAddress string) error {
	metrics.AdminCSRFViolations.Inc()
	return nil
}

// SecurityAlertHandler implements the SecurityAlerter interface
type SecurityAlertHandler struct {
	// Add dependencies as needed (email client, slack client, etc.)
}

// NewSecurityAlertHandler creates a new security alert handler
func NewSecurityAlertHandler() *SecurityAlertHandler {
	return &SecurityAlertHandler{}
}

// TriggerAlert triggers an alert and sends notifications
func (h *SecurityAlertHandler) TriggerAlert(ctx context.Context, alertType, severity, ipAddress string, details map[string]interface{}) error {
	logrus.WithFields(logrus.Fields{
		"alert_type": alertType,
		"severity":   severity,
		"ip_address": ipAddress,
		"details":    details,
	}).Warn("Security alert triggered")

	// TODO: Implement actual notification sending
	return nil
}
