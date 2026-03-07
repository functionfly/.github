package triggers

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

type FollowEvent struct {
	ID             uuid.UUID
	FollowerID     uuid.UUID
	FollowerUserID uuid.UUID
	FunctionID     uuid.UUID
	FunctionName   string
	OwnerID        uuid.UUID
	OwnerName      string
	EventType      string // "followed", "unfollowed", "function_updated", "function_new_version", "function_deprecated"
	Timestamp      int64
}

type FollowTrigger struct {
	name string
}

func NewFollowTrigger() *FollowTrigger {
	return &FollowTrigger{
		name: "follow",
	}
}

func (t *FollowTrigger) Name() string {
	return t.name
}

func (t *FollowTrigger) ShouldTrigger(event interface{}) bool {
	_, ok := event.(*FollowEvent)
	return ok
}

func (t *FollowTrigger) BuildNotification(event interface{}) (*notification.Notification, error) {
	e, ok := event.(*FollowEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	switch e.EventType {
	case "function_updated":
		return &notification.Notification{
			UserID:   e.FollowerUserID,
			Type:     "follow.function_updated",
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Updated: %s", e.FunctionName),
			Body:     fmt.Sprintf("The function %s by %s has been updated.", e.FunctionName, e.OwnerName),
			Data: notification.JSONMap{
				"function_id":   e.FunctionID.String(),
				"function_name": e.FunctionName,
				"owner_id":      e.OwnerID.String(),
				"owner_name":    e.OwnerName,
				"updated_at":    e.Timestamp,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "function_new_version":
		return &notification.Notification{
			UserID:   e.FollowerUserID,
			Type:     "follow.function_new_version",
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("New Version: %s", e.FunctionName),
			Body:     fmt.Sprintf("A new version of %s by %s is now available.", e.FunctionName, e.OwnerName),
			Data: notification.JSONMap{
				"function_id":     e.FunctionID.String(),
				"function_name":   e.FunctionName,
				"owner_id":        e.OwnerID.String(),
				"owner_name":      e.OwnerName,
				"new_version_at":  e.Timestamp,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "function_deprecated":
		return &notification.Notification{
			UserID:   e.FollowerUserID,
			Type:     "follow.function_deprecated",
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Deprecated: %s", e.FunctionName),
			Body:     fmt.Sprintf("The function %s by %s has been deprecated. Consider updating your integrations.", e.FunctionName, e.OwnerName),
			Data: notification.JSONMap{
				"function_id":     e.FunctionID.String(),
				"function_name":   e.FunctionName,
				"owner_id":        e.OwnerID.String(),
				"owner_name":      e.OwnerName,
				"deprecated_at":   e.Timestamp,
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "followed":
		return &notification.Notification{
			UserID:   e.FollowerUserID,
			Type:     "follow.started",
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Now Following: %s", e.FunctionName),
			Body:     fmt.Sprintf("You are now following %s by %s.", e.FunctionName, e.OwnerName),
			Data: notification.JSONMap{
				"function_id":   e.FunctionID.String(),
				"function_name": e.FunctionName,
				"owner_id":      e.OwnerID.String(),
				"owner_name":    e.OwnerName,
				"followed_at":  e.Timestamp,
			},
			Priority: notification.PriorityLow,
			Channels: []string{notification.ChannelInApp},
		}, nil

	default:
		return nil, fmt.Errorf("unknown follow event type: %s", e.EventType)
	}
}

type FollowTriggerRegistry struct {
	triggers map[string]notification.Trigger
}

func NewFollowTriggerRegistry() *FollowTriggerRegistry {
	return &FollowTriggerRegistry{
		triggers: make(map[string]notification.Trigger),
	}
}

func (r *FollowTriggerRegistry) Register(trigger notification.Trigger) {
	r.triggers[trigger.Name()] = trigger
}

func (r *FollowTriggerRegistry) Get(name string) (notification.Trigger, bool) {
	trigger, ok := r.triggers[name]
	return trigger, ok
}

func (r *FollowTriggerRegistry) ProcessEvent(event interface{}) (*notification.Notification, error) {
	for _, trigger := range r.triggers {
		if trigger.ShouldTrigger(event) {
			return trigger.BuildNotification(event)
		}
	}
	return nil, fmt.Errorf("no trigger found for event")
}
