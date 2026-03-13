"""Base provider abstract class for LLM providers.

This module defines the abstract interface that all LLM providers must implement.
"""

from abc import ABC, abstractmethod
from typing import Optional, AsyncGenerator, Any
import asyncio
import logging
import time
import datetime
from datetime import datetime as dt

from ..models.schemas import (
    ChatMessage,
    CompletionResponse,
    EmbeddingResponse,
    ProviderInfo,
    CostTracking,
)


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
    ):
        self.name = name
        self.display_name = display_name
        self.rate_limiter = RateLimiter(rate_limit)
        self.retry_config = retry_config or RetryConfig()
        self._available = True
        self._models: list[str] = []

    @property
    def available(self) -> bool:
        """Whether this provider is currently available."""
        return self._available

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
    ) -> CompletionResponse:
        """Generate a completion response.

        Args:
            messages: List of chat messages
            model: Model to use (provider-specific)
            temperature: Sampling temperature
            max_tokens: Maximum tokens to generate
            top_p: Nucleus sampling parameter
            stop: Stop sequences

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
    ) -> AsyncGenerator[str, None]:
        """Stream completion responses token by token.

        Args:
            messages: List of chat messages
            model: Model to use (provider-specific)
            temperature: Sampling temperature
            max_tokens: Maximum tokens to generate
            top_p: Nucleus sampling parameter
            stop: Stop sequences

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
        """Execute a function with exponential backoff retry logic.

        Args:
            func: Async function to execute
            *args: Positional arguments for func
            **kwargs: Keyword arguments for func

        Returns:
            Result from func

        Raises:
            Last exception if all retries fail
        """
        last_exception = None

        for attempt in range(self.retry_config.max_retries + 1):
            try:
                return await func(*args, **kwargs)
            except Exception as e:
                last_exception = e
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
