# Function Consciousness — Production Plan

> **Feature**: Predictive awareness layer that proactively tells developers what their backend needs before they know.
> **Tagline**: "Your backend tells you what it needs before you know you need it."
> **Plan Integration**: New paid feature tier across Professional, Enterprise, and Agent Enterprise plans.

---

## 1. Executive Summary

Function Consciousness is a **continuous intelligence layer** that analyzes every function, graph, agent, and state across a tenant's account, computes a **System Awareness Score**, and proactively delivers actionable insights through Slack, email, in-app, and webhook channels. It synthesizes existing infrastructure (DNA fitness scoring, anomaly detection, usage forecasting, cost allocation, recommendation engine) into a unified "always-on senior engineer" experience.

### Key Principle: DRY

This feature is built by **composing existing services**, not duplicating them:

| Existing Service | Consciousness Reuse |
|-----------------|-------------------|
| `internal/dna/` fitness scoring, mutation proposals | Function health signals, optimization triggers |
| `ai-service/src/services/anomaly/detector.py` | Traffic anomaly detection (Z-score, 3σ) |
| `internal/services/usage_forecaster.go` | Cost trajectory prediction |
| `internal/storage/billing_repository_cost_allocation.go` | Per-function cost data for redundancy detection |
| `internal/recommendations/` | Marketplace function matching |
| `internal/notification/` | Multi-channel delivery (email, in-app, webhook) |
| `internal/monitoring/alert_engine.go` | Alert rule evaluation patterns |
| `internal/scheduler/` | Cron-based scheduling |
| `internal/plans/` | Feature gating per plan tier |

---

## 2. Plan Integration

### 2.1 Feature Constants

Add to `internal/plans/features.go`:

```go
// Function Consciousness features
const (
    FeatureConsciousnessBasic      = "consciousness_basic"      // Pro+: basic insights, daily digest
    FeatureConsciousnessAdvanced   = "consciousness_advanced"   // Enterprise+: real-time, predictive, auto-fix
    FeatureConsciousnessAutonomous = "consciousness_autonomous" // Agent Enterprise: fully autonomous actions
)
```

### 2.2 Plan Tier Mapping

| Plan | Consciousness Level | Capabilities |
|------|-------------------|-------------|
| **Free** | None | No consciousness features |
| **Starter** | None | No consciousness features |
| **Professional** | Basic | System Awareness Score, daily insight digest (email + in-app), basic cost/health insights, 7-day lookback |
| **Enterprise** | Advanced | All Basic + real-time insights, predictive alerts (Slack + webhook), marketplace recommendations, graph optimization, 30-day lookback, auto-fix proposals |
| **Agent Enterprise** | Autonomous | All Advanced + autonomous fix deployment, unlimited lookback, priority insight queue, dedicated consciousness API |

### 2.3 Files to Modify (Plan Gating)

| File | Change |
|------|--------|
| `internal/plans/features.go` | Add 3 feature constants + definitions + add to plan feature arrays |
| `internal/plans/limits.go` | Add consciousness-specific limits (lookback days, insight frequency, max auto-fixes/day) |
| `web/dashboard/src/lib/plan-utils.ts` | Add `hasConsciousness()`, `hasAdvancedConsciousness()`, `hasAutonomousConsciousness()` helpers |
| `web/dashboard/src/lib/constants.ts` | Add consciousness limits to PLANS object |

### 2.4 Consciousness Limits per Plan

```go
// internal/plans/limits.go
const (
    // Consciousness limits
    StarterConsciousnessLookbackDays    = 0   // Not available
    ProConsciousnessLookbackDays        = 7
    EnterpriseConsciousnessLookbackDays = 30
    AgentEnterpriseConsciousnessLookbackDays = -1 // Unlimited

    ProConsciousnessInsightFrequencyHours        = 24 // Daily digest
    EnterpriseConsciousnessInsightFrequencyHours = 1  // Hourly
    AgentEnterpriseConsciousnessInsightFrequencyHours = 0 // Real-time

    EnterpriseMaxAutoFixesPerDay = 5
    AgentEnterpriseMaxAutoFixesPerDay = -1 // Unlimited
)
```

---

## 3. Architecture

### 3.1 Package Layout

```
internal/consciousness/
├── engine.go              # Main orchestrator — composes all analyzers
├── engine_test.go
├── models.go              # Insight, SystemAwarenessScore, ConsciousnessConfig types
├── repository.go          # PostgreSQL persistence for insights + scores
├── repository_test.go
├── analyzers/
│   ├── analyzer.go        # Analyzer interface (shared contract)
│   ├── traffic.go         # Traffic pattern analysis (delegates to anomaly detector)
│   ├── cost.go            # Cost efficiency analysis (delegates to cost allocation + forecasting)
│   ├── redundancy.go      # Redundancy detection (co-occurrence + code similarity)
│   ├── health.go          # Function health (delegates to DNA fitness scores)
│   ├── marketplace.go     # Marketplace opportunity detection (delegates to recommendations)
│   └── scaling.go         # Scaling trajectory prediction (delegates to usage forecaster)
├── channels/
│   ├── channel.go         # ConsciousnessChannel interface
│   ├── slack.go           # Slack channel (reuses security_alert.go pattern)
│   ├── email.go           # Email channel (delegates to notification service)
│   ├── inapp.go           # In-app channel (delegates to notification service)
│   └── digest.go          # Daily/weekly digest compilation
├── actions/
│   ├── action.go          # Action interface (apply/dismiss)
│   ├── merge_functions.go # Merge redundant functions action
│   ├── scale_config.go    # Scaling config adjustment action
│   └── swap_marketplace.go # Swap to marketplace function action
└── score.go               # System Awareness Score computation
```

