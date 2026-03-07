package state

import (
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
)

// Handler handles state API requests
type Handler struct {
	stateRepo    *staterepo.StateRepository
	triggerEngine *staterepo.TriggerEngine
}

// NewHandler creates a new state handler
func NewHandler(stateRepo *staterepo.StateRepository) *Handler {
	return &Handler{
		stateRepo: stateRepo,
	}
}

// NewHandlerWithTriggerEngine creates a new state handler with trigger engine
func NewHandlerWithTriggerEngine(stateRepo *staterepo.StateRepository, triggerEngine *staterepo.TriggerEngine) *Handler {
	return &Handler{
		stateRepo:    stateRepo,
		triggerEngine: triggerEngine,
	}
}
