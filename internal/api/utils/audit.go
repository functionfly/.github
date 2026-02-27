package utils

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LogAuditEvent logs an admin action to the audit system
func LogAuditEvent(ctx context.Context, repo storage.Repository, r *http.Request, action, resourceType string, resourceID *uuid.UUID, beforeState, afterState interface{}, success bool) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		logrus.Warn("No authenticated user for audit logging")
		return // Don't log if no authenticated user
	}

	// Convert states to simple maps to avoid JSON marshaling issues
	var beforeJSON, afterJSON interface{}

	// Helper function to convert struct to map
	toMap := func(obj interface{}) map[string]interface{} {
		if obj == nil {
			return nil
		}
		data, err := json.Marshal(obj)
		if err != nil {
			logrus.WithError(err).Warn("Failed to marshal audit state")
			return nil
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal audit state")
			return nil
		}
		return m
	}

	beforeJSON = toMap(beforeState)
	afterJSON = toMap(afterState)

	// Extract IP address from RemoteAddr (remove port if present)
	ipAddress := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ipAddress = host
	}

	event := &storage.AuditEvent{
		ActorUserID:  &user.UserID,
		ActorEmail:   user.Email,
		TenantID:     &user.TenantID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		RequestID:    r.Header.Get("X-Request-ID"),
		BeforeState:  beforeJSON,
		AfterState:   afterJSON,
		IPAddress:    ipAddress,
		UserAgent:    r.Header.Get("User-Agent"),
		Timestamp:    time.Now(),
		Success:      success,
	}

	logrus.WithFields(logrus.Fields{
		"action":        action,
		"resource_type": resourceType,
		"actor_email":   user.Email,
		"success":       success,
	}).Info("Logging audit event")

	// Log audit event synchronously for now
	if err := repo.LogAuditEvent(ctx, event); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"action":        action,
			"resource_type": resourceType,
			"actor_email":   user.Email,
		}).Error("Failed to log audit event")
	} else {
		logrus.WithFields(logrus.Fields{
			"action":        action,
			"resource_type": resourceType,
			"actor_email":   user.Email,
		}).Info("Audit event logged successfully")
	}
}