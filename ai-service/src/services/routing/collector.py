"""Latency data collector for routing service.

Collects latency samples from edge executions and stores them in Redis.
"""

import json
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional

import redis.asyncio as redis

from ...config import settings
from ...models.schemas import EdgeProvider, LatencySample
from .models import EdgeMetrics, exponential_decay

logger = logging.getLogger(__name__)


class LatencyCollector:
    """Collects and manages latency data from edge executions.

    Uses Redis for storage with time-based decay for older samples.
    """

    # Redis key patterns
    LATENCY_KEY_PREFIX = "routing:latency:"
    EDGE_STATUS_KEY = "routing:edge:status"
    SAMPLE_EXPIRY_SECONDS = 3600  # 1 hour

    def __init__(self):
        self._redis: Optional[redis.Redis] = None

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

    async def record_latency(
        self,
        function_id: str,
        edge: EdgeProvider,
        latency_ms: float,
        success: bool = True,
    ) -> None:
        """Record a latency sample for an edge.

        Args:
            function_id: The function ID
            edge: The edge provider
            latency_ms: Latency in milliseconds
            success: Whether the execution was successful
        """
        sample = LatencySample(
            function_id=function_id,
            edge=edge,
            latency_ms=latency_ms,
            timestamp=datetime.utcnow(),
            success=success,
        )

        redis_client = await self.get_redis()
        if not redis_client:
            logger.warning("Redis not available, cannot record latency")
            return

        try:
            # Store sample with timestamp
            key = f"{self.LATENCY_KEY_PREFIX}{edge.value}"
            sample_json = json.dumps(sample.model_dump(), default=str)

            # Use sorted set with timestamp as score for time-based queries
            score = sample.timestamp.timestamp()
            await redis_client.zadd(key, {sample_json: score})

            # Set expiry on the key
            await redis_client.expire(key, self.SAMPLE_EXPIRY_SECONDS)

            logger.debug(f"Recorded latency: {edge.value} = {latency_ms}ms for {function_id}")
        except Exception as e:
            logger.error(f"Failed to record latency: {e}")

    async def get_latency_samples(
        self,
        edge: EdgeProvider,
        function_id: Optional[str] = None,
        window_seconds: int = 300,
    ) -> List[LatencySample]:
        """Get latency samples for an edge within a time window.

        Args:
            edge: The edge provider
            function_id: Optional function ID to filter by
            window_seconds: Time window in seconds (default 5 minutes)

        Returns:
            List of latency samples
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return []

        try:
            key = f"{self.LATENCY_KEY_PREFIX}{edge.value}"
            min_score = (datetime.utcnow() - timedelta(seconds=window_seconds)).timestamp()

            samples_json = await redis_client.zrangebyscore(
                key, min_score, "+inf"
            )

            samples = []
            for sample_json in samples_json:
                try:
                    data = json.loads(sample_json)
                    sample = LatencySample(**data)

                    # Filter by function_id if specified
                    if function_id is None or sample.function_id == function_id:
                        samples.append(sample)
                except Exception as e:
                    logger.warning(f"Failed to parse latency sample: {e}")

            return samples
        except Exception as e:
            logger.error(f"Failed to get latency samples: {e}")
            return []

    async def calculate_weighted_latency(
        self,
        edge: EdgeProvider,
        function_id: Optional[str] = None,
        window_seconds: int = 300,
    ) -> float:
        """Calculate weighted average latency using exponential decay.

        Args:
            edge: The edge provider
            function_id: Optional function ID to filter by
            window_seconds: Time window in seconds

        Returns:
            Weighted average latency in milliseconds
        """
        samples = await self.get_latency_samples(edge, function_id, window_seconds)

        if not samples:
            return 0.0

        now = datetime.utcnow()
        total_weight = 0.0
        weighted_sum = 0.0

        for sample in samples:
            age_seconds = (now - sample.timestamp).total_seconds()
            weight = exponential_decay(age_seconds)

            weighted_sum += sample.latency_ms * weight
            total_weight += weight

        if total_weight == 0:
            return 0.0

        return weighted_sum / total_weight

    async def get_edge_metrics(self, edge: EdgeProvider) -> EdgeMetrics:
        """Get current metrics for an edge.

        Args:
            edge: The edge provider

        Returns:
            EdgeMetrics with current values
        """
        redis_client = await self.get_redis()

        # Get weighted latency
        latency = await self.calculate_weighted_latency(edge)

        # Get current load from Redis (updated by edge health checks)
        load = 0.0
        if redis_client:
            try:
                load_key = f"{self.EDGE_STATUS_KEY}:{edge.value}:load"
                load_str = await redis_client.get(load_key)
                if load_str:
                    load = float(load_str)
            except Exception as e:
                logger.warning(f"Failed to get load for {edge.value}: {e}")

        # Get sample count
        samples = await self.get_latency_samples(edge)

        return EdgeMetrics(
            provider=edge,
            avg_latency_ms=latency,
            current_load_percent=load,
            available=True,  # Default to available
            sample_count=len(samples),
        )

    async def get_all_edge_metrics(self) -> Dict[EdgeProvider, EdgeMetrics]:
        """Get metrics for all edges.

        Returns:
            Dictionary of edge to metrics
        """
        metrics = {}

        for edge in EdgeProvider:
            try:
                metrics[edge] = await self.get_edge_metrics(edge)
            except Exception as e:
                logger.warning(f"Failed to get metrics for {edge.value}: {e}")
                metrics[edge] = EdgeMetrics(provider=edge, available=False)

        return metrics

    async def update_edge_load(self, edge: EdgeProvider, load_percent: float) -> None:
        """Update current load for an edge.

        Args:
            edge: The edge provider
            load_percent: Current load percentage (0-100)
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return

        try:
            load_key = f"{self.EDGE_STATUS_KEY}:{edge.value}:load"
            await redis_client.setex(load_key, 60, str(load_percent))  # Expire after 1 minute
        except Exception as e:
            logger.error(f"Failed to update edge load: {e}")

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global latency collector instance
_latency_collector: Optional[LatencyCollector] = None


def get_latency_collector() -> LatencyCollector:
    """Get the global latency collector instance.

    Returns:
        The LatencyCollector instance
    """
    global _latency_collector
    if _latency_collector is None:
        _latency_collector = LatencyCollector()
    return _latency_collector
