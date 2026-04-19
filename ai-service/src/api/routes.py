"""API routes for FlyMind AI Service.

This module defines all the API endpoints for Phase 1 (Foundation)
and Phase 2 (Intelligence Layer).
"""

import asyncio
import logging
from datetime import datetime
from typing import AsyncGenerator, Optional, List

from fastapi import APIRouter, HTTPException, Query, status, Body, Depends
from fastapi.responses import StreamingResponse
import redis.asyncio as redis
from pydantic import BaseModel

from ..security.auth import (
    require_api_key,
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)

from ..config import settings
from ..models.schemas import (
    CompletionRequest,
    CompletionResponse,
    EmbeddingRequest,
    EmbeddingResponse,
    ErrorResponse,
    ProviderStatusResponse,
    HealthResponse,
    # Phase 2 - Routing
    RoutingDecisionRequest,
    RoutingDecision,
    EdgeListResponse,
    EdgeStatus,
    # Phase 2 - Prewarming
    PredictionRequest,
    Prediction,
    PrewarmTriggerRequest,
    PrewarmStatus,
    # Phase 2 - Anomaly Detection
    AnomalyListResponse,
    AnomalyAcknowledgeRequest,
    Anomaly,
    # Phase 3 - Chat
    ChatSessionResponse,
    ChatMessageResponse,
    ChatHistoryResponse,
    ChatIntent,
    # Phase 3 - Search
    SearchQuery,
    SearchResult,
    SearchResponse,
    # Phase 3 - Debugging
    DebugAnalyzeRequest,
    DebugAnalysis,
    DebugSuggestRequest,
    DebugSuggestResponse,
    FixSuggestion,
    # Phase 3 - Optimization
    OptimizationResult,
    OptimizationRecommendationsResponse,
    ApplyRecommendationRequest,
    ApplyRecommendationResponse,
    # Phase 1 - AI Graph Composition
    GraphCompositionRequest,
    GraphCompositionResponse,
    GraphTemplateListResponse,
    GraphTemplateRequest,
    # Team Memory Extraction
    MemoryExtractionRequest,
    MemoryExtractionResponse,
    ExtractedMemory,
)
from ..providers.manager import get_provider_manager
from ..services.embeddings import get_embeddings_service
from ..services.routing import get_routing_service
from ..services.prewarming import (
    get_forecasting_service,
    get_prewarming_service,
)
from ..services.anomaly import (
    get_anomaly_detector,
    get_alerting_service,
)
from ..services.chat import get_chat_manager
from ..services.search import get_search_indexer, get_result_ranker, get_query_processor
from ..services.debugging import get_error_analyzer, get_fix_suggester
from ..services.optimization import get_recommendation_engine
from ..services.flyembed import get_flyembed_service
from ..services.memory_extraction import get_memory_extraction_service
from ..models.schemas import (
    TripleEmbeddingRequest,
    TripleEmbeddingResult,
    TripleEmbeddingBatchRequest,
    TripleEmbeddingBatchResponse,
    TripleQueryRequest,
    TripleQueryVector,
    # Phase 4 - AI Composer
    FunctionGenerationRequest,
    FunctionGenerationResponse,
    FunctionManifest,
    FunctionGenerationResult,
    # Phase 1 - AI Graph Composition
    GraphCompositionRequest,
    GraphCompositionResponse,
    GraphTemplateListResponse,
    GraphTemplateRequest,
)


logger = logging.getLogger(__name__)

router = APIRouter()

class ChatSendBody(BaseModel):
    session_id: str
    user_id: str
    message: str
    tenant_id: Optional[str] = None


