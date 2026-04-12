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

    # New providers for FunctionFly multi-provider architecture
    fireworks_api_key: Optional[str] = Field(default=None, description="Fireworks AI API key (best for function calling)")
    groq_api_key: Optional[str] = Field(default=None, description="Groq API key (best for low-latency)")
    deepinfra_api_key: Optional[str] = Field(default=None, description="DeepInfra API key (best for background/batch)")
    together_api_key: Optional[str] = Field(default=None, description="Together AI API key (alternative provider)")

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

    # Rate limits for new providers
    fireworks_rate_limit: int = 120  # Fireworks: high throughput
    groq_rate_limit: int = 30  # Groq free tier: 30 RPM (upgrade for more)
    deepinfra_rate_limit: int = 100  # DeepInfra: good for batch
    together_rate_limit: int = 60  # Together: standard rate limit

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
    cors_origins: list[str] = ["http://localhost:3000"]
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

    # Cost per 1K tokens (2026 model pricing - estimated continued cost reductions)
    # GPT-4o class models (high performance)
    openai_input_cost: float = 0.0015  # $1.50 per 1M tokens (2026 estimate)
    openai_output_cost: float = 0.006   # $6 per 1M tokens (2026 estimate)
    # GPT-4o-mini (cost-optimized, used for memory extraction)
    openai_mini_input_cost: float = 0.00015   # $0.15 per 1M tokens (2026 estimate)
    openai_mini_output_cost: float = 0.0006   # $0.60 per 1M tokens (2026 estimate)
    # Anthropic Claude (Sonnet 4-6 era)
    anthropic_input_cost: float = 0.002   # $2 per 1M tokens (2026 estimate)
    anthropic_output_cost: float = 0.008   # $8 per 1M tokens (2026 estimate)
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

    # Redis Cache Security
    redis_cache_encryption_key: Optional[str] = Field(
        default=None, description="Fernet key for cache encryption (base64-encoded)"
    )
    redis_cache_namespace: str = "flyembed"
    redis_cache_ttl: int = 3600

    # Rate Limiting - Embedding Specific
    embed_tokens_per_minute: int = 100000  # Token budget for embeddings
    embed_cost_per_day: float = 50.0  # USD budget for embeddings per tenant

    # RAG Security
    rag_validate_content: bool = True
    rag_blocked_file_patterns: list[str] = [".env", ".key", ".pem", ".secret", "password", "credential"]
    rag_max_query_length: int = 1000


# Global settings instance
settings = Settings()


def get_settings() -> Settings:
    """Get the global settings instance.

    Returns:
        The global Settings instance.
    """
    return settings
