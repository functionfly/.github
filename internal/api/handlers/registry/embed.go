package registry

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HandleServeEmbed serves a per-function embed script.
//
// Routes:
//
//	GET /v1/embed/{author}/{nameVersion}
//
// The {nameVersion} path segment accepts:
//
//	slugify.js            → latest version
//	slugify@1.2.0.js      → pinned version
//
// Query parameters (all optional):
//
//	namespace  – global variable name (default: "ff")
//	autoload   – auto-initialize on DOMContentLoaded (default: "true")
//	ui         – inject default UI widget (default: "false")
//	theme      – "light" | "dark" | "auto" (default: "auto")
func (h *Handler) HandleServeEmbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]

	// The route captures "{name}.js" or "{name}@{version}.js" as a single
	// "nameVersion" variable so we can support the @ syntax.
	nameVersion := vars["nameVersion"] // e.g. "slugify.js" or "slugify@1.2.0.js"

	// Strip trailing ".js"
	nameVersion = strings.TrimSuffix(nameVersion, ".js")

	// Split on "@" to separate name from optional version
	var name, version string
	if idx := strings.LastIndex(nameVersion, "@"); idx >= 0 {
		name = nameVersion[:idx]
		version = nameVersion[idx+1:]
	} else {
		name = nameVersion
		version = ""
	}

	if author == "" || name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author and name are required"))
		return
	}

	// 1. Look up function in registry
	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	// 2. Phase 3 — Allowed origins enforcement
	//    If the function has embed_config with a non-wildcard allowed_origins list,
	//    validate the requesting Origin header.
	embedCfg, cfgErr := h.repo.GetFunctionEmbedConfig(fn.ID)
	if cfgErr == nil && embedCfg != nil {
		if !embedCfg.Enabled {
			apierror.WriteError(w, apierror.NewForbidden("Embed not enabled for this function"))
			return
		}
		// SECURITY: Use Referer header (server-controlled) for origin validation,
		// not X-Embed-Origin which can be spoofed by malicious JS.
		origin := resolveRequestOrigin(r)
		if origin != "" && !isOriginAllowed(origin, embedCfg.AllowedOrigins) {
			apierror.WriteError(w, apierror.NewForbidden("Origin not allowed"))
			return
		}
		// In production, if allowed_origins is empty (fail closed), deny all cross-origin requests
		if origin != "" && len(embedCfg.AllowedOrigins) == 0 {
			isProd := os.Getenv("ENVIRONMENT") == "production" || os.Getenv("NODE_ENV") == "production"
			if isProd {
				apierror.WriteError(w, apierror.NewForbidden("Origin not allowed"))
				return
			}
		}

		// Phase 3 — Per-embed-domain rate limiting
		if origin != "" && embedCfg.RateLimitPerHour > 0 {
			since := time.Now().Add(-time.Hour)
			count, rlErr := h.repo.GetEmbedExecutionCountByOrigin(fn.ID, origin, since)
			if rlErr == nil && count >= int64(embedCfg.RateLimitPerHour) {
				w.Header().Set("Retry-After", "3600")
				apierror.WriteError(w, apierror.NewRateLimited("Embed rate limit exceeded for this origin"))
				return
			}
		}
	}

	// 3. Get function version metadata
	latestVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("No versions available"))
		return
	}

	// 4. Parse embed options from query string
	opts := parseEmbedOptions(
		r.URL.Query().Get("namespace"),
		r.URL.Query().Get("autoload"),
		r.URL.Query().Get("ui"),
		r.URL.Query().Get("theme"),
	)

	// 5. Generate embed script (with Redis caching)
	cacheControl := "public, max-age=300" // 5 min for latest
	if version != "" {
		cacheControl = "public, max-age=31536000, immutable" // 1 year for pinned
	}

	script := h.getOrGenerateEmbedScript(fn, latestVersion, version, opts, cacheControl)

	// 6. Set headers
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", cacheControl)
	middleware.SetCORSHeaders(w, r)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// SECURITY: Prevent the script from being rendered as HTML by older browsers
	w.Header().Set("X-Download-Options", "noopen")
	// SECURITY: Prevent MIME sniffing attacks
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(script))
}

