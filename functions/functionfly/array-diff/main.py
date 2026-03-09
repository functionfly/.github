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
        result = [x for x in items if x not in other_set]
    except TypeError:
        other_list = list(other)
        result = [x for x in items if x not in other_list]
    return {"ok": True, "result": result}
