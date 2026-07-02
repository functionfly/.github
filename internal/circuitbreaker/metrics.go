package circuitbreaker

import (
	"github.com/functionfly/functionfly/internal/monitoring"
)

// MetricsOnStateChange returns an OnStateChange callback that updates Prometheus metrics.
// Use this when creating a Config for breakers that should report to the existing
// functionfly_circuit_breaker_state and functionfly_circuit_breaker_transitions_total metrics.
//
// The key format should be "backend:<uuid>" or "provider:<name>" — the provider and region
// are extracted from the breaker's Manager context if available, or default to "unknown".
func MetricsOnStateChange(key string, from, to State) {
	backendID, provider, region := parseKey(key)
	monitoring.UpdateCircuitBreakerState(backendID, provider, region, int(to))
	monitoring.RecordCircuitBreakerTransition(backendID, from.String(), to.String())
}

// parseKey extracts backend_id, provider, and region from a breaker key.
// Key formats:
//   - "backend:<uuid>" → backend_id=<uuid>, provider="unknown", region="unknown"
//   - "provider:<name>" → backend_id=<name>, provider=<name>, region="unknown"
//   - "<provider>:<uuid>" → backend_id=<uuid>, provider=<provider>, region="unknown"
func parseKey(key string) (backendID, provider, region string) {
	provider = "unknown"
	region = "unknown"
	backendID = key

	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			provider = key[:i]
			backendID = key[i+1:]
			break
		}
	}

	return backendID, provider, region
}