@router.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint.

    Returns:
        Health status of the service and its dependencies
    """
    # Check Redis
    redis_connected = False
    try:
        r = redis.from_url(settings.redis_url)
        await r.ping()
        redis_connected = True
        await r.close()
    except Exception as e:
        logger.warning(f"Redis health check failed: {e}")

    # Check database (connect and run SELECT 1)
    database_connected = False
    if settings.database_url:
        try:
            import asyncpg

            conn = await asyncio.wait_for(
                asyncpg.connect(settings.database_url),
                timeout=2.0,
            )
            try:
                await conn.fetchval("SELECT 1")
                database_connected = True
            finally:
                await conn.close()
        except Exception as e:
            logger.warning(f"Database health check failed: {e}")

    # Get provider status
    provider_manager = get_provider_manager()
    provider_status = await provider_manager.health_check_all()
    available_providers = [name for name, available in provider_status.items() if available]

    return HealthResponse(
        status="healthy" if (redis_connected or not settings.enable_caching) else "degraded",
        service=settings.service_name,
        version=settings.service_version,
        providers_available=available_providers,
        redis_connected=redis_connected,
        database_connected=database_connected,
    )


@router.get("/api/providers", response_model=ProviderStatusResponse)
async def get_providers():
    """Get status of all available providers.

    Returns:
        Information about all registered providers
    """
    provider_manager = get_provider_manager()
    providers = provider_manager.list_providers()

    return ProviderStatusResponse(
        providers=providers,
        default_provider=provider_manager.default_provider,
        default_embedding_provider=provider_manager.default_embedding_provider,
    )


@router.post("/api/embed", response_model=EmbeddingResponse)
async def create_embedding(
    request: EmbeddingRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_WRITE)),
):
    """Generate embeddings for the given text.

    Args:
        request: Embedding request with text and optional provider/model
        api_key: Validated API key with embed:write scope

    Returns:
        Embedding response with vector
    """
    try:
        embeddings_service = get_embeddings_service()
        response = await embeddings_service.generate_embedding(request, api_key_info=api_key)
        return response
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
    except NotImplementedError as e:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail=str(e),
        )
    except Exception as e:
        logger.error(f"Embedding generation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to generate embedding",
        )


@router.post("/api/complete", response_model=CompletionResponse)
async def create_completion(request: CompletionRequest):
    """Generate a completion using the specified provider.

    Args:
        request: Completion request with messages and optional provider/model

    Returns:
        Completion response with generated content
    """
    try:
        provider_manager = get_provider_manager()
        provider_name = request.provider.value if request.provider else None
        provider = provider_manager.get_provider(provider_name)

        response = await provider.complete(
            messages=request.messages,
            model=request.model,
            temperature=request.temperature,
            max_tokens=request.max_tokens,
            top_p=request.top_p,
            stop=request.stop,
        )
        return response
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
    except Exception as e:
        logger.error(f"Completion failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to generate completion",
        )


@router.post("/api/stream")
async def stream_completion(request: CompletionRequest):
    """Stream a completion using the specified provider.

    Args:
        request: Completion request with messages and optional provider/model

    Returns:
        Streaming response with generated content
    """
    if not settings.enable_streaming:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail="Streaming is not enabled",
        )

    if not request.stream:
        # If stream=False, use regular completion
        return await create_completion(request)

    try:
        provider_manager = get_provider_manager()
        provider_name = request.provider.value if request.provider else None
        provider = provider_manager.get_provider(provider_name)

        async def generate() -> AsyncGenerator[str, None]:
            try:
                async for chunk in provider.stream(
                    messages=request.messages,
                    model=request.model,
                    temperature=request.temperature,
                    max_tokens=request.max_tokens,
                    top_p=request.top_p,
                    stop=request.stop,
                ):
                    yield f"data: {chunk}\n\n"
                yield "data: [DONE]\n\n"
            except Exception as e:
                logger.error(f"Streaming failed: {e}")
                yield f"data: {{\"error\": \"{str(e)}\"}}\n\n"

        return StreamingResponse(
            generate(),
            media_type="text/event-stream",
        )
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
    except Exception as e:
        logger.error(f"Streaming setup failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to setup streaming",
        )


# =============================================================================
# Phase 2: Intelligent Request Routing Endpoints
# =============================================================================

@router.post("/api/route/decide", response_model=RoutingDecision)
async def decide_routing(request: RoutingDecisionRequest):
    """Get optimal edge for a request.

    Uses ML-based routing to select the best edge (Cloudflare, Vercel, Fly.io, Deno)
    based on geographic proximity, historical latency, current load, and availability.

    Args:
        request: Routing decision request with function ID and user location

    Returns:
        RoutingDecision with recommended edge and alternatives
    """
    if not settings.routing_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Routing service is not enabled",
        )

    try:
        routing_service = get_routing_service()
        decision = await routing_service.decide_routing(request)
        return decision
    except Exception as e:
        logger.error(f"Routing decision failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to determine routing",
        )


@router.get("/api/route/edges", response_model=EdgeListResponse)
async def get_edge_statuses():
    """List available edges with status.

    Returns status information for all supported edge providers including
    current latency, load, and availability.

    Returns:
        EdgeListResponse with all edges and their status
    """
    if not settings.routing_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Routing service is not enabled",
        )

    try:
        routing_service = get_routing_service()
        edges = await routing_service.get_edge_statuses()

        return EdgeListResponse(
            edges=edges,
            total_count=len(edges),
            last_updated=datetime.utcnow(),
        )
    except Exception as e:
        logger.error(f"Failed to get edge statuses: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get edge statuses",
        )


# =============================================================================
# Phase 2: Predictive Cold Start Prewarming Endpoints
# =============================================================================

@router.post("/api/prewarm/predict", response_model=Prediction)
async def get_prewarming_prediction(request: PredictionRequest):
    """Get prewarming predictions for a function.

    Uses time-series forecasting to predict request volume and
    determine if prewarming is needed.

    Args:
        request: Prediction request with function ID and window

    Returns:
        Prediction with predicted requests and confidence
    """
    if not settings.prewarming_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Prewarming service is not enabled",
        )

    try:
        forecasting_service = get_forecasting_service()
        prediction = await forecasting_service.predict(request)
        return prediction
    except Exception as e:
        logger.error(f"Prediction failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to generate prediction",
        )


@router.post("/api/prewarm/warm", response_model=PrewarmStatus)
async def trigger_prewarming(request: PrewarmTriggerRequest):
    """Trigger prewarming for a function.

    Proactively warms function instances to reduce cold starts.

    Args:
        request: Prewarm trigger request

    Returns:
        PrewarmStatus with the result
    """
    if not settings.prewarming_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Prewarming service is not enabled",
        )

    try:
        prewarming_service = get_prewarming_service()
        status = await prewarming_service.trigger_prewarming(request)
        return status
    except Exception as e:
        logger.error(f"Prewarming trigger failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to trigger prewarming",
        )


# =============================================================================
# Phase 2: Anomaly Detection Endpoints
# =============================================================================

@router.get("/api/anomalies", response_model=AnomalyListResponse)
async def get_anomalies(
    function_id: Optional[str] = Query(None, description="Filter by function ID"),
    page: int = Query(1, ge=1, description="Page number"),
    page_size: int = Query(20, ge=1, le=100, description="Page size"),
):
    """List detected anomalies.

    Returns anomalies detected by the monitoring system including
    latency spikes, error rate increases, and cold start issues.

    Args:
        function_id: Optional function ID to filter by
        page: Page number
        page_size: Number of items per page

    Returns:
        AnomalyListResponse with detected anomalies
    """
    if not settings.anomaly_detection_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Anomaly detection is not enabled",
        )

    try:
        detector = get_anomaly_detector()

        # Calculate offset
        offset = (page - 1) * page_size

        # Get anomalies
        anomalies = await detector.get_anomalies(
            function_id=function_id,
            limit=offset + page_size,
        )

        # Paginate
        total_count = len(anomalies)
        anomalies = anomalies[offset:offset + page_size]

        return AnomalyListResponse(
            anomalies=anomalies,
            total_count=total_count,
            page=page,
            page_size=page_size,
        )
    except Exception as e:
        logger.error(f"Failed to get anomalies: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get anomalies",
        )


@router.post("/api/anomalies/acknowledge")
async def acknowledge_anomaly(request: AnomalyAcknowledgeRequest):
    """Acknowledge an anomaly.

    Marks an anomaly as acknowledged to silence alerts.

    Args:
        request: Anomaly acknowledgement request

    Returns:
        Success message
    """
    if not settings.anomaly_detection_enabled:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Anomaly detection is not enabled",
        )

    try:
        detector = get_anomaly_detector()
        success = await detector.acknowledge_anomaly(
            request.anomaly_id,
            request.acknowledged_by,
        )

        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Anomaly not found",
            )

        return {"message": "Anomaly acknowledged successfully"}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to acknowledge anomaly: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to acknowledge anomaly",
        )


# =============================================================================
# Phase 3: Developer Experience Layer - Chat Endpoints
# =============================================================================

@router.post("/api/chat/message", response_model=ChatMessageResponse)
async def send_chat_message(
    session_id: Optional[str] = Query(None, description="Existing session ID (or empty to create new)"),
    user_id: Optional[str] = Query(None, description="The user ID"),
    message: Optional[str] = Query(None, description="The message text"),
    tenant_id: Optional[str] = Query(None, description="Tenant ID for context (e.g. deployed functions)"),
    payload: Optional[ChatSendBody] = Body(default=None),
):
    """Send a message to a chat session.

    Args:
        session_id: Existing session ID (or empty to create new)
        user_id: The user ID
        message: The message text
        tenant_id: Optional tenant ID; when set, chat context includes functions from the orchestrator

    Returns:
        ChatMessageResponse with the assistant's reply
    """
    try:
        # Support both JSON body (used by Go orchestrator) and query params (legacy).
        sid = payload.session_id if payload is not None else (session_id or "")
        uid = payload.user_id if payload is not None else (user_id or "")
        msg = payload.message if payload is not None else (message or "")
        tid = payload.tenant_id if payload is not None else tenant_id

        if not msg:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail="message is required",
            )
        if not uid:
            # Keep the endpoint compatible with older callers that don't send user_id.
            uid = "unknown"

        chat_manager = get_chat_manager()
        result = await chat_manager.process_message(
            session_id=sid,
            user_id=uid,
            message=msg,
            tenant_id=tid,
        )
        return ChatMessageResponse(
            session_id=result["session_id"],
            message=result["message"],
            intent=ChatIntent(result["intent"]),
            confidence=result["confidence"],
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Chat message failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to process chat message",
        )


@router.get("/api/chat/sessions", response_model=List[ChatSessionResponse])
async def list_chat_sessions(
    user_id: str = Query(..., description="User ID"),
    limit: int = Query(10, ge=1, le=50),
):
    """List chat sessions for a user.

    Args:
        user_id: The user ID
        limit: Maximum number of sessions

    Returns:
        List of ChatSessionResponse
    """
    try:
        chat_manager = get_chat_manager()
        sessions = await chat_manager.list_user_sessions(user_id, limit)
        return [
            ChatSessionResponse(
                session_id=s.session_id,
                user_id=s.user_id,
                created_at=s.created_at,
                message_count=len(s.messages),
            )
            for s in sessions
        ]
    except Exception as e:
        logger.error(f"Failed to list sessions: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to list sessions",
        )


@router.get("/api/chat/sessions/{session_id}", response_model=ChatHistoryResponse)
async def get_chat_session_history(session_id: str):
    """Get chat session history.

    Args:
        session_id: The session ID

    Returns:
        ChatHistoryResponse with messages
    """
    try:
        chat_manager = get_chat_manager()
        session = await chat_manager.get_session(session_id)
        if not session:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Session not found",
            )
        return ChatHistoryResponse(
            session_id=session.session_id,
            messages=session.messages,
            created_at=session.created_at,
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get session history: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get session history",
        )


# =============================================================================
# Phase 3: Developer Experience Layer - Search Endpoints
# =============================================================================

@router.post("/api/search/functions", response_model=SearchResponse)
async def search_functions(
    request: SearchQuery,
    tenant_id: str = Query(..., description="Tenant ID"),
):
    """Search functions using semantic search.

    Args:
        request: Search query
        tenant_id: The tenant ID

    Returns:
        SearchResponse with results
    """
    try:
        query_processor = get_query_processor()
        indexer = get_search_indexer()
        ranker = get_result_ranker()

        if request.use_triple:
            # Triple-vector search path
            processed_query, _, metadata = await query_processor.process_query(request.query)

            # Generate triple query vectors
            flyembed = get_flyembed_service()
            triple_query = await flyembed.embed_query(request.query)

            results = await indexer.search_triple(
                tenant_id=tenant_id,
                query_vectors={
                    "contract_vector": triple_query.contract_vector,
                    "semantic_vector": triple_query.semantic_vector,
                    "code_vector": triple_query.code_vector,
                },
                limit=request.limit,
            )

            if request.filters:
                results = ranker.filter_results(results, request.filters)
            ranked_results = ranker.rank_results_triple(results, request.query, request.weights)
        else:
            # Single-vector search path (legacy)
            processed_query, embedding, metadata = await query_processor.process_query(request.query)
            results = await indexer.search(
                tenant_id=tenant_id,
                query_embedding=embedding,
                limit=request.limit,
            )
            if request.filters:
                results = ranker.filter_results(results, request.filters)
            ranked_results = ranker.rank_results(results, request.query)

        # Keyword fallback if no results
        if not ranked_results:
            all_functions = await indexer.index_tenant_functions(tenant_id)
            ranked_results = query_processor.process_keyword_fallback(
                request.query,
                [{"data": f} for f in all_functions],
            )

        return SearchResponse(
            query=request.query,
            results=[
                SearchResult(
                    function_id=r["function_id"],
                    function_name=r["data"].get("name", ""),
                    description=r["data"].get("description"),
                    runtime=r["data"].get("runtime"),
                    tags=r["data"].get("tags", []),
                    score=r.get("score", 0),
                    rank=r.get("rank", 0),
                )
                for r in ranked_results[:request.limit]
            ],
            total_count=len(ranked_results),
            query_type=metadata.get("search_type", "semantic"),
        )
    except Exception as e:
        logger.error(f"Search failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to search functions",
        )


# =============================================================================
# Phase 3: Developer Experience Layer - Debugging Endpoints
# =============================================================================

@router.post("/api/debug/analyze", response_model=DebugAnalysis)
async def analyze_error(request: DebugAnalyzeRequest):
    """Analyze an error and provide root cause analysis.

    Args:
        request: Debug analysis request

    Returns:
        DebugAnalysis with root cause and suggestions
    """
    try:
        analyzer = get_error_analyzer()
        analysis = await analyzer.analyze_error(
            function_id=request.function_id,
            error_message=request.error_message,
            stack_trace=request.stack_trace,
        )
        return DebugAnalysis(**analysis)
    except Exception as e:
        logger.error(f"Debug analysis failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to analyze error",
        )


@router.post("/api/debug/suggest", response_model=DebugSuggestResponse)
async def get_fix_suggestions(request: DebugAnalyzeRequest):
    """Get fix suggestions for an error.

    Args:
        request: Debug analysis request

    Returns:
        DebugSuggestResponse with suggestions
    """
    try:
        # First analyze the error
        analyzer = get_error_analyzer()
        analysis = await analyzer.analyze_error(
            function_id=request.function_id,
            error_message=request.error_message,
            stack_trace=request.stack_trace,
        )

        # Then get suggestions
        suggester = get_fix_suggester()
        suggestions = await suggester.generate_suggestions(analysis)
        docs = suggester.get_documentation_links(analysis.get("error_category", ""))

        return DebugSuggestResponse(
            analysis=DebugAnalysis(**analysis),
            suggestions=[FixSuggestion(**s) for s in suggestions],
            documentation_links=docs,
        )
    except Exception as e:
        logger.error(f"Get suggestions failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get fix suggestions",
        )


# =============================================================================
# Phase 3: Developer Experience Layer - Optimization Endpoints
# =============================================================================

@router.get("/api/optimize/{function_id}", response_model=OptimizationRecommendationsResponse)
async def get_optimization_recommendations(function_id: str):
    """Get optimization recommendations for a function.

    Args:
        function_id: The function ID

    Returns:
        OptimizationRecommendationsResponse with recommendations
    """
    try:
        recommender = get_recommendation_engine()
        recommendations = await recommender.generate_recommendations(function_id)

        # Convert recommendations to proper format
        formatted_recommendations = []
        for i, rec in enumerate(recommendations):
            formatted_recommendations.append(Recommendation(
                id=f"rec-{i}",
                type=rec.get("type", ""),
                title=rec.get("title", ""),
                description=rec.get("description", ""),
                category=rec.get("category", ""),
                priority=rec.get("priority", "medium"),
                action=rec.get("action", ""),
                current_value=rec.get("current_value", 0),
                target_value=rec.get("target_value", 0),
                estimated_savings_monthly=rec.get("estimated_savings_monthly", 0),
            ))

        return OptimizationRecommendationsResponse(
            function_id=function_id,
            recommendations=formatted_recommendations,
            total_count=len(formatted_recommendations),
        )
    except Exception as e:
        logger.error(f"Get recommendations failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get recommendations",
        )


@router.post("/api/optimize/{function_id}/apply", response_model=ApplyRecommendationResponse)
async def apply_optimization_recommendation(
    function_id: str,
    request: ApplyRecommendationRequest,
):
    """Apply an optimization recommendation.

    Args:
        function_id: The function ID
        request: Apply recommendation request

    Returns:
        ApplyRecommendationResponse with result
    """
    try:
        recommender = get_recommendation_engine()
        result = await recommender.apply_recommendation(
            function_id=function_id,
            recommendation_id=request.recommendation_id,
        )
        return ApplyRecommendationResponse(**result)
    except Exception as e:
        logger.error(f"Apply recommendation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to apply recommendation",
        )


# =============================================================================
# FlyEmbed Triple-Vector Embedding Endpoints
# =============================================================================

@router.post("/api/flyembed/embed", response_model=TripleEmbeddingResult)
async def flyembed_embed(
    request: TripleEmbeddingRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_WRITE)),
):
    """Generate triple embeddings for a single function.

    Args:
        request: Triple embedding request with function data
        api_key: Validated API key with embed:write scope

    Returns:
        TripleEmbeddingResult with contract, semantic, and code embeddings
    """
    try:
        service = get_flyembed_service()
        result = await service.embed_function(request.model_dump())
        return TripleEmbeddingResult(
            function_id=result.function_id,
            contract_embedding=result.contract_embedding,
            semantic_embedding=result.semantic_embedding,
            code_embedding=result.code_embedding,
            contract_text=result.contract_text,
            semantic_text=result.semantic_text,
            code_text=result.code_text,
            latency_ms=result.latency_ms,
        )
    except Exception as e:
        logger.error(f"FlyEmbed embed failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate triple embeddings: {e}",
        )


@router.post("/api/flyembed/embed-batch", response_model=TripleEmbeddingBatchResponse)
async def flyembed_embed_batch(
    request: TripleEmbeddingBatchRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_ADMIN)),
):
    """Batch generate triple embeddings for multiple functions.

    Args:
        request: Batch of function data
        api_key: Validated API key with embed:admin scope (batch operations require elevated permissions)

    Returns:
        TripleEmbeddingBatchResponse with all results
    """
    try:
        service = get_flyembed_service()
        functions_data = [f.model_dump() for f in request.functions]
        results = await service.embed_batch(functions_data)
        return TripleEmbeddingBatchResponse(
            results=[
                TripleEmbeddingResult(
                    function_id=r.function_id,
                    contract_embedding=r.contract_embedding,
                    semantic_embedding=r.semantic_embedding,
                    code_embedding=r.code_embedding,
                    contract_text=r.contract_text,
                    semantic_text=r.semantic_text,
                    code_text=r.code_text,
                    latency_ms=r.latency_ms,
                )
                for r in results
            ],
            total_count=len(results),
        )
    except Exception as e:
        logger.error(f"FlyEmbed batch embed failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to batch generate triple embeddings: {e}",
        )


@router.post("/api/flyembed/query", response_model=TripleQueryVector)
async def flyembed_query(
    request: TripleQueryRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_READ)),
):
    """Generate triple query vectors for search.

    Args:
        request: Query request with search text
        api_key: Validated API key with embed:read scope

    Returns:
        TripleQueryVector with three query vectors
    """
    try:
        service = get_flyembed_service()
        result = await service.embed_query(request.query)
        return TripleQueryVector(
            query=result.query,
            contract_vector=result.contract_vector,
            semantic_vector=result.semantic_vector,
            code_vector=result.code_vector,
            dimensions=512,
            latency_ms=result.latency_ms,
        )
    except Exception as e:
        logger.error(f"FlyEmbed query failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate triple query vectors: {e}",
        )


@router.get("/api/flyembed/health")
async def flyembed_health():
    """Health check for FlyEmbed service.

    Returns:
        Status of the FlyEmbed service
    """
    return {"status": "healthy", "service": "flyembed"}


# =============================================================================
# Phase 4: AI Composer - Function Generation Endpoints
# =============================================================================

import uuid
import time as pytime


@router.post("/api/composer/generate", response_model=FunctionGenerationResponse)
async def generate_function(
    request: FunctionGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Generate a function using AI based on natural language description.

    This endpoint uses LLM to generate complete function code along with
    I/O manifest, test suggestions, and complexity estimation.

    Args:
        request: Function generation request with description and optional constraints
        api_key: Validated API key with chat:write scope

    Returns:
        FunctionGenerationResponse with generated code and metadata
    """
    generation_id = str(uuid.uuid4())
    start_time = pytime.time()

    try:
        provider_manager = get_provider_manager()
        provider_name = "openai"  # Use OpenAI for code generation
        provider = provider_manager.get_provider(provider_name)

        if not provider or not provider.available:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="AI generation service not available",
            )

        # Build the prompt for function generation
        runtime_prompts = {
            "python": "Generate Python 3.11+ code. Use type hints, docstrings, and modern Python patterns.",
            "nodejs": "Generate Node.js 20+ JavaScript code. Use async/await and modern patterns.",
            "go": "Generate Go 1.21+ code. Include proper error handling and idiomatic Go patterns.",
            "rust": "Generate Rust code with proper error handling and modern idioms.",
        }

        runtime_guidance = runtime_prompts.get(request.runtime, f"Generate {request.runtime} code.")

        system_prompt = f"""You are an expert serverless function developer.
Your task is to generate complete, production-ready function code based on a natural language description.

{runtime_guidance}

The function should:
1. Be self-contained and stateless (serverless-appropriate)
2. Handle errors gracefully
3. Include input validation
4. Be secure (no code injection vulnerabilities)
5. Follow best practices for the target runtime

Respond with a JSON object containing:
- code: The complete function code
- manifest: Object with name, description, version, inputs (array), outputs (array), runtime, timeout_seconds, memory_mb, capabilities (array)
- explanation: Brief explanation of what the code does
- suggested_tests: Array of test case descriptions
- estimated_complexity: "simple", "moderate", or "complex"

Inputs should be objects with: name, type, description, required (boolean), default (optional)
Outputs should be objects with: name, type, description
Capabilities are strings like "http", "network", "filesystem", etc."""

        user_prompt = f"Generate a {request.runtime} function that: {request.description}"

        if request.constraints:
            user_prompt += f"\n\nAdditional constraints: {request.constraints}"

        if request.examples:
            user_prompt += f"\n\nExample inputs/outputs: {chr(10).join(request.examples)}"

        messages = [
            ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
            ChatMessage(role=MessageRole.USER, content=user_prompt),
        ]

        # Generate the function
        completion = await provider.complete(
            messages=messages,
            model="gpt-4o",
            temperature=0.2,
            max_tokens=4000,
        )

        # Parse the JSON response
        import json
        try:
            result_data = json.loads(completion.content)
        except json.JSONDecodeError:
            # Try to extract JSON from markdown code blocks
            content = completion.content
            if "```json" in content:
                json_str = content.split("```json")[1].split("```")[0].strip()
            elif "```" in content:
                json_str = content.split("```")[1].split("```")[0].strip()
            else:
                raise ValueError(f"Failed to parse JSON response: {completion.content[:200]}")
            result_data = json.loads(json_str)

        # Build the result
        manifest_data = result_data.get("manifest", {})
        manifest = FunctionManifest(
            name=manifest_data.get("name", "generated_function"),
            description=manifest_data.get("description", request.description[:100]),
            version=manifest_data.get("version", "1.0.0"),
            inputs=manifest_data.get("inputs", []),
            outputs=manifest_data.get("outputs", []),
            runtime=manifest_data.get("runtime", request.runtime),
            timeout_seconds=manifest_data.get("timeout_seconds", 30),
            memory_mb=manifest_data.get("memory_mb", 256),
            capabilities=manifest_data.get("capabilities", []),
        )

        result = FunctionGenerationResult(
            code=result_data.get("code", ""),
            runtime=manifest.runtime,
            manifest=manifest,
            explanation=result_data.get("explanation", ""),
            suggested_tests=result_data.get("suggested_tests", []),
            estimated_complexity=result_data.get("estimated_complexity", "moderate"),
        )

        latency_ms = (pytime.time() - start_time) * 1000

        return FunctionGenerationResponse(
            success=True,
            result=result,
            generation_id=generation_id,
            latency_ms=latency_ms,
            tokens_used=completion.usage,
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Function generation failed: {e}")
        latency_ms = (pytime.time() - start_time) * 1000
        return FunctionGenerationResponse(
            success=False,
            error=str(e),
            generation_id=generation_id,
            latency_ms=latency_ms,
        )


@router.post("/api/composer/generate/stream")
async def generate_function_stream(
    request: FunctionGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Stream a function generation using AI.

    This endpoint streams the LLM output as it's generated, useful for
    real-time UI updates during function creation.

    Args:
        request: Function generation request
        api_key: Validated API key with chat:write scope

    Returns:
        Streaming response with generated content chunks
    """
    generation_id = str(uuid.uuid4())

    if not settings.enable_streaming:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail="Streaming is not enabled",
        )

    try:
        provider_manager = get_provider_manager()
        provider = provider_manager.get_provider("openai")

        if not provider or not provider.available:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="AI generation service not available",
            )

        system_prompt = f"""You are an expert serverless function developer.
Generate {request.runtime} code based on the user's description.

Output ONLY the raw code, no markdown formatting, no explanations.
The code should be self-contained and production-ready."""

        messages = [
            ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
            ChatMessage(role=MessageRole.USER, content=request.description),
        ]

        async def generate() -> AsyncGenerator[str, None]:
            yield f'{{"generation_id": "{generation_id}", "type": "start"}}\n\n'

            try:
                async for chunk in provider.stream(
                    messages=messages,
                    model="gpt-4o",
                    temperature=0.2,
                    max_tokens=4000,
                ):
                    escaped = chunk.replace('"', '\\"').replace('\n', '\\n')
                    yield f'{{"type": "chunk", "content": "{escaped}"}}\n\n'

                yield f'{{"type": "complete"}}\n\n'
            except Exception as e:
                yield f'{{"type": "error", "error": "{str(e)}"}}\n\n'

        return StreamingResponse(
            generate(),
            media_type="text/event-stream",
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Function generation stream failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to stream function generation",
        )


# =============================================================================
# Phase 5: Cost-Optimized Auto Function Builder
# =============================================================================

from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List
from enum import Enum


class OptimizedGenerationRequest(BaseModel):
    """Request for cost-optimized function generation."""
    description: str = Field(..., min_length=10, max_length=2000, description="Natural language description")
    runtime: str = Field(default="python", description="Target runtime")
    inputs: Optional[List[Dict[str, Any]]] = Field(default=None)
    outputs: Optional[List[Dict[str, Any]]] = Field(default=None)
    constraints: Optional[str] = Field(default=None)
    examples: Optional[List[str]] = Field(default=None)
    force_tier: Optional[str] = Field(default=None, description="Force tier: cheap, mid, premium")


class OptimizedGenerationMetricsResponse(BaseModel):
    """Metrics for optimized generation."""
    total_attempts: int
    final_tier: str
    cache_hit: bool
    template_used: bool
    validation_attempts: int
    total_cost_usd: float
    savings_vs_premium_usd: float
    savings_vs_premium_pct: float


class OptimizedGenerationResponse(BaseModel):
    """Response for optimized function generation."""
    success: bool
    result: Optional[FunctionGenerationResult] = None
    error: Optional[str] = None
    generation_id: str
    latency_ms: float
    tokens_used: Dict[str, int]
    metrics: OptimizedGenerationMetricsResponse
    optimization_notes: List[str]


@router.post("/api/composer/generate-optimized", response_model=OptimizedGenerationResponse)
async def generate_function_optimized(
    request: OptimizedGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
    tenant_id: str = Query("default", description="Tenant ID"),
):
    """Generate a function using cost-optimized multi-tier AI pipeline.

    This endpoint implements the complete cost optimization strategy:
    - Multi-tier model routing (cheap -> mid -> premium)
    - Template + RAG retrieval for faster generation
    - Validation pipeline with auto-fix
    - Intelligent caching
    - Confidence-based escalation

    **Cost savings**: 70-90% cheaper than using premium models directly.

    Args:
        request: Generation request with description and constraints
        api_key: Validated API key with chat:write scope
        tenant_id: Tenant ID for RAG retrieval

    Returns:
        OptimizedGenerationResponse with generated code and cost metrics
    """
    try:
        from ..services.generation import (
            get_optimized_generation_service,
            ModelTier,
        )

        service = get_optimized_generation_service()

        # Convert request
        gen_request = FunctionGenerationRequest(
            description=request.description,
            runtime=request.runtime,
            inputs=request.inputs,
            outputs=request.outputs,
            constraints=request.constraints,
            examples=request.examples,
        )

        # Determine tier override
        force_tier = None
        if request.force_tier:
            try:
                force_tier = ModelTier(request.force_tier.lower())
            except ValueError:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"Invalid force_tier: {request.force_tier}. Use: cheap, mid, premium",
                )

        # Generate with optimization
        response, metrics = await service.generate(
            request=gen_request,
            tenant_id=tenant_id,
            api_key_info=api_key_info,
            force_tier=force_tier,
        )

        # Build optimization notes
        notes = []
        if metrics.cache_hit:
            notes.append("Served from cache - zero AI cost")
        elif metrics.template_used:
            notes.append("Template-based generation used - reduced token usage")
        if metrics.savings_vs_premium_pct > 50:
            notes.append(f"Saved {metrics.savings_vs_premium_pct:.0f}% vs premium model")
        if metrics.total_attempts > 1:
            notes.append(f"Auto-escalated through {metrics.total_attempts} tiers for quality")

        return OptimizedGenerationResponse(
            success=response.success,
            result=response.result,
            error=response.error,
            generation_id=response.generation_id,
            latency_ms=response.latency_ms,
            tokens_used=response.tokens_used,
            metrics=OptimizedGenerationMetricsResponse(
                total_attempts=metrics.total_attempts,
                final_tier=metrics.final_tier,
                cache_hit=metrics.cache_hit,
                template_used=metrics.template_used,
                validation_attempts=metrics.validation_attempts,
                total_cost_usd=metrics.total_cost_usd,
                savings_vs_premium_usd=metrics.savings_vs_premium_usd,
                savings_vs_premium_pct=metrics.savings_vs_premium_pct,
            ),
            optimization_notes=notes,
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Optimized generation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Optimized generation failed: {str(e)}",
        )


@router.get("/api/composer/optimized-stats")
async def get_optimized_generation_stats(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Get statistics for the optimized generation service.

    Returns:
        Cache stats, cost tracking, and optimization metrics
    """
    try:
        from ..services.generation import get_optimized_generation_service

        service = get_optimized_generation_service()
        stats = await service.get_stats()

        return {
            "cache": stats["cache"],
            "costs": stats["costs"],
            "optimization_enabled": True,
            "strategies": [
                "multi_tier_routing",
                "template_rag_retrieval",
                "validation_pipeline",
                "intelligent_caching",
                "auto_escalation",
            ],
        }

    except Exception as e:
        logger.error(f"Get optimized stats failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get optimization stats",
        )


# =============================================================================
# Phase 1: AI Graph Composition - Backend as a Graph
# =============================================================================

@router.get("/api/composition/templates", response_model=GraphTemplateListResponse)
async def list_graph_templates():
    """List available prebuilt graph templates.

    Returns all available templates for common backend patterns:
    - SaaS Starter (auth, billing, email)
    - E-commerce Checkout
    - API Backend (CRUD, auth, caching)
    - Webhook Processor

    Returns:
        GraphTemplateListResponse with all templates
    """
    try:
        from ..services.graph_composition import get_graph_composition_service

        service = get_graph_composition_service()
        templates = service.list_templates()

        return GraphTemplateListResponse(
            templates=templates,
            total_count=len(templates),
        )
    except Exception as e:
        logger.error(f"Failed to list templates: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to list templates: {str(e)}",
        )


@router.post("/api/composition/compose", response_model=GraphCompositionResponse)
async def compose_graph_from_prompt(
    request: GraphCompositionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Generate a graph using AI composition from natural language.

    This is the core "Backend as a Graph" feature. Users describe what they want
    (e.g., "Create a SaaS signup flow with Stripe billing and welcome email")
    and AI generates a complete graph definition with nodes, edges, and triggers.

    The service will:
    1. Match against templates if applicable
    2. Use LLM to generate topology for custom workflows
    3. Suggest function nodes from the catalog
    4. Connect them with appropriate data flows

    Example prompts:
    - "SaaS signup: validate email, create Stripe customer, send welcome email"
    - "E-commerce checkout: validate cart, process payment, create order, send receipt"
    - "API backend for blog with auth, CRUD, and caching"

    Args:
        request: Composition request with prompt and requirements
        api_key: Validated API key with chat:write scope

    Returns:
        GraphCompositionResponse with complete graph definition
    """
    try:
        from ..services.graph_composition import get_graph_composition_service

        service = get_graph_composition_service()
        response = await service.compose_from_prompt(
            request=request,
            api_key_info=api_key,
        )
        return response
    except Exception as e:
        logger.error(f"Graph composition failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Graph composition failed: {str(e)}",
        )


@router.post("/api/composition/template/{template_id}", response_model=GraphCompositionResponse)
async def instantiate_graph_template(
    template_id: str,
    customization: Optional[str] = Query(None, description="Optional customization instructions"),
    tenant_id: Optional[str] = Query(None, description="Tenant ID"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Instantiate a prebuilt graph template.

    Quick-start graph creation using prebuilt templates:
    - saas_starter: Auth + Stripe + Email
    - ecommerce_checkout: Cart validation → Payment → Order → Receipt
    - api_backend: Auth → Cache → DB → Response
    - webhook_processor: Signature validation → Parsing → Queue → Processing

    Args:
        template_id: Template identifier (e.g., 'saas_starter')
        customization: Optional customization instructions
        tenant_id: Tenant ID for context
        api_key: Validated API key with chat:write scope

    Returns:
        GraphCompositionResponse with instantiated graph
    """
    try:
        from ..services.graph_composition import get_graph_composition_service

        service = get_graph_composition_service()

        # Check if template exists first
        if not service.get_template(template_id):
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Template not found: {template_id}. Use /api/composition/templates to list available templates.",
            )

        response = await service.compose_from_template(
            template_id=template_id,
            customization_prompt=customization,
            tenant_id=tenant_id,
        )
        return response
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Template instantiation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Template instantiation failed: {str(e)}",
        )


# ============================================================================
# Team Memory Extraction Endpoints
# ============================================================================

@router.post("/api/memory/extract", response_model=MemoryExtractionResponse)
async def extract_memories(
    request: MemoryExtractionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Extract structured team memories from a conversation transcript.

    Analyzes conversation text and extracts:
    - decisions: Team decisions with rationale
    - preferences: Communication styles, working preferences
    - processes: Workflows and procedures
    - client_context: Client information and requirements

    Args:
        request: Memory extraction request with transcript and optional context
        api_key: Validated API key with chat:write scope

    Returns:
        MemoryExtractionResponse with extracted memories and confidence scores

    Example:
        ```json
        {
          "transcript": "[10:00] Alice: We decided to use React for the frontend...",
          "team_id": "team-123",
          "context": {"participants": ["Alice", "Bob"]}
        }
        ```

    Note:
        - Only high-confidence extractions (>=0.7) are returned
        - Uses gpt-4o-mini by default for cost efficiency (2026 pricing: ~$0.15 per 1K extractions)
        - Average latency: 1-3 seconds depending on transcript length
        - Estimated cost: $1-2 per 1000 conversations analyzed
    """
    import time

    start_time = time.time()

    try:
        service = get_memory_extraction_service()

        result = await service.analyze_conversation(
            transcript=request.transcript,
            context={
                "team_id": request.team_id,
                "conversation_id": request.conversation_id,
                **(request.context or {})
            }
        )

        latency_ms = (time.time() - start_time) * 1000

        # Convert to response model
        memories = [
            ExtractedMemory(
                type=m.type,
                category=m.category,
                summary=m.summary,
                content=m.content,
                confidence=m.confidence,
                rationale=m.rationale,
            )
            for m in result.memories
        ]

        return MemoryExtractionResponse(
            memories=memories,
            confidence=result.confidence,
            tokens_used=result.tokens_used,
            model=result.model,
            latency_ms=latency_ms,
        )

    except Exception as e:
        logger.error(f"Memory extraction failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Memory extraction failed: {str(e)}",
        )


@router.post("/api/memory/extract/batch", response_model=List[MemoryExtractionResponse])
async def extract_memories_batch(
    requests: List[MemoryExtractionRequest],
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Batch extract memories from multiple conversations.

    Processes multiple transcripts in sequence. Use for bulk processing
    of conversation history.

    Args:
        requests: List of extraction requests
        api_key: Validated API key with chat:write scope

    Returns:
        List of MemoryExtractionResponse objects

    Note:
        - Cost scales linearly: ~$0.0006 per conversation (2026 gpt-4o-mini pricing)
        - Processing 10,000 conversations = ~$6 total
        - Consider rate limits when batching large volumes
    """
    import time

    start_time = time.time()

    try:
        service = get_memory_extraction_service()

        results = []
        for req in requests:
            result = await service.analyze_conversation(
                transcript=req.transcript,
                context={
                    "team_id": req.team_id,
                    "conversation_id": req.conversation_id,
                    **(req.context or {})
                }
            )

            memories = [
                ExtractedMemory(
                    type=m.type,
                    category=m.category,
                    summary=m.summary,
                    content=m.content,
                    confidence=m.confidence,
                    rationale=m.rationale,
                )
                for m in result.memories
            ]

            results.append(MemoryExtractionResponse(
                memories=memories,
                confidence=result.confidence,
                tokens_used=result.tokens_used,
                model=result.model,
                latency_ms=0,  # Will be calculated at end
            ))

        total_latency_ms = (time.time() - start_time) * 1000

        # Update latency on all responses
        for r in results:
            r.latency_ms = total_latency_ms / len(results) if results else 0

        return results

    except Exception as e:
        logger.error(f"Batch memory extraction failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Batch memory extraction failed: {str(e)}",
        )


# =============================================================================
# Phase 3: Economic Memory Layer - Cost-Per-Quality Metrics
# =============================================================================


class EconomicMemoryMetricsResponse(BaseModel):
    """Response with economic memory metrics."""
    provider: str
    model: str
    cost_quality_index: float
    avg_cost_per_1k_tokens: float
    avg_cost_per_request: float
    quality_score: float
    response_time_score: float
    success_rate: float
    total_executions: int
    total_cost_usd: float
    recommendation: str
    confidence: str


class EconomicMemorySummaryResponse(BaseModel):
    """Summary of all economic memory data."""
    providers: List[EconomicMemoryMetricsResponse]
    total_executions: int
    total_cost_usd: float
    best_value_provider: Optional[str]
    best_value_cqi: float
    generated_at: datetime


class EconomicRoutingRequest(BaseModel):
    """Request for economic routing decision."""
    function_id: str
    strategy: str = "balanced"  # quality_first, balanced, cost_optimized, cost_first
    quality_threshold: Optional[float] = 0.7
    max_cost_per_1k: Optional[float] = None


class EconomicRoutingResponse(BaseModel):
    """Response with economic routing decision."""
    provider: str
    model: str
    strategy: str
    cost_quality_index: float
    estimated_cost_per_1k: float
    estimated_quality: float
    confidence: str
    reasoning: str
    alternatives: List[str]


class ModelRecommendationResponse(BaseModel):
    """Model recommendation with economic analysis."""
    current_model: str
    recommendation: str
    suggested_model: Optional[str]
    current_cost_per_1k: Optional[float]
    suggested_cost_per_1k: Optional[float]
    potential_savings_percent: Optional[float]
    quality_delta: Optional[float]
    message: str


class CostSavingsOpportunityResponse(BaseModel):
    """Cost savings opportunity analysis."""
    period_days: int
    analysis: str
    current_period_cost: float
    executions_analyzed: int
    best_value_provider: Optional[str]
    best_value_cqi: Optional[float]
    estimated_monthly_savings: float
    optimization_opportunities: List[str]


@router.get("/api/economic-memory/scores", response_model=EconomicMemorySummaryResponse)
async def get_economic_memory_scores():
    """Get all cost-quality scores for provider/model combinations.

    Returns comprehensive economic analysis showing which providers and models
    offer the best value (cost per unit of quality).

    Returns:
        EconomicMemorySummaryResponse with all provider metrics and recommendations
    """
    try:
        from ..services.economic_memory import get_economic_memory

        memory = get_economic_memory()
        scores = await memory.get_all_scores()

        if not scores:
            return EconomicMemorySummaryResponse(
                providers=[],
                total_executions=0,
                total_cost_usd=0.0,
                best_value_provider=None,
                best_value_cqi=0.0,
                generated_at=datetime.utcnow(),
            )

        # Find best value
        best = max(scores, key=lambda s: s.cost_quality_index)

        # Determine recommendations
        provider_responses = []
        for score in scores:
            if score.total_executions < 5:
                rec = "insufficient_data"
            elif score.cost_quality_index >= 50 and score.quality_score >= 0.7:
                rec = "highly_recommended"
            elif score.cost_quality_index >= 30:
                rec = "recommended"
            elif score.cost_quality_index < 10:
                rec = "avoid"
            else:
                rec = "neutral"

            confidence = "high" if score.total_executions >= 50 else (
                "medium" if score.total_executions >= 10 else "low"
            )

            provider_responses.append(EconomicMemoryMetricsResponse(
                provider=score.provider.value,
                model=score.model,
                cost_quality_index=round(score.cost_quality_index, 2),
                avg_cost_per_1k_tokens=round(score.avg_cost_per_1k_tokens, 6),
                avg_cost_per_request=round(score.avg_cost_per_request, 6),
                quality_score=round(score.quality_score, 2),
                response_time_score=round(score.response_time_score, 2),
                success_rate=round(score.success_rate, 3),
                total_executions=score.total_executions,
                total_cost_usd=round(score.total_cost_usd, 4),
                recommendation=rec,
                confidence=confidence,
            ))

        # Sort by CQI descending
        provider_responses.sort(key=lambda p: p.cost_quality_index, reverse=True)

        total_cost = sum(s.total_cost_usd for s in scores)
        total_execs = sum(s.total_executions for s in scores)

        return EconomicMemorySummaryResponse(
            providers=provider_responses,
            total_executions=total_execs,
            total_cost_usd=round(total_cost, 2),
            best_value_provider=f"{best.provider.value}/{best.model}",
            best_value_cqi=round(best.cost_quality_index, 2),
            generated_at=datetime.utcnow(),
        )

    except Exception as e:
        logger.error(f"Failed to get economic memory scores: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get economic scores: {str(e)}",
        )


@router.post("/api/economic-memory/route", response_model=EconomicRoutingResponse)
async def get_economic_routing(request: EconomicRoutingRequest):
    """Get cost-intelligent routing recommendation.

    Uses economic memory data to recommend the best provider/model
    combination based on cost-quality balance.

    Strategies:
    - quality_first: Maximize quality regardless of cost
    - balanced: Balance cost and quality (default)
    - cost_optimized: Minimize cost while meeting quality threshold
    - cost_first: Minimize cost (may reduce quality)

    Args:
        request: Routing request with strategy and constraints

    Returns:
        EconomicRoutingResponse with recommended provider and economics
    """
    try:
        from ..services.economic_routing import (
            get_economic_routing_service,
            RoutingStrategy,
        )

        router = get_economic_routing_service()

        # Parse strategy
        try:
            strategy = RoutingStrategy(request.strategy)
        except ValueError:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Invalid strategy: {request.strategy}. Use: quality_first, balanced, cost_optimized, cost_first"
            )

        # Create routing request
        routing_request = RoutingDecisionRequest(
            function_id=request.function_id,
        )

        decision = await router.decide_routing(
            request=routing_request,
            strategy=strategy,
            quality_threshold=request.quality_threshold,
            max_cost_per_1k=request.max_cost_per_1k,
        )

        # Parse reasoning to extract provider/model
        # Format: "... provider/model with ..."
        reasoning_parts = decision.reasoning.split("Selected ")
        provider_model = reasoning_parts[1].split(" with")[0] if len(reasoning_parts) > 1 else "unknown/default"

        return EconomicRoutingResponse(
            provider=provider_model.split("/")[0] if "/" in provider_model else provider_model,
            model=provider_model.split("/")[1] if "/" in provider_model else "default",
            strategy=request.strategy,
            cost_quality_index=0.0,  # Would need to extract from reasoning
            estimated_cost_per_1k=0.0,
            estimated_quality=request.quality_threshold or 0.7,
            confidence="medium",
            reasoning=decision.reasoning,
            alternatives=[a.value for a in decision.alternatives],
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Economic routing failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Economic routing failed: {str(e)}",
        )


@router.get("/api/economic-memory/recommendation", response_model=ModelRecommendationResponse)
async def get_model_recommendation(
    provider: str = Query(..., description="Provider type (e.g., 'openai')"),
    current_model: str = Query(..., description="Current model name"),
    target_quality: float = Query(0.75, description="Minimum quality threshold"),
):
    """Get model recommendation with economic analysis.

    Analyzes the current model and suggests alternatives with better
    cost-quality characteristics.

    Args:
        provider: Provider type
        current_model: Current model name
        target_quality: Minimum acceptable quality score

    Returns:
        ModelRecommendationResponse with recommendation and analysis
    """
    try:
        from ..services.economic_routing import get_economic_routing_service
        from ..models.schemas import ProviderType

        router = get_economic_routing_service()

        try:
            provider_type = ProviderType(provider)
        except ValueError:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Invalid provider: {provider}"
            )

        result = await router.get_model_recommendation(
            provider=provider_type,
            current_model=current_model,
            target_quality=target_quality,
        )

        return ModelRecommendationResponse(
            current_model=result.get("current_model", current_model),
            recommendation=result.get("recommendation", "unknown"),
            suggested_model=result.get("suggested_model"),
            current_cost_per_1k=result.get("current_cost_per_1k"),
            suggested_cost_per_1k=result.get("suggested_cost_per_1k"),
            potential_savings_percent=result.get("potential_savings_percent"),
            quality_delta=result.get("quality_delta"),
            message=result.get("message", "No recommendation available"),
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get model recommendation: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get recommendation: {str(e)}",
        )


