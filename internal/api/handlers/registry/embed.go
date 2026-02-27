package registry

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// HandleServeEmbed serves a per-function embed script.
//
// Routes:
//
//	GET /v1/embed/{author}/{name}.js
//	GET /v1/embed/{author}/{name}@{version}.js
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

	// 2. Get function version metadata
	latestVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		http.Error(w, "No versions available", http.StatusNotFound)
		return
	}

	// 3. Parse embed options from query string
	opts := parseEmbedOptions(
		r.URL.Query().Get("namespace"),
		r.URL.Query().Get("autoload"),
		r.URL.Query().Get("ui"),
		r.URL.Query().Get("theme"),
	)

	// 4. Generate embed script
	script := generateEmbedScript(fn, latestVersion, version, opts)

	// 5. Set headers
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
