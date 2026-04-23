"""Tests for traffic-based provider router."""

import pytest
from unittest.mock import MagicMock

from src.providers.router import (
    ProviderRouter,
    RoutingRule,
    init_provider_router,
    get_provider_router,
)
from src.models.schemas import (
    ProviderType,
    TrafficType,
    ChatMessage,
    MessageRole,
    CompletionRequest,
    EmbeddingRequest,
)
from src.providers.base import BaseProvider


class MockProvider:
    """Mock provider for testing."""

    def __init__(self, name: str, available: bool = True, supports_embeddings: bool = False):
        self.name = name
        self.available = available
        self.supports_embeddings = supports_embeddings


def _make_providers(**overrides) -> dict[str, MockProvider]:
    """Create a dict of mock providers."""
    defaults = {
        "groq": MockProvider("groq", available=True),
        "fireworks": MockProvider("fireworks", available=True, supports_embeddings=True),
        "deepinfra": MockProvider("deepinfra", available=True, supports_embeddings=True),
        "openrouter": MockProvider("openrouter", available=True),
        "openai": MockProvider("openai", available=True, supports_embeddings=True),
        "together": MockProvider("together", available=True, supports_embeddings=True),
        "anthropic": MockProvider("anthropic", available=True),
    }
    defaults.update(overrides)
    return defaults


class TestProviderRouterInit:
    """Tests for ProviderRouter initialization."""

    def test_init_with_default_rules(self):
        """Router should use default routing rules."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        table = router.get_routing_table()
        assert "realtime" in table
        assert "structured" in table
        assert "background" in table
        assert "function_calling" in table
        assert "general" in table

    def test_init_with_custom_rules(self):
        """Router should accept custom rules."""
        providers = _make_providers()
        custom_rules = {
            TrafficType.REALTIME: RoutingRule(
                traffic_type=TrafficType.REALTIME,
                primary_provider=ProviderType.OPENAI,
                fallback_providers=[],
            )
        }
        router = ProviderRouter(providers, rules=custom_rules)

        table = router.get_routing_table()
        assert table["realtime"]["primary"]["provider"] == "openai"


class TestClassifyTraffic:
    """Tests for traffic classification."""

    def test_explicit_hint_overrides_classification(self):
        """Explicit hint should be returned regardless of content."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        request = CompletionRequest(messages=[ChatMessage(role=MessageRole.USER, content="hello")])
        result = router.classify_traffic(request, hint=TrafficType.BACKGROUND)
        assert result == TrafficType.BACKGROUND

    def test_short_messages_classified_as_realtime(self):
        """Short messages should be classified as realtime."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        request = CompletionRequest(messages=[ChatMessage(role=MessageRole.USER, content="hello")])
        result = router.classify_traffic(request)
        assert result == TrafficType.REALTIME

    def test_function_keywords_classified_as_function_calling(self):
        """Messages with function/tool keywords should be classified as function_calling."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        request = CompletionRequest(
            messages=[
                ChatMessage(
                    role=MessageRole.USER,
                    content="Please invoke the function to execute this tool call with the given parameters",
                )
            ]
        )
        result = router.classify_traffic(request)
        assert result == TrafficType.FUNCTION_CALLING

    def test_stop_with_json_classified_as_structured(self):
        """Stop sequences containing 'json' should be classified as structured."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        request = CompletionRequest(
            messages=[ChatMessage(role=MessageRole.USER, content="Give me data")],
            stop=["json_output"],
        )
        result = router.classify_traffic(request)
        assert result == TrafficType.STRUCTURED

    def test_long_messages_classified_as_general(self):
        """Long messages without special keywords should be general."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        content = "This is a long message " * 20  # > 200 chars, > 2 messages
        request = CompletionRequest(
            messages=[
                ChatMessage(role=MessageRole.USER, content=content),
                ChatMessage(role=MessageRole.ASSISTANT, content="response"),
                ChatMessage(role=MessageRole.USER, content=content),
            ]
        )
        result = router.classify_traffic(request)
        assert result == TrafficType.GENERAL


