# FunctionFly URL System Migration Plan

## Executive Summary

This document outlines the complete migration from the current URL structure to the new `@username`-based identity system that provides cleaner, more professional profile URLs and consistent function namespaces.

---

## Current State Analysis

### Existing URL Patterns
```
Current Profile URLs:
- GET /v1/users/{username}           → Public profile API
- GET /v1/users/me                  → Current user profile

Current Function URLs:
- GET /v1/fx/{author}/{name}         → Function page (docs + playground)  
- POST /v1/fx/{author}/{name}        → Execute function
- GET /v1/run/{author}/{name}        → Live playground
- POST /v1/run/{author}/{name}/execute → Execute in playground
- GET /v1/registry/functions/{author}/{name} → Registry API

Current Legacy URLs:
- GET /v1/playground/{author}/{name} → Legacy playground
- GET /v1/docs/{author}/{name}       → Function docs
```

---

## Target URL Structure

### Public Identity URLs (Primary)

```
Profile:     functionfly.com/@username
Example:     functionfly.com/@functionfly
```

```
Function:    functionfly.com/@username/v1/fx/function-name
Example:     functionfly.com/@stripe-tools/slugify
```

```
Versioned:   functionfly.com/@username/v1/fx/function-name/v/1.2.0
Example:     functionfly.com/@stripe-tools/slugify/v/1.2.0
```

### Dashboard URLs (Separate Namespace)

```
Dashboard:   account.functionfly.com/settings
Functions:   account.functionfly.com/functions  
Billing:     account.functionfly.com/billing
```

### Execution URLs (High-Traffic)

```
Execute:     run.functionfly.com/@username/function-name
Execute v:   run.functionfly.com/@username/function-name@v1.2.0
Replay:      run.functionfly.com/replay/{execution_id}
```

### Registry API URLs

```
Function:    registry.functionfly.com/@username/function-name
Versions:    registry.functionfly.com/@username/function-name/versions
Stats:       registry.functionfly.com/@username/function-name/stats
```

### Documentation URLs

```
Docs:        docs.functionfly.com/@username/function-name
OpenAPI:     docs.functionfly.com/@username/function-name/openapi.json
```

---

## Migration Components

### 1. Backend API Routing Changes

#### Phase 1: Add New Routes (Non-Breaking)

```go
// New routes to add in internal/api/routes.go

// Profile routes with @ prefix
api.HandleFunc("/@/{username}", usersHandler.HandleGetPublicProfileByAt).Methods("GET")

// Function routes with @ prefix
api.HandleFunc("/@/{username}/v1/fx/{functionName}", registryPlaygroundHandler.HandleFunctionPageAt).Methods("GET")
api.HandleFunc("/@/{username}/v1/fx/{functionName}/execute", registryHandler.HandleExecuteAt).Methods("POST")
api.HandleFunc("/@/{username}/v1/fx/{functionName}/v/{version}", registryPlaygroundHandler.HandleFunctionPageAtVersion).Methods("GET")

// Function version routes
api.HandleFunc("/@/{username}/v1/fx/{functionName}/versions", registryHandler.HandleListVersionsAt).Methods("GET")
api.HandleFunc("/@/{username}/v1/fx/{functionName}/stats", registryHandler.HandleGetFunctionStatsAt).Methods("GET")

// Embed routes with @ prefix  
api.HandleFunc("/embed/@/{username}/{nameVersion}", registryHandler.HandleServeEmbedAt).Methods("GET")
```

#### Phase 2: Update Route Handlers

- Modify [`internal/api/handlers/users/users.go`](internal/api/handlers/users/users.go) to handle both `{username}` and `@{username}` patterns
- Modify [`internal/api/handlers/registry/registry.go`](internal/api/handlers/registry/registry.go) to accept new URL patterns
- Update [`internal/api/handlers/registry/playground.go`](internal/api/handlers/registry/playground.go) for `@username` routing

#### Phase 3: Add Username Validation Middleware

Create new middleware in [`internal/api/middleware/`](internal/api/middleware/):

```go
// username_validation.go
func ValidateUsernameFormat(username string) bool {
    // Must start with @ when passed in URL
    // After @: lowercase alphanumeric, hyphens, underscores
    // Length: 3-30 characters
    // Cannot be reserved name
}
```

### 2. Reserved Username System

#### Implementation Location
New file: [`internal/api/middleware/reserved_usernames.go`](internal/api/middleware/reserved_usernames.go)

