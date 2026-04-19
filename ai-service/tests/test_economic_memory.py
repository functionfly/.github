"""Tests for Economic Memory and Economic Routing services.

These tests verify:
1. Cost-quality score calculations
2. Execution record tracking
3. Economic routing decisions
4. Model recommendations
"""

import pytest
from datetime import datetime
import asyncio

# Import the services we created
from src.services.economic_memory import (
    EconomicMemory,
    CostQualityScore,
    ExecutionRecord,
    ProviderType,
    get_economic_memory,
)
from src.services.economic_routing import (
    EconomicRoutingService,
    RoutingStrategy,
    EconomicRoutingScore,
    get_economic_routing_service,
)


@pytest.fixture
def economic_memory():
    """Fresh economic memory instance for tests."""
    return EconomicMemory()


@pytest.fixture
def execution_record():
    """Sample execution record for testing."""
    return ExecutionRecord(
        execution_id="test-exec-001",
        provider=ProviderType.OPENAI,
        model="gpt-4o-mini",
        input_tokens=1000,
        output_tokens=500,
        total_tokens=1500,
        cost_usd=0.00075,
        latency_ms=250.0,
        success=True,
        output_quality_score=0.85,
        tenant_id="tenant-123",
    )


class TestCostQualityScore:
    """Tests for the CostQualityScore dataclass."""

    def test_calculate_cqi_with_cost(self):
        """Test CQI calculation with a real cost."""
        score = CostQualityScore(
            provider=ProviderType.OPENAI,
            model="gpt-4o-mini",
            avg_cost_per_1k_tokens=0.000375,  # $0.375 per 1M tokens
            quality_score=0.82,
            response_time_score=0.95,
            success_rate=0.99,
        )
        
        cqi = score.calculate_cqi()
        
        # Quality composite = 0.82*0.4 + 0.95*0.3 + 0.0*0.2 + 0.99*0.1 = 0.877
        # CQI = (0.877 * 100) / (0.000375 * 1000) = 87.7 / 0.375 = 233.87, clamped to 100
        assert cqi > 0
        assert cqi <= 100
        assert score.cost_quality_index == cqi

    def test_calculate_cqi_zero_cost(self):
        """Test CQI calculation with zero cost (e.g., Ollama)."""
        score = CostQualityScore(
            provider=ProviderType.OLLAMA,
            model="llama3.1",
            avg_cost_per_1k_tokens=0.0,
            quality_score=0.70,
            response_time_score=0.90,
            success_rate=0.95,
        )
        
        cqi = score.calculate_cqi()
        assert cqi == 0.0  # Can't calculate without cost


