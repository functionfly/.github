package storage

import "database/sql"

// EmployeeRepository handles FWOS employee-related database operations.
type EmployeeRepository struct {
	db *sql.DB
}

// NewEmployeeRepository creates a new employee repository.
func NewEmployeeRepository(db *sql.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}
