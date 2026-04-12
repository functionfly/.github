package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamMemoryRepository provides data access for team memories
type TeamMemoryRepository interface {
	// CRUD operations
	Create(ctx context.Context, memory *TeamMemory) (*TeamMemory, error)
	GetByID(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) (*TeamMemory, error)
	Update(ctx context.Context, memory *TeamMemory) (*TeamMemory, error)
	Delete(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) error

	// Team-scoped queries
	ListByTeam(ctx context.Context, tenantID, teamID uuid.UUID, filter TeamMemoryFilter) ([]*TeamMemory, int64, error)
	ListByType(ctx context.Context, tenantID, teamID uuid.UUID, memoryType string, limit, offset int) ([]*TeamMemory, error)
	ListByCategory(ctx context.Context, tenantID, teamID uuid.UUID, category string) ([]*TeamMemory, error)
	ListPendingValidation(ctx context.Context, tenantID, teamID uuid.UUID, limit int) ([]*TeamMemory, error)

	// Semantic search (vector similarity)
	SearchSimilar(ctx context.Context, tenantID, teamID uuid.UUID, queryVector []float32, limit int) ([]*TeamMemorySearchResult, error)
	SearchByText(ctx context.Context, tenantID, teamID uuid.UUID, searchText string, limit int) ([]*TeamMemorySearchResult, error)

	// Vector-only search for encrypted memories (can't search content JSON)
	SearchByEmbeddingOnly(ctx context.Context, tenantID, teamID uuid.UUID, queryVector []float32, limit int) ([]*TeamMemorySearchResult, error)

	// Auto-update related
	FindByConversation(ctx context.Context, conversationID uuid.UUID) ([]*TeamMemory, error)
	MarkAsAccessed(ctx context.Context, memoryID uuid.UUID) error
	UpdateFromExtraction(ctx context.Context, memoryID uuid.UUID, newContent JSONMap, confidence float64) error

	// Validation workflow
	ValidateMemory(ctx context.Context, memoryID uuid.UUID, validatedBy uuid.UUID) error
	InvalidateMemory(ctx context.Context, memoryID uuid.UUID) error

	// Maintenance
	CleanupExpired(ctx context.Context, batchSize int) (int, error)
	ApplyDecay(ctx context.Context, daysSinceAccess int) (int, error)
	ListForAutoUpdate(ctx context.Context, batchSize int) ([]*TeamMemory, error)

	// Extraction queue
	CreateExtraction(ctx context.Context, extraction *MemoryExtraction) (*MemoryExtraction, error)
	GetExtractionsByTeam(ctx context.Context, teamID uuid.UUID, status string, limit int) ([]*MemoryExtraction, error)
	ApproveExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID) (*TeamMemory, error)
	RejectExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID, reason string) error
	ProcessAutoApplyExtractions(ctx context.Context, batchSize int) (int, error)

	// Encryption-aware helpers
	GetDecryptionPayload(ctx context.Context, memoryID uuid.UUID) (encryptedContent, iv, tag []byte, err error)
	CreateEncryptedMemory(ctx context.Context, memory *TeamMemory, encryptedContent, iv, tag []byte) (*TeamMemory, error)
}

// teamMemoryRepo implements TeamMemoryRepository
type teamMemoryRepo struct {
	db               *gorm.DB
	embeddingService EmbeddingService
}

// NewTeamMemoryRepository creates a new TeamMemoryRepository
func NewTeamMemoryRepository(db *gorm.DB, embedSvc EmbeddingService) TeamMemoryRepository {
	return &teamMemoryRepo{db: db, embeddingService: embedSvc}
}

// Create stores a new team memory
func (r *teamMemoryRepo) Create(ctx context.Context, memory *TeamMemory) (*TeamMemory, error) {
	if memory.ID == uuid.Nil {
		memory.ID = uuid.New()
	}

	err := r.db.WithContext(ctx).Create(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create team memory: %w", err)
	}

	return memory, nil
}

// GetByID retrieves a memory by ID with tenant/team scoping
func (r *teamMemoryRepo) GetByID(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) (*TeamMemory, error) {
	var memory TeamMemory
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND team_id = ?", memoryID, tenantID, teamID).
		First(&memory).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team memory not found")
		}
		return nil, fmt.Errorf("failed to get team memory: %w", err)
	}

	return &memory, nil
}

// Update updates an existing memory
func (r *teamMemoryRepo) Update(ctx context.Context, memory *TeamMemory) (*TeamMemory, error) {
	err := r.db.WithContext(ctx).Save(memory).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update team memory: %w", err)
	}

	return memory, nil
}

