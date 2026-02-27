#!/usr/bin/env python3
"""
Test script for FlyPy performance monitoring.
"""

import flypy
import time
import random

# Sample functions with performance monitoring
@flypy.function(
    name="fast-function",
    description="Fast function for testing",
    enable_performance_monitoring=True
)
def fast_function(data: dict) -> dict:
    """A fast function that completes quickly."""
    time.sleep(0.01)  # 10ms delay
    return {"result": data.get("input", "") + "_fast", "processed": True}

@flypy.function(
    name="slow-function",
    description="Slow function for testing",
    enable_performance_monitoring=True
)
def slow_function(data: dict) -> dict:
    """A slow function that takes time."""
    time.sleep(0.1)  # 100ms delay
    return {"result": data.get("input", "") + "_slow", "processed": True}

@flypy.function(
    name="error-function",
    description="Function that sometimes errors",
    enable_performance_monitoring=True
)
def error_function(data: dict) -> dict:
    """A function that randomly fails."""
    if random.random() < 0.3:  # 30% chance of error
        raise ValueError("Random error for testing")

    time.sleep(0.05)  # 50ms delay
    return {"result": data.get("input", "") + "_error_test", "processed": True}

if __name__ == "__main__":
    print("Testing FlyPy performance monitoring...")

    # Run some test calls
    print("\nRunning test function calls...")

    test_data = {"input": "test_data"}

    for i in range(10):
        try:
            fast_function(test_data)
            slow_function(test_data)
            error_function(test_data)
        except Exception:
            pass  # Expected errors

    print("✅ Test calls completed.")

    # Get performance stats
    print("\n📊 Performance Statistics:")
    stats = flypy.get_performance_stats()

    print(f"Total functions monitored: {stats['summary']['total_functions']}")
    print(f"Total calls: {stats['summary']['total_calls']}")
    print(f"Total errors: {stats['summary']['total_errors']}")

    for func_name, func_stats in stats['functions'].items():
        print(f"\nFunction: {func_name}")
        print(f"  Calls: {func_stats['total_calls']}")
        print(f"  Errors: {func_stats['error_count']}")
        if 'execution_time' in func_stats:
            exec_time = func_stats['execution_time']
            print(f"  Avg Time: {exec_time['mean']:.3f}s")
            print(f"  Min Time: {exec_time['min']:.3f}s")
            print(f"  Max Time: {exec_time['max']:.3f}s")
            print(f"  P95 Time: {exec_time['p95']:.3f}s")

    # Generate performance report
    print("\n📋 Performance Report:")
    report = flypy.get_performance_report("text")
    print(report)

    # Check for alerts
    print("\n🚨 Checking for alerts...")
    alerts = flypy.check_performance_alerts()

    if alerts:
        print(f"Found {len(alerts)} alert(s):")
        for alert in alerts:
            print(f"  {alert['severity'].upper()}: {alert['message']}")
    else:
        print("No alerts found.")

    print("\n✅ Performance monitoring test completed!")

    # Note: Dashboard and background monitoring would be tested separately
    # as they run continuously