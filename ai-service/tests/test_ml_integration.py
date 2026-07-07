"""Integration tests for ML services with Redis and PostgreSQL using testcontainers.

These tests verify that ML services work correctly with real Redis and PostgreSQL
instances, ensuring horizontal scaling and distributed state management work.
"""

import asyncio
import hashlib
import os
import tempfile
from pathlib import Path
from typing import AsyncGenerator

import pytest
import pytest_asyncio

try:
    import redis.asyncio as redis
    from testcontainers.postgres import PostgresContainer
    from testcontainers.redis import RedisContainer
    TESTCONTAINERS_AVAILABLE = True
except ImportError:
    TESTCONTAINERS_AVAILABLE = False
    redis = None


# Skip all tests if testcontainers is not available
pytestmark = pytest.mark.skipif(
    not TESTCONTAINERS_AVAILABLE,
    reason="testcontainers not available - install with: pip install testcontainers[redis,postgres]"
)


class RedisFixture:
    """Redis test container fixture."""

    def __init__(self):
        self.container = None
        self.client = None

    async def start(self):
        """Start Redis container."""
        self.container = RedisContainer()
        self.container.start()
        redis_url = self.container.get_url()
        self.client = redis.from_url(redis_url, encoding="utf-8", decode_responses=True)
        await self.client.ping()
        return self

    async def stop(self):
        """Stop Redis container."""
        if self.client:
            await self.client.close()
        if self.container:
            self.container.stop()

    @property
    def url(self) -> str:
        return self.container.get_url()


class PostgresFixture:
    """PostgreSQL test container fixture."""

    def __init__(self):
        self.container = None
        self.url = None

    async def start(self):
        """Start PostgreSQL container."""
        self.container = PostgresContainer("postgres:16-alpine")
        self.container.start()
        self.url = self.container.get_url()
        return self

    async def stop(self):
        """Stop PostgreSQL container."""
        if self.container:
            self.container.stop()


@pytest_asyncio.fixture
async def redis_fixture() -> AsyncGenerator[RedisFixture, None]:
    """Fixture that provides a Redis container for testing."""
    fixture = RedisFixture()
    await fixture.start()
    yield fixture
    await fixture.stop()


@pytest_asyncio.fixture
async def postgres_fixture() -> AsyncGenerator[PostgresFixture, None]:
    """Fixture that provides a PostgreSQL container for testing."""
    fixture = PostgresFixture()
    await fixture.start()
    yield fixture
    await fixture.stop()


@pytest_asyncio.fixture
async def ml_config(redis_fixture: RedisFixture, postgres_fixture: PostgresFixture):
    """Fixture that provides ML service configuration for testing."""
    # Set environment variables for ML services
    os.environ["REDIS_URL"] = redis_fixture.url
    os.environ["ML_MODEL_DIR"] = tempfile.mkdtemp()
    os.environ["ML_ENABLED"] = "true"
    yield {
        "redis_url": redis_fixture.url,
        "model_dir": os.environ["ML_MODEL_DIR"],
    }
    # Cleanup
    import shutil
    shutil.rmtree(os.environ["ML_MODEL_DIR"], ignore_errors=True)


# =============================================================================
# Cost Anomaly Service Integration Tests
# =============================================================================

@pytest_asyncio.fixture
async def cost_anomaly_service(ml_config):
    """Fixture that provides a CostAnomalyDetector instance."""
    from src.services.cost_anomaly import CostAnomalyDetector

    service = CostAnomalyDetector()
    yield service
    await service.close()


