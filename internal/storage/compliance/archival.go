package compliance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ArchiveOldAuditLogs archives audit logs older than specified age
func (acs *AuditComplianceService) ArchiveOldAuditLogs(ctx context.Context, archiveAge time.Duration) error {
	cutoffDate := time.Now().Add(-archiveAge)
	archiveID := uuid.New()

	acs.logger.WithFields(logrus.Fields{
		"archive_age_days": int(archiveAge.Hours() / 24),
		"archive_id":       archiveID.String(),
	}).Info("Starting audit log archival process")

	// Check if archive storage is configured
	if acs.archiveStore == nil {
		acs.logger.Warn("Archive storage not configured, skipping archival")
		return fmt.Errorf("archive storage not configured")
	}

	// Get count of events to archive
	var oldEventCount int
	err := acs.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM audit_events
		WHERE timestamp < $1
	`, cutoffDate).Scan(&oldEventCount)
	if err != nil {
		return fmt.Errorf("failed to count old audit events: %w", err)
	}

	if oldEventCount == 0 {
		acs.logger.Info("No old audit events to archive")
		return nil
	}

	// Export old audit events to JSON format
	rows, err := acs.db.QueryContext(ctx, `
		SELECT id, action, resource_type, resource_id, before_state, after_state,
		       success, user_id, ip_address, user_agent, timestamp
		FROM audit_events
		WHERE timestamp < $1
		ORDER BY timestamp ASC
	`, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to query old audit events: %w", err)
	}
	defer rows.Close()

	// Convert events to JSON for archiving
	var events []map[string]interface{}
	for rows.Next() {
		var event struct {
			ID          uuid.UUID
			Action      string
			ResourceType *string
			ResourceID   *uuid.UUID
			BeforeState  *string
			AfterState   *string
			Success      bool
			UserID       *uuid.UUID
			IPAddress    *string
			UserAgent    *string
			Timestamp    time.Time
		}

		err := rows.Scan(
			&event.ID, &event.Action, &event.ResourceType, &event.ResourceID,
			&event.BeforeState, &event.AfterState, &event.Success,
			&event.UserID, &event.IPAddress, &event.UserAgent, &event.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("failed to scan audit event: %w", err)
		}

		eventMap := map[string]interface{}{
			"id":            event.ID,
			"action":        event.Action,
			"resource_type": event.ResourceType,
			"resource_id":   event.ResourceID,
			"before_state":  event.BeforeState,
			"after_state":   event.AfterState,
			"success":       event.Success,
			"user_id":       event.UserID,
			"ip_address":    event.IPAddress,
			"user_agent":    event.UserAgent,
			"timestamp":     event.Timestamp,
		}
		events = append(events, eventMap)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating audit events: %w", err)
	}

	// Convert events to JSON bytes
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal audit events to JSON: %w", err)
	}

	// Create archive metadata
	archiveKey := fmt.Sprintf("audit-logs/%s/%s.json", cutoffDate.Format("2006-01"), archiveID.String())
	metadata := &storage.ArchiveMetadata{
		ID:          archiveID,
		ArchiveType: "audit_logs",
		RecordCount: len(events),
		DateRange: storage.ArchiveDateRange{
			Start: cutoffDate.Add(-archiveAge), // Approximate start date
			End:   cutoffDate,
		},
		StorageKey: archiveKey,
		Status:     "pending",
		Metadata: map[string]interface{}{
			"table_name":       "audit_events",
			"archive_criteria": "timestamp < cutoff_date",
			"cutoff_date":      cutoffDate,
		},
	}

	// Store archive using the configured storage backend
	err = acs.archiveStore.StoreArchive(ctx, archiveKey, bytes.NewReader(eventsJSON), metadata)
	if err != nil {
		metadata.Status = "failed"
		metadata.ErrorMessage = err.Error()

		// Log the failure
		acs.logger.WithError(err).WithFields(logrus.Fields{
			"archive_key": archiveKey,
			"event_count": len(events),
		}).Error("Failed to store audit log archive")

		return fmt.Errorf("failed to store archive: %w", err)
	}

	// Delete archived events from main table
	result, err := acs.db.ExecContext(ctx, `
		DELETE FROM audit_events
		WHERE timestamp < $1
	`, cutoffDate)
	if err != nil {
		// Log error but don't fail the operation - data is safely archived
		acs.logger.WithError(err).WithField("archive_key", archiveKey).
			Error("Failed to delete archived audit events from main table")
	} else {
		deletedCount, _ := result.RowsAffected()
		acs.logger.WithFields(logrus.Fields{
			"archive_key":    archiveKey,
			"events_deleted": deletedCount,
			"events_archived": len(events),
		}).Info("Successfully deleted archived audit events from main table")
	}

	// Log successful archival
	acs.logger.WithFields(logrus.Fields{
		"archive_key":      archiveKey,
		"events_archived":  len(events),
		"original_size_kb": len(eventsJSON) / 1024,
		"archive_age_days": int(archiveAge.Hours() / 24),
	}).Info("Audit log archival completed successfully")

	// Log compliance event
	complianceEvent := &ComplianceAuditEvent{
		Framework:    GDPR, // Audit log retention is often GDPR-related
		Section:      "Data Retention",
		RequirementID: "audit-log-archival",
		Action:       "audit_log_archival",
		Severity:     "medium",
		BeforeState: map[string]interface{}{
			"active_events_count": oldEventCount,
		},
		AfterState: map[string]interface{}{
			"events_archived":     len(events),
			"archive_key":         archiveKey,
			"archive_id":          archiveID.String(),
			"archive_age_days":    int(archiveAge.Hours() / 24),
			"cutoff_date":         cutoffDate,
			"compression_ratio":   metadata.CompressionRatio,
			"compressed_size_kb":  metadata.CompressedSize / 1024,
		},
		Success:   true,
		Timestamp: time.Now(),
	}

	return acs.LogComplianceEvent(ctx, complianceEvent)
}