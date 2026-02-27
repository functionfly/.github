package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
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
		http.Error(w, "author and name are required", http.StatusBadRequest)
		return
	}

	// 1. Look up function in registry
	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// 2. Phase 3 — Allowed origins enforcement
	//    If the function has embed_config with a non-wildcard allowed_origins list,
	//    validate the requesting Origin header.
	embedCfg, cfgErr := h.repo.GetFunctionEmbedConfig(fn.ID)
	if cfgErr == nil && embedCfg != nil {
		if !embedCfg.Enabled {
			http.Error(w, "Embed not enabled for this function", http.StatusForbidden)
			return
		}
		origin := r.Header.Get("Origin")
		if origin != "" && !isOriginAllowed(origin, embedCfg.AllowedOrigins) {
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		// Phase 3 — Per-embed-domain rate limiting
		if origin != "" && embedCfg.RateLimitPerHour > 0 {
			since := time.Now().Add(-time.Hour)
			count, rlErr := h.repo.GetEmbedExecutionCountByOrigin(fn.ID, origin, since)
			if rlErr == nil && count >= int64(embedCfg.RateLimitPerHour) {
				w.Header().Set("Retry-After", "3600")
				http.Error(w, "Embed rate limit exceeded for this origin", http.StatusTooManyRequests)
				return
			}
		}
	}

	// 3. Get function version metadata
	latestVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
		return
	}

	// 4. Parse embed options from query string
	opts := parseEmbedOptions(
		r.URL.Query().Get("namespace"),
		r.URL.Query().Get("autoload"),
		r.URL.Query().Get("ui"),
		r.URL.Query().Get("theme"),
	)

	// 5. Generate embed script
	script := generateEmbedScript(fn, latestVersion, version, opts)

	// 6. Set headers
	cacheControl := "public, max-age=300" // 5 min for latest
	if version != "" {
		cacheControl = "public, max-age=31536000, immutable" // 1 year for pinned
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", cacheControl)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(script))
}

// HandleGetEmbedConfig returns the embed configuration for a function (JSON).
//
// Route: GET /v1/registry/functions/{author}/{name}/embed
func (h *Handler) HandleGetEmbedConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	cfg, err := h.repo.GetFunctionEmbedConfig(fn.ID)
	if err != nil {
		http.Error(w, "Failed to get embed config", http.StatusInternalServerError)
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
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	var cfg registry.EmbedConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate theme
	if cfg.UITheme != "" && cfg.UITheme != "light" && cfg.UITheme != "dark" && cfg.UITheme != "auto" {
		http.Error(w, "Invalid ui_theme: must be 'light', 'dark', or 'auto'", http.StatusBadRequest)
		return
	}

	// Default rate limit
	if cfg.RateLimitPerHour == 0 {
		cfg.RateLimitPerHour = 1000
	}

	// Default allowed origins
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"*"}
	}

	if err := h.repo.UpdateFunctionEmbedConfig(fn.ID, &cfg); err != nil {
		http.Error(w, "Failed to update embed config", http.StatusInternalServerError)
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
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	latestVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
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

	// Build example usage snippets
	basicSnippet := fmt.Sprintf(`<script src="%s"></script>
<script>
  const result = await %s.run({ /* your input */ });
  console.log(result.data);
</script>`, scriptURL, opts.Namespace)

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
		"author":        author,
		"name":          name,
		"version":       latestVersion.Version,
		"script_url":    scriptURL,
		"pinned_url":    pinnedURL,
		"namespace":     opts.Namespace,
		"snippets": map[string]string{
			"basic":  basicSnippet,
			"form":   formSnippet,
			"widget": widgetSnippet,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// HandleGetEmbedAnalytics returns embed execution analytics for a function.
//
// Route: GET /v1/registry/functions/{author}/{name}/embed/analytics
// Query parameters:
//
//	days  – number of days to look back (default: 30, max: 90)
//	limit – max number of origin domains to return (default: 20, max: 100)
func (h *Handler) HandleGetEmbedAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
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
		http.Error(w, "Failed to get embed analytics", http.StatusInternalServerError)
		return
	}

	// Phase 3 — Abuse detection: flag origins with unusually high counts
	var totalCount int64
	for _, s := range stats {
		totalCount += s.Count
	}

	type enrichedStat struct {
		Origin    string  `json:"origin"`
		Count     int64   `json:"count"`
		Pct       float64 `json:"pct"`
		Suspicious bool   `json:"suspicious"`
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

