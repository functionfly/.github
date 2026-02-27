package cache

import (
	"github.com/google/uuid"
)

// FunctionVersionData contains the data needed to check cache eligibility
type FunctionVersionData struct {
	FunctionID   uuid.UUID
	Version      string
	Deterministic bool
	CacheTTL     int
	SideEffects  string
}

// EligibilityResult contains cache eligibility decision
type EligibilityResult struct {
	Eligible   bool   // Whether caching is enabled for this function
	TTL        int    // Cache TTL in seconds
	Version    string // Cache namespace (function version)
	FunctionID string // Function ID for cache key
	CanUseCDN  bool   // Whether CDN caching is allowed (public functions)
}

// CheckEligibility determines if a function version is cache-eligible
// based on registry metadata
func CheckEligibility(v FunctionVersionData) EligibilityResult {
	// Must be deterministic
	if !v.Deterministic {
		return EligibilityResult{Eligible: false}
	}

	// Must have explicit cache TTL > 0
	if v.CacheTTL <= 0 {
		return EligibilityResult{Eligible: false}
	}

	// Must have no side effects
	if v.SideEffects != "" && v.SideEffects != "none" {
		return EligibilityResult{Eligible: false}
	}

	return EligibilityResult{
		Eligible:   true,
		TTL:        v.CacheTTL,
		Version:    v.Version,
		FunctionID: v.FunctionID.String(),
		CanUseCDN:  true, // Set based on function visibility in handler
	}
}

// IsCacheable is a convenience function that returns true if the result is eligible
func (e EligibilityResult) IsCacheable() bool {
	return e.Eligible
}
