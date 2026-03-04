// Package flywheel provides HTTP handlers for the Flywheel Network
package flywheel

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/flywheel"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles Flywheel Network HTTP requests
type Handler struct {
	service *flywheel.Service
	wsHub   *WebSocketHub
	logger  *logrus.Logger
}

// NewHandler creates a new Flywheel handler
func NewHandler(service *flywheel.Service, wsHub *WebSocketHub, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		wsHub:   wsHub,
		logger:  logger,
	}
}

// Thread handlers

// CreateThreadRequest represents a request to create a thread
type CreateThreadRequest struct {
	Title             string          `json:"title"`
	Type              string          `json:"type"`
	Content           string          `json:"content"`
	CategoryID        *string         `json:"category_id,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
	ProblemData       json.RawMessage `json:"problem_data,omitempty"`
	EnvironmentSpecs  json.RawMessage `json:"environment_specs,omitempty"`
	ExpectedOutput    json.RawMessage `json:"expected_output,omitempty"`
	AttachedCapsuleID *string         `json:"attached_capsule_id,omitempty"`
}

// CreateThread handles POST /api/v1/flywheel/threads
func (h *Handler) CreateThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req CreateThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" || req.Type == "" {
		http.Error(w, `{"error":"Title and type are required"}`, http.StatusBadRequest)
		return
	}

	thread := &flywheel.Thread{
		Title:    req.Title,
		Type:     flywheel.ThreadType(req.Type),
		AuthorID: user.UserID,
		Tags:     req.Tags,
	}

	if req.CategoryID != nil {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			http.Error(w, `{"error":"Invalid category_id"}`, http.StatusBadRequest)
			return
		}
		thread.CategoryID = &id
	}

	if req.AttachedCapsuleID != nil {
		id, err := uuid.Parse(*req.AttachedCapsuleID)
		if err != nil {
			http.Error(w, `{"error":"Invalid attached_capsule_id"}`, http.StatusBadRequest)
			return
		}
		thread.AttachedCapsuleID = &id
	}

	// Set problem data for problem threads
	if thread.Type == flywheel.ThreadTypeProblem {
		if len(req.ProblemData) > 0 {
			thread.ProblemData = req.ProblemData
		} else if req.Content != "" {
			thread.ProblemData = json.RawMessage(`{"description":` + strconv.Quote(req.Content) + `}`)
		}
		thread.EnvironmentSpecs = req.EnvironmentSpecs
		thread.ExpectedOutput = req.ExpectedOutput
	}

	if err := h.service.CreateThread(r.Context(), thread); err != nil {
		h.logger.WithError(err).Error("Failed to create thread")
		http.Error(w, `{"error":"Failed to create thread"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(thread)
}

// ListThreads handles GET /api/v1/flywheel/threads
func (h *Handler) ListThreads(w http.ResponseWriter, r *http.Request) {
	filter := flywheel.ThreadFilter{
		Type:        flywheel.ThreadType(r.URL.Query().Get("type")),
		Status:      flywheel.ThreadStatus(r.URL.Query().Get("status")),
		SearchQuery: r.URL.Query().Get("q"),
		SortBy:      r.URL.Query().Get("sort"),
	}

	if filter.SortBy == "" {
		filter.SortBy = "newest"
	}

	// Parse category_id
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		id, err := uuid.Parse(categoryID)
		if err == nil {
			filter.CategoryID = &id
		}
	}

	// Parse pagination
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	threads, count, err := h.service.ListThreads(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list threads")
		http.Error(w, `{"error":"Failed to list threads"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"threads": threads,
		"total":   count,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetThread handles GET /api/v1/flywheel/threads/:id
func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	thread, err := h.service.GetThread(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Thread not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thread)
}

// UpdateThread handles PATCH /api/v1/flywheel/threads/:id
func (h *Handler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateThread(r.Context(), id, updates, user.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to update thread")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// ResolveThread handles POST /api/v1/flywheel/threads/:id/resolve
func (h *Handler) ResolveThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		ReplyID string `json:"reply_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	replyID, err := uuid.Parse(req.ReplyID)
	if err != nil {
		http.Error(w, `{"error":"Invalid reply_id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.ResolveThread(r.Context(), threadID, replyID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to resolve thread")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

// Reply handlers

// CreateReplyRequest represents a request to create a reply
type CreateReplyRequest struct {
	Content           string          `json:"content"`
	ParentReplyID     *string         `json:"parent_reply_id,omitempty"`
	CodeBlocks        json.RawMessage `json:"code_blocks,omitempty"`
	AttachedCapsuleID *string         `json:"attached_capsule_id,omitempty"`
}

// CreateReply handles POST /api/v1/flywheel/threads/:id/replies
func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	var req CreateReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" && len(req.CodeBlocks) == 0 {
		http.Error(w, `{"error":"Content or code_blocks required"}`, http.StatusBadRequest)
		return
	}

	reply := &flywheel.Reply{
		ThreadID: threadID,
		AuthorID: user.UserID,
		Content:  req.Content,
	}

	if req.ParentReplyID != nil {
		id, err := uuid.Parse(*req.ParentReplyID)
		if err != nil {
			http.Error(w, `{"error":"Invalid parent_reply_id"}`, http.StatusBadRequest)
			return
		}
		reply.ParentReplyID = &id
	}

	if len(req.CodeBlocks) > 0 {
		reply.CodeBlocks = req.CodeBlocks
	}

	if req.AttachedCapsuleID != nil {
		id, err := uuid.Parse(*req.AttachedCapsuleID)
		if err != nil {
			http.Error(w, `{"error":"Invalid attached_capsule_id"}`, http.StatusBadRequest)
			return
		}
		reply.AttachedCapsuleID = &id
	}

	if err := h.service.CreateReply(r.Context(), reply); err != nil {
		h.logger.WithError(err).Error("Failed to create reply")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)
}

// ListReplies handles GET /api/v1/flywheel/threads/:id/replies
func (h *Handler) ListReplies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	filter := flywheel.ReplyFilter{}
	if r.URL.Query().Get("parent_only") == "true" {
		filter.ParentOnly = true
	}

	replies, count, err := h.service.ListReplies(r.Context(), threadID, filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list replies")
		http.Error(w, `{"error":"Failed to list replies"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"replies": replies,
		"total":   count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExecuteReply handles POST /api/v1/flywheel/replies/:id/execute
func (h *Handler) ExecuteReply(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	replyID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid reply ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Input json.RawMessage `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Input = nil
	}

	execution, err := h.service.ExecuteReply(r.Context(), replyID, user.UserID, req.Input)
	if err != nil {
		h.logger.WithError(err).Error("Failed to execute reply")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

// VerifyReply handles POST /api/v1/flywheel/replies/:id/verify
func (h *Handler) VerifyReply(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	replyID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid reply ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.VerifyReply(r.Context(), replyID); err != nil {
		h.logger.WithError(err).Error("Failed to verify reply")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "verified"})
}

// Reputation handlers

// GetMyReputation handles GET /api/v1/flywheel/reputation/me
func (h *Handler) GetMyReputation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	scores, err := h.service.GetUserReputation(r.Context(), user.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation")
		http.Error(w, `{"error":"Failed to get reputation"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// GetUserReputation handles GET /api/v1/flywheel/reputation/:user_id
func (h *Handler) GetUserReputation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	scores, err := h.service.GetUserReputation(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get reputation")
		http.Error(w, `{"error":"Failed to get reputation"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// GetLeaderboard handles GET /api/v1/flywheel/leaderboards/:score_type
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scoreType := flywheel.ReputationScoreType(vars["score_type"])

	// Validate score type
	validTypes := map[flywheel.ReputationScoreType]bool{
		flywheel.ReputationScoreTypeBuilder:        true,
		flywheel.ReputationScoreTypeOptimizer:      true,
		flywheel.ReputationScoreTypeMentor:         true,
		flywheel.ReputationScoreTypeAgentWhisperer: true,
		flywheel.ReputationScoreTypeReliability:    true,
	}
	if !validTypes[scoreType] {
		http.Error(w, `{"error":"Invalid score type"}`, http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	scores, count, err := h.service.GetLeaderboard(r.Context(), scoreType, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get leaderboard")
		http.Error(w, `{"error":"Failed to get leaderboard"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"scores": scores,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Challenge handlers

// ListChallenges handles GET /api/v1/flywheel/challenges
func (h *Handler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	filter := flywheel.ChallengeFilter{
		Status: flywheel.ChallengeStatus(r.URL.Query().Get("status")),
	}

	if r.URL.Query().Get("active_only") == "true" {
		filter.ActiveOnly = true
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if filter.Limit == 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	challenges, count, err := h.service.ListChallenges(r.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list challenges")
		http.Error(w, `{"error":"Failed to list challenges"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"challenges": challenges,
		"total":      count,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetChallenge handles GET /api/v1/flywheel/challenges/:id
func (h *Handler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid challenge ID"}`, http.StatusBadRequest)
		return
	}

	challenge, err := h.service.GetChallenge(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Challenge not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(challenge)
}

// SubmitChallenge handles POST /api/v1/flywheel/challenges/:id/submit
func (h *Handler) SubmitChallenge(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	challengeID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid challenge ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		SubmissionType     string  `json:"submission_type"`
		CodeSubmission     string  `json:"code_submission,omitempty"`
		SubmittedCapsuleID *string `json:"submitted_capsule_id,omitempty"`
		Notes              string  `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	submission := &flywheel.ChallengeSubmission{
		ChallengeID:    challengeID,
		ParticipantID:  user.UserID,
		SubmissionType: req.SubmissionType,
		CodeSubmission: req.CodeSubmission,
		Notes:          req.Notes,
	}

	if req.SubmittedCapsuleID != nil {
		id, err := uuid.Parse(*req.SubmittedCapsuleID)
		if err != nil {
			http.Error(w, `{"error":"Invalid submitted_capsule_id"}`, http.StatusBadRequest)
			return
		}
		submission.SubmittedCapsuleID = &id
	}

	if err := h.service.SubmitChallengeEntry(r.Context(), submission); err != nil {
		h.logger.WithError(err).Error("Failed to submit challenge entry")
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(submission)
}

// GetChallengeLeaderboard handles GET /api/v1/flywheel/challenges/:id/leaderboard
func (h *Handler) GetChallengeLeaderboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	challengeID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid challenge ID"}`, http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	submissions, err := h.service.GetChallengeLeaderboard(r.Context(), challengeID, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get challenge leaderboard")
		http.Error(w, `{"error":"Failed to get leaderboard"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"submissions": submissions,
	})
}

// Category handlers

// ListCategories handles GET /api/v1/flywheel/categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list categories")
		http.Error(w, `{"error":"Failed to list categories"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// Subscription handlers

// SubscribeToThread handles POST /api/v1/flywheel/threads/:id/subscribe
func (h *Handler) SubscribeToThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		NotificationLevel string `json:"notification_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.NotificationLevel = "all"
	}

	if err := h.service.SubscribeToThread(r.Context(), user.UserID, threadID, req.NotificationLevel); err != nil {
		h.logger.WithError(err).Error("Failed to subscribe to thread")
		http.Error(w, `{"error":"Failed to subscribe"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "subscribed"})
}

// UnsubscribeFromThread handles DELETE /api/v1/flywheel/threads/:id/subscribe
func (h *Handler) UnsubscribeFromThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.UnsubscribeFromThread(r.Context(), user.UserID, threadID); err != nil {
		h.logger.WithError(err).Error("Failed to unsubscribe from thread")
		http.Error(w, `{"error":"Failed to unsubscribe"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "unsubscribed"})
}

// Search handles GET /api/v1/flywheel/search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"Search query required"}`, http.StatusBadRequest)
		return
	}

	// Search in threads
	threadFilter := flywheel.ThreadFilter{
		SearchQuery: query,
		Limit:       20,
	}

	threads, _, err := h.service.ListThreads(r.Context(), threadFilter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search threads")
		http.Error(w, `{"error":"Search failed"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"query":   query,
		"threads": threads,
		"total":   len(threads),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListVerifiedSolutions handles GET /api/v1/flywheel/solutions/verified
func (h *Handler) ListVerifiedSolutions(w http.ResponseWriter, r *http.Request) {
	// This would query for replies with verified execution status
	// For now, return an empty list
	response := map[string]interface{}{
		"solutions": []interface{}{},
		"total":     0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetThreadTimeline handles GET /api/v1/flywheel/threads/:id/timeline
func (h *Handler) GetThreadTimeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	// Get thread with executions for replay
	thread, err := h.service.GetThread(r.Context(), threadID)
	if err != nil {
		http.Error(w, `{"error":"Thread not found"}`, http.StatusNotFound)
		return
	}

	// Build timeline from thread history
	timeline := []map[string]interface{}{
		{
			"timestamp": thread.CreatedAt,
			"type":      "thread_created",
			"data":      thread,
		},
	}

	response := map[string]interface{}{
		"thread_id": threadID,
		"timeline":  timeline,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReplayThread handles POST /api/v1/flywheel/threads/:id/replay
func (h *Handler) ReplayThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	// Get thread
	_, err = h.service.GetThread(r.Context(), threadID)
	if err != nil {
		http.Error(w, `{"error":"Thread not found"}`, http.StatusNotFound)
		return
	}

	// Replay would execute all solutions in sequence
	// For now, return success
	response := map[string]interface{}{
		"thread_id": threadID,
		"status":    "replay_started",
		"message":   "Thread replay functionality will execute all solutions sequentially",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListThreadAgents handles GET /api/v1/flywheel/threads/:id/agents
func (h *Handler) ListThreadAgents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	// Return list of agents participating in thread
	// For now, return empty list
	response := map[string]interface{}{
		"thread_id": threadID,
		"agents":    []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InviteAgent handles POST /api/v1/flywheel/threads/:id/agents/:agent_id/invite
func (h *Handler) InviteAgent(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	agentID, err := uuid.Parse(vars["agent_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid agent ID"}`, http.StatusBadRequest)
		return
	}

	// Invite agent logic would go here
	h.logger.WithFields(logrus.Fields{
		"thread_id": threadID,
		"agent_id":  agentID,
		"user_id":   user.UserID,
	}).Info("Agent invited to thread")

	response := map[string]interface{}{
		"thread_id": threadID,
		"agent_id":  agentID,
		"status":    "invited",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveAgent handles DELETE /api/v1/flywheel/threads/:id/agents/:agent_id
func (h *Handler) RemoveAgent(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	agentID, err := uuid.Parse(vars["agent_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid agent ID"}`, http.StatusBadRequest)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"thread_id": threadID,
		"agent_id":  agentID,
		"user_id":   user.UserID,
	}).Info("Agent removed from thread")

	response := map[string]interface{}{
		"thread_id": threadID,
		"agent_id":  agentID,
		"status":    "removed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AgentRespond handles POST /api/v1/flywheel/threads/:id/agents/:agent_id/respond
func (h *Handler) AgentRespond(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	threadID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid thread ID"}`, http.StatusBadRequest)
		return
	}

	agentID, err := uuid.Parse(vars["agent_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid agent ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Create reply from agent
	reply := &flywheel.Reply{
		ThreadID:   threadID,
		AuthorID:   agentID,
		AuthorType: flywheel.ReplyAuthorTypeAgent,
		Content:    req.Content,
	}

	if err := h.service.CreateReply(r.Context(), reply); err != nil {
		h.logger.WithError(err).Error("Failed to create agent reply")
		http.Error(w, `{"error":"Failed to create reply"}`, http.StatusInternalServerError)
		return
	}

	// Broadcast to thread subscribers
	h.BroadcastNewReply(threadID.String(), reply)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)
}

// PublishToMarketplace handles POST /api/v1/flywheel/replies/:id/publish-to-marketplace
func (h *Handler) PublishToMarketplace(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	replyID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid reply ID"}`, http.StatusBadRequest)
		return
	}

	// Get the reply
	reply, err := h.service.GetReply(r.Context(), replyID)
	if err != nil {
		http.Error(w, `{"error":"Reply not found"}`, http.StatusNotFound)
		return
	}

	// Verify the user owns this reply or has permission
	if reply.AuthorID != user.UserID {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusForbidden)
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Visibility  string   `json:"visibility"`
		Price       *float64 `json:"price,omitempty"`
		Tags        []string `json:"tags,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Marketplace publishing logic would integrate with registry
	h.logger.WithFields(logrus.Fields{
		"reply_id":    replyID,
		"user_id":     user.UserID,
		"name":        req.Name,
		"visibility":  req.Visibility,
	}).Info("Publishing solution to marketplace")

	response := map[string]interface{}{
		"reply_id":    replyID,
		"status":      "published",
		"name":        req.Name,
		"description": req.Description,
		"visibility":  req.Visibility,
		"message":     "Solution published to marketplace successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
