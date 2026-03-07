# GoBetterAuth Integration

This package provides GoBetterAuth integration for the FunctionFly platform as part of Phase 1 of the Better Auth migration plan.

## Overview

GoBetterAuth is embedded directly into the Go backend, providing authentication services without the need for a separate service. This approach offers:

- **Single Language Stack**: Pure Go implementation
- **Better Performance**: No network overhead (in-process)
- **Type Safety**: Native Go types throughout
- **Simpler Deployment**: Single binary deployment

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Go Backend (Single Binary)                │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              API Routes (internal/api/routes.go)      │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                      │
│  ┌────────────────────▼─────────────────────────────────┐  │
│  │           GoBetterAuth (internal/auth/gba)            │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │  │
│  │  │  Config  │  │  OAuth   │  │ Session  │           │  │
│  │  │  Models  │  │  Manager │  │ Manager  │           │  │
│  │  └──────────┘  └──────────┘  └──────────┘           │  │
│  │  ┌──────────┐  ┌──────────┐                          │  │
│  │  │  Hooks   │  │ Handlers │                          │  │
│  │  │  Manager │  │          │                          │  │
│  │  └──────────┘  └──────────┘                          │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                      │
│  ┌────────────────────▼─────────────────────────────────┐  │
│  │                  GORM / PostgreSQL                    │  │
│  └─────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Features

### Phase 1 (Current)
- ✅ Email/Password authentication
- ✅ OAuth providers (GitHub, Google)
- ✅ Session management with cookies
- ✅ Multi-tenancy support
- ✅ JWT token generation
- ✅ Database migrations

### Future Phases
- ⏳ MFA (TOTP) plugin
- ⏳ WebAuthn/Passkeys plugin
- ⏳ SAML SSO plugin
- ⏳ Advanced session policies
- ⏳ IP allowlisting

## Configuration

### Environment Variables

```bash
# GoBetterAuth Master Switch
GBA_ENABLED=true

# Feature Flags (Gradual Migration)
GBA_LOGIN=false        # Set to true to use GoBetterAuth for login
GBA_REGISTER=false     # Set to true to use GoBetterAuth for registration
GBA_OAUTH=false        # Set to true to use GoBetterAuth for OAuth
GBA_SESSION=false      # Set to true to use GoBetterAuth for session validation

# OAuth Providers
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Session Configuration
SESSION_MAX_AGE=168h              # Session duration (7 days default)
SESSION_COOKIE_NAME=ff_session
SESSION_COOKIE_SECURE=false       # Set to true in production (HTTPS)
SESSION_COOKIE_HTTPONLY=true
SESSION_COOKIE_SAMESITE=Lax
```

### Gradual Migration Strategy

The feature flags allow you to migrate gradually:

1. **Start with registration**: Set `GBA_REGISTER=true` for new users
2. **Enable login**: Set `GBA_LOGIN=true` once registration is stable
3. **Enable OAuth**: Set `GBA_OAUTH=true` for social login
4. **Enable sessions**: Set `GBA_SESSION=true` for session validation
5. **Full migration**: Once all features work, remove legacy auth

## Usage

### Initialize GoBetterAuth

```go
import (
    "github.com/functionfly/functionfly/internal/auth/gba"
    "gorm.io/gorm"
)

// Initialize with database
func SetupAuth(db *gorm.DB) (*gba.Auth, error) {
    // Create config from environment
    cfg, err := gba.ConfigFromEnv(db)
    if err != nil {
        return nil, err
    }

    // Create auth instance
    auth, err := gba.New(cfg)
    if err != nil {
        return nil, err
    }

    // Run migrations
    if err := gba.Migrate(db); err != nil {
        return nil, err
    }

    return auth, nil
}
```

### Register Routes

