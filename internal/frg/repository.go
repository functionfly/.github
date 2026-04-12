// Package frg provides the Function Registry + Live Runtime Graph repository layer
package frg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles FRG database operations
type Repository struct {
	db     *gorm.DB
	cache  *cache.RegistryRedisCache
	keyGen *cache.RegistryCacheKey
}

// NewRepository creates a new FRG repository
func NewRepository(db *gorm.DB, redisCache *cache.RegistryRedisCache) *Repository {
	var keyGen *cache.RegistryCacheKey
	if redisCache != nil {
		keyGen = cache.NewRegistryCacheKey()
	}

	return &Repository{
		db:     db,
		cache:  redisCache,
		keyGen: keyGen,
	}
}

// AutoMigrate runs database migrations for FRG models
func (r *Repository) AutoMigrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&GraphDefinition{},
		&GraphInstance{},
		&GraphNodeExecution{},
		&GraphEdgeExecution{},
		&GraphEvent{},
		&GraphOptimizationSuggestion{},
	)
}

// ==================== Graph Definitions ====================

// CreateDefinition creates a new graph definition
func (r *Repository) CreateDefinition(ctx context.Context, def *GraphDefinition) (*GraphDefinition, error) {
	if def.ID == uuid.Nil {
		def.ID = uuid.New()
	}
	if def.Version == "" {
		def.Version = "v1"
	}

	if err := r.db.WithContext(ctx).Create(def).Error; err != nil {
		return nil, fmt.Errorf("failed to create graph definition: %w", err)
	}

	return def, nil
}

// GetDefinitionByName retrieves a graph definition by author/name/version
func (r *Repository) GetDefinitionByName(ctx context.Context, author, name, version string) (*GraphDefinition, error) {
	var def GraphDefinition
	err := r.db.WithContext(ctx).
		Where("author = ? AND name = ? AND version = ?", author, name, version).
		First(&def).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("graph not found: %s/%s@%s", author, name, version)
		}
		return nil, err
	}

	return &def, nil
}

// GetDefinitionByID retrieves a graph definition by ID
func (r *Repository) GetDefinitionByID(ctx context.Context, id uuid.UUID) (*GraphDefinition, error) {
	var def GraphDefinition
	if err := r.db.WithContext(ctx).First(&def, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("graph not found: %s", id)
		}
		return nil, err
	}
	return &def, nil
}

// GetLatestVersion retrieves the latest version of a graph
func (r *Repository) GetLatestVersion(ctx context.Context, author, name string) (*GraphDefinition, error) {
	var def GraphDefinition
	err := r.db.WithContext(ctx).
		Where("author = ? AND name = ?", author, name).
		Order("published_at DESC NULLS LAST, created_at DESC").
		First(&def).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("graph not found: %s/%s", author, name)
		}
		return nil, err
	}

	return &def, nil
}

// ListDefinitions lists graph definitions with filtering
func (r *Repository) ListDefinitions(ctx context.Context, filter *DefinitionFilter) ([]*GraphDefinition, error) {
	query := r.db.WithContext(ctx).Model(&GraphDefinition{})

	if filter.Author != "" {
		query = query.Where("author = ?", filter.Author)
	}
	if filter.Visibility != "" {
		query = query.Where("visibility = ?", filter.Visibility)
	}
	if filter.ExecutionMode != "" {
		query = query.Where("execution_mode = ?", filter.ExecutionMode)
	}
	if filter.TenantID != nil {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.OwnerUserID != nil {
		query = query.Where("owner_user_id = ?", filter.OwnerUserID)
	}

	query = query.Order("created_at DESC")

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var defs []*GraphDefinition
	if err := query.Find(&defs).Error; err != nil {
		return nil, err
	}

	return defs, nil
}

// UpdateDefinition updates a graph definition (only if not published)
func (r *Repository) UpdateDefinition(ctx context.Context, def *GraphDefinition) error {
	// Cannot modify published graphs
	var existing GraphDefinition
	if err := r.db.WithContext(ctx).First(&existing, def.ID).Error; err != nil {
		return err
	}
	if existing.PublishedAt != nil {
		return errors.New("cannot modify published graph - create new version instead")
	}

	def.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(def).Error
}

// PublishVersion marks a graph definition as published
func (r *Repository) PublishVersion(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&GraphDefinition{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"published_at": now,
			"updated_at":   now,
		}).Error
}

// DeleteDefinition deletes a graph definition (only if not published)
func (r *Repository) DeleteDefinition(ctx context.Context, id uuid.UUID) error {
	var existing GraphDefinition
	if err := r.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		return err
	}
	if existing.PublishedAt != nil {
		return errors.New("cannot delete published graph - deprecate instead")
	}

	return r.db.WithContext(ctx).Delete(&GraphDefinition{}, id).Error
}

