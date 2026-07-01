# ML Intelligence Layer for FunctionFly

**Date:** 2026-06-30  
**Status:** Draft  
**Scope:** Four ML services in FlyMind to replace rule-based heuristics with learned models

---

## Problem

FlyMind (the Python AI service) has zero actual ML models. All "intelligence" is hand-rolled Python math:

- **Anomaly detection**: Z-score with hardcoded global thresholds
- **Prewarming**: Simple moving average + linear regression slope
- **Edge routing**: Weighted scoring (latency 30%, load 30%, availability 40%)
- **Cost anomaly**: `anomaly_detected: false` hardcoded in `usage_forecast_handlers.go:669`
- **Recommendations**: Basic FlyEmbed cosine similarity

These work for bootstrapping but don't learn from data, don't adapt to per-function patterns, and miss anomalies that fall within global thresholds.

## Goal

Replace the four highest-impact rule-based systems with ML models that:

1. Learn from execution data (not static thresholds)
2. Adapt per-function, per-tenant (not global)
3. Improve over time as data accumulates
4. Keep existing API contracts (drop-in replacement)

---

## Architecture

### Service Structure

Each ML service follows the same layout in `ai-service/src/services/`:

```
services/<name>/
├── __init__.py          # Public API exports
├── models.py            # Pydantic request/response models
├── trainer.py           # Model training pipeline
├── predictor.py         # Inference/serving
├── persistence.py       # Model serialization (joblib)
└── feature_store.py     # Feature extraction from Redis/Postgres
```

### Shared Infrastructure

```
services/ml_common/
├── __init__.py
├── features.py          # Common feature extraction utilities
├── persistence.py       # Model versioning, save/load with joblib
├── evaluation.py        # Model evaluation metrics
├── training.py          # Training pipeline abstraction
├── synthetic.py         # Synthetic data generation for bootstrapping
└── scheduler.py         # Retraining cron integration
```

### Data Pipeline

```
Go Backend (execution data)
    → Postgres (cost_allocation_entries, usage_events)
    → Redis (real-time metrics, sorted sets)
    ↓
FlyMind Feature Store (services/ml_common/features.py)
    ↓
ML Models (train on schedule, predict on demand)
    ↓
API Endpoints (same contracts as existing)
```

### Model Lifecycle

1. **Bootstrap**: Train on synthetic data + whatever exists in Postgres/Redis
2. **Serve**: Load model from disk, predict on demand via API
3. **Collect**: Log predictions + actual outcomes to Redis/Postgres
4. **Retrain**: Daily cron job retrains on accumulated data
5. **Evaluate**: Track model drift, alert if accuracy degrades

### Dependencies

```toml
# ai-service/pyproject.toml — add to dependencies
"scikit-learn>=1.5.0",
"numpy>=1.26.0",
"pandas>=2.2.0",
"joblib>=1.4.0",
"scipy>=1.14.0",
"statsmodels>=0.14.0",
```

Note: `prophet` is optional Phase 2 addition. `statsmodels` provides Holt-Winters with lighter footprint. Prophet adds Facebook's decomposable time-series model but has heavier C dependencies.

---

## Service 1: Cost Anomaly Detection

**Location:** `ai-service/src/services/cost_anomaly/`

### Problem

`usage_forecast_handlers.go:669` has `anomaly_detected: false` hardcoded. No cost anomaly detection exists. A function with a memory leak or infinite loop can burn through a user's spend cap before anyone notices.

### Approach

**Phase 1 — Adaptive Z-score per function:**
- Track mean and stddev of `total_cost_cents` per function over 7-day sliding window
- Flag when a single execution cost > mean + 3σ (per-function, not global)
- Track error rate per function — flag when error rate > historical 95th percentile
- Track memory trend — flag when memory_used_mb shows monotonic increase over 10+ executions

**Phase 2 — Isolation Forest for multi-dimensional anomalies:**
- Features: cost, latency, memory, error_rate, region, time_of_day
- Train per-tenant Isolation Forest on 30-day history
- Detect complex anomalies (e.g., normal cost but abnormal latency+memory combination)

### Data Sources

| Source | Location | Access |
|--------|----------|--------|
| Per-execution cost | `cost_allocation_entries` table | Go backend → HTTP to FlyMind |
| Real-time metrics | Redis sorted sets | Direct Redis access |
| Usage events | `usage_events` table | Go backend → HTTP to FlyMind |

### API

```
POST /api/anomalies/cost/check
  Body: { function_id, cost_cents, duration_ms, memory_mb, region, timestamp }
  Response: { is_anomaly, score, type, severity, details }

GET /api/anomalies/cost/{tenant_id}
  Response: { anomalies: [...], summary }
```

