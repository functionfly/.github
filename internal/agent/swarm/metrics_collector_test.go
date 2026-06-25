package swarm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateConversionRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		discovered int
		published  int
		expected   float64
	}{
		{"100% conversion", 100, 100, 100.0},
		{"50% conversion", 100, 50, 50.0},
		{"0% conversion", 100, 0, 0.0},
		{"zero discovered", 0, 0, 0.0},
		{"partial", 3, 1, 33.333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := calculateConversionRate(tt.discovered, tt.published)
			assert.InDelta(t, tt.expected, rate, 0.1)
		})
	}
}

func TestCalculateConversionRate_EdgeCases(t *testing.T) {
	t.Parallel()

	// All edge cases
	assert.Equal(t, 0.0, calculateConversionRate(0, 0))
	assert.Equal(t, 0.0, calculateConversionRate(1, 0))
	assert.Equal(t, 100.0, calculateConversionRate(1, 1))
	assert.InDelta(t, 33.33, calculateConversionRate(3, 1), 0.01)
}
