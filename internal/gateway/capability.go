package gateway

import (
	"context"
	"sync"
)

// CapabilitySet is the data-driven capability resolution layer.
// Capabilities live in DB tables (api_key_capabilities,
// agent_capabilities, role_capabilities) and are loaded at request
// time. No Go switch statements — adding a new capability is a row,
// not a deploy.
//
// Capability namespaces:
//   - mcp:tools:list, mcp:tools:call
//   - a2a:tasks:send, a2a:tasks:get, a2a:tasks:cancel, a2a:tasks:subscribe
//   - a2a:agent:invoke, a2a:agent:delegate
//   - receipt:read, receipt:write, receipt:revoke
//   - bundle:publish, bundle:install
type CapabilitySet struct {
	mu           sync.RWMutex
	capabilities map[string]map[string]struct{} // keyID → set of capabilities
}

// NewCapabilitySet creates an empty CapabilitySet.
func NewCapabilitySet() *CapabilitySet {
	return &CapabilitySet{
		capabilities: make(map[string]map[string]struct{}),
	}
}

// CapabilityChecker is the interface the GatewayCore uses to check
// whether a caller has a given capability. Implementations load from
// DB tables at request time.
type CapabilityChecker interface {
	// HasCapability returns true if the caller (identified by keyID)
	// has the specified capability.
	HasCapability(ctx context.Context, keyID, capability string) (bool, error)
	// LoadCapabilities returns all capabilities for the caller.
	LoadCapabilities(ctx context.Context, keyID string) ([]string, error)
}

// InMemoryCapabilityChecker is a simple in-memory implementation for
// development and testing. Production uses a DB-backed implementation.
type InMemoryCapabilityChecker struct {
	mu           sync.RWMutex
	capabilities map[string]map[string]struct{}
}

// NewInMemoryCapabilityChecker creates an empty in-memory checker.
func NewInMemoryCapabilityChecker() *InMemoryCapabilityChecker {
	return &InMemoryCapabilityChecker{
		capabilities: make(map[string]map[string]struct{}),
	}
}

// Grant adds a capability for a keyID.
func (c *InMemoryCapabilityChecker) Grant(keyID, capability string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.capabilities[keyID] == nil {
		c.capabilities[keyID] = make(map[string]struct{})
	}
	c.capabilities[keyID][capability] = struct{}{}
}

// Revoke removes a capability for a keyID.
func (c *InMemoryCapabilityChecker) Revoke(keyID, capability string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.capabilities[keyID], capability)
}

// HasCapability checks if the keyID has the specified capability.
func (c *InMemoryCapabilityChecker) HasCapability(_ context.Context, keyID, capability string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if caps, ok := c.capabilities[keyID]; ok {
		_, has := caps[capability]
		return has, nil
	}
	return false, nil
}

// LoadCapabilities returns all capabilities for the keyID.
func (c *InMemoryCapabilityChecker) LoadCapabilities(_ context.Context, keyID string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []string
	if caps, ok := c.capabilities[keyID]; ok {
		for cap := range caps {
			out = append(out, cap)
		}
	}
	return out, nil
}

// DefaultCapabilityChecker returns a checker that grants all
// capabilities to all callers. Used as a fallback when no capability
// system is configured (development mode).
type AllowAllCapabilityChecker struct{}

func (c *AllowAllCapabilityChecker) HasCapability(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

func (c *AllowAllCapabilityChecker) LoadCapabilities(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
