"""Base provider abstract class for LLM providers.

This module defines the abstract interface that all LLM providers must implement.
"""

from abc import ABC, abstractmethod
from collections.abc import AsyncGenerator
from enum import StrEnum
from typing import Any, Optional
import asyncio
import logging
import threading
import time


class CircuitBreakerOpenError(Exception):
    """Raised when a circuit breaker is open and requests should fail fast."""
    pass


from ..models.schemas import (
    ChatMessage,
    CompletionResponse,
    EmbeddingResponse,
    ProviderInfo,
    CostTracking,
    ThinkingConfig,
)
from ..observability.metrics import get_metrics_collector


logger = logging.getLogger(__name__)


class RateLimiter:
    """Simple rate limiter using token bucket algorithm.

    Args:
        rate: Maximum requests per minute
    """

    def __init__(self, rate: int):
        self.rate = rate
        self.interval = 60.0 / rate  # seconds between requests
        self.last_request = 0.0
        self._lock = asyncio.Lock()

    async def acquire(self) -> None:
        """Wait until a request can be made."""
        async with self._lock:
            now = time.time()
            wait_time = self.last_request + self.interval - now
            if wait_time > 0:
                await asyncio.sleep(wait_time)
            self.last_request = time.time()


class RetryConfig:
    """Configuration for retry logic with exponential backoff."""

    def __init__(
        self,
        max_retries: int = 3,
        base_delay: float = 1.0,
        max_delay: float = 30.0,
        exponential_base: float = 2.0,
    ):
        self.max_retries = max_retries
        self.base_delay = base_delay
        self.max_delay = max_delay
        self.exponential_base = exponential_base

    def get_delay(self, attempt: int) -> float:
        """Calculate delay for the given attempt number."""
        delay = self.base_delay * (self.exponential_base ** attempt)
        return min(delay, self.max_delay)


class CircuitState(StrEnum):
    """Circuit breaker states."""
    CLOSED = "closed"      # Normal operation, requests pass through
    OPEN = "open"          # Failing fast, no requests allowed
    HALF_OPEN = "half_open"  # Testing if service recovered


