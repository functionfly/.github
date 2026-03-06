# Standalone Admin Dashboard SPA - Implementation Plan

**Created:** March 5, 2026  
**Status:** Ready for Implementation  
**Estimated Duration:** 5-6 weeks  
**Complexity:** High

---

## Executive Summary

This plan details the migration of FunctionFly's admin dashboard from an embedded section within the main user dashboard to a fully standalone Single Page Application (SPA) with enhanced security features. The standalone admin dashboard will be deployed at `admin.functionfly.com` with strict security boundaries, dedicated authentication flow, and enhanced monitoring.

### Goals

1. **Security Isolation**: Separate admin dashboard from user dashboard for better security boundaries
2. **Separate Deployment**: Independent SPA deployable at a different subdomain (e.g., admin.functionfly.com)
3. **Enhanced Security**: Additional security layers beyond current implementation
4. **Zero User Impact**: Existing user dashboard remains unchanged
5. **Maintainability**: Clear separation of concerns

---

## Current State

### Existing Implementation
- Admin dashboard embedded within main user dashboard at `/web/dashboard`
- 26+ admin pages located in `/web/dashboard/src/pages/Admin*Page/`
- Admin routes: `/admin/*` (tenants, users, billing, audit, system, backends, providers, content, registry, state-fabric, feedback, etc.)
- Backend API routes: `/v1/admin/*` with RBAC permissions

### Security Features (Current)
- RequirePermission middleware (RBAC)
- RequireHMACSignature for sensitive operations
- MFA enforcement for admin users
- Rate limiting

### Admin Pages to Migrate (26+)
- AdminDashboardPage (overview, activity, revenue, quick stats)
- AdminTenantsPage & AdminTenantDetailPage
- AdminUsersPage & AdminUserDetailPage
- AdminBillingPage (tiers, subscriptions, invoices)
- AdminAuditPage
- AdminSystemPage (health, maintenance)
- AdminBackendsPage, AdminProvidersPage
- AdminContentPage, AdminContentCalendarPage, AdminNewsletterPage, AdminRedirectsPage
- AdminFunctionsPage, AdminFunctionDetailPage
- AdminRegistryPage, AdminStateFabricPage
- AdminFeedbackPage, AdminFeaturesPage, AdminStatusPage
- AdminTrustDashboardPage, AdminExecutionAuditPage, AdminFraudDetectionPage, AdminEconomicLeaderboardPage

---

## Architecture Design

### Directory Structure

```
├── web/
│   ├── dashboard/              # Existing user dashboard (UNCHANGED)
│   └── admin-dashboard/        # NEW standalone admin SPA
│       ├── src/
│       │   ├── pages/          # All 26+ admin pages
│       │   ├── components/
│       │   │   ├── layout/
│       │   │   ├── security/
│       │   │   └── common/
│       │   ├── stores/
│       │   │   ├── adminAuthStore.ts
│       │   │   ├── adminSessionStore.ts
│       │   │   └── adminAuditStore.ts
│       │   ├── hooks/
│       │   ├── lib/
│       │   │   ├── api/
│       │   │   │   ├── adminClient.ts
│       │   │   │   └── hmacSigner.ts
│       │   │   └── security/
│       │   ├── App.tsx
│       │   └── main.tsx
│       ├── Dockerfile
│       ├── package.json
│       └── vite.config.ts
│
├── deploy/
│   ├── caddy/
│   │   ├── admin.Caddyfile
│   │   └── admin-staging.Caddyfile
│   └── docker/
│       └── Dockerfile.admin-dashboard
│
└── docker-compose.admin.yml
```

### Technology Stack

**Frontend:**
- React 19+ with TypeScript
- Vite for build tooling
- React Router v7 for routing
- TanStack Query for data fetching
- Zustand for state management
- Radix UI + Tailwind CSS for UI
- Recharts for admin analytics
- React Hook Form + Zod for forms

**Security:**
- Cloudflare Access (Zero Trust) - **optional**
- Infisical for secrets management
- HMAC-SHA256 request signing
- Content Security Policy (CSP)
- IP Whitelisting

