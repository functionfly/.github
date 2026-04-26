package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Publisher publishes generated functions to the function registry
type Publisher struct {
	db *gorm.DB
}

// NewPublisher creates a new function publisher
func NewPublisher(db *gorm.DB) *Publisher {
	return &Publisher{db: db}
}

// PublishRequest represents a request to publish a function
type PublishRequest struct {
	AgentID         string         `json:"agent_id" validate:"required"`
	GeneratedCodeID uuid.UUID      `json:"generated_code_id" validate:"required"`
	GeneratedCode   *GeneratedCode `json:"-"`
	Author          string         `json:"author" validate:"required"`
	Name            string         `json:"name" validate:"required"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Category        string         `json:"category"`
	Tags            []string       `json:"tags"`
	IsPublic        bool           `json:"is_public"`
	Changelog       *ChangelogSpec `json:"changelog,omitempty"` // Optional changelog for this version
}

// ChangelogSpec represents changelog information for a published function
type ChangelogSpec struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Category    string         `json:"category"` // "feature", "bug_fix", "performance", etc.
	Changes     []ChangeDetail `json:"changes"`
}

// ChangeDetail represents a single change in the changelog
type ChangeDetail struct {
	Component   string `json:"component"`
	Field       string `json:"field"`
	Before      any    `json:"before"`
	After       any    `json:"after"`
	Description string `json:"description"`
}

// PublishedFunction represents a successfully published function
type PublishedFunction struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID            string     `json:"agent_id" gorm:"not null"`
	GeneratedCodeID    uuid.UUID  `json:"generated_code_id" gorm:"type:uuid"`
	FunctionID         string     `json:"function_id"` // The published function ID (author/name)
	RegistryFunctionID *uuid.UUID `json:"registry_function_id" gorm:"type:uuid"`
	Author             string     `json:"author" gorm:"not null"`
	Name               string     `json:"name" gorm:"not null"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Category           string     `json:"category"`
	Tags               pq.StringArray `json:"tags" gorm:"type:text[]"`
	IsPublic           bool       `json:"is_public" gorm:"not null;default:false"`
	Status             string     `json:"status" gorm:"not null;default:'pending'"` // pending, published, failed
	Version            string     `json:"version" gorm:"not null;default:'1.0.0'"`
	ErrorMessage       *string    `json:"error_message"`
	PublishedAt        *time.Time `json:"published_at"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (PublishedFunction) TableName() string {
	return "agent_published_functions"
}

// Publish publishes a generated function to the registry
func (p *Publisher) Publish(ctx context.Context, req *PublishRequest) (*PublishedFunction, error) {
	// Get the generated code
	var generated GeneratedCode
	if req.GeneratedCode != nil {
		generated = *req.GeneratedCode
		if generated.ID == uuid.Nil {
			generated.ID = req.GeneratedCodeID
		}
	} else {
		err := p.db.WithContext(ctx).Where("id = ?", req.GeneratedCodeID).First(&generated).Error
		if err != nil {
			return nil, fmt.Errorf("generated code not found: %w", err)
		}
	}

	if generated.Status != "success" {
		return nil, fmt.Errorf("generated code is not successful: %s", generated.Status)
	}

	published := &PublishedFunction{
		ID:              uuid.New(),
		AgentID:         req.AgentID,
		GeneratedCodeID: req.GeneratedCodeID,
		FunctionID:      fmt.Sprintf("%s/%s", req.Author, req.Name),
		Author:          req.Author,
		Name:            req.Name,
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		Tags:            req.Tags,
		IsPublic:        req.IsPublic,
		Status:          "pending",
		Version:         "1.0.0",
	}

	// Create the function in the registry
	funcID, err := p.createRegistryFunction(ctx, generated, req)
	if err != nil {
		published.Status = "failed"
		errMsg := err.Error()
		published.ErrorMessage = &errMsg
		p.db.WithContext(ctx).Create(published)
		return published, err
	}

	published.RegistryFunctionID = &funcID
	published.Status = "published"
	now := time.Now()
	published.PublishedAt = &now

	if err := p.db.WithContext(ctx).Create(published).Error; err != nil {
		return nil, fmt.Errorf("failed to save published function: %w", err)
	}

	// Update the agent's function ownership
	if err := p.trackOwnership(ctx, req.AgentID, funcID); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to track ownership: %v\n", err)
	}

	// Create changelog entry for the published function
	if err := p.createChangelogEntry(ctx, funcID, req); err != nil {
		// Log but don't fail the publish
		fmt.Printf("Warning: failed to create changelog entry: %v\n", err)
	}

	return published, nil
}

// createRegistryFunction creates a function entry in the registry
func (p *Publisher) createRegistryFunction(ctx context.Context, generated GeneratedCode, req *PublishRequest) (uuid.UUID, error) {
	// The function was already created by GenerateFunction() in generation/service.go
	// We just need to find it and update its metadata
	var existing identity.Function

	err := p.db.WithContext(ctx).
		Where("author = ? AND name = ?", req.Author, req.Name).
		First(&existing).Error

	if err == nil {
		// Function exists from GenerateFunction(), update it
		existing.LatestVersion = "1.0.0"
		existing.Description = req.Description
		existing.Category = req.Category
		existing.Tags = req.Tags
		if req.IsPublic {
			existing.Visibility = "public"
		}
		existing.OwnerAgentID = &req.AgentID
		existing.AgentGenerated = true
		existing.GenerationModel = &generated.ModelUsed

		if err := p.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return uuid.Nil, fmt.Errorf("failed to update function: %w", err)
		}

		existingID, _ := uuid.Parse(existing.ID)
		return existingID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return uuid.Nil, fmt.Errorf("failed to check existing function: %w", err)
	}

	// Function not found - this shouldn't happen if GenerateFunction succeeded
	// But if it does (e.g. partial failure), create it with the generated ID
	newFunc := identity.Function{
		ID:                 generated.ID.String(),
		Author:             req.Author,
		Name:               req.Name,
		LatestVersion:      "1.0.0",
		Title:              req.Title,
		Description:        req.Description,
		Category:           req.Category,
		Tags:               req.Tags,
		Visibility:         "private",
		PopularityScore:    0,
		ReliabilityScore:   0,
		DeterministicScore: 0,
		OwnerAgentID:       &req.AgentID,
		AgentGenerated:     true,
		GenerationModel:    &generated.ModelUsed,
	}

	if req.IsPublic {
		newFunc.Visibility = "public"
	}

	if err := p.db.WithContext(ctx).Create(&newFunc).Error; err != nil {
		return uuid.Nil, fmt.Errorf("failed to create function: %w", err)
	}

	return generated.ID, nil
}

// trackOwnership tracks that an agent owns a function
func (p *Publisher) trackOwnership(ctx context.Context, agentID string, functionID uuid.UUID) error {
	// This could be implemented as a separate table or as part of the agent identity
	// For now, the function model already has OwnerAgentID
	return nil
}

// GetPublishedFunctions retrieves published functions for an agent
func (p *Publisher) GetPublishedFunctions(ctx context.Context, agentID string, limit, offset int) ([]PublishedFunction, int64, error) {
	var total int64
	var functions []PublishedFunction

	query := p.db.WithContext(ctx).Model(&PublishedFunction{}).Where("agent_id = ?", agentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get functions: %w", err)
	}

	return functions, total, nil
}

// GetPublishedFunction retrieves a specific published function
func (p *Publisher) GetPublishedFunction(ctx context.Context, functionID uuid.UUID) (*PublishedFunction, error) {
	var published PublishedFunction
	err := p.db.WithContext(ctx).Where("id = ?", functionID).First(&published).Error
	return &published, err
}

// GetFunctionByURI retrieves a published function by its URI
func (p *Publisher) GetFunctionByURI(ctx context.Context, author, name string) (*PublishedFunction, error) {
	var published PublishedFunction
	err := p.db.WithContext(ctx).Where("author = ? AND name = ?", author, name).First(&published).Error
	return &published, err
}

// Unpublish removes a function from the registry
func (p *Publisher) Unpublish(ctx context.Context, functionID uuid.UUID) error {
	var published PublishedFunction
	if err := p.db.WithContext(ctx).Where("id = ?", functionID).First(&published).Error; err != nil {
		return fmt.Errorf("function not found: %w", err)
	}

	if published.Status != "published" {
		return fmt.Errorf("function is not published")
	}

	// Update status
	published.Status = "unpublished"
	published.UpdatedAt = time.Now()

	if err := p.db.WithContext(ctx).Save(&published).Error; err != nil {
		return fmt.Errorf("failed to unpublish: %w", err)
	}

	// Optionally hide the function in the registry
	if published.RegistryFunctionID != nil {
		var fn identity.Function
		if err := p.db.WithContext(ctx).Where("id = ?", *published.RegistryFunctionID).First(&fn).Error; err == nil {
			fn.Visibility = "private"
			p.db.WithContext(ctx).Save(&fn)
		}
	}

	return nil
}

// AutoMigrate runs auto migration for deployment models
func (p *Publisher) AutoMigrate(ctx context.Context) error {
	return p.db.WithContext(ctx).AutoMigrate(
		&GeneratedCode{},
		&PublishedFunction{},
	)
}

// createChangelogEntry creates a changelog entry for a newly published function
func (p *Publisher) createChangelogEntry(ctx context.Context, functionID uuid.UUID, req *PublishRequest) error {
	// Determine change type and category
	changeType := registry.ChangeTypeAdded
	category := registry.ChangeCategoryFeature
	if req.Changelog != nil && req.Changelog.Category != "" {
		category = registry.ChangeCategory(req.Changelog.Category)
	}

	// Build title and description
	title := "Initial release"
	description := "First version of the function published by agent"
	if req.Changelog != nil {
		if req.Changelog.Title != "" {
			title = req.Changelog.Title
		}
		if req.Changelog.Description != "" {
			description = req.Changelog.Description
		}
	}

	// Convert changes to JSON
	var changesJSON []byte
	if req.Changelog != nil && len(req.Changelog.Changes) > 0 {
		changes := make([]registry.FunctionChangelogChange, len(req.Changelog.Changes))
		for i, c := range req.Changelog.Changes {
			changes[i] = registry.FunctionChangelogChange{
				Component:   c.Component,
				Field:       c.Field,
				Before:      c.Before,
				After:       c.After,
				Description: c.Description,
			}
		}
		var err error
		changesJSON, err = json.Marshal(changes)
		if err != nil {
			return fmt.Errorf("failed to marshal changes: %w", err)
		}
	} else {
		changesJSON = []byte("[]")
	}

	// Get the function version ID
	var fnVersion identity.Function
	if err := p.db.WithContext(ctx).Where("id = ?", functionID.String()).First(&fnVersion).Error; err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	// Create changelog entry
	changelog := &registry.FunctionVersionChangelog{
		FunctionID:        functionID,
		FunctionVersionID: functionID, // Use function ID as version ID for initial version
		Version:           "1.0.0",
		ChangeType:        changeType,
		Category:          category,
		Title:             title,
		Description:       description,
		Changes:           changesJSON,
		Author:            req.AgentID,
	}

	if err := p.db.WithContext(ctx).Create(changelog).Error; err != nil {
		return fmt.Errorf("failed to create changelog: %w", err)
	}

	return nil
}
