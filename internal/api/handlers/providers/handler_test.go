package providers

import (
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIsProviderStale_NeverUsed(t *testing.T) {
	oldProvider := &storage.Provider{
		CreatedAt: time.Now().Add(-31 * 24 * time.Hour),
	}
	assert.True(t, isProviderStale(oldProvider))

	recentProvider := &storage.Provider{
		CreatedAt: time.Now().Add(-29 * 24 * time.Hour),
	}
	assert.False(t, isProviderStale(recentProvider))
}

func TestIsProviderStale_WithLastUsed(t *testing.T) {
	provider := &storage.Provider{
		LastUsedAt: func() *time.Time { t := time.Now().Add(-31 * 24 * time.Hour); return &t }(),
	}
	assert.True(t, isProviderStale(provider))

	provider = &storage.Provider{
		LastUsedAt: func() *time.Time { t := time.Now().Add(-29 * 24 * time.Hour); return &t }(),
	}
	assert.False(t, isProviderStale(provider))
}

func TestIsProviderStale_Exactly30Days(t *testing.T) {
	thirtyDays := 30 * 24 * time.Hour
	now := time.Now()

	provider := &storage.Provider{
		CreatedAt: now.Add(-thirtyDays - time.Minute),
	}
	assert.True(t, isProviderStale(provider))

	provider = &storage.Provider{
		CreatedAt: now.Add(-thirtyDays + time.Minute),
	}
	assert.False(t, isProviderStale(provider))
}

func TestListProviderFromStorage_StatusMapping(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		expectedStatus string
	}{
		{"active provider", "active", "online"},
		{"inactive provider", "inactive", "offline"},
		{"error provider", "error", "degraded"},
		{"pending provider", "pending", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			provider := &storage.Provider{
				ID:         uuid.New().String(),
				Provider:   "functionfly-edge",
				Status:     tt.status,
				CreatedAt:  now,
				LastUsedAt: &now,
			}

			result := listProviderFromStorage(provider)

			assert.Equal(t, provider.ID, result["id"])
			assert.Equal(t, "functionfly-edge", result["name"])
			assert.Equal(t, tt.expectedStatus, result["status"])
			assert.Equal(t, now.Format(time.RFC3339), result["connectedAt"])
			assert.NotNil(t, result["lastUsedAt"])
			assert.NotNil(t, result["isStale"])
		})
	}
}

func TestListProviderFromStorage_NoLastUsed(t *testing.T) {
	now := time.Now()
	provider := &storage.Provider{
		ID:         uuid.New().String(),
		Provider:   "functionfly-edge",
		Status:     "active",
		CreatedAt:  now,
		LastUsedAt: nil,
	}

	result := listProviderFromStorage(provider)

	assert.Equal(t, provider.ID, result["id"])
	assert.Equal(t, "functionfly-edge", result["name"])
	assert.Equal(t, "online", result["status"])
	assert.Equal(t, now.Format(time.RFC3339), result["connectedAt"])
	assert.Nil(t, result["lastUsedAt"])
	assert.False(t, result["isStale"].(bool))
}

func TestConnectedProviderResponse(t *testing.T) {
	now := time.Now()
	provider := &storage.Provider{
		ID:         uuid.New().String(),
		Provider:   "functionfly-edge",
		Status:     "active",
		CreatedAt:  now,
		LastUsedAt: &now,
	}

	result := connectedProviderResponse(provider)

	assert.Equal(t, provider.ID, result["id"])
	assert.Equal(t, "functionfly-edge", result["name"])
	assert.Equal(t, "online", result["status"])
	assert.Equal(t, now.Format(time.RFC3339), result["connectedAt"])
	assert.NotNil(t, result["lastUsedAt"])
}

func TestConnectedProviderResponse_StatusMapping(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		expectedStatus string
	}{
		{"active provider", "active", "online"},
		{"inactive provider", "inactive", "offline"},
		{"error provider", "error", "degraded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &storage.Provider{
				ID:        uuid.New().String(),
				Provider:  "functionfly-edge",
				Status:    tt.status,
				CreatedAt: time.Now(),
			}

			result := connectedProviderResponse(provider)

			assert.Equal(t, tt.expectedStatus, result["status"])
		})
	}
}
