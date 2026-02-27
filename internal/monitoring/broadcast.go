package monitoring

import (
	"encoding/json"
	"fmt"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// broadcastAlert broadcasts an alert using Supabase's LISTEN/NOTIFY
func (s *Service) broadcastAlert(alert *storage.Alert) error {
	// Create payload for real-time broadcast
	payload := map[string]interface{}{
		"type":      "alert_created",
		"alert_id":  alert.ID,
		"alert_type": alert.AlertType,
		"severity":  alert.Severity,
		"title":     alert.Title,
		"message":   alert.Message,
		"tenant_id": alert.TenantID,
		"timestamp": alert.CreatedAt,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert payload: %w", err)
	}

	// Broadcast using Supabase's real-time features via repository
	channelName := "monitoring_alerts"
	if alert.TenantID != nil {
		channelName = fmt.Sprintf("tenant_%s_alerts", alert.TenantID.String())
	}

	return s.notifyChannel(channelName, string(payloadJSON))
}

// broadcastMonitoringEvent broadcasts a monitoring event using Supabase real-time
func (s *Service) broadcastMonitoringEvent(event *storage.MonitoringEvent) error {
	payload := map[string]interface{}{
		"type":       "monitoring_event",
		"event_type": event.EventType,
		"event_id":   event.ID,
		"tenant_id":  event.TenantID,
		"app_id":     event.AppID,
		"backend_id": event.BackendID,
		"request_id": event.RequestID,
		"data":       event.Data,
		"timestamp":  event.Timestamp,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	channelName := "monitoring_events"
	if event.TenantID != nil {
		channelName = fmt.Sprintf("tenant_%s_events", event.TenantID.String())
	}

	return s.notifyChannel(channelName, string(payloadJSON))
}

// notifyChannel sends a notification to a PostgreSQL channel
func (s *Service) notifyChannel(channel, payload string) error {
	return s.db.PgNotify(channel, payload)
}

// broadcastToChannels sends an alert to all registered subscriber channels
func (s *Service) broadcastToChannels(alert *storage.Alert) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for subscriberID, ch := range s.alertChans {
		select {
		case ch <- alert:
			// Successfully sent
		default:
			// Channel is full, skip this subscriber
			logrus.WithField("subscriber_id", subscriberID).Warn("Alert channel is full, skipping subscriber")
		}
	}
}