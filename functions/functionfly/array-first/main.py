def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    default = event.get("default")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if len(items) == 0:
        return {"ok": True, "result": default}
    return {"ok": True, "result": items[0]}
