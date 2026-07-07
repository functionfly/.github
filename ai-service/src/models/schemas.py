"""Pydantic models for FlyMind AI Service API.

This module contains all schemas for Phase 1 (Foundation) and Phase 2 (Intelligence).
"""

import uuid
from datetime import datetime
from enum import StrEnum
from typing import Any, Literal

from pydantic import BaseModel, Field


class InteractionType(StrEnum):
    """Valid interaction types for recommendation system."""
    
    VIEW = "view"
    INSTALL = "install"
    EXECUTE = "execute"
    RATE = "rate"
    SEARCH_CLICK = "search_click"
    SEARCH_IMPRESSION = "search_impression"


class ProviderType(StrEnum):
    """Supported LLM providers."""

    OPENAI = "openai"
    ANTHROPIC = "anthropic"
    OLLAMA = "ollama"
    OPENROUTER = "openrouter"
    FIREWORKS = "fireworks"
    GROQ = "groq"
    DEEPINFRA = "deepinfra"
    TOGETHER = "together"
    MIMO = "mimo"
    MINIMAX = "minimax"
    STEPFUN = "stepfun"


class TrafficType(StrEnum):
    """Traffic types for provider routing."""

    REALTIME = "realtime"  # Low-latency agent function calls
    STRUCTURED = "structured"  # Structured output / tool use / JSON mode
    BACKGROUND = "background"  # Embeddings, batch processing
    FUNCTION_CALLING = "function_calling"  # Function calling optimized
    GENERAL = "general"  # Default routing


class MessageRole(StrEnum):
    """Chat message roles."""

    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"


class ChatMessage(BaseModel):
    """A single chat message."""

    role: MessageRole
    content: str


class ThinkingConfig(BaseModel):
    """Configuration for provider-native thinking/reasoning."""

    mode: str = Field(default="off", description="Thinking mode: off, auto, always")
    budget_tokens: int = Field(default=10000, ge=1000, le=100000, description="Max tokens for thinking")


class CompletionRequest(BaseModel):
    """Request for LLM completion."""

    provider: ProviderType | None = None
    model: str | None = None
    messages: list[ChatMessage]
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: int | None = Field(default=None, ge=1)
    stream: bool = False
    top_p: float | None = Field(default=None, ge=0.0, le=1.0)
    stop: list[str] | None = None
    thinking: ThinkingConfig | None = None


