#!/usr/bin/env python3
"""
Basic test for FlyPy SDK functionality.
"""

import sys
import os

# Add the flypy package to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'flypy'))

import flypy
from flypy.decorators import get_registered_functions, clear_registry


def test_basic_decorator():
    """Test basic function decorator."""
    print("Testing basic decorator...")

    @flypy.function(name="test-func", deterministic=True)
    def test_function(x: int, y: int) -> int:
        return x + y

    # Check that function was registered
    functions = get_registered_functions()
    assert "test-func" in functions

    func_def = functions["test-func"]
    assert func_def.metadata.name == "test-func"
    assert func_def.metadata.deterministic == True
    assert func_def.source_code is not None

    print("✓ Basic decorator test passed")


def test_schema_decorators():
    """Test schema decorators."""
    print("Testing schema decorators...")

    input_schema = {"type": "object", "properties": {"a": {"type": "number"}, "b": {"type": "number"}}}
    output_schema = {"type": "object", "properties": {"result": {"type": "number"}}}

    @flypy.input_schema(input_schema)
    @flypy.output_schema(output_schema)
    @flypy.function(name="test-schema")
    def test_function(data: dict) -> dict:
        return {"result": data["a"] + data["b"]}

    functions = get_registered_functions()
    func_def = functions["test-schema"]

    assert func_def.metadata.input_schema == input_schema
    assert func_def.metadata.output_schema == output_schema

    print("✓ Schema decorators test passed")


def test_schema_inference():
    """Test automatic schema inference."""
    print("Testing schema inference...")

    from flypy.schema import Schema
    from typing import Dict

    def test_func(x: str) -> Dict[str, int]:
        return {"result": int(x)}

    # Test inference from type hints
    input_schema, output_schema = Schema.infer_from_function(test_func)

    assert "x" in input_schema.properties
    assert input_schema.properties["x"].type == "string"
    assert "result" in output_schema.properties

    print("✓ Schema inference test passed")


def run_tests():
    """Run all tests."""
    print("Running FlyPy SDK tests...\n")

    try:
        # Clear any existing functions
        clear_registry()

        test_basic_decorator()
        test_schema_decorators()
        test_schema_inference()

        print("\n🎉 All tests passed!")

    except Exception as e:
        print(f"\n❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(run_tests())