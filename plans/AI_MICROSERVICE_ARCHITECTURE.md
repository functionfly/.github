# FunctionFly AI Microservice Architecture

## Executive Summary: "FlyMind" - The Intelligence Layer for Serverless Functions

### Core Concept

FlyMind is an innovative Python AI microservice that transforms FunctionFly from a pure serverless platform into an **intelligent serverless ecosystem**. Rather than being a generic AI wrapper, FlyMind leverages unique insights from function execution patterns, user behavior, and deployment analytics to provide differentiated intelligence that directly improves function performance, developer productivity, and cost efficiency.

### What Makes This Different

Traditional AI services offer generic capabilities. FlyMind is purpose-built for a serverless function platform, giving it access to:

- **Execution telemetry** from millions of function invocations
- **Deployment patterns** across multiple edge targets (Cloudflare, Vercel, Fly.io, Deno Deploy)
- **User behavior data** from the function registry and dashboard
- **Cost and latency metrics** at granular function-level detail

This positions FlyMind as a **product differentiator** rather than a commodity AI wrapper.

---

## 1. Innovative Capabilities

### 1.1 Intelligent Request Routing & Load Balancing

**What it does**: Analyzes real-time request patterns, user geography, edge target health, and historical latency to route function invocations to the optimal edge target.

**Value**: Reduces average latency by 20-40% by avoiding congested or slow edges.

**Implementation**:
- Collects latency metrics from function executions
- Monitors edge target health via existing `/healthz` endpoints
- Uses a lightweight ML model (gradient boosted trees via scikit-learn) to predict optimal routing
- Falls back to geo-based routing when insufficient training data

### 1.2 Predictive Cold Start Prewarming

**What it does**: Predicts incoming request spikes and proactively warms function instances before they arrive.

**Value**: Eliminates cold starts for predictable traffic patterns (scheduled jobs, marketing campaigns, API windows).

**Implementation**:
- Analyzes time-series execution data to identify patterns
- Uses Prophet or similar forecasting to predict request volume
- Triggers prewarming via orchestrator API before predicted spikes
- Integrates with Redis for distributed warming coordination

### 1.3 AI-Powered Function Optimization

**What it does**: Analyzes function code and execution patterns to recommend or automatically apply optimizations.

**Value**: Helps developers reduce function latency, memory usage, and costs.

**Capabilities**:
- **Code analysis**: Static analysis of Python/Go/Node functions for optimization opportunities
- **Dependency audit**: Identifies heavy dependencies that could be replaced with lighter alternatives
- **Memory profiling**: Recommends memory allocation changes based on execution telemetry
- **Auto-tuning**: (Phase 2) Automatically adjusts timeout, memory, and concurrency settings

### 1.4 Smart Caching with ML-Based Prepopulation

**What it does**: Learns request patterns and proactively caches function outputs that are likely to be requested again.

**Value**: Reduces execution costs for deterministic functions by 60-90%.

**Implementation**:
- Analyzes input/output pairs to identify cacheable patterns
- Uses locality-sensitive hashing (LSH) for fast input matching
- Coordinates cache population across distributed Redis instances
- Supports TTL-based and pattern-based invalidation

### 1.5 Anomaly Detection in Function Executions

**What it does**: Continuously monitors function execution metrics to detect anomalies in latency, errors, and resource usage.

**Value**: Enables proactive alerting before users notice issues.

**Implementation**:
- Collects per-execution metrics via existing status handlers
- Uses isolation forest or one-class SVM for anomaly detection
- Generates alerts via existing notification system
- Provides root cause analysis via LLM-powered investigation

### 1.6 Natural Language Interface for Platform Operations

**What it does**: Allows developers to interact with FunctionFly using natural language through a chat interface.

**Value**: Lowers barrier to entry for platform operations.

**Capabilities**:
- "Show me functions with high error rates"
- "Deploy my function to eu-west"
- "What's consuming my most quota?"
- "Create a scheduled job for data processing"
- Generates appropriate API calls from natural language

### 1.7 Semantic Search Across Function Registry

**What it does**: Enhances existing keyword search with semantic understanding of function purpose and usage.

**Value**: Helps developers find relevant functions even when they don't know exact names.

**Implementation**:
- Uses existing pgvector infrastructure for embeddings
- Generates embeddings for function code, documentation, and metadata
- Supports hybrid search (keyword + semantic)
- Integrates with existing registry handlers

### 1.8 AI-Assisted Function Debugging

**What it does**: Analyzes function execution failures and provides intelligent debugging assistance.

**Value**: Reduces debugging time for complex function failures.

**Capabilities**:
- Analyzes error traces and function code
- Identifies likely root causes
- Suggests fixes with explanations
- Can generate corrected code snippets

### 1.9 Content Moderation for Function I/O

**What it does**: Scans function inputs and outputs for policy violations, sensitive data, and security concerns.

**Value**: Ensures platform safety without manual review overhead.

