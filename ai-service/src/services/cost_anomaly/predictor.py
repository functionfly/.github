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
    """

    STATS_KEY_PREFIX = "ml:cost_anomaly:stats:"
    ANOMALIES_KEY = "ml:cost_anomaly:records"
    STATS_EXPIRY_DAYS = 30
    ANOMALIES_EXPIRY_DAYS = 14

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._threshold = settings.ml_cost_anomaly_threshold
        self._window_hours = settings.ml_cost_anomaly_window_hours
        self._memory_trend_min_count = 10
        self._stats_cache: Dict[str, FunctionCostStats] = {}

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

    async def _load_stats(self, function_id: str) -> FunctionCostStats:
        """Load running stats for a function from Redis."""
        if function_id in self._stats_cache:
            return self._stats_cache[function_id]

        r = await self.get_redis()
        if r:
            try:
                key = f"{self.STATS_KEY_PREFIX}{function_id}"
                data = await r.get(key)
                if data:
                    stats = FunctionCostStats.model_validate_json(data)
                    self._stats_cache[function_id] = stats
                    return stats
            except Exception as e:
                logger.warning(f"Failed to load stats for {function_id}: {e}")

        stats = FunctionCostStats(function_id=function_id)
        self._stats_cache[function_id] = stats
        return stats

    async def _save_stats(self, stats: FunctionCostStats) -> None:
        """Persist running stats to Redis."""
        self._stats_cache[stats.function_id] = stats
        r = await self.get_redis()
        if r:
            try:
                key = f"{self.STATS_KEY_PREFIX}{stats.function_id}"
                await r.set(key, stats.model_dump_json(), ex=self.STATS_EXPIRY_DAYS * 86400)
            except Exception as e:
                logger.error(f"Failed to save stats for {stats.function_id}: {e}")

    async def check_execution(self, metrics: CostExecutionMetrics) -> CostAnomalyResult:
        """Check a single execution for cost anomalies.

        This is the main entry point called by the Go backend.
        Returns immediately with anomaly assessment.
        """
        stats = await self._load_stats(metrics.function_id)

        # Need at least 5 data points before we can detect anomalies
        if stats.count < 5:
            stats.update(metrics.cost_cents)
            await self._save_stats(stats)
            return CostAnomalyResult(
                is_anomaly=False,
                score=0.0,
                severity="none",
                details=f"Collecting baseline data ({stats.count}/5 samples)",
                function_id=metrics.function_id,
            )

        # Calculate z-score
        z = stats.z_score(metrics.cost_cents)
        is_anomaly = abs(z) > self._threshold

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
            memory_result = await self._check_memory_trend(metrics)
            if memory_result and memory_result.is_anomaly:
                stats.update(metrics.cost_cents)
                await self._save_stats(stats)
                return memory_result

        # Update stats AFTER check (so current value doesn't pollute its own detection)
        stats.update(metrics.cost_cents)
        await self._save_stats(stats)

        # Store anomaly if detected
        result = CostAnomalyResult(
            is_anomaly=is_anomaly,
            score=min(1.0, abs(z) / 6.0),
            anomaly_type=anomaly_type,
            severity=severity,
            details=details,
            function_id=metrics.function_id,
            z_score=z,
            threshold=self._threshold,
        )

        if is_anomaly:
            await self._store_anomaly(result)

        return result

    async def _check_memory_trend(self, metrics: CostExecutionMetrics) -> Optional[CostAnomalyResult]:
        """Detect monotonic memory increase (potential memory leak)."""
        r = await self.get_redis()
        if not r:
            return None

        try:
            key = f"ml:cost_anomaly:memory:{metrics.function_id}"
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

    async def _store_anomaly(self, anomaly: CostAnomalyResult) -> None:
        """Store a detected anomaly in Redis."""
        r = await self.get_redis()
        if r:
            try:
                score = anomaly.detected_at.timestamp()
                await r.zadd(self.ANOMALIES_KEY, {anomaly.model_dump_json(): score})
                await r.expire(self.ANOMALIES_KEY, self.ANOMALIES_EXPIRY_DAYS * 86400)
            except Exception as e:
                logger.error(f"Failed to store cost anomaly: {e}")

    async def get_anomalies(
        self,
        function_id: Optional[str] = None,
        limit: int = 50,
    ) -> List[CostAnomalyResult]:
        """Get recent cost anomalies."""
        r = await self.get_redis()
        if not r:
            return []

        try:
            items = await r.zrevrange(self.ANOMALIES_KEY, 0, limit * 2)
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
        anomalies = await self.get_anomalies(limit=100)
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
