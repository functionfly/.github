package client

import (
	"context"

	wasmpool "github.com/functionfly/functionfly/internal/wasm"
	"github.com/sirupsen/logrus"
)

// LocalPoolClient wraps the existing in-process wasm.InstancePool so the
// router can use it as the byte-identical fallback. It mirrors the orchestrator's
// current direct pool.Get/Put usage in engines.go and wasm_integration.go.
type LocalPoolClient struct {
	pool *wasmpool.InstancePool
}

// NewLocalPoolClient constructs a client backed by the given pool.
func NewLocalPoolClient(pool *wasmpool.InstancePool) *LocalPoolClient {
	return &LocalPoolClient{pool: pool}
}

// Name returns the transport identifier used in metrics.
func (c *LocalPoolClient) Name() string { return "local" }

// Close is a no-op; the orchestrator owns the pool's lifecycle.
func (c *LocalPoolClient) Close() error { return nil }

// Execute acquires a per-tenant instance, runs the input, and returns the
// output. The instance is always returned to the pool, even on error.
//
// Mirrors the orchestrator's previous direct pool.Get/Put usage:
//   - Inits the pool instance (the pool factory may not have inited it).
//   - Loads source code via LoadCode() when req.Code is non-empty
//     (source-based functions: pool instance is the interpreter).
//   - Calls ExecuteWithContext with the per-request input.
//   - Returns the instance to the pool.
func (c *LocalPoolClient) Execute(ctx context.Context, req *Request) (*Response, error) {
	pi, err := c.pool.Get(ctx, req.TenantID, req.Runtime)
	if err != nil {
		return nil, err
	}
	cold := pi.ExecuteCount == 0
	pi.ExecuteCount++

	if initErr := pi.Instance.Init(); initErr != nil {
		_ = c.pool.Put(pi)
		return nil, initErr
	}
	if len(req.Code) > 0 {
		if loadErr := pi.Instance.LoadCode(string(req.Code)); loadErr != nil {
			_ = c.pool.Put(pi)
			return nil, loadErr
		}
	}

	execCtx := ctx
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	output, err := pi.Instance.ExecuteWithContext(execCtx, req.Input)
	mem := pi.Instance.GetMemoryUsage()

	if putErr := c.pool.Put(pi); putErr != nil {
		logrus.WithError(putErr).Warn("local pool: put failed")
	}

	resp := &Response{
		Output:      output,
		Latency:     0, // caller measures
		MemoryBytes: uint64(mem),
		ColdStarted: cold,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}
