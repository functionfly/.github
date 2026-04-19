"""Cost tracking integration for recording completions to economic memory.

Wraps provider calls to automatically record execution data including
cost, latency, and quality metrics.
"""

import logging
import time
import uuid
from typing import Optional, Callable, Any
from functools import wraps

from ..models.schemas import ChatMessage, CompletionResponse, ProviderType, CostTracking
from ..services.economic_memory import ExecutionRecord, get_economic_memory

logger = logging.getLogger(__name__)


# Provider cost rates (per 1K tokens) - used for cost estimation
# These should be updated as pricing changes
PROVIDER_COST_RATES = {
    ProviderType.OPENAI: {
        "gpt-4o": {"input": 0.005, "output": 0.015},
        "gpt-4o-mini": {"input": 0.00015, "output": 0.0006},
        "gpt-3.5-turbo": {"input": 0.0005, "output": 0.0015},
        "default": {"input": 0.002, "output": 0.006},
    },
    ProviderType.ANTHROPIC: {
        "claude-3-opus": {"input": 0.015, "output": 0.075},
        "claude-3-sonnet": {"input": 0.003, "output": 0.015},
        "claude-3-haiku": {"input": 0.00025, "output": 0.00125},
        "default": {"input": 0.003, "output": 0.015},
    },
    ProviderType.GROQ: {
        "default": {"input": 0.0001, "output": 0.0001},  # Very cheap
    },
    ProviderType.TOGETHER: {
        "default": {"input": 0.0002, "output": 0.0002},
    },
    ProviderType.DEEPINFRA: {
        "default": {"input": 0.00015, "output": 0.00015},
    },
    ProviderType.FIREWORKS: {
        "default": {"input": 0.0002, "output": 0.0002},
    },
    ProviderType.OPENROUTER: {
        "default": {"input": 0.0005, "output": 0.0015},
    },
    ProviderType.OLLAMA: {
        "default": {"input": 0.0, "output": 0.0},  # Self-hosted
    },
}


def estimate_cost(
    provider: ProviderType,
    model: str,
    input_tokens: int,
    output_tokens: int,
) -> float:
    """Estimate cost for a completion based on provider and model.
    
    Args:
        provider: Provider type
        model: Model name
        input_tokens: Number of input tokens
        output_tokens: Number of output tokens
        
    Returns:
        Estimated cost in USD
    """
    provider_rates = PROVIDER_COST_RATES.get(provider, {})
    model_rates = provider_rates.get(model, provider_rates.get("default", {"input": 0.0, "output": 0.0}))
    
    input_cost = (input_tokens / 1000) * model_rates["input"]
    output_cost = (output_tokens / 1000) * model_rates["output"]
    
    return round(input_cost + output_cost, 6)


class EconomicTracker:
    """Tracks completions for economic memory recording."""
    
    def __init__(
        self,
        provider: ProviderType,
        model: str,
        tenant_id: Optional[str] = None,
        function_id: Optional[str] = None,
    ):
        self.provider = provider
        self.model = model
        self.tenant_id = tenant_id
        self.function_id = function_id
        self.execution_id = str(uuid.uuid4())
        self.start_time: Optional[float] = None
        self.end_time: Optional[float] = None
        
    def __enter__(self) -> "EconomicTracker":
        """Start tracking."""
        self.start_time = time.time()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        """End tracking - record to economic memory."""
        self.end_time = time.time()
        
    async def record_completion(
        self,
        response: CompletionResponse,
        success: bool = True,
        error: Optional[str] = None,
    ) -> None:
        """Record a completion response to economic memory.
        
        Args:
            response: The completion response
            success: Whether the request succeeded
            error: Error message if failed
        """
        if self.start_time is None:
            self.start_time = time.time()
        if self.end_time is None:
            self.end_time = time.time()
        
        latency_ms = (self.end_time - self.start_time) * 1000
        
        # Estimate cost if not in response
        cost_usd = 0.0
        if hasattr(response, 'cost_usd') and response.cost_usd is not None:
            cost_usd = response.cost_usd
        else:
            cost_usd = estimate_cost(
                self.provider,
                self.model,
                response.usage.get("prompt_tokens", 0),
                response.usage.get("completion_tokens", 0),
            )
        
        # Create execution record
        record = ExecutionRecord(
            execution_id=self.execution_id,
            provider=self.provider,
            model=self.model,
            input_tokens=response.usage.get("prompt_tokens", 0),
            output_tokens=response.usage.get("completion_tokens", 0),
            total_tokens=response.usage.get("total_tokens", 0),
            cost_usd=cost_usd,
            latency_ms=latency_ms,
            success=success,
            error_type=error if error else None,
            tenant_id=self.tenant_id,
            function_id=self.function_id,
        )
        
        # Record to economic memory
        try:
            memory = get_economic_memory()
            await memory.record_execution(record)
            logger.debug(
                f"Recorded execution {self.execution_id} to economic memory: "
                f"{self.provider.value}/{self.model}, cost=${cost_usd:.6f}"
            )
        except Exception as e:
            # Don't fail the request if recording fails
            logger.warning(f"Failed to record to economic memory: {e}")
    
    async def record_failure(
        self,
        error: Exception,
        input_tokens: int = 0,
    ) -> None:
        """Record a failed completion to economic memory.
        
        Args:
            error: The exception that occurred
            input_tokens: Tokens consumed before failure
        """
        if self.start_time is None:
            self.start_time = time.time()
        if self.end_time is None:
            self.end_time = time.time()
        
        latency_ms = (self.end_time - self.start_time) * 1000
        
        # Failed requests still have cost (usually)
        cost_usd = estimate_cost(self.provider, self.model, input_tokens, 0)
        
        error_type = type(error).__name__
        
        record = ExecutionRecord(
            execution_id=self.execution_id,
            provider=self.provider,
            model=self.model,
            input_tokens=input_tokens,
            output_tokens=0,
            total_tokens=input_tokens,
            cost_usd=cost_usd,
            latency_ms=latency_ms,
            success=False,
            error_type=error_type,
            tenant_id=self.tenant_id,
            function_id=self.function_id,
        )
        
        try:
            memory = get_economic_memory()
            await memory.record_execution(record)
        except Exception as e:
            logger.warning(f"Failed to record failure to economic memory: {e}")


