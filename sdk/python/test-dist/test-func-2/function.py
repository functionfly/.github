@flypy.function(name="test-func-2")
def test_function_2(data: dict) -> dict:
    """Test function 2."""
    return {"result": data.get("input", "") + "_processed_2"}
