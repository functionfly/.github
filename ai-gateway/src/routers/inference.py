"""Inference Router for AI Gateway.

Provides the /infer endpoint for running AI inference requests.
"""

import asyncio
import logging
import time
import uuid
from typing import Any, Dict, Optional

from fastapi import APIRouter, Depends, Header, HTTPException, Request, status
from fastapi.responses import JSONResponse, StreamingResponse

from ..config import Settings, get_settings
from ..models.schemas import (
    BatchInferenceRequest,
    BatchInferenceResponse,
    ErrorResponse,
    InferenceRequest,
    InferenceResponse,
    ModelListResponse,
    ModelInfo,
    InferenceBackend,
    ModelProvider,
)
from ..services.inference_engine import get_inference_engine, InferenceEngine

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/v1", tags=["inference"])


# Supported models registry
SUPPORTED_MODELS = [
    ModelInfo(
        model_id="onnx://phi-3-mini",
        provider=ModelProvider.ONNX,
        backend=InferenceBackend.ONNX_RUNTIME,
        context_length=4096,
        supported_parameters=["temperature", "max_tokens", "top_p", "top_k", "repeat_penalty"],
        is_available=True,
    ),
    ModelInfo(
        model_id="openai://gpt-4",
        provider=ModelProvider.OPENAI,
        backend=InferenceBackend.OPENAI_API,
        context_length=8192,
        supported_parameters=["temperature", "max_tokens", "top_p"],
        is_available=True,
    ),
    ModelInfo(
        model_id="openai://gpt-3.5-turbo",
        provider=ModelProvider.OPENAI,
        backend=InferenceBackend.OPENAI_API,
        context_length=4096,
        supported_parameters=["temperature", "max_tokens", "top_p"],
        is_available=True,
    ),
    ModelInfo(
        model_id="runpod://phi-3-mini",
        provider=ModelProvider.RUNPOD,
        backend=InferenceBackend.RUNPOD_API,
        context_length=4096,
        supported_parameters=["temperature", "max_tokens", "top_p", "top_k", "repeat_penalty"],
        is_available=True,
    ),
]


async def get_inference_engine_dep() -> InferenceEngine:
    """Dependency to get inference engine."""
    return get_inference_engine()


async def verify_api_key(
    x_api_key: Optional[str] = Header(None, alias="X-API-Key"),
    settings: Settings = Depends(get_settings),
) -> Optional[str]:
    """Verify API key if required.

    Args:
        x_api_key: API key from header
        settings: Application settings

    Returns:
        Validated API key or None

    Raises:
        HTTPException: If API key is invalid
    """
    if settings.REQUIRED_API_KEY:
        if not x_api_key:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="API key required",
            )
        if x_api_key != settings.REQUIRED_API_KEY:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid API key",
            )
    return x_api_key


def get_tenant_id(
    x_tenant_id: Optional[str] = Header(None, alias="X-Tenant-ID"),
) -> Optional[str]:
    """Get tenant ID from header.

    Args:
        x_tenant_id: Tenant ID from header

    Returns:
        Tenant ID or None
    """
    return x_tenant_id


@router.post(
    "/infer",
    response_model=InferenceResponse,
    responses={
        400: {"model": ErrorResponse},
        401: {"model": ErrorResponse},
        429: {"model": ErrorResponse},
        500: {"model": ErrorResponse},
        503: {"model": ErrorResponse},
    },
    summary="Run inference",
    description="Run AI inference request with the specified model and parameters.",
)
async def infer(
    request: InferenceRequest,
    engine: InferenceEngine = Depends(get_inference_engine_dep),
    api_key: Optional[str] = Depends(verify_api_key),
    tenant_id: Optional[str] = Depends(get_tenant_id),
) -> InferenceResponse:
    """Run inference request.

    Args:
        request: Inference request with model, input, and parameters
        engine: Inference engine
        api_key: Validated API key
        tenant_id: Tenant ID for multi-tenant isolation

    Returns:
        Inference response with output, latency, and cost

    Raises:
        HTTPException: On inference error
    """
    # Override tenant_id if provided in request
    if request.tenant_id:
        tenant_id = request.tenant_id

    request_id = str(uuid.uuid4())
    logger.info(
        f" Inference request {request_id}: model={request.model}, "
        f"tenant={tenant_id}, stream={request.stream}"
    )

    try:
        response = await engine.infer(request)
        logger.info(
            f"Inference {request_id} completed: latency={response.latency_ms:.2f}ms, "
            f"cost=${response.cost_usd:.6f}"
        )
        return response

    except RuntimeError as e:
        error_msg = str(e)
        logger.error(f"Inference {request_id} failed: {error_msg}")

        if "circuit breaker open" in error_msg.lower():
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail=ErrorResponse(
                    error="circuit_breaker_open",
                    message="Backend temporarily unavailable",
                    request_id=request_id,
                ).model_dump(),
            )
        elif "rate limit" in error_msg.lower():
            raise HTTPException(
                status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                detail=ErrorResponse(
                    error="rate_limit_exceeded",
                    message=error_msg,
                    request_id=request_id,
                ).model_dump(),
            )
        else:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=ErrorResponse(
                    error="inference_failed",
                    message=error_msg,
                    request_id=request_id,
                ).model_dump(),
            )


