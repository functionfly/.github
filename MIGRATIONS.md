# Database Migration Policy

This document defines the migration naming convention and best practices for the FunctionFly project.

## Migration Naming Convention

**All new migrations MUST use the timestamp format:**

```
YYYYMMDDHHMMSS_description.sql
```

Example: `20260418160000_add_tax_compliance.up.sql`

### Why Timestamp Format?

1. **Prevents collisions**: Multiple developers can create migrations simultaneously without conflicts
2. **Chronological ordering**: Natural sort order matches execution order
3. **Industry standard**: Used by Rails, Django, golang-migrate, and others
4. **No gaps**: No need to "reserve" sequence numbers

## Creating a New Migration

Use the provided helper script:

```bash
./scripts/create-migration.sh "description_of_changes"
```

This will:
1. Generate timestamp based on current time
2. Create both `.up.sql` and `.down.sql` files
3. Add the standard header comment

Or manually:

```bash
timestamp=$(date +%Y%m%d%H%M%S)
touch migrations/${timestamp}_my_feature.up.sql
touch migrations/${timestamp}_my_feature.down.sql
```

## Validation

Run validation before committing:

```bash
./scripts/validate-migrations.sh
```

This checks for:
- Duplicate version numbers
- Missing `.down.sql` files
- Mixed naming conventions (warning)

## Historical Context

The project previously used two conventions:
- **Sequential**: `000XXX_description.sql` (first 194 migrations)
- **Timestamp**: `YYYYMMDDHHMMSS_description.sql` (newer migrations)

This mixed convention caused several issues that have been resolved:
- Duplicate sequence numbers (000250, 20260412175400) have been renamed
- All new migrations should use timestamp format

## Migration Content Guidelines

### Up Migrations
- Make changes idempotent where possible (`IF NOT EXISTS`, `IF EXISTS`)
- Include comments explaining the purpose
- Keep transactions small and focused

### Down Migrations
- Must exactly reverse the up migration
- Use `IF EXISTS` for drops to handle partial failures
- Test in a fresh database before committing

Example:
```sql
-- Migration: Add new column to users table
-- Purpose: Store user timezone preference

-- Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) DEFAULT 'UTC';

-- Down
ALTER TABLE users DROP COLUMN IF EXISTS timezone;
```

## Pre-commit Hook

Add to `.git/hooks/pre-commit` or `.pre-commit-config.yaml`:

```bash
#!/bin/bash
./scripts/validate-migrations.sh || exit 1
```

## CI/CD Integration

The validation script runs in CI to prevent merging duplicate migrations:

```yaml
# .github/workflows/migrations.yml or equivalent
- name: Validate Migrations
  run: ./scripts/validate-migrations.sh
```
