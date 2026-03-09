# AI Function Factory Implementation Plan

## Executive Summary

This document outlines the complete implementation of the **FunctionFly AI Function Factory** - an autonomous system that automatically discovers opportunities, generates functions, tests them, publishes them, and continuously optimizes them. The platform already has ~85% of the infrastructure built; this plan completes the remaining 15% to create a fully operational factory.

### Current State

The project already has substantial infrastructure:

| Component | Location | Status |
|-----------|----------|--------|
| Generation Service | `internal/agent/generation/service.go` | ✅ Interface exists, needs LLM implementation |
| Evolution Service | `internal/agent/evolution/service.go` | ✅ Complete |
| Autonomy Service | `internal/agent/autonomy/service.go` | ✅ Complete |
| Publisher | `internal/agent/deployment/publisher.go` | ✅ Complete |
| Learning/Optimizer | `internal/agent/learning/optimizer.go` | ✅ Complete |
| Execution Queue | `internal/queue/execution_queue.go` | ✅ Complete |
| Agent Models | `internal/agent/identity/models.go` | ✅ Complete |

### What's Missing

| Component | Priority | Description |
|-----------|----------|-------------|
| Discovery Engine | HIGH | Scans GitHub, StackOverflow, Reddit for opportunities |
| Testing Pipeline | HIGH | Validation, security scanning, sandbox execution |
| Pipeline Orchestrator | HIGH | Main loop connecting all agents |
| OpenRouter Integration | HIGH | Actual LLM code generation |
| API Handlers | MEDIUM | REST endpoints for factory operations |
| Quality Guardrails | MEDIUM | Code quality thresholds, benchmarks |
| Admin Dashboard UI | LOW | Factory monitoring and control |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      AI FUNCTION FACTORY                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐              │
│  │   DISCOVERY │───►│  GENERATOR  │───►│   TESTING   │              │
│  │    ENGINE   │    │   (Mercury) │    │   PIPELINE  │              │
│  └──────────────┘    └──────────────┘    └──────────────┘              │
│         │                                        │                       │
│         │                                        ▼                       │
│         │                              ┌──────────────┐                │
│         │                              │   PUBLISHER  │                │
│         │                              └──────────────┘                │
│         │                                        │                       │
│         │                                        ▼                       │
│         │                              ┌──────────────┐                │
│         └─────────────────────────────►│  OPTIMIZER   │                │
│                                        └──────────────┘                │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     ORCHESTRATION LAYER                          │   │
│  │  • Pipeline Scheduler (cron)  • Queue Management  • Error Handling│   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Component Specifications

### 1. Discovery Engine (`internal/agent/discovery/`)

The Discovery Engine continuously scans for function opportunities.

#### Sources
- GitHub Issues (most valuable - explicit feature requests)
- StackOverflow Questions (problem descriptions)
- Reddit r/dev, r/learnprogramming (pain points)
- API Documentation (missing integrations)
- Product Forums (feature requests)

#### Implementation

```go
// Service: internal/agent/discovery/service.go

type Source interface {
    Scan(ctx context.Context) ([]Opportunity, error)
    Name() string
}

type Opportunity struct {
    ID          uuid.UUID
    Source      string
    SourceID    string
    Title       string
    Description string
    Category    string
    Tags        []string
    DemandScore float64 // 0-100
    Complexity  int     // 1-10
    Validated   bool
}

type Service struct {
    db        *gorm.DB
    sources   []Source
    llmClient LLMClient
}
```

#### Sources to Implement
1. **GitHubScanner** - Scans issues with "feature request", "would be nice", "need" labels
2. **StackOverflowScanner** - Searches for unanswered questions with high views
3. **RedditScanner** - Monitors dev subreddits for common problems
4. **APIDocScanner** - Identifies gaps in popular API integrations

---

### 2. Generator Service Enhancement

The existing generation service needs an actual LLM implementation.

#### OpenRouter Integration

```go
// internal/agent/generation/llm.go

type LLMClient interface {
    GenerateCode(ctx context.Context, req *GenerationRequest) (string, error)
}

type OpenRouterClient struct {
    apiKey     string
    model      string // "mistralai/ministral-8b" or "deepseek/deepseek-chat"
    baseURL    string
    httpClient *http.Client
}

type GenerationRequest struct {
    AgentID        string
    Name           string
    Description    string
    Category       string
    InputSchema    map[string]any
    OutputSchema   map[string]any
    Runtime        string // "python3.11", "nodejs20"
    Prompt         string
    Model          string
    Deterministic  bool
    Tags           []string
}
```

#### Prompt Engineering

The generator uses structured prompts for consistent output:

```
You are FunctionFly's AI Function Generator. Generate a production-ready serverless function.

Function Name: {name}
Description: {description}
Category: {category}
Runtime: {runtime}
Input Schema: {input_schema}
Output Schema: {output_schema}

Requirements:
1. Write clean, documented code
2. Include error handling
3. Include test cases
4. Follow security best practices
5. Optimize for cold start performance

Output format: JSON with "code", "tests", "manifest"
```

