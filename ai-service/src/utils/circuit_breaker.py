"""Circuit breaker utility for external API calls.

Provides a standalone circuit breaker implementation with Prometheus metrics integration.
"""

from enum import Enum
from typing import Callable, Any, Optional
import asyncio
import time
import logging
from collections import deque

logger = logging.getLogger(__name__)


class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"


class CircuitOpenError(Exception):
    """Raised when a circuit breaker is open and requests should fail fast."""
    pass


class CircuitBreaker:
    def __init__(
        self,
        name: str,
        failure_threshold: int = 5,
        recovery_timeout: float = 60.0,
        half_open_max_calls: int = 3,
        expected_exceptions: tuple = (Exception,),
    ):
        self.name = name
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.half_open_max_calls = half_open_max_calls
        self.expected_exceptions = expected_exceptions

        self._state = CircuitState.CLOSED
        self._failure_count = 0
        self._last_failure_time: float | None = None
        self._half_open_calls = 0
        self._recent_errors: deque = deque(maxlen=100)
        self._lock = asyncio.Lock()
        self._sync_lock = __import__("threading").Lock()

    @property
    def state(self) -> CircuitState:
        return self._state

    async def call(self, func: Callable, *args, **kwargs) -> Any:
        if self._state == CircuitState.OPEN:
            if self._last_failure_time and \
               time.time() - self._last_failure_time >= self.recovery_timeout:
                async with self._lock:
                    if self._state == CircuitState.OPEN:
                        self._state = CircuitState.HALF_OPEN
                        self._half_open_calls = 0
                        logger.info(f"Circuit breaker {self.name}: OPEN -> HALF_OPEN")
            else:
                raise CircuitOpenError(
                    f"Circuit breaker {self.name} is open, retry after {self.recovery_timeout}s"
                )

        if self._state == CircuitState.HALF_OPEN:
            async with self._lock:
                if self._half_open_calls >= self.half_open_max_calls:
                    raise CircuitOpenError(
                        f"Circuit breaker {self.name} half-open limit reached"
                    )
                self._half_open_calls += 1

        try:
            result = await func(*args, **kwargs)
            await self._on_success()
            return result
        except self.expected_exceptions as e:
            await self._on_failure(e)
            raise

    async def _on_success(self):
        async with self._lock:
            self._failure_count = 0
            if self._state == CircuitState.HALF_OPEN:
                logger.info(f"Circuit breaker {self.name}: HALF_OPEN -> CLOSED")
                self._state = CircuitState.CLOSED

    async def _on_failure(self, exc: Exception):
        async with self._lock:
            self._failure_count += 1
            self._last_failure_time = time.time()
            self._recent_errors.append(str(exc))

            if self._state == CircuitState.HALF_OPEN:
                logger.warning(
                    f"Circuit breaker {self.name}: HALF_OPEN -> OPEN (failure during recovery)"
                )
                self._state = CircuitState.OPEN
            elif self._failure_count >= self.failure_threshold:
                logger.error(
                    f"Circuit breaker {self.name}: CLOSED -> OPEN "
                    f"(failed {self._failure_count} times, threshold={self.failure_threshold})"
                )
                self._state = CircuitState.OPEN

    def get_stats(self) -> dict:
        with self._sync_lock:
            return {
                "name": self.name,
                "state": self._state.value,
                "failure_count": self._failure_count,
                "failure_threshold": self.failure_threshold,
                "recovery_timeout": self.recovery_timeout,
                "last_failure_time": self._last_failure_time,
                "recent_errors": list(self._recent_errors),
            }


class CircuitBreakerManager:
    _instance: Optional["CircuitBreakerManager"] = None

    def __init__(self):
        self._breakers: dict[str, CircuitBreaker] = {}
        self._lock = asyncio.Lock()

    @classmethod
    def get_instance(cls) -> "CircuitBreakerManager":
        if cls._instance is None:
            cls._instance = CircuitBreakerManager()
        return cls._instance

    def get_or_create(
        self,
        name: str,
        failure_threshold: int = 5,
        recovery_timeout: float = 60.0,
        half_open_max_calls: int = 3,
    ) -> CircuitBreaker:
        if name not in self._breakers:
            self._breakers[name] = CircuitBreaker(
                name=name,
                failure_threshold=failure_threshold,
                recovery_timeout=recovery_timeout,
                half_open_max_calls=half_open_max_calls,
            )
        return self._breakers[name]

    def get_all_stats(self) -> dict:
        return {name: cb.get_stats() for name, cb in self._breakers.items()}

    def reset_all(self):
        for cb in self._breakers.values():
            cb._state = CircuitState.CLOSED
            cb._failure_count = 0
            cb._half_open_calls = 0
