# FunctionFly AI Agents Offerings

## Overview

FunctionFly provides a comprehensive platform for AI agents to discover, execute, and manage serverless functions as tools. The platform is designed to support AI agents with trust scoring, policy enforcement, economic controls, and multi-agent orchestration.

---

## 1. Agent Registration & Identity Management

### API Endpoints
- `POST /v1/agent/register` - Register a new agent
- `GET /v1/agent` - List all agents
- `GET /v1/agent/{agent_id}` - Get agent details
- `DELETE /v1/agent/{agent_id}` - Delete an agent

### Features
- **Agent Identity**: Agents are identified by `org/agent-name` format
- **Plan Tiers**: `agent_starter`, `agent_pro`, `agent_enterprise`
- **Swarm Roles**: `worker`, `manager`, `infrastructure`
- **Capabilities**: Map of agent capabilities (e.g., `{"ai": true, "web": true}`)
- **API Key Authentication**: Agents authenticate via `X-Agent-API-Key` header

### Key Models
```go
type AgentIdentity struct {
    AgentID           string         // "org/agent-name"
    Name              string
    Description       string
    PlanTier          string         // agent_starter | agent_pro | agent_enterprise
    SwarmRole         string         // worker | manager | infrastructure
    Capabilities      map[string]any
    AutonomousEnabled bool
    EvolutionEnabled  bool
    TrustScore        float64        // 0-100
    EconomicScore     float64        // 0-100
}
```

---

## 2. Function Discovery (Tool Calling)

### API Endpoints
- `GET /v1/agent/discover` - Search functions with filters
- `GET /v1/agent/discover/{author}/{name}` - Get specific function

### Query Parameters
- `q` - Search query
- `category` - Function category
- `deterministic` - Filter for deterministic functions only
- `trust_score_min` - Minimum trust score
- `tags` - Filter by tags
- `limit` / `offset` - Pagination

### Discovery Response
```go
type DiscoveryResult struct {
    URI            string          // fx://org/name
    Author         string
    Name           string
    Version        string
    Title          string
    Description    string
    Schema         json.RawMessage // Input/output schema
    PricingPerCall float64
    Deterministic  bool
    TrustScore     float64
    SuccessRate    float64
    Tags           json.RawMessage
    Capabilities   json.RawMessage
    SideEffects    string
    Category       string
}
```

---

## 3. Function Execution (Tool Calling)

### API Endpoints
- `POST /v1/agent/execute/{author}/{name}` - Execute function
- `POST /v1/agent/execute/{author}/{name}/{version}` - Execute specific version

### Request
```go
type AgentExecuteRequest struct {
    Input     json.RawMessage
    SessionID string          // Optional session for tracking
    CallDepth int             // Recursion depth tracking
}
```

### Response
```go
type AgentExecuteResponse struct {
    OK          bool
    Data        json.RawMessage
    ExecutionID string
    SessionID   string
    DurationMs  int
    Version     string
    CostUSD     float64
    CallDepth   int
    Cached      bool
}
```

### Function URI Format
- `fx://author/function-name` - Latest version
- `fx://author/function-name@v1.0.0` - Specific version

---

## 4. Policy & Safety Controls

### API Endpoints
- `GET /v1/agent/{agent_id}/policy` - Get agent policy
- `PUT /v1/agent/{agent_id}/policy` - Update agent policy

### Policy Features
```go
type BehavioralPolicy struct {
    AgentID             string
    MaxExecutionDepth   int      // Default: 10
    MaxRecursionDepth   int      // Default: 3
    MaxWallTimeMs       int      // Default: 300000 (5 min)
    MaxMemoryGrowthMB   int      // Default: 512
    ForbiddenFunctions  []string // Block specific functions
    DeterministicOnly   bool     // Only allow deterministic functions
    AllowedCapabilities []string // Restrict capabilities
}
```

### Policy Violation Codes
- `LOOP_DETECTED` - Same function called with identical input repeatedly
- `DEPTH_EXCEEDED` - Max execution depth exceeded
- `FUNCTION_BLOCKED` - Function in forbidden list
- `CAPABILITY_DENIED` - Required capability not allowed
- `DETERMINISTIC_ONLY` - Non-deterministic function called

---

## 5. Quota & Rate Limiting

### API Endpoints
- `PUT /v1/agent/{agent_id}/quota` - Update quota
- `GET /v1/agent/{agent_id}/usage` - Get usage statistics

### Quota Configuration
```go
type AgentQuotaConfig struct {
    MaxCallsPerMinute   int
    MaxCallsPerDay      int
    MaxStateWritesPerHr int
    MaxCostPerExecution float64
    MaxDailySpendUSD    float64
    AllowedFunctions    []string
    ForbiddenFunctions  []string
}
```

---

## 6. Session Management

### API Endpoints
- `POST /v1/agent/{agent_id}/session/start` - Start session
- `POST /v1/agent/{agent_id}/session/{session_id}/end` - End session
- `GET /v1/agent/{agent_id}/session/{session_id}` - Get session info

### Session Tracking
- Call count per session
- Total cost per session
- Call graph visualization
- Session duration

---