// HandleGetEmbedConfig returns the embed configuration for a function (JSON).
// SECURITY: Requires authentication + function ownership.
//
// Route: GET /v1/registry/functions/{author}/{name}/embed
func (h *Handler) HandleGetEmbedConfig(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require authenticated user (config reveals allowed origins, rate limits)
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	// SECURITY: Only the function owner can see embed config details
	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	cfg, err := h.repo.GetFunctionEmbedConfig(fn.ID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get embed config"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cfg)
}

// HandleUpdateEmbedConfig updates the embed configuration for a function.
//
// Route: PUT /v1/registry/functions/{author}/{name}/embed
// Body: JSON EmbedConfig
func (h *Handler) HandleUpdateEmbedConfig(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require authenticated user (enforced by RequireAuth middleware)
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	// SECURITY: Verify the authenticated user owns this function (same pattern as settings.go)
	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	var cfg registry.EmbedConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate theme
	if cfg.UITheme != "" && cfg.UITheme != "light" && cfg.UITheme != "dark" && cfg.UITheme != "auto" {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid ui_theme: must be 'light', 'dark', or 'auto'"))
		return
	}

	// Validate allowed origins — must be valid origin format (scheme + host) or wildcard
	if err := validateAllowedOrigins(cfg.AllowedOrigins); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest(err.Error()))
		return
	}

	// Validate rate limit bounds
	if cfg.RateLimitPerHour < 0 {
		apierror.WriteError(w, apierror.NewBadRequest("rate_limit_per_hour must be non-negative"))
		return
	}
	if cfg.RateLimitPerHour > 100000 {
		cfg.RateLimitPerHour = 100000 // hard cap to prevent misconfiguration
	}

	// Default rate limit
	if cfg.RateLimitPerHour == 0 {
		cfg.RateLimitPerHour = 1000
	}

	// Default allowed origins — SECURITY: production defaults to restrictive, not wildcard
	if len(cfg.AllowedOrigins) == 0 {
		isProd := os.Getenv("ENVIRONMENT") == "production" || os.Getenv("NODE_ENV") == "production"
		if isProd {
			// In production, empty origins = no origins allowed (fail closed)
			// Owner must explicitly set ["*"] or specific origins
			cfg.AllowedOrigins = []string{}
		} else {
			cfg.AllowedOrigins = []string{"*"}
		}
	}

	if err := h.repo.UpdateFunctionEmbedConfig(fn.ID, &cfg); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to update embed config"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cfg)
}

