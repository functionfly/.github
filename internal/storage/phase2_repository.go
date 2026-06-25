package storage

import "database/sql"

// Phase2Repository handles FWOS Phase 2 database operations.
type Phase2Repository struct {
	db *sql.DB
}

// NewPhase2Repository creates a new Phase 2 repository.
func NewPhase2Repository(db *sql.DB) *Phase2Repository {
	return &Phase2Repository{db: db}
}
