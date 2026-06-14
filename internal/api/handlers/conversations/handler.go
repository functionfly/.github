package conversations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/conversations"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/security"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles conversation (DM) and message API requests.
type Handler struct {
	repo         *storage.ConversationRepository
	registryRepo *registry.RegistryRepository
	notify       *notification.Service
	users        userByIDGetter
	logger       *logrus.Logger
	// memoryEvents is the team memory event publisher (may be nil if team memory not configured)
	memoryEvents ConversationEventPublisher
	// wsHub broadcasts real-time events to connected WebSocket clients (may be nil).
	wsHub *ConversationWebSocketHub
}

// ConversationEventPublisher defines the interface for publishing conversation events to team memory
type ConversationEventPublisher interface {
	PublishResolved(ctx context.Context, conv *conversations.Conversation) error
}

// userByIDGetter resolves users for notification copy (minimal surface for tests).
type userByIDGetter interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*storage.User, error)
}

// NewHandler creates a new conversations handler. notify and users may be nil.
// registryRepo may be nil for stub behaviour.
func NewHandler(
	repo *storage.ConversationRepository,
	registryRepo *registry.RegistryRepository,
	notify *notification.Service,
	users userByIDGetter,
	logger *logrus.Logger,
) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	return &Handler{
		repo:         repo,
		registryRepo: registryRepo,
		notify:       notify,
		users:        users,
		logger:       logger,
		memoryEvents: nil, // Set via SetMemoryEventPublisher if team memory is enabled
	}
}

// SetMemoryEventPublisher sets the team memory event publisher for conversation webhooks
func (h *Handler) SetMemoryEventPublisher(publisher ConversationEventPublisher) {
	h.memoryEvents = publisher
}

// SetWebSocketHub attaches a WebSocket hub for real-time message broadcasting.
func (h *Handler) SetWebSocketHub(hub *ConversationWebSocketHub) {
	h.wsHub = hub
}

