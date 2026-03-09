package generation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoreComplexityCapsAtTen(t *testing.T) {
	t.Parallel()

	req := &GenerationRequest{
		Name:        "oauth webhook batch transform",
		Description: "cache retry security json csv stream auth webhook batch transform",
		Prompt:      "oauth auth stream webhook batch transform csv json retry cache security",
		Category:    "security",
		InputSchema: map[string]any{"a": 1, "b": 2, "c": 3, "d": 4},
		OutputSchema: map[string]any{"a": 1, "b": 2, "c": 3},
	}

	assert.Equal(t, 10, ScoreComplexity(req), "complexity should not exceed the documented upper bound")
}

func TestHeuristicModelSelectorRequiresReviewForComplexRequests(t *testing.T) {
	t.Parallel()

	selection := HeuristicModelSelector{}.SelectModel(&GenerationRequest{
		Name:        "oauth webhook batch transform",
		Description: "security sensitive batch workflow",
		Prompt:      "oauth auth stream webhook batch transform retry cache security",
		InputSchema: map[string]any{"a": 1, "b": 2, "c": 3, "d": 4},
		OutputSchema: map[string]any{"a": 1, "b": 2, "c": 3},
	})

	assert.Equal(t, defaultComplexModel, selection.Model, "complex requests should use the higher capability model")
	assert.True(t, selection.Review.Required, "complex requests should trigger manual review")
	assert.Equal(t, "human_approval", selection.Review.Tier, "manual review tier should be explicit for complex requests")
}

func TestInMemoryGenerationCacheHonorsExpiration(t *testing.T) {
	t.Parallel()

	cache := NewInMemoryGenerationCache(5 * time.Millisecond)
	req := &GenerationRequest{Name: "demo", Prompt: "hello", Runtime: "python3.11"}
	cache.Put(context.Background(), req, CachedGeneration{Code: "print('ok')", Model: "demo-model"})

	entry, ok := cache.Get(context.Background(), req)
	require.True(t, ok, "fresh cache entries should be retrievable")
	assert.Equal(t, "demo-model", entry.Model)

	time.Sleep(15 * time.Millisecond)
	entry, ok = cache.Get(context.Background(), req)
	assert.False(t, ok, "expired cache entries should not be returned")
	assert.Nil(t, entry, "expired cache lookups should return a nil entry")
}
