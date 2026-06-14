package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// DiscoveryResult is the agent-optimized function discovery response
type DiscoveryResult struct {
	URI            string          `json:"uri"`           // fx://org/name
	Author         string          `json:"author"`
	Name           string          `json:"name"`
	Version        string          `json:"version,omitempty"`
	Title          string          `json:"title,omitempty"`
	Description    string          `json:"description,omitempty"`
	Schema         json.RawMessage `json:"schema,omitempty"`
	PricingPerCall float64         `json:"pricing_per_call"`
	Deterministic  bool            `json:"deterministic"`
	TrustScore     float64         `json:"trust_score"`
	SuccessRate    float64         `json:"success_rate"`
	Tags           json.RawMessage `json:"tags,omitempty"`
	Capabilities   json.RawMessage `json:"capabilities,omitempty"`
	SideEffects    string          `json:"side_effects"`
	Category       string          `json:"category,omitempty"`
}

// HandleDiscover handles function discovery for agents
// GET /v1/agent/discover
func (h *Handler) HandleDiscover(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parse query parameters
	deterministicOnly := q.Get("deterministic") == "true"
	author := q.Get("author")
	trustScoreMinStr := q.Get("trust_score_min")
	category := q.Get("category")
	searchQuery := q.Get("q")
	tagsParam := q.Get("tags")

	limit := 20
	offset := 0
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	var trustScoreMin float64
	if trustScoreMinStr != "" {
		if v, err := strconv.ParseFloat(trustScoreMinStr, 64); err == nil {
			trustScoreMin = v
		}
	}

	// Use the registry's search function
	// SearchFunctions(query string, category, runtime string, minRating float64, limit, offset int)
	functions, total, err := h.registryRepo.SearchFunctions(searchQuery, category, "", 0, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("discovery search failed")
		writeError(w, http.StatusInternalServerError, "DISCOVERY_FAILED", "failed to search functions")
		return
	}

	// Filter and transform results
	results := make([]DiscoveryResult, 0, len(functions))
	for _, fn := range functions {
		// Apply filters
		if deterministicOnly && fn.DeterministicScore < 0.9 {
			continue
		}
		if author != "" && fn.Author != author {
			continue
		}
		if trustScoreMin > 0 && fn.ReliabilityScore < trustScoreMin {
			continue
		}
		if tagsParam != "" && !fnHasTag(fn.Tags, tagsParam) {
			continue
		}

		title := ""
		if fn.Title.Valid {
			title = fn.Title.String
		}
		description := ""
		if fn.Description.Valid {
			description = fn.Description.String
		}
		cat := ""
		if fn.Category.Valid {
			cat = fn.Category.String
		}

		result := DiscoveryResult{
			URI:            "fx://" + fn.Author + "/" + fn.Name,
			Author:         fn.Author,
			Name:           fn.Name,
			Title:          title,
			Description:    description,
			PricingPerCall: fn.PricePerCall,
			Deterministic:  fn.DeterministicScore >= 0.9,
			TrustScore:     fn.ReliabilityScore,
			SuccessRate:    fn.ReliabilityScore,
			Tags:           fn.Tags,
			Capabilities:   fn.Capabilities,
			Category:       cat,
		}

		results = append(results, result)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"results": results,
		"total":   total,
		"page":    offset/limit + 1,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleDiscoverFunction returns detailed function info for agents
// GET /v1/agent/discover/{author}/{name}
func (h *Handler) HandleDiscoverFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := h.registryRepo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "function not found")
		return
	}

	fnVersion, err := h.registryRepo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no versions available")
		return
	}

	title := ""
	if fn.Title.Valid {
		title = fn.Title.String
	}
	description := ""
	if fn.Description.Valid {
		description = fn.Description.String
	}
	cat := ""
	if fn.Category.Valid {
		cat = fn.Category.String
	}

	result := DiscoveryResult{
		URI:            "fx://" + fn.Author + "/" + fn.Name,
		Author:         fn.Author,
		Name:           fn.Name,
		Version:        fnVersion.Version,
		Title:          title,
		Description:    description,
		PricingPerCall: fn.PricePerCall,
		Deterministic:  fnVersion.Deterministic,
		SideEffects:    fnVersion.SideEffects,
		TrustScore:     fn.ReliabilityScore,
		SuccessRate:    fn.ReliabilityScore,
		Tags:           fn.Tags,
		Capabilities:   fnVersion.Capabilities,
		Category:       cat,
		Schema:         fnVersion.Manifest,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"function": result,
	})
}

// fnHasTag checks if a function's tags JSON contains any of the requested tags
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
		for _, tag := range tags {
			if strings.EqualFold(tag, req) {
				return true
			}
		}
	}
	return false
}
