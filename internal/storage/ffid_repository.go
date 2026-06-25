package storage

import "database/sql"

// FFIDRepository handles FFID identity system database operations.
type FFIDRepository struct {
	db *sql.DB
}

// NewFFIDRepository creates a new FFID repository.
func NewFFIDRepository(db *sql.DB) *FFIDRepository {
	return &FFIDRepository{db: db}
}
