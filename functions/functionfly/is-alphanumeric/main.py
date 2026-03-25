def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value)
    result = val.isalnum() and len(val) > 0
    return {"ok": True, "value": value, "result": result}