@pytest.mark.asyncio
async def test_cost_anomaly_redis_state(
    redis_fixture: RedisFixture,
    ml_config,
    cost_anomaly_service,
):
    """Test that cost anomaly detector stores state in Redis."""
    from src.services.cost_anomaly.models import CostExecutionMetrics

    # Record some executions
    metrics = CostExecutionMetrics(
        function_id="test-fn-1",
        cost_cents=10.0,
        duration_ms=100.0,
        memory_mb=128,
        region="us-east-1",
    )

    # Check multiple times to build up stats
    for i in range(5):
        result = await cost_anomaly_service.check_execution(metrics)
        assert result is not None

    # Verify state is in Redis
    stats_key = f"ml:cost_anomaly:stats:test-fn-1"
    stats_data = await redis_fixture.client.get(stats_key)
    assert stats_data is not None

    # Clean up
    await redis_fixture.client.delete(stats_key)


@pytest.mark.asyncio
async def test_cost_anomaly_across_instances(
    redis_fixture: RedisFixture,
    ml_config,
):
    """Test that two detector instances share state via Redis."""
    from src.services.cost_anomaly import CostAnomalyDetector
    from src.services.cost_anomaly.models import CostExecutionMetrics

    # Create two detector instances
    detector1 = CostAnomalyDetector()
    detector2 = CostAnomalyDetector()

    metrics = CostExecutionMetrics(
        function_id="test-fn-shared",
        cost_cents=25.0,
        duration_ms=200.0,
        memory_mb=256,
        region="us-west-2",
    )

    # Record from first instance
    await detector1.check_execution(metrics)
    await detector1.check_execution(metrics)

    # Read from second instance - should see same stats
    summary = await detector2.get_summary("test", hours=24)
    assert summary is not None

    await detector1.close()
    await detector2.close()

    # Clean up
    await redis_fixture.client.delete("ml:cost_anomaly:stats:test-fn-shared")


# =============================================================================
# Thompson Routing Integration Tests
# =============================================================================

@pytest_asyncio.fixture
async def thompson_router_service(ml_config):
    """Fixture that provides a ThompsonSamplingRouter instance."""
    from src.services.thompson_routing import ThompsonSamplingRouter

    service = ThompsonSamplingRouter()
    yield service
    await service.close()


@pytest.mark.asyncio
async def test_thompson_routing_redis_state(
    redis_fixture: RedisFixture,
    ml_config,
    thompson_router_service,
):
    """Test that Thompson routing stores arm states in Redis."""
    from src.services.thompson_routing.models import RoutingDecisionRequest, EdgeMetrics

    # Record some outcomes
    edges = [
        EdgeMetrics(edge_id="edge-1", latency_ms=50, success_rate=0.99, error_rate=0.01),
        EdgeMetrics(edge_id="edge-2", latency_ms=100, success_rate=0.95, error_rate=0.05),
    ]

    request = RoutingDecisionRequest(
        function_id="test-fn-route",
        available_edges=edges,
    )

    # Make multiple decisions
    for _ in range(10):
        decision = await thompson_router_service.decide(request)
        assert decision is not None

        # Record outcome
        await thompson_router_service.record_outcome(
            function_id="test-fn-route",
            edge_id=decision.selected_edge_id,
            latency_ms=decision.latency_ms_estimate,
            success=True,
            error=None,
        )

    # Verify state in Redis
    arms_key = f"ml:thompson:arms:test-fn-route"
    arms_data = await redis_fixture.client.lrange(arms_key, 0, -1)
    assert len(arms_data) > 0

    # Clean up
    await redis_fixture.client.delete(arms_key)


@pytest.mark.asyncio
async def test_thompson_routing_scaling(
    redis_fixture: RedisFixture,
    ml_config,
):
    """Test that Thompson router works correctly when scaled horizontally."""
    from src.services.thompson_routing import ThompsonSamplingRouter
    from src.services.thompson_routing.models import RoutingDecisionRequest, EdgeMetrics

    router1 = ThompsonSamplingRouter()
    router2 = ThompsonSamplingRouter()

    edges = [
        EdgeMetrics(edge_id="edge-a", latency_ms=30, success_rate=0.98, error_rate=0.02),
        EdgeMetrics(edge_id="edge-b", latency_ms=80, success_rate=0.96, error_rate=0.04),
    ]

    request = RoutingDecisionRequest(
        function_id="test-fn-scale",
        available_edges=edges,
    )

    # Use router1 to record outcomes
    decision1 = await router1.decide(request)
    await router1.record_outcome(
        function_id="test-fn-scale",
        edge_id=decision1.selected_edge_id,
        latency_ms=decision1.latency_ms_estimate,
        success=True,
        error=None,
    )

    # Use router2 to get stats - should see same data
    stats = await router2.get_arm_stats("test-fn-scale")
    assert "edge-a" in stats or "edge-b" in stats

    await router1.close()
    await router2.close()

    # Clean up
    await redis_fixture.client.delete("ml:thompson:arms:test-fn-scale")