#### Reserved Names List (Priority Order)

```go
var ReservedUsernames = []string{
    // Platform names
    "functionfly", "function", "flypy", "registry", "api",
    
    // System accounts  
    "system", "admin", "support", "root", "nobody",
    
    // Dashboard routes
    "account", "dashboard", "billing", "settings", "login",
    "logout", "signup", "register", "auth",
    
    // Feature routes
    "run", "play", "docs", "blog", "market", "enterprise",
    "security", "trust", "core", "debug", "status",
    
    // API paths
    "v1", "v2", "v3", "latest",
    
    // Common reserved
    "www", "mail", "ftp", "localhost", "static",
    
    // OAuth providers (for future)
    "google", "github", "twitter", "facebook",
}
```

#### Database Constraints

Add unique constraint exception handling in user creation:

```sql
-- In user creation, check against reserved list
INSERT INTO users (id, username, ...) 
VALUES ($1, $2, ...)
WHERE NOT EXISTS (
    SELECT 1 FROM reserved_usernames WHERE username = $2
);
```

### 3. Database Schema Updates

#### New Tables

```sql
-- Reserved usernames table
CREATE TABLE reserved_usernames (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(30) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- User verification badges
ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN verification_badge VARCHAR(50);
ALTER TABLE users ADD COLUMN trust_score INTEGER DEFAULT 0;

-- Profile enhancement fields
ALTER TABLE users ADD COLUMN total_calls BIGINT DEFAULT 0;
ALTER TABLE users ADD COLUMN profile_views BIGINT DEFAULT 0;
```

#### Username Validation Updates

In [`internal/storage/`](internal/storage/) repository:

```go
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
    // Validate username format
    if !isValidUsernameFormat(user.Username) {
        return errors.New("invalid username format")
    }
    
    // Check reserved names
    if isReservedUsername(user.Username) {
        return errors.New("username is reserved")
    }
    
    // Proceed with creation
}
```

### 4. Frontend Changes

#### New Page Routes

Create new Astro pages in [`web/site/src/pages/`](web/site/src/pages/):

```astro
---
// src/pages/@[username].astro
// Profile page with @ prefix
---

// src/pages/@[username]/v1/fx/[function].astro  
// Function page with @ prefix
```

#### Existing Page Updates

Update [`web/site/src/pages/index.astro`](web/site/src/pages/index.astro) and related files to:

1. Use new `@username` URLs for profile links
2. Update function cards to link to `/@username/v1/fx/function-name`
3. Add Open Graph tags for social sharing

#### API Client Updates

Update [`web/site/src/lib/api.ts`](web/site/src/lib/api.ts):

```typescript
// New API methods
export async function getProfile(username: string) {
  return fetch(`/v1/@/${username}`).then(r => r.json());
}

export async function getFunction(username: string, functionName: string) {
  return fetch(`/v1/@/${username}/v1/fx/${functionName}`).then(r => r.json());
}
```

### 5. SEO Enhancement Strategy

#### Profile Page SEO

Enhance public profile responses to include SEO metadata:

```json
{
  "username": "stripe-tools",
  "name": "Stripe Tools Suite", 
  "bio": "Production-ready payment and billing",
  "is functionsVerified": true,
  "trustScore": 95,
  "totalCalls": 1250000,
  "functions": [...],
  "tags": ["payments", "billing", "stripe", "api"],
  "ogImage": "https://functionfly.com/og/stripe-tools.png",
  "schema": {
    "@type": "SoftwareApplication",
    "name": "Stripe Tools",
    "description": "..."
  }
}
```

#### Structured Data Markup

