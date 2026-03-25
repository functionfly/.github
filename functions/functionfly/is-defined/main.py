def handler(event):
    if "value" not in event:
        return {"ok": False, "error": "value is required"}
    value = event.get("value")
    result = value is not None
    return {"ok": True, "value": value, "result": result}
