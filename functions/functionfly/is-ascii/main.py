def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    try:
        result = str(value).isascii()
    except Exception:
        result = False
    return {"ok": True, "value": value, "result": result}
