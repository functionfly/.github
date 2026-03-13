"""Anthropic provider implementation.

This module provides the Anthropic Claude provider (2026-era models).
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


class AnthropicProvider(BaseProvider):
    """Anthropic Claude provider.

    Supports Claude Sonnet 4.6, Claude 3.5 Sonnet/Haiku, and Claude 3 Opus.
    """

    def __init__(self):
        super().__init__(
            name="anthropic",
            display_name="Anthropic",
            rate_limit=settings.anthropic_rate_limit,
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
        )
        self.api_key = settings.anthropic_api_key or os.environ.get("ANTHROPIC_API_KEY")
        self.model = settings.anthropic_model
        self.max_tokens = settings.anthropic_max_tokens
        self._models = [
            "claude-sonnet-4-6",
            "claude-3-5-sonnet-20241022",
            "claude-3-5-haiku-20241022",
            "claude-3-5-sonnet-20240620",
            "claude-3-opus-20240229",
            "claude-3-sonnet-20240229",
            "claude-3-haiku-20240307",
        ]

        # Try to import anthropic
        try:
            from anthropic import AsyncAnthropic
            self.client = AsyncAnthropic(
                api_key=self.api_key,
                max_retries=0,
            ) if self.api_key else None
        except ImportError:
            self.client = None
            self._available = False

    @property
    def available(self) -> bool:
        """Check if Anthropic is available."""
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
        """Generate a completion using Anthropic API."""
        if not self.available:
            raise RuntimeError("Anthropic provider not available. Check API key.")

        await self.rate_limiter.acquire()

        start_time = time.time()

        # Convert messages to Anthropic format
        # System messages need special handling - combine into first user message if present
        system_prompt = None
        anthropic_messages = []

        for msg in messages:
            if msg.role.value == "system":
                system_prompt = msg.content
            else:
                anthropic_messages.append({
                    "role": msg.role.value,
                    "content": msg.content
                })

        def _do_completion():
            return self.client.messages.create(
                model=model or self.model,
                system=system_prompt,
                messages=anthropic_messages,
                temperature=temperature,
                max_tokens=max_tokens or self.max_tokens,
                top_p=top_p,
                stop_sequences=stop,
            )

        response = await self._retry_with_backoff(_do_completion)

        latency_ms = (time.time() - start_time) * 1000

        # Extract content from response
        content = ""
        if response.content:
            content = response.content[0].text if hasattr(response.content[0], 'text') else str(response.content[0])

        return CompletionResponse(
            content=content,
            provider=ProviderType.ANTHROPIC,
            model=response.model,
            usage={
                "prompt_tokens": response.usage.input_tokens,
                "completion_tokens": response.usage.output_tokens,
                "total_tokens": response.usage.input_tokens + response.usage.output_tokens,
            },
            finish_reason=response.stop_reason,
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
        """Stream completion using Anthropic API."""
        if not self.available:
            raise RuntimeError("Anthropic provider not available. Check API key.")

        await self.rate_limiter.acquire()

        # Convert messages to Anthropic format
        system_prompt = None
        anthropic_messages = []

        for msg in messages:
            if msg.role.value == "system":
                system_prompt = msg.content
            else:
                anthropic_messages.append({
                    "role": msg.role.value,
                    "content": msg.content
                })

        async with self.client.messages.stream(
            model=model or self.model,
            system=system_prompt,
            messages=anthropic_messages,
            temperature=temperature,
            max_tokens=max_tokens or self.max_tokens,
            top_p=top_p,
            stop_sequences=stop,
        ) as stream:
            async for text in stream.text_stream:
                yield text

    async def embed(
        self,
        text: str,
        model: Optional[str] = None,
        dimensions: Optional[int] = None,
    ) -> EmbeddingResponse:
        """Generate embeddings using Anthropic API.

        Note: Anthropic doesn't have a native embeddings API.
        This falls back to using the OpenAI provider for embeddings.
        """
        # For now, raise an error as Anthropic doesn't have embeddings
        raise NotImplementedError(
            "Anthropic does not have a native embeddings API. "
            "Use OpenAI or Ollama for embeddings."
        )

    def get_provider_info(self) -> ProviderInfo:
        """Get information about Anthropic provider."""
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self.available,
            models=self._models,
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=0,  # Not supported
            supports_streaming=settings.enable_streaming,
            supports_embeddings=False,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Calculate cost for Anthropic request."""
        # Pricing per 1K tokens (Sonnet 4.6 ~$3/M in, $15/M out)
        if "opus" in model:
            input_cost = settings.anthropic_input_cost * 5
            output_cost = settings.anthropic_output_cost * 5
        elif "sonnet-4" in model or "sonnet-4-6" in model:
            input_cost = 0.003  # $3/M
            output_cost = 0.015  # $15/M
        elif "sonnet" in model:
            input_cost = settings.anthropic_input_cost
            output_cost = settings.anthropic_output_cost
        else:  # haiku
            input_cost = settings.anthropic_input_cost / 10
            output_cost = settings.anthropic_output_cost / 10

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
