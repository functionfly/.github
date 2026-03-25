# Admin Dashboard Production Security Plan

**Document Version:** 1.1  
**Last Updated:** 2026-03-22  
**Status:** Draft - Awaiting Implementation  
**Owner:** Security Team  

---

## Executive Summary

This document outlines a comprehensive security plan for hardening the FunctionFly Admin Dashboard to production-ready standards. The admin dashboard, located at `web/admin-dashboard/`, is a React 19 + TypeScript SPA with Go backend handlers in `internal/api/handlers/admin/`. Currently in Phase 1, significant security gaps exist that must be addressed before production deployment.

### Critical Security Gaps Summary

| Gap | Priority | Estimated Effort |
|-----|----------|------------------|
| No CSRF Protection | P0 | 3 days |
| Mock Authentication | P0 | 10 days |
| No Backend Rate Limiting | P0 | 3 days |
| IP Allowlist Schema Missing | P1 | 2 days |
| Admin Session Table Missing | P1 | 2 days |
| Device Fingerprinting Not Connected | P1 | 3 days |
| 26+ Admin Pages Pending Migration | P2 | 20 days |

---

## 1. Authentication & Authorization Hardening

### 1.1 Replace Mock Login with Real OAuth2/SAML Integration

**Priority:** P0  
**Estimated Effort:** 10 days  
**Dependencies:** None  

#### Current State

- Login page at `web/admin-dashboard/src/pages/LoginPage.tsx` uses placeholder auth
- Accepts `test@example.com` with any password
- No integration with identity providers

#### Implementation Requirements

**Phase 1: OAuth2 Provider Integration (5 days)**

```
1. Configure OAuth2 providers (Google, GitHub, Microsoft)
   - Register applications in each provider
   - Set redirect URIs to /admin/auth/callback/{provider}
   - Store client IDs/secrets in environment variables

2. Implement OAuth2 callback handlers
   - Create handlers in internal/api/handlers/admin/auth.go
   - Verify state parameter to prevent CSRF
   - Exchange code for access token
   - Fetch user profile from provider

3. Create user provisioning logic
   - Auto-create users on first OAuth login
   - Map OAuth email to admin_roles table
   - Store OAuth provider and provider_user_id
```

**Phase 2: SAML 2.0 Integration (3 days)**

```
1. Configure SAML IdP (Okta, Azure AD, Auth0)
   - Metadata URL: /admin/auth/saml/{idp}/metadata
   - ACS URL: /admin/auth/saml/{idp}/acs
   - Implement SAML SSO flow

2. Add SAML attribute mapping
   - Map IdP groups to admin_roles
   - Support Just-in-Time provisioning
```

**Phase 3: Session Enhancement (2 days)**

```
1. Implement secure session creation
   - Generate cryptographically secure session IDs
   - Store sessions in admin_sessions table (see 3.1)
   - Set HttpOnly, Secure, SameSite=Strict cookies

2. Add multi-factor authentication (MFA)
   - TOTP support (Google Authenticator, Authy)
   - WebAuthn/FIDO2 support for hardware keys
   - Backup codes generation
```

**Reference Files:**

- `web/admin-dashboard/src/pages/LoginPage.tsx`
- `internal/api/handlers/admin/auth.go`
- `internal/auth/session.go`

---

### 1.2 Implement CSRF Token Generation and Validation

**Priority:** P0  
**Estimated Effort:** 3 days  
**Dependencies:** Session table (3.1)  

#### Implementation

**Backend (Go) - `internal/api/middleware/csrf.go`**

```go
// Generate CSRF token on session creation
// Token format: HMAC-SHA256(session_id + nonce, secret)
// Store nonce in Redis with 1-hour TTL

// Middleware for mutating requests (POST, PUT, PATCH, DELETE)
// Validate X-CSRF-Token header against session's token
// Reject if token missing, expired, or invalid
```

**Frontend (React) - `web/admin-dashboard/src/lib/csrf.ts`**

```typescript
// Fetch CSRF token on app init via GET /api/admin/csrf
// Store in memory (not localStorage to prevent XSS theft)
// Include in all mutating requests via X-CSRF-Token header
// Refresh token on 401 response
```

**API Endpoints:**

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/admin/csrf` | GET | Get CSRF token for current session |
| `/v1/admin/auth/login` | POST | Login (exempt, uses separate flow) |

**Reference Files:**

- `web/admin-dashboard/src/api/adminClient.ts`
- `internal/api/routes.go`
- `internal/api/middleware/auth.go`

---

### 1.3 Add Rate Limiting to Admin Login Endpoints

**Priority:** P0  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Implementation

**Rate Limit Rules:**

| Endpoint | Limit | Window | Action on Exceed |
|----------|-------|--------|------------------|
| `/v1/admin/auth/login` | 5 attempts | 15 minutes | Block IP for 1 hour |
| `/v1/admin/auth/mfa` | 3 attempts | 15 minutes | Block IP for 4 hours |
| `/v1/admin/auth/password-reset` | 3 attempts | 15 minutes | Block IP for 1 hour |

**Implementation Location:** `internal/api/middleware/ratelimit/admin.go`

```go
// Use sliding window algorithm with Redis
// Key format: ratelimit:admin:login:{ip_address}
// Include X-Forwarded-For handling for proxied requests
// Return 429 Too Many Requests with Retry-After header
```

**Reference Files:**

- `internal/api/middleware/ratelimit/ratelimit.go`
- `internal/api/routes.go`

---

### 1.4 Implement Secure Session Storage with Server-Side Validation

**Priority:** P0  
**Estimated Effort:** 3 days  
**Dependencies:** Admin sessions table (3.1)  

#### Implementation

**Session Storage Schema (see 3.1 for full schema)**

**Validation Flow:**

```
1. JWT contains: session_id, user_id, roles, exp
2. On each request:
   a. Extract session_id from JWT
   b. Query admin_sessions table for session_id
   c. Verify: device_fingerprint, IP address, user_agent
   d. Check: not revoked, not expired, idle_timeout not exceeded
   e. Update last_activity timestamp