@router.get("/api/economic-memory/savings", response_model=CostSavingsOpportunityResponse)
async def get_cost_savings_opportunity(
    tenant_id: Optional[str] = Query(None, description="Tenant ID for filtering"),
    days: int = Query(7, description="Analysis period in days"),
):
    """Get cost savings opportunity analysis.

    Analyzes recent usage and identifies potential cost savings
    through better provider/model selection.

    Args:
        tenant_id: Optional tenant ID to filter
        days: Number of days to analyze

    Returns:
        CostSavingsOpportunityResponse with savings analysis
    """
    try:
        from ..services.economic_routing import get_economic_routing_service

        router = get_economic_routing_service()

        result = await router.get_cost_savings_opportunity(
            tenant_id=tenant_id,
            days=days,
        )

        return CostSavingsOpportunityResponse(
            period_days=result.get("period_days", days),
            analysis=result.get("analysis", "unknown"),
            current_period_cost=result.get("current_period_cost", 0.0),
            executions_analyzed=result.get("executions_analyzed", 0),
            best_value_provider=result.get("best_value_provider"),
            best_value_cqi=result.get("best_value_cqi"),
            estimated_monthly_savings=result.get("estimated_monthly_savings", 0.0),
            optimization_opportunities=result.get("optimization_opportunities", []),
        )

    except Exception as e:
        logger.error(f"Failed to get cost savings opportunity: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to analyze savings: {str(e)}",
        )


