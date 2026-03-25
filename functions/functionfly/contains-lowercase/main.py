def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value)
    result = any(c.islower() for c in val)
    return {"ok": True, "value": value, "result": result}
