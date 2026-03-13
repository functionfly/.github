"""Cache strategy selection and management.

This module provides different caching strategies including:
- LRU (Least Recently Used)
- TTL (Time To Live)
- ML-predicted caching
"""

import time
import hashlib
from abc import ABC, abstractmethod
from enum import Enum
from typing import Any, Dict, List, Optional, Tuple
from dataclasses import dataclass, field
import logging

logger = logging.getLogger(__name__)


class CacheStrategyType(str, Enum):
    """Types of cache strategies."""
    LRU = "lru"
    TTL = "ttl"
    ML_PREDICTED = "ml_predicted"
    ADAPTIVE = "adaptive"


@dataclass
class CacheEntry:
    """A single cache entry."""
    key: str
    value: Any
    created_at: float = field(default_factory=time.time)
    last_accessed: float = field(default_factory=time.time)
    access_count: int = 0
    ttl: Optional[int] = None  # seconds, None means no expiry
    tags: List[str] = field(default_factory=list)
    size_bytes: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)

    def is_expired(self) -> bool:
        """Check if the entry has expired.

        Returns:
            True if expired, False otherwise
        """
        if self.ttl is None:
            return False
        return time.time() > (self.created_at + self.ttl)

    def touch(self) -> None:
        """Update the last accessed time."""
        self.last_accessed = time.time()
        self.access_count += 1


class CacheStrategy(ABC):
    """Abstract base class for cache strategies."""

    @abstractmethod
    def get(self, key: str) -> Optional[Any]:
        """Get a value from the cache.

        Args:
            key: The cache key

        Returns:
            The cached value or None if not found
        """
        pass

    @abstractmethod
    def set(self, key: str, value: Any, ttl: Optional[int] = None, tags: Optional[List[str]] = None) -> None:
        """Set a value in the cache.

        Args:
            key: The cache key
            value: The value to cache
            ttl: Time to live in seconds
            tags: Tags for the cache entry
        """
        pass

    @abstractmethod
    def delete(self, key: str) -> bool:
        """Delete a value from the cache.

        Args:
            key: The cache key

        Returns:
            True if deleted, False if not found
        """
        pass

    @abstractmethod
    def clear(self) -> None:
        """Clear all entries from the cache."""
        pass

    @abstractmethod
    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics.

        Returns:
            Dictionary with cache statistics
        """
        pass

    @abstractmethod
    def evict_if_needed(self, required_space: int) -> int:
        """Evict entries if needed to make space.

        Args:
            required_space: Required space in bytes

        Returns:
            Number of bytes freed
        """
        pass


class LRUCacheStrategy(CacheStrategy):
    """Least Recently Used cache strategy."""

    def __init__(self, max_size: int = 1000, max_memory_bytes: int = 100 * 1024 * 1024):
        """Initialize LRU cache.

        Args:
            max_size: Maximum number of entries
            max_memory_bytes: Maximum memory in bytes
        """
        self._cache: Dict[str, CacheEntry] = {}
        self._max_size = max_size
        self._max_memory_bytes = max_memory_bytes
        self._current_memory = 0
        self._hits = 0
        self._misses = 0

    def get(self, key: str) -> Optional[Any]:
        """Get a value from the cache."""
        entry = self._cache.get(key)
        if entry is None:
            self._misses += 1
            return None

        if entry.is_expired():
            self._delete_entry(key)
            self._misses += 1
            return None

        entry.touch()
        self._hits += 1
        return entry.value

    def set(self, key: str, value: Any, ttl: Optional[int] = None, tags: Optional[List[str]] = None) -> None:
        """Set a value in the cache."""
        # Calculate size
        import sys
        size = sys.getsizeof(str(value))

        # Evict if needed
        self.evict_if_needed(size)

        # Create entry
        entry = CacheEntry(
            key=key,
            value=value,
            ttl=ttl,
            tags=tags or [],
            size_bytes=size,
        )

        # Add to cache
        old_entry = self._cache.get(key)
        if old_entry:
            self._current_memory -= old_entry.size_bytes

        self._cache[key] = entry
        self._current_memory += size

    def delete(self, key: str) -> bool:
        """Delete a value from the cache."""
        return self._delete_entry(key)

    def _delete_entry(self, key: str) -> bool:
        """Delete an entry and update memory."""
        entry = self._cache.pop(key, None)
        if entry:
            self._current_memory -= entry.size_bytes
            return True
        return False

    def clear(self) -> None:
        """Clear all entries."""
        self._cache.clear()
        self._current_memory = 0
        self._hits = 0
        self._misses = 0

    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics."""
        total = self._hits + self._misses
        hit_rate = self._hits / total if total > 0 else 0

        return {
            "strategy": "lru",
            "size": len(self._cache),
            "max_size": self._max_size,
            "memory_bytes": self._current_memory,
            "max_memory_bytes": self._max_memory_bytes,
            "hits": self._hits,
            "misses": self._misses,
            "hit_rate": hit_rate,
        }

    def evict_if_needed(self, required_space: int) -> int:
        """Evict entries if needed."""
        freed = 0

        # Check size limit
        while len(self._cache) >= self._max_size:
            self._evict_lru()
            freed += 1

        # Check memory limit
        while self._current_memory + required_space > self._max_memory_bytes and self._cache:
            self._evict_lru()
            freed += 1

        return freed

    def _evict_lru(self) -> None:
        """Evict the least recently used entry."""
        if not self._cache:
            return

        # Find LRU entry
        lru_key = min(
            self._cache.keys(),
            key=lambda k: self._cache[k].last_accessed
        )

        self._delete_entry(lru_key)


