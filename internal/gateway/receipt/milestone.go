// Package receipt (under gateway) — milestone wrapper.
//
// This is a thin wrapper that the receipt emitter uses to trigger
// milestone checks. The actual milestone logic lives in
// internal/api/handlers/receipt/milestone.go; this wrapper provides
// the protocol-agnostic trigger point for the gateway.
package receipt

import (
	"context"

	"github.com/google/uuid"
)

// MilestoneChecker is the interface the emitter uses for milestone
// detection. The concrete implementation is the existing Milestone
// struct in handlers/receipt/milestone.go.
type MilestoneChecker interface {
	OnExecution(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string)
}

// Milestone wraps MilestoneChecker for the emitter.
type Milestone struct {
	checker MilestoneChecker
}

// NewMilestone creates a milestone wrapper.
func NewMilestone(checker MilestoneChecker) *Milestone {
	if checker == nil {
		return nil
	}
	return &Milestone{checker: checker}
}

// OnExecution triggers a milestone check. Delegates to the underlying
// checker (which is the existing handlers/receipt/milestone.go worker).
func (m *Milestone) OnExecution(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string) {
	if m == nil || m.checker == nil {
		return
	}
	m.checker.OnExecution(ctx, functionID, tenantID, publicID)
}
