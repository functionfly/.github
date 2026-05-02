"""FlyMind AI Service - Main Application.

This is the FastAPI application entry point.
Includes Phase 1 (Foundation) and Phase 2 (Intelligence Layer).
"""

import hashlib
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
from .security.auth import get_api_key_validator, KeyScope


# Configure logging (uppercase so "info" maps to logging.INFO, not logging.info)
_log_level = getattr(logging, settings.log_level.upper(), logging.INFO)
if not isinstance(_log_level, int):
    _log_level = logging.INFO
logging.basicConfig(
    level=_log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for startup and shutdown events."""
    # Startup
    logger.info(f"Starting {settings.service_name} v{settings.service_version}")

    # Initialize API key validator and load persistent key from environment
    try:
        validator = get_api_key_validator()
        persistent_key = os.environ.get("AI_SERVICE_API_KEY")
        if persistent_key:
            # Check if this key already exists
            key_info = validator.validate_key(persistent_key)
            if not key_info:
                # Add the persistent key to the validator
                key_hash = hashlib.sha256(persistent_key.encode()).hexdigest()
                from .security.auth import APIKeyInfo, KeyStatus
                from datetime import datetime
                from typing import List

                info = APIKeyInfo(
                    key_id="persistent",
                    tenant_id="system",
                    name="Persistent API Key",
                    scopes=[KeyScope.FULL, KeyScope.CHAT_WRITE],
                    status=KeyStatus.ACTIVE,
                    created_at=datetime.utcnow(),
                    rate_limit=120,
                )
                validator._keys["persistent"] = (key_hash, info)
                validator._key_lookup[key_hash] = "persistent"
                logger.info("Loaded persistent API key from AI_SERVICE_API_KEY environment")
        else:
            logger.info("No AI_SERVICE_API_KEY environment variable set")
    except Exception as e:
        logger.warning(f"Could not initialize persistent API key: {e}")

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


# Create FastAPI application
app = FastAPI(
    title="FlyMind AI Service",
    description="Intelligent capabilities for FunctionFly serverless platform",
    version=settings.service_version,
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
)


# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=settings.cors_allow_credentials,
    allow_methods=settings.cors_allow_methods,
    allow_headers=settings.cors_allow_headers,
)


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
