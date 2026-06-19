# Database Migration Rollback Guide

This document describes how to safely roll back database migrations in FunctionFly.

## Overview

FunctionFly uses golang-migrate for database migrations. Each migration has:
- `.up.sql` - The forward migration (applies changes)
- `.down.sql` - The rollback migration (reverts changes)

## Before You Begin

### Prerequisites
- Access to the database (via `DATABASE_URL` or `DB_HOST`/`DB_USER`/`DB_PASSWORD`/`DB_NAME`)
- Migration files present in `migrations/` directory
- Adequate disk space for backup

### Important Warnings

> **⚠️ WARNING: Always test rollbacks in a non-production environment first!**

> **⚠️ WARNING: Never roll back in production without a verified database backup!**

> **⚠️ WARNING: Some migrations cannot be rolled back (e.g., data deletions, cascade drops). Check the `.down.sql` file before proceeding.**

## Rollback Procedures

### 1. Check Migration Status

Before rolling back, check the current migration state:

```bash
# Show detailed migration status
./scripts/rollback-migration.sh status

# List all available migrations
./scripts/rollback-migration.sh list
```

### 2. Create a Database Backup

**Always backup before rolling back:**

```bash
# Full database backup
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > backup_before_rollback_$(date +%Y%m%d_%H%M%S).sql

# Or use the project's backup script
./scripts/backup-database.sh
```

### 3. Preview the Rollback (Dry Run)

Test what will happen without applying changes:

```bash
DRY_RUN=1 ./scripts/rollback-migration.sh down
```

This shows the SQL that would be executed without actually running it.

### 4. Apply the Rollback

#### Roll Back One Migration

```bash
# Roll back the last applied migration
./scripts/rollback-migration.sh down
```

#### Roll Back to Specific Version

```bash
# Roll back all migrations to a specific version
./scripts/rollback-migration.sh version 20250101000000
```

### 5. Verify the Rollback

After rolling back:

```bash
# Check migration status
./scripts/rollback-migration.sh status

# Verify application functionality
curl https://your-api.com/api/health
```

## Using golang-migrate Directly

For more control, use golang-migrate commands directly:

```bash
# Install migrate CLI (if not already installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Roll back last migration
migrate -path ./migrations -database $DATABASE_URL down 1

# Roll back to specific version
migrate -path ./migrations -database $DATABASE_URL down -version 20250101000000

# Force to specific version (if down migrations are missing)
migrate -path ./migrations -database $DATABASE_URL force 20250101000000
```

## Common Rollback Scenarios

### Scenario 1: Migration Failed Mid-Way

If a migration failed during application:

1. Check what was applied:
```bash
psql $DATABASE_URL -c "SELECT * FROM schema_migrations ORDER BY applied_at DESC LIMIT 5;"
```

2. Identify the broken migration

3. Roll back to the version before the failure:
```bash
./scripts/rollback-migration.sh version <previous_version>
```

### Scenario 2: Migration Causes Application Errors

If a migration passes but breaks the application:

1. **Immediately:**
```bash
# Roll back the problematic migration
./scripts/rollback-migration.sh down
```

2. **Investigate:**
   - Review the migration SQL
   - Check application logs
   - Test in staging environment

3. **Fix and re-migrate properly**

### Scenario 3: Missing Down Migration

If a migration doesn't have a `.down.sql` file:

1. **Check if data loss is acceptable**

2. **Options:**
   - Manually write a compensating migration
   - Use `migrate force` to mark the version as applied without running it
   - Restore from backup

```bash
# Force to a version without running its up migration
migrate -path ./migrations -database $DATABASE_URL force 20250101000000
```

## Rollback Safety Checklist

Before performing any rollback in production:

- [ ] Database backup created and verified
- [ ] Staging environment tested rollback
- [ ] Application downtime window communicated
- [ ] Rollback command tested in dry-run mode
- [ ] Someone is designated to monitor after rollback
- [ ] Rollback plan documented

## Automated Backup Before Migration

The recommended approach is to backup before ANY migration:

```bash
# Automated pre-migration backup
./scripts/backup-database.sh && ./scripts/rollback-migration.sh status
```

## Emergency Rollback Procedure

If a migration causes a production incident:

1. **Immediate action:**
```bash
# Stop the application to prevent further issues
# Scale down orchestrator
fly scale count 0 --app functionfly-orchestrator

# Roll back the migration
./scripts/rollback-migration.sh down

# Verify
./scripts/rollback-migration.sh status
```

2. **Restore from backup if rollback fails:**
```bash
# Drop existing database
psql $DATABASE_URL -c "DROP DATABASE $DB_NAME WITH (force)"

# Restore from backup
psql $DATABASE_URL < backup_before_migration.sql
```

3. **Notify team and start investigation**

## Monitoring After Rollback

After rolling back:

1. Check application health:
```bash
curl https://your-api.com/api/health
```

2. Monitor error rates:
```bash
# Check Prometheus metrics
curl https://your-prometheus/api/v1/query?query=functionfly_api_error_rate
```

3. Verify data integrity:
```bash
# Run application health checks
./scripts/db-health-check.sh
```

## Rollback Limitations

Some operations **cannot** be rolled back safely:

| Operation | Can Roll Back? | Notes |
|-----------|----------------|-------|
| `DROP TABLE` | ❌ | Data lost permanently |
| `DROP COLUMN` | ⚠️ | Only if down migration exists |
| `ALTER TABLE DROP` | ⚠️ | Only if down migration exists |
| `DELETE FROM` | ❌ | Data lost permanently |
| `TRUNCATE` | ❌ | Data lost permanently |
| `DROP INDEX` | ✅ | Usually safe |
| `ALTER TABLE ADD COLUMN` | ⚠️ | Only if down migration drops it |
| `CREATE INDEX` | ✅ | Safe - just re-create the index |

## Related Documentation

- [Deployment Guide](./DEPLOYMENT.md)
- [Database Backup Strategy](./FUNCTION_BACKUP_STRATEGY.md)
- [Disaster Recovery](./DISASTER_RECOVERY.md)
