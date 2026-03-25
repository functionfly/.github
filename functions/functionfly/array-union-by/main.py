def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    other = event.get("other")
    key = event.get("key")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if other is None:
        return {"ok": False, "error": "other is required"}
    if not isinstance(other, (list, tuple)):
        return {"ok": False, "error": "other must be an array"}
    if not key or not isinstance(key, str):
        return {"ok": False, "error": "key is required and must be a string"}

    seen = set()
    result = []
    for x in list(items) + list(other):
        if not isinstance(x, dict):
            return {"ok": False, "error": "all elements must be objects when using key comparator"}
        k = x.get(key)
        if k not in seen:
            seen.add(k)
            result.append(x)
    return {"ok": True, "result": result}
