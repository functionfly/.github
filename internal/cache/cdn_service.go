package cache

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// CDNService handles CDN configuration and headers for static assets
type CDNService struct {
	cdnURL     string
	enabled    bool
	maxAge     int
	providers  map[string]*CDNProvider
}

// CDNProvider represents a CDN provider configuration
type CDNProvider struct {
	Name       string
	BaseURL    string
	Regions    []string
	Enabled    bool
	Priority   int // Higher priority = preferred
}

// CDNConfig is defined in cdn.go

// NewCDNService creates a new CDN service
func NewCDNService(config *CDNConfig) *CDNService {
	if config == nil {
		config = NewCDNConfig()
	}

	service := &CDNService{
		cdnURL:    "",
		enabled:   config.EnableCDNCaching,
		maxAge:    config.CDNMaxAge,
		providers: make(map[string]*CDNProvider),
	}

	// Initialize default providers if enabled
	if config.EnableCDNCaching {
		service.initializeProviders()
	}

	return service
}

// initializeProviders sets up CDN providers
func (c *CDNService) initializeProviders() {
	// Cloudflare CDN (default)
	c.providers["cloudflare"] = &CDNProvider{
		Name:     "cloudflare",
		BaseURL:  "https://cdn.functionfly.dev",
		Regions:  []string{"global"},
		Enabled:  true,
		Priority: 100,
	}

	// AWS CloudFront
	c.providers["cloudfront"] = &CDNProvider{
		Name:     "cloudfront",
		BaseURL:  "https://d1234567890.cloudfront.net",
		Regions:  []string{"us-east-1", "us-west-2", "eu-west-1"},
		Enabled:  false, // Disabled by default
		Priority: 90,
	}

	// Fastly
	c.providers["fastly"] = &CDNProvider{
		Name:     "fastly",
		BaseURL:  "https://functionfly.global.ssl.fastly.net",
		Regions:  []string{"global"},
		Enabled:  false, // Disabled by default
		Priority: 80,
	}
}

// EnableProvider enables a CDN provider
func (c *CDNService) EnableProvider(name string) {
	if provider, exists := c.providers[name]; exists {
		provider.Enabled = true
		logrus.Infof("Enabled CDN provider: %s", name)
	}
}

// DisableProvider disables a CDN provider
func (c *CDNService) DisableProvider(name string) {
	if provider, exists := c.providers[name]; exists {
		provider.Enabled = false
		logrus.Infof("Disabled CDN provider: %s", name)
	}
}

// GetCDNURL returns the CDN URL for a given path
func (c *CDNService) GetCDNURL(path string) string {
	if !c.enabled {
		return path // Return original path if CDN is disabled
	}

	provider := c.getBestProvider()
	if provider == nil {
		return path
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return provider.BaseURL + path
}

// getBestProvider returns the highest priority enabled provider
func (c *CDNService) getBestProvider() *CDNProvider {
	var bestProvider *CDNProvider
	bestPriority := -1

	for _, provider := range c.providers {
		if provider.Enabled && provider.Priority > bestPriority {
			bestProvider = provider
			bestPriority = provider.Priority
		}
	}

	return bestProvider
}

// SetCDNHeaders sets appropriate CDN headers for static assets
func (c *CDNService) SetCDNHeaders(w http.ResponseWriter, path string) {
	if !c.enabled {
		return
	}

	// Determine content type and cache settings based on path
	contentType := c.getContentType(path)
	maxAge := c.getMaxAge(path)

	// Set cache control headers
	cacheControl := fmt.Sprintf("public, max-age=%d, s-maxage=%d", maxAge, maxAge*2)
	w.Header().Set("Cache-Control", cacheControl)

	// Set content type if detected
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	// Set CDN-specific headers
	w.Header().Set("X-CDN-Provider", c.getBestProvider().Name)
	w.Header().Set("X-CDN-Cache", "HIT") // This would be set by CDN, but we set it for debugging

	// Set Last-Modified and ETag if available (simplified)
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, c.generateETag(path)))
}

// getContentType determines content type based on file extension
func (c *CDNService) getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".js", ".mjs":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	default:
		return ""
	}
}

// getMaxAge determines cache max-age based on content type
func (c *CDNService) getMaxAge(path string) int {
	// SDK files - cache for 1 hour
	if strings.Contains(path, "/sdk/") {
		return 3600
	}

	// Documentation - cache for 30 minutes
	if strings.Contains(path, "/docs/") {
		return 1800
	}

	// Static assets - cache for 24 hours
	if strings.Contains(path, "/static/") {
		return 86400
	}

	// Default
	return c.maxAge
}

// generateETag generates a simple ETag for the path
func (c *CDNService) generateETag(path string) string {
	// In a real implementation, this would be based on file content hash
	// For now, use a simple hash of path and timestamp
	return fmt.Sprintf("%x", len(path))
}

// IsCDNEnabled returns whether CDN is enabled
func (c *CDNService) IsCDNEnabled() bool {
	return c.enabled
}

// GetProviders returns all configured CDN providers
func (c *CDNService) GetProviders() map[string]*CDNProvider {
	return c.providers
}

// ServeStaticAsset serves a static asset with CDN headers
func (c *CDNService) ServeStaticAsset(w http.ResponseWriter, r *http.Request, filePath string) {
	// Set CDN headers
	c.SetCDNHeaders(w, filePath)

	// In a real implementation, you would serve the actual file
	// For now, just return a placeholder response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Static asset: %s (served via CDN)", filePath)
}

// PurgeCDNCache purges the CDN cache for a specific path
func (c *CDNService) PurgeCDNCache(path string) error {
	if !c.enabled {
		return nil
	}

	provider := c.getBestProvider()
	if provider == nil {
		return fmt.Errorf("no CDN provider available")
	}

	// In a real implementation, this would call the CDN provider's API
	// to purge the cache for the given path
	logrus.Infof("Purging CDN cache for path: %s via provider: %s", path, provider.Name)

	return nil
}

// GetCDNStats returns CDN usage statistics
func (c *CDNService) GetCDNStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":     c.enabled,
		"providers":   len(c.providers),
		"default_max_age": c.maxAge,
		"active_provider": func() string {
			if provider := c.getBestProvider(); provider != nil {
				return provider.Name
			}
			return "none"
		}(),
	}
}