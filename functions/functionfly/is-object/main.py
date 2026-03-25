def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None and "value" not in event:
        return {"ok": False, "error": "value is required"}
    result = isinstance(value, dict)
    keys = list(value.keys()) if result else None
    return {"ok": True, "value": value, "result": result, "keys": keys}