// GetCollaborationProfile handles GET /api/v1/conversations/collaboration-profile/:user_id
// Returns reputation and collaboration insight for the sidebar.
func (h *Handler) GetCollaborationProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	if userIDStr == "" {
		http.Error(w, `{"error":"user_id required"}`, http.StatusBadRequest)
		return
	}
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid user_id"}`, http.StatusBadRequest)
		return
	}

	// Verify target user exists
	targetUser, err := h.users.GetUserByID(r.Context(), targetUserID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Get real collaboration stats from database
	_ = r.Context() // TODO: fix - ctx was unused
	reputation := map[string]int{
		"builder": 0,
		"mentor":  0,
	}
	sharedThreads := 0
	functionsOverlap := []string{}

	// Query trust score from registry ratings
	var avgTrustScore float64
	if err := h.repo.DB().Raw(`
		SELECT COALESCE(AVG(rat.overall_score), 0)
		FROM registry_functions rf
		LEFT JOIN registry_function_ratings rat ON rat.function_id = rf.id
		WHERE rf.owner_user_id = ?`, targetUserID).Scan(&avgTrustScore).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to get avg trust score for collaboration profile")
	}

	// Query function count (builder reputation)
	var functionCount int64
	if err := h.repo.DB().Model(&storage.RegistryFunction{}).Where("owner_user_id = ?", targetUserID).Count(&functionCount).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to get function count for collaboration profile")
	}

	// Query shared conversations (threads) between current user and target user
	if err := h.repo.DB().Raw(`
		SELECT COUNT(DISTINCT cc.id)
		FROM conversation_conversations cc
		JOIN conversation_conversation_participants ccp1 ON cc.id = ccp1.conversation_id AND ccp1.user_id = ?
		JOIN conversation_conversation_participants ccp2 ON cc.id = ccp2.conversation_id AND ccp2.user_id = ?
		WHERE cc.id IN (
			SELECT conversation_id FROM conversation_conversation_participants WHERE user_id = ?
			INTERSECT
			SELECT conversation_id FROM conversation_conversation_participants WHERE user_id = ?
		)`, user.UserID, targetUserID, user.UserID, targetUserID).Scan(&sharedThreads).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to get shared threads for collaboration profile")
	}

	// Query execution count as mentor indicator
	var executionCount int64
	if err := h.repo.DB().Raw(`
		SELECT COALESCE(SUM(execution_count), 0)
		FROM registry_function_executions
		WHERE function_id IN (SELECT id FROM registry_functions WHERE owner_user_id = ?)
		AND created_at > NOW() - INTERVAL '30 days'`, targetUserID).Scan(&executionCount).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to get execution count for collaboration profile")
	}

	// Set builder reputation based on function count (0-100 scale)
	builderRep := int(functionCount * 10)
	if builderRep > 100 {
		builderRep = 100
	}
	reputation["builder"] = builderRep

	// Set mentor reputation based on recent executions (0-100 scale)
	mentorRep := int(executionCount)
	if mentorRep > 100 {
		mentorRep = 100
	}
	reputation["mentor"] = mentorRep

	// Get overlapping functions (functions both users have interacted with)
	type functionOverlap struct {
		FunctionName string
	}
	var overlaps []functionOverlap
	if err := h.repo.DB().Raw(`
		SELECT rf.author || '/' || rf.name as function_name
		FROM registry_functions rf
		WHERE rf.owner_user_id = ?
		AND rf.name IN (
			SELECT DISTINCT rf2.name
			FROM registry_functions rf2
			JOIN registry_function_executions rfe ON rfe.function_id = rf2.id
			WHERE rf2.owner_user_id = ?
			AND rfe.created_at > NOW() - INTERVAL '90 days'
		)
		LIMIT 10`, targetUserID, user.UserID).Scan(&overlaps).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to get functions overlap for collaboration profile")
	}
	for _, f := range overlaps {
		functionsOverlap = append(functionsOverlap, f.FunctionName)
	}

	trustScore := int(avgTrustScore * 100)
	if trustScore < 0 {
		trustScore = 0
	}
	if trustScore > 100 {
		trustScore = 100
	}

	profile := map[string]interface{}{
		"user_id":            targetUserID.String(),
		"username":           targetUser.Username,
		"reputation":         reputation,
		"trust_score":       trustScore,
		"shared_threads":    sharedThreads,
		"functions_overlap": functionsOverlap,
		"function_count":    functionCount,
		"recent_executions": executionCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// GetConversationContext handles GET /api/v1/conversations/context?function=author/name
// Returns smart context for composing messages about a function (last failures, trust, open issues).
func (h *Handler) GetConversationContext(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	fn := r.URL.Query().Get("function")
	if fn == "" {
		http.Error(w, `{"error":"function query required (e.g. author/name)"}`, http.StatusBadRequest)
		return
	}

	if h.registryRepo == nil {
		http.Error(w, `{"error":"Registry not available","code":"REGISTRY_UNAVAILABLE","message":"Function context lookup is not available"}`, http.StatusServiceUnavailable)
		return
	}

	parts := strings.SplitN(fn, "/", 2)
	author, name := "", ""
	if len(parts) >= 1 {
		author = strings.TrimSpace(parts[0])
	}
	if len(parts) == 2 {
		name = strings.TrimSpace(parts[1])
	}
	if author == "" || name == "" {
		http.Error(w, `{"error":"function must be author/name (e.g. author/name)"}`, http.StatusBadRequest)
		return
	}

	fnRecord, err := h.registryRepo.GetFunctionByAuthorName(r.Context(), author, name)
	if err != nil {
		h.logger.WithError(err).WithField("function", fn).Error("GetFunctionByAuthorName failed")
		http.Error(w, `{"error":"Failed to lookup function","code":"LOOKUP_FAILED","message":"Unable to retrieve function information"}`, http.StatusInternalServerError)
		return
	}
	if fnRecord == nil {
		http.Error(w, fmt.Sprintf(`{"error":"Function not found","code":"FUNCTION_NOT_FOUND","message":"Function '%s' does not exist"}`, fn), http.StatusNotFound)
		return
	}

	trustScore := 50
	since := time.Now().AddDate(0, 0, -30)
	if total, successRate, _, _, _, _, _, err := h.registryRepo.GetFunctionTrustStats(r.Context(), fnRecord.ID, since); err == nil && total > 0 {
		trustScore = int(successRate)
		if trustScore < 0 {
			trustScore = 0
		}
		if trustScore > 100 {
			trustScore = 100
		}
	}

	lastFailures := []interface{}{}
	if failures, err := h.registryRepo.GetRecentFailedExecutions(r.Context(), fnRecord.ID, 10); err == nil {
		for _, e := range failures {
			entry := map[string]interface{}{
				"id":        e.ID.String(),
				"timestamp": e.Timestamp.Format(time.RFC3339),
				"outcome":   e.Outcome,
				"version":   e.Version,
			}
			if e.ErrorCode.Valid {
				entry["error_code"] = e.ErrorCode.String
			}
			lastFailures = append(lastFailures, entry)
		}
	}

	ctx := map[string]interface{}{
		"function":        fn,
		"trust_score":     trustScore,
		"last_failures":   lastFailures,
		"open_issues":     []interface{}{},
		"suggested_hints": []string{"Include execution hash if reporting a bug.", "Mention version for reproducibility."},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ctx)
}

// ListConversations handles GET /api/v1/conversations
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := h.repo.ListConversationsForUser(r.Context(), user.UserID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("List conversations failed")
		http.Error(w, `{"error":"Failed to list conversations"}`, http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []conversations.ConversationListEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"conversations": list})
}

