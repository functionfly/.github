"""Tests for model router and tier escalation logic."""

import sys
import pytest

# Import directly from the module to avoid pulling in the generation package init
# which has heavy dependencies (redis, etc.)
from src.services.generation.model_router import (
    ModelTier,
    ModelConfig,
    ModelRouter,
    ComplexityAnalyzer,
    RoutingDecision,
    TIER_MODELS,
)
from src.models.schemas import ProviderType


class TestComplexityAnalyzer:
    """Tests for ComplexityAnalyzer."""

    def test_simple_description_returns_cheap(self):
        """Simple descriptions should route to CHEAP tier."""
        tier, confidence = ComplexityAnalyzer.analyze(
            "Create a simple hello world function that validates a string"
        )
        assert tier == ModelTier.CHEAP

    def test_complex_description_returns_premium(self):
        """Complex descriptions should route to PREMIUM tier."""
        tier, confidence = ComplexityAnalyzer.analyze(
            "Build a distributed machine learning pipeline with orchestrate multi-step workflow"
        )
        assert tier == ModelTier.PREMIUM
        assert confidence >= 0.7

    def test_moderate_description_returns_mid(self):
        """Moderate descriptions should route to MID tier."""
        tier, confidence = ComplexityAnalyzer.analyze(
            "Create an API integration that processes requests"
        )
        assert tier == ModelTier.MID

    def test_constraints_affect_complexity(self):
        """Constraints text should be included in complexity analysis."""
        tier, _ = ComplexityAnalyzer.analyze(
            "Parse a file", constraints="Must use machine learning and real-time streaming"
        )
        assert tier == ModelTier.PREMIUM

    def test_explicit_simple_keyword_boosts_cheap(self):
        """Explicit 'simple' keyword should push toward CHEAP."""
        tier, _ = ComplexityAnalyzer.analyze("A simple basic function for parsing JSON")
        assert tier == ModelTier.CHEAP

    def test_explicit_very_complex_boosts_premium(self):
        """Explicit 'very complex' should push toward PREMIUM."""
        tier, _ = ComplexityAnalyzer.analyze("A very complex function for data processing")
        assert tier == ModelTier.PREMIUM


class TestModelRouterRouting:
    """Tests for ModelRouter.route()."""

    def test_route_returns_routing_decision(self):
        """route() should return a RoutingDecision."""
        router = ModelRouter()
        decision = router.route("Create a simple hello world function")

        assert isinstance(decision, RoutingDecision)
        assert decision.tier in ModelTier
        assert decision.model_config is not None
        assert decision.confidence > 0
        assert decision.estimated_cost_usd >= 0

    def test_route_with_preferred_tier(self):
        """route() should respect preferred_tier."""
        router = ModelRouter()
        decision = router.route("anything", preferred_tier=ModelTier.PREMIUM)

        assert decision.tier == ModelTier.PREMIUM
        assert decision.confidence == 0.9

    def test_route_escalates_when_no_cheap_models(self):
        """route() should escalate from CHEAP to MID when no cheap models available."""
        router = ModelRouter()
        router.update_provider_availability(
            {
                ProviderType.OLLAMA: False,
                ProviderType.OPENROUTER: True,
                ProviderType.OPENAI: True,
            }
        )

        decision = router.route("simple hello world function")
        assert decision.tier in (ModelTier.MID, ModelTier.PREMIUM)

    def test_route_escalates_from_mid_to_premium(self):
        """route() should escalate from MID to PREMIUM when mid models unavailable."""
        router = ModelRouter()
        router.update_provider_availability(
            {
                ProviderType.OLLAMA: False,
                ProviderType.OPENROUTER: False,
                ProviderType.OPENAI: False,
                ProviderType.ANTHROPIC: True,
            }
        )

        decision = router.route("api database integration")
        assert decision.tier == ModelTier.PREMIUM

    def test_route_raises_when_no_models_available(self):
        """route() should raise RuntimeError when all providers unavailable."""
        router = ModelRouter()
        router.update_provider_availability(
            {
                ProviderType.OLLAMA: False,
                ProviderType.OPENROUTER: False,
                ProviderType.OPENAI: False,
                ProviderType.ANTHROPIC: False,
            }
        )

        with pytest.raises(RuntimeError, match="No models available"):
            router.route("simple hello world function")


