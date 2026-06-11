package agent_observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/api/middleware"
	atlaspkg "github.com/functionfly/functionfly/internal/atlas"
	"github.com/functionfly/functionfly/internal/storage"
)

type Handler struct {
	db          *gorm.DB
	atlasClient *atlaspkg.Client
	repo        *storage.AgentObservabilityRepository
	logger      *logrus.Logger
	upgrader    websocket.Upgrader
}

func NewHandler(db *gorm.DB, atlasClient *atlaspkg.Client) *Handler {
	return &Handler{
		db:          db,
		atlasClient: atlasClient,
		repo:        storage.NewAgentObservabilityRepository(db),
		logger:      logrus.New(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

type CreateRunRequest struct {
	AgentID      string            `json:"agent_id"`
	AgentType    string            `json:"agent_type"`
	SpanID       string            `json:"span_id,omitempty"`
	ParentSpanID string            `json:"parent_span_id,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type RunResponse struct {
	ID                string     `json:"id"`
	AtlasRunID        string     `json:"atlas_run_id"`
	AtlasTenantID     string     `json:"atlas_tenant_id"`
	AgentID           string     `json:"agent_id"`
	AgentType         string     `json:"agent_type"`
	SpanID            *string    `json:"span_id,omitempty"`
	ParentSpanID      *string    `json:"parent_span_id,omitempty"`
	Status            string     `json:"status"`
	TotalCostUSD      float64    `json:"total_cost_usd"`
	TotalInputTokens  int        `json:"total_input_tokens"`
	TotalOutputTokens int        `json:"total_output_tokens"`
	EventCount        int        `json:"event_count"`
	ErrorCount        int        `json:"error_count"`
	ToolCallCount     int        `json:"tool_call_count"`
	StartedAt         time.Time  `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
}

type EventResponse struct {
	EventID   string          `json:"event_id"`
	Sequence  uint64          `json:"sequence"`
	Kind      string          `json:"kind"`
	Timestamp time.Time       `json:"timestamp"`
	SystemID  string          `json:"system_id"`
	Payload   json.RawMessage `json:"payload"`
	ParentID  string          `json:"parent_id,omitempty"`
	SpanID    string          `json:"span_id,omitempty"`
}

type DecisionGraphResponse struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	EventID   string          `json:"event_id"`
	Kind      string          `json:"kind"`
	Sequence  uint64          `json:"sequence"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type StatsResponse struct {
	AtlasRunID    string  `json:"atlas_run_id"`
	DurationMs    int64   `json:"duration_ms"`
	EventCount    int     `json:"event_count"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
	ErrorCount    int     `json:"error_count"`
	ToolCallCount int     `json:"tool_call_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	CostPerToken  float64 `json:"cost_per_token"`
}

type ConfigResponse struct {
	SamplingRate      float64 `json:"sampling_rate"`
	TraceErrorsOnly   bool    `json:"trace_errors_only"`
	SampleHeadPercent float64 `json:"sample_head_percent"`
	SampleTailCount   int     `json:"sample_tail_count"`
	RetentionDays     int     `json:"retention_days"`
}

type UpdateConfigRequest struct {
	SamplingRate      *float64 `json:"sampling_rate,omitempty"`
	TraceErrorsOnly   *bool    `json:"trace_errors_only,omitempty"`
	SampleHeadPercent *float64 `json:"sample_head_percent,omitempty"`
	SampleTailCount   *int     `json:"sample_tail_count,omitempty"`
	RetentionDays     *int     `json:"retention_days,omitempty"`
}

func (h *Handler) HandleCreateRun(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentID == "" || req.AgentType == "" {
		http.Error(w, "agent_id and agent_type are required", http.StatusBadRequest)
		return
	}

	validAgentTypes := map[string]bool{"flymind": true, "agent": true, "workflow": true, "team": true}
	if !validAgentTypes[req.AgentType] {
		http.Error(w, "invalid agent_type", http.StatusBadRequest)
		return
	}

	atlasTenantID := atlaspkg.DeriveAtlasTenantID(claims.TenantID)

	metadata := map[string]string{
		"tenant_id":  claims.TenantID.String(),
		"agent_id":   req.AgentID,
		"agent_type": req.AgentType,
	}
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	atlasRunID, err := h.atlasClient.CreateRun(r.Context(), metadata)
	if err != nil {
		h.logger.WithError(err).Error("failed to create Atlas run")
		http.Error(w, "failed to create observability run", http.StatusInternalServerError)
		return
	}

	var spanID, parentSpanID *string
	if req.SpanID != "" {
		spanID = &req.SpanID
	}
	if req.ParentSpanID != "" {
		parentSpanID = &req.ParentSpanID
	}

	run := &storage.ObservabilityRun{
		ID:            uuid.New(),
		TenantID:      claims.TenantID,
		AtlasTenantID: atlasTenantID,
		AtlasRunID:    atlasRunID,
		AgentID:       req.AgentID,
		AgentType:     req.AgentType,
		SpanID:        spanID,
		ParentSpanID:  parentSpanID,
		Status:        "running",
		Metadata:      storage.JSONMap(metadata),
	}

	if err := h.repo.CreateRun(r.Context(), run); err != nil {
		h.logger.WithError(err).Error("failed to create run in database")
		http.Error(w, "failed to create run", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toRunResponse(run))
}

func (h *Handler) HandleListRuns(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	runs, total, err := h.repo.ListRuns(r.Context(), claims.TenantID, agentID, status, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("failed to list runs")
		http.Error(w, "failed to list runs", http.StatusInternalServerError)
		return
	}

	responses := make([]*RunResponse, len(runs))
	for i, run := range runs {
		responses[i] = toRunResponse(run)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"runs":   responses,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) HandleGetRun(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toRunResponse(run))
}

func (h *Handler) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	afterSeq, _ := strconv.ParseUint(r.URL.Query().Get("after_sequence"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	events, err := h.atlasClient.Replay(r.Context(), run.AtlasRunID, afterSeq)
	if err != nil {
		h.logger.WithError(err).Error("failed to get events from Atlas")
		http.Error(w, "failed to get events", http.StatusInternalServerError)
		return
	}

	if len(events) > limit {
		events = events[:limit]
	}

	responses := make([]*EventResponse, len(events))
	for i, event := range events {
		responses[i] = &EventResponse{
			EventID:   event.EventID,
			Sequence:  event.Sequence,
			Kind:      event.Kind,
			Timestamp: event.Timestamp,
			SystemID:  event.SystemID,
			Payload:   event.Payload,
			ParentID:  event.ParentID,
			SpanID:    event.SpanID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": responses,
	})
}

func (h *Handler) HandleReplay(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	afterSeq, _ := strconv.ParseUint(r.URL.Query().Get("after_sequence"), 10, 64)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("["))

	events, err := h.atlasClient.Replay(r.Context(), run.AtlasRunID, afterSeq)
	if err != nil {
		h.logger.WithError(err).Error("failed to replay events")
		http.Error(w, "failed to replay events", http.StatusInternalServerError)
		return
	}

	first := true
	for _, event := range events {
		if !first {
			w.Write([]byte(","))
		}
		first = false

		resp := &EventResponse{
			EventID:   event.EventID,
			Sequence:  event.Sequence,
			Kind:      event.Kind,
			Timestamp: event.Timestamp,
			SystemID:  event.SystemID,
			Payload:   event.Payload,
			ParentID:  event.ParentID,
			SpanID:    event.SpanID,
		}
		data, _ := json.Marshal(resp)
		w.Write(data)
	}

	w.Write([]byte("]"))
}

func (h *Handler) HandleStreamEvents(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.WithError(err).Error("failed to upgrade WebSocket")
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	events, errs := h.atlasClient.StreamEvents(ctx, run.AtlasRunID)

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}
			resp := &EventResponse{
				EventID:   event.EventID,
				Sequence:  event.Sequence,
				Kind:      event.Kind,
				Timestamp: event.Timestamp,
				SystemID:  event.SystemID,
				Payload:   event.Payload,
				ParentID:  event.ParentID,
				SpanID:    event.SpanID,
			}
			if err := conn.WriteJSON(resp); err != nil {
				h.logger.WithError(err).Error("failed to write event to WebSocket")
				return
			}
		case err, ok := <-errs:
			if !ok {
				return
			}
			h.logger.WithError(err).Error("error from Atlas stream")
			return
		}
	}
}

func (h *Handler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	stats, err := h.atlasClient.GetStats(r.Context(), run.AtlasRunID)
	if err != nil {
		h.logger.WithError(err).Error("failed to get stats from Atlas")
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&StatsResponse{
		AtlasRunID:    stats.AtlasRunID,
		DurationMs:    stats.DurationMs,
		EventCount:    stats.EventCount,
		InputTokens:   stats.InputTokens,
		OutputTokens:  stats.OutputTokens,
		TotalCostUSD:  stats.TotalCostUSD,
		ErrorCount:    stats.ErrorCount,
		ToolCallCount: stats.ToolCallCount,
		AvgLatencyMs:  stats.AvgLatencyMs,
		CostPerToken:  stats.CostPerToken,
	})
}

func (h *Handler) HandleGetGraph(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	eventID := r.URL.Query().Get("event_id")
	maxDepth, _ := strconv.Atoi(r.URL.Query().Get("max_depth"))
	if maxDepth == 0 {
		maxDepth = 10
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	graph, err := h.atlasClient.GetGraph(r.Context(), run.AtlasRunID, eventID, maxDepth)
	if err != nil {
		h.logger.WithError(err).Error("failed to get graph from Atlas")
		http.Error(w, "failed to get graph", http.StatusInternalServerError)
		return
	}

	nodes := make([]GraphNode, len(graph.Nodes))
	for i, node := range graph.Nodes {
		nodes[i] = GraphNode{
			EventID:   node.EventID,
			Kind:      node.Kind,
			Sequence:  node.Sequence,
			Payload:   node.Payload,
			Timestamp: node.Timestamp,
		}
	}

	edges := make([]GraphEdge, len(graph.Edges))
	for i, edge := range graph.Edges {
		edges[i] = GraphEdge{
			From: edge.From,
			To:   edge.To,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&DecisionGraphResponse{
		Nodes: nodes,
		Edges: edges,
	})
}

func (h *Handler) HandleEndRun(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "completed"
	}

	if req.Status != "completed" && req.Status != "failed" {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	if err := h.atlasClient.EndRun(r.Context(), run.AtlasRunID, req.Status); err != nil {
		h.logger.WithError(err).Error("failed to end Atlas run")
	}

	if err := h.repo.EndRun(r.Context(), runID, req.Status); err != nil {
		h.logger.WithError(err).Error("failed to end run in database")
		http.Error(w, "failed to end run", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) HandleCreateSpan(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	var req struct {
		SpanID       string `json:"span_id"`
		ParentSpanID string `json:"parent_span_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SpanID == "" {
		http.Error(w, "span_id is required", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	metadata := map[string]string{
		"span_id":        req.SpanID,
		"parent_span_id": req.ParentSpanID,
		"run_id":         run.AtlasRunID,
	}

	payload, _ := json.Marshal(metadata)
	_, err = h.atlasClient.AppendEvent(r.Context(), run.AtlasRunID, "DECISION", payload, "flymind-span", "")
	if err != nil {
		h.logger.WithError(err).Error("failed to create span event in Atlas")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"span_id":        req.SpanID,
		"parent_span_id": req.ParentSpanID,
	})
}

func (h *Handler) HandleListSpans(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	runID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid run ID", http.StatusBadRequest)
		return
	}

	run, err := h.repo.GetRun(r.Context(), claims.TenantID, runID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		h.logger.WithError(err).Error("failed to get run")
		http.Error(w, "failed to get run", http.StatusInternalServerError)
		return
	}

	events, err := h.atlasClient.Replay(r.Context(), run.AtlasRunID, 0)
	if err != nil {
		h.logger.WithError(err).Error("failed to get events from Atlas")
		http.Error(w, "failed to get spans", http.StatusInternalServerError)
		return
	}

	spanMap := make(map[string]map[string]interface{})
	for _, event := range events {
		if event.Kind != "DECISION" {
			continue
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			continue
		}
		if spanID, ok := payload["span_id"].(string); ok && spanID != "" {
			if _, exists := spanMap[spanID]; !exists {
				spanMap[spanID] = map[string]interface{}{
					"span_id": spanID,
				}
			}
			if parentSpanID, ok := payload["parent_span_id"].(string); ok && parentSpanID != "" {
				spanMap[spanID]["parent_span_id"] = parentSpanID
			}
		}
	}

	spans := make([]map[string]interface{}, 0, len(spanMap))
	for _, span := range spanMap {
		spans = append(spans, span)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"spans": spans,
	})
}

