# Supabase Migration Guide

This guide provides a comprehensive strategy for migrating your FunctionFly application from PostgreSQL to Supabase while maintaining data integrity and adding real-time capabilities.

## Overview

Your current application uses a traditional PostgreSQL database with custom migrations. Supabase provides PostgreSQL with additional features like Row Level Security (RLS), real-time subscriptions, and built-in authentication.

## Migration Strategy

### Phase 1: Database Setup

#### 1.1 Create Supabase Project

1. Go to [supabase.com](https://supabase.com) and create a new project
2. Choose your organization and project name
3. Select the region closest to your users
4. Set a secure database password
5. Wait for the project to be fully provisioned

#### 1.2 Configure Environment Variables

Add these to your `.env` file (create `.env.supabase` for Supabase-specific config):

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key

# Optional: For gradual migration
USE_SUPABASE=false  # Set to true when ready to switch
```

#### 1.3 Run Supabase Migrations

Execute the migration files in order:

```bash
# Connect to your Supabase database (get connection string from Supabase dashboard)
psql "postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"

# Run migrations in order
\i supabase/migrations/20240217000000_initial_user_schema.sql
\i supabase/migrations/20240217000001_row_level_security.sql
\i supabase/migrations/20240217000002_realtime_setup.sql
```

### Phase 2: Data Migration

#### 2.1 Export Current Data

Create a script to export your current PostgreSQL data:

```sql
-- Export users
COPY (
    SELECT
        id, tenant_id, email, password_hash, role,
        email_verified, verification_token, verification_expires_at,
        provider, provider_id, provider_data,
        created_at, updated_at
    FROM users
) TO '/tmp/users.csv' WITH CSV HEADER;

-- Export tenants
COPY (SELECT id, name, plan, status, created_at, updated_at FROM tenants)
TO '/tmp/tenants.csv' WITH CSV HEADER;

-- Export other tables as needed...
```

#### 2.2 Transform and Import Data

Create user profiles for existing users and import data:

```sql
-- Import tenants first
COPY tenants (id, name, plan, status, created_at, updated_at)
FROM '/tmp/tenants.csv' WITH CSV HEADER;

-- Import users
COPY users (
    id, tenant_id, email, password_hash, role,
    email_verified, verification_token, verification_expires_at,
    provider, provider_id, provider_data,
    created_at, updated_at
)
FROM '/tmp/users.csv' WITH CSV HEADER;

-- Create profiles for existing users
INSERT INTO user_profiles (id, user_id, timezone, preferences, created_at, updated_at)
SELECT
    gen_random_uuid(),
    id,
    'UTC',
    '{}',
    created_at,
    updated_at
FROM users;
```

#### 2.3 Data Validation

Verify data integrity:

```sql
-- Check user counts match
SELECT COUNT(*) FROM users;

-- Verify profiles were created
SELECT COUNT(*) FROM user_profiles;

-- Check for any data issues
SELECT * FROM users WHERE tenant_id NOT IN (SELECT id FROM tenants);
```

### Phase 3: Application Integration

#### 3.1 Add Supabase Configuration

Update your application configuration to support both databases:

```go
type Config struct {
    // Existing config...
    Database *DatabaseConfig

    // New Supabase config
    Supabase *supabase.Config
    UseSupabase bool
}
```

#### 3.2 Create Supabase Client Factory

```go
func NewSupabaseClient(config *Config) (*supabase.Client, error) {
    if !config.UseSupabase {
        return nil, nil // Return nil to use existing database
    }

    return supabase.NewClient(&supabase.Config{
        URL:       config.Supabase.URL,
        AnonKey:   config.Supabase.AnonKey,
        ServiceKey: config.Supabase.ServiceKey,
    })
}
```

#### 3.3 Update Repository Layer

Modify your existing repositories to support both databases:

```go
type UserRepository struct {
    postgres *PostgresUserRepo
    supabase *supabase.UserRepository
    useSupabase bool
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
    if r.useSupabase && r.supabase != nil {
        return r.supabase.GetUserByID(ctx, id)
    }
    return r.postgres.GetUserByID(ctx, id)
}
```

#### 3.4 Add Real-time Subscriptions

Initialize real-time handlers in your application startup:

```go
func initRealtimeHandlers(app *App) error {
    if !app.Config.UseSupabase {
        return nil
    }

    handler := supabase.NewRealtimeHandler(app.SupabaseClient)

    // Subscribe to tenant-wide user events
    err := handler.SubscribeToTenantUsers(context.Background(), tenantID, func(event supabase.RealtimeEvent) {
        // Handle real-time user events
        app.BroadcastToWebSocketClients(event)
    })
    if err != nil {
        return fmt.Errorf("failed to subscribe to tenant users: %w", err)
    }

    app.RealtimeHandler = handler
    return nil
}
```

### Phase 4: Real-time Features

#### 4.1 WebSocket Integration

Update your WebSocket handlers to support real-time events:

```go
func (h *WebSocketHandler) HandleConnection(conn *websocket.Conn, userID uuid.UUID) {
    client := supabase.NewWebSocketConnection(conn, &userID)

    // Subscribe to user-specific events
    h.realtimeHandler.SubscribeToUserNotifications(context.Background(), userID, func(event supabase.RealtimeEvent) {
        client.SendEvent(event)
    })

    h.realtimeHandler.SubscribeToUserProfile(context.Background(), userID, func(event supabase.RealtimeEvent) {
        client.SendEvent(event)
    })

    client.Start()
}
```

#### 4.2 Frontend Integration

Update your React dashboard to handle real-time events:

```typescript
// Use Supabase real-time in your React components
import { useEffect } from 'react';
import { supabase } from './supabaseClient';

export function useRealtimeNotifications(userId: string) {
  useEffect(() => {
    const channel = supabase
      .channel(`user_${userId}_notifications`)
      .on('broadcast', { event: 'new_notification' }, (payload) => {
        // Handle new notification
        showNotification(payload);
      })
      .subscribe();

    return () => {
      supabase.removeChannel(channel);
    };
  }, [userId]);
}
```

### Phase 5: Testing and Rollback

#### 5.1 Dual Database Testing

Test both database systems simultaneously:

```go
func TestDualDatabaseConsistency(t *testing.T) {
    // Create test data in both databases
    // Verify results are identical
    // Test real-time features
}
```

#### 5.2 Gradual Rollout

Implement feature flags for gradual rollout:

```go
// Feature flag for real-time notifications
if app.Config.EnableRealtimeNotifications {
    // Use Supabase real-time
} else {
    // Use polling or existing system
}
```

#### 5.3 Rollback Strategy

Prepare rollback procedures:

```sql
-- If rollback needed, export from Supabase and import back to PostgreSQL
-- Keep both databases in sync during transition period
```

### Phase 6: Production Deployment

#### 6.1 Environment Setup

Configure production environment variables:

```bash
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
USE_SUPABASE=true
```

#### 6.2 Monitoring

Set up monitoring for Supabase:

- Monitor database performance
- Set up alerts for real-time subscription failures
- Monitor RLS policy effectiveness
- Track real-time event throughput

#### 6.3 Backup Strategy

Configure Supabase backups:

- Enable automated backups
- Test backup restoration
- Document backup and recovery procedures

## Migration Checklist

- [ ] Supabase project created
- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] Data exported from current database
- [ ] Data transformed and imported to Supabase
- [ ] Data validation completed
- [ ] Application code updated for Supabase integration
- [ ] Real-time subscriptions implemented
- [ ] WebSocket handlers updated
- [ ] Frontend real-time features implemented
- [ ] Testing completed in staging
- [ ] Gradual rollout to production
- [ ] Monitoring and alerts configured
- [ ] Rollback procedures documented

## Benefits After Migration

1. **Real-time Features**: Instant notifications and live updates
2. **Built-in Security**: Row Level Security policies
3. **Scalability**: Supabase handles connection pooling and scaling
4. **Developer Experience**: Rich client libraries and dashboard
5. **Authentication**: Built-in user management and OAuth
6. **Edge Functions**: Serverless functions for background tasks

## Common Challenges and Solutions

### Challenge: Complex RLS Policies

**Solution**: Start with simple policies and gradually add complexity. Test thoroughly.

### Challenge: Real-time Performance

**Solution**: Use appropriate subscription scopes and implement client-side throttling.

### Challenge: Data Synchronization

**Solution**: Implement dual writes during transition and validate data consistency.

### Challenge: Migration Downtime

**Solution**: Use blue-green deployment with gradual traffic shifting.

## Support and Resources

- [Supabase Documentation](https://supabase.com/docs)
- [Supabase Go Client](https://github.com/supabase-community/supabase-go)
- [Supabase Real-time Guide](https://supabase.com/docs/guides/realtime)
- [Row Level Security](https://supabase.com/docs/guides/auth/row-level-security)
