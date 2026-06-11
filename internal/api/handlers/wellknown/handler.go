// Package wellknown serves the FunctionFly AI discovery document at
// GET /.well-known/functionfly.json. It exposes the public registry to LLMs and
// agents via a standard, cacheable manifest with OpenAI-compatible tool schemas.
package wellknown

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/gateway"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// registryReader is the minimal registry interface needed for discovery (for testability).
type registryReader interface {
	SearchFunctions(query, category, runtime string, minRating float64, limit, offset int) ([]registry.RegistryFunction, int, error)
	GetLatestFunctionVersion(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error)
}

const (
	defaultLimit       = 50
	maxLimit           = 200
	maxQueryLen        = 256
	maxCategoryLen     = 128
	maxAuthorLen       = 128
	maxTagsParamLen    = 512
	maxDescriptionLen  = 4000 // OpenAI tool description guidance
	schemaVersion      = "1.0"
	provider           = "functionfly"
	cacheMaxAgeSeconds = 300
)

// OpenAI tool names: only [a-zA-Z0-9_]; replace other runes with underscore.
var toolNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

// FunctionFlyManifest is the response for GET /.well-known/functionfly.json.
// It is designed for LLM consumption and CDN caching.
type FunctionFlyManifest struct {
	SchemaVersion     string           `json:"schema_version"`
	Provider          string           `json:"provider"`
	ProviderURL       string           `json:"provider_url"`
	APIBase           string           `json:"api_base"`
	ExecutionEndpoint string           `json:"execution_endpoint"`
	AgentEndpoint     string           `json:"agent_endpoint"`
	DiscoveryEndpoint string           `json:"discovery_endpoint"`
	GeneratedAt       time.Time        `json:"generated_at"`
	TotalFunctions    int              `json:"total_functions"`
	Functions         []AICallableFunc `json:"functions"`
}

