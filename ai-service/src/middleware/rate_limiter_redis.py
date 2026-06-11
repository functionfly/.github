"""Redis-backed rate limiting for FlyMind AI Service.

This module provides distributed rate limiting using Redis as a backend,
supporting multi-instance deployments with consistent rate limit enforcement.
"""

import logging
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, Optional
import threading

from ..services.redis_client import RedisClient
from .rate_limiter import RateLimitExceeded

logger = logging.getLogger(__name__)


@dataclass
class RateLimitConfig:
    """Configuration for rate limiting."""
    requests_per_minute: int = 60
    requests_per_hour: int = 1000
    requests_per_day: int = 10000
    burst_size: int = 10
    enabled: bool = True
    embed_tokens_per_minute: int = 100000
    embed_cost_per_day: float = 50.0


class RedisRateLimiter:
    """Token bucket rate limiter with Redis backend for distributed deployments.

    Uses Redis atomic operations (INCR, EXPIRE) to ensure consistent rate limiting
    across multiple service instances.
    """

    RATE_PREFIX = "ratelimit:v1"

    def __init__(self, redis_client: Optional[RedisClient] = None):
        """Initialize the Redis rate limiter.

        Args:
            redis_client: Optional Redis client. If not provided, will attempt to create one.
        """
        self._logger = logging.getLogger(__name__)
        self._redis: Optional[RedisClient] = redis_client
        self._lock = threading.Lock()
        self._local_fallback: Dict[str, Dict] = {}
        self._use_local_fallback = False

    async def _get_redis(self) -> Optional[RedisClient]:
        """Get or create Redis client."""
        if self._redis is None:
            self._redis = await RedisClient.create()
            if self._redis and not await self._redis.ping():
                self._redis = None
        return self._redis

    def _get_key(self, tenant_id: str, window: str) -> str:
        """Generate Redis key for rate limit window."""
        return f"{self.RATE_PREFIX}:{tenant_id}:{window}"

    async def check_limit(
        self,
        tenant_id: str,
        cost: int = 1,
    ) -> bool:
        """Check if request is within rate limits using Redis atomic operations.

        Args:
            tenant_id: Tenant ID
            cost: Request cost (for burst calculations)

        Returns:
            True if within limits

        Raises:
            RateLimitExceeded: If limit exceeded
        """
        redis = await self._get_redis()

        if not redis:
            return await self._check_limit_local(tenant_id, cost)

        now = time.time()
        minute_key = self._get_key(tenant_id, "minute")
        hour_key = self._get_key(tenant_id, "hour")
        day_key = self._get_key(tenant_id, "day")
        burst_key = self._get_key(tenant_id, "burst")

        try:
            pipe = redis._client.pipeline() if hasattr(redis._client, 'pipeline') else None

            if pipe:
                pipe.incr(minute_key)
                pipe.expire(minute_key, 60)
                pipe.incr(hour_key)
                pipe.expire(hour_key, 3600)
                pipe.incr(day_key)
                pipe.expire(day_key, 86400)
                results = await self._execute_pipeline(pipe)

                minute_count = results[0]
                hour_count = results[2]
                day_count = results[4]
            else:
                minute_count = await redis._client.incr(minute_key)
                await redis._client.expire(minute_key, 60)
                hour_count = await redis._client.incr(hour_key)
                await redis._client.expire(hour_key, 3600)
                day_count = await redis._client.incr(day_key)
                await redis._client.expire(day_key, 86400)

            if minute_count > 60:
                retry_after = 60 - (now % 60)
                raise RateLimitExceeded(
                    f"Rate limit exceeded: 60 requests per minute",
                    tenant_id=tenant_id,
                    limit=60,
                    window_seconds=60,
                    retry_after=int(retry_after) + 1,
                )

            if hour_count > 1000:
                retry_after = 3600 - (now % 3600)
                raise RateLimitExceeded(
                    f"Rate limit exceeded: 1000 requests per hour",
                    tenant_id=tenant_id,
                    limit=1000,
                    window_seconds=3600,
                    retry_after=int(retry_after) + 1,
                )

            if day_count > 10000:
                retry_after = 86400 - (now % 86400)
                raise RateLimitExceeded(
                    f"Rate limit exceeded: 10000 requests per day",
                    tenant_id=tenant_id,
                    limit=10000,
                    window_seconds=86400,
                    retry_after=int(retry_after) + 1,
                )

            return True

        except RateLimitExceeded:
            raise
        except Exception as e:
            self._logger.warning(f"Redis rate limit check failed, falling back to local: {e}")
            return await self._check_limit_local(tenant_id, cost)

    async def _execute_pipeline(self, pipe):
        """Execute Redis pipeline and return results."""
        results = await pipe.execute()
        return results

    async def _check_limit_local(self, tenant_id: str, cost: int = 1) -> bool:
        """Local fallback rate limiting (in-process).

        This is used when Redis is unavailable. Not suitable for multi-instance deployments.
        """
        with self._lock:
            if tenant_id not in self._local_fallback:
                self._local_fallback[tenant_id] = {
                    "minute_count": 0,
                    "hour_count": 0,
                    "day_count": 0,
                    "minute_reset": time.time() + 60,
                    "hour_reset": time.time() + 3600,
                    "day_reset": time.time() + 86400,
                }

            state = self._local_fallback[tenant_id]
            now = time.time()

            if now >= state["minute_reset"]:
                state["minute_count"] = 0
                state["minute_reset"] = now + 60

            if now >= state["hour_reset"]:
                state["hour_count"] = 0
                state["hour_reset"] = now + 3600

            if now >= state["day_reset"]:
                state["day_count"] = 0
                state["day_reset"] = now + 86400

            if state["minute_count"] >= 60:
                retry_after = int(state["minute_reset"] - now) + 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: 60 requests per minute",
                    tenant_id=tenant_id,
                    limit=60,
                    window_seconds=60,
                    retry_after=retry_after,
                )

            if state["hour_count"] >= 1000:
                retry_after = int(state["hour_reset"] - now) + 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: 1000 requests per hour",
                    tenant_id=tenant_id,
                    limit=1000,
                    window_seconds=3600,
                    retry_after=retry_after,
                )

            if state["day_count"] >= 10000:
                retry_after = int(state["day_reset"] - now) + 1
                raise RateLimitExceeded(
                    f"Rate limit exceeded: 10000 requests per day",
                    tenant_id=tenant_id,
                    limit=10000,
                    window_seconds=86400,
                    retry_after=retry_after,
                )

            state["minute_count"] += 1
            state["hour_count"] += 1
            state["day_count"] += 1

            return True

    async def check_embed_limits(
        self,
        tenant_id: str,
        tokens: int,
        cost_usd: float = 0.0,
    ) -> bool:
        """Check if embedding operation is within token and cost limits.

        Args:
            tenant_id: Tenant ID
            tokens: Number of tokens to be consumed
            cost_usd: Cost in USD

        Returns:
            True if within limits

        Raises:
            RateLimitExceeded: If limit exceeded
        """
        redis = await self._get_redis()

        if not redis:
            return await self._check_embed_limits_local(tenant_id, tokens, cost_usd)

        now = time.time()
        token_minute_key = self._get_key(tenant_id, "embed_tokens_minute")
        cost_day_key = self._get_key(tenant_id, "embed_cost_day")

        try:
            token_count = await redis._client.incr(token_minute_key)
            if token_count == 1:
                await redis._client.expire(token_minute_key, 60)

            if token_count + tokens > 100000:
                raise RateLimitExceeded(
                    f"Embedding token limit exceeded: 100000 tokens per minute",
                    tenant_id=tenant_id,
                    limit=100000,
                    window_seconds=60,
                    retry_after=60,
                )

            current_cost_str = await redis.get(cost_day_key)
            current_cost = float(current_cost_str) if current_cost_str else 0.0

            if current_cost + cost_usd > 50.0:
                raise RateLimitExceeded(
                    f"Embedding cost limit exceeded: $50.00 per day",
                    tenant_id=tenant_id,
                    limit=50,
                    window_seconds=86400,
                    retry_after=86400 - int(now % 86400),
                )

            new_cost = current_cost + cost_usd
            await redis.set(cost_day_key, str(new_cost), ex=86400)

            return True

        except RateLimitExceeded:
            raise
        except Exception as e:
            self._logger.warning(f"Redis embed limit check failed, falling back to local: {e}")
            return await self._check_embed_limits_local(tenant_id, tokens, cost_usd)

    async def _check_embed_limits_local(
        self,
        tenant_id: str,
        tokens: int,
        cost_usd: float = 0.0,
    ) -> bool:
        """Local fallback for embedding limits."""
        with self._lock:
            if tenant_id not in self._local_fallback:
                self._local_fallback[tenant_id] = {
                    "embed_tokens_minute": 0,
                    "embed_cost_day": 0.0,
                    "embed_token_reset": time.time() + 60,
                    "day_reset": time.time() + 86400,
                }

            state = self._local_fallback[tenant_id]
            now = time.time()

            if now >= state["embed_token_reset"]:
                state["embed_tokens_minute"] = 0
                state["embed_token_reset"] = now + 60

            if now >= state["day_reset"]:
                state["embed_cost_day"] = 0.0
                state["day_reset"] = now + 86400

            if state["embed_tokens_minute"] + tokens > 100000:
                raise RateLimitExceeded(
                    f"Embedding token limit exceeded: 100000 tokens per minute",
                    tenant_id=tenant_id,
                    limit=100000,
                    window_seconds=60,
                    retry_after=60,
                )

            if state["embed_cost_day"] + cost_usd > 50.0:
                raise RateLimitExceeded(
                    f"Embedding cost limit exceeded: $50.00 per day",
                    tenant_id=tenant_id,
                    limit=50,
                    window_seconds=86400,
                    retry_after=int(state["day_reset"] - now),
                )

            state["embed_tokens_minute"] += tokens
            state["embed_cost_day"] += cost_usd

            return True

    async def get_usage(self, tenant_id: str) -> Dict[str, int]:
        """Get current usage for a tenant.

        Args:
            tenant_id: Tenant ID

        Returns:
            Dictionary with usage statistics
        """
        redis = await self._get_redis()

        if not redis:
            return await self._get_usage_local(tenant_id)

        try:
            minute_key = self._get_key(tenant_id, "minute")
            hour_key = self._get_key(tenant_id, "hour")
            day_key = self._get_key(tenant_id, "day")

            minute_count = await redis._client.get(minute_key)
            hour_count = await redis._client.get(hour_key)
            day_count = await redis._client.get(day_key)

            return {
                "minute": int(minute_count) if minute_count else 0,
                "hour": int(hour_count) if hour_count else 0,
                "day": int(day_count) if day_count else 0,
                "limit_minute": 60,
                "limit_hour": 1000,
                "limit_day": 10000,
            }
        except Exception as e:
            self._logger.warning(f"Redis get_usage failed, falling back to local: {e}")
            return await self._get_usage_local(tenant_id)

    async def _get_usage_local(self, tenant_id: str) -> Dict[str, int]:
        """Local fallback for get_usage."""
        with self._lock:
            state = self._local_fallback.get(tenant_id, {})
            return {
                "minute": state.get("minute_count", 0),
                "hour": state.get("hour_count", 0),
                "day": state.get("day_count", 0),
                "limit_minute": 60,
                "limit_hour": 1000,
                "limit_day": 10000,
            }

    async def reset(self, tenant_id: str) -> None:
        """Reset rate limits for a tenant.

        Args:
            tenant_id: Tenant ID
        """
        redis = await self._get_redis()

        if redis:
            try:
                minute_key = self._get_key(tenant_id, "minute")
                hour_key = self._get_key(tenant_id, "hour")
                day_key = self._get_key(tenant_id, "day")
                burst_key = self._get_key(tenant_id, "burst")
                token_minute_key = self._get_key(tenant_id, "embed_tokens_minute")
                cost_day_key = self._get_key(tenant_id, "embed_cost_day")

                await redis.delete(
                    minute_key, hour_key, day_key, burst_key,
                    token_minute_key, cost_day_key
                )
            except Exception as e:
                self._logger.warning(f"Redis reset failed: {e}")

        with self._lock:
            if tenant_id in self._local_fallback:
                del self._local_fallback[tenant_id]

    async def get_stats(self) -> Dict[str, int]:
        """Get rate limiter statistics.

        Returns:
            Dictionary with stats
        """
        with self._lock:
            return {
                "local_fallback_tenants": len(self._local_fallback),
                "redis_available": self._redis is not None,
            }


_redis_rate_limiter: Optional[RedisRateLimiter] = None


async def get_redis_rate_limiter() -> RedisRateLimiter:
    """Get the global Redis-backed rate limiter.

    Returns:
        RedisRateLimiter instance
    """
    global _redis_rate_limiter
    if _redis_rate_limiter is None:
        _redis_rate_limiter = RedisRateLimiter()
    return _redis_rate_limiter
