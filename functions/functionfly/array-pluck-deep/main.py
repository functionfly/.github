def _get_path(obj, path):
    parts = path.split(".")
    for p in parts:
        if obj is None:
            return None
        if isinstance(obj, dict) and p in obj:
            obj = obj[p]
        else:
            return None
    return obj


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    path = event.get("path")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if not path or not isinstance(path, str):
        return {"ok": False, "error": "path is required and must be a string"}
    result = [_get_path(x, path) for x in items]
    return {"ok": True, "result": result}
