"""Provider manager for FlyMind AI Service.

This module manages all LLM providers and provides a unified interface.
"""

from typing import Optional, Dict
import logging

from .base import BaseProvider
from .openai import OpenAIProvider
from .anthropic import AnthropicProvider
from .ollama import OllamaProvider
from .openrouter import OpenRouterProvider
from .fireworks import FireworksProvider
from .groq import GroqProvider
from .deepinfra import DeepInfraProvider
from .together import TogetherProvider
from .router import ProviderRouter, init_provider_router
from ..config import settings
from ..models.schemas import ProviderType, ProviderInfo


logger = logging.getLogger(__name__)


class ProviderManager:
    """Manages all LLM providers and provides a unified interface."""

    def __init__(self):
        self._providers: Dict[str, BaseProvider] = {}
        self._default_provider: str = settings.default_provider
        self._default_embedding_provider: str = settings.default_embedding_provider

        # Initialize all providers
        self._init_providers()

    def _init_providers(self) -> None:
        """Initialize all available providers."""
        # Initialize OpenAI
        try:
            openai = OpenAIProvider()
            self._providers[ProviderType.OPENAI.value] = openai
            logger.info(f"Initialized OpenAI provider (available: {openai.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize OpenAI provider: {e}")

        # Initialize Anthropic
        try:
            anthropic = AnthropicProvider()
            self._providers[ProviderType.ANTHROPIC.value] = anthropic
            logger.info(f"Initialized Anthropic provider (available: {anthropic.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize Anthropic provider: {e}")

        # Initialize Ollama
        try:
            ollama = OllamaProvider()
            self._providers[ProviderType.OLLAMA.value] = ollama
            logger.info(f"Initialized Ollama provider (available: {ollama.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize Ollama provider: {e}")

        # Initialize OpenRouter
        try:
            openrouter = OpenRouterProvider()
            self._providers[ProviderType.OPENROUTER.value] = openrouter
            logger.info(f"Initialized OpenRouter provider (available: {openrouter.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize OpenRouter provider: {e}")

        # Initialize Fireworks AI (primary for function calling)
        try:
            fireworks = FireworksProvider()
            self._providers[ProviderType.FIREWORKS.value] = fireworks
            logger.info(f"Initialized Fireworks AI provider (available: {fireworks.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize Fireworks AI provider: {e}")

        # Initialize Groq (low-latency real-time)
        try:
            groq = GroqProvider()
            self._providers[ProviderType.GROQ.value] = groq
            logger.info(f"Initialized Groq provider (available: {groq.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize Groq provider: {e}")

        # Initialize DeepInfra (cost-effective background/batch)
        try:
            deepinfra = DeepInfraProvider()
            self._providers[ProviderType.DEEPINFRA.value] = deepinfra
            logger.info(f"Initialized DeepInfra provider (available: {deepinfra.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize DeepInfra provider: {e}")

        # Initialize Together AI (alternative/batch)
        try:
            together = TogetherProvider()
            self._providers[ProviderType.TOGETHER.value] = together
            logger.info(f"Initialized Together AI provider (available: {together.available})")
        except Exception as e:
            logger.warning(f"Failed to initialize Together AI provider: {e}")

        # Initialize traffic-based provider router
        if settings.enable_traffic_based_routing:
            try:
                init_provider_router(self._providers)
                logger.info("Initialized traffic-based provider router")
            except Exception as e:
                logger.warning(f"Failed to initialize provider router: {e}")

    def get_provider(self, name: Optional[str] = None) -> BaseProvider:
        """Get a provider by name.

        Args:
            name: Provider name (openai, anthropic, ollama, openrouter)
                  If None, uses the default provider

        Returns:
            The provider instance

        Raises:
            ValueError: If provider is not available or not found
        """
        provider_name = name or self._default_provider

        if provider_name not in self._providers:
            raise ValueError(f"Provider '{provider_name}' not found")

        provider = self._providers[provider_name]

        if not provider.available:
            raise ValueError(f"Provider '{provider_name}' is not available")

        return provider

    def get_provider_for_chat(self, name: Optional[str] = None) -> BaseProvider:
        """Resolve an LLM provider for chat/completions.

        Tries *name* (if given), then :attr:`default_provider`, then common fallbacks
        so local dev works with Ollama when ``OPENAI_API_KEY`` is unset.
        """
        order: list[str] = []
        if name:
            order.append(name)
        order.append(self._default_provider)
        for extra in ("ollama", "openrouter", "openai", "anthropic"):
            order.append(extra)
        seen: set[str] = set()
        for pname in order:
            if pname in seen:
                continue
            seen.add(pname)
            if pname not in self._providers:
                continue
            provider = self._providers[pname]
            if provider.available:
                preferred = name or self._default_provider
                if pname != preferred:
                    logger.info(
                        "LLM provider fallback: using '%s' (preferred '%s' unavailable)",
                        pname,
                        preferred,
                    )
                return provider
        raise ValueError("No LLM provider available")

    def get_embedding_provider(self, name: Optional[str] = None) -> BaseProvider:
        """Get a provider that supports embeddings.

        Args:
            name: Provider name (optional)

        Returns:
            The provider instance that supports embeddings

        Raises:
            ValueError: If no embedding provider is available
        """
        provider_name = name or self._default_embedding_provider

        # Try the requested provider first
        if provider_name in self._providers:
            provider = self._providers[provider_name]
            if provider.available and provider.supports_embeddings:
                return provider

        # Fall back to finding any available embedding provider
        for pname, provider in self._providers.items():
            if provider.available and hasattr(provider, 'supports_embeddings'):
                if provider.get_provider_info().supports_embeddings:
                    logger.info(f"Falling back to {pname} for embeddings")
                    return provider

        raise ValueError("No embedding provider available")

    def list_providers(self) -> list[ProviderInfo]:
        """Get information about all providers.

        Returns:
            List of ProviderInfo for all registered providers
        """
        providers = []
        for name, provider in self._providers.items():
            try:
                providers.append(provider.get_provider_info())
            except Exception as e:
                logger.warning(f"Failed to get provider info for {name}: {e}")
                # Add unavailable provider info
                providers.append(ProviderInfo(
                    name=name,
                    display_name=name.title(),
                    available=False,
                    models=[],
                    rate_limit=0,
                    embedding_dimensions=0,
                    supports_streaming=False,
                    supports_embeddings=False,
                ))
        return providers

    async def health_check_all(self) -> Dict[str, bool]:
        """Run health checks on all providers.

        Returns:
            Dictionary of provider name to health status
        """
        results = {}
        for name, provider in self._providers.items():
            try:
                results[name] = await provider.health_check()
            except Exception as e:
                logger.warning(f"Health check failed for {name}: {e}")
                results[name] = False
        return results

    @property
    def default_provider(self) -> str:
        """Get the default provider name."""
        return self._default_provider

    @property
    def default_embedding_provider(self) -> str:
        """Get the default embedding provider name."""
        return self._default_embedding_provider

    def get_provider_router(self) -> Optional[ProviderRouter]:
        """Get the traffic-based provider router if available.

        Returns:
            ProviderRouter instance or None if not initialized
        """
        try:
            from .router import get_provider_router
            return get_provider_router()
        except RuntimeError:
            return None

    def get_functionfly_recommendations(self) -> list[dict]:
        """Get FunctionFly-specific provider recommendations.

        Returns:
            List of recommendations for different use cases
        """
        try:
            router = self.get_provider_router()
            if router:
                return router.get_recommendations()
        except Exception as e:
            logger.debug(f"Could not get recommendations: {e}")

        # Fallback recommendations
        return [
            {"use_case": "Real-time agent calls", "provider": "groq", "reason": "Lowest latency"},
            {"use_case": "Function calling / JSON", "provider": "fireworks", "reason": "FireAttention optimized"},
            {"use_case": "Background/batch work", "provider": "deepinfra", "reason": "Cost-effective"},
            {"use_case": "Multi-model routing", "provider": "openrouter", "reason": "Agnostic routing"},
        ]


# Global provider manager instance
_provider_manager: Optional[ProviderManager] = None


def get_provider_manager() -> ProviderManager:
    """Get the global provider manager instance.

    Returns:
        The ProviderManager instance
    """
    global _provider_manager
    if _provider_manager is None:
        _provider_manager = ProviderManager()
    return _provider_manager
