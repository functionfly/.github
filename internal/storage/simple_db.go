package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// OpenSimpleDB opens a minimal PostgreSQL connection for CLI/short-lived use.
// It uses the same env-based config as NewPostgresDB but skips health monitoring,
// prepared statements, and GORM so the process can complete and exit quickly.
// Caller must call db.Close() when done.
func OpenSimpleDB(ctx context.Context) (*sql.DB, error) {
	config, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}
	connStr := buildConnectionString(config)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(1 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
