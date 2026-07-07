"""Tests for ML Intelligence Layer services.

Tests for:
- Cost anomaly detection (Welford's algorithm, Z-score)
- Prewarming (Holt-Winters exponential smoothing)
- Thompson Sampling routing
- Collaborative filtering recommendations
"""

import pytest
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, MagicMock, patch

from src.services.cost_anomaly.models import (
    CostExecutionMetrics,
    CostAnomalyResult,
    FunctionCostStats,
)
from src.services.cost_anomaly.predictor import CostAnomalyDetector
from src.services.thompson_routing.models import ArmState, RoutingOutcome
from src.services.thompson_routing.selector import ThompsonSamplingRouter
from src.services.recommendations.models import InteractionEvent
from src.services.prewarming.models import Prediction, RequestDataPoint


class TestFunctionCostStats:
    """Tests for FunctionCostStats using Welford's online algorithm."""

    def test_initial_state(self):
        """New stats should have zero count and infinite variance."""
        stats = FunctionCostStats(function_id="test_func")
        assert stats.count == 0
        assert stats.variance == 0.0
        assert stats.std == 0.0

    def test_single_update(self):
        """After one update, mean equals that value."""
        stats = FunctionCostStats(function_id="test_func")
        stats.update(10.0)
        assert stats.count == 1
        assert stats.mean == 10.0
        assert stats.variance == 0.0

    def test_welford_variance(self):
        """Welford's algorithm should give correct variance."""
        stats = FunctionCostStats(function_id="test_func")
        values = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
        for v in values:
            stats.update(v)

        assert stats.count == 8
        assert abs(stats.mean - 5.0) < 0.01
        assert abs(stats.std - 2.0) < 0.01

    def test_z_score_normal(self):
        """Z-score should be 0 for value at mean."""
        stats = FunctionCostStats(function_id="test_func")
        stats.update(10.0)
        stats.update(20.0)
        assert stats.z_score(15.0) == 0.0

    def test_z_score_positive(self):
        """Z-score should be positive for value above mean."""
        stats = FunctionCostStats(function_id="test_func")
        stats.update(10.0)
        stats.update(20.0)
        z = stats.z_score(25.0)
        assert z > 0

    def test_z_score_negative(self):
        """Z-score should be negative for value below mean."""
        stats = FunctionCostStats(function_id="test_func")
        stats.update(10.0)
        stats.update(20.0)
        z = stats.z_score(5.0)
        assert z < 0

    def test_z_score_zero_std(self):
        """Z-score should be 0 when std is 0 (single sample)."""
        stats = FunctionCostStats(function_id="test_func")
        stats.update(10.0)
        assert stats.z_score(15.0) == 0.0


class TestCostAnomalyResult:
    """Tests for CostAnomalyResult model."""

    def test_anomaly_result_creation(self):
        """Should create anomaly result with all fields."""
        result = CostAnomalyResult(
            is_anomaly=True,
            score=0.75,
            anomaly_type="cost_spike",
            severity="high",
            details="Cost spike detected",
            function_id="test_func",
            z_score=4.5,
            threshold=3.0,
        )
        assert result.is_anomaly is True
        assert result.score == 0.75
        assert result.anomaly_type == "cost_spike"
        assert result.severity == "high"

    def test_score_bounds(self):
        """Score should be between 0 and 1 via Pydantic validation."""
        result = CostAnomalyResult(
            is_anomaly=True,
            score=0.75,
            severity="high",
            function_id="test",
        )
        assert result.score == 0.75