class TestModelRouterEscalation:
    """Tests for ModelRouter.should_escalate()."""

    def test_no_escalation_from_premium(self):
        """Should not escalate from PREMIUM tier."""
        router = ModelRouter()
        result = router.should_escalate(
            result="some code",
            validation_errors=["error"],
            current_tier=ModelTier.PREMIUM,
        )
        assert result is None

    def test_escalate_from_cheap_on_syntax_error(self):
        """Should escalate from CHEAP to MID on syntax errors."""
        router = ModelRouter()
        result = router.should_escalate(
            result="some code",
            validation_errors=["Syntax error at line 5: invalid syntax"],
            current_tier=ModelTier.CHEAP,
        )
        assert result == ModelTier.MID

    def test_escalate_from_mid_on_syntax_error(self):
        """Should escalate from MID to PREMIUM on syntax errors."""
        router = ModelRouter()
        result = router.should_escalate(
            result="some code",
            validation_errors=["Syntax error at line 5: invalid syntax"],
            current_tier=ModelTier.MID,
        )
        assert result == ModelTier.PREMIUM

    def test_escalate_on_short_result(self):
        """Should escalate when result is too short."""
        router = ModelRouter()
        result = router.should_escalate(
            result="hi",
            validation_errors=[],
            current_tier=ModelTier.CHEAP,
        )
        assert result == ModelTier.MID

    def test_escalate_on_multiple_errors(self):
        """Should escalate when 3+ validation errors."""
        router = ModelRouter()
        result = router.should_escalate(
            result="some longer code that is at least 100 chars" * 3,
            validation_errors=["error1", "error2", "error3"],
            current_tier=ModelTier.CHEAP,
        )
        assert result == ModelTier.MID

    def test_no_escalation_when_no_errors(self):
        """Should not escalate when no errors and sufficient code."""
        router = ModelRouter()
        result = router.should_escalate(
            result="def hello():\n    return 'world'\n" * 10,
            validation_errors=[],
            current_tier=ModelTier.CHEAP,
        )
        assert result is None


class TestModelRouterAvailability:
    """Tests for provider availability filtering."""

    def test_get_available_models_filters_unavailable(self):
        """get_available_models should filter out unavailable providers."""
        router = ModelRouter()
        router.update_provider_availability(
            {
                ProviderType.OLLAMA: False,
            }
        )

        cheap_models = router.get_available_models(ModelTier.CHEAP)
        assert all(m.provider != ProviderType.OLLAMA for m in cheap_models)

    def test_get_available_models_returns_all_when_available(self):
        """get_available_models should return all models when all providers available."""
        router = ModelRouter()
        router.update_provider_availability(
            {
                ProviderType.OLLAMA: True,
                ProviderType.OPENROUTER: True,
                ProviderType.OPENAI: True,
                ProviderType.ANTHROPIC: True,
            }
        )

        for tier in ModelTier:
            models = router.get_available_models(tier)
            assert len(models) > 0


class TestTierModelsConfig:
    """Tests for TIER_MODELS configuration."""

    def test_all_tiers_have_models(self):
        """Each tier should have at least one model configured."""
        for tier in ModelTier:
            assert len(TIER_MODELS[tier]) > 0

    def test_cheap_models_are_free(self):
        """CHEAP tier models should have zero cost."""
        for model in TIER_MODELS[ModelTier.CHEAP]:
            assert model.cost_per_1k_input == 0.0
            assert model.cost_per_1k_output == 0.0

    def test_premium_models_cost_more_than_mid(self):
        """PREMIUM tier models should cost more than MID tier."""
        mid_avg = sum(m.cost_per_1k_input for m in TIER_MODELS[ModelTier.MID]) / len(
            TIER_MODELS[ModelTier.MID]
        )
        premium_avg = sum(m.cost_per_1k_input for m in TIER_MODELS[ModelTier.PREMIUM]) / len(
            TIER_MODELS[ModelTier.PREMIUM]
        )
        assert premium_avg > mid_avg
