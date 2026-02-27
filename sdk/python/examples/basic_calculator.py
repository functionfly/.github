"""
Basic calculator example for FlyPy.

This example demonstrates a simple deterministic function that performs
mathematical calculations.
"""

import flypy


@flypy.function(
    name="calculate",
    description="Perform mathematical calculations",
    deterministic=True,
    idempotent=True,
    pure=True,
    cache_ttl=3600
)
def calculate(operation: str, a: float, b: float) -> dict:
    """
    Perform a mathematical operation on two numbers.

    Args:
        operation: The operation to perform ("add", "subtract", "multiply", "divide")
        a: First number
        b: Second number

    Returns:
        Dictionary with the result
    """
    if operation == "add":
        result = a + b
    elif operation == "subtract":
        result = a - b
    elif operation == "multiply":
        result = a * b
    elif operation == "divide":
        if b == 0:
            raise ValueError("Division by zero")
        result = a / b
    else:
        raise ValueError(f"Unknown operation: {operation}")

    return {
        "operation": operation,
        "a": a,
        "b": b,
        "result": result
    }


@flypy.function(
    name="batch-calculate",
    description="Perform multiple calculations",
    deterministic=True,
    idempotent=True
)
def batch_calculate(calculations: list) -> dict:
    """
    Perform multiple calculations in batch.

    Args:
        calculations: List of calculation requests

    Returns:
        Dictionary with all results
    """
    results = []
    total = 0

    for calc in calculations:
        result = calculate(calc["operation"], calc["a"], calc["b"])
        results.append(result)
        total += result["result"]

    return {
        "results": results,
        "total": total,
        "count": len(results)
    }