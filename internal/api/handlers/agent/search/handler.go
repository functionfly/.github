package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	agenttools "github.com/functionfly/functionfly/internal/agent/tools"
	"github.com/functionfly/functionfly/internal/agent/tools/search"
	"github.com/functionfly/functionfly/internal/agent/tools/search/providers"
	searchrepo "github.com/functionfly/functionfly/internal/storage/search"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SearchHandler handles search tool HTTP requests
type SearchHandler struct {
	toolRegistry  *search.Registry
	execRepo      *searchrepo.ExecutionRepository
	cacheRepo     *searchrepo.CacheRepository
	quotaEnforcer agenttools.QuotaEnforcer
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(
	db *gorm.DB,
	redisClient agenttools.RedisClient,
	quotaEnforcer agenttools.QuotaEnforcer,
	cacheTTLSeconds int,
) *SearchHandler {
	execRepo := searchrepo.NewExecutionRepository(db)
	cacheRepo := searchrepo.NewCacheRepository(db, cacheTTLSeconds)

	cfg := search.LoadSearchConfig()

	var provider search.SearchProvider
	var reader *search.JinaReader

	if cfg.Provider == "mock" && cfg.SearXNGURL == "" && cfg.BraveAPIKey == "" && cfg.SERPAPIKey == "" {
		provider = providers.NewMockProvider()
	} else {
		cascade, err := providers.NewCascadeProviderFromConfig(cfg)
		if err != nil {
			logrus.WithError(err).Warn("failed to create cascade provider, falling back to mock")
			provider = providers.NewMockProvider()
		} else {
			provider = cascade
		}
		if cfg.JinaAPIKey != "" || cfg.AutoReadTop > 0 {
			reader = search.NewJinaReader(cfg.JinaAPIKey)
		}
	}

	if reader != nil {
		if err := search.InitializeWithReader(provider, reader); err != nil {
			logrus.WithError(err).Error("failed to initialize search tools")
		}
	} else {
		if err := search.Initialize(provider); err != nil {
			logrus.WithError(err).Error("failed to initialize search tools")
		}
	}

	return &SearchHandler{
		toolRegistry:  search.GetRegistry(),
		execRepo:       execRepo,
		cacheRepo:     cacheRepo,
		quotaEnforcer: quotaEnforcer,
	}
}

// SetProvider sets the search provider (for testing or custom providers)
func (h *SearchHandler) SetProvider(provider search.SearchProvider) {
	if err := search.Initialize(provider); err != nil {
		logrus.WithError(err).Error("failed to set search provider")
	}
}

// HandleListTools returns all available search tools
// GET /v1/agent/tools/search
func (h *SearchHandler) HandleListTools(w http.ResponseWriter, r *http.Request) {
	definitions := h.toolRegistry.ListDefinitions()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"tools": definitions,
		"count": len(definitions),
	})
}

// HandleGetToolSchema returns the schema for a specific tool
// GET /v1/agent/tools/search/{tool_name}/schema
func (h *SearchHandler) HandleGetToolSchema(w http.ResponseWriter, r *http.Request) {
	toolName := mux.Vars(r)["tool_name"]

	schema, err := h.toolRegistry.GetSchema(toolName)
	if err != nil {
		writeError(w, http.StatusNotFound, "TOOL_NOT_FOUND", fmt.Sprintf("tool %s not found", toolName))
		return
	}

	baseCost, perResultCost, err := h.toolRegistry.GetCostInfo(toolName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "COST_INFO_FAILED", "failed to get cost info")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"tool_name":      toolName,
		"schema":         schema,
		"base_cost":      baseCost,
		"cost_per_result": perResultCost,
	})
}

