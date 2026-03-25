def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    if isinstance(value, float):
        return {"ok": True, "value": value, "result": True}
    if isinstance(value, int):
        return {"ok": True, "value": value, "result": False}
    try:
        f = float(str(value))
        result = '.' in str(value) or 'e' in str(value).lower()
    except (ValueError, TypeError):
        result = False
    return {"ok": True, "value": value, "result": result}
