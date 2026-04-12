"""Traffic-based provider router for FunctionFly.

Routes AI inference requests to the optimal provider based on traffic type:
- Real-time agent function calls: Groq (lowest latency)
- Structured output / tool use: Fireworks (best function calling perf)
- Background/batch tasks: DeepInfra (cost-optimized)
- Embeddings: DeepInfra or Fireworks
- Fallback: OpenRouter (multi-model routing)

Architecture:
| Traffic type              | Provider              |
|---------------------------|-----------------------|
| Real-time agent calls     | Groq or Fireworks     |
| Structured output/tool    | Fireworks             |
| Embeddings/background     | DeepInfra             |
| Model-agnostic routing    | OpenRouter            |

All providers are OpenAI-compatible, so swapping is seamless.
"""

from typing import Optional, Dict, List, Any, Callable
from dataclasses import dataclass
from enum import Enum
import logging

from ..models.schemas import (
    ProviderType,
    TrafficType,
    CompletionRequest,
    EmbeddingRequest,
)
from .base import BaseProvider

logger = logging.getLogger(__name__)


@dataclass
class RoutingRule:
    """A routing rule for selecting providers."""
    traffic_type: TrafficType
    primary_provider: ProviderType
    fallback_providers: List[ProviderType]
    model_override: Optional[str] = None
    description: str = ""


