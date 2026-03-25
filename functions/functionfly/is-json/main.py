import json


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    try:
        parsed = json.loads(str(value))
        return {"ok": True, "value": value, "result": True, "type": type(parsed).__name__}
    except (json.JSONDecodeError, TypeError):
        return {"ok": True, "value": value, "result": False}
