"""Advanced cache service combining strategy, prediction, and invalidation.

This module provides a unified interface for advanced caching operations.
"""

import hashlib
import json
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Callable, Dict, List, Optional
import logging

from .strategy import (
    CacheStrategy,
    get_cache_strategy,
    CacheEntry,
)
from .predictor import (
    CachePredictor,
    get_cache_predictor,
    CachePrediction,
)
from .invalidator import (
    CacheInvalidator,
    get_cache_invalidator,
    InvalidationRule,
)

logger = logging.getLogger(__name__)


@dataclass
class CacheStats:
    """Cache statistics."""
    hits: int = 0
    misses: int = 0
    sets: int = 0
    deletes: int = 0
    evictions: int = 0
    total_bytes: int = 0
    entry_count: int = 0
    hit_rate: float = 0.0
    avg_latency_ms: float = 0.0


class AdvancedCacheService:
    """Advanced caching service with ML-based predictions."""

    def __init__(
        self,
        strategy: Optional[CacheStrategy] = None,
        predictor: Optional[CachePredictor] = None,
        invalidator: Optional[CacheInvalidator] = None,
    ):
        """Initialize the advanced cache service.

        Args:
            strategy: Cache strategy to use
            predictor: Cache predictor to use
            invalidator: Cache invalidator to use
        """
        self._strategy = strategy or get_cache_strategy()
        self._predictor = predictor or get_cache_predictor()
        self._invalidator = invalidator or get_cache_invalidator()
        self._logger = logging.getLogger(__name__)

        # Stats tracking
        self._stats = CacheStats()
        self._latencies: List[float] = []

    def get(
        self,
        key: str,
        function_id: Optional[str] = None,
    ) -> Optional[Any]:
        """Get a value from the cache.

        Args:
            key: The cache key
            function_id: Optional function ID for tracking

        Returns:
            The cached value or None
        """
        start_time = time.time()

        try:
            value = self._strategy.get(key)

            if value is not None:
                self._stats.hits += 1

                # Record access for prediction
                if function_id:
                    output_hash = hashlib.sha256(
                        str(value).encode()
                    ).hexdigest()[:16]
                    self._predictor.record_access(key, function_id, output_hash)
            else:
                self._stats.misses += 1

            # Track latency
            latency = (time.time() - start_time) * 1000
            self._track_latency(latency)

            return value

        except Exception as e:
            self._logger.error(f"Cache get error for {key}: {e}")
            return None

    def set(
        self,
        key: str,
        value: Any,
        ttl: Optional[int] = None,
        tags: Optional[List[str]] = None,
        function_id: Optional[str] = None,
    ) -> bool:
        """Set a value in the cache.

        Args:
            key: The cache key
            value: The value to cache
            ttl: Time to live in seconds
            tags: Tags for the cache entry
            function_id: Function ID for prediction

        Returns:
            True if successful
        """
        start_time = time.time()

        try:
            # Check if we should cache this
            if function_id:
                output_hash = hashlib.sha256(
                    str(value).encode()
                ).hexdigest()[:16]

                if not self._predictor.should_cache(function_id, output_hash):
                    self._logger.debug(f"Skipping cache for {function_id}")
                    return False

            # Set the value
            self._strategy.set(key, value, ttl, tags)
            self._stats.sets += 1

            # Register tags with invalidator
            if tags:
                self._invalidator.register_tags(key, tags)

            # Track latency
            latency = (time.time() - start_time) * 1000
            self._track_latency(latency)

            return True

        except Exception as e:
            self._logger.error(f"Cache set error for {key}: {e}")
            return False

    def delete(self, key: str) -> bool:
        """Delete a value from the cache.

        Args:
            key: The cache key

        Returns:
            True if deleted
        """
        try:
            result = self._strategy.delete(key)
            if result:
                self._stats.deletes += 1
            return result
        except Exception as e:
            self._logger.error(f"Cache delete error for {key}: {e}")
            return False

    def invalidate_by_key(self, key: str) -> List[str]:
        """Invalidate a key and its dependencies.

        Args:
            key: The cache key

        Returns:
            List of invalidated keys
        """
        def delete_callback(k: str) -> bool:
            return self._strategy.delete(k)

        return self._invalidator.invalidate_by_key(key, delete_callback)

    def invalidate_by_tags(self, tags: List[str]) -> List[str]:
        """Invalidate all keys with the given tags.

        Args:
            tags: Tags to invalidate

        Returns:
            List of invalidated keys
        """
        def delete_callback(k: str) -> bool:
            return self._strategy.delete(k)

        return self._invalidator.invalidate_by_tags(tags, delete_callback)

    def invalidate_by_pattern(self, pattern: str) -> List[str]:
        """Invalidate keys matching a pattern.

        Args:
            pattern: Regex pattern

        Returns:
            List of invalidated keys
        """
        def delete_callback(k: str) -> bool:
            return self._strategy.delete(k)

        return self._invalidator.invalidate_by_pattern(pattern, delete_callback)

    def warm_cache(
        self,
        predictions: List[CachePrediction],
        fetch_callback: Callable[[str], Any],
    ) -> Dict[str, bool]:
        """Warm the cache based on predictions.

        Args:
            predictions: List of cache predictions
            fetch_callback: Callback to fetch the value for a key

        Returns:
            Dictionary of key -> success
        """
        results = {}

        for prediction in predictions:
            if not prediction.is_valid():
                continue

            try:
                # Fetch the value
                value = fetch_callback(prediction.key)

                if value is not None:
                    # Cache it
                    success = self.set(
                        key=prediction.key,
                        value=value,
                        ttl=prediction.suggested_ttl,
                        tags=prediction.suggested_tags,
                        function_id=prediction.function_id,
                    )
                    results[prediction.key] = success
                else:
                    results[prediction.key] = False

            except Exception as e:
                self._logger.error(
                    f"Error warming cache for {prediction.key}: {e}"
                )
                results[prediction.key] = False

        return results

    def get_predictions(
        self,
        function_id: Optional[str] = None,
        limit: int = 100,
    ) -> List[CachePrediction]:
        """Get cache predictions.

        Args:
            function_id: Optional function ID to filter by
            limit: Maximum number of predictions

        Returns:
            List of CachePrediction
        """
        return self._predictor.predict(function_id, limit)

    def add_invalidation_rule(self, rule: InvalidationRule) -> None:
        """Add an invalidation rule.

        Args:
            rule: The rule to add
        """
        self._invalidator.add_rule(rule)

    def _track_latency(self, latency_ms: float) -> None:
        """Track request latency.

        Args:
            latency_ms: Latency in milliseconds
        """
        self._latencies.append(latency_ms)

        # Keep only recent latencies
        if len(self._latencies) > 1000:
            self._latencies = self._latencies[-1000:]

    def get_stats(self) -> CacheStats:
        """Get cache statistics.

        Returns:
            CacheStats
        """
        # Calculate hit rate
        total = self._stats.hits + self._stats.misses
        if total > 0:
            self._stats.hit_rate = self._stats.hits / total

        # Calculate average latency
        if self._latencies:
            self._stats.avg_latency_ms = sum(self._latencies) / len(self._latencies)

        # Get strategy stats
        try:
            strategy_stats = self._strategy.get_stats()
            self._stats.entry_count = strategy_stats.get("size", 0)
            self._stats.total_bytes = strategy_stats.get("memory_bytes", 0)
        except Exception as e:
            self._logger.error(f"Error getting strategy stats: {e}")

        return self._stats

    def clear(self) -> None:
        """Clear all cache entries."""
        def clear_callback() -> None:
            self._strategy.clear()

        count = self._invalidator.invalidate_all(clear_callback)
        self._logger.info(f"Cleared {count} cache entries")

    @property
    def predictor(self) -> CachePredictor:
        """Get the cache predictor."""
        return self._predictor

    @property
    def invalidator(self) -> CacheInvalidator:
        """Get the cache invalidator."""
        return self._invalidator


# Global service instance
_cache_service: Optional[AdvancedCacheService] = None


def get_cache_service() -> AdvancedCacheService:
    """Get the global advanced cache service instance.

    Returns:
        AdvancedCacheService instance
    """
    global _cache_service
    if _cache_service is None:
        _cache_service = AdvancedCacheService()

    return _cache_service