### 3.2 Scheduler

```
internal/scheduler/consciousness_scheduler.go  # Periodic analysis runner
```

### 3.3 API Handler

```
internal/api/handlers/consciousness/
├── handler.go             # HTTP handlers
├── registrar.go           # Route registration
└── models.go              # Request/response DTOs
```

### 3.4 Database Migration

```
migrations/YYYYMMDDHHMMSS_function_consciousness.up.sql
migrations/YYYYMMDDHHMMSS_function_consciousness.down.sql
```

### 3.5 Dashboard

```
web/dashboard/src/pages/ConsciousnessPage/
├── index.tsx              # Main consciousness dashboard
├── components/
│   ├── AwarenessScore.tsx # System Awareness Score widget
│   ├── InsightFeed.tsx    # Real-time insight feed
│   ├── InsightCard.tsx    # Individual insight card
│   ├── CostInsights.tsx   # Cost efficiency panel
│   ├── HealthInsights.tsx # Function health panel
│   └── ActionDialog.tsx   # Apply/dismiss action dialog
└── hooks/
    ├── useConsciousness.ts
    └── useInsightFeed.ts
```

---

## 4. Data Model

### 4.1 Core Tables

```sql
-- Consciousness insights (the main table)
CREATE TABLE IF NOT EXISTS consciousness_insights (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Insight classification
    category        VARCHAR(50) NOT NULL,  -- 'traffic', 'cost', 'redundancy', 'health', 'marketplace', 'scaling'
    severity        VARCHAR(20) NOT NULL,  -- 'info', 'warning', 'critical', 'opportunity'
    priority        INT NOT NULL DEFAULT 0, -- Higher = more important
    
    -- Content
    title           VARCHAR(500) NOT NULL,
    message         TEXT NOT NULL,          -- Plain English message (the "senior engineer" voice)
    summary         VARCHAR(200),           -- One-line summary for digest emails
    
    -- Context
    function_id     UUID,                   -- Primary function involved (nullable for account-level insights)
    graph_id        UUID,                   -- Primary graph involved (nullable)
    agent_id        UUID,                   -- Primary agent involved (nullable)
    related_function_ids UUID[] DEFAULT '{}', -- Other functions involved
    
    -- Data payload
    insight_data    JSONB NOT NULL DEFAULT '{}',  -- Structured data (metrics, projections, comparisons)
    action_type     VARCHAR(50),           -- 'merge_functions', 'scale_config', 'swap_marketplace', 'optimize', 'none'
    action_data     JSONB DEFAULT '{}',    -- Action parameters (what to do if user accepts)
    action_preview  JSONB DEFAULT '{}',    -- Preview of what the action would change
    
    -- Trajectory (for predictive insights)
    trajectory      VARCHAR(20),           -- 'improving', 'stable', 'degrading', 'critical'
    projected_days  INT,                   -- Days until issue becomes critical (nullable)
    confidence      NUMERIC(5,4),          -- 0.0 - 1.0 confidence in prediction
    
    -- Lifecycle
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- 'active', 'dismissed', 'applied', 'expired', 'superseded'
    dismissed_at    TIMESTAMPTZ,
    applied_at      TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    superseded_by   UUID REFERENCES consciousness_insights(id),
    
    -- Delivery tracking
    channels_sent   TEXT[] DEFAULT '{}',   -- Which channels this was delivered to
    read_at         TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consciousness_insights_tenant_status ON consciousness_insights(tenant_id, status);
CREATE INDEX idx_consciousness_insights_tenant_category ON consciousness_insights(tenant_id, category);
CREATE INDEX idx_consciousness_insights_created ON consciousness_insights(created_at DESC);
CREATE INDEX idx_consciousness_insights_function ON consciousness_insights(function_id) WHERE function_id IS NOT NULL;
CREATE INDEX idx_consciousness_insights_priority ON consciousness_insights(tenant_id, priority DESC, created_at DESC) WHERE status = 'active';

-- System Awareness Score (one per tenant, updated periodically)
CREATE TABLE IF NOT EXISTS system_awareness_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Overall score (0-100)
    overall_score   NUMERIC(5,2) NOT NULL DEFAULT 0,
    
    -- Component scores (0-100 each)
    health_score    NUMERIC(5,2) NOT NULL DEFAULT 0,   -- Function health (DNA fitness, error rates)
    efficiency_score NUMERIC(5,2) NOT NULL DEFAULT 0,  -- Cost efficiency (spend trends, waste)
    scalability_score NUMERIC(5,2) NOT NULL DEFAULT 0, -- Scaling readiness (trajectory, headroom)
    reliability_score NUMERIC(5,2) NOT NULL DEFAULT 0, -- Reliability (uptime, error patterns)
    optimization_score NUMERIC(5,2) NOT NULL DEFAULT 0, -- Optimization (unused features, marketplace gaps)
    
    -- Metadata
    functions_analyzed  INT NOT NULL DEFAULT 0,
    graphs_analyzed     INT NOT NULL DEFAULT 0,
    agents_analyzed     INT NOT NULL DEFAULT 0,
    active_insights     INT NOT NULL DEFAULT 0,
    critical_insights   INT NOT NULL DEFAULT 0,
    
    -- Trend
    previous_score  NUMERIC(5,2),          -- Score from previous period
    trend           VARCHAR(20),           -- 'improving', 'stable', 'declining'
    
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_system_awareness_scores_tenant ON system_awareness_scores(tenant_id);

-- Consciousness notification preferences (one per tenant)
CREATE TABLE IF NOT EXISTS consciousness_preferences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Channel preferences
    email_enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    slack_enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    slack_webhook_url   VARCHAR(1000),
    inapp_enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    webhook_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    webhook_url         VARCHAR(1000),
    webhook_secret      VARCHAR(256),
    
    -- Frequency preferences
    digest_frequency    VARCHAR(20) NOT NULL DEFAULT 'daily',  -- 'realtime', 'hourly', 'daily', 'weekly'
    quiet_hours_start   TIME,                                  -- No notifications during quiet hours
    quiet_hours_end     TIME,
    timezone            VARCHAR(50) DEFAULT 'UTC',
    
    -- Category preferences (which types of insights to receive)
    enabled_categories  TEXT[] DEFAULT ARRAY['traffic','cost','redundancy','health','marketplace','scaling'],
    
    -- Severity filter (minimum severity to notify)
    min_notify_severity VARCHAR(20) NOT NULL DEFAULT 'warning',  -- 'info', 'warning', 'critical'
    
    -- Autonomous actions (Agent Enterprise only)
    auto_apply_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    auto_apply_categories TEXT[] DEFAULT '{}',
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_consciousness_preferences_tenant ON consciousness_preferences(tenant_id);

-- Insight delivery log (for analytics and dedup)
CREATE TABLE IF NOT EXISTS consciousness_delivery_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id      UUID NOT NULL REFERENCES consciousness_insights(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel         VARCHAR(20) NOT NULL,   -- 'email', 'slack', 'in_app', 'webhook'
    status          VARCHAR(20) NOT NULL,   -- 'sent', 'failed', 'skipped'
    error_message   TEXT,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_consciousness_delivery_log_tenant ON consciousness_delivery_log(tenant_id, sent_at DESC);
```

