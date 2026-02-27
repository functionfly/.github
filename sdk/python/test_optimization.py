#!/usr/bin/env python3
"""
Test script for FlyPy bundle optimization.
"""

import flypy
from flypy.optimizer import optimize_bundle, analyze_bundle_size

# Sample function with optimization opportunities
@flypy.function(name="test-optimization")
def sample_function(data: dict) -> dict:
    """
    Sample function with various optimization opportunities.
    """
    # Unused import (will be removed)
    import math
    import json

    # Dead code (will be removed)
    unused_variable = "this is never used"
    another_unused = 42

    # Used code (will be kept)
    result = []
    for item in data.get("items", []):
        processed = {
            "id": item["id"],
            "value": item.get("value", 0) * 2,
            "name": item["name"].upper()
        }
        result.append(processed)

    # Constant folding opportunity
    total = 10 + 20 + 30

    return {
        "processed_items": result,
        "total": total,
        "count": len(result)
    }

if __name__ == "__main__":
    # Get the function definition
    func_def = flypy.get_function_definition("test-optimization")
    if not func_def:
        print("Function not found")
        exit(1)

    print("Original source code:")
    print("=" * 50)
    print(func_def.source_code)
    print("=" * 50)

    # Test optimization
    print("\nOptimizing bundle...")
    optimized_code, stats = optimize_bundle(
        func_def.source_code,
        func_def.metadata.model_dump(),
        func_def.dependencies,
        optimization_level="balanced"
    )

    print("Optimization stats:")
    print(f"- Original size: {stats['original_size']} bytes")
    print(f"- Optimized size: {stats['optimized_size']} bytes")
    print(f"- Code removed: {stats['code_removed']} bytes")
    print(f"- Optimizations applied: {stats['optimizations_applied']}")

    print("\nOptimized source code:")
    print("=" * 50)
    print(optimized_code)
    print("=" * 50)

    # Test bundle analysis
    print("\nBundle analysis:")
    analysis = analyze_bundle_size(optimized_code, func_def.metadata.model_dump())
    print(f"- Bundle size: {analysis['bundle_size_bytes']} bytes ({analysis['bundle_size_kb']:.1f} KB)")
    print(f"- Category: {analysis['category']}")
    print(f"- Recommendations: {len(analysis['recommendations'])}")
    for rec in analysis['recommendations']:
        print(f"  - {rec['type'].upper()}: {rec['message']}")

    print("\n✅ Bundle optimization test completed!")