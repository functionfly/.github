# Database Migrations

This project uses an embedded SQL migration system that compiles migration files directly into the Go binary.

## Directory Structure

```
internal/storage/sql/migrations/
├── 20240101000000_initial_schema.up.sql
├── 20240101000000_initial_schema.down.sql
├── 20260214141003_add_user_roles.up.sql
└── 20260214141003_add_user_roles.down.sql
```

## Migration Files

Each migration consists of two files:

- `VERSION_DESCRIPTION.up.sql` - SQL to apply the migration
- `VERSION_DESCRIPTION.down.sql` - SQL to rollback the migration

### Naming Convention

- **Version**: `YYYYMMDDHHMMSS` format (timestamp)
- **Description**: lowercase with underscores instead of spaces
- **Extensions**: `.up.sql` for applying, `.down.sql` for rolling back

## CLI Commands

The migration CLI tool provides several commands:

### Build the CLI

```bash
go build -o migrate cmd/migrate/main.go
```

### Create New Migration

```bash
./migrate create -desc "add user roles"
```

This creates two files:

- `internal/storage/sql/migrations/YYYYMMDDHHMMSS_add_user_roles.up.sql`
- `internal/storage/sql/migrations/YYYYMMDDHHMMSS_add_user_roles.down.sql`

### Run Migrations

```bash
./migrate run
```

Applies all pending migrations to the database.

### Check Migration Status

```bash
./migrate status
```

Shows which migrations are applied and which are pending.

## Writing Migrations

### Up Migrations

Write SQL that transforms the database to the new state:

```sql
-- Add a new column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user';

-- Create a new table
CREATE TABLE user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add an index
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
```

### Down Migrations

Write SQL that reverses the up migration:

```sql
-- Reverse the changes in reverse order
DROP INDEX IF EXISTS idx_user_permissions_user_id;
DROP TABLE IF EXISTS user_permissions;
ALTER TABLE users DROP COLUMN IF EXISTS role;
```

## Best Practices

1. **Always provide down migrations** - Even if rollback is rare, it's essential for development
2. **Use transactions** - The system wraps each migration in a transaction automatically
3. **Test migrations** - Test both up and down migrations on a copy of production data
4. **Keep migrations small** - Prefer multiple small migrations over one large one
5. **Use IF EXISTS/NOT EXISTS** - Make migrations idempotent where possible
6. **Version control** - Commit migration files with your code changes

## Rebuilding After Changes

After creating new migration files, rebuild the binary:

```bash
go build -o migrate cmd/migrate/main.go
```

The embedded files are compiled into the binary at build time.

## Database Setup

The system automatically creates the `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Error Handling

- Each migration runs in its own transaction
- If a migration fails, the transaction is rolled back
- Failed migrations are not recorded as applied
- Use `./migrate status` to see which migrations failed