@router.post(
    "/infer/stream",
    response_class=StreamingResponse,
    responses={
        400: {"model": ErrorResponse},
        401: {"model": ErrorResponse},
        429: {"model": ErrorResponse},
        500: {"model": ErrorResponse},
        503: {"model": ErrorResponse},
    },
    summary="Run streaming inference",
    description="Run streaming AI inference request for real-time output.",
)
async def infer_stream(
    request: InferenceRequest,
    engine: InferenceEngine = Depends(get_inference_engine_dep),
    api_key: Optional[str] = Depends(verify_api_key),
    tenant_id: Optional[str] = Depends(get_tenant_id),
) -> StreamingResponse:
    """Run streaming inference request.

    Args:
        request: Inference request
        engine: Inference engine
        api_key: Validated API key
        tenant_id: Tenant ID

    Returns:
        Streaming response
    """
    # Ensure stream is True
    request.stream = True

    if request.tenant_id:
        tenant_id = request.tenant_id

    request_id = str(uuid.uuid4())
    logger.info(
        f"Streaming inference request {request_id}: model={request.model}, "
        f"tenant={tenant_id}"
    )

    async def generate():
        """Generate streaming response."""
        try:
            # For streaming, we yield chunks as they come
            # In real implementation, this would stream from the backend
            response = await engine.infer(request)

            # Send response as SSE
            yield f"data: {response.model_dump_json()}\n\n"
            yield "data: [DONE]\n\n"

        except Exception as e:
            logger.error(f"Streaming inference {request_id} failed: {e}")
            error_response = ErrorResponse(
                error="inference_failed",
                message=str(e),
                request_id=request_id,
            )
            yield f"data: {error_response.model_dump_json()}\n\n"

    return StreamingResponse(
        generate(),
        media_type="text/event-stream",
        headers={
            "X-Request-ID": request_id,
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
        },
    )


@router.post(
    "/infer/batch",
    response_model=BatchInferenceResponse,
    responses={
        400: {"model": ErrorResponse},
        401: {"model": ErrorResponse},
        429: {"model": ErrorResponse},
        500: {"model": ErrorResponse},
    },
    summary="Run batch inference",
    description="Run batch inference with multiple requests (max 8).",
)
async def infer_batch(
    request: BatchInferenceRequest,
    engine: InferenceEngine = Depends(get_inference_engine_dep),
    api_key: Optional[str] = Depends(verify_api_key),
    tenant_id: Optional[str] = Depends(get_tenant_id),
) -> BatchInferenceResponse:
    """Run batch inference request.

    Args:
        request: Batch inference request
        engine: Inference engine
        api_key: Validated API key
        tenant_id: Tenant ID

    Returns:
        Batch inference response
    """
    request_id = str(uuid.uuid4())
    logger.info(
        f"Batch inference request {request_id}: size={len(request.requests)}"
    )

    start_time = time.time()
    outputs: list[InferenceResponse] = []
    errors: list[str] = []

    # Process all requests
    for i, req in enumerate(request.requests):
        try:
            if tenant_id:
                req.tenant_id = tenant_id
            response = await engine.infer(req)
            outputs.append(response)
        except Exception as e:
            errors.append(f"Request {i}: {str(e)}")
            logger.error(f"Batch request {i} failed: {e}")

    total_latency_ms = (time.time() - start_time) * 1000
    total_cost_usd = sum(o.cost_usd for o in outputs)

    if errors and not outputs:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=ErrorResponse(
                error="batch_inference_failed",
                message=f"All batch requests failed: {errors}",
                request_id=request_id,
            ).model_dump(),
        )

    logger.info(
        f"Batch inference {request_id} completed: "
        f"size={len(outputs)}/{len(request.requests)}, "
        f"latency={total_latency_ms:.2f}ms, cost=${total_cost_usd:.6f}"
    )

    return BatchInferenceResponse(
        outputs=outputs,
        total_latency_ms=total_latency_ms,
        total_cost_usd=total_cost_usd,
        batch_size=len(outputs),
    )


@router.get(
    "/models",
    response_model=ModelListResponse,
    summary="List available models",
    description="Get list of available models and their configurations.",
)
async def list_models(
    api_key: Optional[str] = Depends(verify_api_key),
) -> ModelListResponse:
    """List available models.

    Args:
        api_key: Validated API key

    Returns:
        List of available models
    """
    return ModelListResponse(
        models=SUPPORTED_MODELS,
        default_model="phi-3-mini",
    )


