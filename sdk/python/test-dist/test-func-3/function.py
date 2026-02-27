@flypy.function(name="test-func-3")
def test_function_3(data: dict) -> dict:
    """Test function 3."""
    return {"result": data.get("input", "") + "_processed_3"}
