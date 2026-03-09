# AI Function Factory - Implementation Todo List

## Overview

Based on the blueprint analysis and existing codebase review, here's the detailed implementation plan for the AI Function Factory.

### What's Already Built (85% Complete)

| Component | Location | Status |
|-----------|----------|--------|
| Generation Service | `internal/agent/generation/service.go` | ✅ Interface exists, needs LLM implementation |
| Evolution Service | `internal/agent/evolution/service.go` | ✅ Complete |
| Autonomy Service | `internal/agent/autonomy/service.go` | ✅ Complete |
| Publisher | `internal/agent/deployment/publisher.go` | ✅ Complete |
| Learning/Optimizer | `internal/agent/learning/optimizer.go` | ✅ Complete |
| Execution Queue | `internal/queue/execution_queue.go` | ✅ Complete |
| Agent Models | `internal/agent/identity/models.go` | ✅ Complete |
| Discovery Handler | `internal/api/handlers/agent/discovery.go` | ✅ Exists |

---

## Phase 1: Core Pipeline Components (Priority: HIGH)

### 1.1 Discovery Engine Service

- [ ] **Create** `internal/agent/discovery/service.go` - Main discovery orchestration
- [ ] **Create** `internal/agent/discovery/github_scanner.go` - GitHub issues scanner
- [ ] **Create** `internal/agent/discovery/stackoverflow_scanner.go` - StackOverflow scanner
- [ ] **Create** `internal/agent/discovery/reddit_scanner.go` - Reddit dev forums scanner
- [ ] **Create** `internal/agent/discovery/models.go` - Opportunity data models
- [ ] **Create** database migration for `factory_opportunities` table

### 1.2 OpenRouter LLM Integration

- [ ] **Create** `internal/agent/generation/llm.go` - OpenRouter client implementation
- [ ] **Implement** code generation using Mercury 2 model
- [ ] **Add** prompt engineering for consistent function generation
- [ ] **Configure** environment variables for API keys
- [ ] **Plan** RunPod integration for self-hosted inference (Phase 4)

### 1.4 Caching Layer (Priority: HIGH)

- [ ] **Create** template cache service `internal/agent/generation/cache.go`
- [ ] **Implement** semantic caching for similar prompts
- [ ] **Add** cache invalidation strategy
- [ ] **Target:** Reduce LLM calls by 30-50%

### 1.5 Hybrid Model Strategy (Priority: HIGH)

- [ ] **Integrate** Mistral 7B for simple functions
- [ ] **Configure** Mercury 2 for complex functions only
- [ ] **Implement** complexity scoring algorithm
- [ ] **Create** auto-model selector based on function complexity

### 1.6 Quality Gate & Manual Review (Priority: HIGH)

- [ ] **Implement** automatic quality scoring (70% minimum)
- [ ] **Add** manual review queue for high-impact functions
- [ ] **Configure** auto-publish for simple functions
- [ ] **Add** human approval tier for complex functions

### 1.7 Feedback Loop (Priority: MEDIUM)

- [ ] **Implement** function usage tracking
- [ ] **Add** invocation analytics
- [ ] **Create** popularity-based generation priority
- [ ] **Implement** auto-retirement for unused functions
- [ ] **Add** feedback data to discovery scoring

### 1.8 Testing Pipeline Service

- [ ] **Create** `internal/agent/testing/service.go` - Main testing orchestration
- [ ] **Implement** syntax validation (lint, compile)
- [ ] **Implement** security scanner with pattern detection
- [ ] **Implement** sandbox execution testing
- [ ] **Implement** performance benchmarking
- [ ] **Create** database migration for `factory_test_results` table

---

## Phase 2: Pipeline Orchestration (Priority: HIGH)

### 2.1 Factory Orchestrator

- [ ] **Create** `internal/agent/factory/service.go` - Main pipeline orchestrator
- [ ] **Create** `internal/agent/factory/config.go` - Pipeline configuration
- [ ] **Implement** discovery → generation → testing → publish workflow
- [ ] **Add** error handling and retry logic
- [ ] **Implement** quality gate enforcement (70% minimum score)

