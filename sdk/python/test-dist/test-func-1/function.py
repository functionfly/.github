@flypy.function(name="test-func-1")
def test_function_1(data: dict) -> dict:
    """Test function 1."""
    return {"result": data.get("input", "") + "_processed_1"}
