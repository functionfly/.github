"""Pydantic models for FlyMind AI Service API.

This module contains all schemas for Phase 1 (Foundation) and Phase 2 (Intelligence).
"""

from typing import Optional, Any, Dict, List
from datetime import datetime, timedelta
from enum import Enum
from pydantic import BaseModel, Field
import uuid


class ProviderType(str, Enum):
    """Supported LLM providers."""
    OPENAI = "openai"
    ANTHROPIC = "anthropic"
    OLLAMA = "ollama"
    OPENROUTER = "openrouter"


class MessageRole(str, Enum):
    """Chat message roles."""
    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"


class ChatMessage(BaseModel):
    """A single chat message."""
    role: MessageRole
    content: str


class CompletionRequest(BaseModel):
    """Request for LLM completion."""
    provider: Optional[ProviderType] = None
    model: Optional[str] = None
    messages: list[ChatMessage]
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: Optional[int] = Field(default=None, ge=1)
    stream: bool = False
    top_p: Optional[float] = Field(default=None, ge=0.0, le=1.0)
    stop: Optional[list[str]] = None


class CompletionResponse(BaseModel):
    """Response from LLM completion."""
    content: str
    provider: ProviderType
    model: str
    usage: dict[str, int] = Field(
        default_factory=lambda: {
            "prompt_tokens": 0,
            "completion_tokens": 0,
            "total_tokens": 0
        }
    )
    finish_reason: Optional[str] = None
    latency_ms: float = 0.0


class EmbeddingRequest(BaseModel):
    """Request for embeddings generation."""
    provider: Optional[ProviderType] = None
    model: Optional[str] = None
    text: str = Field(..., min_length=1, max_length=8192)
    dimensions: Optional[int] = Field(default=None, ge=1, le=4096)


class EmbeddingResponse(BaseModel):
    """Response with embeddings."""
    embedding: list[float]
    provider: ProviderType
    model: str
    dimensions: int
    usage: dict[str, int] = Field(
        default_factory=lambda: {"tokens": 0}
    )
    latency_ms: float = 0.0


class ProviderInfo(BaseModel):
    """Information about a provider."""
    name: str
    display_name: str
    available: bool
    models: list[str]
    rate_limit: int  # requests per minute
    embedding_dimensions: int
    supports_streaming: bool
    supports_embeddings: bool


class ProviderStatusResponse(BaseModel):
    """Status of all providers."""
    providers: list[ProviderInfo]
    default_provider: ProviderType
    default_embedding_provider: ProviderType


class HealthResponse(BaseModel):
    """Health check response."""
    status: str
    service: str
    version: str
    providers_available: list[str]
    redis_connected: bool
    database_connected: bool


class ErrorResponse(BaseModel):
    """Error response."""
    error: str
    detail: Optional[str] = None
    provider: Optional[str] = None
    retry_after: Optional[int] = None  # seconds


class CostTracking(BaseModel):
    """Cost tracking information."""
    provider: str
    model: str
    input_tokens: int
    output_tokens: int
    total_tokens: int
    estimated_cost: float  # in USD


# =============================================================================
# Phase 2: Intelligent Request Routing Models
# =============================================================================

class EdgeProvider(str, Enum):
    """Supported edge providers."""
    CLOUDFLARE = "cloudflare"
    VERCEL = "vercel"
    FLY = "fly"
    DENO = "deno"
    FUNCTIONFLY = "functionfly"  # FunctionFly edge


class EdgeLocation(BaseModel):
    """Geographic location of an edge."""
    region: str
    country: str
    latitude: float
    longitude: float


class EdgeStatus(BaseModel):
    """Status information for an edge provider."""
    provider: EdgeProvider
    location: EdgeLocation
    available: bool
    current_load_percent: float = 0.0
    avg_latency_ms: float = 0.0
    last_check: datetime


class RoutingDecisionRequest(BaseModel):
    """Request for routing decision."""
    function_id: str
    user_geography: Optional[str] = None  # e.g., "us-east", "eu-west"
    user_country: Optional[str] = None
    user_latitude: Optional[float] = None
    user_longitude: Optional[float] = None
    metadata: Dict[str, str] = Field(default_factory=dict)