// Delete removes a memory
func (r *teamMemoryRepo) Delete(ctx context.Context, tenantID, teamID, memoryID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND team_id = ?", memoryID, tenantID, teamID).
		Delete(&TeamMemory{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete team memory: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("team memory not found")
	}

	return nil
}

// ListByTeam retrieves memories with filtering
func (r *teamMemoryRepo) ListByTeam(ctx context.Context, tenantID, teamID uuid.UUID, filter TeamMemoryFilter) ([]*TeamMemory, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&TeamMemory{}).
		Where("tenant_id = ? AND team_id = ?", tenantID, teamID).
		Where("expires_at IS NULL OR expires_at > NOW()")

	// Apply filters
	if filter.MemoryType != nil {
		query = query.Where("memory_type = ?", *filter.MemoryType)
	}
	if filter.Category != nil {
		query = query.Where("category = ?", *filter.Category)
	}
	if filter.IsValidated != nil {
		query = query.Where("is_validated = ?", *filter.IsValidated)
	}
	if filter.MinConfidence != nil {
		query = query.Where("confidence_score >= ?", *filter.MinConfidence)
	}
	if filter.IsEncrypted != nil {
		query = query.Where("is_encrypted = ?", *filter.IsEncrypted)
	}
	if filter.CreatedAfter != nil {
		query = query.Where("created_at > ?", *filter.CreatedAfter)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count memories: %w", err)
	}

	// Apply pagination and ordering
	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	offset := filter.Offset

	var memories []*TeamMemory
	err := query.
		Order("importance_score DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memories: %w", err)
	}

	return memories, total, nil
}

// ListByType retrieves memories by type
func (r *teamMemoryRepo) ListByType(ctx context.Context, tenantID, teamID uuid.UUID, memoryType string, limit, offset int) ([]*TeamMemory, error) {
	var memories []*TeamMemory
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND team_id = ? AND memory_type = ?", tenantID, teamID, memoryType).
		Where("expires_at IS NULL OR expires_at > NOW()").
		Order("importance_score DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list memories by type: %w", err)
	}

	return memories, nil
}

// ListByCategory retrieves memories by category
func (r *teamMemoryRepo) ListByCategory(ctx context.Context, tenantID, teamID uuid.UUID, category string) ([]*TeamMemory, error) {
	var memories []*TeamMemory
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND team_id = ? AND category = ?", tenantID, teamID, category).
		Where("expires_at IS NULL OR expires_at > NOW()").
		Order("created_at DESC").
		Find(&memories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list memories by category: %w", err)
	}

	return memories, nil
}

// ListPendingValidation retrieves memories awaiting validation
func (r *teamMemoryRepo) ListPendingValidation(ctx context.Context, tenantID, teamID uuid.UUID, limit int) ([]*TeamMemory, error) {
	var memories []*TeamMemory
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND team_id = ? AND is_validated = ?", tenantID, teamID, false).
		Where("confidence_score >= 0.7"). // Show higher confidence unvalidated memories
		Order("confidence_score DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list pending memories: %w", err)
	}

	return memories, nil
}

// SearchSimilar performs vector similarity search
func (r *teamMemoryRepo) SearchSimilar(ctx context.Context, tenantID, teamID uuid.UUID, queryVector []float32, limit int) ([]*TeamMemorySearchResult, error) {
	if limit == 0 {
		limit = 10
	}

	// Convert vector to PostgreSQL format
	vectorStr := pgVectorToString(queryVector)

	// Search unencrypted memories first (can use content + embedding)
	var results []*TeamMemorySearchResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			m.*,
			1.0 - (m.embedding <=> ?) as relevance_score
		FROM team_memories m
		WHERE m.tenant_id = ? 
		  AND m.team_id = ?
		  AND (m.expires_at IS NULL OR m.expires_at > NOW())
		  AND m.is_encrypted = false
		ORDER BY m.embedding <=> ?
		LIMIT ?
	`, vectorStr, tenantID, teamID, vectorStr, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	return results, nil
}

// SearchByText generates embedding from text then searches
func (r *teamMemoryRepo) SearchByText(ctx context.Context, tenantID, teamID uuid.UUID, searchText string, limit int) ([]*TeamMemorySearchResult, error) {
	// Generate embedding for search text
	embedding, err := r.embeddingService.GenerateEmbedding(ctx, searchText)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	return r.SearchSimilar(ctx, tenantID, teamID, embedding, limit)
}

// SearchByEmbeddingOnly searches including encrypted memories (embedding-only, no content match)
func (r *teamMemoryRepo) SearchByEmbeddingOnly(ctx context.Context, tenantID, teamID uuid.UUID, queryVector []float32, limit int) ([]*TeamMemorySearchResult, error) {
	if limit == 0 {
		limit = 10
	}

	vectorStr := pgVectorToString(queryVector)

	var results []*TeamMemorySearchResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			m.*,
			1.0 - (m.embedding <=> ?) as relevance_score
		FROM team_memories m
		WHERE m.tenant_id = ? 
		  AND m.team_id = ?
		  AND (m.expires_at IS NULL OR m.expires_at > NOW())
		ORDER BY m.embedding <=> ?
		LIMIT ?
	`, vectorStr, tenantID, teamID, vectorStr, limit).Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search all memories: %w", err)
	}

	return results, nil
}