---

## 5. Analyzer Interface & Implementations

### 5.1 Analyzer Interface

```go
// internal/consciousness/analyzers/analyzer.go
package analyzers

import (
    "context"
    "github.com/functionfly/functionfly/internal/consciousness"
)

// Analyzer is the interface that all consciousness analyzers implement.
// Each analyzer examines a specific aspect of a tenant's backend and
// produces Insights (or nil if nothing noteworthy).
type Analyzer interface {
    // Name returns the analyzer identifier (e.g., "traffic", "cost").
    Name() string

    // Category returns the insight category this analyzer produces.
    Category() string

    // Analyze examines the tenant's backend and returns insights.
    // Returns nil slice if nothing noteworthy is found.
    Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*consciousness.Insight, error)
}

// AnalysisParams provides context for the analysis run.
type AnalysisParams struct {
    LookbackDays   int       // How far back to analyze
    FunctionIDs    []uuid.UUID // Specific functions to analyze (empty = all)
    IncludeGraphs  bool
    IncludeAgents  bool
    PlanTier       string    // For plan-aware analysis depth
}
```

### 5.2 Traffic Analyzer

```go
// internal/consciousness/analyzers/traffic.go
// DELEGATES TO: ai-service/src/services/anomaly/detector.py + internal/services/usage_forecaster.go
package analyzers

// TrafficAnalyzer detects traffic pattern changes and scaling trajectories.
//
// It reuses:
//   - AI anomaly detector (Z-score, 3σ threshold, 5-min sliding window)
//   - Usage forecaster (predictive analytics for spend + execution volume)
//   - Prometheus metrics (functionfly_function_invocations_total)
//
// Produces insights like:
//   - "Your payment function is handling 3x more traffic than last Tuesday.
//     Based on this trajectory you'll hit scaling issues in roughly 6 days."
//   - "Traffic to your auth function has dropped 80% in the last 48 hours.
//     This could indicate an upstream issue."
type TrafficAnalyzer struct {
    anomalyURL    string                           // AI service anomaly endpoint
    forecaster    *usage_forecaster.Service         // Existing usage forecaster
    metricsRepo   storage.MetricsRepository         // Prometheus/DB metrics
}
```

### 5.3 Cost Analyzer

