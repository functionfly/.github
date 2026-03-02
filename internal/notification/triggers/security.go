package triggers

import (
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Type       string // "password_changed", "mfa_enabled", "new_device_login", "suspicious_activity"
	DeviceInfo string
	Location   string
	IPAddress  string
	Timestamp  time.Time
	Metadata   map[string]interface{}
}

// SecurityTrigger handles security events and creates notifications
type SecurityTrigger struct {
	name string
}

// NewSecurityTrigger creates a new security trigger
func NewSecurityTrigger() *SecurityTrigger {
	return &SecurityTrigger{
		name: "security",
	}
}

// Name returns the trigger name
func (t *SecurityTrigger) Name() string {
	return t.name
}

// ShouldTrigger determines if this trigger should handle the event
func (t *SecurityTrigger) ShouldTrigger(event interface{}) bool {
	_, ok := event.(*SecurityEvent)
	return ok
}

// BuildNotification creates a notification from a security event
func (t *SecurityTrigger) BuildNotification(event interface{}) (*notification.Notification, error) {
	e, ok := event.(*SecurityEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	switch e.Type {
	case "password_changed":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeSecurityPasswordChanged,
			Category: notification.CategorySecurity,
			Title:    "Password Changed",
			Body:     fmt.Sprintf("Your password was changed on %s at %s.", e.Timestamp.Format("2006-01-02"), e.Timestamp.Format("15:04")),
			Data: notification.JSONMap{
				"changed_at":  e.Timestamp.Format(time.RFC3339),
				"device_info": e.DeviceInfo,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "mfa_enabled":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeSecurityMFAEnabled,
			Category: notification.CategorySecurity,
			Title:    "Multi-Factor Authentication Enabled",
			Body:     "Multi-factor authentication has been enabled on your account.",
			Data: notification.JSONMap{
				"enabled_at":  e.Timestamp.Format(time.RFC3339),
				"device_info": e.DeviceInfo,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "new_device_login":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeSecurityNewDeviceLogin,
			Category: notification.CategorySecurity,
			Title:    "New Device Login Detected",
			Body:     fmt.Sprintf("A new device logged into your account from %s using %s.", e.Location, e.DeviceInfo),
			Data: notification.JSONMap{
				"login_at":    e.Timestamp.Format(time.RFC3339),
				"device_info": e.DeviceInfo,
				"location":    e.Location,
				"ip_address":  e.IPAddress,
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "suspicious_activity":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeSecuritySuspiciousActivity,
			Category: notification.CategorySecurity,
			Title:    "Suspicious Activity Detected",
			Body:     fmt.Sprintf("We detected suspicious activity on your account from %s. Please review your recent activity and change your password if you don't recognize this activity.", e.Location),
			Data: notification.JSONMap{
				"detected_at": e.Timestamp.Format(time.RFC3339),
				"location":    e.Location,
				"ip_address":  e.IPAddress,
				"device_info": e.DeviceInfo,
			},
			Priority: notification.PriorityUrgent,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	default:
		return nil, fmt.Errorf("unknown security event type: %s", e.Type)
	}
}
