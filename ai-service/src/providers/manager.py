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
