"""API routes for FlyMind AI Service.

This module aggregates all domain-specific routers and defines
core endpoints (health, providers, embed, complete, stream).

Domain routers:
  - routes_chat:       /api/chat/*
  - routes_search:     /api/search/*
  - routes_debug:      /api/debug/*
  - routes_optimize:   /api/optimize/*
  - routes_flyembed:   /api/flyembed/*
  - routes_composer:   /api/composer/*, /api/ai/composer/*
  - routes_routing:    /api/route/*, /api/prewarm/*, /api/anomalies/*
  - routes_composition: /api/composition/*
  - routes_dna:        /api/dna/*
"""

import asyncio
import json
import logging
import time
from datetime import datetime
from typing import AsyncGenerator, Optional, List

from fastapi import APIRouter, HTTPException, Query, status, Depends
from fastapi.responses import StreamingResponse
import redis.asyncio as redis
from pydantic import BaseModel

from .routes_chat import router as chat_router
from .routes_search import router as search_router
from .routes_debug import router as debug_router
from .routes_optimize import router as optimize_router
from .routes_flyembed import router as flyembed_router
from .routes_composer import router as composer_router
from .routes_routing import router as routing_router
from .routes_composition import router as composition_router
from .routes_dna import router as dna_router
from .routes_ml import router as ml_router

from ..security.auth import (
    require_api_key,
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)

from ..config import settings
from ..utils.security import sanitize_error_message
from ..models.schemas import (
    ChatMessage,
    CompletionRequest,
    CompletionResponse,
    EmbeddingRequest,
    EmbeddingResponse,
    ErrorResponse,
    ProviderStatusResponse,
    HealthResponse,
    # Memory Extraction
    MemoryExtractionRequest,
    MemoryExtractionResponse,
    ExtractedMemory,
    # Future AI Expansion Placeholders
    AIChatRequest,
    AISuggestRequest,
    AIPlaceholderResponse,
)
from ..providers.manager import get_provider_manager
from ..services.embeddings import get_embeddings_service
from ..services.memory_extraction import get_memory_extraction_service

logger = logging.getLogger(__name__)

router = APIRouter()

# =============================================================================
# Include domain-specific routers
# =============================================================================

router.include_router(chat_router)
router.include_router(search_router)
router.include_router(debug_router)
router.include_router(optimize_router)
router.include_router(flyembed_router)
router.include_router(composer_router)
router.include_router(routing_router)
router.include_router(composition_router)
router.include_router(dna_router)
router.include_router(ml_router)


# =============================================================================
# Core endpoints: health, providers, embed, complete, stream
# =============================================================================


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