**Implementation**:
- Integrates with existing verification services (ClamAV, YARA)
- Adds PII detection and redaction
- Supports custom moderation rules per function
- Uses lightweight ML models for content classification

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              FunctionFly Platform                                │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                    │
│  │   Dashboard  │    │  Go CLI/SDK  │    │   Web/API    │                    │
│  │   (React)    │    │   (Users)     │    │   (Agents)   │                    │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘                    │
│         │                    │                    │                            │
│         └────────────────────┼────────────────────┘                            │
│                              │                                                  │
│                              ▼                                                  │
│                 ┌────────────────────────┐                                      │
│                 │   Go Orchestrator API   │                                      │
│                 │        (Port 8080)      │                                      │
│                 └────────────┬─────────────┘                                      │
│                              │                                                  │
│         ┌────────────────────┼────────────────────┐                              │
│         │                    │                    │                              │
│         ▼                    ▼                    ▼                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                    │
│  │  PostgreSQL  │    │     Redis     │    │   FlyMind    │                    │
│  │   (Primary)  │    │   (Cache)     │    │  (AI Service) │                    │
│  └──────────────┘    └──────────────┘    └───────┬────────┘                    │
│                                                   │                               │
│                              ┌────────────────────┼────────────────────┐        │
│                              │                    │                    │        │
│                              ▼                    ▼                    ▼        │
│                     ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│                     │   Ollama     │    │   Vector DB  │    │  Model Cache │   │
│                     │  (Local LLM) │    │  (pgvector)  │    │    (Redis)    │   │
│                     └──────────────┘    └──────────────┘    └──────────────┘   │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 FlyMind Service Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              FlyMind AI Microservice                             │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                           API Gateway (FastAPI)                             │ │
│  │                    Port 8081 | /api/v1/ai/* endpoints                       │ │
│  └─────────────────────────────────┬───────────────────────────────────────────┘ │
│                                      │                                            │
│         ┌────────────────────────────┼────────────────────────────┐             │
│         │                            │                            │             │
│         ▼                            ▼                            ▼             │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────────┐    │
│  │  Route Optimizer │    │  Anomaly Detector│    │  Chat/NL Interface   │    │
│  │  (Prediction)    │    │  (Monitoring)    │    │  (LLM-powered)       │    │
│  └──────────────────┘    └──────────────────┘    └──────────────────────┘    │
│                                                                                  │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────────┐    │
│  │ Prewarming Engine│    │  Code Optimizer  │    │  Semantic Search     │    │
│  │  (Forecasting)   │    │  (Analysis)      │    │  (Embeddings)        │    │
│  └──────────────────┘    └──────────────────┘    └──────────────────────┘    │
│                                                                                  │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────────┐    │
│  │  Cache Predictor │    │  Debug Assistant │    │  Content Moderator  │    │
│  │  (ML-based)      │    │  (LLM-powered)    │    │  (Classification)    │    │
│  └──────────────────┘    └──────────────────┘    └──────────────────────┘    │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                      Provider Abstraction Layer                              │ │
│  │         ┌─────────┐  ┌──────────┐  ┌─────────┐  ┌────────────┐            │ │
│  │         │ OpenAI  │  │Anthropic │  │ Ollama  │  │  Azure    │            │ │
│  │         └─────────┘  └──────────┘  └─────────┘  └────────────┘            │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                           Data Layer                                         │ │
│  │    ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                 │ │
│  │    │ PostgreSQL│  │  Redis   │  │ Vector DB│  │   S3/    │                 │ │
│  │    │(Metrics) │  │ (Cache)  │  │(pgvector)│  │   R2     │                 │ │
│  │    └──────────┘  └──────────┘  └──────────┘  └──────────┘                 │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 Component Responsibilities

| Component | Responsibility | Dependencies |
|-----------|---------------|--------------|
| API Gateway | Request routing, auth, rate limiting | Go orchestrator (auth), Redis |
| Route Optimizer | ML-based edge routing decisions | Execution metrics, edge health |
| Anomaly Detector | Real-time anomaly detection | Execution metrics stream |
| Prewarming Engine | Forecasting and prewarming | Execution history, orchestrator API |
| Code Optimizer | Static analysis and recommendations | Function code, execution profiles |
| Semantic Search | Embedding generation and similarity | pgvector, function registry |
| Cache Predictor | ML-based cache prediction | Execution patterns, Redis |
| Chat Interface | Natural language processing | LLM providers, function metadata |
| Debug Assistant | Error analysis and suggestions | Execution traces, function code |
| Content Moderator | Input/output scanning | Existing verification services |

---

## 3. API Design

### 3.1 REST API Endpoints

#### Core AI Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/ai/chat` | Natural language interface |
| GET | `/api/v1/ai/chat/history` | Get chat history |
| POST | `/api/v1/ai/optimize` | Get function optimization suggestions |
| POST | `/api/v1/ai/optimize/apply` | Apply optimization recommendations |
| GET | `/api/v1/ai/search` | Semantic search across functions |
| POST | `/api/v1/ai/analyze` | Analyze function for issues |
| POST | `/api/v1/ai/debug` | Debug failed execution |
| POST | `/api/v1/ai/moderate` | Moderate function I/O |

#### Intelligence Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/ai/routing/recommend` | Get routing recommendations |
| POST | `/api/v1/ai/routing/feedback` | Report routing decision quality |
| GET | `/api/v1/ai/prewarm/status` | Get prewarming status |
| POST | `/api/v1/ai/prewarm/trigger` | Manually trigger prewarming |
| GET | `/api/v1/ai/cache/recommendations` | Get cache recommendations |
| GET | `/api/v1/ai/anomalies` | Get detected anomalies |
| POST | `/api/v1/ai/anomalies/acknowledge` | Acknowledge anomaly |

#### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/ai/config` | Get AI service configuration |
| PUT | `/api/v1/ai/config` | Update AI service configuration |
| GET | `/api/v1/ai/providers` | List configured providers |
| POST | `/api/v1/ai/providers/test` | Test provider configuration |
| GET | `/api/v1/ai/health` | Health check endpoint |

### 3.2 gRPC for Inter-Service Communication

For high-performance communication with the Go orchestrator:

```protobuf
// flymind.proto
service FlyMindService {
  // Get routing recommendation for a request
  rpc GetRoutingRecommendation(RoutingRequest) returns (RoutingRecommendation);
  
  // Report execution outcome for learning
  rpc ReportExecutionOutcome(ExecutionOutcome) returns (Empty);
  
  // Check if function should be prewarmed
  rpc ShouldPrewarm(PrewarmQuery) returns (PrewarmDecision);
  
  // Get cache key for input
  rpc GetCacheKey(CacheQuery) returns (CacheDecision);
  
  // Moderate function I/O
  rpc ModerateContent(ModerationRequest) returns (ModerationResponse);
}

message RoutingRequest {
  string function_id = 1;
  string user_id = 2;
  string geography = 3;
  map<string, string> metadata = 4;
}

message RoutingRecommendation {
  string target = 1;  // cloudflare, vercel, fly, deno
  float confidence = 2;
}

message ExecutionOutcome {
  string execution_id = 1;
  string function_id = 2;
  string target = 3;
  int64 latency_ms = 4;
  bool success = 5;
}

message PrewarmQuery {
  string function_id = 1;
  int64 timestamp = 2;
}

message PrewarmDecision {
  bool should_prewarm = 1;
  int32 instances = 2;
}

message CacheQuery {
  string function_id = 1;
  bytes input_hash = 2;
}

message CacheDecision {
  bool should_cache = 1;
  string cached_output_id = 2;
  int32 ttl_seconds = 3;
}

message ModerationRequest {
  string function_id = 1;
  bytes content = 2;
  ContentType type = 3;  // INPUT or OUTPUT
}

message ModerationResponse {
  bool allowed = 1;
  repeated string violations = 2;
  float risk_score = 3;
}
```

### 3.3 Request/Response Examples

#### Chat Interface

```json
// POST /api/v1/ai/chat
{
  "message": "Show me functions with high error rates",
  "context": {
    "tenant_id": "uuid",
    "user_id": "uuid"
  }
}

// Response
{
  "response": "Here are your functions with elevated error rates:",
  "data": {
    "functions": [
      {
        "name": "process-image",
        "error_rate": 0.15,
        "recent_errors": 47,
        "suggestion": "Consider adding retry logic for network timeouts"
      }
    ]
  },
  "actions": [
    {
      "type": "VIEW_FUNCTION",
      "payload": {"function_id": "uuid"}
    }
  ]
}
```

#### Function Optimization

```json
// POST /api/v1/ai/optimize
{
  "function_id": "uuid"
}

// Response
{
  "optimizations": [
    {
      "type": "dependency_replacement",
      "description": "Replace 'pillow' with 'Pillow-SIMD' for 3x faster image processing",
      "impact": {
        "latency_reduction": 0.65,
        "cost_reduction": 0.40
      },
      "auto_applicable": false
    },
    {
      "type": "memory_tuning",
      "description": "Reduce memory allocation from 512MB to 256MB based on profiling",
      "impact": {
        "cost_reduction": 0.50
      },
      "auto_applicable": true
    }
  ]
}
```

---

## 4. Data Models

### 4.1 Core Data Structures

```python
# Models for FlyMind AI Microservice

from pydantic import BaseModel
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum

class ContentType(str, Enum):
    INPUT = "input"
    OUTPUT = "output"

class FunctionMetrics(BaseModel):
    function_id: str
    timestamp: datetime
    latency_ms: int
    cold_start: bool
    memory_mb: int
    error: Optional[str] = None
    target: str  # cloudflare, vercel, fly, deno

class Anomaly(BaseModel):
    id: str
    function_id: str
    type: str  # latency_spike, error_rate, memory_leak
    severity: float  # 0-1
    detected_at: datetime
    description: str
    acknowledged: bool = False

class OptimizationSuggestion(BaseModel):
    id: str
    function_id: str
    type: str  # dependency, memory, timeout, concurrency
    description: str
    impact: Dict[str, float]  # e.g., {"latency": -0.3, "cost": -0.5}
    auto_applicable: bool
    applied: bool = False
    created_at: datetime

class CacheRecommendation(BaseModel):
    function_id: str
    cacheable: bool
    similarity_threshold: float
    suggested_ttl_seconds: int
    estimated_hit_rate: float

class RoutingDecision(BaseModel):
    function_id: str
    recommended_target: str
    confidence: float
    reasoning: str
    alternatives: List[str] = []

class ChatMessage(BaseModel):
    id: str
    role: str  # user, assistant
    content: str
    timestamp: datetime
    context: Dict[str, Any] = {}

class ChatSession(BaseModel):
    id: str
    tenant_id: str
    user_id: str
    messages: List[ChatMessage]
    created_at: datetime
    updated_at: datetime
```

### 4.2 Database Schema (PostgreSQL)

```sql
-- AI-specific tables for FlyMind

-- Function execution metrics (aggregated)
CREATE TABLE IF NOT EXISTS function_metrics_ai (
    id BIGSERIAL PRIMARY KEY,
    function_id UUID NOT NULL REFERENCES functions(id),
    timestamp TIMESTAMPTZ NOT NULL,
    latency_ms INTEGER NOT NULL,
    cold_start BOOLEAN NOT NULL,
    memory_mb INTEGER NOT NULL,
    target VARCHAR(50) NOT NULL,  -- cloudflare, vercel, fly, deno
    error VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_metrics_function_time ON function_metrics_ai(function_id, timestamp DESC);

-- Anomaly detection results
CREATE TABLE IF NOT EXISTS anomalies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES functions(id),
    anomaly_type VARCHAR(50) NOT NULL,
    severity FLOAT NOT NULL CHECK (severity BETWEEN 0 AND 1),
    detected_at TIMESTAMPTZ NOT NULL,
    description TEXT NOT NULL,
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_anomalies_function ON anomalies(function_id, detected_at DESC);

-- Optimization suggestions
CREATE TABLE IF NOT EXISTS optimization_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES functions(id),
    suggestion_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    impact_latency FLOAT,
    impact_cost FLOAT,
    auto_applicable BOOLEAN DEFAULT FALSE,
    applied BOOLEAN DEFAULT FALSE,
    applied_by UUID REFERENCES users(id),
    applied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_optimizations_function ON optimization_suggestions(function_id, created_at DESC);

-- Chat sessions and history
CREATE TABLE IF NOT EXISTS ai_chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES ai_chat_sessions(id),
    role VARCHAR(20) NOT NULL,  -- user, assistant, system
    content TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chat_messages_session ON ai_chat_messages(session_id, created_at DESC);

-- Routing decisions for learning
CREATE TABLE IF NOT EXISTS routing_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES functions(id),
    user_geo VARCHAR(10),
    recommended_target VARCHAR(50) NOT NULL,
    confidence FLOAT NOT NULL,
    actual_latency_ms INTEGER,
    feedback_received BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_routing_function_time ON routing_decisions(function_id, created_at DESC);

-- Cache predictions
CREATE TABLE IF NOT EXISTS cache_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES functions(id),
    input_pattern_hash VARCHAR(64) NOT NULL,
    cacheable BOOLEAN NOT NULL,
    similarity_threshold FLOAT NOT NULL,
    estimated_hit_rate FLOAT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cache_predictions_function ON cache_predictions(function_id);
```

### 4.3 Redis Data Structures

```python
# Redis key patterns for FlyMind

# Chat sessions (TTL: 30 days)
"ai:chat:{session_id}" -> Hash of messages

# Real-time routing cache (TTL: 5 minutes)
"ai:routing:{function_id}:{geo_hash}" -> JSON of routing decision

# Prewarming state (TTL: 1 hour)
"ai:prewarm:{function_id}" -> Hash with instances, last_warmed

# Rate limiting for AI endpoints (TTL: 1 minute)
"ai:ratelimit:{tenant_id}:{endpoint}" -> Counter

# LLM response cache (TTL: configurable, default 1 hour)
"ai:llm:cache:{prompt_hash}" -> JSON of cached response

# Provider fallback state (TTL: 30 seconds)
"ai:provider:status:{provider}" -> Hash with latency, errors
```

---

## 5. Provider Abstraction Strategy

### 5.1 Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Provider Abstraction Layer                           │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│   │   Router    │    │   Fallback  │    │   Retry     │    │    Rate     │  │
│   │   (Select)  │    │   Manager   │    │   Logic     │    │   Limiter   │  │
│   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘    └──────┬──────┘  │
│          │                  │                  │                  │          │
│          └──────────────────┼──────────────────┼──────────────────┘          │
│                             │                  │                              │
│                             ▼                  ▼                              │
│   ┌───────────────────────────────────────────────────────────────────────┐  │
│   │                        Base Provider Interface                         │  │
│   │  - complete(prompt: str, **kwargs) -> str                             │  │
│   │  - stream(prompt: str, **kwargs) -> Generator[str, None, None]        │  │
│   │  - embed(texts: List[str], **kwargs) -> List[List[float]]            │  │
│   │  - get_token_count(text: str) -> int                                 │  │
│   │  - get_model_info() -> ModelInfo                                      │  │
│   └───────────────────────────────────────────────────────────────────────┘  │
│                             │                                                │
│          ┌──────────────────┼──────────────────┐                              │
│          │                  │                  │                              │
│          ▼                  ▼                  ▼                              │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                        │
│   │  OpenAI     │    │  Anthropic  │    │   Ollama    │                        │
│   │  Adapter    │    │   Adapter   │    │   Adapter   │                        │
│   └─────────────┘    └─────────────┘    └─────────────┘                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 Provider Interface Definition

```python
# provider/base.py
from abc import ABC, abstractmethod
from typing import Generator, List, Dict, Any, Optional
from pydantic import BaseModel

class ModelInfo(BaseModel):
    name: str
    max_tokens: int
    supports_streaming: bool
    supports_embeddings: bool
    embedding_dimensions: Optional[int] = None
    pricing_per_1k_input: Optional[float] = None
    pricing_per_1k_output: Optional[float] = None

class ProviderResponse(BaseModel):
    content: str
    model: str
    usage: Dict[str, int]  # input_tokens, output_tokens
    finish_reason: str

class BaseProvider(ABC):
    """Base class for all LLM providers"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self._client = self._initialize_client()
    
    @abstractmethod
    def _initialize_client(self) -> Any:
        """Initialize the provider-specific client"""
        pass
    
    @abstractmethod
    async def complete(
        self, 
        prompt: str, 
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        **kwargs
    ) -> ProviderResponse:
        """Synchronous completion"""
        pass
    
    @abstractmethod
    async def stream(
        self,
        prompt: str,
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        **kwargs
    ) -> Generator[str, None, None]:
        """Streaming completion"""
        pass
    
    @abstractmethod
    async def embed(
        self,
        texts: List[str],
        model: Optional[str] = None,
        **kwargs
    ) -> List[List[float]]:
        """Generate embeddings"""
        pass
    
    @abstractmethod
    def get_token_count(self, text: str) -> int:
        """Count tokens in text"""
        pass
    
    @abstractmethod
    def get_model_info(self, model: Optional[str] = None) -> ModelInfo:
        """Get model information"""
        pass
    
    @abstractmethod
    def validate_config(self) -> bool:
        """Validate provider configuration"""
        pass
```

### 5.3 Provider Implementations

```python
# provider/openai.py
import openai
from typing import Generator, List, Dict, Any, Optional
from .base import BaseProvider, ProviderResponse, ModelInfo

class OpenAIProvider(BaseProvider):
    """OpenAI API provider (GPT-4, GPT-3.5, etc.)"""
    
    DEFAULT_MODELS = {
        "gpt-4o": ModelInfo(
            name="gpt-4o",
            max_tokens=128000,
            supports_streaming=True,
            supports_embeddings=True,
            embedding_dimensions=3072,
            pricing_per_1k_input=0.005,
            pricing_per_1k_output=0.015
        ),
        "gpt-4o-mini": ModelInfo(
            name="gpt-4o-mini",
            max_tokens=128000,
            supports_streaming=True,
            supports_embeddings=True,
            embedding_dimensions=1536,
            pricing_per_1k_input=0.00015,
            pricing_per_1k_output=0.0006
        )
    }
    
    def _initialize_client(self) -> openai.AsyncOpenAI:
        api_key = self.config.get("api_key")
        base_url = self.config.get("base_url")  # For OpenRouter compatibility
        return openai.AsyncOpenAI(api_key=api_key, base_url=base_url)
    
    async def complete(
        self,
        prompt: str,
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        **kwargs
    ) -> ProviderResponse:
        response = await self._client.chat.completions.create(
            model=model,
            messages=[{"role": "user", "content": prompt}],
            temperature=temperature,
            max_tokens=max_tokens,
            **kwargs
        )
        
        return ProviderResponse(
            content=response.choices[0].message.content,
            model=response.model,
            usage={
                "input_tokens": response.usage.prompt_tokens,
                "output_tokens": response.usage.completion_tokens
            },
            finish_reason=response.choices[0].finish_reason
        )
    
    async def stream(self, prompt: str, **kwargs) -> Generator[str, None, None]:
        # Implementation for streaming
        pass
    
    async def embed(self, texts: List[str], model: str = "text-embedding-3-small") -> List[List[float]]:
        response = await self._client.embeddings.create(
            model=model,
            input=texts
        )
        return [embedding.embedding for embedding in response.data]
    
    def get_token_count(self, text: str) -> int:
        # Use tiktoken for accurate counting
        import tiktoken
        encoding = tiktoken.encoding_for_model("gpt-4o")
        return len(encoding.encode(text))
    
    def get_model_info(self, model: Optional[str] = None) -> ModelInfo:
        return self.DEFAULT_MODELS.get(model or "gpt-4o-mini")
    
    def validate_config(self) -> bool:
        return bool(self.config.get("api_key"))
```

### 5.4 Provider Router

```python
# provider/router.py
from typing import List, Dict, Any, Optional
import asyncio
from .base import BaseProvider, ProviderResponse
from .openai import OpenAIProvider
from .anthropic import AnthropicProvider
from .ollama import OllamaProvider

class ProviderRouter:
    """Routes requests to appropriate provider with fallback"""
    
    def __init__(self, config: Dict[str, Any]):
        self.providers: Dict[str, BaseProvider] = {}
        self.default_provider = config.get("default", "openai")
        self._initialize_providers(config)
    
    def _initialize_providers(self, config: Dict[str, Any]):
        """Initialize all configured providers"""
        if "openai" in config.get("enabled", []):
            self.providers["openai"] = OpenAIProvider(config.get("openai", {}))
        
        if "anthropic" in config.get("enabled", []):
            self.providers["anthropic"] = AnthropicProvider(config.get("anthropic", {}))
        
        if "ollama" in config.get("enabled", []):
            self.providers["ollama"] = OllamaProvider(config.get("ollama", {}))
    
    async def complete(
        self,
        prompt: str,
        provider: Optional[str] = None,
        fallback: bool = True,
        **kwargs
    ) -> ProviderResponse:
        """Complete with optional fallback to other providers"""
        primary = provider or self.default_provider
        
        if primary in self.providers:
            try:
                return await self.providers[primary].complete(prompt, **kwargs)
            except Exception as e:
                if not fallback:
                    raise
                # Log failure and try fallback
                return await self._try_fallback(primary, prompt, **kwargs)
        
        raise ValueError(f"Provider {primary} not configured")
    
    async def _try_fallback(self, failed_provider: str, prompt: str, **kwargs) -> ProviderResponse:
        """Try other providers in order"""
        for name, provider in self.providers.items():
            if name != failed_provider:
                try:
                    return await provider.complete(prompt, **kwargs)
                except Exception:
                    continue
        
        raise RuntimeError("All providers failed")
    
    async def embed(
        self,
        texts: List[str],
        provider: Optional[str] = None
    ) -> List[List[float]]:
        """Generate embeddings via configured provider"""
        provider = provider or self.default_provider
        
        if provider not in self.providers:
            raise ValueError(f"Provider {provider} not configured")
        
        return await self.providers[provider].embed(texts)
```

### 5.5 Configuration

```yaml
# config/ai.yaml
providers:
  enabled:
    - openai
    - anthropic
    - ollama
  
  default: openai
  
  openai:
    api_key: ${OPENAI_API_KEY}
    base_url: https://openrouter.ai/v1  # Optional: for OpenRouter
    models:
      - gpt-4o
      - gpt-4o-mini
      - gpt-4-turbo
    timeout: 30
    max_retries: 3
  
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    models:
      - claude-3-5-sonnet-20241022
      - claude-3-opus-20240229
    timeout: 30
    max_retries: 3
  
  ollama:
    base_url: http://localhost:11434
    models:
      - llama3.1:70b
      - mistral:7b
    timeout: 60

routing:
  strategy: latency_based  # latency_based, cost_based, quality_based
  fallback_enabled: true
  health_check_interval: 30

cache:
  enabled: true
  redis_db: 1
  ttl_seconds: 3600
  max_size_mb: 512
```

---

## 6. Integration Points with FunctionFly

### 6.1 Communication with Go Orchestrator

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        FlyMind ↔ Orchestrator Integration                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────┐                          ┌───────────────────────┐     │
│  │   FlyMind        │                          │   Go Orchestrator     │     │
│  │   (Python)       │                          │   (Port 8080)         │     │
│  └────────┬─────────┘                          └───────────┬───────────┘     │
│           │                                               │                    │
│           │  1. HTTP/REST (orchestrator APIs)            │                    │
│           │  ─────────────────────────────────────────────│                    │
│           │  • GET /api/v1/functions/{id}                │                    │
│           │  • POST /api/v1/functions/{id}/deploy        │                    │
│           │  • GET /api/v1/status/executions            │                    │
│           │  • POST /api/v1/webhooks/notify             │                    │
│           │                                              │                    │
│           │  2. gRPC (high-perf routing)                 │                    │
│           │  ─────────────────────────────────────────────│                    │
│           │  • GetRoutingRecommendation                 │                    │
│           │  • ShouldPrewarm                             │                    │
│           │  • ModerateContent                          │                    │
│           │                                              │                    │
│           │  3. Pub/Sub (event-driven)                   │                    │
│           │  ─────────────────────────────────────────────│                    │
│           │  • function.execution.completed              │                    │
│           │  • function.deployed                         │                    │
│           │  • function.error                             │                    │
│           │                                              │                    │
│           │  4. Shared Database                           │                    │
│           │  ─────────────────────────────────────────────│                    │
│           │  • function_metrics_ai                       │                    │
│           │  • anomalies                                 │                    │
│           │  • optimization_suggestions                  │                    │
│           │                                              │                    │
└───────────┴──────────────────────────────────────────────┴────────────────────┘
```

### 6.2 Authentication

FlyMind authenticates with the orchestrator using:

1. **API Key**: Shared secret in `X-API-Key` header
2. **JWT Token**: For user-context operations
3. **Service Account**: For system-level operations

```python
# Integration with orchestrator auth

import httpx
from typing import Optional

class OrchestratorClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.client = httpx.AsyncClient(
            headers={"X-API-Key": api_key},
            timeout=30.0
        )
    
    async def get_function(self, function_id: str) -> Dict:
        response = await self.client.get(
            f"{self.base_url}/api/v1/functions/{function_id}"
        )
        response.raise_for_status()
        return response.json()
    
    async def get_execution_metrics(
        self, 
        function_id: str, 
        since: datetime,
        until: Optional[datetime] = None
    ) -> List[Dict]:
        params = {"since": since.isoformat()}
        if until:
            params["until"] = until.isoformat()
        
        response = await self.client.get(
            f"{self.base_url}/api/v1/status/executions/{function_id}",
            params=params
        )
        response.raise_for_status()
        return response.json()["executions"]
```

### 6.3 Event Subscription

```python
# Subscribe to orchestrator events via Redis pub/sub

import asyncio
import json
import redis.asyncio as redis

class EventSubscriber:
    def __init__(self, redis_url: str):
        self.redis = redis.from_url(redis_url)
        self.pubsub = self.redis.pubsub()
    
    async def subscribe_to_events(self):
        await self.pubsub.subscribe(
            "function.execution.completed",
            "function.deployed",
            "function.error"
        )
        
        async for message in self.pubsub.listen():
            if message["type"] == "message":
                event = json.loads(message["data"])
                await self.handle_event(message["channel"], event)
    
    async def handle_event(self, channel: str, event: Dict):
        if channel == "function.execution.completed":
            await self.process_execution(event)
        elif channel == "function.deployed":
            await self.process_deployment(event)
        elif channel == "function.error":
            await self.process_error(event)
    
    async def process_execution(self, event: Dict):
        # Update metrics, check for anomalies
        await self.store_metrics(event)
        await self.check_anomaly(event)
```

---

## 7. Deployment Strategy

### 7.1 Docker Configuration

```dockerfile
# Dockerfile.flymind
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Create non-root user
RUN useradd -m -u 1000 flymind
USER flymind

EXPOSE 8081

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8081"]
```

```yaml
# docker-compose.ai.yml
services:
  flymind:
    build:
      context: ./ai-service
      dockerfile: Dockerfile.flymind
    ports:
      - "8081:8081"
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/functionfly
      - REDIS_URL=redis://redis:6379
      - ORCHESTRATOR_URL=http://orchestrator:8080
      - ORCHESTRATOR_API_KEY=${ORCHESTRATOR_API_KEY}
    depends_on:
      - postgres
      - redis
      - orchestrator
    volumes:
      - ./ai-service/config:/app/config:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/api/v1/ai/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  flymind-worker:
    build:
      context: ./ai-service
      dockerfile: Dockerfile.flymind
    command: ["python", "-m", "worker"]
    environment:
      - DATABASE_URL=postgresql://user:pass@postgres:5432/functionfly
      - REDIS_URL=redis://redis:6379
      - WORKER_PROCESSES=4
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
```

### 7.2 Kubernetes Deployment

```yaml
# k8s/flymind-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flymind
  namespace: functionfly
spec:
  replicas: 3
  selector:
    matchLabels:
      app: flymind
  template:
    metadata:
      labels:
        app: flymind
    spec:
      containers:
        - name: flymind
          image: functionfly/flymind:latest
          ports:
            - containerPort: 8081
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: functionfly-secrets
                  key: database-url
            - name: REDIS_URL
              valueFrom:
                secretKeyRef:
                  name: functionfly-secrets
                  key: redis-url
            - name: ORCHESTRATOR_API_KEY
              valueFrom:
                secretKeyRef:
                  name: functionfly-secrets
                  key: orchestrator-api-key
          resources:
            requests:
              memory: "512Mi"
              cpu: "250m"
            limits:
              memory: "2Gi"
              cpu: "1000m"
          livenessProbe:
            httpGet:
              path: /api/v1/ai/health
              port: 8081
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /api/v1/ai/health
              port: 8081
            initialDelaySeconds: 10
            periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: flymind
  namespace: functionfly
spec:
  selector:
    app: flymind
  ports:
    - port: 8081
      targetPort: 8081
  type: ClusterIP
```

### 7.3 Scaling Considerations

| Component | Scaling Strategy | Notes |
|-----------|------------------|-------|
| API Pods | Horizontal (HPA) | CPU > 70%, replicas 3-20 |
| ML Workers | Horizontal (HPA) | Queue depth based |
| Vector DB | Vertical + Read Replicas | pgvector with streaming replica |
| Redis | Cluster mode | For distributed caching |
| LLM Inference | Provider abstraction | Rate limits apply |

### 7.4 Resource Requirements

| Service | CPU | Memory | Storage |
|---------|-----|--------|---------|
| FlyMind API | 500m | 512Mi | - |
| FlyMind Worker | 1000m | 1Gi | - |
| PostgreSQL (AI tables) | 1000m | 2Gi | 50Gi SSD |
| Redis | 500m | 1Gi | 10Gi |

---

## 8. Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)

**Goal**: Basic AI service with chat interface and semantic search

| Week | Deliverables |
|------|-------------|
| 1 | Project setup, Docker, basic FastAPI structure |
| 2 | Provider abstraction layer (OpenAI, Ollama) |
| 3 | Chat interface endpoint with context |
| 4 | Semantic search integration with existing pgvector |

**Milestone**: Developers can chat with their functions and search semantically

### Phase 2: Intelligence (Weeks 5-8)

**Goal**: Core intelligence features

| Week | Deliverables |
|------|-------------|
| 5 | Route optimization - collect training data |
| 6 | Route optimization - ML model training and API |
| 7 | Anomaly detection - metric collection and detection |
| 8 | Prewarming engine - forecasting and triggering |

**Milestone**: Platform demonstrates intelligent routing and proactive prewarming

### Phase 3: Optimization (Weeks 9-12)

**Goal**: Developer productivity features

| Week | Deliverables |
|------|-------------|
| 9 | Code optimization - static analysis |
| 10 | Cache prediction - ML model |
| 11 | Debug assistant - LLM-powered |
| 12 | Content moderation - classification |

**Milestone**: Comprehensive AI-assisted developer experience

### Phase 4: Production Hardening (Weeks 13-16)

**Goal**: Production readiness

| Week | Deliverables |
|------|-------------|
| 13 | gRPC integration with orchestrator |
| 14 | Kubernetes deployment manifests |
| 15 | Observability (metrics, logging, tracing) |
| 16 | Load testing, security audit |

**Milestone**: Production-ready AI service

---

## 9. Security Considerations

### 9.1 API Security

- **Authentication**: API key + JWT for user context
- **Rate Limiting**: Per-tenant, per-endpoint limits via Redis
- **Input Validation**: Pydantic models for all inputs
- **Output Filtering**: Sanitize LLM outputs before returning

### 9.2 Data Privacy

- **No training data retention**: LLM interactions not stored long-term
- **PII handling**: Anonymize user identifiers in metrics
- **Encryption**: TLS in transit, encrypted volumes at rest

### 9.3 LLM Safety

- **Prompt injection protection**: Sanitize user inputs
- **Output validation**: Validate LLM outputs before use
- **Content moderation**: Scan inputs/outputs for policy violations

---

## 10. Success Metrics

### 10.1 Technical Metrics

| Metric | Target |
|--------|--------|
| API Latency (p95) | < 200ms |
| Chat response time | < 5s |
| Semantic search latency | < 100ms |
| Anomaly detection latency | < 1s |
| Route prediction latency | < 50ms |

### 10.2 Business Metrics

| Metric | Target |
|--------|--------|
| Latency reduction | 20-40% |
| Cold start reduction | 50% |
| Cache hit rate | 60%+ |
| Developer engagement | 30% using AI features |
| Cost savings | 30% on function execution |

---

## Appendix: Environment Variables

```bash
# FlyMind Configuration
FLYMIND_PORT=8081
FLYMIND_ENV=development  # development, staging, production

# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/functionfly

# Redis
REDIS_URL=redis://localhost:6379

# Orchestrator
ORCHESTRATOR_URL=http://localhost:8080
ORCHESTRATOR_API_KEY=your-api-key

# AI Providers
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
OLLAMA_BASE_URL=http://localhost:11434

# Feature Flags
ENABLE_ROUTING_OPTIMIZATION=true
ENABLE_PREWARMING=true
ENABLE_ANOMALY_DETECTION=true
ENABLE_CHAT_INTERFACE=true
ENABLE_CODE_OPTIMIZATION=true
ENABLE_CONTENT_MODERATION=true

# ML Models
ROUTING_MODEL_PATH=/app/models/routing
CACHE_MODEL_PATH=/app/models/cache
ANOMALY_MODEL_PATH=/app/models/anomaly
```

---

*Document Version: 1.0*  
*Created: 2026-03-12*  
*Architecture: FlyMind - FunctionFly AI Microservice*