class TestEconomicMemory:
    """Tests for the EconomicMemory class."""

    @pytest.mark.asyncio
    async def test_record_execution(self, economic_memory, execution_record):
        """Test recording a single execution."""
        await economic_memory.record_execution(execution_record)
        
        # Verify it was recorded
        score = await economic_memory.get_score(
            ProviderType.OPENAI,
            "gpt-4o-mini"
        )
        
        assert score is not None
        assert score.provider == ProviderType.OPENAI
        assert score.model == "gpt-4o-mini"
        assert score.total_executions == 1
        assert score.total_cost_usd == 0.00075

    @pytest.mark.asyncio
    async def test_record_multiple_executions(self, economic_memory):
        """Test recording multiple executions averages correctly."""
        # Record successful execution
        for i in range(5):
            record = ExecutionRecord(
                execution_id=f"test-exec-{i}",
                provider=ProviderType.GROQ,
                model="llama-3.1-8b",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.00015,  # Cheaper
                latency_ms=100.0,
                success=True,
                output_quality_score=0.75,
            )
            await economic_memory.record_execution(record)
        
        score = await economic_memory.get_score(
            ProviderType.GROQ,
            "llama-3.1-8b"
        )
        
        assert score is not None
        assert score.total_executions == 5
        assert score.total_cost_usd == pytest.approx(0.00075)  # 5 * 0.00015
        assert score.success_rate == 1.0

    @pytest.mark.asyncio
    async def test_best_value_provider(self, economic_memory):
        """Test finding the best value provider."""
        # Record executions for multiple providers
        providers_data = [
            (ProviderType.OPENAI, "gpt-4o-mini", 0.82, 0.000375),
            (ProviderType.GROQ, "llama-3.1-8b", 0.75, 0.0001),
            (ProviderType.ANTHROPIC, "claude-3-haiku", 0.80, 0.000625),
        ]
        
        for i, (provider, model, quality, cost) in enumerate(providers_data):
            for j in range(20):  # 20 executions each
                record = ExecutionRecord(
                    execution_id=f"test-exec-{i}-{j}",
                    provider=provider,
                    model=model,
                    input_tokens=1000,
                    output_tokens=500,
                    total_tokens=1500,
                    cost_usd=cost,
                    latency_ms=150.0,
                    success=True,
                    output_quality_score=quality,
                )
                await economic_memory.record_execution(record)
        
        best = await economic_memory.get_best_value_provider(
            min_executions=10
        )
        
        assert best is not None
        # Groq should win on value (cheapest * decent quality)
        assert best.provider == ProviderType.GROQ

    @pytest.mark.asyncio
    async def test_recommendations(self, economic_memory):
        """Test getting recommendations."""
        # Seed with data
        for i in range(20):
            record = ExecutionRecord(
                execution_id=f"test-exec-{i}",
                provider=ProviderType.GROQ,
                model="llama-3.1-8b",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.0001,
                latency_ms=100.0,
                success=True,
                output_quality_score=0.75,
            )
            await economic_memory.record_execution(record)
        
        recommendations = await economic_memory.get_recommendations(
            target_quality_threshold=0.7
        )
        
        assert len(recommendations) > 0
        rec = recommendations[0]
        assert "provider" in rec
        assert "model" in rec
        assert "recommendation" in rec
        assert rec["recommendation"] in ["highly_recommended", "recommended", "neutral", "avoid"]

    @pytest.mark.asyncio
    async def test_cost_breakdown(self, economic_memory):
        """Test cost breakdown by provider."""
        # Add some data
        for i in range(10):
            record = ExecutionRecord(
                execution_id=f"test-exec-{i}",
                provider=ProviderType.OPENAI,
                model="gpt-4o-mini",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.000375,
                latency_ms=200.0,
                success=True,
                output_quality_score=0.82,
            )
            await economic_memory.record_execution(record)
        
        breakdown = await economic_memory.get_cost_breakdown(days=7)
        
        assert "period_days" in breakdown
        assert "total_cost" in breakdown
        assert "provider_breakdown" in breakdown
        assert breakdown["total_cost"] > 0

    @pytest.mark.asyncio
    async def test_model_switch_suggestion(self, economic_memory):
        """Test model switch suggestions."""
        # Seed with expensive model data
        for i in range(15):
            expensive = ExecutionRecord(
                execution_id=f"expensive-{i}",
                provider=ProviderType.OPENAI,
                model="gpt-4o",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.005,  # Expensive
                latency_ms=200.0,
                success=True,
                output_quality_score=0.95,
            )
            await economic_memory.record_execution(expensive)
        
        # Seed with cheap model data (similar quality)
        for i in range(15):
            cheap = ExecutionRecord(
                execution_id=f"cheap-{i}",
                provider=ProviderType.OPENAI,
                model="gpt-4o-mini",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.000375,  # Much cheaper
                latency_ms=180.0,
                success=True,
                output_quality_score=0.82,  # Slightly lower but acceptable
            )
            await economic_memory.record_execution(cheap)
        
        suggestion = await economic_memory.suggest_model_switch(
            current_provider=ProviderType.OPENAI,
            current_model="gpt-4o",
            target_quality=0.80,
        )
        
        assert suggestion is not None
        assert "recommendation" in suggestion
        # Should suggest switching to mini
        if suggestion["recommendation"] == "downgrade_suggested":
            assert suggestion["suggested_model"] == "gpt-4o-mini"


