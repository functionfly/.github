package dna

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/google/uuid"
)

// NotificationServiceNotifier implements MutationNotifier using the notification service.
type NotificationServiceNotifier struct {
	svc *notification.Service
}

// NewNotificationServiceNotifier creates a new notifier.
func NewNotificationServiceNotifier(svc *notification.Service) *NotificationServiceNotifier {
	return &NotificationServiceNotifier{svc: svc}
}

// NotifyMutationProposed sends a real-time notification when a DNA mutation is proposed.
func (n *NotificationServiceNotifier) NotifyMutationProposed(ctx context.Context, tenantID, functionID, mutationType, triggerReason string) error {
	// Parse tenant ID as user ID (tenant owner receives the notification)
	userID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	mutationLabels := map[string]string{
		"optimize_latency":  "Latency Optimization",
		"reduce_memory":     "Memory Reduction",
		"fix_error_pattern": "Error Pattern Fix",
		"improve_reliability": "Reliability Improvement",
		"refactor_hotpath":  "Hot Path Refactor",
	}

	label := mutationLabels[mutationType]
	if label == "" {
		label = mutationType
	}

	_, err = n.svc.Send(ctx, notification.SendRequest{
		UserID:   userID,
		Type:     "function.dna_mutation_proposed",
		Category: notification.CategoryFunction,
		Title:    fmt.Sprintf("New DNA Evolution Proposed: %s", label),
		Body:     fmt.Sprintf("Function %s has a new evolution proposal. %s", functionID, triggerReason),
		Data: notification.JSONMap{
			"function_id":   functionID,
			"mutation_type": mutationType,
			"trigger_reason": triggerReason,
		},
		Channels: []string{notification.ChannelInApp},
		Priority: notification.PriorityNormal,
	})
	return err
}
