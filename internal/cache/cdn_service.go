package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	// Cloudflare cache purge
	cloudflareZoneID string
	cloudflareToken  string
	httpClient       *http.Client
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

	baseURL := config.CDNBaseURL
	if baseURL == "" {
		baseURL = "https://cdn.functionfly.com"
	}
	// Normalize: no trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	service := &CDNService{
		cdnURL:           baseURL,
		enabled:          config.EnableCDNCaching,
		maxAge:           config.CDNMaxAge,
		providers:        make(map[string]*CDNProvider),
		cloudflareZoneID: config.CloudflareZoneID,
		cloudflareToken:  config.CloudflareToken,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
	}

	// Initialize default providers if enabled
	if config.EnableCDNCaching {
		service.initializeProviders(baseURL)
	}

	return service
}

// initializeProviders sets up CDN providers (baseURL is the configured CDN base, e.g. https://cdn.functionfly.com)
func (c *CDNService) initializeProviders(baseURL string) {
	if baseURL == "" {
		baseURL = "https://cdn.functionfly.com"
	}
	// Cloudflare CDN (default) – use configured base URL
	c.providers["cloudflare"] = &CDNProvider{
		Name:     "cloudflare",
		BaseURL:  baseURL,
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

// generateETag generates an ETag based on file path and modification time
func (c *CDNService) generateETag(path string) string {
	// Use path + mtime to generate a deterministic ETag
	// This provides proper cache invalidation when files change
	info, err := os.Stat(path)
	if err != nil {
		// Fall back to path hash if file info unavailable
		h := sha256.Sum256([]byte(path))
		return fmt.Sprintf(`"%x"`, h[:16])
	}

	// Combine path with mtime for content-aware ETag
	data := fmt.Sprintf("%s:%d", path, info.ModTime().Unix())
	h := sha256.Sum256([]byte(data))
	return fmt.Sprintf(`"%x"`, h[:16])
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
	// Resolve and validate the file path to prevent path traversal
	absPath, err := c.resolveFilePath(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if file exists and is readable
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Failed to access file", http.StatusInternalServerError)
		return
	}

	// Don't serve directories
	if info.IsDir() {
		http.Error(w, "Requested path is a directory", http.StatusForbidden)
		return
	}

	// Set CDN headers
	c.SetCDNHeaders(w, filePath)

	// Handle range requests for efficient partial content delivery (streaming media, large files)
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		served, err := c.serveFileWithRange(w, r, absPath, info)
		if err != nil {
			if !served {
				http.Error(w, "Failed to serve file", http.StatusInternalServerError)
			}
			return
		}
		return
	}

	// For HEAD requests or other methods, serve the full file
	file, err := os.Open(absPath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	http.ServeContent(w, r, filePath, info.ModTime(), file)
}

// resolveFilePath resolves a request path to an absolute filesystem path
// It prevents path traversal attacks by sanitizing the input
func (c *CDNService) resolveFilePath(requestPath string) (string, error) {
	// Get the static files root directory from config, default to ./static
	root := "./static"

	// Clean the request path to remove any traversal components
	cleanPath := filepath.Clean(requestPath)

	// Ensure the resolved path is within the static root directory
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("invalid static root path")
	}
	absPath := filepath.Join(absRoot, cleanPath)

	// Verify the resolved path is within the root (prevent traversal)
	if !strings.HasPrefix(absPath, absRoot) {
		return "", fmt.Errorf("access denied: path outside static root")
	}

	return absPath, nil
}

// serveFileWithRange handles HTTP range requests for efficient partial content delivery
func (c *CDNService) serveFileWithRange(w http.ResponseWriter, r *http.Request, absPath string, info os.FileInfo) (bool, error) {
	file, err := os.Open(absPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Get content type
	contentType := c.getContentType(absPath)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Check for Range header
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// No range requested, serve full file
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
		if r.Method == http.MethodHead {
			return true, nil
		}
		_, err = io.Copy(w, file)
		return true, err
	}

	// Parse the range header (simplified - supports single range)
	// Format: "bytes=start-end"
	rangePart := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangePart, "-")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid range header")
	}

	start, _ := strconv.ParseInt(parts[0], 10, 64)
	end, _ := strconv.ParseInt(parts[1], 10, 64)

	if end == 0 {
		end = info.Size() - 1
	}
	if start > end || start >= info.Size() {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", info.Size()))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return true, nil
	}

	// Seek to start position
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return false, err
	}

	contentLength := end - start + 1
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
	w.WriteHeader(http.StatusPartialContent)

	if r.Method == http.MethodHead {
		return true, nil
	}

	// Copy the range
	_, err = io.CopyN(w, file, contentLength)
	return true, err
}

// PurgeCDNCache purges the CDN cache for a specific path via Cloudflare's Cache Purge API.
// For R2-backed CDN (cdn.functionfly.com), this clears the file from Cloudflare's edge cache,
// causing the next request to re-fetch from the R2 origin.
func (c *CDNService) PurgeCDNCache(path string) error {
	return c.PurgeCDNURLs([]string{c.cdnURL + path})
}