// AICallableFunc is a single function entry for AI tool discovery.
// ToolSchema is directly usable in OpenAI tools, Anthropic tools, LangChain, etc.
type AICallableFunc struct {
	URI               string          `json:"uri"`
	Name              string          `json:"name"`
	Title             string          `json:"title,omitempty"`
	Description       string          `json:"description,omitempty"`
	Version           string          `json:"version,omitempty"`
	Category          string          `json:"category,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	ExecutionURL      string          `json:"execution_url"`
	AgentExecutionURL string          `json:"agent_execution_url"`
	PricingPerCall    float64         `json:"pricing_per_call"`
	Deterministic     bool            `json:"deterministic"`
	SideEffects       string          `json:"side_effects"`
	TrustScore        float64         `json:"trust_score"`
	SuccessRate       float64         `json:"success_rate"`
	ToolSchema        json.RawMessage `json:"tool_schema,omitempty"`
}

// Handler serves the well-known FunctionFly AI discovery document.
type Handler struct {
	registryRepo registryReader
}

// NewHandler creates a new wellknown handler.
func NewHandler(registryRepo *registry.RegistryRepository) *Handler {
	return &Handler{registryRepo: registryRepo}
}

// HandleWellKnown serves GET /.well-known/functionfly.json.
// Query params (all optional): category, tags, author, q, limit (default 50, max 200), offset.
// Response is public, cacheable (Cache-Control: public, max-age=300), and CORS-enabled.
func (h *Handler) HandleWellKnown(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method == http.MethodOptions {
		setCORSHeaders(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET, OPTIONS")
		writeErrorJSON(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only GET and OPTIONS are allowed")
		return
	}

	q := r.URL.Query()
	category := trimAndCap(q.Get("category"), maxCategoryLen)
	tagsParam := trimAndCap(q.Get("tags"), maxTagsParamLen)
	author := trimAndCap(q.Get("author"), maxAuthorLen)
	searchQuery := trimAndCap(q.Get("q"), maxQueryLen)

	limit := defaultLimit
	if l := strings.TrimSpace(q.Get("limit")); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > maxLimit {
				limit = maxLimit
			} else {
				limit = v
			}
		}
	}
	offset := 0
	if o := strings.TrimSpace(q.Get("offset")); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	functions, total, err := h.registryRepo.SearchFunctions(searchQuery, category, "", 0, limit, offset)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"query": searchQuery, "category": category, "limit": limit, "offset": offset,
		}).Error("wellknown: search functions failed")
		writeErrorJSON(w, r, http.StatusInternalServerError, "SEARCH_FAILED", "failed to search functions")
		return
	}

	apiBase := getAPIBase(r)
	results := make([]AICallableFunc, 0, len(functions))
	skipped := 0

	for _, fn := range functions {
		if author != "" && fn.Author != author {
			continue
		}
		if tagsParam != "" && !fnHasTag(fn.Tags, tagsParam) {
			continue
		}
		if len(results) >= limit {
			break
		}

		fnVersion, err := h.registryRepo.GetLatestFunctionVersion(fn.ID)
		if err != nil {
			skipped++
			logrus.WithError(err).WithFields(logrus.Fields{
				"function_id": fn.ID, "author": fn.Author, "name": fn.Name,
			}).Debug("wellknown: skip function, no latest version")
			continue
		}

		title := nullStr(fn.Title)
		description := nullStr(fn.Description)
		cat := nullStr(fn.Category)
		version := nullStr(fn.LatestVersion)
		toolName := sanitizeToolName(fn.Author + "_" + fn.Name)
		execURL := apiBase + "/v1/fx/" + fn.Author + "/" + fn.Name
		agentExecURL := apiBase + "/v1/agent/execute/" + fn.Author + "/" + fn.Name

		sideEffects := fnVersion.SideEffects
		if sideEffects == "" {
			sideEffects = "none"
		}

		toolDesc := description
		if len(toolDesc) > maxDescriptionLen {
			toolDesc = toolDesc[:maxDescriptionLen-3] + "..."
		}
		toolSchema := buildOpenAIToolSchema(toolName, toolDesc, fnVersion.Manifest)

		var tags []string
		if fn.Tags != nil {
			_ = json.Unmarshal(fn.Tags, &tags)
		}

		results = append(results, AICallableFunc{
			URI:               "fx://" + fn.Author + "/" + fn.Name,
			Name:              toolName,
			Title:             title,
			Description:       description,
			Version:           version,
			Category:          cat,
			Tags:              tags,
			ExecutionURL:      execURL,
			AgentExecutionURL: agentExecURL,
			PricingPerCall:    fn.PricePerCall,
			Deterministic:     fnVersion.Deterministic,
			SideEffects:       sideEffects,
			TrustScore:        fn.ReliabilityScore,
			SuccessRate:       fn.ReliabilityScore,
			ToolSchema:        toolSchema,
		})
	}

	totalFunctions := total
	if author != "" || tagsParam != "" {
		totalFunctions = len(results)
	}

	manifest := FunctionFlyManifest{
		SchemaVersion:     schemaVersion,
		Provider:          provider,
		ProviderURL:       getProviderURL(),
		APIBase:           apiBase + "/v1",
		ExecutionEndpoint: "POST /v1/fx/{author}/{name}",
		AgentEndpoint:     "POST /v1/agent/execute/{author}/{name}",
		DiscoveryEndpoint: "GET /v1/agent/discover",
		GeneratedAt:       time.Now().UTC(),
		TotalFunctions:    totalFunctions,
		Functions:         results,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(cacheMaxAgeSeconds))
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Vary", "Accept-Encoding")
	setCORSHeaders(w, r)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(manifest); err != nil {
		logrus.WithError(err).Error("wellknown: encode response failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"path": r.URL.Path, "total": totalFunctions, "returned": len(results),
		"skipped": skipped, "duration_ms": time.Since(start).Milliseconds(),
	}).Debug("wellknown: request completed")
}

// setCORSHeaders writes the well-known CORS headers. Public read-only
// endpoint — GET/OPTIONS only. Thin wrapper over gateway.SetCORSHeaders
// (see P0 of the Two-Protocol Gateway plan). The middleware-level
// security headers are still applied via middleware.SetCORSHeaders
// before the gateway call so we don't lose Caddy/CDN semantics.
func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	middleware.SetCORSHeaders(w, r)
	gateway.SetCORSHeaders(w, r, gateway.CORSOptions{
		AllowMethods: "GET, OPTIONS",
	})
}

func writeErrorJSON(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	setCORSHeaders(w, r)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func getAPIBase(r *http.Request) string {
	if base := os.Getenv("BASE_URL"); base != "" {
		return strings.TrimSuffix(strings.TrimSpace(base), "/")
	}
	scheme := "https"
	if r.TLS == nil {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto == "http" {
			scheme = "http"
		} else if host := r.Host; host != "" && (strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1")) {
			scheme = "http"
		}
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	if host == "" {
		host = "api.functionfly.com"
	}
	return scheme + "://" + host
}

// getProviderURL returns the public/marketing site URL for the manifest (env-derived, staging-safe).
func getProviderURL() string {
	if u := os.Getenv("PUBLIC_SITE_URL"); u != "" {
		return strings.TrimSuffix(strings.TrimSpace(u), "/")
	}
	base := strings.TrimSpace(os.Getenv("BASE_URL"))
	if base != "" {
		if idx := strings.Index(base, "://"); idx != -1 {
			scheme := base[:idx+3]
			rest := base[idx+3:]
			if firstDot := strings.Index(rest, "."); firstDot != -1 {
				return scheme + rest[firstDot+1:]
			}
		}
	}
	return "https://functionfly.com"
}

func trimAndCap(s string, max int) string {
	s = strings.TrimSpace(s)
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}

func nullStr(n sql.NullString) string {
	if n.Valid {
		return n.String
	}
	return ""
}

// sanitizeToolName returns a name safe for OpenAI/Anthropic tools: [a-zA-Z0-9_].
func sanitizeToolName(name string) string {
	if name == "" {
		return "fn"
	}
	s := toolNameSanitizer.ReplaceAllString(name, "_")
	s = strings.Trim(s, "_")
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "fn"
	}
	if len(out) > 64 {
		return out[:64]
	}
	return out
}

func fnHasTag(tagsJSON json.RawMessage, tagsParam string) bool {
	if tagsJSON == nil {
		return false
	}
	var tags []string
	if err := json.Unmarshal(tagsJSON, &tags); err != nil {
		return false
	}
	requested := strings.Split(tagsParam, ",")
	for _, req := range requested {
		req = strings.TrimSpace(req)
		if req == "" {
			continue
		}
		for _, tag := range tags {
			if strings.EqualFold(tag, req) {
				return true
			}
		}
	}
	return false
}

// buildOpenAIToolSchema builds an OpenAI-compatible tool definition from the stored manifest.
// Parameters schema is derived from manifest input (Schema or Properties+Required).
func buildOpenAIToolSchema(name, description string, manifestRaw json.RawMessage) json.RawMessage {
	params := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
	if len(manifestRaw) > 0 {
		var manifest functionregistry.FunctionManifest
		if err := json.Unmarshal(manifestRaw, &manifest); err == nil && manifest.Input != nil {
			if manifest.Input.Schema != nil {
				var schema map[string]interface{}
				if err := json.Unmarshal(manifest.Input.Schema, &schema); err == nil {
					params = schema
				}
			} else if manifest.Input.Properties != nil {
				params["type"] = "object"
				if manifest.Input.Type != "" {
					params["type"] = manifest.Input.Type
				}
				var props map[string]interface{}
				if err := json.Unmarshal(manifest.Input.Properties, &props); err == nil {
					params["properties"] = props
				}
				if manifest.Input.Required.Array != nil && len(manifest.Input.Required.Array) > 0 {
					params["required"] = manifest.Input.Required.Array
				}
			}
		}
	}
	tool := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        name,
			"description": description,
			"parameters":  params,
		},
	}
	out, err := json.Marshal(tool)
	if err != nil {
		out, _ = json.Marshal(map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        name,
				"description": description,
				"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
			},
		})
	}
	return out
}
