package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ValidateAuditIntegrity checks the integrity of audit logs
func (acs *AuditComplianceService) ValidateAuditIntegrity(ctx context.Context) error {
	// Check for gaps in audit log sequence
	// Check for tampering attempts
	// Verify cryptographic signatures if implemented

	// This is a simplified implementation
	rows, err := acs.db.Query(`
		SELECT id, timestamp
		FROM audit_events
		WHERE timestamp > NOW() - INTERVAL '30 days'
		ORDER BY timestamp ASC
	`)

	if err != nil {
		return fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var prevTime time.Time
	gapCount := 0

	for rows.Next() {
		var id string
		var timestamp time.Time

		if err := rows.Scan(&id, &timestamp); err != nil {
			continue
		}

		if !prevTime.IsZero() {
			// Check for suspicious time gaps (more than 1 hour without events during business hours)
			if timestamp.Sub(prevTime) > time.Hour && acs.isBusinessHours(timestamp) {
				gapCount++
				acs.logger.WithFields(logrus.Fields{
					"gap_duration": timestamp.Sub(prevTime),
					"from_time":    prevTime,
					"to_time":      timestamp,
				}).Warn("Audit log gap detected")
			}
		}

		prevTime = timestamp
	}

	if gapCount > 0 {
		event := &ComplianceAuditEvent{
			Action: "audit_integrity_check",
			BeforeState: nil,
			AfterState: map[string]interface{}{
				"gaps_found": gapCount,
				"status":     "warning",
			},
			Success:   false,
			Timestamp: time.Now(),
		}

		if err := acs.LogComplianceEvent(ctx, event); err != nil {
			acs.logger.WithError(err).Warn("Failed to log audit integrity check")
		}
	}

	return nil
}

// isBusinessHours checks if timestamp is during business hours
func (acs *AuditComplianceService) isBusinessHours(t time.Time) bool {
	hour := t.Hour()
	weekday := t.Weekday()

	// Business hours: Monday-Friday, 9 AM - 6 PM
	return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour <= 18
}

// GetAuditTrailSummary generates a summary of audit activities
func (acs *AuditComplianceService) GetAuditTrailSummary(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Count total events
	var totalEvents int
	err := acs.db.QueryRow(`
		SELECT COUNT(*) FROM audit_events
		WHERE timestamp BETWEEN $1 AND $2
	`, startDate, endDate).Scan(&totalEvents)

	if err != nil {
		return nil, fmt.Errorf("failed to count audit events: %w", err)
	}

	summary["total_events"] = totalEvents

	// Count events by action type
	rows, err := acs.db.Query(`
		SELECT action, COUNT(*) as count
		FROM audit_events
		WHERE timestamp BETWEEN $1 AND $2
		GROUP BY action
		ORDER BY count DESC
	`, startDate, endDate)

	if err != nil {
		return nil, fmt.Errorf("failed to query action counts: %w", err)
	}
	defer rows.Close()

	actionCounts := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		actionCounts[action] = count
	}

	summary["action_counts"] = actionCounts

	// Count failed operations
	var failedEvents int
	acs.db.QueryRow(`
		SELECT COUNT(*) FROM audit_events
		WHERE timestamp BETWEEN $1 AND $2 AND success = false
	`, startDate, endDate).Scan(&failedEvents)

	summary["failed_events"] = failedEvents

	// Get most active users
	userRows, err := acs.db.Query(`
		SELECT COALESCE(actor_email, 'unknown') as user_email, COUNT(*) as count
		FROM audit_events
		WHERE timestamp BETWEEN $1 AND $2
		GROUP BY actor_email
		ORDER BY count DESC
		LIMIT 10
	`, startDate, endDate)

	if err != nil {
		return nil, fmt.Errorf("failed to query user activity: %w", err)
	}
	defer userRows.Close()

	userActivity := make(map[string]int)
	for userRows.Next() {
		var userEmail string
		var count int
		if err := userRows.Scan(&userEmail, &count); err != nil {
			continue
		}
		userActivity[userEmail] = count
	}

	summary["user_activity"] = userActivity

	return summary, nil
}