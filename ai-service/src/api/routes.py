"""API routes for FlyMind AI Service.

This module defines all the API endpoints for Phase 1 (Foundation)
and Phase 2 (Intelligence Layer).
"""

import asyncio
import logging
from datetime import datetime
from typing import AsyncGenerator, Optional, List

from fastapi import APIRouter, HTTPException, Query, status, Body
from fastapi.responses import StreamingResponse
import redis.asyncio as redis
from pydantic import BaseModel

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
from ..models.schemas import (
    TripleEmbeddingRequest,
    TripleEmbeddingResult,
    TripleEmbeddingBatchRequest,
    TripleEmbeddingBatchResponse,
    TripleQueryRequest,
    TripleQueryVector,
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
async def create_embedding(request: EmbeddingRequest):
    """Generate embeddings for the given text.

    Args:
        request: Embedding request with text and optional provider/model

    Returns:
        Embedding response with vector
    """
    try:
        embeddings_service = get_embeddings_service()
        response = await embeddings_service.generate_embedding(request)
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
async def flyembed_embed(request: TripleEmbeddingRequest):
    """Generate triple embeddings for a single function.

    Args:
        request: Triple embedding request with function data

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
async def flyembed_embed_batch(request: TripleEmbeddingBatchRequest):
    """Batch generate triple embeddings for multiple functions.

    Args:
        request: Batch of function data

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
async def flyembed_query(request: TripleQueryRequest):
    """Generate triple query vectors for search.

    Args:
        request: Query request with search text

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
