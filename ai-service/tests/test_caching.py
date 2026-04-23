"""Tests for caching service, strategy, and predictor."""

import time
import pytest

from src.services.caching.strategy import (
    LRUCacheStrategy,
    TTLCacheStrategy,
    AdaptiveCacheStrategy,
    CacheEntry,
    get_cache_strategy,
    set_cache_strategy,
)
from src.services.caching.predictor import (
    CachePredictor,
    CachePrediction,
    AccessPattern,
    DeterminismLevel,
    get_cache_predictor,
)
from src.services.caching.service import (
    AdvancedCacheService,
    CacheStats,
)


# =============================================================================
# CacheEntry tests
# =============================================================================


class TestCacheEntry:
    """Tests for CacheEntry."""

    def test_entry_not_expired_without_ttl(self):
        """Entry without TTL should never expire."""
        entry = CacheEntry(key="k", value="v", ttl=None)
        assert entry.is_expired() is False

    def test_entry_expired_after_ttl(self):
        """Entry should expire after TTL."""
        entry = CacheEntry(key="k", value="v", ttl=1, created_at=time.time() - 2)
        assert entry.is_expired() is True

    def test_entry_not_expired_before_ttl(self):
        """Entry should not expire before TTL."""
        entry = CacheEntry(key="k", value="v", ttl=60, created_at=time.time())
        assert entry.is_expired() is False

    def test_touch_updates_access_count(self):
        """touch() should increment access count."""
        entry = CacheEntry(key="k", value="v")
        assert entry.access_count == 0
        entry.touch()
        assert entry.access_count == 1
        entry.touch()
        assert entry.access_count == 2


# =============================================================================
# LRUCacheStrategy tests
# =============================================================================


class TestLRUCacheStrategy:
    """Tests for LRUCacheStrategy."""

    def test_set_and_get(self):
        """Should store and retrieve values."""
        cache = LRUCacheStrategy()
        cache.set("key1", "value1")
        assert cache.get("key1") == "value1"

    def test_get_missing_returns_none(self):
        """Getting a missing key should return None."""
        cache = LRUCacheStrategy()
        assert cache.get("missing") is None

    def test_delete(self):
        """Should delete entries."""
        cache = LRUCacheStrategy()
        cache.set("key1", "value1")
        assert cache.delete("key1") is True
        assert cache.get("key1") is None

    def test_delete_missing_returns_false(self):
        """Deleting a missing key should return False."""
        cache = LRUCacheStrategy()
        assert cache.delete("missing") is False

    def test_clear(self):
        """clear() should remove all entries."""
        cache = LRUCacheStrategy()
        cache.set("k1", "v1")
        cache.set("k2", "v2")
        cache.clear()
        assert cache.get("k1") is None
        assert cache.get("k2") is None

    def test_eviction_on_max_size(self):
        """Should evict LRU entry when max size reached."""
        cache = LRUCacheStrategy(max_size=3)
        cache.set("a", 1)
        cache.set("b", 2)
        cache.set("c", 3)
        # Access 'a' to make it recently used
        cache.get("a")
        cache.set("d", 4)  # Should evict 'b' (least recently used)

        assert cache.get("a") == 1
        assert cache.get("b") is None  # Evicted
        assert cache.get("c") == 3
        assert cache.get("d") == 4

    def test_get_stats(self):
        """get_stats() should return correct statistics."""
        cache = LRUCacheStrategy()
        cache.set("k1", "v1")
        cache.get("k1")  # Hit
        cache.get("k2")  # Miss

        stats = cache.get_stats()
        assert stats["strategy"] == "lru"
        assert stats["size"] == 1
        assert stats["hits"] == 1
        assert stats["misses"] == 1
        assert stats["hit_rate"] == 0.5

    def test_expired_entry_returns_none(self):
        """Expired entries should return None on get."""
        cache = LRUCacheStrategy()
        entry = CacheEntry(key="k", value="v", ttl=1, created_at=time.time() - 2)
        cache._cache["k"] = entry

        assert cache.get("k") is None


# =============================================================================
# TTLCacheStrategy tests
# =============================================================================


