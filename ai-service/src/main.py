"""FlyMind AI Service - Main Application.

This is the FastAPI application entry point.
Includes Phase 1 (Foundation) and Phase 2 (Intelligence Layer).
"""

import logging
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from .config import settings, get_settings
from .api.routes import router
from .providers.manager import get_provider_manager
from .services.embeddings import get_embeddings_service
from .middleware.security_headers import SecurityHeadersMiddleware


# Configure logging (uppercase so "info" maps to logging.INFO, not logging.info)
_log_level = getattr(logging, settings.log_level.upper(), logging.INFO)
if not isinstance(_log_level, int):
    _log_level = logging.INFO
logging.basicConfig(
    level=_log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


def _init_sentry() -> None:
    """Initialize Sentry error tracking if SENTRY_DSN is configured.

    Sentry provides distributed tracing and error tracking for production.
    Only initialized if the SENTRY_DSN environment variable is set.
    """
    sentry_dsn = os.getenv("SENTRY_DSN")
    if not sentry_dsn:
        logger.info("Sentry not configured (SENTRY_DSN not set)")
        return

    try:
        import sentry_sdk
        from sentry_sdk.integrations.fastapi import FastApiIntegration

        sentry_sdk.init(
            dsn=sentry_dsn,
            integrations=[FastApiIntegration()],
            environment=os.getenv("FLY_ENV", "production"),
            release=os.getenv("SENTRY_RELEASE", settings.service_version),
            traces_sample_rate=0.1,
            attach_stacktrace=True,
        )
        logger.info("Sentry initialized for error tracking")
    except Exception as e:
        logger.warning(f"Failed to initialize Sentry: {e}")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup and shutdown events."""
    # Startup
    logger.info(f"Starting {settings.service_name} v{settings.service_version}")

    # Initialize Sentry first (before other components to catch startup errors)
    _init_sentry()

    # Validate Redis is available at startup (fail fast in production)
    await _validate_redis_startup()

    # Initialize API key validator (uses Go orchestrator for persistent storage)
    # NOTE: This now gracefully degrades if orchestrator is unreachable
    try:
        from .security.auth import initialize_api_key_validator
        from .observability.health import get_health_checker

        checker = get_health_checker()
        healthy = await initialize_api_key_validator()
        if healthy:
            logger.info("API key validator initialized (orchestrator-backed)")
        else:
            checker.set_degraded("Orchestrator unreachable at startup")
            logger.warning(
                "Orchestrator unreachable at startup - running in degraded mode. "
                "API key validation will retry periodically."
            )
    except Exception as e:
        from .observability.health import get_health_checker
        checker = get_health_checker()
        checker.set_degraded(f"API key validator initialization failed: {e}")
        logger.warning(
            f"Failed to initialize API key validator: {e}. "
            "Service starting in degraded mode - orchestrator-backed auth will retry."
        )

    # Initialize providers
    try:
        provider_manager = get_provider_manager()
        all_providers = provider_manager.get_all_providers()
        logger.info(f"Initialized providers: {list(all_providers.keys())}")

        # Update model router with provider availability
        try:
            from .services.generation import get_model_router
            from .models.schemas import ProviderType

            router = get_model_router()
            availability = {}
            for name, provider in all_providers.items():
                try:
                    provider_type = ProviderType(name)
                    availability[provider_type] = getattr(provider, "available", False)
                except ValueError:
                    pass
            router.update_provider_availability(availability)
            logger.info(
                f"Model router initialized with provider availability: {[f'{k.value}: {v}' for k, v in availability.items()]}"
            )
        except Exception as e:
            logger.warning(f"Could not initialize model router: {e}")

    except Exception as e:
        logger.error(f"Failed to initialize providers: {e}")

    # Initialize embeddings service
    try:
        embeddings_service = get_embeddings_service()
        logger.info("Embeddings service initialized")
    except Exception as e:
        logger.error(f"Failed to initialize embeddings service: {e}")

    yield

    # Shutdown
    logger.info("Shutting down...")

    # Close embeddings service
    try:
        embeddings_service = get_embeddings_service()
        await embeddings_service.close()
    except Exception as e:
        logger.warning(f"Error closing embeddings service: {e}")

    logger.info("Shutdown complete")


async def _validate_redis_startup() -> None:
    """Validate Redis is available at startup.

    In production, Redis is required for rate limiting and cost tracking
    to work correctly across multiple instances. Fail fast if Redis is
    configured as required but not reachable.
    """
    import os

    # Check if Redis is explicitly required
    require_redis = os.getenv("REQUIRE_REDIS", "false").lower() == "true"

    if not require_redis:
        logger.info("Redis startup validation skipped (REQUIRE_REDIS not set)")
        return

    try:
        from .services.redis_client import init_redis_client

        redis_client = await init_redis_client()
        if not redis_client or not await redis_client.ping():
            raise RuntimeError("Redis ping failed")

        logger.info("Redis startup validation passed")
    except Exception as e:
        logger.error(f"Redis is required but not available: {e}")
        raise RuntimeError(
            f"Redis is required for production deployment but is not reachable. "
            "Set REQUIRE_REDIS=false to disable this check."
        )


# Create FastAPI application
app = FastAPI(
    title="FlyMind AI Service",
    description="Intelligent capabilities for FunctionFly serverless platform",
    version=settings.service_version,
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
)


# Add CORS middleware with secure defaults
# If cors_origins is empty (default), don't add CORS at all (restrictive)
# If cors_origins is set, use those values
if settings.cors_origins:
    # Validate origins are not wildcard in production
    if "*" in settings.cors_origins:
        import logging
        logging.getLogger(__name__).warning(
            "CORSOrigins contains '*' which allows all origins. "
            "This is insecure for production!"
        )

    # Explicitly specify allowed methods and headers (never use "*" in production)
    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origins,
        allow_credentials=settings.cors_allow_credentials,
allow_methods=settings.cors_allow_methods if settings.cors_allow_methods != ["*"] else ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allow_headers=settings.cors_allow_headers if settings.cors_allow_headers != ["*"] else ["Authorization", "Content-Type", "X-API-Key"],
        max_age=600,
    )

# Add security middleware
from .middleware.security import (
    SecurityHeadersMiddleware,
    RequestSizeLimitMiddleware,
    TimeoutMiddleware,
    TenantValidationMiddleware,
)

app.add_middleware(SecurityHeadersMiddleware)
app.add_middleware(RequestSizeLimitMiddleware)
app.add_middleware(TimeoutMiddleware)
app.add_middleware(TenantValidationMiddleware)

# Add rate limiting middleware (if enabled)
from .middleware.rate_limit_middleware import setup_rate_limit_middleware
setup_rate_limit_middleware(app, enabled=settings.enable_rate_limiting)


# Include API routes
app.include_router(router)


@app.get("/")
async def root():
    """Root endpoint."""
    return JSONResponse(
        {
            "service": settings.service_name,
            "version": settings.service_version,
            "status": "running",
            "docs": "/docs",
        }
    )


@app.get("/health")
async def health_check():
    """Health check endpoint with degraded state reporting.

    Returns:
        JSON health status including whether service is in degraded mode.
        Load balancers can use this to route requests away from degraded instances.
    """
    from .observability.health import get_health_checker

    checker = get_health_checker()
    health = await checker.check_all()
    overall = checker.get_overall_status()

    response = {
        "status": overall.value,
        "service": settings.service_name,
        "version": settings.service_version,
        "components": {
            name: {
                "status": comp.status.value,
                "message": comp.message,
                "latency_ms": round(comp.latency_ms, 2) if comp.latency_ms else None,
            }
            for name, comp in health.items()
        },
    }

    # Include degraded reason if applicable
    if checker.is_degraded():
        response["degraded_reason"] = checker.get_degraded_reason()
        response["degraded"] = True

    # Return 200 for healthy/degraded, 503 for unhealthy
    if overall.value == "unhealthy":
        return JSONResponse(status_code=503, content=response)

    return JSONResponse(content=response)


@app.get("/metrics")
async def prometheus_metrics():
    """Prometheus metrics endpoint.

    Returns:
        Prometheus-formatted metrics text
    """
    from .observability.metrics import get_metrics_collector
    from fastapi.responses import PlainTextResponse

    collector = get_metrics_collector()
    metrics_text = collector.get_metrics_text()
    return PlainTextResponse(content=metrics_text, media_type="text/plain")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
    )
