package agent_memory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	statestorage "github.com/functionfly/functionfly/internal/storage/state"
)

// AuditAction represents the type of audit action performed on agent memory
type AuditAction string

const (
	AuditActionCreate      AuditAction = "create"
	AuditActionRead        AuditAction = "read"
	AuditActionUpdate      AuditAction = "update"
	AuditActionDelete      AuditAction = "delete"
	AuditActionSearch      AuditAction = "search"
	AuditActionMarkAccessed AuditAction = "mark_accessed"
	AuditActionRebuildIndex AuditAction = "rebuild_index"
)

// ============================================
// Handler and Repository Types
// ============================================

// AgentMemoryHandler handles HTTP requests for agent memory management
type AgentMemoryHandler struct {
	db *gorm.DB
}

const (
	// MaxEmbeddingDimensions is the maximum allowed embedding vector dimensions
	MaxEmbeddingDimensions = 1536
)

// getClientIP extracts the client IP address from the request, considering proxies
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// logAuditEvent logs an audit event for agent memory operations
func logAuditEvent(r *http.Request, claims *auth.Claims, action AuditAction, memoryID *uuid.UUID, metadata map[string]interface{}) {
	entry := logrus.WithFields(logrus.Fields{
		"service":     "agent_memory",
		"action":      action,
		"tenant_id":   claims.TenantID,
		"user_id":     claims.UserID,
		"user_email":  claims.Email,
		"ip_address":  getClientIP(r),
		"user_agent":  r.UserAgent(),
		"request_id":  r.Header.Get("X-Request-ID"),
	})
	if memoryID != nil {
		entry = entry.WithField("memory_id", *memoryID)
	}
	if metadata != nil {
		entry = entry.WithFields(metadata)
	}
	entry.Info("agent_memory audit event")
}