// PurgeCDNURLs purges multiple full URLs from the CDN cache.
// Use this for precise control over which URLs to invalidate.
func (c *CDNService) PurgeCDNURLs(urls []string) error {
	if !c.enabled {
		return nil
	}

	provider := c.getBestProvider()
	if provider == nil {
		return fmt.Errorf("no CDN provider available")
	}

	switch provider.Name {
	case "cloudflare":
		return c.purgeCloudflare(urls)
	case "cloudfront":
		return c.purgeCloudFront(urls)
	case "fastly":
		return c.purgeFastly(urls)
	default:
		logrus.Warnf("CDN provider %s does not support cache purge", provider.Name)
		return nil
	}
}

// PurgeCDNByPrefix purges all URLs under a given CDN path prefix.
// Cloudflare's prefix purge uses a cache tag / URL pattern approach.
// Falls back to prefix-based invalidation by issuing individual purges for
// known asset paths under the prefix.
func (c *CDNService) PurgeCDNByPrefix(prefix string) error {
	if !c.enabled {
		return nil
	}

	provider := c.getBestProvider()
	if provider == nil {
		return fmt.Errorf("no CDN provider available")
	}

	fullPrefix := c.cdnURL + prefix
	logrus.Infof("Purging CDN cache for prefix: %s via provider: %s", fullPrefix, provider.Name)

	// Cloudflare supports prefix-based purge via the "prefixes" field
	if provider.Name == "cloudflare" {
		return c.purgeCloudflarePrefix(fullPrefix)
	}

	// For providers that don't support prefix purge, warn and skip
	logrus.Warnf("Provider %s does not support prefix-based purge", provider.Name)
	return nil
}

// purgeCloudflare calls Cloudflare's Cache Purge API to invalidate cached URLs.
// API docs: https://developers.cloudflare.com/cache/how-to/purge-cache/
func (c *CDNService) purgeCloudflare(urls []string) error {
	if c.cloudflareZoneID == "" || c.cloudflareToken == "" {
		return fmt.Errorf("Cloudflare Zone ID and API token are required for cache purge")
	}

	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/cache/purge", c.cloudflareZoneID)

	payload := map[string]interface{}{
		"files": urls,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal purge request: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create purge request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cloudflareToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Cloudflare cache purge API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Cloudflare cache purge failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result cfPurgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse Cloudflare purge response: %w", err)
	}

	if !result.Success {
		var errMsgs []string
		for _, e := range result.Errors {
			errMsgs = append(errMsgs, e.Message)
		}
		return fmt.Errorf("Cloudflare purge errors: %v", errMsgs)
	}

	logrus.Infof("Cloudflare cache purged for %d URL(s)", len(urls))
	return nil
}

// purgeCloudflarePrefix calls Cloudflare's Cache Purge API with a URL prefix.
// This invalidates all cached URLs matching the prefix in one API call.
func (c *CDNService) purgeCloudflarePrefix(prefix string) error {
	if c.cloudflareZoneID == "" || c.cloudflareToken == "" {
		return fmt.Errorf("Cloudflare Zone ID and API token are required for cache purge")
	}

	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/cache/purge", c.cloudflareZoneID)

	payload := map[string]interface{}{
		"prefixes": []string{prefix},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal purge request: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create purge request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cloudflareToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Cloudflare cache purge API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Cloudflare cache purge failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result cfPurgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse Cloudflare purge response: %w", err)
	}

	if !result.Success {
		var errMsgs []string
		for _, e := range result.Errors {
			errMsgs = append(errMsgs, e.Message)
		}
		return fmt.Errorf("Cloudflare purge errors: %v", errMsgs)
	}

	logrus.Infof("Cloudflare cache purged for prefix: %s", prefix)
	return nil
}

// cfPurgeResponse represents Cloudflare Cache Purge API response
type cfPurgeResponse struct {
	Success bool        `json:"success"`
	Errors  []cfPurgeError `json:"errors"`
}

type cfPurgeError struct {
	Message string `json:"message"`
}

// purgeCloudFront calls AWS CloudFront's cache purge API.
// Requires AWS credentials with CloudFront distribution access.
func (c *CDNService) purgeCloudFront(urls []string) error {
	// TODO: Implement CloudFront cache purge using AWS SDK
	// CloudFront has a CreateInvalidation API that can be called via AWS SDK for Go v2
	logrus.Warnf("CloudFront cache purge not yet implemented, %d URL(s) not purged", len(urls))
	return nil
}

// purgeFastly calls Fastly's purge API for a single URL.
// Fastly also supports purging by Surrogate-Key (prefix) but that requires additional setup.
func (c *CDNService) purgeFastly(urls []string) error {
	// TODO: Implement Fastly cache purge
	// Fastly API: POST https://api.fastly.com/purge/<url>
	// Or use soft purge: POST https://api.fastly.com/purge/<url> with Surrogate-Key header
	logrus.Warnf("Fastly cache purge not yet implemented, %d URL(s) not purged", len(urls))
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