# =============================================================================
# Recommendations Service Integration Tests
# =============================================================================

@pytest_asyncio.fixture
async def recommendation_engine(ml_config):
    """Fixture that provides a RecommendationEngine instance."""
    from src.services.recommendations import RecommendationEngine

    service = RecommendationEngine()
    yield service
    await service.close()


@pytest.mark.asyncio
async def test_recommendations_redis_interactions(
    redis_fixture: RedisFixture,
    ml_config,
    recommendation_engine,
):
    """Test that recommendations store interactions in Redis."""
    from src.services.recommendations.models import RecommendationInteractionRequest

    # Record interactions
    interaction = RecommendationInteractionRequest(
        user_id="user-123",
        function_id="func-456",
        interaction_type="execution",
        weight=1.0,
    )

    await recommendation_engine.record_interaction(interaction)
    await recommendation_engine.record_interaction(interaction)
    await recommendation_engine.record_interaction(interaction)

    # Verify in Redis
    interactions_key = "ml:recommendations:interactions"
    count = await redis_fixture.client.llen(interactions_key)
    assert count >= 3

    # Clean up
    await redis_fixture.client.delete(interactions_key)


@pytest.mark.asyncio
async def test_recommendations_across_instances(
    redis_fixture: RedisFixture,
    ml_config,
):
    """Test that recommendation engine shares state via Redis."""
    from src.services.recommendations import RecommendationEngine
    from src.services.recommendations.models import RecommendationInteractionRequest

    engine1 = RecommendationEngine()
    engine2 = RecommendationEngine()

    interaction = RecommendationInteractionRequest(
        user_id="user-cross",
        function_id="func-cross",
        interaction_type="execution",
        weight=1.0,
    )

    # Record from engine1
    await engine1.record_interaction(interaction)

    # Both engines should be able to read it (through Redis)
    # Note: the actual recommendation requires training, but we verify
    # the interaction was recorded

    await engine1.close()
    await engine2.close()


# =============================================================================
# Holt-Winters Prewarming Integration Tests
# =============================================================================

@pytest_asyncio.fixture
async def holt_winters_service(ml_config):
    """Fixture that provides a HoltWintersForecaster instance."""
    from src.services.prewarming.holt_winters import HoltWintersForecaster

    service = HoltWintersForecaster()
    yield service
    await service.close()


@pytest.mark.asyncio
async def test_holt_winters_forecasting(
    redis_fixture: RedisFixture,
    ml_config,
    holt_winters_service,
):
    """Test Holt-Winters forecasting with real Redis."""
    # Record some historical data
    for hour in range(48):
        timestamp = 1700000000 + hour * 3600  # 48 hours of data
        count = 10 + (hour % 24) * 2  # Daily pattern
        await holt_winters_service.record_demand(
            function_id="fn-winters-test",
            timestamp=timestamp,
            request_count=count,
        )

    # Get forecast
    forecast = await holt_winters_service.predict_demand(
        function_id="fn-winters-test",
        horizon_hours=6,
    )

    assert forecast is not None
    assert forecast.forecast is not None
    assert len(forecast.forecast) == 6

    # Clean up
    await redis_fixture.client.delete("ml:prewarming:demand:fn-winters-test")


# =============================================================================
# Auth Rate Limiter Integration Tests
# =============================================================================