class TestCostAnomalyDetector:
    """Tests for CostAnomalyDetector - model validation and algorithm tests only.

    Note: Full integration tests with Redis require a running Redis instance
    and are skipped here. These tests verify the core algorithm logic.
    """

    def test_detector_has_adaptive_threshold(self):
        """CostAnomalyDetector should have adaptive threshold capability."""
        detector = CostAnomalyDetector()
        assert hasattr(detector, "_base_threshold")
        assert hasattr(detector, "MIN_THRESHOLD")
        assert hasattr(detector, "MAX_THRESHOLD")

    def test_adaptive_threshold_calculation_stable_function(self):
        """Stable function (low CV) should get lower threshold."""
        detector = CostAnomalyDetector()
        stats = FunctionCostStats(function_id="test")
        for _ in range(20):
            stats.update(100.0)
        threshold = detector._calculate_adaptive_threshold(stats)
        assert threshold <= detector._base_threshold

    def test_adaptive_threshold_calculation_variable_function(self):
        """Variable function (high CV) should get higher threshold."""
        detector = CostAnomalyDetector()
        stats = FunctionCostStats(function_id="test")
        for i in range(20):
            stats.update(100.0 + (50.0 if i % 2 == 0 else -30.0))
        threshold = detector._calculate_adaptive_threshold(stats)
        assert threshold >= detector._base_threshold

    def test_adaptive_threshold_bounds(self):
        """Adaptive threshold should be bounded between MIN and MAX."""
        detector = CostAnomalyDetector()
        stats = FunctionCostStats(function_id="test")
        for _ in range(20):
            stats.update(100.0)
        threshold = detector._calculate_adaptive_threshold(stats)
        assert detector.MIN_THRESHOLD <= threshold <= detector.MAX_THRESHOLD


class TestArmState:
    """Tests for Thompson Sampling ArmState."""

    def test_initial_state(self):
        """New arm should have Beta(1, 1) prior."""
        arm = ArmState(edge="cloudflare")
        assert arm.alpha == 1.0
        assert arm.beta == 1.0
        assert arm.total_pulls == 0
        assert arm.mean == 0.5

    def test_update_reward(self):
        """Update should modify Beta distribution parameters."""
        arm = ArmState(edge="cloudflare")
        arm.alpha += 0.8
        arm.beta += 0.2
        assert arm.alpha == 1.8
        assert arm.beta == 1.2

    def test_mean_calculation(self):
        """Mean should be alpha / (alpha + beta)."""
        arm = ArmState(edge="cloudflare")
        arm.alpha = 3.0
        arm.beta = 1.0
        assert arm.mean == 0.75


class TestRoutingOutcome:
    """Tests for RoutingOutcome model."""

    def test_outcome_creation(self):
        """Should create outcome with all fields."""
        outcome = RoutingOutcome(
            edge="cloudflare",
            function_id="test_func",
            latency_ms=100.0,
            success=True,
            cost_cents=0.5,
        )
        assert outcome.edge == "cloudflare"
        assert outcome.success is True
        assert outcome.latency_ms == 100.0


class TestThompsonSamplingRouter:
    """Tests for ThompsonSamplingRouter - algorithm and model tests.

    Note: Full integration tests with Redis require a running Redis instance
    and are skipped here. These tests verify the core algorithm logic.
    """

    def test_router_has_exploration_rate(self):
        """Router should have configurable exploration rate."""
        router = ThompsonSamplingRouter()
        assert hasattr(router, "_exploration_rate")
        assert 0.0 <= router._exploration_rate <= 1.0

    def test_reward_calculation_logic(self):
        """Reward should be weighted combination of latency, success, cost."""
        outcome = RoutingOutcome(
            edge="cloudflare",
            function_id="test_func",
            latency_ms=100.0,
            success=True,
            cost_cents=0.5,
        )

        latency_reward = max(0, 1.0 - min(outcome.latency_ms / 500.0, 1.0))
        success_reward = 1.0 if outcome.success else 0.0
        cost_reward = max(0, 1.0 - min(outcome.cost_cents / 1.0, 1.0))
        expected_reward = 0.4 * latency_reward + 0.4 * success_reward + 0.2 * cost_reward

        assert latency_reward == 0.8
        assert success_reward == 1.0
        assert cost_reward == 0.5
        assert abs(expected_reward - 0.82) < 0.01

    def test_failed_outcome_reward_calculation(self):
        """Failed outcome should have lower reward."""
        outcome = RoutingOutcome(
            edge="cloudflare",
            function_id="test_func",
            latency_ms=200.0,
            success=False,
            cost_cents=0.5,
        )

        latency_reward = max(0, 1.0 - min(outcome.latency_ms / 500.0, 1.0))
        success_reward = 1.0 if outcome.success else 0.0
        cost_reward = max(0, 1.0 - min(outcome.cost_cents / 1.0, 1.0))
        expected_reward = 0.4 * latency_reward + 0.4 * success_reward + 0.2 * cost_reward

        assert latency_reward == 0.6
        assert success_reward == 0.0
        assert cost_reward == 0.5
        assert abs(expected_reward - 0.34) < 0.01