// ForkGraph creates a fork of an existing graph
func (r *Repository) ForkGraph(ctx context.Context, sourceAuthor, sourceName, sourceVersion, newAuthor, newName string, ownerUserID *uuid.UUID) (*GraphDefinition, error) {
	// Get source graph
	source, err := r.GetDefinitionByName(ctx, sourceAuthor, sourceName, sourceVersion)
	if err != nil {
		return nil, err
	}

	// Create new graph with fork lineage
	fork := &GraphDefinition{
		Author:            newAuthor,
		Name:              newName,
		Version:           "v1",
		NodeRefs:          source.NodeRefs,
		Edges:             source.Edges,
		ExecutionMode:     source.ExecutionMode,
		TriggerConfig:     source.TriggerConfig,
		InputSchema:       source.InputSchema,
		OutputSchema:      source.OutputSchema,
		ForkedFromAuthor:  &sourceAuthor,
		ForkedFromName:    &sourceName,
		ForkedFromVersion: &sourceVersion,
		OwnerUserID:       ownerUserID,
		Visibility:        "public",
	}

	return r.CreateDefinition(ctx, fork)
}

// DefinitionFilter provides filtering for graph definitions
type DefinitionFilter struct {
	Author        string
	Visibility    string
	ExecutionMode string
	TenantID      *uuid.UUID
	OwnerUserID   *uuid.UUID
	Limit         int
	Offset        int
}

// ==================== Semantic Search ====================

// SearchByEmbedding finds graphs by semantic similarity
func (r *Repository) SearchByEmbedding(ctx context.Context, embedding []byte, limit int) ([]*GraphDefinition, error) {
	var defs []*GraphDefinition

	// pgvector cosine similarity search
	err := r.db.WithContext(ctx).
		Raw("SELECT * FROM graph_definitions ORDER BY ai_embedding <=> ? LIMIT ?", embedding, limit).
		Scan(&defs).Error

	if err != nil {
		return nil, err
	}

	return defs, nil
}

// SearchByText finds graphs by full-text search
func (r *Repository) SearchByText(ctx context.Context, query string, limit int) ([]*GraphDefinition, error) {
	var defs []*GraphDefinition

	err := r.db.WithContext(ctx).
		Where("to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(ai_description, '')) @@ plainto_tsquery(?)", query).
		Limit(limit).
		Find(&defs).Error

	if err != nil {
		return nil, err
	}

	return defs, nil
}

// ==================== Graph Instances ====================

// CreateInstance creates a new graph execution instance
func (r *Repository) CreateInstance(ctx context.Context, def *GraphDefinition, input json.RawMessage) (*GraphInstance, error) {
	instance := &GraphInstance{
		ID:                uuid.New(),
		DefinitionID:      def.ID,
		Status:            InstanceStatusPending,
		InputData:         input,
		FrozenNodes:       def.NodeRefs,
		FrozenEdges:       def.Edges,
		NodeStates:        json.RawMessage("{}"),
		ExecutionRootHash: "", // Will be computed at completion
	}

	if err := r.db.WithContext(ctx).Create(instance).Error; err != nil {
		return nil, fmt.Errorf("failed to create graph instance: %w", err)
	}

	return instance, nil
}

// GetInstanceByID retrieves an instance by ID
func (r *Repository) GetInstanceByID(ctx context.Context, id uuid.UUID) (*GraphInstance, error) {
	var instance GraphInstance
	if err := r.db.WithContext(ctx).First(&instance, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance not found: %s", id)
		}
		return nil, err
	}
	return &instance, nil
}