class TestTTLCacheStrategy:
    """Tests for TTLCacheStrategy."""

    def test_set_and_get(self):
        """Should store and retrieve values."""
        cache = TTLCacheStrategy()
        cache.set("key1", "value1")
        assert cache.get("key1") == "value1"

    def test_get_missing_returns_none(self):
        """Getting a missing key should return None."""
        cache = TTLCacheStrategy()
        assert cache.get("missing") is None

    def test_delete(self):
        """Should delete entries."""
        cache = TTLCacheStrategy()
        cache.set("key1", "value1")
        assert cache.delete("key1") is True
        assert cache.get("key1") is None

    def test_default_ttl_applied(self):
        """Default TTL should be applied when not specified."""
        cache = TTLCacheStrategy(default_ttl=3600)
        cache.set("k", "v")
        entry = cache._cache["k"]
        assert entry.ttl == 3600

    def test_custom_ttl_overrides_default(self):
        """Custom TTL should override default."""
        cache = TTLCacheStrategy(default_ttl=3600)
        cache.set("k", "v", ttl=60)
        entry = cache._cache["k"]
        assert entry.ttl == 60

    def test_cleanup_expired(self):
        """cleanup_expired should remove expired entries."""
        cache = TTLCacheStrategy()
        entry = CacheEntry(key="old", value="v", ttl=1, created_at=time.time() - 2)
        cache._cache["old"] = entry
        cache.set("new", "v")

        count = cache.cleanup_expired()
        assert count >= 1
        assert cache.get("old") is None
        assert cache.get("new") == "v"

    def test_get_stats(self):
        """get_stats() should return correct statistics."""
        cache = TTLCacheStrategy()
        cache.set("k1", "v1")
        cache.get("k1")
        cache.get("missing")

        stats = cache.get_stats()
        assert stats["strategy"] == "ttl"
        assert stats["hits"] == 1
        assert stats["misses"] == 1


# =============================================================================
# AdaptiveCacheStrategy tests
# =============================================================================


class TestAdaptiveCacheStrategy:
    """Tests for AdaptiveCacheStrategy."""

    def test_delegates_to_lru_by_default(self):
        """Should delegate to LRU strategy by default."""
        strategy = AdaptiveCacheStrategy()
        assert strategy.current_strategy == "lru"

    def test_set_and_get(self):
        """Should store and retrieve values."""
        strategy = AdaptiveCacheStrategy()
        strategy.set("k", "v")
        assert strategy.get("k") == "v"

    def test_delete_from_both(self):
        """delete should remove from both LRU and TTL."""
        strategy = AdaptiveCacheStrategy()
        strategy.set("k", "v")
        strategy.delete("k")
        assert strategy.get("k") is None

    def test_clear_clears_both(self):
        """clear should remove from both strategies."""
        strategy = AdaptiveCacheStrategy()
        strategy.set("k1", "v1")
        strategy.clear()
        assert strategy.get("k1") is None

    def test_get_stats_includes_both(self):
        """get_stats should include stats from both strategies."""
        strategy = AdaptiveCacheStrategy()
        stats = strategy.get_stats()
        assert "lru" in stats
        assert "ttl" in stats
        assert "current" in stats


# =============================================================================
# CachePredictor tests
# =============================================================================


class TestAccessPattern:
    """Tests for AccessPattern."""

    def test_record_access(self):
        """Recording access should update count and times."""
        pattern = AccessPattern(key="k", function_id="f")
        pattern.record_access()
        assert pattern.access_count == 1
        assert pattern.last_access is not None

    def test_get_avg_interval(self):
        """Average interval should be calculated from recorded accesses."""
        pattern = AccessPattern(key="k", function_id="f")
        now = time.time()
        pattern.record_access(timestamp=now)
        pattern.record_access(timestamp=now + 10)
        pattern.record_access(timestamp=now + 20)

        avg = pattern.get_avg_interval()
        assert avg is not None
        assert abs(avg - 10.0) < 1.0

    def test_get_avg_interval_no_data(self):
        """Average interval should be None with insufficient data."""
        pattern = AccessPattern(key="k", function_id="f")
        assert pattern.get_avg_interval() is None


