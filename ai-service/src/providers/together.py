"""Together AI provider implementation.

Together recently achieved up to 2x faster serverless inference vs competitors
on models like GPT-OSS and DeepSeek, and offers batch processing at up to
50% less cost. Good 200+ model catalog.

See https://together.ai
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

TOGETHER_BASE_URL = "https://api.together.xyz/v1"


class TogetherProvider(BaseProvider):
    """Together AI provider.

    Best for: batch processing, cost-effective inference, wide model catalog.
    Key features: Up to 2x faster serverless inference, batch processing
    at 50% less cost, 200+ model catalog.
    """

    def __init__(self):
        super().__init__(
            name="together",
            display_name="Together AI",
            rate_limit=getattr(settings, 'together_rate_limit', 60),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
        )
        self.api_key = (
            getattr(settings, 'together_api_key', None) or
            os.environ.get("TOGETHER_API_KEY")
        )
        # Default to fast model for general use
        self.model = getattr(settings, 'together_model', "meta-llama/Llama-3.3-70B-Instruct-Turbo")
        self.base_url = getattr(
            settings, 'together_base_url', TOGETHER_BASE_URL
        )
        self._models = [
            # Meta Llama models
            "meta-llama/Llama-3.3-70B-Instruct-Turbo",
            "meta-llama/Llama-3.2-11B-Vision-Instruct",
            "meta-llama/Llama-3.2-90B-Vision-Instruct",
            "meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo",
            "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
            "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo",
            "meta-llama/Llama-3-70b-hf",
            "meta-llama/Llama-3-8b-hf",
            # DeepSeek
            "deepseek-ai/DeepSeek-V3",
            "deepseek-ai/DeepSeek-R1",
            "deepseek-ai/DeepSeek-R1-Distill-Llama-70B",
            # Qwen
            "Qwen/Qwen2.5-72B-Instruct-Turbo",
            "Qwen/Qwen2.5-Coder-32B-Instruct",
            # Mistral
            "mistralai/Mixtral-8x7B-Instruct-v0.1",
            "mistralai/Mistral-7B-Instruct-v0.3",
            "mistralai/Mistral-Small-24B-Instruct-2501",
            # Google Gemma
            "google/gemma-2-27b-it",
            "google/gemma-2-9b-it",
            # NVIDIA
            "nvidia/Llama-3.1-Nemotron-70B-Instruct-HF",
            # Microsoft Phi
            "microsoft/phi-4",
            "microsoft/Phi-3-medium-128k-instruct",
            # Embedding models
            "togethercomputer/m2-bert-80M-2k-retrieval",
            "togethercomputer/m2-bert-80M-8k-retrieval",
            "BAAI/bge-large-en-v1.5",
            "BAAI/bge-base-en-v1.5",
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
        """Generate a completion via Together AI."""
        if not self.available:
            raise RuntimeError(
                "Together AI provider not available. Set TOGETHER_API_KEY."
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
            provider=ProviderType.TOGETHER,
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
        """Stream completion via Together AI."""
        if not self.available:
            raise RuntimeError(
                "Together AI provider not available. Set TOGETHER_API_KEY."
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
        """Generate embeddings via Together AI."""
        if not self.available:
            raise RuntimeError(
                "Together AI provider not available. Set TOGETHER_API_KEY."
            )
        await self.rate_limiter.acquire()
        start_time = time.time()

        # Together embedding models
        embed_model = model or "BAAI/bge-large-en-v1.5"

        async def _do():
            return await self.client.embeddings.create(
                model=embed_model,
                input=text,
            )

        response = await self._retry_with_backoff(_do)
        latency_ms = (time.time() - start_time) * 1000
        embedding_data = response.data[0].embedding

        # Handle dimensions
        if dimensions and len(embedding_data) != dimensions:
            if len(embedding_data) > dimensions:
                embedding_data = embedding_data[:dimensions]
            else:
                embedding_data = embedding_data + [0.0] * (dimensions - len(embedding_data))

        return EmbeddingResponse(
            embedding=embedding_data,
            provider=ProviderType.TOGETHER,
            model=response.model,
            dimensions=len(embedding_data),
            usage={"tokens": response.usage.total_tokens if response.usage else 0},
            latency_ms=latency_ms,
        )

    def get_provider_info(self) -> ProviderInfo:
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self.available,
            models=self._models,
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=1024,  # bge-large-en-v1.5
            supports_streaming=True,
            supports_embeddings=True,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Estimate cost for Together AI.

        Together pricing (as of 2024):
        - Llama 3.3 70B: $0.88/1M input, $0.88/1M output
        - Llama 3.1 8B: $0.18/1M input, $0.18/1M output
        - Llama 3.1 405B: $5.00/1M input, $5.00/1M output
        - DeepSeek V3: $1.00/1M input, $2.50/1M output
        - Mixtral 8x7B: $0.60/1M input, $0.60/1M output
        - Embeddings: $0.02/1M tokens
        """
        total_tokens = input_tokens + output_tokens

        # Estimate based on model
        if "405b" in model.lower():
            input_cost_per_m = 5.00
            output_cost_per_m = 5.00
        elif "70b" in model.lower() or "90b" in model.lower():
            input_cost_per_m = 0.88
            output_cost_per_m = 0.88
        elif "8b" in model.lower() or "7b" in model.lower():
            input_cost_per_m = 0.18
            output_cost_per_m = 0.18
        elif "mixtral" in model.lower():
            input_cost_per_m = 0.60
            output_cost_per_m = 0.60
        elif "deepseek" in model.lower():
            input_cost_per_m = 1.00
            output_cost_per_m = 2.50
        else:
            # Default pricing
            input_cost_per_m = 0.50
            output_cost_per_m = 0.50

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
        """Together AI health: client and key present."""
        try:
            self._available = bool(self.client and self.api_key)
            return self._available
        except Exception:
            self._available = False
            return False
