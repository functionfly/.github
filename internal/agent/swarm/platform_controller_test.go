package swarm

import (
	"testing"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/stretchr/testify/assert"
)

func TestPlatformStatus_TotalPendingMessages(t *testing.T) {
	// No DB needed - pure unit test
	status := &PlatformStatus{
		Children: []ChildAgentStatus{
			{PendingMessages: 3},
			{PendingMessages: 5},
			{PendingMessages: 0},
		},
	}

	total := status.TotalPendingMessages()
	assert.Equal(t, 8, total)
}

func TestPlatformController_extractRoleFromCapabilities(t *testing.T) {
	// No DB needed - pure unit test
	pc := &PlatformController{}

	tests := []struct {
		name         string
		capabilities identity.JSONBMap
		expected     string
	}{
		{
			name:         "has role capability",
			capabilities: identity.JSONBMap{"role": "github-scanner", "source": "github"},
			expected:     "github-scanner",
		},
		{
			name:         "no role capability",
			capabilities: identity.JSONBMap{"source": "github"},
			expected:     "unknown",
		},
		{
			name:         "nil capabilities",
			capabilities: nil,
			expected:     "unknown",
		},
		{
			name:         "empty capabilities",
			capabilities: identity.JSONBMap{},
			expected:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := pc.extractRoleFromCapabilities(tt.capabilities)
			assert.Equal(t, tt.expected, role)
		})
	}
}
