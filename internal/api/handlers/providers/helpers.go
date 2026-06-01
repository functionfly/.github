package providers

import (
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// isProviderStale returns true if the provider hasn't been used in more than 30 days
func isProviderStale(p *storage.Provider) bool {
	if p.LastUsedAt == nil {
		return time.Since(p.CreatedAt) > 30*24*time.Hour
	}
	return time.Since(*p.LastUsedAt) > 30*24*time.Hour
}

// providerResponse builds the dashboard JSON shape for a connected provider.
func providerResponse(p *storage.Provider) map[string]interface{} {
	status := "pending"
	switch p.Status {
	case "active":
		status = "online"
	case "inactive":
		status = "offline"
	case "error":
		status = "degraded"
	}

	result := map[string]interface{}{
		"id":          p.ID,
		"name":        p.Provider,
		"status":      status,
		"connectedAt": p.CreatedAt.Format(time.RFC3339),
		"isStale":     isProviderStale(p),
	}

	if p.LastUsedAt != nil {
		result["lastUsedAt"] = p.LastUsedAt.Format(time.RFC3339)
	}

	return result
}

// listProviderFromStorage maps a storage.Provider to the list view format.
func listProviderFromStorage(p *storage.Provider) map[string]interface{} {
	return providerResponse(p)
}

// connectedProviderResponse is an alias for providerResponse.
func connectedProviderResponse(p *storage.Provider) map[string]interface{} {
	return providerResponse(p)
}

// maskAPIKey returns a masked version of an API key showing first 4 and last 4 characters.
// For pipe-delimited AWS credentials, masks each field individually.
func maskAPIKey(key string) string {
	if strings.Contains(key, "|") {
		parts := strings.SplitN(key, "|", 4)
		masked := make([]string, len(parts))
		for i, p := range parts {
			masked[i] = maskAPIKey(p)
		}
		return strings.Join(masked, "|")
	}
	if len(key) <= 12 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// mapProviderIDForValidation maps frontend provider slugs to backend validation names.
func mapProviderIDForValidation(providerID string) string {
	if providerID == "workers" {
		return "cloudflare"
	}
	return providerID
}
