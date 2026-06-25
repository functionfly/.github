package storage

import "database/sql"

// Phase6Repository handles FWOS Phase 6 database operations.
type Phase6Repository struct {
	db *sql.DB
}

// NewPhase6Repository creates a new Phase 6 repository.
func NewPhase6Repository(db *sql.DB) *Phase6Repository {
	return &Phase6Repository{db: db}
}