class RoutingDecision(BaseModel):
    """Routing decision response."""
    function_id: str
    recommended_edge: EdgeProvider
    confidence: float = Field(ge=0.0, le=1.0)
    reasoning: str
    alternatives: List[EdgeProvider] = Field(default_factory=list)
    latency_estimate_ms: float
    decided_at: datetime = Field(default_factory=datetime.utcnow)


class EdgeListResponse(BaseModel):
    """Response containing all available edges with status."""
    edges: List[EdgeStatus]
    total_count: int
    last_updated: datetime


class LatencySample(BaseModel):
    """A latency sample from an edge execution."""
    function_id: str
    edge: EdgeProvider
    latency_ms: float
    timestamp: datetime
    success: bool = True


# =============================================================================
# Phase 2: Predictive Cold Start Prewarming Models
# =============================================================================

class PredictionRequest(BaseModel):
    """Request for prewarming predictions."""
    function_id: str
    prediction_window_minutes: int = Field(default=10, ge=1, le=60)


class Prediction(BaseModel):
    """Prediction for function demand."""
    function_id: str
    predicted_requests: int
    confidence: float = Field(ge=0.0, le=1.0)
    window_start: datetime
    window_end: datetime
    trend: str  # "increasing", "decreasing", "stable"
    generated_at: datetime = Field(default_factory=datetime.utcnow)


class PrewarmTriggerRequest(BaseModel):
    """Request to trigger prewarming."""
    function_id: str
    instances: int = Field(default=1, ge=1, le=10)
    edge: Optional[EdgeProvider] = None


class PrewarmStatus(BaseModel):
    """Status of prewarming operation."""
    function_id: str
    instances_requested: int
    instances_warmed: int
    status: str  # "pending", "warming", "complete", "failed"
    triggered_at: datetime
    completed_at: Optional[datetime] = None


class HistoricalRequestData(BaseModel):
    """Historical request data point for forecasting."""
    function_id: str
    request_count: int
    timestamp: datetime


# =============================================================================
# Phase 2: Anomaly Detection Models
# =============================================================================

class AnomalyType(str, Enum):
    """Types of anomalies that can be detected."""
    LATENCY_SPIKE = "latency_spike"
    ERROR_RATE_INCREASE = "error_rate_increase"
    COLD_START_SPIKE = "cold_start_spike"
    MEMORY_LEAK = "memory_leak"
    UNUSUAL_PATTERN = "unusual_pattern"


