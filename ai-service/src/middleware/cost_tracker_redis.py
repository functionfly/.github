"""Redis-backed cost tracking for FlyMind AI Service.

This module provides persistent cost tracking using Redis as a backend,
ensuring cost data survives service restarts and is consistent across instances.
"""

import json
import logging
import time
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Dict, List, Optional
import threading

from ..services.redis_client import RedisClient
from .cost_tracker import CostLimitExceeded

logger = logging.getLogger(__name__)


@dataclass
class CostLimit:
    """Cost limit configuration."""
    limit: float
    period: str = "day"
    alert_threshold: float = 0.8


@dataclass
class CostEntry:
    """A single cost entry."""
    tenant_id: str
    provider: str
    model: str
    input_tokens: int
    output_tokens: int
    cost: float
    timestamp: datetime = field(default_factory=datetime.utcnow)


class RedisCostTracker:
    """Cost tracker with Redis backend for persistent, distributed cost tracking.

    Uses Redis sorted sets for time-windowed cost aggregation and
    Hash maps for efficient cost lookups.
    """

    COST_PREFIX = "cost:v1"
    COST_ENTRIES_KEY = "cost:entries"

    def __init__(self, redis_client: Optional[RedisClient] = None):
        """Initialize the Redis cost tracker.

        Args:
            redis_client: Optional Redis client. If not provided, will attempt to create one.
        """
        self._logger = logging.getLogger(__name__)
        self._redis: Optional[RedisClient] = redis_client
        self._lock = threading.Lock()
        self._local_fallback: Dict[str, List[CostEntry]] = []
        self._total_cost = 0.0

    async def _get_redis(self) -> Optional[RedisClient]:
        """Get or create Redis client."""
        if self._redis is None:
            self._redis = await RedisClient.create()
            if self._redis and not await self._redis.ping():
                self._redis = None
        return self._redis

    def _get_cost_key(self, tenant_id: str, period: str) -> str:
        """Generate Redis key for cost tracking."""
        return f"{self.COST_PREFIX}:{tenant_id}:{period}"

    def _get_period_cutoff(self, period: str) -> datetime:
        """Get the cutoff time for a period."""
        now = datetime.utcnow()
        if period == "hour":
            return now - timedelta(hours=1)
        elif period == "day":
            return now - timedelta(days=1)
        elif period == "month":
            return now - timedelta(days=30)
        else:
            return now - timedelta(days=1)

    async def record_cost(
        self,
        tenant_id: str,
        provider: str,
        model: str,
        input_tokens: int,
        output_tokens: int,
        cost: float,
    ) -> None:
        """Record an API cost.

        Args:
            tenant_id: Tenant ID
            provider: Provider name
            model: Model name
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens
            cost: Cost in USD
        """
        entry = CostEntry(
            tenant_id=tenant_id,
            provider=provider,
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            cost=cost,
        )

        redis = await self._get_redis()

        if redis:
            try:
                period_keys = ["hour", "day", "month"]
                now = time.time()

                for period in period_keys:
                    cost_key = self._get_cost_key(tenant_id, period)
                    await redis._client.zincrby(cost_key, cost, str(now))

                    if period == "hour":
                        await redis._client.expire(cost_key, 3600 * 2)
                    elif period == "day":
                        await redis._client.expire(cost_key, 86400 * 2)
                    else:
                        await redis._client.expire(cost_key, 86400 * 35)

                entry_key = f"{self.COST_ENTRIES_KEY}:{tenant_id}"
                entry_data = json.dumps({
                    "provider": provider,
                    "model": model,
                    "input_tokens": input_tokens,
                    "output_tokens": output_tokens,
                    "cost": cost,
                    "timestamp": datetime.utcnow().isoformat(),
                })
                await redis._client.rpush(entry_key, entry_data)
                await redis._client.expire(entry_key, 86400 * 7)

                return
            except Exception as e:
                self._logger.warning(f"Redis cost record failed, falling back to local: {e}")

        with self._lock:
            self._local_fallback.append(entry)
            self._total_cost += cost
            if len(self._local_fallback) > 10000:
                self._local_fallback = self._local_fallback[-5000:]

    async def check_limit(
        self,
        tenant_id: str,
        additional_cost: float = 0.0,
    ) -> bool:
        """Check if adding cost would exceed limit.

        Args:
            tenant_id: Tenant ID
            additional_cost: Additional cost to check

        Returns:
            True if within limits

        Raises:
            CostLimitExceeded: If limit exceeded
        """
        current_cost = await self.get_current_cost(tenant_id, "day")

        if current_cost + additional_cost > 100.0:
            raise CostLimitExceeded(
                f"Cost limit exceeded: ${current_cost:.2f} / $100.00 per day",
                tenant_id=tenant_id,
                current_cost=current_cost,
                limit=100.0,
                period="day",
            )

        return True

    async def get_current_cost(
        self,
        tenant_id: str,
        period: str = "day",
    ) -> float:
        """Get current cost for a period.

        Args:
            tenant_id: Tenant ID
            period: Period (hour, day, month)

        Returns:
            Current cost
        """
        redis = await self._get_redis()

        if redis:
            try:
                cost_key = self._get_cost_key(tenant_id, period)
                cutoff = self._get_period_cutoff(period)
                cutoff_ts = cutoff.timestamp()

                entries = await redis._client.zrangebyscore(
                    cost_key, cutoff_ts, "+inf"
                )

                total = sum(float(e) for e in entries)
                return total
            except Exception as e:
                self._logger.warning(f"Redis get_current_cost failed, falling back to local: {e}")

        return await self._get_current_cost_local(tenant_id, period)

    async def _get_current_cost_local(
        self,
        tenant_id: str,
        period: str = "day",
    ) -> float:
        """Local fallback for get_current_cost."""
        with self._lock:
            cutoff = self._get_period_cutoff(period)
            total = 0.0
            for entry in self._local_fallback:
                if entry.tenant_id == tenant_id and entry.timestamp >= cutoff:
                    total += entry.cost
            return total

    async def get_usage(self, tenant_id: str) -> Dict[str, float]:
        """Get current usage for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dictionary with usage statistics
        """
        hourly = await self.get_current_cost(tenant_id, "hour")
        daily = await self.get_current_cost(tenant_id, "day")
        monthly = await self.get_current_cost(tenant_id, "month")

        return {
            "hourly": round(hourly, 4),
            "daily": round(daily, 4),
            "monthly": round(monthly, 4),
            "limit": 100.0,
            "period": "day",
            "alert_threshold": 0.8,
            "alert_at": 80.0,
        }

    async def get_costs_by_provider(
        self,
        tenant_id: str,
    ) -> Dict[str, float]:
        """Get costs breakdown by provider.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dictionary of provider -> cost
        """
        redis = await self._get_redis()

        if redis:
            try:
                entry_key = f"{self.COST_ENTRIES_KEY}:{tenant_id}"
                entries = await redis._client.lrange(entry_key, 0, -1)

                costs: Dict[str, float] = {}
                for entry_data in entries:
                    try:
                        data = json.loads(entry_data)
                        key = f"{data['provider']}:{data['model']}"
                        costs[key] = costs.get(key, 0.0) + data['cost']
                    except (json.JSONDecodeError, KeyError):
                        continue

                return costs
            except Exception as e:
                self._logger.warning(f"Redis get_costs_by_provider failed, falling back to local: {e}")

        with self._lock:
            costs: Dict[str, float] = {}
            for entry in self._local_fallback:
                if entry.tenant_id == tenant_id:
                    key = f"{entry.provider}:{entry.model}"
                    costs[key] = costs.get(key, 0.0) + entry.cost
            return costs

    async def get_stats(self) -> Dict[str, any]:
        """Get cost tracker statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "total_cost": round(self._total_cost, 4),
                "local_entries": len(self._local_fallback),
                "redis_available": self._redis is not None,
            }

    async def reset(self, tenant_id: str) -> None:
        """Reset cost tracking for a tenant.

        Args:
            tenant_id: Tenant ID
        """
        redis = await self._get_redis()

        if redis:
            try:
                for period in ["hour", "day", "month"]:
                    cost_key = self._get_cost_key(tenant_id, period)
                    await redis.delete(cost_key)

                entry_key = f"{self.COST_ENTRIES_KEY}:{tenant_id}"
                await redis.delete(entry_key)
            except Exception as e:
                self._logger.warning(f"Redis reset failed: {e}")

        with self._lock:
            self._local_fallback = [
                e for e in self._local_fallback
                if e.tenant_id != tenant_id
            ]

    async def get_cost_history(
        self,
        tenant_id: str,
        limit: int = 100,
    ) -> List[CostEntry]:
        """Get cost history for a tenant.

        Args:
            tenant_id: Tenant ID
            limit: Maximum entries to return

        Returns:
            List of CostEntry
        """
        redis = await self._get_redis()

        if redis:
            try:
                entry_key = f"{self.COST_ENTRIES_KEY}:{tenant_id}"
                entries = await redis._client.lrange(entry_key, -limit, -1)

                result = []
                for entry_data in entries:
                    try:
                        data = json.loads(entry_data)
                        result.append(CostEntry(
                            tenant_id=tenant_id,
                            provider=data['provider'],
                            model=data['model'],
                            input_tokens=data['input_tokens'],
                            output_tokens=data['output_tokens'],
                            cost=data['cost'],
                            timestamp=datetime.fromisoformat(data['timestamp']),
                        ))
                    except (json.JSONDecodeError, KeyError):
                        continue

                return result
            except Exception as e:
                self._logger.warning(f"Redis get_cost_history failed, falling back to local: {e}")

        with self._lock:
            entries = [
                e for e in reversed(self._local_fallback)
                if e.tenant_id == tenant_id
            ]
            return entries[:limit]


_redis_cost_tracker: Optional[RedisCostTracker] = None


async def get_redis_cost_tracker() -> RedisCostTracker:
    """Get the global Redis-backed cost tracker.

    Returns:
        RedisCostTracker instance
    """
    global _redis_cost_tracker
    if _redis_cost_tracker is None:
        _redis_cost_tracker = RedisCostTracker()
    return _redis_cost_tracker
