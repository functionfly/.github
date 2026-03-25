package capabilities

import (
	"path"
	"strings"
)

// Capabilities defines what an agent is allowed to do
type Capabilities struct {
	AgentID          string   `json:"agent_id"`
	AllowedFunctions []string `json:"allowed_functions,omitempty"` // glob patterns: "fx://functionfly/*"
	DeniedFunctions  []string `json:"denied_functions,omitempty"`
	AllowedProviders []string `json:"allowed_providers,omitempty"` // ["workers", "vercel"]
	MaxCallDepth     int      `json:"max_call_depth,omitempty"`
	MaxConcurrent    int      `json:"max_concurrent,omitempty"`
	AllowedRegions   []string `json:"allowed_regions,omitempty"`
	MaxExecutionTime int      `json:"max_execution_time,omitempty"` // seconds
	MaxMemoryMB      int      `json:"max_memory_mb,omitempty"`
}

// DefaultCapabilities returns default capabilities for an agent
func DefaultCapabilities(agentID string) Capabilities {
	return Capabilities{
		AgentID:          agentID,
		MaxCallDepth:     10,
		MaxConcurrent:    10,
		MaxExecutionTime: 300, // 5 minutes
		MaxMemoryMB:      512,
	}
}

// CanExecute checks if the agent can execute the given function
func (c *Capabilities) CanExecute(functionURI string) bool {
	// Check denied first
	for _, pattern := range c.DeniedFunctions {
		if matchPattern(pattern, functionURI) {
			return false
		}
	}

	// Check allowed
	if len(c.AllowedFunctions) == 0 {
		return true // no restrictions
	}

	for _, pattern := range c.AllowedFunctions {
		if matchPattern(pattern, functionURI) {
			return true
		}
	}

	return false
}

// CanUseProvider checks if the agent can use the given provider
func (c *Capabilities) CanUseProvider(provider string) bool {
	if len(c.AllowedProviders) == 0 {
		return true // no restrictions
	}

	for _, p := range c.AllowedProviders {
		if strings.EqualFold(p, provider) {
			return true
		}
	}

	return false
}

// CanUseRegion checks if the agent can use the given region
func (c *Capabilities) CanUseRegion(region string) bool {
	if len(c.AllowedRegions) == 0 {
		return true // no restrictions
	}

	for _, r := range c.AllowedRegions {
		if strings.EqualFold(r, region) {
			return true
		}
	}

	return false
}

// ValidateCallDepth checks if the call depth is within limits
func (c *Capabilities) ValidateCallDepth(depth int) bool {
	if c.MaxCallDepth == 0 {
		return true // no limit
	}
	return depth <= c.MaxCallDepth
}

// ValidateExecutionTime checks if the execution time is within limits
func (c *Capabilities) ValidateExecutionTime(seconds int) bool {
	if c.MaxExecutionTime == 0 {
		return true // no limit
	}
	return seconds <= c.MaxExecutionTime
}

// ValidateMemory checks if the memory usage is within limits
func (c *Capabilities) ValidateMemory(mb int) bool {
	if c.MaxMemoryMB == 0 {
		return true // no limit
	}
	return mb <= c.MaxMemoryMB
}

// matchPattern matches a glob pattern against a value
func matchPattern(pattern, value string) bool {
	// Support glob patterns
	if strings.Contains(pattern, "*") {
		matched, _ := path.Match(pattern, value)
		return matched
	}
	return pattern == value
}

// Merge merges two capability sets (more restrictive wins)
func (c *Capabilities) Merge(other *Capabilities) *Capabilities {
	merged := &Capabilities{
		AgentID: c.AgentID,
	}

	// Functions: intersection of allowed, union of denied
	if len(c.AllowedFunctions) > 0 && len(other.AllowedFunctions) > 0 {
		merged.AllowedFunctions = intersect(c.AllowedFunctions, other.AllowedFunctions)
	} else if len(c.AllowedFunctions) > 0 {
		merged.AllowedFunctions = c.AllowedFunctions
	} else {
		merged.AllowedFunctions = other.AllowedFunctions
	}

	merged.DeniedFunctions = union(c.DeniedFunctions, other.DeniedFunctions)

	// Providers: intersection
	if len(c.AllowedProviders) > 0 && len(other.AllowedProviders) > 0 {
		merged.AllowedProviders = intersect(c.AllowedProviders, other.AllowedProviders)
	} else if len(c.AllowedProviders) > 0 {
		merged.AllowedProviders = c.AllowedProviders
	} else {
		merged.AllowedProviders = other.AllowedProviders
	}

	// Regions: intersection
	if len(c.AllowedRegions) > 0 && len(other.AllowedRegions) > 0 {
		merged.AllowedRegions = intersect(c.AllowedRegions, other.AllowedRegions)
	} else if len(c.AllowedRegions) > 0 {
		merged.AllowedRegions = c.AllowedRegions
	} else {
		merged.AllowedRegions = other.AllowedRegions
	}

	// Numeric limits: minimum (more restrictive)
	merged.MaxCallDepth = min(c.MaxCallDepth, other.MaxCallDepth)
	merged.MaxConcurrent = min(c.MaxConcurrent, other.MaxConcurrent)
	merged.MaxExecutionTime = min(c.MaxExecutionTime, other.MaxExecutionTime)
	merged.MaxMemoryMB = min(c.MaxMemoryMB, other.MaxMemoryMB)

	return merged
}

func intersect(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range a {
		set[strings.ToLower(s)] = true
	}

	var result []string
	for _, s := range b {
		if set[strings.ToLower(s)] {
			result = append(result, s)
		}
	}
	return result
}

func union(a, b []string) []string {
	set := make(map[string]bool)
	for _, s := range a {
		set[strings.ToLower(s)] = true
	}
	for _, s := range b {
		set[strings.ToLower(s)] = true
	}

	var result []string
	for s := range set {
		result = append(result, s)
	}
	return result
}

func min(a, b int) int {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}