---

## Security Architecture

### Multi-Layer Security Model

1. **Layer 1**: Zero Trust Access (optional - Cloudflare Access with SSO + MFA + Device Check)
2. **Layer 2**: Network (IP Whitelist + GeoIP)
3. **Layer 3**: Edge (Rate Limit + DDoS protection)
4. **Layer 4**: Application (JWT + RBAC)
5. **Layer 5**: API (HMAC + Session validation)
6. **Layer 6**: Data (Row-Level Security)

### Session Management

- **Session Duration**: 30 minutes (vs 24 hours for users)
- **Idle Timeout**: 15 minutes of inactivity
- **Storage**: Memory only (no localStorage/sessionStorage)
- **MFA Re-verification**: Required every 4 hours
- **Concurrent Sessions**: Maximum 2 per admin
- **Session Binding**: IP + User-Agent validation
- **Revocation**: Immediate via Redis

### IP Whitelisting

**Database Schema:**

```sql
CREATE TABLE admin_ip_whitelist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address INET,
    ip_range CIDR,
    description TEXT,
    enabled BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    use_count INTEGER DEFAULT 0,
    CONSTRAINT check_ip_or_range CHECK (
        (ip_address IS NOT NULL AND ip_range IS NULL) OR
        (ip_address IS NULL AND ip_range IS NOT NULL)
    )
);
```

### Enhanced Security Headers

```
Content-Security-Policy: 
    default-src 'none';
    script-src 'self';
    style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
    img-src 'self' data:;
    font-src 'self' https://fonts.gstatic.com;
    connect-src 'self' https://api.functionfly.com;
    frame-ancestors 'none';
    base-uri 'self';
    form-action 'self';
    upgrade-insecure-requests;

X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
```

### Comprehensive Audit Logging

**Enhanced Schema:**

```sql
-- Extend existing audit_events table
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS admin_context JSONB;
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS risk_score INTEGER;
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS session_id TEXT;
ALTER TABLE audit_events ADD COLUMN IF NOT EXISTS device_fingerprint TEXT;

-- Admin session tracking
CREATE TABLE admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) NOT NULL,
    session_token_hash TEXT NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    device_fingerprint TEXT,
    mfa_verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    last_activity_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    revoked_by UUID REFERENCES users(id),
    revoke_reason TEXT
);
```

---

## Migration Strategy

### Timeline Overview

**Phase 1**: Setup Infrastructure (Week 1)  
**Phase 2**: Migrate Core Admin Pages (Week 2-3)  
**Phase 3**: Implement Security Enhancements (Week 3-4)  
**Phase 4**: Testing & Validation (Week 4-5)  
**Phase 5**: Deployment & Monitoring (Week 5-6)

---

### Phase 1: Setup New SPA Infrastructure (Week 1)

**Objective**: Create standalone admin SPA project structure with build tooling.

#### Tasks:

1. **Create Admin SPA Project** (Day 1-2)
   - Initialize new Vite + React + TypeScript project
   - Setup package.json with dependencies
   - Configure TypeScript, ESLint, Prettier
   - Setup Tailwind CSS
   - Configure path aliases

2. **Docker Configuration** (Day 2-3)
   - Create multi-stage Dockerfile
   - Update docker-compose.admin.yml
   - Configure build scripts

3. **Caddy/Nginx Configuration** (Day 3)
   - Configure admin subdomain routing
   - Setup CORS
   - Configure security headers
   - Setup SSL/TLS

4. **Environment Variables Setup** (Day 4)
   - Create `.env.example`
   - Configure Infisical integration
   - Setup separate secrets for admin dashboard

5. **Basic Layout & Routing** (Day 5)
   - Create AdminLayout component
   - Setup React Router configuration
   - Create placeholder pages
   - Implement route guards

**Deliverables:**
- ✅ Admin SPA project initialized
- ✅ Docker configuration complete
- ✅ Caddy/Nginx configured
- ✅ Environment variables documented
- ✅ Basic routing working

