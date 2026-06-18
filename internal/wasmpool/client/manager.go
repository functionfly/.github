package client

import (
	"context"
	"fmt"
	"sync"

	wasmpool "github.com/functionfly/functionfly/internal/wasm"
	"github.com/sirupsen/logrus"
)

// Manager is the package-level entry point the orchestrator uses to
// construct and access the WasmPoolRouter. It is safe for concurrent use.
type Manager struct {
	mu     sync.RWMutex
	router *WasmPoolRouter
	ext    *ExternalPoolClient
	loc    *LocalPoolClient
}

// NewManagerFromConfig builds the router and its sub-clients from the
// env-driven RouterConfig and the orchestrator's existing local pool.
//
// Behavior:
//   - WASM_POOL_EXTERNAL_PERCENT == 0 (default) → external is not constructed
//     and every request goes to Local. The router is fully functional but
//     never reaches the breaker or override paths.
//   - WASM_POOL_EXTERNAL_PERCENT > 0 → the external client is constructed
//     (gRPC dial to the headless service). If the dial fails, NewManager
//     returns an error — the plan says the SDK must "pre-populate the ring
//     from a config map or DNS lookup at startup, before the first Execute
//     call".
func NewManagerFromConfig(localPool *wasmpool.InstancePool) (*Manager, error) {
	cfg := LoadRouterConfigFromConfig()
	return NewManager(localPool, cfg)
}

// LoadRouterConfigFromConfig reads RouterConfig from env. Exposed for tests
// that want to construct a config without touching the global env.
func LoadRouterConfigFromConfig() RouterConfig {
	return LoadRouterConfigFromEnv()
}

// NewManager is the explicit constructor that takes a RouterConfig. Use
// this from tests; use NewManagerFromConfig in production wiring.
func NewManager(localPool *wasmpool.InstancePool, cfg RouterConfig) (*Manager, error) {
	loc := NewLocalPoolClient(localPool)
	mgr := &Manager{loc: loc}

	if cfg.ExternalPercent > 0 || len(cfg.ExternalTenants) > 0 {
		extCfg := ExternalConfig{
			Addr:      EnvAddr(),
			AuthToken: envOrDefault("WASM_POOL_GRPC_AUTH_TOKEN", ""),
			TLS:       envOrBool("WASM_POOL_GRPC_TLS", false),
			CertFile:  envOrDefault("WASM_POOL_GRPC_CERT_FILE", ""),
			KeyFile:   envOrDefault("WASM_POOL_GRPC_KEY_FILE", ""),
			CAFile:    envOrDefault("WASM_POOL_GRPC_CA_FILE", ""),
		}
		ext, err := NewExternalPoolClient(extCfg)
		if err != nil {
			return nil, fmt.Errorf("wasm pool: build external client: %w", err)
		}
		mgr.ext = ext
		logrus.WithField("endpoints", ext.Endpoints()).Info("wasm pool: external client ready")
	}

	mgr.router = NewRouter(loc, mgr.ext, cfg)
	return mgr, nil
}

// Execute is the convenience method the orchestrator's call sites use.
func (m *Manager) Execute(ctx context.Context, req *Request) (*Response, error) {
	m.mu.RLock()
	r := m.router
	m.mu.RUnlock()
	return r.Execute(ctx, req)
}

// Router returns the underlying router (for tests and metrics introspection).
func (m *Manager) Router() *WasmPoolRouter { return m.router }

// Close releases the external gRPC connections (the local pool is owned
// by the caller and not closed here).
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ext != nil {
		return m.ext.Close()
	}
	return nil
}

