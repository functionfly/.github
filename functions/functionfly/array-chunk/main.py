def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    size = event.get("size")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if size is None:
        return {"ok": False, "error": "size is required"}
    try:
        size = int(size)
    except (TypeError, ValueError):
        return {"ok": False, "error": "size must be an integer"}
    if size < 1:
        return {"ok": False, "error": "size must be at least 1"}
    items = list(items)
    chunks = [items[i : i + size] for i in range(0, len(items), size)]
    return {"ok": True, "chunks": chunks}
