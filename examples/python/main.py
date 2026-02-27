import json
import os

def handler(event):
    """
    Function entry point.
    Receives: dict or string
    Returns: dict or string
    """
    # Access environment
    api_key = os.environ.get("API_KEY", "")

    # Process input
    if isinstance(event, dict):
        name = event.get("name", "World")
    else:
        name = str(event)

    # Return result
    return {
        "message": f"Hello, {name}!",
        "api_key_set": bool(api_key),
        "runtime": "python-wasm",
        "timestamp": "2026-02-18",  # Current date
    }