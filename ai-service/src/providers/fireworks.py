"""Fireworks AI provider implementation.

Fireworks AI delivers 4x lower latency than vLLM with their proprietary
FireAttention engine. Optimized for structured output, function calling,
and JSON mode - perfect for a callable function marketplace.

SOC 2 + HIPAA compliant, 15+ global regions.
See https://fireworks.ai

SDK Options:
- Native SDK (recommended): `pip install --pre fireworks-ai`
  Benefits: Better concurrency defaults, Fireworks-exclusive features,
  platform automation for datasets/evals/fine-tuning
- OpenAI SDK: OpenAI-compatible endpoint (fallback)
  Benefits: Consistent interface with other providers
"""

import os
from typing import Optional, AsyncGenerator, Any
import time
import logging

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

logger = logging.getLogger(__name__)

FIREWORKS_BASE_URL = "https://api.fireworks.ai/inference/v1"


class FireworksProvider(BaseProvider):
    """Fireworks AI provider.

    Best for: structured output, function calling, JSON mode.
    Key features: FireAttention engine, 4x lower latency than vLLM,
    SOC 2 + HIPAA compliant, 15+ global regions.

    Uses native fireworks-ai SDK when available for optimal performance,
    falls back to OpenAI-compatible API otherwise.
    """

    def __init__(self, api_key: Optional[str] = None):
        super().__init__(
            name="fireworks",
            display_name="Fireworks AI",
            rate_limit=getattr(settings, 'fireworks_rate_limit', 120),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = (
            api_key or
            getattr(settings, 'fireworks_api_key', None) or
            os.environ.get("FIREWORKS_API_KEY")
        )
        self.model = getattr(settings, 'fireworks_model', "accounts/fireworks/models/llama-v3p1-405b-instruct")
        self.base_url = getattr(
            settings, 'fireworks_base_url', FIREWORKS_BASE_URL
        )
        self._models = [
            # Llama models (optimized for function calling)
            "accounts/fireworks/models/llama-v3p1-405b-instruct",
            "accounts/fireworks/models/llama-v3p1-70b-instruct",
            "accounts/fireworks/models/llama-v3p1-8b-instruct",
            "accounts/fireworks/models/llama-v3p3-70b-instruct",
            # Qwen models
            "accounts/fireworks/models/qwen2p5-72b-instruct",
            "accounts/fireworks/models/qwen2p5-coder-32b-instruct",
            # Mixtral
            "accounts/fireworks/models/mixtral-8x22b-instruct",
            "accounts/fireworks/models/mixtral-8x7b-instruct",
            # DeepSeek
            "accounts/fireworks/models/deepseek-v3",
            "accounts/fireworks/models/deepseek-r1",
        ]
        self._available = False
        self._native_client: Optional[Any] = None
        self._openai_client: Optional[Any] = None
        self._use_native = False

        if not self.api_key:
            logger.warning("Fireworks AI: No API key configured")
            return

        # Try native fireworks-ai SDK first (recommended)
        try:
            from fireworks.client import Fireworks
            self._native_client = Fireworks(
                api_key=self.api_key,
                # Optimized defaults for high-throughput
                max_retries=0,  # We handle retries
                timeout=120,    # 2 minute timeout
            )
            self._use_native = True
            self._available = True
            logger.info("Fireworks AI: Using native SDK (optimal performance)")
        except ImportError:
            logger.debug("fireworks-ai SDK not installed, falling back to OpenAI SDK")
            self._use_native = False

        # Fallback to OpenAI SDK if native not available
        if not self._use_native:
            try:
                from openai import AsyncOpenAI
                self._openai_client = AsyncOpenAI(
                    api_key=self.api_key,
                    base_url=self.base_url,
                    max_retries=0,
                    # Optimized connection pooling for high-throughput
                    http_client=None,  # Use default with connection pooling
                )
                self._available = True
                logger.info("Fireworks AI: Using OpenAI SDK (OpenAI-compatible mode)")
            except ImportError:
                logger.warning("Neither fireworks-ai nor openai SDK available")
                self._available = False

    @property
    def available(self) -> bool:
        return (
            super().available
            and self._available
            and self.api_key is not None
            and (self._native_client is not None or self._openai_client is not None)
        )

    @property
    def use_native_sdk(self) -> bool:
        """Whether using the native fireworks-ai SDK."""
        return self._use_native

    @property
    def supports_embeddings(self) -> bool:
        return True

    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
    ) -> CompletionResponse:
        """Generate a completion via Fireworks AI."""
        if not self.available:
            raise RuntimeError(
                "Fireworks AI provider not available. Set FIREWORKS_API_KEY."
            )
        await self.rate_limiter.acquire()
        start_time = time.time()

        model_to_use = model or self.model
        msgs = [{"role": m.role.value, "content": m.content} for m in messages]

        # Detect function calling / structured output
        extra_body = {}
        if any(m.role.value == "function" or "function_call" in str(m.content) for m in messages):
            extra_body["response_format"] = {"type": "json_object"}

        if self._use_native and self._native_client:
            response = await self._complete_native(
                msgs, model_to_use, temperature, max_tokens, top_p, stop, extra_body
            )
        else:
            response = await self._complete_openai(
                msgs, model_to_use, temperature, max_tokens, top_p, stop, extra_body
            )

        latency_ms = (time.time() - start_time) * 1000
        return CompletionResponse(
            content=response["content"],
            provider=ProviderType.FIREWORKS,
            model=response["model"],
            usage=response["usage"],
            finish_reason=response["finish_reason"],
            latency_ms=latency_ms,
        )

    async def _complete_native(
        self,
        messages: list[dict],
        model: str,
        temperature: float,
        max_tokens: Optional[int],
        top_p: Optional[float],
        stop: Optional[list[str]],
        extra_body: dict,
    ) -> dict:
        """Complete using native fireworks-ai SDK."""
        from fireworks.client import CompletionResponse as FireworksResponse

        async def _do():
            # Native SDK sync call (wrapped in async)
            import asyncio
            loop = asyncio.get_event_loop()

            def _sync_complete():
                return self._native_client.chat.completions.create(
                    model=model,
                    messages=messages,
                    temperature=temperature,
                    max_tokens=max_tokens,
                    top_p=top_p,
                    stop=stop,
                    **extra_body,
                )

            resp = await loop.run_in_executor(None, _sync_complete)
            return {
                "content": resp.choices[0].message.content or "",
                "model": resp.model,
                "usage": {
                    "prompt_tokens": resp.usage.prompt_tokens,
                    "completion_tokens": resp.usage.completion_tokens,
                    "total_tokens": resp.usage.total_tokens,
                },
                "finish_reason": resp.choices[0].finish_reason,
            }

        return await self._retry_with_backoff(_do)

    async def _complete_openai(
        self,
        messages: list[dict],
        model: str,
        temperature: float,
        max_tokens: Optional[int],
        top_p: Optional[float],
        stop: Optional[list[str]],
        extra_body: dict,
    ) -> dict:
        """Complete using OpenAI-compatible API."""
        async def _do():
            resp = await self._openai_client.chat.completions.create(
                model=model,
                messages=messages,
                temperature=temperature,
                max_tokens=max_tokens,
                top_p=top_p,
                stop=stop,
                **extra_body,
            )
            return {
                "content": resp.choices[0].message.content or "",
                "model": resp.model,
                "usage": {
                    "prompt_tokens": resp.usage.prompt_tokens,
                    "completion_tokens": resp.usage.completion_tokens,
                    "total_tokens": resp.usage.total_tokens,
                },
                "finish_reason": resp.choices[0].finish_reason,
            }

        return await self._retry_with_backoff(_do)

    async def stream(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream completion via Fireworks AI."""
        if not self.available:
            raise RuntimeError(
                "Fireworks AI provider not available. Set FIREWORKS_API_KEY."
            )
        await self.rate_limiter.acquire()

        msgs = [{"role": m.role.value, "content": m.content} for m in messages]
        model_to_use = model or self.model

        if self._use_native and self._native_client:
            # Native streaming
            import asyncio
            loop = asyncio.get_event_loop()

            def _sync_stream():
                return self._native_client.chat.completions.create(
                    model=model_to_use,
                    messages=msgs,
                    temperature=temperature,
                    max_tokens=max_tokens,
                    top_p=top_p,
                    stop=stop,
                    stream=True,
                )

            stream_resp = await loop.run_in_executor(None, _sync_stream)
            for chunk in stream_resp:
                if chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content
        else:
            # OpenAI streaming
            stream = await self._openai_client.chat.completions.create(
                model=model_to_use,
                messages=msgs,
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
        """Generate embeddings via Fireworks AI."""
        if not self.available:
            raise RuntimeError(
                "Fireworks AI provider not available. Set FIREWORKS_API_KEY."
            )
        await self.rate_limiter.acquire()
        start_time = time.time()

        embed_model = model or "nomic-ai/nomic-embed-text-v1.5"

        if self._use_native and self._native_client:
            # Native embedding
            import asyncio
            loop = asyncio.get_event_loop()

            def _sync_embed():
                return self._native_client.embeddings.create(
                    model=embed_model,
                    input=text,
                )

            response = await loop.run_in_executor(None, _sync_embed)
            embedding_data = response.data[0].embedding
        else:
            # OpenAI embedding
            async def _do():
                return await self._openai_client.embeddings.create(
                    model=embed_model,
                    input=text,
                )
            response = await self._retry_with_backoff(_do)
            embedding_data = response.data[0].embedding

        latency_ms = (time.time() - start_time) * 1000

        # Handle dimensions
        if dimensions and len(embedding_data) != dimensions:
            if len(embedding_data) > dimensions:
                embedding_data = embedding_data[:dimensions]
            else:
                embedding_data = embedding_data + [0.0] * (dimensions - len(embedding_data))

        # Get token count from response
        tokens = 0
        if hasattr(response, 'usage') and response.usage:
            tokens = response.usage.total_tokens

        return EmbeddingResponse(
            embedding=embedding_data,
            provider=ProviderType.FIREWORKS,
            model=embed_model,
            dimensions=len(embedding_data),
            usage={"tokens": tokens},
            latency_ms=latency_ms,
        )

    def get_provider_info(self) -> ProviderInfo:
        sdk_mode = "native (fireworks-ai)" if self._use_native else "openai-compatible"
        return ProviderInfo(
            name=self.name,
            display_name=f"{self.display_name} ({sdk_mode})",
            available=self.available,
            models=self._models,
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=768,  # nomic-embed-text-v1.5
            supports_streaming=True,
            supports_embeddings=True,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Estimate cost for Fireworks AI.

        Pricing varies by model. Rough estimates:
        - Llama 3.1 405B: $3/1M input, $3/1M output
        - Llama 3.1 70B: $0.9/1M input, $0.9/1M output
        - Llama 3.1 8B: $0.2/1M input, $0.2/1M output
        """
        total_tokens = input_tokens + output_tokens

        # Estimate based on model tier
        if "405b" in model.lower():
            input_cost_per_m = 3.0
            output_cost_per_m = 3.0
        elif "70b" in model.lower():
            input_cost_per_m = 0.9
            output_cost_per_m = 0.9
        elif "8b" in model.lower() or "32b" in model.lower():
            input_cost_per_m = 0.2
            output_cost_per_m = 0.2
        else:
            # Default to mid-tier pricing
            input_cost_per_m = 0.5
            output_cost_per_m = 0.5

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
        """Fireworks health check."""
        try:
            has_client = (self._native_client is not None) or (self._openai_client is not None)
            self._available = bool(has_client and self.api_key)
            return self._available
        except Exception:
            self._available = False
            return False
