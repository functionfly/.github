package team_memory

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles team memory API requests (Team Memory Engine / Shared Brain)
type Handler struct {
	repo storage.Repository
}

// NewHandler creates a new team memory handler
func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
}

// ============================================
// Request/Response Types
// ============================================

// CreateMemoryRequest represents a request to create a team memory
type CreateMemoryRequest struct {
	MemoryType    string                 `json:"memory_type"` // decision, preference, process, client_context
	Category      string                 `json:"category,omitempty"`
	Summary       string                 `json:"summary"`
	Content       map[string]interface{} `json:"content"`
	Confidence    float64                `json:"confidence,omitempty"`
	TTLDays       int                    `json:"ttl_days,omitempty"`
	IsEncrypted   bool                   `json:"is_encrypted,omitempty"`
	EncryptedData *EncryptedData         `json:"encrypted_data,omitempty"` // Required if is_encrypted=true
}

type EncryptedData struct {
	Ciphertext string `json:"ciphertext"` // base64 encoded
	IV         string `json:"iv"`         // base64 encoded
	Tag        string `json:"tag"`        // base64 encoded
}

// UpdateMemoryRequest represents a request to update a team memory
type UpdateMemoryRequest struct {
	Category        *string                `json:"category,omitempty"`
	Summary         *string                `json:"summary,omitempty"`
	Content         map[string]interface{} `json:"content,omitempty"`
	Confidence      *float64               `json:"confidence,omitempty"`
	ImportanceScore *float64               `json:"importance_score,omitempty"`
	TTLDays         *int                   `json:"ttl_days,omitempty"`
	AutoUpdate      *bool                  `json:"auto_update_enabled,omitempty"`
}