---

### 3. Testing Pipeline (`internal/agent/testing/`)

Before publishing, functions must pass validation.

#### Testing Stages

| Stage | Description | Timeout |
|-------|-------------|---------|
| Syntax Validation | Parse, lint, type check | 5s |
| Security Scan | Malware, injection patterns | 10s |
| Unit Tests | Run generated tests | 30s |
| Sandbox Execution | Run with sample input | 30s |
| Performance Benchmark | Cold start, latency | 60s |

#### Implementation

```go
// internal/agent/testing/service.go

type TestResult struct {
    FunctionID    uuid.UUID
    Stage         string    // "syntax", "security", "unit", "execution", "performance"
    Passed        bool
    Score         float64   // 0-100
    DurationMs    int
    Error         string
    Details       map[string]any
}

type Service struct {
    db           *gorm.DB
    sandbox      SandboxExecutor
    security     SecurityScanner
    benchmark    BenchmarkRunner
}

func (s *Service) RunTests(ctx context.Context, functionID uuid.UUID, code, runtime string) ([]TestResult, error)
```

#### Security Checks

```go
type SecurityScanner struct {
    malwarePatterns   []string
    injectionPatterns []string
    networkPatterns   []string
}

func (s *SecurityScanner) Scan(code string) (bool, []string)
```

Security patterns to detect:
- `eval()`, `exec()`, `compile()` - Code execution
- `subprocess`, `os.system` - Shell commands
- `requests`, `urllib` without restrictions - External network
- `__import__`, `importlib` - Dynamic imports
- Infinite loops (detect via AST analysis)
- Memory exhaustion patterns

---

### 4. Publisher Enhancement

The existing publisher needs enhancement for the factory.

```go
// internal/agent/deployment/publisher.go (enhancement)

type PublishOptions struct {
    FunctionID    uuid.UUID
    Author        string // "functionfly-ai" for platform-owned
    Name          string
    Title         string
    Description   string
    Category      string
    Tags          []string
    Runtime       string
    IsPublic      bool
    PricingModel  string // "free", "per_call"
    PricePerCall  *float64
}

func (p *Publisher) PublishFromFactory(ctx context.Context, opts *PublishOptions) (*PublishedFunction, error)
```

---

### 5. Optimizer Integration

The learning/optimizer system already exists; integrate it into the factory loop.

```go
// Integration point after publishing

func (s *FactoryService) MonitorAndOptimize(ctx context.Context, functionID uuid.UUID) {
    // Schedule performance monitoring
    schedule := &identity.AutonomySchedule{
        AgentID:      "functionfly-ai",
        ScheduleType: "recurring",
        CronExpression: strPtr("0 */6 * * *"), // Every 6 hours
        ActionType:   "evolve",
        ActionPayload: map[string]any{
            "function_id": functionID,
            "action":     "optimize",
        },
    }
    s.autonomy.CreateSchedule(ctx, schedule)
}
```

---

### 6. Pipeline Orchestrator (`internal/agent/factory/`)

The main orchestration layer that ties everything together.

```go
// internal/agent/factory/service.go

type FactoryService struct {
    db           *gorm.DB
    discovery    *discovery.Service
    generator    *generation.Service
    testing      *testing.Service
    publisher    *deployment.Publisher
    optimizer    *learning.Optimizer
    autonomy     *autonomy.Service
}

type PipelineConfig struct {
    DailyTarget      int  // 10-20 functions/day initially
    MaxConcurrent    int  // Parallel pipeline runs
    TimeoutMinutes   int  // Max time per function
    MinQualityScore  float64 // 70% minimum
}

func (s *FactoryService) RunPipeline(ctx context.Context) (*PipelineResult, error)
```

#### Pipeline Flow

```
1. Discovery
   └─> Get opportunities from all sources
   └─> Score and rank by demand, complexity
   └─> Select top N for generation

2. Generation
   └─> For each opportunity:
       └─> Generate function code via LLM
       └─> Generate tests
       └─> Create manifest

3. Testing
   └─> Run syntax validation
   └─> Run security scan
   └─> Run unit tests
   └─> Run sandbox execution
   └─> Run performance benchmark

4. Publishing
   └─> If all tests pass (score >= 70):
       └─> Create function in registry
       └─> Set visibility (public/private)
       └─> Configure pricing

5. Optimization Loop
   └─> Schedule performance monitoring
   └─> Schedule evolution analysis
   └─> Auto-optimize if needed
```

---

### 7. API Handlers

