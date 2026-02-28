package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
)

// Handler contains registry API handlers
type Handler struct {
	repo            *registry.RegistryRepository
	backendRepo     storage.Repository
	cacheService    *cache.CacheService
	cdnService      *cache.CDNService
	edgeCache       *cache.EdgeCacheService
	realtimeMonitor *monitoring.RealtimeMonitor
}

// NewHandler creates a new registry handler
func NewHandler(repo *registry.RegistryRepository, backendRepo storage.Repository, cacheService *cache.CacheService, cdnService *cache.CDNService, edgeCache *cache.EdgeCacheService, realtimeMonitor *monitoring.RealtimeMonitor) *Handler {
	return &Handler{
		repo:            repo,
		backendRepo:     backendRepo,
		cacheService:    cacheService,
		cdnService:      cdnService,
		edgeCache:       edgeCache,
		realtimeMonitor: realtimeMonitor,
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in case of multiple
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx > 0 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// HandleGetSDKCode handles generating SDK code for a function
func (h *Handler) HandleGetSDKCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	sdk := r.URL.Query().Get("sdk") // javascript, python, go

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
		return
	}

	var code string
	switch sdk {
	case "python":
		code = generatePythonSDK(fn.Author, fn.Name, fnVersion.Version, fn.Title.String, fn.Description.String)
	case "go":
		code = generateGoSDK(fn.Author, fn.Name, fnVersion.Version, fn.Title.String, fn.Description.String)
	default:
		code = generateJavaScriptSDK(fn.Author, fn.Name, fnVersion.Version, fn.Title.String, fn.Description.String)
	}

	response := map[string]interface{}{
		"sdk":    sdk,
		"code":   code,
		"author": author,
		"name":   name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleServeSDK serves static SDK files via CDN
func (h *Handler) HandleServeSDK(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sdkType := vars["sdk"]       // javascript, python, go
	version := vars["version"]   // latest, v1.0.0, etc.
	filename := vars["filename"] // functionfly.js, functionfly.py, etc.

	if h.cdnService == nil || !h.cdnService.IsCDNEnabled() {
		cache.RecordCDNMiss()
		h.serveSDKLocally(w, r, sdkType, version, filename)
		return
	}

	// Construct CDN path
	cdnPath := fmt.Sprintf("/sdk/%s/%s/%s", sdkType, version, filename)

	// Set CDN headers
	h.cdnService.SetCDNHeaders(w, cdnPath)

	// Record CDN hit (in production, this would be determined by CDN)
	cache.RecordCDNHit()

	// For now, serve locally with CDN headers
	// In production, this would redirect to or proxy from CDN
	h.serveSDKLocally(w, r, sdkType, version, filename)
}

// HandleServeDocs serves documentation files via CDN
func (h *Handler) HandleServeDocs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docType := vars["type"]    // api, sdk, guides
	version := vars["version"] // latest, v1.0.0, etc.
	path := vars["path"]       // getting-started.md, api-reference.html, etc.

	if h.cdnService == nil || !h.cdnService.IsCDNEnabled() {
		cache.RecordCDNMiss()
		h.serveDocsLocally(w, r, docType, version, path)
		return
	}

	// Construct CDN path
	cdnPath := fmt.Sprintf("/docs/%s/%s/%s", docType, version, path)

	// Set CDN headers
	h.cdnService.SetCDNHeaders(w, cdnPath)

	// Record CDN hit
	cache.RecordCDNHit()

	// For now, serve locally with CDN headers
	h.serveDocsLocally(w, r, docType, version, path)
}

// HandleServeStatic serves other static assets via CDN
func (h *Handler) HandleServeStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"] // images, css, js, fonts, etc.
	path := vars["path"]         // logo.png, styles.css, etc.

	if h.cdnService == nil || !h.cdnService.IsCDNEnabled() {
		cache.RecordCDNMiss()
		h.serveStaticLocally(w, r, category, path)
		return
	}

	// Construct CDN path
	cdnPath := fmt.Sprintf("/static/%s/%s", category, path)

	// Set CDN headers
	h.cdnService.SetCDNHeaders(w, cdnPath)

	// Record CDN hit
	cache.RecordCDNHit()

	// For now, serve locally with CDN headers
	h.serveStaticLocally(w, r, category, path)
}

// HandleGetCacheStats returns comprehensive cache statistics
// This endpoint provides public cache metrics for monitoring and debugging
func (h *Handler) HandleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"cache_enabled":      h.cacheService != nil,
		"redis_enabled":      h.cacheService != nil && h.cacheService.IsRedisCacheEnabled(),
		"cdn_enabled":        h.cdnService != nil && h.cdnService.IsCDNEnabled(),
		"edge_cache_enabled": h.edgeCache != nil,
	}

	// Get comprehensive cache service stats if available
	if h.cacheService != nil {
		// Memory cache stats (L1)
		if memStats := h.cacheService.GetMemoryStats(); memStats != nil {
			stats["memory_cache"] = map[string]interface{}{
				"layer":        "L1",
				"type":         "memory",
				"hits":         memStats.Hits,
				"misses":       memStats.Misses,
				"hit_ratio":    memStats.Ratio,
				"size_bytes":   memStats.SizeBytes,
				"evictions":    memStats.Evictions,
			}
		}

		// Disk cache stats (L2)
		if diskStats, err := h.cacheService.GetDiskStats(); err == nil && diskStats != nil {
			var hitRatio float64
			if diskStats.TotalHits > 0 {
				totalLookups := diskStats.TotalEntries + diskStats.TotalHits // Approximation
				if totalLookups > 0 {
					hitRatio = float64(diskStats.TotalHits) / float64(totalLookups)
				}
			}
			stats["disk_cache"] = map[string]interface{}{
				"layer":           "L2",
				"type":            "disk",
				"total_entries":   diskStats.TotalEntries,
				"total_size_bytes": diskStats.TotalSizeBytes,
				"total_hits":      diskStats.TotalHits,
				"hit_ratio":       hitRatio,
				"expired_entries": diskStats.ExpiredEntries,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