// hasPermission checks if the user claims contain the required permission
func hasPermission(claims *auth.Claims, permission string) bool {
	if claims == nil {
		return false
	}
	// Allow super_admin and admin roles to bypass permission checks
	if claims.Role == auth.RoleSuperAdmin || claims.Role == auth.RoleAdmin {
		return true
	}
	if claims.Permissions == nil {
		return false
	}
	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// checkMemoryRead checks if the user has memory.read permission
func checkMemoryRead(claims *auth.Claims) bool {
	return hasPermission(claims, auth.PermMemoryRead)
}

// checkMemoryWrite checks if the user has memory.write permission
func checkMemoryWrite(claims *auth.Claims) bool {
	return hasPermission(claims, auth.PermMemoryWrite)
}

// validateEmbedding validates the embedding vector length
func validateEmbedding(embedding []float32) error {
	if len(embedding) > MaxEmbeddingDimensions {
		return fmt.Errorf("embedding dimensions (%d) exceed maximum allowed (%d)", len(embedding), MaxEmbeddingDimensions)
	}
	return nil
}

// NewHandler creates a new AgentMemoryHandler
func NewHandler(db *gorm.DB) *AgentMemoryHandler {
	return &AgentMemoryHandler{db: db}
}

// ============================================
// Request/Response Types
// ============================================

// CreateMemoryRequest represents the request body for creating a memory
type CreateMemoryRequest struct {
	AgentID         string                 `json:"agent_id"`
	MemoryType      string                 `json:"memory_type"`
	Content         string                 `json:"content"`
	StructuredData  map[string]interface{} `json:"structured_data,omitempty"`
	Embedding       []float32              `json:"embedding,omitempty"`
	ImportanceScore float32                `json:"importance_score,omitempty"`
	TTLDays         int                    `json:"ttl_days,omitempty"`
}

// UpdateMemoryRequest represents the request body for updating a memory
type UpdateMemoryRequest struct {
	Content         string                 `json:"content,omitempty"`
	StructuredData  map[string]interface{} `json:"structured_data,omitempty"`
	Embedding       []float32              `json:"embedding,omitempty"`
	ImportanceScore float32                `json:"importance_score,omitempty"`
}

// SearchMemoryRequest represents the request body for searching memories
type SearchMemoryRequest struct {
	AgentID    string    `json:"agent_id,omitempty"`
	Query      string    `json:"query"`
	Embedding  []float32 `json:"embedding,omitempty"`
	MemoryType string    `json:"memory_type,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Threshold  float32   `json:"threshold,omitempty"`
}

// RebuildIndexRequest represents the request body for rebuilding the search index
type RebuildIndexRequest struct {
	AgentID    string `json:"agent_id,omitempty"`
	MemoryType string `json:"memory_type,omitempty"`
}

// ListMemoriesResponse represents the response for listing memories
type ListMemoriesResponse struct {
	Memories []*statestorage.AgentMemory `json:"memories"`
	Total    int64                       `json:"total"`
	Limit    int                         `json:"limit"`
	Offset   int                         `json:"offset"`
}

// SearchMemoriesResponse represents the response for searching memories
type SearchMemoriesResponse struct {
	Memories []*statestorage.AgentMemory `json:"memories"`
	Count    int                         `json:"count"`
}

// RebuildIndexResponse represents the response for rebuilding the index
type RebuildIndexResponse struct {
	Success     bool   `json:"success"`
	IndexID     string `json:"index_id,omitempty"`
	MemoryCount int    `json:"memory_count"`
	Message     string `json:"message"`
}

// ============================================
// HTTP Handlers
// ============================================

// HandleCreateMemory handles POST /agent-memories - Create a new memory
func (h *AgentMemoryHandler) HandleCreateMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.write permission
	if !checkMemoryWrite(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	var req CreateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	if req.AgentID == "" || req.MemoryType == "" {
		apierror.WriteError(w, apierror.NewBadRequest("agent_id and memory_type are required"))
		return
	}

	// Validate embedding if provided
	if len(req.Embedding) > 0 {
		if err := validateEmbedding(req.Embedding); err != nil {
			apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeValidation, "invalid embedding")
			return
		}
	}

	memory := &statestorage.AgentMemory{
		TenantID:        claims.TenantID,
		AgentID:         req.AgentID,
		MemoryType:      req.MemoryType,
		Content:         strPtr(req.Content),
		StructuredData:  req.StructuredData,
		Embedding:       req.Embedding,
		ImportanceScore: req.ImportanceScore,
		TTLDays:         req.TTLDays,
	}

	if req.TTLDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, req.TTLDays)
		memory.ExpiresAt = &expiresAt
	}

	created, err := h.createMemory(r.Context(), memory)
	if err != nil {
		logrus.Errorf("failed to create memory: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to create memory"))
		return
	}

	// Log audit event
	logAuditEvent(r, claims, AuditActionCreate, &created.ID, map[string]interface{}{
		"agent_id":     created.AgentID,
		"memory_type":  created.MemoryType,
		"importance":   created.ImportanceScore,
		"has_embedding": len(created.Embedding) > 0,
		"ttl_days":     created.TTLDays,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleListMemories handles GET /agent-memories - List memories with optional filters
func (h *AgentMemoryHandler) HandleListMemories(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.read permission
	if !checkMemoryRead(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	// Parse query parameters
	agentID := r.URL.Query().Get("agent_id")
	memoryType := r.URL.Query().Get("memory_type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var memories []*statestorage.AgentMemory
	var total int64
	var err error

	// Query based on filters
	if agentID != "" && memoryType != "" {
		memories, total, err = h.listMemoriesByAgentAndType(r.Context(), claims.TenantID, agentID, memoryType, limit, offset)
	} else if agentID != "" {
		memories, total, err = h.listMemoriesByAgent(r.Context(), claims.TenantID, agentID, limit, offset)
	} else {
		memories, total, err = h.listMemoriesByTenant(r.Context(), claims.TenantID, limit, offset)
	}

	if err != nil {
		logrus.Errorf("failed to list memories: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to list memories"))
		return
	}

	// Log audit event for list operation
	logAuditEvent(r, claims, AuditActionRead, nil, map[string]interface{}{
		"operation": "list",
		"agent_id":  agentID,
		"count":     len(memories),
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListMemoriesResponse{
		Memories: memories,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// HandleGetMemory handles GET /agent-memories/{id} - Get a specific memory
func (h *AgentMemoryHandler) HandleGetMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.read permission
	if !checkMemoryRead(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	memory, err := h.getMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("memory not found"))
		return
	}

	// Log audit event
	logAuditEvent(r, claims, AuditActionRead, &memoryUUID, map[string]interface{}{
		"agent_id":    memory.AgentID,
		"memory_type": memory.MemoryType,
		"operation":   "get",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memory)
}

// HandleUpdateMemory handles PATCH /agent-memories/{id} - Update an existing memory
func (h *AgentMemoryHandler) HandleUpdateMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.write permission
	if !checkMemoryWrite(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	var req UpdateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	// Validate embedding if provided
	if req.Embedding != nil && len(req.Embedding) > 0 {
		if err := validateEmbedding(req.Embedding); err != nil {
			apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeValidation, "invalid embedding")
			return
		}
	}

	// Get existing memory
	memory, err := h.getMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("memory not found"))
		return
	}

	// Update fields if provided
	if req.Content != "" {
		memory.Content = &req.Content
	}
	if req.StructuredData != nil {
		memory.StructuredData = req.StructuredData
	}
	if req.Embedding != nil {
		memory.Embedding = req.Embedding
	}
	if req.ImportanceScore > 0 {
		memory.ImportanceScore = req.ImportanceScore
	}

	updated, err := h.updateMemory(r.Context(), memory)
	if err != nil {
		logrus.Errorf("failed to update memory: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to update memory"))
		return
	}

	// Log audit event
	logAuditEvent(r, claims, AuditActionUpdate, &memoryUUID, map[string]interface{}{
		"agent_id":     updated.AgentID,
		"memory_type":  updated.MemoryType,
		"importance":   updated.ImportanceScore,
		"has_embedding": len(updated.Embedding) > 0,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteMemory handles DELETE /agent-memories/{id} - Delete a memory
func (h *AgentMemoryHandler) HandleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.write permission
	if !checkMemoryWrite(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	err = h.deleteMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		logrus.Errorf("failed to delete memory: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to delete memory"))
		return
	}

	// Log audit event
	logAuditEvent(r, claims, AuditActionDelete, &memoryUUID, map[string]interface{}{
		"agent_id": memoryID,
	})

	w.WriteHeader(http.StatusNoContent)
}

// HandleMarkAccessed handles POST /agent-memories/{id}/accessed - Mark memory as accessed
func (h *AgentMemoryHandler) HandleMarkAccessed(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.read permission (marking as accessed is a read operation)
	if !checkMemoryRead(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	// Verify memory exists and belongs to tenant
	_, err = h.getMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("memory not found"))
		return
	}

	// Update access count
	err = h.markMemoryAccessed(r.Context(), memoryUUID)
	if err != nil {
		logrus.Errorf("failed to mark memory accessed: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to update access count"))
		return
	}

	// Log audit event
	logAuditEvent(r, claims, AuditActionMarkAccessed, &memoryUUID, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memory_id": memoryID,
		"accessed":  true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleSearchMemories handles POST /agent-memories/search - Search memories using vector similarity
func (h *AgentMemoryHandler) HandleSearchMemories(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.read permission
	if !checkMemoryRead(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	var req SearchMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	// Validate embedding if provided
	if len(req.Embedding) > 0 {
		if err := validateEmbedding(req.Embedding); err != nil {
			apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeValidation, "invalid embedding")
			return
		}
	}

	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// If embedding is provided, use vector search
	// Otherwise, fall back to keyword/score-based search
	var memories []*statestorage.AgentMemory
	var err error

	if len(req.Embedding) > 0 {
		memories, err = h.searchMemoriesByVector(r.Context(), claims.TenantID, req.AgentID, req.MemoryType, req.Embedding, req.Limit, req.Threshold)
	} else {
		memories, err = h.searchMemoriesByFilter(r.Context(), claims.TenantID, req.AgentID, req.MemoryType, req.Limit, req.Threshold)
	}

	if err != nil {
		logrus.Errorf("failed to search memories: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to search memories"))
		return
	}

	// Log audit event
	searchType := "filter"
	if len(req.Embedding) > 0 {
		searchType = "vector"
	}
	logAuditEvent(r, claims, AuditActionSearch, nil, map[string]interface{}{
		"search_type":  searchType,
		"agent_id":     req.AgentID,
		"memory_type":  req.MemoryType,
		"result_count": len(memories),
		"limit":        req.Limit,
		"threshold":    req.Threshold,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SearchMemoriesResponse{
		Memories: memories,
		Count:    len(memories),
	})
}

// HandleRebuildIndex handles POST /agent-memories/index - Rebuild the search index
func (h *AgentMemoryHandler) HandleRebuildIndex(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	// Check memory.write permission (rebuilding index is a write operation)
	if !checkMemoryWrite(claims) {
		apierror.WriteError(w, apierror.NewForbidden("forbidden: insufficient permissions"))
		return
	}

	var req RebuildIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	if req.AgentID == "" {
		apierror.WriteError(w, apierror.NewBadRequest("agent_id is required"))
		return
	}
	if req.MemoryType == "" {
		apierror.WriteError(w, apierror.NewBadRequest("memory_type is required"))
		return
	}

	// For now, we just update the index record with the current timestamp
	// In a production system, this would rebuild the actual pgvector index
	index, err := h.rebuildIndex(r.Context(), claims.TenantID, req.AgentID, req.MemoryType)
	if err != nil {
		logrus.Errorf("failed to rebuild index: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to rebuild index"))
		return
	}

	// Log audit event
	logAuditEvent(r, claims, AuditActionRebuildIndex, nil, map[string]interface{}{
		"agent_id":     req.AgentID,
		"memory_type":  req.MemoryType,
		"memory_count": index.MemoryCount,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RebuildIndexResponse{
		Success:     true,
		IndexID:     index.ID.String(),
		MemoryCount: index.MemoryCount,
		Message:     "Index rebuilt successfully",
	})
}

// ============================================
// Repository Methods
// ============================================

func (h *AgentMemoryHandler) createMemory(ctx context.Context, memory *statestorage.AgentMemory) (*statestorage.AgentMemory, error) {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}
	if memory.TTLDays == 0 {
		memory.TTLDays = 30
	}
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()

	err := h.db.WithContext(ctx).Create(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}
	return memory, nil
}

func (h *AgentMemoryHandler) getMemory(ctx context.Context, tenantID, memoryID uuid.UUID) (*statestorage.AgentMemory, error) {
	var memory statestorage.AgentMemory
	err := h.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", memoryID, tenantID).
		First(&memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}
	return &memory, nil
}

func (h *AgentMemoryHandler) deleteMemory(ctx context.Context, tenantID, memoryID uuid.UUID) error {
	result := h.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", memoryID, tenantID).
		Delete(&statestorage.AgentMemory{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete memory: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("memory not found")
	}
	return nil
}

func (h *AgentMemoryHandler) updateMemory(ctx context.Context, memory *statestorage.AgentMemory) (*statestorage.AgentMemory, error) {
	memory.UpdatedAt = time.Now()
	err := h.db.WithContext(ctx).Save(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}
	return memory, nil
}

func (h *AgentMemoryHandler) markMemoryAccessed(ctx context.Context, memoryID uuid.UUID) error {
	err := h.db.WithContext(ctx).Model(&statestorage.AgentMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"access_count":     gorm.Expr("access_count + 1"),
			"last_accessed_at": time.Now(),
			"updated_at":       time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to mark memory accessed: %w", err)
	}
	return nil
}

func (h *AgentMemoryHandler) listMemoriesByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit, offset int) ([]*statestorage.AgentMemory, int64, error) {
	var memories []*statestorage.AgentMemory
	var total int64

	err := h.db.WithContext(ctx).Model(&statestorage.AgentMemory{}).
		Where("tenant_id = ? AND agent_id = ?", tenantID, agentID).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	err = h.db.WithContext(ctx).
		Where("tenant_id = ? AND agent_id = ?", tenantID, agentID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memories: %w", err)
	}

	return memories, total, nil
}

func (h *AgentMemoryHandler) listMemoriesByAgentAndType(ctx context.Context, tenantID uuid.UUID, agentID, memoryType string, limit, offset int) ([]*statestorage.AgentMemory, int64, error) {
	var memories []*statestorage.AgentMemory
	var total int64

	err := h.db.WithContext(ctx).Model(&statestorage.AgentMemory{}).
		Where("tenant_id = ? AND agent_id = ? AND memory_type = ?", tenantID, agentID, memoryType).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	err = h.db.WithContext(ctx).
		Where("tenant_id = ? AND agent_id = ? AND memory_type = ?", tenantID, agentID, memoryType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memories: %w", err)
	}

	return memories, total, nil
}

func (h *AgentMemoryHandler) listMemoriesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*statestorage.AgentMemory, int64, error) {
	var memories []*statestorage.AgentMemory
	var total int64

	err := h.db.WithContext(ctx).Model(&statestorage.AgentMemory{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	err = h.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memories: %w", err)
	}

	return memories, total, nil
}

func (h *AgentMemoryHandler) searchMemoriesByVector(ctx context.Context, tenantID uuid.UUID, agentID, memoryType string, embedding []float32, limit int, threshold float32) ([]*statestorage.AgentMemory, error) {
	var memories []*statestorage.AgentMemory

	// Build the base query conditions using GORM's parameterized queries
	query := h.db.WithContext(ctx).
		Where("tenant_id = ? AND (expires_at IS NULL OR expires_at > NOW()) AND embedding IS NOT NULL", tenantID)

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}
	if threshold > 0 {
		query = query.Where("importance_score >= ?", threshold)
	}

	// Use GORM's raw query with pgvector's cosine distance operator (<=>)
	// The embedding is passed as a pq.Array which GORM properly parameterizes
	// This prevents SQL injection since the vector values are never interpolated directly
	sqlQuery := `SELECT id, tenant_id, agent_id, memory_type, content, structured_data,
				 embedding, importance_score, access_count, last_accessed_at,
				 ttl_days, expires_at, source_event_id, created_at, updated_at
		  FROM agent_memories
		  WHERE tenant_id = ? AND (expires_at IS NULL OR expires_at > NOW()) AND embedding IS NOT NULL`

	args := []interface{}{tenantID}

	if agentID != "" {
		sqlQuery += " AND agent_id = ?"
		args = append(args, agentID)
	}
	if memoryType != "" {
		sqlQuery += " AND memory_type = ?"
		args = append(args, memoryType)
	}
	if threshold > 0 {
		sqlQuery += " AND importance_score >= ?"
		args = append(args, threshold)
	}

	// Use pgvector's <=> (cosine distance) operator for similarity search
	// Pass embedding as a properly parameterized pq.Array to prevent SQL injection
	sqlQuery += " ORDER BY embedding <=> ? LIMIT ?"
	args = append(args, pq.Array(embedding), limit)

	err := h.db.WithContext(ctx).Raw(sqlQuery, args...).Scan(&memories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	return memories, nil
}

func (h *AgentMemoryHandler) searchMemoriesByFilter(ctx context.Context, tenantID uuid.UUID, agentID, memoryType string, limit int, threshold float32) ([]*statestorage.AgentMemory, error) {
	var memories []*statestorage.AgentMemory

	query := h.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("expires_at IS NULL OR expires_at > NOW()")

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}
	if threshold > 0 {
		query = query.Where("importance_score >= ?", threshold)
	}

	err := query.
		Order("importance_score DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	return memories, nil
}

func (h *AgentMemoryHandler) rebuildIndex(ctx context.Context, tenantID uuid.UUID, agentID, memoryType string) (*storage.AgentMemoryIndex, error) {
	// Don't create index with empty memory_type - it's an enum and '' is invalid
	if memoryType == "" {
		return nil, fmt.Errorf("memory_type is required for index operation")
	}
	// Don't create index with empty agent_id - it's NOT NULL in the schema
	if agentID == "" {
		return nil, fmt.Errorf("agent_id is required for index operation")
	}

	// Count memories for the index
	var count int64
	query := h.db.WithContext(ctx).Model(&statestorage.AgentMemory{}).
		Where("tenant_id = ?", tenantID).
		Where("expires_at IS NULL OR expires_at > NOW()")

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	if memoryType != "" {
		query = query.Where("memory_type = ?", memoryType)
	}

	err := query.Count(&count).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count memories: %w", err)
	}

	// Find or create the index record
	var index storage.AgentMemoryIndex
	err = h.db.WithContext(ctx).
		Where("tenant_id = ? AND agent_id = COALESCE(?, agent_id) AND memory_type = COALESCE(?, memory_type)", tenantID, strPtr(agentID), strPtr(memoryType)).
		First(&index).Error

	if err == gorm.ErrRecordNotFound {
		// Create new index record
		index = storage.AgentMemoryIndex{
			ID:               uuid.New(),
			TenantID:         tenantID,
			AgentID:          agentID,
			MemoryType:       memoryType,
			Dimension:        1536,
			SimilarityMetric: "cosine",
			MemoryCount:      int(count),
			LastIndexedAt:    timePtr(time.Now()),
			CreatedAt:        time.Now(),
		}
		err = h.db.WithContext(ctx).Create(&index).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	} else {
		// Update existing index
		index.LastIndexedAt = timePtr(time.Now())
		index.MemoryCount = int(count)
		err = h.db.WithContext(ctx).Save(&index).Error
		if err != nil {
			return nil, fmt.Errorf("failed to update index: %w", err)
		}
	}

	return &index, nil
}

// ============================================
// Helper Functions
// ============================================

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