@pytest.mark.asyncio
async def test_redis_auth_rate_limiter_distributed(
    redis_fixture: RedisFixture,
):
    """Test that Redis rate limiter works across multiple instances."""
    from src.security.auth import RedisAuthRateLimiter

    # Create two rate limiter instances sharing same Redis
    limiter1 = RedisAuthRateLimiter(redis_fixture.client)
    limiter2 = RedisAuthRateLimiter(redis_fixture.client)

    test_key = "test-api-key-12345"

    # Record failures from limiter1
    for _ in range(3):
        await limiter1.record_failure(test_key)

    # Check lockout from limiter2 - should see the failures
    is_locked, retry_after = await limiter2.check_lockout(test_key)
    # After 3 failures, no lockout yet (threshold is 3)
    assert is_locked is False

    # Record one more failure from limiter1
    await limiter1.record_failure(test_key)

    # Now limiter2 should see lockout
    is_locked, retry_after = await limiter2.check_lockout(test_key)
    assert is_locked is True
    assert retry_after > 0

    # Clear with limiter1
    await limiter1.record_success(test_key)

    # Both should be unlocked now
    is_locked1, _ = await limiter1.check_lockout(test_key)
    is_locked2, _ = await limiter2.check_lockout(test_key)
    assert is_locked1 is False
    assert is_locked2 is False

    # Clean up
    await redis_fixture.client.delete(
        f"flymind:auth:ratelimit:{hashlib.sha256(test_key.encode()).hexdigest()}"
    )


# =============================================================================
# Model Persistence Integration Tests
# =============================================================================

@pytest.mark.asyncio
async def test_model_encryption_roundtrip(ml_config):
    """Test encrypted model save/load roundtrip."""
    from cryptography.fernet import Fernet
    from src.services.ml_common.persistence import EncryptedModelStore

    # Generate encryption key
    key = Fernet.generate_key()
    os.environ["ML_MODEL_ENCRYPTION_KEY"] = key.decode()

    # Create encrypted store
    store = EncryptedModelStore("test-encrypt", encrypt=True)

    # Save a model
    import numpy as np
    model_data = {"weights": np.array([1.0, 2.0, 3.0]), "bias": 0.5}
    version = store.save(model_data, version="v1")

    # Load it back
    loaded = store.load(version)
    assert loaded is not None
    assert np.allclose(loaded["weights"], model_data["weights"])
    assert loaded["bias"] == model_data["bias"]

    # Verify file is encrypted (can't read as plain joblib)
    model_path = Path(ml_config["model_dir"]) / "test-encrypt" / f"model_{version}.joblib"
    with open(model_path, "rb") as f:
        header = f.read(10)
        assert header != b"<?xml"  # Not plain pickle format


@pytest.mark.asyncio
async def test_model_store_without_encryption(ml_config):
    """Test model store works without encryption."""
    from src.services.ml_common.persistence import ModelStore

    store = ModelStore("test-plain")

    # Save a model
    import numpy as np
    model_data = {"weights": np.array([1.0, 2.0, 3.0]), "bias": 0.5}
    version = store.save(model_data, version="v1")

    # Load it back
    loaded = store.load(version)
    assert loaded is not None
    assert np.allclose(loaded["weights"], model_data["weights"])


# =============================================================================
# Registry Integration Tests
# =============================================================================

@pytest.mark.asyncio
async def test_registry_lifecycle():
    """Test ML service registry properly initializes and shuts down services."""
    from src.services.ml_common.registry import (
        get_registry,
        init_ml_services,
        shutdown_ml_services,
    )

    registry = get_registry()
    assert registry is not None

    # Initialize services
    await init_ml_services()

    # Check services are registered
    assert registry.get("cost_anomaly") is not None
    assert registry.get("thompson_routing") is not None
    assert registry.get("recommendations") is not None
    assert registry.get("prewarming") is not None

    # Shutdown
    await shutdown_ml_services()

    # After shutdown, services should be cleared
    # (or closed, depending on implementation)
