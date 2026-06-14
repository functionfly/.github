package statefabric

import (
	"fmt"
)

type FabricStatus string

const (
	FabricStatusPending  FabricStatus = "pending"
	FabricStatusOnline  FabricStatus = "online"
	FabricStatusDegraded FabricStatus = "degraded"
	FabricStatusOffline FabricStatus = "offline"
	FabricStatusSuspended FabricStatus = "suspended"
)

type StoreStatus string

const (
	StoreStatusActive  StoreStatus = "active"
	StoreStatusPaused StoreStatus = "paused"
	StoreStatusStopped StoreStatus = "stopped"
)

type PipelineStatus string

const (
	PipelineStatusDraft     PipelineStatus = "draft"
	PipelineStatusActive   PipelineStatus = "active"
	PipelineStatusPaused   PipelineStatus = "paused"
	PipelineStatusStopped  PipelineStatus = "stopped"
)

type FabricStatusTransition struct {
	From FabricStatus
	To   FabricStatus
}

var allowedFabricTransitions = map[FabricStatusTransition]bool{
	{FabricStatusPending, FabricStatusOnline}:   true,
	{FabricStatusPending, FabricStatusSuspended}: true,
	{FabricStatusOnline, FabricStatusDegraded}:  true,
	{FabricStatusOnline, FabricStatusOffline}:   true,
	{FabricStatusOnline, FabricStatusSuspended}: true,
	{FabricStatusDegraded, FabricStatusOnline}:  true,
	{FabricStatusDegraded, FabricStatusOffline}: true,
	{FabricStatusDegraded, FabricStatusSuspended}: true,
	{FabricStatusOffline, FabricStatusOnline}: true,
	{FabricStatusSuspended, FabricStatusOnline}: true,
}

type StoreStatusTransition struct {
	From StoreStatus
	To   StoreStatus
}

var allowedStoreTransitions = map[StoreStatusTransition]bool{
	{StoreStatusActive, StoreStatusPaused}:  true,
	{StoreStatusActive, StoreStatusStopped}: true,
	{StoreStatusPaused, StoreStatusActive}:  true,
	{StoreStatusPaused, StoreStatusStopped}: true,
	{StoreStatusStopped, StoreStatusActive}: true,
	{StoreStatusStopped, StoreStatusPaused}: true,
}

type PipelineStatusTransition struct {
	From PipelineStatus
	To   PipelineStatus
}

var allowedPipelineTransitions = map[PipelineStatusTransition]bool{
	{PipelineStatusDraft, PipelineStatusActive}:   true,
	{PipelineStatusDraft, PipelineStatusStopped}:  true,
	{PipelineStatusActive, PipelineStatusPaused}:  true,
	{PipelineStatusActive, PipelineStatusStopped}: true,
	{PipelineStatusPaused, PipelineStatusActive}:  true,
	{PipelineStatusPaused, PipelineStatusStopped}: true,
	{PipelineStatusStopped, PipelineStatusDraft}:  true,
	{PipelineStatusStopped, PipelineStatusActive}: true,
}

type StateTransitionValidator struct{}

func NewStateTransitionValidator() *StateTransitionValidator {
	return &StateTransitionValidator{}
}

func (v *StateTransitionValidator) ValidateFabricTransition(from, to FabricStatus) error {
	transition := FabricStatusTransition{From: from, To: to}
	if !allowedFabricTransitions[transition] {
		return fmt.Errorf("invalid fabric status transition from %s to %s", from, to)
	}
	return nil
}

func (v *StateTransitionValidator) ValidateStoreTransition(from, to StoreStatus) error {
	transition := StoreStatusTransition{From: from, To: to}
	if !allowedStoreTransitions[transition] {
		return fmt.Errorf("invalid store status transition from %s to %s", from, to)
	}
	return nil
}

func (v *StateTransitionValidator) ValidatePipelineTransition(from, to PipelineStatus) error {
	transition := PipelineStatusTransition{From: from, To: to}
	if !allowedPipelineTransitions[transition] {
		return fmt.Errorf("invalid pipeline status transition from %s to %s", from, to)
	}
	return nil
}

func (v *StateTransitionValidator) IsValidFabricStatus(status FabricStatus) bool {
	switch status {
	case FabricStatusPending, FabricStatusOnline, FabricStatusDegraded, FabricStatusOffline, FabricStatusSuspended:
		return true
	default:
		return false
	}
}

func (v *StateTransitionValidator) IsValidStoreStatus(status StoreStatus) bool {
	switch status {
	case StoreStatusActive, StoreStatusPaused, StoreStatusStopped:
		return true
	default:
		return false
	}
}

func (v *StateTransitionValidator) IsValidPipelineStatus(status PipelineStatus) bool {
	switch status {
	case PipelineStatusDraft, PipelineStatusActive, PipelineStatusPaused, PipelineStatusStopped:
		return true
	default:
		return false
	}
}

func fabricStatusFromString(s string) FabricStatus {
	switch s {
	case "pending":
		return FabricStatusPending
	case "online":
		return FabricStatusOnline
	case "degraded":
		return FabricStatusDegraded
	case "offline":
		return FabricStatusOffline
	case "suspended":
		return FabricStatusSuspended
	default:
		return ""
	}
}

func storeStatusFromString(s string) StoreStatus {
	switch s {
	case "active":
		return StoreStatusActive
	case "paused":
		return StoreStatusPaused
	case "stopped":
		return StoreStatusStopped
	default:
		return ""
	}
}

func pipelineStatusFromString(s string) PipelineStatus {
	switch s {
	case "draft":
		return PipelineStatusDraft
	case "active":
		return PipelineStatusActive
	case "paused":
		return PipelineStatusPaused
	case "stopped":
		return PipelineStatusStopped
	default:
		return ""
	}
}

func (r *Repository) ValidateFabricStatusTransition(currentStatus string, newStatus string) error {
	validator := NewStateTransitionValidator()
	current := fabricStatusFromString(currentStatus)
	target := fabricStatusFromString(newStatus)

	if current == "" || target == "" {
		return fmt.Errorf("invalid status value")
	}

	return validator.ValidateFabricTransition(current, target)
}

func (r *Repository) ValidateStoreStatusTransition(currentStatus string, newStatus string) error {
	validator := NewStateTransitionValidator()
	current := storeStatusFromString(currentStatus)
	target := storeStatusFromString(newStatus)

	if current == "" || target == "" {
		return fmt.Errorf("invalid status value")
	}

	return validator.ValidateStoreTransition(current, target)
}

func (r *Repository) ValidatePipelineStatusTransition(currentStatus string, newStatus string) error {
	validator := NewStateTransitionValidator()
	current := pipelineStatusFromString(currentStatus)
	target := pipelineStatusFromString(newStatus)

	if current == "" || target == "" {
		return fmt.Errorf("invalid status value")
	}

	return validator.ValidatePipelineTransition(current, target)
}