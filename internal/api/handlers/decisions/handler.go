package decisions

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage/trustapi/decisions"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles Team Decision API requests
type Handler struct {
	repo   *decisions.Repository
	logger *logrus.Logger
}

// NewHandler creates a new decisions handler
func NewHandler(repo *decisions.Repository) *Handler {
	return &Handler{
		repo:   repo,
		logger: logrus.New(),
	}
}

// RegisterRoutes registers decision routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	decisionsRouter := r.PathPrefix("/v1/teams/{team_id}/decisions").Subrouter()
	decisionsRouter.HandleFunc("", h.HandleCreateDecision).Methods("POST")
	decisionsRouter.HandleFunc("", h.HandleListDecisions).Methods("GET")
	decisionsRouter.HandleFunc("/search", h.HandleSearchDecisions).Methods("GET")
	decisionsRouter.HandleFunc("/{decision_id}", h.HandleGetDecision).Methods("GET")
	decisionsRouter.HandleFunc("/{decision_id}", h.HandleUpdateDecision).Methods("PUT")
	decisionsRouter.HandleFunc("/{decision_id}", h.HandleDeleteDecision).Methods("DELETE")
	decisionsRouter.HandleFunc("/{decision_id}/approve", h.HandleApproveDecision).Methods("POST")
}

// getUserFromContext extracts the authenticated user from context
func getUserFromContext(r *http.Request) *auth.Claims {
	return middleware.GetUserFromContext(r)
}

// parseTeamID parses the team_id from URL params
func parseTeamID(vars map[string]string) (uuid.UUID, error) {
	teamIDStr := vars["team_id"]
	return uuid.Parse(teamIDStr)
}

// parseDecisionID parses the decision_id from URL params
func parseDecisionID(vars map[string]string) (uuid.UUID, error) {
	decisionIDStr := vars["decision_id"]
	return uuid.Parse(decisionIDStr)
}

// HandleCreateDecision handles POST /v1/teams/{team_id}/decisions
func (h *Handler) HandleCreateDecision(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	teamID, err := parseTeamID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid team ID", "invalid_team_id")
		return
	}

	var req decisions.CreateDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_body")
		return
	}

	// Validate required fields
	if req.Title == "" {
		h.writeError(w, http.StatusBadRequest, "Title is required", "missing_title")
		return
	}

	decision, err := h.repo.Create(teamID, user.UserID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create decision")
		h.writeError(w, http.StatusInternalServerError, "Failed to create decision", "create_failed")
		return
	}

	h.writeJSON(w, http.StatusCreated, decision.ToResponse())
}

// HandleListDecisions handles GET /v1/teams/{team_id}/decisions
func (h *Handler) HandleListDecisions(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	teamID, err := parseTeamID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid team ID", "invalid_team_id")
		return
	}

	// Parse query params
	status := r.URL.Query().Get("status")
	tag := r.URL.Query().Get("tag")
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parsePositiveInt(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := parsePositiveInt(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	decisionsList, total, err := h.repo.ListByTeam(teamID, status, tag, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list decisions")
		h.writeError(w, http.StatusInternalServerError, "Failed to list decisions", "list_failed")
		return
	}

	responses := make([]decisions.DecisionResponse, len(decisionsList))
	for i, d := range decisionsList {
		responses[i] = *d.ToResponse()
	}

	h.writeJSON(w, http.StatusOK, decisions.ListDecisionsResponse{
		Decisions:  responses,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	})
}

// HandleGetDecision handles GET /v1/teams/{team_id}/decisions/{decision_id}
func (h *Handler) HandleGetDecision(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	decisionID, err := parseDecisionID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid decision ID", "invalid_decision_id")
		return
	}

	decision, err := h.repo.GetByID(decisionID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Decision not found", "not_found")
		return
	}

	h.writeJSON(w, http.StatusOK, decision.ToResponse())
}

// HandleUpdateDecision handles PUT /v1/teams/{team_id}/decisions/{decision_id}
func (h *Handler) HandleUpdateDecision(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	decisionID, err := parseDecisionID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid decision ID", "invalid_decision_id")
		return
	}

	var req decisions.UpdateDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_body")
		return
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{
			"pending": true, "approved": true, "superseded": true, "deprecated": true,
		}
		if !validStatuses[*req.Status] {
			h.writeError(w, http.StatusBadRequest, "Invalid status", "invalid_status")
			return
		}
	}

	// Validate title if provided
	if req.Title != nil && len(*req.Title) < 3 {
		h.writeError(w, http.StatusBadRequest, "Title must be at least 3 characters", "title_too_short")
		return
	}

	decision, err := h.repo.Update(decisionID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update decision")
		h.writeError(w, http.StatusInternalServerError, "Failed to update decision", "update_failed")
		return
	}

	h.writeJSON(w, http.StatusOK, decision.ToResponse())
}

// HandleDeleteDecision handles DELETE /v1/teams/{team_id}/decisions/{decision_id}
func (h *Handler) HandleDeleteDecision(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	decisionID, err := parseDecisionID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid decision ID", "invalid_decision_id")
		return
	}

	if err := h.repo.Delete(decisionID); err != nil {
		h.logger.WithError(err).Error("Failed to delete decision")
		h.writeError(w, http.StatusInternalServerError, "Failed to delete decision", "delete_failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleApproveDecision handles POST /v1/teams/{team_id}/decisions/{decision_id}/approve
func (h *Handler) HandleApproveDecision(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	decisionID, err := parseDecisionID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid decision ID", "invalid_decision_id")
		return
	}

	var req decisions.ApproveDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "invalid_body")
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"approved": true, "superseded": true, "deprecated": true,
	}
	if !validStatuses[req.Status] {
		h.writeError(w, http.StatusBadRequest, "Status must be approved, superseded, or deprecated", "invalid_status")
		return
	}

	decision, err := h.repo.Approve(decisionID, user.UserID, req.Status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to approve decision")
		h.writeError(w, http.StatusInternalServerError, "Failed to approve decision", "approve_failed")
		return
	}

	h.writeJSON(w, http.StatusOK, decision.ToResponse())
}

// HandleSearchDecisions handles GET /v1/teams/{team_id}/decisions/search?q=
func (h *Handler) HandleSearchDecisions(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "unauthorized")
		return
	}

	vars := mux.Vars(r)
	teamID, err := parseTeamID(vars)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid team ID", "invalid_team_id")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		h.writeError(w, http.StatusBadRequest, "Query parameter 'q' is required", "missing_query")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := parsePositiveInt(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.repo.SearchByText(teamID, query, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search decisions")
		h.writeError(w, http.StatusInternalServerError, "Failed to search decisions", "search_failed")
		return
	}

	responses := make([]decisions.DecisionResponse, len(results))
	for i, d := range results {
		responses[i] = *d.ToResponse()
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"decisions": responses,
		"query":     query,
		"count":     len(responses),
	})
}

// Helper functions

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errMsg string, code string) {
	h.writeJSON(w, status, map[string]string{
		"error": errMsg,
		"code":  code,
	})
}

func parsePositiveInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{s}
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

type parseError struct {
	s string
}

func (e *parseError) Error() string {
	return "invalid integer: " + e.s
}