class TestCachePredictor:
    """Tests for CachePredictor."""

    def test_record_access_creates_pattern(self):
        """Recording access should create a pattern."""
        predictor = CachePredictor()
        predictor.record_access("key1", "func1")

        assert "key1" in predictor._patterns

    def test_predict_returns_empty_for_new_patterns(self):
        """Predictions should be empty for patterns with <2 accesses."""
        predictor = CachePredictor()
        predictor.record_access("key1", "func1")

        predictions = predictor.predict()
        assert len(predictions) == 0

    def test_predict_returns_predictions_for_frequent_access(self):
        """Frequent access patterns should generate predictions."""
        predictor = CachePredictor(prediction_window_minutes=10, min_confidence=0.1)
        past = time.time() - 600  # 10 minutes ago

        # Create pattern with sufficient interval data and access rate
        pattern = AccessPattern(key="key1", function_id="func1")
        pattern.access_count = 50
        for i in range(50):
            pattern.access_times.append(past + i * 5)
        for i in range(49):
            pattern.access_intervals.append(5.0)
        import datetime

        pattern.first_access = datetime.datetime.fromtimestamp(past)
        pattern.last_access = datetime.datetime.fromtimestamp(past + 245)
        predictor._patterns["key1"] = pattern

        predictions = predictor.predict()
        assert len(predictions) > 0

    def test_should_cache_deterministic(self):
        """Deterministic functions should be cached."""
        predictor = CachePredictor()
        # Record many accesses with same output hash to build determinism score
        for _ in range(10):
            predictor.record_access("key1", "func1", output_hash="hash1")

        assert predictor.should_cache("func1") is True

    def test_should_not_cache_non_deterministic(self):
        """Non-deterministic functions should not be cached."""
        predictor = CachePredictor()
        # Record many different outputs
        for i in range(20):
            predictor.record_access("key1", "func1", output_hash=f"hash{i}")

        assert predictor.should_cache("func1") is False

    def test_get_stats(self):
        """get_stats should return pattern statistics."""
        predictor = CachePredictor()
        predictor.record_access("k1", "f1")
        predictor.record_access("k2", "f2")

        stats = predictor.get_stats()
        assert stats["total_patterns"] == 2

    def test_prediction_is_valid(self):
        """CachePrediction validity should work correctly."""
        from datetime import datetime, timedelta

        prediction = CachePrediction(
            key="k",
            function_id="f",
            predicted_accesses=10,
            confidence=0.8,
            valid_until=datetime.utcnow() + timedelta(minutes=5),
        )
        assert prediction.is_valid() is True

        expired = CachePrediction(
            key="k",
            function_id="f",
            predicted_accesses=10,
            confidence=0.8,
            valid_until=datetime.utcnow() - timedelta(minutes=1),
        )
        assert expired.is_valid() is False


# =============================================================================
# AdvancedCacheService tests
# =============================================================================


class TestAdvancedCacheService:
    """Tests for AdvancedCacheService."""

    def test_set_and_get(self):
        """Should store and retrieve values."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("key1", "value1")
        assert service.get("key1") == "value1"

    def test_get_miss_returns_none(self):
        """Cache miss should return None."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        assert service.get("missing") is None

    def test_cache_hit_tracking(self):
        """Cache hits should be tracked in stats."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("k", "v")
        service.get("k")  # Hit
        service.get("missing")  # Miss

        stats = service.get_stats()
        assert stats.hits == 1
        assert stats.misses == 1
        assert stats.hit_rate == 0.5

    def test_delete(self):
        """delete should remove entry."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("k", "v")
        assert service.delete("k") is True
        assert service.get("k") is None

    def test_invalidate_by_key(self):
        """invalidate_by_key should remove entry and dependencies."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("k", "v")
        invalidated = service.invalidate_by_key("k")
        assert "k" in invalidated

    def test_invalidate_by_tags(self):
        """invalidate_by_tags should remove all entries with given tags."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("k1", "v1", tags=["tag1"])
        service.set("k2", "v2", tags=["tag2"])
        service.set("k3", "v3", tags=["tag1", "tag2"])

        invalidated = service.invalidate_by_tags(["tag1"])
        assert "k1" in invalidated
        assert "k3" in invalidated
        assert "k2" not in invalidated

    def test_clear(self):
        """clear should remove all entries."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("k1", "v1")
        service.set("k2", "v2")
        service.clear()
        assert service.get("k1") is None
        assert service.get("k2") is None

    def test_set_records_access_for_prediction(self):
        """get with function_id should record access for prediction."""
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=CachePredictor(),
        )
        service.set("k", "v", function_id="func1")
        service.get("k", function_id="func1")

        assert "k" in service.predictor._patterns

    def test_should_not_cache_when_predictor_rejects(self):
        """set should return False when predictor says not to cache."""
        predictor = CachePredictor()
        service = AdvancedCacheService(
            strategy=LRUCacheStrategy(),
            predictor=predictor,
        )

        # Make predictor reject caching for this function
        for i in range(20):
            predictor.record_access("k", "func1", output_hash=f"hash{i}")

        result = service.set("k", "v", function_id="func1")
        assert result is False