class TestGetProviderForTraffic:
    """Tests for provider selection by traffic type."""

    def test_realtime_routes_to_groq(self):
        """Realtime traffic should route to Groq."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        provider, provider_type = router.get_provider_for_traffic(TrafficType.REALTIME)
        assert provider_type == ProviderType.GROQ

    def test_structured_routes_to_fireworks(self):
        """Structured traffic should route to Fireworks."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        provider, provider_type = router.get_provider_for_traffic(TrafficType.STRUCTURED)
        assert provider_type == ProviderType.FIREWORKS

    def test_background_routes_to_deepinfra(self):
        """Background traffic should route to DeepInfra."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        provider, provider_type = router.get_provider_for_traffic(TrafficType.BACKGROUND)
        assert provider_type == ProviderType.DEEPINFRA

    def test_fallback_when_primary_unavailable(self):
        """Should fall back to next provider when primary is unavailable."""
        providers = _make_providers(groq=MockProvider("groq", available=False))
        router = ProviderRouter(providers)

        provider, provider_type = router.get_provider_for_traffic(TrafficType.REALTIME)
        # Should fall back to Fireworks (second in REALTIME fallback list)
        assert provider_type == ProviderType.FIREWORKS

    def test_raises_when_no_provider_available(self):
        """Should raise ValueError when no providers available."""
        providers = {
            k: MockProvider(k, available=False)
            for k in ["groq", "fireworks", "openrouter", "openai"]
        }
        router = ProviderRouter(providers)

        with pytest.raises(ValueError, match="No suitable provider"):
            router.get_provider_for_traffic(TrafficType.REALTIME)

    def test_embeddings_require_support(self):
        """Embedding routing should skip providers without embedding support."""
        providers = _make_providers(
            deepinfra=MockProvider("deepinfra", available=True, supports_embeddings=False),
            fireworks=MockProvider("fireworks", available=True, supports_embeddings=True),
        )
        router = ProviderRouter(providers)

        provider, provider_type = router.get_provider_for_traffic(
            TrafficType.BACKGROUND, require_embeddings=True
        )
        # DeepInfra is primary for BACKGROUND but lacks embedding support, should skip to Together
        assert provider.supports_embeddings is True


class TestRouteCompletion:
    """Tests for async route_completion."""

    @pytest.mark.asyncio
    async def test_route_completion_returns_provider_and_type(self):
        """route_completion should return provider, type, and model override."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        request = CompletionRequest(messages=[ChatMessage(role=MessageRole.USER, content="hello")])
        provider, provider_type, model_override = await router.route_completion(request)

        assert provider is not None
        assert provider_type == ProviderType.GROQ
        assert model_override is None


class TestRouteEmbedding:
    """Tests for async route_embedding."""

    @pytest.mark.asyncio
    async def test_route_embedding_routes_to_background(self):
        """Embedding requests should route to background traffic type."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        request = EmbeddingRequest(text="test embedding")
        provider, provider_type = await router.route_embedding(request)

        assert provider is not None
        assert provider.supports_embeddings is True


class TestUpdateRule:
    """Tests for runtime rule updates."""

    def test_update_existing_rule(self):
        """Should update an existing routing rule."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        router.update_rule(
            TrafficType.REALTIME,
            primary_provider=ProviderType.OPENAI,
        )

        table = router.get_routing_table()
        assert table["realtime"]["primary"]["provider"] == "openai"

    def test_update_creates_new_rule(self):
        """Should create a rule for a traffic type that doesn't exist yet."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        # Use a custom traffic type by passing a new rule
        custom_rules = {}
        router2 = ProviderRouter(providers, rules=custom_rules)
        router2.update_rule(
            TrafficType.REALTIME,
            primary_provider=ProviderType.GROQ,
            fallback_providers=[ProviderType.OPENAI],
        )

        table = router2.get_routing_table()
        assert "realtime" in table
        assert table["realtime"]["primary"]["provider"] == "groq"


class TestGetRoutingTable:
    """Tests for routing table output."""

    def test_routing_table_has_all_traffic_types(self):
        """Routing table should include all default traffic types."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        table = router.get_routing_table()

        for tt in ["realtime", "structured", "function_calling", "background", "general"]:
            assert tt in table
            assert "primary" in table[tt]
            assert "fallbacks" in table[tt]
            assert "description" in table[tt]

    def test_routing_table_shows_availability(self):
        """Routing table should correctly show provider availability."""
        providers = _make_providers(groq=MockProvider("groq", available=False))
        router = ProviderRouter(providers)

        table = router.get_routing_table()
        assert table["realtime"]["primary"]["available"] is False


class TestGetRecommendations:
    """Tests for FunctionFly recommendations."""

    def test_recommendations_returns_list(self):
        """Should return a non-empty list of recommendations."""
        providers = _make_providers()
        router = ProviderRouter(providers)

        recs = router.get_recommendations()
        assert len(recs) > 0
        for rec in recs:
            assert "use_case" in rec
            assert "provider" in rec
            assert "reason" in rec


class TestInitAndGetProviderRouter:
    """Tests for global router initialization."""

    def test_init_and_get_provider_router(self):
        """init_provider_router should create a global instance."""
        providers = _make_providers()
        router = init_provider_router(providers)

        assert router is get_provider_router()

    def test_get_provider_router_raises_before_init(self):
        """get_provider_router should raise RuntimeError before init."""
        import src.providers.router as router_module

        old = router_module._provider_router
        router_module._provider_router = None
        try:
            with pytest.raises(RuntimeError):
                get_provider_router()
        finally:
            router_module._provider_router = old
