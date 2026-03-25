def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    val = str(value).strip()
    try:
        float(val)
        result = True
    except ValueError:
        result = False
    return {"ok": True, "value": value, "result": result}