Add JSON-LD to profile pages:

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Person",
  "name": "Stripe Tools",
  "url": "https://functionfly.com/@stripe-tools",
  "description": "Production-ready payment functions",
  "applicationCategory": "DeveloperApplication"
}
</script>
```

### 6. URL Mapping & Backwards Compatibility

#### Redirect Strategy

Implement 301 redirects for SEO preservation:

```go
// In internal/api/routes.go - add at end of route setup
func (s *Server) setupRedirects() {
    // Old -> New redirects (301 Permanent Redirect)
    
    // /v1/users/{username} -> /@{username}
    s.router.HandleFunc("/v1/users/{username}", func(w http.ResponseWriter, r *http.Request) {
        username := mux.Vars(r)["username"]
        http.Redirect(w, r, "/@"+username, http.StatusMovedPermanently)
    })
    
    // /v1/fx/{author}/{name} -> /@{author}/v1/fx/{name}
    s.router.HandleFunc("/v1/fx/{author}/{name}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        http.Redirect(w, r, fmt.Sprintf("/@%s/v1/fx/%s", vars["author"], vars["name"]), http.StatusMovedPermanently)
    })
    
    // /v1/run/{author}/{name} -> https://run.functionfly.com/@{author}/{name}
    s.router.HandleFunc("/v1/run/{author}/{name}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        http.Redirect(w, r, fmt.Sprintf("https://run.functionfly.com/@%s/%s", vars["author"], vars["name"]), http.StatusMovedPermanently)
    })
}
```

#### API Versioning Consideration

Keep v1 API routes functional but mark as deprecated:

```go
// Mark as deprecated but functional
api.HandleFunc("/v1/users/{username}", usersHandler.HandleGetPublicProfileDeprecated).Methods("GET")
// Add deprecation header
w.Header().Set("Deprecation", "true")
w.Header().Set("Link", "</v1/@/{username}>; rel=\"successor-version\"")
```

### 7. Infrastructure Updates

#### CDN/Proxy Configuration

Update [`deploy/caddy/Caddyfile`](deploy/caddy/Caddyfile) for new subdomain routing:

```caddy
# Main domain - profile and function pages
functionfly.com {
    reverse_proxy localhost:8080
    
    # New @username routes
    @atUser {
        path /@*
    }
    handle @atUser {
        reverse_proxy localhost:8080
    }
}

# Dashboard subdomain  
account.functionfly.com {
    reverse_proxy localhost:8081
    # Require authentication
}

# Execution subdomain (high traffic optimization)
run.functionfly.com {
    reverse_proxy localhost:8082
    # Cache headers for responses
    header Cache-Control "public, max-age=60"
}

# Registry API subdomain
registry.functionfly.com {
    reverse_proxy localhost:8080
    # API rate limiting
    rate_limit 100r/m
}

# Docs subdomain
docs.functionfly.com {
    reverse_proxy localhost:8080
}
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [ ] Implement reserved username system
- [ ] Add database schema updates
- [ ] Create username validation middleware
- [ ] Add new API routes (non-functional first)

### Phase 2: Backend Integration (Week 2)
- [ ] Connect new routes to handlers
- [ ] Update profile response format with SEO fields
- [ ] Implement 301 redirects
- [ ] Test all new URL patterns

### Phase 3: Frontend Development (Week 3)
- [ ] Create new @username profile pages
- [ ] Update existing function pages for new URLs
- [ ] Update API client
- [ ] Add Open Graph tags

### Phase 4: Infrastructure (Week 4)
- [ ] Update Caddy configuration for subdomains
- [ ] Set up new DNS records
- [ ] Configure SSL certificates
- [ ] Load test new infrastructure

### Phase 5: Migration & Testing (Week 5)
- [ ] Deploy to staging
- [ ] Run migration scripts
- [ ] Test all redirects
- [ ] Verify SEO metadata
- [ ] Performance testing

### Phase 6: Production Launch (Week 6)
- [ ] Deploy to production
- [ ] Monitor 301 redirects
- [ ] Track profile page analytics
- [ ] Update documentation

---

## Success Metrics

1. **URL Adoption**: % of users using new @username URLs within 3 months
2. **SEO Performance**: Search ranking improvement for profile pages
3. **Redirect Health**: <1% error rate on 301 redirects
4. **Performance**: <100ms additional latency from new routing layer
5. **User Satisfaction**: Positive feedback on cleaner URLs

---

## Risk Mitigation

1. **Backward Compatibility**: Maintain old URLs with 301 redirects for minimum 12 months
2. **Reserved Name Conflicts**: Graceful handling for existing users with reserved names
3. **Crawler Preservation**: Submit sitemap with new URLs to search engines
4. **Database Performance**: Index username column for fast lookups
5. **Cache Invalidation**: Ensure CDN properly caches new URL patterns

---

## Related Documentation

- [API Routes Documentation](../internal/api/routes.go)
- [User Handler Implementation](../internal/api/handlers/users/users.go)  
- [Registry Handler Implementation](../internal/api/handlers/registry/registry.go)
- [Frontend Pages Structure](../web/site/src/pages/)
- [Deployment Configuration](../deploy/caddy/Caddyfile)