REST endpoints for factory operations.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/factory/status` | GET | Factory health and statistics |
| `/api/v1/factory/pipeline/run` | POST | Manually trigger pipeline |
| `/api/v1/factory/discover` | GET | Get discovered opportunities |
| `/api/v1/factory/generate` | POST | Generate a specific function |
| `/api/v1/factory/functions` | GET | List factory-generated functions |
| `/api/v1/factory/config` | GET/PUT | Factory configuration |

---

## Database Schema Additions

```sql
-- Discovery opportunities
CREATE TABLE factory_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source VARCHAR(100) NOT NULL,
    source_id VARCHAR(255),
    title TEXT NOT NULL,
    description TEXT,
    category VARCHAR(100),
    tags TEXT[],
    demand_score DECIMAL(5,2),
    complexity INT,
    status VARCHAR(50) DEFAULT 'discovered', -- discovered, generating, testing, published, rejected
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Factory pipeline runs
CREATE TABLE factory_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(50) DEFAULT 'running',
    functions_created INT DEFAULT 0,
    functions_failed INT DEFAULT 0,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    error_message TEXT
);

-- Factory function versions
CREATE TABLE factory_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL,
    version VARCHAR(20) NOT NULL,
    generation_model VARCHAR(100),
    generation_prompt_hash VARCHAR(64),
    quality_score DECIMAL(5,2),
    test_results JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## Quality Guardrails

### Minimum Standards

| Metric | Threshold | Action |
|--------|-----------|--------|
| Code Quality Score | >= 70% | Pass/Fail |
| Test Pass Rate | >= 90% | Pass/Fail |
| Cold Start Time | <= 2000ms | Warning if exceeded |
| Execution Time | <= 500ms (avg) | Flag for optimization |
| Security Issues | 0 Critical | Reject if found |
| Dependencies | <= 10 | Warning if exceeded |

### Safety Rules

1. **No External Network by Default** - Functions cannot make outbound requests unless explicitly enabled
2. **Deterministic Preferred** - Functions with deterministic cert get priority publishing
3. **Sandboxed Execution** - All tests run in isolated containers
4. **Rate Limiting** - Max 100 functions/day initial, scale with trust

---

## Configuration

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

## Implementation Priority

### Phase 1: Core Pipeline (Week 1-2)
1. [x] Discovery Engine service and GitHub scanner
2. [x] OpenRouter LLM integration
3. [x] Testing pipeline with security scanning
4. [x] Basic publisher integration
5. [x] Pipeline orchestrator

### Phase 2: Enhancement (Week 3-4)
1. [ ] Additional discovery sources (StackOverflow, Reddit)
2. [ ] Performance benchmarking
3. [ ] Quality guardrails enforcement
4. [ ] API handlers

### Phase 3: Scale (Week 5-8)
1. [ ] Admin dashboard UI
2. [ ] Advanced optimization
3. [ ] Monitoring and alerting
4. [ ] Production hardening

---

## Cost Analysis

### Running Costs (Initial)

| Component | Monthly Cost |
|-----------|--------------|
| Mercury 2 (LLM) | ~$50 (100 functions/day) |
| Execution Infra | ~$100 |
| Monitoring | ~$20 |
| **Total** | **~$170/month** |

### Cost Per Function

| Stage | Cost |
|-------|------|
| Discovery | ~$0.0001 (API calls) |
| Generation | ~$0.004 (Mercury 2) |
| Testing | ~$0.001 (sandbox) |
| **Total** | **~$0.005/function** |

---

## Success Metrics

### KPIs

| Metric | Month 1 Target | Month 3 Target | Month 6 Target |
|--------|----------------|----------------|----------------|
| Functions Published | 300 | 2,000 | 10,000 |
| Daily Production | 10/day | 65/day | 200/day |
| Quality Pass Rate | 60% | 75% | 85% |
| Avg Function Quality | 75% | 80% | 85% |
| Cold Start Time | <2s | <1.5s | <1s |

---

## Integration Points

### With Existing Systems

1. **Agent Identity** - Use existing `identity.Function` model
2. **Autonomy** - Use existing `autonomy.Service` for scheduling
3. **Evolution** - Use existing `evolution.Service` for optimization
4. **Publisher** - Use existing `deployment.Publisher`
5. **Queue** - Use existing `ExecutionQueue` for sandboxed tests

### With External Services

1. **OpenRouter** - Code generation via API
2. **GitHub API** - Issue scanning
3. **StackExchange API** - Question search
4. **Reddit API** - Subreddit monitoring

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| LLM generates bad code | Mandatory testing pipeline, quality thresholds |
| Discovery finds spam | Human review queue for new categories |
| Cost overruns | Daily budget limits, auto-pause at threshold |
| Security vulnerabilities | Sandboxed execution, security scanning |
| Platform abuse | Rate limiting, trust scoring |

---

## Conclusion

This implementation plan completes the AI Function Factory by building on the existing 85% of infrastructure. The factory will enable FunctionFly to automatically generate and publish functions at scale, creating a self-growing ecosystem without waiting for user submissions.

The initial target of 10-20 functions/day at ~$0.005/function makes this economically viable from day one, with the platform capturing 100% of revenue until the user marketplace launches.
