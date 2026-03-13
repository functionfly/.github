"""Cache prediction service for intelligent cache warming.

This module predicts which functions/results should be cached based on:
- Historical access patterns
- Time-based patterns (time of day, day of week)
- Phase 2 forecasting data
- Determinism analysis
"""

import time
import hashlib
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from typing import Any, Dict, List, Optional, Set
from collections import defaultdict
import logging

logger = logging.getLogger(__name__)


class DeterminismLevel(str, Enum):
    """Level of determinism in function outputs."""
    DETERMINISTIC = "deterministic"  # Same input = same output
    PARTIALLY_DETERMINISTIC = "partially_deterministic"  # Sometimes same output
    NON_DETERMINISTIC = "non_deterministic"  # Different output each time


@dataclass
class CachePrediction:
    """A prediction for cache warming."""
    key: str
    function_id: str
    predicted_accesses: int
    confidence: float = field(ge=0.0, le=1.0)
    suggested_ttl: int = 3600  # seconds
    suggested_tags: List[str] = field(default_factory=list)
    predicted_at: datetime = field(default_factory=datetime.utcnow)
    valid_until: datetime = field(default_factory=lambda: datetime.utcnow() + timedelta(minutes=10))

    def is_valid(self) -> bool:
        """Check if prediction is still valid."""
        return datetime.utcnow() < self.valid_until


@dataclass
class AccessPattern:
    """Historical access pattern for a key."""
    key: str
    function_id: str
    access_count: int = 0
    first_access: Optional[datetime] = None
    last_access: Optional[datetime] = None
    access_times: List[float] = field(default_factory=list)
    access_intervals: List[float] = field(default_factory=list)  # seconds between accesses
    is_deterministic: bool = True
    determinism_score: float = 1.0  # 0.0 to 1.0

    def record_access(self, timestamp: Optional[float] = None) -> None:
        """Record an access to this key."""
        now = timestamp or time.time()

        if self.last_access is not None:
            interval = now - self.last_access
            if interval > 0:
                self.access_intervals.append(interval)
                # Keep only last 100 intervals
                if len(self.access_intervals) > 100:
                    self.access_intervals = self.access_intervals[-100:]

        self.access_count += 1
        self.access_times.append(now)

        # Keep only last 1000 access times
        if len(self.access_times) > 1000:
            self.access_times = self.access_times[-1000:]

        if self.first_access is None:
            self.first_access = datetime.fromtimestamp(now)
        self.last_access = datetime.fromtimestamp(now)

    def get_avg_interval(self) -> Optional[float]:
        """Get average access interval in seconds."""
        if not self.access_intervals:
            return None
        return sum(self.access_intervals) / len(self.access_intervals)

    def get_access_rate(self) -> float:
        """Get accesses per minute."""
        if not self.access_times or self.first_access is None:
            return 0.0

        duration_minutes = (datetime.utcnow() - self.first_access).total_seconds() / 60
        if duration_minutes <= 0:
            return 0.0

        return self.access_count / duration_minutes

    def predict_next_access(self) -> Optional[datetime]:
        """Predict when the next access will occur."""
        avg_interval = self.get_avg_interval()
        if avg_interval is None or self.last_access is None:
            return None

        next_ts = self.last_access.timestamp() + avg_interval
        return datetime.fromtimestamp(next_ts)


