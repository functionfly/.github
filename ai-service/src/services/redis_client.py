"""Redis client factory supporting both standard Redis and Upstash Redis.

This module provides a robust Redis client with:
- Automatic retry with exponential backoff on connection failures
- Circuit breaker pattern to prevent cascading failures
- Connection health monitoring
- Support for both standard Redis and Upstash Redis
- In-memory fallback when Redis is unavailable
"""

import asyncio
import logging
import threading
import time
from typing import Optional, Callable, Any

logger = logging.getLogger(__name__)


class CircuitOpenError(Exception):
    """Raised when circuit breaker is open and operations should fail fast."""
    pass


class RedisCircuitBreaker:
    """Circuit breaker for Redis operations.

    States:
    - CLOSED: Normal operation, requests pass through
    - OPEN: Too many failures, requests fail immediately
    - HALF_OPEN: Testing if Redis recovered
    """

    def __init__(
        self,
        failure_threshold: int = 5,
        recovery_timeout: float = 30.0,
        half_open_max_calls: int = 1,
    ):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.half_open_max_calls = half_open_max_calls

        self._state = "closed"
        self._failure_count = 0
        self._last_failure_time: float | None = None
        self._half_open_calls = 0
        self._lock = threading.Lock()

        # Metrics
        self._total_failures = 0
        self._total_successes = 0
        self._total_retries = 0

    def _get_state(self) -> str:
        """Get current state with timeout-based transitions."""
        with self._lock:
            if self._state == "open":
                if self._last_failure_time and \
                   time.time() - self._last_failure_time >= self.recovery_timeout:
                    logger.info("Redis circuit breaker: OPEN -> HALF_OPEN")
                    self._state = "half_open"
                    self._half_open_calls = 0
            return self._state

    def can_attempt(self) -> bool:
        """Check if we can attempt a Redis operation."""
        state = self._get_state()
        if state == "closed":
            return True
        if state == "open":
            return False
        # HALF_OPEN
        with self._lock:
            if self._half_open_calls >= self.half_open_max_calls:
                return False
            self._half_open_calls += 1
            return True

    def record_success(self) -> None:
        """Record a successful Redis operation."""
        with self._lock:
            self._total_successes += 1
            if self._state == "half_open":
                logger.info("Redis circuit breaker: HALF_OPEN -> CLOSED")
                self._state = "closed"
                self._failure_count = 0
            elif self._state == "closed":
                self._failure_count = 0

    def record_failure(self, exc: Exception | None = None) -> None:
        """Record a failed Redis operation."""
        with self._lock:
            self._total_failures += 1
            self._failure_count += 1
            self._last_failure_time = time.time()

            if self._state == "half_open":
                logger.warning("Redis circuit breaker: HALF_OPEN -> OPEN (failure during recovery)")
                self._state = "open"
                self._half_open_calls = 0
            elif self._state == "closed":
                if self._failure_count >= self.failure_threshold:
                    logger.error(
                        f"Redis circuit breaker: CLOSED -> OPEN "
                        f"(failed {self._failure_count} times, threshold={self.failure_threshold})"
                    )
                    self._state = "open"

    def record_retry(self) -> None:
        """Record a retry attempt."""
        with self._lock:
            self._total_retries += 1

    def get_stats(self) -> dict[str, Any]:
        """Get circuit breaker statistics."""
        with self._lock:
            return {
                "state": self._state,
                "failure_count": self._failure_count,
                "failure_threshold": self.failure_threshold,
                "recovery_timeout": self.recovery_timeout,
                "last_failure_time": self._last_failure_time,
                "total_failures": self._total_failures,
                "total_successes": self._total_successes,
                "total_retries": self._total_retries,
            }


