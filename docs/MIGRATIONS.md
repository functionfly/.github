# Database Migrations Guide

This guide explains how to apply database migrations to your PostgreSQL databases (local development and Neon production).

## Quick Start

### Option 1: Using the Migration Tool (Recommended)

```bash
# Apply to local PostgreSQL only
go run cmd/apply-migrations/main.go -target=local

# Apply to Neon only
go run cmd/apply-migrations/main.go -target=neon

# Apply to both
make migrate

# Check migration status
go run cmd/apply-migrations/main.go -target=local -status
go run cmd/apply-migrations/main.go -target=neon -status

# Dry-run (show what would be applied without running)
go run cmd/apply-migrations/main.go -target=local -dry-run

# Skip validation (useful in development)
go run cmd/apply-migrations/main.go -target=local -skip-validation
```

### Option 2: Using the Orchestrator API

The orchestrator API automatically applies migrations on startup (unless `--skip-migrations` is provided):

```bash
# Start API with migrations
./bin/orchestrator-api

# Skip migrations (use existing schema)
./bin/orchestrator-api --skip-migrations
```

### Option 3: Using golang-migrate CLI

If you have the `migrate` CLI installed:

```bash
# Local PostgreSQL
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/functionfly?sslmode=disable" up

# Neon
migrate -path migrations -database "$DATABASE_URL" up
```

Install golang-migrate CLI:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# See: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
```

## Environment Setup

### Local PostgreSQL

Ensure these are set in your `.env` file:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=functionfly
DB_SSLMODE=disable
```

### Neon PostgreSQL

Ensure this is set (from Neon console):
```env
DATABASE_URL=postgresql://user:pass@neon-host.neon.tech/dbname?sslmode=require
```

## Migration Status

To check which migrations have been applied:

```bash
# Using psql (local)
psql -h localhost -U postgres -d functionfly -c "SELECT * FROM schema_migrations;"

# Using psql (Neon)
psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations;"
```

## Creating New Migrations

1. Create new migration files:
```bash
# Using golang-migrate CLI
migrate create -ext sql -dir migrations -seq my_new_feature

# Or manually create:
# migrations/000XXX_description.up.sql
# migrations/000XXX_description.down.sql
```

2. Write your SQL in the `.up.sql` file
3. Write rollback SQL in the `.down.sql` file (optional but recommended)

## Troubleshooting

### "Dirty" Migration State

If migrations fail and leave the database in a "dirty" state:

```bash
# Check current state
psql "$DATABASE_URL" -c "SELECT version, dirty FROM schema_migrations;"

# Force to a specific version (use with caution!)
migrate -path migrations -database "$DATABASE_URL" force <VERSION_NUMBER>
```

### Duplicate Sequence Numbers

If you encounter errors about duplicate migration sequence numbers:
```bash
# The migration system handles this automatically via the repair mechanism
# Just run migrations normally and they will be repaired if needed
```

### Connection Issues (Neon)

Neon uses PgBouncer which can cause issues with prepared statements. The migration tool handles this automatically by using a dedicated connection for migrations.

## Available Migrations

Check `migrations/` directory for all available migrations. Key recent migrations:

- `000244_create_cost_allocation_entries` - Detailed cost tracking for chargebacks
- Registry feature migrations
- Billing and usage tracking
- Security and authentication

## Migration Validation

After migrations run, the system performs validation checks:

1. Foreign key integrity
2. Required indexes exist
3. Required tables exist
4. Data integrity constraints

Validation failures are logged but don't stop the application in development mode (when `SKIP_MIGRATION_VALIDATION=true`).

## Rollback Testing

To test migration rollbacks:
```bash
export DB_TEST_MIGRATION_ROLLBACK=true
go run cmd/apply-migrations/main.go
```

This will test that the latest migration can be rolled back successfully on a temporary database.
