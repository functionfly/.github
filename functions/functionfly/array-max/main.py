def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if not items:
        return {"ok": False, "error": "items must not be empty"}
    try:
        result = max(items)
    except TypeError as e:
        return {"ok": False, "error": str(e)}
    return {"ok": True, "result": result}
