# Phase 3: Economic Memory Layer & FlyMindClient Economic Routing

This document describes the implementation of the Economic Memory layer and FlyMindClient economic routing for the FunctionFly AI Service.

## Overview

The Economic Memory layer tracks cost-per-quality metrics for LLM executions, enabling cost-intelligent model selection and routing decisions. This implementation fulfills Phase 3 of the Stateful Cloud Agents (SCA) roadmap.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     API Layer (FastAPI)                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐  │
│  │ /complete   │ │ /route      │ │ /economic-memory/*  │  │
│  │ (tracks)    │ │ (decides)   │ │ (queries/analyze)   │  │
│  └──────┬──────┘ └──────┬──────┘ └──────────┬──────────┘  │
└─────────┼───────────────┼───────────────────┼─────────────┘
          │               │                   │
          ▼               ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer (Python)                     │
│  ┌───────────────────┐    ┌───────────────────────────────┐  │
│  │ Economic Memory   │    │ Economic Routing Service    │  │
│  │ - ExecutionRecord │◄──►│ - Provider scoring          │  │
│  │ - CostQualityScore│    │ - Strategy-based routing    │  │
│  │ - Cost tracking   │    │ - Model recommendations       │  │
│  └─────────┬─────────┘    └───────────────┬───────────────┘  │
│            │                              │                  │
│            │         ┌────────────────────┘                  │
│            │         │                                       │
│            ▼         ▼                                       │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │         Provider Manager (with tracking wrapper)      │  │
│  │  All completions automatically recorded with costs    │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
          │
          │  HTTP / gRPC
          ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Orchestrator                           │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              FlyMindClient (Go)                        │ │
│  │  - Economic routing API calls                         │ │
│  │  - Provider scoring queries                           │ │
│  │  - Cost savings analysis                              │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL                              │
│  - economic_memory_executions (individual records)         │
│  - economic_memory_scores (aggregated metrics)               │
│  - economic_memory_tenant_summary (billing view)             │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. Economic Memory Service (`ai-service/src/services/economic_memory/`)

#### `__init__.py` - Core Memory Module
- **CostQualityScore**: Dataclass tracking cost and quality metrics per provider/model
- **ExecutionRecord**: Records individual LLM executions with cost and quality data
- **EconomicMemory**: In-memory store with PostgreSQL persistence

**Key Features:**
- Tracks: cost per 1K tokens, cost per request, quality scores, latency, success rate
- Calculates: Cost-Quality Index (CQI) = quality / cost
- Provides: Best value provider selection, model switch suggestions, cost breakdowns

#### `repository.py` - Database Persistence
- PostgreSQL integration with asyncpg
- Execution records table with indexes for efficient querying
- Aggregated scores table with automatic updates via trigger
- Migration SQL included

#### `tracking.py` - Provider Tracking Integration
- Wraps providers to automatically record all completions
- Tracks latency, tokens, cost, and quality
- Non-blocking (failures don't affect request completion)
- Cost estimation for providers based on known pricing

### 2. Economic Routing Service (`ai-service/src/services/economic_routing/`)

#### `__init__.py` - Routing Engine
- **EconomicRoutingService**: Makes routing decisions based on cost-quality analysis
- **EconomicRoutingScore**: Combined routing score with economic factors
- **RoutingStrategy**: Available strategies (quality_first, balanced, cost_optimized, cost_first)

**Routing Strategies:**

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `quality_first` | Maximize quality regardless of cost | Critical accuracy requirements |
| `balanced` | Balance cost and quality (default) | General purpose |
| `cost_optimized` | Minimize cost while meeting quality threshold | Budget-conscious |
| `cost_first` | Minimize cost (may reduce quality) | Non-critical workloads |

### 3. API Endpoints (`ai-service/src/api/routes.py`)

New endpoints added:

```
GET  /api/economic-memory/scores           # All cost-quality scores
POST /api/economic-memory/route             # Get routing recommendation
GET  /api/economic-memory/recommendation    # Model recommendation
GET  /api/economic-memory/savings           # Cost savings analysis
GET  /api/economic-memory/executions        # Recent execution records
GET  /api/economic-memory/health            # Health check
```

### 4. FlyMindClient (Go) - `internal/support/flymind_economic.go`

Go client for the Go orchestrator to:
- Query economic scores
- Get routing recommendations
- Analyze cost savings opportunities
- Select best value providers

### 5. Database Migration - `migrations/00160_economic_memory_phase3.sql`

Creates:
- `economic_memory_executions` table
- `economic_memory_scores` table (aggregated)
- `economic_memory_tenant_summary` view (billing/insights)
- Trigger function for automatic score updates

## Key Metrics

### Cost-Quality Index (CQI)

```
CQI = (quality_score × response_time_score × success_rate × 100) / (cost_per_1k_tokens × 1000)
```

**Interpretation:**
- Higher is better
- Measures quality per dollar spent
- Ranges 0-100 (clamped)

### Quality Composite

```
Quality = quality_score × 0.4 + response_time × 0.3 + token_efficiency × 0.2 + success_rate × 0.1
```

## Usage Examples

### 1. Python - Get Economic Routing Recommendation

```python
from services.economic_routing import get_economic_routing_service, RoutingStrategy

router = get_economic_routing_service()
decision = await router.decide_routing(
    request=RoutingDecisionRequest(function_id="my-function"),
    strategy=RoutingStrategy.BALANCED,
    quality_threshold=0.75,
)

print(f"Recommended: {decision.recommended_edge}")
print(f"CQI: {decision.cost_quality_index}")
```

### 2. Go - Get Best Value Provider

```go
client := support.NewFlyMindClient(nil, logger)

// Get best provider
best, err := client.GetBestValueProvider(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Best value: %s/%s (CQI: %.1f)\n",
    best.Provider, best.Model, best.CostQualityIndex)

// Get routing recommendation
routing, err := client.SelectBestProviderWithEconomicRouting(
    ctx,
    "my-function",
    support.Balanced,
)
```

### 3. API - Query Economic Scores

```bash
# Get all scores
curl http://localhost:8081/api/economic-memory/scores

# Get routing recommendation
curl -X POST http://localhost:8081/api/economic-memory/route \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "my-func",
    "strategy": "cost_optimized",
    "quality_threshold": 0.8
  }'

# Get cost savings analysis
curl "http://localhost:8081/api/economic-memory/savings?days=7&tenant_id=tenant-123"
```

## Provider Cost Estimates

Built-in cost estimates (per 1K tokens):

| Provider | Model | Input | Output |
|----------|-------|-------|--------|
| OpenAI | gpt-4o | $0.005 | $0.015 |
| OpenAI | gpt-4o-mini | $0.00015 | $0.0006 |
| Anthropic | claude-3-haiku | $0.00025 | $0.00125 |
| Groq | llama-3.1-8b | $0.0001 | $0.0001 |
| DeepInfra | various | $0.00015 | $0.00015 |
| Ollama | local | $0 | $0 |

## Testing

Run the test suite:

```bash
cd ai-service
python -m pytest tests/test_economic_memory.py -v
```

Integration test included that simulates 100 executions across 3 providers.

## Database Maintenance

### Automatic Score Updates

The database includes a trigger that automatically updates aggregated scores when new execution records are inserted:

```sql
-- Trigger: trg_update_economic_scores
-- Function: update_economic_memory_scores()
```

### Archiving Old Data

Consider partitioning or archiving for high-volume installations:

```sql
-- Optional: Partition by month for >1M records per month
-- See migration file for commented example
```

## Future Enhancements

1. **Tenant-aware routing**: Route based on tenant's historical quality preferences
2. **Adaptive thresholds**: Automatically adjust quality thresholds based on workload type
3. **Predictive cost alerts**: Alert when daily/hourly spend exceeds projections
4. **A/B testing integration**: Use canary system to test provider changes

## References

- Phase 3 Plan: `.kilo/plans/1776486791750-calm-squid.md`
- AI Service Routes: `ai-service/src/api/routes.py`
- Go Client: `internal/support/flymind_economic.go`