```go
// internal/consciousness/analyzers/cost.go
// DELEGATES TO: internal/storage/billing_repository_cost_allocation.go + billing_repository_forecast.go
package analyzers

// CostAnalyzer detects cost inefficiencies and optimization opportunities.
//
// It reuses:
//   - Cost allocation entries (per-function cost breakdown)
//   - Billing forecast (projected monthly cost)
//   - Billing materialized views (mv_function_usage_stats)
//
// Produces insights like:
//   - "Your image-resize function costs $47/mo but is only used 12 times/day.
//     Switching to a cached approach would save ~$35/mo."
//   - "Your overall spend is up 34% vs last month, driven primarily by
//     the data-pipeline function (now 62% of total cost)."
type CostAnalyzer struct {
    costRepo     storage.CostAllocationRepository
    forecastRepo storage.BillingForecastRepository
}
```

### 5.4 Redundancy Analyzer

```go
// internal/consciousness/analyzers/redundancy.go
// DELEGATES TO: internal/recommendations/ (co-occurrence + similarity) + function_embedding_triples
package analyzers

// RedundancyAnalyzer detects duplicate or overlapping function work.
//
// It reuses:
//   - Function co-occurrence tracking (FunctionCooccurrence)
//   - FlyEmbed triple-vector similarity (contract, semantic, code)
//   - Execution metrics (what each function does, with what inputs/outputs)
//
// Produces insights like:
//   - "Two functions in your graph are doing redundant work.
//     Merging them would save you $140 this month."
//   - "functions validate-email and check-email-format overlap by 87%.
//     Consider consolidating."
type RedundancyAnalyzer struct {
    recoRepo     recommendations.Repository
    functionRepo storage.FunctionRepository
    embedURL     string  // AI service flyembed endpoint
}
```

### 5.5 Health Analyzer

```go
// internal/consciousness/analyzers/health.go
// DELEGATES TO: internal/dna/service.go (fitness scores) + internal/storage/dna/repository.go
package analyzers

// HealthAnalyzer assesses function health via DNA fitness scores and error patterns.
//
// It reuses:
//   - DNA fitness scores (0-100, success rate, P99 latency, cold start rate)
//   - DNA mutation proposals (optimization suggestions)
//   - Error distribution maps
//
// Produces insights like:
//   - "3 of your functions have fitness scores below 60. The worst is
//     order-processor at 42 — it has a 23% timeout rate."
//   - "A new marketplace function was published today that replaces
//     something you built manually. It's faster and cheaper."
type HealthAnalyzer struct {
    dnaRepo   *dna.Repository
    dnaSvc    *dna.Service
}
```

### 5.6 Marketplace Analyzer

```go
// internal/consciousness/analyzers/marketplace.go
// DELEGATES TO: internal/recommendations/ (triple-vector search) + registry search
package analyzers

// MarketplaceAnalyzer discovers marketplace functions that could replace
// or improve the tenant's existing functions.
//
// It reuses:
//   - FlyEmbed triple-vector search (contract + semantic + code matching)
//   - Registry trending/popular functions
//   - Trust scores + pricing data
//
// Produces insights like:
//   - "A new marketplace function (author/sendgrid-email) was published
//     this week. It replaces your custom email-sender with 3x better
//     latency and costs $0.001/call vs your current $0.004."
type MarketplaceAnalyzer struct {
    recoRepo     recommendations.Repository
    registryRepo storage.RegistryRepository
}
```

### 5.7 Scaling Analyzer

```go
// internal/consciousness/analyzers/scaling.go
// DELEGATES TO: internal/services/usage_forecaster.go + monitoring metrics
package analyzers

// ScalingAnalyzer predicts scaling bottlenecks before they happen.
//
// It reuses:
//   - Usage forecaster (trend extrapolation)
//   - Plan limits (concurrency, requests, agents)
//   - Prometheus metrics (execution rate, latency percentiles)
//
// Produces insights like:
//   - "At current growth rate, you'll hit your 1M request limit in 12 days.
//     Consider upgrading to Enterprise or optimizing your top 3 functions."
//   - "Your agent concurrency is at 87% of your plan limit during peak hours."
type ScalingAnalyzer struct {
    forecaster  *usage_forecaster.Service
    planLimits  *plans.LimitChecker
    metricsRepo storage.MetricsRepository
}
```

---

## 6. Engine (Orchestrator)

```go
// internal/consciousness/engine.go
package consciousness

// Engine is the main orchestrator that runs all analyzers, computes the
// System Awareness Score, and dispatches insights to notification channels.
//
// It is NOT a new microservice — it's a composable Go package that
// integrates with the existing scheduler infrastructure.
type Engine struct {
    analyzers    []analyzers.Analyzer       // Pluggable analyzer set
    repo         *Repository                // Persistence
    notifier     *NotificationDispatcher    // Multi-channel delivery
    scoreComputer *ScoreComputer            // System Awareness Score
    planChecker   *plans.FeatureChecker     // Plan-aware feature gating
    logger        *logrus.Logger
    
    // Concurrency control
    maxConcurrent int                       // Max parallel tenant analyses
}

// AnalyzeTenant runs all analyzers for a single tenant and produces insights.
func (e *Engine) AnalyzeTenant(ctx context.Context, tenantID uuid.UUID) (*AnalysisResult, error)

// ComputeAwarenessScore computes the System Awareness Score for a tenant.
func (e *Engine) ComputeAwarenessScore(ctx context.Context, tenantID uuid.UUID) (*SystemAwarenessScore, error)

// DispatchInsights sends new insights to configured notification channels.
func (e *Engine) DispatchInsights(ctx context.Context, tenantID uuid.UUID, insights []*Insight) error
```

