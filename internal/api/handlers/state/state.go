package state

import (
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
)

// Handler handles state API requests
type Handler struct {
	stateRepo *staterepo.StateRepository
}

// NewHandler creates a new state handler
func NewHandler(stateRepo *staterepo.StateRepository) *Handler {
	return &Handler{
		stateRepo: stateRepo,
	}
}