## 7. Attribution & Observability

### API Endpoints
- `GET /v1/agent/{agent_id}/executions` - List executions
- `GET /v1/agent/{agent_id}/executions/{exec_id}` - Get execution details
- `GET /v1/agent/{agent_id}/analytics` - Get analytics

### Execution Records
```go
type AgentExecutionRecord struct {
    AgentID          string
    FunctionURI      string    // fx://org/name@version
    ExecutionID      string
    SessionID        string
    CallDepth        int
    RetryCount       *int
    InputHash        string    // SHA-256
    OutputHash       string    // SHA-256
    MemoryBeforeHash string    // Agent context before
    MemoryAfterHash  string    // Agent context after
    CostUSD          float64
    LatencyMs        int
    Outcome          string    // success | error | timeout | policy_violation
    ErrorCode        *string
    PolicyViolation  *string
}
```

---

## 8. Billing & Economics

### API Endpoints
- `GET /v1/agent/{agent_id}/billing/summary` - Billing summary
- `PUT /v1/agent/{agent_id}/billing/spend-cap` - Set spend cap
- `GET /v1/agent/{agent_id}/cost-breakdown` - Cost breakdown
- `GET /v1/agent/{agent_id}/credits/balance` - Credit balance
- `POST /v1/agent/{agent_id}/credits/purchase` - Purchase credits

### Wallet System
```go
type AgentWallet struct {
    AgentID          string
    BalanceUSD       float64
    EscrowBalanceUSD float64
    TotalEarnedUSD   float64
    TotalSpentUSD    float64
}
```

---

## 9. Multi-Agent Swarm Orchestration

### Swarm Roles
- **Worker**: Executes tasks
- **Manager**: Coordinates workers
- **Infrastructure**: Provides shared services

### API Endpoints (via swarm handler)
- Spawn child agents
- Task delegation between agents
- Capability discovery between agents
- Inter-agent messaging

### Message Types (A2A Protocol)
- `task_delegation`
- `task_result`
- `query` / `response`
- `capability_discovery`
- `heartbeat`
- `evolution_proposal`
- `budget_request`

---

## 10. Agent Autonomy (Scheduled Execution)

### API Endpoints
- Create autonomy schedules
- Get schedules for agent
- Get active schedules

### Schedule Types
- **One-time**: Execute at specific time
- **Recurring**: Cron-based execution

---

## 11. Agent Evolution & Learning

### Performance Analysis
- Success rate tracking
- Latency analysis
- Cost analysis
- Failure categorization

### Self-Optimization
```go
type Optimization struct {
    OptimizationType string  // timeout_adjustment, caching, batch_processing, etc.
    Description      string
    ExpectedImpact  map[string]any
    Implementation   string  // low | medium | high
    Status           string  // pending | approved | rejected | applied
}
```

---

## 12. Function Marketplace

### Marketplace Features
- Trust score ranking (30% weight)
- Economic score (25% weight)
- Reliability score (20% weight)
- ROI score (15% weight)
- Call volume (10% weight)

### Pricing Models
- **Free**: No charge
- **Per-call**: Fixed price per execution
- **Subscription**: Monthly recurring
- **Revenue share**: Percentage of delegate value

### Agent/Function Listings
- Public marketplace for agents
- Public marketplace for functions
- Rating and review system

---

## 13. Trust & Verification System

### Trust Score Components
- Success rate
- Reliability score
- Deterministic score
- Execution latency
- Historical performance

### Function Verification
- Source code signing
- Malware scanning
- Approval workflows
- Verification status tracking

---

## 14. Capabilities System

### Available Capabilities
Functions declare capabilities in their manifest:
```json
{
  "capabilities": ["ai", "network", "storage", "io"]
}
```

### Agent Capability Filtering
- Agents can filter functions by required capabilities
- Policy engine enforces capability restrictions

---

## 15. AI-Powered Function Generation

### Generation Service
- Generate functions from natural language prompts
- Support for multiple runtimes (python3.11, nodejs20, etc.)
- Input/output schema specification
- Deterministic flagging

### Generation Request
```go
type GenerationRequest struct {
    AgentID       string
    Name          string
    Description   string
    Category      string
    InputSchema   map[string]any
    OutputSchema  map[string]any
    Runtime       string
    Prompt        string
    Model         string
    Deterministic bool
    Tags          []string
}
```

---

## Summary

FunctionFly provides a comprehensive AI agent platform with:

| Category | Features |
|----------|----------|
| **Discovery** | Search, filter, trust-scored function registry |
| **Execution** | Tool calling with versioning, caching, cost tracking |
| **Safety** | Policy engine, loop detection, depth limits |
| **Economics** | Quotas, rate limits, billing, credits |
| **Observability** | Execution records, sessions, analytics |
| **Orchestration** | Multi-agent swarms, task delegation |
| **Autonomy** | Scheduled execution, self-optimization |
| **Marketplace** | Listings, ratings, pricing models |
| **Trust** | Verification, signing, malware scanning |
| **Generation** | AI-powered function creation |

The platform follows the `fx://author/name@version` URI scheme for function identification and provides RESTful APIs for all agent operations.
