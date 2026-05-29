//go:build cgo

package wasm

import (
	"context"

	"github.com/google/uuid"

	statestore "github.com/functionfly/functionfly/internal/storage/state"
	statefabricrepo "github.com/functionfly/functionfly/internal/storage/statefabric"
)

// StateFabricHostConfig configures edge StateFabric host functions for a single execution.
type StateFabricHostConfig struct {
	Repo          *statefabricrepo.Repository
	TriggerEngine *statestore.TriggerEngine
	TenantID      uuid.UUID
	DefaultFabric uuid.UUID
}

// NewStateFabricHostHandler creates a host handler backed by durable state_values storage.
func NewStateFabricHostHandler(
	baseHandler *DefaultHostHandler,
	cfg StateFabricHostConfig,
) *StateFabricHostHandler {
	if baseHandler == nil {
		baseHandler = NewDefaultHostHandler(nil)
	}
	return &StateFabricHostHandler{
		DefaultHostHandler: baseHandler,
		repo:               cfg.Repo,
		triggerEngine:      cfg.TriggerEngine,
		tenantID:           cfg.TenantID,
		fabricID:           cfg.DefaultFabric,
		ctx:                context.Background(),
	}
}

// HostHandlerForExecution builds a StateFabric-aware handler when a repo is configured.
func HostHandlerForExecution(cfg StateFabricHostConfig) HostFunctionHandler {
	if cfg.Repo == nil || cfg.TenantID == uuid.Nil {
		return NewDefaultHostHandler(nil)
	}
	return NewStateFabricHostHandler(nil, cfg)
}
