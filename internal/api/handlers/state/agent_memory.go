package state

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
	storagestate "github.com/functionfly/functionfly/internal/storage/state"
	"github.com/functionfly/functionfly/internal/apierror"
)

type AgentMemory = storagestate.AgentMemory

type AgentMemoryHandler struct {
	memoryRepo *AgentMemoryRepository
}

func NewAgentMemoryHandler(repo *AgentMemoryRepository) *AgentMemoryHandler {
	return &AgentMemoryHandler{memoryRepo: repo}
}

type CreateMemoryRequest struct {
	AgentID         string                 `json:"agent_id"`
	MemoryType      string                 `json:"memory_type"`
	Content         string                 `json:"content"`
	StructuredData  map[string]interface{} `json:"structured_data,omitempty"`
	Embedding       []float32              `json:"embedding,omitempty"`
	ImportanceScore float32                `json:"importance_score,omitempty"`
	TTLDays         int                    `json:"ttl_days,omitempty"`
}

type UpdateMemoryRequest struct {
	Content         string                 `json:"content,omitempty"`
	StructuredData  map[string]interface{} `json:"structured_data,omitempty"`
	Embedding       []float32              `json:"embedding,omitempty"`
	ImportanceScore float32                `json:"importance_score,omitempty"`
}

type SearchMemoryRequest struct {
	AgentID    string    `json:"agent_id,omitempty"`
	Query      string    `json:"query"`
	Embedding  []float32 `json:"embedding,omitempty"`
	MemoryType string    `json:"memory_type,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Threshold  float32   `json:"threshold,omitempty"`
}

func (h *AgentMemoryHandler) HandleCreateMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
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

	memory := &storagestate.AgentMemory{
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

	created, err := h.memoryRepo.CreateMemory(r.Context(), memory)
	if err != nil {
		logrus.Errorf("failed to create memory: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to create memory"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

func (h *AgentMemoryHandler) HandleGetMemory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	memory, err := h.memoryRepo.GetMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("memory not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memory)
}

func (h *AgentMemoryHandler) HandleUpdateMemory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	var req UpdateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	memory, err := h.memoryRepo.GetMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("memory not found"))
		return
	}

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

	updated, err := h.memoryRepo.UpdateMemory(r.Context(), memory)
	if err != nil {
		logrus.Errorf("failed to update memory: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to update memory"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *AgentMemoryHandler) HandleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	memoryID := vars["id"]

	memoryUUID, err := uuid.Parse(memoryID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid memory ID"))
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	err = h.memoryRepo.DeleteMemory(r.Context(), claims.TenantID, memoryUUID)
	if err != nil {
		logrus.Errorf("failed to delete memory: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to delete memory"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AgentMemoryHandler) HandleListMemories(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	memoryType := r.URL.Query().Get("memory_type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 20
	}

	var memories []*storagestate.AgentMemory
	var total int64
	var err error

	if agentID != "" && memoryType != "" {
		memories, total, err = h.memoryRepo.ListMemoriesByAgentAndType(r.Context(), claims.TenantID, agentID, memoryType, limit, offset)
	} else if agentID != "" {
		memories, total, err = h.memoryRepo.ListMemoriesByAgent(r.Context(), claims.TenantID, agentID, limit, offset)
	} else {
		memories, total, err = h.memoryRepo.ListMemoriesByTenant(r.Context(), claims.TenantID, limit, offset)
	}

	if err != nil {
		logrus.Errorf("failed to list memories: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to list memories"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memories": memories,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *AgentMemoryHandler) HandleSearchMemories(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("unauthorized"))
		return
	}

	var req SearchMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid request body"))
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	memories, err := h.memoryRepo.SearchMemories(r.Context(), claims.TenantID, req.AgentID, req.MemoryType, req.Embedding, req.Limit, req.Threshold)
	if err != nil {
		logrus.Errorf("failed to search memories: %v", err)
		apierror.WriteError(w, apierror.NewInternal("failed to search memories"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"memories": memories,
		"count":    len(memories),
	})
}

type AgentMemoryRepository struct {
	db *gorm.DB
}

func NewAgentMemoryRepository(db *gorm.DB) *AgentMemoryRepository {
	return &AgentMemoryRepository{db: db}
}

func (r *AgentMemoryRepository) CreateMemory(ctx context.Context, memory *storagestate.AgentMemory) (*storagestate.AgentMemory, error) {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}
	if memory.TTLDays == 0 {
		memory.TTLDays = 30
	}
	memory.CreatedAt = time.Now()
	memory.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Create(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}
	return memory, nil
}

func (r *AgentMemoryRepository) GetMemory(ctx context.Context, tenantID, memoryID uuid.UUID) (*storagestate.AgentMemory, error) {
	var memory storagestate.AgentMemory
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", memoryID, tenantID).
		First(&memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}
	return &memory, nil
}

func (r *AgentMemoryRepository) UpdateMemory(ctx context.Context, memory *storagestate.AgentMemory) (*storagestate.AgentMemory, error) {
	memory.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}
	return memory, nil
}

func (r *AgentMemoryRepository) DeleteMemory(ctx context.Context, tenantID, memoryID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", memoryID, tenantID).
		Delete(&storagestate.AgentMemory{})
	return result.Error
}

func (r *AgentMemoryRepository) ListMemoriesByAgent(ctx context.Context, tenantID uuid.UUID, agentID string, limit, offset int) ([]*storagestate.AgentMemory, int64, error) {
	var memories []*storagestate.AgentMemory
	var total int64

	err := r.db.WithContext(ctx).Model(&storagestate.AgentMemory{}).
		Where("tenant_id = ? AND agent_id = ?", tenantID, agentID).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	err = r.db.WithContext(ctx).
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

func (r *AgentMemoryRepository) ListMemoriesByAgentAndType(ctx context.Context, tenantID uuid.UUID, agentID, memoryType string, limit, offset int) ([]*storagestate.AgentMemory, int64, error) {
	var memories []*storagestate.AgentMemory
	var total int64

	err := r.db.WithContext(ctx).Model(&storagestate.AgentMemory{}).
		Where("tenant_id = ? AND agent_id = ? AND memory_type = ?", tenantID, agentID, memoryType).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	err = r.db.WithContext(ctx).
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

func (r *AgentMemoryRepository) ListMemoriesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*storagestate.AgentMemory, int64, error) {
	var memories []*storagestate.AgentMemory
	var total int64

	err := r.db.WithContext(ctx).Model(&storagestate.AgentMemory{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	err = r.db.WithContext(ctx).
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

func (r *AgentMemoryRepository) SearchMemories(ctx context.Context, tenantID uuid.UUID, agentID, memoryType string, embedding []float32, limit int, threshold float32) ([]*storagestate.AgentMemory, error) {
	var memories []*storagestate.AgentMemory

	query := r.db.WithContext(ctx).
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

