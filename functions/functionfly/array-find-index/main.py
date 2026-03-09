def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    value = event.get("value")
    from_index = event.get("from_index", 0)
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if "value" not in event:
        return {"ok": False, "error": "value is required"}
    try:
        from_index = int(from_index)
    except (TypeError, ValueError):
        from_index = 0
    arr = list(items)
    if from_index < 0:
        from_index = max(0, len(arr) + from_index)
    try:
        index = arr.index(value, from_index)
    except ValueError:
        index = -1
    return {"ok": True, "index": index}
