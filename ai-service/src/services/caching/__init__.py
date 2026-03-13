"""Advanced Caching Service for FlyMind AI Service.

This module provides intelligent caching with ML-based predictions:
- Cache strategy selection (LRU, TTL, ML-predicted)
- Prediction-based cache warming
- Intelligent cache invalidation
"""

from .strategy import (
    CacheStrategy,
    CacheStrategyType,
    get_cache_strategy,
    set_cache_strategy,
)
from .predictor import (
    CachePredictor,
    CachePrediction,
    get_cache_predictor,
)
from .invalidator import (
    CacheInvalidator,
    InvalidationRule,
    get_cache_invalidator,
)
from .service import (
    AdvancedCacheService,
    CacheEntry,
    CacheStats,
    get_cache_service,
)

__all__ = [
    "CacheStrategy",
    "CacheStrategyType",
    "get_cache_strategy",
    "set_cache_strategy",
    "CachePredictor",
    "CachePrediction",
    "get_cache_predictor",
    "CacheInvalidator",
    "InvalidationRule",
    "get_cache_invalidator",
    "AdvancedCacheService",
    "CacheEntry",
    "CacheStats",
    "get_cache_service",
]
