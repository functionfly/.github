package decisions

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles database operations for Team Decisions
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Decisions repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new team decision
func (r *Repository) Create(teamID, userID uuid.UUID, req *CreateDecisionRequest) (*TeamDecision, error) {
	decision := &TeamDecision{
		ID:              uuid.New(),
		TeamID:          teamID,
		Title:           req.Title,
		Description:     req.Description,
		Rationale:       req.Rationale,
		Outcome:         req.Outcome,
		CreatedBy:       userID,
		Status:          string(DecisionStatusPending),
		SourceType:      string(DecisionSourceManual),
		ImportanceScore: req.ImportanceScore,
	}

	if len(req.Alternatives) > 0 {
		if err := decision.SetAlternativesSlice(req.Alternatives); err != nil {
			return nil, err
		}
	}

	if len(req.Tags) > 0 {
		decision.Tags = req.Tags
	}

	if err := r.db.Create(decision).Error; err != nil {
		return nil, err
	}

	return decision, nil
}

// GetByID retrieves a decision by ID
func (r *Repository) GetByID(id uuid.UUID) (*TeamDecision, error) {
	var decision TeamDecision
	err := r.db.Where("id = ?", id).First(&decision).Error
	if err != nil {
		return nil, err
	}
	return &decision, nil
}

// ListByTeam retrieves all decisions for a team with optional filtering
func (r *Repository) ListByTeam(teamID uuid.UUID, status, tag string, limit, offset int) ([]TeamDecision, int64, error) {
	var decisions []TeamDecision
	var total int64

	query := r.db.Model(&TeamDecision{}).Where("team_id = ?", teamID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if tag != "" {
		// Search in the tags array/text field
		query = query.Where("tags::text ILIKE ?", "%"+tag+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Set defaults
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Fetch with pagination, ordered by importance then created_at
	if err := query.
		Order("importance_score DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&decisions).Error; err != nil {
		return nil, 0, err
	}

	return decisions, total, nil
}

// Update updates a decision
func (r *Repository) Update(id uuid.UUID, req *UpdateDecisionRequest) (*TeamDecision, error) {
	decision, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Rationale != nil {
		updates["rationale"] = *req.Rationale
	}
	if req.Outcome != nil {
		updates["outcome"] = *req.Outcome
	}
	if req.Alternatives != nil {
		if err := decision.SetAlternativesSlice(*req.Alternatives); err != nil {
			return nil, err
		}
		updates["alternatives"] = decision.Alternatives
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ImportanceScore != nil {
		updates["importance_score"] = *req.ImportanceScore
	}

	if len(updates) > 0 {
		if err := r.db.Model(decision).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return r.GetByID(id)
}

// Delete deletes a decision
func (r *Repository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&TeamDecision{}).Error
}

// Approve approves, supersedes, or deprecates a decision
func (r *Repository) Approve(id, approverID uuid.UUID, status string) (*TeamDecision, error) {
	decision, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":      status,
		"approved_by": approverID,
		"approved_at": now,
	}

	if err := r.db.Model(decision).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

// SearchByText searches decisions by text query (title, description, rationale, outcome)
func (r *Repository) SearchByText(teamID uuid.UUID, query string, limit int) ([]TeamDecision, error) {
	var decisions []TeamDecision

	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	searchPattern := "%" + query + "%"
	err := r.db.Model(&TeamDecision{}).
		Where("team_id = ?", teamID).
		Where("(title ILIKE ? OR description ILIKE ? OR rationale ILIKE ? OR outcome ILIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Order("importance_score DESC, created_at DESC").
		Limit(limit).
		Find(&decisions).Error

	if err != nil {
		return nil, err
	}

	return decisions, nil
}

// CountByStatus returns counts of decisions by status for a team
func (r *Repository) CountByStatus(teamID uuid.UUID) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount
	err := r.db.Model(&TeamDecision{}).
		Select("status, COUNT(*) as count").
		Where("team_id = ?", teamID).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int64)
	for _, rc := range results {
		countMap[rc.Status] = rc.Count
	}

	return countMap, nil
}

// ListBySourceType retrieves decisions filtered by source type
func (r *Repository) ListBySourceType(teamID uuid.UUID, sourceType string, limit, offset int) ([]TeamDecision, int64, error) {
	var decisions []TeamDecision
	var total int64

	query := r.db.Model(&TeamDecision{}).Where("team_id = ? AND source_type = ?", teamID, sourceType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}

	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&decisions).Error; err != nil {
		return nil, 0, err
	}

	return decisions, total, nil
}

// GetPendingForTeam returns all pending decisions for a team (useful for AI extraction review)
func (r *Repository) GetPendingForTeam(teamID uuid.UUID) ([]TeamDecision, error) {
	var decisions []TeamDecision
	err := r.db.Model(&TeamDecision{}).
		Where("team_id = ? AND status = ?", teamID, string(DecisionStatusPending)).
		Order("importance_score DESC, created_at DESC").
		Find(&decisions).Error

	if err != nil {
		return nil, err
	}

	return decisions, nil
}
