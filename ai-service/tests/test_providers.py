"""Tests for LLM providers."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch

from src.providers.base import BaseProvider, RateLimiter, RetryConfig
from src.providers.openai import OpenAIProvider
from src.providers.ollama import OllamaProvider
from src.models.schemas import ChatMessage, MessageRole


class TestRateLimiter:
    """Tests for RateLimiter class."""

    def test_rate_limiter_initialization(self):
        """Test rate limiter initializes correctly."""
        limiter = RateLimiter(rate=60)
        assert limiter.rate == 60
        assert limiter.interval == 1.0  # 60/60

    @pytest.mark.asyncio
    async def test_rate_limiter_acquire(self):
        """Test rate limiter acquire."""
        limiter = RateLimiter(rate=60)
        # Should not raise
        await limiter.acquire()


class TestRetryConfig:
    """Tests for RetryConfig class."""

    def test_retry_config_initialization(self):
        """Test retry config initializes correctly."""
        config = RetryConfig(max_retries=3, base_delay=1.0, max_delay=30.0)
        assert config.max_retries == 3
        assert config.base_delay == 1.0
        assert config.max_delay == 30.0

    def test_exponential_backoff(self):
        """Test exponential backoff calculation."""
        config = RetryConfig(max_retries=3, base_delay=1.0, max_delay=30.0)

        assert config.get_delay(0) == 1.0
        assert config.get_delay(1) == 2.0
        assert config.get_delay(2) == 4.0
        assert config.get_delay(3) == 8.0
        assert config.get_delay(10) == 30.0  # Capped at max_delay


class TestOpenAIProvider:
    """Tests for OpenAI provider."""

    def test_openai_provider_initialization(self):
        """Test OpenAI provider initializes correctly."""
        with patch('src.providers.openai.settings') as mock_settings:
            mock_settings.openai_api_key = None
            mock_settings.openai_rate_limit = 60
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.openai_model = "gpt-4o"
            mock_settings.openai_embedding_model = "text-embedding-3-small"
            mock_settings.openai_embedding_dimensions = 1536

            provider = OpenAIProvider()

            assert provider.name == "openai"
            assert provider.display_name == "OpenAI"
            assert "gpt-4o" in provider.models

    def test_openai_provider_not_available_without_key(self):
        """Test OpenAI provider is not available without API key."""
        with patch('src.providers.openai.settings') as mock_settings:
            mock_settings.openai_api_key = None
            mock_settings.openai_rate_limit = 60
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.openai_model = "gpt-4o"
            mock_settings.openai_embedding_model = "text-embedding-3-small"
            mock_settings.openai_embedding_dimensions = 1536

            provider = OpenAIProvider()

            assert provider.available is False

    def test_openai_get_provider_info(self):
        """Test getting provider info."""
        with patch('src.providers.openai.settings') as mock_settings:
            mock_settings.openai_api_key = None
            mock_settings.openai_rate_limit = 60
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.openai_model = "gpt-4o"
            mock_settings.openai_embedding_model = "text-embedding-3-small"
            mock_settings.openai_embedding_dimensions = 1536
            mock_settings.enable_streaming = True

            provider = OpenAIProvider()
            info = provider.get_provider_info()

            assert info.name == "openai"
            assert info.display_name == "OpenAI"
            assert info.available is False
            assert "gpt-4o" in info.models
            assert info.embedding_dimensions == 1536
            assert info.supports_embeddings is True


class TestOllamaProvider:
    """Tests for Ollama provider."""

    def test_ollama_provider_initialization(self):
        """Test Ollama provider initializes correctly."""
        with patch('src.providers.ollama.settings') as mock_settings:
            mock_settings.ollama_rate_limit = 100
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.ollama_base_url = "http://localhost:11434"
            mock_settings.ollama_model = "llama2"
            mock_settings.ollama_embedding_model = "nomic-embed-text"

            provider = OllamaProvider()

            assert provider.name == "ollama"
            assert provider.display_name == "Ollama (Local)"
            assert provider.base_url == "http://localhost:11434"
            assert provider.model == "llama2"

    def test_ollama_get_provider_info(self):
        """Test getting provider info."""
        with patch('src.providers.ollama.settings') as mock_settings:
            mock_settings.ollama_rate_limit = 100
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.ollama_base_url = "http://localhost:11434"
            mock_settings.ollama_model = "llama2"
            mock_settings.ollama_embedding_model = "nomic-embed-text"

            provider = OllamaProvider()
            info = provider.get_provider_info()

            assert info.name == "ollama"
            assert info.display_name == "Ollama (Local)"
            assert info.supports_streaming is True
            assert info.supports_embeddings is True


class TestCostCalculation:
    """Tests for cost calculation."""

    def test_ollama_cost_is_free(self):
        """Test Ollama cost is zero (free for local)."""
        with patch('src.providers.ollama.settings') as mock_settings:
            mock_settings.ollama_rate_limit = 100
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.ollama_base_url = "http://localhost:11434"
            mock_settings.ollama_model = "llama2"
            mock_settings.ollama_embedding_model = "nomic-embed-text"

            provider = OllamaProvider()
            cost = provider.calculate_cost("llama2", 100, 50)

            assert cost.estimated_cost == 0.0

    def test_openai_cost_calculation(self):
        """Test OpenAI cost calculation."""
        with patch('src.providers.openai.settings') as mock_settings:
            mock_settings.openai_api_key = None
            mock_settings.openai_rate_limit = 60
            mock_settings.max_retries = 3
            mock_settings.retry_base_delay = 1.0
            mock_settings.retry_max_delay = 30.0
            mock_settings.openai_model = "gpt-4o"
            mock_settings.openai_embedding_model = "text-embedding-3-small"
            mock_settings.openai_embedding_dimensions = 1536
            mock_settings.openai_input_cost = 0.0025
            mock_settings.openai_output_cost = 0.01

            provider = OpenAIProvider()
            cost = provider.calculate_cost("gpt-4o", 1000, 500)

            # Input: 1000 * 0.0025 / 1000 = 0.0025
            # Output: 500 * 0.01 / 1000 = 0.005
            # Total: 0.0075
            assert cost.estimated_cost == 0.0075
            assert cost.total_tokens == 1500


class TestProviderManager:
    """Tests for ProviderManager."""

    def test_provider_manager_initialization(self):
        """Test provider manager initializes providers."""
        with patch('src.providers.openai.settings') as mock_openai, \
             patch('src.providers.ollama.settings') as mock_ollama:

            # OpenAI settings
            mock_openai.openai_api_key = None
            mock_openai.openai_rate_limit = 60
            mock_openai.max_retries = 3
            mock_openai.retry_base_delay = 1.0
            mock_openai.retry_max_delay = 30.0
            mock_openai.openai_model = "gpt-4o"
            mock_openai.openai_embedding_model = "text-embedding-3-small"
            mock_openai.openai_embedding_dimensions = 1536

            # Ollama settings
            mock_ollama.ollama_rate_limit = 100
            mock_ollama.max_retries = 3
            mock_ollama.retry_base_delay = 1.0
            mock_ollama.retry_max_delay = 30.0
            mock_ollama.ollama_base_url = "http://localhost:11434"
            mock_ollama.ollama_model = "llama2"
            mock_ollama.ollama_embedding_model = "nomic-embed-text"

            from src.providers.manager import ProviderManager
            manager = ProviderManager()

            assert "openai" in manager._providers
            assert "ollama" in manager._providers
