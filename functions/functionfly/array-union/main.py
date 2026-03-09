def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    other = event.get("other")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if other is None:
        return {"ok": False, "error": "other is required"}
    if not isinstance(other, (list, tuple)):
        return {"ok": False, "error": "other must be an array"}
    seen = set()
    result = []
    for x in list(items) + list(other):
        try:
            k = x
            if k not in seen:
                seen.add(k)
                result.append(x)
        except TypeError:
            result.append(x)
    return {"ok": True, "result": result}
