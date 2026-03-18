"""Ollama provider implementation.

This module provides the Ollama provider for local/development use.
"""

from typing import Optional, AsyncGenerator
import json
import time
import httpx

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


class OllamaProvider(BaseProvider):
    """Ollama local provider.

    Supports local LLM models via Ollama API.
    """

    def __init__(self):
        super().__init__(
            name="ollama",
            display_name="Ollama (Local)",
            rate_limit=settings.ollama_rate_limit,
            retry_config=RetryConfig(
                max_retries=settings.max_retries,
                base_delay=settings.retry_base_delay,
                max_delay=settings.retry_max_delay,
            ),
        )
        self.base_url = settings.ollama_base_url
        self.model = settings.ollama_model
        self.embedding_model = settings.ollama_embedding_model
        self.embedding_dimensions = 768  # Default for nomic-embed-text
        self._models = []

        # Create HTTP client
        self.client = httpx.AsyncClient(
            base_url=self.base_url,
            timeout=60.0,
        )

    @property
    def available(self) -> bool:
        """Check if Ollama is available."""
        return self._available

    async def _check_available(self) -> bool:
        """Check if Ollama is running and accessible."""
        try:
            response = await self.client.get("/api/tags")
            if response.status_code == 200:
                data = response.json()
                self._models = [m.get("name", "") for m in data.get("models", [])]
                self._available = True
                return True
        except Exception:
            pass
        self._available = False
        return False

    async def health_check(self) -> bool:
        """Check if Ollama is healthy."""
        return await self._check_available()

    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
    ) -> CompletionResponse:
        """Generate a completion using Ollama API."""
        if not await self._check_available():
            raise RuntimeError("Ollama provider not available. Is Ollama running?")

        await self.rate_limiter.acquire()

        start_time = time.time()

        # Convert messages to Ollama format
        ollama_messages = [
            {"role": m.role.value, "content": m.content}
            for m in messages
        ]

        payload = {
            "model": model or self.model,
            "messages": ollama_messages,
            "temperature": temperature,
            "stream": False,
        }

        if max_tokens:
            payload["options"] = {"num_predict": max_tokens}
        if top_p:
            if "options" not in payload:
                payload["options"] = {}
            payload["options"]["top_p"] = top_p
        if stop:
            if "options" not in payload:
                payload["options"] = {}
            payload["options"]["stop"] = stop

        def _do_completion():
            return self.client.post("/api/chat", json=payload)

        response = await self._retry_with_backoff(_do_completion)

        if response.status_code != 200:
            raise RuntimeError(f"Ollama API error: {response.text}")

        data = response.json()

        latency_ms = (time.time() - start_time) * 1000

        return CompletionResponse(
            content=data.get("message", {}).get("content", ""),
            provider=ProviderType.OLLAMA,
            model=data.get("model", model or self.model),
            usage={
                "prompt_tokens": data.get("prompt_eval_count", 0),
                "completion_tokens": data.get("eval_count", 0),
                "total_tokens": data.get("prompt_eval_count", 0) + data.get("eval_count", 0),
            },
            finish_reason=data.get("done_reason"),
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
        """Stream completion using Ollama API."""
        if not await self._check_available():
            raise RuntimeError("Ollama provider not available. Is Ollama running?")

        await self.rate_limiter.acquire()

        # Convert messages to Ollama format
        ollama_messages = [
            {"role": m.role.value, "content": m.content}
            for m in messages
        ]

        payload = {
            "model": model or self.model,
            "messages": ollama_messages,
            "temperature": temperature,
            "stream": True,
        }

        if max_tokens:
            payload["options"] = {"num_predict": max_tokens}
        if top_p:
            if "options" not in payload:
                payload["options"] = {}
            payload["options"]["top_p"] = top_p
        if stop:
            if "options" not in payload:
                payload["options"] = {}
            payload["options"]["stop"] = stop

        async with self.client.stream("POST", "/api/chat", json=payload) as response:
            if response.status_code != 200:
                raise RuntimeError(f"Ollama API error: {response.text}")

            async for line in response.aiter_lines():
                if line.strip():
                    try:
                        data = json.loads(line)  # Parse the JSON line safely
                        if "message" in data and "content" in data["message"]:
                            content = data["message"]["content"]
                            if content:
                                yield content
                    except Exception:
                        pass

    async def embed(
        self,
        text: str,
        model: Optional[str] = None,
        dimensions: Optional[int] = None,
    ) -> EmbeddingResponse:
        """Generate embeddings using Ollama API."""
        if not await self._check_available():
            raise RuntimeError("Ollama provider not available. Is Ollama running?")

        await self.rate_limiter.acquire()

        start_time = time.time()

        payload = {
            "model": model or self.embedding_model,
            "input": text,
        }

        try:
            # Try the new embeddings API
            response = await self.client.post("/api/embeddings", json=payload)

            if response.status_code != 200:
                raise RuntimeError(f"Ollama embeddings API error: {response.text}")

            data = response.json()

            latency_ms = (time.time() - start_time) * 1000

            embedding = data.get("embedding", [])

            return EmbeddingResponse(
                embedding=embedding,
                provider=ProviderType.OLLAMA,
                model=data.get("model", model or self.embedding_model),
                dimensions=len(embedding),
                usage={"tokens": data.get("token_count", 0)},
                latency_ms=latency_ms,
            )
        except Exception as e:
            # Fallback to older API format
            raise RuntimeError(
                f"Ollama embeddings not available. Make sure the embedding model is installed. Error: {e}"
            )

    def get_provider_info(self) -> ProviderInfo:
        """Get information about Ollama provider."""
        return ProviderInfo(
            name=self.name,
            display_name=self.display_name,
            available=self._available,
            models=self._models if self._models else [self.model],
            rate_limit=self.rate_limiter.rate,
            embedding_dimensions=self.embedding_dimensions,
            supports_streaming=True,
            supports_embeddings=True,
        )

    def calculate_cost(
        self,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> CostTracking:
        """Calculate cost for Ollama request.

        Ollama is local, so there's no API cost.
        """
        total_tokens = input_tokens + output_tokens

        return CostTracking(
            provider=self.name,
            model=model,
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            total_tokens=total_tokens,
            estimated_cost=0.0,  # Free for local
        )

    async def close(self):
        """Close the HTTP client."""
        await self.client.aclose()
