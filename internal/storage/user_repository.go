package storage

// UserRepository handles user-related database operations.
type UserRepository struct {
	db *PostgresDB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *PostgresDB) *UserRepository {
	return &UserRepository{db: db}
}
