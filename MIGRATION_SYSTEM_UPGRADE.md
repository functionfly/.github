# Migration System Upgrade: Custom → golang-migrate

## Overview

The project's database migration system has been upgraded from a custom implementation to the industry-standard [golang-migrate](https://github.com/golang-migrate/migrate) library.

## What Changed

### Before (Custom System)
- Embedded SQL files in Go binary (`go:embed`)
- Custom migration runner with timestamp-based versioning
- Manual migration state tracking in `schema_migrations` table
- Basic CLI commands (up, down, reset, status, version)

### After (golang-migrate)
- File-based migrations in `migrations/` directory
- Sequential numbering (000001, 000002, etc.)
- golang-migrate's robust state management
- Rich CLI with additional commands and features
- Industry-standard tooling and ecosystem

## Migration Details

### File Structure
```
migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_deployments.up.sql
├── 000002_add_deployments.down.sql
├── ...
├── 000037_function_capabilities.up.sql
├── 000037_function_capabilities.down.sql
└── 000038_migrate_state.up.sql
```

### Database Schema
The migration state table changed from:
```sql
-- Old format
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

To golang-migrate format:
```sql
-- New format
CREATE TABLE schema_migrations (
    version BIGINT NOT NULL PRIMARY KEY,
    dirty BOOLEAN NOT NULL DEFAULT FALSE
);
```

## Commands

### Basic Usage
```bash
# Run all pending migrations
go run cmd/migrate/main.go up

# Rollback latest migration
go run cmd/migrate/main.go down

# Show migration status
go run cmd/migrate/main.go status

# Show current version
go run cmd/migrate/main.go version

# Reset database (drop everything and re-run)
go run cmd/migrate/main.go reset
```

### Advanced Usage
```bash
# Create new migration
go run cmd/migrate/main.go create add_new_feature

# Direct golang-migrate usage (if needed)
export PATH=$PATH:$(go env GOPATH)/bin
migrate -path migrations -database "postgres://..." up
```

## Migration State Transition

The upgrade includes a special migration (`000038_migrate_state.up.sql`) that handles the transition from the old system to the new one. This migration:

1. Detects if an old `schema_migrations` table exists
2. Migrates the highest applied version to the new format
3. Backs up the old table as `schema_migrations_old`
4. Creates the new golang-migrate compatible table

## Benefits

### Better Tooling
- **Dry runs**: Test migrations without applying them
- **Force version**: Set migration version manually for recovery
- **Lock timeout**: Configurable database lock acquisition time
- **Prefetch**: Load multiple migrations in advance for performance

### Developer Experience
- **Standard format**: Compatible with existing golang-migrate tooling
- **Rich ecosystem**: Integrates with existing DevOps tools
- **Better error handling**: More robust transaction management
- **Schema validation**: Built-in migration file validation

### Production Ready
- **Dirty state detection**: Prevents corrupted migration states
- **Concurrent safety**: Database-level locking prevents race conditions
- **Rollback safety**: Safer rollback operations with better error reporting
- **Monitoring**: Better visibility into migration status

## Compatibility

### Go Code Changes
- `internal/storage/migrations.go`: Updated to use golang-migrate API
- `cmd/migrate/main.go`: Simplified to delegate to golang-migrate CLI
- All existing migration functions maintain the same interface for backward compatibility

### Migration Files
- All 37 existing migrations converted to sequential numbering
- Up/down pairs maintained
- SQL content preserved exactly
- Added state transition migration

## Troubleshooting

### Migration Fails
```bash
# Check migration status
go run cmd/migrate/main.go status

# Force a specific version (use carefully!)
export PATH=$PATH:$(go env GOPATH)/bin
migrate -path migrations -database "postgres://..." force VERSION
```

### Database Connection Issues
Ensure your `DATABASE_URL` environment variable is set:
```bash
export DATABASE_URL="postgres://user:pass@localhost/dbname?sslmode=disable"
```

### Rollback Issues
If a migration fails during rollback:
```bash
# Check which migrations are applied
go run cmd/migrate/main.go status

# Manually fix the issue, then force the version
migrate -path migrations -database "postgres://..." force VERSION
```

## Future Improvements

With golang-migrate, you can now easily add:

1. **Migration templates/scaffolding**
2. **Schema diffing tools**
3. **CI/CD integration**
4. **Multi-environment migration strategies**
5. **Migration testing frameworks**

## Rollback Plan

If you need to rollback to the old system:

1. The old `schema_migrations` table is backed up as `schema_migrations_old`
2. Restore the old Go code from git history
3. Restore the embedded migration files
4. Run `go run cmd/migrate/main.go reset` to clean up

The migration system upgrade is backward compatible and includes safety measures for rollbacks.