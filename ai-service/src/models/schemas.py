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
    FIREWORKS = "fireworks"
    GROQ = "groq"
    DEEPINFRA = "deepinfra"
    TOGETHER = "together"
    MIMO = "mimo"
    MINIMAX = "minimax"
    STEPFUN = "stepfun"


class TrafficType(str, Enum):
    """Traffic types for provider routing."""

    REALTIME = "realtime"  # Low-latency agent function calls
    STRUCTURED = "structured"  # Structured output / tool use / JSON mode
    BACKGROUND = "background"  # Embeddings, batch processing
    FUNCTION_CALLING = "function_calling"  # Function calling optimized
    GENERAL = "general"  # Default routing


class MessageRole(str, Enum):
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

    provider: Optional[ProviderType] = None
    model: Optional[str] = None
    messages: list[ChatMessage]
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: Optional[int] = Field(default=None, ge=1)
    stream: bool = False
    top_p: Optional[float] = Field(default=None, ge=0.0, le=1.0)
    stop: Optional[list[str]] = None
    thinking: Optional[ThinkingConfig] = None


class CompletionResponse(BaseModel):
    """Response from LLM completion."""

    content: str
    provider: ProviderType
    model: str
    usage: dict[str, int] = Field(
        default_factory=lambda: {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
    )
    finish_reason: Optional[str] = None
    latency_ms: float = 0.0
    thinking_content: Optional[str] = None
    thinking_tokens: int = 0


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
    thinking_content: Optional[str] = None
    thinking_tokens: int = 0


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
    use_triple: bool = Field(default=True, description="Enable triple-vector search")
    weights: Optional[Dict[str, float]] = Field(
        default=None, description="Custom weights for contract/semantic/code triple scoring"
    )


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


# =============================================================================
# FlyEmbed Triple-Vector Embedding Schemas
# =============================================================================


class TripleEmbeddingRequest(BaseModel):
    """Request to generate triple embeddings for a function."""

    function_id: str
    name: str
    title: Optional[str] = None
    description: Optional[str] = None
    category: Optional[str] = None
    tags: List[str] = Field(default_factory=list)
    manifest: Dict[str, Any] = Field(default_factory=dict)
    source_code: Optional[str] = None
    runtime: Optional[str] = None
    capabilities: List[str] = Field(default_factory=list)


class TripleEmbeddingResult(BaseModel):
    """Response with triple embeddings for a function."""

    function_id: str
    contract_embedding: List[float]
    semantic_embedding: List[float]
    code_embedding: List[float]
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
    contract_vector: List[float]
    semantic_vector: List[float]
    code_vector: List[float]
    dimensions: int = 512
    latency_ms: float = 0.0


class TripleEmbeddingBatchRequest(BaseModel):
    """Request to batch generate triple embeddings."""

    functions: List[TripleEmbeddingRequest] = Field(..., max_length=50)


class TripleEmbeddingBatchResponse(BaseModel):
    """Response for batch triple embeddings."""

    results: List[TripleEmbeddingResult]
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
    inputs: Optional[List[Dict[str, Any]]] = Field(
        default=None, description="Optional input schema hints"
    )
    outputs: Optional[List[Dict[str, Any]]] = Field(
        default=None, description="Optional output schema hints"
    )
    constraints: Optional[str] = Field(
        default=None, description="Optional constraints or requirements"
    )
    examples: Optional[List[str]] = Field(
        default=None, description="Optional example inputs/outputs"
    )


class FunctionManifest(BaseModel):
    """Function I/O manifest schema."""

    name: str
    description: str
    version: str = "1.0.0"
    inputs: List[Dict[str, Any]] = Field(default_factory=list)
    outputs: List[Dict[str, Any]] = Field(default_factory=list)
    runtime: str
    timeout_seconds: int = 30
    memory_mb: int = 256
    capabilities: List[str] = Field(default_factory=list)


class FunctionGenerationResult(BaseModel):
    """Generated function result."""

    code: str
    runtime: str
    manifest: FunctionManifest
    explanation: str
    suggested_tests: List[str] = Field(default_factory=list)
    estimated_complexity: str  # simple, moderate, complex


class FunctionGenerationResponse(BaseModel):
    """Response for AI function generation."""

    success: bool
    result: Optional[FunctionGenerationResult] = None
    error: Optional[str] = None
    generation_id: str
    latency_ms: float = 0.0
    tokens_used: Dict[str, int] = Field(
        default_factory=lambda: {"prompt": 0, "completion": 0, "total": 0}
    )


class GallerySearchRequest(BaseModel):
    """Request to search the function gallery."""

    query: str = Field(..., min_length=1, max_length=500)
    category: Optional[str] = None
    runtime: Optional[str] = None
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
    category: Optional[str] = None
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
    results: List[GalleryFunctionInfo]
    total_count: int
    limit: int
    offset: int


class RemixRequest(BaseModel):
    """Request to remix/fork a gallery function."""

    source_author: str
    source_name: str
    target_tenant_id: str
    new_name: Optional[str] = None
    customizations: Optional[str] = Field(
        default=None, description="Optional customization instructions"
    )


class RemixResponse(BaseModel):
    """Response for function remix."""

    success: bool
    new_function_id: Optional[str] = None
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
    default: Optional[Any] = None


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
    config: Dict[str, Any] = Field(default_factory=dict, description="Node-specific configuration")
    description: Optional[str] = None


class GraphEdgeMapping(BaseModel):
    """Data mapping between nodes."""

    source_path: Optional[str] = Field(
        default=None, description="JSONPath in source output (e.g., '$.user.id', or '*' for all)"
    )
    target_path: Optional[str] = Field(
        default=None, description="JSONPath in target input (e.g., '$.userId')"
    )
    transform: Optional[str] = Field(
        default=None,
        description="Transformation: 'map', 'filter', 'reduce', 'flat', or custom script",
    )
    script: Optional[str] = Field(
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
    condition: Optional[GraphEdgeCondition] = None
    type: str = Field(default="sync", description="Edge type: 'sync', 'async', 'stream'")
    fallback_node_id: Optional[str] = None


class GraphTriggerConfig(BaseModel):
    """Configuration for what triggers graph execution."""

    type: str = Field(
        ..., description="Trigger type: 'webhook', 'schedule', 'state_trigger', 'manual'"
    )
    config: Dict[str, Any] = Field(default_factory=dict, description="Trigger-specific config")
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
    nodes: List[GraphNodeRef] = Field(default_factory=list)
    edges: List[GraphEdge] = Field(default_factory=list)
    input_schema: Optional[Dict[str, Any]] = None
    output_schema: Optional[Dict[str, Any]] = None
    trigger_config: Optional[GraphTriggerConfig] = None
    visibility: str = Field(default="public", description="public, unlisted, private")
    estimated_cost_usd: Optional[float] = None
    estimated_latency_ms: Optional[int] = None


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
    requirements: List[str] = Field(
        default_factory=list,
        description="Requirements: 'low_latency', 'cost_optimized', 'high_availability'",
    )
    preferred_runtime: str = Field(default="python", description="Preferred function runtime")
    tenant_id: Optional[str] = None


class GraphCompositionExplanation(BaseModel):
    """Explanation of the composed graph."""

    summary: str
    node_purposes: Dict[str, str] = Field(default_factory=dict, description="What each node does")
    data_flow_description: str
    trigger_explanation: str
    suggested_tests: List[str] = Field(default_factory=list)
    estimated_monthly_cost_usd: Optional[float] = None


class GraphCompositionResponse(BaseModel):
    """Response for AI graph composition."""

    success: bool
    graph: Optional[GraphDefinition] = None
    explanation: Optional[GraphCompositionExplanation] = None
    confidence: float = Field(ge=0.0, le=1.0, description="AI confidence score")
    generation_id: str
    latency_ms: float
    tokens_used: Dict[str, int] = Field(
        default_factory=lambda: {"prompt": 0, "completion": 0, "total": 0}
    )
    error: Optional[str] = None
    suggestions: List[str] = Field(
        default_factory=list, description="Follow-up suggestions or improvements"
    )


class TemplateCategory(str, Enum):
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
    tags: List[str] = Field(default_factory=list)
    node_count: int
    complexity: str  # simple, moderate, complex
    estimated_setup_time_minutes: int
    popular_use_cases: List[str] = Field(default_factory=list)


class GraphTemplateListResponse(BaseModel):
    """Response listing available templates."""

    templates: List[GraphTemplateInfo]
    total_count: int


class GraphTemplateRequest(BaseModel):
    """Request to instantiate a template."""

    template_id: str
    customization_prompt: Optional[str] = None
    tenant_id: Optional[str] = None


# ============================================================================
# Team Memory Extraction Schemas
# ============================================================================


class MemoryContent(BaseModel):
    """Structured content for a memory (type-specific)."""

    pass


class ExtractedMemory(BaseModel):
    """A single extracted memory from conversation analysis."""

    type: str = Field(..., description="Memory type: decision, preference, process, client_context")
    category: Optional[str] = Field(None, description="Optional category like 'client:acme-corp'")
    summary: str = Field(..., description="Human-readable summary (max 100 chars)")
    content: Dict[str, Any] = Field(default_factory=dict, description="Structured content object")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Extraction confidence score")
    rationale: str = Field(..., description="Why this memory is important")


class MemoryExtractionRequest(BaseModel):
    """Request to extract memories from a conversation."""

    transcript: str = Field(
        ..., min_length=10, max_length=50000, description="Conversation transcript to analyze"
    )
    team_id: Optional[str] = Field(None, description="Team ID for context")
    conversation_id: Optional[str] = Field(None, description="Conversation ID")
    context: Optional[Dict[str, Any]] = Field(None, description="Additional context")


class MemoryExtractionResponse(BaseModel):
    """Response with extracted memories."""

    memories: List[ExtractedMemory] = Field(default_factory=list, description="Extracted memories")
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

    session_id: Optional[str] = Field(
        None, description="Existing session ID or null for new session"
    )
    user_id: str = Field(..., description="User ID")
    message: str = Field(..., min_length=1, max_length=10000, description="User message")
    context: Optional[Dict[str, Any]] = Field(None, description="Additional context")
    streaming: bool = Field(default=False, description="Enable streaming response")


class AIChatResponse(BaseModel):
    """Response from AI conversational interface.

    Reserved for future conversational AI features.
    """

    session_id: str = Field(..., description="Session ID")
    message: str = Field(..., description="AI response message")
    messages: List[AIChatMessage] = Field(
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
    cursor_position: Optional[int] = Field(None, description="Cursor position in code")
    file_path: Optional[str] = Field(None, description="Path to the file being edited")
    function_id: Optional[str] = Field(None, description="Related function ID if applicable")
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
    range: Optional[Dict[str, Any]] = Field(None, description="Suggested edit range")
    confidence: float = Field(ge=0.0, le=1.0, default=0.8)


class AISuggestResponse(BaseModel):
    """Response from AI code suggestion interface.

    Reserved for future code suggestion/intellisense features.
    """

    suggestions: List[AISuggestion] = Field(default_factory=list, description="AI suggestions")
    context_used: str = Field(default="", description="Context that was analyzed")
    latency_ms: float = Field(0.0, description="Processing time in milliseconds")


class AIPlaceholderResponse(BaseModel):
    """Placeholder response for future AI features."""

    feature: str = Field(..., description="Feature name")
    status: str = Field(default="coming_soon", description="Feature status")
    message: str = Field(
        default="This feature is coming soon. Check back later!", description="Status message"
    )
    estimated_release: Optional[str] = Field(None, description="Estimated release timeframe")
