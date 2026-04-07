"""Configuration management for FlyMind AI Service.

Uses Pydantic Settings for environment-based configuration.
"""

from typing import Optional
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict
from pathlib import Path


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore"
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

    # Ollama configuration (local/development)
    ollama_base_url: str = "http://localhost:11434"
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

    # Redis configuration
    redis_url: str = "redis://localhost:6379"
    redis_cache_ttl: int = 3600  # seconds

    # Upstash Redis configuration (alternative to standard Redis)
    # Get these from https://console.upstash.com
    upstash_redis_url: Optional[str] = Field(default=None, description="Upstash Redis REST URL")
    upstash_redis_token: Optional[str] = Field(default=None, description="Upstash Redis REST Token")
    use_upstash_redis: bool = False

    # Database configuration
    database_url: str = "postgresql://postgres:postgres@localhost:5432/functionfly"

    # CORS settings
    cors_origins: list[str] = ["http://localhost:3000", "http://"]
    cors_allow_credentials: bool = True
    cors_allow_methods: list[str] = ["*"]
    cors_allow_headers: list[str] = ["*"]

    # Retry configuration
    max_retries: int = 3
    retry_base_delay: float = 1.0  # seconds
    retry_max_delay: float = 30.0  # seconds

    # Provider-specific settings (2026-era defaults)
    openai_model: str = "gpt-4o"
    openai_embedding_model: str = "text-embedding-3-small"
    openai_embedding_dimensions: int = 1536

    anthropic_model: str = "claude-sonnet-4-6"
    anthropic_max_tokens: int = 8192

    openrouter_model: str = "openrouter/free"
    openrouter_base_url: str = "https://openrouter.ai/api/v1"

    # Feature flags
    enable_streaming: bool = True
    enable_caching: bool = True
    enable_cost_tracking: bool = True

    # RAG (retrieval-augmented generation) for chat
    enable_rag: bool = True
    rag_docs_dir: str = Field(
        default=str((Path(__file__).resolve().parents[2] / "web" / "docs" / "src" / "content" / "docs")),
        description="Directory containing markdown docs to use for RAG",
    )
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

    # Cost per 1K tokens (for tracking)
    openai_input_cost: float = 0.0025  # $ per 1K tokens
    openai_output_cost: float = 0.01
    anthropic_input_cost: float = 0.003
    anthropic_output_cost: float = 0.015

    # Phase 2: Intelligent Routing Configuration
    routing_enabled: bool = True
    routing_latency_weight: float = 0.30  # 30%
    routing_load_weight: float = 0.30  # 30%
    routing_availability_weight: float = 0.40  # 40%
    routing_cache_ttl_seconds: int = 60

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
        default="http://localhost:8080",
        description="URL for the Go orchestrator API"
    )
    orchestrator_api_key: Optional[str] = Field(
        default=None,
        description="API key for orchestrator authentication"
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


# Global settings instance
settings = Settings()


def get_settings() -> Settings:
    """Get the global settings instance.

    Returns:
        The global Settings instance.
    """
    return settings
