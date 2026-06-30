"""StepFun AI provider implementation.

StepFun provides OpenAI-compatible APIs for Step series models (step-1/2/3).
Best for: reasoning, coding, Chinese/English bilingual tasks.

API base: https://api.stepfun.ai/v1  (OpenAI-compatible)
See https://platform.stepfun.ai/
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
)

STEPFUN_BASE_URL = "https://api.stepfun.ai/v1"


class StepFunProvider(BaseProvider):
    """StepFun AI provider (OpenAI-compatible)."""

    def __init__(self, api_key: Optional[str] = None):
        super().__init__(
            name="stepfun",
            display_name="StepFun",
            rate_limit=getattr(settings, "stepfun_rate_limit", 60),
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
            api_key=api_key,
        )
        self.api_key = (
            api_key
            or getattr(settings, "stepfun_api_key", None)
            or os.environ.get("STEPFUN_API_KEY")
        )
        self.model = getattr(settings, "stepfun_model", "step-3.5-flash")
        self.base_url = getattr(settings, "stepfun_base_url", STEPFUN_BASE_URL)
        self._models = model_ids_for_provider("stepfun")
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
    ) -> CompletionResponse:
        if not self.available:
            raise RuntimeError("StepFun provider not available. Set STEPFUN_API_KEY.")
        await self.rate_limiter.acquire()
        start_time = time.time()

        async def _do():
            return await self.client.chat.completions.create(
                model=model or self.model,
                messages=[{"role": m.role.value, "content": m.content} for m in messages],
                temperature=temperature,
                max_tokens=max_tokens or 8192,
                top_p=top_p,
                stop=stop,
            )

        response = await self._retry_with_backoff(_do)
        latency_ms = (time.time() - start_time) * 1000
        choice = response.choices[0]
        usage = response.usage
        return CompletionResponse(
            content=choice.message.content or "",
            provider=ProviderType.STEPFUN,
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
        if not self.available:
            raise RuntimeError("StepFun provider not available. Set STEPFUN_API_KEY.")
        await self.rate_limiter.acquire()

        stream_resp = await self.client.chat.completions.create(
            model=model or self.model,
            messages=[{"role": m.role.value, "content": m.content} for m in messages],
            temperature=temperature,
            max_tokens=max_tokens or 8192,
            top_p=top_p,
            stop=stop,
            stream=True,
        )
        async for chunk in stream_resp:
            if chunk.choices[0].delta.content:
                yield chunk.choices[0].delta.content

    async def embed(
        self,
        text: str,
        model: Optional[str] = None,
        dimensions: Optional[int] = None,
    ) -> EmbeddingResponse:
        raise NotImplementedError("StepFun does not provide an embeddings API.")

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
        """Estimate cost from published StepFun API pricing (per 1M tokens)."""
        total_tokens = input_tokens + output_tokens
        model_lower = model.lower()

        if "3.5-flash" in model_lower or "flash" in model_lower:
            input_per_m, output_per_m = 0.05, 0.15
        elif "3" in model_lower and "2" not in model_lower:
            input_per_m, output_per_m = 0.50, 1.50
        elif "2-mini" in model_lower:
            input_per_m, output_per_m = 0.10, 0.30
        elif "2" in model_lower:
            input_per_m, output_per_m = 0.50, 1.50
        elif "1" in model_lower:
            input_per_m, output_per_m = 0.20, 0.60
        else:
            input_per_m, output_per_m = 0.20, 0.60

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