class CompletionResponse(BaseModel):
    """Response from LLM completion."""

    content: str
    provider: ProviderType
    model: str
    usage: dict[str, int] = Field(
        default_factory=lambda: {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
    )
    finish_reason: str | None = None
    latency_ms: float = 0.0
    thinking_content: str | None = None
    thinking_tokens: int = 0


class EmbeddingRequest(BaseModel):
    """Request for embeddings generation."""

    provider: ProviderType | None = None
    model: str | None = None
    text: str = Field(..., min_length=1, max_length=8192)
    dimensions: int | None = Field(default=None, ge=1, le=4096)


class EmbeddingResponse(BaseModel):
    """Response with embeddings."""

    embedding: list[float]
    provider: ProviderType
    model: str
    dimensions: int
    usage: dict[str, int] = Field(default_factory=lambda: {"tokens": 0})
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
    detail: str | None = None
    provider: str | None = None
    retry_after: int | None = None  # seconds


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


class EdgeProvider(StrEnum):
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
    user_geography: str | None = None  # e.g., "us-east", "eu-west"
    user_country: str | None = None
    user_latitude: float | None = None
    user_longitude: float | None = None
    metadata: dict[str, str] = Field(default_factory=dict)


class RoutingDecision(BaseModel):
    """Routing decision response."""

    function_id: str
    recommended_edge: EdgeProvider
    confidence: float = Field(ge=0.0, le=1.0)
    reasoning: str
    alternatives: list[EdgeProvider] = Field(default_factory=list)
    latency_estimate_ms: float
    decided_at: datetime = Field(default_factory=datetime.utcnow)


class EdgeListResponse(BaseModel):
    """Response containing all available edges with status."""

    edges: list[EdgeStatus]
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
    edge: EdgeProvider | None = None


class PrewarmStatus(BaseModel):
    """Status of prewarming operation."""

    function_id: str
    instances_requested: int
    instances_warmed: int
    status: str  # "pending", "warming", "complete", "failed"
    triggered_at: datetime
    completed_at: datetime | None = None


class HistoricalRequestData(BaseModel):
    """Historical request data point for forecasting."""

    function_id: str
    request_count: int
    timestamp: datetime


# =============================================================================
# Phase 2: Anomaly Detection Models
# =============================================================================


class AnomalyType(StrEnum):
    """Types of anomalies that can be detected."""

    LATENCY_SPIKE = "latency_spike"
    ERROR_RATE_INCREASE = "error_rate_increase"
    COLD_START_SPIKE = "cold_start_spike"
    MEMORY_LEAK = "memory_leak"
    UNUSUAL_PATTERN = "unusual_pattern"


class AnomalySeverity(StrEnum):
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
    z_score: float | None = None
    acknowledged: bool = False
    acknowledged_by: str | None = None
    acknowledged_at: datetime | None = None


class AnomalyListResponse(BaseModel):
    """Response containing list of anomalies."""

    anomalies: list[Anomaly]
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
    error: str | None = None
    target: EdgeProvider


# =============================================================================
# Phase 3: Developer Experience Layer - Chat Service Models
# =============================================================================


class ChatIntent(StrEnum):
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
    thinking_content: str | None = None
    thinking_tokens: int = 0


class ChatHistoryResponse(BaseModel):
    """Chat history response."""

    session_id: str
    messages: list[dict[str, Any]]
    created_at: datetime


# =============================================================================
# Phase 3: Developer Experience Layer - Search Service Models
# =============================================================================


class SearchQuery(BaseModel):
    """Search query request."""

    query: str = Field(..., min_length=1, max_length=500)
    limit: int = Field(default=20, ge=1, le=50)
    filters: dict[str, Any] | None = None
    use_triple: bool = Field(default=True, description="Enable triple-vector search")
    weights: dict[str, float] | None = Field(
        default=None, description="Custom weights for contract/semantic/code triple scoring"
    )


class SearchResult(BaseModel):
    """Single search result."""

    function_id: str
    function_name: str
    description: str | None = None
    runtime: str | None = None
    tags: list[str] = Field(default_factory=list)
    score: float
    rank: int


class SearchResponse(BaseModel):
    """Search response."""

    query: str
    results: list[SearchResult]
    total_count: int
    query_type: str = "semantic"


# =============================================================================
# Phase 3: Developer Experience Layer - Debugging Service Models
# =============================================================================


class DebugAnalyzeRequest(BaseModel):
    """Debug analysis request."""

    function_id: str
    error_message: str
    stack_trace: str | None = None


class DebugAnalysis(BaseModel):
    """Debug analysis result."""

    function_id: str
    error_message: str
    error_category: str
    category_confidence: float
    root_cause: str
    confidence: float
    details: dict[str, Any]
    suggestions: list[str]


class FixSuggestion(BaseModel):
    """Fix suggestion."""

    id: str
    title: str
    description: str
    code_example: str | None = None
    effort: str  # low, medium, high
    impact: str  # low, medium, high


class DebugSuggestRequest(BaseModel):
    """Debug suggestion request."""

    analysis: DebugAnalysis


class DebugSuggestResponse(BaseModel):
    """Debug suggestion response."""

    analysis: DebugAnalysis
    suggestions: list[FixSuggestion]
    documentation_links: list[dict[str, str]]


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
    patterns: list[dict[str, Any]]
    issues: list[dict[str, Any]]


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
    recommendations: list[Recommendation]
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
    api_key: str | None = None
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
    output: str | None = None
    error: str | None = None
    started_at: datetime
    completed_at: datetime


# =============================================================================
# ML Intelligence Layer Schemas
# =============================================================================


class CostAnomalyCheckRequest(BaseModel):
    """Request for cost anomaly check."""

    function_id: str = Field(..., max_length=64)
    cost_cents: float = Field(ge=0, le=1e6)
    duration_ms: float = Field(ge=0, le=1e9)
    memory_mb: float = Field(ge=0, le=1e6)
    region: str = Field(max_length=32, default="unknown")


class CostAnomalyCheckResponse(BaseModel):
    """Response from cost anomaly check."""

    is_anomaly: bool
    z_score: float | None = None
    severity: str | None = None
    message: str | None = None


class PrewarmRecordRequest(BaseModel):
    """Request to record a prewarm event."""

    function_id: str = Field(..., max_length=64)
    count: int = Field(default=1, ge=1, le=10000)


class RoutingOutcomeRequest(BaseModel):
    """Request to record routing outcome for Thompson Sampling."""

    edge: str = Field(..., max_length=128)
    function_id: str = Field(..., max_length=64)
    latency_ms: float = Field(ge=0, le=60000)
    success: bool = Field(default=True)
    cost_cents: float = Field(ge=0, le=100)


class RecommendationInteractionRequest(BaseModel):
    """Request to record user-function interaction."""

    user_id: str = Field(..., max_length=64)
    function_id: str = Field(..., max_length=64)
    interaction_type: InteractionType = Field(default=InteractionType.VIEW)
    context: dict[str, Any] | None = None


class RecommendationResponse(BaseModel):
    """Response with recommendations."""

    user_id: str
    recommendations: list[str]
    scores: dict[str, float]
    fallback: bool = False
    generated_at: datetime = Field(default_factory=datetime.utcnow)


class MLHealthResponse(BaseModel):
    """Health check response for ML services."""

    status: str
    services: dict[str, Any]
    model_dir: str
    synthetic_data: bool
    redis_connected: bool


# =============================================================================
# FlyEmbed Triple-Vector Embedding Schemas
# =============================================================================


class TripleEmbeddingRequest(BaseModel):
    """Request to generate triple embeddings for a function."""

    function_id: str
    name: str
    title: str | None = None
    description: str | None = None
    category: str | None = None
    tags: list[str] = Field(default_factory=list)
    manifest: dict[str, Any] = Field(default_factory=dict)
    source_code: str | None = None
    runtime: str | None = None
    capabilities: list[str] = Field(default_factory=list)


class TripleEmbeddingResult(BaseModel):
    """Response with triple embeddings for a function."""

    function_id: str
    contract_embedding: list[float]
    semantic_embedding: list[float]
    code_embedding: list[float]
    contract_text: str
    semantic_text: str
    code_text: str
    embedding_model: str = "flyembed-v1"
    dimensions: int = 512
    latency_ms: float = 0.0


class TripleQueryRequest(BaseModel):
    """Request to generate triple query vectors for search."""

    query: str = Field(..., min_length=1, max_length=500)


class TripleQueryVector(BaseModel):
    """Triple query vectors for search."""

    query: str
    contract_vector: list[float]
    semantic_vector: list[float]
    code_vector: list[float]
    dimensions: int = 512
    latency_ms: float = 0.0


class TripleEmbeddingBatchRequest(BaseModel):
    """Request to batch generate triple embeddings."""

    functions: list[TripleEmbeddingRequest] = Field(..., max_length=50)


class TripleEmbeddingBatchResponse(BaseModel):
    """Response for batch triple embeddings."""

    results: list[TripleEmbeddingResult]
    total_count: int
    latency_ms: float = 0.0


# =============================================================================
# Phase 4: AI Composer - Function Generation Models
# =============================================================================


class FunctionGenerationRequest(BaseModel):
    """Request to generate a function using AI."""

    description: str = Field(
        ...,
        min_length=10,
        max_length=2000,
        description="Natural language description of what the function should do",
    )
    runtime: str = Field(default="python", description="Target runtime (python, nodejs, go, etc.)")
    inputs: list[dict[str, Any]] | None = Field(
        default=None, description="Optional input schema hints"
    )
    outputs: list[dict[str, Any]] | None = Field(
        default=None, description="Optional output schema hints"
    )
    constraints: str | None = Field(
        default=None, description="Optional constraints or requirements"
    )
    examples: list[str] | None = Field(
        default=None, description="Optional example inputs/outputs"
    )


class FunctionManifest(BaseModel):
    """Function I/O manifest schema."""

    name: str
    description: str
    version: str = "1.0.0"
    inputs: list[dict[str, Any]] = Field(default_factory=list)
    outputs: list[dict[str, Any]] = Field(default_factory=list)
    runtime: str
    timeout_seconds: int = 30
    memory_mb: int = 256
    capabilities: list[str] = Field(default_factory=list)


class FunctionGenerationResult(BaseModel):
    """Generated function result."""

    code: str
    runtime: str
    manifest: FunctionManifest
    explanation: str
    suggested_tests: list[str] = Field(default_factory=list)
    estimated_complexity: str  # simple, moderate, complex


class FunctionGenerationResponse(BaseModel):
    """Response for AI function generation."""

    success: bool
    result: FunctionGenerationResult | None = None
    error: str | None = None
    generation_id: str
    latency_ms: float = 0.0
    tokens_used: dict[str, int] = Field(
        default_factory=lambda: {"prompt": 0, "completion": 0, "total": 0}
    )


class GallerySearchRequest(BaseModel):
    """Request to search the function gallery."""

    query: str = Field(..., min_length=1, max_length=500)
    category: str | None = None
    runtime: str | None = None
    sort_by: str = Field(default="popular", description="popular, recent, rating, name")
    limit: int = Field(default=20, ge=1, le=100)
    offset: int = Field(default=0, ge=0)


class GalleryFunctionInfo(BaseModel):
    """Function info for gallery display."""

    id: str
    author: str
    name: str
    title: str
    description: str
    category: str | None = None
    runtime: str
    trust_score: float = Field(ge=0.0, le=100.0)
    popularity_score: int = 0
    remix_count: int = 0
    like_count: int = 0
    created_at: datetime
    updated_at: datetime


class GallerySearchResponse(BaseModel):
    """Response for gallery search."""

    query: str
    results: list[GalleryFunctionInfo]
    total_count: int
    limit: int
    offset: int


class RemixRequest(BaseModel):
    """Request to remix/fork a gallery function."""

    source_author: str
    source_name: str
    target_tenant_id: str
    new_name: str | None = None
    customizations: str | None = Field(
        default=None, description="Optional customization instructions"
    )


class RemixResponse(BaseModel):
    """Response for function remix."""

    success: bool
    new_function_id: str | None = None
    message: str
    remix_id: str


# =============================================================================
# Phase 1: AI Graph Composition - Backend as a Graph
# =============================================================================


class GraphNodeInput(BaseModel):
    """Input schema for a graph node."""

    name: str
    type: str  # string, number, boolean, object, array
    description: str
    required: bool = True
    default: Any | None = None


class GraphNodeOutput(BaseModel):
    """Output schema for a graph node."""

    name: str
    type: str
    description: str


class GraphNodeRef(BaseModel):
    """Reference to a function node in a graph."""

    node_id: str = Field(
        ..., description="Unique identifier within the graph (e.g., 'node-1', 'auth-node')"
    )
    author: str = Field(..., description="Function author (namespace)")
    name: str = Field(..., description="Function name")
    version: str = Field(default="latest", description="Function version or 'latest'")
    config: dict[str, Any] = Field(default_factory=dict, description="Node-specific configuration")
    description: str | None = None


class GraphEdgeMapping(BaseModel):
    """Data mapping between nodes."""

    source_path: str | None = Field(
        default=None, description="JSONPath in source output (e.g., '$.user.id', or '*' for all)"
    )
    target_path: str | None = Field(
        default=None, description="JSONPath in target input (e.g., '$.userId')"
    )
    transform: str | None = Field(
        default=None,
        description="Transformation: 'map', 'filter', 'reduce', 'flat', or custom script",
    )
    script: str | None = Field(
        default=None, description="Custom transformation script if transform is 'custom'"
    )


class GraphEdgeCondition(BaseModel):
    """Conditional routing for edges."""

    operator: str = Field(
        ..., description="Operator: 'eq', 'ne', 'gt', 'lt', 'contains', 'regex', 'exists'"
    )
    field: str = Field(..., description="JSONPath to field in output")
    value: Any = Field(..., description="Comparison value")


class GraphEdge(BaseModel):
    """Edge connecting two nodes in the graph."""

    id: str = Field(..., description="Unique edge identifier")
    source_node_id: str
    target_node_id: str
    mapping: GraphEdgeMapping = Field(default_factory=lambda: GraphEdgeMapping())
    condition: GraphEdgeCondition | None = None
    type: str = Field(default="sync", description="Edge type: 'sync', 'async', 'stream'")
    fallback_node_id: str | None = None


class GraphTriggerConfig(BaseModel):
    """Configuration for what triggers graph execution."""

    type: str = Field(
        ..., description="Trigger type: 'webhook', 'schedule', 'state_trigger', 'manual'"
    )
    config: dict[str, Any] = Field(default_factory=dict, description="Trigger-specific config")
    # Examples:
    # webhook: { "path": "/webhook/signup", "method": "POST" }
    # schedule: { "cron": "0 9 * * *", "timezone": "UTC" }
    # state_trigger: { "table": "users", "event": "INSERT" }
    # manual: {}


class GraphDefinition(BaseModel):
    """Complete graph definition for AI composition."""

    name: str = Field(..., description="Graph name (URL-friendly)")
    description: str
    execution_mode: str = Field(default="sync", description="sync, async, streaming, event_driven")
    nodes: list[GraphNodeRef] = Field(default_factory=list)
    edges: list[GraphEdge] = Field(default_factory=list)
    input_schema: dict[str, Any] | None = None
    output_schema: dict[str, Any] | None = None
    trigger_config: GraphTriggerConfig | None = None
    visibility: str = Field(default="public", description="public, unlisted, private")
    estimated_cost_usd: float | None = None
    estimated_latency_ms: int | None = None


class GraphCompositionRequest(BaseModel):
    """Request to generate a graph using AI composition.

    Example prompts:
    - "Create a SaaS signup flow with auth, Stripe billing, and welcome email"
    - "Build an e-commerce checkout: validate cart, process payment, send receipt"
    - "API backend for blog: CRUD posts, auth, caching"
    """

    prompt: str = Field(
        ...,
        min_length=10,
        max_length=2000,
        description="Natural language description of the backend workflow",
    )
    requirements: list[str] = Field(
        default_factory=list,
        description="Requirements: 'low_latency', 'cost_optimized', 'high_availability'",
    )
    preferred_runtime: str = Field(default="python", description="Preferred function runtime")
    tenant_id: str | None = None


class GraphCompositionExplanation(BaseModel):
    """Explanation of the composed graph."""

    summary: str
    node_purposes: dict[str, str] = Field(default_factory=dict, description="What each node does")
    data_flow_description: str
    trigger_explanation: str
    suggested_tests: list[str] = Field(default_factory=list)
    estimated_monthly_cost_usd: float | None = None


class GraphCompositionResponse(BaseModel):
    """Response for AI graph composition."""

    success: bool
    graph: GraphDefinition | None = None
    explanation: GraphCompositionExplanation | None = None
    confidence: float = Field(ge=0.0, le=1.0, description="AI confidence score")
    generation_id: str
    latency_ms: float
    tokens_used: dict[str, int] = Field(
        default_factory=lambda: {"prompt": 0, "completion": 0, "total": 0}
    )
    error: str | None = None
    suggestions: list[str] = Field(
        default_factory=list, description="Follow-up suggestions or improvements"
    )


class TemplateCategory(StrEnum):
    """Categories for prebuilt graph templates."""

    SAAS_STARTER = "saas_starter"
    MARKETPLACE = "marketplace"
    API_BACKEND = "api_backend"
    AUTHENTICATION = "authentication"
    PAYMENTS = "payments"
    WEBHOOK_PROCESSOR = "webhook_processor"
    DATA_PIPELINE = "data_pipeline"
    AI_WORKFLOW = "ai_workflow"


class GraphTemplateInfo(BaseModel):
    """Information about a prebuilt graph template."""

    id: str
    name: str
    description: str
    category: TemplateCategory
    tags: list[str] = Field(default_factory=list)
    node_count: int
    complexity: str  # simple, moderate, complex
    estimated_setup_time_minutes: int
    popular_use_cases: list[str] = Field(default_factory=list)


class GraphTemplateListResponse(BaseModel):
    """Response listing available templates."""

    templates: list[GraphTemplateInfo]
    total_count: int


class GraphTemplateRequest(BaseModel):
    """Request to instantiate a template."""

    template_id: str
    customization_prompt: str | None = None
    tenant_id: str | None = None


# ============================================================================
# Team Memory Extraction Schemas
# ============================================================================


class MemoryContent(BaseModel):
    """Structured content for a memory (type-specific)."""

    pass


class ExtractedMemory(BaseModel):
    """A single extracted memory from conversation analysis."""

    type: str = Field(..., description="Memory type: decision, preference, process, client_context")
    category: str | None = Field(None, description="Optional category like 'client:acme-corp'")
    summary: str = Field(..., description="Human-readable summary (max 100 chars)")
    content: dict[str, Any] = Field(default_factory=dict, description="Structured content object")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Extraction confidence score")
    rationale: str = Field(..., description="Why this memory is important")


class MemoryExtractionRequest(BaseModel):
    """Request to extract memories from a conversation."""

    transcript: str = Field(
        ..., min_length=10, max_length=50000, description="Conversation transcript to analyze"
    )
    team_id: str | None = Field(None, description="Team ID for context")
    conversation_id: str | None = Field(None, description="Conversation ID")
    context: dict[str, Any] | None = Field(None, description="Additional context")


class MemoryExtractionResponse(BaseModel):
    """Response with extracted memories."""

    memories: list[ExtractedMemory] = Field(default_factory=list, description="Extracted memories")
    confidence: float = Field(0.0, description="Average confidence score")
    tokens_used: int = Field(0, description="Tokens consumed")
    model: str = Field("unknown", description="Model used for extraction")
    latency_ms: float = Field(0.0, description="Processing time in milliseconds")


# =============================================================================
# Future AI Expansion - Placeholder Schemas
# =============================================================================


class AIChatMessage(BaseModel):
    """Message in AI chat conversation."""

    role: str = Field(..., description="Message role: user, assistant, system")
    content: str = Field(..., description="Message content")
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class AIChatRequest(BaseModel):
    """Request for AI conversational interface.

    Reserved for future conversational AI features.
    """

    session_id: str | None = Field(
        None, description="Existing session ID or null for new session"
    )
    user_id: str = Field(..., description="User ID")
    message: str = Field(..., min_length=1, max_length=10000, description="User message")
    context: dict[str, Any] | None = Field(None, description="Additional context")
    streaming: bool = Field(default=False, description="Enable streaming response")


class AIChatResponse(BaseModel):
    """Response from AI conversational interface.

    Reserved for future conversational AI features.
    """

    session_id: str = Field(..., description="Session ID")
    message: str = Field(..., description="AI response message")
    messages: list[AIChatMessage] = Field(
        default_factory=list, description="Full conversation history"
    )
    intent: str = Field(default="general", description="Detected intent")
    confidence: float = Field(ge=0.0, le=1.0, default=1.0, description="Response confidence")
    latency_ms: float = Field(0.0, description="Processing time in milliseconds")


class AISuggestRequest(BaseModel):
    """Request for AI code suggestion interface.

    Reserved for future code suggestion/intellisense features.
    """

    code_context: str = Field(..., description="Current code context (file content or snippet)")
    cursor_position: int | None = Field(None, description="Cursor position in code")
    file_path: str | None = Field(None, description="Path to the file being edited")
    function_id: str | None = Field(None, description="Related function ID if applicable")
    suggestion_type: str = Field(
        default="completion", description="Type: completion, fix, explain, optimize"
    )


class AISuggestion(BaseModel):
    """Single AI code suggestion."""

    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    type: str = Field(..., description="Suggestion type")
    label: str = Field(..., description="Display label")
    description: str = Field(..., description="Detailed description")
    insert_text: str = Field(..., description="Text to insert")
    range: dict[str, Any] | None = Field(None, description="Suggested edit range")
    confidence: float = Field(ge=0.0, le=1.0, default=0.8)


class AISuggestResponse(BaseModel):
    """Response from AI code suggestion interface.

    Reserved for future code suggestion/intellisense features.
    """

    suggestions: list[AISuggestion] = Field(default_factory=list, description="AI suggestions")
    context_used: str = Field(default="", description="Context that was analyzed")
    latency_ms: float = Field(0.0, description="Processing time in milliseconds")


class AIPlaceholderResponse(BaseModel):
    """Placeholder response for future AI features."""

    feature: str = Field(..., description="Feature name")
    status: str = Field(default="coming_soon", description="Feature status")
    message: str = Field(
        default="This feature is coming soon. Check back later!", description="Status message"
    )
    estimated_release: str | None = Field(None, description="Estimated release timeframe")
