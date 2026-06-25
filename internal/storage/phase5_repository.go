package storage

import "database/sql"

// Phase5Repository handles FWOS Phase 5 database operations.
type Phase5Repository struct {
	db *sql.DB
}

// NewPhase5Repository creates a new Phase 5 repository.
func NewPhase5Repository(db *sql.DB) *Phase5Repository {
	return &Phase5Repository{db: db}
}
