#!/usr/bin/env python3
"""
Test script for FlyPy build optimization.
"""

import flypy
from flypy.build_optimizer import optimize_build_process, clear_build_cache

# Create some test functions
@flypy.function(name="test-func-1")
def test_function_1(data: dict) -> dict:
    """Test function 1."""
    return {"result": data.get("input", "") + "_processed_1"}

@flypy.function(name="test-func-2")
def test_function_2(data: dict) -> dict:
    """Test function 2."""
    return {"result": data.get("input", "") + "_processed_2"}

@flypy.function(name="test-func-3")
def test_function_3(data: dict) -> dict:
    """Test function 3."""
    return {"result": data.get("input", "") + "_processed_3"}

if __name__ == "__main__":
    # Get registered functions
    functions = flypy.get_registered_functions()
    if not functions:
        print("No functions registered")
        exit(1)

    print(f"Testing build optimization with {len(functions)} functions...")

    # Convert to format expected by optimizer
    func_defs = []
    for func_name, func_def in functions.items():
        func_dict = {
                "metadata": func_def.metadata.model_dump(),
            "source": {
                "code": func_def.source_code,
                "file": func_def.metadata.source_file,
                "dependencies": func_def.dependencies,
                "imports": func_def.imports,
            },
            "ast": None
        }
        func_defs.append(func_dict)

    # Test build optimization
    build_config = {
        "mode": "deterministic",
        "optimize_bundle": True,
        "optimization_level": "balanced",
        "optimize_cold_start": True,
    }

    print("\nRunning build optimization...")
    results = optimize_build_process(
        func_defs,
        "/usr/local/bin/flypy-go",  # Mock Go binary path
        "./test-dist",
        build_config,
        enable_parallel=True,
        enable_incremental=True,
        verbose=True
    )

    print("\nBuild optimization results:")
    print(f"- Success: {results['success']}")
    print(f"- Total functions: {results['total_functions']}")
    print(f"- Successful builds: {results['successful_builds']}")
    print(f"- Failed builds: {results['failed_builds']}")
    print(f"- Cached builds: {results['cached_builds']}")
    print(f"- Total build time: {results['total_build_time']:.3f}s")
    print(f"- Estimated time saved: {results['estimated_time_saved_ms']}ms")
    print(f"- Dependency levels: {results['dependency_levels']}")

    print("\n✅ Build optimization test completed!")

    # Test cache clearing
    print("\nClearing build cache...")
    clear_build_cache()
    print("✅ Build cache cleared!")