class TTLCacheStrategy(CacheStrategy):
    """Time To Live cache strategy."""

    def __init__(self, default_ttl: int = 3600, max_entries: int = 10000):
        """Initialize TTL cache.

        Args:
            default_ttl: Default time to live in seconds
            max_entries: Maximum number of entries
        """
        self._cache: Dict[str, CacheEntry] = {}
        self._default_ttl = default_ttl
        self._max_entries = max_entries
        self._hits = 0
        self._misses = 0
        self._evictions = 0

    def get(self, key: str) -> Optional[Any]:
        """Get a value from the cache."""
        entry = self._cache.get(key)
        if entry is None:
            self._misses += 1
            return None

        if entry.is_expired():
            self._cache.pop(key, None)
            self._misses += 1
            return None

        entry.touch()
        self._hits += 1
        return entry.value

    def set(self, key: str, value: Any, ttl: Optional[int] = None, tags: Optional[List[str]] = None) -> None:
        """Set a value in the cache."""
        import sys
        size = sys.getsizeof(str(value))

        # Use default TTL if not specified
        effective_ttl = ttl if ttl is not None else self._default_ttl

        entry = CacheEntry(
            key=key,
            value=value,
            ttl=effective_ttl,
            tags=tags or [],
            size_bytes=size,
        )

        self._cache[key] = entry

        # Evict if over limit
        if len(self._cache) > self._max_entries:
            self._evict_expired()

    def delete(self, key: str) -> bool:
        """Delete a value from the cache."""
        entry = self._cache.pop(key, None)
        return entry is not None

    def clear(self) -> None:
        """Clear all entries."""
        self._cache.clear()
        self._hits = 0
        self._misses = 0
        self._evictions = 0

    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics."""
        total = self._hits + self._misses
        hit_rate = self._hits / total if total > 0 else 0

        return {
            "strategy": "ttl",
            "size": len(self._cache),
            "max_entries": self._max_entries,
            "default_ttl": self._default_ttl,
            "hits": self._hits,
            "misses": self._misses,
            "evictions": self._evictions,
            "hit_rate": hit_rate,
        }

    def evict_if_needed(self, required_space: int = 0) -> int:
        """Evict expired entries."""
        return self._evict_expired()

    def _evict_expired(self) -> int:
        """Evict all expired entries."""
        expired_keys = [
            key for key, entry in self._cache.items()
            if entry.is_expired()
        ]

        for key in expired_keys:
            self._cache.pop(key, None)
            self._evictions += 1

        return len(expired_keys)

    def cleanup_expired(self) -> int:
        """Manually trigger cleanup of expired entries."""
        return self._evict_expired()


class AdaptiveCacheStrategy(CacheStrategy):
    """Adaptive cache strategy that switches between LRU and TTL based on patterns."""

    def __init__(
        self,
        lru_strategy: Optional[LRUCacheStrategy] = None,
        ttl_strategy: Optional[TTLCacheStrategy] = None,
        switch_threshold: float = 0.7,
    ):
        """Initialize adaptive cache.

        Args:
            lru_strategy: LRU strategy to use
            ttl_strategy: TTL strategy to use
            switch_threshold: Hit rate threshold to switch strategies
        """
        self._lru = lru_strategy or LRUCacheStrategy()
        self._ttl = ttl_strategy or TTLCacheStrategy()
        self._switch_threshold = switch_threshold
        self._current_strategy: CacheStrategy = self._lru
        self._strategy_switches = 0

    @property
    def current_strategy(self) -> str:
        """Get the current strategy name."""
        if self._current_strategy == self._lru:
            return "lru"
        return "ttl"

    def get(self, key: str) -> Optional[Any]:
        """Get a value from the cache."""
        return self._current_strategy.get(key)

    def set(self, key: str, value: Any, ttl: Optional[int] = None, tags: Optional[List[str]] = None) -> None:
        """Set a value in the cache."""
        self._current_strategy.set(key, value, ttl, tags)

    def delete(self, key: str) -> bool:
        """Delete a value from the cache."""
        deleted = self._lru.delete(key)
        deleted = self._ttl.delete(key) or deleted
        return deleted

    def clear(self) -> None:
        """Clear all entries."""
        self._lru.clear()
        self._ttl.clear()

    def get_stats(self) -> Dict[str, Any]:
        """Get cache statistics."""
        lru_stats = self._lru.get_stats()
        ttl_stats = self._ttl.get_stats()

        # Check if we should switch
        hit_rate = lru_stats.get("hit_rate", 0)

        if hit_rate < self._switch_threshold and self._current_strategy == self._lru:
            self._current_strategy = self._ttl
            self._strategy_switches += 1
        elif hit_rate >= self._switch_threshold and self._current_strategy == self._ttl:
            self._current_strategy = self._lru
            self._strategy_switches += 1

        return {
            "strategy": "adaptive",
            "current": self.current_strategy,
            "switch_threshold": self._switch_threshold,
            "strategy_switches": self._strategy_switches,
            "lru": lru_stats,
            "ttl": ttl_stats,
        }

    def evict_if_needed(self, required_space: int) -> int:
        """Evict entries if needed."""
        return self._current_strategy.evict_if_needed(required_space)


# Global strategy instance
_cache_strategy: Optional[CacheStrategy] = None


def get_cache_strategy() -> CacheStrategy:
    """Get the global cache strategy instance.

    Returns:
        CacheStrategy instance
    """
    global _cache_strategy
    if _cache_strategy is None:
        _cache_strategy = AdaptiveCacheStrategy()

    return _cache_strategy


def set_cache_strategy(strategy: CacheStrategy) -> None:
    """Set the global cache strategy.

    Args:
        strategy: The cache strategy to use
    """
    global _cache_strategy
    _cache_strategy = strategy