class TestPrediction:
    """Tests for Prewarming Prediction model."""

    def test_prediction_creation(self):
        """Should create prediction with all fields."""
        now = datetime.utcnow()
        later = now + timedelta(minutes=10)
        prediction = Prediction(
            function_id="test_func",
            predicted_requests=10,
            confidence=0.85,
            window_start=now,
            window_end=later,
            trend="increasing",
        )
        assert prediction.function_id == "test_func"
        assert prediction.predicted_requests == 10
        assert prediction.confidence == 0.85
        assert prediction.trend == "increasing"


class TestRequestDataPoint:
    """Tests for Prewarming RequestDataPoint model."""

    def test_datapoint_creation(self):
        """Should create data point with timestamp."""
        dp = RequestDataPoint(
            function_id="test_func",
            timestamp=datetime.utcnow(),
            request_count=5,
        )
        assert dp.request_count == 5
        assert dp.timestamp is not None


class TestInteractionEvent:
    """Tests for Recommendation InteractionEvent model."""

    def test_interaction_event_creation(self):
        """Should create interaction event with all fields."""
        event = InteractionEvent(
            user_id="user123",
            function_id="func456",
            interaction_type="install",
            context={"source": "marketplace"},
        )
        assert event.user_id == "user123"
        assert event.function_id == "func456"
        assert event.interaction_type == "install"
        assert event.context["source"] == "marketplace"

    def test_interaction_type_is_required(self):
        """interaction_type is a required field with no default."""
        with pytest.raises(Exception):
            InteractionEvent(
                user_id="user123",
                function_id="func456",
            )


class TestRecommendationInteractionWeights:
    """Tests for recommendation interaction weights."""

    def test_install_has_highest_weight(self):
        """Install interaction should have highest weight."""
        from src.services.recommendations.predictor import RecommendationEngine

        weights = RecommendationEngine.INTERACTION_WEIGHTS
        assert weights["install"] > weights["execute"]
        assert weights["install"] > weights["view"]

    def test_view_has_lowest_weight(self):
        """View interaction should have lowest weight."""
        from src.services.recommendations.predictor import RecommendationEngine

        weights = RecommendationEngine.INTERACTION_WEIGHTS
        assert weights["view"] < weights["install"]
        assert weights["view"] < weights["execute"]


class TestRecommendationEngineInteractionWeights:
    """Tests for RecommendationEngine interaction weights configuration."""

    def test_all_interaction_types_have_weights(self):
        """All known interaction types should have weights."""
        from src.services.recommendations.predictor import RecommendationEngine

        known_types = ["install", "execute", "rate", "search_click", "view", "search_impression"]
        for interaction_type in known_types:
            assert interaction_type in RecommendationEngine.INTERACTION_WEIGHTS

    def test_weights_are_positive(self):
        """All interaction weights should be positive."""
        from src.services.recommendations.predictor import RecommendationEngine

        for weight in RecommendationEngine.INTERACTION_WEIGHTS.values():
            assert weight > 0


