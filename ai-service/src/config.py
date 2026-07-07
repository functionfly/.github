"""Configuration management for FlyMind AI Service.

Uses Pydantic Settings for environment-based configuration.
"""

import os
from pathlib import Path
from typing import Optional

from pydantic import Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


def _validate_rag_path(path: str) -> str:
    """Validate RAG docs directory path to prevent path traversal.

    Args:
        path: The path to validate

    Returns:
        The validated path

    Raises:
        ValueError: If path is invalid or could enable path traversal
    """
    if not path:
        return path

    # Resolve the path to its absolute form
    resolved_path = Path(path).resolve()

    # Define allowed base directories (adjust for your deployment)
    allowed_bases = [
        Path("/app/docs").resolve(),
        Path("/app/web/docs/src/content/docs").resolve(),
        Path(__file__).resolve().parents[2] / "web" / "docs" / "src" / "content" / "docs",
    ]

    # Check if path is within an allowed directory
    is_allowed = any(
        resolved_path.is_relative_to(base) for base in allowed_bases
    )

    # Also allow paths that are explicitly configured and exist
    if not is_allowed:
        # For production, we should be strict - only allow explicitly configured paths
        if os.getenv("ENVIRONMENT", "development") == "production":
            raise ValueError(
                f"RAG docs directory '{path}' is not in an allowed location. "
                f"Path traversal is not permitted for security reasons."
            )
        else:
            # In development, warn but allow
            import logging
            logging.getLogger(__name__).warning(
                f"RAG docs directory '{path}' is not in a standard location. "
                f"Ensure this is intentional for development only."
            )

    # Check for dangerous path components
    dangerous_components = ["..", "~", "$", "`", "|", ";", "&", "\n", "\r", "\0"]
    path_str = str(resolved_path)
    for component in dangerous_components:
        if component in path_str:
            raise ValueError(
                f"RAG docs directory contains dangerous path component: {component}"
            )

    return str(resolved_path)


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    model_config = SettingsConfigDict(
        env_file=".env", env_file_encoding="utf-8", case_sensitive=False, extra="ignore"
    )

    # Service_name: str = configuration
    service_name: str = "flymind-ai-service"
    service_version: str = "1.0.0"
    host: str = "0.0.0.0"
    port: int = 8081
    debug: bool = False
    log_level: str = "INFO"

    # API keys for LLM providers
    openai_api_key: Optional[str] = Field(default=None, description="OpenAI API key")
    anthropic_api_key: Optional[str] = Field(default=None, description="Anthropic API key")
    openrouter_api_key: Optional[str] = Field(default=None, description="OpenRouter API key")

    # New providers for FunctionFly multi-provider architecture
    fireworks_api_key: Optional[str] = Field(
        default=None, description="Fireworks AI API key (best for function calling)"
    )
    groq_api_key: Optional[str] = Field(
        default=None, description="Groq API key (best for low-latency)"
    )
    deepinfra_api_key: Optional[str] = Field(
        default=None, description="DeepInfra API key (best for background/batch)"
    )
    together_api_key: Optional[str] = Field(
        default=None, description="Together AI API key (alternative provider)"
    )

    # Ollama configuration (local/development)
    ollama_base_url: str = Field(default="", description="Ollama base URL - must be set explicitly")
    # Smaller default so local dev fits typical RAM; override with OLLAMA_MODEL (e.g. llama3.3 when you have headroom)
    ollama_model: str = "llama3.2:3b"
    ollama_embedding_model: str = "nomic-embed-text"

    # Default provider
    default_provider: str = "openrouter"
    default_embedding_provider: str = "openai"

    # Rate limits per provider (requests per minute)
    openai_rate_limit: int = 60
    anthropic_rate_limit: int = 50
    ollama_rate_limit: int = 100
    openrouter_rate_limit: int = 60

    # Rate limits for new providers
    fireworks_rate_limit: int = 120  # Fireworks: high throughput
    groq_rate_limit: int = 30  # Groq free tier: 30 RPM (upgrade for more)
    deepinfra_rate_limit: int = 100  # DeepInfra: good for batch
    together_rate_limit: int = 60  # Together: standard rate limit

    # Redis configuration
    redis_url: str = "redis://localhost:6379"
    redis_password: Optional[str] = Field(default=None, description="Redis password (if auth enabled)")
    redis_cache_ttl: int = 3600  # seconds
    redis_use_tls: bool = Field(
        default=False, description="Enable TLS for Redis connections"
    )

    # Redis connection retry configuration
    redis_max_connection_retries: int = 5
    redis_base_retry_delay: float = 0.5
    redis_max_retry_delay: float = 30.0
    redis_retry_on_connect: bool = True

    # Database configuration
    database_url: str = Field(
        default="", description="Database connection string - must be set explicitly"
    )

    # CORS settings
    cors_origins: list[str] = Field(
        default=[], description="CORS allowed origins - must be set explicitly"
    )
    cors_allow_credentials: bool = True
    cors_allow_methods: list[str] = ["GET", "POST", "OPTIONS"]
    cors_allow_headers: list[str] = []
    cors_max_age: int = Field(default=600, description="CORS preflight cache duration (seconds)")
    cors_expose_headers: list[str] = Field(default=[], description="CORS exposed headers")

    @field_validator("cors_allow_methods", "cors_allow_headers", "cors_expose_headers", mode="after")
    @classmethod
    def _warn_on_wildcard_cors(cls, v: list[str], info) -> list[str]:
        """Warn if wildcard '*' is used in CORS settings."""
        if "*" in v:
            import logging
            logger = logging.getLogger(__name__)
            logger.warning(
                f"CORS setting '{info.field_name}' contains wildcard '*' which is insecure. "
                f"Consider using explicit method/header names instead."
            )
        return v

    # Retry configuration
    max_retries: int = 3
    retry_base_delay: float = 1.0  # seconds
    retry_max_delay: float = 30.0  # seconds

    # Circuit breaker configuration
    circuit_breaker_enabled: bool = Field(default=True, description="Enable circuit breaker for external API calls")
    circuit_breaker_failure_threshold: int = Field(default=5, description="Number of failures before opening circuit")
    circuit_breaker_recovery_timeout: float = Field(default=60.0, description="Seconds before attempting recovery from open state")
    circuit_breaker_half_open_max_calls: int = Field(default=3, description="Max test calls in half-open state")

    # Provider-specific settings (2026-era defaults)
    openai_model: str = "gpt-4o"
    openai_embedding_model: str = "text-embedding-3-small"
    openai_embedding_dimensions: int = 1536

    anthropic_model: str = "claude-sonnet-4-6"
    anthropic_max_tokens: int = 8192

    openrouter_model: str = "poolside/laguna-xs.2:free"
    openrouter_base_url: str = "https://openrouter.ai/api/v1"

    # Fireworks AI configuration (primary for structured output)
    fireworks_model: str = "accounts/fireworks/models/llama-v3p1-405b-instruct"
    fireworks_base_url: str = "https://api.fireworks.ai/inference/v1"

    # Groq configuration (latency-critical paths)
    groq_model: str = "llama-4-scout-17b-16e-instruct"
    groq_base_url: str = "https://api.groq.com/openai/v1"

    # DeepInfra configuration (background/batch work)
    deepinfra_model: str = "meta-llama/Llama-3.3-70B-Instruct-Turbo"
    deepinfra_embedding_model: str = "BAAI/bge-large-en-v1.5"
    deepinfra_base_url: str = "https://api.deepinfra.com/v1/openai"

    # Together AI configuration (alternative/batch)
    together_model: str = "meta-llama/Llama-3.3-70B-Instruct-Turbo"
    together_base_url: str = "https://api.together.xyz/v1"

    # Feature flags
    enable_streaming: bool = True
    enable_caching: bool = True
    enable_cost_tracking: bool = True
    enable_rate_limiting: bool = Field(default=True, description="Enable rate limiting middleware")

    # RAG (retrieval-augmented generation) for chat
    enable_rag: bool = True
    rag_docs_dir: str = Field(
        default=str(
            (Path(__file__).resolve().parents[2] / "web" / "docs" / "src" / "content" / "docs")
        ),
        description="Directory containing markdown docs to use for RAG",
    )

    @field_validator("rag_docs_dir")
    @classmethod
    def validate_rag_docs_dir(cls, v: str) -> str:
        """Validate RAG docs directory path."""
        return _validate_rag_path(v)
    rag_top_k: int = 4
    rag_candidate_chunks: int = 24
    # Limit expensive embedding rerank work per request for snappy chat UX.
    rag_embedding_rerank_chunks: int = 6
    rag_embedding_max_seconds: float = 3.0
    rag_chunk_max_chars: int = 1600
    rag_chunk_min_chars: int = 250
    rag_max_chunks: int = 2000
    # RAG retrieval uses embeddings; default to Ollama (local) so no cloud API key is required.
    # Set to openai | anthropic | openrouter | ollama
    rag_embedding_provider: str = "ollama"

    # Cost per 1K tokens (2026 model pricing - estimated continued cost reductions)
    # GPT-4o class models (high performance)
    openai_input_cost: float = 0.0015  # $1.50 per 1M tokens (2026 estimate)
    openai_output_cost: float = 0.006  # $6 per 1M tokens (2026 estimate)
    # GPT-4o-mini (cost-optimized, used for memory extraction)
    openai_mini_input_cost: float = 0.00015  # $0.15 per 1M tokens (2026 estimate)
    openai_mini_output_cost: float = 0.0006  # $0.60 per 1M tokens (2026 estimate)
    # Anthropic Claude (Sonnet 4-6 era)
    anthropic_input_cost: float = 0.002  # $2 per 1M tokens (2026 estimate)
    anthropic_output_cost: float = 0.008  # $8 per 1M tokens (2026 estimate)
    # Embedding costs (text-embedding-3-small)
    embedding_cost_per_1k: float = 0.00002  # $0.02 per 1K tokens (2026 estimate)

    # Phase 2: Intelligent Routing Configuration
    routing_enabled: bool = True
    routing_latency_weight: float = 0.30  # 30%
    routing_load_weight: float = 0.30  # 30%
    routing_availability_weight: float = 0.40  # 40%
    routing_cache_ttl_seconds: int = 60

    # FunctionFly Multi-Provider Traffic-Based Routing
    # These control which providers are used for different traffic types
    traffic_realtime_provider: str = "groq"  # Low-latency: Groq
    traffic_structured_provider: str = "fireworks"  # JSON/tool: Fireworks
    traffic_function_calling_provider: str = "fireworks"  # Function calls: Fireworks
    traffic_background_provider: str = "deepinfra"  # Batch: DeepInfra
    traffic_fallback_provider: str = "openrouter"  # Fallback: OpenRouter
    enable_traffic_based_routing: bool = True  # Enable smart traffic routing

    # Phase 2: Prewarming Configuration
    prewarming_enabled: bool = True
    prewarming_threshold: int = 10  # Predicted requests threshold
    prewarming_window_minutes: int = 10
    prewarming_instances_default: int = 1

    # Phase 2: Anomaly Detection Configuration
    anomaly_detection_enabled: bool = True
    anomaly_latency_threshold: float = 3.0  # Z-score threshold
    anomaly_error_rate_threshold: float = 0.01  # 1%
    anomaly_cold_start_threshold: float = 0.10  # 10%
    anomaly_window_minutes: int = 5
    anomaly_check_interval_seconds: int = 30

    # Orchestrator integration
    # In Fly.io production, use internal DNS: http://functionfly-api.internal:8080
    orchestrator_url: str = Field(
        default="http://localhost:8080", description="URL for the Go orchestrator API"
    )
    orchestrator_api_key: Optional[str] = Field(
        default=None, description="API key for orchestrator authentication"
    )
    ai_service_api_key: Optional[str] = Field(
        default=None, description="API key the orchestrator uses to call this AI service (supports comma-separated list for key rotation)"
    )
    ai_service_api_key_deprecated: Optional[str] = Field(
        default=None, description="Deprecated API key - still accepted but logs warning (for key rotation grace period)"
    )
    ai_service_api_key_rotation_warning_days: int = Field(
        default=30, description="Days to warn about deprecated key usage before rejecting"
    )

    # Internal API secret for service-to-service authentication
    # Required for accessing internal endpoints like /internal/composer/generate
    internal_api_secret: Optional[str] = Field(
        default=None, description="Secret for internal API authentication"
    )

    # Content moderation (2026: OpenAI Moderation API recommended when key is set)
    # auto = use OpenAI if key present, else Detoxify, else keyword fallback
    moderation_provider: str = "auto"  # auto | openai | detoxify | keywords
    openai_moderation_model: str = "omni-moderation-latest"

    # FlyEmbed Triple-Vector Configuration
    flyembed_model: str = "text-embedding-3-small"
    flyembed_dimensions: int = 512
    flyembed_default_weight_contract: float = 0.35
    flyembed_default_weight_semantic: float = 0.40
    flyembed_default_weight_code: float = 0.25
    flyembed_batch_size: int = 10
    flyembed_max_source_code_chars: int = 2000

    # Production Security Settings
    security_mode: str = "strict"  # strict | relaxed | disabled
    max_embedding_batch_size: int = 100  # Limit batch operations
    embedding_timeout_seconds: float = 30.0  # Timeout for embedding calls
    require_auth_for_embed: bool = True
    log_all_embedding_requests: bool = True
    alert_on_high_cost_threshold: float = 0.8  # Alert at 80% of daily budget

    # Embedding Security - PII Detection
    embedding_pii_check_enabled: bool = True
    embedding_pii_mode: str = "block"  # block | redact | warn
    embedding_max_input_length: int = 8000  # Prevent token bombing
    embedding_blocked_patterns: list[str] = []  # Regex patterns to block

    # Request Security - Timeouts and Size Limits
    request_timeout_seconds: float = 30.0  # Max request duration
    max_request_body_bytes: int = 10 * 1024 * 1024  # 10MB max

    # Redis Cache Security
    redis_cache_encryption_key: Optional[str] = Field(
        default=None, description="Fernet key for cache encryption (base64-encoded, 32 bytes)"
    )
    redis_cache_namespace: str = "flyembed"

    # Redis Retry and Resilience
    redis_max_connection_retries: int = Field(
        default=3, description="Max retries for Redis connection with exponential backoff"
    )
    redis_base_retry_delay: float = Field(
        default=0.5, description="Base delay for Redis connection retry (seconds)"
    )
    redis_max_retry_delay: float = Field(
        default=30.0, description="Max delay for Redis connection retry (seconds)"
    )
    redis_retry_on_connection: bool = Field(
        default=True, description="Enable retry logic for Redis connection failures"
    )
    redis_circuit_breaker_failure_threshold: int = Field(
        default=5, description="Failure threshold before opening Redis circuit breaker"
    )
    redis_circuit_breaker_recovery_timeout: float = Field(
        default=30.0, description="Seconds before trying to reconnect after circuit opens"
    )

    # Internal API Key (for service-to-service auth)
    internal_api_key: Optional[str] = Field(
        default=None, description="Internal API key for /internal endpoints (service-to-service auth)"
    )

    # Security - Auth Degraded Mode
    # If True (default), when orchestrator is unreachable, all API key auth requests are REJECTED.
    # If False, cached keys may still work (less secure but higher availability).
    reject_auth_in_degraded_mode: bool = Field(
        default=True, description="Reject auth requests when orchestrator is unreachable"
    )

    # Security - Allow Cached Auth in Degraded Mode
    # If True and reject_auth_in_degraded_mode is True, cached keys will still work during degraded mode.
    # If False, even cached keys are rejected during degraded mode (most secure).
    # WARNING: Setting to True means revoked keys may still work for up to cache TTL (60s).
    allow_cached_auth_in_degraded_mode: bool = Field(
        default=True,
        description="Allow cached API keys to work when orchestrator is unreachable (less secure)"
    )

    # Security - Auth Cache TTL
    # How long to cache API key validation results (in seconds).
    # Shorter TTL = faster revocation propagation, longer TTL = better performance.
    auth_cache_ttl_seconds: int = Field(
        default=60,
        description="How long to cache API key validation results (seconds)"
    )

    # Security - Degraded Mode Cache TTL
    # When orchestrator is unreachable, how long to allow cached keys to work (seconds).
    # Only used when allow_cached_auth_in_degraded_mode is True.
    # Set to a low value for higher security (faster revocation propagation).
    degraded_auth_cache_ttl_seconds: int = Field(
        default=30,
        description="TTL for cached auth when in degraded mode (seconds)"
    )

    # Security - Require API Key for All Endpoints
    # If True, all endpoints require API key authentication (except /health, /metrics, /docs)
    # If False, some endpoints may be accessible without auth (not recommended for production)
    require_auth_all_endpoints: bool = Field(
        default=True, description="Require API key authentication for all protected endpoints"
    )

    # Rate Limiting - Embedding Specific
    embed_tokens_per_minute: int = 100000  # Token budget for embeddings
    embed_cost_per_day: float = 50.0  # USD budget for embeddings per tenant

    # Cost Limits
    daily_cost_limit_usd: float = Field(default=100.0, description="Daily cost limit per tenant in USD")

    # RAG Security
    rag_validate_content: bool = True
    rag_blocked_file_patterns: list[str] = [
        ".env",
        ".key",
        ".pem",
        ".secret",
        "password",
        "credential",
    ]
    rag_max_query_length: int = 1000

    # Atlas Observability
    atlas_enabled: bool = Field(default=True, description="Enable Atlas observability")
    atlas_base_url: str = Field(default="http://localhost:7447", description="Atlas Cloud base URL")
    atlas_grpc_host: str = Field(default="localhost", description="Atlas gRPC host")
    atlas_grpc_port: int = Field(default=50051, description="Atlas gRPC port")
    atlas_api_key: Optional[str] = Field(default=None, description="Atlas API key")
    atlas_agent_id_prefix: str = Field(default="flymind", description="Prefix for agent IDs in Atlas")

    # Atlas Sampling
    atlas_sample_rate: float = Field(default=1.0, description="Sampling rate 0.0-1.0")
    atlas_trace_errors_only: bool = Field(default=False, description="Only trace errors")
    atlas_sample_head_percent: float = Field(default=100.0, description="Head-based sampling %")
    atlas_sample_tail_count: int = Field(default=10, description="Tail-based: keep last N events")

    # Atlas OpenTelemetry Export
    atlas_otel_exporter_enabled: bool = Field(default=False, description="Export Atlas events as OTEL spans")
    atlas_otel_endpoint: Optional[str] = Field(default=None, description="OTLP endpoint for Atlas export")

    # ML Intelligence Layer
    ml_enabled: bool = Field(default=True, description="Enable ML services")
    ml_retrain_cron: str = Field(default="0 3 * * *", description="Cron for model retraining")
    ml_model_dir: str = Field(default="/var/lib/flymind/models", description="Model storage directory")
    ml_model_encryption_key: Optional[str] = Field(
        default=None,
        description="Fernet AES-256-GCM key for model encryption at rest (base64-encoded, generate with: python -c 'from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())')"
    )
    ml_cost_anomaly_threshold: float = Field(default=3.0, description="Z-score threshold for cost anomalies")
    ml_cost_anomaly_window_hours: int = Field(default=168, description="Cost anomaly sliding window (hours)")
    ml_prewarm_seasonality_periods: int = Field(default=24, description="Seasonality periods for Holt-Winters")
    ml_routing_exploration: float = Field(default=0.1, description="Thompson Sampling exploration budget")
    ml_routing_use_informed_priors: bool = Field(
        default=True,
        description="Use function-type-based priors for Thompson Sampling cold-start"
    )

    # ML Per-Tenant Resource Quotas
    ml_quota_enabled: bool = Field(
        default=False,
        description="Enable per-tenant ML resource quotas"
    )
    ml_quota_predictions_per_minute: int = Field(
        default=1000,
        description="Maximum ML predictions per minute per tenant"
    )
    ml_quota_predictions_per_day: int = Field(
        default=100000,
        description="Maximum ML predictions per day per tenant"
    )
    ml_quota_training_per_day: int = Field(
        default=10,
        description="Maximum ML training requests per day per tenant"
    )
    ml_quota_concurrent_requests: int = Field(
        default=10,
        description="Maximum concurrent ML requests per tenant"
    )
    ml_recommendation_latent_dims: int = Field(default=50, description="ALS latent factor dimensions")
    ml_synthetic_data_enabled: bool = Field(default=True, description="Use synthetic data for bootstrap")
    ml_training_timeout_seconds: float = Field(default=300.0, description="Max duration for ML model training (5 minutes)")
    ml_training_max_retries: int = Field(default=1, description="Max retries for ML training on failure")
    ml_training_rate_limit_per_hour: int = Field(
        default=4,
        description="Maximum training requests per hour per tenant (to prevent resource abuse)"
    )

    # ML Production Security
    ml_require_encryption_in_production: bool = Field(
        default=True,
        description="Require model encryption in production (Enforced when ENVIRONMENT=production)"
    )
    ml_model_backup_enabled: bool = Field(
        default=True,
        description="Enable automatic model backup before retraining"
    )
    ml_model_backup_dir: str = Field(
        default="/var/lib/flymind/backups",
        description="Directory for model backups"
    )
    ml_model_max_backups: int = Field(
        default=5,
        description="Maximum number of model backups to retain per service"
    )

    # ML Drift Detection & Monitoring
    ml_drift_detection_enabled: bool = Field(
        default=True,
        description="Enable model drift detection for cost anomaly and recommendations"
    )
    ml_drift_threshold: float = Field(
        default=0.15,
        description="Drift threshold (0.0-1.0) for alerting - percentage change in distribution"
    )
    ml_drift_check_interval_hours: int = Field(
        default=6,
        description="How often to check for model drift (hours)"
    )
    ml_drift_alert_webhook_url: Optional[str] = Field(
        default=None,
        description="Webhook URL for drift alerts (POST with JSON payload)"
    )

    # ML Automatic Rollback
    ml_auto_rollback_enabled: bool = Field(
        default=False,
        description="Enable automatic model rollback when significant drift is detected"
    )
    ml_auto_rollback_threshold: float = Field(
        default=0.5,
        description="Minimum drift score (0.0-1.0) to trigger automatic rollback"
    )
    ml_max_rollback_per_day: int = Field(
        default=3,
        description="Maximum automatic rollbacks per day per service per tenant"
    )
    ml_rollback_alert_webhook_url: Optional[str] = Field(
        default=None,
        description="Webhook URL for rollback alerts (POST with JSON payload)"
    )

    # ML File-based Fallback (when Redis is unavailable)
    ml_fallback_to_file: bool = Field(
        default=True,
        description="Enable file-based fallback for ML state when Redis is unavailable"
    )
    ml_fallback_file_dir: str = Field(
        default="/var/lib/flymind/ml-state",
        description="Directory for file-based fallback state"
    )

    # Cache Warming
    cache_warming_enabled: bool = Field(
        default=True,
        description="Enable cache warming on startup for ML services"
    )
    cache_warming_timeout_seconds: float = Field(
        default=300.0,
        description="Maximum time to spend warming caches (seconds)"
    )
    cache_warming_max_tenants: int = Field(
        default=1000,
        description="Maximum number of tenants to warm caches for on startup"
    )

    # Cache Warming Configuration
    cache_warming_enabled: bool = Field(
        default=True,
        description="Enable cache warming on startup to preload ML state"
    )
    cache_warming_timeout_seconds: float = Field(
        default=300.0,
        description="Maximum time to spend warming cache on startup"
    )
    cache_warming_max_tenants: int = Field(
        default=1000,
        description="Maximum number of tenants to warm cache for"
    )


# Global settings instance
settings = Settings()


def get_settings() -> Settings:
    """Get the global settings instance.

    Returns:
        The global Settings instance.
    """
    return settings
