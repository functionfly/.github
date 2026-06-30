"""OpenAI provider implementation.

This module provides the OpenAI GPT provider (2026-era models).
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


class OpenAIProvider(BaseProvider):
    """OpenAI GPT provider.

    Supports GPT-4o, o1, GPT-4.1, and text-embedding-3 models.
    """

    def __init__(self, api_key: Optional[str] = None):
        super().__init__(
            name="openai",
            display_name="OpenAI",
            rate_limit=settings.openai_rate_limit,
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = api_key or settings.openai_api_key or os.environ.get("OPENAI_API_KEY")
        self.model = settings.openai_model
        self.embedding_model = settings.openai_embedding_model
        self.embedding_dimensions = settings.openai_embedding_dimensions
        self._models = [
            "gpt-4o",
            "gpt-4o-mini",
            "gpt-4.1",
            "gpt-4.1-mini",
            "o1",
            "o1-mini",
            "gpt-4-turbo",
            "gpt-4",
            "gpt-3.5-turbo",
        ]

        # Try to import openai
        try:
            from openai import AsyncOpenAI
            self.client = AsyncOpenAI(
                api_key=self.api_key,
                max_retries=0,  # We handle retries ourselves
            ) if self.api_key else None
        except ImportError:
            self.client = None
            self._available = False

    @property
    def available(self) -> bool:
        """Check if OpenAI is available."""
        return super().available and self.client is not None and self.api_key is not None

    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
    ) -> CompletionResponse:
        """Generate a completion using OpenAI API."""
        if not self.available:
            raise RuntimeError("OpenAI provider not available. Check API key.")

        await self.rate_limiter.acquire()

        start_time = time.time()

        def _do_completion():
            return self.client.chat.completions.create(
                model=model or self.model,
                messages=[{"role": m.role.value, "content": m.content} for m in messages],
                temperature=temperature,
                max_tokens=max_tokens,
                top_p=top_p,
                stop=stop,
            )

        response = await self._retry_with_backoff(_do_completion)

        latency_ms = (time.time() - start_time) * 1000

        return CompletionResponse(
            content=response.choices[0].message.content or "",
            provider=ProviderType.OPENAI,
            model=response.model,
            usage={
                "prompt_tokens": response.usage.prompt_tokens,
                "completion_tokens": response.usage.completion_tokens,
                "total_tokens": response.usage.total_tokens,
            },
            finish_reason=response.choices[0].finish_reason,
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
        """Stream completion using OpenAI API."""
        if not self.available:
            raise RuntimeError("OpenAI provider not available. Check API key.")

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
        """Generate embeddings using OpenAI API."""
        if not self.available:
            raise RuntimeError("OpenAI provider not available. Check API key.")

        await self.rate_limiter.acquire()

        start_time = time.time()

        dims = dimensions or self.embedding_dimensions

        def _do_embedding():
            return self.client.embeddings.create(
                model=model or self.embedding_model,
                input=text,
                dimensions=dims,
            )

        response = await self._retry_with_backoff(_do_embedding)

        latency_ms = (time.time() - start_time) * 1000

        return EmbeddingResponse(
            embedding=response.data[0].embedding,
            provider=ProviderType.OPENAI,
            model=response.model,
            dimensions=len(response.data[0].embedding),
            usage={"tokens": response.usage.total_tokens},
            latency_ms=latency_ms,
        )

    def get_provider_info(self) -> ProviderInfo:
        """Get information about OpenAI provider."""
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self.available,
            models=self._models,
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=self.embedding_dimensions,
            supports_streaming=settings.enable_streaming,
            supports_embeddings=True,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Calculate cost for OpenAI request."""
        # Pricing per 1K tokens (2026-era: o1/gpt-4.1 premium)
        if "o1" in model or "gpt-4.1" in model:
            input_cost = settings.openai_input_cost * 5
            output_cost = settings.openai_output_cost * 5
        elif "gpt-4" in model:
            input_cost = settings.openai_input_cost * 4
            output_cost = settings.openai_output_cost * 4
        else:
            input_cost = settings.openai_input_cost
            output_cost = settings.openai_output_cost

        total_tokens = input_tokens + output_tokens
        estimated_cost = (input_tokens / 1000 * input_cost) + (output_tokens / 1000 * output_cost)

        return CostTracking(
            provider=self.name,
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            total_tokens=total_tokens,
            estimated_cost=estimated_cost,
        )