### 2.2 Database Schema

- [ ] **Create** migration: `factory_opportunities` table
- [ ] **Create** migration: `factory_runs` table (pipeline execution tracking)
- [ ] **Create** migration: `factory_versions` table (version tracking)
- [ ] **Add** GORM models for factory entities

### 2.3 Quality Guardrails

- [ ] **Define** minimum code quality threshold (70%)
- [ ] **Define** test pass rate threshold (90%)
- [ ] **Define** cold start time threshold (<= 2000ms)
- [ ] **Define** execution time threshold (<= 500ms avg)
- [ ] **Implement** security issue rejection (0 critical)
- [ ] **Implement** dependency limit warnings (> 10)

---

## Phase 3: API Handlers (Priority: MEDIUM)

### 3.1 Factory API Endpoints

- [ ] **Create** `internal/api/handlers/factory/handler.go` - Main factory handler
- [ ] **Implement** `GET /api/v1/factory/status` - Factory health and statistics
- [ ] **Implement** `POST /api/v1/factory/pipeline/run` - Manual pipeline trigger
- [ ] **Implement** `GET /api/v1/factory/discover` - Get discovered opportunities
- [ ] **Implement** `POST /api/v1/factory/generate` - Generate specific function
- [ ] **Implement** `GET /api/v1/factory/functions` - List factory-generated functions
- [ ] **Implement** `GET/PUT /api/v1/factory/config` - Factory configuration

### 3.2 Route Registration

- [ ] **Add** factory routes to `internal/api/routes.go`
- [ ] **Add** authentication middleware for factory endpoints
- [ ] **Add** rate limiting for factory operations

---

## Phase 4: Monitoring & Optimization (Priority: MEDIUM)

### 4.1 Performance Monitoring

- [ ] **Integrate** existing optimizer with factory pipeline
- [ ] **Schedule** performance monitoring (every 6 hours)
- [ ] **Implement** latency tracking for factory functions
- [ ] **Implement** error rate monitoring
- [ ] **Implement** automatic optimization triggers

### 4.2 Factory Dashboard Metrics

- [ ] **Track** functions created per day
- [ ] **Track** pipeline success rate
- [ ] **Track** average generation time
- [ ] **Track** quality scores distribution

---

## Phase 5: Admin UI (Priority: LOW)

### 5.1 Factory Dashboard

- [ ] **Create** factory status page in admin dashboard
- [ ] **Display** pipeline statistics
- [ ] **Show** recent function generations
- [ ] **Provide** manual pipeline trigger UI
- [ ] **Show** quality metrics and trends

---

## Phase 6: Scale & Cost Optimization (Priority: LOW)

### 6.1 RunPod Integration

- [ ] **Evaluate** RunPod for self-hosted GPU inference
- [ ] **Design** on-demand GPU provisioning strategy
- [ ] **Implement** GPU instance lifecycle management
- [ ] **Compare** costs: API vs self-hosted at scale

### 6.2 Cost Management

- [ ] **Implement** batch processing for cost savings
- [ ] **Add** cost monitoring and alerts
- [ ] **Create** cost-per-function tracking
- [ ] **Optimize** for target cost efficiency

---

## Implementation Sequence

### Week 1-2: Core Pipeline + Intelligence

```
1. Create Discovery Engine service
2. Implement GitHub scanner
3. Add OpenRouter LLM integration
4. Create Testing Pipeline service
5. Add Caching Layer (reduce costs 30-50%)
6. Implement Hybrid Model Strategy
```

### Week 3-4: Orchestration + Quality

```
1. Build Factory Orchestrator
2. Add database migrations
3. Implement quality guardrails
4. Connect all pipeline stages
5. Add manual review queue
6. Implement feedback loop
```

### Week 5-6: API & Integration

```
1. Create Factory API handlers
2. Add route registration
3. Connect to existing agent services
4. Test end-to-end pipeline
```

### Week 7-8: Polish & Scale