class CachePredictor:
    """Predicts which cache entries should be warmed."""

    def __init__(
        self,
        prediction_window_minutes: int = 10,
        min_confidence: float = 0.5,
        history_window_hours: int = 24,
    ):
        """Initialize the cache predictor.

        Args:
            prediction_window_minutes: Window for predictions
            min_confidence: Minimum confidence for predictions
            history_window_hours: How long to keep history
        """
        self._patterns: Dict[str, AccessPattern] = {}
        self._prediction_window = prediction_window_minutes * 60  # Convert to seconds
        self._min_confidence = min_confidence
        self._history_window = history_window_hours * 3600  # Convert to seconds
        self._determinism_cache: Dict[str, DeterminismLevel] = {}
        self._predictions: List[CachePrediction] = []
        self._logger = logging.getLogger(__name__)

    def record_access(
        self,
        key: str,
        function_id: str,
        output_hash: Optional[str] = None,
    ) -> None:
        """Record an access to a cache key.

        Args:
            key: The cache key
            function_id: The function ID
            output_hash: Hash of the output (for determinism tracking)
        """
        # Create or update pattern
        if key not in self._patterns:
            self._patterns[key] = AccessPattern(
                key=key,
                function_id=function_id,
            )

        pattern = self._patterns[key]
        pattern.record_access()

        # Update determinism
        if output_hash is not None:
            self._update_determinism(key, output_hash)

        # Clean old patterns
        self._cleanup_old_patterns()

    def _update_determinism(self, key: str, output_hash: str) -> None:
        """Update determinism score for a key.

        Args:
            key: The cache key
            output_hash: Hash of the output
        """
        pattern = self._patterns.get(key)
        if pattern is None:
            return

        # Count unique outputs
        if not hasattr(pattern, '_output_hashes'):
            pattern._output_hashes = []

        pattern._output_hashes.append(output_hash)

        # Keep only recent outputs
        if len(pattern._output_hashes) > 100:
            pattern._output_hashes = pattern._output_hashes[-100:]

        # Calculate uniqueness ratio
        unique_outputs = len(set(pattern._output_hashes))
        total_outputs = len(pattern._output_hashes)

        if total_outputs > 0:
            # Higher ratio = less deterministic
            uniqueness = unique_outputs / total_outputs
            pattern.determinism_score = 1.0 - uniqueness

            if uniqueness > 0.5:
                pattern.is_deterministic = False

    def _cleanup_old_patterns(self) -> None:
        """Remove old patterns that haven't been accessed recently."""
        now = time.time()
        cutoff = now - self._history_window

        old_keys = [
            key for key, pattern in self._patterns.items()
            if pattern.last_access and pattern.last_access.timestamp() < cutoff
        ]

        for key in old_keys:
            del self._patterns[key]

    def predict(
        self,
        function_id: Optional[str] = None,
        limit: int = 100,
    ) -> List[CachePrediction]:
        """Predict which keys should be cached/warmed.

        Args:
            function_id: Optional function ID to filter by
            limit: Maximum number of predictions

        Returns:
            List of cache predictions
        """
        predictions = []

        # Filter patterns
        patterns = self._patterns.values()
        if function_id:
            patterns = [p for p in patterns if p.function_id == function_id]

        for pattern in patterns:
            # Skip non-deterministic patterns (low cache value)
            if pattern.determinism_score < 0.3:
                continue

            # Calculate prediction
            prediction = self._make_prediction(pattern)
            if prediction and prediction.confidence >= self._min_confidence:
                predictions.append(prediction)

        # Sort by confidence and access rate
        predictions.sort(
            key=lambda p: (p.confidence, p.predicted_accesses),
            reverse=True
        )

        # Limit results
        return predictions[:limit]

    def _make_prediction(self, pattern: AccessPattern) -> Optional[CachePrediction]:
        """Make a prediction for a single pattern.

        Args:
            pattern: The access pattern

        Returns:
            CachePrediction or None
        """
        if pattern.access_count < 2:
            return None

        # Calculate access rate
        access_rate = pattern.get_access_rate()
        if access_rate <= 0:
            return None

        # Predict accesses in the window
        window_minutes = self._prediction_window / 60
        predicted_accesses = int(access_rate * window_minutes)

        if predicted_accesses < 1:
            return None

        # Calculate confidence based on:
        # - Number of accesses
        # - Consistency of intervals
        # - Determinism score
        confidence = self._calculate_confidence(pattern)

        # Determine TTL based on access pattern
        suggested_ttl = self._calculate_ttl(pattern)

        # Build tags
        tags = [pattern.function_id]
        if pattern.is_deterministic:
            tags.append("deterministic")

        # Create prediction
        prediction = CachePrediction(
            key=pattern.key,
            function_id=pattern.function_id,
            predicted_accesses=predicted_accesses,
            confidence=confidence,
            suggested_ttl=suggested_ttl,
            suggested_tags=tags,
        )

        return prediction

    def _calculate_confidence(self, pattern: AccessPattern) -> float:
        """Calculate confidence score for a prediction.

        Args:
            pattern: The access pattern

        Returns:
            Confidence score between 0 and 1
        """
        # Base confidence from access count
        count_factor = min(pattern.access_count / 50, 1.0) * 0.3

        # Interval consistency
        interval_factor = 0.0
        if pattern.access_intervals:
            avg = sum(pattern.access_intervals) / len(pattern.access_intervals)
            if avg > 0:
                std_dev = (
                    sum((x - avg) ** 2 for x in pattern.access_intervals)
                    / len(pattern.access_intervals)
                ) ** 0.5
                cv = std_dev / avg  # Coefficient of variation
                interval_factor = max(0, 1 - cv) * 0.4

        # Determinism factor
        determinism_factor = pattern.determinism_score * 0.3

        return min(count_factor + interval_factor + determinism_factor, 1.0)

    def _calculate_ttl(self, pattern: AccessPattern) -> int:
        """Calculate suggested TTL based on access pattern.

        Args:
            pattern: The access pattern

        Returns:
            Suggested TTL in seconds
        """
        avg_interval = pattern.get_avg_interval()

        if avg_interval is None:
            return 3600  # Default 1 hour

        # TTL should be several times the average interval
        # But capped at 24 hours
        suggested_ttl = int(avg_interval * 3)
        return max(60, min(suggested_ttl, 86400))

    def get_determinism(self, function_id: str) -> DeterminismLevel:
        """Get the determinism level for a function.

        Args:
            function_id: The function ID

        Returns:
            DeterminismLevel
        """
        if function_id in self._determinism_cache:
            return self._determinism_cache[function_id]

        # Check patterns for this function
        patterns = [p for p in self._patterns.values() if p.function_id == function_id]

        if not patterns:
            return DeterminismLevel.UNKNOWN

        # Average determinism score
        avg_score = sum(p.determinism_score for p in patterns) / len(patterns)

        if avg_score >= 0.8:
            level = DeterminismLevel.DETERMINISTIC
        elif avg_score >= 0.5:
            level = DeterminismLevel.PARTIALLY_DETERMINISTIC
        else:
            level = DeterminismLevel.NON_DETERMINISTIC

        self._determinism_cache[function_id] = level
        return level

    def should_cache(
        self,
        function_id: str,
        output_hash: Optional[str] = None,
    ) -> bool:
        """Determine if a function output should be cached.

        Args:
            function_id: The function ID
            output_hash: Hash of the output

        Returns:
            True if should cache
        """
        determinism = self.get_determinism(function_id)

        if determinism == DeterminismLevel.NON_DETERMINISTIC:
            return False

        if determinism == DeterminismLevel.PARTIALLY_DETERMINISTIC:
            # Only cache if we have a hash to verify
            return output_hash is not None

        return True

    def get_stats(self) -> Dict[str, Any]:
        """Get predictor statistics.

        Returns:
            Dictionary with statistics
        """
        total_patterns = len(self._patterns)
        deterministic = sum(1 for p in self._patterns.values() if p.is_deterministic)

        return {
            "total_patterns": total_patterns,
            "deterministic_patterns": deterministic,
            "non_deterministic_patterns": total_patterns - deterministic,
            "predictions_count": len(self._predictions),
            "prediction_window_seconds": self._prediction_window,
            "min_confidence": self._min_confidence,
        }


# Global predictor instance
_predictor: Optional[CachePredictor] = None


def get_cache_predictor() -> CachePredictor:
    """Get the global cache predictor instance.

    Returns:
        CachePredictor instance
    """
    global _predictor
    if _predictor is None:
        _predictor = CachePredictor()

    return _predictor
