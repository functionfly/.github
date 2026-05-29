//go:build !cgo

package wasm

import (
	"github.com/google/uuid"
)

// StateFabricHostConfig configures edge StateFabric host functions for a single execution.
type StateFabricHostConfig struct {
	Repo          interface{}
	TriggerEngine interface{}
	TenantID      uuid.UUID
	DefaultFabric uuid.UUID
}

// HostHandlerForExecution returns the default handler when CGO/WASM host functions are unavailable.
func HostHandlerForExecution(_ StateFabricHostConfig) HostFunctionHandler {
	return NewDefaultHostHandler(nil)
}
