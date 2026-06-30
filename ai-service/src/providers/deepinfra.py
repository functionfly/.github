"""DeepInfra provider implementation.

DeepInfra is best for background and bulk traffic where cost matters more
than peak speed - summarization, batch processing, embeddings.
Serverless pricing can cut costs up to 90% vs provisioned instances.

See https://deepinfra.com
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

DEEPINFRA_BASE_URL = "https://api.deepinfra.com/v1/openai"


class DeepInfraProvider(BaseProvider):
    """DeepInfra provider.

    Best for: background tasks, batch processing, embeddings, summarization.
    Key features: Serverless pricing (up to 90% cost reduction),
    batch processing support, good for non-latency-sensitive work.
    """

    def __init__(self, api_key: Optional[str] = None):
        super().__init__(
            name="deepinfra",
            display_name="DeepInfra",
            rate_limit=getattr(settings, 'deepinfra_rate_limit', 100),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = (
            api_key or
            getattr(settings, 'deepinfra_api_key', None) or
            os.environ.get("DEEPINFRA_API_KEY")
        )
        # Default to cost-effective model for batch work
        self.model = getattr(settings, 'deepinfra_model', "meta-llama/Llama-3.3-70B-Instruct-Turbo")
        self.embedding_model = getattr(
            settings, 'deepinfra_embedding_model', "BAAI/bge-large-en-v1.5"
        )
        self.base_url = getattr(
            settings, 'deepinfra_base_url', DEEPINFRA_BASE_URL
        )
        self._models = [
            # Meta Llama models (cost-effective)
            "meta-llama/Llama-3.3-70B-Instruct-Turbo",
            "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
            "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo",
            "meta-llama/Llama-3.2-11B-Vision-Instruct",
            "meta-llama/Llama-3.2-90B-Vision-Instruct",
            # Mistral models
            "mistralai/Mistral-7B-Instruct-v0.3",
            "mistralai/Mixtral-8x7B-Instruct-v0.1",
            "mistralai/Mixtral-8x22B-Instruct-v0.1",
            "mistralai/Mistral-Nemo-Instruct-2407",
            "mistralai/Mistral-Small-Instruct-2409",
            "mistralai/Pixtral-12B-2409",
            # Qwen models
            "Qwen/Qwen2.5-72B-Instruct",
            "Qwen/Qwen2.5-Coder-32B-Instruct",
            # DeepSeek
            "deepseek-ai/DeepSeek-V3",
            "deepseek-ai/DeepSeek-R1",
            # Microsoft Phi
            "microsoft/Phi-4",
            "microsoft/Phi-3.5-mini-instruct",
            # Google
            "google/gemma-2-27b-it",
            "google/gemma-2-9b-it",
            # Embedding models
            "BAAI/bge-large-en-v1.5",
            "BAAI/bge-base-en-v1.5",
            "sentence-transformers/all-MiniLM-L6-v2",
            "intfloat/multilingual-e5-large-instruct",
            "nomic-ai/nomic-embed-text-v1",
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
        """Generate a completion via DeepInfra."""
        if not self.available:
            raise RuntimeError(
                "DeepInfra provider not available. Set DEEPINFRA_API_KEY."
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
            provider=ProviderType.DEEPINFRA,
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
        """Stream completion via DeepInfra."""
        if not self.available:
            raise RuntimeError(
                "DeepInfra provider not available. Set DEEPINFRA_API_KEY."
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
        """Generate embeddings via DeepInfra."""
        if not self.available:
            raise RuntimeError(
                "DeepInfra provider not available. Set DEEPINFRA_API_KEY."
            )
        await self.rate_limiter.acquire()
        start_time = time.time()

        embed_model = model or self.embedding_model

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
            provider=ProviderType.DEEPINFRA,
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
        """Estimate cost for DeepInfra.

        DeepInfra pricing (serverless, ~90% cheaper than provisioned):
        - Llama 3.3 70B Turbo: $0.30/1M input, $0.50/1M output
        - Llama 3.1 8B: $0.06/1M input, $0.10/1M output
        - Mixtral 8x7B: $0.15/1M input, $0.15/1M output
        - DeepSeek V3: $0.40/1M input, $1.00/1M output
        - Embeddings (bge-large): $0.02/1M tokens
        """
        total_tokens = input_tokens + output_tokens

        # Estimate based on model
        if "70b" in model.lower() or "90b" in model.lower():
            input_cost_per_m = 0.30
            output_cost_per_m = 0.50
        elif "8b" in model.lower() or "7b" in model.lower() or "9b" in model.lower():
            input_cost_per_m = 0.06
            output_cost_per_m = 0.10
        elif "mixtral" in model.lower() or "8x" in model.lower():
            input_cost_per_m = 0.15
            output_cost_per_m = 0.15
        elif "deepseek" in model.lower():
            input_cost_per_m = 0.40
            output_cost_per_m = 1.00
        else:
            # Default pricing
            input_cost_per_m = 0.20
            output_cost_per_m = 0.30

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
        """DeepInfra health: client and key present."""
        try:
            self._available = bool(self.client and self.api_key)
            return self._available
        except Exception:
            self._available = False
            return False