class ProviderRouter:
    """Routes inference requests to optimal providers based on traffic type.

    For FunctionFly specifically, this implements the recommended architecture:
    - Fireworks as primary for structured output/function calling
    - Groq for latency-critical real-time calls
    - DeepInfra for background/batch work
    - OpenRouter as fallback for model-agnostic routing
    """

    # Default routing rules for FunctionFly
    DEFAULT_RULES: Dict[TrafficType, RoutingRule] = {
        TrafficType.REALTIME: RoutingRule(
            traffic_type=TrafficType.REALTIME,
            primary_provider=ProviderType.GROQ,
            fallback_providers=[
                ProviderType.FIREWORKS,
                ProviderType.OPENROUTER,
                ProviderType.OPENAI,
            ],
            model_override=None,  # Uses provider default
            description="Low-latency real-time agent function calls",
        ),
        TrafficType.STRUCTURED: RoutingRule(
            traffic_type=TrafficType.STRUCTURED,
            primary_provider=ProviderType.FIREWORKS,
            fallback_providers=[
                ProviderType.OPENROUTER,
                ProviderType.OPENAI,
                ProviderType.TOGETHER,
            ],
            model_override=None,
            description="Structured output, JSON mode, tool use",
        ),
        TrafficType.FUNCTION_CALLING: RoutingRule(
            traffic_type=TrafficType.FUNCTION_CALLING,
            primary_provider=ProviderType.FIREWORKS,
            fallback_providers=[
                ProviderType.OPENROUTER,
                ProviderType.OPENAI,
                ProviderType.GROQ,
            ],
            model_override=None,
            description="Function calling optimized with FireAttention engine",
        ),
        TrafficType.BACKGROUND: RoutingRule(
            traffic_type=TrafficType.BACKGROUND,
            primary_provider=ProviderType.DEEPINFRA,
            fallback_providers=[
                ProviderType.TOGETHER,
                ProviderType.FIREWORKS,
                ProviderType.OPENAI,
            ],
            model_override=None,
            description="Cost-optimized batch processing and background tasks",
        ),
        TrafficType.GENERAL: RoutingRule(
            traffic_type=TrafficType.GENERAL,
            primary_provider=ProviderType.FIREWORKS,
            fallback_providers=[
                ProviderType.OPENROUTER,
                ProviderType.TOGETHER,
                ProviderType.OPENAI,
            ],
            model_override=None,
            description="Default routing with Fireworks as primary",
        ),
    }

    def __init__(
        self,
        providers: Dict[str, BaseProvider],
        rules: Optional[Dict[TrafficType, RoutingRule]] = None,
    ):
        """Initialize the provider router.

        Args:
            providers: Dictionary of provider name to provider instance
            rules: Optional custom routing rules (defaults to DEFAULT_RULES)
        """
        self._providers = providers
        self._rules = rules or self.DEFAULT_RULES.copy()
        self._traffic_classifiers: List[Callable[[CompletionRequest], TrafficType]] = []

    def classify_traffic(
        self,
        request: CompletionRequest,
        hint: Optional[TrafficType] = None,
    ) -> TrafficType:
        """Classify the traffic type for a completion request.

        Args:
            request: The completion request
            hint: Optional explicit traffic type hint

        Returns:
            Classified traffic type
        """
        # Use explicit hint if provided
        if hint:
            return hint

        # Check for function calling indicators
        messages_str = " ".join(
            m.content for m in request.messages if hasattr(m, 'content')
        ).lower()

        # Check for structured output / JSON mode
        if request.stop and any(
            s in str(request.stop) for s in ["json", "schema", "function"]
        ):
            return TrafficType.STRUCTURED

        # Check for function calling keywords
        function_keywords = [
            "function", "call", "tool", "invoke", "execute",
            "parameters", "arguments", "return type",
        ]
        if any(kw in messages_str for kw in function_keywords):
            return TrafficType.FUNCTION_CALLING

        # Check for real-time indicators (short, urgent messages)
        if len(request.messages) <= 2 and len(messages_str) < 200:
            return TrafficType.REALTIME

        # Default to general
        return TrafficType.GENERAL

    def get_provider_for_traffic(
        self,
        traffic_type: TrafficType,
        require_embeddings: bool = False,
    ) -> tuple[BaseProvider, ProviderType]:
        """Get the best available provider for a traffic type.

        Args:
            traffic_type: The type of traffic
            require_embeddings: Whether the provider must support embeddings

        Returns:
            Tuple of (provider_instance, provider_type)

        Raises:
            ValueError: If no suitable provider is available
        """
        rule = self._rules.get(traffic_type, self._rules[TrafficType.GENERAL])

        # Try primary first
        providers_to_try = [rule.primary_provider] + rule.fallback_providers

        for provider_type in providers_to_try:
            provider_name = provider_type.value
            if provider_name not in self._providers:
                continue

            provider = self._providers[provider_name]

            if not provider.available:
                logger.debug(f"Provider {provider_name} not available, trying fallback")
                continue

            if require_embeddings and not provider.supports_embeddings:
                logger.debug(f"Provider {provider_name} doesn't support embeddings, trying fallback")
                continue

            return provider, provider_type

        raise ValueError(
            f"No suitable provider available for traffic type '{traffic_type.value}'. "
            f"Tried: {[p.value for p in providers_to_try]}"
        )

    async def route_completion(
        self,
        request: CompletionRequest,
        traffic_hint: Optional[TrafficType] = None,
    ) -> tuple[BaseProvider, ProviderType, Optional[str]]:
        """Route a completion request to the optimal provider.

        Args:
            request: The completion request
            traffic_hint: Optional explicit traffic type hint

        Returns:
            Tuple of (provider_instance, provider_type, model_override)
        """
        traffic_type = self.classify_traffic(request, traffic_hint)
        provider, provider_type = self.get_provider_for_traffic(
            traffic_type,
            require_embeddings=False,
        )

        # Get model override from routing rule if set
        rule = self._rules.get(traffic_type)
        model_override = rule.model_override if rule else None

        logger.debug(
            f"Routed {traffic_type.value} traffic to {provider_type.value} "
            f"(requested model: {request.model or 'default'})"
        )

        return provider, provider_type, model_override

    async def route_embedding(
        self,
        request: EmbeddingRequest,
        traffic_hint: Optional[TrafficType] = None,
    ) -> tuple[BaseProvider, ProviderType]:
        """Route an embedding request to the optimal provider.

        Args:
            request: The embedding request
            traffic_hint: Optional explicit traffic type hint

        Returns:
            Tuple of (provider_instance, provider_type)
        """
        # Embeddings always route to BACKGROUND traffic type for cost optimization
        traffic_type = traffic_hint or TrafficType.BACKGROUND

        provider, provider_type = self.get_provider_for_traffic(
            traffic_type,
            require_embeddings=True,
        )

        logger.debug(
            f"Routed embedding request to {provider_type.value} "
            f"(model: {request.model or 'default'})"
        )

        return provider, provider_type

    def update_rule(
        self,
        traffic_type: TrafficType,
        primary_provider: Optional[ProviderType] = None,
        fallback_providers: Optional[List[ProviderType]] = None,
        model_override: Optional[str] = None,
    ) -> None:
        """Update a routing rule at runtime.

        Args:
            traffic_type: The traffic type to update
            primary_provider: New primary provider
            fallback_providers: New fallback provider list
            model_override: New model override
        """
        if traffic_type not in self._rules:
            self._rules[traffic_type] = RoutingRule(
                traffic_type=traffic_type,
                primary_provider=primary_provider or ProviderType.FIREWORKS,
                fallback_providers=fallback_providers or [],
                model_override=model_override,
                description="Custom routing rule",
            )
        else:
            rule = self._rules[traffic_type]
            if primary_provider:
                rule.primary_provider = primary_provider
            if fallback_providers:
                rule.fallback_providers = fallback_providers
            if model_override:
                rule.model_override = model_override

        logger.info(f"Updated routing rule for {traffic_type.value}")

    def get_routing_table(self) -> Dict[str, Any]:
        """Get the current routing configuration.

        Returns:
            Dictionary describing the routing table
        """
        table = {}
        for traffic_type, rule in self._rules.items():
            # Check availability of each provider in the chain
            primary_available = (
                rule.primary_provider.value in self._providers and
                self._providers[rule.primary_provider.value].available
            )
            fallback_availability = [
                {
                    "provider": p.value,
                    "available": p.value in self._providers and self._providers[p.value].available,
                }
                for p in rule.fallback_providers
            ]

            table[traffic_type.value] = {
                "primary": {
                    "provider": rule.primary_provider.value,
                    "available": primary_available,
                },
                "fallbacks": fallback_availability,
                "description": rule.description,
            }
        return table

    def get_recommendations(self) -> List[Dict[str, str]]:
        """Get provider recommendations for FunctionFly use cases.

        Returns:
            List of recommendations with use cases
        """
        return [
            {
                "use_case": "Real-time agent function calls",
                "provider": "groq",
                "reason": "LPU hardware delivers 0.6-0.9s time-to-first-token consistently",
            },
            {
                "use_case": "Structured output / JSON mode / Tool use",
                "provider": "fireworks",
                "reason": "FireAttention engine optimized for function calling, 4x lower latency than vLLM",
            },
            {
                "use_case": "Function generation / code completion",
                "provider": "fireworks",
                "reason": "Best structured output for function schemas",
            },
            {
                "use_case": "Background tasks / Batch processing",
                "provider": "deepinfra",
                "reason": "Serverless pricing cuts costs up to 90% vs provisioned",
            },
            {
                "use_case": "Embeddings / Vector search",
                "provider": "deepinfra",
                "reason": "Cost-effective embedding models with good throughput",
            },
            {
                "use_case": "Multi-model A/B testing",
                "provider": "openrouter",
                "reason": "Single API surface across 100+ models",
            },
            {
                "use_case": "Fallback / Generic routing",
                "provider": "openrouter",
                "reason": "Automatic fallback across multiple backends",
            },
        ]


# Global router instance (initialized by manager)
_provider_router: Optional[ProviderRouter] = None


def get_provider_router() -> ProviderRouter:
    """Get the global provider router instance.

    Returns:
        The ProviderRouter instance

    Raises:
        RuntimeError: If router hasn't been initialized
    """
    global _provider_router
    if _provider_router is None:
        raise RuntimeError(
            "ProviderRouter not initialized. "
            "Call init_provider_router() first."
        )
    return _provider_router


def init_provider_router(providers: Dict[str, BaseProvider]) -> ProviderRouter:
    """Initialize the global provider router.

    Args:
        providers: Dictionary of provider name to provider instance

    Returns:
        The initialized ProviderRouter
    """
    global _provider_router
    _provider_router = ProviderRouter(providers)
    logger.info("ProviderRouter initialized with traffic-based routing")
    return _provider_router