### 6.1 Analysis Flow

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Consciousness Scheduler (every 30min)             │
│                                                                      │
│  1. Get all tenants with consciousness feature enabled               │
│  2. For each tenant (bounded concurrency):                           │
│     a. Load consciousness_preferences                                │
│     b. Run all analyzers in parallel:                                │
│        ┌─────────┐ ┌────────┐ ┌─────────────┐ ┌────────┐           │
│        │ Traffic │ │  Cost  │ │ Redundancy  │ │ Health │           │
│        │(anomaly)│ │(alloc) │ │(embeddings) │ │ (DNA)  │           │
│        └────┬────┘ └───┬────┘ └──────┬──────┘ └───┬────┘           │
│        ┌────┴────┐ ┌───┴─────────┐   │            │                │
│        │Scaling  │ │ Marketplace │   │            │                │
│        │(forecast)│ │(reco+embed)│   │            │                │
│        └────┬────┘ └───┬─────────┘   │            │                │
│             └──────────┼─────────────┼────────────┘                │
│                        ▼             ▼                              │
│              ┌─────────────────────────────┐                        │
│              │   Deduplicate + Rank +      │                        │
│              │   Filter by Preferences     │                        │
│              └──────────────┬──────────────┘                        │
│                             ▼                                       │
│              ┌─────────────────────────────┐                        │
│              │  Persist New Insights       │                        │
│              │  (consciousness_insights)   │                        │
│              └──────────────┬──────────────┘                        │
│                             ▼                                       │
│              ┌─────────────────────────────┐                        │
│              │  Compute Awareness Score    │                        │
│              │  (system_awareness_scores)  │                        │
│              └──────────────┬──────────────┘                        │
│                             ▼                                       │
│              ┌─────────────────────────────┐                        │
│              │  Dispatch via Channels      │                        │
│              │  (email/slack/in-app/wh)    │                        │
│              └─────────────────────────────┘                        │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 7. System Awareness Score

### 7.1 Computation

```go
// internal/consciousness/score.go

// ScoreComputer computes the System Awareness Score (0-100).
type ScoreComputer struct {
    dnaRepo      *dna.Repository
    costRepo     storage.CostAllocationRepository
    forecaster   *usage_forecaster.Service
    metricsRepo  storage.MetricsRepository
}

// Component weights (configurable per plan)
var defaultWeights = ScoreWeights{
    Health:      0.25,  // DNA fitness scores, error rates
    Efficiency:  0.20,  // Cost per execution, waste ratio
    Scalability: 0.20,  // Headroom before limits, trajectory
    Reliability: 0.20,  // Uptime, P99 latency, cold start rate
    Optimization: 0.15, // Marketplace adoption, unused features
}
```

### 7.2 Score Interpretation

| Score Range | Label | Color | Message |
|------------|-------|-------|---------|
| 90-100 | Excellent | Green | "Your backend is in top shape. No action needed." |
| 70-89 | Good | Blue | "Your backend is healthy with a few optimization opportunities." |
| 50-69 | Needs Attention | Yellow | "Several areas need attention. Review your insights." |
| 30-49 | At Risk | Orange | "Your backend has critical issues that should be addressed soon." |
| 0-29 | Critical | Red | "Immediate action required. Multiple critical issues detected." |

---

## 8. Notification Channels

### 8.1 Channel Interface

```go
// internal/consciousness/channels/channel.go
package channels

type ConsciousnessChannel interface {
    Name() string
    IsAvailable(prefs *consciousness.Preferences) bool
    SendInsight(ctx context.Context, insight *consciousness.Insight, prefs *consciousness.Preferences) error
    SendDigest(ctx context.Context, digest *consciousness.Digest, prefs *consciousness.Preferences) error
}
```

### 8.2 Channel Implementations

| Channel | Reuses | Delivery |
|---------|--------|----------|
| **Email** | `internal/notification/email_channel.go`, `internal/email/email.go` | New consciousness email templates |
| **Slack** | `internal/api/middleware/security_alert.go` (webhook pattern) | Block Kit messages with action buttons |
| **In-App** | `internal/notification/inapp_channel.go` | New consciousness notification types |
| **Webhook** | `internal/notification/webhook_channel.go` | HMAC-SHA256 signed POST |

### 8.3 Message Format (Plain English)

Every insight message is generated in **plain English** by the engine, following a template system:

