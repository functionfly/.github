package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/gateway"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// CardRepository is the storage surface for agent cards.
type CardRepository interface {
	GetCard(ctx context.Context, agentID string) (*AgentCardInfo, error)
	ListCards(ctx context.Context, limit, offset int) ([]AgentCardInfo, int, error)
	UpsertCard(ctx context.Context, card *AgentCardInfo) error
	DeleteCard(ctx context.Context, agentID string) error
}

// Handler is the HTTP boundary for the A2A protocol surface.
// It is ONLY A2A plumbing (JSON shaping, task state transitions, SSE).
// Zero execution logic lives here.
type Handler struct {
	core      *gateway.GatewayCore
	engine    *TaskEngine
	cards     CardRepository
	logger    *logrus.Logger
}

// NewHandler creates an A2A handler.
func NewHandler(core *gateway.GatewayCore, engine *TaskEngine, cards CardRepository, logger *logrus.Logger) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		core:   core,
		engine: engine,
		cards:  cards,
		logger: logger,
	}
}

// SendTask serves POST /v1/a2a/{agent_id}/tasks/send.
func (h *Handler) SendTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		a2aCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_AGENT_ID", "agent_id is required")
		return
	}

	var req SendTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeA2AError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Parse the message into inputs.
	inputs, err := ParseTaskMessage(req.Message)
	if err != nil {
		writeA2AError(w, http.StatusBadRequest, "INVALID_MESSAGE", err.Error())
		return
	}

	// Generate a task ID.
	taskID := GenerateTaskID()

	// Resolve caller from auth context.
	caller := gateway.CallerFromContext(r.Context())

	// Build a CallRequest for the GatewayCore.
	callReq := gateway.CallRequest{
		Protocol: gateway.ProtocolA2A,
		Caller:   caller,
		Target: gateway.Target{
			AgentID: agentID,
		},
		Inputs: inputs,
		Metadata: req.Metadata,
	}

	// Execute via GatewayCore.
	result, err := h.core.Call(r.Context(), callReq)
	if err != nil {
		h.logger.WithError(err).WithField("agent_id", agentID).Error("a2a: task execution failed")
		writeA2AError(w, http.StatusInternalServerError, "EXECUTION_FAILED", err.Error())
		return
	}

	// Build the response.
	resp := GetTaskResponse{
		ID: taskID,
		Status: TaskStatusInfo{
			State:     TaskState(result.State),
			Timestamp: time.Now().UTC(),
		},
	}

	// If we got output, add it as an artifact.
	if len(result.Output) > 0 {
		resp.Artifacts = []Artifact{
			{
				Name: "result",
				Parts: []MessagePart{
					{Type: "text", Text: string(result.Output)},
				},
			},
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	a2aCORS(w)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetTask serves GET /v1/a2a/tasks/{task_id}.
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		a2aCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	taskID := mux.Vars(r)["task_id"]
	if taskID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_TASK_ID", "task_id is required")
		return
	}

	// Look up the task state from the receipt store.
	state, err := h.engine.store.GetTaskState(r.Context(), taskID)
	if err != nil {
		writeA2AError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}

	resp := GetTaskResponse{
		ID: taskID,
		Status: TaskStatusInfo{
			State:     state,
			Timestamp: time.Now().UTC(),
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	a2aCORS(w)
	_ = json.NewEncoder(w).Encode(resp)
}

// CancelTask serves POST /v1/a2a/tasks/{task_id}/cancel.
func (h *Handler) CancelTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		a2aCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	taskID := mux.Vars(r)["task_id"]
	if taskID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_TASK_ID", "task_id is required")
		return
	}

	if err := h.engine.Cancel(r.Context(), taskID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeA2AError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
			return
		}
		writeA2AError(w, http.StatusConflict, "INVALID_TRANSITION", err.Error())
		return
	}

	resp := CancelTaskResponse{
		ID: taskID,
		Status: TaskStatusInfo{
			State:     StateCanceled,
			Timestamp: time.Now().UTC(),
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	a2aCORS(w)
	_ = json.NewEncoder(w).Encode(resp)
}

// SubscribeSSE serves POST /v1/a2a/tasks/{task_id}/subscribe.
// Returns an SSE stream of task state changes.
func (h *Handler) SubscribeSSE(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["task_id"]
	if taskID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_TASK_ID", "task_id is required")
		return
	}

	// Verify the task exists.
	_, err := h.engine.store.GetTaskState(r.Context(), taskID)
	if err != nil {
		writeA2AError(w, http.StatusNotFound, "TASK_NOT_FOUND", "task not found")
		return
	}

	// Set SSE headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	a2aCORS(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeA2AError(w, http.StatusInternalServerError, "SSE_NOT_SUPPORTED", "SSE not supported")
		return
	}

	// Subscribe to task state changes.
	ch, unsubscribe := h.engine.Subscribe(taskID)
	defer unsubscribe()

	// Send initial state.
	state, _ := h.engine.store.GetTaskState(r.Context(), taskID)
	initialEvent := SSEEvent{
		Event: "task_status_change",
		Data: map[string]interface{}{
			"task_id":    taskID,
			"to_state":   string(state),
			"timestamp":  time.Now().UTC(),
		},
	}
	data, _ := json.Marshal(initialEvent.Data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", initialEvent.Event, string(data))
	flusher.Flush()

	// Stream events until the client disconnects.
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, string(data))
			flusher.Flush()
		}
	}
}