class AnomalySeverity(str, Enum):
    """Severity levels for anomalies."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class Anomaly(BaseModel):
    """Detected anomaly."""
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    function_id: str
    type: AnomalyType
    severity: AnomalySeverity
    detected_at: datetime = Field(default_factory=datetime.utcnow)
    description: str
    metric_name: str  # e.g., "latency_ms", "error_rate"
    metric_value: float
    threshold: float
    z_score: Optional[float] = None
    acknowledged: bool = False
    acknowledged_by: Optional[str] = None
    acknowledged_at: Optional[datetime] = None


class AnomalyListResponse(BaseModel):
    """Response containing list of anomalies."""
    anomalies: List[Anomaly]
    total_count: int
    page: int = 1
    page_size: int = 20


class AnomalyAcknowledgeRequest(BaseModel):
    """Request to acknowledge an anomaly."""
    anomaly_id: str
    acknowledged_by: str = "system"


class AnomalyThresholds(BaseModel):
    """Configurable thresholds for anomaly detection."""
    latency_z_score_threshold: float = 3.0  # Standard deviations
    error_rate_threshold: float = 0.01  # 1%
    cold_start_rate_threshold: float = 0.10  # 10%
    sliding_window_minutes: int = 5
    check_interval_seconds: int = 30


class ExecutionMetrics(BaseModel):
    """Execution metrics for a function."""
    function_id: str
    timestamp: datetime
    latency_ms: float
    cold_start: bool
    memory_mb: int
    error: Optional[str] = None
    target: EdgeProvider


# =============================================================================
# Phase 3: Developer Experience Layer - Chat Service Models
# =============================================================================

class ChatIntent(str, Enum):
    """Chat intent types."""
    EXPLAIN = "explain_intent"
    QUERY = "query_intent"
    DEBUG = "debugging_intent"
    OPTIMIZE = "optimization_intent"
    HELP = "help_intent"
    UNKNOWN = "unknown_intent"


class ChatSessionResponse(BaseModel):
    """Chat session response."""
    session_id: str
    user_id: str
    created_at: datetime
    message_count: int = 0


class ChatMessageResponse(BaseModel):
    """Chat message response."""
    session_id: str
    message: str
    intent: ChatIntent
    confidence: float


class ChatHistoryResponse(BaseModel):
    """Chat history response."""
    session_id: str
    messages: List[Dict[str, Any]]
    created_at: datetime


# =============================================================================
# Phase 3: Developer Experience Layer - Search Service Models
# =============================================================================

class SearchQuery(BaseModel):
    """Search query request."""
    query: str = Field(..., min_length=1, max_length=500)
    limit: int = Field(default=20, ge=1, le=50)
    filters: Optional[Dict[str, Any]] = None


class SearchResult(BaseModel):
    """Single search result."""
    function_id: str
    function_name: str
    description: Optional[str] = None
    runtime: Optional[str] = None
    tags: List[str] = Field(default_factory=list)
    score: float
    rank: int


class SearchResponse(BaseModel):
    """Search response."""
    query: str
    results: List[SearchResult]
    total_count: int
    query_type: str = "semantic"


# =============================================================================
# Phase 3: Developer Experience Layer - Debugging Service Models
# =============================================================================

class DebugAnalyzeRequest(BaseModel):
    """Debug analysis request."""
    function_id: str
    error_message: str
    stack_trace: Optional[str] = None


class DebugAnalysis(BaseModel):
    """Debug analysis result."""
    function_id: str
    error_message: str
    error_category: str
    category_confidence: float
    root_cause: str
    confidence: float
    details: Dict[str, Any]
    suggestions: List[str]


class FixSuggestion(BaseModel):
    """Fix suggestion."""
    id: str
    title: str
    description: str
    code_example: Optional[str] = None
    effort: str  # low, medium, high
    impact: str  # low, medium, high


class DebugSuggestRequest(BaseModel):
    """Debug suggestion request."""
    analysis: DebugAnalysis


class DebugSuggestResponse(BaseModel):
    """Debug suggestion response."""
    analysis: DebugAnalysis
    suggestions: List[FixSuggestion]
    documentation_links: List[Dict[str, str]]


# =============================================================================
# Phase 3: Developer Experience Layer - Optimization Service Models
# =============================================================================

class OptimizationResult(BaseModel):
    """Optimization result."""
    function_id: str
    function_name: str
    runtime: str
    memory_mb: int
    analyzed_at: datetime
    patterns: List[Dict[str, Any]]
    issues: List[Dict[str, Any]]


class Recommendation(BaseModel):
    """Optimization recommendation."""
    id: str
    type: str
    title: str
    description: str
    category: str
    priority: str  # critical, high, medium, low
    action: str
    current_value: float
    target_value: float
    estimated_savings_monthly: float = 0.0
    savings_currency: str = "USD"


class OptimizationRecommendationsResponse(BaseModel):
    """Optimization recommendations response."""
    function_id: str
    recommendations: List[Recommendation]
    total_count: int


class ApplyRecommendationRequest(BaseModel):
    """Apply recommendation request."""
    recommendation_id: str


class ApplyRecommendationResponse(BaseModel):
    """Apply recommendation response."""
    success: bool
    function_id: str
    recommendation_id: str
    message: str


# =============================================================================
# Phase 2: Orchestrator Integration Models
# =============================================================================

class OrchestratorConfig(BaseModel):
    """Configuration for connecting to the Go orchestrator."""
    orchestrator_url: str = "http://localhost:8080"
    api_key: Optional[str] = None
    timeout_seconds: int = 30


class FunctionInfo(BaseModel):
    """Function information from orchestrator."""
    id: str
    name: str
    tenant_id: str
    runtime: str
    memory_mb: int
    timeout_seconds: int
    created_at: datetime
    updated_at: datetime


class ExecutionResult(BaseModel):
    """Execution result from orchestrator."""
    execution_id: str
    function_id: str
    status: str
    latency_ms: float
    output: Optional[str] = None
    error: Optional[str] = None
    started_at: datetime
    completed_at: datetime
