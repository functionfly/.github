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
    try:
        other_set = set(other)
        seen = set()
        result = []
        for x in items:
            if x in other_set and x not in seen:
                seen.add(x)
                result.append(x)
    except TypeError:
        other_list = list(other)
        result = []
        for x in items:
            if x in other_list and x not in result:
                result.append(x)
    return {"ok": True, "result": result}
