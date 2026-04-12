"""Test script for cost-optimized function generation.

Run this to verify the optimized generation pipeline works correctly.
"""

import asyncio
import json
import sys
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from services.generation import (
    get_model_router,
    get_validation_pipeline,
    get_function_rag_retriever,
    get_generation_cache,
    get_optimized_generation_service,
    ModelTier,
)
from models.schemas import FunctionGenerationRequest, ProviderType


async def test_complexity_analyzer():
    """Test complexity analysis and routing."""
    print("\n=== Testing Complexity Analyzer ===")

    router = get_model_router()

    test_cases = [
        ("Summarize text from emails", None, ModelTier.CHEAP),
        ("Parse JSON and validate schema", None, ModelTier.CHEAP),
        ("Create a REST API client with authentication", None, ModelTier.MID),
        ("Build a workflow that processes data through multiple stages", None, ModelTier.MID),
        ("Implement machine learning pipeline with vector embeddings", None, ModelTier.PREMIUM),
        ("Create a secure multi-step authentication workflow", None, ModelTier.PREMIUM),
    ]

    for description, constraints, expected_tier in test_cases:
        routing = router.route(description, constraints)
        status = "✓" if routing.tier == expected_tier else "✗"
        print(f"{status} '{description[:40]}...' -> {routing.tier.value} (expected: {expected_tier.value})")
        print(f"   Confidence: {routing.confidence:.2f}, Est. cost: ${routing.estimated_cost_usd:.4f}")


async def test_validation_pipeline():
    """Test validation pipeline."""
    print("\n=== Testing Validation Pipeline ===")

    validator = get_validation_pipeline()

    # Valid Python code
    valid_code = """
def handler(input):
    \"\"\"Handle incoming request.\"\"\"\n    data = input.get("data", {})\n    return {"result": data}
"""

    # Invalid Python code
    invalid_code = """
def handler(input)\n    data = input.get("data"  # Missing closing bracket
    return {"result" data}  # Missing colon
"""

    test_cases = [
        ("python", valid_code, True),
        ("python", invalid_code, False),
    ]

    for runtime, code, should_pass in test_cases:
        report = validator.validate(code, runtime, skip_runtime=True)
        status = "✓" if report.overall_passed == should_pass else "✗"
        print(f"{status} {runtime} code validation: confidence={report.confidence_score:.2f}")
        if not report.overall_passed and report.stages:
            print(f"   Errors: {[e for s in report.stages for e in s.errors][:2]}")


async def test_template_retrieval():
    """Test template matching."""
    print("\n=== Testing Template Retrieval ===")

    rag = get_function_rag_retriever()

    test_cases = [
        ("Create a webhook endpoint to receive events", "python"),
        ("Make API calls to external services", "nodejs"),
        ("Transform and parse data structures", "python"),
        ("Verify JWT tokens for authentication", "python"),
    ]

    for description, runtime in test_cases:
        template = rag.find_template(description, runtime)
        if template:
            print(f"✓ Found template: {template.template_name} (score: {template.match_score:.2f})")
            print(f"   Tokens saved: ~{template.estimated_tokens_saved}")
        else:
            print(f"✗ No matching template for: {description[:40]}")


async def test_caching():
    """Test generation caching."""
    print("\n=== Testing Generation Cache ===")

    cache = get_generation_cache()

    # Test cache key generation
    key1 = cache._generate_key("Summarize text from emails", "python", None)
    key2 = cache._generate_key("summarize and email text from", "python", None)

    print(f"✓ Cache key normalization working: {key1 == key2}")

    # Test local cache
    await cache.set(
        description="Test function",
        runtime="python",
        code="def handler(): pass",
        manifest={"name": "test"},
        explanation="Test",
        complexity="simple",
    )

    cached = await cache.get("Test function", "python")
    if cached:
        print(f"✓ Cache retrieval working: {cached.cache_key}")
    else:
        print("✗ Cache retrieval failed")


async def test_end_to_end():
    """Test the full optimized generation pipeline."""
    print("\n=== Testing End-to-End Generation ===")

    service = get_optimized_generation_service()

    # Simple request that should use cheap tier + template
    simple_request = FunctionGenerationRequest(
        description="Create a webhook handler that receives POST requests and processes the payload",
        runtime="python",
    )

    # Complex request that should escalate to premium
    complex_request = FunctionGenerationRequest(
        description="Build a multi-step workflow that orchestrates distributed transactions with rollback capability",
        runtime="python",
    )

    print("\nTesting simple request (should use cheap tier or template)...")
    try:
        response, metrics = await service.generate(
            request=simple_request,
            tenant_id="test",
        )
        print(f"✓ Generation success: {response.success}")
        print(f"  Tier: {metrics.final_tier}, Cache hit: {metrics.cache_hit}")
        print(f"  Template used: {metrics.template_used}, Cost: ${metrics.total_cost_usd:.4f}")
        if response.result:
            print(f"  Code length: {len(response.result.code)} chars")
    except Exception as e:
        print(f"✗ Simple generation failed: {e}")

    print("\nTesting complex request (should use premium tier)...")
    try:
        response, metrics = await service.generate(
            request=complex_request,
            tenant_id="test",
        )
        print(f"✓ Generation success: {response.success}")
        print(f"  Tier: {metrics.final_tier}, Attempts: {metrics.total_attempts}")
        print(f"  Cost: ${metrics.total_cost_usd:.4f}")
    except Exception as e:
        print(f"✗ Complex generation failed: {e}")


async def main():
    """Run all tests."""
    print("=" * 60)
    print("Cost-Optimized Function Generation Test Suite")
    print("=" * 60)

    try:
        await test_complexity_analyzer()
    except Exception as e:
        print(f"Complexity analyzer test failed: {e}")

    try:
        await test_validation_pipeline()
    except Exception as e:
        print(f"Validation pipeline test failed: {e}")

    try:
        await test_template_retrieval()
    except Exception as e:
        print(f"Template retrieval test failed: {e}")

    try:
        await test_caching()
    except Exception as e:
        print(f"Caching test failed: {e}")

    print("\n" + "=" * 60)
    print("Tests complete. Note: End-to-end tests require running services.")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
