"""OpenRouter provider implementation.

OpenRouter provides a unified API to many LLM backends (OpenAI, Anthropic,
Google, etc.). Uses OpenAI-compatible request/response format.
See https://openrouter.ai/docs
"""

import os
from typing import Optional, AsyncGenerator
import time

from .base import BaseProvider, RetryConfig
from ..config import settings
from ..models.schemas import (
    ChatMessage,
    CompletionResponse,
    EmbeddingResponse,
    ProviderInfo,
    ProviderType,
    CostTracking,
)

OPENROUTER_BASE_URL = "https://openrouter.ai/api/v1"


class OpenRouterProvider(BaseProvider):
    """OpenRouter provider.

    Routes to multiple backends via OpenRouter (openai/gpt-4o,
    anthropic/claude-3.5-sonnet, google/gemini-pro, etc.).
    Chat and stream only; no native embeddings (use OpenAI/other for embeddings).
    """

    def __init__(self):
        super().__init__(
            name="openrouter",
            display_name="OpenRouter",
            rate_limit=settings.openrouter_rate_limit,
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
        )
        self.api_key = (
            settings.openrouter_api_key or os.environ.get("OPENROUTER_API_KEY")
        )
        self.model = settings.openrouter_model
        self.base_url = getattr(
            settings, "openrouter_base_url", OPENROUTER_BASE_URL
        )
        self._models = [
            "anthropic/claude-sonnet-4-6",
            "openai/gpt-4o",
            "openai/gpt-4o-mini",
            "openai/gpt-4-turbo",
            "anthropic/claude-3.5-sonnet",
            "anthropic/claude-3.5-haiku",
            "anthropic/claude-3-opus",
            "google/gemini-2.0-flash-exp",
            "google/gemini-pro",
            "meta-llama/llama-3.3-70b-instruct",
            "mistralai/mistral-large",
            "openrouter/hunter-alpha",
            "nvidia/nemotron-3-super-120b-a12b:free",
            "inception/mercury-2",


        ]
        self._available = True
        try:
            from openai import AsyncOpenAI
            self.client = AsyncOpenAI(
                api_key=self.api_key,
                base_url=self.base_url,
                max_retries=0,
            ) if self.api_key else None
        except ImportError:
            self.client = None
            self._available = False

    @property
    def available(self) -> bool:
        return (
            super().available
            and self.client is not None
            and self.api_key is not None
        )

    @property
    def supports_embeddings(self) -> bool:
        return False

    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
    ) -> CompletionResponse:
        """Generate a completion via OpenRouter (OpenAI-compatible API)."""
        if not self.available:
            raise RuntimeError(
                "OpenRouter provider not available. Set OPENROUTER_API_KEY."
            )
        await self.rate_limiter.acquire()
        start_time = time.time()
        async def _do():
            return await self.client.chat.completions.create(
                model=model or self.model,
                messages=[{"role": m.role.value, "content": m.content} for m in messages],
                temperature=temperature,
                max_tokens=max_tokens,
                top_p=top_p,
                stop=stop,
            )

        response = await self._retry_with_backoff(_do)
        latency_ms = (time.time() - start_time) * 1000
        choice = response.choices[0]
        usage = response.usage
        return CompletionResponse(
            content=choice.message.content or "",
            provider=ProviderType.OPENROUTER,
            model=response.model,
            usage={
                "prompt_tokens": usage.prompt_tokens,
                "completion_tokens": usage.completion_tokens,
                "total_tokens": usage.total_tokens,
            },
            finish_reason=choice.finish_reason,
            latency_ms=latency_ms,
        )

    async def stream(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream completion via OpenRouter."""
        if not self.available:
            raise RuntimeError(
                "OpenRouter provider not available. Set OPENROUTER_API_KEY."
            )
        await self.rate_limiter.acquire()
        stream = await self.client.chat.completions.create(
            model=model or self.model,
            messages=[{"role": m.role.value, "content": m.content} for m in messages],
            temperature=temperature,
            max_tokens=max_tokens,
            top_p=top_p,
            stop=stop,
            stream=True,
        )
        async for chunk in stream:
            if chunk.choices[0].delta.content:
                yield chunk.choices[0].delta.content

    async def embed(
        self,
        text: str,
        model: Optional[str] = None,
        dimensions: Optional[int] = None,
    ) -> EmbeddingResponse:
        """OpenRouter does not expose a unified embeddings API."""
        raise NotImplementedError(
            "OpenRouter does not provide a unified embeddings API. "
            "Use OpenAI or Ollama for embeddings."
        )

    def get_provider_info(self) -> ProviderInfo:
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self.available,
            models=self._models,
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=0,
            supports_streaming=settings.enable_streaming,
            supports_embeddings=False,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Estimate cost (OpenRouter pricing varies by backend)."""
        total_tokens = input_tokens + output_tokens
        # Rough average; real cost depends on OpenRouter's model pricing
        estimated_cost = (input_tokens / 1_000_000 * 2.5) + (
            output_tokens / 1_000_000 * 10.0
        )
        return CostTracking(
            provider=self.name,
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            total_tokens=total_tokens,
            estimated_cost=estimated_cost,
        )

    async def health_check(self) -> bool:
        """OpenRouter health: client and key present (no embed)."""
        try:
            self._available = bool(self.client and self.api_key)
            return self._available
        except Exception:
            self._available = False
            return False