@router.get("/api/health", response_model=HealthResponse)
async def api_health_check():
    """Health check endpoint at /api/health path.

    Returns:
        Health status of the service and its dependencies
    """
    return await health_check()


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
async def create_completion(
    request: CompletionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Generate a completion using the specified provider.

    Args:
        request: Completion request with messages and optional provider/model
        api_key: Validated API key with chat:write scope

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
async def stream_completion(
    request: CompletionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Stream a completion using the specified provider.

    Args:
        request: Completion request with messages and optional provider/model
        api_key: Validated API key with chat:write scope

    Returns:
        Streaming response with generated content
    """
    if not settings.enable_streaming:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail="Streaming is not enabled",
        )

    if not request.stream:
        return await create_completion(request, api_key)

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
                    escaped = chunk.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r')
                    yield f'{{"type": "chunk", "content": "{escaped}"}}\n\n'
                yield '{"type": "complete"}\n\n'
            except Exception as e:
                logger.error(f"Streaming failed: {e}")
                error_json = json.dumps({"type": "error", "error": str(e)})
                yield f'{error_json}\n\n'

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
# Team Memory Extraction Endpoints
# =============================================================================


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
    """
    start_time = time.time()

    try:
        service = get_memory_extraction_service()

        result = await service.analyze_conversation(
            transcript=request.transcript,
            context={
                "team_id": request.team_id,
                "conversation_id": request.conversation_id,
                **(request.context or {}),
            },
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
            detail=sanitize_error_message(e, include_details=settings.debug),
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
    """
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
                    **(req.context or {}),
                },
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

            results.append(
                MemoryExtractionResponse(
                    memories=memories,
                    confidence=result.confidence,
                    tokens_used=result.tokens_used,
                    model=result.model,
                    latency_ms=0,  # Will be calculated at end
                )
            )

        total_latency_ms = (time.time() - start_time) * 1000

        # Update latency on all responses
        for r in results:
            r.latency_ms = total_latency_ms / len(results) if results else 0

        return results

    except Exception as e:
        logger.error(f"Batch memory extraction failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


# =============================================================================
# Economic Memory Layer - Cost-Per-Quality Metrics
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
async def get_economic_memory_scores(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Get all cost-quality scores for provider/model combinations.

    Returns comprehensive economic analysis showing which providers and models
    offer the best value (cost per unit of quality).

    Args:
        api_key: Validated API key with chat:read scope

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

            confidence = (
                "high"
                if score.total_executions >= 50
                else ("medium" if score.total_executions >= 10 else "low")
            )

            provider_responses.append(
                EconomicMemoryMetricsResponse(
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
                )
            )

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
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.post("/api/economic-memory/route", response_model=EconomicRoutingResponse)
async def get_economic_routing(
    request: EconomicRoutingRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
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
        api_key: Validated API key with chat:write scope

    Returns:
        EconomicRoutingResponse with recommended provider and economics
    """
    try:
        from ..services.economic_routing import (
            get_economic_routing_service,
            RoutingStrategy,
        )

        svc = get_economic_routing_service()

        # Parse strategy
        try:
            strategy = RoutingStrategy(request.strategy)
        except ValueError:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Invalid strategy: {request.strategy}. Use: quality_first, balanced, cost_optimized, cost_first",
            )

        # Create routing request
        routing_request = RoutingDecisionRequest(
            function_id=request.function_id,
        )

        decision = await svc.decide_routing(
            request=routing_request,
            strategy=strategy,
            quality_threshold=request.quality_threshold,
            max_cost_per_1k=request.max_cost_per_1k,
        )

        # Parse reasoning to extract provider/model
        reasoning_parts = decision.reasoning.split("Selected ")
        provider_model = (
            reasoning_parts[1].split(" with")[0] if len(reasoning_parts) > 1 else "unknown/default"
        )

        return EconomicRoutingResponse(
            provider=provider_model.split("/")[0] if "/" in provider_model else provider_model,
            model=provider_model.split("/")[1] if "/" in provider_model else "default",
            strategy=request.strategy,
            cost_quality_index=0.0,
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
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.get("/api/economic-memory/recommendation", response_model=ModelRecommendationResponse)
async def get_model_recommendation(
    provider: str = Query(..., description="Provider type (e.g., 'openai')"),
    current_model: str = Query(..., description="Current model name"),
    target_quality: float = Query(0.75, description="Minimum quality threshold"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Get model recommendation with economic analysis.

    Args:
        provider: Provider type
        current_model: Current model name
        target_quality: Minimum quality threshold
        api_key: Validated API key with chat:read scope

    Returns:
        ModelRecommendationResponse with recommendation
    """
    try:
        from ..services.economic_routing import get_economic_routing_service
        from ..models.schemas import ProviderType

        svc = get_economic_routing_service()

        try:
            provider_type = ProviderType(provider)
        except ValueError:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST, detail=f"Invalid provider: {provider}"
            )

        result = await svc.get_model_recommendation(
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
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.get("/api/economic-memory/savings", response_model=CostSavingsOpportunityResponse)
async def get_cost_savings_opportunity(
    tenant_id: Optional[str] = Query(None, description="Tenant ID for filtering"),
    days: int = Query(7, description="Analysis period in days"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Get cost savings opportunity analysis.

    Args:
        tenant_id: Optional tenant ID for filtering
        days: Analysis period in days
        api_key: Validated API key with chat:read scope

    Returns:
        CostSavingsOpportunityResponse with savings analysis
    """
    try:
        from ..services.economic_routing import get_economic_routing_service

        svc = get_economic_routing_service()

        result = await svc.get_cost_savings_opportunity(
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
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.get("/api/economic-memory/executions")
async def get_recent_executions(
    provider: Optional[str] = Query(None, description="Filter by provider"),
    model: Optional[str] = Query(None, description="Filter by model"),
    limit: int = Query(100, ge=1, le=1000, description="Number of executions to return"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Get recent execution records with cost-quality data."""
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
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.post("/api/economic-memory/record")
async def record_execution_quality(
    execution_id: str = Query(..., description="Execution ID"),
    quality_score: float = Query(..., ge=0, le=1, description="Quality score 0-1"),
    user_rating: Optional[float] = Query(None, ge=0, le=5, description="User rating 0-5"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Record quality feedback for an execution."""
    return {
        "success": True,
        "execution_id": execution_id,
        "quality_recorded": quality_score,
        "user_rating_recorded": user_rating,
        "message": "Quality feedback recorded (persistence pending)",
    }


@router.get("/api/economic-memory/health")
async def economic_memory_health(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Health check for economic memory service.

    Args:
        api_key: Validated API key with chat:read scope

    Returns:
        Health status of economic memory service
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


# =============================================================================
# Future AI Expansion - Placeholder Endpoints
# =============================================================================


# -----------------------------------------------------------------------------
# Future: /ai/chat - Conversational AI Interface
# -----------------------------------------------------------------------------


@router.post("/api/ai/chat/message", response_model=AIPlaceholderResponse)
async def ai_chat_message(
    request: AIChatRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Conversational AI interface - RESERVED FOR FUTURE."""
    return AIPlaceholderResponse(
        feature="ai_chat",
        status="coming_soon",
        message="Conversational AI interface is coming soon! Use /api/chat/message for current chat functionality.",
        estimated_release="Q3 2026",
    )


@router.get("/api/ai/chat/sessions")
async def ai_chat_list_sessions_placeholder(
    user_id: str = Query(..., description="User ID"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """List AI chat sessions - RESERVED FOR FUTURE."""
    return AIPlaceholderResponse(
        feature="ai_chat_sessions",
        status="coming_soon",
        message="AI chat sessions API coming soon! Use /api/chat/sessions for now.",
    )


# -----------------------------------------------------------------------------
# Future: /ai/suggest - Code Suggestion/Intellisense
# -----------------------------------------------------------------------------


@router.post("/api/ai/suggest/completions", response_model=AIPlaceholderResponse)
async def ai_suggest_completions(
    request: AISuggestRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """AI-powered code suggestions - RESERVED FOR FUTURE."""
    return AIPlaceholderResponse(
        feature="ai_suggest",
        status="coming_soon",
        message="AI code suggestions coming soon! This will provide intelligent intellisense for function development.",
        estimated_release="Q4 2026",
    )


@router.post("/api/ai/suggest/fixes", response_model=AIPlaceholderResponse)
async def ai_suggest_fixes(
    request: AISuggestRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """AI-powered error fixes - RESERVED FOR FUTURE."""
    return AIPlaceholderResponse(
        feature="ai_suggest_fixes",
        status="coming_soon",
        message="AI fix suggestions coming soon! Use /api/debug/suggest for debugging help now.",
        estimated_release="Q4 2026",
    )


@router.get("/api/ai/suggest/status")
async def ai_suggest_status():
    """Status of AI suggestion service."""
    return {
        "service": "ai_suggest",
        "status": "planned",
        "features": [
            "smart_completions",
            "error_detection",
            "optimization_hints",
            "pattern_matching",
        ],
        "estimated_release": "Q4 2026",
        "message": "AI suggestion service is planned. Stay tuned for updates!",
    }


# -----------------------------------------------------------------------------
# AI Namespace Health & Discovery
# -----------------------------------------------------------------------------


@router.get("/api/ai/status")
async def ai_namespace_status():
    """Get status of all AI namespace features."""
    return {
        "namespace": "ai",
        "version": "1.0.0",
        "features": {
            "composer": {
                "path": "/api/ai/composer/*",
                "status": "active",
                "description": "AI function generation",
                "endpoints": [
                    "POST /api/ai/composer/generate",
                    "POST /api/ai/composer/generate/stream",
                    "POST /api/ai/composer/generate-optimized",
                ],
            },
            "chat": {
                "path": "/api/ai/chat/*",
                "status": "coming_soon",
                "description": "Conversational AI interface",
                "estimated_release": "Q3 2026",
                "endpoints": [
                    "POST /api/ai/chat/message",
                    "GET /api/ai/chat/sessions",
                ],
            },
            "suggest": {
                "path": "/api/ai/suggest/*",
                "status": "planned",
                "description": "AI code suggestions and intellisense",
                "estimated_release": "Q4 2026",
                "endpoints": [
                    "POST /api/ai/suggest/completions",
                    "POST /api/ai/suggest/fixes",
                    "GET /api/ai/suggest/status",
                ],
            },
        },
        "message": "AI namespace is active. Use /api/ai/composer/* for current features.",
    }


@router.get("/api/models/catalog")
async def get_model_catalog():
    """Return the full model catalog with provider metadata."""
    from ..providers.model_registry import CURATED_MODELS

    return {"models": CURATED_MODELS}