---

### Phase 2: Migrate Core Admin Pages (Week 2-3)

**Objective**: Migrate 26+ admin pages from user dashboard to standalone admin SPA.

#### Migration Priority:

**Priority 1 - Critical (Days 1-4):**
1. AdminDashboardPage (overview, stats)
2. AdminTenantsPage + AdminTenantDetailPage
3. AdminUsersPage + AdminUserDetailPage
4. AdminAuditPage

**Priority 2 - Important (Days 5-7):**
5. AdminBillingPage
6. AdminSystemPage
7. AdminBackendsPage
8. AdminProvidersPage

**Priority 3 - Standard (Days 8-10):**
9. AdminFunctionsPage + AdminFunctionDetailPage
10. AdminRegistryPage
11. AdminStateFabricPage
12. AdminFeedbackPage
13. AdminFeaturesPage
14. AdminStatusPage
15. AdminContentPage, AdminContentCalendarPage, AdminNewsletterPage, AdminRedirectsPage

**Priority 4 - Additional (Days 10-12):**
16. AdminTrustDashboardPage
17. AdminExecutionAuditPage
18. AdminFraudDetectionPage
19. AdminEconomicLeaderboardPage

#### Migration Process per Page:

1. Copy page directory from `web/dashboard/src/pages/AdminXxxPage/`
2. Paste to `web/admin-dashboard/src/pages/AdminXxxPage/`
3. Update imports to admin-dashboard paths
4. Update authStore to adminAuthStore
5. Update API client to use admin-specific client with HMAC
6. Add comprehensive audit logging
7. Test functionality and permissions

**Deliverables:**
- ✅ All 26+ admin pages migrated
- ✅ Shared components created/adapted
- ✅ API integration functional
- ✅ Basic testing completed

---

### Phase 3: Implement Security Enhancements (Week 3-4)

**Objective**: Implement enhanced security features specific to admin dashboard.

#### Tasks:

1. **Enhanced Authentication System** (Days 1-2)
   - Memory-only session storage
   - Automatic session expiry checking
   - Activity monitoring
   - Idle timeout enforcement

2. **IP Whitelisting Frontend** (Day 2)
   - IP check component
   - Access denied page
   - Admin IP management UI

3. **MFA Re-verification Component** (Day 3)
   - 4-hour re-verification requirement
   - MFA prompt UI
   - Verification flow

4. **Session Timeout Warning** (Day 3)
   - 5-minute warning before expiry
   - Countdown timer
   - Session extension option

5. **Backend Security Enhancements** (Days 4-5)
   - IP whitelist repository methods
   - Admin session management
   - Enhanced audit logging middleware

6. **Rate Limiting Enhancement** (Day 6)
   - Admin-specific rate limits (60/min)
   - Violation tracking
   - Automatic cleanup

7. **CSRF Protection** (Day 7)
   - CSRF token generation
   - Request validation
   - Token lifecycle management

**Deliverables:**
- ✅ Enhanced authentication with session monitoring
- ✅ IP whitelisting frontend & backend
- ✅ MFA re-verification implemented
- ✅ Session timeout warnings functional
- ✅ Backend security middleware complete
- ✅ Rate limiting enhanced
- ✅ CSRF protection implemented

---

### Phase 4: Testing & Validation (Week 4-5)

**Objective**: Comprehensive testing of all admin dashboard features and security measures.

#### Testing Strategy:

1. **Unit Testing** (Days 1-2)
   - Auth store tests
   - Security component tests
   - API client tests
   - Target coverage: >80%

2. **Integration Testing** (Days 2-3)
   - Full authentication flow
   - Page navigation
   - API integration
   - Permission checks

3. **E2E Testing with Playwright** (Days 3-4)
   - Login flow
   - Session timeout
   - IP whitelist
   - Rate limiting
   - Critical user journeys

4. **Security Testing** (Days 5-6)
   - CSRF protection
   - HMAC signature requirement
   - IP whitelist enforcement
   - Rate limiting
   - Session timeout
   - XSS/SQL injection attempts

