"""Groq provider implementation.

Groq uses custom LPU (Language Processing Unit) hardware with consistently
low time-to-first-token (0.6-0.9s). Best for latency-critical paths like
user-facing, real-time function calls.

Free tier includes Llama 4 Scout, Qwen3 32B, and others with 30 RPM.
See https://groq.com
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
    ThinkingConfig,
)

GROQ_BASE_URL = "https://api.groq.com/openai/v1"


class GroqProvider(BaseProvider):
    """Groq provider.

    Best for: latency-critical paths, real-time function calls.
    Key features: LPU hardware, 0.6-0.9s time-to-first-token,
    free tier with 30 RPM for prototyping.
    """

    def __init__(self, api_key: Optional[str] = None):
        super().__init__(
            name="groq",
            display_name="Groq",
            rate_limit=getattr(settings, 'groq_rate_limit', 30),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = (
            api_key or
            getattr(settings, 'groq_api_key', None) or
            os.environ.get("GROQ_API_KEY")
        )
        # Default to Llama 4 Scout for fast inference
        self.model = getattr(settings, 'groq_model', "llama-4-scout-17b-16e-instruct")
        self.base_url = getattr(
            settings, 'groq_base_url', GROQ_BASE_URL
        )
        self._models = [
            # Llama 4 models (fastest)
            "llama-4-scout-17b-16e-instruct",
            "llama-4-maverick-17b-128k-instruct",
            # Llama 3.x models
            "llama-3.3-70b-versatile",
            "llama-3.1-8b-instant",
            "llama-3.1-70b-versatile",
            "llama3-70b-8192",
            "llama3-8b-8192",
            # Qwen models
            "qwen-2.5-32b",
            "qwen-2.5-coder-32b",
            "qwen-qwq-32b",
            # Google Gemma
            "gemma2-9b-it",
            # Mixtral
            "mixtral-8x7b-32768",
            # DeepSeek
            "deepseek-r1-distill-llama-70b",
            "deepseek-r1-distill-qwen-32b",
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
        # Groq does not currently support embeddings
        return False

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
        """Generate a completion via Groq (lowest latency)."""
        if not self.available:
            raise RuntimeError(
                "Groq provider not available. Set GROQ_API_KEY."
            )
        await self.rate_limiter.acquire()
        start_time = time.time()

        # Groq optimized for speed - use lower temperature for function calls
        effective_temp = temperature if temperature < 0.5 else 0.3

        async def _do():
            return await self.client.chat.completions.create(
                model=model or self.model,
                messages=[{"role": m.role.value, "content": m.content} for m in messages],
                temperature=effective_temp,
                max_tokens=max_tokens or 4096,  # Groq default
                top_p=top_p,
                stop=stop,
            )

        response = await self._retry_with_backoff(_do)
        latency_ms = (time.time() - start_time) * 1000
        choice = response.choices[0]
        usage = response.usage
        return CompletionResponse(
            content=choice.message.content or "",
            provider=ProviderType.GROQ,
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
        thinking: Optional[ThinkingConfig] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream completion via Groq (extremely fast streaming)."""
        if not self.available:
            raise RuntimeError(
                "Groq provider not available. Set GROQ_API_KEY."
            )
        await self.rate_limiter.acquire()

        effective_temp = temperature if temperature < 0.5 else 0.3

        stream = await self.client.chat.completions.create(
            model=model or self.model,
            messages=[{"role": m.role.value, "content": m.content} for m in messages],
            temperature=effective_temp,
            max_tokens=max_tokens or 4096,
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
        """Groq does not support embeddings."""
        raise NotImplementedError(
            "Groq does not provide an embeddings API. "
            "Use Fireworks, DeepInfra, or OpenAI for embeddings."
        )

    def get_provider_info(self) -> ProviderInfo:
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self.available,
            models=self._models,
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=0,
            supports_streaming=True,
            supports_embeddings=False,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Estimate cost for Groq.

        Groq pricing (as of 2024):
        - Llama 4 Scout: $0.40/1M input, $0.60/1M output
        - Llama 4 Maverick: $0.40/1M input, $0.60/1M output
        - Llama 3.3 70B: $0.59/1M input, $0.79/1M output
        - Llama 3.1 8B: $0.05/1M input, $0.08/1M output
        - Mixtral 8x7B: $0.24/1M input, $0.24/1M output
        """
        total_tokens = input_tokens + output_tokens

        # Estimate based on model
        if "scout" in model.lower() or "maverick" in model.lower() or "llama-4" in model.lower():
            input_cost_per_m = 0.40
            output_cost_per_m = 0.60
        elif "70b" in model.lower():
            input_cost_per_m = 0.59
            output_cost_per_m = 0.79
        elif "8b" in model.lower():
            input_cost_per_m = 0.05
            output_cost_per_m = 0.08
        elif "mixtral" in model.lower():
            input_cost_per_m = 0.24
            output_cost_per_m = 0.24
        else:
            # Default pricing
            input_cost_per_m = 0.50
            output_cost_per_m = 0.70

        estimated_cost = (
            (input_tokens / 1_000_000 * input_cost_per_m) +
            (output_tokens / 1_000_000 * output_cost_per_m)
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
        """Groq health: client and key present."""
        try:
            self._available = bool(self.client and self.api_key)
            return self._available
        except Exception:
            self._available = False
            return False
