package agent_memory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	statestorage "github.com/functionfly/functionfly/internal/storage/state"
)

// ============================================
// Handler and Repository Types
// ============================================

// AgentMemoryHandler handles HTTP requests for agent memory management
type AgentMemoryHandler struct {
	db *gorm.DB
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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentID == "" || req.MemoryType == "" {
		http.Error(w, "agent_id and memory_type are required", http.StatusBadRequest)
		return
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
		http.Error(w, "failed to create memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// HandleListMemories handles GET /agent-memories - List memories with optional filters
func (h *AgentMemoryHandler) HandleListMemories(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "failed to list memories", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		http.Error(w, "invalid memory ID", http.StatusBadRequest)
		return
	}

	memory, err := h.getMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		http.Error(w, "memory not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memory)
}

// HandleUpdateMemory handles PATCH /agent-memories/{id} - Update an existing memory
func (h *AgentMemoryHandler) HandleUpdateMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		http.Error(w, "invalid memory ID", http.StatusBadRequest)
		return
	}

	var req UpdateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing memory
	memory, err := h.getMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		http.Error(w, "memory not found", http.StatusNotFound)
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
		http.Error(w, "failed to update memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteMemory handles DELETE /agent-memories/{id} - Delete a memory
func (h *AgentMemoryHandler) HandleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		http.Error(w, "invalid memory ID", http.StatusBadRequest)
		return
	}

	err = h.deleteMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		logrus.Errorf("failed to delete memory: %v", err)
		http.Error(w, "failed to delete memory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleMarkAccessed handles POST /agent-memories/{id}/accessed - Mark memory as accessed
func (h *AgentMemoryHandler) HandleMarkAccessed(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		http.Error(w, "invalid memory ID", http.StatusBadRequest)
		return
	}

	// Verify memory exists and belongs to tenant
	_, err = h.getMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		http.Error(w, "memory not found", http.StatusNotFound)
		return
	}

	// Update access count
	err = h.markMemoryAccessed(r.Context(), memoryUUID)
	if err != nil {
		logrus.Errorf("failed to mark memory accessed: %v", err)
		http.Error(w, "failed to update access count", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req SearchMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
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
		http.Error(w, "failed to search memories", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req RebuildIndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentID == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}
	if req.MemoryType == "" {
		http.Error(w, "memory_type is required", http.StatusBadRequest)
		return
	}

	// For now, we just update the index record with the current timestamp
	// In a production system, this would rebuild the actual pgvector index
	index, err := h.rebuildIndex(r.Context(), claims.TenantID, req.AgentID, req.MemoryType)
	if err != nil {
		logrus.Errorf("failed to rebuild index: %v", err)
		http.Error(w, "failed to rebuild index", http.StatusInternalServerError)
		return
	}

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
	// Convert float32 slice to PostgreSQL vector format for pgvector
	vectorStr := "["
	for i, v := range embedding {
		if i > 0 {
			vectorStr += ","
		}
		vectorStr += fmt.Sprintf("%.6f", v)
	}
	vectorStr += "]"

	var memories []*statestorage.AgentMemory

	// Build the base query conditions
	baseConditions := []interface{}{}
	whereClause := "tenant_id = ? AND (expires_at IS NULL OR expires_at > NOW()) AND embedding IS NOT NULL"
	baseConditions = append(baseConditions, tenantID)

	if agentID != "" {
		whereClause += " AND agent_id = ?"
		baseConditions = append(baseConditions, agentID)
	}
	if memoryType != "" {
		whereClause += " AND memory_type = ?"
		baseConditions = append(baseConditions, memoryType)
	}
	if threshold > 0 {
		whereClause += " AND importance_score >= ?"
		baseConditions = append(baseConditions, threshold)
	}

	// Build the full SQL query with vector similarity ordering
	sqlQuery := fmt.Sprintf(`
		SELECT id, tenant_id, agent_id, memory_type, content, structured_data,
				   embedding, importance_score, access_count, last_accessed_at,
				   ttl_days, expires_at, source_event_id, created_at, updated_at
		FROM agent_memories
		WHERE %s
		ORDER BY embedding <=> '%s'
		LIMIT ?
	`, whereClause, vectorStr)

	baseConditions = append(baseConditions, limit)

	err := h.db.WithContext(ctx).Raw(sqlQuery, baseConditions...).Scan(&memories).Error
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