class CircuitBreaker:
    """Circuit breaker pattern to prevent cascading failures.

    States:
    - CLOSED: Normal operation. Failures count towards threshold.
    - OPEN: Too many failures. Requests fail immediately without trying.
    - HALF_OPEN: After recovery_timeout, allow one test request.

    Transitions:
    - CLOSED -> OPEN: When failure_count >= failure_threshold within window
    - OPEN -> HALF_OPEN: When recovery_timeout elapsed since opening
    - HALF_OPEN -> CLOSED: On successful request
    - HALF_OPEN -> OPEN: On failed request
    """

    def __init__(
        self,
        name: str,
        failure_threshold: int = 5,
        recovery_timeout: float = 30.0,
        half_open_max_calls: int = 1,
    ):
        self.name = name
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.half_open_max_calls = half_open_max_calls

        self._state = CircuitState.CLOSED
        self._failure_count = 0
        self._failure_window_start: float | None = None
        self._last_failure_time: float | None = None
        self._half_open_calls = 0
        self._lock = threading.Lock()

    @property
    def state(self) -> CircuitState:
        """Get current circuit state, checking for timeout-based transitions."""
        with self._lock:
            if self._state == CircuitState.OPEN:
                if self._last_failure_time and \
                   time.time() - self._last_failure_time >= self.recovery_timeout:
                    logger.info(f"Circuit breaker {self.name}: OPEN -> HALF_OPEN (recovery timeout elapsed)")
                    self._state = CircuitState.HALF_OPEN
                    self._half_open_calls = 0
            return self._state

    def is_available(self) -> bool:
        """Check if requests can be made through this circuit."""
        state = self.state
        if state == CircuitState.CLOSED:
            return True
        if state == CircuitState.OPEN:
            return False
        # HALF_OPEN: allow limited calls
        with self._lock:
            return self._half_open_calls < self.half_open_max_calls

    def record_success(self) -> None:
        """Record a successful request, potentially closing the circuit."""
        with self._lock:
            old_state = self._state
            if self._state == CircuitState.HALF_OPEN:
                logger.info(f"Circuit breaker {self.name}: HALF_OPEN -> CLOSED (success)")
                self._state = CircuitState.CLOSED
                self._failure_count = 0
                self._failure_window_start = None
                self._half_open_calls = 0
            elif self._state == CircuitState.CLOSED:
                # Reset failure count on success
                self._failure_count = 0
                self._failure_window_start = None

            if old_state != self._state:
                try:
                    metrics = get_metrics_collector()
                    service_name = self.name.replace("provider.", "")
                    metrics.record_circuit_breaker_state(service_name, self._state.value)
                except Exception:
                    pass

    def record_failure(self, exc: Exception | None = None) -> None:
        """Record a failed request, potentially opening the circuit."""
        with self._lock:
            old_state = self._state
            now = time.time()

            # Reset window if it has expired (1 minute window)
            if self._failure_window_start and now - self._failure_window_start > 60:
                self._failure_count = 0
                self._failure_window_start = now
            elif not self._failure_window_start:
                self._failure_window_start = now

            self._failure_count += 1
            self._last_failure_time = now

            if self._state == CircuitState.HALF_OPEN:
                logger.warning(f"Circuit breaker {self.name}: HALF_OPEN -> OPEN (failure during recovery test)")
                self._state = CircuitState.OPEN
                self._half_open_calls = 0

            elif self._state == CircuitState.CLOSED:
                if self._failure_count >= self.failure_threshold:
                    logger.error(
                        f"Circuit breaker {self.name}: CLOSED -> OPEN "
                        f"(failed {self._failure_count} times in window, threshold={self.failure_threshold})"
                    )
                    self._state = CircuitState.OPEN

            if old_state != self._state:
                try:
                    metrics = get_metrics_collector()
                    service_name = self.name.replace("provider.", "")
                    metrics.record_circuit_breaker_state(service_name, self._state.value)
                    error_type = type(exc).__name__ if exc else "unknown"
                    metrics.record_circuit_breaker_error(service_name, error_type)
                except Exception:
                    pass

    def can_attempt(self) -> bool:
        """Check if we can attempt a request (must be called before attempting)."""
        state = self.state
        if state == CircuitState.CLOSED:
            return True
        if state == CircuitState.OPEN:
            return False
        # HALF_OPEN
        with self._lock:
            if self._half_open_calls >= self.half_open_max_calls:
                return False
            self._half_open_calls += 1
            return True

    def get_stats(self) -> dict[str, int | str | None]:
        """Get circuit breaker statistics."""
        with self._lock:
            return {
                "name": self.name,
                "state": self._state.value,
                "failure_count": self._failure_count,
                "failure_threshold": self.failure_threshold,
                "recovery_timeout": self.recovery_timeout,
                "last_failure_time": self._last_failure_time,
            }


