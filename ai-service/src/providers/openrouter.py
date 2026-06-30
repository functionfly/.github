"""OpenRouter provider implementation.

OpenRouter provides a unified API to many LLM backends (OpenAI, Anthropic,
Google, etc.). Uses OpenAI-compatible request/response format.
See https://openrouter.ai/docs
"""

import os
import time
from typing import AsyncGenerator, Optional

from ..config import settings
from ..models.schemas import (
    ChatMessage,
    CompletionResponse,
    CostTracking,
    EmbeddingResponse,
    ProviderInfo,
    ProviderType,
)
from .base import BaseProvider, RetryConfig

OPENROUTER_BASE_URL = "https://openrouter.ai/api/v1"


# Free models on OpenRouter (models that don't cost credits)
FREE_MODELS = frozenset(
    [
        "nvidia/nemotron-3-super-120b-a12b:free",
        "nvidia/nemotron-3-ultra-550b-a55b:free",
        "nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free",
        "inclusionai/ling-2.6-flash:free",
        "inclusionai/ling-2.6-1t:free",
        "poolside/laguna-m.1:free",
        "poolside/laguna-xs.2:free",
    ]
)


def is_free_model(model: str) -> bool:
    """Check if a model is a free model."""
    return model in FREE_MODELS


def get_default_free_model() -> str:
    """Get the default free model (first available free model)."""
    return "poolside/laguna-xs.2:free"


class OpenRouterProvider(BaseProvider):
    """OpenRouter provider.

    Routes to multiple backends via OpenRouter (openai/gpt-4o,
    anthropic/claude-3.5-sonnet, google/gemini-pro, etc.).
    Chat and stream only; no native embeddings (use OpenAI/other for embeddings).
    """

    def __init__(self, api_key: Optional[str] = None):
        super().__init__(
            name="openrouter",
            display_name="OpenRouter",
            rate_limit=settings.openrouter_rate_limit,
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = api_key or settings.openrouter_api_key or os.environ.get("OPENROUTER_API_KEY")
        self.model = settings.openrouter_model
        self.base_url = getattr(settings, "openrouter_base_url", OPENROUTER_BASE_URL)
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
            "openrouter/owl-alpha",
            "nvidia/nemotron-3-super-120b-a12b:free",
            "nvidia/nemotron-3-ultra-550b-a55b:free",
            "nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free",
            "inception/mercury-2",
            "inclusionai/ling-2.6-flash:free",
            "inclusionai/ling-2.6-1t:free",
            "minimax/minimax-m3",
            "stepfun/step-3.7-flash",
            "x-ai/grok-build-0.1",
            "poolside/laguna-m.1:free",
            "poolside/laguna-xs.2:free",
        ]
        self._available = True
        try:
            from openai import AsyncOpenAI

            self.client = (
                AsyncOpenAI(
                    api_key=self.api_key,
                    base_url=self.base_url,
                    max_retries=0,
                )
                if self.api_key
                else None
            )
        except ImportError:
            self.client = None
            self._available = False

    @property
    def available(self) -> bool:
        return super().available and self.client is not None and self.api_key is not None

    @property
    def supports_embeddings(self) -> bool:
        return False

    async def _get_user_preference_model(
        self, user_id: Optional[str], tenant_id: Optional[str]
    ) -> Optional[str]:
        """Get user's preferred model from Redis preferences store.


        Checks if the user has set an explicit model preference.
        Returns None if no preference is set.
        """
        if not user_id and not tenant_id:
            return None

        try:
            import redis.asyncio as redis

            r = redis.from_url(settings.redis_url, encoding="utf-8", decode_responses=True)
            key = f"ai:preference:{tenant_id or 'default'}:{user_id or 'default'}:model"
            pref = await r.get(key)
            await r.aclose()
            return pref if pref else None
        except Exception:
            return None

    async def _get_model_for_request(
        self, model: Optional[str], user_id: Optional[str], tenant_id: Optional[str]
    ) -> str:
        """Determine which model to use for a request.


        Logic:
        1. If model is explicitly specified, use it
        2. If user has set a non-free model preference, use that
        3. Otherwise, use a free model (cost-saving default)
        """
        if model:
            return model

        # Check user preference
        user_pref = await self._get_user_preference_model(user_id, tenant_id)
        if user_pref:
            return user_pref

        # Default to free model
        return get_default_free_model()

    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
        user_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
    ) -> CompletionResponse:
        """Generate a completion via OpenRouter (OpenAI-compatible API).


        Args:
            messages: Chat messages
            model: Optional model override
            temperature: Sampling temperature
            max_tokens: Max tokens to generate
            top_p: Top-p sampling parameter
            stop: Stop sequences
            user_id: User ID for preference lookup (prefers free models by default)
            tenant_id: Tenant ID for preference lookup
        """
        if not self.available:
            raise RuntimeError("OpenRouter provider not available. Set OPENROUTER_API_KEY.")
        await self.rate_limiter.acquire()
        start_time = time.time()

        # Resolve model: explicit preference > user preference > free model
        resolved_model = await self._get_model_for_request(model, user_id, tenant_id)

        async def _do():
            return await self.client.chat.completions.create(
                model=resolved_model,
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
        user_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream completion via OpenRouter.


        Args:
            messages: Chat messages
            model: Optional model override
            temperature: Sampling temperature
            max_tokens: Max tokens to generate
            top_p: Top-p sampling parameter
            stop: Stop sequences
            user_id: User ID for preference lookup (prefers free models by default)
            tenant_id: Tenant ID for preference lookup
        """
        if not self.available:
            raise RuntimeError("OpenRouter provider not available. Set OPENROUTER_API_KEY.")
        await self.rate_limiter.acquire()

        # Resolve model: explicit preference > user preference > free model
        resolved_model = await self._get_model_for_request(model, user_id, tenant_id)

        stream = await self.client.chat.completions.create(
            model=resolved_model,
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
        estimated_cost = (input_tokens / 1_000_000 * 2.5) + (output_tokens / 1_000_000 * 10.0)
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
