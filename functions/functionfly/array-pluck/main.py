def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    key = event.get("key")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if not key or not isinstance(key, str):
        return {"ok": False, "error": "key is required and must be a string"}
    result = []
    for x in items:
        if isinstance(x, dict) and key in x:
            result.append(x[key])
        else:
            result.append(None)
    return {"ok": True, "result": result}