// HandleExecuteTool executes a search tool
// POST /v1/agent/tools/search/execute
func (h *SearchHandler) HandleExecuteTool(w http.ResponseWriter, r *http.Request) {
	// Extract agent context from headers
	agentIDStr := r.Header.Get("X-Agent-ID")
	sessionID := r.Header.Get("X-Agent-Session-ID")

	var agentID *uuid.UUID
	if agentIDStr != "" {
		id, err := uuid.Parse(agentIDStr)
		if err == nil {
			agentID = &id
		}
	}

	// Parse request
	var req ExecuteToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.ToolName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOOL_NAME", "tool_name is required")
		return
	}

	// Check if tool exists
	tool, err := h.toolRegistry.Get(req.ToolName)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusNotFound, "TOOL_NOT_FOUND", "get search tool", err)
		return
	}

	// Validate parameters
	if err := tool.Validate(req.Parameters); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_PARAMETERS", "validate search parameters", err)
		return
	}

	// Check quota if agent is specified
	if agentID != nil && h.quotaEnforcer != nil {
		allowed, reason, err := h.quotaEnforcer.CheckQuota(r.Context(), agentID.String(), req.ToolName)
		if err != nil {
			logrus.WithError(err).Warn("quota check failed, proceeding anyway")
		} else if !allowed {
			writeError(w, http.StatusTooManyRequests, "QUOTA_EXCEEDED", reason)
			return
		}
	}

	// Check cache for GET requests (cache enabled by default for search)
	cacheKey := search.GenerateCacheKey(req.ToolName, getQueryFromParams(req.Parameters), marshalParams(req.Parameters))
	if req.EnableCache {
		cached, err := h.cacheRepo.Get(r.Context(), cacheKey)
		if err == nil && cached != nil {
			// Cache hit - return cached result
			result, _ := searchrepo.ValidateCacheResult[interface{}](cached)
			if result != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":       true,
					"cached":   true,
					"result":   result,
					"cachedAt": cached.CachedAt,
				})
				return
			}
		}
	}

	// Create execution context
	startTime := time.Now()
	execCtx := &agenttools.ExecutionContext{
		AgentID:   agentIDStr,
		SessionID: sessionID,
		CallDepth: 0,
	}

	// Execute tool
	toolResult, err := tool.Execute(r.Context(), req.Parameters, execCtx.AgentID, "", execCtx.SessionID, execCtx.CallDepth)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "EXECUTION_FAILED", "execute search tool", err)
		return
	}

	// Extract result data
	result := toolResult.Data

	// Calculate execution time and cost
	executionTime := time.Since(startTime)
	resultCount := getResultCount(result)
	cost := toolResult.CostUSD
	if cost == 0 {
		cost = tool.Cost(req.Parameters, resultCount)
	}

	// Log execution
	execution := &searchrepo.Execution{
		ToolName:        req.ToolName,
		Query:           getQueryFromParams(req.Parameters),
		Parameters:      marshalParams(req.Parameters),
		ResultsCount:    resultCount,
		CreditsUsed:     cost,
		ExecutionTimeMs: int(executionTime.Milliseconds()),
		AgentID:         agentID,
	}

	if err := h.execRepo.Create(r.Context(), execution); err != nil {
		logrus.WithError(err).Warn("failed to log search execution")
	}

	// Update quota usage
	if agentID != nil && h.quotaEnforcer != nil {
		if err := h.quotaEnforcer.RecordUsage(r.Context(), agentID.String(), req.ToolName, cost); err != nil {
			logrus.WithError(err).Warn("failed to record quota usage")
		}
	}

	// Cache result if enabled
	if req.EnableCache {
		if err := h.cacheRepo.Set(r.Context(), &searchrepo.CacheEntry{
			CacheKey:   cacheKey,
			ToolName:   req.ToolName,
			QueryHash:  uuid.NewSHA1(uuid.NameSpaceDNS, []byte(getQueryFromParams(req.Parameters))).String(),
			Parameters: marshalParams(req.Parameters),
			Results:    mustMarshal(result),
			ExpiresAt:  time.Now().Add(time.Hour),
		}); err != nil {
			logrus.WithError(err).Warn("failed to cache search result")
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"cached":           false,
		"result":           result,
		"executionTimeMs":  executionTime.Milliseconds(),
		"creditsUsed":      cost,
		"resultsCount":     resultCount,
	})
}

