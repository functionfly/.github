#!/usr/bin/env python3
"""
Test script for FlyPy cold start optimization.
"""

import flypy
from flypy.cold_start_optimizer import optimize_function_cold_start, warmup_all_functions

# Sample function with cold start optimization
@flypy.function(
    name="cold-start-test",
    description="Function with cold start optimization",
    optimize_cold_start=True,
    warmup_data=[
        {"input": "test1", "expected": "processed_test1"},
        {"input": "test2", "expected": "processed_test2"},
    ]
)
def sample_function(data: dict) -> dict:
    """
    Sample function for cold start testing.
    """
    # Simulate some processing time
    import time
    time.sleep(0.01)  # 10ms delay

    result = f"processed_{data.get('input', 'unknown')}"
    return {
        "result": result,
        "timestamp": time.time(),
        "input_length": len(str(data))
    }

if __name__ == "__main__":
    # Get the function definition
    func_def = flypy.get_function_definition("cold-start-test")
    if not func_def:
        print("Function not found")
        exit(1)

    print("Testing cold start optimization...")

    # Test cold start optimization
    dependencies = ["time", "json"]  # Mock dependencies
    warmup_data = [
        {"input": "warmup1"},
        {"input": "warmup2"},
    ]

    print("\nOptimizing cold start...")
    optimization_results = optimize_function_cold_start(
        "cold-start-test",
        func_def.source_code,
        dependencies,
        warmup_data
    )

    print("Cold start optimization results:")
    print(f"- Function: {optimization_results['function_name']}")
    print(f"- Optimizations applied: {optimization_results['optimizations_applied']}")
    print(f"- Estimated time saved: {optimization_results['estimated_time_saved_ms']}ms")

    if optimization_results.get('preload_results'):
        preload = optimization_results['preload_results']
        print(f"- Dependencies preloaded: {preload['modules_loaded']}")
        print(f"- Preload time: {preload['preload_time']:.3f}s")

    print("\n✅ Cold start optimization test completed!")