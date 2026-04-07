"""Redis client factory supporting both standard Redis and Upstash Redis."""

import logging
from typing import Optional

logger = logging.getLogger(__name__)


class RedisClient:
    """Abstraction layer for Redis operations supporting both standard Redis and Upstash."""

    def __init__(self, client):
        self._client = client
        self._is_upstash = False

    @classmethod
    async def create(cls) -> "RedisClient":
        """Create a Redis client based on configuration.

        Returns Upstash Redis if configured, otherwise standard Redis.
        """
        from upstash_redis import Redis as UpstashRedis

        from ..config import settings

        # Check if using Upstash
        if settings.use_upstash_redis and settings.upstash_redis_url and settings.upstash_redis_token:
            try:
                client = UpstashRedis(url=settings.upstash_redis_url, token=settings.upstash_redis_token)
                instance = cls(client)
                instance._is_upstash = True
                await client.ping()
                logger.info("Using Upstash Redis")
                return instance
            except Exception as e:
                logger.warning(f"Failed to connect to Upstash Redis: {e}, falling back to standard Redis")

        # Fall back to standard Redis
        import redis.asyncio as redis

        try:
            client = redis.from_url(
                settings.redis_url,
                encoding="utf-8",
                decode_responses=True,
            )
            await client.ping()
            logger.info("Using standard Redis")
            return cls(client)
        except Exception as e:
            logger.warning(f"Failed to connect to Redis: {e}")
            return cls(None)

    async def get(self, key: str) -> Optional[str]:
        """Get a value from Redis."""
        if self._client is None:
            return None
        if self._is_upstash:
            return await self._client.get(key)
        return await self._client.get(key)

    async def set(self, key: str, value: str, ex: Optional[int] = None) -> bool:
        """Set a value in Redis with optional expiration in seconds."""
        if self._client is None:
            return False
        if self._is_upstash:
            if ex:
                await self._client.setex(key, ex, value)
            else:
                await self._client.set(key, value)
            return True
        if ex:
            await self._client.setex(key, ex, value)
        else:
            await self._client.set(key, value)
        return True

    async def setex(self, key: str, seconds: int, value: str) -> bool:
        """Set a value with expiration."""
        return await self.set(key, value, ex=seconds)

    async def delete(self, *keys: str) -> int:
        """Delete one or more keys."""
        if self._client is None:
            return 0
        if self._is_upstash:
            deleted = 0
            for key in keys:
                result = await self._client.delete(key)
                if result:
                    deleted += 1
            return deleted
        return await self._client.delete(*keys)

    async def sadd(self, key: str, *values: str) -> int:
        """Add to a set."""
        if self._client is None:
            return 0
        if self._is_upstash:
            result = await self._client.sadd(key, *values)
            return result if isinstance(result, int) else len(values)
        return await self._client.sadd(key, *values)

    async def srem(self, key: str, *values: str) -> int:
        """Remove from a set."""
        if self._client is None:
            return 0
        if self._is_upstash:
            result = await self._client.srem(key, *values)
            return result if isinstance(result, int) else len(values)
        return await self._client.srem(key, *values)

    async def smembers(self, key: str) -> set:
        """Get all members of a set."""
        if self._client is None:
            return set()
        if self._is_upstash:
            result = await self._client.smembers(key)
            return set(result) if result else set()
        return await self._client.smembers(key)

    async def expire(self, key: str, seconds: int) -> bool:
        """Set expiration on a key."""
        if self._client is None:
            return False
        if self._is_upstash:
            # Upstash handles this via setex
            return True
        return await self._client.expire(key, seconds)

    async def ttl(self, key: str) -> int:
        """Get TTL of a key."""
        if self._client is None:
            return -1
        if self._is_upstash:
            result = await self._client.ttl(key)
            return result if isinstance(result, int) else -1
        return await self._client.ttl(key)

    async def ping(self) -> bool:
        """Ping Redis to check connection."""
        if self._client is None:
            return False
        try:
            if self._is_upstash:
                await self._client.ping()
            else:
                await self._client.ping()
            return True
        except Exception:
            return False

    async def close(self):
        """Close the Redis connection."""
        if self._client is None:
            return
        if not self._is_upstash:
            await self._client.close()
        # Upstash HTTP client doesn't need explicit close
