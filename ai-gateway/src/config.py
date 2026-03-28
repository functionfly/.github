"""Configuration settings for AI Gateway.

Environment variables:
    HOST: Server bind host (default: 0.0.0.0)
    PORT: Server port (default: 8082)
    RUNPOD_API_KEY: RunPod API key for cluster management
    RUNPOD_CLUSTER_URL: Base URL for RunPod cluster manager API
    DEFAULT_MODEL: Default model for inference (default: phi-3-mini)
    MAX_CONTEXT_LENGTH: Maximum context length (default: 4096)
    API_KEY_HEADER: Header name for API key authentication (default: X-API-Key)
    TENANT_ID_HEADER: Header name for tenant identification (default: X-Tenant-ID)
    MAX_BATCH_SIZE: Maximum batch size for inference (default: 8)
    BATCH_TIMEOUT_MS: Batch timeout in milliseconds (default: 100)
"""

import os
from functools import lru_cache
from typing import List, Optional

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """AI Gateway configuration settings."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=True,
        extra="ignore",
    )

    # Server Configuration
    HOST: str = "0.0.0.0"
    PORT: int = 8082
    DEBUG: bool = False

    # RunPod Integration
    RUNPOD_API_KEY: str = Field(default="", description="RunPod API key")
    RUNPOD_CLUSTER_URL: str = Field(
        default="http://localhost:8080",
        description="Base URL for cluster manager API",
    )
    RUNPOD_API_BASE_URL: str = Field(
        default="https://api.runpod.io/graphql",
        description="RunPod GraphQL API base URL",
    )

    # Model Configuration
    DEFAULT_MODEL: str = "phi-3-mini"
    MAX_CONTEXT_LENGTH: int = 4096
    MODEL_CACHE_DIR: str = "/tmp/model-cache"

    # Security
    API_KEY_HEADER: str = "X-API-Key"
    TENANT_ID_HEADER: str = "X-Tenant-ID"
    REQUIRED_API_KEY: Optional[str] = Field(
        default=None, description="Static API key for service authentication"
    )

    # Rate Limiting
    RATE_LIMIT_REQUESTS: int = 100
    RATE_LIMIT_WINDOW_SECONDS: int = 60

    # Performance
    MAX_BATCH_SIZE: int = 8
    BATCH_TIMEOUT_MS: int = 100
    MAX_CONCURRENT_REQUESTS: int = 50
    REQUEST_TIMEOUT_SECONDS: int = 300

    # Circuit Breaker
    CIRCUIT_BREAKER_FAILURE_THRESHOLD: int = 5
    CIRCUIT_BREAKER_RECOVERY_TIMEOUT_SECONDS: int = 60

    # CORS
    CORS_ORIGINS: List[str] = Field(
        default=["*"], description="Allowed CORS origins"
    )
    CORS_ALLOW_CREDENTIALS: bool = True
    CORS_ALLOW_METHODS: List[str] = Field(
        default=["*"], description="Allowed HTTP methods"
    )
    CORS_ALLOW_HEADERS: List[str] = Field(
        default=["*"], description="Allowed HTTP headers"
    )

    # Health Check
    HEALTH_CHECK_INTERVAL_SECONDS: int = 30
    HEALTH_CHECK_TIMEOUT_SECONDS: int = 5

    # Logging
    LOG_LEVEL: str = "INFO"

    # Inference Backend (select: runpod, onnx, openai)
    INFERENCE_BACKEND: str = "runpod"

    # OpenAI-compatible API (fallback)
    OPENAI_API_KEY: Optional[str] = None
    OPENAI_API_BASE: str = "https://api.openai.com/v1"

    # Cost tracking
    COST_PER_TOKEN: float = 0.00001
    ENABLE_COST_TRACKING: bool = True


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()


# Convenience function for dependency injection
def get_config() -> Settings:
    """Alias for get_settings() for FastAPI dependency injection."""
    return get_settings()
