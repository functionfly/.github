package triggers

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

// DeploymentEvent represents a deployment event
type DeploymentEvent struct {
	ID        uuid.UUID
	AppID     uuid.UUID
	AppName   string
	UserID    uuid.UUID
	Status    string // "started", "success", "failed"
	Message   string
	ErrorMsg  string
	LogsURL   string
	Timestamp int64
}

// DeploymentTrigger handles deployment events and creates notifications
type DeploymentTrigger struct {
	name string
}

// NewDeploymentTrigger creates a new deployment trigger
func NewDeploymentTrigger() *DeploymentTrigger {
	return &DeploymentTrigger{
		name: "deployment",
	}
}

// Name returns the trigger name
func (t *DeploymentTrigger) Name() string {
	return t.name
}

// ShouldTrigger determines if this trigger should handle the event
func (t *DeploymentTrigger) ShouldTrigger(event interface{}) bool {
	_, ok := event.(*DeploymentEvent)
	return ok
}

// BuildNotification creates a notification from a deployment event
func (t *DeploymentTrigger) BuildNotification(event interface{}) (*notification.Notification, error) {
	e, ok := event.(*DeploymentEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	switch e.Status {
	case "success":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeDeploymentSuccess,
			Category: notification.CategoryDeployment,
			Title:    fmt.Sprintf("Deployment Successful: %s", e.AppName),
			Body:     fmt.Sprintf("Your deployment of %s was successful.", e.AppName),
			Data: notification.JSONMap{
				"app_id":      e.AppID.String(),
				"app_name":    e.AppName,
				"deployed_at": e.Timestamp,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "failed":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeDeploymentFailed,
			Category: notification.CategoryDeployment,
			Title:    fmt.Sprintf("Deployment Failed: %s", e.AppName),
			Body:     fmt.Sprintf("Your deployment of %s failed. %s", e.AppName, e.ErrorMsg),
			Data: notification.JSONMap{
				"app_id":        e.AppID.String(),
				"app_name":      e.AppName,
				"error_message": e.ErrorMsg,
				"logs_url":      e.LogsURL,
				"failed_at":     e.Timestamp,
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "started":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeDeploymentStarted,
			Category: notification.CategoryDeployment,
			Title:    fmt.Sprintf("Deployment Started: %s", e.AppName),
			Body:     fmt.Sprintf("Your deployment of %s has started.", e.AppName),
			Data: notification.JSONMap{
				"app_id":     e.AppID.String(),
				"app_name":   e.AppName,
				"started_at": e.Timestamp,
			},
			Priority: notification.PriorityLow,
			Channels: []string{notification.ChannelInApp},
		}, nil

	default:
		return nil, fmt.Errorf("unknown deployment status: %s", e.Status)
	}
}

// DeploymentTriggerRegistry manages deployment triggers
type DeploymentTriggerRegistry struct {
	triggers map[string]notification.Trigger
}

// NewDeploymentTriggerRegistry creates a new deployment trigger registry
func NewDeploymentTriggerRegistry() *DeploymentTriggerRegistry {
	return &DeploymentTriggerRegistry{
		triggers: make(map[string]notification.Trigger),
	}
}

// Register registers a deployment trigger
func (r *DeploymentTriggerRegistry) Register(trigger notification.Trigger) {
	r.triggers[trigger.Name()] = trigger
}

// Get retrieves a deployment trigger by name
func (r *DeploymentTriggerRegistry) Get(name string) (notification.Trigger, bool) {
	trigger, ok := r.triggers[name]
	return trigger, ok
}

// ProcessEvent processes a deployment event and returns a notification
func (r *DeploymentTriggerRegistry) ProcessEvent(event interface{}) (*notification.Notification, error) {
	for _, trigger := range r.triggers {
		if trigger.ShouldTrigger(event) {
			return trigger.BuildNotification(event)
		}
	}
	return nil, fmt.Errorf("no trigger found for event")
}
