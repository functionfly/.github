def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    if isinstance(value, bool):
        return {"ok": True, "value": value, "result": False}
    if isinstance(value, int):
        return {"ok": True, "value": value, "result": True}
    if isinstance(value, float):
        result = value == int(value)
        return {"ok": True, "value": value, "result": result}
    try:
        int(str(value))
        result = '.' not in str(value)
    except (ValueError, TypeError):
        result = False
    return {"ok": True, "value": value, "result": result}