### Integration

- Go backend calls `POST /api/anomalies/cost/check` after each cost allocation batch (every 100 executions or 5 minutes)
- Wire into existing `AlertingService` for notifications (Slack, email, in-app)
- Enhance existing `GET /api/anomalies` response with cost anomaly type
- Dashboard: show cost anomalies in the existing anomaly timeline

### Synthetic Data

Generate synthetic cost distributions per function type:
- Normal: Gaussian around function's typical cost
- Anomalous: 5-10x cost spike, gradual drift, error burst
- Use to bootstrap model before real data accumulates

### Metrics

- Precision: % of flagged anomalies that are real
- Recall: % of real anomalies that are caught
- Alert latency: time from anomaly occurrence to alert delivery
- False positive rate: target < 5%

---

## Service 2: Predictive Prewarming

**Location:** `ai-service/src/services/prewarming/` (replace existing)

### Problem

`ForecastingService` uses simple moving average + linear regression. No seasonality awareness. A function that gets traffic every weekday at 9am has the same forecast as one with random traffic.

### Approach

**Phase 1 — Holt-Winters Exponential Smoothing:**
- Triple exponential smoothing (level + trend + seasonality)
- Hourly seasonality (24 periods) with daily cycle
- Weekly seasonality (168 periods) with weekly cycle
- Per-function model (each function gets its own forecast)
- Automatic parameter optimization via grid search on training data

**Phase 2 — Ensemble with Prophet (optional):**
- Add Prophet for functions with complex seasonality (e.g., business-hours patterns)
- Ensemble: average Holt-Winters and Prophet predictions
- Prophet handles holidays and special events automatically

### Data Sources

| Source | Location | Access |
|--------|----------|--------|
| Request rate per function | Redis sorted sets (`prewarm:{function_id}`) | Direct Redis |
| Execution timestamps | `cost_allocation_entries` table | Go backend → HTTP |
| Cold start events | Redis sorted sets | Direct Redis |

### API (same contract as existing)

```
POST /api/prewarm/predict
  Body: { function_id, horizon_hours }
  Response: { predictions: [{ timestamp, predicted_requests, lower_bound, upper_bound, confidence }], method_used }

POST /api/prewarm/warm
  Body: { function_id, instances }
  Response: { warmed, instances }
```

### Integration

- Drop-in replacement for `ForecastingService` — same API contract
- `PrewarmingService` (warmer.py) calls the new forecaster
- Go orchestrator triggers prewarming based on predictions
- Model retraining: daily cron, per-function

### Synthetic Data

Generate synthetic request patterns:
- Constant rate with noise
- Business-hours pattern (peak 9am-5pm, low overnight)
- Bursty pattern (random spikes)
- Growing/declining trend
- Weekend vs weekday differentiation

### Metrics

- MAE (Mean Absolute Error): target < 20% of actual request count
- MAPE (Mean Absolute Percentage Error): target < 30%
- Cold start reduction: target 40-60% fewer cold starts vs current SMA

---

## Service 3: Intelligent Edge Routing

**Location:** `ai-service/src/services/routing/` (replace existing)

### Problem

`RoutingService` uses static weights (latency 30%, load 30%, availability 40%). No learning from execution outcomes. A function that consistently fails on Cloudflare but works on Fly.io will keep getting routed to Cloudflare.

### Approach

**Phase 1 — Thompson Sampling (Multi-Armed Bandit):**
- Arms: edge providers (Cloudflare, Vercel, Fly.io, Deno, FunctionFly-native)
- Reward: `r = 0.4 * (1 - normalize(latency_ms)) + 0.4 * success + 0.2 * (1 - normalize(cost_cents))` where `success` is 1.0 for success, 0.0 for error, and normalization is per-edge min-max over 24h window
- Each arm maintains a Beta(α, β) distribution
- Update: α += reward, β += (1 - reward) after each execution
- Exploration: naturally explores via sampling from distributions
- Per-function models: each function learns its own optimal edge

**Phase 2 — Contextual Bandit (LinUCB):**
- Context features: function language, payload size, user region, time-of-day, function age
- Linear model per arm: reward ≈ θᵀ context
- Upper Confidence Bound for exploration: select arm with highest θᵀ context + α√(contextᵀ A⁻¹ context)
- Shared exploration across similar functions

### Data Sources

| Source | Location | Access |
|--------|----------|--------|
| Per-edge latency | Redis sorted sets (`routing:edge:{name}`) | Direct Redis |
| Execution outcomes | `cost_allocation_entries` table | Go backend → HTTP |
| Adapter health | Health check results in Redis | Direct Redis |

### API (same contract as existing)