```go
// internal/consciousness/messages.go

var insightTemplates = map[string]string{
    "traffic_spike":        "Your {{.FunctionName}} function is handling {{.Multiplier}}x more traffic than {{.ComparisonPeriod}}. Based on this trajectory you'll hit scaling issues in roughly {{.ProjectedDays}} days. {{.ActionHint}}",
    "cost_savings":         "Merging {{.FunctionA}} and {{.FunctionB}} would save you ${{.SavingsAmount}} this month. They're doing redundant work with {{.OverlapPercent}}% overlap.",
    "marketplace_better":   "A new marketplace function ({{.MarketplaceAuthor}}/{{.MarketplaceName}}) was published {{.PublishedAgo}}. It replaces your {{.LocalFunction}} with {{.LatencyImprovement}}x better latency at ${{.CostPerCall}}/call.",
    "health_degraded":      "{{.FunctionName}} fitness has dropped to {{.FitnessScore}}/100 (was {{.PreviousScore}}). The main issue: {{.TopIssue}}.",
    "scaling_trajectory":   "At current growth rate, you'll hit your {{.LimitType}} limit in {{.ProjectedDays}} days. {{.Recommendation}}",
    "redundant_work":       "{{.Count}} functions in your {{.GraphName}} graph are doing overlapping work. Consolidating them would reduce your monthly cost by ${{.SavingsEstimate}}.",
}
```

---

## 9. API Endpoints

### 9.1 Routes

Registered in `internal/api/routes.go`:

```go
// Consciousness endpoints (plan-gated)
consciousnessGroup := v1.Group("/consciousness")
consciousnessGroup.Use(authMiddleware.RequireAuth())
consciousnessGroup.Use(featureMiddleware.RequireFeature(plans.FeatureConsciousnessBasic))

consciousnessGroup.GET("/score", consciousnessHandler.GetAwarenessScore)
consciousnessGroup.GET("/insights", consciousnessHandler.ListInsights)
consciousnessGroup.GET("/insights/:id", consciousnessHandler.GetInsight)
consciousnessGroup.POST("/insights/:id/dismiss", consciousnessHandler.DismissInsight)
consciousnessGroup.POST("/insights/:id/apply", consciousnessHandler.ApplyAction)  // Enterprise+
consciousnessGroup.GET("/feed", consciousnessHandler.StreamInsights)              // SSE for real-time
consciousnessGroup.GET("/digest", consciousnessHandler.GetDigest)                 // Daily digest preview
consciousnessGroup.GET("/preferences", consciousnessHandler.GetPreferences)
consciousnessGroup.PUT("/preferences", consciousnessHandler.UpdatePreferences)
consciousnessGroup.GET("/history", consciousnessHandler.GetScoreHistory)

// Admin endpoints
adminGroup.GET("/consciousness/status", adminHandler.GetConsciousnessStatus)
adminGroup.POST("/consciousness/run/:tenant_id", adminHandler.TriggerAnalysis)
```

### 9.2 API Contract

```typescript
// GET /api/v1/consciousness/score
interface AwarenessScoreResponse {
    score: number;                    // 0-100
    label: string;                    // "excellent" | "good" | "needs_attention" | "at_risk" | "critical"
    components: {
        health: number;
        efficiency: number;
        scalability: number;
        reliability: number;
        optimization: number;
    };
    trend: "improving" | "stable" | "declining";
    previousScore: number | null;
    functionsAnalyzed: number;
    activeInsights: number;
    criticalInsights: number;
    computedAt: string;               // ISO 8601
}

// GET /api/v1/consciousness/insights?category=traffic&severity=warning&status=active&limit=20
interface InsightListResponse {
    insights: Insight[];
    total: number;
    categories: Record<string, number>;  // Count per category
}

interface Insight {
    id: string;
    category: "traffic" | "cost" | "redundancy" | "health" | "marketplace" | "scaling";
    severity: "info" | "warning" | "critical" | "opportunity";
    priority: number;
    title: string;
    message: string;                   // The plain English message
    summary: string;
    functionId: string | null;
    functionName: string | null;
    trajectory: "improving" | "stable" | "degrading" | "critical";
    projectedDays: number | null;
    confidence: number;
    actionType: string | null;
    actionPreview: Record<string, unknown> | null;
    status: "active" | "dismissed" | "applied" | "expired";
    channelsSent: string[];
    createdAt: string;
    updatedAt: string;
}

// POST /api/v1/consciousness/insights/:id/apply
// (Enterprise+ only — applies the suggested action)
interface ApplyActionRequest {
    dryRun?: boolean;                  // Preview the change without applying
    confirmationNote?: string;         // User's note for audit trail
}

interface ApplyActionResponse {
    success: boolean;
    actionType: string;
    changes: Record<string, unknown>;  // What was changed
    dryRun: boolean;
}
```

---

## 10. Scheduler Integration

```go
// internal/scheduler/consciousness_scheduler.go
package scheduler

// ConsciousnessScheduler runs the consciousness analysis engine periodically.
// Follows the exact pattern of DNAInsightsScheduler.
type ConsciousnessScheduler struct {
    cron      *cron.Cron
    engine    *consciousness.Engine
    repo      storage.Repository
    logger    *logrus.Logger
    config    *ConsciousnessSchedulerConfig
    stopOnce  sync.Once
    cancel    context.CancelFunc
}

// Default config: every 30 minutes
// Enterprise+: every 15 minutes (configurable)
// Agent Enterprise: every 5 minutes (near real-time)
```

