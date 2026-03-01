# Implementation Plan: Tier-Specific Features

## Overview
This plan outlines the implementation steps for adding tier-specific features to the FunctionFly platform before MVP launch.

## Enterprise-Only Features to Implement

### Phase 1: Core Feature Infrastructure
1. **Create `internal/plans/features.go`**
   - Define feature constants (string keys)
   - Create feature categories (Core, Security, Analytics, Support)
   - Define feature availability map per plan

2. **Create `internal/plans/feature_checker.go`**
   - `HasFeature(plan, feature string) bool`
   - `GetFeaturesForPlan(plan string) []string`
   - `IsEnterpriseOnly(feature string) bool`

### Phase 2: Feature Gating Middleware
3. **Create `internal/api/middleware/features.go`**
   - Feature check middleware
   - RequireFeature handler wrapper
   - Feature-gated route registration

### Phase 3: API Endpoints
4. **Create `internal/api/handlers/admin/features.go`**
   - `GET /admin/features` - List all available features
   - `GET /admin/plans/{plan}/features` - Get features for a plan
   - Feature schema endpoint

### Phase 4: Database
5. **Create migration for default features**
   - Add feature definitions to `pricing_tiers` table
   - Seed default features for each tier

## Feature Definitions

### Enterprise-Only Features
```
FEATURE_MICRO_VMS           - Python MicroVM runtime
FEATURE_DEDICATED_POOL     - Unlimited agent concurrency  
FEATURE_CUSTOM_LIMITS       - Configurable request/limits
FEATURE_ADVANCED_SECURITY  - Enhanced security features
FEATURE_SLA                - Service Level Agreement
FEATURE_PRIORITY_SUPPORT   - 24/7 priority support
FEATURE_CUSTOM_DOMAINS     - Branded custom domains
FEATURE_SSO_SAML           - SSO/SAML integration
FEATURE_AUDIT_LOGS         - Extended audit logging
FEATURE_DATA_RESIDENCY     - Region-specific data storage
FEATURE_API_RATE_LIMITS    - Custom rate limiting
FEATURE_WEBHOOK_SIGNING    - Webhook signature verification
FEATURE_ADVANCED_ANALYTICS - Extended analytics
FEATURE_TEAM_RBAC          - Role-based access control
FEATURE_SECRET_ROTATION    - Automatic secret rotation
```

### Pro-Only Features
```
FEATURE_EXTENDED_PROVIDERS   - 3 providers per app (vs 2)
FEATURE_HIGHER_REQUESTS     - 500K requests/month (vs 100K)
FEATURE_AGENT_SCALE_TIER    - Access to agent_scale plan
FEATURE_BASIC_ANALYTICS    - Usage analytics dashboard
FEATURE_WEBHOOK_RETRIES    - Automatic webhook retries
FEATURE_CUSTOM_HEADERS      - Custom HTTP headers
FEATURE_LONG_TIMEOUT        - Extended function timeout
FEATURE_BULK_OPERATIONS    - Bulk function operations
```

### Starter Features (Included)
```
FEATURE_BASIC_PROVIDERS     - 2 providers per app
FEATURE_BASE_REQUESTS       - 100K requests/month
FEATURE_AGENT_STARTER      - Basic agent access
FEATURE_COMMUNITY_SUPPORT   - Community support
FEATURE_BASIC_LOGGING      - 7-day log retention
FEATURE_STANDARD_SLA        - Basic SLA (99.5%)
FEATURE_PUBLISH_FUNCTIONS  - Publish functions to registry
```

## Code Implementation Details

