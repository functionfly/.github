"""Embeddings service for FlyMind AI Service.

This module provides the embeddings service with caching support.
"""

import hashlib
import json
import logging
import time
from typing import Optional, Any, Dict

import redis.asyncio as redis

from ..config import settings
from ..models.schemas import EmbeddingRequest, EmbeddingResponse, ProviderType
from ..providers.manager import get_provider_manager
from ..security.auth import APIKeyInfo
from ..security.audit import get_audit_logger, AuditOperation, AuditEventStatus
from ..middleware.cost_tracker import get_cost_tracker, CostLimitExceeded

# Optional cryptography import
try:
    from cryptography.fernet import Fernet
    CRYPTOGRAPHY_AVAILABLE = True
except ImportError:
    CRYPTOGRAPHY_AVAILABLE = False
    Fernet = None


logger = logging.getLogger(__name__)


class EmbeddingsService:
    """Service for generating embeddings with Redis caching."""

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._cache_ttl = settings.redis_cache_ttl

    async def get_redis(self) -> Optional[redis.Redis]:
        """Get Redis connection."""
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url,
                    encoding="utf-8",
                    decode_responses=True,
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Failed to connect to Redis: {e}")
                self._redis = None
        return self._redis

    def _get_cache_key(
        self,
        text: str,
        model: str,
        dimensions: Optional[int],
        provider_name: Optional[str] = None,
        tenant_id: Optional[str] = "default",
    ) -> str:
        """Generate a tenant-scoped cache key for the embedding request.
        
        Args:
            text: Text to embed
            model: Model name
            dimensions: Embedding dimensions
            provider_name: Provider name
            tenant_id: Tenant ID for multi-tenant isolation
            
        Returns:
            Tenant-scoped cache key
        """
        key_data = {
            "tenant": tenant_id,
            "text": text,
            "model": model,
            "dimensions": dimensions,
            "provider": provider_name or "",
        }
        key_str = json.dumps(key_data, sort_keys=True)
        return f"embedding:{tenant_id}:{hashlib.sha256(key_str.encode()).hexdigest()}"

    async def generate_embedding(
        self,
        request: EmbeddingRequest,
        api_key_info: Optional[APIKeyInfo] = None,
    ) -> EmbeddingResponse:
        """Generate an embedding for the given text.

        Args:
            request: The embedding request
            api_key_info: API key info for audit logging and cost tracking

        Returns:
            EmbeddingResponse with the embedding vector
        """
        start_time = time.monotonic()
        audit_logger = get_audit_logger()
        cost_tracker = get_cost_tracker()
        
        provider_manager = get_provider_manager()

        # Get the provider
        provider_name = request.provider.value if request.provider else None
        provider = provider_manager.get_embedding_provider(provider_name)
        resolved_model = request.model or getattr(provider, "embedding_model", None) or provider.model
        resolved_dimensions = request.dimensions or getattr(provider, "embedding_dimensions", None) or 1536
        
        # Get tenant ID from API key or default
        tenant_id = getattr(api_key_info, 'tenant_id', 'default') if api_key_info else 'default'
        api_key_id = getattr(api_key_info, 'key_id', 'unknown') if api_key_info else 'unknown'
        
        # Estimate cost and check budget (only for cloud providers)
        estimated_tokens = len(request.text.split()) + 10  # Rough estimate
        if settings.enable_cost_tracking and api_key_info:
            try:
                # Estimate cost based on provider
                estimated_cost = self._estimate_embedding_cost(estimated_tokens, resolved_model)
                cost_tracker.check_limit(tenant_id, estimated_cost)
            except CostLimitExceeded as e:
                # Log the blocked attempt
                latency_ms = (time.monotonic() - start_time) * 1000
                audit_logger.log_embedding_event(
                    tenant_id=tenant_id,
                    api_key_id=api_key_id,
                    operation=AuditOperation.EMBED_GENERATE,
                    model=resolved_model,
                    dimensions=resolved_dimensions,
                    text=request.text,
                    success=False,
                    status=AuditEventStatus.BLOCKED,
                    latency_ms=latency_ms,
                    error_message=f"Cost limit exceeded: {e}",
                    token_count=estimated_tokens,
                )
                raise

        # PII/Content filtering before embedding
        if settings.embedding_pii_check_enabled:
            from ..services.moderation import get_content_scanner
            scanner = get_content_scanner()
            sanitized_text, violations = await scanner.sanitize_for_embedding(
                request.text,
                mode=settings.embedding_pii_mode
            )
            
            if violations:
                violation_types = [v.matched_pattern for v in violations]
                
                if settings.embedding_pii_mode == "block":
                    # Log blocked attempt
                    latency_ms = (time.monotonic() - start_time) * 1000
                    audit_logger.log_embedding_event(
                        tenant_id=tenant_id,
                        api_key_id=api_key_id,
                        operation=AuditOperation.EMBED_GENERATE,
                        model=resolved_model,
                        dimensions=resolved_dimensions,
                        text=request.text,
                        success=False,
                        status=AuditEventStatus.BLOCKED,
                        latency_ms=latency_ms,
                        error_message=f"PII/Secrets detected: {violation_types}",
                        token_count=estimated_tokens,
                        metadata={"violations": violation_types},
                    )
                    raise ValueError(
                        f"Content blocked: PII/secrets detected ({', '.join(violation_types)}). "
                        f"Use 'redact' mode to mask sensitive data."
                    )
                elif settings.embedding_pii_mode == "redact":
                    # Use sanitized text
                    request.text = sanitized_text
                    logger.info(f"PII redacted from embedding input: {violation_types}")
                # In "warn" mode, we continue with original text but violations are logged

        # Check input length limit
        if len(request.text) > settings.embedding_max_input_length:
            raise ValueError(
                f"Input text exceeds maximum length of {settings.embedding_max_input_length} characters"
            )

        # Check cache if enabled
        cache_key = None
        cache_hit = False
        if settings.enable_caching:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = self._get_cache_key(
                    request.text,
                    resolved_model,
                    request.dimensions,
                    provider_name or provider.name,
                    tenant_id,
                )
                try:
                    cached = await redis_client.get(cache_key)
                    if cached:
                        logger.debug(f"Cache hit for embedding: {cache_key}")
                        # Try to decrypt if encryption is enabled, otherwise parse as JSON
                        data = self._decrypt_cache_data(cached)
                        latency_ms = (time.monotonic() - start_time) * 1000
                        
                        # Log cache hit audit event
                        audit_logger.log_embedding_event(
                            tenant_id=tenant_id,
                            api_key_id=api_key_id,
                            operation=AuditOperation.EMBED_GENERATE,
                            model=resolved_model,
                            dimensions=resolved_dimensions,
                            text=request.text,
                            success=True,
                            status=AuditEventStatus.SUCCESS,
                            latency_ms=latency_ms,
                            token_count=data.get("usage", {}).get("tokens", 0),
                            cost_usd=0.0,  # Cache hits are free
                            metadata={"cache_hit": True},
                        )
                        
                        return EmbeddingResponse(
                            embedding=data["embedding"],
                            provider=ProviderType(data["provider"]),
                            model=data["model"],
                            dimensions=data["dimensions"],
                            usage=data.get("usage", {"tokens": 0}),
                            latency_ms=latency_ms,
                        )
                except Exception as e:
                    logger.warning(f"Cache lookup failed: {e}")

        # Generate embedding
        error_message = None
        response = None
        try:
            response = await provider.embed(
                text=request.text,
                model=request.model,
                dimensions=request.dimensions,
            )
            success = True
            status = AuditEventStatus.SUCCESS
        except Exception as e:
            error_message = str(e)
            success = False
            status = AuditEventStatus.FAILURE
            logger.error(f"Embedding generation failed: {e}")
            raise
        finally:
            # Calculate latency
            latency_ms = (time.monotonic() - start_time) * 1000
            
            # Get token count and cost
            token_count = getattr(response, 'usage', {}).get('tokens', estimated_tokens) if response else estimated_tokens
            cost_usd = self._estimate_embedding_cost(token_count, resolved_model) if success and response else 0.0
            
            # Log audit event
            audit_logger.log_embedding_event(
                tenant_id=tenant_id,
                api_key_id=api_key_id,
                operation=AuditOperation.EMBED_GENERATE,
                model=resolved_model,
                dimensions=resolved_dimensions,
                text=request.text,
                success=success,
                status=status,
                latency_ms=latency_ms,
                error_message=error_message,
                token_count=token_count,
                cost_usd=cost_usd,
                metadata={"cache_hit": cache_hit},
            )
            
            # Record cost if successful
            if success and response and settings.enable_cost_tracking:
                cost_tracker.record_cost(
                    tenant_id=tenant_id,
                    provider=provider_name or provider.name,
                    model=resolved_model,
                    input_tokens=token_count,
                    output_tokens=0,
                    cost=cost_usd,
                )

        # Cache the result if enabled
        if settings.enable_caching and cache_key and redis_client:
            try:
                cache_data = {
                    "embedding": response.embedding,
                    "provider": response.provider.value,
                    "model": response.model,
                    "dimensions": response.dimensions,
                    "usage": response.usage,
                }
                
                # Encrypt cache data if encryption is enabled
                if CRYPTOGRAPHY_AVAILABLE and settings.redis_cache_encryption_key:
                    encrypted_data = self._encrypt_cache_data(cache_data)
                    await redis_client.setex(
                        cache_key,
                        self._cache_ttl,
                        encrypted_data,
                    )
                else:
                    await redis_client.setex(
                        cache_key,
                        self._cache_ttl,
                        json.dumps(cache_data),
                    )
                logger.debug(f"Cached embedding: {cache_key}")
            except Exception as e:
                logger.warning(f"Cache write failed: {e}")

        return response
    
    def _encrypt_cache_data(self, data: Dict[str, Any]) -> str:
        """Encrypt cache data using Fernet symmetric encryption.
        
        Args:
            data: Data to encrypt
            
        Returns:
            Encrypted data as base64 string
        """
        if not CRYPTOGRAPHY_AVAILABLE or not settings.redis_cache_encryption_key:
            return json.dumps(data)
        
        try:
            cipher = Fernet(settings.redis_cache_encryption_key.encode())
            json_bytes = json.dumps(data).encode()
            encrypted = cipher.encrypt(json_bytes)
            return encrypted.decode()
        except Exception as e:
            logger.warning(f"Encryption failed, storing unencrypted: {e}")
            return json.dumps(data)
    
    def _decrypt_cache_data(self, encrypted: str) -> Dict[str, Any]:
        """Decrypt cache data.
        
        Args:
            encrypted: Encrypted data string
            
        Returns:
            Decrypted data dictionary
        """
        if not CRYPTOGRAPHY_AVAILABLE or not settings.redis_cache_encryption_key:
            return json.loads(encrypted)
        
        try:
            cipher = Fernet(settings.redis_cache_encryption_key.encode())
            decrypted = cipher.decrypt(encrypted.encode())
            return json.loads(decrypted)
        except Exception as e:
            logger.warning(f"Decryption failed, trying plain JSON: {e}")
            return json.loads(encrypted)
    
    def _estimate_embedding_cost(self, tokens: int, model: str) -> float:
        """Estimate the cost of an embedding request.
        
        Args:
            tokens: Number of tokens
            model: Model name
            
        Returns:
            Estimated cost in USD
        """
        # OpenAI pricing (as of 2026): $0.02 per 1K tokens for text-embedding-ada-002
        # text-embedding-3-small: $0.02 per 1K tokens
        # text-embedding-3-large: $0.13 per 1K tokens
        
        model_lower = model.lower()
        if "text-embedding-3-large" in model_lower:
            rate_per_1k = 0.13
        elif "text-embedding-3" in model_lower or "text-embedding-ada" in model_lower:
            rate_per_1k = 0.02
        else:
            # Default rate for unknown models (local/Ollama is free)
            rate_per_1k = 0.0
        
        return (tokens / 1000.0) * rate_per_1k

    async def normalize_dimensions(
        self,
        embedding: list[float],
        target_dimensions: int,
    ) -> list[float]:
        """Normalize embedding to target dimensions.

        If the embedding is smaller than target, pad with zeros.
        If larger, truncate to target.

        Args:
            embedding: The embedding vector
            target_dimensions: Target number of dimensions

        Returns:
            Normalized embedding
        """
        if len(embedding) == target_dimensions:
            return embedding

        if len(embedding) > target_dimensions:
            return embedding[:target_dimensions]

        # Pad with zeros
        return embedding + [0.0] * (target_dimensions - len(embedding))

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global embeddings service instance
_embeddings_service: Optional[EmbeddingsService] = None


def get_embeddings_service() -> EmbeddingsService:
    """Get the global embeddings service instance.

    Returns:
        The EmbeddingsService instance
    """
    global _embeddings_service
    if _embeddings_service is None:
        _embeddings_service = EmbeddingsService()
    return _embeddings_service
