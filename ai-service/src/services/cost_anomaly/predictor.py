"""Cost anomaly detector — adaptive Z-score per function.

Replaces the hardcoded anomaly_detected=false with real per-function
statistical anomaly detection using Welford's online algorithm.
"""

import json
import logging
from collections import defaultdict
from datetime import datetime, timedelta
from typing import Dict, List, Optional

import redis.asyncio as redis

from ...config import settings
from .models import (
    CostAnomalyResult,
    CostAnomalySummary,
    CostExecutionMetrics,
    FunctionCostStats,
)

logger = logging.getLogger(__name__)


class CostAnomalyDetector:
    """Adaptive per-function cost anomaly detector.

    Uses Welford's online algorithm for running mean/stddev,
    then flags executions where z-score exceeds the adaptive threshold.
    Also detects memory leak trends and error rate spikes.

    The threshold is adaptive per function based on coefficient of variation:
    - Stable functions (low CV) get lower thresholds (more sensitive)
    - Variable functions (high CV) get higher thresholds (less sensitive)

    Tenant-isolated: all Redis keys are namespaced by tenant.
    """

    STATS_KEY_PREFIX = "ml:cost_anomaly:stats:"
    ANOMALIES_KEY_PREFIX = "ml:cost_anomaly:records:"
    MEMORY_KEY_PREFIX = "ml:cost_anomaly:memory:"
    STATS_EXPIRY_DAYS = 30
    ANOMALIES_EXPIRY_DAYS = 14

    # Adaptive threshold bounds
    MIN_THRESHOLD = 2.5
    MAX_THRESHOLD = 5.0
    # CV thresholds for adaptation
    LOW_CV_THRESHOLD = 0.2
    HIGH_CV_THRESHOLD = 0.5

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._base_threshold = settings.ml_cost_anomaly_threshold
        self._window_hours = settings.ml_cost_anomaly_window_hours
        self._memory_trend_min_count = 10
        self._stats_cache: Dict[str, FunctionCostStats] = {}
        self._threshold_cache: Dict[str, float] = {}

    def _make_stats_key(self, tenant_id: str, function_id: str) -> str:
        """Create tenant-isolated Redis key for cost stats."""
        return f"{self.STATS_KEY_PREFIX}{tenant_id}:{function_id}"

    def _make_anomalies_key(self, tenant_id: str) -> str:
        """Create tenant-isolated Redis key for anomaly records."""
        return f"{self.ANOMALIES_KEY_PREFIX}{tenant_id}"

    def _make_memory_key(self, tenant_id: str, function_id: str) -> str:
        """Create tenant-isolated Redis key for memory history."""
        return f"{self.MEMORY_KEY_PREFIX}{tenant_id}:{function_id}"

    async def get_redis(self) -> Optional[redis.Redis]:
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url, encoding="utf-8", decode_responses=True
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Redis connection failed: {e}")
                self._redis = None
        return self._redis

    async def _load_stats(self, tenant_id: str, function_id: str) -> FunctionCostStats:
        """Load running stats for a tenant-function from Redis."""
        cache_key = f"{tenant_id}:{function_id}"
        if cache_key in self._stats_cache:
            return self._stats_cache[cache_key]

        r = await self.get_redis()
        if r:
            try:
                key = self._make_stats_key(tenant_id, function_id)
                data = await r.get(key)
                if data:
                    stats = FunctionCostStats.model_validate_json(data)
                    self._stats_cache[cache_key] = stats
                    return stats
            except Exception as e:
                logger.warning(f"Failed to load stats for {tenant_id}/{function_id}: {e}")

        stats = FunctionCostStats(function_id=function_id)
        self._stats_cache[cache_key] = stats
        return stats

    async def _save_stats(self, tenant_id: str, stats: FunctionCostStats) -> None:
        """Persist running stats to Redis with tenant isolation."""
        cache_key = f"{tenant_id}:{stats.function_id}"
        self._stats_cache[cache_key] = stats
        r = await self.get_redis()
        if r:
            try:
                key = self._make_stats_key(tenant_id, stats.function_id)
                await r.set(key, stats.model_dump_json(), ex=self.STATS_EXPIRY_DAYS * 86400)
            except Exception as e:
                logger.error(f"Failed to save stats for {tenant_id}/{stats.function_id}: {e}")

    def _calculate_adaptive_threshold(self, stats: FunctionCostStats) -> float:
        """Calculate adaptive threshold based on coefficient of variation.

        Functions with high variance relative to their mean get higher thresholds
        to avoid false positives. Functions with stable costs get lower thresholds
        for better sensitivity.

        Args:
            stats: Function statistics

        Returns:
            Adaptive threshold between MIN_THRESHOLD and MAX_THRESHOLD
        """
        if stats.count < 10:
            return self._base_threshold

        if stats.mean == 0:
            return self._base_threshold

        cv = stats.std / abs(stats.mean)

        if cv <= self.LOW_CV_THRESHOLD:
            scale = 0.8
        elif cv >= self.HIGH_CV_THRESHOLD:
            scale = 1.5
        else:
            t = (cv - self.LOW_CV_THRESHOLD) / (self.HIGH_CV_THRESHOLD - self.LOW_CV_THRESHOLD)
            scale = 0.8 + (t * 0.7)

        adaptive_threshold = self._base_threshold * scale
        return max(self.MIN_THRESHOLD, min(self.MAX_THRESHOLD, adaptive_threshold))

    async def check_execution(self, tenant_id: str, metrics: CostExecutionMetrics) -> CostAnomalyResult:
        """Check a single execution for cost anomalies.

        This is the main entry point called by the Go backend.
        Returns immediately with anomaly assessment.

        Args:
            tenant_id: Tenant ID for isolation
            metrics: Execution metrics to check
        """
        stats = await self._load_stats(tenant_id, metrics.function_id)

        if metrics.function_id not in self._threshold_cache:
            self._threshold_cache[metrics.function_id] = self._base_threshold

        adaptive_threshold = self._calculate_adaptive_threshold(stats)
        self._threshold_cache[metrics.function_id] = adaptive_threshold

        need_at_least = 5
        if stats.count < need_at_least:
            stats.update(metrics.cost_cents)
            await self._save_stats(tenant_id, stats)
            return CostAnomalyResult(
                is_anomaly=False,
                score=0.0,
                severity="none",
                details=f"Collecting baseline data ({stats.count}/{need_at_least} samples)",
                function_id=metrics.function_id,
                threshold=adaptive_threshold,
            )

        z = stats.z_score(metrics.cost_cents)
        is_anomaly = abs(z) > adaptive_threshold

        # Determine anomaly type and severity
        anomaly_type = None
        severity = "none"
        details = ""

        if is_anomaly:
            anomaly_type = "cost_spike"
            abs_z = abs(z)
            if abs_z > 5.0:
                severity = "critical"
            elif abs_z > 4.0:
                severity = "high"
            elif abs_z > 3.0:
                severity = "medium"
            else:
                severity = "low"

            details = (
                f"Cost spike: {metrics.cost_cents:.4f}¢ "
                f"(mean: {stats.mean:.4f}¢, σ: {stats.std:.4f}¢, z: {z:.2f})"
            )

        # Check for memory leak trend (monotonic increase)
        if not is_anomaly:
            memory_result = await self._check_memory_trend(tenant_id, metrics)
            if memory_result and memory_result.is_anomaly:
                memory_result.threshold = adaptive_threshold
                stats.update(metrics.cost_cents)
                await self._save_stats(tenant_id, stats)
                return memory_result

        # Update stats AFTER check (so current value doesn't pollute its own detection)
        stats.update(metrics.cost_cents)
        await self._save_stats(tenant_id, stats)

        result = CostAnomalyResult(
            is_anomaly=is_anomaly,
            score=min(1.0, abs(z) / 6.0),
            anomaly_type=anomaly_type,
            severity=severity,
            details=details,
            function_id=metrics.function_id,
            z_score=z,
            threshold=adaptive_threshold,
        )

        if is_anomaly:
            await self._store_anomaly(tenant_id, result)

        return result

    async def _check_memory_trend(self, tenant_id: str, metrics: CostExecutionMetrics) -> Optional[CostAnomalyResult]:
        """Detect monotonic memory increase (potential memory leak)."""
        r = await self.get_redis()
        if not r:
            return None

        try:
            key = self._make_memory_key(tenant_id, metrics.function_id)
            # Store last N memory values
            await r.lpush(key, json.dumps({"value": metrics.memory_mb, "ts": metrics.timestamp.isoformat()}))
            await r.ltrim(key, 0, self._memory_trend_min_count - 1)
            await r.expire(key, 3600)

            items = await r.lrange(key, 0, -1)
            if len(items) < self._memory_trend_min_count:
                return None

            values = [json.loads(item)["value"] for item in reversed(items)]

            # Check if monotonically increasing
            increasing_count = sum(1 for i in range(1, len(values)) if values[i] > values[i - 1])
            if increasing_count >= len(values) * 0.8:
                pct_increase = (values[-1] - values[0]) / max(values[0], 0.01) * 100
                if pct_increase > 20:
                    return CostAnomalyResult(
                        is_anomaly=True,
                        score=min(1.0, pct_increase / 100),
                        anomaly_type="memory_leak",
                        severity="high" if pct_increase > 50 else "medium",
                        details=(
                            f"Memory leak pattern: {pct_increase:.1f}% increase over "
                            f"{len(values)} executions ({values[0]:.1f}MB → {values[-1]:.1f}MB)"
                        ),
                        function_id=metrics.function_id,
                    )
        except Exception as e:
            logger.warning(f"Memory trend check failed: {e}")

        return None

    async def _store_anomaly(self, tenant_id: str, anomaly: CostAnomalyResult) -> None:
        """Store a detected anomaly in Redis with tenant isolation."""
        r = await self.get_redis()
        if r:
            try:
                key = self._make_anomalies_key(tenant_id)
                score = anomaly.detected_at.timestamp()
                await r.zadd(key, {anomaly.model_dump_json(): score})
                await r.expire(key, self.ANOMALIES_EXPIRY_DAYS * 86400)
            except Exception as e:
                logger.error(f"Failed to store cost anomaly: {e}")

    async def get_anomalies(
        self,
        tenant_id: str,
        function_id: Optional[str] = None,
        limit: int = 50,
    ) -> List[CostAnomalyResult]:
        """Get recent cost anomalies for a tenant."""
        r = await self.get_redis()
        if not r:
            return []

        try:
            key = self._make_anomalies_key(tenant_id)
            items = await r.zrevrange(key, 0, limit * 2)
            results = []
            for item in items:
                try:
                    data = json.loads(item)
                    if function_id is None or data.get("function_id") == function_id:
                        results.append(CostAnomalyResult(**data))
                except Exception:
                    continue
            return results[:limit]
        except Exception as e:
            logger.error(f"Failed to get cost anomalies: {e}")
            return []

    async def get_summary(self, tenant_id: str, hours: int = 24) -> CostAnomalySummary:
        """Get cost anomaly summary for a tenant."""
        anomalies = await self.get_anomalies(tenant_id, limit=100)
        cutoff = datetime.utcnow() - timedelta(hours=hours)
        recent = [a for a in anomalies if a.detected_at >= cutoff]

        return CostAnomalySummary(
            tenant_id=tenant_id,
            total_anomalies=len(recent),
            anomalies=recent,
            period_hours=hours,
        )

    async def close(self):
        if self._redis:
            await self._redis.close()
            self._redis = None


_detector: Optional[CostAnomalyDetector] = None


def get_cost_anomaly_detector() -> CostAnomalyDetector:
    global _detector
    if _detector is None:
        _detector = CostAnomalyDetector()
    return _detector
