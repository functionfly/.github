# Neon Postgres Migration Plan

## Executive Summary

This document outlines the migration strategy from Supabase to Neon Postgres. The project already has native PostgreSQL support via `internal/storage/`, so the migration primarily involves:
1. Creating a Neon Postgres project
2. Configuring connection strings
3. Running existing migrations
4. Removing/disabling Supabase dependencies

## Current Architecture

### Existing Components

```mermaid
graph TB
    subgraph "Current Architecture"
        A[API Server] --> B{USE_SUPABASE}
        B -->|true| C[Supabase Client]
        B -->|false| D[PostgreSQL Client]
        C --> E[Supabase Database]
        D --> F[PostgreSQL Database]
        C --> G[Supabase Storage]
        C --> H[Supabase Realtime]
    end
```

### Supabase Dependencies Found

| Component | Location | Purpose |
|-----------|----------|---------|
| SupabaseRepository | `internal/repositories/supabase/` | Database access layer |
| Supabase Client | `internal/supabase/client.go` | API client |
| Storage Service | `internal/services/storage.go` | File storage |
| Realtime Handler | `internal/supabase/realtime.go` | Real-time events |
| Feature Flag | `USE_SUPABASE` env var | Switch between backends |

## Migration Strategy

### Phase 1: Neon Project Setup

1. **Create Neon Project**
   - Go to [Neon Console](https://console.neon.tech)
   - Create new project: `functionfly`
   - Select region closest to users
   - Save connection string

2. **Configure Connection String**
   ```
   postgresql://username:password@ep-{project-id}.us-east-1.aws.neon.tech/functionfly?sslmode=require
   ```

### Phase 2: Database Migration

The existing schema migrations in `internal/storage/sql/migrations/` are compatible with Neon Postgres (both are PostgreSQL 15+).

**Run migrations on Neon:**
```bash
# Using psql
psql "postgresql://user:pass@host/neon?sslmode=require" -f internal/storage/sql/migrations/20240101000000_initial_schema.up.sql
# ... run all migrations in order
```

Or use the migration CLI:
```bash
./migrate run
```

### Phase 3: Data Migration (If data exists in Supabase)

**Export from Supabase:**
```sql
-- Export all tables to CSV
COPY tenants TO '/tmp/tenants.csv' WITH CSV HEADER;
COPY users TO '/tmp/users.csv' WITH CSV HEADER;
COPY apps TO '/tmp/apps.csv' WITH CSV HEADER;
-- ... etc for all tables
```

**Import to Neon:**
```sql
-- Import in correct order (respecting foreign keys)
COPY tenants FROM '/tmp/tenants.csv' WITH CSV HEADER;
COPY users FROM '/tmp/users.csv' WITH CSV HEADER;
COPY apps FROM '/tmp/apps.csv' WITH CSV HEADER;
-- ... etc
```

### Phase 4: Application Configuration

**Update Environment Variables:**

In `.env`:
```bash
# Disable Supabase - use PostgreSQL
USE_SUPABASE=false

# Neon Postgres connection
DB_HOST=ep-xxx.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=your-username
DB_PASSWORD=your-password
DB_NAME=functionfly
DB_SSLMODE=require

# Remove or comment Supabase variables
# SUPABASE_URL=
# SUPABASE_ANON_KEY=
# SUPABASE_SERVICE_KEY=
```

**Update docker-compose.yml:**

Since Neon is cloud-hosted, remove the local PostgreSQL container:
```yaml
# Remove the postgres service - use Neon instead
# postgres:
#   image: postgres:15
#   ...

orchestrator-api:
  environment:
    - DB_HOST=ep-xxx.us-east-1.aws.neon.tech
    - DB_SSLMODE=require
```

### Phase 5: Remove Supabase Dependencies

**Code Changes Required:**

1. **Remove Supabase imports and initialization** in `internal/api/server.go`
2. **Optional: Remove Supabase packages** from `go.mod`
3. **Keep or remove** Supabase storage:
   - If keeping file storage: Configure alternative (S3, local, etc.)
   - If removing: Update artifact storage to use Redis/filesystem

## Implementation Checklist

- [x] 1. Create Neon project and obtain connection string
- [x] 2. Test Neon connection with `psql`
- [x] 3. Run all migrations on Neon database
- [ ] 4. Export data from Supabase (if applicable)
- [ ] 5. Import data to Neon (if applicable)
- [x] 6. Update environment configuration
- [ ] 7. Test application with Neon database
- [x] 8. Remove Supabase configuration
- [ ] 9. Update documentation
- [ ] 10. Monitor production after switch

## Environment Configuration Changes

### Current (.env.example)

```bash
# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=functionfly
DB_SSLMODE=disable

# Supabase (Frontend)
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=xxx
```

### After Migration

```bash
# Database - Neon Postgres
DB_HOST=ep-xxx.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=your-neon-user
DB_PASSWORD=your-neon-password
DB_NAME=functionfly
DB_SSLMODE=require

# Disable Supabase
USE_SUPABASE=false

# Supabase (Frontend - if still needed for auth UI)
VITE_SUPABASE_URL=https://xxx.supabase.co
VITE_SUPABASE_ANON_KEY=xxx
```

## Connection String Format for Neon

The existing `buildConnectionString()` function in `internal/storage/config.go` builds:
```
host=%s port=%d user=%s password=%s dbname=%s sslmode=%s
```

For Neon, set these environment variables:
- `DB_SSLMODE=require` (Neon requires SSL)
- `DB_HOST` = your Neon host (e.g., `ep-xxx.us-east-1.aws.neon.tech`)

## Rollback Plan

If migration fails:

1. **Revert environment variables** to point back to original database
2. **Keep Supabase project** active for at least 7 days
3. **Data recovery**: Re-export from Neon if needed, import to Supabase

## Frontend Supabase Dependencies

The React dashboard (`web/dashboard/`) has significant Supabase integration:

| Component | Files | Purpose |
|-----------|-------|---------|
| Auth UI | `web/dashboard/src/stores/authStore.ts` | User authentication |
| Realtime | `web/dashboard/src/hooks/useRealtime.ts` | Real-time notifications |
| Presence | `web/dashboard/src/hooks/usePresence.ts` | User presence tracking |
| Supabase Client | `web/dashboard/src/lib/supabase.ts` | Database queries |

**Frontend Changes Required:**
1. Replace Supabase Auth with your backend API authentication
2. Replace real-time subscriptions with WebSocket/polling alternative
3. Update API client to use your backend instead of Supabase

## Benefits of Neon Postgres

| Feature | Benefit |
|---------|---------|
| Serverless | Auto-scales to zero, cost-effective |
| Branching | Create preview databases for testing |
| Built-in SSL | Simplified security |
| PostgreSQL 15+ | Latest features and performance |
| Point-in-time恢复 | Restore to any point |

## Timeline Estimate

| Task | Effort |
|------|--------|
| Create Neon project | 10 min |
| Test connection | 10 min |
| Run migrations | 5 min |
| Data export/import | Varies (data size) |
| Configuration update | 15 min |
| Frontend auth migration | 2-4 hours |
| Frontend realtime migration | 1-2 hours |
| Testing | 30 min |
| **Total** | ~4-8 hours |

## Next Steps

1. Create Neon project at https://console.neon.tech
2. Run the migration CLI to apply schema
3. Test the connection in development
4. Migrate frontend authentication
5. Migrate frontend real-time features
6. Schedule production cutover