def track_completion(
    provider: ProviderType,
    model: Optional[str],
    tenant_id: Optional[str] = None,
    function_id: Optional[str] = None,
) -> EconomicTracker:
    """Create an economic tracker for a completion.
    
    Usage:
        with track_completion(ProviderType.OPENAI, "gpt-4o-mini") as tracker:
            response = await provider.complete(...)
            await tracker.record_completion(response)
    
    Args:
        provider: Provider type
        model: Model name
        tenant_id: Optional tenant ID
        function_id: Optional function ID
        
    Returns:
        EconomicTracker context manager
    """
    return EconomicTracker(
        provider=provider,
        model=model or "default",
        tenant_id=tenant_id,
        function_id=function_id,
    )


class TrackedProvider:
    """Wrapper that adds economic tracking to a base provider.
    
    This wraps a BaseProvider and automatically records all completions
    to the economic memory system.
    """
    
    def __init__(
        self,
        provider: Any,  # BaseProvider
        provider_type: ProviderType,
    ):
        self._provider = provider
        self._provider_type = provider_type
        self._name = provider.name
        self._display_name = provider.display_name
        
    @property
    def name(self) -> str:
        return self._name
    
    @property
    def display_name(self) -> str:
        return self._display_name
    
    @property
    def available(self) -> bool:
        return self._provider.available
    
    @property
    def models(self) -> list[str]:
        return self._provider.models
    
    async def complete(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
        tenant_id: Optional[str] = None,
        **kwargs,
    ) -> CompletionResponse:
        """Complete with tracking."""
        tracker = EconomicTracker(
            provider=self._provider_type,
            model=model or "default",
            tenant_id=tenant_id,
        )
        
        with tracker:
            try:
                response = await self._provider.complete(
                    messages=messages,
                    model=model,
                    temperature=temperature,
                    max_tokens=max_tokens,
                    top_p=top_p,
                    stop=stop,
                )
                await tracker.record_completion(response, success=True)
                return response
            except Exception as e:
                # Estimate input tokens from messages for failure recording
                input_tokens = sum(len(m.content.split()) for m in messages) * 1.5  # Rough estimate
                await tracker.record_failure(e, int(input_tokens))
                raise
    
    async def stream(
        self,
        messages: list[ChatMessage],
        model: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        top_p: Optional[float] = None,
        stop: Optional[list[str]] = None,
        **kwargs,
    ):
        """Stream with tracking."""
        tracker = EconomicTracker(
            provider=self._provider_type,
            model=model or "default",
        )
        
        with tracker:
            try:
                async for chunk in self._provider.stream(
                    messages=messages,
                    model=model,
                    temperature=temperature,
                    max_tokens=max_tokens,
                    top_p=top_p,
                    stop=stop,
                ):
                    yield chunk
                    
                # Stream completed successfully
                # Note: We don't have full usage stats for streams
                # so this is a partial record
                
            except Exception as e:
                raise
    
    async def embed(
        self,
        text: str,
        model: Optional[str] = None,
        dimensions: Optional[int] = None,
    ) -> Any:
        """Embed with tracking."""
        return await self._provider.embed(text, model, dimensions)


def wrap_provider_with_tracking(
    provider: Any,
    provider_type: ProviderType,
) -> TrackedProvider:
    """Wrap a provider with economic tracking.
    
    Args:
        provider: Base provider to wrap
        provider_type: Provider type enum value
        
    Returns:
        TrackedProvider wrapper
    """
    return TrackedProvider(provider, provider_type)
