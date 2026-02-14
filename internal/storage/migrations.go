package storage

import (
	"database/sql"
	"fmt"
)

func RunMigrations(db *PostgresDB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db.DB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// TODO: Implement migration runner that reads from migrations/ directory
	// For now, just run the basic schema setup
	return runInitialSchema(db.DB)
}

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	return err
}

func runInitialSchema(db *sql.DB) error {
	schema := `
		-- Tenants and users
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Apps
		CREATE TABLE IF NOT EXISTS apps (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Backends
		CREATE TABLE IF NOT EXISTS backends (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			app_id UUID NOT NULL REFERENCES apps(id),
			provider VARCHAR(50) NOT NULL, -- 'workers', 'vercel', 'fly', 'deno-deploy'
			region VARCHAR(10) NOT NULL,
			url VARCHAR(500) NOT NULL,
			shared_secret VARCHAR(255) NOT NULL,
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Health checks
		CREATE TABLE IF NOT EXISTS health_checks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			backend_id UUID NOT NULL REFERENCES backends(id),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ok BOOLEAN NOT NULL,
			status_code INTEGER,
			latency_ms INTEGER,
			error_message TEXT
		);

		-- Circuit breaker state
		CREATE TABLE IF NOT EXISTS circuit_state (
			backend_id UUID PRIMARY KEY REFERENCES backends(id),
			state VARCHAR(20) NOT NULL DEFAULT 'closed', -- 'closed', 'open', 'half-open'
			since_ts TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			fail_count INTEGER DEFAULT 0,
			success_count INTEGER DEFAULT 0,
			last_failure_ts TIMESTAMP WITH TIME ZONE,
			last_success_ts TIMESTAMP WITH TIME ZONE
		);

		-- Routing events (for observability)
		CREATE TABLE IF NOT EXISTS routing_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			app_id UUID NOT NULL REFERENCES apps(id),
			backend_id UUID NOT NULL REFERENCES backends(id),
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			latency_ms INTEGER,
			outcome VARCHAR(20) NOT NULL, -- 'success', 'failure', 'timeout'
			request_id VARCHAR(255)
		);

		-- Indexes for performance
		CREATE INDEX IF NOT EXISTS idx_backends_app_id ON backends(app_id);
		CREATE INDEX IF NOT EXISTS idx_health_checks_backend_id ON health_checks(backend_id);
		CREATE INDEX IF NOT EXISTS idx_routing_events_app_id ON routing_events(app_id);
		CREATE INDEX IF NOT EXISTS idx_routing_events_timestamp ON routing_events(timestamp);
	`

	_, err := db.Exec(schema)
	return err
}