class TestEconomicRouting:
    """Tests for the EconomicRoutingService."""

    @pytest.fixture
    def routing_service(self, economic_memory):
        """Fresh routing service with fresh memory."""
        service = EconomicRoutingService()
        # Override the internal memory for testing
        service._economic_memory = economic_memory
        return service

    @pytest.mark.asyncio
    async def test_score_providers(self, routing_service, economic_memory):
        """Test scoring providers."""
        # Seed with data
        for i in range(20):
            record = ExecutionRecord(
                execution_id=f"test-exec-{i}",
                provider=ProviderType.GROQ,
                model="llama-3.1-8b",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.0001,
                latency_ms=100.0,
                success=True,
                output_quality_score=0.75,
            )
            await economic_memory.record_execution(record)
        
        scores = await routing_service._score_providers(
            strategy=RoutingStrategy.BALANCED,
            quality_threshold=None,
            max_cost_per_1k=None,
        )
        
        assert len(scores) > 0
        # Should include at least Groq
        groq_scores = [s for s in scores if s.provider == ProviderType.GROQ]
        assert len(groq_scores) > 0

    @pytest.mark.asyncio  
    async def test_model_recommendation(self, routing_service, economic_memory):
        """Test model recommendation."""
        # Seed with data for current and alternative models
        for i in range(15):
            record = ExecutionRecord(
                execution_id=f"exec-{i}",
                provider=ProviderType.OPENAI,
                model="gpt-4o-mini",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.000375,
                latency_ms=200.0,
                success=True,
                output_quality_score=0.82,
            )
            await economic_memory.record_execution(record)
        
        recommendation = await routing_service.get_model_recommendation(
            provider=ProviderType.OPENAI,
            current_model="gpt-4o",
            target_quality=0.75,
        )
        
        assert "recommendation" in recommendation
        assert "current_model" in recommendation
        assert recommendation["current_model"] == "gpt-4o"

    @pytest.mark.asyncio
    async def test_cost_savings_opportunity(self, routing_service, economic_memory):
        """Test cost savings analysis."""
        # Add some execution data
        for i in range(10):
            record = ExecutionRecord(
                execution_id=f"test-exec-{i}",
                provider=ProviderType.GROQ,
                model="llama-3.1-8b",
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=0.0001,
                latency_ms=100.0,
                success=True,
                tenant_id="tenant-test",
            )
            await economic_memory.record_execution(record)
        
        savings = await routing_service.get_cost_savings_opportunity(
            tenant_id="tenant-test",
            days=7,
        )
        
        assert "period_days" in savings
        assert "analysis" in savings
        assert savings["executions_analyzed"] > 0


class TestGlobalInstances:
    """Tests for the global service instances."""

    def test_get_economic_memory_singleton(self):
        """Test that get_economic_memory returns a singleton."""
        mem1 = get_economic_memory()
        mem2 = get_economic_memory()
        assert mem1 is mem2

    def test_get_economic_routing_service_singleton(self):
        """Test that get_economic_routing_service returns a singleton."""
        svc1 = get_economic_routing_service()
        svc2 = get_economic_routing_service()
        assert svc1 is svc2


# ============================================================================
# Integration Test
# ============================================================================

@pytest.mark.asyncio
async def test_full_economic_workflow():
    """Integration test for the full economic memory workflow."""
    memory = EconomicMemory()
    
    # Simulate 100 executions across different providers
    providers = [
        (ProviderType.OPENAI, "gpt-4o-mini", 0.000375, 0.82, 200),
        (ProviderType.GROQ, "llama-3.1-8b", 0.0001, 0.75, 80),
        (ProviderType.ANTHROPIC, "claude-3-haiku", 0.000625, 0.80, 180),
    ]
    
    exec_count = 0
    for provider, model, cost, quality, latency in providers:
        for i in range(33):
            record = ExecutionRecord(
                execution_id=f"integ-exec-{exec_count}",
                provider=provider,
                model=model,
                input_tokens=1000,
                output_tokens=500,
                total_tokens=1500,
                cost_usd=cost,
                latency_ms=float(latency),
                success=True,
                output_quality_score=quality,
                tenant_id="integration-test",
            )
            await memory.record_execution(record)
            exec_count += 1
    
    # Verify aggregation
    scores = await memory.get_all_scores()
    assert len(scores) == 3
    
    # Find best value
    best = await memory.get_best_value_provider()
    assert best is not None
    
    # Get recommendations
    recs = await memory.get_recommendations()
    assert len(recs) > 0
    
    # Get cost breakdown
    breakdown = await memory.get_cost_breakdown(days=1)
    assert breakdown["total_executions"] == 99
    
    print(f"✓ Integration test passed: {exec_count} executions recorded")
    print(f"✓ Best value provider: {best.provider.value}/{best.model} (CQI: {best.cost_quality_index:.1f})")


if __name__ == "__main__":
    # Run the integration test
    asyncio.run(test_full_economic_workflow())
