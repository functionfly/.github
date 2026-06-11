// Package storage (under a2a) provides the persistence layer for A2A
// agent cards. Task storage REUSES the existing registry_executions_public
// table — there is no a2a_tasks table.
package storage

import (
	"context"
	"time"

	"github.com/functionfly/functionfly/internal/a2a"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// A2ACardRepository manages agent_cards rows.
type A2ACardRepository struct {
	db *gorm.DB
}

// NewA2ACardRepository creates a new agent card repository.
func NewA2ACardRepository(db *gorm.DB) *A2ACardRepository {
	return &A2ACardRepository{db: db}
}

// GetCard returns a single agent card by ID.
func (r *A2ACardRepository) GetCard(_ context.Context, agentID string) (*a2a.AgentCardInfo, error) {
	var card a2a.AgentCardInfo
	if err := r.db.Where("id = ?", agentID).First(&card).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &card, nil
}

// ListCards returns paginated agent cards.
func (r *A2ACardRepository) ListCards(_ context.Context, limit, offset int) ([]a2a.AgentCardInfo, int, error) {
	var cards []a2a.AgentCardInfo
	var total int64

	if err := r.db.Model(&a2a.AgentCardInfo{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("trust_score DESC, published_at DESC").Limit(limit).Offset(offset).Find(&cards).Error; err != nil {
		return nil, 0, err
	}
	return cards, int(total), nil
}

// UpsertCard creates or updates an agent card.
func (r *A2ACardRepository) UpsertCard(_ context.Context, card *a2a.AgentCardInfo) error {
	card.UpdatedAt = time.Now().UTC()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"version", "name", "description", "url", "protocol_version", "capabilities", "skills", "auth_schemes", "input_modes", "output_modes", "peer_jwks_url", "updated_at"}),
	}).Create(card).Error
}

// DeleteCard removes an agent card.
func (r *A2ACardRepository) DeleteCard(_ context.Context, agentID string) error {
	return r.db.Where("id = ?", agentID).Delete(&a2a.AgentCardInfo{}).Error
}

// A2ATaskStore implements a2a.TaskStore using the registry_executions_public table.
type A2ATaskStore struct {
	db *gorm.DB
}

// NewA2ATaskStore creates a new task store.
func NewA2ATaskStore(db *gorm.DB) *A2ATaskStore {
	return &A2ATaskStore{db: db}
}

// UpdateTaskState updates the state column of a receipt.
func (s *A2ATaskStore) UpdateTaskState(_ context.Context, publicID string, fromState, to a2a.TaskState) error {
	result := s.db.Exec(
		`UPDATE registry_executions_public SET state = ? WHERE public_id = ? AND state = ? AND protocol = 'a2a'`,
		string(to), publicID, string(fromState),
	)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// GetTaskState returns the current state of a task receipt.
func (s *A2ATaskStore) GetTaskState(_ context.Context, publicID string) (a2a.TaskState, error) {
	var state string
	err := s.db.Raw(
		`SELECT state FROM registry_executions_public WHERE public_id = ? AND protocol = 'a2a'`,
		publicID,
	).Scan(&state).Error
	if err != nil {
		return "", err
	}
	return a2a.TaskState(state), nil
}

// SetTaskOutput sets the output column when a task completes.
func (s *A2ATaskStore) SetTaskOutput(_ context.Context, publicID string, output []byte) error {
	return s.db.Exec(
		`UPDATE registry_executions_public SET output_json = ?::jsonb WHERE public_id = ? AND protocol = 'a2a'`,
		string(output), publicID,
	).Error
}

// Ensure interfaces are satisfied.
var _ a2a.TaskStore = (*A2ATaskStore)(nil)
var _ a2a.CardRepository = (*A2ACardRepository)(nil)
