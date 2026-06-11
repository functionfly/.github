package gateway

import (
	"context"
	"sync"
)

// FallbackChain defines the ordered list of routes/adapters that
// GatewayCore tries when the primary execution path fails. Each
// element is a label like "primary", "cache", "peer-agent", "degraded".
//
// The chain decision is logged in CallResult.FallbackChain AND in the
// metric gateway_fallback_fired_total{from,to,reason}.
type FallbackChain struct {
	mu       sync.RWMutex
	strategies []FallbackStrategy
}

// FallbackStrategy is a single step in the fallback chain.
type FallbackStrategy struct {
	// Name is the human-readable label (e.g. "primary", "cache", "peer").
	Name string
	// Execute is the function to call. If it returns (result, nil), the
	// chain stops. If it returns (nil, err), the next strategy is tried.
	Execute func(ctx context.Context, req CallRequest) (*CallResult, error)
	// Healthy returns true if this strategy is available. Unhealthy
	// strategies are skipped.
	Healthy func(ctx context.Context) bool
}

// NewFallbackChain creates a FallbackChain from the given strategies.
func NewFallbackChain(strategies ...FallbackStrategy) *FallbackChain {
	return &FallbackChain{
		strategies: strategies,
	}
}

// Execute walks the fallback chain and returns the first successful
// result. The CallResult.FallbackChain field records which strategies
// were attempted.
func (fc *FallbackChain) Execute(ctx context.Context, req CallRequest) (*CallResult, []string, error) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	chain := make([]string, 0, len(fc.strategies))
	for _, s := range fc.strategies {
		if s.Healthy != nil && !s.Healthy(ctx) {
			chain = append(chain, s.Name+":skip")
			continue
		}

		result, err := s.Execute(ctx, req)
		if err == nil && result != nil {
			chain = append(chain, s.Name+":ok")
			result.FallbackChain = chain
			return result, chain, nil
		}
		chain = append(chain, s.Name+":fail")
	}

	return nil, chain, ErrAllFallbacksFailed
}

// ErrAllFallbacksFailed is returned when every strategy in the chain failed.
var ErrAllFallbacksFailed = &FallbackError{Message: "all fallback strategies failed"}

// FallbackError is the error type for fallback chain exhaustion.
type FallbackError struct {
	Message string
}

func (e *FallbackError) Error() string {
	return e.Message
}

// DefaultFallbackChain returns a simple single-strategy chain that
// delegates to the provided execute function. This is the default for
// environments without cache or peer fallback.
func DefaultFallbackChain(execute func(ctx context.Context, req CallRequest) (*CallResult, error)) *FallbackChain {
	return NewFallbackChain(FallbackStrategy{
		Name: "primary",
		Execute: execute,
		Healthy: func(_ context.Context) bool { return true },
	})
}