// HandleGetExecutionStats returns execution statistics for search tools
// GET /v1/agent/tools/search/stats
func (h *SearchHandler) HandleGetExecutionStats(w http.ResponseWriter, r *http.Request) {
	toolName := r.URL.Query().Get("tool_name")
	if toolName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOOL_NAME", "tool_name query parameter is required")
		return
	}

	// Default: last 7 days
	since := time.Now().UTC().AddDate(0, 0, -7)
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			since = t
		}
	}

	stats, err := h.execRepo.GetStats(r.Context(), toolName, since)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "STATS_FAILED", "get search stats", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"toolName":    toolName,
		"since":       since,
		"stats":       stats,
	})
}

// HandleListExecutions lists search executions for an agent
// GET /v1/agent/tools/search/executions
func (h *SearchHandler) HandleListExecutions(w http.ResponseWriter, r *http.Request) {
	agentIDStr := r.Header.Get("X-Agent-ID")
	if agentIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_AGENT_ID", "X-Agent-ID header is required")
		return
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_AGENT_ID", "invalid agent ID format")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	executions, err := h.execRepo.ListByAgent(r.Context(), agentID, limit, offset)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "LIST_FAILED", "list search executions", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"executions": executions,
		"limit":      limit,
		"offset":     offset,
	})
}

// Helper functions

func getQueryFromParams(params map[string]interface{}) string {
	if q, ok := params["query"].(string); ok {
		return q
	}
	if q, ok := params["Query"].(string); ok {
		return q
	}
	return ""
}

func marshalParams(params map[string]interface{}) json.RawMessage {
	if params == nil {
		return nil
	}
	b, err := json.Marshal(params)
	if err != nil {
		return nil
	}
	return b
}

func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func getResultCount(result interface{}) int {
	if result == nil {
		return 0
	}

	// Try to extract result count from common response types
	switch r := result.(type) {
	case *search.WebSearchResponse:
		if r != nil {
			return len(r.Results)
		}
	case *search.NewsSearchResponse:
		if r != nil {
			return len(r.Articles)
		}
	case *search.DocsSearchResponse:
		if r != nil {
			return len(r.Documents)
		}
	case *search.CompanySearchResponse:
		if r != nil {
			count := 0
			if r.Company != nil {
				count++
			}
			count += len(r.News) + len(r.Technologies)
			if r.Funding != nil {
				count += len(r.Funding.Rounds)
			}
			return count
		}
	case map[string]interface{}:
		if results, ok := r["results"].([]interface{}); ok {
			return len(results)
		}
		if articles, ok := r["articles"].([]interface{}); ok {
			return len(articles)
		}
		if documents, ok := r["documents"].([]interface{}); ok {
			return len(documents)
		}
	}

	return 0
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// writeError is a helper to write error responses
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}

// writeErrorFromErr logs err server-side with context and writes a generic
// client-visible message. Use in place of writeError(w, status, code, err.Error()).
func writeErrorFromErr(r *http.Request, w http.ResponseWriter, status int, code, contextMsg string, err error) {
	if err != nil {
		fields := logrus.Fields{
			"status":  status,
			"code":    code,
			"context": contextMsg,
			"method":  "",
			"path":    "",
		}
		if r != nil {
			fields["method"] = r.Method
			if r.URL != nil {
				fields["path"] = r.URL.Path
			}
		}
		entry := logrus.WithError(err).WithFields(fields)
		if status >= 500 {
			entry.Error("search handler error")
		} else {
			entry.Info("search handler client error")
		}
	}
	message := sanitizedSearchErrorMessage(status, code, contextMsg)
	writeError(w, status, code, message)
}

func sanitizedSearchErrorMessage(status int, code, contextMsg string) string {
	if status >= 500 {
		switch code {
		case "EXECUTION_FAILED":
			return "Search execution failed. Check server logs for details."
		case "STATS_FAILED":
			return "Failed to retrieve search statistics."
		case "LIST_FAILED":
			return "Failed to list search tools."
		}
		return "Internal server error"
	}
	switch code {
	case "TOOL_NOT_FOUND":
		return "Search tool not found"
	case "INVALID_PARAMETERS":
		return "Invalid search parameters"
	}
	return contextMsg
}