// CreateConversationRequest is the body for creating a conversation.
type CreateConversationRequest struct {
	Type           string   `json:"type"`
	ParticipantIDs []string `json:"participant_ids"` // UUID strings
	SourceThreadID *string  `json:"source_thread_id,omitempty"`
	OrganizationID *string  `json:"organization_id,omitempty"`
}

// CreateFromThreadRequest is the body for "Move to Private Debug Thread".
type CreateFromThreadRequest struct {
	ThreadID string `json:"thread_id"` // Flywheel thread UUID
}

// CreateFromThread handles POST /api/v1/conversations/from-thread
// Creates a conversation linked to a Flywheel thread (move to private debug).
func (h *Handler) CreateFromThread(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	var req CreateFromThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	threadID, err := uuid.Parse(req.ThreadID)
	if err != nil {
		http.Error(w, `{"error":"Invalid thread_id"}`, http.StatusBadRequest)
		return
	}
	// Create conversation with current user and thread reference; participants would be filled by caller or lookup
	participantIDs := []string{user.UserID.String()}
	participantIDsJSON, _ := json.Marshal(participantIDs)
	c := &conversations.Conversation{
		Type:           conversations.TypeIssueThread,
		ParticipantIDs: participantIDsJSON,
		SourceThreadID: &threadID,
	}
	if err := h.repo.CreateConversation(r.Context(), c); err != nil {
		h.logger.WithError(err).Error("Create from thread failed")
		http.Error(w, `{"error":"Failed to create conversation"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// CreateConversation handles POST /api/v1/conversations
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	// Ensure current user is in participants
	participantIDs := req.ParticipantIDs
	me := user.UserID.String()
	hasMe := false
	for _, id := range participantIDs {
		if id == me {
			hasMe = true
			break
		}
	}
	if !hasMe {
		participantIDs = append([]string{me}, participantIDs...)
	}
	participantIDsJSON, _ := json.Marshal(participantIDs)
	c := &conversations.Conversation{
		Type:           conversations.ConversationType(req.Type),
		ParticipantIDs: participantIDsJSON,
	}
	if req.Type == "" {
		c.Type = conversations.TypeDM
	}
	if req.SourceThreadID != nil {
		id, err := uuid.Parse(*req.SourceThreadID)
		if err == nil {
			c.SourceThreadID = &id
		}
	}
	if req.OrganizationID != nil {
		id, err := uuid.Parse(*req.OrganizationID)
		if err == nil {
			c.OrganizationID = &id
		}
	}
	if err := h.repo.CreateConversation(r.Context(), c); err != nil {
		h.logger.WithError(err).Error("Create conversation failed")
		http.Error(w, `{"error":"Failed to create conversation"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GetConversation handles GET /api/v1/conversations/:id
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	c, err := h.repo.GetConversationByID(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Get conversation failed")
		http.Error(w, `{"error":"Failed to get conversation"}`, http.StatusInternalServerError)
		return
	}
	if c == nil {
		http.Error(w, `{"error":"Conversation not found"}`, http.StatusNotFound)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// ListMessages handles GET /api/v1/conversations/:id/messages
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	list, err := h.repo.ListMessages(r.Context(), id, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("List messages failed")
		http.Error(w, `{"error":"Failed to list messages"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"messages": list})
}

// MarkConversationRead handles POST /api/v1/conversations/:id/read
// Sets the current user's read cursor to the latest message (clears sidebar unread for this thread).
func (h *Handler) MarkConversationRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	if err := h.repo.MarkConversationRead(r.Context(), id, user.UserID); err != nil {
		h.logger.WithError(err).Error("Mark conversation read failed")
		http.Error(w, `{"error":"Failed to update read state"}`, http.StatusInternalServerError)
		return
	}

	// Broadcast read status to other participants via WebSocket.
	if h.wsHub != nil {
		readPayload, _ := json.Marshal(map[string]interface{}{
			"user_id":         user.UserID,
			"conversation_id": id,
		})
		h.wsHub.BroadcastToConversation(id, &ConvWSMessage{
			Type:    "conversation_read",
			Payload: readPayload,
		}, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ValidateMessageRequest is the body for validating a message (security scan).
type ValidateMessageRequest struct {
	Content string `json:"content"`
}

// ValidateMessage handles POST /api/v1/conversations/:id/messages/validate
func (h *Handler) ValidateMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	var req ValidateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	scan := security.ScanMessageContent(req.Content)
	warning := ""
	if len(req.Content) > 5000 {
		warning = "Message is very long; consider using a snippet instead."
	}
	for _, w := range scan.Warnings {
		if warning != "" {
			warning += " "
		}
		warning += w
	}
	if len(scan.SecretsFound) > 0 {
		warning = "Suspected secrets or tokens detected (e.g. " + strings.Join(scan.SecretsFound, ", ") + "). Remove them before sending."
	}
	resp := map[string]interface{}{
		"valid":   scan.Valid,
		"warning": warning,
	}
	if len(scan.SecretsFound) > 0 {
		resp["secrets_found"] = scan.SecretsFound
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateMessageRequest is the body for creating a message.
type CreateMessageRequest struct {
	Content    string          `json:"content"`
	Embeddings json.RawMessage `json:"embeddings,omitempty"`
}

// CreateMessage handles POST /api/v1/conversations/:id/messages
func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Embeddings == nil {
		req.Embeddings = json.RawMessage("{}")
	}

	// Server-side message content length validation
	if req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"message content is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Content) > 10000 {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"message content exceeds maximum length"}`, http.StatusBadRequest)
		return
	}
	m := &conversations.ConversationMessage{
		ConversationID: id,
		AuthorID:       user.UserID,
		Content:        req.Content,
		Embeddings:     req.Embeddings,
	}
	if err := h.repo.CreateMessage(r.Context(), m); err != nil {
		h.logger.WithError(err).Error("Create message failed")
		http.Error(w, `{"error":"Failed to send message"}`, http.StatusInternalServerError)
		return
	}

	if err := h.repo.MarkConversationRead(r.Context(), id, user.UserID); err != nil {
		h.logger.WithError(err).Debug("Mark conversation read after send failed")
	}

	h.notifyConversationMessage(r.Context(), id, user.UserID, m)

	// Broadcast the new message to all WebSocket-connected participants.
	if h.wsHub != nil {
		h.wsHub.BroadcastNewMessage(m)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// ResolveConversationRequest is the body for resolving a conversation.
type ResolveConversationRequest struct {
	MessageID string `json:"message_id"` // UUID of the message that is the accepted solution
}

// ResolveConversation handles POST /api/v1/conversations/:id/resolve
func (h *Handler) ResolveConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	var req ResolveConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	var messageID *uuid.UUID
	if req.MessageID != "" {
		parsed, err := uuid.Parse(req.MessageID)
		if err != nil {
			http.Error(w, `{"error":"Invalid message_id"}`, http.StatusBadRequest)
			return
		}
		messageID = &parsed
	}
	if err := h.repo.ResolveConversation(r.Context(), id, user.UserID, messageID); err != nil {
		h.logger.WithError(err).Error("Resolve conversation failed")
		http.Error(w, `{"error":"Failed to resolve conversation"}`, http.StatusInternalServerError)
		return
	}
	c, _ := h.repo.GetConversationByID(r.Context(), id)

	// Trigger team memory extraction webhook if enabled
	if h.memoryEvents != nil && c != nil {
		go func(conv *conversations.Conversation) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := h.memoryEvents.PublishResolved(ctx, conv); err != nil {
				h.logger.WithError(err).Warn("Failed to publish conversation resolved event for memory extraction")
			}
		}(c)
	}

	// Broadcast resolution to all WebSocket-connected participants.
	if h.wsHub != nil && c != nil {
		h.wsHub.BroadcastConversationResolved(c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// ListBounties handles GET /api/v1/conversations/:id/bounties
func (h *Handler) ListBounties(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	list, err := h.repo.ListBountiesForConversation(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("List bounties failed")
		http.Error(w, `{"error":"Failed to list bounties"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"bounties": list})
}

// CreateBountyRequest is the body for attaching a bounty.
type CreateBountyRequest struct {
	AmountReputation         int     `json:"amount_reputation"`
	AmountCents              int     `json:"amount_cents,omitempty"`
	SecurityWeightMultiplier float64 `json:"security_weight_multiplier,omitempty"`
}

// CreateBounty handles POST /api/v1/conversations/:id/bounties
func (h *Handler) CreateBounty(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), id, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	var req CreateBountyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.AmountReputation < 0 {
		req.AmountReputation = 0
	}
	if req.SecurityWeightMultiplier <= 0 {
		req.SecurityWeightMultiplier = 1.0
	}
	b := &conversations.ConversationBounty{
		ConversationID:           id,
		OfferedBy:                user.UserID,
		AmountReputation:         req.AmountReputation,
		AmountCents:              req.AmountCents,
		SecurityWeightMultiplier: req.SecurityWeightMultiplier,
	}
	if err := h.repo.CreateBounty(r.Context(), b); err != nil {
		h.logger.WithError(err).Error("Create bounty failed")
		http.Error(w, `{"error":"Failed to attach bounty"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

// ClaimBounty handles POST /api/v1/conversations/:id/bounties/:bounty_id/claim
func (h *Handler) ClaimBounty(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	conversationID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	bountyID, err := uuid.Parse(vars["bounty_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid bounty ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), conversationID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	b, err := h.repo.GetBountyByID(r.Context(), bountyID)
	if err != nil || b == nil || b.ConversationID != conversationID {
		http.Error(w, `{"error":"Bounty not found"}`, http.StatusNotFound)
		return
	}
	if b.ClaimedBy != nil {
		http.Error(w, `{"error":"Bounty already claimed"}`, http.StatusConflict)
		return
	}
	if err := h.repo.ClaimBounty(r.Context(), bountyID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Claim bounty failed")
		http.Error(w, `{"error":"Failed to claim bounty"}`, http.StatusInternalServerError)
		return
	}
	b.ClaimedBy = &user.UserID
	now := time.Now()
	b.ClaimedAt = &now
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

// EditMessageRequest is the body for editing a message.
type EditMessageRequest struct {
	Content string `json:"content"`
}

// EditMessage handles PATCH /api/v1/conversations/:id/messages/:message_id
func (h *Handler) EditMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	if msg.AuthorID != user.UserID {
		http.Error(w, `{"error":"Can only edit your own messages"}`, http.StatusForbidden)
		return
	}
	var req EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, `{"error":"Content cannot be empty"}`, http.StatusBadRequest)
		return
	}
	updated, err := h.repo.EditMessage(r.Context(), msgID, req.Content)
	if err != nil {
		h.logger.WithError(err).Error("Edit message failed")
		http.Error(w, `{"error":"Failed to edit message"}`, http.StatusInternalServerError)
		return
	}

	if h.wsHub != nil {
		h.wsHub.BroadcastMessageUpdated(updated)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteMessage handles DELETE /api/v1/conversations/:id/messages/:message_id
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	if msg.AuthorID != user.UserID {
		http.Error(w, `{"error":"Can only delete your own messages"}`, http.StatusForbidden)
		return
	}
	if err := h.repo.SoftDeleteMessage(r.Context(), msgID); err != nil {
		h.logger.WithError(err).Error("Delete message failed")
		http.Error(w, `{"error":"Failed to delete message"}`, http.StatusInternalServerError)
		return
	}

	if h.wsHub != nil {
		h.wsHub.BroadcastMessageDeleted(convID, msgID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetMessage handles GET /api/v1/conversations/:conversation_id/messages/:message_id
func (h *Handler) GetMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil {
		h.logger.WithError(err).Error("Get message failed")
		http.Error(w, `{"error":"Failed to get message"}`, http.StatusInternalServerError)
		return
	}
	if msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	attachments, _ := h.repo.ListAttachmentsForMessage(r.Context(), msgID)
	if len(attachments) > 0 {
		msg.Attachments = make([]conversations.MessageAttachment, len(attachments))
		for i, a := range attachments {
			msg.Attachments[i] = *a
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

// ListAttachments handles GET /api/v1/conversations/:conversation_id/messages/:message_id/attachments
func (h *Handler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	attachments, err := h.repo.ListAttachmentsForMessage(r.Context(), msgID)
	if err != nil {
		h.logger.WithError(err).Error("List attachments failed")
		http.Error(w, `{"error":"Failed to list attachments"}`, http.StatusInternalServerError)
		return
	}
	if attachments == nil {
		attachments = []*conversations.MessageAttachment{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"attachments": attachments})
}

// GetAttachment handles GET /api/v1/conversations/:conversation_id/messages/:message_id/attachments/:attachment_id
func (h *Handler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	attachmentID, err := uuid.Parse(vars["attachment_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid attachment ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	// Fetch all attachments for messages in this conversation, then filter by ID
	// Alternatively, use a direct lookup if available
	attachment, err := h.repo.GetAttachmentByID(r.Context(), attachmentID)
	if err != nil {
		h.logger.WithError(err).Error("Get attachment failed")
		http.Error(w, `{"error":"Failed to get attachment"}`, http.StatusInternalServerError)
		return
	}
	if attachment == nil || attachment.ConversationID != convID {
		http.Error(w, `{"error":"Attachment not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachment)
}

// SearchMessages handles GET /api/v1/conversations/search?q=...
func (h *Handler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	q := r.URL.Query().Get("q")
	if strings.TrimSpace(q) == "" {
		http.Error(w, `{"error":"q query parameter required"}`, http.StatusBadRequest)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	var convID *uuid.UUID
	if cid := r.URL.Query().Get("conversation_id"); cid != "" {
		parsed, err := uuid.Parse(cid)
		if err == nil {
			convID = &parsed
		}
	}
	results, err := h.repo.SearchMessages(r.Context(), user.UserID, q, convID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Search messages failed")
		http.Error(w, `{"error":"Search failed"}`, http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []*conversations.ConversationMessage{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"messages": results})
}

// UploadAttachment handles POST /api/v1/conversations/:id/messages/:message_id/attachments
func (h *Handler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	var req struct {
		Filename     string  `json:"filename"`
		ContentType  string  `json:"content_type"`
		SizeBytes    int64   `json:"size_bytes"`
		StorageURL   string  `json:"storage_url"`
		ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Filename == "" || req.StorageURL == "" {
		http.Error(w, `{"error":"filename and storage_url are required"}`, http.StatusBadRequest)
		return
	}
	if req.ContentType == "" {
		req.ContentType = "application/octet-stream"
	}
	att := &conversations.MessageAttachment{
		MessageID:      msgID,
		ConversationID: convID,
		UploadedBy:     user.UserID,
		Filename:       req.Filename,
		ContentType:    req.ContentType,
		SizeBytes:      req.SizeBytes,
		StorageURL:     req.StorageURL,
		ThumbnailURL:   req.ThumbnailURL,
	}
	if err := h.repo.CreateAttachment(r.Context(), att); err != nil {
		h.logger.WithError(err).Error("Create attachment failed")
		http.Error(w, `{"error":"Failed to create attachment"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(att)
}

// DeleteAttachment handles DELETE /api/v1/conversations/:id/messages/:message_id/attachments/:attachment_id
func (h *Handler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	attachmentID, err := uuid.Parse(vars["attachment_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid attachment ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	if err := h.repo.DeleteAttachment(r.Context(), attachmentID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Delete attachment failed")
		http.Error(w, `{"error":"Failed to delete attachment"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// AddReactionRequest is the body for adding a reaction.
type AddReactionRequest struct {
	Reaction string `json:"reaction"`
}

// AddReaction handles POST /api/v1/conversations/:id/messages/:message_id/reactions
func (h *Handler) AddReaction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	var req AddReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Reaction) == "" {
		http.Error(w, `{"error":"Reaction cannot be empty"}`, http.StatusBadRequest)
		return
	}
	rxn, err := h.repo.AddReaction(r.Context(), msgID, user.UserID, req.Reaction)
	if err != nil {
		h.logger.WithError(err).Error("Add reaction failed")
		http.Error(w, `{"error":"Failed to add reaction"}`, http.StatusInternalServerError)
		return
	}

	// Broadcast reaction added to all participants
	if h.wsHub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"user_id":    user.UserID,
			"reaction":   req.Reaction,
		})
		h.wsHub.BroadcastToConversation(convID, &ConvWSMessage{
			Type:    "message_reaction_added",
			Payload: payload,
		}, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rxn)
}

// RemoveReaction handles DELETE /api/v1/conversations/:id/messages/:message_id/reactions/:reaction
func (h *Handler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	reaction := vars["reaction"]
	if reaction == "" {
		http.Error(w, `{"error":"Reaction is required"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	if err := h.repo.RemoveReaction(r.Context(), msgID, user.UserID, reaction); err != nil {
		h.logger.WithError(err).Error("Remove reaction failed")
		http.Error(w, `{"error":"Failed to remove reaction"}`, http.StatusInternalServerError)
		return
	}

	// Broadcast reaction removed to all participants
	if h.wsHub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"user_id":    user.UserID,
			"reaction":   reaction,
		})
		h.wsHub.BroadcastToConversation(convID, &ConvWSMessage{
			Type:    "message_reaction_removed",
			Payload: payload,
		}, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ListReactions handles GET /api/v1/conversations/:id/messages/:message_id/reactions
func (h *Handler) ListReactions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	summary, err := h.repo.GetMessageReactionSummary(r.Context(), msgID)
	if err != nil {
		h.logger.WithError(err).Error("List reactions failed")
		http.Error(w, `{"error":"Failed to list reactions"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"reactions": summary})
}

// MarkMessageRead handles POST /api/v1/conversations/:id/messages/:message_id/read
func (h *Handler) MarkMessageRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	if err := h.repo.MarkMessageRead(r.Context(), msgID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Mark message read failed")
		http.Error(w, `{"error":"Failed to mark message read"}`, http.StatusInternalServerError)
		return
	}

	// Broadcast read receipt to all participants
	if h.wsHub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": msgID,
			"user_id":    user.UserID,
		})
		h.wsHub.BroadcastToConversation(convID, &ConvWSMessage{
			Type:    "message_read",
			Payload: payload,
		}, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetMessageReadReceipts handles GET /api/v1/conversations/:id/messages/:message_id/read-receipts
func (h *Handler) GetMessageReadReceipts(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	convID, err := uuid.Parse(vars["conversation_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid conversation ID"}`, http.StatusBadRequest)
		return
	}
	msgID, err := uuid.Parse(vars["message_id"])
	if err != nil {
		http.Error(w, `{"error":"Invalid message ID"}`, http.StatusBadRequest)
		return
	}
	ok, err := h.repo.IsParticipant(r.Context(), convID, user.UserID)
	if err != nil || !ok {
		http.Error(w, `{"error":"Not a participant"}`, http.StatusForbidden)
		return
	}
	msg, err := h.repo.GetMessageByID(r.Context(), msgID)
	if err != nil || msg == nil || msg.ConversationID != convID {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	receipts, err := h.repo.GetMessageReadReceipts(r.Context(), msgID)
	if err != nil {
		h.logger.WithError(err).Error("Get message read receipts failed")
		http.Error(w, `{"error":"Failed to get read receipts"}`, http.StatusInternalServerError)
		return
	}
	if receipts == nil {
		receipts = []*conversations.MessageRead{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"read_receipts": receipts})
}

func conversationMessagePreview(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return "(empty message)"
	}
	r := []rune(s)
	const max = 160
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func senderDisplayName(u *storage.User) string {
	if u == nil {
		return "Someone"
	}
	if strings.TrimSpace(u.Name) != "" {
		return u.Name
	}
	if u.Username != nil && strings.TrimSpace(*u.Username) != "" {
		return *u.Username
	}
	return "Someone"
}

// notifyConversationMessage creates in-app notifications for other participants (unread bell + counts).
func (h *Handler) notifyConversationMessage(ctx context.Context, conversationID, authorID uuid.UUID, m *conversations.ConversationMessage) {
	if h.notify == nil {
		return
	}
	conv, err := h.repo.GetConversationByID(ctx, conversationID)
	if err != nil || conv == nil {
		return
	}
	var participantStrs []string
	if err := json.Unmarshal(conv.ParticipantIDs, &participantStrs); err != nil {
		h.logger.WithError(err).Warn("notifyConversationMessage: parse participant_ids")
		return
	}

	var author *storage.User
	if h.users != nil {
		var err error
		author, err = h.users.GetUserByID(ctx, authorID)
		if err != nil {
			h.logger.WithError(err).Debug("notifyConversationMessage: sender lookup")
		}
	}
	fromName := senderDisplayName(author)
	title := fmt.Sprintf("New message from %s", fromName)
	body := conversationMessagePreview(m.Content)
	actionPath := fmt.Sprintf("/conversations/%s", conversationID.String())

	for _, ps := range participantStrs {
		pid, err := uuid.Parse(strings.TrimSpace(ps))
		if err != nil || pid == authorID {
			continue
		}
		_, err = h.notify.Send(ctx, notification.SendRequest{
			UserID:   pid,
			Type:     notification.TypeTeamDirectMessage,
			Category: notification.CategoryMessages,
			Title:    title,
			Body:     body,
			Data: notification.JSONMap{
				"action_url":      actionPath,
				"conversation_id": conversationID.String(),
				"message_id":      m.ID.String(),
				"from_user_id":    authorID.String(),
			},
			Channels: []string{notification.ChannelInApp},
			Priority: notification.PriorityNormal,
		})
		if err != nil {
			h.logger.WithError(err).WithField("recipient", pid).Warn("notifyConversationMessage: send failed")
		}
	}
}