### 1. Feature Constants (features.go)
```go
package plans

// Feature categories
const (
    CategoryCore       = "core"
    CategorySecurity   = "security"
    CategoryAnalytics  = "analytics"
    CategorySupport    = "support"
)

// Enterprise-only features
const (
    FeatureMicroVMs         = "micro_vms"
    FeatureDedicatedPool    = "dedicated_pool"
    FeatureCustomLimits     = "custom_limits"
    FeatureAdvancedSecurity = "advanced_security"
    FeatureSLA             = "sla"
    FeaturePrioritySupport = "priority_support"
    FeatureCustomDomains   = "custom_domains"
    FeatureSSOSAML        = "sso_saml"
    FeatureAuditLogs      = "audit_logs"
    FeatureDataResidency  = "data_residency"
    FeatureAPIRateLimits  = "api_rate_limits"
    FeatureWebhookSigning = "webhook_signing"
    FeatureAdvancedAnalytics = "advanced_analytics"
    FeatureTeamRBAC       = "team_rbac"
    FeatureSecretRotation = "secret_rotation"
)

// Pro features
const (
    FeatureExtendedProviders = "extended_providers"
    FeatureHigherRequests   = "higher_requests"
    FeatureAgentScaleTier   = "agent_scale_tier"
    FeatureBasicAnalytics   = "basic_analytics"
    FeatureWebhookRetries   = "webhook_retries"
    FeatureCustomHeaders    = "custom_headers"
    FeatureLongTimeout      = "long_timeout"
    FeatureBulkOperations   = "bulk_operations"
)

// Starter features
const (
    FeatureBasicProviders    = "basic_providers"
    FeatureBaseRequests      = "base_requests"
    FeatureAgentStarter      = "agent_starter"
    FeatureCommunitySupport  = "community_support"
    FeatureBasicLogging      = "basic_logging"
    FeatureStandardSLA       = "standard_sla"
    FeaturePublishFunctions  = "publish_functions"
)
```

### 2. Feature Checker (feature_checker.go)
```go
// GetFeaturesForPlan returns all features available for a plan
func GetFeaturesForPlan(plan string) []string {
    switch plan {
    case PlanEnterprise:
        return enterpriseFeatures
    case PlanPro:
        return proFeatures
    case PlanStarter:
        return starterFeatures
    case PlanAgentEnterprise:
        return agentEnterpriseFeatures
    case PlanAgentPro:
        return agentProFeatures
    case PlanAgentScale:
        return agentScaleFeatures
    case PlanAgentStarter:
        return agentStarterFeatures
    default:
        return starterFeatures
    }
}

// HasFeature checks if a plan has a specific feature
func HasFeature(plan string, feature string) bool {
    features := GetFeaturesForPlan(plan)
    for _, f := range features {
        if f == feature {
            return true
        }
    }
    return false
}

// IsEnterpriseOnly checks if a feature is only available on enterprise
func IsEnterpriseOnly(feature string) bool {
    for _, f := range enterpriseFeatures {
        if f == feature {
            return true
        }
    }
    return false
}
```

### 3. Middleware (middleware/features.go)
```go
// RequireFeature returns a handler that checks for a feature before proceeding
func RequireFeature(feature string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenant := getTenantFromContext(r)
            if !plans.HasFeature(tenant.Plan, feature) {
                http.Error(w, fmt.Sprintf("Feature '%s' requires a higher tier", feature), http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Migration

### Database Schema
```sql
-- Add features column to pricing_tiers if not already present
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS features JSONB DEFAULT '{}';

-- Default features will be seeded via application code
-- Features are defined in plans/features.go
```

## Testing Plan

1. Unit tests for feature checking functions
2. Integration tests for feature middleware
3. Manual testing of tier-restricted features
4. Verify correct 403 responses for unauthorized access

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `internal/plans/features.go` | Create | Feature constants and types |
| `internal/plans/feature_checker.go` | Create | Feature validation functions |
| `internal/api/middleware/features.go` | Create | Feature checking middleware |
| `internal/api/handlers/admin/features.go` | Create | Feature API endpoints |
| `internal/plans/limits.go` | Modify | Add feature checks to existing limits |
| `plans/FEATURE_SPEC.md` | Create | Feature specification (done) |
| `plans/IMPLEMENTATION_PLAN.md` | Create | This plan (done) |

## Implementation Order

1. Create feature constants in `features.go`
2. Implement feature checking in `feature_checker.go`  
3. Add feature checks to existing endpoints in `limits.go`
4. Create middleware for feature gating
5. Add API endpoints for feature management
6. Update documentation