**Environment Variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `CONSCIOUSNESS_ENABLED` | `true` | Master switch |
| `CONSCIOUSNESS_CRON` | `*/30 * * * *` | Analysis frequency |
| `CONSCIOUSNESS_MAX_CONCURRENT` | `10` | Max parallel tenant analyses |
| `CONSCIOUSNESS_AI_SERVICE_URL` | `http://localhost:8081` | AI service for anomaly/embedding calls |
| `CONSCIOUSNESS_INSIGHT_TTL_HOURS` | `168` | Auto-expire insights after 7 days |

---

## 11. Dashboard UI

### 11.1 New Page Route

Add to `web/dashboard/src/pages/App.tsx` (or router config):

```tsx
<Route path="/consciousness" element={<ConsciousnessPage />} />
```

### 11.2 Navigation

Add "Consciousness" to the sidebar navigation, gated by `hasFeature(plan, 'consciousness_basic')`.

### 11.3 Components

| Component | Description | Reuses |
|-----------|-------------|--------|
| `AwarenessScore` | Circular gauge (0-100) with component breakdown | Existing chart components |
| `InsightFeed` | Scrollable feed of insight cards | Existing notification card pattern |
| `InsightCard` | Individual insight with message, action button, dismiss | Existing card components |
| `CostInsights` | Cost efficiency panel with savings opportunities | Existing cost allocation charts |
| `HealthInsights` | Function health panel with DNA fitness scores | Existing DNA dashboard components |
| `ScalingProjection` | Timeline chart showing projected limit hits | Existing usage forecast charts |
| `ActionDialog` | Confirmation dialog for applying suggested actions | Existing dialog components |
| `PreferencesPanel` | Channel + frequency configuration | Existing settings form patterns |

### 11.4 Widget for Main Dashboard

Add a "Consciousness" widget to the existing draggable widget grid on `/dashboard`:

```tsx
// Shows: Awareness Score, active insight count, top 3 insights
<ConsciousnessWidget score={score} insights={topInsights} />
```

---

## 12. Notification Types

Add to `internal/notification/types.go`:

```go
// Function Consciousness notifications
TypeConsciousnessInsight       = "consciousness.insight"
TypeConsciousnessDigest        = "consciousness.digest"
TypeConsciousnessCritical      = "consciousness.critical"
TypeConsciousnessAutoApplied   = "consciousness.auto_applied"
TypeConsciousnessScoreChanged  = "consciousness.score_changed"
```

Add category:

```go
CategoryConsciousness = "consciousness"
```

---

## 13. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Database migration (4 tables)
- [ ] `internal/consciousness/models.go` — all types
- [ ] `internal/consciousness/repository.go` — CRUD for insights + scores
- [ ] `internal/consciousness/analyzers/analyzer.go` — interface
- [ ] Feature constants in `internal/plans/features.go` + limits
- [ ] Frontend plan-utils updates

### Phase 2: Core Analyzers (Week 3-4)
- [ ] `analyzers/health.go` — DNA fitness delegation
- [ ] `analyzers/cost.go` — cost allocation delegation
- [ ] `analyzers/traffic.go` — anomaly detection delegation
- [ ] `analyzers/scaling.go` — usage forecast delegation
- [ ] `analyzers/redundancy.go` — embedding similarity delegation
- [ ] `analyzers/marketplace.go` — recommendation delegation
- [ ] `score.go` — System Awareness Score computation

### Phase 3: Engine + Scheduler (Week 5)
- [ ] `engine.go` — orchestrator
- [ ] `channels/` — all 4 channel implementations
- [ ] `messages.go` — plain English template system
- [ ] `scheduler/consciousness_scheduler.go`
- [ ] Wire into `cmd/orchestrator-api/main.go`

### Phase 4: API + Actions (Week 6)
- [ ] `handlers/consciousness/` — all HTTP handlers
- [ ] Route registration in `routes.go`
- [ ] `actions/` — merge, scale, swap action implementations
- [ ] Notification type registration

### Phase 5: Dashboard (Week 7-8)
- [ ] `ConsciousnessPage/` — full page with all components
- [ ] Consciousness widget for main dashboard
- [ ] Preferences panel
- [ ] Real-time insight feed (SSE)
- [ ] Action dialogs

### Phase 6: Polish + Testing (Week 9-10)
- [ ] Unit tests for all analyzers
- [ ] Integration tests for engine
- [ ] Load testing (1000+ tenants)
- [ ] Email templates (consciousness digest)
- [ ] Slack Block Kit templates
- [ ] Documentation updates
- [ ] Marketing site pricing page update

---

## 14. Testing Strategy

### 14.1 Unit Tests
- Each analyzer tested independently with mock data
- Score computation tested with known inputs
- Message template rendering tested
- Plan gating tested for all tiers

### 14.2 Integration Tests
- Engine end-to-end: analyzer → insight → score → notification
- Repository CRUD with real Postgres
- Channel delivery (email/slack/in-app/webhook) with mock HTTP

### 14.3 Load Tests
- 1000 tenants × 6 analyzers × 30-min interval
- Insight deduplication under high volume
- Score computation performance (<5s per tenant)