5. **Performance Testing** (Day 7)
   - Page load times (<3s)
   - Large data set rendering (<2s)
   - Memory leak detection
   - API response times

**Deliverables:**
- ✅ Unit tests for all stores and utilities
- ✅ Integration tests for auth flow
- ✅ E2E tests for critical paths
- ✅ Security tests passed
- ✅ Performance benchmarks met
- ✅ Test coverage > 80%

---

### Phase 5: Deployment & Monitoring (Week 5-6)

**Objective**: Deploy admin dashboard to staging and production with comprehensive monitoring.

#### Tasks:

1. **Staging Deployment** (Days 1-3)
   - Deploy to `admin-staging.functionfly.com`
   - Configure staging environment
   - Run smoke tests
   - UAT with stakeholders

2. **CI/CD Pipeline** (Day 2)
   - GitHub Actions workflow
   - Automated testing
   - Docker image building
   - Deployment automation
   - Rollback procedures

3. **Monitoring Setup** (Days 4-5)
   - Sentry error tracking
   - Prometheus metrics
   - Grafana dashboards
   - Custom admin metrics

4. **Alerting Configuration** (Day 5)
   - High failed login rate
   - IP whitelist denials
   - Session anomalies
   - API errors
   - Performance degradation

5. **Production Deployment** (Day 6)
   - Deploy to `admin.functionfly.com`
   - DNS configuration
   - SSL/TLS setup
   - Smoke tests
   - Monitoring verification

6. **Documentation** (Day 7)
   - Admin user guide
   - Deployment documentation
   - Security documentation
   - Troubleshooting guide

**Deliverables:**
- ✅ Staging environment deployed and tested
- ✅ CI/CD pipeline functional
- ✅ Monitoring and alerting configured
- ✅ Production deployment successful
- ✅ Documentation complete
- ✅ Rollback procedures tested

---

## API Changes Required

### New Endpoints

```
POST   /v1/admin/auth/session              # Create admin session
DELETE /v1/admin/auth/session              # Revoke admin session
GET    /v1/admin/security/check-ip         # Check IP whitelist
POST   /v1/admin/security/ip-whitelist     # Add IP to whitelist
GET    /v1/admin/security/ip-whitelist     # List whitelisted IPs
PATCH  /v1/admin/security/ip-whitelist/:id # Update whitelist entry
DELETE /v1/admin/security/ip-whitelist/:id # Remove IP from whitelist
GET    /v1/admin/sessions                  # List admin sessions
DELETE /v1/admin/sessions/:id              # Revoke specific session
```

### Middleware Stack for Admin Routes

```go
adminRoutes.Use(
    corsMiddleware.Handle,              // CORS for admin subdomain
    rateLimitMiddleware.RateLimit,      // 60 req/min
    ipWhitelistMiddleware.RequireWhitelistedIP, // IP check
    authMiddleware.RequireAuth,         // JWT validation
    authMiddleware.RequirePermission,   // RBAC
    mfaMiddleware.RequireMFA,           // MFA check
    csrfMiddleware.RequireCSRF,         // CSRF token
    auditMiddleware.AuditAdminAction,   // Audit logging
)

// For mutating operations, add HMAC
adminRoutes.HandleFunc("/tenants", 
    advancedSecurityMiddleware.RequireHMACSignature(
        adminHandler.HandleCreateTenant
    )
).Methods("POST")
```

---

## Configuration & Infrastructure

### DNS Configuration

```
admin.functionfly.com        A     <server-ip>
admin-staging.functionfly.com A    <staging-ip>
```

### Caddy Configuration

```caddyfile
admin.functionfly.com {
    header {
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
        Permissions-Policy "geolocation=(), microphone=(), camera=()"
        Strict-Transport-Security "max-age=63072000; includeSubDomains; preload"
        Content-Security-Policy "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data:; font-src 'self' https://fonts.gstatic.com; connect-src 'self' https://api.functionfly.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; upgrade-insecure-requests;"
    }

    rate_limit {
        zone admin_zone {
            key {remote_host}
            window 1m
            events 60
        }
    }

    root * /var/www/admin-dashboard
    try_files {path} /index.html
    file_server

    /health {
        respond "OK" 200
    }

    log {
        output file /var/log/caddy/admin-access.log
        format json
    }
}
```

