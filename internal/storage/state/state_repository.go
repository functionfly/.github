package state

import (
	"gorm.io/gorm"
)

// StateRepository handles state-related database operations using GORM
type StateRepository struct {
	db *gorm.DB
}

// NewStateRepository creates a new state repository
func NewStateRepository(db *gorm.DB) *StateRepository {
	return &StateRepository{db: db}
}