class RedisClient:
    """Abstraction layer for Redis operations supporting both standard Redis and Upstash.

    Features:
    - Automatic retry with exponential backoff on connection failures
    - Circuit breaker pattern to prevent cascading failures
    - Connection health monitoring
    - Encryption support
    """

    def __init__(self, client):
        self._client = client
        self._is_upstash = False
        self._encryption = None
        self._circuit_breaker = RedisCircuitBreaker()
        self._connection_retries = 0
        self._last_connection_time: float | None = None
        self._connection_failures = 0

        # Retry configuration
        self._max_retries = 3
        self._base_delay = 0.5
        self._max_delay = 30.0

    @classmethod
    async def create(cls) -> "RedisClient":
        """Create a Redis client based on configuration.

        Returns Upstash Redis if configured, otherwise standard Redis.
        """
        from upstash_redis import Redis as UpstashRedis

        from ..config import settings
        from ..security.encryption import get_cache_encryption

        instance = cls.__new__(cls)
        instance._client = None
        instance._is_upstash = False
        instance._encryption = get_cache_encryption()
        instance._circuit_breaker = RedisCircuitBreaker()
        instance._connection_retries = 0
        instance._last_connection_time = None
        instance._connection_failures = 0

        # Load retry config from settings
        instance._max_retries = getattr(settings, 'redis_max_connection_retries', 3)
        instance._base_delay = getattr(settings, 'redis_base_retry_delay', 0.5)
        instance._max_delay = getattr(settings, 'redis_max_retry_delay', 30.0)

        if instance._encryption.is_enabled:
            logger.info("Cache encryption is enabled")

        # Check if using Upstash
        if settings.use_upstash_redis and settings.upstash_redis_url and settings.upstash_redis_token:
            instance = await cls._connect_upstash(instance, settings)
            if instance._client:
                return instance
            logger.warning("Upstash connection failed, falling back to standard Redis")

        # Fall back to standard Redis
        instance = await cls._connect_standard(instance, settings)
        return instance

    @classmethod
    async def _connect_upstash(cls, instance: "RedisClient", settings) -> "RedisClient":
        """Connect to Upstash Redis with retry logic."""
        from upstash_redis import Redis as UpstashRedis

        for attempt in range(instance._max_retries + 1):
            try:
                client = UpstashRedis(url=settings.upstash_redis_url, token=settings.upstash_redis_token)
                await client.ping()
                instance._client = client
                instance._is_upstash = True
                instance._last_connection_time = time.time()
                instance._connection_retries = attempt
                logger.info(f"Using Upstash Redis (connected after {attempt} retries)")
                return instance
            except Exception as e:
                instance._connection_failures += 1
                instance._circuit_breaker.record_failure(e)
                if attempt < instance._max_retries:
                    delay = min(instance._base_delay * (2 ** attempt), instance._max_delay)
                    logger.warning(
                        f"Upstash connection attempt {attempt + 1} failed: {e}. "
                        f"Retrying in {delay:.2f}s..."
                    )
                    instance._circuit_breaker.record_retry()
                    await asyncio.sleep(delay)
                else:
                    logger.error(f"Upstash connection failed after {attempt + 1} attempts: {e}")
        return instance

    @classmethod
    async def _connect_standard(cls, instance: "RedisClient", settings) -> "RedisClient":
        """Connect to standard Redis with retry logic."""
        import redis.asyncio as redis

        for attempt in range(instance._max_retries + 1):
            try:
                # Build Redis URL with password if provided
                redis_url = settings.redis_url
                if settings.redis_password:
                    if "://" in redis_url:
                        parts = redis_url.split("://", 1)
                        auth_part = f":{settings.redis_password}@"
                        redis_url = parts[0] + "://" + auth_part + parts[1]
                    else:
                        redis_url = f"redis://:{settings.redis_password}@{redis_url.replace('redis://', '')}"

                # Auto-detect cloud Redis and enable TLS if not explicitly set
                use_tls = settings.redis_use_tls
                if not use_tls:
                    # Check if this looks like a cloud Redis URL
                    if cls._is_cloud_redis(redis_url):
                        logger.info(
                            f"Cloud Redis detected ('{redis_url}'), auto-enabling TLS. "
                            "Set REDIS_USE_TLS=false to disable if not needed."
                        )
                        use_tls = True

                client = redis.from_url(
                    redis_url,
                    encoding="utf-8",
                    decode_responses=True,
                    ssl=use_tls,
                )
                await client.ping()
                instance._client = client
                instance._last_connection_time = time.time()
                instance._connection_retries = attempt
                logger.info(f"Using standard Redis (TLS: {use_tls}, retries: {attempt})")
                return instance
            except Exception as e:
                instance._connection_failures += 1
                instance._circuit_breaker.record_failure(e)
                if attempt < instance._max_retries:
                    delay = min(instance._base_delay * (2 ** attempt), instance._max_delay)
                    logger.warning(
                        f"Redis connection attempt {attempt + 1} failed: {e}. "
                        f"Retrying in {delay:.2f}s..."
                    )
                    instance._circuit_breaker.record_retry()
                    await asyncio.sleep(delay)
                else:
                    logger.error(f"Redis connection failed after {attempt + 1} attempts: {e}")
        return instance

    CLOUD_REDIS_PATTERNS = [
        r"\.upstash\.io$",
        r"\.redis\.cloud$",
        r"\.redislabs\.com$",
        r"\.aws\.amazon\.com/elasticache",
        r"\.gcp\.googleapis\.com/memstore",
        r"\.azure\.com/cache",
        r"clustered\.redis\.com$",
    ]

    @classmethod
    def _is_cloud_redis(cls, redis_url: str) -> bool:
        """Check if Redis URL appears to be a cloud-hosted instance."""
        import re
        redis_url_lower = redis_url.lower()
        for pattern in cls.CLOUD_REDIS_PATTERNS:
            if re.search(pattern, redis_url_lower):
                return True
        return False

    def is_connected(self) -> bool:
        """Check if Redis client is initialized."""
        return self._client is not None

    def connection_health(self) -> dict[str, Any]:
        """Get connection health statistics."""
        return {
            "is_connected": self.is_connected(),
            "is_upstash": self._is_upstash,
            "connection_retries": self._connection_retries,
            "connection_failures": self._connection_failures,
            "last_connection_time": self._last_connection_time,
            "circuit_breaker": self._circuit_breaker.get_stats(),
        }

    async def reconnect(self) -> bool:
        """Explicitly attempt to reconnect to Redis.

        Returns:
            True if reconnection successful, False otherwise
        """
        if self._client is not None:
            try:
                await self._client.close()
            except Exception:
                pass

        self._client = None
        new_instance = await self.__class__.create()
        if new_instance._client:
            self._client = new_instance._client
            self._is_upstash = new_instance._is_upstash
            self._last_connection_time = time.time()
            logger.info("Redis reconnection successful")
            return True
        return False

        async def _try_standard_redis():
            import redis.asyncio as redis_async
            redis_url = settings.redis_url
            if settings.redis_password:
                if "://" in redis_url:
                    parts = redis_url.split("://", 1)
                    auth_part = f":{settings.redis_password}@"
                    redis_url = parts[0] + "://" + auth_part + parts[1]
                else:
                    redis_url = f"redis://:{settings.redis_password}@{redis_url.replace('redis://', '')}"

            client = redis_async.from_url(
                redis_url,
                encoding="utf-8",
                decode_responses=True,
                ssl=settings.redis_use_tls,
            )
            await client.ping()
            logger.info(f"Using standard Redis (TLS: {settings.redis_use_tls})")
            instance._client = client
            return instance

        try:
            if settings.use_upstash_redis and settings.upstash_redis_url and settings.upstash_redis_token:
                try:
                    await instance._circuit_breaker.call(_try_upstash)
                    return instance
                except CircuitOpenError:
                    logger.warning("Redis circuit breaker open, cannot connect to Upstash")
                except Exception as e:
                    logger.warning(f"Failed to connect to Upstash Redis: {e}, falling back to standard Redis")

            await instance._circuit_breaker.call(_try_standard_redis)
            return instance
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, cannot connect to standard Redis")
            return instance
        except Exception as e:
            logger.warning(f"Failed to connect to Redis: {e}")
            return instance

    def _encrypt_value(self, value: str) -> str:
        """Encrypt a value if encryption is enabled."""
        if self._encryption and self._encryption.is_enabled:
            return self._encryption.encrypt(value)
        return value

    def _decrypt_value(self, value: str) -> str:
        """Decrypt a value if encryption is enabled."""
        if self._encryption and self._encryption.is_enabled:
            try:
                return self._encryption.decrypt(value)
            except ValueError:
                return value
        return value

    async def get(self, key: str) -> Optional[str]:
        """Get a value from Redis."""
        if self._client is None:
            return self._get_from_fallback(key)

        async def _do_get():
            return await self._client.get(key)

        try:
            value = await self._circuit_breaker.call(_do_get)
            if value and self._encryption and self._encryption.is_enabled:
                return self._decrypt_value(value)
            return value
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return self._get_from_fallback(key)
        except Exception as e:
            logger.warning(f"Redis get failed: {e}, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", type(e).__name__)
            return self._get_from_fallback(key)

    def _get_from_fallback(self, key: str) -> Optional[str]:
        with _fallback_lock:
            return _fallback_storage.get(key)

    async def set(self, key: str, value: str, ex: Optional[int] = None) -> bool:
        """Set a value in Redis with optional expiration in seconds."""
        encrypted_value = self._encrypt_value(value) if self._encryption and self._encryption.is_enabled else value

        if self._client is None:
            return self._set_to_fallback(key, encrypted_value, ex)

        async def _do_set():
            if self._is_upstash:
                if ex:
                    await self._client.setex(key, ex, encrypted_value)
                else:
                    await self._client.set(key, encrypted_value)
            else:
                if ex:
                    await self._client.setex(key, ex, encrypted_value)
                else:
                    await self._client.set(key, encrypted_value)
            return True

        try:
            return await self._circuit_breaker.call(_do_set)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return self._set_to_fallback(key, encrypted_value, ex)
        except Exception as e:
            logger.warning(f"Redis set failed: {e}, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", type(e).__name__)
            return self._set_to_fallback(key, encrypted_value, ex)

    def _set_to_fallback(self, key: str, value: str, ex: Optional[int] = None) -> bool:
        with _fallback_lock:
            _fallback_storage[key] = value
            return True

    async def setex(self, key: str, seconds: int, value: str) -> bool:
        """Set a value with expiration."""
        return await self.set(key, value, ex=seconds)

    async def delete(self, *keys: str) -> int:
        """Delete one or more keys."""
        if self._client is None:
            return self._delete_from_fallback(*keys)

        async def _do_delete():
            if self._is_upstash:
                deleted = 0
                for key in keys:
                    result = await self._client.delete(key)
                    if result:
                        deleted += 1
                return deleted
            return await self._client.delete(*keys)

        try:
            return await self._circuit_breaker.call(_do_delete)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return self._delete_from_fallback(*keys)
        except Exception as e:
            logger.warning(f"Redis delete failed: {e}, using fallback")
            return self._delete_from_fallback(*keys)

    def _delete_from_fallback(self, *keys: str) -> int:
        with _fallback_lock:
            deleted = 0
            for key in keys:
                if key in _fallback_storage:
                    del _fallback_storage[key]
                    deleted += 1
            return deleted

    async def sadd(self, key: str, *values: str) -> int:
        """Add to a set."""
        if self._client is None:
            return self._sadd_to_fallback(key, *values)

        encrypted_values = tuple(
            self._encryption.encrypt(v) if self._encryption and self._encryption.is_enabled else v
            for v in values
        )

        async def _do_sadd():
            if self._is_upstash:
                result = await self._client.sadd(key, *encrypted_values)
                return result if isinstance(result, int) else len(values)
            return await self._client.sadd(key, *encrypted_values)

        try:
            return await self._circuit_breaker.call(_do_sadd)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return self._sadd_to_fallback(key, *values)
        except Exception as e:
            logger.warning(f"Redis sadd failed: {e}, using fallback")
            return self._sadd_to_fallback(key, *values)

    def _sadd_to_fallback(self, key: str, *values: str) -> int:
        with _fallback_lock:
            if key not in _fallback_storage:
                _fallback_storage[key] = set()
            existing = _fallback_storage[key]
            if isinstance(existing, set):
                for v in values:
                    existing.add(v)
                return len(values)
            return 0

    async def srem(self, key: str, *values: str) -> int:
        """Remove from a set."""
        if self._client is None:
            return self._srem_from_fallback(key, *values)

        encrypted_values = tuple(
            self._encryption.encrypt(v) if self._encryption and self._encryption.is_enabled else v
            for v in values
        )

        async def _do_srem():
            if self._is_upstash:
                result = await self._client.srem(key, *encrypted_values)
                return result if isinstance(result, int) else len(values)
            return await self._client.srem(key, *encrypted_values)

        try:
            return await self._circuit_breaker.call(_do_srem)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return self._srem_from_fallback(key, *values)
        except Exception as e:
            logger.warning(f"Redis srem failed: {e}, using fallback")
            return self._srem_from_fallback(key, *values)

    def _srem_from_fallback(self, key: str, *values: str) -> int:
        with _fallback_lock:
            if key not in _fallback_storage:
                return 0
            existing = _fallback_storage[key]
            if isinstance(existing, set):
                removed = 0
                for v in values:
                    if v in existing:
                        existing.discard(v)
                        removed += 1
                return removed
            return 0

    async def smembers(self, key: str) -> set:
        """Get all members of a set."""
        if self._client is None:
            return self._smembers_from_fallback(key)

        async def _do_smembers():
            if self._is_upstash:
                result = await self._client.smembers(key)
                return set(result) if result else set()
            return await self._client.smembers(key)

        try:
            members = await self._circuit_breaker.call(_do_smembers)
            if self._encryption and self._encryption.is_enabled:
                decrypted = set()
                for m in members:
                    try:
                        decrypted.add(self._encryption.decrypt(m))
                    except ValueError:
                        decrypted.add(m)
                return decrypted
            return members
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return self._smembers_from_fallback(key)
        except Exception as e:
            logger.warning(f"Redis smembers failed: {e}, using fallback")
            return self._smembers_from_fallback(key)

    def _smembers_from_fallback(self, key: str) -> set:
        with _fallback_lock:
            val = _fallback_storage.get(key)
            if isinstance(val, set):
                return val.copy()
            return set()

    async def expire(self, key: str, seconds: int) -> bool:
        """Set expiration on a key."""
        if self._client is None:
            return True

        async def _do_expire():
            if self._is_upstash:
                return True
            return await self._client.expire(key, seconds)

        try:
            return await self._circuit_breaker.call(_do_expire)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return True
        except Exception as e:
            logger.warning(f"Redis expire failed: {e}, using fallback")
            return True

    async def ttl(self, key: str) -> int:
        """Get TTL of a key."""
        if self._client is None:
            return -1

        async def _do_ttl():
            if self._is_upstash:
                result = await self._client.ttl(key)
                return result if isinstance(result, int) else -1
            return await self._client.ttl(key)

        try:
            return await self._circuit_breaker.call(_do_ttl)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open, using fallback")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return -1
        except Exception as e:
            logger.warning(f"Redis ttl failed: {e}, using fallback")
            return -1

    async def ping(self) -> bool:
        """Ping Redis to check connection."""
        if self._client is None:
            return False

        async def _do_ping():
            await self._client.ping()
            return True

        try:
            return await self._circuit_breaker.call(_do_ping)
        except CircuitOpenError:
            logger.warning("Redis circuit breaker open")
            metrics = get_metrics_collector()
            metrics.record_circuit_breaker_error("redis", "circuit_open")
            return False
        except Exception:
            return False

    async def close(self):
        """Close the Redis connection."""
        if self._client is None:
            return
        if not self._is_upstash:
            await self._client.close()


_instance: Optional[RedisClient] = None


async def init_redis_client() -> RedisClient:
    """Create and store the module-level Redis client singleton."""
    global _instance
    _instance = await RedisClient.create()
    return _instance


def get_redis_client() -> Optional[RedisClient]:
    """Return the module-level Redis client singleton (may be None if not initialized)."""
    return _instance


# In-memory fallback storage when Redis is unavailable
_fallback_storage: dict = {}
_fallback_lock = threading.Lock()


def get_metrics_collector():
    """Get the metrics collector if available."""
    try:
        from ..observability.metrics import get_metrics_collector as _get
        return _get()
    except Exception:
        # Return a no-op metrics collector if not available
        return _NoOpMetricsCollector()


class _NoOpMetricsCollector:
    """No-op metrics collector when real one is not available."""

    def record_circuit_breaker_error(self, service: str, error_type: str):
        pass

    def increment(self, name: str, value: float = 1, **labels):
        pass
