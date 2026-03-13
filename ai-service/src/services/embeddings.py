"""Embeddings service for FlyMind AI Service.

This module provides the embeddings service with caching support.
"""

import hashlib
import json
import logging
from typing import Optional

import redis.asyncio as redis

from ..config import settings
from ..models.schemas import EmbeddingRequest, EmbeddingResponse, ProviderType
from ..providers.manager import get_provider_manager


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

    def _get_cache_key(self, text: str, model: str, dimensions: Optional[int]) -> str:
        """Generate a cache key for the embedding request."""
        key_data = {
            "text": text,
            "model": model,
            "dimensions": dimensions,
        }
        key_str = json.dumps(key_data, sort_keys=True)
        return f"embedding:{hashlib.sha256(key_str.encode()).hexdigest()}"

    async def generate_embedding(
        self,
        request: EmbeddingRequest,
    ) -> EmbeddingResponse:
        """Generate an embedding for the given text.

        Args:
            request: The embedding request

        Returns:
            EmbeddingResponse with the embedding vector
        """
        provider_manager = get_provider_manager()

        # Get the provider
        provider_name = request.provider.value if request.provider else None
        provider = provider_manager.get_embedding_provider(provider_name)

        # Check cache if enabled
        cache_key = None
        if settings.enable_caching:
            redis_client = await self.get_redis()
            if redis_client:
                cache_key = self._get_cache_key(
                    request.text,
                    request.model or provider.model,
                    request.dimensions,
                )
                try:
                    cached = await redis_client.get(cache_key)
                    if cached:
                        logger.debug(f"Cache hit for embedding: {cache_key}")
                        data = json.loads(cached)
                        return EmbeddingResponse(
                            embedding=data["embedding"],
                            provider=ProviderType(data["provider"]),
                            model=data["model"],
                            dimensions=data["dimensions"],
                            usage=data.get("usage", {"tokens": 0}),
                            latency_ms=0.0,  # Cache hit
                        )
                except Exception as e:
                    logger.warning(f"Cache lookup failed: {e}")

        # Generate embedding
        response = await provider.embed(
            text=request.text,
            model=request.model,
            dimensions=request.dimensions,
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
                await redis_client.setex(
                    cache_key,
                    self._cache_ttl,
                    json.dumps(cache_data),
                )
                logger.debug(f"Cached embedding: {cache_key}")
            except Exception as e:
                logger.warning(f"Cache write failed: {e}")

        return response

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
