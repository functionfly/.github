package triggers

import (
	"fmt"

	"github.com/functionfly/functionfly/internal/notification"
)

// TriggerRegistry manages all notification triggers
type TriggerRegistry struct {
	triggers map[string]notification.Trigger
}

// NewTriggerRegistry creates a new trigger registry with all triggers registered
func NewTriggerRegistry() *TriggerRegistry {
	registry := &TriggerRegistry{
		triggers: make(map[string]notification.Trigger),
	}

	// Register all triggers
	registry.Register(NewDeploymentTrigger())
	registry.Register(NewBillingTrigger())
	registry.Register(NewSecurityTrigger())
	registry.Register(NewTeamTrigger())
	registry.Register(NewFunctionTrigger())
	registry.Register(NewFollowTrigger())

	return registry
}

// Register registers a trigger
func (r *TriggerRegistry) Register(trigger notification.Trigger) {
	r.triggers[trigger.Name()] = trigger
}

// Get retrieves a trigger by name
func (r *TriggerRegistry) Get(name string) (notification.Trigger, bool) {
	trigger, ok := r.triggers[name]
	return trigger, ok
}

// ProcessEvent processes an event through all registered triggers and returns a notification
func (r *TriggerRegistry) ProcessEvent(event interface{}) (*notification.Notification, error) {
	for _, trigger := range r.triggers {
		if trigger.ShouldTrigger(event) {
			return trigger.BuildNotification(event)
		}
	}
	return nil, fmt.Errorf("no trigger found for event type")
}

// ListTriggers returns a list of all registered trigger names
func (r *TriggerRegistry) ListTriggers() []string {
	names := make([]string, 0, len(r.triggers))
	for name := range r.triggers {
		names = append(names, name)
	}
	return names
}