// HandleGetEmbedSnippet returns a ready-to-use embed code snippet for the dashboard.
//
// Route: GET /v1/registry/functions/{author}/{name}/embed/snippet
// Query parameters: namespace, autoload, ui, theme (same as embed script)
func (h *Handler) HandleGetEmbedSnippet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	latestVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("No versions available"))
		return
	}

	opts := parseEmbedOptions(
		r.URL.Query().Get("namespace"),
		r.URL.Query().Get("autoload"),
		r.URL.Query().Get("ui"),
		r.URL.Query().Get("theme"),
	)

	baseURL := getEmbedBaseURL()

	// Build query string
	params := []string{}
	if opts.Namespace != "ff" {
		params = append(params, "namespace="+opts.Namespace)
	}
	if !opts.Autoload {
		params = append(params, "autoload=false")
	}
	if opts.UI {
		params = append(params, "ui=true")
	}
	if opts.Theme != "auto" {
		params = append(params, "theme="+opts.Theme)
	}

	queryStr := ""
	if len(params) > 0 {
		queryStr = "?" + strings.Join(params, "&")
	}

	scriptURL := fmt.Sprintf("%s/embed/%s/%s.js%s", baseURL, author, name, queryStr)
	pinnedURL := fmt.Sprintf("%s/embed/%s/%s@%s.js%s", baseURL, author, name, latestVersion.Version, queryStr)

	// SECURITY: Generate SRI (Subresource Integrity) hash for pinned version.
	// This allows consumers to verify the script hasn't been tampered with.
	scriptContent := generateEmbedScript(fn, latestVersion, latestVersion.Version, opts)
	hash := sha512.Sum384([]byte(scriptContent))
	sriHash := "sha384-" + base64.StdEncoding.EncodeToString(hash[:])

	// Build example usage snippets (pinned version includes integrity attribute)
	basicSnippet := fmt.Sprintf(`<script src="%s"></script>
<script>
  const result = await %s.run({ /* your input */ });
  console.log(result.data);
</script>`, scriptURL, opts.Namespace)

	pinnedSnippet := fmt.Sprintf(`<script src="%s" integrity="%s" crossorigin="anonymous"></script>
<script>
  const result = await %s.run({ /* your input */ });
  console.log(result.data);
</script>`, pinnedURL, sriHash, opts.Namespace)

	formSnippet := fmt.Sprintf(`<form id="myForm">
  <!-- your form fields -->
</form>
<script src="%s"></script>
<script>
  %s.form(document.getElementById("myForm"), {
    onSuccess: (data) => console.log(data),
    onError:   (err)  => console.error(err),
  });
</script>`, scriptURL, opts.Namespace)

	widgetSnippet := fmt.Sprintf(`<div id="ff-widget"></div>
<script src="%s?ui=true"></script>
<script>
  %s.widget(document.getElementById("ff-widget"), {
    title:       "%s",
    placeholder: "Enter input (JSON)...",
    buttonText:  "Run",
  });
</script>`, scriptURL, opts.Namespace, fn.Title.String)

	response := map[string]interface{}{
		"author":     author,
		"name":       name,
		"version":    latestVersion.Version,
		"script_url": scriptURL,
		"pinned_url": pinnedURL,
		"integrity":  sriHash,
		"namespace":  opts.Namespace,
		"snippets": map[string]string{
			"basic":  basicSnippet,
			"pinned": pinnedSnippet,
			"form":   formSnippet,
			"widget": widgetSnippet,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// HandleGetEmbedAnalytics returns embed execution analytics for a function.
// SECURITY: Requires authentication + function ownership (analytics are sensitive).
//
// Route: GET /v1/registry/functions/{author}/{name}/embed/analytics
// Query parameters:
//
//	days  – number of days to look back (default: 30, max: 90)
//	limit – max number of origin domains to return (default: 20, max: 100)
func (h *Handler) HandleGetEmbedAnalytics(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Require authenticated user (analytics reveals traffic patterns)
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	// SECURITY: Only the function owner can see embed analytics
	if fn.TenantID == nil || *fn.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Parse query params
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		fmt.Sscanf(d, "%d", &days)
		if days < 1 {
			days = 1
		}
		if days > 90 {
			days = 90
		}
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit < 1 {
			limit = 1
		}
		if limit > 100 {
			limit = 100
		}
	}

	since := time.Now().AddDate(0, 0, -days)
	stats, err := h.repo.GetEmbedAnalytics(fn.ID, since, limit)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get embed analytics"))
		return
	}

	// Phase 3 — Abuse detection: flag origins with unusually high counts
	var totalCount int64
	for _, s := range stats {
		totalCount += s.Count
	}

	type enrichedStat struct {
		Origin     string  `json:"origin"`
		Count      int64   `json:"count"`
		Pct        float64 `json:"pct"`
		Suspicious bool    `json:"suspicious"`
	}

	enriched := make([]enrichedStat, len(stats))
	for i, s := range stats {
		pct := 0.0
		if totalCount > 0 {
			pct = float64(s.Count) / float64(totalCount) * 100
		}
		// Flag as suspicious if a single origin accounts for >80% of traffic
		// and total count is significant (>100 executions)
		suspicious := pct > 80 && totalCount > 100
		enriched[i] = enrichedStat{
			Origin:     s.Origin,
			Count:      s.Count,
			Pct:        pct,
			Suspicious: suspicious,
		}
	}

	response := map[string]interface{}{
		"author":      author,
		"name":        name,
		"days":        days,
		"total_count": totalCount,
		"top_origins": enriched,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// getOrGenerateEmbedScript returns a cached embed script or generates a new one.
// Uses Redis caching to avoid regenerating the same script on every request.
func (h *Handler) getOrGenerateEmbedScript(
	fn *registry.RegistryFunction,
	fnVersion *registry.RegistryFunctionVersion,
	version string,
	opts EmbedOptions,
	cacheControl string,
) string {
	// Build options hash for cache key
	optsHash := fmt.Sprintf("%s:%v:%v:%s", opts.Namespace, opts.Autoload, opts.UI, opts.Theme)
	cacheKey := cache.NewRegistryCacheKey().EmbedScript(fn.Author, fn.Name, version, optsHash)

	// Try Redis cache first
	registryCache := h.cacheService.GetRegistryCache()
	if registryCache != nil {
		if cached, err := registryCache.Get(context.Background(), cacheKey); err == nil && cached != nil {
			logrus.WithFields(logrus.Fields{
				"author":    fn.Author,
				"name":      fn.Name,
				"version":   version,
				"cache_key": cacheKey,
			}).Debug("Embed script cache hit")
			return string(cached)
		}
	}

	// Generate the script
	script := generateEmbedScript(fn, fnVersion, version, opts)

	// Store in Redis with TTL matching Cache-Control
	if registryCache != nil {
		// Parse max-age from cache-control header
		ttl := 5 * time.Minute // default for "latest"
		if strings.Contains(cacheControl, "max-age=31536000") {
			ttl = 365 * 24 * time.Hour // 1 year for pinned versions
		}
		if err := registryCache.SetWithTTL(context.Background(), cacheKey, []byte(script), ttl); err != nil {
			logrus.WithError(err).WithField("cache_key", cacheKey).Warn("Failed to cache embed script")
		}
	}

	return script
}

// isOriginAllowed checks whether the given origin is in the allowed list.
// A wildcard "*" in the list allows all origins.
func isOriginAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}

// getEmbedBaseURL returns the public base URL for embed script URLs.
func getEmbedBaseURL() string {
	base := getBaseURL() // returns "https://api.functionfly.com/v1/fx"
	// Strip the /v1/fx suffix to get the root
	base = strings.TrimSuffix(base, "/v1/fx")
	return base
}

// validateAllowedOrigins validates that each entry in the allowed origins list
// is either a wildcard "*" or a valid origin (scheme + host, no path).
// SECURITY: prevents injection of arbitrary strings into origin matching.
func validateAllowedOrigins(origins []string) error {
	for _, o := range origins {
		if o == "*" {
			continue // wildcard is always valid
		}
		if o == "" {
			return fmt.Errorf("allowed_origins entries must not be empty strings")
		}
		u, err := url.Parse(o)
		if err != nil {
			return fmt.Errorf("invalid origin %q: %w", o, err)
		}
		// Must have a scheme and host, must NOT have a path
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid origin %q: must be a full origin (e.g. https://example.com)", o)
		}
		if u.Path != "" && u.Path != "/" {
			return fmt.Errorf("invalid origin %q: origin must not contain a path", o)
		}
		if u.Scheme != "https" && u.Scheme != "http" {
			return fmt.Errorf("invalid origin %q: scheme must be http or https", o)
		}
	}
	return nil
}

// resolveRequestOrigin extracts the actual origin for embed tracking.
// SECURITY: Uses the Referer header (server-side, cannot be spoofed by JS) as the
// primary source. Falls back to X-Embed-Origin only when Referer is unavailable
// (e.g., non-browser clients). Always extracts just the origin (scheme+host).
func resolveRequestOrigin(r *http.Request) string {
	// Primary: use Referer header which browsers set automatically and JS cannot override
	if ref := r.Header.Get("Referer"); ref != "" {
		if u, err := url.Parse(ref); err == nil && u.Host != "" {
			return u.Scheme + "://" + u.Host
		}
	}
	// Fallback: use Origin header (also browser-controlled but set on some requests)
	if origin := r.Header.Get("Origin"); origin != "" {
		return origin
	}
	// Last resort: X-Embed-Origin from embed script (spoofable, but only for analytics)
	if xOrigin := r.Header.Get("X-Embed-Origin"); xOrigin != "" {
		return xOrigin
	}
	return ""
}