---

## 15. Security Considerations

| Concern | Mitigation |
|---------|-----------|
| Insight data exposure | Insights scoped to tenant_id, RLS enforced |
| Auto-apply actions | Only Agent Enterprise, requires explicit opt-in, audit logged |
| Slack webhook URLs | Encrypted at rest (AES-256-GCM), never returned in API responses |
| Webhook signatures | HMAC-SHA256 with per-tenant secret (reuse webhook_channel.go pattern) |
| Rate limiting | Analysis scheduler uses bounded concurrency, per-tenant cooldown |
| Cost of AI calls | Cached anomaly/embedding results, configurable analysis frequency |

---

## 16. Observability

### Prometheus Metrics

```go
// internal/consciousness/metrics.go
var (
    consciousness_analysis_duration = prometheus.NewHistogramVec(...)
    consciousness_insights_generated = prometheus.NewCounterVec(...)
    consciousness_score_computed = prometheus.NewHistogramVec(...)
    consciousness_notifications_sent = prometheus.NewCounterVec(...)
    consciousness_actions_applied = prometheus.NewCounterVec(...)
    consciousness_analyzer_errors = prometheus.NewCounterVec(...)
)
```

### Logging

All analysis runs logged with:
- Tenant ID, analyzers run, insights generated, score computed, notifications dispatched
- Duration per analyzer, total run duration
- Error details for failed analyzers (non-fatal — other analyzers continue)

---

## 17. Files to Create/Modify Summary

### New Files (23)

| File | Purpose |
|------|---------|
| `internal/consciousness/engine.go` | Main orchestrator |
| `internal/consciousness/models.go` | Types: Insight, Score, Preferences, Digest |
| `internal/consciousness/repository.go` | PostgreSQL persistence |
| `internal/consciousness/score.go` | System Awareness Score computation |
| `internal/consciousness/messages.go` | Plain English message templates |
| `internal/consciousness/analyzers/analyzer.go` | Analyzer interface |
| `internal/consciousness/analyzers/traffic.go` | Traffic pattern analysis |
| `internal/consciousness/analyzers/cost.go` | Cost efficiency analysis |
| `internal/consciousness/analyzers/redundancy.go` | Redundancy detection |
| `internal/consciousness/analyzers/health.go` | Function health analysis |
| `internal/consciousness/analyzers/marketplace.go` | Marketplace opportunities |
| `internal/consciousness/analyzers/scaling.go` | Scaling trajectory prediction |
| `internal/consciousness/channels/channel.go` | Channel interface |
| `internal/consciousness/channels/slack.go` | Slack delivery |
| `internal/consciousness/channels/email.go` | Email delivery |
| `internal/consciousness/channels/inapp.go` | In-app delivery |
| `internal/consciousness/channels/webhook.go` | Webhook delivery |
| `internal/consciousness/channels/digest.go` | Digest compilation |
| `internal/consciousness/actions/action.go` | Action interface |
| `internal/consciousness/actions/merge_functions.go` | Merge redundant functions |
| `internal/consciousness/actions/scale_config.go` | Adjust scaling config |
| `internal/consciousness/actions/swap_marketplace.go` | Swap to marketplace fn |
| `internal/api/handlers/consciousness/handler.go` | HTTP handlers |
| `internal/api/handlers/consciousness/registrar.go` | Route registration |
| `internal/api/handlers/consciousness/models.go` | Request/response DTOs |
| `internal/scheduler/consciousness_scheduler.go` | Cron scheduler |
| `migrations/YYYYMMDDHHMMSS_function_consciousness.up.sql` | Schema |
| `migrations/YYYYMMDDHHMMSS_function_consciousness.down.sql` | Rollback |
| `web/dashboard/src/pages/ConsciousnessPage/index.tsx` | Dashboard page |
| `web/dashboard/src/pages/ConsciousnessPage/components/AwarenessScore.tsx` | Score widget |
| `web/dashboard/src/pages/ConsciousnessPage/components/InsightFeed.tsx` | Insight feed |
| `web/dashboard/src/pages/ConsciousnessPage/components/InsightCard.tsx` | Insight card |
| `web/dashboard/src/pages/ConsciousnessPage/components/ActionDialog.tsx` | Action dialog |
| `web/dashboard/src/pages/ConsciousnessPage/hooks/useConsciousness.ts` | API hooks |
| `web/dashboard/src/pages/ConsciousnessPage/hooks/useInsightFeed.ts` | SSE hook |

### Modified Files (8)

| File | Change |
|------|--------|
| `internal/plans/features.go` | Add 3 consciousness feature constants + plan assignments |
| `internal/plans/limits.go` | Add consciousness limits |
| `internal/notification/types.go` | Add 5 consciousness notification types + category |
| `internal/api/routes.go` | Register consciousness routes |
| `web/dashboard/src/lib/plan-utils.ts` | Add consciousness feature helpers |
| `web/dashboard/src/lib/constants.ts` | Add consciousness limits to PLANS |
| `web/dashboard/src/pages/App.tsx` | Add /consciousness route |
| `cmd/orchestrator-api/main.go` | Wire consciousness scheduler + engine |
