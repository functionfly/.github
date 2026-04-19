package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/evolution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EvolutionHandler exposes the agent evolution API over HTTP.
type EvolutionHandler struct {
	db           *gorm.DB
	evolutionSvc *evolution.Service
	identityRepo *identity.Repository
}

// NewEvolutionHandler creates a new evolution handler.
func NewEvolutionHandler(db *gorm.DB, evolutionSvc *evolution.Service, identityRepo *identity.Repository) *EvolutionHandler {
	return &EvolutionHandler{
		db:           db,
		evolutionSvc: evolutionSvc,
		identityRepo: identityRepo,
	}
}

// requireAgentTenant verifies the request is authenticated and the agent belongs to the caller's tenant.
func (h *EvolutionHandler) requireAgentTenant(w http.ResponseWriter, r *http.Request, agentID string) bool {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return false
	}
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
		return false
	}
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return false
	}
	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return false
	}
	return true
}

// ListSuggestions handles GET /v1/agents/{id}/evolution/suggestions
// Query params: status (pending/approved/rejected/implemented), limit, offset
func (h *EvolutionHandler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	status := r.URL.Query().Get("status")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	proposals, err := h.evolutionSvc.GetProposals(r.Context(), agentID, status)
	if err != nil {
		logrus.WithError(err).Error("failed to list evolution suggestions")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list suggestions")
		return
	}

	// Apply limit/offset in memory (for small result sets)
	total := len(proposals)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	if offset < total {
		proposals = proposals[offset:end]
	} else {
		proposals = []identity.EvolutionProposal{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"suggestions": proposals,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// ApproveSuggestion handles POST /v1/agents/{id}/evolution/suggestions/{suggestion_id}/approve
func (h *EvolutionHandler) ApproveSuggestion(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	suggestionIDStr := mux.Vars(r)["suggestion_id"]
	suggestionID, err := uuid.Parse(suggestionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid suggestion_id")
		return
	}

	claims := middleware.GetUserFromContext(r)
	approverID := claims.UserID.String()

	proposal, err := h.evolutionSvc.ApproveProposal(r.Context(), suggestionID, approverID)
	if err != nil {
		logrus.WithError(err).Error("failed to approve evolution suggestion")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Auto-implement if approved (can be made async in production)
	if proposal.Status == "approved" {
		if err := h.evolutionSvc.ImplementProposal(r.Context(), suggestionID); err != nil {
			logrus.WithError(err).Warn("failed to auto-implement approved proposal")
			// Don't fail the request - the proposal is approved, implementation can be retried
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"suggestion_id": suggestionID,
		"status":        "approved",
		"implemented":   proposal.Status == "approved",
	})
}

// RejectSuggestion handles POST /v1/agents/{id}/evolution/suggestions/{suggestion_id}/reject
func (h *EvolutionHandler) RejectSuggestion(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	suggestionIDStr := mux.Vars(r)["suggestion_id"]
	suggestionID, err := uuid.Parse(suggestionIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid suggestion_id")
		return
	}

	claims := middleware.GetUserFromContext(r)
	rejectedBy := claims.UserID.String()

	proposal, err := h.evolutionSvc.RejectProposal(r.Context(), suggestionID, rejectedBy)
	if err != nil {
		logrus.WithError(err).Error("failed to reject evolution suggestion")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"suggestion_id": suggestionID,
		"status":        proposal.Status,
	})
}

// ToggleEvolutionMode handles POST /v1/agents/{id}/evolution/auto-enable
// Body: { "enabled": true/false }
func (h *EvolutionHandler) ToggleEvolutionMode(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Update the agent's evolution_enabled flag
	result := h.db.WithContext(r.Context()).
		Model(&identity.AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Update("evolution_enabled", req.Enabled)
	if result.Error != nil {
		logrus.WithError(result.Error).Error("failed to toggle evolution mode")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update agent")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                true,
		"agent_id":          agentID,
		"evolution_enabled": req.Enabled,
	})
}

// GetEvolutionHistory handles GET /v1/agents/{id}/evolution/history
// Returns implemented evolutions audit log
func (h *EvolutionHandler) GetEvolutionHistory(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	// Get implemented proposals
	proposals, err := h.evolutionSvc.GetProposals(r.Context(), agentID, "implemented")
	if err != nil {
		logrus.WithError(err).Error("failed to get evolution history")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get history")
		return
	}

	// Build history entries
	history := make([]EvolutionHistoryEntry, 0, len(proposals))
	for _, p := range proposals {
		entry := EvolutionHistoryEntry{
			ID:            p.ID,
			Type:          p.ProposalType,
			Status:        p.Status,
			Data:          p.ProposalData,
			ImplementedAt: p.ImplementedAt,
			ApprovedBy:    p.ApprovedBy,
			CreatedAt:     p.CreatedAt,
		}
		history = append(history, entry)
	}

	if len(history) > limit {
		history = history[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"history": history,
		"count":   len(history),
	})
}

// GetEvolutionStatus handles GET /v1/agents/{id}/evolution/status
// Returns current evolution status for an agent
func (h *EvolutionHandler) GetEvolutionStatus(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	// Count pending suggestions
	pending, err := h.evolutionSvc.GetProposals(r.Context(), agentID, "pending")
	if err != nil {
		logrus.WithError(err).Warn("failed to count pending proposals")
		pending = []identity.EvolutionProposal{}
	}

	// Count implemented
	implemented, err := h.evolutionSvc.GetProposals(r.Context(), agentID, "implemented")
	if err != nil {
		logrus.WithError(err).Warn("failed to count implemented proposals")
		implemented = []identity.EvolutionProposal{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"status": EvolutionStatus{
			AgentID:          agentID,
			EvolutionEnabled: agent.EvolutionEnabled,
			PendingCount:     len(pending),
			ImplementedCount: len(implemented),
			CanEvolve:        agent.EvolutionEnabled && len(pending) < 10,
		},
	})
}

// TriggerAnalysis handles POST /v1/agents/{id}/evolution/analyze
// Manually triggers performance analysis and may generate new suggestions
func (h *EvolutionHandler) TriggerAnalysis(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req struct {
		TimeWindowHours int `json:"time_window_hours,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to 24 hours
		req.TimeWindowHours = 24
	}
	if req.TimeWindowHours <= 0 {
		req.TimeWindowHours = 24
	}
	if req.TimeWindowHours > 168 { // Max 1 week
		req.TimeWindowHours = 168
	}

	timeWindow := time.Duration(req.TimeWindowHours) * time.Hour

	// Perform analysis
	analysis, err := h.evolutionSvc.AnalyzePerformance(r.Context(), agentID, timeWindow)
	if err != nil {
		logrus.WithError(err).Error("failed to analyze agent performance")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "analysis failed")
		return
	}

	// Generate proposal if analysis indicates need
	var proposal *identity.EvolutionProposal
	if analysis.SuccessRate < 80 || analysis.AvgLatencyMs > 10000 {
		proposal, err = h.evolutionSvc.ProposeEvolution(r.Context(), agentID, analysis)
		if err != nil {
			logrus.WithError(err).Warn("failed to generate evolution proposal")
			// Don't fail - analysis succeeded even if proposal generation failed
		}
	}

	result := map[string]any{
		"ok":               true,
		"analysis":         analysis,
		"proposal_created": proposal != nil,
	}
	if proposal != nil {
		result["proposal_id"] = proposal.ID
		result["proposal_type"] = proposal.ProposalType
	}

	writeJSON(w, http.StatusOK, result)
}

// EvolutionStatus represents the evolution status for an agent
type EvolutionStatus struct {
	AgentID          string `json:"agent_id"`
	EvolutionEnabled bool   `json:"evolution_enabled"`
	PendingCount     int    `json:"pending_count"`
	ImplementedCount int    `json:"implemented_count"`
	CanEvolve        bool   `json:"can_evolve"`
}

// EvolutionHistoryEntry represents a single evolution history record
type EvolutionHistoryEntry struct {
	ID            uuid.UUID      `json:"id"`
	Type          string         `json:"type"`
	Status        string         `json:"status"`
	Data          map[string]any `json:"data"`
	ImplementedAt *time.Time     `json:"implemented_at,omitempty"`
	ApprovedBy    *string        `json:"approved_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// RegisterRoutes registers all evolution routes
func (h *EvolutionHandler) RegisterRoutes(router *mux.Router, basePath string, authMiddleware *middleware.AuthMiddleware) {
	auth := authMiddleware.RequireAuth
	agent := router.PathPrefix(basePath + "/agents").Subrouter()

	// Core evolution endpoints
	agent.HandleFunc("/{id}/evolution/suggestions", auth(h.ListSuggestions)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/evolution/suggestions/{suggestion_id}/approve", auth(h.ApproveSuggestion)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/evolution/suggestions/{suggestion_id}/reject", auth(h.RejectSuggestion)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/evolution/auto-enable", auth(h.ToggleEvolutionMode)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/evolution/history", auth(h.GetEvolutionHistory)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/evolution/status", auth(h.GetEvolutionStatus)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/evolution/analyze", auth(h.TriggerAnalysis)).Methods(http.MethodPost)
}
