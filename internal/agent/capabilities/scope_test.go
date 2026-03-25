package capabilities

import (
	"testing"
)

func TestCapabilities_CanExecute(t *testing.T) {
	tests := []struct {
		name        string
		caps        Capabilities
		functionURI string
		expected    bool
	}{
		{
			name:        "no restrictions",
			caps:        Capabilities{AgentID: "test"},
			functionURI: "fx://functionfly/test",
			expected:    true,
		},
		{
			name: "allowed function",
			caps: Capabilities{
				AgentID:          "test",
				AllowedFunctions: []string{"fx://functionfly/*"},
			},
			functionURI: "fx://functionfly/test",
			expected:    true,
		},
		{
			name: "denied function",
			caps: Capabilities{
				AgentID:         "test",
				DeniedFunctions: []string{"fx://functionfly/test"},
			},
			functionURI: "fx://functionfly/test",
			expected:    false,
		},
		{
			name: "denied takes precedence over allowed",
			caps: Capabilities{
				AgentID:          "test",
				AllowedFunctions: []string{"fx://functionfly/*"},
				DeniedFunctions:  []string{"fx://functionfly/test"},
			},
			functionURI: "fx://functionfly/test",
			expected:    false,
		},
		{
			name: "glob pattern match",
			caps: Capabilities{
				AgentID:          "test",
				AllowedFunctions: []string{"fx://functionfly/array-*"},
			},
			functionURI: "fx://functionfly/array-sum",
			expected:    true,
		},
		{
			name: "glob pattern no match",
			caps: Capabilities{
				AgentID:          "test",
				AllowedFunctions: []string{"fx://functionfly/array-*"},
			},
			functionURI: "fx://functionfly/http-get",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.CanExecute(tt.functionURI)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCapabilities_CanUseProvider(t *testing.T) {
	tests := []struct {
		name     string
		caps     Capabilities
		provider string
		expected bool
	}{
		{
			name:     "no restrictions",
			caps:     Capabilities{AgentID: "test"},
			provider: "workers",
			expected: true,
		},
		{
			name: "allowed provider",
			caps: Capabilities{
				AgentID:          "test",
				AllowedProviders: []string{"workers", "vercel"},
			},
			provider: "workers",
			expected: true,
		},
		{
			name: "denied provider",
			caps: Capabilities{
				AgentID:          "test",
				AllowedProviders: []string{"workers"},
			},
			provider: "vercel",
			expected: false,
		},
		{
			name: "case insensitive",
			caps: Capabilities{
				AgentID:          "test",
				AllowedProviders: []string{"Workers"},
			},
			provider: "workers",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.CanUseProvider(tt.provider)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCapabilities_CanUseRegion(t *testing.T) {
	tests := []struct {
		name     string
		caps     Capabilities
		region   string
		expected bool
	}{
		{
			name:     "no restrictions",
			caps:     Capabilities{AgentID: "test"},
			region:   "iad",
			expected: true,
		},
		{
			name: "allowed region",
			caps: Capabilities{
				AgentID:        "test",
				AllowedRegions: []string{"iad", "sfo"},
			},
			region:   "iad",
			expected: true,
		},
		{
			name: "denied region",
			caps: Capabilities{
				AgentID:        "test",
				AllowedRegions: []string{"iad"},
			},
			region:   "sfo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.CanUseRegion(tt.region)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCapabilities_ValidateCallDepth(t *testing.T) {
	tests := []struct {
		name     string
		caps     Capabilities
		depth    int
		expected bool
	}{
		{
			name:     "no limit",
			caps:     Capabilities{AgentID: "test"},
			depth:    100,
			expected: true,
		},
		{
			name: "within limit",
			caps: Capabilities{
				AgentID:      "test",
				MaxCallDepth: 10,
			},
			depth:    5,
			expected: true,
		},
		{
			name: "at limit",
			caps: Capabilities{
				AgentID:      "test",
				MaxCallDepth: 10,
			},
			depth:    10,
			expected: true,
		},
		{
			name: "exceeds limit",
			caps: Capabilities{
				AgentID:      "test",
				MaxCallDepth: 10,
			},
			depth:    11,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.ValidateCallDepth(tt.depth)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCapabilities_ValidateExecutionTime(t *testing.T) {
	tests := []struct {
		name     string
		caps     Capabilities
		seconds  int
		expected bool
	}{
		{
			name:     "no limit",
			caps:     Capabilities{AgentID: "test"},
			seconds:  1000,
			expected: true,
		},
		{
			name: "within limit",
			caps: Capabilities{
				AgentID:          "test",
				MaxExecutionTime: 300,
			},
			seconds:  100,
			expected: true,
		},
		{
			name: "exceeds limit",
			caps: Capabilities{
				AgentID:          "test",
				MaxExecutionTime: 300,
			},
			seconds:  301,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.ValidateExecutionTime(tt.seconds)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCapabilities_ValidateMemory(t *testing.T) {
	tests := []struct {
		name     string
		caps     Capabilities
		mb       int
		expected bool
	}{
		{
			name:     "no limit",
			caps:     Capabilities{AgentID: "test"},
			mb:       1000,
			expected: true,
		},
		{
			name: "within limit",
			caps: Capabilities{
				AgentID:     "test",
				MaxMemoryMB: 512,
			},
			mb:       256,
			expected: true,
		},
		{
			name: "exceeds limit",
			caps: Capabilities{
				AgentID:     "test",
				MaxMemoryMB: 512,
			},
			mb:       513,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.ValidateMemory(tt.mb)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCapabilities_Merge(t *testing.T) {
	caps1 := Capabilities{
		AgentID:          "test",
		AllowedFunctions: []string{"fx://functionfly/*"},
		AllowedProviders: []string{"workers", "vercel"},
		MaxCallDepth:     10,
		MaxConcurrent:    5,
	}

	caps2 := Capabilities{
		AgentID:          "test",
		AllowedFunctions: []string{"fx://functionfly/array-*"},
		AllowedProviders: []string{"workers"},
		MaxCallDepth:     5,
		MaxConcurrent:    10,
	}

	merged := caps1.Merge(&caps2)

	// Functions should be intersection
	if len(merged.AllowedFunctions) != 1 {
		t.Errorf("expected 1 allowed function, got %d", len(merged.AllowedFunctions))
	}

	// Providers should be intersection
	if len(merged.AllowedProviders) != 1 {
		t.Errorf("expected 1 allowed provider, got %d", len(merged.AllowedProviders))
	}

	// Numeric limits should be minimum
	if merged.MaxCallDepth != 5 {
		t.Errorf("expected MaxCallDepth=5, got %d", merged.MaxCallDepth)
	}
	if merged.MaxConcurrent != 5 {
		t.Errorf("expected MaxConcurrent=5, got %d", merged.MaxConcurrent)
	}
}

func TestDefaultCapabilities(t *testing.T) {
	caps := DefaultCapabilities("test-agent")

	if caps.AgentID != "test-agent" {
		t.Errorf("expected AgentID=test-agent, got %s", caps.AgentID)
	}
	if caps.MaxCallDepth != 10 {
		t.Errorf("expected MaxCallDepth=10, got %d", caps.MaxCallDepth)
	}
	if caps.MaxConcurrent != 10 {
		t.Errorf("expected MaxConcurrent=10, got %d", caps.MaxConcurrent)
	}
	if caps.MaxExecutionTime != 300 {
		t.Errorf("expected MaxExecutionTime=300, got %d", caps.MaxExecutionTime)
	}
	if caps.MaxMemoryMB != 512 {
		t.Errorf("expected MaxMemoryMB=512, got %d", caps.MaxMemoryMB)
	}
}
