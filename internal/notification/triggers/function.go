package triggers

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

type FunctionEvent struct {
	ID          uuid.UUID
	FunctionID  uuid.UUID
	FunctionName string
	TenantID    uuid.UUID
	UserID      uuid.UUID
	EventType   string // "executed", "error", "published", "updated", "deprecated"
	Message     string
	ErrorMsg    string
	ExecutionTimeMs int64
	LogsURL     string
	Timestamp   int64
}

type FunctionTrigger struct {
	name string
}

func NewFunctionTrigger() *FunctionTrigger {
	return &FunctionTrigger{
		name: "function",
	}
}

func (t *FunctionTrigger) Name() string {
	return t.name
}

func (t *FunctionTrigger) ShouldTrigger(event interface{}) bool {
	_, ok := event.(*FunctionEvent)
	return ok
}

func (t *FunctionTrigger) BuildNotification(event interface{}) (*notification.Notification, error) {
	e, ok := event.(*FunctionEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event type")
	}

	switch e.EventType {
	case "executed":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     "function.executed",
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Executed: %s", e.FunctionName),
			Body:     fmt.Sprintf("Your function %s executed successfully in %dms.", e.FunctionName, e.ExecutionTimeMs),
			Data: notification.JSONMap{
				"function_id":      e.FunctionID.String(),
				"function_name":   e.FunctionName,
				"execution_time":   e.ExecutionTimeMs,
				"logs_url":        e.LogsURL,
				"executed_at":     e.Timestamp,
			},
			Priority: notification.PriorityLow,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "error":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     "function.error",
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Error: %s", e.FunctionName),
			Body:     fmt.Sprintf("Your function %s encountered an error: %s", e.FunctionName, e.ErrorMsg),
			Data: notification.JSONMap{
				"function_id":    e.FunctionID.String(),
				"function_name":  e.FunctionName,
				"error_message":  e.ErrorMsg,
				"logs_url":       e.LogsURL,
				"failed_at":      e.Timestamp,
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	case "published":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeFunctionPublished,
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Published: %s", e.FunctionName),
			Body:     fmt.Sprintf("Your function %s has been published to the registry.", e.FunctionName),
			Data: notification.JSONMap{
				"function_id":    e.FunctionID.String(),
				"function_name":  e.FunctionName,
				"published_at":   e.Timestamp,
			},
			Priority: notification.PriorityNormal,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "updated":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeFunctionUpdated,
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Updated: %s", e.FunctionName),
			Body:     fmt.Sprintf("Your function %s has been updated.", e.FunctionName),
			Data: notification.JSONMap{
				"function_id":  e.FunctionID.String(),
				"function_name": e.FunctionName,
				"updated_at":   e.Timestamp,
			},
			Priority: notification.PriorityLow,
			Channels: []string{notification.ChannelInApp},
		}, nil

	case "deprecated":
		return &notification.Notification{
			UserID:   e.UserID,
			Type:     notification.TypeFunctionDeprecated,
			Category: notification.CategoryFunction,
			Title:    fmt.Sprintf("Function Deprecated: %s", e.FunctionName),
			Body:     fmt.Sprintf("Your function %s has been deprecated. Please update to a newer version.", e.FunctionName),
			Data: notification.JSONMap{
				"function_id":    e.FunctionID.String(),
				"function_name":  e.FunctionName,
				"deprecated_at":  e.Timestamp,
			},
			Priority: notification.PriorityHigh,
			Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		}, nil

	default:
		return nil, fmt.Errorf("unknown function event type: %s", e.EventType)
	}
}

type FunctionTriggerRegistry struct {
	triggers map[string]notification.Trigger
}

func NewFunctionTriggerRegistry() *FunctionTriggerRegistry {
	return &FunctionTriggerRegistry{
		triggers: make(map[string]notification.Trigger),
	}
}

func (r *FunctionTriggerRegistry) Register(trigger notification.Trigger) {
	r.triggers[trigger.Name()] = trigger
}

func (r *FunctionTriggerRegistry) Get(name string) (notification.Trigger, bool) {
	trigger, ok := r.triggers[name]
	return trigger, ok
}

func (r *FunctionTriggerRegistry) ProcessEvent(event interface{}) (*notification.Notification, error) {
	for _, trigger := range r.triggers {
		if trigger.ShouldTrigger(event) {
			return trigger.BuildNotification(event)
		}
	}
	return nil, fmt.Errorf("no trigger found for event")
}