### Environment Variables

```bash
# Admin Dashboard
VITE_API_BASE_URL=https://api.functionfly.com
VITE_ADMIN_API_BASE_URL=https://api.functionfly.com/v1/admin
VITE_ADMIN_SHARED_SECRET=<from-infisical>
VITE_SESSION_TIMEOUT=1800000          # 30 minutes
VITE_IDLE_TIMEOUT=900000              # 15 minutes
VITE_MFA_REVERIFY_INTERVAL=14400000   # 4 hours
VITE_ENABLE_IP_WHITELIST=true
VITE_ENABLE_DEVICE_FINGERPRINT=true
VITE_ENABLE_AUDIT_LOGGING=true
VITE_SENTRY_DSN=<from-infisical>
VITE_SENTRY_ENVIRONMENT=production
```

---

## Rollback Plan

### Scenario: Critical Issue in Production

1. **Immediate Rollback**
   ```bash
   cd /opt/functionfly
   docker compose -f docker-compose.admin.production.yml down
   docker compose -f docker-compose.admin.production.yml up -d --force-recreate
   ```

2. **Revert to Previous Image**
   ```bash
   docker tag ghcr.io/functionfly/admin-dashboard:previous-stable ghcr.io/functionfly/admin-dashboard:latest
   docker compose -f docker-compose.admin.production.yml up -d
   ```

3. **Fallback to User Dashboard**
   - Update DNS to point `admin.functionfly.com` to main dashboard
   - Add routing rule in main dashboard to handle `/admin/*` routes
   - Restore admin pages in user dashboard temporarily

4. **Communication**
   - Notify admin users via Slack/email
   - Post status update
   - Provide timeline for resolution

---

## Success Metrics

### Security Metrics
- Zero unauthorized access attempts successful
- 100% of admin actions audited
- Session timeout enforcement: 100%
- MFA compliance: 100%
- IP whitelist violations: tracked and alerted

### Performance Metrics
- Page load time: <3 seconds
- API response time: <500ms (p95)
- Time to interactive: <2 seconds
- No memory leaks detected

### Reliability Metrics
- Uptime: >99.9%
- Error rate: <0.1%
- Failed deployments: <5%
- Rollback time: <5 minutes

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Data breach via new attack surface | High | Multi-layer security, IP whitelist, comprehensive audit logging |
| Admin lockout due to IP whitelist | Medium | Emergency access procedure, bypass mechanism for super admins |
| Session timeout disrupting admin work | Low | Configurable timeouts, auto-save, session extension |
| Migration bugs breaking admin functionality | Medium | Comprehensive testing, staged rollout, easy rollback |
| Performance degradation | Low | Performance testing, caching, CDN |
| DNS/SSL issues on deployment | Low | Pre-deployment verification, fallback procedures |

---

## Next Steps

1. **Review & Approval** (This week)
   - Review plan with stakeholders
   - Get security team approval
   - Finalize timeline

2. **Resource Allocation** (This week)
   - Assign developers
   - Schedule design reviews
   - Set up project tracking

3. **Start Phase 1** (Next week)
   - Create admin SPA project
   - Setup infrastructure
   - Begin migration planning

4. **Ongoing**
   - Daily standups
   - Weekly progress reviews
   - Risk monitoring

---

## References

- Current Admin Dashboard: `/web/dashboard/src/pages/Admin*`
- Security Documentation: `/SECURITY.md`
- API Routes: `/internal/api/routes.go`
- Authentication: `/internal/auth/`
- Docker Compose: `/docker-compose.yml`

---

**Plan Version:** 1.0  
**Last Updated:** March 5, 2026  
**Status:** Ready for Implementation
