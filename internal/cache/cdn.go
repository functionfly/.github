package cache

import (
	"net/http"
	"strconv"
)

// SetCDNHeaders sets appropriate Cache-Control headers based on cache eligibility
// This enables CDN-level caching for deterministic public functions
func SetCDNHeaders(w http.ResponseWriter, eligibility EligibilityResult, isPublic bool) {
	if !eligibility.Eligible {
		// Never cache non-eligible responses
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		return
	}

	if !isPublic || !eligibility.CanUseCDN {
		// Private cache - can be cached by browser/CDN with auth, but not shared
		w.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(eligibility.TTL))
		return
	}

	// Public CDN cache - can be cached at edge and shared
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(eligibility.TTL))
	w.Header().Set("X-Cache-Status", "MISS") // Will be "HIT" at edge
}

// SetNoCacheHeaders sets headers to prevent caching
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetCacheControlHeader sets a custom Cache-Control header value
func SetCacheControlHeader(w http.ResponseWriter, directive string, maxAge int) {
	if maxAge > 0 {
		w.Header().Set("Cache-Control", directive+", max-age="+strconv.Itoa(maxAge))
	} else {
		w.Header().Set("Cache-Control", directive)
	}
}

// GetCDNCacheKey returns the CDN cache key components for debugging
func GetCDNCacheKey(functionID, version, inputHash string) string {
	return "fx:" + functionID + ":" + version + ":" + inputHash[:8]
}

// CDNConfig holds CDN configuration
type CDNConfig struct {
	EnableCDNCaching bool   // Toggle for CDN caching
	CDNMaxAge        int    // Default CDN max-age (seconds)
	SDKBasePath      string // Base path for SDK assets
	DocsBasePath     string // Base path for documentation assets
	StaticBasePath   string // Base path for other static assets
}

// NewCDNConfig creates a default CDN configuration
func NewCDNConfig() *CDNConfig {
	return &CDNConfig{
		EnableCDNCaching: true,
		CDNMaxAge:        3600,
		SDKBasePath:      "/sdk",
		DocsBasePath:     "/docs",
		StaticBasePath:   "/static",
	}
}