// ServeAgentCard serves GET /v1/a2a/agents/{agent_id}/card.
func (h *Handler) ServeAgentCard(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		a2aCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_AGENT_ID", "agent_id is required")
		return
	}

	card, err := h.cards.GetCard(r.Context(), agentID)
	if err != nil || card == nil {
		writeA2AError(w, http.StatusNotFound, "AGENT_NOT_FOUND", "agent card not found")
		return
	}

	a2aCard := card.ToAgentCard()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	a2aCORS(w)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(a2aCard)
}

// ListAgentCards serves GET /v1/a2a/agents/cards.
func (h *Handler) ListAgentCards(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		a2aCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	cards, total, err := h.cards.ListCards(r.Context(), 50, 0)
	if err != nil {
		h.logger.WithError(err).Error("a2a: list agent cards failed")
		writeA2AError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list agent cards")
		return
	}

	result := make([]gateway.AgentCard, 0, len(cards))
	for _, c := range cards {
		result = append(result, c.ToAgentCard())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	a2aCORS(w)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"cards": result,
		"total": total,
	})
}

// PublishAgentCard serves POST /v1/a2a/agents/cards.
func (h *Handler) PublishAgentCard(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		a2aCORS(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var card AgentCardInfo
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		writeA2AError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if card.ID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_ID", "card id is required")
		return
	}
	if card.Name == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_NAME", "card name is required")
		return
	}
	if card.ProtocolVersion == "" {
		card.ProtocolVersion = "0.3.0"
	}

	if err := h.cards.UpsertCard(r.Context(), &card); err != nil {
		h.logger.WithError(err).Error("a2a: upsert agent card failed")
		writeA2AError(w, http.StatusInternalServerError, "UPSERT_FAILED", "failed to publish agent card")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(card)
}

// DeleteAgentCard serves DELETE /v1/a2a/agents/cards/{agent_id}.
func (h *Handler) DeleteAgentCard(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeA2AError(w, http.StatusBadRequest, "MISSING_AGENT_ID", "agent_id is required")
		return
	}

	if err := h.cards.DeleteCard(r.Context(), agentID); err != nil {
		writeA2AError(w, http.StatusInternalServerError, "DELETE_FAILED", "failed to delete agent card")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// =============================================================================
// helpers
// =============================================================================

func a2aCORS(w http.ResponseWriter) {
	gateway.SetCORSHeaders(w, nil, gateway.CORSOptions{
		AllowMethods:  "GET, POST, DELETE, OPTIONS",
		AllowHeaders:  "Content-Type, Authorization, X-Agent-API-Key",
		ExposeHeaders: "X-Request-Id",
	})
}

func writeA2AError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