```
1. Performance monitoring integration
2. Admin dashboard UI
3. Production hardening
4. Documentation
```

---

## Key Files to Create

```
internal/agent/
├── discovery/
│   ├── service.go         # Main discovery orchestration
│   ├── github_scanner.go  # GitHub issues scanner
│   ├── stackoverflow_scanner.go
│   ├── reddit_scanner.go
│   └── models.go          # Opportunity models
├── testing/
│   ├── service.go         # Main testing orchestration
│   ├── security_scanner.go
│   ├── sandbox_executor.go
│   └── benchmark_runner.go
├── factory/
│   ├── service.go         # Main orchestrator
│   └── config.go          # Configuration
└── generation/
    └── llm.go             # OpenRouter client

internal/api/handlers/
└── factory/
    └── handler.go         # Factory API handlers

migrations/
└── factory/
    ├── 001_create_opportunities.sql
    ├── 002_create_factory_runs.sql
    └── 003_create_factory_versions.sql
```

---

## Configuration Requirements

### Environment Variables

```bash
# Factory Configuration
FACTORY_ENABLED=true
FACTORY_DAILY_TARGET=20
FACTORY_MAX_CONCURRENT=5
FACTORY_MIN_QUALITY_SCORE=70

# LLM Configuration (OpenRouter)
OPENROUTER_API_KEY=sk-or-...
OPENROUTER_MODEL=deepseek/deepseek-chat
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1

# Discovery Sources
GITHUB_TOKEN=ghp_...
STACKOVERFLOW_KEY=...
REDDIT_CLIENT_ID=...
REDDIT_CLIENT_SECRET=...
```

---

## Cost Estimates

Based on the blueprint, Mercury 2 pricing is:
- **Input tokens:** $0.25 per 1M tokens
- **Output tokens:** $0.75 per 1M tokens
- **Cost per function:** ~$0.002-0.01

| Daily Volume | Monthly Cost |
|--------------|-------------|
| 100 functions/day | $6-$30/month |
| 1,000 functions/day | $60-$300/month |
| 10,000 functions/day | $600-$3,000/month |

**Note:** These are raw LLM costs. Add sandbox execution costs (~50% more) for testing pipeline.

---

## Cost Optimization Strategies

### Compute Options

1. **Mercury 2 via API** - Fast, managed, good for initial scale (~$0.002-0.01/function)
2. **RunPod On-Demand** - Self-hosted GPUs, pay per second, 60-80% cheaper at scale
3. **Hybrid approach** - API for speed, RunPod for batch processing

### Optimization Tactics

1. **Use smaller models for simpler functions** - Not every function needs Mercury 2
2. **Cache generated templates** - Reuse common patterns
3. **Batch discovery scans** - Reduce API calls
4. **Smart testing** - Skip expensive tests for low-risk functions
5. **Self-host smaller models** - Once volume justifies it
6. **RunPod integration** - Spin up GPU instances on-demand for batch generation

## Additional Recommendations

### 1. Caching Layer
- Cache common function templates to reduce LLM calls by 30-50%
- Store successful generation patterns for reuse
- Implement semantic caching for similar prompts

### 2. Quality-First Strategy
- Start with private functions → gradually make public
- Add manual review tier for high-impact functions
- Track usage metrics to identify high-value functions

### 3. Hybrid Model Strategy
- Use fast/cheap models (e.g., Mistral 7B) for simple functions
- Reserve Mercury 2 for complex functions only
- Auto-select model based on complexity scoring

### 4. Feedback Loop
- Track which factory functions are actually invoked
- Prioritize generation of functions similar to popular ones
- Retire unused functions automatically

### 5. Human-in-the-Loop (Optional)
- Add manual approval queue for certain categories
- Allow community voting on desired functions
- Partner with top developers for curated functions

---

## Success Criteria

- [ ] Pipeline generates 10-20 functions/day initially
- [ ] Quality score >= 70% for all published functions
- [ ] Zero critical security issues
- [ ] Cold start time <= 2000ms
- [ ] 90%+ test pass rate
- [ ] Fully automated discovery → publish workflow