// pgVectorToString converts float32 slice to PostgreSQL vector format
func pgVectorToString(vec []float32) string {
	var parts []string
	for _, v := range vec {
		parts = append(parts, fmt.Sprintf("%.6f", v))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// FindByConversation retrieves memories sourced from a conversation
func (r *teamMemoryRepo) FindByConversation(ctx context.Context, conversationID uuid.UUID) ([]*TeamMemory, error) {
	var memories []*TeamMemory
	err := r.db.WithContext(ctx).
		Where("source_conversation_id = ?", conversationID).
		Find(&memories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find memories by conversation: %w", err)
	}

	return memories, nil
}

// MarkAsAccessed increments access counter and updates timestamp
func (r *teamMemoryRepo) MarkAsAccessed(ctx context.Context, memoryID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&TeamMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"access_count":     gorm.Expr("access_count + 1"),
			"last_accessed_at": time.Now(),
		}).Error
}

// UpdateFromExtraction merges AI-extracted content into existing memory
func (r *teamMemoryRepo) UpdateFromExtraction(ctx context.Context, memoryID uuid.UUID, newContent JSONMap, confidence float64) error {
	return r.db.WithContext(ctx).Model(&TeamMemory{}).
		Where("id = ? AND auto_update_enabled = ?", memoryID, true).
		Updates(map[string]interface{}{
			"content":              newContent,
			"confidence_score":     gorm.Expr("GREATEST(confidence_score, ?)", confidence),
			"last_auto_updated_at": time.Now(),
			"updated_at":           time.Now(),
		}).Error
}

// ValidateMemory marks a memory as validated
func (r *teamMemoryRepo) ValidateMemory(ctx context.Context, memoryID uuid.UUID, validatedBy uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&TeamMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"is_validated": true,
			"validated_by": validatedBy,
			"validated_at": &now,
			"updated_at":   now,
		}).Error
}

// InvalidateMemory removes validation status
func (r *teamMemoryRepo) InvalidateMemory(ctx context.Context, memoryID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&TeamMemory{}).
		Where("id = ?", memoryID).
		Updates(map[string]interface{}{
			"is_validated": false,
			"validated_by": nil,
			"validated_at": nil,
			"updated_at":   time.Now(),
		}).Error
}