class BaseProvider(ABC):
    """Abstract base class for LLM providers.

    All providers must implement the abstract methods defined here.
    """

    def __init__(
        self,
        name: str,
        display_name: str,
        rate_limit: int,
        retry_config: Optional[RetryConfig] = None,
        api_key: Optional[str] = None,
        circuit_breaker: Optional[CircuitBreaker] = None,
    ):
        self.name = name
        self.display_name = display_name
        self.rate_limiter = RateLimiter(rate_limit)
        self.retry_config = retry_config or RetryConfig()
        self._available = True
        self._models: list[str] = []
        self._byok = api_key is not None  # True if using a BYOK key
        self._circuit_breaker = circuit_breaker or CircuitBreaker(
            name=f"provider.{name}",
            failure_threshold=5,
            recovery_timeout=30.0,
        )

    @property
    def circuit_breaker(self) -> CircuitBreaker:
        """Get the circuit breaker for this provider."""
        return self._circuit_breaker

    @property
    def available(self) -> bool:
        """Whether this provider is currently available."""
        return self._available and self._circuit_breaker.is_available()

    @property
    def models(self) -> list[str]:
        """List of available models for this provider."""
        return self._models

    @property
    def supports_embeddings(self) -> bool:
        """Whether this provider supports embeddings."""
        return True  # Default to True, subclasses can override

    @abstractmethod
    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
        thinking: Optional[ThinkingConfig] = None,
    ) -> CompletionResponse:
        """Generate a completion response.

        Args:
            messages: List of chat messages
            model: Model to use (provider-specific)
            temperature: Sampling temperature
            max_tokens: Maximum tokens to generate
            top_p: Nucleus sampling parameter
            stop: Stop sequences
            thinking: Optional thinking/reasoning configuration

        Returns:
            CompletionResponse with the generated content
        """
        pass

    @abstractmethod
    async def stream(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
        thinking: Optional[ThinkingConfig] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream completion responses token by token.

        Args:
            messages: List of chat messages
            model: Model to use (provider-specific)
            temperature: Sampling temperature
            max_tokens: Maximum tokens to generate
            top_p: Nucleus sampling parameter
            stop: Stop sequences
            thinking: Optional thinking/reasoning configuration

        Yields:
            Text chunks as they are generated
        """
        pass

    @abstractmethod
    async def embed(
        self,
        text: str,
        model: Optional[str] = None,
        dimensions: Optional[int] = None,
    ) -> EmbeddingResponse:
        """Generate embeddings for the given text.

        Args:
            text: Text to embed
            model: Embedding model to use
            dimensions: Number of dimensions for the output

        Returns:
            EmbeddingResponse with the embedding vector
        """
        pass

    @abstractmethod
    def get_provider_info(self) -> ProviderInfo:
        """Get information about this provider.

        Returns:
            ProviderInfo with provider details
        """
        pass

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Calculate the cost for a request.

        Args:
            model: Model used
            input_tokens: Number of input tokens
            output_tokens: Number of output tokens

        Returns:
            CostTracking with cost breakdown
        """
        # Default implementation - subclasses should override
        total_tokens = input_tokens + output_tokens
        return CostTracking(
            provider=self.name,
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            total_tokens=total_tokens,
            estimated_cost=0.0,
        )

    async def _retry_with_backoff(
        self,
        func,
        *args,
        **kwargs,
    ) -> Any:
        """Execute a function with exponential backoff retry logic and circuit breaker.

        Args:
            func: Async function to execute
            *args: Positional arguments for func
            **kwargs: Keyword arguments for func

        Returns:
            Result from func

        Raises:
            Last exception if all retries fail
        """
        # Check circuit breaker before attempting
        if not self._circuit_breaker.can_attempt():
            logger.warning(
                f"Provider {self.name} circuit breaker is OPEN - failing fast"
            )
            raise CircuitBreakerOpenError(
                f"Provider {self.name} circuit breaker is open, not attempting request"
            )

        last_exception = None

        for attempt in range(self.retry_config.max_retries + 1):
            try:
                result = await func(*args, **kwargs)
                self._circuit_breaker.record_success()
                return result
            except CircuitBreakerOpenError:
                # Don't retry if circuit is open
                raise
            except Exception as e:
                last_exception = e
                # Don't back off/retry for deterministic client/auth failures.
                status_code = getattr(e, "status_code", None)
                msg = str(e)
                non_retryable = (
                    (isinstance(status_code, int) and 400 <= status_code < 500 and status_code != 429)
                    or "401" in msg
                    or "403" in msg
                )
                if non_retryable:
                    logger.error(
                        f"Provider {self.name} non-retryable error: {e}"
                    )
                    self._circuit_breaker.record_failure(e)
                    raise
                if attempt < self.retry_config.max_retries:
                    delay = self.retry_config.get_delay(attempt)
                    logger.warning(
                        f"Provider {self.name} attempt {attempt + 1} failed: {e}. "
                        f"Retrying in {delay:.2f}s..."
                    )
                    await asyncio.sleep(delay)
                else:
                    logger.error(
                        f"Provider {self.name} all retries exhausted: {e}"
                    )
                    self._circuit_breaker.record_failure(e)

        raise last_exception

    async def health_check(self) -> bool:
        """Check if the provider is healthy.

        Returns:
            True if provider is healthy, False otherwise
        """
        try:
            # Simple health check - try to get provider info
            await asyncio.wait_for(
                self.embed("health check"),
                timeout=10.0,
            )
            self._available = True
            return True
        except Exception as e:
            logger.warning(f"Provider {self.name} health check failed: {e}")
            self._available = False
            return False
