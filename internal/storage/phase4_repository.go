package storage

import "database/sql"

// Phase4Repository handles FWOS Phase 4 database operations.
type Phase4Repository struct {
	db *sql.DB
}

// NewPhase4Repository creates a new Phase 4 repository.
func NewPhase4Repository(db *sql.DB) *Phase4Repository {
	return &Phase4Repository{db: db}
}