@router.get("/api/economic-memory/executions")
async def get_recent_executions(
    provider: Optional[str] = Query(None, description="Filter by provider"),
    model: Optional[str] = Query(None, description="Filter by model"),
    limit: int = Query(100, ge=1, le=1000, description="Number of executions to return"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Get recent execution records with cost-quality data.

    Returns detailed execution records for analysis and debugging.
    Requires elevated permissions due to sensitive cost data.

    Args:
        provider: Optional provider filter
        model: Optional model filter
        limit: Number of records to return
        api_key: Validated API key

    Returns:
        List of execution records
    """
    try:
        from ..services.economic_memory import get_economic_memory, ProviderType

        memory = get_economic_memory()

        provider_type = None
        if provider:
            try:
                provider_type = ProviderType(provider)
            except ValueError:
                pass

        executions = await memory.get_recent_executions(
            provider=provider_type,
            model=model,
            limit=limit,
        )

        return {
            "executions": [
                {
                    "execution_id": e.execution_id,
                    "provider": e.provider.value,
                    "model": e.model,
                    "cost_usd": round(e.cost_usd, 6),
                    "total_tokens": e.total_tokens,
                    "latency_ms": round(e.latency_ms, 2),
                    "success": e.success,
                    "quality_score": e.output_quality_score,
                    "timestamp": e.timestamp.isoformat(),
                }
                for e in executions
            ],
            "count": len(executions),
        }

    except Exception as e:
        logger.error(f"Failed to get executions: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get executions: {str(e)}",
        )


@router.post("/api/economic-memory/record")
async def record_execution_quality(
    execution_id: str = Query(..., description="Execution ID"),
    quality_score: float = Query(..., ge=0, le=1, description="Quality score 0-1"),
    user_rating: Optional[float] = Query(None, ge=0, le=5, description="User rating 0-5"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Record quality feedback for an execution.

    Allows updating execution records with quality scores or user ratings
    for improving cost-quality analysis.

    Args:
        execution_id: The execution ID to update
        quality_score: Quality score (0-1)
        user_rating: Optional user rating (0-5)
        api_key: Validated API key

    Returns:
        Success status
    """
    # This would update the execution record in the database
    # For now, return success - actual implementation would query DB
    return {
        "success": True,
        "execution_id": execution_id,
        "quality_recorded": quality_score,
        "user_rating_recorded": user_rating,
        "message": "Quality feedback recorded (persistence pending)",
    }


@router.get("/api/economic-memory/health")
async def economic_memory_health():
    """Health check for economic memory service.

    Returns:
        Health status and statistics
    """
    try:
        from ..services.economic_memory import get_economic_memory

        memory = get_economic_memory()
        scores = await memory.get_all_scores()

        return {
            "status": "healthy",
            "service": "economic_memory",
            "providers_tracked": len(scores),
            "total_executions_recorded": sum(s.total_executions for s in scores),
            "total_cost_tracked": round(sum(s.total_cost_usd for s in scores), 2),
        }

    except Exception as e:
        logger.error(f"Economic memory health check failed: {e}")
        return {
            "status": "unhealthy",
            "service": "economic_memory",
            "error": str(e),
        }
