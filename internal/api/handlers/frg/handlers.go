// Package frg provides HTTP handlers for the Function Registry + Live Runtime Graph API
package frg

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/frg"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// AICompositionClient calls the AI service for graph composition
type AICompositionClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewAICompositionClient creates a new AI composition client
func NewAICompositionClient() *AICompositionClient {
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		logrus.Warn("AI_SERVICE_URL not set for AICompositionClient")
		baseURL = "" // Must be set explicitly
	}

	return &AICompositionClient{
		baseURL: baseURL,
		apiKey:  os.Getenv("AI_SERVICE_API_KEY"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// EmbeddingServiceClient calls the AI service for generating embeddings
type EmbeddingServiceClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewEmbeddingServiceClient creates a new embedding service client
func NewEmbeddingServiceClient() *EmbeddingServiceClient {
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		logrus.Warn("AI_SERVICE_URL not set for EmbeddingServiceClient")
		baseURL = "" // Must be set explicitly
	}

	return &EmbeddingServiceClient{
		baseURL: baseURL,
		apiKey:  os.Getenv("AI_SERVICE_API_KEY"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// embeddingRequest is the request to generate an embedding
type embeddingRequest struct {
	Text string `json:"text"`
}

// embeddingResponse is the response with the generated embedding
type embeddingResponse struct {
	Embedding  []float64 `json:"embedding"`
	Model      string    `json:"model"`
	Dimensions int       `json:"dimensions"`
}

// GenerateEmbedding generates an embedding for the given text
func (c *EmbeddingServiceClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	url := fmt.Sprintf("%s/api/embed", c.baseURL)

	reqBody := embeddingRequest{Text: text}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var embedResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(embedResp.Embedding))
	for i, v := range embedResp.Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// CompositionRequest is the request to AI service for graph composition
type CompositionRequest struct {
	Prompt           string   `json:"prompt"`
	Requirements     []string `json:"requirements,omitempty"`
	PreferredRuntime string   `json:"preferred_runtime"`
	// TenantID intentionally omitted - do not leak tenant identifiers to external AI services
}

// CompositionResponse is the response from AI service
type CompositionResponse struct {
	Success      bool                    `json:"success"`
	Graph        *AIComposedGraph        `json:"graph,omitempty"`
	Explanation  *CompositionExplanation `json:"explanation,omitempty"`
	Confidence   float64                 `json:"confidence"`
	GenerationID string                  `json:"generation_id"`
	LatencyMs    float64                 `json:"latency_ms"`
	Error        string                  `json:"error,omitempty"`
	Suggestions  []string                `json:"suggestions,omitempty"`
}

// AIComposedGraph is the graph structure from AI service
type AIComposedGraph struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	ExecutionMode string                 `json:"execution_mode"`
	Nodes         []GraphNodeRef         `json:"nodes"`
	Edges         []GraphEdge            `json:"edges"`
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema  map[string]interface{} `json:"output_schema,omitempty"`
	TriggerConfig *TriggerConfig         `json:"trigger_config,omitempty"`
	Visibility    string                 `json:"visibility"`
}

// GraphNodeRef is a node reference from AI
type GraphNodeRef struct {
	NodeID      string                 `json:"node_id"`
	Author      string                 `json:"author"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Config      map[string]interface{} `json:"config"`
	Description string                 `json:"description,omitempty"`
}

// GraphEdge is an edge from AI
type GraphEdge struct {
	ID             string `json:"id"`
	SourceNodeID   string `json:"source_node_id"`
	TargetNodeID   string `json:"target_node_id"`
	Type           string `json:"type"`
	FallbackNodeID string `json:"fallback_node_id,omitempty"`
}

// TriggerConfig is trigger configuration from AI
type TriggerConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// CompositionExplanation explains the composed graph
type CompositionExplanation struct {
	Summary             string            `json:"summary"`
	NodePurposes        map[string]string `json:"node_purposes"`
	DataFlowDescription string            `json:"data_flow_description"`
	TriggerExplanation  string            `json:"trigger_explanation"`
	SuggestedTests      []string          `json:"suggested_tests"`
}

// Compose sends a composition request to the AI service
func (c *AICompositionClient) Compose(ctx context.Context, req CompositionRequest) (*CompositionResponse, error) {
	url := fmt.Sprintf("%s/api/composition/compose", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result CompositionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// Handler contains dependencies for FRG handlers
type Handler struct {
	// Repositories
	frgRepo      *frg.Repository
	registryRepo *registry.RegistryRepository
	functionRepo *storage.FunctionRepository

	// Services
	engine       *frg.ExecutionEngine
	graphService *graph.Service

	// Caching
	cacheService *cache.CacheService

	// AI Service
	aiClient *AICompositionClient

	// Embedding Service for semantic search
	embedClient *EmbeddingServiceClient

	// UsageTracker provides real-time quota enforcement and usage tracking
	UsageTracker services.RealtimeUsageTrackerInterface
}

// NewHandler creates a new FRG handler
func NewHandler(
	frgRepo *frg.Repository,
	registryRepo *registry.RegistryRepository,
	functionRepo *storage.FunctionRepository,
	engine *frg.ExecutionEngine,
	graphService *graph.Service,
	cacheService *cache.CacheService,
	aiClient *AICompositionClient,
	embedClient *EmbeddingServiceClient,
	usageTracker services.RealtimeUsageTrackerInterface,
) *Handler {
	if aiClient == nil {
		aiClient = NewAICompositionClient()
	}
	if embedClient == nil {
		embedClient = NewEmbeddingServiceClient()
	}
	return &Handler{
		frgRepo:      frgRepo,
		registryRepo: registryRepo,
		functionRepo: functionRepo,
		engine:       engine,
		graphService: graphService,
		cacheService: cacheService,
		aiClient:     aiClient,
		embedClient:  embedClient,
		UsageTracker: usageTracker,
	}
}

// ==================== Graph Definition Handlers ====================

// CreateGraphRequest represents a request to create a new graph
type CreateGraphRequest struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	ExecutionMode string                 `json:"execution_mode,omitempty"` // sync, async, streaming, event_driven
	Nodes         []frg.GraphNodeRef     `json:"nodes"`
	Edges         []frg.GraphEdge        `json:"edges"`
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema  map[string]interface{} `json:"output_schema,omitempty"`
	TriggerConfig *frg.TriggerConfig     `json:"trigger_config,omitempty"`
	Visibility    string                 `json:"visibility,omitempty"`
	BasePrice     float64                `json:"base_price,omitempty"`
}

// GraphResponse represents a graph definition response
type GraphResponse struct {
	frg.GraphDefinition
	FullName string `json:"full_name"`
}

// CreateGraph creates a new graph definition
func (h *Handler) CreateGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Input size limits to prevent memory exhaustion
	const (
		maxNodes     = 100
		maxEdges     = 500
		maxGraphJSON = 1_000_000 // ~1MB
	)
	if len(req.Nodes) > maxNodes {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Graph cannot have more than %d nodes", maxNodes))
		return
	}
	if len(req.Edges) > maxEdges {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Graph cannot have more than %d edges", maxEdges))
		return
	}

	// Validate required fields
	if req.Name == "" || len(req.Nodes) == 0 {
		respondError(w, http.StatusBadRequest, "Name and nodes are required")
		return
	}

	// Validate name format
	if len(req.Name) > 100 {
		respondError(w, http.StatusBadRequest, "Graph name too long (max 100 characters)")
		return
	}

	// Validate all function refs exist
	for _, node := range req.Nodes {
		_, err := h.registryRepo.GetFunctionByAuthorName(context.Background(), node.Author, node.Name)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Function not found: "+node.Author+"/"+node.Name)
			return
		}
	}

	// Detect cycles using graph service
	for _, edge := range req.Edges {
		// Simple cycle check: ensure source and target are valid
		sourceValid, targetValid := false, false
		for _, node := range req.Nodes {
			if node.NodeID == edge.SourceNodeID {
				sourceValid = true
			}
			if node.NodeID == edge.TargetNodeID {
				targetValid = true
			}
		}
		if !sourceValid || !targetValid {
			respondError(w, http.StatusBadRequest, "Edge references invalid node")
			return
		}
	}

	// Build graph definition
	nodeRefsJSON, _ := json.Marshal(req.Nodes)
	edgesJSON, _ := json.Marshal(req.Edges)
	inputSchemaJSON, _ := json.Marshal(req.InputSchema)
	outputSchemaJSON, _ := json.Marshal(req.OutputSchema)
	triggerConfigJSON, _ := json.Marshal(req.TriggerConfig)

	if req.ExecutionMode == "" {
		req.ExecutionMode = "sync"
	}
	if req.Visibility == "" {
		req.Visibility = "public"
	}

	def := &frg.GraphDefinition{
		Author:        user.Username,
		Name:          req.Name,
		Version:       "v1",
		NodeRefs:      nodeRefsJSON,
		Edges:         edgesJSON,
		ExecutionMode: frg.ExecutionMode(req.ExecutionMode),
		InputSchema:   inputSchemaJSON,
		OutputSchema:  outputSchemaJSON,
		TriggerConfig: triggerConfigJSON,
		OwnerUserID:   &user.UserID,
		Visibility:    req.Visibility,
		BasePrice:     req.BasePrice,
	}

	created, err := h.frgRepo.CreateDefinition(ctx, def)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, GraphResponse{
		GraphDefinition: *created,
		FullName:        created.Author + "/" + created.Name + "@" + created.Version,
	})
}

// GetGraph retrieves a graph definition
func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "v1"
	}

	def, err := h.frgRepo.GetDefinitionByName(ctx, author, name, version)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GraphResponse{
		GraphDefinition: *def,
		FullName:        def.Author + "/" + def.Name + "@" + def.Version,
	})
}

// ListGraphs lists graph definitions
func (h *Handler) ListGraphs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := &frg.DefinitionFilter{
		Author:        r.URL.Query().Get("author"),
		Visibility:    r.URL.Query().Get("visibility"),
		ExecutionMode: r.URL.Query().Get("execution_mode"),
		Limit:         20,
	}

	defs, err := h.frgRepo.ListDefinitions(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var responses []GraphResponse
	for _, def := range defs {
		responses = append(responses, GraphResponse{
			GraphDefinition: *def,
			FullName:        def.Author + "/" + def.Name + "@" + def.Version,
		})
	}

	respondJSON(w, http.StatusOK, responses)
}

// UpdateGraph updates a graph definition (only if not published)
func (h *Handler) UpdateGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get existing graph
	def, err := h.frgRepo.GetDefinitionByName(ctx, author, name, "v1")
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Check ownership
	if def.OwnerUserID == nil || *def.OwnerUserID != user.UserID {
		respondError(w, http.StatusForbidden, "Not authorized to modify this graph")
		return
	}

	// Parse updates
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates (only allowed fields before publish)
	if nodes, ok := updates["nodes"].([]interface{}); ok {
		nodeRefsJSON, _ := json.Marshal(nodes)
		def.NodeRefs = nodeRefsJSON
	}
	if edges, ok := updates["edges"].([]interface{}); ok {
		edgesJSON, _ := json.Marshal(edges)
		def.Edges = edgesJSON
	}

	if err := h.frgRepo.UpdateDefinition(ctx, def); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GraphResponse{
		GraphDefinition: *def,
		FullName:        def.Author + "/" + def.Name + "@" + def.Version,
	})
}

// PublishGraphVersion publishes a graph version
func (h *Handler) PublishGraphVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get existing graph
	def, err := h.frgRepo.GetDefinitionByName(ctx, author, name, "v1")
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Check ownership
	if def.OwnerUserID == nil || *def.OwnerUserID != user.UserID {
		respondError(w, http.StatusForbidden, "Not authorized")
		return
	}

	// Publish
	if err := h.frgRepo.PublishVersion(ctx, def.ID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "published",
	})
}

// DeleteGraph deletes a graph (only if not published)
func (h *Handler) DeleteGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get existing graph
	def, err := h.frgRepo.GetDefinitionByName(ctx, author, name, "v1")
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Check ownership
	if def.OwnerUserID == nil || *def.OwnerUserID != user.UserID {
		respondError(w, http.StatusForbidden, "Not authorized")
		return
	}

	// Delete
	if err := h.frgRepo.DeleteDefinition(ctx, def.ID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ==================== Graph Execution Handlers ====================

// ExecuteGraphRequest represents a graph execution request
type ExecuteGraphRequest struct {
	Input map[string]interface{} `json:"input"`
}

// ExecuteGraphResponse represents a graph execution response
type ExecuteGraphResponse struct {
	InstanceID  string                 `json:"instance_id"`
	Status      string                 `json:"status"`
	Output      interface{}            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	DurationMs  int                    `json:"duration_ms,omitempty"`
	NodeResults map[string]interface{} `json:"node_results,omitempty"`
}

// ExecuteGraph executes a graph synchronously or asynchronously
func (h *Handler) ExecuteGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "latest"
	}

	// Get graph definition
	var def *frg.GraphDefinition
	var err error
	if version == "latest" {
		def, err = h.frgRepo.GetLatestVersion(ctx, author, name)
	} else {
		def, err = h.frgRepo.GetDefinitionByName(ctx, author, name, version)
	}
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	// Tenant isolation: private graphs can only be executed by their owners
	if def.Visibility == "private" {
		// Check ownership: either user owns the graph directly, or their tenant owns it
		if def.OwnerUserID != nil && *def.OwnerUserID != user.UserID {
			// Check tenant ownership
			if def.TenantID != nil && *def.TenantID != user.TenantID {
				respondError(w, http.StatusForbidden, "Not authorized to execute this graph")
				return
			}
		}
	}

	// Real-time quota enforcement
	if h.UsageTracker != nil && h.UsageTracker.IsEnabled() && user.TenantID != uuid.Nil {
		quotaResult, err := h.UsageTracker.RecordExecution(ctx, user.TenantID, "")
		if err != nil {
			logrus.WithError(err).WithField("tenant_id", user.TenantID).Warn("Quota check failed, allowing execution")
		} else if !quotaResult.Allowed {
			logrus.WithFields(logrus.Fields{
				"tenant_id": user.TenantID,
				"reason":    quotaResult.Reason,
			}).Warn("Quota exceeded, blocking FRG graph execution")
			monitoring.RecordFRGQuotaExceeded(user.TenantID.String())
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "QUOTA_EXCEEDED",
					"message": quotaResult.Reason,
					"type":    "quota_exceeded",
				},
				"quota_status": quotaResult.Status,
				"upgrade_url": "/settings/billing",
			})
			return
		}

		// Add quota headers and record metrics
		if quotaResult.Status != nil {
			w.Header().Set("X-Quota-Executions-Used", fmt.Sprintf("%d", quotaResult.Status.ExecutionsUsed))
			w.Header().Set("X-Quota-Executions-Limit", fmt.Sprintf("%d", quotaResult.Status.ExecutionsLimit))
			w.Header().Set("X-Quota-Executions-Percent", fmt.Sprintf("%.1f", quotaResult.Status.ExecutionsPercent))
			w.Header().Set("X-Quota-Status", quotaResult.Status.Status)
			monitoring.RecordFRGQuotaUsagePercent(user.TenantID.String(), quotaResult.Status.ExecutionsPercent)
		}
	}

	// Parse input
	var req ExecuteGraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Input = make(map[string]interface{})
	}

	// Audit logging for graph execution
	executionStart := time.Now()
	graphID := def.ID.String()

	// Default execution timeout (5 minutes for sync execution)
	const defaultExecutionTimeout = 5 * time.Minute
	execTimeout := defaultExecutionTimeout
	if timeoutStr := r.URL.Query().Get("timeout"); timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			// Cap timeout at maximum to prevent abuse
			if parsedTimeout < defaultExecutionTimeout {
				execTimeout = parsedTimeout
			}
		}
	}
	execCtx, execCancel := context.WithTimeout(ctx, execTimeout)
	defer execCancel()

	// Execute based on mode
	switch def.ExecutionMode {
	case frg.ExecutionModeSync:
		result, err := h.engine.ExecuteSync(execCtx, def, req.Input)
		if err != nil {
			if execCtx.Err() == context.DeadlineExceeded {
				monitoring.RecordFRGGraphExecution(user.TenantID.String(), graphID, "execute", "timeout")
				respondError(w, http.StatusGatewayTimeout, "Graph execution timed out")
				return
			}
			monitoring.RecordFRGGraphExecution(user.TenantID.String(), graphID, "execute", "error")
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var output interface{}
		if result.Output != nil {
			json.Unmarshal(result.Output, &output)
		}

		monitoring.RecordFRGGraphExecution(user.TenantID.String(), graphID, "execute", "success")
		monitoring.RecordFRGGraphExecutionDuration(user.TenantID.String(), graphID, time.Since(executionStart))

		respondJSON(w, http.StatusOK, ExecuteGraphResponse{
			InstanceID: result.InstanceID.String(),
			Status:     string(result.Status),
			Output:     output,
			DurationMs: result.DurationMs,
		})

	case frg.ExecutionModeAsync, frg.ExecutionModeStreaming, frg.ExecutionModeEventDriven:
		// Async execution - return instance ID immediately
		instance, err := h.engine.ExecuteAsync(execCtx, def, req.Input)
		if err != nil {
			monitoring.RecordFRGGraphExecution(user.TenantID.String(), graphID, "execute_async", "error")
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		monitoring.RecordFRGGraphExecution(user.TenantID.String(), graphID, "execute_async", "started")

		respondJSON(w, http.StatusAccepted, ExecuteGraphResponse{
			InstanceID: instance.ID.String(),
			Status:     string(instance.Status),
		})
	}
}

// GetInstanceStatus returns the status of a graph instance
func (h *Handler) GetInstanceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]

	// Parse UUID
	id, err := parseUUID(instanceID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}

	instance, err := h.frgRepo.GetInstanceByID(ctx, id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, instance)
}

// ListInstances lists graph instances
func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get graph to find definition ID
	def, err := h.frgRepo.GetLatestVersion(ctx, author, name)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	filter := &frg.InstanceFilter{
		DefinitionID: &def.ID,
		Status:       r.URL.Query().Get("status"),
		Limit:        20,
	}

	instances, err := h.frgRepo.ListInstances(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, instances)
}

// StopInstance stops a running graph instance
func (h *Handler) StopInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	instanceID := vars["instance_id"]

	// Parse UUID
	id, err := parseUUID(instanceID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}

	// Stop the instance
	if err := h.engine.StopInstance(id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update database
	if err := h.frgRepo.UpdateInstanceStatus(ctx, id, frg.InstanceStatusCompleted); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "stopped",
	})
}

// ==================== Fork/Remix Handlers ====================

// RemixGraph forks a graph
func (h *Handler) RemixGraph(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	// Get target name from request
	var req struct {
		NewName string `json:"new_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewName == "" {
		respondError(w, http.StatusBadRequest, "new_name is required")
		return
	}

	// Fork the graph
	forked, err := h.frgRepo.ForkGraph(ctx, author, name, "v1", user.Username, req.NewName, &user.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, GraphResponse{
		GraphDefinition: *forked,
		FullName:        forked.Author + "/" + forked.Name + "@" + forked.Version,
	})
}

// ==================== AI Discovery Handlers ====================

// AIComposeRequest represents a request for AI graph composition
type AIComposeRequest struct {
	Prompt       string   `json:"prompt"`
	Requirements []string `json:"requirements,omitempty"` // e.g., ["low_latency", "cost_optimized"]
}

// AIComposeHandlerResponse represents the response from AI composition endpoint
type AIComposeHandlerResponse struct {
	Success      bool                   `json:"success"`
	Graph        *frg.GraphDefinition   `json:"graph,omitempty"`
	Explanation  map[string]interface{} `json:"explanation,omitempty"`
	Confidence   float64                `json:"confidence"`
	GenerationID string                 `json:"generation_id"`
	LatencyMs    float64                `json:"latency_ms"`
	Suggestions  []string               `json:"suggestions,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// AICompose generates a graph using AI composition
func (h *Handler) AICompose(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req AIComposeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Prompt == "" {
		respondError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	// Call AI service for composition (without TenantID to avoid leaking to external AI)
	composeReq := CompositionRequest{
		Prompt:           req.Prompt,
		Requirements:     req.Requirements,
		PreferredRuntime: "python", // Default to Python
	}

	aiResp, err := h.aiClient.Compose(ctx, composeReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("AI composition failed: %v", err))
		return
	}

	if !aiResp.Success {
		respondJSON(w, http.StatusOK, AIComposeHandlerResponse{
			Success:      false,
			Error:        aiResp.Error,
			Confidence:   aiResp.Confidence,
			GenerationID: aiResp.GenerationID,
			LatencyMs:    aiResp.LatencyMs,
			Suggestions:  aiResp.Suggestions,
		})
		return
	}

	// Convert AI response nodes to frg.GraphNodeRef
	nodeRefs := make([]frg.GraphNodeRef, len(aiResp.Graph.Nodes))
	for i, node := range aiResp.Graph.Nodes {
		nodeRefs[i] = frg.GraphNodeRef{
			NodeID:   node.NodeID,
			Author:   node.Author,
			Name:     node.Name,
			Version:  node.Version,
			Config:   node.Config,
			Metadata: map[string]interface{}{"description": node.Description},
		}
	}

	// Convert AI response edges to frg.GraphEdge
	edges := make([]frg.GraphEdge, len(aiResp.Graph.Edges))
	for i, edge := range aiResp.Graph.Edges {
		edges[i] = frg.GraphEdge{
			ID:           edge.ID,
			SourceNodeID: edge.SourceNodeID,
			TargetNodeID: edge.TargetNodeID,
			Mapping: frg.DataMapping{
				SourcePath: "*",
				TargetPath: "",
			},
			Type: frg.EdgeType(edge.Type),
		}
	}

	// Build trigger config
	var triggerConfigJSON []byte
	if aiResp.Graph.TriggerConfig != nil {
		triggerConfigJSON, _ = json.Marshal(aiResp.Graph.TriggerConfig)
	}

	// Build input/output schemas
	inputSchemaJSON, _ := json.Marshal(aiResp.Graph.InputSchema)
	outputSchemaJSON, _ := json.Marshal(aiResp.Graph.OutputSchema)

	// Create the graph definition (sanitize AI-generated content to prevent XSS)
	graphDef := &frg.GraphDefinition{
		Author:        user.Username,
		Name:          aiResp.Graph.Name,
		Version:       "v1",
		AIDescription: sanitizeString(aiResp.Graph.Description),
		ExecutionMode: frg.ExecutionMode(aiResp.Graph.ExecutionMode),
		Visibility:    aiResp.Graph.Visibility,
		OwnerUserID:   &user.UserID,
	}

	// Marshal node refs and edges
	nodeRefsJSON, _ := json.Marshal(nodeRefs)
	edgesJSON, _ := json.Marshal(edges)

	graphDef.NodeRefs = nodeRefsJSON
	graphDef.Edges = edgesJSON
	graphDef.InputSchema = inputSchemaJSON
	graphDef.OutputSchema = outputSchemaJSON
	if len(triggerConfigJSON) > 0 {
		graphDef.TriggerConfig = triggerConfigJSON
	}

	// Build explanation
	explanation := map[string]interface{}{}
	if aiResp.Explanation != nil {
		explanation["summary"] = aiResp.Explanation.Summary
		explanation["node_purposes"] = aiResp.Explanation.NodePurposes
		explanation["data_flow_description"] = aiResp.Explanation.DataFlowDescription
		explanation["trigger_explanation"] = aiResp.Explanation.TriggerExplanation
		explanation["suggested_tests"] = aiResp.Explanation.SuggestedTests
	}

	respondJSON(w, http.StatusOK, AIComposeHandlerResponse{
		Success:      true,
		Graph:        graphDef,
		Explanation:  explanation,
		Confidence:   aiResp.Confidence,
		GenerationID: aiResp.GenerationID,
		LatencyMs:    aiResp.LatencyMs,
		Suggestions:  aiResp.Suggestions,
	})
}

// SemanticSearch performs semantic search over graphs
func (h *Handler) SemanticSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	var results []*frg.GraphDefinition
	var err error

	// Try semantic search with embedding service if available
	if h.embedClient != nil {
		// Generate embedding for the query
		embedding, embedErr := h.embedClient.GenerateEmbedding(ctx, query)
		if embedErr != nil {
			// Log the error but fall back to text search
			// (In production, you might want to log this properly)
			_ = embedErr
		}

		if embedding != nil && len(embedding) > 0 {
			// Convert []float32 to []byte for pgvector
			embeddingBytes := embeddingToBytes(embedding)
			results, err = h.frgRepo.SearchByEmbedding(ctx, embeddingBytes, 10)
			if err != nil {
				// Fall back to text search on error
				results, err = h.frgRepo.SearchByText(ctx, query, 10)
			}
		} else {
			// No embedding generated, fall back to text search
			results, err = h.frgRepo.SearchByText(ctx, query, 10)
		}
	} else {
		// No embedding client configured, use text search
		results, err = h.frgRepo.SearchByText(ctx, query, 10)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var responses []GraphResponse
	for _, def := range results {
		responses = append(responses, GraphResponse{
			GraphDefinition: *def,
			FullName:        def.Author + "/" + def.Name + "@" + def.Version,
		})
	}

	respondJSON(w, http.StatusOK, responses)
}

// GenerateFunctionRequest is the request to generate a function using AI
type GenerateFunctionRequest struct {
	Author      string `json:"author"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Runtime     string `json:"runtime"`
}

// GenerateFunction generates a production-ready function using OpenRouter
func (h *Handler) GenerateFunction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUserFromContext(r)
	if user == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req GenerateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Author == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "author and name are required")
		return
	}

	// Check if function already exists
	existingFn, err := h.registryRepo.GetFunctionByAuthorName(context.Background(), req.Author, req.Name)
	if err == nil && existingFn != nil {
		respondError(w, http.StatusConflict, "Function already exists")
		return
	}

	// Call AI service for code generation using internal bypass endpoint
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://localhost:18081"
	}

	genReq := map[string]interface{}{
		"description": req.Description,
		"runtime":      req.Runtime,
	}
	body, _ := json.Marshal(genReq)

	// Use internal endpoint that bypasses auth
	httpReq, err := http.NewRequestWithContext(ctx, "POST", aiServiceURL+"/internal/composer/generate", bytes.NewReader(body))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Use internal AI service endpoint that bypasses auth
	httpReq.Header.Set("X-API-Key", os.Getenv("AI_SERVICE_API_KEY"))

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to call AI generation service")
		respondError(w, http.StatusServiceUnavailable, "AI generation service unavailable")
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{"status": resp.StatusCode, "body": string(bodyBytes[:min(500, len(bodyBytes))])}).Info("AI generation response")
	}

	if resp.StatusCode != http.StatusOK {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("AI generation failed with status %d: %s", resp.StatusCode, string(bodyBytes)))
		return
	}

	var genResp struct {
		Success bool `json:"success"`
		Result  struct {
			Code        string                  `json:"code"`
			Manifest    functionregistry.FunctionManifest `json:"manifest"`
			Explanation string                  `json:"explanation"`
		} `json:"result"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse generation response: %s", err.Error()))
		return
	}

	if !genResp.Success {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("AI generation failed: %s", genResp.Error))
		return
	}

	manifest := genResp.Result.Manifest

	// Create the function in the registry
	fn := &registry.RegistryFunction{
		Author:      req.Author,
		Name:        req.Name,
		Title:       sql.NullString{String: manifest.Name, Valid: manifest.Name != ""},
		Description: sql.NullString{String: manifest.Description, Valid: manifest.Description != ""},
		Category:    sql.NullString{String: req.Runtime, Valid: true},
		Visibility:  "private", // Default to private for auto-generated functions
		TenantID:     &user.TenantID,
		OwnerUserID: &user.UserID,
	}

	if err := h.registryRepo.CreateFunction(context.Background(), fn); err != nil {
		logrus.WithError(err).Error("Failed to create function in registry")
		respondError(w, http.StatusInternalServerError, "Failed to create function")
		return
	}

	// Create version with generated code
	manifestJSON, _ := json.Marshal(manifest)
	version := &registry.RegistryFunctionVersion{
		ID:          uuid.New(),
		FunctionID:   fn.ID,
		Version:      "1.0.0",
		Manifest:     manifestJSON,
		SourceCode:   sql.NullString{String: genResp.Result.Code, Valid: genResp.Result.Code != ""},
		Runtime:      req.Runtime,
		MemoryMB:     256,
		TimeoutMs:    30000,
		PublishedAt:  time.Now(),
	}

	if err := h.registryRepo.CreateFunctionVersion(version); err != nil {
		logrus.WithError(err).Error("Failed to create function version")
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success":    true,
		"function":   fn,
		"version":    version,
		"explanation": genResp.Result.Explanation,
	})
}

// embeddingToBytes converts a float32 slice to bytes for pgvector
func embeddingToBytes(embedding []float32) []byte {
	// pgvector expects the embedding as a string representation of the array
	// e.g., "[0.1,0.2,0.3]"
	if len(embedding) == 0 {
		return nil
	}

	// Convert to pgvector string format
	var parts []string
	for _, v := range embedding {
		parts = append(parts, fmt.Sprintf("%.6f", v))
	}
	vectorStr := "[" + strings.Join(parts, ",") + "]"
	return []byte(vectorStr)
}

// sanitizeString removes/replaces characters that could be used for XSS attacks
func sanitizeString(s string) string {
	// Use basic HTML escaping to prevent XSS in stored/generated content
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// ==================== Optimization Handlers ====================

// GetOptimizations returns AI optimization suggestions for a graph
func (h *Handler) GetOptimizations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	def, err := h.frgRepo.GetLatestVersion(ctx, author, name)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	suggestions, err := h.frgRepo.GetOptimizationSuggestions(ctx, def.ID, false)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, suggestions)
}

// ==================== Helper Functions ====================

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