// SearchRequest represents a memory search request
type SearchRequest struct {
	Query      string `json:"query"` // Natural language query
	MemoryType string `json:"memory_type,omitempty"`
	Category   string `json:"category,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Memories []*storage.TeamMemorySearchResult `json:"memories"`
	Total    int64                             `json:"total,omitempty"`
}

// QueryRequest represents an agent-facing natural language query
type QueryRequest struct {
	Query      string   `json:"query"` // Natural language query
	Categories []string `json:"categories,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// QueryResponse represents agent-facing query results
type QueryResponse struct {
	Context string                `json:"context"` // Formatted for LLM consumption
	Sources []*storage.TeamMemory `json:"sources"` // Full memory records
}

// ValidateRequest represents a validation request
type ValidateRequest struct {
	Validated bool `json:"validated"` // true to validate, false to unvalidate
}

// ExtractionResponse represents a memory extraction in the queue
type ExtractionResponse struct {
	ID           uuid.UUID              `json:"id"`
	MemoryType   string                 `json:"memory_type"`
	Category     string                 `json:"category,omitempty"`
	Summary      string                 `json:"summary"`
	Content      map[string]interface{} `json:"content"`
	Confidence   float64                `json:"confidence"`
	Rationale    string                 `json:"rationale"`
	SourceConvID uuid.UUID              `json:"source_conversation_id"`
	Status       string                 `json:"status"`
	CreatedAt    string                 `json:"created_at"`
}

// BulkCreateRequest represents a bulk memory creation request (e.g., from conversation)
type BulkCreateRequest struct {
	Memories []CreateMemoryRequest `json:"memories"`
}

// BulkCreateResponse represents the result of bulk creation
type BulkCreateResponse struct {
	Created []string `json:"created"`          // IDs of created memories
	Errors  []string `json:"errors,omitempty"` // Any errors encountered
}

// ============================================
// HTTP Handlers
// ============================================

// HandleCreateMemory handles POST /v1/teams/{teamId}/memories
func (h *Handler) HandleCreateMemory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req CreateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.MemoryType == "" || req.Summary == "" {
		http.Error(w, "memory_type and summary are required", http.StatusBadRequest)
		return
	}

	// Validate memory type
	validTypes := map[string]bool{"decision": true, "preference": true, "process": true, "client_context": true}
	if !validTypes[req.MemoryType] {
		http.Error(w, "Invalid memory_type. Must be: decision, preference, process, client_context", http.StatusBadRequest)
		return
	}

	// Set default confidence
	confidence := req.Confidence
	if confidence == 0 {
		confidence = 0.9
	}

	// Create memory
	memory := &storage.TeamMemory{
		TenantID:        user.TenantID,
		TeamID:          teamID,
		MemoryType:      req.MemoryType,
		Category:        &req.Category,
		Summary:         &req.Summary,
		Content:         req.Content,
		CreatedBy:       user.UserID,
		ConfidenceScore: confidence,
		IsValidated:     true, // User-created memories are pre-validated
		TTLDays:         req.TTLDays,
	}

	var result *storage.TeamMemory
	if req.IsEncrypted {
		// Handle encrypted memory creation
		if req.EncryptedData == nil {
			http.Error(w, "encrypted_data required when is_encrypted=true", http.StatusBadRequest)
			return
		}
		// Decode base64
		cipherBytes, err := decodeBase64(req.EncryptedData.Ciphertext)
		if err != nil {
			http.Error(w, "Invalid ciphertext encoding", http.StatusBadRequest)
			return
		}
		ivBytes, err := decodeBase64(req.EncryptedData.IV)
		if err != nil {
			http.Error(w, "Invalid IV encoding", http.StatusBadRequest)
			return
		}
		tagBytes, err := decodeBase64(req.EncryptedData.Tag)
		if err != nil {
			http.Error(w, "Invalid tag encoding", http.StatusBadRequest)
			return
		}

		result, err = h.repo.CreateEncryptedTeamMemory(r.Context(), memory, cipherBytes, ivBytes, tagBytes)
	} else {
		result, err = h.repo.CreateTeamMemory(r.Context(), memory)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to create team memory")
		http.Error(w, "Failed to create memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// HandleListMemories handles GET /v1/teams/{teamId}/memories
func (h *Handler) HandleListMemories(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse query parameters
	filter := storage.TeamMemoryFilter{}

	if memoryType := r.URL.Query().Get("memory_type"); memoryType != "" {
		filter.MemoryType = &memoryType
	}
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = &category
	}
	if validatedStr := r.URL.Query().Get("validated"); validatedStr != "" {
		validated := validatedStr == "true"
		filter.IsValidated = &validated
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		filter.Limit = limit
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, _ := strconv.Atoi(offsetStr)
		filter.Offset = offset
	}

	memories, total, err := h.repo.ListTeamMemories(r.Context(), user.TenantID, teamID, filter)
	if err != nil {
		logrus.WithError(err).Error("Failed to list team memories")
		http.Error(w, "Failed to list memories", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"memories": memories,
		"total":    total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetMemory handles GET /v1/teams/{teamId}/memories/{memoryId}
func (h *Handler) HandleGetMemory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}
	memoryID, err := uuid.Parse(vars["memoryId"])
	if err != nil {
		http.Error(w, "Invalid memory ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	memory, err := h.repo.GetTeamMemoryByID(r.Context(), user.TenantID, teamID, memoryID)
	if err != nil {
		if err.Error() == "team memory not found" {
			http.Error(w, "Memory not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get team memory")
		http.Error(w, "Failed to get memory", http.StatusInternalServerError)
		return
	}

	// Mark as accessed (async)
	go h.repo.MarkTeamMemoryAsAccessed(r.Context(), memoryID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memory)
}

// HandleUpdateMemory handles PUT /v1/teams/{teamId}/memories/{memoryId}
func (h *Handler) HandleUpdateMemory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}
	memoryID, err := uuid.Parse(vars["memoryId"])
	if err != nil {
		http.Error(w, "Invalid memory ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get existing memory
	memory, err := h.repo.GetTeamMemoryByID(r.Context(), user.TenantID, teamID, memoryID)
	if err != nil {
		if err.Error() == "team memory not found" {
			http.Error(w, "Memory not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get team memory for update")
		http.Error(w, "Failed to get memory", http.StatusInternalServerError)
		return
	}

	// Cannot update encrypted memories via this endpoint (client-side only)
	if memory.IsEncrypted {
		http.Error(w, "Cannot update encrypted memory via API. Use client-side operations.", http.StatusBadRequest)
		return
	}

	var req UpdateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.Category != nil {
		memory.Category = req.Category
	}
	if req.Summary != nil {
		memory.Summary = req.Summary
	}
	if req.Content != nil {
		memory.Content = req.Content
	}
	if req.Confidence != nil {
		memory.ConfidenceScore = *req.Confidence
	}
	if req.ImportanceScore != nil {
		memory.ImportanceScore = *req.ImportanceScore
	}
	if req.TTLDays != nil {
		memory.TTLDays = *req.TTLDays
	}
	if req.AutoUpdate != nil {
		memory.AutoUpdateEnabled = *req.AutoUpdate
	}

	updated, err := h.repo.UpdateTeamMemory(r.Context(), memory)
	if err != nil {
		logrus.WithError(err).Error("Failed to update team memory")
		http.Error(w, "Failed to update memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteMemory handles DELETE /v1/teams/{teamId}/memories/{memoryId}
func (h *Handler) HandleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}
	memoryID, err := uuid.Parse(vars["memoryId"])
	if err != nil {
		http.Error(w, "Invalid memory ID", http.StatusBadRequest)
		return
	}

	// Check team admin/owner for deletion
	membership, err := h.repo.GetTeamMembership(teamID, user.UserID)
	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	if membership.Role != auth.TeamRoleOwner && membership.Role != auth.TeamRoleAdmin {
		http.Error(w, "Only team owners or admins can delete memories", http.StatusForbidden)
		return
	}

	if err := h.repo.DeleteTeamMemory(r.Context(), user.TenantID, teamID, memoryID); err != nil {
		if err.Error() == "team memory not found" {
			http.Error(w, "Memory not found", http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to delete team memory")
		http.Error(w, "Failed to delete memory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleSearchMemories handles POST /v1/teams/{teamId}/memories/search
func (h *Handler) HandleSearchMemories(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	results, err := h.repo.SearchTeamMemories(r.Context(), user.TenantID, teamID, req.Query, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to search team memories")
		http.Error(w, "Failed to search memories", http.StatusInternalServerError)
		return
	}

	// Mark results as accessed (async)
	for _, result := range results {
		go h.repo.MarkTeamMemoryAsAccessed(r.Context(), result.ID)
	}

	response := SearchResponse{
		Memories: results,
		Total:    int64(len(results)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleQueryMemories handles POST /v1/teams/{teamId}/memories/query (agent-facing)
func (h *Handler) HandleQueryMemories(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	limit := req.Limit
	if limit == 0 {
		limit = 5
	}
	if limit > 10 {
		limit = 10
	}

	// Search for relevant memories
	results, err := h.repo.SearchTeamMemories(r.Context(), user.TenantID, teamID, req.Query, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to query team memories")
		http.Error(w, "Failed to query memories", http.StatusInternalServerError)
		return
	}

	// Build formatted context for LLM consumption
	context := h.buildAgentContext(results, req.Query)

	// Extract sources
	sources := make([]*storage.TeamMemory, 0, len(results))
	for _, result := range results {
		sources = append(sources, &result.TeamMemory)
		go h.repo.MarkTeamMemoryAsAccessed(r.Context(), result.ID)
	}

	response := QueryResponse{
		Context: context,
		Sources: sources,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleValidateMemory handles POST /v1/teams/{teamId}/memories/{memoryId}/validate
func (h *Handler) HandleValidateMemory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}
	memoryID, err := uuid.Parse(vars["memoryId"])
	if err != nil {
		http.Error(w, "Invalid memory ID", http.StatusBadRequest)
		return
	}

	// Check team admin/owner for validation
	membership, err := h.repo.GetTeamMembership(teamID, user.UserID)
	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	if membership.Role != auth.TeamRoleOwner && membership.Role != auth.TeamRoleAdmin {
		http.Error(w, "Only team owners or admins can validate memories", http.StatusForbidden)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Validated {
		err = h.repo.ValidateTeamMemory(r.Context(), memoryID, user.UserID)
	} else {
		// Unvalidate - set is_validated to false
		memory, err := h.repo.GetTeamMemoryByID(r.Context(), user.TenantID, teamID, memoryID)
		if err != nil {
			http.Error(w, "Memory not found", http.StatusNotFound)
			return
		}
		memory.IsValidated = false
		memory.ValidatedBy = nil
		_, err = h.repo.UpdateTeamMemory(r.Context(), memory)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to validate team memory")
		http.Error(w, "Failed to validate memory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleListExtractions handles GET /v1/teams/{teamId}/memories/extractions
func (h *Handler) HandleListExtractions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	extractions, err := h.repo.GetMemoryExtractionsByTeam(r.Context(), teamID, status, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to list memory extractions")
		http.Error(w, "Failed to list extractions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(extractions)
}

// HandleApproveExtraction handles POST /v1/teams/{teamId}/memories/extractions/{extractionId}/approve
func (h *Handler) HandleApproveExtraction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}
	extractionID, err := uuid.Parse(vars["extractionId"])
	if err != nil {
		http.Error(w, "Invalid extraction ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	memory, err := h.repo.ApproveMemoryExtraction(r.Context(), extractionID, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to approve extraction")
		http.Error(w, "Failed to approve extraction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memory)
}

// HandleRejectExtraction handles POST /v1/teams/{teamId}/memories/extractions/{extractionId}/reject
func (h *Handler) HandleRejectExtraction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["teamId"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}
	extractionID, err := uuid.Parse(vars["extractionId"])
	if err != nil {
		http.Error(w, "Invalid extraction ID", http.StatusBadRequest)
		return
	}

	// Check team membership
	if !h.isTeamMember(user.UserID, teamID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var req struct {
		Reason string `json:"reason,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.repo.RejectMemoryExtraction(r.Context(), extractionID, user.UserID, req.Reason); err != nil {
		logrus.WithError(err).Error("Failed to reject extraction")
		http.Error(w, "Failed to reject extraction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// Helper Methods
// ============================================

func (h *Handler) isTeamMember(userID, teamID uuid.UUID) bool {
	_, err := h.repo.GetTeamMembership(teamID, userID)
	return err == nil
}

func (h *Handler) buildAgentContext(results []*storage.TeamMemorySearchResult, query string) string {
	if len(results) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Relevant Team Knowledge\n")

	for i, result := range results {
		validity := ""
		if !result.IsValidated {
			validity = " [UNVALIDATED]"
		}

		parts = append(parts, fmt.Sprintf("### %d. %s (%s)%s\n%s\n",
			i+1,
			*result.Summary,
			result.MemoryType,
			validity,
			formatContentForAgent(result.Content, result.MemoryType),
		))
	}

	return strings.Join(parts, "\n")
}

func formatContentForAgent(content map[string]interface{}, memoryType string) string {
	if content == nil {
		return ""
	}

	switch memoryType {
	case "decision":
		if rationale, ok := content["rationale"].(string); ok {
			return rationale
		}
	case "preference":
		subject, _ := content["subject"].(string)
		value, _ := content["value"].(string)
		ctx, _ := content["context"].(string)
		if ctx != "" {
			return fmt.Sprintf("%s: %s (%s)", subject, value, ctx)
		}
		return fmt.Sprintf("%s: %s", subject, value)
	case "process":
		if name, ok := content["name"].(string); ok {
			return name
		}
	case "client_context":
		if notes, ok := content["notes"].(string); ok {
			return notes
		}
	}

	return fmt.Sprintf("%v", content)
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
