"""MiniMax provider implementation.

MiniMax exposes an OpenAI-compatible chat API at api.minimaxi.com.
Best for: agentic workflows, tool use, interleaved thinking.

See https://platform.minimaxi.com/docs/api-reference/text-chat-openai.md
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

MINIMAX_BASE_URL = "https://api.minimaxi.com/v1"


class MiniMaxProvider(BaseProvider):
    """MiniMax provider (OpenAI-compatible)."""

    def __init__(self, api_key: Optional[str] = None, base_url: Optional[str] = None):
        super().__init__(
            name="minimax",
            display_name="MiniMax",
            rate_limit=getattr(settings, "minimax_rate_limit", 100),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = (
            api_key
            or getattr(settings, "minimax_api_key", None)
            or os.environ.get("MINIMAX_API_KEY")
        )
        self.model = getattr(settings, "minimax_model", "MiniMax-M3")
        self.base_url = base_url or getattr(settings, "minimax_base_url", MINIMAX_BASE_URL)
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
            raise RuntimeError("MiniMax provider not available. Set MINIMAX_API_KEY.")
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
            raise RuntimeError("MiniMax provider not available. Set MINIMAX_API_KEY.")
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
        raise NotImplementedError("MiniMax does not provide an embeddings API via this provider.")

    def get_provider_info(self) -> ProviderInfo:
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self.available,
            models=["MiniMax-M3"],
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
        total_tokens = input_tokens + output_tokens
        input_per_m, output_per_m = 1.00, 8.00
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
