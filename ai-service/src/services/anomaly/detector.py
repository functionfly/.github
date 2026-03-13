"""Statistical anomaly detector.

Uses Z-score based detection for latency, error rate, and cold start rate.
"""

import json
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Set

import redis.asyncio as redis

from ...config import settings
from ...models.schemas import ExecutionMetrics
from .models import (
    AnomalyRecord,
    MetricData,
    AnomalyThresholds,
    StatisticalSummary,
    calculate_mean,
    calculate_std,
    calculate_z_score,
    determine_severity,
)

logger = logging.getLogger(__name__)


class AnomalyDetector:
    """Statistical anomaly detector for function executions.

    Uses Z-score based detection with sliding windows:
    - latency: alerts when latency > 3σ
    - error_rate: alerts when error_rate > 1%
    - cold_start_rate: alerts when cold_start_rate > 10%
    """

    # Redis key patterns
    METRICS_KEY_PREFIX = "anomaly:metrics:"
    ANOMALIES_KEY = "anomaly:records"
    THRESHOLDS_KEY = "anomaly:thresholds"
    METRICS_EXPIRY_HOURS = 24
    ANOMALIES_EXPIRY_DAYS = 7

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._thresholds = AnomalyThresholds(
            latency_z_score=settings.anomaly_latency_threshold,
            error_rate=settings.anomaly_error_rate_threshold,
            cold_start_rate=settings.anomaly_cold_start_threshold,
            window_minutes=settings.anomaly_window_minutes,
            check_interval=settings.anomaly_check_interval_seconds,
        )

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

    async def record_metric(self, metric: MetricData) -> None:
        """Record a metric data point.

        Args:
            metric: The metric data to record
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return

        try:
            key = f"{self.METRICS_KEY_PREFIX}{metric.function_id}:{metric.metric_name}"
            metric_json = json.dumps(metric.model_dump(), default=str)

            # Use sorted set with timestamp as score
            score = metric.timestamp.timestamp()
            await redis_client.zadd(key, {metric_json: score})

            # Set expiry
            await redis_client.expire(key, self.METRICS_EXPIRY_HOURS * 3600)

        except Exception as e:
            logger.error(f"Failed to record metric: {e}")

    async def get_function_ids_with_metrics(self) -> List[str]:
        """Get function IDs that have metrics stored (for periodic anomaly checks).

        Returns:
            List of unique function IDs
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return []

        try:
            prefix = self.METRICS_KEY_PREFIX
            seen: Set[str] = set()
            async for key in redis_client.scan_iter(match=f"{prefix}*", count=100):
                # Key format: anomaly:metrics:{function_id}:{metric_name}
                suffix = key[len(prefix) :]
                if ":" in suffix:
                    fid = suffix.split(":", 1)[0]
                    seen.add(fid)
            return list(seen)
        except Exception as e:
            logger.error(f"Failed to get function IDs with metrics: {e}")
            return []

    async def record_execution_metrics(self, metrics: ExecutionMetrics) -> None:
        """Record execution metrics for a function.

        Args:
            metrics: Execution metrics
        """
        # Record latency
        await self.record_metric(MetricData(
            function_id=metrics.function_id,
            metric_name="latency_ms",
            value=metrics.latency_ms,
            timestamp=metrics.timestamp,
        ))

        # Record cold start (as binary)
        await self.record_metric(MetricData(
            function_id=metrics.function_id,
            metric_name="cold_start",
            value=1.0 if metrics.cold_start else 0.0,
            timestamp=metrics.timestamp,
        ))

        # Record error (as binary)
        error_value = 1.0 if metrics.error is not None else 0.0
        await self.record_metric(MetricData(
            function_id=metrics.function_id,
            metric_name="error",
            value=error_value,
            timestamp=metrics.timestamp,
        ))

    async def get_metrics_window(
        self,
        function_id: str,
        metric_name: str,
        window_minutes: Optional[int] = None,
    ) -> List[MetricData]:
        """Get metrics within a time window.

        Args:
            function_id: The function ID
            metric_name: The metric name
            window_minutes: Time window in minutes

        Returns:
            List of metric data points
        """
        window_minutes = window_minutes or self._thresholds.window_minutes

        redis_client = await self.get_redis()
        if not redis_client:
            return []

        try:
            key = f"{self.METRICS_KEY_PREFIX}{function_id}:{metric_name}"
            min_score = (datetime.utcnow() - timedelta(minutes=window_minutes)).timestamp()

            metrics_json = await redis_client.zrangebyscore(key, min_score, "+inf")

            metrics = []
            for item in metrics_json:
                try:
                    data = json.loads(item)
                    metrics.append(MetricData(**data))
                except Exception as e:
                    logger.warning(f"Failed to parse metric: {e}")

            return sorted(metrics, key=lambda m: m.timestamp)
        except Exception as e:
            logger.error(f"Failed to get metrics window: {e}")
            return []

    async def calculate_statistics(
        self,
        function_id: str,
        metric_name: str,
    ) -> Optional[StatisticalSummary]:
        """Calculate statistical summary for a metric.

        Args:
            function_id: The function ID
            metric_name: The metric name

        Returns:
            StatisticalSummary or None if not enough data
        """
        metrics = await self.get_metrics_window(function_id, metric_name)

        if len(metrics) < 3:
            return None

        values = [m.value for m in metrics]
        mean = calculate_mean(values)
        std = calculate_std(values, mean)

        return StatisticalSummary(
            count=len(values),
            mean=mean,
            std=std,
            min=min(values),
            max=max(values),
        )

    async def check_latency_anomaly(
        self,
        function_id: str,
    ) -> Optional[AnomalyRecord]:
        """Check for latency anomalies.

        Args:
            function_id: The function ID

        Returns:
            AnomalyRecord if anomaly detected, None otherwise
        """
        stats = await self.calculate_statistics(function_id, "latency_ms")

        if not stats or stats.std == 0:
            return None

        # Get the most recent latency
        metrics = await self.get_metrics_window(function_id, "latency_ms")
        if not metrics:
            return None

        latest = metrics[-1]

        # Calculate Z-score
        z_score = calculate_z_score(latest.value, stats.mean, stats.std)

        # Check if anomalous
        if abs(z_score) > self._thresholds.latency_z_score:
            severity = determine_severity(z_score)

            return AnomalyRecord(
                function_id=function_id,
                type="latency_spike",
                severity=severity,
                description=f"Latency spike detected: {latest.value:.2f}ms "
                           f"(mean: {stats.mean:.2f}ms, σ: {stats.std:.2f})",
                metric_name="latency_ms",
                metric_value=latest.value,
                threshold=self._thresholds.latency_z_score * stats.std + stats.mean,
                z_score=z_score,
            )

        return None

    async def check_error_rate_anomaly(
        self,
        function_id: str,
    ) -> Optional[AnomalyRecord]:
        """Check for error rate anomalies.

        Args:
            function_id: The function ID

        Returns:
            AnomalyRecord if anomaly detected, None otherwise
        """
        stats = await self.calculate_statistics(function_id, "error")

        if not stats or stats.count < 5:
            return None

        error_rate = stats.mean

        # Check if error rate exceeds threshold
        if error_rate > self._thresholds.error_rate:
            severity = "high" if error_rate > 0.05 else "medium"

            return AnomalyRecord(
                function_id=function_id,
                type="error_rate_increase",
                severity=severity,
                description=f"Error rate increased: {error_rate*100:.2f}% "
                           f"(threshold: {self._thresholds.error_rate*100}%)",
                metric_name="error_rate",
                metric_value=error_rate,
                threshold=self._thresholds.error_rate,
            )

        return None

    async def check_cold_start_anomaly(
        self,
        function_id: str,
    ) -> Optional[AnomalyRecord]:
        """Check for cold start rate anomalies.

        Args:
            function_id: The function ID

        Returns:
            AnomalyRecord if anomaly detected, None otherwise
        """
        stats = await self.calculate_statistics(function_id, "cold_start")

        if not stats or stats.count < 5:
            return None

        cold_start_rate = stats.mean

        # Check if cold start rate exceeds threshold
        if cold_start_rate > self._thresholds.cold_start_rate:
            severity = "high" if cold_start_rate > 0.2 else "medium"

            return AnomalyRecord(
                function_id=function_id,
                type="cold_start_spike",
                severity=severity,
                description=f"Cold start rate elevated: {cold_start_rate*100:.2f}% "
                           f"(threshold: {self._thresholds.cold_start_rate*100}%)",
                metric_name="cold_start_rate",
                metric_value=cold_start_rate,
                threshold=self._thresholds.cold_start_rate,
            )

        return None

    async def check_all_anomalies(
        self,
        function_id: str,
    ) -> List[AnomalyRecord]:
        """Check for all types of anomalies.

        Args:
            function_id: The function ID

        Returns:
            List of detected anomalies
        """
        anomalies = []

        # Check latency
        latency_anomaly = await self.check_latency_anomaly(function_id)
        if latency_anomaly:
            anomalies.append(latency_anomaly)

        # Check error rate
        error_anomaly = await self.check_error_rate_anomaly(function_id)
        if error_anomaly:
            anomalies.append(error_anomaly)

        # Check cold start rate
        cold_start_anomaly = await self.check_cold_start_anomaly(function_id)
        if cold_start_anomaly:
            anomalies.append(cold_start_anomaly)

        return anomalies

    async def store_anomaly(self, anomaly: AnomalyRecord) -> None:
        """Store a detected anomaly.

        Args:
            anomaly: The anomaly to store
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return

        try:
            anomaly_json = json.dumps(anomaly.model_dump(), default=str)

            # Store in a list with timestamp as score
            score = anomaly.detected_at.timestamp()
            await redis_client.zadd(self.ANOMALIES_KEY, {anomaly_json: score})

            # Set expiry
            await redis_client.expire(self.ANOMALIES_KEY, self.ANOMALIES_EXPIRY_DAYS * 86400)

        except Exception as e:
            logger.error(f"Failed to store anomaly: {e}")

    async def get_anomalies(
        self,
        function_id: Optional[str] = None,
        limit: int = 20,
    ) -> List[AnomalyRecord]:
        """Get stored anomalies.

        Args:
            function_id: Optional function ID to filter by
            limit: Maximum number of anomalies to return

        Returns:
            List of anomaly records
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return []

        try:
            # Get most recent anomalies
            anomalies_json = await redis_client.zrevrange(
                self.ANOMALIES_KEY, 0, limit - 1
            )

            anomalies = []
            for item in anomalies_json:
                try:
                    data = json.loads(item)

                    # Filter by function_id if specified
                    if function_id is None or data.get("function_id") == function_id:
                        anomalies.append(AnomalyRecord(**data))
                except Exception as e:
                    logger.warning(f"Failed to parse anomaly: {e}")

            return anomalies
        except Exception as e:
            logger.error(f"Failed to get anomalies: {e}")
            return []

    async def acknowledge_anomaly(
        self,
        anomaly_id: str,
        acknowledged_by: str,
    ) -> bool:
        """Acknowledge an anomaly.

        Args:
            anomaly_id: The anomaly ID
            acknowledged_by: Who is acknowledging

        Returns:
            True if successful
        """
        # Get all anomalies
        anomalies = await self.get_anomalies(limit=100)

        # Find and update the anomaly
        for anomaly in anomalies:
            if anomaly.id == anomaly_id:
                anomaly.acknowledged = True
                anomaly.acknowledged_by = acknowledged_by
                anomaly.acknowledged_at = datetime.utcnow()

                # Re-store the updated anomaly
                await self.store_anomaly(anomaly)
                return True

        return False

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global anomaly detector instance
_anomaly_detector: Optional[AnomalyDetector] = None


def get_anomaly_detector() -> AnomalyDetector:
    """Get the global anomaly detector instance.

    Returns:
        The AnomalyDetector instance
    """
    global _anomaly_detector
    if _anomaly_detector is None:
        _anomaly_detector = AnomalyDetector()
    return _anomaly_detector
