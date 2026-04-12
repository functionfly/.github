"""Caching layer for function generation.

Caches generated functions to avoid repeated AI calls for similar requests.
Uses semantic hashing for cache keys.
"""

import hashlib
import json
import logging
from typing import Optional, Dict, Any, Tuple
from dataclasses import dataclass
from datetime import datetime, timedelta

from ...services.redis_client import get_redis_client
from ...config import settings

logger = logging.getLogger(__name__)


@dataclass
class CachedGeneration:
    """A cached function generation result."""
    cache_key: str
    code: str
    manifest: Dict[str, Any]
    explanation: str
    runtime: str
    complexity: str
    created_at: datetime
    ttl_seconds: int
    hit_count: int


class GenerationCache:
    """Cache for function generations."""

    CACHE_PREFIX = "funcgen:v1"
    DEFAULT_TTL = 3600 * 24 * 7  # 7 days
    SIMILARITY_THRESHOLD = 0.92  # Fuzzy match threshold

    def __init__(self):
        self._redis = None
        self._local_cache: Dict[str, CachedGeneration] = {}
        self._max_local_size = 100

    async def _get_redis(self):
        """Get Redis connection."""
        if self._redis is None:
            self._redis = get_redis_client()
        return self._redis

    def _generate_key(
        self,
        description: str,
        runtime: str,
        constraints: Optional[str] = None,
    ) -> str:
        """Generate cache key from request parameters.

        Uses semantic normalization to increase cache hits.
        """
        # Normalize the description
        normalized = self._normalize_description(description)

        # Create hash components
        key_parts = [
            runtime.lower().strip(),
            normalized.lower().strip(),
            (constraints or "").lower().strip(),
        ]

        # Create deterministic hash
        key_string = "|".join(key_parts)
        hash_val = hashlib.sha256(key_string.encode()).hexdigest()[:24]

        return f"{self.CACHE_PREFIX}:{hash_val}"

    def _normalize_description(self, description: str) -> str:
        """Normalize description for better cache matching."""
        # Convert to lowercase
        normalized = description.lower()

        # Remove articles and filler words
        filler_words = [
            "a", "an", "the", "please", "can you", "could you",
            "i need", "i want", "create", "make", "build", "write",
            "that", "which", "who", "this", "these", "those"
        ]
        for word in filler_words:
            normalized = normalized.replace(f" {word} ", " ")

        # Normalize whitespace
        normalized = " ".join(normalized.split())

        # Sort key phrases (for requests like "summarize and email" vs "email and summarize")
        if " and " in normalized:
            parts = [p.strip() for p in normalized.split(" and ")]
            parts.sort()
            normalized = " and ".join(parts)

        return normalized.strip()

    def _fuzzy_match_key(self, description: str, runtime: str) -> Optional[str]:
        """Try to find a similar cached key.

        Uses local cache for fast fuzzy matching.
        """
        normalized = self._normalize_description(description)

        best_match = None
        best_score = 0.0

        for cached in self._local_cache.values():
            if cached.runtime != runtime:
                continue

            # Calculate simple similarity
            score = self._calculate_similarity(normalized, cached.cache_key)
            if score > best_score and score >= self.SIMILARITY_THRESHOLD:
                best_score = score
                best_match = cached.cache_key

        return best_match

    def _calculate_similarity(self, desc1: str, key2: str) -> float:
        """Calculate simple similarity between descriptions."""
        # Extract from cache key (we store normalized in key generation)
        # For now, use exact match on normalized
        key_normalized = key2.split(":")[-1]  # Get hash part

        # Simple comparison - in production would use embeddings
        return 1.0 if desc1 in key2 or key2 in desc1 else 0.0

    async def get(
        self,
        description: str,
        runtime: str,
        constraints: Optional[str] = None,
    ) -> Optional[CachedGeneration]:
        """Get cached generation if available.

        Args:
            description: Function description
            runtime: Target runtime
            constraints: Optional constraints

        Returns:
            Cached generation or None
        """
        cache_key = self._generate_key(description, runtime, constraints)

        # Try local cache first
        if cache_key in self._local_cache:
            cached = self._local_cache[cache_key]
            cached.hit_count += 1
            logger.debug(f"Local cache hit: {cache_key}")
            return cached

        # Try Redis
        try:
            redis = await self._get_redis()
            if redis:
                data = await redis.get(cache_key)
                if data:
                    parsed = json.loads(data)
                    cached = CachedGeneration(
                        cache_key=cache_key,
                        code=parsed["code"],
                        manifest=parsed["manifest"],
                        explanation=parsed["explanation"],
                        runtime=parsed["runtime"],
                        complexity=parsed["complexity"],
                        created_at=datetime.fromisoformat(parsed["created_at"]),
                        ttl_seconds=parsed.get("ttl_seconds", self.DEFAULT_TTL),
                        hit_count=1,
                    )

                    # Add to local cache
                    self._add_to_local(cached)
                    logger.info(f"Redis cache hit: {cache_key}")
                    return cached

        except Exception as e:
            logger.warning(f"Cache retrieval failed: {e}")

        return None

    async def set(
        self,
        description: str,
        runtime: str,
        code: str,
        manifest: Dict[str, Any],
        explanation: str,
        complexity: str,
        constraints: Optional[str] = None,
        ttl: Optional[int] = None,
    ) -> str:
        """Cache a generation result.

        Args:
            description: Function description
            runtime: Target runtime
            code: Generated code
            manifest: Function manifest
            explanation: Generation explanation
            complexity: Complexity estimate
            constraints: Optional constraints
            ttl: Cache TTL in seconds

        Returns:
            Cache key
        """
        cache_key = self._generate_key(description, runtime, constraints)
        ttl = ttl or self.DEFAULT_TTL

        cached = CachedGeneration(
            cache_key=cache_key,
            code=code,
            manifest=manifest,
            explanation=explanation,
            runtime=runtime,
            complexity=complexity,
            created_at=datetime.utcnow(),
            ttl_seconds=ttl,
            hit_count=0,
        )

        # Add to local cache
        self._add_to_local(cached)

        # Store in Redis
        try:
            redis = await self._get_redis()
            if redis:
                data = json.dumps({
                    "code": code,
                    "manifest": manifest,
                    "explanation": explanation,
                    "runtime": runtime,
                    "complexity": complexity,
                    "created_at": datetime.utcnow().isoformat(),
                    "ttl_seconds": ttl,
                })
                await redis.setex(cache_key, ttl, data)
                logger.debug(f"Cached generation: {cache_key}")

        except Exception as e:
            logger.warning(f"Cache storage failed: {e}")

        return cache_key

    def _add_to_local(self, cached: CachedGeneration) -> None:
        """Add to local cache with LRU eviction."""
        if len(self._local_cache) >= self._max_local_size:
            # Evict least recently used (lowest hit count)
            lru_key = min(
                self._local_cache.keys(),
                key=lambda k: self._local_cache[k].hit_count
            )
            del self._local_cache[lru_key]

        self._local_cache[cached.cache_key] = cached

    async def invalidate(
        self,
        description: str,
        runtime: str,
        constraints: Optional[str] = None,
    ) -> bool:
        """Invalidate a cached entry."""
        cache_key = self._generate_key(description, runtime, constraints)

        # Remove from local
        if cache_key in self._local_cache:
            del self._local_cache[cache_key]

        # Remove from Redis
        try:
            redis = await self._get_redis()
            if redis:
                await redis.delete(cache_key)
                return True
        except Exception as e:
            logger.warning(f"Cache invalidation failed: {e}")

        return False

    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics."""
        total_hits = sum(c.hit_count for c in self._local_cache.values())
        return {
            "local_entries": len(self._local_cache),
            "total_hits": total_hits,
            "estimated_savings_usd": total_hits * 0.05,  # Rough estimate
        }


class GenerationCostTracker:
    """Track generation costs and savings from caching."""

    def __init__(self):
        self._total_generations = 0
        self._cached_generations = 0
        self._total_cost_usd = 0.0
        self._savings_usd = 0.0

    def record_generation(
        self,
        tier: str,
        model: str,
        tokens_in: int,
        tokens_out: int,
        cost_usd: float,
        was_cached: bool = False,
    ) -> None:
        """Record a generation event."""
        self._total_generations += 1

        if was_cached:
            self._cached_generations += 1
            self._savings_usd += cost_usd
        else:
            self._total_cost_usd += cost_usd

    def get_stats(self) -> Dict[str, Any]:
        """Get cost tracking statistics."""
        if self._total_generations == 0:
            cache_hit_rate = 0.0
        else:
            cache_hit_rate = self._cached_generations / self._total_generations

        return {
            "total_generations": self._total_generations,
            "cached_generations": self._cached_generations,
            "cache_hit_rate": cache_hit_rate,
            "total_cost_usd": round(self._total_cost_usd, 4),
            "savings_usd": round(self._savings_usd, 4),
            "effective_cost_usd": round(self._total_cost_usd - self._savings_usd, 4),
        }


# Global instances
_cache: Optional[GenerationCache] = None
_cost_tracker: Optional[GenerationCostTracker] = None


def get_generation_cache() -> GenerationCache:
    """Get global generation cache."""
    global _cache
    if _cache is None:
        _cache = GenerationCache()
    return _cache


def get_cost_tracker() -> GenerationCostTracker:
    """Get global cost tracker."""
    global _cost_tracker
    if _cost_tracker is None:
        _cost_tracker = GenerationCostTracker()
    return _cost_tracker
