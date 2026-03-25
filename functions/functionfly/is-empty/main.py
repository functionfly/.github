def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if "value" not in event:
        return {"ok": False, "error": "value is required"}

    if value is None:
        result = True
    elif isinstance(value, (str, list, dict, tuple, set)):
        result = len(value) == 0
    elif isinstance(value, (int, float, bool)):
        result = False
    else:
        result = not bool(value)

    return {"ok": True, "value": value, "result": result}