3. On validation failure: revoke session, return 401
```

**Memory-Only Storage Issues to Fix:**

- Current: JWT decoded client-side, no server validation
- Problem: Sessions cannot be revoked, device fingerprint not verified
- Solution: Implement above validation flow

**Reference Files:**

- `internal/auth/jwt.go`
- `internal/auth/session.go`
- `web/admin-dashboard/src/stores/authStore.ts`

---

### 1.5 Add Device Fingerprint Verification

**Priority:** P1  
**Estimated Effort:** 3 days  
**Dependencies:** Admin sessions table (3.1)  

#### Implementation

**Frontend Fingerprint Collection - `web/admin-dashboard/src/lib/deviceFingerprint.ts`**

```typescript
// Collect fingerprint using:
interface DeviceFingerprint {
  screenResolution: string;      // e.g., "1920x1080"
  timezone: string;              // e.g., "America/Chicago"
  language: string;             // e.g., "en-US"
  platform: string;             // e.g., "Win32"
  canvasFingerprint: string;    // Hash of canvas rendering
  webglRenderer: string;        // GPU renderer string
  installedFonts: string[];    // Hash of common fonts
  audioFingerprint: number;     // AudioContext hash
}

// Send fingerprint with login request
// Store hash in admin_sessions.device_fingerprint
```

**Backend Verification - `internal/auth/device.go`**

```go
// On login: Store device fingerprint hash
// On subsequent requests: Compare fingerprint
// Flag session if fingerprint mismatch (allow with warning)
// Block if significantly different device (new device notification)
```

**Reference Files:**

- `web/admin-dashboard/src/pages/LoginPage.tsx`
- `internal/api/handlers/admin/auth.go`

---

## 2. API Security

### 2.1 Add Rate Limiting Middleware for All `/v1/admin/*` Routes

**Priority:** P0  
**Estimated Effort:** 1 day  
**Dependencies:** None  

#### Implementation

**Rate Limit Configuration:**

| Route Pattern | Limit | Window | Scope |
|---------------|-------|--------|-------|
| `/v1/admin/auth/*` | 20 req | 1 minute | Per IP |
| `/v1/admin/users/*` | 100 req | 1 minute | Per user |
| `/v1/admin/roles/*` | 50 req | 1 minute | Per user |
| `/v1/admin/audit/*` | 30 req | 1 minute | Per user |
| `/v1/admin/*` (other) | 200 req | 1 minute | Per user |

**Implementation:** Extend `internal/api/middleware/ratelimit/ratelimit.go`

```go
// Add admin-specific rate limiter
// Use token bucket algorithm for smoother limiting
// Implement burst allowance
// Add per-route configuration
```

---

### 2.2 Implement CSRF Token Validation for Mutating Requests

**Priority:** P0  
**Estimated Effort:** 2 days  
**Dependencies:** CSRF token generation (1.2)  

#### Implementation

**Protected HTTP Methods:** POST, PUT, PATCH, DELETE

**Validation Rules:**

```
1. All mutating requests must include X-CSRF-Token header
2. Token must match server-side token for session
3. Token must not be expired (1 hour TTL)
4. Double-submit cookie pattern as fallback for API clients
5. Exempt: /v1/admin/auth/* (use separate protection)
```

**Error Response (401 Unauthorized):**

```json
{
  "error": "csrf_token_invalid",
  "message": "CSRF token is missing, expired, or invalid",
  "required_header": "X-CSRF-Token"
}
```

---

### 2.3 Add Request Signing Verification

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Current State

- HMAC-SHA256 request signing exists for mutations
- See `web/admin-dashboard/src/lib/hmac.ts`

#### Implementation

**Enhancement: Add Request Signing to Admin Client**

```typescript
// In web/admin-dashboard/src/api/adminClient.ts
interface SignedRequest {
  method: string;
  path: string;
  timestamp: number;
  bodyHash: string;      // SHA256 of request body
  signature: string;     // HMAC-SHA256 of above
}

// Validation in Go handler
// Reject if timestamp > 5 minutes old (replay protection)
// Reject if signature invalid
```

**Reference Files:**

- `web/admin-dashboard/src/api/adminClient.ts`
- `web/admin-dashboard/src/lib/hmac.ts`

---

### 2.4 Add Audit Logging for All Admin Operations

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** Audit schema  

#### Implementation

**Audit Log Schema:**

```sql
CREATE TABLE admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES admin_users(id),
    session_id UUID REFERENCES admin_sessions(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(100),
    ip_address INET NOT NULL,
    user_agent TEXT,
    request_body JSONB,
    response_status INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Logged Actions:**

| Category | Actions |
|----------|---------|
| Authentication | login, logout, mfa_verify, password_change, password_reset |
| User Management | user_create, user_update, user_delete, user_role_change |
| Role Management | role_create, role_update, role_delete, permission_change |
| Session Management | session_revoke, session_revoke_all, device_approve |
| Configuration | settings_update, api_key_create, api_key_revoke |

**Reference Files:**

- `internal/api/handlers/admin/audit.go`
- `docs/SECURITY.md`

---

### 2.5 Implement IP Allowlist Enforcement at API Level

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** IP allowlist table (3.2)  

#### Implementation

**Middleware Flow:**

```
1. Extract client IP (X-Forwarded-For, X-Real-IP, RemoteAddr)
2. For super_admin role: Skip allowlist check
3. Query ip_allowlist for user's allowed CIDR blocks
4. Check if client IP matches any allowed CIDR
5. If no match: Return 403 Forbidden
```

**Reference Files:**

- `internal/api/middleware/adminip.go`
- `internal/api/routes.go`

---

## 3. Database Security

### 3.1 Create Admin Sessions Table for Session Tracking

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Migration: `migrations/000002_admin_sessions.sql`

```sql
CREATE TABLE admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    
    -- Session identification
    session_token_hash VARCHAR(64) NOT NULL UNIQUE,
    device_fingerprint_hash VARCHAR(64),
    user_agent TEXT,
    
    -- Timing
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    last_activity_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Security
    ip_address INET NOT NULL,
    ip_initial INET NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    revoked_at TIMESTAMPTZ,
    revocation_reason VARCHAR(100),
    
    -- Metadata
    device_name VARCHAR(100),
    device_trusted BOOLEAN DEFAULT FALSE NOT NULL,
    mfa_verified_at TIMESTAMPTZ,
    fingerprint_mismatch_warnings INTEGER DEFAULT 0
);

-- Indexes
CREATE INDEX idx_admin_sessions_user_id ON admin_sessions(user_id);
CREATE INDEX idx_admin_sessions_token_hash ON admin_sessions(session_token_hash);
CREATE INDEX idx_admin_sessions_expires_at ON admin_sessions(expires_at);
CREATE INDEX idx_admin_sessions_last_activity ON admin_sessions(last_activity_at);

-- Security
ALTER TABLE admin_sessions ENABLE ROW LEVEL SECURITY;

CREATE POLICY admin_sessions_user_policy ON admin_sessions
    FOR ALL
    USING (user_id = current_setting('request.jwt.claim.user_id', true)::uuid)
    WITH CHECK (true);

CREATE POLICY admin_sessions_superadmin_policy ON admin_sessions
    FOR ALL
    TO admin_service
    USING (EXISTS (
        SELECT 1 FROM admin_users au
        JOIN admin_roles ar ON ar.user_id = au.id
        WHERE au.id = admin_sessions.user_id AND ar.role = 'super_admin'
    ));

-- Trigger to clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_admin_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM admin_sessions WHERE expires_at < NOW() OR is_revoked = TRUE;
END;
$$ LANGUAGE plpgsql;

-- Run daily at 3 AM
SELECT cron.schedule('cleanup-expired-admin-sessions', '0 3 * * *', 
    'SELECT cleanup_expired_admin_sessions()');
```

**Reference Files:**

- `migrations/000001_initial_schema.sql` (existing)
- `docs/MIGRATIONS.md`

---

### 3.2 Create IP Allowlist Table with CIDR Support

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Migration: `migrations/000003_ip_allowlist.sql`

```sql
CREATE TABLE ip_allowlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES admin_users(id) ON DELETE CASCADE,
    role VARCHAR(50),  -- NULL for user-specific, e.g., 'super_admin' for role-based
    
    -- CIDR block
    cidr INET NOT NULL,
    description VARCHAR(255),
    
    -- Metadata
    created_by UUID REFERENCES admin_users(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    
    -- Constraints
    CONSTRAINT ip_allowlist_user_or_role CHECK (
        (user_id IS NOT NULL AND role IS NULL) OR
        (user_id IS NULL AND role IS NOT NULL)
    ),
    CONSTRAINT ip_allowlist_cidr_valid CHECK (
        family(cidr) = 4 OR family(cidr) = 6
    )
);

-- Indexes
CREATE INDEX idx_ip_allowlist_user_id ON ip_allowlist(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_ip_allowlist_role ON ip_allowlist(role) WHERE role IS NOT NULL;
CREATE INDEX idx_ip_allowlist_cidr ON ip_allowlist USING gist (cidr inet_ops);

-- RLS
ALTER TABLE ip_allowlist ENABLE ROW LEVEL SECURITY;

CREATE POLICY ip_allowlist_superadmin_policy ON ip_allowlist
    FOR ALL
    TO admin_service
    USING (EXISTS (
        SELECT 1 FROM admin_users au
        JOIN admin_roles ar ON ar.user_id = au.id
        WHERE au.id = ip_allowlist.created_by AND ar.role = 'super_admin'
    ));

-- Helper function to check if IP is allowed
CREATE OR REPLACE FUNCTION is_ip_allowed(
    check_user_id UUID,
    check_role VARCHAR(50),
    check_ip INET
) RETURNS BOOLEAN AS $$
DECLARE
    allowed BOOLEAN;
BEGIN
    -- super_admin bypass for emergency access
    IF check_role = 'super_admin' THEN
        RETURN TRUE;
    END IF;
    
    -- Check user-specific allowlist
    IF EXISTS (
        SELECT 1 FROM ip_allowlist
        WHERE user_id = check_user_id
        AND is_active = TRUE
        AND check_ip << cidr
    ) THEN
        RETURN TRUE;
    END IF;
    
    -- Check role-based allowlist
    IF EXISTS (
        SELECT 1 FROM ip_allowlist
        WHERE role = check_role
        AND user_id IS NULL
        AND is_active = TRUE
        AND check_ip << cidr
    ) THEN
        RETURN TRUE;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

---

### 3.3 Add Row-Level Security for Admin Tables

**Priority:** P2  
**Estimated Effort:** 3 days  
**Dependencies:** Sessions table (3.1)  

#### Implementation

**Tables Requiring RLS:**

| Table | Policy Type | Implementation |
|-------|-------------|----------------|
| `admin_users` | Role-based | super_admin: full access, others: own row only |
| `admin_roles` | Role-based | super_admin: full access, others: own roles only |
| `admin_audit_log` | Role-based | super_admin + support: read all, others: own only |
| `api_keys` | User-based | Users can only see their own keys |

**Reference Files:**

- `migrations/000001_initial_schema.sql`
- `internal/storage/postgres/admin.go`

---

### 3.4 Implement Connection Pooling with TLS

**Priority:** P2  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Implementation

**Connection Pool Configuration:**

```go
// In internal/storage/sql/pool.go
type PoolConfig struct {
    MaxOpenConns:    25
    MaxIdleConns:    5
    ConnMaxLifetime: 5 * time.Minute
    ConnMaxIdleTime: 1 * time.Minute
    
    // TLS configuration
    TLSMode: "require"  // Options: disable, allow, prefer, require, verify-ca, verify-full
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS13,
        CurvePreferences: []tls.CurveID{
            tls.X25519,
            tls.CurveP256,
        },
    }
}
```

**Environment Variables:**

```bash
DB_SSLMODE=require           # or "verify-full" for production
DB_SSLROOTCERT=/path/to/ca   # for verify-full mode
```

**Reference Files:**

- `internal/storage/sql/pool.go`
- `docs/LOCAL_POSTGRES_17.md`

---

## 4. Frontend Security

### 4.1 Implement CSRF Token Handling in adminClient

**Priority:** P0  
**Estimated Effort:** 2 days  
**Dependencies:** CSRF middleware (1.2)  

#### Implementation

**File: `web/admin-dashboard/src/api/adminClient.ts`**

```typescript
// Add CSRF token to all mutating requests
// Get CSRF token from auth store on app initialization
// Include in X-CSRF-Token header for all POST/PUT/PATCH/DELETE

// Refresh token on 401 response
// Handle token expiration gracefully
```

**Reference Files:**

- `web/admin-dashboard/src/api/adminClient.ts`
- `web/admin-dashboard/src/stores/authStore.ts`

---

### 4.2 Add Device Fingerprint Collection and Verification

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** Device fingerprint backend (1.5)  

#### Implementation

**File: `web/admin-dashboard/src/lib/deviceFingerprint.ts`**

```typescript
// Implement fingerprint collection using FingerprintJS or custom
// Collect: screen, timezone, language, canvas, webgl, fonts, audio
// Hash fingerprint and include in login request
// Store trusted fingerprints after MFA verification
```

**Reference Files:**

- `web/admin-dashboard/src/pages/LoginPage.tsx`
- `web/admin-dashboard/src/lib/`

---

### 4.3 Ensure All API Calls Use HMAC Signing

**Priority:** P1  
**Estimated Effort:** 1 day  
**Dependencies:** Request signing verification (2.3)  

#### Current Implementation

HMAC signing exists in `web/admin-dashboard/src/lib/hmac.ts`

#### Enhancement Required

```typescript
// Extend adminClient to automatically sign all requests
// Include timestamp in signed payload (5-minute validity)
// Add retry logic for replay-attack detection
```

**Reference Files:**

- `web/admin-dashboard/src/lib/hmac.ts`
- `web/admin-dashboard/src/api/adminClient.ts`

---

### 4.4 Implement Secure Session Handling with Proper Timeout UI

**Priority:** P2  
**Estimated Effort:** 2 days  
**Dependencies:** Session validation backend (1.4)  

#### Implementation

**Session Timeout Warning UI:**

```typescript
// Show warning modal at 5 minutes before session expiry
// Display countdown timer
// Options: "Stay Logged In" (refresh session), "Logout Now"
// Auto-logout on session expiry with redirect to login
```

**Reference Files:**

- `web/admin-dashboard/src/stores/authStore.ts`
- `web/admin-dashboard/src/components/SessionTimeoutModal.tsx` (new)

---

## 5. Infrastructure Security

### 5.1 Configure WAF Rules for Admin Endpoints

**Priority:** P1  
**Estimated Effort:** 3 days  
**Dependencies:** None  

#### Cloudflare WAF Rules

**Rule Set:**

| Rule Name | Condition | Action |
|-----------|-----------|--------|
| admin-block-known-bad-ips | ip.src in $threat_feed | Block |
| admin-rate-limit | cf.threat_score > 15 | Challenge |
| admin-geo-block | ip.geoip.country in ['XX', 'YY'] | Block |
| admin-sql-injection | cf.threat_score > 30 OR sql injection patterns | Block |
| admin-xss | xss attack patterns | Block |
| admin-cf-bot-management | cf.bot_management.verified_bot != true | JS Challenge |

**Expression:**

```
(http.request.uri.path matches "^/admin" OR 
 http.request.uri.path matches "^/api/admin")
AND cf.threat_score > 15
```

**Reference Files:**

- `deploy/cloudflare/waf-rules.json`
- `docs/CLOUDFLARE.md`

---

### 5.2 Add Bot Detection and Prevention

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Implementation

**Cloudflare Bot Management:**

```yaml
# Enable bot detection for admin endpoints
bot_management:
  enabled: true
  fight_mode: high
  
# Custom rules
- expression: http.request.uri.path contains "/admin"
  action: managed_challenge
  bot_score_threshold: 30
```

**Backend Bot Detection:**

```go
// In internal/api/middleware/bot.go
// Check for:
- Missing User-Agent or known bot UA
- Suspicious request patterns (rapid-fire)
- JavaScript challenge cookies
- Turnstile/reCAPTCHA tokens
```

**Reference Files:**

- `docs/CLOUDFLARE.md`
- `internal/api/middleware/auth.go`

---

### 5.3 Implement DDoS Protection

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** Cloudflare configuration  

#### Implementation

**Rate Limiting (Cloudflare):**

```json
{
  "description": "Admin endpoint rate limit",
  "matchers": [
    {
      "target": "endpoint",
      "operator": "matches",
      "value": "/admin*"
    }
  ],
  "rules": [
    {
      "threshold": 100,
      "period": 60,
      "action": "simulate"
    }
  ]
}
```

**Under-Attack Mode:**

- Enable automatically when origin traffic > 10x normal
- Manual activation procedure documented in runbook

**Reference Files:**

- `docs/CLOUDFLARE.md`
- `docs/runbooks/`

---

### 5.4 Add Network Segmentation (Admin in Isolated VLAN)

**Priority:** P2  
**Estimated Effort:** 5 days  
**Dependencies:** Infrastructure team  

#### Implementation

**Network Architecture:**

```
Internet
    |
Cloudflare (DDOS protection, WAF)
    |
Load Balancer (internal only)
    |
VLAN 30: Admin API (10.0.30.0/24)
    |   - orchestrator-api (admin endpoints)
    |   - Port 8080 only from load balancer
    |
VLAN 20: Database (10.0.20.0/24)
    |   - PostgreSQL (port 5432 from VLAN 30 only)
    |   - Redis (port 6379 from VLAN 30 only)
    |
VLAN 10: Main API (10.0.10.0/24)
    - orchestrator-api (main endpoints)
    - Other services
```

**Security Groups:**

| SG Name | Inbound | Outbound |
|---------|---------|----------|
| admin-api | LB:8080 | DB:5432, Redis:6379, MainAPI |
| postgres-admin | VLAN30:5432 | - |

---

### 5.5 Configure TLS 1.3 Minimum

**Priority:** P2  
**Estimated Effort:** 1 day  
**Dependencies:** Certificate rotation plan  

#### Implementation

**Nginx/Caddy Configuration:**

```nginx
# TLS settings for admin endpoints
ssl_protocols TLSv1.3;
ssl_ciphers 'TLS_AES_256_GCM_SHA384:TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256';
ssl_prefer_server_ciphers off;
ssl_ecdh_curve X25519:P-256;
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:50m;
ssl_session_tickets off;

# HSTS (only after confirming all assets load over HTTPS)
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
```

**Reference Files:**

- `deploy/caddy/Caddyfile`
- `nginx.conf`

---

## 6. Monitoring & Alerting

### 6.1 Add Admin Login Failure Alerting

**Priority:** P1  
**Estimated Effort:** 1 day  
**Dependencies:** None  

#### Alert Rules

**Prometheus/Grafana:**

```yaml
groups:
  - name: admin_auth_alerts
    rules:
      - alert: AdminLoginFailureRate
        expr: |
          rate(auth_login_failures_total[5m]) / 
          rate(auth_login_attempts_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High admin login failure rate"
          
      - alert: AdminLoginBruteForce
        expr: |
          increase(auth_login_failures_total[15m]) > 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Potential brute force attack on admin login"
```

**Notification Channels:**

- Slack: #security-alerts (critical only)
- PagerDuty: All severity (critical)
- Email: Weekly digest (warning)

**Reference Files:**

- `internal/api/handlers/admin/auth.go`
- `deploy/monitoring/prometheus.yml`

---

### 6.2 Add Anomalous Activity Detection

**Priority:** P2  
**Estimated Effort:** 3 days  
**Dependencies:** Audit logging (2.4)  

#### Detection Rules

**Behavioral Analysis:**

```
1. New IP for user (first time from this IP)
2. Multiple sessions for same user
3. Access at unusual hours (outside 9am-6pm user timezone)
4. Rapid page navigation (crawler behavior)
5. Bulk data access patterns
6. Failed authorization attempts
```

**Implementation:**

```go
// In internal/api/middleware/anomaly.go
// Score each request 0-100
// Alert if score > threshold
// Store anomaly events for review
```

---

### 6.3 Add Rate Limit Breach Alerting

**Priority:** P1  
**Estimated Effort:** 1 day  
**Dependencies:** Rate limiting (2.1)  

#### Alert Rules

```yaml
- alert: AdminRateLimitBreached
  expr: |
    increase(ratelimit_admin_blocked_total[5m]) > 5
  for: 1m
  labels:
    severity: warning
  annotations:
    summary: "Admin endpoint rate limit being exceeded"
```

---

### 6.4 Add Session Anomaly Detection

**Priority:** P2  
**Estimated Effort:** 2 days  
**Dependencies:** Session table (3.1)  

#### Detection Rules

```
1. Concurrent sessions exceeding limit (e.g., > 3)
2. Session from blocked country
3. Fingerprint mismatch warnings
4. Session duration anomaly (extremely short or long)
5. Stale session activity (long gap between requests)
```

---

### 6.5 Configure Audit Log Aggregation

**Priority:** P1  
**Estimated Effort:** 2 days  
**Dependencies:** Audit logging (2.4)  

#### Implementation

**Log Shipping:**

```yaml
# Fluentd/Fluent Bit config
<source>
  @type postgres
  host localhost
  database functionfly
  query SELECT * FROM admin_audit_log WHERE created_at > ?
  tag admin.audit
</source>

<match admin.audit>
  @type elasticsearch
  host elasticsearch.internal
  index_name admin-audit-%Y%m%d
</match>
```

**Retention:**

- Hot storage (Elasticsearch): 30 days
- Cold storage (S3): 1 year
- Archival: 7 years (compliance)

**Reference Files:**

- `deploy/monitoring/`
- `docs/MONITORING.md`

---

## 7. Compliance & Documentation

### 7.1 Document Incident Response Procedures

**Priority:** P1  
**Estimated Effort:** 3 days  
**Dependencies:** None  

#### Incident Response Playbook

**Severity Levels:**

| Level | Definition | Response Time |
|-------|------------|---------------|
| P1 - Critical | Active breach, data exfiltration | 15 minutes |
| P2 - High | Unauthorized access, system compromise | 1 hour |
| P3 - Medium | Policy violation, suspicious activity | 4 hours |
| P4 - Low | Minor incident, no immediate threat | 24 hours |

**Playbook Sections:**

1. Detection & Triage
2. Containment
3. Investigation
4. Eradication & Recovery
5. Post-Incident Review

**Reference Files:**

- `docs/runbooks/security-incident-response.md` (new)
- `docs/DISASTER_RECOVERY_RUNBOOK.md`

---

### 7.2 Add Security Runbooks for Common Attacks

**Priority:** P2  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Runbooks Required

| Runbook | Description |
|---------|-------------|
| `brute-force-admin-login.md` | Handling login brute force attempts |
| `session-hijacking.md` | Detecting and responding to session theft |
| `insider-threat.md` | Investigating suspicious admin activity |
| `api-abuse.md` | Rate limiting and API abuse handling |
| `emergency-access.md` | Emergency super_admin access procedures |
| `data-breach.md` | Breach notification and containment |

**Reference Files:**

- `docs/runbooks/security/`
- `docs/runbooks/agent/`

---

### 7.3 Document Admin Access Approval Workflow

**Priority:** P2  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Workflow

```
1. Request Submission
   └─ User requests admin access via support ticket
   └─ Required: Business justification, duration, access level

2. Manager Approval
   └─ Direct manager reviews and approves/denies
   └─ Auto-notify security team for P1+ roles

3. Security Review (for super_admin, billing_admin)
   └─ Security team reviews request
   └─ Background check verification
   └─ 48-hour SLA

4. Access Provisioning
   └─ Create account with expiration date
   └─ Configure IP allowlist
   └─ Enable MFA enforcement
   └─ Notify user with onboarding instructions

5. Access Expiration
   └─ Automated removal on expiration date
   └─ Manager notification for renewal
   └─ Quarterly access review for permanent access
```

**Reference Files:**

- `docs/ADMIN_SETUP_README.md`
- `internal/api/handlers/admin/users.go`

---

### 7.4 Add Data Retention Policies

**Priority:** P3  
**Estimated Effort:** 2 days  
**Dependencies:** None  

#### Retention Schedule

| Data Type | Retention | Storage | Destruction |
|-----------|-----------|---------|-------------|
| Admin audit logs | 7 years | S3 Glacier | Cryptographic erasure |
| Session records | 90 days | PostgreSQL | DELETE |
| API access logs | 1 year | Elasticsearch | DELETE |
| IP allowlist history | 2 years | PostgreSQL | DELETE |
| MFA recovery codes | Until used | PostgreSQL | DELETE |
| Password reset tokens | 1 hour | Redis | Automatic expiry |

---

## 8. Deployment Checklist

### Pre-Deployment Security Verification

#### Environment Variables (Required)

```bash
# Authentication
ADMIN_JWT_SECRET=<64-byte hex, minimum>
ADMIN_SESSION_SECRET=<64-byte hex, minimum>
ADMIN_CSRF_SECRET=<64-byte hex, minimum>
OAUTH_GOOGLE_CLIENT_SECRET=<from secrets manager>
OAUTH_GITHUB_CLIENT_SECRET=<from secrets manager>

# Database
DB_SSLMODE=verify-full
DB_SSLROOTCERT=/path/to/ca-cert.pem
DATABASE_URL=<postgres connection string with sslmode>

# Redis
REDIS_TLS=true
REDIS_PASSWORD=<strong password>

# Security
ADMIN_MAX_SESSIONS_PER_USER=3
ADMIN_SESSION_TIMEOUT_MINUTES=30
ADMIN_IDLE_TIMEOUT_MINUTES=15
ADMIN_MFA_REQUIRED=true
ADMIN_IP_ALLOWLIST_ENABLED=true

# Rate Limiting
ADMIN_RATE_LIMIT_ENABLED=true
ADMIN_RATE_LIMIT_REQUESTS=100
ADMIN_RATE_LIMIT_WINDOW_MS=60000
```

#### Database Migrations

```
[x] Run migrations/000250_admin_sessions.up.sql
[x] Run migrations/000251_ip_allowlist.up.sql
[x] Verify tables created with correct indexes
[ ] Verify RLS policies applied
[ ] Verify cron job for session cleanup created
[ ] Test cleanup_expired_admin_sessions() function
[x] Security migrations applied (000250, 000251, 000252, 000253)
[x] Redis connection verified for rate limiting
[ ] IP allowlist populated with production IPs
[ ] Security alerts configured
[x] Prometheus metrics endpoint accessible
[ ] Audit log retention policy configured
[ ] Admin access policy documented
[ ] Incident response contacts updated
```

#### SSL/TLS Certificate Verification

**Current Configuration Status:**

| Item | Status | Location |
|------|--------|----------|
| HSTS header | ✅ Configured | [`nginx.conf:14`](web/admin-dashboard/nginx.conf:14), [`admin.Caddyfile:15`](deploy/caddy/admin.Caddyfile:15) |
| TLS 1.3 on load balancer | ✅ TLS 1.3 + TLS_AES_256_GCM_SHA384 | Cloudflare (verified) |
| TLS 1.1/1.2 disabled | ⚠️ Cloudflare default | [`docs/CLOUDFLARE.md`](docs/CLOUDFLARE.md) |
| Perfect forward secrecy | ⚠️ Caddy defaults | [`admin.Caddyfile`](deploy/caddy/admin.Caddyfile) |
| DB SSL mode | ✅ Production: `require` | [`.env.production`](.env.production) |

```
[x] Certificate valid and not expiring within 90 days         ✅ Valid until Jun 20, 2026 (90 days)
[x] Certificate covers admin.functionfly.com                  ✅ CN=admin.functionfly.com
[x] Certificate issued by trusted CA                          ✅ Google Trust Services
[x] TLS 1.3 configured on load balancer                       ✅ Protocol: TLSv1.3, Cipher: TLS_AES_256_GCM_SHA384
[ ] TLS 1.1 and 1.2 disabled for admin endpoints              ⚠️ Requires Cloudflare Dashboard verification
[x] HSTS header configured (max-age: 63072000)                 ✅ x-frame-options, x-content-type-options, referrer-policy all present
[ ] Perfect forward secrecy enabled                           ⚠️ Requires SSL Labs verification
```

> **🔔 Certificate Renewal Reminder:** Cert expires June 20, 2026 - set reminder for 30 days before expiry (May 21, 2026).

**Verification Commands:**

```bash
# 1. Certificate validity and expiration
# Replace admin.functionfly.com with your actual admin domain
openssl s_client -connect admin.functionfly.com:443 -servername admin.functionfly.com </dev/null 2>/dev/null | openssl x509 -noout -dates

# 2. Check certificate covers the correct subdomain
openssl s_client -connect admin.functionfly.com:443 -servername admin.functionfly.com </dev/null 2>/dev/null | openssl x509 -noout -subject -ext subjectAltName

# 3. Verify certificate is from trusted CA (Let's Encrypt/Cloudflare)
openssl s_client -connect admin.functionfly.com:443 -servername admin.functionfly.com </dev/null 2>/dev/null | openssl x509 -noout -issuer

# 4. Check TLS version negotiated (should be TLS 1.3)
echo | openssl s_client -connect admin.functionfly.com:443 -tls1_3 2>&1 | grep -i " TLS "
echo | openssl s_client -connect admin.functionfly.com:443 -tls1_2 2>&1 | grep -i "connect\|failed"

# 5. Test TLS 1.1 and 1.2 are disabled (should fail)
echo | openssl s_client -connect admin.functionfly.com:443 -tls1_1 2>&1 | grep -i "alert\|wrong\|no"
echo | openssl s_client -connect admin.functionfly.com:443 -tls1 2>&1 | grep -i "alert\|wrong\|no"

# 6. Verify HSTS header is present and correct
curl -sI https://admin.functionfly.com/ | grep -i "strict-transport-security"
# Expected: Strict-Transport-Security: max-age=63072000; includeSubDomains; preload

# 7. Check for perfect forward secrecy (ECDHE cipher suites)
openssl s_client -connect admin.functionfly.com:443 -servername admin.functionfly.com </dev/null 2>/dev/null | openssl s_client 2>&1 | grep -i "cipher\|ECDHE"

# 8. Full SSL Labs assessment (online - requires public domain)
# Visit: https://www.ssllabs.com/ssltest/analyze.html?d=admin.functionfly.com

# 9. Check DB SSL configuration in production
grep -E "DB_SSL|SSLMODE" .env.production

# 10. Verify Caddy's TLS configuration in admin.Caddyfile
grep -E "tls|protocol|cipher" deploy/caddy/admin.Caddyfile
```

**Infrastructure Verification (Cloudflare Dashboard):**

1. **TLS 1.3 Configuration:**
   - Go to Cloudflare Dashboard → SSL/TLS → Configuration
   - Ensure "TLS 1.3" is enabled
   - Ensure "Minimum TLS Version" is set to "TLS 1.3" for admin endpoints

2. **TLS 1.1/1.2 Disabled:**
   - Cloudflare doesn't allow disabling older TLS versions per-hostname
   - Alternative: Use WAF rule to block connections with TLS < 1.3

3. **Perfect Forward Secrecy:**
   - Cloudflare enables PFS by default with ECDHE cipher suites
   - Verify in SSL Labs report that cipher suites include ECDHE

4. **Certificate:**
   - Cloudflare Dashboard → SSL/TLS → Edge Certificates
   - Check certificate status and expiration
   - Ensure "Always Use HTTPS" is ON

**Reference Configuration Files:**

- [`web/admin-dashboard/nginx.conf`](web/admin-dashboard/nginx.conf:14) - HSTS header (line 14)
- [`deploy/caddy/admin.Caddyfile`](deploy/caddy/admin.Caddyfile:15) - HSTS header (line 15)
- [`.env.production`](.env.production:14) - DB SSL mode
- [`docs/CLOUDFLARE.md`](docs/CLOUDFLARE.md) - Cloudflare SSL/TLS guidance

#### WAF Configuration

> **⚠️ Note:** WAF items (22-31) require Cloudflare Dashboard verification. Basic security headers are present on the application, but full WAF rules must be verified through the Cloudflare Dashboard.

**Legend:**

- 🖥️ = Code/Infrastructure change (file, deployment, or config file)
- 🌐 = Cloudflare Dashboard configuration required
- ✅ = Already implemented in codebase
- ⚠️ = Partial implementation - verification needed

```
[ ] Cloudflare WAF enabled for admin routes          🌐 Dashboard
[ ] Rate limiting rules deployed                      🖥️ ✅ (Caddyfile) + 🌐 Dashboard
[ ] Bot management enabled                            🌐 Dashboard
[ ] IP reputation feed configured                    🌐 Dashboard
[ ] SQL injection rules set to "Block"               🌐 Dashboard + 🖥️ ✅ (middleware)
[ ] XSS rules set to "Block"                          🌐 Dashboard + 🖥️ ✅ (middleware)
[ ] Geo-blocking configured (if applicable)           🖥️ ✅ (middleware) + 🌐 Dashboard
[ ] OWASP ruleset enabled                             🌐 Dashboard
[ ] DDoS protection configured                        🌐 Dashboard
[ ] Admin origin IP not publicly exposed             🖥️ ✅ (VLAN/Network config)
```

**Verification Steps by Item:**

| # | Item | Location | Verification Command/Action |
|---|------|----------|----------------------------|
| 22 | Cloudflare WAF for admin | Cloudflare Dashboard → Security → WAF | Check that admin.* domain has WAF rules applied |
| 23 | Rate limiting rules | [`deploy/caddy/admin.Caddyfile:24`](deploy/caddy/admin.Caddyfile:24) + Cloudflare Dashboard | `grep -n "rate_limit" deploy/caddy/admin.Caddyfile`; Verify Cloudflare Rate Limiting rules |
| 24 | Bot management | Cloudflare Dashboard → Security → Bots | Enable Bot Fight Mode or Super Bot Fight Mode for admin |
| 25 | IP reputation | Cloudflare Dashboard → Security → WAF → IP Access Rules | Configure reputation-based blocking |
| 26 | SQL injection rules | Cloudflare Dashboard → Security → WAF → Managed Rules | Set SQLi rule to "Block"; [`internal/api/middleware/advanced_security/middleware.go`](internal/api/middleware/advanced_security/middleware.go) has [`SQLInjectionFilter`](internal/api/middleware/advanced_security/filters.go) |
| 27 | XSS rules | Cloudflare Dashboard → Security → WAF → Managed Rules | Set XSS rule to "Block"; [`internal/api/middleware/advanced_security/filters.go`](internal/api/middleware/advanced_security/filters.go) has [`XSSFilter`](internal/api/middleware/advanced_security/filters.go) |
| 28 | Geo-blocking | [`internal/api/middleware/advanced_security/middleware.go:140-145`](internal/api/middleware/advanced_security/middleware.go:140) + Cloudflare Dashboard | `grep -n "GeoBlocking\|geo.*block" internal/`; Verify Cloudflare IP Access Rules for geography |
| 29 | OWASP ruleset | Cloudflare Dashboard → Security → WAF → Managed Rules | Enable OWASP ModSecurity Core Rule Set |
| 30 | DDoS protection | Cloudflare Dashboard → Security → DDoS | Configure DDoS alert thresholds; Orange-cloud proxy enabled |
| 31 | Admin origin IP | Network/VLAN config | Verify admin origin has no public IP; uses private VLAN or Cloudflare Tunnel |

**Existing Code References:**

1. **Rate Limiting (Caddy level):**
   - [`deploy/caddy/admin.Caddyfile:23-25`](deploy/caddy/admin.Caddyfile:23) - 60 req/min per IP for `/admin*` routes
   - [`internal/api/middleware/admin_rate_limit.go`](internal/api/middleware/admin_rate_limit.go) - Admin-specific Redis rate limiting
   - [`internal/api/middleware/advanced_security/ratelimit.go`](internal/api/middleware/advanced_security/ratelimit.go) - Sliding window rate limiter

2. **Application-level Security Filters:**
   - [`internal/api/middleware/advanced_security/filters.go`](internal/api/middleware/advanced_security/filters.go) - SQL injection, XSS, path traversal detection
   - [`internal/api/middleware/advanced_security/middleware.go`](internal/api/middleware/advanced_security/middleware.go) - Geo-blocking, IP reputation checks

3. **Cloudflare Configuration (from [`docs/CLOUDFLARE.md`](docs/CLOUDFLARE.md:212-217)):**

   ```markdown
   ## WAF and security
   - **Proxy:** Enable "Proxied" (orange cloud) for public hostnames
   - **SSL/TLS:** Use "Full (strict)"
   - **Security level:** In Security → Settings, set Security Level (e.g. Medium)
   - **Bot Fight Mode / Under Attack:** Enable when under abuse
   ```

**⚠️ Known Gap:** `docker-compose.admin.yml` does not exist in the repository. Admin deployments use [`deploy/caddy/admin.Caddyfile`](deploy/caddy/admin.Caddyfile) directly.

#### Authentication & Authorization

```
[x] OAuth2 providers configured (Google, GitHub, Microsoft)
[ ] SAML IdP configured (if required)
[ ] MFA enforced for all admin users
[ ] Session timeout configured (30 min)
[ ] Idle timeout configured (15 min)
[ ] IP allowlist table populated
[ ] RBAC roles verified (super_admin, support, billing_admin, developer_admin)
```

#### API Security

```
[x] CSRF middleware deployed ✅ (now properly wired)
[x] Rate limiting middleware deployed ✅ (now properly wired)
[x] Request signing verified on all mutations ✅ (already in adminClient)
[x] Audit logging enabled (adminAuditHandler exists in routes.go:454)
[ ] Error responses don't leak sensitive data
[x] Admin endpoints only accessible from designated IPs ✅ (IP allowlist now wired)
```

#### Monitoring & Alerting

```
[x] Prometheus metrics for admin endpoints (internal/api/metrics/admin_security.go)
[ ] Grafana dashboard created
[x] Login failure alerts configured (SecurityAlertMiddleware.checkFailedLoginThreshold)
[x] Rate limit breach alerts configured (SecurityAlertMiddleware.checkRateLimitExceeded)
[x] Session anomaly alerts configured (SecurityAlertMiddleware.checkSessionAnomaly)
[ ] Slack integration for critical alerts
[ ] PagerDuty integration for P1/P2 alerts
[ ] Log aggregation to Elasticsearch
```

#### Backup & Disaster Recovery

```
[ ] Database backup schedule verified (daily + point-in-time)
[ ] Backup retention: 30 days hot, 1 year cold
[ ] Disaster recovery runbook tested
[ ] Backup restoration tested in past 30 days
[ ] Failover procedure documented and tested
```

### Staging Environment Testing

**Before Production Deployment, Complete in Staging:**

1. **Authentication Flow Testing**
   - [ ] OAuth login (all providers)
   - [ ] SAML login (if applicable)
   - [ ] MFA verification
   - [ ] Session timeout behavior
   - [ ] Idle timeout behavior
   - [ ] Password reset flow

2. **Authorization Testing**
   - [ ] Role-based access control for each role
   - [ ] IP allowlist enforcement
   - [ ] Device fingerprint verification
   - [ ] Session revocation

3. **API Security Testing**
   - [ ] CSRF token validation (positive and negative)
   - [ ] Rate limiting (positive and breach)
   - [ ] Request signing validation
   - [ ] SQL injection prevention
   - [ ] XSS prevention
   - [ ] Input validation

4. **Penetration Testing**
   - [ ] Automated scanning (Burp Suite, OWASP ZAP)
   - [ ] Manual penetration test
   - [ ] Credential stuffing protection test
   - [ ] Session hijacking attempt

---

## Implementation Timeline

```
Week 1 (P0 - Critical):
├── Day 1-2: CSRF token implementation (1.2, 2.2)
├── Day 3-4: Rate limiting middleware (1.3, 2.1)
├── Day 5:   Session validation backend (1.4)

Week 2 (P0 - Critical):
├── Day 1-3: OAuth2 integration (1.1 Phase 1)
├── Day 4-5: Device fingerprinting (1.5)

Week 3 (P1 - High):
├── Day 1-2: Database migrations (3.1, 3.2)
├── Day 3-4: IP allowlist enforcement (2.5)
├── Day 5:   Audit logging (2.4)

Week 4 (P1 - High):
├── Day 1-2: WAF configuration (5.1)
├── Day 3-4: Bot detection (5.2)
├── Day 5:   Monitoring alerts (6.1, 6.3)

Week 5 (P2 - Medium):
├── Day 1-2: Network segmentation (5.4)
├── Day 3-4: Anomaly detection (6.2, 6.4)
├── Day 5:   Documentation (7.1)

Week 6 (P2 - Medium):
├── Day 1-2: Runbooks (7.2)
├── Day 3-4: Access workflow (7.3)
├── Day 5:   Staging testing

Week 7 (P3 - Lower):
├── Day 1-2: RLS implementation (3.3)
├── Day 3-4: Connection pooling (3.4)
├── Day 5:   TLS hardening (5.5)

Week 8 (P3 - Lower):
├── Day 1-3: Admin page migration (continuing)
├── Day 4-5: Penetration testing
├── Day 5:   Production deployment checklist review
```

---

## File References

### Existing Files to Modify

| File | Changes Required |
|------|-------------------|
| `web/admin-dashboard/src/pages/LoginPage.tsx` | Replace mock auth with OAuth/MFA |
| `web/admin-dashboard/src/api/adminClient.ts` | Add CSRF, signing, fingerprint |
| `web/admin-dashboard/src/stores/authStore.ts` | Session validation integration |
| `web/admin-dashboard/src/lib/hmac.ts` | Extend for request signing |
| `internal/api/routes.go` | Add rate limiting, CSRF middleware |
| `internal/api/middleware/auth.go` | Session validation, IP checking |
| `internal/api/handlers/admin/auth.go` | OAuth, MFA, device verification |
| `internal/auth/jwt.go` | Server-side session validation |
| `internal/auth/session.go` | Session management |
| `internal/storage/sql/pool.go` | TLS configuration |
| `deploy/cloudflare/waf-rules.json` | Admin endpoint WAF rules |
| `deploy/monitoring/prometheus.yml` | Add admin metrics |

### New Files to Create

| File | Purpose |
|------|---------|
| `migrations/000002_admin_sessions.sql` | Session tracking table |
| `migrations/000003_ip_allowlist.sql` | IP allowlist table |
| `internal/api/middleware/csrf.go` | CSRF validation |
| `internal/api/middleware/ratelimit/admin.go` | Admin rate limiting |
| `internal/api/middleware/adminip.go` | IP allowlist middleware |
| `internal/api/handlers/admin/audit.go` | Audit logging |
| `internal/auth/device.go` | Device fingerprint verification |
| `web/admin-dashboard/src/lib/deviceFingerprint.ts` | Fingerprint collection |
| `web/admin-dashboard/src/lib/csrf.ts` | CSRF token handling |
| `docs/runbooks/security-incident-response.md` | Incident response |
| `docs/runbooks/security/brute-force-admin-login.md` | Brute force runbook |
| `docs/runbooks/security/session-hijacking.md` | Session theft runbook |
| `deploy/monitoring/admin-dashboard.json` | Grafana dashboard |

---

## Success Criteria

The admin dashboard will be considered production-ready when:

1. **Security Testing Passed**
   - [ ] All P0 items implemented and tested
   - [ ] Penetration test completed with no critical findings
   - [ ] CSRF protection verified
   - [ ] Rate limiting verified

2. **Compliance Requirements Met**
   - [ ] Audit logging implemented and operational
   - [ ] Data retention policies configured
   - [ ] Access approval workflow documented

3. **Operational Readiness**
   - [ ] All runbooks created and reviewed
   - [ ] Monitoring and alerting active
   - [ ] Disaster recovery tested

4. **Performance Baseline Established**
   - [ ] Load testing completed
   - [ ] Response time < 200ms p95
   - [ ] No rate limiting false positives under normal load

---

**Document Approval:**

| Role | Name | Date |
|------|------|------|
| Author | Security Team | 2026-03-21 |
| Reviewer | | |
| Approver | | |
