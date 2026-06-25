package storage

import "database/sql"

// Phase3Repository handles FWOS Phase 3 database operations.
type Phase3Repository struct {
	db *sql.DB
}

// NewPhase3Repository creates a new Phase 3 repository.
func NewPhase3Repository(db *sql.DB) *Phase3Repository {
	return &Phase3Repository{db: db}
}
