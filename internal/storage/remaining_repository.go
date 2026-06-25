package storage

import "database/sql"

// RemainingRepository handles the final batch of FWOS database operations.
type RemainingRepository struct {
	db *sql.DB
}

// NewRemainingRepository creates a new Remaining repository.
func NewRemainingRepository(db *sql.DB) *RemainingRepository {
	return &RemainingRepository{db: db}
}