```
POST /api/route/decide
  Body: { function_id, user_region, payload_size_bytes, runtime }
  Response: { recommended_edge, confidence, latency_estimate_ms, alternatives: [...] }

GET /api/route/edges
  Response: { edges: [{ name, status, avg_latency_ms, success_rate }] }
```

### Integration

- Drop-in replacement for `RoutingService` — same API contract
- `LatencyCollector` continues collecting latency data (already working)
- Add exploration budget: 10% of requests try non-optimal edges
- Log routing decisions + outcomes for model evaluation
- Model retraining: hourly (Thompson Sampling updates are online, no batch training needed)

### Synthetic Data

- Generate synthetic edge performance profiles per function type
- Simulate edge failures and recovery patterns
- Bootstrap with equal priors (Beta(1, 1) for all arms)

### Metrics

- Average latency improvement vs static routing
- Success rate improvement
- Exploration rate (target: 5-15%)
- Regret: cumulative difference from optimal edge

---

## Service 4: Function Recommendations

**Location:** `ai-service/src/services/recommendations/`

### Problem

Recommendations use only FlyEmbed cosine similarity. No personalization based on user behavior. Every user sees the same recommendations regardless of what they've installed, used, or rated.

### Approach

**Phase 1 — Collaborative Filtering (ALS Matrix Factorization):**
- User-function interaction matrix (implicit feedback: views, installs, executions)
- Alternating Least Squares factorization
- Latent factors: 50-dimensional user and item embeddings
- Handle cold-start: fall back to FlyEmbed similarity for new users/functions

**Phase 2 — Learning-to-Rank (LambdaMART):**
- Training data: user interactions with search results (click, install, skip)
- Features: FlyEmbed similarity (3 vectors), collaborative filtering score, function popularity, recency, category match, author reputation
- LambdaMART optimizes NDCG (ranking quality)
- Re-rank search results from `ResultRanker`

### Data Sources

| Source | Location | Access |
|--------|----------|--------|
| Function embeddings | FlyEmbed triple vectors (Redis) | Direct Redis |
| User interactions | Need new tracking table | Go backend → HTTP |
| Search queries | Need new logging | Go backend → HTTP |
| Function metadata | `registry_functions` table | Go backend → HTTP |

### New Data Collection (required)

Create interaction tracking:

```sql
CREATE TABLE IF NOT EXISTS recommendation_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    function_id UUID NOT NULL,
    interaction_type VARCHAR(32) NOT NULL,  -- 'view', 'install', 'execute', 'rate', 'search_impression', 'search_click'
    context JSONB,                          -- search query, position, etc.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rec_interactions_user ON recommendation_interactions(tenant_id, user_id, created_at);
CREATE INDEX idx_rec_interactions_function ON recommendation_interactions(function_id, created_at);
```

### API

```
GET /api/recommendations/{user_id}
  Query: ?limit=20&exclude_installed=true
  Response: { recommendations: [{ function_id, score, reason }], strategy }

POST /api/recommendations/interactions
  Body: { user_id, function_id, interaction_type, context }
  Response: { recorded: true }

POST /api/search/rerank
  Body: { user_id, query, candidates: [function_id, ...] }
  Response: { ranked: [{ function_id, score }] }
```

### Integration

- New Go handler: `internal/api/handlers/recommendations/`
- Wire interaction tracking into existing registry handlers (view, install, execute endpoints)
- Enhance `ResultRanker` in `services/search/ranker.py` with collaborative filtering scores
- Dashboard: personalized "For You" section in function gallery
- Cold-start: new users get FlyEmbed similarity; new functions get content-based recommendations

### Synthetic Data

- Generate synthetic user-function interaction matrices
- Simulate different user archetypes (backend dev, data scientist, DevOps)
- Bootstrap with FlyEmbed similarity as initial "interactions"

### Metrics

- NDCG@10: ranking quality (target > 0.7)
- Click-through rate on recommendations
- Install rate from recommendations
- Coverage: % of functions recommended to at least one user

---

## Implementation Phases

### Phase 1 — Quick Wins (Week 1-2)

| Task | Effort | Impact |
|------|--------|--------|
| Cost anomaly detection (adaptive Z-score) | 3 days | High — catches cost runaway |
| Usage forecasting with seasonality | 2 days | Medium — better spend predictions |
| Interaction tracking table + Go handlers | 2 days | Foundation for recommendations |
| `services/ml_common/` shared infrastructure | 1 day | Foundation for all ML services |

### Phase 2 — Core ML (Week 3-4)

| Task | Effort | Impact |
|------|--------|--------|
| Holt-Winters prewarming | 3 days | High — reduces cold starts |
| Thompson Sampling routing | 3 days | Medium — adaptive routing |
| Isolation Forest cost anomaly (Phase 2) | 2 days | Higher accuracy anomaly detection |

