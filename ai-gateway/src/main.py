"""AI Gateway - FastAPI Main Application.

This is the main FastAPI application entry point for the AI Gateway service.
Provides a unified interface for AI inference across RunPod GPU clusters,
ONNX Runtime, and OpenAI-compatible APIs.
"""

import logging
from contextlib import asynccontextmanager
from datetime import datetime

from fastapi import FastAPI, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from .config import get_settings, Settings
from .routers import inference_router
from .services.inference_engine import get_inference_engine
from .services.cluster_client import get_cluster_client
from .models.schemas import HealthResponse, HealthStatus, ReadinessResponse

# Configure logging
_log_level = getattr(logging, get_settings().LOG_LEVEL.upper(), logging.INFO)
if not isinstance(_log_level, int):
    _log_level = logging.INFO

logging.basicConfig(
    level=_log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Version info
SERVICE_VERSION = "0.1.0"
SERVICE_NAME = "ai-gateway"


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup and shutdown events."""
    settings = get_settings()
    logger.info(f"Starting {SERVICE_NAME} v{SERVICE_VERSION}")

    # Initialize services
    try:
        engine = get_inference_engine()
        await engine.initialize()
        logger.info("Inference engine initialized")
    except Exception as e:
        logger.error(f"Failed to initialize inference engine: {e}")

    try:
        cluster_client = get_cluster_client()
        logger.info("Cluster client initialized")
    except Exception as e:
        logger.warning(f"Failed to initialize cluster client: {e}")

    yield

    # Shutdown
    logger.info("Shutting down...")

    try:
        engine = get_inference_engine()
        await engine.shutdown()
        logger.info("Inference engine shutdown complete")
    except Exception as e:
        logger.warning(f"Error during inference engine shutdown: {e}")

    try:
        cluster_client = get_cluster_client()
        await cluster_client.close()
        logger.info("Cluster client shutdown complete")
    except Exception as e:
        logger.warning(f"Error during cluster client shutdown: {e}")

    logger.info("Shutdown complete")


def create_app() -> FastAPI:
    """Create and configure FastAPI application.

    Returns:
        Configured FastAPI application
    """
    settings = get_settings()

    app = FastAPI(
        title="AI Gateway",
        description=(
            "Unified AI inference gateway for RunPod GPU clusters, "
            "ONNX Runtime, and OpenAI-compatible APIs."
        ),
        version=SERVICE_VERSION,
        docs_url="/docs",
        redoc_url="/redoc",
        openapi_url="/openapi.json",
        lifespan=lifespan,
    )

    # Add CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.CORS_ORIGINS,
        allow_credentials=settings.CORS_ALLOW_CREDENTIALS,
        allow_methods=settings.CORS_ALLOW_METHODS,
        allow_headers=settings.CORS_ALLOW_HEADERS,
    )

    # Include routers
    app.include_router(inference_router)

    # Exception handlers
    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(
        request: Request, exc: RequestValidationError
    ):
        """Handle validation errors."""
        return JSONResponse(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            content={
                "error": "validation_error",
                "message": "Request validation failed",
                "details": exc.errors(),
                "timestamp": datetime.utcnow().isoformat(),
            },
        )

    @app.exception_handler(Exception)
    async def general_exception_handler(request: Request, exc: Exception):
        """Handle general exceptions."""
        logger.exception(f"Unhandled exception: {exc}")
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content={
                "error": "internal_error",
                "message": "An internal error occurred",
                "timestamp": datetime.utcnow().isoformat(),
            },
        )

    # Root endpoints
    @app.get("/", include_in_schema=False)
    async def root():
        """Root endpoint."""
        return JSONResponse({
            "service": SERVICE_NAME,
            "version": SERVICE_VERSION,
            "status": "running",
            "docs": "/docs",
        })

    @app.get(
        "/health",
        response_model=HealthResponse,
        tags=["health"],
        summary="Liveness probe",
        description="Check if the service is alive and running.",
    )
    async def health():
        """Liveness probe endpoint."""
        return HealthResponse(
            status=HealthStatus.HEALTHY,
            version=SERVICE_VERSION,
            timestamp=datetime.utcnow(),
        )

    @app.get(
        "/ready",
        response_model=ReadinessResponse,
        tags=["health"],
        summary="Readiness probe",
        description="Check if the service is ready to accept traffic.",
    )
    async def ready():
        """Readiness probe endpoint."""
        from .services.cluster_client import get_cluster_client

        cluster_client = get_cluster_client()
        clusters_healthy = await cluster_client.health_check()

        # Get cluster health info
        cluster_info = []
        if clusters_healthy:
            try:
                cluster_info = await cluster_client.get_cluster_health_info()
            except Exception as e:
                logger.warning(f"Failed to get cluster health info: {e}")

        healthy_count = sum(
            1 for c in cluster_info if c.status == HealthStatus.HEALTHY
        )

        return ReadinessResponse(
            status=HealthStatus.READY if healthy_count > 0 else HealthStatus.DEGRADED,
            clusters=cluster_info,
            total_clusters=len(cluster_info),
            healthy_clusters=healthy_count,
            version=SERVICE_VERSION,
            timestamp=datetime.utcnow(),
        )

    @app.get(
        "/metrics",
        tags=["metrics"],
        summary="Prometheus metrics",
        description="Get Prometheus-formatted metrics for monitoring.",
    )
    async def metrics():
        """Prometheus metrics endpoint."""
        from .services.inference_engine import get_inference_engine

        engine = get_inference_engine()

        # Build metrics output
        metrics_lines = [
            "# HELP ai_gateway_info AI Gateway service information",
            "# TYPE ai_gateway_info gauge",
            f'ai_gateway_info{{version="{SERVICE_VERSION}"}} 1',
            "",
            "# HELP ai_gateway_up AI Gateway service availability",
            "# TYPE ai_gateway_up gauge",
            "ai_gateway_up 1",
            "",
            "# HELP ai_gateway_inference_requests_total Total inference requests",
            "# TYPE ai_gateway_inference_requests_total counter",
            "ai_gateway_inference_requests_total 0",
            "",
            "# HELP ai_gateway_inference_latency_ms Inference latency histogram",
            "# TYPE ai_gateway_inference_latency_ms histogram",
            'ai_gateway_inference_latency_ms_bucket{le="50"} 0',
            'ai_gateway_inference_latency_ms_bucket{le="100"} 0',
            'ai_gateway_inference_latency_ms_bucket{le="200"} 0',
            'ai_gateway_inference_latency_ms_bucket{le="500"} 0',
            'ai_gateway_inference_latency_ms_bucket{le="1000"} 0',
            'ai_gateway_inference_latency_ms_bucket{le="+Inf"} 0',
            "# TYPE ai_gateway_inference_latency_ms histogram",
            "",
            "# HELP ai_gateway_inference_cost_usd_total Total inference cost in USD",
            "# TYPE ai_gateway_inference_cost_usd_total counter",
            "ai_gateway_inference_cost_usd_total 0",
            "",
            "# HELP ai_gateway_rate_limit_remaining Rate limit tokens remaining",
            "# TYPE ai_gateway_rate_limit_remaining gauge",
            f'ai_gateway_rate_limit_remaining{{tenant_id="default"}} {settings.RATE_LIMIT_REQUESTS}',
        ]

        # Add rate limit metrics for configured tenants
        rate_limit_status = engine.get_rate_limit_status("default")
        metrics_lines.append(
            f'ai_gateway_rate_limit_remaining{{tenant_id="default"}} {rate_limit_status["tokens_remaining"]}'
        )

        return JSONResponse(
            content="\n".join(metrics_lines),
            media_type="text/plain; charset=utf-8",
        )

    return app


# Create app instance
app = create_app()


if __name__ == "__main__":
    import uvicorn

    settings = get_settings()
    uvicorn.run(
        "main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
    )