func (h *Handler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	config, err := h.repo.GetConfig(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("failed to get config")
		http.Error(w, "failed to get config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&ConfigResponse{
		SamplingRate:      config.SamplingRate,
		TraceErrorsOnly:   config.TraceErrorsOnly,
		SampleHeadPercent: config.SampleHeadPercent,
		SampleTailCount:   config.SampleTailCount,
		RetentionDays:     config.RetentionDays,
	})
}

func (h *Handler) HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.repo.GetConfig(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("failed to get config")
		http.Error(w, "failed to get config", http.StatusInternalServerError)
		return
	}

	if req.SamplingRate != nil {
		config.SamplingRate = *req.SamplingRate
	}
	if req.TraceErrorsOnly != nil {
		config.TraceErrorsOnly = *req.TraceErrorsOnly
	}
	if req.SampleHeadPercent != nil {
		config.SampleHeadPercent = *req.SampleHeadPercent
	}
	if req.SampleTailCount != nil {
		config.SampleTailCount = *req.SampleTailCount
	}
	if req.RetentionDays != nil {
		config.RetentionDays = *req.RetentionDays
	}

	if err := h.repo.UpsertConfig(r.Context(), config); err != nil {
		h.logger.WithError(err).Error("failed to update config")
		http.Error(w, "failed to update config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&ConfigResponse{
		SamplingRate:      config.SamplingRate,
		TraceErrorsOnly:   config.TraceErrorsOnly,
		SampleHeadPercent: config.SampleHeadPercent,
		SampleTailCount:   config.SampleTailCount,
		RetentionDays:     config.RetentionDays,
	})
}

func toRunResponse(run *storage.ObservabilityRun) *RunResponse {
	return &RunResponse{
		ID:                run.ID.String(),
		AtlasRunID:        run.AtlasRunID,
		AtlasTenantID:     run.AtlasTenantID,
		AgentID:           run.AgentID,
		AgentType:         run.AgentType,
		SpanID:            run.SpanID,
		ParentSpanID:      run.ParentSpanID,
		Status:            run.Status,
		TotalCostUSD:      run.TotalCostUSD,
		TotalInputTokens:  run.TotalInputTokens,
		TotalOutputTokens: run.TotalOutputTokens,
		EventCount:        run.EventCount,
		ErrorCount:        run.ErrorCount,
		ToolCallCount:     run.ToolCallCount,
		StartedAt:         run.StartedAt,
		EndedAt:           run.EndedAt,
	}
}