### Phase 3 — Advanced (Week 5-6)

| Task | Effort | Impact |
|------|--------|--------|
| ALS collaborative filtering | 4 days | Medium — personalized recommendations |
| Contextual bandit routing | 3 days | Higher accuracy routing |
| LambdaMART learning-to-rank | 3 days | Better search results |
| Model retraining cron + monitoring | 2 days | Operational maturity |

---

## File Changes Summary

### New Files (Python — ai-service/)

```
ai-service/src/services/ml_common/__init__.py
ai-service/src/services/ml_common/features.py
ai-service/src/services/ml_common/persistence.py
ai-service/src/services/ml_common/evaluation.py
ai-service/src/services/ml_common/training.py
ai-service/src/services/ml_common/synthetic.py
ai-service/src/services/ml_common/scheduler.py

ai-service/src/services/cost_anomaly/__init__.py
ai-service/src/services/cost_anomaly/models.py
ai-service/src/services/cost_anomaly/trainer.py
ai-service/src/services/cost_anomaly/predictor.py
ai-service/src/services/cost_anomaly/persistence.py
ai-service/src/services/cost_anomaly/feature_store.py

ai-service/src/services/recommendations/__init__.py
ai-service/src/services/recommendations/models.py
ai-service/src/services/recommendations/trainer.py
ai-service/src/services/recommendations/predictor.py
ai-service/src/services/recommendations/persistence.py
ai-service/src/services/recommendations/feature_store.py

ai-service/src/api/routes_ml.py
```

### Modified Files (Python — ai-service/)

```
ai-service/pyproject.toml                          # Add ML dependencies
ai-service/src/api/routes.py                       # Mount ML routes
ai-service/src/services/prewarming/forecaster.py   # Replace with Holt-Winters
ai-service/src/services/routing/selector.py        # Replace with Thompson Sampling
ai-service/src/services/anomaly/detector.py        # Add cost anomaly type
ai-service/src/config.py                           # Add ML config vars
```

### New Files (Go — backend)

```
internal/api/handlers/recommendations/handler.go
internal/storage/recommendations/interactions.go
migrations/YYYYMMDDHHMMSS_add_recommendation_interactions.up.sql
migrations/YYYYMMDDHHMMSS_add_recommendation_interactions.down.sql
```

### Modified Files (Go — backend)

```
internal/api/routes.go                              # Register recommendation routes
internal/api/handlers/billing/usage_forecast_handlers.go  # Wire cost anomaly check
```

---

## Configuration

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `ML_ENABLED` | Master switch for ML services | `true` |
| `ML_RETRAIN_CRON` | Cron expression for model retraining | `0 3 * * *` (3 AM daily) |
| `ML_MODEL_DIR` | Directory for serialized models | `/var/lib/flymind/models` |
| `ML_COST_ANOMALY_THRESHOLD` | Z-score threshold for cost anomalies | `3.0` |
| `ML_COST_ANOMALY_WINDOW` | Sliding window size (hours) | `168` (7 days) |
| `ML_PREWARM_SEASONALITY` | Seasonality period (hours) | `24` |
| `ML_ROUTING_EXPLORATION` | Exploration budget (0.0-1.0) | `0.1` |
| `ML_RECOMMENDATION_LATENT_DIMS` | Latent factor dimensions | `50` |
| `ML_SYNTHETIC_DATA_ENABLED` | Use synthetic data for bootstrapping | `true` |

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Insufficient training data | Synthetic data generation in Phase 1; models gracefully degrade to heuristics |
| Model drift as platform grows | Daily retraining + evaluation metrics + alerting on accuracy degradation |
| Latency impact of ML inference | Thompson Sampling is O(1); Z-score is O(n) on small windows; models loaded in memory |
| Cold-start for new functions/users | Fall back to global models, FlyEmbed similarity, and rule-based heuristics |
| False positive cost anomalies | Start with conservative thresholds (3σ); tune based on user feedback |
| Dependency bloat | scikit-learn + numpy + pandas is ~100MB; acceptable for a dedicated AI service |

---

## Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Cost anomaly detection | Catch 90%+ of cost runaways | Manual review of flagged vs actual anomalies |
| Prewarming accuracy | MAPE < 30% | Compare predicted vs actual request counts |
| Cold start reduction | 40-60% fewer cold starts | Compare cold start rate before/after |
| Routing improvement | 10%+ latency reduction | A/B test Thompson Sampling vs static weights |
| Recommendation relevance | NDCG@10 > 0.7 | Offline evaluation on held-out interactions |
| API latency | < 50ms p95 for ML endpoints | Prometheus metrics |