@router.get(
    "/models/{model_id}",
    response_model=ModelInfo,
    summary="Get model information",
    description="Get information about a specific model.",
)
async def get_model(
    model_id: str,
    api_key: Optional[str] = Depends(verify_api_key),
) -> ModelInfo:
    """Get model information.

    Args:
        model_id: Model identifier
        api_key: Validated API key

    Returns:
        Model information

    Raises:
        HTTPException: If model not found
    """
    for model in SUPPORTED_MODELS:
        if model.model_id == model_id:
            return model

    raise HTTPException(
        status_code=status.HTTP_404_NOT_FOUND,
        detail=ErrorResponse(
            error="model_not_found",
            message=f"Model not found: {model_id}",
        ).model_dump(),
    )


@router.get(
    "/health",
    summary="Health check",
    description="Check if the inference service is healthy.",
)
async def health_check() -> Dict[str, Any]:
    """Health check endpoint.

    Returns:
        Health status
    """
    return {
        "status": "healthy",
        "service": "ai-gateway",
        "version": "0.1.0",
        "timestamp": time.time(),
    }


@router.get(
    "/ready",
    summary="Readiness check",
    description="Check if the inference service is ready to accept requests.",
)
async def readiness_check(
    engine: InferenceEngine = Depends(get_inference_engine_dep),
) -> Dict[str, Any]:
    """Readiness check endpoint.

    Args:
        engine: Inference engine

    Returns:
        Readiness status
    """
    from ..services.cluster_client import get_cluster_client

    cluster_client = get_cluster_client()
    clusters_healthy = await cluster_client.health_check()

    # Check backends
    backends_status = {}
    for backend in InferenceBackend:
        health = await engine.get_backend_health(backend)
        backends_status[backend.value] = "healthy" if health else "unhealthy"

    is_ready = clusters_healthy or any(
        v == "healthy" for v in backends_status.values()
    )

    return {
        "status": "ready" if is_ready else "not_ready",
        "clusters_reachable": clusters_healthy,
        "backends": backends_status,
        "timestamp": time.time(),
    }


@router.get(
    "/metrics",
    summary="Prometheus metrics",
    description="Get Prometheus-formatted metrics.",
)
async def metrics(
    engine: InferenceEngine = Depends(get_inference_engine_dep),
) -> JSONResponse:
    """Get Prometheus metrics.

    Args:
        engine: Inference engine

    Returns:
        Prometheus-formatted metrics
    """
    from ..services.model_cache import get_model_cache

    # Gather metrics
    metrics_lines = [
        "# HELP ai_gateway_inference_requests_total Total inference requests",
        "# TYPE ai_gateway_inference_requests_total counter",
        "ai_gateway_inference_requests_total 0",
        "",
        "# HELP ai_gateway_inference_latency_ms Inference latency in milliseconds",
        "# TYPE ai_gateway_inference_latency_ms histogram",
        "ai_gateway_inference_latency_ms_bucket{le=\"100\"} 0",
        "ai_gateway_inference_latency_ms_bucket{le=\"500\"} 0",
        "ai_gateway_inference_latency_ms_bucket{le=\"1000\"} 0",
        "ai_gateway_inference_latency_ms_bucket{le=\"+Inf\"} 0",
        "",
        "# HELP ai_gateway_inference_cost_usd Inference cost in USD",
        "# TYPE ai_gateway_inference_cost_usd counter",
        "ai_gateway_inference_cost_usd 0",
        "",
        "# HELP ai_gateway_model_cache_size_bytes Model cache size in bytes",
        "# TYPE ai_gateway_model_cache_size_bytes gauge",
    ]

    try:
        model_cache = get_model_cache()
        cache_stats = await model_cache.get_cache_stats()
        metrics_lines.append(
            f"ai_gateway_model_cache_size_bytes {cache_stats['current_size_mb'] * 1024 * 1024}"
        )
        metrics_lines.append(
            f"ai_gateway_cached_models {cache_stats['cached_models']}"
        )
    except Exception:
        metrics_lines.append("ai_gateway_model_cache_size_bytes 0")
        metrics_lines.append("ai_gateway_cached_models 0")

    metrics_lines.append("")
    metrics_lines.append("# HELP ai_gateway_rate_limit_remaining Rate limit tokens remaining")
    metrics_lines.append("# TYPE ai_gateway_rate_limit_remaining gauge")
    for tenant_id in ["default"]:
        status = engine.get_rate_limit_status(tenant_id)
        metrics_lines.append(
            f'ai_gateway_rate_limit_remaining{{tenant_id="{tenant_id}"}} {status["tokens_remaining"]}'
        )

    return JSONResponse(
        content="\n".join(metrics_lines),
        media_type="text/plain; charset=utf-8",
    )