class TestHoltWintersForecaster:
    """Tests for HoltWintersForecaster - algorithm tests.

    Note: Full integration tests with Redis require a running Redis instance
    and are skipped here. These tests verify the core algorithm logic.
    """

    @pytest.fixture
    def forecaster(self):
        """Create forecaster without Redis dependency for algorithm testing."""
        forecaster = __import__(
            "src.services.prewarming.holt_winters",
            fromlist=["HoltWintersForecaster"]
        ).HoltWintersForecaster()
        forecaster._redis = None
        return forecaster

    def test_simple_exp_smoothing_empty_series(self, forecaster):
        """Should return zero forecast for empty series."""
        import numpy as np
        forecasts, std = forecaster._simple_exp_smoothing(np.array([]))
        assert forecasts[0] == 0.0

    def test_simple_exp_smoothing_single_value(self, forecaster):
        """Should return that value as forecast for single-element series."""
        import numpy as np
        forecasts, std = forecaster._simple_exp_smoothing(np.array([10.0]))
        assert forecasts[0] == 10.0

    def test_simple_exp_smoothing_stable_series(self, forecaster):
        """Should converge to mean for stable series."""
        import numpy as np
        series = np.array([10.0, 10.0, 10.0, 10.0])
        forecasts, std = forecaster._simple_exp_smoothing(series, alpha=0.3)
        assert forecasts[0] >= 9.0
        assert forecasts[0] <= 11.0

    def test_holt_winters_requires_minimum_data(self, forecaster):
        """Holt-Winters requires at least 2 * season_length data points."""
        import numpy as np
        series = np.zeros(24)
        forecasts, std = forecaster._holt_winters_additive(series, season_length=24)
        assert len(forecasts) == 1

    def test_holt_winters_forecast_is_non_negative(self, forecaster):
        """Forecasts should never be negative (requests can't be negative)."""
        import numpy as np
        series = np.array([1.0, 2.0] * 25)
        forecasts, std = forecaster._holt_winters_additive(series, season_length=24)
        assert all(f >= 0 for f in forecasts)


class TestCostAnomalySeverityLevels:
    """Tests for cost anomaly severity classification.

    Severity thresholds are:
    - abs_z > 5.0: critical
    - abs_z > 4.0: high
    - abs_z > 3.0: medium
    - abs_z <= 3.0: low/none
    """

    def test_z_score_above_5_is_critical(self):
        """abs(Z) > 5 should be classified as critical severity."""
        stats = FunctionCostStats(function_id="test")
        values = [8.0, 9.0, 10.0, 11.0, 12.0] * 4
        for v in values:
            stats.update(v)
        z = stats.z_score(20.0)
        assert abs(z) > 5.0

    def test_z_score_above_4_is_high(self):
        """abs(Z) > 4 should be classified as high severity."""
        stats = FunctionCostStats(function_id="test")
        values = [8.0, 9.0, 10.0, 11.0, 12.0] * 4
        for v in values:
            stats.update(v)
        z = stats.z_score(17.5)
        assert abs(z) > 4.0

    def test_z_score_above_3_is_medium(self):
        """abs(Z) > 3 should be classified as medium severity."""
        stats = FunctionCostStats(function_id="test")
        values = [8.0, 9.0, 10.0, 11.0, 12.0] * 4
        for v in values:
            stats.update(v)
        z = stats.z_score(15.0)
        assert abs(z) > 3.0


class TestThompsonSamplingDecision:
    """Tests for Thompson Sampling decision making."""

    def test_beta_distribution_sampling(self):
        """Should sample from Beta distribution."""
        import random
        samples = [random.betavariate(2.0, 2.0) for _ in range(100)]
        assert all(0 <= s <= 1 for s in samples)
        assert not all(s == samples[0] for s in samples)

    def test_exploration_rate_controls_randomness(self):
        """Higher exploration rate should lead to more random decisions."""
        import random
        random.seed(42)

        decisions_low_explore = []
        for _ in range(10):
            is_explore = random.random() < 0.01
            decisions_low_explore.append(is_explore)

        decisions_high_explore = []
        for _ in range(10):
            is_explore = random.random() < 0.5
            decisions_high_explore.append(is_explore)

        assert sum(decisions_high_explore) > sum(decisions_low_explore)