// CleanupExpired removes expired memories
func (r *teamMemoryRepo) CleanupExpired(ctx context.Context, batchSize int) (int, error) {
	if batchSize == 0 {
		batchSize = 100
	}

	result := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < NOW()").
		Limit(batchSize).
		Delete(&TeamMemory{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired memories: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// ApplyDecay reduces importance score for stale memories
func (r *teamMemoryRepo) ApplyDecay(ctx context.Context, daysSinceAccess int) (int, error) {
	if daysSinceAccess == 0 {
		daysSinceAccess = 90
	}

	cutoff := time.Now().AddDate(0, 0, -daysSinceAccess)

	result := r.db.WithContext(ctx).Model(&TeamMemory{}).
		Where("last_accessed_at < ? AND importance_score > 0.1", cutoff).
		Where("is_encrypted = ?", false). // Only decay unencrypted (encrypted handled client-side)
		Update("importance_score", gorm.Expr("GREATEST(importance_score * 0.9, 0.1)"))

	if result.Error != nil {
		return 0, fmt.Errorf("failed to apply decay: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// ListForAutoUpdate retrieves memories eligible for AI updates
func (r *teamMemoryRepo) ListForAutoUpdate(ctx context.Context, batchSize int) ([]*TeamMemory, error) {
	if batchSize == 0 {
		batchSize = 50
	}

	var memories []*TeamMemory
	err := r.db.WithContext(ctx).
		Where("auto_update_enabled = ?", true).
		Where("is_encrypted = ?", false). // Can't auto-update encrypted
		Where("(last_auto_updated_at IS NULL OR last_auto_updated_at < ?)", time.Now().AddDate(0, 0, -7)).
		Limit(batchSize).
		Find(&memories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list memories for auto update: %w", err)
	}

	return memories, nil
}

// CreateExtraction adds a new extraction to the queue
func (r *teamMemoryRepo) CreateExtraction(ctx context.Context, extraction *MemoryExtraction) (*MemoryExtraction, error) {
	if extraction.ID == uuid.Nil {
		extraction.ID = uuid.New()
	}

	err := r.db.WithContext(ctx).Create(extraction).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create extraction: %w", err)
	}

	return extraction, nil
}

// GetExtractionsByTeam retrieves extractions by status
func (r *teamMemoryRepo) GetExtractionsByTeam(ctx context.Context, teamID uuid.UUID, status string, limit int) ([]*MemoryExtraction, error) {
	if limit == 0 {
		limit = 20
	}

	query := r.db.WithContext(ctx).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var extractions []*MemoryExtraction
	err := query.
		Order("confidence DESC, created_at DESC").
		Limit(limit).
		Find(&extractions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get extractions: %w", err)
	}

	return extractions, nil
}

// ApproveExtraction converts an extraction to a validated memory
func (r *teamMemoryRepo) ApproveExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID) (*TeamMemory, error) {
	// Get the extraction
	var extraction MemoryExtraction
	if err := r.db.WithContext(ctx).First(&extraction, "id = ?", extractionID).Error; err != nil {
		return nil, fmt.Errorf("extraction not found: %w", err)
	}

	if extraction.Status != "pending" {
		return nil, fmt.Errorf("extraction already processed")
	}

	// Update extraction status
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&MemoryExtraction{}).
		Where("id = ?", extractionID).
		Updates(map[string]interface{}{
			"status":      "approved",
			"reviewed_by": reviewedBy,
			"reviewed_at": &now,
		}).Error

	if err != nil {
		return nil, fmt.Errorf("failed to update extraction: %w", err)
	}

	// Create the memory
	memory := extraction.ToTeamMemory(reviewedBy)
	memory.IsValidated = true
	memory.ValidatedBy = &reviewedBy
	memory.ValidatedAt = &now

	// Generate embedding
	if r.embeddingService != nil {
		searchText := extraction.Summary + " " + extractTextFromContent(extraction.Content, extraction.MemoryType)
		embedding, err := r.embeddingService.GenerateEmbedding(ctx, searchText)
		if err == nil {
			memory.Embedding = embedding
		}
	}

	return r.Create(ctx, memory)
}

// RejectExtraction marks an extraction as rejected
func (r *teamMemoryRepo) RejectExtraction(ctx context.Context, extractionID uuid.UUID, reviewedBy uuid.UUID, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&MemoryExtraction{}).
		Where("id = ?", extractionID).
		Updates(map[string]interface{}{
			"status":           "rejected",
			"reviewed_by":      reviewedBy,
			"reviewed_at":      &now,
			"rejection_reason": reason,
		}).Error
}

// ProcessAutoApplyExtractions auto-approves high-confidence extractions (MVP: >= 0.9)
func (r *teamMemoryRepo) ProcessAutoApplyExtractions(ctx context.Context, batchSize int) (int, error) {
	if batchSize == 0 {
		batchSize = 10
	}

	// Find pending extractions with high confidence
	var extractions []*MemoryExtraction
	err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Where("confidence >= auto_apply_threshold").
		Limit(batchSize).
		Find(&extractions).Error

	if err != nil {
		return 0, fmt.Errorf("failed to find extractions for auto-apply: %w", err)
	}

	count := 0
	for _, extraction := range extractions {
		// Use a system user ID for auto-approval (or extraction team creator)
		autoReviewer := uuid.Nil // In production, use a system user ID

		_, err := r.ApproveExtraction(ctx, extraction.ID, autoReviewer)
		if err != nil {
			continue // Log error but continue processing
		}

		// Mark as auto_applied
		r.db.WithContext(ctx).Model(&MemoryExtraction{}).
			Where("id = ?", extraction.ID).
			Update("status", "auto_applied")

		count++
	}

	return count, nil
}

// GetDecryptionPayload retrieves encrypted data for client-side decryption
func (r *teamMemoryRepo) GetDecryptionPayload(ctx context.Context, memoryID uuid.UUID) (encryptedContent, iv, tag []byte, err error) {
	var memory TeamMemory
	err = r.db.WithContext(ctx).
		Select("encrypted_content, encryption_iv, encryption_tag").
		Where("id = ? AND is_encrypted = ?", memoryID, true).
		First(&memory).Error

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get decryption payload: %w", err)
	}

	return memory.EncryptedContent, memory.EncryptionIV, memory.EncryptionTag, nil
}

// CreateEncryptedMemory creates a memory with pre-encrypted content (client-side encryption)
func (r *teamMemoryRepo) CreateEncryptedMemory(ctx context.Context, memory *TeamMemory, encryptedContent, iv, tag []byte) (*TeamMemory, error) {
	memory.IsEncrypted = true
	memory.EncryptedContent = encryptedContent
	memory.EncryptionIV = iv
	memory.EncryptionTag = tag
	memory.Content = nil // Clear plaintext content

	return r.Create(ctx, memory)
}

// EmbeddingService interface for generating embeddings
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}