```go
import (
    "github.com/functionfly/functionfly/internal/auth/gba"
    "github.com/gorilla/mux"
)

func SetupAuthRoutes(router *mux.Router, auth *gba.Auth) {
    handler := gba.NewHandler(auth)
    middleware := gba.NewMiddleware(auth)

    // Public routes
    router.HandleFunc("/v1/auth/sign-up", handler.HandleSignUp).Methods("POST")
    router.HandleFunc("/v1/auth/sign-in", handler.HandleSignIn).Methods("POST")
    router.HandleFunc("/v1/auth/sign-out", handler.HandleSignOut).Methods("POST")
    router.HandleFunc("/v1/auth/session", handler.HandleGetSession).Methods("GET")
    
    // OAuth routes
    router.HandleFunc("/v1/auth/oauth/init", handler.HandleOAuthInit).Methods("GET")
    router.HandleFunc("/v1/auth/callback/{provider}", handler.HandleOAuthCallback).Methods("GET")

    // Protected route example
    protected := router.PathPrefix("/v1/protected").Subrouter()
    protected.Use(middleware.RequireAuth)
    protected.HandleFunc("/data", handleProtectedData).Methods("GET")
}
```

### Use Middleware

```go
func handleProtectedData(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context
    userID, ok := gba.GetUserID(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get tenant ID from context
    tenantID, ok := gba.GetTenantID(r.Context())
    if !ok {
        http.Error(w, "Tenant context required", http.StatusBadRequest)
        return
    }

    // Handle request...
}
```

### Register Custom Hooks

```go
// Register a custom hook for tenant validation
auth.GetHooks().Register("before:signup", func(ctx context.Context, req *gba.HookRequest) error {
    // Custom validation logic
    if req.TenantID == uuid.Nil {
        return fmt.Errorf("tenant is required")
    }
    return nil
})

// Register post-signup hook
auth.GetHooks().Register("after:signup", func(ctx context.Context, req *gba.HookRequest) error {
    // Send welcome email, etc.
    return nil
})
```

## Database Schema

GoBetterAuth uses separate tables with the `gba_` prefix:

- `gba_users` - User accounts
- `gba_accounts` - OAuth provider accounts
- `gba_sessions` - Active sessions
- `gba_verification_tokens` - Email verification tokens
- `gba_tenants` - Tenant/organization data
- `gba_tenant_ip_allowlist` - IP allowlist entries
- `gba_auth_audit_logs` - Authentication audit trail

### Migration

Run migrations automatically on startup:

```go
if err := gba.Migrate(db); err != nil {
    log.Fatal("Failed to migrate auth tables:", err)
}
```

## Multi-tenancy

GoBetterAuth supports multi-tenancy through:

1. **Tenant ID in tables**: All user data includes `tenant_id`
2. **Header-based**: `X-Tenant-ID` header
3. **Subdomain-based**: Extract tenant from subdomain (e.g., `tenant.functionfly.com`)
4. **Hooks**: Custom validation and policies per tenant

### Tenant Context Extraction

The middleware extracts tenant context in order of priority:

1. `X-Tenant-ID` header
2. Subdomain from `Host` header
3. Existing session (for authenticated requests)

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/auth/sign-up` | POST | Register with email/password |
| `/v1/auth/sign-in` | POST | Login with email/password |
| `/v1/auth/sign-out` | POST | Logout current session |
| `/v1/auth/session` | GET | Get current session info |
| `/v1/auth/oauth/init?provider=` | GET | Initiate OAuth flow |
| `/v1/auth/callback/github` | GET | GitHub OAuth callback |
| `/v1/auth/callback/google` | GET | Google OAuth callback |

## Testing

```bash
# Run GoBetterAuth tests
cd internal/auth/gba
go test -v ./...

# Run with race detector
go test -race ./...
```

## Migration from Legacy Auth

See [plans/BETTER_AUTH_MIGRATION_PLAN.md](../../../plans/BETTER_AUTH_MIGRATION_PLAN.md) for the full migration strategy.

## Contributing

When adding features:

1. Update models in `models.go`
2. Add handlers in `handlers.go`
3. Update middleware if needed
4. Add migrations
5. Update tests
6. Update documentation

## References

- [Better Auth Migration Plan](../../../plans/BETTER_AUTH_MIGRATION_PLAN.md)
- [GoBetterAuth Library](https://pkg.go.dev/github.com/GoBetterAuth/go-better-auth/v2)