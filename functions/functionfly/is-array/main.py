def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None and "value" not in event:
        return {"ok": False, "error": "value is required"}
    result = isinstance(value, list)
    length = len(value) if result else None
    return {"ok": True, "value": value, "result": result, "length": length}
