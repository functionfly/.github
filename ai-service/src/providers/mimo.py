"""Xiaomi MiMo provider implementation.

MiMo exposes an OpenAI-compatible chat API at api.xiaomimimo.com.
Best for: long-context reasoning, agent workflows, cost-efficient flash inference.

See https://platform.xiaomimimo.com/docs/en-US/api/chat/openai-api
"""

import os
from typing import Optional, AsyncGenerator
import time

from .base import BaseProvider, RetryConfig
from .model_registry import model_ids_for_provider
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

MIMO_BASE_URL = "https://api.xiaomimimo.com/v1"


class MiMoProvider(BaseProvider):
    """Xiaomi MiMo provider (OpenAI-compatible)."""

    def __init__(self, api_key: Optional[str] = None, base_url: Optional[str] = None):
        super().__init__(
            name="mimo",
            display_name="Xiaomi MiMo",
            rate_limit=getattr(settings, "mimo_rate_limit", 100),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = (
            api_key
            or getattr(settings, "mimo_api_key", None)
            or os.environ.get("MIMO_API_KEY")
            or os.environ.get("XIAOMIMIMO_API_KEY")
        )
        self.model = getattr(settings, "mimo_model", "mimo-v2-flash")
        self.base_url = base_url or getattr(settings, "mimo_base_url", MIMO_BASE_URL)
        self._models = model_ids_for_provider("mimo")
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
        if not self.available:
            raise RuntimeError("MiMo provider not available. Set MIMO_API_KEY.")
        await self.rate_limiter.acquire()
        start_time = time.time()

        async def _do():
            return await self.client.chat.completions.create(
                model=model or self.model,
                messages=[{"role": m.role.value, "content": m.content} for m in messages],
                temperature=temperature,
                max_tokens=max_tokens or 4096,
                top_p=top_p,
                stop=stop,
            )

        response = await self._retry_with_backoff(_do)
        latency_ms = (time.time() - start_time) * 1000
        choice = response.choices[0]
        usage = response.usage
        return CompletionResponse(
            content=choice.message.content or "",
            provider=ProviderType.MIMO,
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
        if not self.available:
            raise RuntimeError("MiMo provider not available. Set MIMO_API_KEY.")
        await self.rate_limiter.acquire()

        stream = await self.client.chat.completions.create(
            model=model or self.model,
            messages=[{"role": m.role.value, "content": m.content} for m in messages],
            temperature=temperature,
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
        raise NotImplementedError("MiMo does not provide an embeddings API.")

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
        """Estimate cost from published MiMo API pricing (per 1M tokens)."""
        total_tokens = input_tokens + output_tokens
        model_lower = model.lower()

        if "flash" in model_lower:
            input_per_m, output_per_m = 0.10, 0.30
        elif "2.5-pro" in model_lower or model_lower.endswith("v2-pro"):
            input_per_m, output_per_m = 1.00, 3.00
        elif "omni" in model_lower or "2.5" in model_lower:
            input_per_m, output_per_m = 0.40, 0.80
        else:
            input_per_m, output_per_m = 0.50, 1.50

        estimated_cost = (input_tokens / 1_000_000 * input_per_m) + (
            output_tokens / 1_000_000 * output_per_m
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
        try:
            self._available = bool(self.client and self.api_key)
            return self._available
        except Exception:
            self._available = False
            return False
