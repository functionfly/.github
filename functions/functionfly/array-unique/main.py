def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    seen = set()
    result = []
    for x in items:
        try:
            k = (x, type(x).__name__) if not isinstance(x, (str, int, float, bool, type(None))) else x
        except TypeError:
            k = id(x)
        if k not in seen:
            seen.add(k)
            result.append(x)
    return {"ok": True, "result": result}