// UpdateInstanceStatus updates the runtime status of an instance
func (r *Repository) UpdateInstanceStatus(ctx context.Context, id uuid.UUID, status InstanceStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case InstanceStatusRunning, InstanceStatusStreaming:
		updates["started_at"] = now
	case InstanceStatusCompleted, InstanceStatusFailed:
		updates["completed_at"] = now
	}

	return r.db.WithContext(ctx).
		Model(&GraphInstance{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateInstanceNodeState updates the state of a specific node
func (r *Repository) UpdateInstanceNodeState(ctx context.Context, instanceID uuid.UUID, nodeID string, state *NodeState) error {
	// Get current node states
	var instance GraphInstance
	if err := r.db.WithContext(ctx).First(&instance, instanceID).Error; err != nil {
		return err
	}

	var states map[string]*NodeState
	if err := json.Unmarshal(instance.NodeStates, &states); err != nil {
		states = make(map[string]*NodeState)
	}

	states[nodeID] = state
	statesJSON, _ := json.Marshal(states)

	return r.db.WithContext(ctx).
		Model(&GraphInstance{}).
		Where("id = ?", instanceID).
		Update("node_states", statesJSON).Error
}

// ListInstances lists graph instances with filtering
func (r *Repository) ListInstances(ctx context.Context, filter *InstanceFilter) ([]*GraphInstance, error) {
	query := r.db.WithContext(ctx).Model(&GraphInstance{})

	if filter.DefinitionID != nil {
		query = query.Where("definition_id = ?", filter.DefinitionID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	query = query.Order("created_at DESC")

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var instances []*GraphInstance
	if err := query.Find(&instances).Error; err != nil {
		return nil, err
	}

	return instances, nil
}

// InstanceFilter provides filtering for graph instances
type InstanceFilter struct {
	DefinitionID *uuid.UUID
	Status       string
	Limit        int
}

// ==================== Graph Executions ====================

// CreateNodeExecution records a node execution
func (r *Repository) CreateNodeExecution(ctx context.Context, exec *GraphNodeExecution) error {
	if exec.ID == uuid.Nil {
		exec.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(exec).Error
}

// UpdateNodeExecution updates a node execution
func (r *Repository) UpdateNodeExecution(ctx context.Context, exec *GraphNodeExecution) error {
	return r.db.WithContext(ctx).Save(exec).Error
}

// CreateEdgeExecution records an edge execution
func (r *Repository) CreateEdgeExecution(ctx context.Context, exec *GraphEdgeExecution) error {
	if exec.ID == uuid.Nil {
		exec.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(exec).Error
}

// UpdateEdgeExecution updates an edge execution
func (r *Repository) UpdateEdgeExecution(ctx context.Context, exec *GraphEdgeExecution) error {
	return r.db.WithContext(ctx).Save(exec).Error
}

// ==================== Event Stream ====================

// AppendEvent adds an event to the event stream
func (r *Repository) AppendEvent(ctx context.Context, event *GraphEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return r.db.WithContext(ctx).Create(event).Error
}

// GetEvents retrieves events for an instance
func (r *Repository) GetEvents(ctx context.Context, instanceID uuid.UUID, fromSequence int64, limit int) ([]*GraphEvent, error) {
	var events []*GraphEvent
	query := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Order("sequence_num ASC")

	if fromSequence > 0 {
		query = query.Where("sequence_num > ?", fromSequence)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

// GetLatestSequenceNumber gets the highest sequence number for an instance
func (r *Repository) GetLatestSequenceNumber(ctx context.Context, instanceID uuid.UUID) (int64, error) {
	var maxSeq int64
	result := r.db.WithContext(ctx).
		Model(&GraphEvent{}).
		Where("instance_id = ?", instanceID).
		Select("COALESCE(MAX(sequence_num), 0)").
		Scan(&maxSeq)

	if result.Error != nil {
		return 0, result.Error
	}

	return maxSeq, nil
}

// ==================== Optimization Suggestions ====================

// CreateOptimizationSuggestion stores an AI-generated suggestion
func (r *Repository) CreateOptimizationSuggestion(ctx context.Context, suggestion *GraphOptimizationSuggestion) error {
	if suggestion.ID == uuid.Nil {
		suggestion.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(suggestion).Error
}

// GetOptimizationSuggestions retrieves pending suggestions for a graph
func (r *Repository) GetOptimizationSuggestions(ctx context.Context, definitionID uuid.UUID, includeDismissed bool) ([]*GraphOptimizationSuggestion, error) {
	query := r.db.WithContext(ctx).
		Where("definition_id = ?", definitionID).
		Where("applied = ?", false)

	if !includeDismissed {
		query = query.Where("dismissed = ?", false)
	}

	var suggestions []*GraphOptimizationSuggestion
	if err := query.Find(&suggestions).Error; err != nil {
		return nil, err
	}

	return suggestions, nil
}

// DismissSuggestion marks a suggestion as dismissed
func (r *Repository) DismissSuggestion(ctx context.Context, suggestionID uuid.UUID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&GraphOptimizationSuggestion{}).
		Where("id = ?", suggestionID).
		Updates(map[string]interface{}{
			"dismissed":    true,
			"dismissed_at": now,
			"dismissed_by": userID,
		}).Error
}

// ApplySuggestion marks a suggestion as applied
func (r *Repository) ApplySuggestion(ctx context.Context, suggestionID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&GraphOptimizationSuggestion{}).
		Where("id = ?", suggestionID).
		Updates(map[string]interface{}{
			"applied":    true,
			"applied_at": now,
		}).Error
}

// UpdateNodeExecutionCertID updates the execution certificate ID for a node execution
func (r *Repository) UpdateNodeExecutionCertID(executionID uuid.UUID, certID *uuid.UUID) error {
	return r.db.Model(&GraphNodeExecution{}).
		Where("id = ?", executionID).
		Update("execution_cert_id", certID).Error
}

// QueryPublishedGraphsWithTriggers loads all published graphs that have trigger configurations
func (r *Repository) QueryPublishedGraphsWithTriggers(ctx context.Context, out *[]*GraphDefinition) error {
	var graphs []*GraphDefinition
	err := r.db.WithContext(ctx).
		Where("published_at IS NOT NULL").
		Where("trigger_config IS NOT NULL").
		Where("visibility IN ?", []string{"public", "unlisted"}).
		Find(&graphs).Error
	if err != nil {
		return err
	}
	*out = graphs
	return nil
}
