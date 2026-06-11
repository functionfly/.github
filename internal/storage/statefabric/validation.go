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
	{Pending, Online}:   true,
	{Pending, Suspended}: true,
	{Online, Degraded}:  true,
	{Online, Offline}:   true,
	{Online, Suspended}: true,
	{Degraded, Online}:  true,
	{Degraded, Offline}: true,
	{Degraded, Suspended}: true,
	{Offline, Online}: true,
	{Suspended, Online}: true,
}

type StoreStatusTransition struct {
	From StoreStatus
	To   StoreStatus
}

var allowedStoreTransitions = map[StoreStatusTransition]bool{
	{Active, Paused}:  true,
	{Active, Stopped}: true,
	{Paused, Active}:  true,
	{Paused, Stopped}: true,
	{Stopped, Active}: true,
	{Stopped, Paused}: true,
}

type PipelineStatusTransition struct {
	From PipelineStatus
	To   PipelineStatus
}

var allowedPipelineTransitions = map[PipelineStatusTransition]bool{
	{Draft, Active}:   true,
	{Draft, Stopped}:  true,
	{Active, Paused}:  true,
	{Active, Stopped}: true,
	{Paused, Active}:  true,
	{Paused, Stopped}: true,
	{Stopped, Draft}:  true,
	{Stopped, Active}: true,
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
	case Pending, Online, Degraded, Offline, Suspended:
		return true
	default:
		return false
	}
}

func (v *StateTransitionValidator) IsValidStoreStatus(status StoreStatus) bool {
	switch status {
	case Active, Paused, Stopped:
		return true
	default:
		return false
	}
}

func (v *StateTransitionValidator) IsValidPipelineStatus(status PipelineStatus) bool {
	switch status {
	case Draft, Active, Paused, Stopped:
		return true
	default:
		return false
	}
}

func fabricStatusFromString(s string) FabricStatus {
	switch s {
	case "pending":
		return Pending
	case "online":
		return Online
	case "degraded":
		return Degraded
	case "offline":
		return Offline
	case "suspended":
		return Suspended
	default:
		return ""
	}
}

func storeStatusFromString(s string) StoreStatus {
	switch s {
	case "active":
		return Active
	case "paused":
		return Paused
	case "stopped":
		return Stopped
	default:
		return ""
	}
}

func pipelineStatusFromString(s string) PipelineStatus {
	switch s {
	case "draft":
		return Draft
	case "active":
		return Active
	case "paused":
		return Paused
	case "stopped":
		return Stopped
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